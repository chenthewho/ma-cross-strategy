// Package main is the CLI entry point for ma-cross — an AI-enhanced
// cryptocurrency quantitative trading backtester.
//
// Usage:
//
//	ma-cross run BTCUSDT 5 20 1h          # backtest
//	ma-cross price BTCUSDT                  # real-time price
//	ma-cross sentiment                      # fear & greed index
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ma-cross",
	Short: "MA Cross Strategy — AI-enhanced quantitative backtester",
	Long: `MA Cross Strategy is a multi-dimensional signal fusion backtester
that combines SMA cross, RSI, MACD, and market sentiment (Fear & Greed Index)
to generate trading signals and simulate trades with fees and slippage.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
