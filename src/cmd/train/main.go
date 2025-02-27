package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"pivetta.se/crypro-spotter/src/connectors"
	"pivetta.se/crypro-spotter/src/genetics"
	"pivetta.se/crypro-spotter/src/repositories"
	"pivetta.se/crypro-spotter/src/utils"
)

func main() {
	days := flag.Int("days", 3, "Days to train")
	count := flag.Int("count", 1, "Number of symbols to train")
	flag.Parse()

	bc := connectors.BinanceConnector{
		Url: connectors.LIVE,
	}

	s, err := bc.GetSymbols(*count)
	if err != nil {
		log.Fatalf("Error fetching symbols: %v\n", err)
	}

	for _, symbol := range s {
		fmt.Printf("Training Symbol: %s\n", symbol)
		geneticsRun(*days, symbol)
	}

}

func geneticsRun(days int, asset string) {
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
	err = storeWeights(asset, best)
	if err != nil {
		log.Fatalf("Error storing weights: %v", err)
	}
}

func storeWeights(a string, best *genetics.Score) error {
	db := utils.GetDb()

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
