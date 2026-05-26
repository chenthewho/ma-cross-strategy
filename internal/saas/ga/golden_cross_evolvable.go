// Package ga provides the genetic algorithm evolution engine.
// This file implements EvolvableStrategy for the golden_cross strategy.
package ga

import (
	"context"
	"encoding/json"
	"math"
	"math/rand"

	"github.com/chenthewho/ma-cross-strategy/internal/quant"
	gc "github.com/chenthewho/ma-cross-strategy/internal/strategies/golden_cross"
)

// GoldenCrossEvolvable adapts the golden_cross strategy to the EvolvableStrategy interface.
type GoldenCrossEvolvable struct{}

func (e *GoldenCrossEvolvable) StrategyID() string { return gc.StrategyID }

func (e *GoldenCrossEvolvable) Sample(rng *rand.Rand) Gene {
	c := quant.DefaultSeedChromosome
	c.A = rng.Float64()*6 - 3
	c.B = rng.Float64()*6 - 3
	c.C = rng.Float64()*6 - 3
	c.Beta = 0.1 + rng.Float64()*4.9
	c.Gamma = rng.Float64() * 2.0
	c.SigmaFloor = 0.001 + rng.Float64()*0.049
	c.MacroIntervalDays = 7 + rng.Intn(84)
	c.MacroAccelThreshold = 0.01 + rng.Float64()*0.19
	c.MacroAccelMultiplier = 1.0 + rng.Float64()*4.0
	c.SoftReleaseMonths = 1 + rng.Intn(24)
	c.MaxSoftReleasePct = 0.05 + rng.Float64()*0.45
	c.DeadHoldTarget = 0.10 + rng.Float64()*0.80
	c.MicroReservePct = 0.05 + rng.Float64()*0.45
	c.EMAShortBars = 5 + rng.Intn(51)
	c.EMALongBars = 20 + rng.Intn(181)
	quant.ClampChromosome(&c)
	return &c
}

func (e *GoldenCrossEvolvable) Mutate(gene Gene, prob, scale float64, rng *rand.Rand) Gene {
	c, ok := gene.(*quant.Chromosome)
	if !ok { return gene }
	steps := quant.GeneStep()
	if rng.Float64() < prob { c.A += rng.NormFloat64() * steps["A"] * scale }
	if rng.Float64() < prob { c.B += rng.NormFloat64() * steps["B"] * scale }
	if rng.Float64() < prob { c.C += rng.NormFloat64() * steps["C"] * scale }
	if rng.Float64() < prob { c.Beta += rng.NormFloat64() * steps["Beta"] * scale }
	if rng.Float64() < prob { c.Gamma += rng.NormFloat64() * steps["Gamma"] * scale }
	if rng.Float64() < prob { c.SigmaFloor += rng.NormFloat64() * steps["SigmaFloor"] * scale }
	quant.ClampChromosome(c)
	return c
}

func (e *GoldenCrossEvolvable) Crossover(p1, p2 Gene, rng *rand.Rand) Gene {
	a, ok1 := p1.(*quant.Chromosome)
	b, ok2 := p2.(*quant.Chromosome)
	if !ok1 || !ok2 { return p1 }
	child := &quant.Chromosome{}
	pick := func(v1, v2 float64) float64 { if rng.Float64() < 0.5 { return v1 }; return v2 }
	pickI := func(v1, v2 int) int { if rng.Float64() < 0.5 { return v1 }; return v2 }
	child.A, child.B, child.C = pick(a.A, b.A), pick(a.B, b.B), pick(a.C, b.C)
	child.Beta, child.Gamma = pick(a.Beta, b.Beta), pick(a.Gamma, b.Gamma)
	child.SigmaFloor = pick(a.SigmaFloor, b.SigmaFloor)
	child.MacroIntervalDays = pickI(a.MacroIntervalDays, b.MacroIntervalDays)
	child.MacroAccelThreshold = pick(a.MacroAccelThreshold, b.MacroAccelThreshold)
	child.MacroAccelMultiplier = pick(a.MacroAccelMultiplier, b.MacroAccelMultiplier)
	child.SoftReleaseMonths = pickI(a.SoftReleaseMonths, b.SoftReleaseMonths)
	child.MaxSoftReleasePct = pick(a.MaxSoftReleasePct, b.MaxSoftReleasePct)
	child.DeadHoldTarget = pick(a.DeadHoldTarget, b.DeadHoldTarget)
	child.MicroReservePct = pick(a.MicroReservePct, b.MicroReservePct)
	child.EMAShortBars = pickI(a.EMAShortBars, b.EMAShortBars)
	child.EMALongBars = pickI(a.EMALongBars, b.EMALongBars)
	quant.ClampChromosome(child)
	return child
}

