package ga

import (
	"context"
	"encoding/json"
	"math"
	"math/rand"
	"runtime"
	"sort"
	"sync"

	"github.com/chenthewho/ma-cross-strategy/internal/quant"
)

// ─── Engine Configuration ──────────────────────────────────

// EpochConfig holds the parameters for a single evolution run.
type EpochConfig struct {
	PopSize         int
	MaxGenerations  int
	LotStepSize     float64
	LotMinQty       float64
	OnProgress      func(generation int, bestScore float64, mutProb, mutScale float64)
	SpawnPointOverride *json.RawMessage // if non-nil, override champion/default spawn
}

// DefaultEpochConfig returns sensible defaults.
func DefaultEpochConfig() EpochConfig {
	return EpochConfig{
		PopSize:        300,
		MaxGenerations: 25,
		LotStepSize:    100,
		LotMinQty:      100,
	}
}

// EpochResult holds the final output of an evolution run.
type EpochResult struct {
	Champion   Gene
	ScoreTotal float64
	Scores     []float64
	MaxDD      float64
	Generations int
	EarlyStop  bool
	GenerationsRun int // alias
}

// ─── Engine Hyperparameters ────────────────────────────────

type evoHyperParams struct {
	EliteCount           int
	MutationProbability  float64
	MutationScale        float64
	MutationProbMax      float64
	MutationScaleMax     float64
	MutationRampFactor   float64
	EarlyStopPatience    int
	EarlyStopMinDelta    float64
	TournamentSize       int
}

func defaultHyperParams() evoHyperParams {
	return evoHyperParams{
		EliteCount:          8,
		MutationProbability: 0.15,
		MutationScale:       1.0,
		MutationProbMax:     0.55,
		MutationScaleMax:    3.0,
		MutationRampFactor:  1.25,
		EarlyStopPatience:   5,
		EarlyStopMinDelta:   0.001,
		TournamentSize:      3,
	}
}

// ─── Engine ────────────────────────────────────────────────

// EvolutionEngine drives the GA lifecycle.
type EvolutionEngine struct {
	evolvable EvolvableStrategy
	rng       *rand.Rand
	hp        evoHyperParams
}

// NewEngine creates a new evolution engine for the given strategy.
func NewEngine(evolvable EvolvableStrategy, seed int64) *EvolutionEngine {
	return &EvolutionEngine{
		evolvable: evolvable,
		rng:       rand.New(rand.NewSource(seed)),
		hp:        defaultHyperParams(),
	}
}

