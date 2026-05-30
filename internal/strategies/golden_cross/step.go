package golden_cross

import (
	"math"

	"github.com/chenthewho/ma-cross-strategy/internal/quant"
)

// Step is the strategy's sole entry point. Pure function, no I/O.
func Step(input quant.StrategyInput, params Params) quant.StrategyOutput {
	c := params.Chromosome
	sp := params.SpawnPoint

	// ── 1. Data sufficiency ──
	minBars := max(c.EMAShortBars, c.EMALongBars, quant.MarketEMALongBars)
	if len(input.Closes) < minBars {
		return quant.StrategyOutput{
			MarketState: quant.MarketState{State: quant.MarketQuiet, IsQuiet: true,
				TimeDilationMultiplier: 1.0, BetaMultiplier: 1.0},
			NewRuntime: input.Runtime,
		}
	}

	// ── 2. Core metrics ──
	totalEquity := input.Portfolio.TotalEquity
	spendableCNY := math.Max(0, input.Portfolio.CNYBalance-totalEquity*c.MicroReservePct)
	currentMicroWeight := 0.0
	if totalEquity > 0 {
		currentMicroWeight = input.Portfolio.FloatHold / totalEquity
	}

	// ── 3. Market state ──
	marketState := quant.ComputeMarketState(input.Closes, input.Timestamps)

	// ── 4. Macro engine ──
	var macroIntent *quant.MacroIntent
	// DCA is new money injection — always run macro engine if MonthlyInject > 0.
	// (spendableCNY only matters for acceleration/deadline catch-up, not base DCA.)
	runMacro := spendableCNY >= 100 || sp.Policy.MonthlyInject >= 100
	if runMacro {
		emaLong := quant.EMA(input.Closes, c.EMALongBars)
		priceDeviation := 0.0
		if !math.IsNaN(emaLong) && emaLong > 0 {
			priceDeviation = (input.CurrentPrice - emaLong) / emaLong
		}
		var daysSince float64
		if input.Runtime.LastMacroAction > 0 {
			daysSince = float64(input.Timestamps[len(input.Timestamps)-1]-input.Runtime.LastMacroAction) / 86400000.0
		}
		macroIntent = quant.ComputeMacroDecision(quant.MacroDecisionInput{
			TotalEquity:          totalEquity,
			SpendableCNY:         spendableCNY,
			CurrentPrice:         input.CurrentPrice,
			DeadHold:             input.Portfolio.DeadHold,
			DeadHoldValue:        input.Portfolio.DeadHold * input.CurrentPrice,
			MonthlyInject:        sp.Policy.MonthlyInject,
			TimeDilation:         marketState.TimeDilationMultiplier,
			PriceDeviation:       priceDeviation,
			DaysSinceLastMacro:   daysSince,
			MacroIntervalDays:    c.MacroIntervalDays,
			MacroAccelThreshold:  c.MacroAccelThreshold,
			MacroAccelMultiplier: c.MacroAccelMultiplier,
		})
	}

	// ── 5. Micro engine ──
	var microIntent *quant.MicroIntent
	microOut := quant.ComputeMicroDecision(quant.MicroDecisionInput{
		Closes: input.Closes, CurrentPrice: input.CurrentPrice,
		TotalEquity: totalEquity, CurrentMicroWeight: currentMicroWeight,
		IsQuiet: marketState.IsQuiet, BetaMultiplier: marketState.BetaMultiplier,
		A: c.A, B: c.B, C: c.C, Beta: c.Beta, Gamma: c.Gamma,
		SigmaFloor: c.SigmaFloor, EMAShort: c.EMAShortBars, EMALong: c.EMALongBars,
	})
	if microOut.OrderCNY != 0 {
		act := "BUY"
		amt := microOut.OrderCNY
		if amt < 0 {
			act = "SELL"
			amt = -amt
		}
		if act == "BUY" {
			amt = math.Min(amt, spendableCNY)
		}
		if amt >= 100 {
			microIntent = &quant.MicroIntent{Action: act, AmountCNY: quant.RoundToCNY(amt), Engine: "MICRO", LotType: "FLOATING"}
		}
	}

	// ── 6. Dead hold release ──
	var releaseIntent *quant.ReleaseIntent
	if input.Portfolio.DeadHold > 0 && len(input.Portfolio.DeadHoldLots) > 0 {
		lastBar := input.Timestamps[len(input.Timestamps)-1]
		_, softAmt := quant.SoftRelease(input.Portfolio.DeadHoldLots, lastBar, input.CurrentPrice,
			totalEquity, c.SoftReleaseMonths, c.MaxSoftReleasePct, c.DeadHoldTarget)
		if softAmt > 0 {
			releaseIntent = &quant.ReleaseIntent{ReleaseType: "SOFT", ReleaseAmount: softAmt, Reason: "aged_dead_hold"}
		}
		if microIntent != nil && microIntent.Action == "SELL" {
			needShares := microIntent.AmountCNY / input.CurrentPrice
			if quant.TotalFloatHold(input.Portfolio.DeadHoldLots) < needShares {
				_, hardAmt := quant.HardRelease(input.Portfolio.DeadHoldLots, needShares)
				if hardAmt > 0 {
					releaseIntent = &quant.ReleaseIntent{ReleaseType: "HARD", ReleaseAmount: hardAmt, Reason: "micro_sell_shortfall"}
				}
			}
		}
	}

	// ── 7. Update runtime ──
	newRuntime := input.Runtime
	newRuntime.LastProcessedBar = input.Timestamps[len(input.Timestamps)-1]
	if macroIntent != nil {
		newRuntime.LastMacroAction = input.Timestamps[len(input.Timestamps)-1]
	}

	return quant.StrategyOutput{
		MacroIntent: macroIntent, MicroIntent: microIntent,
		ReleaseIntent: releaseIntent, MarketState: marketState, NewRuntime: newRuntime,
	}
}
