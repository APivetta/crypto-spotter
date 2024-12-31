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

type Kline struct {
	OpenTime  time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	CloseTime time.Time
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
				klines, err := getPrices(d.Symbol, d.LastFetched)
				if err != nil {
					return
				}

				for _, kline := range klines {
					d.ClosePrices <- kline.Close
				}

				d.LastFetched = time.Now()
				sleep := 1 * time.Minute

				if len(klines) > 0 {
					sleep = time.Until(klines[len(klines)-1].CloseTime.Truncate(time.Minute).Add(time.Minute).Add(5 * time.Second))
				}

				fmt.Printf("Sleeping for %s\n", sleep)

				time.Sleep(sleep)
			}
		}(d)
	}
}

func getPrices(symbol string, from time.Time) ([]Kline, error) {
	var result []Kline
	baseUrl := "https://fapi.binance.com/fapi/v1/klines"
	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var raw [][]interface{}
	err = json.Unmarshal(body, &raw)
	if err != nil {
		return nil, err
	}

	for _, data := range raw {
		high := data[2].(string)
		low := data[3].(string)
		open := data[1].(string)
		close := data[4].(string)
		volume := data[5].(string)

		openTime := time.Unix(0, int64(data[0].(float64))*int64(time.Millisecond))
		closeTime := time.Unix(0, int64(data[6].(float64))*int64(time.Millisecond))

		high64, err := strconv.ParseFloat(high, 64)
		if err != nil {
			return nil, err
		}

		low64, err := strconv.ParseFloat(low, 64)
		if err != nil {
			return nil, err
		}

		open64, err := strconv.ParseFloat(open, 64)
		if err != nil {
			return nil, err
		}

		close64, err := strconv.ParseFloat(close, 64)
		if err != nil {
			return nil, err
		}

		volume64, err := strconv.ParseFloat(volume, 64)
		if err != nil {
			return nil, err
		}

		result = append(result, Kline{
			OpenTime:  openTime,
			High:      high64,
			Low:       low64,
			Open:      open64,
			Close:     close64,
			Volume:    volume64,
			CloseTime: closeTime,
		})
	}

	return result, nil
}
