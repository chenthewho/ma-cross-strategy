package quant

import (
	"math"
	"sort"
)

// ComputeMarketState evaluates the current market regime from a series of
// closing prices. It uses dual EMA crossovers combined with volatility
// regime detection to classify the market into one of four states: quiet,
// bull, bear, or panic.
//
// Returns a MarketState with the appropriate TimeDilationMultiplier,
// BetaMultiplier, and IsQuiet flag for downstream strategy use.
func ComputeMarketState(closes []float64) MarketState {
	// Not enough data: return quiet with default ratios.
	if len(closes) < 112 {
		return MarketState{
			State:                 "quiet",
			TimeDilationMultiplier: 1.0,
			BetaMultiplier:        1.0,
			IsQuiet:               true,
		}
	}

	// 1. Compute EMAs on the full series.
	emaShort := EMA(closes, 21)
	emaLong := EMA(closes, 55)

	// 2. Compute log-returns.
	logReturns := make([]float64, len(closes)-1)
	for i := 1; i < len(closes); i++ {
		if closes[i] > 0 && closes[i-1] > 0 {
			logReturns[i-1] = math.Log(closes[i] / closes[i-1])
		}
	}

	// 3. Current volatility: StdDev of log-returns over the last 21 bars.
	volatility := StdDev(logReturns, 21)

	// 4. Rolling median of 21-bar volatility over the last 112 bars.
	// For each window of length 21, we compute a vol value, then take the
	// median across all those vol values.
	var volWindows []float64
	// We need at least 112 log-return bars to compute 112-21+1=92 windows.
	// Each vol window needs 21 log-return bars.
	if len(logReturns) >= 112 {
		for start := len(logReturns) - 112; start <= len(logReturns)-21; start++ {
			window := logReturns[start : start+21]
			v := StdDev(window, 21)
			volWindows = append(volWindows, v)
		}
	} else if len(logReturns) >= 21 {
		// Fallback: use whatever windows we can.
		for start := 0; start <= len(logReturns)-21; start++ {
			window := logReturns[start : start+21]
			v := StdDev(window, 21)
			volWindows = append(volWindows, v)
		}
	}

	var volMedian float64
	if len(volWindows) > 0 {
		sorted := make([]float64, len(volWindows))
		copy(sorted, volWindows)
		sort.Float64s(sorted)
		volMedian = sorted[len(sorted)/2]
	} else {
		volMedian = volatility
	}

	// 5. Volatility ratio (clamped).
	var volRatio float64
	if volMedian > 0 {
		volRatio = ClipFloat64(volatility/volMedian, 0.5, 4.0)
	} else {
		volRatio = 1.0
	}

	// 6. Classify market state.
	if volRatio < 0.7 {
		return MarketState{
			State:                 "quiet",
			TimeDilationMultiplier: 1.0,
			BetaMultiplier:        1.0,
			IsQuiet:               true,
		}
	}

	if emaShort > emaLong {
		// Bull regime.
		if volRatio > 2.5 {
			return MarketState{
				State:                 "panic",
				TimeDilationMultiplier: 0.0,
				BetaMultiplier:        2.0,
				IsQuiet:               false,
			}
		}
		return MarketState{
			State:                 "bull",
			TimeDilationMultiplier: 1.5,
			BetaMultiplier:        1.0,
			IsQuiet:               false,
		}
	}

	// Bear regime (emaShort <= emaLong).
	if volRatio > 2.0 {
		return MarketState{
			State:                 "panic",
			TimeDilationMultiplier: 0.0,
			BetaMultiplier:        2.0,
			IsQuiet:               false,
		}
	}
	return MarketState{
		State:                 "bear",
		TimeDilationMultiplier: 0.5,
		BetaMultiplier:        1.5,
		IsQuiet:               false,
	}
}
