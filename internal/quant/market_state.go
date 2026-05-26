package quant

import "math"

// Market state perception layer constants.
const (
	MarketEMAShortBars = 21  // short EMA window for trend detection
	MarketEMALongBars  = 55  // long EMA window for trend detection
	MarketVolStdDevBars = 21 // volatility window
	MarketVolMedianBars  = 112 // volatility median history
)

// ComputeMarketState classifies the current market regime.
//
// Logic:
//  1. Compute EMA_short(21), EMA_long(55), volatility = StdDev(logReturns, 21)
//  2. Compute rolling volatility median over 112 bars → VolRatio
//  3. Classify into quiet/bull/bear/panic based on trend + volatility
func ComputeMarketState(closes []float64, timestamps []int64) MarketState {
	// Need enough data for indicators
	minLen := max(MarketEMALongBars, MarketVolMedianBars)
	if len(closes) < minLen {
		return MarketState{
			State:                  MarketQuiet,
			TimeDilationMultiplier: 1.0,
			BetaMultiplier:         1.0,
			IsQuiet:                true,
		}
	}

	// Compute EMAs
	emaShort := EMA(closes, MarketEMAShortBars)
	emaLong := EMA(closes, MarketEMALongBars)

	// Compute volatility and rolling median
	logRets := LogReturns(closes)
	volatility := StdDev(logRets, MarketVolStdDevBars)

	// Build rolling volatility series for median
	volHistory := make([]float64, 0, len(logRets)-MarketVolStdDevBars+1)
	for i := MarketVolStdDevBars; i <= len(logRets); i++ {
		v := StdDev(logRets[:i], MarketVolStdDevBars)
		if !math.IsNaN(v) {
			volHistory = append(volHistory, v)
		}
	}
	volMedian := MedianFloat64(volHistory)

	var volRatio float64
	if volMedian > 0 {
		volRatio = ClipFloat64(volatility/volMedian, 0.5, 4.0)
	} else {
		volRatio = 1.0
	}

	// Classify
	ms := MarketState{}

	switch {
	case volRatio < 0.7:
		ms.State = MarketQuiet
		ms.TimeDilationMultiplier = 1.0
		ms.BetaMultiplier = 1.0
		ms.IsQuiet = true

	case emaShort > emaLong:
		if volRatio > 2.5 {
			ms.State = MarketPanic
			ms.TimeDilationMultiplier = 0.0
			ms.BetaMultiplier = 2.0
			ms.IsQuiet = false
		} else {
			ms.State = MarketBull
			ms.TimeDilationMultiplier = 1.5
			ms.BetaMultiplier = 1.0
			ms.IsQuiet = false
		}

	default: // emaShort <= emaLong
		if volRatio > 2.0 {
			ms.State = MarketPanic
			ms.TimeDilationMultiplier = 0.0
			ms.BetaMultiplier = 2.0
			ms.IsQuiet = false
		} else {
			ms.State = MarketBear
			ms.TimeDilationMultiplier = 0.5
			ms.BetaMultiplier = 1.5
			ms.IsQuiet = false
		}
	}

	return ms
}
