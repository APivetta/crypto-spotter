package main

import (
	"flag"
	"fmt"
	"log"

	"pivetta.se/crypro-spotter/src/connectors"
	"pivetta.se/crypro-spotter/src/lib/db"
	"pivetta.se/crypro-spotter/src/lib/helpers"
)

func main() {
	count := flag.Int("count", 1, "Number of symbols to fetch")
	flag.Parse()

	db := db.GetDb()

	bc := connectors.BinanceConnector{
		Url: connectors.LIVE,
	}

	s, err := bc.GetSymbols(*count)
	if err != nil {
		log.Fatalf("Error fetching symbols: %v\n", err)
	}

	for _, symbol := range s {
		fmt.Printf("Fetching Symbol: %s\n", symbol)
		helpers.FetchSnapshots(db, symbol, bc)
	}
}
