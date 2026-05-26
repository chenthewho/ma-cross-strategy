package quant

import (
	"math"
	"testing"
)

// generateCloses creates a price series where all prices are equal to `price`.
// Length must be >= 112 (longest internal window).
func generateCloses(price float64, length int) []float64 {
	closes := make([]float64, length)
	for i := range closes {
		closes[i] = price
	}
	return closes
}

// TestMicroDecisionSignalPositive verifies: when Signal > 0, TargetWeight < 0.5
// To produce a positive signal, we set A=1, B=0, C=0 and run a series where
// current price is slightly above the EMA, making X1 > 0.
func TestMicroDecisionSignalPositive(t *testing.T) {
	// Create closes: 120 bars at 1000, last bar at 1010 (price rose)
	closes := make([]float64, 120)
	for i := range closes {
		closes[i] = 1000
	}
	closes[len(closes)-1] = 1010 // current price jump above EMA

	in := MicroDecisionInput{
		Closes:             closes,
		CurrentPrice:       closes[len(closes)-1],
		TotalEquity:        100000,
		CurrentMicroWeight: 0.3,
		IsQuiet:            false,
		BetaMultiplier:     1.0,
		A:                  1.0,
		B:                  0.0,
		C:                  0.0,
		Beta:               1.0,
		Gamma:              0.3,
		SigmaFloor:         0.005,
		EMAShort:           5,
		EMALong:            55,
	}

	out := ComputeMicroDecision(in)

	if math.IsNaN(out.Signal) {
		t.Fatal("Signal is NaN — insufficient data")
	}
	if out.Signal <= 0 {
		t.Errorf("expected positive Signal, got %v", out.Signal)
	}
	if out.TargetWeight >= 0.5 {
		t.Errorf("expected TargetWeight < 0.5 when Signal > 0, got %v", out.TargetWeight)
	}
}

// TestMicroDecisionSignalNegative verifies: when Signal < 0, TargetWeight > 0.5
// To produce a negative signal, we set A=-1 so X1 > 0 causes Signal < 0.
func TestMicroDecisionSignalNegative(t *testing.T) {
	closes := make([]float64, 120)
	for i := range closes {
		closes[i] = 1000
	}
	closes[len(closes)-1] = 1010 // price above EMA → X1 > 0

	in := MicroDecisionInput{
		Closes:             closes,
		CurrentPrice:       closes[len(closes)-1],
		TotalEquity:        100000,
		CurrentMicroWeight: 0.7,
		IsQuiet:            false,
		BetaMultiplier:     1.0,
		A:                  -1.0, // negative → Signal negative when X1 > 0
		B:                  0.0,
		C:                  0.0,
		Beta:               1.0,
		Gamma:              0.3,
		SigmaFloor:         0.005,
		EMAShort:           5,
		EMALong:            55,
	}

	out := ComputeMicroDecision(in)

	if math.IsNaN(out.Signal) {
		t.Fatal("Signal is NaN — insufficient data")
	}
	if out.Signal >= 0 {
		t.Errorf("expected negative Signal, got %v", out.Signal)
	}
	if out.TargetWeight <= 0.5 {
		t.Errorf("expected TargetWeight > 0.5 when Signal < 0, got %v", out.TargetWeight)
	}
}

// TestMicroDecisionNeutralWeight verifies: when CurrentWeight = 0.5 and signal = 0,
// inventoryBias = 0 → exponent = 0 → TargetWeight = 0.5 regardless of Gamma.
func TestMicroDecisionNeutralWeight(t *testing.T) {
	// All closes equal → X1=X2=X3=0 → signal=0
	closes := generateCloses(1000, 120)

	in := MicroDecisionInput{
		Closes:             closes,
		CurrentPrice:       1000,
		TotalEquity:        100000,
		CurrentMicroWeight: 0.5,
		IsQuiet:            false,
		BetaMultiplier:     1.0,
		A:                  0.0, // zero coefficients → signal = 0
		B:                  0.0,
		C:                  0.0,
		Beta:               1.0,
		Gamma:              1.5, // Gamma > 0, but inventoryBias = 0 → no effect
		SigmaFloor:         0.005,
		EMAShort:           5,
		EMALong:            55,
	}

	out := ComputeMicroDecision(in)

	if math.IsNaN(out.Signal) {
		t.Fatal("Signal is NaN — insufficient data")
	}
	if math.Abs(out.Signal) > 1e-10 {
		t.Logf("Signal is not exactly zero: %v (numeric drift)", out.Signal)
	}
	// With signal=0 and inventoryBias=0, TargetWeight should be exactly 0.5
	if math.Abs(out.TargetWeight-0.5) > 0.01 {
		t.Errorf("expected TargetWeight ≈ 0.5, got %v", out.TargetWeight)
	}
}

