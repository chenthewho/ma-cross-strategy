// Package golden_cross implements the golden_cross dynamic equilibrium strategy.
//
// Strategy ID: golden_cross
// Type: spot (A-share / gold ETF), IsSpot = true
//
// This is a pure function package. The sole entry point is Step().
// The package must NOT import any I/O packages (net/http, database/sql, os, time.Now(), etc.).
//
// Architecture:
//
//	Step(StrategyInput) -> StrategyOutput
//	   ├── ComputeMarketState     (market regime classification)
//	   ├── ComputeMacroDecision   (DCA buy engine)
//	   ├── ComputeMicroDecision   (Sigmoid dynamic balance)
//	   └── ComputeDeadRelease     (soft/hard release)
//
// Iron law: the same Step() function is used for both backtesting and live trading.
// No `if isBacktest` branches allowed.
package golden_cross
