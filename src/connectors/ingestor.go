package connectors

import (
	"time"

	"github.com/cinar/indicator/v2/asset"
)

type PollData struct {
	Symbol      string
	Klines      chan *asset.Snapshot
	LastPrice   chan float64
	LastFetched time.Time
}

type Connector interface {
	Poll() ([]PollData, error)
	GetHistory(symbol string, from time.Time) chan *asset.Snapshot
	GetSymbols(count int) ([]string, error)
}