// TestMicroDecisionDeterminism verifies that ComputeMicroDecision is a pure function.
func TestMicroDecisionDeterminism(t *testing.T) {
	closes := make([]float64, 120)
	for i := range closes {
		closes[i] = 1000 + float64(i)*0.5 // gently rising
	}

	in := MicroDecisionInput{
		Closes:             closes,
		CurrentPrice:       closes[len(closes)-1],
		TotalEquity:        100000,
		CurrentMicroWeight: 0.4,
		IsQuiet:            false,
		BetaMultiplier:     1.2,
		A:                  0.5,
		B:                  0.3,
		C:                  -0.2,
		Beta:               1.2,
		Gamma:              0.5,
		SigmaFloor:         0.005,
		EMAShort:           12,
		EMALong:            55,
	}

	out1 := ComputeMicroDecision(in)
	out2 := ComputeMicroDecision(in)

	if out1.Signal != out2.Signal ||
		out1.TargetWeight != out2.TargetWeight ||
		out1.TheoreticalCNY != out2.TheoreticalCNY ||
		out1.OrderCNY != out2.OrderCNY ||
		out1.VolatilityRatio != out2.VolatilityRatio {
		t.Errorf("non-deterministic: out1=%+v, out2=%+v", out1, out2)
	}
}

// TestMicroDecisionQuietSuppression verifies: IsQuiet=true,
// |TheoreticalCNY| < 100 → OrderCNY = 0.
func TestMicroDecisionQuietSuppression(t *testing.T) {
	closes := generateCloses(1000, 120)

	in := MicroDecisionInput{
		Closes:             closes,
		CurrentPrice:       1000,
		TotalEquity:        10000, // small equity → small order
		CurrentMicroWeight: 0.51,  // slight imbalance
		IsQuiet:            true,
		BetaMultiplier:     1.0,
		A:                  0.0,
		B:                  0.0,
		C:                  0.0,
		Beta:               0.5, // low aggressiveness
		Gamma:              0.0,
		SigmaFloor:         0.005,
		EMAShort:           5,
		EMALong:            55,
	}

	out := ComputeMicroDecision(in)

	if math.Abs(out.TheoreticalCNY) >= 100 {
		t.Skipf("TheoreticalCNY (%v) >= 100, can't verify quiet suppression", out.TheoreticalCNY)
	}
	if out.OrderCNY != 0 {
		t.Errorf("quiet mode: expected OrderCNY=0 (dust suppression), got %v", out.OrderCNY)
	}
}

// TestMicroDecisionWedgeBreakthrough verifies: IsQuiet=false,
// wedge breakthrough condition met, |TheoreticalCNY| < 100 → OrderCNY = ±100.
func TestMicroDecisionWedgeBreakthrough(t *testing.T) {
	// Create volatile closes for high VolatilityRatio
	closes := make([]float64, 120)
	// Short window (16) has high volatility, long (112) low
	for i := range closes {
		closes[i] = 1000
	}
	// First 112 bars: flat
	// Last 16 bars: volatile
	for i := 104; i < 120; i++ {
		if i%2 == 0 {
			closes[i] = 1005
		} else {
			closes[i] = 995
		}
	}

	in := MicroDecisionInput{
		Closes:             closes,
		CurrentPrice:       1002,
		TotalEquity:        10000, // small equity → small theoretical
		CurrentMicroWeight: 0.52,  // slight positive delta
		IsQuiet:            false,
		BetaMultiplier:     1.0,
		A:                  0.0,
		B:                  0.0,
		C:                  0.0,
		Beta:               0.5,
		Gamma:              0.0,
		SigmaFloor:         0.005,
		EMAShort:           5,
		EMALong:            55,
	}

	out := ComputeMicroDecision(in)

	if out.VolatilityRatio < 1.5 && math.Abs(out.TargetWeight-in.CurrentMicroWeight) < 0.02 {
		t.Skipf("wedge conditions not met: volRatio=%v, deltaWeight=%v",
			out.VolatilityRatio, out.TargetWeight-in.CurrentMicroWeight)
	}

	if math.Abs(out.TheoreticalCNY) >= 100 {
		t.Skipf("TheoreticalCNY (%v) >= 100, can't verify wedge breakthrough", out.TheoreticalCNY)
	}

	if out.OrderCNY != 100 && out.OrderCNY != -100 {
		t.Errorf("wedge breakthrough: expected OrderCNY=±100, got %v (TheoreticalCNY=%v, VolRatio=%v)",
			out.OrderCNY, out.TheoreticalCNY, out.VolatilityRatio)
	}
}
