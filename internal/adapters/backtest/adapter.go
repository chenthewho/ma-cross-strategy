// Package backtest provides a deterministic backtest adapter that runs
// strategy Step() over historical bars and accumulates results.
// This adapter is used for GA evaluation AND for determinism verification.
package backtest

import (
	"math"

	"github.com/chenthewho/ma-cross-strategy/internal/quant"
	gc "github.com/chenthewho/ma-cross-strategy/internal/strategies/golden_cross"
)

// BacktestResult holds the output of a complete backtest run.
type BacktestResult struct {
	FinalEquity    float64
	FinalCNY       float64
	FinalDeadHold  float64
	FinalFloatHold float64
	TotalTrades    int
	MaxDrawdown    float64
	PeakEquity     float64
	EquityCurve    []float64 // equity at each bar
}

// RunBacktest executes the golden_cross strategy over historical bars.
// Returns nil if there is insufficient data.
func RunBacktest(bars []quant.Bar, params gc.Params) *BacktestResult {
	if len(bars) < 55 {
		return nil
	}

	closes := quant.ExtractCloses(bars)
	timestamps := quant.ExtractTimestamps(bars)

	spawn := params.SpawnPoint
	if spawn.Policy.InitialCapital <= 0 {
		spawn.Policy.InitialCapital = 100000
	}

	portfolio := quant.PortfolioSnapshot{
		CNYBalance:  spawn.Policy.InitialCapital,
		TotalEquity: spawn.Policy.InitialCapital,
	}
	runtime := quant.RuntimeState{}

	result := &BacktestResult{
		FinalEquity: spawn.Policy.InitialCapital,
		FinalCNY:    spawn.Policy.InitialCapital,
		PeakEquity:  spawn.Policy.InitialCapital,
		EquityCurve: make([]float64, 0, len(bars)),
	}
	totalTrades := 0

	// Warm-up phase: just accumulate bars without trading
	warmupIdx := 55

	for i := warmupIdx; i < len(bars); i++ {
		currentPrice := closes[i]

		// Portfolio state for this tick
		input := quant.StrategyInput{
			Closes:       closes[:i+1],
			Timestamps:   timestamps[:i+1],
			CurrentPrice: currentPrice,
			Portfolio:    portfolio,
			Runtime:      runtime,
		}

		output := gc.Step(input, params)

		// Execute trade intents
		if output.MacroIntent != nil {
			shares := output.MacroIntent.AmountCNY / currentPrice
			shares = math.Floor(shares/spawn.Risk.LotStep) * spawn.Risk.LotStep
			cost := shares * currentPrice * (1 + spawn.Risk.FeeRate + spawn.Risk.Slippage)
			if shares > 0 && portfolio.CNYBalance >= cost {
				portfolio.DeadHold += shares
				portfolio.CNYBalance -= cost
				portfolio.DeadHoldLots = append(portfolio.DeadHoldLots, quant.SpotLot{
					LotType:   quant.LotDeadStack,
					Amount:    shares,
					CostPrice: currentPrice,
					CreatedAt: timestamps[i],
				})
				totalTrades++
			}
		}

		if output.MicroIntent != nil {
			shares := output.MicroIntent.AmountCNY / currentPrice
			shares = math.Floor(shares/spawn.Risk.LotStep) * spawn.Risk.LotStep
			cost := math.Abs(shares * currentPrice * (1 + spawn.Risk.FeeRate + spawn.Risk.Slippage))
			if output.MicroIntent.Action == "BUY" && shares > 0 && portfolio.CNYBalance >= cost {
				portfolio.FloatHold += shares
				portfolio.CNYBalance -= cost
				totalTrades++
			} else if output.MicroIntent.Action == "SELL" && shares > 0 && portfolio.FloatHold >= shares {
				portfolio.FloatHold -= shares
				portfolio.CNYBalance += shares*currentPrice*(1-spawn.Risk.FeeRate-spawn.Risk.Slippage)
				totalTrades++
			}
		}

		// Update portfolio equity
		equity := portfolio.CNYBalance +
			portfolio.DeadHold*currentPrice +
			portfolio.FloatHold*currentPrice +
			portfolio.ColdSealedHold*currentPrice
		portfolio.TotalEquity = equity

		if equity > result.PeakEquity {
			result.PeakEquity = equity
		}
		dd := (result.PeakEquity - equity) / result.PeakEquity
		if dd > result.MaxDrawdown {
			result.MaxDrawdown = dd
		}

		result.EquityCurve = append(result.EquityCurve, equity)
		runtime = output.NewRuntime
	}

	result.FinalEquity = portfolio.TotalEquity
	result.FinalCNY = portfolio.CNYBalance
	result.FinalDeadHold = portfolio.DeadHold
	result.FinalFloatHold = portfolio.FloatHold
	result.TotalTrades = totalTrades

	return result
}
