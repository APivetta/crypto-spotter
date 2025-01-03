package main

import (
	"log"
	"time"

	"github.com/cinar/indicator/v2/helper"
	"github.com/cinar/indicator/v2/momentum"
	"github.com/cinar/indicator/v2/trend"
	"github.com/cinar/indicator/v2/volatility"
	"pivetta.se/crypro-spotter/src/ingestors"
)

// Adjust your rolling window to 200 minutes
// const rollingWindow = 200

func main() {
	bd, err := ingestors.BinancePoller()
	closeSource := make(chan float64, 50)
	lowSource := make(chan float64, 50)
	highSource := make(chan float64, 50)
	close := helper.Duplicate(closeSource, 6)
	low := helper.Duplicate(lowSource, 1)
	high := helper.Duplicate(highSource, 1)

	if err != nil {
		log.Fatalf("Error polling Binance: %v", err)
	}

	btc := bd[0]

	// init indicators
	bollingerBands := volatility.NewBollingerBands[float64]()
	emaFast := trend.NewEmaWithPeriod[float64](5)
	emaSlow := trend.NewEmaWithPeriod[float64](20)
	rsi := momentum.NewRsi[float64]()
	macd := trend.NewMacd[float64]()
	superTrend := volatility.NewSuperTrend[float64]()

	var st, ub, mb, lb, e5, e20, r14, ml, ms float64
	go func() {
		superTrend := superTrend.Compute(high[0], low[0], close[0])
		upperBand, middleBand, lowerBand := bollingerBands.Compute(close[1])
		ema5 := emaFast.Compute(close[2])
		ema20 := emaSlow.Compute(close[3])
		rsi14 := rsi.Compute(close[4])
		macdLine, macdSignal := macd.Compute(close[5])

		go func() {
			for {
				st = <-superTrend
			}
		}()

		go func() {
			for {
				ub = <-upperBand
				mb = <-middleBand
				lb = <-lowerBand
			}
		}()

		go func() {
			for {
				e5 = <-ema5
			}
		}()

		go func() {
			for {
				e20 = <-ema20
			}
		}()

		go func() {
			for {
				r14 = <-rsi14
			}
		}()

		go func() {
			for {
				ml = <-macdLine
				ms = <-macdSignal
			}
		}()
	}()

	stabilization := 0
	for kline := range btc.Klines {
		closeSource <- kline.Close
		lowSource <- kline.Low
		highSource <- kline.High

		if stabilization < 275 {
			stabilization++
		} else {
			time.Sleep(500 * time.Millisecond)
			log.Printf("SuperTrend: %.2f", st)
			log.Printf("UpperBand: %.2f", ub)
			log.Printf("MiddleBand: %.2f", mb)
			log.Printf("LowerBand: %.2f", lb)
			log.Printf("EMA5: %.2f", e5)
			log.Printf("EMA20: %.2f", e20)
			log.Printf("RSI14: %.2f", r14)
			log.Printf("MACDLine: %.2f", ml)
			log.Printf("MACDSignal: %.2f", ms)
		}
	}
}
