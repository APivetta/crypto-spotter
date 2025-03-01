package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	_ "github.com/lib/pq"

	"pivetta.se/crypro-spotter/src/genetics"
	"pivetta.se/crypro-spotter/src/strategies"
)

var db *sql.DB
var once sync.Once

func GetDb() *sql.DB {
	once.Do(func() {
		host := os.Getenv("POSTGRES_HOST")
		port := os.Getenv("POSTGRES_PORT")
		user := os.Getenv("POSTGRES_USER")
		password := os.Getenv("POSTGRES_PASSWORD")
		dbname := os.Getenv("POSTGRES_DB")

		var err error
		p, err := strconv.Atoi(port)
		if err != nil {
			log.Fatalf("Error converting port to int: %v", err)
		}

		db, err = connectDB(host, p, user, password, dbname)
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
	})
	return db
}

func connectDB(host string, port int, user, password, dbname string) (*sql.DB, error) {
	// Build connection string
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	log.Printf("Connecting to database: %s\n", psqlInfo)

	// Connect to the database
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("connectDB: %w", err)
	}

	// Verify connection
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("connectDB: %w", err)
	}

	return db, nil
}

func GetLatestWeights(a string) (*strategies.StrategyWeights, error) {
	db := GetDb()

	var rawJson string
	var weights *strategies.StrategyWeights
	query := `SELECT genome FROM genomes WHERE asset = $1 ORDER BY date DESC LIMIT 1`
	row := db.QueryRow(query, a)
	err := row.Scan(&rawJson)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getWeights: %w", err)
	}

	fmt.Println(rawJson)

	err = json.Unmarshal([]byte(rawJson), &weights)
	if err != nil {
		return nil, fmt.Errorf("getWeights, unmarshal: %w", err)
	}

	return weights, nil
}

func StoreWeights(a string, best *genetics.Score) error {
	db := GetDb()

	jsonData, err := json.Marshal(best.Individual)
	if err != nil {
		return err
	}
	genome := string(jsonData) // Store JSON as string in DB

	query := `INSERT INTO genomes (asset, date, genome, fitness) VALUES ($1, $2, $3, $4)`
	_, err = db.Exec(query, a, time.Now(), genome, best.Value)
	if err != nil {
		return fmt.Errorf("storeWeights: %w", err)
	}

	return nil
}
