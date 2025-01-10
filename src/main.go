package main

import (
	"log"
	"time"

	"github.com/cinar/indicator/v2/asset"
	"pivetta.se/crypro-spotter/src/ingestors"
	"pivetta.se/crypro-spotter/src/strategies"
)

func main() {
	// bd, err := ingestors.BinancePoller()
	// if err != nil {
	// 	log.Fatalf("Error polling Binance: %v", err)
	// }

	// btc := bd[0]

	// scalp := strategies.Scalping{
	// 	Weights: strategies.StrategyWeights{
	// 		EmaWeight:         0.5,
	// 		RsiWeight:         1.2,
	// 		MacdWeight:        1.3,
	// 		SuperTrendWeight:  1.5,
	// 		BollingerWeight:   0.8,
	// 		StrengthThreshold: 2,
	// 	},
	// 	Stabilization: 100,
	// }

	// ac := scalp.Compute(helper.Buffered(btc.Klines, 50))

	// for a := range ac {
	// 	log.Printf("Action: %v", a.Annotation())
	// }

	klines := ingestors.GetHistory("BTCUSDT", time.Now().Add(-24*time.Hour))

	repo := asset.NewInMemoryRepository()

	err := repo.Append("btc", klines)
	if err != nil {
		log.Fatalf("Error appending BTC data: %v", err)
	}

	// r := reports.NewConsoleReport()

	// bt := backtest.NewBacktest(repo, r)

	// bt.Strategies = []strategy.Strategy{
	// 	// strategy.NewBuyAndHoldStrategy(),
	// 	// momentum.NewAwesomeOscillatorStrategy(),
	// 	scalp,
	// }

	// bt.Run()
	// if err != nil {
	// 	log.Fatalf("Error running backtest: %v", err)
	// }

	// genetics.RunGenetic(repo)

	scalp := strategies.Scalping{
		Weights: strategies.StrategyWeights{
			SuperTrendWeight:  2.4102742056013584,
			BollingerWeight:   2.241427642411903,
			EmaWeight:         0.488966530016531,
			RsiWeight:         1.4152026049544952,
			MacdWeight:        0.9197451562266742,
			StrengthThreshold: 0.6512799251718062},
		Stabilization: 100,
	}

	btc, err := repo.Get("btc")
	if err != nil {
		log.Fatalf("Error getting BTC data: %v", err)
	}

	ac, outcome := scalp.ComputeWithOutcome(btc, true)

	for range outcome {
		<-ac
	}
}
