package golden_cross

import (
	"github.com/chenthewho/ma-cross-strategy/internal/quant"
)

// ComputeMicro is a thin wrapper around quant.ComputeMicroDecision.
// It maps strategy-level Chromosome fields onto the engine's
// MicroDecisionInput and converts the MicroDecisionOutput into a
// *quant.MicroIntent (or nil if no action is required).
func ComputeMicro(
	closes []float64,
	currentPrice float64,
	totalEquity float64,
	currentMicroWeight float64,
	marketState quant.MarketState,
	chromo quant.Chromosome,
) *quant.MicroIntent {
	in := quant.MicroDecisionInput{
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

	out := quant.ComputeMicroDecision(in)
	if out.OrderCNY == 0 {
		return nil
	}

	action := "BUY"
	if out.OrderCNY < 0 {
		action = "SELL"
	}

	return &quant.MicroIntent{
		Action:    action,
		AmountCNY: out.OrderCNY,
		Engine:    "MICRO",
		LotType:   string(quant.LotFloating),
	}
}