func (e *GoldenCrossEvolvable) Fingerprint(gene Gene) uint64 {
	c, ok := gene.(*quant.Chromosome)
	if !ok { return 0 }
	h := fnv.New64a()
	writeFloat := func(v float64) {
		q := math.Round(v * 1e6)
		h.Write([]byte{byte(int64(q) >> 56), byte(int64(q) >> 48), byte(int64(q) >> 40),
			byte(int64(q) >> 32), byte(int64(q) >> 24), byte(int64(q) >> 16),
			byte(int64(q) >> 8), byte(int64(q))})
	}
	writeFloat(c.A); writeFloat(c.B); writeFloat(c.C)
	writeFloat(c.Beta); writeFloat(c.Gamma); writeFloat(c.SigmaFloor)
	writeFloat(float64(c.MacroIntervalDays)); writeFloat(c.MacroAccelThreshold)
	writeFloat(c.MacroAccelMultiplier); writeFloat(float64(c.SoftReleaseMonths))
	writeFloat(c.MaxSoftReleasePct); writeFloat(c.DeadHoldTarget)
	writeFloat(c.MicroReservePct); writeFloat(float64(c.EMAShortBars))
	writeFloat(float64(c.EMALongBars))
	return h.Sum64()
}

func (e *GoldenCrossEvolvable) Evaluate(ctx context.Context, gene Gene, plan *EvaluablePlan) EvaluationResult {
	c, ok := gene.(*quant.Chromosome)
	if !ok { return EvaluationResult{ScoreTotal: -99999} }

	totalScore := 0.0
	scores := make([]float64, len(plan.Windows))
	maxDD := 0.0
	roi := 0.0

	for i, w := range plan.Windows {
		score, dd, r := evaluateWindow(ctx, c, w, plan)
		scores[i] = score
		if dd > maxDD { maxDD = dd }
		roi = r
		if score <= -99998 { // fatal
			for j := i + 1; j < len(scores); j++ { scores[j] = -99999 }
			return EvaluationResult{ScoreTotal: -99999, Scores: scores, MaxDrawdown: maxDD, ROI: roi}
		}
		totalScore += score * w.Weight
	}
	return EvaluationResult{ScoreTotal: totalScore, Scores: scores, MaxDrawdown: maxDD, ROI: roi}
}

// evaluateWindow runs a single backtest window and returns (score, maxDD, roi).
func evaluateWindow(ctx context.Context, c *quant.Chromosome, w quant.CrucibleWindow, plan *EvaluablePlan) (float64, float64, float64) {
	spawn := plan.Spawn
	if spawn == nil { spawn = &quant.SpawnPoint{} }

	closes := quant.ExtractCloses(w.Bars)
	timestamps := quant.ExtractTimestamps(w.Bars)
	if len(closes) < 55 { return 0, 0, 0 }

	params := gc.Params{Chromosome: *c, SpawnPoint: *spawn}
	portfolio := quant.PortfolioSnapshot{
		CNYBalance: spawn.Policy.InitialCapital,
		TotalEquity: spawn.Policy.InitialCapital,
	}
	runtime := quant.RuntimeState{}
	currentPrice := closes[len(closes)-1]

	input := quant.StrategyInput{
		Closes: closes, Timestamps: timestamps, CurrentPrice: currentPrice,
		Portfolio: portfolio, Runtime: runtime,
	}
	_ = gc.Step(input, params) // placeholder — needs full backtest loop

	return 0, 0, 0
}

func (e *GoldenCrossEvolvable) DecodeElite(raw json.RawMessage) Gene {
	if len(raw) == 0 {
		c := quant.DefaultSeedChromosome
		return &c
	}
	var c quant.Chromosome
	if err := json.Unmarshal(raw, &c); err != nil {
		d := quant.DefaultSeedChromosome
		return &d
	}
	quant.ClampChromosome(&c)
	return &c
}

func (e *GoldenCrossEvolvable) EncodeResult(gene Gene, spawn *quant.SpawnPoint) json.RawMessage {
	c, ok := gene.(*quant.Chromosome)
	if !ok { return nil }
	payload := map[string]any{"chromosome": c}
	if spawn != nil { payload["spawn_point"] = spawn }
	data, _ := json.Marshal(payload)
	return data
}
