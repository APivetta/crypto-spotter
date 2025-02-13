package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"pivetta.se/crypro-spotter/src/genetics"
	"pivetta.se/crypro-spotter/src/repositories"
	"pivetta.se/crypro-spotter/src/strategies"
	"pivetta.se/crypro-spotter/src/utils"
)

func main() {
	days := flag.Int("days", 3, "Days to backtest")
	asset := flag.String("asset", "BTCUSDT", "Asset to backtest")
	flag.Parse()

	geneticsRun(*days, *asset)
}

func geneticsRun(days int, asset string) {
	historyMinutes := 24 * 60 * days
	repo, err := repositories.NewDBRepository(asset, historyMinutes)
	if err != nil {
		log.Fatalf("Error creating repository: %v", err)
	}

	best, err := genetics.RunGenetic(repo, asset)
	if err != nil {
		log.Fatalf("Error running genetic algorithm: %v", err)
	}

	log.Printf("Best strategy: %+v", best)
	err = storeWeights(asset, best)
	if err != nil {
		log.Fatalf("Error storing weights: %v", err)
	}
}

func storeWeights(a string, weights *strategies.StrategyWeights) error {
	db := utils.GetDb()

	jsonData, err := json.Marshal(weights)
	if err != nil {
		return err
	}
	genome := string(jsonData) // Store JSON as string in DB

	query := `INSERT INTO genomes (asset, date, genome) VALUES ($1, $2, $3)`
	_, err = db.Exec(query, a, time.Now(), genome)
	if err != nil {
		return fmt.Errorf("insertSnapshots: %w", err)
	}

	return nil
}
