package strategies

import (
	"log"
	"sync"
	"time"

	"github.com/cinar/indicator/v2/asset"
	"github.com/cinar/indicator/v2/helper"
	"github.com/cinar/indicator/v2/momentum"
	"github.com/cinar/indicator/v2/strategy"
	"github.com/cinar/indicator/v2/trend"
	"github.com/cinar/indicator/v2/volatility"
)

type Scalping struct {
	strategy.Strategy
	Weights         StrategyWeights
	Stabilization   int
	CurrentPosition *Position
	WithSL          bool
	WithTP          bool
}

// TODO: move to a better place
type PositionType int

const (
	FLAT PositionType = iota
	LONG
	SHORT
)

type Position struct {
	Type       PositionType
	EntryPrice float64
	EntryTime  time.Time
}

const Close strategy.Action = 2

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
	atr        <-chan float64
}

type StrategyWeights struct {
	SuperTrendWeight  float64 `json:"superTrendWeight"`
	BollingerWeight   float64 `json:"bollingerWeight"`
	EmaWeight         float64 `json:"emaWeight"`
	RsiWeight         float64 `json:"rsiWeight"`
	MacdWeight        float64 `json:"macdWeight"`
	StrengthThreshold float64 `json:"strengthThreshold"`
	AtrMultiplier     float64 `json:"atrMultiplier"`
}

type StrategyParams struct {
	Snapshot   asset.Snapshot
	SuperTrend float64
	UpperBand  float64
	MiddleBand float64
	LowerBand  float64
	Ema5       float64
	Ema20      float64
	Rsi14      float64
	MacdLine   float64
	MacdSignal float64
	Atr        float64
}

func (Scalping) Name() string {
	return "Scalping"
}

func (Scalping) getIndicators(snapshots <-chan *asset.Snapshot) Indicators {
	// init channels
	ss := helper.Duplicate(snapshots, 3)
	close := helper.Duplicate(asset.SnapshotsAsClosings(ss[0]), 7)
	low := helper.Duplicate(asset.SnapshotsAsLows(ss[1]), 2)
	high := helper.Duplicate(asset.SnapshotsAsHighs(ss[2]), 2)

	// init indicators
	emaFast := trend.NewEmaWithPeriod[float64](5)
	emaSlow := trend.NewEmaWithPeriod[float64](20)
	bollingerBands := volatility.NewBollingerBands[float64]()
	rsi := momentum.NewRsi[float64]()
	macd := trend.NewMacd[float64]()
	superTrend := volatility.NewSuperTrend[float64]()
	atrInd := volatility.NewAtr[float64]()

	// compute
	st := superTrend.Compute(high[0], low[0], close[0])
	atr := atrInd.Compute(high[1], low[1], close[6])
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
		atr:        atr,
	}
}

