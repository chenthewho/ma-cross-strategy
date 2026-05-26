package golden_cross

import (
	"math"

	"github.com/chenthewho/ma-cross-strategy/internal/quant"
)

// ComputeMacro is a thin wrapper around quant.ComputeMacroDecision.
// It translates strategy-level parameters (Chromosome, SpawnPoint, close
// sequence) into the quant.MacroDecisionInput expected by the engine.
func ComputeMacro(
	closes []float64,
	timestamps []int64,
	currentPrice float64,
	totalEquity float64,
	spendableCNY float64,
	deadHold float64,
	marketState quant.MarketState,
	runtime quant.RuntimeState,
	chromo quant.Chromosome,
	spawn quant.SpawnPoint,
) *quant.MacroIntent {
	// Compute long EMA for price-deviation calculation.
	emaLong := quant.EMA(closes, chromo.EMALongBars)
	if math.IsNaN(emaLong) {
		return nil
	}

	priceDeviation := (currentPrice - emaLong) / emaLong

	// Days since last macro action.
	daysSinceLastMacro := 0
	if runtime.LastMacroAction > 0 && len(timestamps) > 0 {
		lastTs := timestamps[len(timestamps)-1]
		daysSinceLastMacro = int((lastTs - runtime.LastMacroAction) / (24 * 3600 * 1000))
	} else {
		// Never acted — force eligibility.
		daysSinceLastMacro = chromo.MacroIntervalDays + 1
	}

	deadHoldValue := deadHold * currentPrice

	in := quant.MacroDecisionInput{
		TotalEquity:          totalEquity,
		SpendableCNY:         spendableCNY,
		CurrentPrice:         currentPrice,
		DeadHold:             deadHold,
		DeadHoldValue:        deadHoldValue,
		MonthlyInject:        spawn.Policy.MonthlyInject,
		TimeDilation:         marketState.TimeDilationMultiplier,
		PriceDeviation:       priceDeviation,
		DaysSinceLastMacro:   daysSinceLastMacro,
		MacroIntervalDays:    chromo.MacroIntervalDays,
		MacroAccelThreshold:  chromo.MacroAccelThreshold,
		MacroAccelMultiplier: chromo.MacroAccelMultiplier,
	}

	return quant.ComputeMacroDecision(in)
}
