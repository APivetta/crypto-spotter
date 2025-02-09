package main

import (
	"flag"
	"log"
	"time"

	"pivetta.se/crypro-spotter/src/repositories"
	"pivetta.se/crypro-spotter/src/strategies"
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

	// TODO: read weights from file
	scalp := strategies.Scalping{
		Weights: strategies.StrategyWeights{
			// SuperTrendWeight:  0.26982423289769386,
			// BollingerWeight:   0.10180437138661834,
			// EmaWeight:         0.5024262049861916,
			// RsiWeight:         2.9068689490984467,
			// MacdWeight:        0.6119196362815509,
			// StrengthThreshold: 0.9852577989940008,
			// AtrMultiplier:     3.3836230562644802,
			SuperTrendWeight:  0.09757534157306982,
			BollingerWeight:   2.297643212734978,
			EmaWeight:         0.20797171496165825,
			RsiWeight:         1.2438732022732464,
			MacdWeight:        0.09309921828877975,
			StrengthThreshold: 2.6881009994704934,
			AtrMultiplier:     2.6788428493513603,
		},
		Stabilization: 100,
		WithSL:        true,
	}

	r, err := repo.Get(asset)
	if err != nil {
		log.Fatalf("Error getting BTC data: %v", err)
	}

	ac, outcome := scalp.ComputeWithOutcome(r, true)

	for range outcome {
		<-ac
	}

	time.Sleep(1 * time.Second)
}