func (s *Scalping) decide(params StrategyParams) strategy.Action {
	// Initialize signal strength
	signalStrength := 0.0
	rsiOverbought := 70.0
	rsiOversold := 30.0
	macdThreshold := 0.5
	// maxStrength := s.Weights.SuperTrendWeight + s.Weights.BollingerWeight + s.Weights.EmaWeight + s.Weights.RsiWeight + s.Weights.MacdWeight

	// check if we have a position and TP and SL levels
	var tp, sl float64
	multi := s.Weights.AtrMultiplier * params.Atr
	if s.CurrentPosition != nil && s.WithSL {
		if s.CurrentPosition.Type == LONG {
			sl = s.CurrentPosition.EntryPrice - (multi / 2)

			if params.Snapshot.Low <= sl {
				s.CurrentPosition = nil
				return Close
			}
		} else if s.CurrentPosition.Type == SHORT {
			sl = s.CurrentPosition.EntryPrice + (multi / 2)

			if params.Snapshot.High >= sl {
				s.CurrentPosition = nil
				return Close
			}
		}
	}

	if s.CurrentPosition != nil && s.WithTP {
		if s.CurrentPosition.Type == LONG {
			tp = s.CurrentPosition.EntryPrice + multi

			if params.Snapshot.High >= tp {
				s.CurrentPosition = nil
				return Close
			}
		} else if s.CurrentPosition.Type == SHORT {
			tp = s.CurrentPosition.EntryPrice - multi

			if params.Snapshot.Low <= tp {
				s.CurrentPosition = nil
				return Close
			}
		}
	}

	// EMA Crossover Logic
	if params.Ema5 > params.Ema20 {
		signalStrength += s.Weights.EmaWeight // Bullish signal
	} else if params.Ema5 < params.Ema20 {
		signalStrength -= s.Weights.EmaWeight // Bearish signal
	}

	// RSI Logic
	if params.Rsi14 > rsiOverbought {
		signalStrength -= s.Weights.RsiWeight // Bearish signal (overbought)
	} else if params.Rsi14 < rsiOversold {
		signalStrength += s.Weights.RsiWeight // Bullish signal (oversold)
	}

	// MACD Logic
	macdDifference := params.MacdLine - params.MacdSignal
	if macdDifference > macdThreshold {
		signalStrength += s.Weights.MacdWeight // Bullish signal
	} else if macdDifference < -macdThreshold {
		signalStrength -= s.Weights.MacdWeight // Bearish signal
	}

	// SuperTrend Logic
	if params.Snapshot.Close > params.SuperTrend {
		signalStrength += s.Weights.SuperTrendWeight // Bullish signal
	} else {
		signalStrength -= s.Weights.SuperTrendWeight // Bearish signal
	}

	// Bollinger Band Logic
	if params.Snapshot.Close < params.LowerBand {
		signalStrength += s.Weights.BollingerWeight // Bullish signal (price near lower band)
	} else if params.Snapshot.Close > params.UpperBand {
		signalStrength -= s.Weights.BollingerWeight // Bearish signal (price near upper band)
	} else if params.Snapshot.Close > params.MiddleBand {
		signalStrength += s.Weights.BollingerWeight / 2 // Slightly bullish
	} else if params.Snapshot.Close < params.MiddleBand {
		signalStrength -= s.Weights.BollingerWeight / 2 // Slightly bearish
	}

	// Decision Logic
	if signalStrength > s.Weights.StrengthThreshold {
		// log.Printf("Signal Strength: %.2f, confidence: %.2f", signalStrength, signalStrength/maxStrength)

		if s.CurrentPosition != nil && s.CurrentPosition.Type == SHORT {
			s.CurrentPosition = nil
			return Close
		}

		s.CurrentPosition = &Position{
			Type:       LONG,
			EntryPrice: params.Snapshot.Close,
			EntryTime:  params.Snapshot.Date,
		}

		return strategy.Buy
	} else if signalStrength < -s.Weights.StrengthThreshold {
		// log.Printf("Signal Strength: %.2f, confidence: %.2f", signalStrength, signalStrength/maxStrength)

		if s.CurrentPosition != nil && s.CurrentPosition.Type == LONG {
			s.CurrentPosition = nil
			return Close
		}

		s.CurrentPosition = &Position{
			Type:       SHORT,
			EntryPrice: params.Snapshot.Close,
			EntryTime:  params.Snapshot.Date,
		}
		return strategy.Sell
	} else {
		return strategy.Hold
	}
}

func (s Scalping) Compute(snapshots <-chan *asset.Snapshot) <-chan strategy.Action {
	var st, ub, mb, lb, e5, e20, r14, ml, ms, atr float64
	stable := false
	ac := make(chan strategy.Action, 50)
	ss := make(chan *asset.Snapshot, 50)
	ind := s.getIndicators(ss)

	var wg sync.WaitGroup
	go func() {
		for {
			atr = <-ind.atr
			if stable {
				wg.Done()
			}
		}
	}()

	go func() {
		for {
			st = <-ind.superTrend
			if stable {
				wg.Done()
			}
		}
	}()

	go func() {
		for {
			ub = <-ind.upperBand
			mb = <-ind.middleBand
			lb = <-ind.lowerBand
			if stable {
				wg.Done()
			}
		}
	}()

	go func() {
		for {
			e5 = <-ind.ema5
			if stable {
				wg.Done()
			}
		}
	}()

	go func() {
		for {
			e20 = <-ind.ema20
			if stable {
				wg.Done()
			}
		}
	}()

	go func() {
		for {
			r14 = <-ind.rsi14
			if stable {
				wg.Done()
			}
		}
	}()

	go func() {
		for {
			ml = <-ind.macdLine
			ms = <-ind.macdSignal
			if stable {
				wg.Done()
			}
		}
	}()

	go func() {
		i := 0
		for asset := range snapshots {
			// log.Printf("Asset: %v, line %d", asset, i)
			ss <- asset
			wg.Wait()

			// one time stabilization to let all indicators catch up before syncing
			if !stable && i >= s.Stabilization {
				time.Sleep(500 * time.Millisecond)
				stable = true
			}

			if i >= s.Stabilization {
				// log.Printf("SuperTrend: %.2f", st)
				// log.Printf("UpperBand: %.2f", ub)
				// log.Printf("MiddleBand: %.2f", mb)
				// log.Printf("LowerBand: %.2f", lb)
				// log.Printf("EMA5: %.2f", e5)
				// log.Printf("EMA20: %.2f", e20)
				// log.Printf("RSI14: %.2f", r14)
				// log.Printf("MACDLine: %.2f", ml)
				// log.Printf("MACDSignal: %.2f", ms)

				action := s.decide(StrategyParams{
					Snapshot:   *asset,
					SuperTrend: st,
					UpperBand:  ub,
					MiddleBand: mb,
					LowerBand:  lb,
					Ema5:       e5,
					Ema20:      e20,
					Rsi14:      r14,
					MacdLine:   ml,
					MacdSignal: ms,
					Atr:        atr,
				})
				ac <- action
				wg.Add(7)
			} else {
				ac <- strategy.Hold
			}

			i++

		}
		close(ac)
	}()

	return ac
}

