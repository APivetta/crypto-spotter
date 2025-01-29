package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"

	"pivetta.se/crypro-spotter/src/ingestors"
	"pivetta.se/crypro-spotter/src/repositories"
)

func main() {
	host := flag.String("host", "localhost", "Database host")
	port := flag.Int("port", 5433, "Database port")
	user := flag.String("user", "postgres", "Database user")
	password := flag.String("password", "postgres", "Database password")
	dbname := flag.String("dbname", "spotter", "Database name")
	flag.Parse()

	db, err := repositories.ConnectDB(*host, *port, *user, *password, *dbname)
	if err != nil {
		log.Fatalf("connectDB: %v\n", err)
	}
	defer db.Close()

	s, err := ingestors.GetSymbols(20)
	if err != nil {
		log.Fatalf("Error fetching symbols: %v\n", err)
	}

	for _, symbol := range s {
		fmt.Printf("Fetching Symbol: %s\n", symbol)
		fetchSnapshots(db, symbol)
	}
}

func fetchSnapshots(db *sql.DB, symbol string) {
	recentSnapshot, err := repositories.GetLatestSnapshot(db, symbol)
	if err != nil {
		log.Fatalf("Error fetching most recent snapshot: %v\n", err)
	}
	fmt.Printf("Most recent snapshot for %s: %+v\n", symbol, recentSnapshot)

	var date time.Time
	if recentSnapshot == nil {
		date = time.Now().Add(-15 * 24 * time.Hour).Truncate(24 * time.Hour).UTC()
		log.Printf("No snapshots found for %s. Fetching all snapshots since %v\n", symbol, date)
	} else {
		repositories.Cleanup(db, symbol)
		date = recentSnapshot.Date
		log.Printf("Fetching snapshots since %v\n", date)
	}

	ss := ingestors.GetHistory(symbol, date)
	repositories.InsertSnapshots(db, symbol, ss)
}
