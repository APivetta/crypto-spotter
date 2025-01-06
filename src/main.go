package main

import (
	"log"
	"time"

	"github.com/cinar/indicator/v2/asset"
	"github.com/cinar/indicator/v2/backtest"
	"github.com/cinar/indicator/v2/strategy"
	"pivetta.se/crypro-spotter/src/ingestors"
	"pivetta.se/crypro-spotter/src/reports"
	"pivetta.se/crypro-spotter/src/strategies"
)

func main() {
	// bd, err := ingestors.BinancePoller()
	// if err != nil {
	// 	log.Fatalf("Error polling Binance: %v", err)
	// }

	// btc := bd[0]

	scalp := strategies.Scalping{
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
		Stabilization: 100,
	}

	// ac := scalp.Compute(helper.Buffered(btc.Klines, 50))

	// for a := range ac {
	// 	log.Printf("Action: %v", a.Annotation())
	// }

	klines := ingestors.GetHistory("BTCUSDT", time.Now().Add(-6*time.Hour))

	repo := asset.NewInMemoryRepository()

	err := repo.Append("btc", klines)
	if err != nil {
		log.Fatalf("Error appending BTC data: %v", err)
	}

	r := reports.NewConsoleReport()

	bt := backtest.NewBacktest(repo, r)

	bt.Strategies = []strategy.Strategy{
		// strategy.NewBuyAndHoldStrategy(),
		// momentum.NewAwesomeOscillatorStrategy(),
		scalp,
	}

	bt.Run()
	if err != nil {
		log.Fatalf("Error running backtest: %v", err)
	}
}
