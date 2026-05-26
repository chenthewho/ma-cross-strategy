package quant

import "sort"

// CrucibleWindow represents a time-window slice of bars used for backtesting
// with an associated label and weight for result aggregation.
type CrucibleWindow struct {
	Label      string
	Weight     float64
	Bars       []Bar
	EvalStartMs int64
}

// CrucibleResult holds performance metrics for a single crucible window.
type CrucibleResult struct {
	Window string
	Score  float64
	ROI    float64
	MaxDD  float64
	Alpha  float64
}

// windowDef is an internal helper for defining crucible window parameters.
type windowDef struct {
	Label    string
	EvalDays int // 0 means "all bars" (full window)
	Weight   float64
}

// BuildCrucibleWindows slices bars into multiple time windows for
// multi-window backtesting. Bars must be sorted by OpenTime ascending.
// warmupDays specifies how many days of warmup data to include before the
// evaluation start (warmup < EvalStartMs always — no future data leak).
//
// Windows:
//
//	"6m"  — 183-day eval,  weight 0.10
//	"2y"  — 730-day eval,  weight 0.20
//	"5y"  — 1825-day eval, weight 0.30
//	"full" — all bars,      weight 0.40
//
// Returned windows are sorted by len(Bars) ascending.
func BuildCrucibleWindows(bars []Bar, warmupDays int) []CrucibleWindow {
	if len(bars) == 0 {
		return nil
	}

	latest := bars[len(bars)-1].OpenTime

	defs := []windowDef{
		{Label: "6m", EvalDays: 183, Weight: 0.10},
		{Label: "2y", EvalDays: 730, Weight: 0.20},
		{Label: "5y", EvalDays: 1825, Weight: 0.30},
		{Label: "full", EvalDays: 0, Weight: 0.40},
	}

	msPerDay := int64(24 * 3600 * 1000)

	windows := make([]CrucibleWindow, 0, len(defs))

	for _, d := range defs {
		var evalStartMs int64
		if d.EvalDays > 0 {
			evalStartMs = latest - int64(d.EvalDays)*msPerDay
		} else {
			// "full" window: evaluation covers the entire data set
			evalStartMs = bars[0].OpenTime
		}

		warmupStartMs := evalStartMs - int64(warmupDays)*msPerDay

		// Collect bars from warmupStartMs onward.
		// warmupStartMs is always strictly less than evalStartMs,
		// guaranteeing no future data leak.
		var windowBars []Bar
		for _, b := range bars {
			if b.OpenTime >= warmupStartMs {
				windowBars = append(windowBars, b)
			}
		}

		windows = append(windows, CrucibleWindow{
			Label:       d.Label,
			Weight:      d.Weight,
			Bars:        windowBars,
			EvalStartMs: evalStartMs,
		})
	}

	// Sort by number of bars ascending (shortest window first).
	sort.Slice(windows, func(i, j int) bool {
		return len(windows[i].Bars) < len(windows[j].Bars)
	})

	return windows
}
