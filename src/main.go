package main

import (
	"log"

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

	bb := volatility.NewBollingerBands[float64]()

	// Compute the Bollinger Bands
	upperBand, middleBand, lowerBand := bb.Compute(btc.ClosePrices)

	for {
		ub := <-upperBand
		log.Printf("UpperBand: %.2f", ub)
		mb := <-middleBand
		log.Printf("MiddleBand: %.2f", mb)
		lb := <-lowerBand
		log.Printf("lowerBand: %.2f", lb)
	}

	// symbols, err := getSymbols()
	// if err != nil {
	// 	fmt.Printf("Error getting symbols: %v\n", err)
	// 	return
	// }
	// fmt.Println("Symbols:", symbols)

	// highPrices, lowPrices, closePrices, err := getPrices(symbols[0])
	// if err != nil {
	// 	fmt.Printf("Error getting symbol data: %v\n", err)
	// 	return
	// }

	// // EMA with a 5-period (fast) and 20-period (slow)
	// fastPeriod := 5
	// slowPeriod := 20

	// emaFast := indicator.Ema(fastPeriod, closePrices)
	// emaSlow := indicator.Ema(slowPeriod, closePrices)

	// // Print the latest EMA values
	// fmt.Printf("EMA(%d): %.2f, EMA(%d): %.2f\n",
	// 	fastPeriod, emaFast[len(emaFast)-1], slowPeriod, emaSlow[len(emaSlow)-1])

	// // Example strategy: EMA crossover
	// if emaFast[len(emaFast)-1] > emaSlow[len(emaSlow)-1] {
	// 	fmt.Println("Buy signal: Fast EMA crossed above Slow EMA")
	// } else if emaFast[len(emaFast)-1] < emaSlow[len(emaSlow)-1] {
	// 	fmt.Println("Sell signal: Fast EMA crossed below Slow EMA")
	// }

	// // RSI with a 14-period
	// rsiPeriod := 14

	// _, rsi := indicator.RsiPeriod(rsiPeriod, closePrices)

	// // Print the latest RSI value
	// fmt.Printf("RSI(%d): %.2f\n", rsiPeriod, rsi[len(rsi)-1])

	// // Example strategy: RSI thresholds
	// if rsi[len(rsi)-1] < 30 {
	// 	fmt.Println("Buy signal: RSI is oversold")
	// } else if rsi[len(rsi)-1] > 70 {
	// 	fmt.Println("Sell signal: RSI is overbought")
	// } else {
	// 	fmt.Println("Hold: RSI is within normal range")
	// }

	// // Bollinger Bands with a 20-period SMA and multiplier of 2
	// upperBand, middleBand, lowerBand := indicators.GetBollingerBands(closePrices)

	// // Print the latest Bollinger Band values
	// fmt.Printf("Upper Band: %.2f, Middle Band: %.2f, Lower Band: %.2f\n",
	// 	upperBand[len(upperBand)-1], middleBand[len(middleBand)-1], lowerBand[len(lowerBand)-1])

	// // Example strategy: Check the latest price against Bollinger Bands
	// latestPrice := closePrices[len(closePrices)-1]
	// if latestPrice < lowerBand[len(lowerBand)-1] {
	// 	fmt.Println("Buy signal: Price is below the lower Bollinger Band")
	// } else if latestPrice > upperBand[len(upperBand)-1] {
	// 	fmt.Println("Sell signal: Price is above the upper Bollinger Band")
	// } else {
	// 	fmt.Println("Hold: Price is within Bollinger Bands")
	// }

	// // MACD settings: short EMA (12), long EMA (26), signal EMA (9)
	// macdLine, signalLine := indicator.Macd(closePrices)

	// // Print the latest MACD values
	// fmt.Printf("MACD Line: %.2f, Signal Line: %.2f\n",
	// 	macdLine[len(macdLine)-1], signalLine[len(signalLine)-1])

	// // Example strategy: MACD crossover
	// if macdLine[len(macdLine)-1] > signalLine[len(signalLine)-1] {
	// 	fmt.Println("Buy signal: MACD Line crossed above Signal Line")
	// } else if macdLine[len(macdLine)-1] < signalLine[len(signalLine)-1] {
	// 	fmt.Println("Sell signal: MACD Line crossed below Signal Line")
	// } else {
	// 	fmt.Println("Hold: No clear MACD signal")
	// }

	// // Calculate Parabolic SAR
	// psar, _ := indicator.ParabolicSar(highPrices, lowPrices, closePrices)

	// // Print the latest PSAR value
	// fmt.Printf("Parabolic SAR: %.2f\n", psar[len(psar)-1])

	// // Example strategy: PSAR trend-following
	// latestPrice = closePrices[len(closePrices)-1]
	// if latestPrice > psar[len(psar)-1] {
	// 	fmt.Println("Buy signal: PSAR is below the price")
	// } else if latestPrice < psar[len(psar)-1] {
	// 	fmt.Println("Sell signal: PSAR is above the price")
	// } else {
	// 	fmt.Println("Hold: No clear PSAR signal")
	// }
}
