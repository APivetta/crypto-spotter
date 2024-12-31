package ingestors

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type BinanceData struct {
	Symbol      string
	HighPrices  chan float64
	LowPrices   chan float64
	ClosePrices chan float64
	LastFetched time.Time
}

func BinancePoller() ([]BinanceData, error) {
	symbols, err := getSymbols()
	if err != nil {
		return nil, err
	}

	var data []BinanceData

	for _, symbol := range symbols {
		data = append(data, BinanceData{
			Symbol:      symbol,
			HighPrices:  make(chan float64),
			LowPrices:   make(chan float64),
			ClosePrices: make(chan float64),
		})
	}

	startPolling(data)

	return data, nil
}

func getSymbols() ([]string, error) {
	result := []string{}
	count := 1
	resp, err := http.Get("https://fapi.binance.com/fapi/v1/exchangeInfo")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var r map[string]interface{}
	err = json.Unmarshal(body, &r)
	if err != nil {
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

func startPolling(data []BinanceData) {
	for _, d := range data {
		go func(d BinanceData) {
			for {
				_, _, closePrices, err := getPrices(d.Symbol, d.LastFetched)
				if err != nil {
					return
				}

				fmt.Println("close:", closePrices)

				// for _, high := range highPrices {
				// 	fmt.Println("Sending high price:", high)
				// 	d.HighPrices <- high
				// }

				// for _, low := range lowPrices {
				// 	fmt.Println("Sending low price:", low)
				// 	d.LowPrices <- low
				// }

				for _, close := range closePrices {
					fmt.Println("Sending close price:", close)
					d.ClosePrices <- close
				}

				d.LastFetched = time.Now()

				time.Sleep(1 * time.Minute)
			}
		}(d)
	}
}

func getPrices(symbol string, from time.Time) ([]float64, []float64, []float64, error) {
	var highPrices, lowPrices, closePrices []float64
	baseUrl := "https://fapi.binance.com/fapi/v1/klines"
	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, nil, nil, err
	}

	q := u.Query()
	q.Set("symbol", symbol)
	q.Set("interval", "1m")
	q.Set("limit", "300")
	if !from.IsZero() {
		q.Set("startTime", strconv.FormatInt(from.UnixMilli(), 10))
	}
	u.RawQuery = q.Encode()

	fmt.Println("Constructed URL:", u.String()) // Print the full URL

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, nil, nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, nil, err
	}

	var raw [][]interface{}
	err = json.Unmarshal(body, &raw)
	if err != nil {
		return nil, nil, nil, err
	}

	for _, data := range raw {
		high := data[2].(string)
		low := data[3].(string)
		close := data[4].(string)

		high64, err := strconv.ParseFloat(high, 64)
		if err != nil {
			return nil, nil, nil, err
		}

		low64, err := strconv.ParseFloat(low, 64)
		if err != nil {
			return nil, nil, nil, err
		}

		close64, err := strconv.ParseFloat(close, 64)
		if err != nil {
			return nil, nil, nil, err
		}

		highPrices = append(highPrices, high64)
		lowPrices = append(lowPrices, low64)
		closePrices = append(closePrices, close64)
	}

	return highPrices, lowPrices, closePrices, nil
}
