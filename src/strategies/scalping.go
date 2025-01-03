package strategies

import (
	"log"
	"time"

	"github.com/cinar/indicator/v2/asset"
	"github.com/cinar/indicator/v2/helper"
	"github.com/cinar/indicator/v2/momentum"
	"github.com/cinar/indicator/v2/strategy"
	"github.com/cinar/indicator/v2/trend"
	"github.com/cinar/indicator/v2/volatility"
)

type Scalping struct {
	LastPrice <-chan float64
}

type Indicators struct {
	superTrend <-chan float64
	upperBand  <-chan float64
	middleBand <-chan float64
	lowerBand  <-chan float64
	ema5       <-chan float64
	ema20      <-chan float64
	rsi14      <-chan float64
	macdLine   <-chan float64
	macdSignal <-chan float64
}

func (*Scalping) Name() string {
	return "Scalping"
}

func (*Scalping) getIndicators(snapshots <-chan *asset.Snapshot) Indicators {
	// init channels
	ss := helper.Duplicate(snapshots, 3)
	close := helper.Duplicate(asset.SnapshotsAsClosings(ss[0]), 6)
	low := helper.Duplicate(asset.SnapshotsAsLows(ss[1]), 1)
	high := helper.Duplicate(asset.SnapshotsAsHighs(ss[2]), 1)

	// init indicators
	bollingerBands := volatility.NewBollingerBands[float64]()
	emaFast := trend.NewEmaWithPeriod[float64](5)
	emaSlow := trend.NewEmaWithPeriod[float64](20)
	rsi := momentum.NewRsi[float64]()
	macd := trend.NewMacd[float64]()
	superTrend := volatility.NewSuperTrend[float64]()

	// compute
	st := superTrend.Compute(high[0], low[0], close[0])
	upperBand, middleBand, lowerBand := bollingerBands.Compute(close[1])
	ema5 := emaFast.Compute(close[2])
	ema20 := emaSlow.Compute(close[3])
	rsi14 := rsi.Compute(close[4])
	macdLine, macdSignal := macd.Compute(close[5])

	return Indicators{
		superTrend: st,
		upperBand:  upperBand,
		middleBand: middleBand,
		lowerBand:  lowerBand,
		ema5:       ema5,
		ema20:      ema20,
		rsi14:      rsi14,
		macdLine:   macdLine,
		macdSignal: macdSignal,
	}
}

func (i *Scalping) Compute(snapshots <-chan *asset.Snapshot) <-chan strategy.Action {
	var lp, st, ub, mb, lb, e5, e20, r14, ml, ms float64

	ss := helper.Duplicate(snapshots, 2)

	ac := make(chan strategy.Action, 50)
	go func() {
		ind := i.getIndicators(ss[0])
		go func() {
			for {
				lp = <-i.LastPrice
				log.Printf("Last Price: %.2f", lp)
			}
		}()

		go func() {
			for {
				st = <-ind.superTrend
			}
		}()

		go func() {
			for {
				ub = <-ind.upperBand
				mb = <-ind.middleBand
				lb = <-ind.lowerBand
			}
		}()

		go func() {
			for {
				e5 = <-ind.ema5
			}
		}()

		go func() {
			for {
				e20 = <-ind.ema20
			}
		}()

		go func() {
			for {
				r14 = <-ind.rsi14
			}
		}()

		go func() {
			for {
				ml = <-ind.macdLine
				ms = <-ind.macdSignal
			}
		}()
	}()

	stabilization := 0
	for range ss[1] {
		if stabilization < 299 {
			stabilization++
		} else {
			time.Sleep(500 * time.Millisecond)
			log.Printf("SuperTrend: %.2f", st)
			log.Printf("UpperBand: %.2f", ub)
			log.Printf("MiddleBand: %.2f", mb)
			log.Printf("LowerBand: %.2f", lb)
			log.Printf("EMA5: %.2f", e5)
			log.Printf("EMA20: %.2f", e20)
			log.Printf("RSI14: %.2f", r14)
			log.Printf("MACDLine: %.2f", ml)
			log.Printf("MACDSignal: %.2f", ms)
		}
	}

	return ac
}

func (i *Scalping) Report(c <-chan *asset.Snapshot) *helper.Report {
	//
	// snapshots[0] -> dates
	// snapshots[1] -> Compute     -> actions -> annotations
	// snapshots[2] -> closings[0] -> close
	//
	snapshots := helper.Duplicate(c, 3)

	dates := asset.SnapshotsAsDates(snapshots[0])
	closings := asset.SnapshotsAsClosings(snapshots[2])

	actions, outcomes := strategy.ComputeWithOutcome(i, snapshots[1])
	annotations := strategy.ActionsToAnnotations(actions)
	outcomes = helper.MultiplyBy(outcomes, 100)

	report := helper.NewReport(i.Name(), dates)
	report.AddChart()
	report.AddChart()

	report.AddColumn(helper.NewNumericReportColumn("Close", closings))
	report.AddColumn(helper.NewAnnotationReportColumn(annotations), 0, 1)

	report.AddColumn(helper.NewNumericReportColumn("Outcome", outcomes), 2)

	return report
}
