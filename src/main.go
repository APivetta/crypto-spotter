package main

import (
	"log"

	"github.com/cinar/indicator/v2/helper"
	"pivetta.se/crypro-spotter/src/ingestors"
	"pivetta.se/crypro-spotter/src/strategies"
)

func main() {
	bd, err := ingestors.BinancePoller()

	if err != nil {
		log.Fatalf("Error polling Binance: %v", err)
	}

	btc := bd[0]

	scalp := strategies.Scalping{
		LastPrice: btc.LastPrice,
		Weights: strategies.StrategyWeights{
			Ema5Weight:        0.5,
			Ema20Weight:       0.5,
			RsiWeight:         1.2,
			MacdWeight:        1.3,
			SuperTrendWeight:  1.5,
			BollingerWeight:   0.8,
			RsiOverbought:     70.0,
			RsiOversold:       30.0,
			MacdThreshold:     0.8,
			StrengthThreshold: 2,
		},
	}
	scalp.Compute(helper.Buffered(btc.Klines, 50))
}
