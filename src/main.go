package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cinar/indicator/v2/helper"
	"github.com/cinar/indicator/v2/strategy"
	"pivetta.se/crypro-spotter/src/ingestors"
	"pivetta.se/crypro-spotter/src/strategies"
	"pivetta.se/crypro-spotter/src/utils"
)

func main() {
	asset := flag.String("asset", "BTCUSDT", "Asset to backtest")
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

	liveRun(bi, *asset)
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
	var outcome float64
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

	log.Println("Day results:", outcome)
	err = storeOutcome(asset, outcome)
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
