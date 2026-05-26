package quant

import "math"

// ─── Micro-engine constants ──────────────────────────────────────────────────

const (
	microSignalEMABars   = 21   // EMA period for the noise-estimate price series
	microSignalStdDevBars = 21   // lookback for log-return standard deviation
	microVolRatioShortBars = 16  // short window for mean-absolute-change
	microVolRatioLongBars  = 112 // long  window for mean-absolute-change
	minOrderCNY          = 100.0 // minimum notional order size in CNY
)

// ─── MicroInput / MicroOutput ─────────────────────────────────────────────

// MicroInput bundles all parameters needed for one micro-decision evaluation.
type MicroInput struct {
	Closes             []float64 // historical close prices, oldest first
	CurrentPrice       float64   // latest market price
	TotalEquity        float64   // current portfolio equity (CNY)
	CurrentMicroWeight float64   // current micro-engine position weight [0, 1]

	// Strategy chromosome parameters.
	A, B, C     float64 // signal-coefficient weights
	Beta        float64 // base conviction steepness
	Gamma       float64 // inventory-pull coefficient
	SigmaFloor  float64 // minimum allowed volatility floor

	// Market-state modifiers.
	BetaMultiplier float64 // multiplies Beta for regime scaling
	IsQuiet        bool    // suppress forced minimum orders when true

	// EMA lookback bars (from chromosome).
	EMAShortBars int
	EMALongBars  int
}

// MicroOutput holds the result of ComputeMicroDecision.
type MicroOutput struct {
	TargetWeight    float64 // desired micro position weight [0, 1]
	Signal          float64 // raw composite signal (A·X1 + B·X2 + C·X3)
	TheoreticalCNY  float64 // target notional delta (uncapped)
	OrderCNY        float64 // actual CNY order to place
	VolatilityRatio float64 // short-term / long-term volatility ratio
}

// ComputeMicroDecision is the sigmoid-based micro-allocation engine.
//
// # Sigmoid Philosophy
//
// The micro engine converts a composite signal (price-vs-EMA divergence,
// EMA crossover, short-term momentum) into a position weight between 0 and 1
// using the logistic sigmoid 1/(1+eˣ).  Three forces compete inside the
// exponent:
//
//   - **EffectiveBeta · Signal** — how strongly the price evidence tilts the
//     weight.  High Beta makes the sigmoid steeper (more decisive);
//     BetaMultiplier allows the market-state layer to dampen or amplify that
//     steepness for different regimes.
//
//   - **Gamma · InventoryBias** — a mean-reverting pull that draws the target
//     weight toward 0.5 (neutral).  When CurrentMicroWeight is above 0.5 the
//     bias is positive and pushes the exponent higher → sigmoid lower,
//     nudging toward liquidation.  Conversely, a weight below 0.5 produces a
//     negative bias that raises the sigmoid toward accumulation.
//
// The net effect is a smooth, bounded controller: extreme signals cannot
// drive the weight beyond [0,1], and the inventory term ensures the engine
// does not drift indefinitely in one direction without sustained evidence.
//
// After the target weight is computed, volatility-ratio gating and a minimum
// order threshold decide whether an order is emitted.
func ComputeMicroDecision(in MicroInput) MicroOutput {
	// ── Step 1: volatility estimate via log-returns ─────────────────────
	ema := EMA(in.Closes, microSignalEMABars)
	_ = ema // reserved for future use (noise baseline)

	logReturns := make([]float64, 0, len(in.Closes)-1)
	for i := 1; i < len(in.Closes); i++ {
		if in.Closes[i-1] > 0 && in.Closes[i] > 0 {
			logReturns = append(logReturns, math.Log(in.Closes[i]/in.Closes[i-1]))
		}
	}
	sigma := math.Max(StdDev(logReturns, microSignalStdDevBars), in.SigmaFloor)
	if sigma == 0 || math.IsNaN(sigma) {
		return MicroOutput{}
	}

	// ── Step 2: composite signal X1 / X2 / X3 ───────────────────────────
	emaShort := EMA(in.Closes, in.EMAShortBars)
	emaLong  := EMA(in.Closes, in.EMALongBars)

	x1 := 0.0
	if emaShort != 0 {
		x1 = (in.CurrentPrice - emaShort) / (emaShort * sigma)
	}

	x2 := 0.0
	if emaLong != 0 {
		x2 = (emaShort - emaLong) / (emaLong * sigma)
	}

	x3 := 0.0
	if len(in.Closes) >= 6 && in.CurrentPrice != 0 {
		x3 = (in.CurrentPrice - in.Closes[len(in.Closes)-6]) / (in.CurrentPrice * sigma)
	}

	signal := in.A*x1 + in.B*x2 + in.C*x3

	// ── Step 3: sigmoid target weight ───────────────────────────────────
	effectiveBeta := math.Max(0.01, in.Beta*in.BetaMultiplier)
	inventoryBias := ClipFloat64(in.CurrentMicroWeight, 0, 1) - 0.5
	exponent := effectiveBeta*signal + in.Gamma*inventoryBias
	targetWeight := ClipFloat64(1.0/(1.0+math.Exp(exponent)), 0, 1)

	// ── Step 4: delta weight → theoretical CNY ─────────────────────────
	deltaWeight := targetWeight - in.CurrentMicroWeight
	theoreticalCNY := deltaWeight * in.TotalEquity

	// ── Step 5: volatility ratio gating ────────────────────────────────
	volRatio := 1.0
	shortMAV := MAVAbsChange(in.Closes, microVolRatioShortBars)
	longMAV  := MAVAbsChange(in.Closes, microVolRatioLongBars)
	if shortMAV > 0 && longMAV > 0 {
		volRatio = ClipFloat64(shortMAV/longMAV, 0.1, 3.0)
	}

	// ── Step 6: order decision ─────────────────────────────────────────
	var orderCNY float64
	absTheoretical := math.Abs(theoreticalCNY)

	if absTheoretical >= minOrderCNY {
		orderCNY = theoreticalCNY
	} else if in.IsQuiet {
		orderCNY = 0
	} else if math.Abs(deltaWeight) >= 0.02 || volRatio >= 1.5 {
		orderCNY = math.Copysign(minOrderCNY, theoreticalCNY)
	} else {
		orderCNY = 0
	}

	return MicroOutput{
		TargetWeight:    targetWeight,
		Signal:          signal,
		TheoreticalCNY:  theoreticalCNY,
		OrderCNY:        orderCNY,
		VolatilityRatio: volRatio,
	}
}