func (s Scalping) ComputeWithOutcome(c <-chan *asset.Snapshot, withLog bool) (<-chan strategy.Action, <-chan float64) {
	snapshots := helper.Duplicate(c, 3)
	actions := helper.Duplicate(s.Compute(snapshots[0]), 2)

	high := 0.0
	low := 0.0
	position := FLAT
	entryPrice := 0.0
	totalDiff := 0.0
	minutes := 0

	go func() {
		for range snapshots[2] {
		}
		if withLog {
			log.Printf("Total Diff: %.2f", totalDiff)
		}
	}()

	outcomes := helper.Operate(snapshots[1], actions[1], func(ss *asset.Snapshot, action strategy.Action) float64 {
		close := ss.Close
		minutes++

		// record high and low of the current position
		if position != FLAT {
			if ss.High > high {
				high = ss.High
			}

			if ss.Low < low {
				low = ss.Low
			}
		}

		if position == FLAT {
			if action == strategy.Buy {
				position = LONG
				entryPrice = close
				high = close
				low = close
				minutes = 0
			} else if action == strategy.Sell {
				position = SHORT
				entryPrice = close
				high = close
				low = close
				minutes = 0
			}
		} else if position == LONG && action == Close {
			diff := close - entryPrice
			totalDiff += diff
			position = FLAT
			if withLog {
				log.Printf("Long position closed. entry: %.2f, exit: %.2f, diff: %.2f, minutes %d, high: %.2f, low: %.2f, total diff: %.2f", entryPrice, close, diff, minutes, high, low, totalDiff)
			}

		} else if position == SHORT && action == Close {
			diff := entryPrice - close
			totalDiff += diff
			position = FLAT
			if withLog {
				log.Printf("Short position closed. entry: %.2f, exit: %.2f, diff: %.2f, minutes %d, high: %.2f, low: %.2f, total diff: %.2f", entryPrice, close, diff, minutes, high, low, totalDiff)
			}
		}

		return totalDiff
	})

	return actions[0], outcomes
}

func (s Scalping) Report(c <-chan *asset.Snapshot) *helper.Report {
	//
	// snapshots[0] -> dates
	// snapshots[1] -> Compute     -> actions -> annotations
	// snapshots[2] -> closings[0] -> close
	//
	snapshots := helper.Duplicate(c, 3)

	dates := asset.SnapshotsAsDates(snapshots[0])
	closings := asset.SnapshotsAsClosings(snapshots[2])

	actions, outcomes := strategy.ComputeWithOutcome(s, snapshots[1])
	annotations := strategy.ActionsToAnnotations(actions)
	outcomes = helper.MultiplyBy(outcomes, 100)

	report := helper.NewReport(s.Name(), dates)
	report.AddChart()
	report.AddChart()

	report.AddColumn(helper.NewNumericReportColumn("Close", closings))
	report.AddColumn(helper.NewAnnotationReportColumn(annotations), 0, 1)

	report.AddColumn(helper.NewNumericReportColumn("Outcome", outcomes), 2)

	return report
}
