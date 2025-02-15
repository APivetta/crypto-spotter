package main

import (
	"flag"
	"log"
	"os"

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
		Url:    ingestors.TESTNET,
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

	ac := scalp.Compute(helper.Buffered(bd.Klines, 50))

	for a := range ac {
		log.Printf("Action: %v", ExtendedAnnotation(a))
	}
}

func ExtendedAnnotation(a strategy.Action) string {
	switch a {
	case strategy.Sell:
		return "SHORT"

	case strategy.Buy:
		return "LONG"

	case strategies.Close:
		return "CLOSE"

	default:
		return ""
	}
}
