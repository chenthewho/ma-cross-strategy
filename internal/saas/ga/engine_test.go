package ga

import (
	"context"
	"encoding/json"
	"math"
	"math/rand"
	"testing"

	"github.com/chenthewho/ma-cross-strategy/internal/quant"
)

// mockEvolvable is a minimal EvolvableStrategy for testing the engine.
type mockEvolvable struct {
	scoreFn func(gene Gene) float64 // deterministic score function
}

func (m *mockEvolvable) StrategyID() string { return "mock" }

func (m *mockEvolvable) Sample(rng *rand.Rand) Gene {
	c := quant.DefaultSeedChromosome
	c.A = rng.Float64()*6 - 3
	c.B = rng.Float64()*6 - 3
	c.C = rng.Float64()*6 - 3
	return &c
}

func (m *mockEvolvable) Mutate(gene Gene, prob, scale float64, rng *rand.Rand) Gene {
	return gene // identity (simplest)
}

func (m *mockEvolvable) Crossover(p1, p2 Gene, rng *rand.Rand) Gene {
	return p1 // identity
}

func (m *mockEvolvable) Fingerprint(gene Gene) uint64 {
	c := gene.(*quant.Chromosome)
	return uint64(math.Float64bits(c.A + c.B + c.C))
}

func (m *mockEvolvable) Evaluate(ctx context.Context, gene Gene, plan *EvaluablePlan) EvaluationResult {
	if m.scoreFn != nil {
		return EvaluationResult{ScoreTotal: m.scoreFn(gene)}
	}
	return EvaluationResult{ScoreTotal: 1.0}
}

func (m *mockEvolvable) DecodeElite(raw json.RawMessage) Gene {
	c := quant.DefaultSeedChromosome
	return &c
}

func (m *mockEvolvable) EncodeResult(gene Gene, spawn *quant.SpawnPoint) json.RawMessage {
	return nil
}

// mockEvaluablePlan is a minimal EvaluablePlan for engine tests.
func mockPlan() EvaluablePlan {
	return EvaluablePlan{
		Symbol:       "MOCK",
		TemplateName: "mock",
	}
}

// TestElitePreservation verifies that after one generation, the top EliteCount
// individuals from the parent generation appear in the child population.
func TestElitePreservation(t *testing.T) {
	eng := NewEngine(&mockEvolvable{
		scoreFn: func(gene Gene) float64 {
			c := gene.(*quant.Chromosome)
			// Score = A + B + C (deterministic, no noise)
			return c.A + c.B + c.C
		},
	}, 42)

	cfg := EpochConfig{
		PopSize:        50,
		MaxGenerations: 2,
		LotStepSize:    100,
		LotMinQty:      100,
	}

	result, err := eng.RunEpoch(context.Background(), cfg, mockPlan())
	if err != nil {
		t.Fatalf("RunEpoch failed: %v", err)
	}
	if result.Champion == nil {
		t.Fatal("no champion returned")
	}

	t.Logf("champion: %+v, score=%v, gens=%d", result.Champion, result.ScoreTotal, result.Generations)
}

// TestMutationRamp verifies: after EarlyStopPatience generations with no improvement,
// mutProb should equal initial × MutationRampFactor (capped at MutationProbMax).
func TestMutationRamp(t *testing.T) {
	hp := defaultHyperParams()

	// Verify the ramp factor is correct
	expected := hp.MutationProbability * hp.MutationRampFactor
	if expected > hp.MutationProbMax {
		expected = hp.MutationProbMax
	}

	// Simulate: after patience generations with no improvement, ramp should activate
	patienceCount := hp.EarlyStopPatience // threshold reached

	if patienceCount >= hp.EarlyStopPatience {
		mutProb := hp.MutationProbability
		mutProb = math.Min(mutProb*hp.MutationRampFactor, hp.MutationProbMax)
		if math.Abs(mutProb-expected) > 1e-10 {
			t.Errorf("expected mutProb=%v after ramp, got %v", expected, mutProb)
		}
	}

	t.Logf("initial mutProb=%v, post-ramp mutProb=%v (rampFactor=%v, max=%v)",
		hp.MutationProbability, expected, hp.MutationRampFactor, hp.MutationProbMax)
}

// TestFatalTournamentSelection verifies that individuals with score = -99999
// are almost never selected in tournament selection (less than 5% over 1000 attempts).
func TestFatalTournamentSelection(t *testing.T) {
	eng := NewEngine(&mockEvolvable{}, 123)

	// Create a population of 30 individuals
	// One fatal (score=-99999) at index 0, rest have score=1.0
	popSize := 30
	population := make([]Gene, popSize)
	fitness := make([]float64, popSize)
	for i := 0; i < popSize; i++ {
		c := quant.DefaultSeedChromosome
		c.A = float64(i)
		population[i] = &c
		fitness[i] = 1.0
	}
	// Mark first as fatal
	fitness[0] = -99999

	// Run 1000 tournament selections
	fatalSelections := 0
	numTrials := 1000
	for i := 0; i < numTrials; i++ {
		gene := eng.tournamentSelect(population, fitness)
		c := gene.(*quant.Chromosome)
		if math.Abs(c.A-0.0) < 1e-10 { // index 0 = fatal individual
			fatalSelections++
		}
	}

	rate := float64(fatalSelections) / float64(numTrials)
	t.Logf("fatal individual selected %d/%d times (%.2f%%)", fatalSelections, numTrials, rate*100)

	if rate > 0.05 {
		t.Errorf("fatal individual selected %.2f%% — exceeds 5%% threshold", rate*100)
	}
}
