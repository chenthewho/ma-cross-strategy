package quant

import (
	"math"
)

// GhostDCAResult holds the output of a passive DCA benchmark simulation.
type GhostDCAResult struct {
	FinalEquity   float64 // final portfolio value in CNY
	TotalInjected float64 // total capital injected
	MaxDrawdown   float64 // maximum drawdown (fraction, e.g. 0.25 = 25%)
	ROI           float64 // Modified Dietz return (fraction)
}

// navPoint is a single point on the NAV curve.
type navPoint struct {
	ts  int64
	nav float64
}

// GhostDCAConfig holds DCA benchmark parameters.
type GhostDCAConfig struct {
	InitialCapital float64 // seed capital at first bar
	MonthlyInject  float64 // monthly injection amount
}

// SimulateGhostDCA runs a passive DCA benchmark.
//
// Algorithm:
//   1. Buy all initial capital at the first bar's close
//   2. At each calendar month boundary, inject MonthlyInject CNY and buy all
//   3. Track NAV curve for max drawdown calculation
//   4. Compute Modified Dietz ROI (removes injection jumps from NAV)
//
// Bars must be sorted by OpenTime ascending.
func SimulateGhostDCA(bars []Bar, cfg GhostDCAConfig) GhostDCAResult {
	if len(bars) == 0 || cfg.InitialCapital <= 0 {
		return GhostDCAResult{}
	}

	var navs []navPoint
	var cashFlows []struct {
		ts     int64
		amount float64
	}

	shares := 0.0
	cash := cfg.InitialCapital
	totalInjected := cfg.InitialCapital

	// Track current month to detect month boundaries
	currentMonth := int64(-1)

	for i, bar := range bars {
		price := bar.Close
		if price <= 0 {
			continue
		}

		// Detect month boundary for injection
		// Approximate: every 30 days = one injection
		monthIdx := bar.OpenTime / (30 * 24 * 3600 * 1000)
		if i == 0 {
			// Initial buy at first bar
			shares = cash / price
			cash = 0
			currentMonth = monthIdx
		} else if monthIdx > currentMonth {
			// Month boundary: inject and buy
			cash += cfg.MonthlyInject
			totalInjected += cfg.MonthlyInject
			additionalShares := cash / price
			shares += additionalShares
			cashFlows = append(cashFlows, struct {
				ts     int64
				amount float64
			}{bar.OpenTime, cfg.MonthlyInject})
			cash = 0
			currentMonth = monthIdx
		}

		nav := shares*price + cash
		navs = append(navs, navPoint{bar.OpenTime, nav})
	}

	if len(navs) == 0 {
		return GhostDCAResult{TotalInjected: totalInjected}
	}

	// Final equity
	finalEquity := navs[len(navs)-1].nav

	// Max drawdown from NAV curve
	maxDD := MaxDrawdownNAV(navs)

	// Modified Dietz ROI
	roi := ModifiedDietzROI(finalEquity, cfg.InitialCapital, cashFlows, bars[0].OpenTime, bars[len(bars)-1].OpenTime)

	return GhostDCAResult{
		FinalEquity:   finalEquity,
		TotalInjected: totalInjected,
		MaxDrawdown:   maxDD,
		ROI:           roi,
	}
}

// MaxDrawdownNAV computes the maximum drawdown from a NAV curve.
// Returns the ratio (e.g. 0.25 = 25% peak-to-trough decline).
func MaxDrawdownNAV(navs []navPoint) float64 {
	if len(navs) < 2 {
		return 0
	}

	peak := navs[0].nav
	maxDD := 0.0

	for _, p := range navs {
		if p.nav > peak {
			peak = p.nav
		}
		dd := (peak - p.nav) / peak
		if dd > maxDD {
			maxDD = dd
		}
	}
	return maxDD
}

// ModifiedDietzROI computes the Modified Dietz return.
//
// Formula:
//
//	ROI = (EndValue - StartValue - SumCF) /
//	      (StartValue + Sum(CF_i × (TotalDays - DayOfCF) / TotalDays))
//
// This removes injection jumps from NAV, giving a time-weighted return.
func ModifiedDietzROI(
	endValue, startValue float64,
	cashFlows []struct {
		ts     int64
		amount float64
	},
	startTime, endTime int64,
) float64 {
	if startValue <= 0 {
		return 0
	}

	totalDays := float64(endTime-startTime) / (24 * 3600 * 1000)
	if totalDays <= 0 {
		// Handle single-bar case
		return (endValue - startValue) / startValue
	}

	totalCF := 0.0
	weightedCF := 0.0
	for _, cf := range cashFlows {
		totalCF += cf.amount
		daysSinceStart := float64(cf.ts-startTime) / (24 * 3600 * 1000)
		weight := (totalDays - daysSinceStart) / totalDays
		weightedCF += cf.amount * weight
	}

	denominator := startValue + weightedCF
	if math.Abs(denominator) < 1e-10 {
		return 0
	}

	return (endValue - startValue - totalCF) / denominator
}
