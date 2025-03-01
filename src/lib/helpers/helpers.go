package helpers

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"pivetta.se/crypro-spotter/src/connectors"
	"pivetta.se/crypro-spotter/src/genetics"
	"pivetta.se/crypro-spotter/src/lib/db"
	"pivetta.se/crypro-spotter/src/repositories"
)

func CalculateQuantity(usd float64, price float64) float64 {
	return usd / price
}

func FetchSnapshots(db *sql.DB, symbol string, bc connectors.BinanceConnector) {
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

	ss := bc.GetHistory(symbol, date)
	repositories.InsertSnapshots(db, symbol, ss)
}
func GeneticsRun(days int, asset string) {
	historyMinutes := 24 * 60 * days
	repo, err := repositories.NewDBRepository(asset, historyMinutes+60)
	if err != nil {
		log.Fatalf("Error creating repository: %v", err)
	}

	best, err := genetics.RunGenetic(repo, asset)
	if err != nil {
		log.Fatalf("Error running genetic algorithm: %v", err)
	}

	log.Printf("Best strategy: %+v", best)
	err = db.StoreWeights(asset, best)
	if err != nil {
		log.Fatalf("Error storing weights: %v", err)
	}
}
