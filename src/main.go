package main

import (
	"log"

	"github.com/cinar/indicator/v2/momentum"
	"github.com/cinar/indicator/v2/trend"
	"github.com/cinar/indicator/v2/volatility"
	"pivetta.se/crypro-spotter/src/ingestors"
)

// Adjust your rolling window to 200 minutes
// const rollingWindow = 200

func main() {
	bd, err := ingestors.BinancePoller()
	if err != nil {
		log.Fatalf("Error polling Binance: %v", err)
	}

	btc := bd[0]

	// init indicators
	bb := volatility.NewBollingerBands[float64]()
	emaFast := trend.NewEmaWithPeriod[float64](5)
	emaSlow := trend.NewEmaWithPeriod[float64](20)
	rsi := momentum.NewRsi[float64]()
	macd := trend.NewMacd[float64]()

	// Compute indicators
	upperBand, middleBand, lowerBand := bb.Compute(btc.ClosePrices)
	ema5 := emaFast.Compute(btc.ClosePrices)
	ema20 := emaSlow.Compute(btc.ClosePrices)
	rsi14 := rsi.Compute(btc.ClosePrices)
	macdLine, macdSignal := macd.Compute(btc.ClosePrices)

	for {
		ub := <-upperBand
		log.Printf("UpperBand: %.2f", ub)
		mb := <-middleBand
		log.Printf("MiddleBand: %.2f", mb)
		lb := <-lowerBand
		log.Printf("LowerBand: %.2f", lb)
		e5 := <-ema5
		log.Printf("EMA5: %.2f", e5)
		e20 := <-ema20
		log.Printf("EMA20: %.2f", e20)
		r14 := <-rsi14
		log.Printf("RSI14: %.2f", r14)
		ml := <-macdLine
		log.Printf("MACDLine: %.2f", ml)
		ms := <-macdSignal
		log.Printf("MACDSignal: %.2f", ms)
	}
}
