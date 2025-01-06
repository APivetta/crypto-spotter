package reports

import (
	"log"

	"github.com/cinar/indicator/v2/asset"
	"github.com/cinar/indicator/v2/backtest"
	"github.com/cinar/indicator/v2/strategy"
)

type ConsoleReport struct {
	backtest.Report
}

func NewConsoleReport() backtest.Report {
	return &ConsoleReport{}
}

func (r *ConsoleReport) Begin(assetNames []string, strategies []strategy.Strategy) error {
	log.Printf("Backtest started for assets: %v", assetNames)
	return nil
}

func (r *ConsoleReport) AssetBegin(name string, strategies []strategy.Strategy) error {
	log.Printf("Backtesting started for asset: %s", name)
	return nil
}

func (r *ConsoleReport) Write(assetName string, currentStrategy strategy.Strategy, snapshots <-chan *asset.Snapshot, actions <-chan strategy.Action, outcomes <-chan float64) error {
	log.Printf("Writing backtest results for asset: %s", assetName)
	var sc, ac, oc bool
	for {
		snapshot, ok := <-snapshots
		if !ok {
			sc = true
		}
		log.Printf("Snapshot: %v", snapshot)
		action, ok := <-actions
		if !ok {
			ac = true
		}
		log.Printf("Action: %v", action.Annotation())
		outcome, ok := <-outcomes
		if !ok {
			oc = true
		}
		log.Printf("Outcome: %v", outcome)

		if sc && ac && oc {
			break
		}
	}

	return nil
}

func (r *ConsoleReport) AssetEnd(name string) error {
	log.Printf("Backtesting ended for asset: %s", name)
	return nil
}

func (r *ConsoleReport) End() error {
	log.Printf("Backtest ended")
	return nil
}
