package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cinar/indicator/v2/helper"
	"github.com/cinar/indicator/v2/strategy"
	"pivetta.se/crypro-spotter/src/genetics"
	"pivetta.se/crypro-spotter/src/ingestors"
	"pivetta.se/crypro-spotter/src/repositories"
	"pivetta.se/crypro-spotter/src/strategies"
	"pivetta.se/crypro-spotter/src/utils"
)

// yes i know, global bad
var outcome float64
var asset *string

// this is dup code, find a place to put it and consolidate
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

// end of dup code

func main() {
	// Create a channel to listen for OS signals
	sigChan := make(chan os.Signal, 1)
	// Notify the channel when receiving an Interrupt (Ctrl+C) or Termination signal
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	// Run a goroutine to handle the signal
	go func() {
		sig := <-sigChan
		fmt.Println("\nReceived signal:", sig)
		cleanup()
		os.Exit(0)
	}()

	asset = flag.String("asset", "BTCUSDT", "Asset to backtest")
	flag.Parse()
	apiKey := os.Getenv("API_KEY")
	apiSecret := os.Getenv("API_SECRET")

	if apiKey == "" || apiSecret == "" {
		log.Fatalf("API_KEY and API_SECRET must be set")
	}

	bi := ingestors.BinanceIngestor{
		Url:    ingestors.LIVE,
		Key:    apiKey,
		Secret: apiSecret,
	}
	db := utils.GetDb()

	for {
		fetchSnapshots(db, *asset, bi)
		geneticsRun(3, *asset)
		liveRun(bi, *asset)
	}
	// b, err := bi.GetBalance()
	// if err != nil {
	// 	log.Fatalf("Error getting balance: %v", err)
	// }
	// log.Printf("Balance: %v", b)
}

func liveRun(bi ingestors.BinanceIngestor, asset string) {
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	bd, err := bi.Poll(asset)
	if err != nil {
		log.Fatalf("Error polling Binance: %v", err)
	}

	w, err := utils.GetLatestWeights("BTCUSDT")
	if err != nil {
		log.Fatalf("Error getting weights: %v", err)
	}

	scalp := strategies.Scalping{
		Weights:       *w,
		Stabilization: 299,
		WithSL:        true,
	}

	ss := helper.Duplicate(bd.Klines, 2)
	ac, oc := scalp.ComputeWithOutcome(ss[0], true)

	for a := range ac {
		kline := <-ss[1]
		outcome = <-oc
		log.Printf("Action: %v, Price: %.2f, Outcome: %.2f", extendedAnnotation(a), kline.Close, outcome)

		// we should run the compute until midnight, store the outcome and retrain the weights
		if time.Since(startDate) >= 24*time.Hour {
			log.Printf("End of day, retraining weights")
			// TODO: CLOSE ALL POSITIONS when live!!
			break
		}
	}

	cleanup()
}

func cleanup() {
	log.Println("Trade results:", outcome)
	err := storeOutcome(*asset, outcome)
	if err != nil {
		log.Fatalf("Error storing outcome: %v", err)
	}
}

func extendedAnnotation(a strategy.Action) string {
	switch a {
	case strategy.Sell:
		return "SHORT"

	case strategy.Buy:
		return "LONG"

	case strategies.Close:
		return "CLOSE"

	default:
		return "HOLD"
	}
}

func storeOutcome(a string, outcome float64) error {
	db := utils.GetDb()

	query := `INSERT INTO trade_results (asset, date, result) VALUES ($1, $2, $3)`
	_, err := db.Exec(query, a, time.Now(), outcome)
	if err != nil {
		return fmt.Errorf("storeOutcome: %w", err)
	}

	return nil
}
