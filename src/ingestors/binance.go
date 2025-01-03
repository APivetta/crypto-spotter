package ingestors

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/cinar/indicator/v2/asset"
)

type BinanceData struct {
	Symbol      string
	Klines      chan *asset.Snapshot
	LastPrice   chan float64
	LastFetched time.Time
}

const KLINE_LIMIT = "300"
const KLINE_INTERVAL = "1m"

func BinancePoller() ([]BinanceData, error) {
	symbols, err := getSymbols()
	if err != nil {
		return nil, err
	}

	var data []BinanceData

	for _, symbol := range symbols {
		data = append(data, BinanceData{
			Symbol:    symbol,
			Klines:    make(chan *asset.Snapshot),
			LastPrice: make(chan float64),
		})
	}

	pollKlines(data)
	pollLastPrice(data)

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

func pollKlines(data []BinanceData) {
	for _, d := range data {
		go func(d BinanceData) {
			for {
				klines, err := getKlines(d.Symbol, d.LastFetched)
				if err != nil {
					return
				}

				for _, kline := range klines {
					d.Klines <- &kline
				}

				d.LastFetched = time.Now()
				sleep := 1 * time.Minute

				if len(klines) > 0 {
					sleep = time.Until(klines[len(klines)-1].Date.Truncate(time.Minute).Add(time.Minute).Add(5 * time.Second))
				}

				time.Sleep(sleep)
			}
		}(d)
	}
}
func pollLastPrice(data []BinanceData) {
	for _, d := range data {
		go func(d BinanceData) {
			for {
				p, err := getLastPrice(d.Symbol)
				if err != nil {
					return
				}

				d.LastPrice <- p
				time.Sleep(5 * time.Second)
			}
		}(d)
	}
}

func getKlines(symbol string, from time.Time) ([]asset.Snapshot, error) {
	var result []asset.Snapshot
	baseUrl := "https://fapi.binance.com/fapi/v1/klines"
	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("symbol", symbol)
	q.Set("interval", KLINE_INTERVAL)
	q.Set("limit", KLINE_LIMIT)
	if !from.IsZero() {
		q.Set("startTime", strconv.FormatInt(from.UnixMilli(), 10))
	}
	u.RawQuery = q.Encode()

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
		// closeTime := time.Unix(0, int64(data[6].(float64))*int64(time.Millisecond))

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

		result = append(result, asset.Snapshot{
			Date:   openTime,
			High:   high64,
			Low:    low64,
			Open:   open64,
			Close:  close64,
			Volume: volume64,
			// CloseTime: closeTime,
		})
	}

	return result, nil
}

func getLastPrice(symbol string) (float64, error) {
	baseUrl := "https://fapi.binance.com/fapi/v2/ticker/price"
	u, err := url.Parse(baseUrl)
	if err != nil {
		return 0, err
	}

	q := u.Query()
	q.Set("symbol", symbol)
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var raw map[string]interface{}
	err = json.Unmarshal(body, &raw)
	if err != nil {
		return 0, err
	}

	p, err := strconv.ParseFloat(raw["price"].(string), 64)
	if err != nil {
		return 0, err
	}

	return p, nil
}