// RunEpoch executes a full evolution epoch.
func (eng *EvolutionEngine) RunEpoch(ctx context.Context, cfg EpochConfig, plan EvaluablePlan) (EpochResult, error) {
	hp := eng.hp

	// ── Step 1: Population initialization ──
	population := make([]Gene, cfg.PopSize)

	// Try to load elite genes from DB (stored in plan context or passed via cfg)
	// For now, index 0 = default seed, rest random
	population[0] = eng.evolvable.DecodeElite(nil) // default seed

	for i := 1; i < cfg.PopSize; i++ {
		population[i] = eng.evolvable.Sample(eng.rng)
	}

	// ── Step 2: Initial evaluation ──
	fitness := eng.evaluatePopulation(ctx, population, plan)
	bestFitness := fitness[0]
	for _, f := range fitness {
		if f > bestFitness {
			bestFitness = f
		}
	}

	// ── Step 3: Evolution loop ──
	patienceCount := 0
	generation := 0
	earlyStop := false
	mutProb := hp.MutationProbability
	mutScale := hp.MutationScale

	for generation = 0; generation < cfg.MaxGenerations; generation++ {
		// Sort population by fitness (descending)
		type idxFit struct {
			idx int
			f   float64
		}
		sorted := make([]idxFit, len(population))
		for i, f := range fitness {
			sorted[i] = idxFit{i, f}
		}
		sort.Slice(sorted, func(i, j int) bool { return sorted[i].f > sorted[j].f })

		currentBest := sorted[0].f

		// Convergence detection
		if math.Abs(currentBest-bestFitness) < hp.EarlyStopMinDelta {
			patienceCount++
		} else {
			bestFitness = currentBest
			patienceCount = 0
		}

		// Mutation ramp
		if patienceCount >= hp.EarlyStopPatience {
			if mutProb < hp.MutationProbMax || mutScale < hp.MutationScaleMax {
				mutProb = math.Min(mutProb*hp.MutationRampFactor, hp.MutationProbMax)
				mutScale = math.Min(mutScale*hp.MutationRampFactor, hp.MutationScaleMax)
			} else {
				// Both at max and still no improvement → early stop
				earlyStop = true
				break
			}
		}

		// Progress callback
		if cfg.OnProgress != nil {
			cfg.OnProgress(generation, currentBest, mutProb, mutScale)
		}

		// ── Produce next generation ──
		nextGen := make([]Gene, cfg.PopSize)

		// Elitism: top EliteCount go straight to next generation
		for i := 0; i < hp.EliteCount && i < len(sorted); i++ {
			nextGen[i] = population[sorted[i].idx]
		}

		// Fill rest with crossover + mutation
		for i := hp.EliteCount; i < cfg.PopSize; i++ {
			p1 := eng.tournamentSelect(population, fitness)
			p2 := eng.tournamentSelect(population, fitness)
			child := eng.evolvable.Crossover(p1, p2, eng.rng)
			child = eng.evolvable.Mutate(child, mutProb, mutScale, eng.rng)
			nextGen[i] = child
		}

		population = nextGen

		// Evaluate new population
		fitness = eng.evaluatePopulation(ctx, population, plan)
	}

	// ── Step 4: Return champion ──
	// Find best
	finalBest := fitness[0]
	finalIdx := 0
	for i, f := range fitness {
		if f > finalBest {
			finalBest = f
			finalIdx = i
		}
	}

	champion := population[finalIdx]
	result := eng.evolvable.Evaluate(ctx, champion, &plan)

	return EpochResult{
		Champion:    champion,
		ScoreTotal:  result.ScoreTotal,
		Scores:      result.Scores,
		MaxDD:       result.MaxDrawdown,
		Generations: generation + 1,
		EarlyStop:   earlyStop,
	}, nil
}

// evaluatePopulation evaluates the entire population concurrently with fingerprint caching.
func (eng *EvolutionEngine) evaluatePopulation(ctx context.Context, pop []Gene, plan EvaluablePlan) []float64 {
	n := len(pop)
	fitness := make([]float64, n)
	workers := min(runtime.NumCPU(), n)

	var cache sync.Map // fingerprint -> fitness

	type task struct {
		idx  int
		gene Gene
	}
	tasks := make(chan task, n)
	for i, g := range pop {
		tasks <- task{i, g}
	}
	close(tasks)

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range tasks {
				select {
				case <-ctx.Done():
					fitness[t.idx] = -99999
					continue
				default:
				}
				fp := eng.evolvable.Fingerprint(t.gene)
				if cached, ok := cache.Load(fp); ok {
					fitness[t.idx] = cached.(float64)
					continue
				}
				result := eng.evolvable.Evaluate(ctx, t.gene, &plan)
				fitness[t.idx] = result.ScoreTotal
				cache.Store(fp, result.ScoreTotal)
			}
		}()
	}
	wg.Wait()

	return fitness
}

// tournamentSelect picks the fittest among TournamentSize randomly chosen individuals.
func (eng *EvolutionEngine) tournamentSelect(pop []Gene, fitness []float64) Gene {
	bestIdx := -1
	bestFit := math.Inf(-1)
	for k := 0; k < eng.hp.TournamentSize; k++ {
		idx := eng.rng.Intn(len(pop))
		if fitness[idx] > bestFit {
			bestFit = fitness[idx]
			bestIdx = idx
		}
	}
	return pop[bestIdx]
}

// EncodeResult delegates to the evolvable strategy's EncodeResult.
func (eng *EvolutionEngine) EncodeResult(gene Gene, spawn json.RawMessage) json.RawMessage {
	var sp *quant.SpawnPoint
	if len(spawn) > 0 {
		var s quant.SpawnPoint
		if json.Unmarshal(spawn, &s) == nil {
			sp = &s
		}
	}
	return eng.evolvable.EncodeResult(gene, sp)
}
