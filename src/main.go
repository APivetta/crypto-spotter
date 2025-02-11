package main

import (
	"log"

	"github.com/cinar/indicator/v2/helper"
	"pivetta.se/crypro-spotter/src/ingestors"
	"pivetta.se/crypro-spotter/src/strategies"
)

func main() {
	liveRun()
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
