package golden_cross

import (
	"github.com/chenthewho/ma-cross-strategy/internal/quant"
)

// Step is the single entry point for the golden_cross strategy.
// It implements the full 8-step pipeline described in doc/策略数学引擎.md:
//
//  1. Data sufficiency check
//  2. Calculate TotalEquity / SpendableCNY / CurrentMicroWeight
//  3. Market state perception (quant.ComputeMarketState)
//  4. Macro engine decision (macro.go wrapper)
//  5. Micro engine decision (micro.go wrapper)
//  6. Dead-hold release rules (dead_release.go wrapper)
//  7. Update RuntimeState
//  8. Assemble and return StrategyOutput
//
// Step is a pure function: no HTTP, SQL, file I/O, or external side effects.
// Backtest and live trading call the same implementation — there are no
// isBacktest branches.
func Step(input quant.StrategyInput, params Params) quant.StrategyOutput {
	chromo := params.Chromosome

	// ── 1. Data sufficiency ──────────────────────────────────────────
	minBars := max(chromo.EMAShortBars, chromo.EMALongBars, 21)
	if len(input.Closes) < minBars {
		return emptyOutput(input.Runtime)
	}

	// ── 2. Core calculations ─────────────────────────────────────────
	totalEquity := input.Portfolio.TotalEquity

	spendableCNY := 0.0
	if totalEquity > 0 {
		spendableCNY = max(0, input.Portfolio.CNYBalance-totalEquity*chromo.MicroReservePct)
	}

	currentMicroWeight := 0.0
	if totalEquity > 0 && input.CurrentPrice > 0 {
		currentMicroWeight = input.Portfolio.FloatHold * input.CurrentPrice / totalEquity
	}

	// ── 3. Market state ──────────────────────────────────────────────
	marketState := quant.ComputeMarketState(input.Closes, input.Timestamps)

	// ── 4. Macro engine ──────────────────────────────────────────────
	macroIntent := ComputeMacro(
		input.Closes,
		input.Timestamps,
		input.CurrentPrice,
		totalEquity,
		spendableCNY,
		input.Portfolio.DeadHold,
		marketState,
		input.Runtime,
		chromo,
		params.SpawnPoint,
	)

	// ── 5. Micro engine ──────────────────────────────────────────────
	microIntent := ComputeMicro(
		input.Closes,
		input.CurrentPrice,
		totalEquity,
		currentMicroWeight,
		marketState,
		chromo,
	)

	// ── 6. Dead release ──────────────────────────────────────────────
	currentTime := int64(0)
	if len(input.Timestamps) > 0 {
		currentTime = input.Timestamps[len(input.Timestamps)-1]
	}
	releaseIntent := ComputeDeadRelease(
		input.Portfolio.DeadHoldLots,
		currentTime,
		input.CurrentPrice,
		totalEquity,
		input.Portfolio.FloatHold,
		microIntent,
		chromo,
	)

	// ── 7. Update RuntimeState ───────────────────────────────────────
	newRuntime := input.Runtime
	if len(input.Timestamps) > 0 {
		newRuntime.LastProcessedBar = input.Timestamps[len(input.Timestamps)-1]
	}
	if macroIntent != nil && len(input.Timestamps) > 0 {
		newRuntime.LastMacroAction = input.Timestamps[len(input.Timestamps)-1]
	}

	// ── 8. Assemble and return ───────────────────────────────────────
	return quant.StrategyOutput{
		MacroIntent:   macroIntent,
		MicroIntent:   microIntent,
		ReleaseIntent: releaseIntent,
		MarketState:   marketState,
		NewRuntime:    newRuntime,
	}
}

// emptyOutput returns a safe zero-value StrategyOutput when data is
// insufficient for any decision-making.
func emptyOutput(runtime quant.RuntimeState) quant.StrategyOutput {
	return quant.StrategyOutput{
		MarketState: quant.MarketState{
			State:                  quant.MarketQuiet,
			TimeDilationMultiplier: 1.0,
			BetaMultiplier:         1.0,
			IsQuiet:                true,
		},
		NewRuntime: runtime,
	}
}
