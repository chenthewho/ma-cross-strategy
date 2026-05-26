package quant

import (
	"math"
)

// Micro-engine constants (non-evolvable).
const (
	MicroSignalEMABars    = 21   // Signal EMA window
	MicroSignalStdDevBars = 21   // Volatility window
	MicroVolRatioShortBars = 16  // Wedge filter short window
	MicroVolRatioLongBars  = 112 // Wedge filter long window
	MinOrderCNY            = 100 // Minimum order size (dust suppression)
)

// ─── Sigmoid Dynamic Balance ───────────────────────────────
//
// Design philosophy:
//   Signal    = external force (market signal)
//   InventoryBias = spring restoring force toward 0.5
//   Beta      = spring stiffness (aggressiveness)
//   Gamma     = whether the spring is enabled
//   VolatilityRatio = wedge filter to suppress quiet-period dust

// ComputeMicroDecision runs the full sigmoid dynamic balance engine.
//
// Steps:
//   1. Compute EMA and sigma (volatility)
//   2. Compute dimensionless 3-factor signal
//   3. Sigmoid target weight
//   4. Theoretical order
//   5. Volatility ratio
//   6. Wedge filter → final OrderCNY
func ComputeMicroDecision(in MicroDecisionInput) MicroDecisionOutput {
	out := MicroDecisionOutput{}

	// ── Step 1: EMA and sigma ──
	ema := EMA(in.Closes, MicroSignalEMABars)
	logRets := LogReturns(in.Closes)
	sigma := math.Max(StdDev(logRets, MicroSignalStdDevBars), in.SigmaFloor)
	if sigma <= 0 || math.IsNaN(ema) || math.IsNaN(sigma) {
		return out // insufficient data
	}

	// ── Step 2: Dimensionless 3-factor signal ──
	// X1 = price deviation from short EMA (normalized)
	X1 := (in.CurrentPrice - ema) / (ema * sigma)

	// X2 = EMA divergence (short EMA vs long EMA)
	emaShort := EMA(in.Closes, in.EMAShort)
	emaLong := EMA(in.Closes, in.EMALong)
	var X2 float64
	if !math.IsNaN(emaShort) && !math.IsNaN(emaLong) && emaLong > 0 {
		X2 = (emaShort - emaLong) / (emaLong * sigma)
	}

	// X3 = 5-period momentum
	var X3 float64
	if len(in.Closes) >= 6 {
		refClose := in.Closes[len(in.Closes)-6] // 5 bars back
		if refClose > 0 {
			X3 = (in.CurrentPrice - refClose) / (in.CurrentPrice * sigma)
		}
	}

	signal := in.A*X1 + in.B*X2 + in.C*X3
	out.Signal = signal

	// ── Step 3: Sigmoid target weight ──
	effectiveBeta := math.Max(0.01, in.Beta*in.BetaMultiplier)
	inventoryBias := ClipFloat64(in.CurrentMicroWeight, 0, 1) - 0.5
	exponent := effectiveBeta*signal + in.Gamma*inventoryBias
	out.TargetWeight = ClipFloat64(1.0/(1.0+math.Exp(exponent)), 0, 1)

	// ── Step 4: Theoretical order ──
	deltaWeight := out.TargetWeight - in.CurrentMicroWeight
	out.TheoreticalCNY = deltaWeight * in.TotalEquity

	// ── Step 5: Volatility ratio ──
	mavShort := MAVAbsChange(in.Closes, MicroVolRatioShortBars)
	mavLong := MAVAbsChange(in.Closes, MicroVolRatioLongBars)
	if math.IsNaN(mavShort) || math.IsNaN(mavLong) || mavLong == 0 {
		out.VolatilityRatio = 1.0
	} else {
		out.VolatilityRatio = ClipFloat64(mavShort/mavLong, 0.1, 3.0)
	}

	// ── Step 6: Wedge filter ──
	absCNY := math.Abs(out.TheoreticalCNY)
	if absCNY >= MinOrderCNY {
		// Direct order — above minimum threshold
		out.OrderCNY = out.TheoreticalCNY
	} else if in.IsQuiet {
		// Quiet state — dust orders go to zero
		out.OrderCNY = 0
	} else if math.Abs(deltaWeight) >= 0.02 || out.VolatilityRatio >= 1.5 {
		// Wedge breakthrough — force minimum order
		if out.TheoreticalCNY > 0 {
			out.OrderCNY = MinOrderCNY
		} else if out.TheoreticalCNY < 0 {
			out.OrderCNY = -MinOrderCNY
		}
	} else {
		out.OrderCNY = 0
	}

	return out
}
