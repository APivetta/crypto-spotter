package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/cinar/indicator"
	"pivetta.se/crypro-spotter/src/indicators"
)

// Adjust your rolling window to 200 minutes
// const rollingWindow = 200

func main() {
	symbols, err := getSymbols()
	if err != nil {
		fmt.Printf("Error getting symbols: %v\n", err)
		return
	}
	fmt.Println("Symbols:", symbols)

	highPrices, lowPrices, closePrices, err := getPrices(symbols[0])
	if err != nil {
		fmt.Printf("Error getting symbol data: %v\n", err)
		return
	}

	// EMA with a 5-period (fast) and 20-period (slow)
	fastPeriod := 5
	slowPeriod := 20

	emaFast := indicator.Ema(fastPeriod, closePrices)
	emaSlow := indicator.Ema(slowPeriod, closePrices)

	// Print the latest EMA values
	fmt.Printf("EMA(%d): %.2f, EMA(%d): %.2f\n",
		fastPeriod, emaFast[len(emaFast)-1], slowPeriod, emaSlow[len(emaSlow)-1])

	// Example strategy: EMA crossover
	if emaFast[len(emaFast)-1] > emaSlow[len(emaSlow)-1] {
		fmt.Println("Buy signal: Fast EMA crossed above Slow EMA")
	} else if emaFast[len(emaFast)-1] < emaSlow[len(emaSlow)-1] {
		fmt.Println("Sell signal: Fast EMA crossed below Slow EMA")
	}

	// RSI with a 14-period
	rsiPeriod := 14

	_, rsi := indicator.RsiPeriod(rsiPeriod, closePrices)

	// Print the latest RSI value
	fmt.Printf("RSI(%d): %.2f\n", rsiPeriod, rsi[len(rsi)-1])

	// Example strategy: RSI thresholds
	if rsi[len(rsi)-1] < 30 {
		fmt.Println("Buy signal: RSI is oversold")
	} else if rsi[len(rsi)-1] > 70 {
		fmt.Println("Sell signal: RSI is overbought")
	} else {
		fmt.Println("Hold: RSI is within normal range")
	}

	// Bollinger Bands with a 20-period SMA and multiplier of 2
	upperBand, middleBand, lowerBand := indicators.GetBollingerBands(closePrices)

	// Print the latest Bollinger Band values
	fmt.Printf("Upper Band: %.2f, Middle Band: %.2f, Lower Band: %.2f\n",
		upperBand[len(upperBand)-1], middleBand[len(middleBand)-1], lowerBand[len(lowerBand)-1])

	// Example strategy: Check the latest price against Bollinger Bands
	latestPrice := closePrices[len(closePrices)-1]
	if latestPrice < lowerBand[len(lowerBand)-1] {
		fmt.Println("Buy signal: Price is below the lower Bollinger Band")
	} else if latestPrice > upperBand[len(upperBand)-1] {
		fmt.Println("Sell signal: Price is above the upper Bollinger Band")
	} else {
		fmt.Println("Hold: Price is within Bollinger Bands")
	}

	// MACD settings: short EMA (12), long EMA (26), signal EMA (9)
	macdLine, signalLine := indicator.Macd(closePrices)

	// Print the latest MACD values
	fmt.Printf("MACD Line: %.2f, Signal Line: %.2f\n",
		macdLine[len(macdLine)-1], signalLine[len(signalLine)-1])

	// Example strategy: MACD crossover
	if macdLine[len(macdLine)-1] > signalLine[len(signalLine)-1] {
		fmt.Println("Buy signal: MACD Line crossed above Signal Line")
	} else if macdLine[len(macdLine)-1] < signalLine[len(signalLine)-1] {
		fmt.Println("Sell signal: MACD Line crossed below Signal Line")
	} else {
		fmt.Println("Hold: No clear MACD signal")
	}

	// Calculate Parabolic SAR
	psar, _ := indicator.ParabolicSar(highPrices, lowPrices, closePrices)

	// Print the latest PSAR value
	fmt.Printf("Parabolic SAR: %.2f\n", psar[len(psar)-1])

	// Example strategy: PSAR trend-following
	latestPrice = closePrices[len(closePrices)-1]
	if latestPrice > psar[len(psar)-1] {
		fmt.Println("Buy signal: PSAR is below the price")
	} else if latestPrice < psar[len(psar)-1] {
		fmt.Println("Sell signal: PSAR is above the price")
	} else {
		fmt.Println("Hold: No clear PSAR signal")
	}
}

func getSymbols() ([]string, error) {
	result := []string{}
	count := 5
	resp, err := http.Get("https://fapi.binance.com/fapi/v1/exchangeInfo")
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return nil, err
	}

	var r map[string]interface{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return nil, err
	}

	if symbols, ok := r["symbols"].([]interface{}); ok {
		for i := 0; i < count; i++ {
			symbolsMap := symbols[i].(map[string]interface{})
			s := symbolsMap["symbol"].(string)
			result = append(result, s)
		}
	}

	return result, nil
}

func getPrices(symbol string) ([]float64, []float64, []float64, error) {
	var highPrices, lowPrices, closePrices []float64
	baseUrl := "https://fapi.binance.com/fapi/v1/klines"
	u, err := url.Parse(baseUrl)
	if err != nil {
		fmt.Printf("Error parsing URL: %v\n", err)
		return nil, nil, nil, err
	}

	q := u.Query()
	q.Set("symbol", symbol)
	q.Set("interval", "1m")
	q.Set("limit", "300")
	u.RawQuery = q.Encode()

	fmt.Println("Constructed URL:", u.String()) // Print the full URL

	resp, err := http.Get(u.String())
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return nil, nil, nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return nil, nil, nil, err
	}

	// JSON array
	var raw [][]interface{}
	err = json.Unmarshal(body, &raw)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, nil, nil, err
	}

	for _, data := range raw {
		high := data[2].(string)
		low := data[3].(string)
		close := data[4].(string)

		high64, err := strconv.ParseFloat(high, 64)
		if err != nil {
			fmt.Printf("Error parsing float: %v\n", err)
			return nil, nil, nil, err
		}

		low64, err := strconv.ParseFloat(low, 64)
		if err != nil {
			fmt.Printf("Error parsing float: %v\n", err)
			return nil, nil, nil, err
		}

		close64, err := strconv.ParseFloat(close, 64)
		if err != nil {
			fmt.Printf("Error parsing float: %v\n", err)
			return nil, nil, nil, err
		}

		highPrices = append(highPrices, high64)
		lowPrices = append(lowPrices, low64)
		closePrices = append(closePrices, close64)
	}

	return highPrices, lowPrices, closePrices, nil
}
