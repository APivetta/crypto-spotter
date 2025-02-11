package ingestors

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/cinar/indicator/v2/asset"
)

const KLINE_LIMIT = "300"
const KLINE_INTERVAL = "1m"
const LIVE = "https://fapi.binance.com"
const TESTNET = "https://testnet.binancefuture.com"

type BinanceIngestor struct {
	Ingestor
	Url    string
	Key    string
	Secret string
}

func (i *BinanceIngestor) Poll() ([]PollData, error) {
	symbols, err := i.GetSymbols(1)
	if err != nil {
		return nil, err
	}

	var data []PollData

	for _, symbol := range symbols {
		data = append(data, PollData{
			Symbol:    symbol,
			Klines:    make(chan *asset.Snapshot),
			LastPrice: make(chan float64),
		})
	}

	i.pollKlines(data)
	i.pollLastPrice(data)

	return data, nil
}

func (i *BinanceIngestor) GetHistory(symbol string, from time.Time) chan *asset.Snapshot {
	res := make(chan *asset.Snapshot)
	f := from

	go func() {
		for {
			klines, err := i.getKlines(symbol, f)
			if err != nil {
				log.Printf("Error getting klines: %v", err)
				close(res)
				break
			}

			for _, kline := range klines {
				res <- &kline
			}

			if len(klines) == 0 || klines[len(klines)-1].Date.Truncate(time.Minute).Equal(time.Now().Truncate(time.Minute)) {
				close(res)
				break
			}

			f = klines[len(klines)-1].Date
		}
	}()
	return res
}

func (i *BinanceIngestor) GetSymbols(count int) ([]string, error) {
	result := []string{}
	resp, err := http.Get(i.Url + "/fapi/v1/exchangeInfo")
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

func (i *BinanceIngestor) pollKlines(data []PollData) {
	for _, d := range data {
		go func(d PollData) {
			for {
				klines, err := i.getKlines(d.Symbol, d.LastFetched)
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
func (i *BinanceIngestor) pollLastPrice(data []PollData) {
	for _, d := range data {
		go func(d PollData) {
			for {
				p, err := i.getLastPrice(d.Symbol)
				if err != nil {
					return
				}

				d.LastPrice <- p
				time.Sleep(5 * time.Second)
			}
		}(d)
	}
}

func (i *BinanceIngestor) getKlines(symbol string, from time.Time) ([]asset.Snapshot, error) {
	var result []asset.Snapshot
	baseUrl := i.Url + "/fapi/v1/klines"
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

func (i *BinanceIngestor) getLastPrice(symbol string) (float64, error) {
	baseUrl := i.Url + "/fapi/v2/ticker/price"
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

func (BinanceIngestor) generateHMAC(message, secretKey string) string {
	// Create a new HMAC using SHA256
	h := hmac.New(sha256.New, []byte(secretKey))

	// Write message to the HMAC hash
	h.Write([]byte(message))

	// Compute final hash and return it as a hexadecimal string
	return hex.EncodeToString(h.Sum(nil))
}

func (i *BinanceIngestor) GetBalance() (float64, error) {
	baseUrl := i.Url + "/fapi/v3/balance"
	u, err := url.Parse(baseUrl)
	if err != nil {
		return 0, err
	}

	q := u.Query()
	q.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	u.RawQuery = q.Encode()
	signature := i.generateHMAC(u.RawQuery, i.Secret)
	u.RawQuery += "&signature=" + signature

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return 0, err
	}

	req.Header.Set("X-MBX-APIKEY", i.Key)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var raw []interface{}
	err = json.Unmarshal(body, &raw)
	if err != nil {
		return 0, err
	}

	for _, data := range raw {
		d := data.(map[string]interface{})
		if d["asset"].(string) == "USDT" {
			balance, err := strconv.ParseFloat(d["balance"].(string), 64)
			if err != nil {
				return 0, err
			}
			return balance, nil
		}
	}

	return 0, nil
}
