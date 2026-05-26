package golden_cross

import (
	"math"

	"github.com/chenthewho/ma-cross-strategy/internal/quant"
)

// ComputeMacro wraps quant.ComputeMacroDecision for the golden_cross strategy.
// It builds the MacroDecisionInput from the strategy-level parameters and
// delegates to the quant engine.
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
	if totalEquity <= 0 || currentPrice <= 0 {
		return nil
	}

	deadHoldValue := deadHold * currentPrice

	// Price deviation: (Price - EMA_long) / EMA_long
	var priceDeviation float64
	emaLong := quant.EMA(closes, chromo.EMALongBars)
	if !math.IsNaN(emaLong) && emaLong > 0 {
		priceDeviation = (currentPrice - emaLong) / emaLong
	}

	// Days since last macro action
	daysSinceLastMacro := 0
	if runtime.LastMacroAction > 0 && len(timestamps) > 0 {
		lastBar := timestamps[len(timestamps)-1]
		daysSinceLastMacro = int((lastBar - runtime.LastMacroAction) / (24 * 3600 * 1000))
	}

	input := quant.MacroDecisionInput{
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

	return quant.ComputeMacroDecision(input)
}
