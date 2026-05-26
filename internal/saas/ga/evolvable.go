// Package ga provides the genetic algorithm evolution engine.
package ga

import (
	"context"
	"encoding/json"
	"math/rand"

	"github.com/chenthewho/ma-cross-strategy/internal/quant"
)

// Gene is an opaque chromosome — the engine never reads its internal fields.
type Gene = any

// DCABaseline holds pre-computed Ghost DCA results for one crucible window.
type DCABaseline struct {
	FinalEquity   float64
	TotalInjected float64
	MaxDrawdown   float64
	ROI           float64
}

// EvaluablePlan is immutable evaluation context built at Epoch start,
// shared by all Evaluate calls within the epoch.
type EvaluablePlan struct {
	Symbol       string
	TemplateName string
	Spawn        *quant.SpawnPoint
	LotStep      float64
	LotMin       float64
	Windows      []quant.CrucibleWindow
	DCABaselines []DCABaseline // one per window, pre-computed
}

// EvaluationResult holds the fitness score and per-window breakdown.
type EvaluationResult struct {
	ScoreTotal  float64
	Scores      []float64 // per-window scores
	MaxDrawdown float64
	ROI         float64
}

// EvolvableStrategy is the 8-verb contract between the engine and a specific strategy.
// The engine drives population lifecycle through this interface without ever
// inspecting chromosome internals.
type EvolvableStrategy interface {
	// StrategyID returns the unique strategy identifier (e.g. "golden_cross").
	StrategyID() string

	// Sample generates a random gene within legal bounds.
	Sample(rng *rand.Rand) Gene

	// Mutate applies additive Gaussian mutation to a gene.
	Mutate(gene Gene, prob, scale float64, rng *rand.Rand) Gene

	// Crossover performs uniform crossover between two parents.
	Crossover(p1, p2 Gene, rng *rand.Rand) Gene

	// Fingerprint returns a unique hash for the gene (FNV-1a-64, 1e-6 precision).
	Fingerprint(gene Gene) uint64

	// Evaluate assesses a gene on the given crucible plan, returning fitness
	// and per-window breakdown. Should cascade-short-circuit (fatal=-99999).
	Evaluate(ctx context.Context, gene Gene, plan *EvaluablePlan) EvaluationResult

	// DecodeElite decodes an elite gene from DB JSON. nil/empty returns default seed.
	DecodeElite(raw json.RawMessage) Gene

	// EncodeResult serializes a champion gene + spawn point to JSON for DB storage.
	EncodeResult(gene Gene, spawn *quant.SpawnPoint) json.RawMessage
}
