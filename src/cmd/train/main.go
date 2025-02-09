package main

import (
	"flag"
	"log"

	"pivetta.se/crypro-spotter/src/genetics"
	"pivetta.se/crypro-spotter/src/repositories"
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

	err = genetics.RunGenetic(repo, asset)
	if err != nil {
		log.Fatalf("Error running genetic algorithm: %v", err)
	}
}
