// Package quant provides mathematical tools and data types shared by all strategies.
// All functions are pure: no I/O, no state, no side effects.
// Price-related calculations use dimensionless expressions (log returns / ratios).
package quant

import "math"

// EMA computes the most recent Exponential Moving Average from a sequence.
// Uses α = 2/(period+1). Returns NaN if len(closes) < period.
func EMA(closes []float64, period int) float64 {
	if period <= 0 || len(closes) < period {
		return math.NaN()
	}
	if period == 1 {
		return closes[len(closes)-1]
	}
	alpha := 2.0 / float64(period+1)
	ema := closes[0]
	for _, c := range closes[1:] {
		ema = alpha*c + (1-alpha)*ema
	}
	return ema
}

// StdDev computes the sample standard deviation over the most recent 'period' values.
// Returns NaN if len(closes) < 2 or period < 2.
func StdDev(closes []float64, period int) float64 {
	if period < 2 || len(closes) < period {
		return math.NaN()
	}
	window := closes[len(closes)-period:]
	mean := 0.0
	for _, v := range window {
		mean += v
	}
	mean /= float64(period)

	sumSq := 0.0
	for _, v := range window {
		d := v - mean
		sumSq += d * d
	}
	return math.Sqrt(sumSq / float64(period-1))
}

// LogReturns computes log returns from a close price series.
// logReturn[i] = ln(closes[i] / closes[i-1])
func LogReturns(closes []float64) []float64 {
	if len(closes) < 2 {
		return nil
	}
	r := make([]float64, len(closes)-1)
	for i := 1; i < len(closes); i++ {
		r[i-1] = math.Log(closes[i] / closes[i-1])
	}
	return r
}

// MAVAbsChange computes Mean Absolute Value of price changes over period L.
// Formula: sum(|close[i] - close[i-1]|) / (L-1) for the last L bars.
// This is NOT ATR — it doesn't depend on High/Low.
func MAVAbsChange(closes []float64, period int) float64 {
	if period < 2 || len(closes) < period {
		return math.NaN()
	}
	window := closes[len(closes)-period:]
	sum := 0.0
	for i := 1; i < len(window); i++ {
		sum += math.Abs(window[i] - window[i-1])
	}
	return sum / float64(period-1)
}

// ClipFloat64 clamps x to the interval [lo, hi].
func ClipFloat64(x, lo, hi float64) float64 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

// RoundToCNY rounds a float64 to 2 decimal places (CNY cents).
func RoundToCNY(x float64) float64 {
	return math.Round(x*100) / 100
}

// MedianFloat64 calculates the median of a slice. Modifies the input by sorting (in-place).
func MedianFloat64(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := make([]float64, len(values))
	copy(sorted, values)
	// simple insertion sort for small slices
	for i := 1; i < len(sorted); i++ {
		key := sorted[i]
		j := i - 1
		for j >= 0 && sorted[j] > key {
			sorted[j+1] = sorted[j]
			j--
		}
		sorted[j+1] = key
	}
	mid := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[mid]
	}
	return (sorted[mid-1] + sorted[mid]) / 2
}
