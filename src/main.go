package main

import (
	"log"
	"time"

	"github.com/cinar/indicator/v2/helper"
	"pivetta.se/crypro-spotter/src/genetics"
	"pivetta.se/crypro-spotter/src/ingestors"
	"pivetta.se/crypro-spotter/src/repositories"
	"pivetta.se/crypro-spotter/src/strategies"
)

func main() {
	//geneticsRun()
	backtestRun()
	// liveRun()
}

func geneticsRun() {
	historyMinutes := 24 * 60 * 1
	repo, err := repositories.NewDBRepository("BTCUSDT", historyMinutes)
	if err != nil {
		log.Fatalf("Error creating repository: %v", err)
	}

	err = genetics.RunGenetic(repo, "BTCUSDT")
	if err != nil {
		log.Fatalf("Error running genetic algorithm: %v", err)
	}
}

func backtestRun() {
	historyMinutes := 24 * 60 * 7
	repo, err := repositories.NewDBRepository("BTCUSDT", historyMinutes)
	if err != nil {
		log.Fatalf("Error creating repository: %v", err)
	}

	scalp := strategies.Scalping{
		Weights: strategies.StrategyWeights{
			SuperTrendWeight:  0.26982423289769386,
			BollingerWeight:   0.10180437138661834,
			EmaWeight:         0.5024262049861916,
			RsiWeight:         2.9068689490984467,
			MacdWeight:        0.6119196362815509,
			StrengthThreshold: 0.9852577989940008,
			AtrMultiplier:     3.3836230562644802,
		},
		Stabilization: 100,
		WithSL:        true,
	}

	btc, err := repo.Get("BTCUSDT")
	if err != nil {
		log.Fatalf("Error getting BTC data: %v", err)
	}

	ac, outcome := scalp.ComputeWithOutcome(btc, true)

	for range outcome {
		<-ac
	}

	time.Sleep(1 * time.Second)
}

func liveRun() {
	bd, err := ingestors.BinancePoller()
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
