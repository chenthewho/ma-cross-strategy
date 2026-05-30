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
	// ── Trigger 1: Base DCA ──
	effectiveInterval := float64(in.MacroIntervalDays) * in.TimeDilation
	baseTriggered := effectiveInterval > 0 && in.DaysSinceLastMacro >= effectiveInterval

	// ── Trigger 2: Acceleration (price far below long EMA) ──
	accelTriggered := in.PriceDeviation < -in.MacroAccelThreshold

	// ── Trigger 3: Deadline (capital piling up) ──
	deadlineTriggered := in.SpendableCNY > in.MonthlyInject*3

	if !baseTriggered && !accelTriggered && !deadlineTriggered {
		return nil
	}

	// Base DCA is new money injection — not constrained by spendableCNY.
	if baseTriggered {
		orderCNY := in.MonthlyInject
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

	// Accel / Deadline use available funds.
	var orderCNY float64
	switch {
	case accelTriggered:
		orderCNY = in.MonthlyInject * in.MacroAccelMultiplier
	case deadlineTriggered:
		orderCNY = math.Min(in.SpendableCNY, in.MonthlyInject*2)
	default:
		return nil
	}

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
