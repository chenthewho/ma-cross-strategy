// +build ignore

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/chenthewho/ma-cross-strategy/internal/quant"
	gc "github.com/chenthewho/ma-cross-strategy/internal/strategies/golden_cross"
)

func main() {
	dsn := "host=/var/run/postgresql user=postgres dbname=quantsaas sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("DB error:", err)
		os.Exit(1)
	}

	type KLine struct {
		OpenTime int64
		Open     float64
		High     float64
		Low      float64
		Close    float64
		Volume   float64
	}

	var rawBars []KLine
	db.Table("k_lines").
		Where("symbol = ? AND interval = ?", "BTCUSDT", "1h").
		Order("open_time ASC").
		Limit(200).
		Find(&rawBars)

	var bars []quant.Bar
	for _, k := range rawBars {
		bars = append(bars, quant.Bar{
			OpenTime: k.OpenTime,
			Open:     k.Open, High: k.High, Low: k.Low,
			Close: k.Close, Volume: k.Volume,
		})
	}

	// Params: old vs new
	type ParamSet struct {
		Name  string
		Gamma float64
		Beta  float64
		B     float64
		C     float64
	}

	oldP := ParamSet{"旧参数", 0.3, 1.0, 0.3, 0.1}
	newP := ParamSet{"新参数", 1.5, 0.8, 0.08, 0.04}

	for _, ps := range []ParamSet{oldP, newP} {
		fmt.Printf("\n━━━ %s (Gamma=%.1f Beta=%.1f B=%.2f C=%.02f) ━━━\n", ps.Name, ps.Gamma, ps.Beta, ps.B, ps.C)
		fmt.Printf("%-6s %-10s %-6s %8s %8s %8s %8s %s\n",
			"Tick", "时间", "信号", "仓位%", "目标%", "delta%", "金额¥", "方向")

		params := gc.Params{
			Chromosome: quant.DefaultSeedChromosome,
			SpawnPoint: quant.DefaultSpawnPoint,
		}
		params.Chromosome.Gamma = ps.Gamma
		params.Chromosome.Beta = ps.Beta
		params.Chromosome.B = ps.B
		params.Chromosome.C = ps.C
		params.Chromosome.A = 0
		params.Chromosome.MacroIntervalDays = 1
		params.Chromosome.EMAShortBars = 21
		params.Chromosome.EMALongBars = 55
		params.SpawnPoint.Policy.InitialCapital = 10000
		params.SpawnPoint.Policy.MonthlyInject = 100

		var runtime quant.RuntimeState
		totalEquity := 10000.0
		cnyBalance := 10000.0
		floatHold := 0.0
		flipCount := 0
		lastAction := ""

		minBars := max(params.Chromosome.EMAShortBars, params.Chromosome.EMALongBars, quant.MarketEMALongBars)

		for i := minBars; i < len(bars); i++ {
			window := bars[:i+1]
			closes := quant.ExtractCloses(window)
			timestamps := quant.ExtractTimestamps(window)
			currentPrice := closes[len(closes)-1]

			currentMicroWeight := 0.0
			if totalEquity > 0 {
				currentMicroWeight = floatHold * currentPrice / totalEquity
			}

			input := quant.StrategyInput{
				Closes:       closes,
				Timestamps:   timestamps,
				CurrentPrice: currentPrice,
				Portfolio: quant.PortfolioSnapshot{
					CNYBalance:  cnyBalance,
					FloatHold:   floatHold,
					TotalEquity: totalEquity,
				},
				Runtime: runtime,
			}

			output := gc.Step(input, params)

			// Apply trade
			if output.MicroIntent != nil && output.MicroIntent.AmountCNY >= 100 {
				amt := output.MicroIntent.AmountCNY
				act := output.MicroIntent.Action

				if act == "BUY" {
					_ = amt / currentPrice
					cnyBalance -= amt
					floatHold += amt
					totalEquity = cnyBalance + floatHold
				} else {
					available := floatHold
					if amt > available {
						amt = available
					}
					cnyBalance += amt
					floatHold -= amt
					totalEquity = cnyBalance + floatHold
				}

				newWeight := floatHold * currentPrice / totalEquity
				flip := ""
				if lastAction != "" && lastAction != act {
					flipCount++
					flip = " ⚡翻转"
				}
				lastAction = act

				fmt.Printf("%-6d %-10d %-6.3f %7.1f%% %7.1f%% %+7.1f%% %8.0f %s%s\n",
					i-minBars+1,
					bars[i].OpenTime,
					output.MarketState.State, // placeholder
					currentMicroWeight*100,
					newWeight*100,
					(newWeight-currentMicroWeight)*100,
					amt,
					act,
					flip,
				)
			}

			runtime = output.NewRuntime
		}
		fmt.Printf("  翻转次数: %d\n", flipCount)
	}
}

// Marshal helpers
func init() {
	// quiet unused imports
	_ = json.Marshal
	_ = strings.Join
	_ = sort.Ints
}
