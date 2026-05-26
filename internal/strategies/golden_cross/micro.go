package golden_cross

import (
	"github.com/chenthewho/ma-cross-strategy/internal/quant"
)

// ComputeMicro wraps quant.ComputeMicroDecision for the golden_cross strategy.
// It builds the MicroDecisionInput from the strategy-level parameters and
// delegates to the quant engine.  The quant engine returns a MicroDecisionOutput;
// this wrapper converts it to a MicroIntent that the backtest can execute.
func ComputeMicro(
	closes []float64,
	currentPrice float64,
	totalEquity float64,
	currentMicroWeight float64,
	marketState quant.MarketState,
	chromo quant.Chromosome,
) *quant.MicroIntent {
	if totalEquity <= 0 || currentPrice <= 0 {
		return nil
	}

	input := quant.MicroDecisionInput{
		Closes:             closes,
		CurrentPrice:       currentPrice,
		TotalEquity:        totalEquity,
		CurrentMicroWeight: currentMicroWeight,
		IsQuiet:            marketState.IsQuiet,
		BetaMultiplier:     marketState.BetaMultiplier,
		A:                  chromo.A,
		B:                  chromo.B,
		C:                  chromo.C,
		Beta:               chromo.Beta,
		Gamma:              chromo.Gamma,
		SigmaFloor:         chromo.SigmaFloor,
		EMAShort:           chromo.EMAShortBars,
		EMALong:            chromo.EMALongBars,
	}

	out := quant.ComputeMicroDecision(input)

	if out.OrderCNY == 0 {
		return nil
	}

	action := "BUY"
	amountCNY := out.OrderCNY
	if out.OrderCNY < 0 {
		action = "SELL"
		amountCNY = -out.OrderCNY
	}

	return &quant.MicroIntent{
		Action:    action,
		AmountCNY: amountCNY,
		Engine:    "MICRO",
		LotType:   "FLOATING",
	}
}
