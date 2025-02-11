package main

import (
	"log"
	"os"

	"github.com/cinar/indicator/v2/helper"
	"pivetta.se/crypro-spotter/src/ingestors"
	"pivetta.se/crypro-spotter/src/strategies"
)

func main() {
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

	b, err := bi.GetBalance()
	if err != nil {
		log.Fatalf("Error getting balance: %v", err)
	}
	log.Printf("Balance: %v", b)
}

func liveRun() {
	bi := ingestors.BinanceIngestor{
		Url: ingestors.LIVE,
	}
	bd, err := bi.Poll()
	if err != nil {
		log.Fatalf("Error polling Binance: %v", err)
	}

	btc := bd[0]

	scalp := strategies.Scalping{
		Weights: strategies.StrategyWeights{
			SuperTrendWeight:  0.11305080149086622,
			BollingerWeight:   2.5045589544222526,
			EmaWeight:         0.029109615148465218,
			RsiWeight:         2.5698574046573275,
			MacdWeight:        1.284617282123353,
			StrengthThreshold: 3.5556241277450615,
			AtrMultiplier:     2.740351814448069,
		},
		Stabilization: 100,
	}

	ac := scalp.Compute(helper.Buffered(btc.Klines, 50))

	for a := range ac {
		log.Printf("Action: %v", a.Annotation())
	}
}
