package main

import (
	"flag"
	"fmt"
	"log"

	"pivetta.se/crypro-spotter/src/repositories"
	"pivetta.se/crypro-spotter/src/strategies"
	"pivetta.se/crypro-spotter/src/utils"
)

func main() {
	days := flag.Int("days", 1, "Days to backtest")
	asset := flag.String("asset", "BTCUSDT", "Asset to backtest")
	flag.Parse()

	backtestRun(*days, *asset)
}

func backtestRun(days int, asset string) {
	historyMinutes := 24 * 60 * days
	repo, err := repositories.NewDBRepository(asset, historyMinutes+60)
	if err != nil {
		log.Fatalf("Error creating repository: %v", err)
	}

	w, err := utils.GetLatestWeights(asset)
	if err != nil {
		log.Fatalf("Error getting weights: %v", err)
	}

	scalp := strategies.Scalping{
		Weights:       *w,
		Stabilization: 60,
		WithSL:        true,
	}

	r, err := repo.Get(asset)
	if err != nil {
		log.Fatalf("Error getting BTC data: %v", err)
	}

	ac, oc := scalp.ComputeWithOutcome(r, true)
	var outcome float64
	for o := range oc {
		<-ac
		outcome = o
	}

	fmt.Printf("Outcome: %f\n", outcome)
}
