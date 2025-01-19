package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/cinar/indicator/v2/asset"
	_ "github.com/lib/pq"
	"pivetta.se/crypro-spotter/src/ingestors"
)

// TODO: implement for more assets later
const ASSET = "BTCUSDT"

func main() {
	host := flag.String("host", "localhost", "Database host")
	port := flag.Int("port", 5433, "Database port")
	user := flag.String("user", "postgres", "Database user")
	password := flag.String("password", "postgres", "Database password")
	dbname := flag.String("dbname", "spotter", "Database name")
	flag.Parse()

	db, err := connectDB(*host, *port, *user, *password, *dbname)
	if err != nil {
		log.Fatalf("connectDB: %v\n", err)
	}
	defer db.Close()

	recentSnapshot, err := getLatestSnapshot(db, ASSET)
	if err != nil {
		log.Fatalf("Error fetching most recent snapshot: %v\n", err)
	}
	fmt.Printf("Most recent snapshot for %s: %+v\n", ASSET, recentSnapshot)

	var date time.Time
	if recentSnapshot == nil {
		date = time.Now().Add(-15 * 24 * time.Hour).Truncate(24 * time.Hour).UTC()
		log.Printf("No snapshots found for %s. Fetching all snapshots since %v\n", ASSET, date)
	} else {
		cleanup(db, ASSET)
		date = recentSnapshot.Date
		log.Printf("Fetching snapshots since %v\n", date)
	}

	ss := ingestors.GetHistory(ASSET, date)
	insertSnapshots(db, ss)
}

func connectDB(host string, port int, user, password, dbname string) (*sql.DB, error) {
	// Build connection string
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	log.Printf("Connecting to database: %s\n", psqlInfo)

	// Connect to the database
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	// Verify connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Unable to reach database: %v\n", err)
	}
	fmt.Println("Successfully connected to database!")
	return db, nil
}

func getLatestSnapshot(db *sql.DB, a string) (*asset.Snapshot, error) {
	query := `SELECT date, open, high, low, close, volume FROM snapshots WHERE asset = $1 ORDER BY date DESC LIMIT 1`
	layout := "2006-01-02T15:04:05Z"

	row := db.QueryRow(query, a)
	var dateStr string
	var snapshot asset.Snapshot
	err := row.Scan(&dateStr, &snapshot.Open, &snapshot.High, &snapshot.Low, &snapshot.Close, &snapshot.Volume)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("getMostRecentSnapshot: %w", err)
	}

	localLocation := time.Now().Location()
	parsed, err := time.ParseInLocation(layout, dateStr, localLocation)
	if err != nil {
		return nil, fmt.Errorf("getMostRecentSnapshot, parse date: %w", err)
	}

	snapshot.Date = parsed

	return &snapshot, nil
}

func insertSnapshots(db *sql.DB, ss chan *asset.Snapshot) error {
	query := `INSERT INTO snapshots (asset, date, open, high, low, close, volume) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	for snapshot := range ss {
		log.Printf("Inserting snapshot: %+v\n", snapshot)
		_, err := db.Exec(query, ASSET, snapshot.Date, snapshot.Open, snapshot.High, snapshot.Low, snapshot.Close, snapshot.Volume)
		if err != nil {
			return fmt.Errorf("insertSnapshots: %w", err)
		}
	}
	return nil
}

func cleanup(db *sql.DB, a string) error {
	date := time.Now().Add(-15 * 24 * time.Hour).Truncate(24 * time.Hour).UTC()
	log.Printf("Cleaning up snapshots before %v\n", date)
	query := `DELETE FROM snapshots WHERE asset = $1 AND date < $2`
	_, err := db.Exec(query, a, date)
	if err != nil {
		return fmt.Errorf("cleanup: %w", err)
	}
	return nil
}
