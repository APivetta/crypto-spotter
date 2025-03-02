package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cinar/indicator/v2/helper"
	"github.com/cinar/indicator/v2/strategy"
	"pivetta.se/crypro-spotter/src/connectors"
	"pivetta.se/crypro-spotter/src/lib/db"
	"pivetta.se/crypro-spotter/src/lib/helpers"
	"pivetta.se/crypro-spotter/src/strategies"
)

// yes i know, global bad
var outcome float64
var asset *string

func main() {
	// Create a channel to listen for OS signals
	sigChan := make(chan os.Signal, 1)
	// Notify the channel when receiving an Interrupt (Ctrl+C) or Termination signal
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	// Run a goroutine to handle the signal
	go func() {
		sig := <-sigChan
		fmt.Println("\nReceived signal:", sig)
		cleanup()
		os.Exit(0)
	}()

	asset = flag.String("asset", "BTCUSDT", "Asset to backtest")
	flag.Parse()
	apiKey := os.Getenv("API_KEY")
	apiSecret := os.Getenv("API_SECRET")
	mode := os.Getenv("MODE")
	skip := os.Getenv("SKIP")
	trade := os.Getenv("TRADE")

	url := connectors.TESTNET
	if mode == "live" {
		log.Println("Running in live mode")
		url = connectors.LIVE
	}

	if apiKey == "" || apiSecret == "" {
		log.Fatalf("API_KEY and API_SECRET must be set")
	}

	bc := connectors.BinanceConnector{
		Url:    url,
		Key:    apiKey,
		Secret: apiSecret,
	}
	db := db.GetDb()

	for {
		if skip != "true" {
			helpers.FetchSnapshots(db, *asset, bc)
			helpers.GeneticsRun(3, *asset)
		}
		liveRun(bc, *asset, trade == "true")
	}
}

// func main() {
// 	apiKey := os.Getenv("API_KEY")
// 	apiSecret := os.Getenv("API_SECRET")

// 	if apiKey == "" || apiSecret == "" {
// 		log.Fatalf("API_KEY and API_SECRET must be set")
// 	}

// 	bc := connectors.BinanceConnector{
// 		Url:    connectors.TESTNET,
// 		Key:    apiKey,
// 		Secret: apiSecret,
// 	}

// 	b, err := bc.GetBalance()
// 	if err != nil {
// 		log.Fatalf("Error getting balance: %v", err)
// 	}
// 	log.Printf("Balance: %v", b)

// 	bc.TestOrder("BTCUSDT", "BUY", "LONG")
// 	time.Sleep(10 * time.Second)
// 	bc.TestOrder("BTCUSDT", "SELL", "SHORT")

// }

func liveRun(bc connectors.BinanceConnector, asset string, trade bool) {
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	bd, err := bc.Poll(asset)
	if err != nil {
		log.Fatalf("Error polling Binance: %v", err)
	}

	w, err := db.GetLatestWeights("BTCUSDT")
	if err != nil {
		log.Fatalf("Error getting weights: %v", err)
	}

	scalp := strategies.Scalping{
		Weights:       *w,
		Stabilization: 299,
		WithSL:        true,
	}

	ss := helper.Duplicate(bd.Klines, 2)
	ac, oc := scalp.ComputeWithOutcome(ss[0], true)

	var pos *strategies.Position
	for a := range ac {
		kline := <-ss[1]
		outcome = <-oc
		log.Printf("Action: %v, Price: %.2f, Outcome: %.2f", extendedAnnotation(a), kline.Close, outcome)

		// we should run the compute until midnight, store the outcome and retrain the weights
		if time.Since(startDate) >= 24*time.Hour {
			log.Printf("End of day, retraining weights")
			if pos != nil {
				// generate opposite order to close position
				var orderType connectors.Side
				if pos.Type == strategies.LONG {
					orderType = connectors.SELL
				} else {
					orderType = connectors.BUY
				}

				err := bc.PlaceOrder(asset, orderType, helpers.CalculateQuantity(250, pos.EntryPrice))
				if err != nil {
					// figure what to do here, serious here
					log.Printf("Error placing order: %v", err)
				} else {
					pos = nil
				}
			}

			break
		}

		if !trade {
			continue
		}

		if a == strategy.Buy {
			err := bc.PlaceOrder(asset, connectors.BUY, helpers.CalculateQuantity(250, kline.Close))
			if err != nil {
				// figure what to do here
				log.Printf("Error placing order: %v", err)
			} else {
				log.Printf("Placed order: %v", asset)
				pos = &strategies.Position{
					Type:       strategies.LONG,
					EntryPrice: kline.Close,
					EntryTime:  time.Now(),
				}
			}
		} else if a == strategy.Sell {
			err := bc.PlaceOrder(asset, connectors.SELL, helpers.CalculateQuantity(250, kline.Close))
			if err != nil {
				// figure what to do here
				log.Printf("Error placing order: %v", err)
			} else {
				log.Printf("Placed order: %v", asset)
				pos = &strategies.Position{
					Type:       strategies.SHORT,
					EntryPrice: kline.Close,
					EntryTime:  time.Now(),
				}
			}
		} else if a == strategies.Close && pos != nil {
			// generate opposite order to close position
			var orderType connectors.Side
			if pos.Type == strategies.LONG {
				orderType = connectors.SELL
			} else {
				orderType = connectors.BUY
			}

			err := bc.PlaceOrder(asset, orderType, helpers.CalculateQuantity(250, pos.EntryPrice))
			if err != nil {
				// figure what to do here, serious here
				log.Printf("Error placing order: %v", err)
			} else {
				pos = nil
			}
		}
	}

	cleanup()
}

func cleanup() {
	log.Println("Trade results:", outcome)
	err := storeOutcome(*asset, outcome)
	if err != nil {
		log.Fatalf("Error storing outcome: %v", err)
	}
}

func extendedAnnotation(a strategy.Action) string {
	switch a {
	case strategy.Sell:
		return "SHORT"

	case strategy.Buy:
		return "LONG"

	case strategies.Close:
		return "CLOSE"

	default:
		return "HOLD"
	}
}

func storeOutcome(a string, outcome float64) error {
	db := db.GetDb()

	query := `INSERT INTO trade_results (asset, date, result) VALUES ($1, $2, $3)`
	_, err := db.Exec(query, a, time.Now(), outcome)
	if err != nil {
		return fmt.Errorf("storeOutcome: %w", err)
	}

	return nil
}
