package backtest

import (
	"math"
	"testing"

	"github.com/chenthewho/ma-cross-strategy/internal/quant"
	gc "github.com/chenthewho/ma-cross-strategy/internal/strategies/golden_cross"
)

// generateSyntheticBars creates a simple synthetic bar series for testing.
// N bars, prices gently trending + small noise.
func generateSyntheticBars(n int) []quant.Bar {
	bars := make([]quant.Bar, n)
	price := 1000.0
	for i := 0; i < n; i++ {
		price += math.Sin(float64(i)*0.02) * 2.0 + 0.5
		// Clamp to positive
		if price < 100 {
			price = 100
		}
		bars[i] = quant.Bar{
			OpenTime: int64(i * 3600 * 1000), // hourly
			Open:     price - 0.5,
			High:     price + 1,
			Low:      price - 1,
			Close:    price,
			Volume:   1000,
		}
	}
	return bars
}

// TestBacktestDeterminism verifies that running the same backtest twice
// with the same parameters produces byte-identical results.
func TestBacktestDeterminism(t *testing.T) {
	bars := generateSyntheticBars(200)

	chr := quant.DefaultSeedChromosome
	spawn := quant.DefaultSpawnPoint
	params := gc.Params{Chromosome: chr, SpawnPoint: spawn}

	result1 := RunBacktest(bars, params)
	if result1 == nil {
		t.Fatal("RunBacktest returned nil")
	}

	result2 := RunBacktest(bars, params)
	if result2 == nil {
		t.Fatal("second RunBacktest returned nil")
	}

	// Assert all output fields are exactly equal
	if result1.FinalEquity != result2.FinalEquity {
		t.Errorf("FinalEquity: %v != %v", result1.FinalEquity, result2.FinalEquity)
	}
	if result1.FinalCNY != result2.FinalCNY {
		t.Errorf("FinalCNY: %v != %v", result1.FinalCNY, result2.FinalCNY)
	}
	if result1.FinalDeadHold != result2.FinalDeadHold {
		t.Errorf("FinalDeadHold: %v != %v", result1.FinalDeadHold, result2.FinalDeadHold)
	}
	if result1.FinalFloatHold != result2.FinalFloatHold {
		t.Errorf("FinalFloatHold: %v != %v", result1.FinalFloatHold, result2.FinalFloatHold)
	}
	if result1.TotalTrades != result2.TotalTrades {
		t.Errorf("TotalTrades: %v != %v", result1.TotalTrades, result2.TotalTrades)
	}
	if result1.MaxDrawdown != result2.MaxDrawdown {
		t.Errorf("MaxDrawdown: %v != %v", result1.MaxDrawdown, result2.MaxDrawdown)
	}
	if result1.PeakEquity != result2.PeakEquity {
		t.Errorf("PeakEquity: %v != %v", result1.PeakEquity, result2.PeakEquity)
	}

	if len(result1.EquityCurve) != len(result2.EquityCurve) {
		t.Fatalf("EquityCurve length: %d != %d", len(result1.EquityCurve), len(result2.EquityCurve))
	}
	for i := range result1.EquityCurve {
		if result1.EquityCurve[i] != result2.EquityCurve[i] {
			t.Fatalf("EquityCurve[%d]: %v != %v", i, result1.EquityCurve[i], result2.EquityCurve[i])
		}
	}
}
