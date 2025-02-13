package utils

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"sync"
)

var db *sql.DB
var once sync.Once

func GetDb() *sql.DB {
	once.Do(func() {
		host := flag.String("host", "localhost", "Database host")
		port := flag.Int("port", 5433, "Database port")
		user := flag.String("user", "postgres", "Database user")
		password := flag.String("password", "postgres", "Database password")
		dbname := flag.String("dbname", "spotter", "Database name")
		flag.Parse()

		var err error
		db, err = connectDB(*host, *port, *user, *password, *dbname)
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
