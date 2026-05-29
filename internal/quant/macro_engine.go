package quant

import "math"

// ComputeMacroDecision generates a DCA buy intention from the macro engine.
//
// Three trigger conditions (evaluated in order):
//   1. BASE DCA: DaysSinceLastMacro >= MacroIntervalDays × TimeDilation
//      → order = MonthlyInject
//   2. ACCELERATION: PriceDeviation < -MacroAccelThreshold
//      → extra buy = MonthlyInject × MacroAccelMultiplier (resets counter)
//   3. DEADLINE: SpendableCNY > MonthlyInject × 3
//      → force DCA, amount = min(SpendableCNY, MonthlyInject × 2)
//
// Iron law: only BUY intentions, never SELL.
// Minimum order: 100 CNY.
func ComputeMacroDecision(in MacroDecisionInput) *MacroIntent {
	var orderCNY float64

	// ── Trigger 1: Base DCA ──
	effectiveInterval := float64(in.MacroIntervalDays) * in.TimeDilation
	baseTriggered := effectiveInterval > 0 && float64(in.DaysSinceLastMacro) >= effectiveInterval

	// ── Trigger 2: Acceleration (price far below long EMA) ──
	accelTriggered := in.PriceDeviation < -in.MacroAccelThreshold

	// ── Trigger 3: Deadline (capital piling up) ──
	deadlineTriggered := in.SpendableCNY > in.MonthlyInject*3

	if !baseTriggered && !accelTriggered && !deadlineTriggered {
		return nil
	}

	switch {
	case accelTriggered:
		// Acceleration overrides base — extra buy
		orderCNY = in.MonthlyInject * in.MacroAccelMultiplier

	case baseTriggered:
		// Base DCA, optionally amplified by deadline catch-up
		orderCNY = in.MonthlyInject
		if deadlineTriggered {
			deadlineOrder := math.Min(in.SpendableCNY, in.MonthlyInject*2)
			orderCNY = math.Max(orderCNY, deadlineOrder)
		}

	default:
		return nil
	}

	// Clamp to available funds and minimum order
	orderCNY = ClipFloat64(orderCNY, 100, in.SpendableCNY)
	if orderCNY < 100 {
		return nil
	}

	return &MacroIntent{
		Action:    "BUY",
		AmountCNY: RoundToCNY(orderCNY),
		Engine:    "MACRO",
		LotType:   "DEAD_STACK",
	}
}
