package quant

import (
	"math"
	"time"
)

// GhostDCAConfig holds parameters for the ghost DCA benchmark simulator.
type GhostDCAConfig struct {
	InitialCapital float64 // Starting CNY capital, deployed at bar[0]
	MonthlyInject  float64 // CNY injected on each calendar month boundary
}

// GhostDCAResult holds the output of a ghost DCA simulation.
type GhostDCAResult struct {
	FinalEquity   float64 // NAV at last bar (shares × final close)
	TotalInjected float64 // InitialCapital + all monthly injections
	MaxDrawdown   float64 // Peak-to-trough drawdown as fraction [0, 1]
	ROI           float64 // Modified Dietz return on investment
}

// SimulateGhostDCA runs a ghost dollar-cost-averaging benchmark against
// historical bars.
//
// Mechanics:
//   - Buy cfg.InitialCapital worth of shares at bars[0].Close.
//   - On every calendar-month boundary, inject cfg.MonthlyInject CNY and buy
//     shares at that bar's close.
//   - Track the NAV curve (shares × current close) for drawdown calculation.
//   - Compute Modified Dietz ROI:
//     (FinalEquity - InitialCapital - SumInjected) / (InitialCapital + Σ(CFᵢ × wᵢ))
//     where wᵢ = (totalBars - injectionBarIdx) / totalBars.
func SimulateGhostDCA(bars []Bar, cfg GhostDCAConfig) GhostDCAResult {
	if len(bars) == 0 || cfg.InitialCapital <= 0 {
		return GhostDCAResult{}
	}

	// Resolve timestamp units: milliseconds if > 1e10, else seconds.
	tsToTime := func(ts int64) time.Time {
		if ts > 1e10 {
			return time.UnixMilli(ts)
		}
		return time.Unix(ts, 0)
	}

	firstPrice := bars[0].Close
	if firstPrice <= 0 {
		return GhostDCAResult{}
	}

	shares := cfg.InitialCapital / firstPrice
	totalInjected := cfg.InitialCapital
	sumMonthly := 0.0 // sum of all monthly injections (excludes InitialCapital)

	nav := make([]float64, len(bars))

	// cashFlows records every monthly injection amount and its bar index.
	type cashFlow struct {
		amount float64
		barIdx int
	}
	cashFlows := make([]cashFlow, 0)

	lastMonth := tsToTime(bars[0].OpenTime).Month()
	lastYear := tsToTime(bars[0].OpenTime).Year()

	for i, bar := range bars {
		t := tsToTime(bar.OpenTime)

		// Detect calendar-month boundary (skip the first bar).
		if i > 0 && (t.Year() != lastYear || t.Month() != lastMonth) {
			if cfg.MonthlyInject > 0 && bar.Close > 0 {
				added := cfg.MonthlyInject / bar.Close
				shares += added
				totalInjected += cfg.MonthlyInject
				sumMonthly += cfg.MonthlyInject
				cashFlows = append(cashFlows, cashFlow{
					amount: cfg.MonthlyInject,
					barIdx: i,
				})
			}
			lastMonth = t.Month()
			lastYear = t.Year()
		}

		nav[i] = shares * bar.Close
	}

	// Modified Dietz ROI
	totalBars := len(bars)
	if totalBars < 2 {
		totalBars = 2 // avoid division by zero
	}
	denomBars := totalBars - 1

	var weightedCFSum float64
	for _, cf := range cashFlows {
		weight := float64(denomBars-cf.barIdx) / float64(denomBars)
		weightedCFSum += cf.amount * weight
	}

	finalEquity := nav[len(nav)-1]

	var roi float64
	denom := cfg.InitialCapital + weightedCFSum
	if denom != 0 {
		roi = (finalEquity - cfg.InitialCapital - sumMonthly) / denom
	}

	return GhostDCAResult{
		FinalEquity:   math.Round(finalEquity*100) / 100,
		TotalInjected: math.Round(totalInjected*100) / 100,
		MaxDrawdown:   math.Round(MaxDrawdownFromNAV(nav)*10000) / 10000,
		ROI:           math.Round(roi*10000) / 10000,
	}
}

// MaxDrawdownFromNAV computes the peak-to-trough maximum drawdown from a NAV
// curve. Returns a fraction in [0, 1]; 0 means no drawdown occurred.
func MaxDrawdownFromNAV(nav []float64) float64 {
	if len(nav) < 2 {
		return 0
	}

	peak := nav[0]
	var maxDD float64

	for _, v := range nav {
		if v > peak {
			peak = v
		}
		dd := (peak - v) / peak
		if dd > maxDD {
			maxDD = dd
		}
	}

	return maxDD
}
