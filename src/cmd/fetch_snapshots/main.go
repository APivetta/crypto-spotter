package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"

	"pivetta.se/crypro-spotter/src/ingestors"
	"pivetta.se/crypro-spotter/src/repositories"
	"pivetta.se/crypro-spotter/src/utils"
)

func main() {
	count := flag.Int("count", 1, "Number of symbols to fetch")
	flag.Parse()

	db := utils.GetDb()

	bi := ingestors.BinanceIngestor{
		Url: ingestors.LIVE,
	}

	s, err := bi.GetSymbols(*count)
	if err != nil {
		log.Fatalf("Error fetching symbols: %v\n", err)
	}

	for _, symbol := range s {
		fmt.Printf("Fetching Symbol: %s\n", symbol)
		fetchSnapshots(db, symbol, bi)
	}
}

func fetchSnapshots(db *sql.DB, symbol string, bi ingestors.BinanceIngestor) {
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

	ss := bi.GetHistory(symbol, date)
	repositories.InsertSnapshots(db, symbol, ss)
}
