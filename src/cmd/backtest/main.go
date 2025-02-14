package main

import (
	"database/sql"
	"encoding/json"
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
	repo, err := repositories.NewDBRepository(asset, historyMinutes)
	if err != nil {
		log.Fatalf("Error creating repository: %v", err)
	}

	w, err := getWeights(asset)
	if err != nil {
		log.Fatalf("Error getting weights: %v", err)
	}

	scalp := strategies.Scalping{
		Weights:       *w,
		Stabilization: 100,
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

func getWeights(a string) (*strategies.StrategyWeights, error) {
	db := utils.GetDb()

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
