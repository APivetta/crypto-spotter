package main

import (
	"flag"
	"fmt"
	"log"

	"pivetta.se/crypro-spotter/src/connectors"
	"pivetta.se/crypro-spotter/src/lib/helpers"
)

func main() {
	days := flag.Int("days", 3, "Days to train")
	count := flag.Int("count", 1, "Number of symbols to train")
	flag.Parse()

	bc := connectors.BinanceConnector{
		Url: connectors.LIVE,
	}

	s, err := bc.GetSymbols(*count)
	if err != nil {
		log.Fatalf("Error fetching symbols: %v\n", err)
	}

	for _, symbol := range s {
		fmt.Printf("Training Symbol: %s\n", symbol)
		helpers.GeneticsRun(*days, symbol)
	}

}
