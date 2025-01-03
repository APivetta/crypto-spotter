package main

import (
	"log"

	"github.com/cinar/indicator/v2/helper"
	"pivetta.se/crypro-spotter/src/ingestors"
	"pivetta.se/crypro-spotter/src/strategies"
)

func main() {
	bd, err := ingestors.BinancePoller()

	if err != nil {
		log.Fatalf("Error polling Binance: %v", err)
	}

	btc := bd[0]

	scalp := strategies.Scalping{
		LastPrice: btc.LastPrice,
	}
	scalp.Compute(helper.Buffered(btc.Klines, 50))
}
