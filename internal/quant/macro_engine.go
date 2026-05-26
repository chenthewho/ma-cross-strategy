package quant

// MacroInput bundles all the parameters needed to compute a macro-level
// trading decision.
type MacroInput struct {
	TotalEquity   float64 // Total portfolio equity in CNY
	SpendableCNY  float64 // Available CNY balance for trading
	CurrentPrice  float64 // Current asset price
	DeadHoldValue float64 // Total value locked in dead-hold lots
	MonthlyInject float64 // Scheduled monthly CNY injection

	TimeDilation float64 // Time-dilation multiplier from market state
	EMALong      float64 // Long-term EMA value for price deviation

	MacroIntervalDays    int     // Minimum days between macro actions
	MacroAccelThreshold  float64 // Price deviation threshold for acceleration
	MacroAccelMultiplier float64 // Acceleration multiplier on monthly injection

	DaysSinceLastMacro int // Days elapsed since the last macro action
}

// MacroOutput carries the computed macro-level order decision.
type MacroOutput struct {
	OrderCNY float64 // CNY amount to allocate (0 if no action)
	Action   string  // "BUY" or "" (empty string means no action)
}

// ComputeMacroDecision evaluates whether a macro-level buy order should be
// placed based on time-based injection schedules and price deviation from
// the long-term EMA.
//
// Rules:
//  1. Base order is 0 by default.
//  2. If enough days have passed (adjusted by time dilation), base = MonthlyInject.
//  3. If price is significantly below the long EMA, accelerate by adding
//     MonthlyInject * MacroAccelMultiplier.
//  4. If spendable CNY exceeds 3× MonthlyInject, boost base to at least
//     min(SpendableCNY, MonthlyInject * 2).
//  5. Final order CNY is clamped to [100, SpendableCNY].
//  6. Action is "BUY" if OrderCNY >= 100, otherwise "".
func ComputeMacroDecision(in MacroInput) MacroOutput {
	base := 0.0

	// Time-based injection: if enough calendar days have passed (adjusted by
	// time dilation), the base is the monthly injection amount.
	effectiveInterval := float64(in.MacroIntervalDays) * in.TimeDilation
	if effectiveInterval < 1 {
		effectiveInterval = 1 // safety: never allow zero/negative
	}
	if float64(in.DaysSinceLastMacro) >= effectiveInterval {
		base = in.MonthlyInject
	}

	// Price deviation from long EMA.
	if in.EMALong > 0 {
		priceDev := (in.CurrentPrice - in.EMALong) / in.EMALong

		// Acceleration: if price is significantly below the long EMA.
		if priceDev < -in.MacroAccelThreshold {
			base += in.MonthlyInject * in.MacroAccelMultiplier
		}
	}

	// Liquidity boost: if we have excess spendable CNY (>3× monthly inject),
	// increase base to at least min(SpendableCNY, MonthlyInject*2).
	if in.SpendableCNY > in.MonthlyInject*3 {
		floor := in.MonthlyInject * 2
		if in.SpendableCNY < floor {
			floor = in.SpendableCNY
		}
		if base < floor {
			base = floor
		}
	}

	// Clamp final order amount to valid range.
	orderCNY := ClipFloat64(base, 100, in.SpendableCNY)

	action := ""
	if orderCNY >= 100 {
		action = "BUY"
	}

	return MacroOutput{
		OrderCNY: orderCNY,
		Action:   action,
	}
}
