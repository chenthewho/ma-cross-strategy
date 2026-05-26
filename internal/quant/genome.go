package quant

import (
	"encoding/json"
	"math"
)

// ─── Chromosome — Evolvable Strategy Parameters ─────────────
//
// 15 fields optimized by the GA engine.
// These participate in crossover and mutation during evolution.

// Chromosome holds all evolvable strategy parameters.
type Chromosome struct {
	// Signal coefficients (three-factor linear synthesis)
	A float64 `json:"a"` // Price deviation weight (X1), range [-3, 3]
	B float64 `json:"b"` // EMA divergence weight (X2), range [-3, 3]
	C float64 `json:"c"` // Momentum weight (X3), range [-3, 3]

	// Sigmoid engine parameters
	Beta       float64 `json:"beta"`        // Sigmoid aggressiveness, range [0.1, 5.0]
	Gamma      float64 `json:"gamma"`       // Inventory bias coefficient, range [0, 2.0]
	SigmaFloor float64 `json:"sigma_floor"` // Volatility floor (prevents division by zero), range [0.001, 0.05]

	// Macro engine parameters
	MacroIntervalDays    int     `json:"macro_interval_days"`    // Base DCA interval, range [7, 90]
	MacroAccelThreshold  float64 `json:"macro_accel_threshold"`  // Price deviation acceleration trigger, range [0.01, 0.20]
	MacroAccelMultiplier float64 `json:"macro_accel_multiplier"` // Acceleration multiplier, range [1.0, 5.0]

	// Dead hold release parameters
	SoftReleaseMonths int     `json:"soft_release_months"` // Aging months for soft release, range [1, 24]
	MaxSoftReleasePct float64 `json:"max_soft_release_pct"` // Max soft release fraction, range [0.05, 0.50]
	DeadHoldTarget    float64 `json:"dead_hold_target"`     // Target dead hold ratio, range [0.10, 0.90]

	// Reserve & indicator parameters
	MicroReservePct float64 `json:"micro_reserve_pct"` // Micro reserve fraction, range [0.05, 0.50]
	EMAShortBars    int     `json:"ema_short_bars"`    // Short EMA window, range [5, 55]
	EMALongBars     int     `json:"ema_long_bars"`     // Long EMA window, range [20, 200]
}

// HardBounds defines the legal value range for each chromosome field.
// [min, max] inclusive.
var HardBounds = map[string][2]float64{
	"a":                      {-3, 3},
	"b":                      {-3, 3},
	"c":                      {-3, 3},
	"beta":                   {0.1, 5.0},
	"gamma":                  {0, 2.0},
	"sigma_floor":            {0.001, 0.05},
	"macro_interval_days":    {7, 90},
	"macro_accel_threshold":  {0.01, 0.20},
	"macro_accel_multiplier": {1.0, 5.0},
	"soft_release_months":    {1, 24},
	"max_soft_release_pct":   {0.05, 0.50},
	"dead_hold_target":       {0.10, 0.90},
	"micro_reserve_pct":      {0.05, 0.50},
	"ema_short_bars":         {5, 55},
	"ema_long_bars":          {20, 200},
}

// DefaultSeedChromosome is the product default champion seed.
// Used as GA cold-start individual and JSON decode fallback.
var DefaultSeedChromosome = Chromosome{
	A: 0.0, B: 0.3, C: 0.1,
	Beta:       1.0,
	Gamma:      0.3,
	SigmaFloor: 0.005,
	MacroIntervalDays:    30,
	MacroAccelThreshold:  0.05,
	MacroAccelMultiplier: 2.0,
	SoftReleaseMonths:    6,
	MaxSoftReleasePct:    0.30,
	DeadHoldTarget:       0.50,
	MicroReservePct:      0.25,
	EMAShortBars:         21,
	EMALongBars:          55,
}

// ClampChromosome clamps all fields to HardBounds and repairs structural constraints.
// MUST be called after mutation and crossover.
func ClampChromosome(c *Chromosome) {
	// Clamp float64 fields
	c.A = clipToBounds("a", c.A)
	c.B = clipToBounds("b", c.B)
	c.C = clipToBounds("c", c.C)
	c.Beta = clipToBounds("beta", c.Beta)
	c.Gamma = clipToBounds("gamma", c.Gamma)
	c.SigmaFloor = clipToBounds("sigma_floor", c.SigmaFloor)
	c.MacroAccelThreshold = clipToBounds("macro_accel_threshold", c.MacroAccelThreshold)
	c.MacroAccelMultiplier = clipToBounds("macro_accel_multiplier", c.MacroAccelMultiplier)
	c.MaxSoftReleasePct = clipToBounds("max_soft_release_pct", c.MaxSoftReleasePct)
	c.DeadHoldTarget = clipToBounds("dead_hold_target", c.DeadHoldTarget)
	c.MicroReservePct = clipToBounds("micro_reserve_pct", c.MicroReservePct)

	// Clamp int fields
	c.MacroIntervalDays = intClipToBounds("macro_interval_days", c.MacroIntervalDays)
	c.SoftReleaseMonths = intClipToBounds("soft_release_months", c.SoftReleaseMonths)
	c.EMAShortBars = intClipToBounds("ema_short_bars", c.EMAShortBars)
	c.EMALongBars = intClipToBounds("ema_long_bars", c.EMALongBars)

	// ── Structural constraints ──

	// EMA order: short < long
	if c.EMAShortBars >= c.EMALongBars {
		c.EMALongBars = c.EMAShortBars + 1
		if c.EMALongBars > 200 {
			c.EMAShortBars = 199
			c.EMALongBars = 200
		}
	}

	// DCA interval minimum
	if c.MacroIntervalDays < 7 {
		c.MacroIntervalDays = 7
	}

	// Soft release aging at least 1 month
	if c.SoftReleaseMonths < 1 {
		c.SoftReleaseMonths = 1
	}

	// Avoid over-reservation: dead_hold_target + micro_reserve_pct <= 0.95
	if c.DeadHoldTarget+c.MicroReservePct > 0.95 {
		excess := c.DeadHoldTarget + c.MicroReservePct - 0.95
		c.MicroReservePct = math.Max(0.05, c.MicroReservePct-excess/2)
		c.DeadHoldTarget = math.Max(0.10, c.DeadHoldTarget-excess/2)
	}
}

func clipToBounds(name string, v float64) float64 {
	b, ok := HardBounds[name]
	if !ok {
		return v
	}
	return ClipFloat64(v, b[0], b[1])
}

func intClipToBounds(name string, v int) int {
	b, ok := HardBounds[name]
	if !ok {
		return v
	}
	if float64(v) < b[0] {
		return int(b[0])
	}
	if float64(v) > b[1] {
		return int(b[1])
	}
	return v
}

// ─── SpawnPoint — Epoch-level frozen parameters ─────────────

// SpawnPoint contains capital policy and risk parameters.
// These are frozen at Epoch start and shared across the entire population.
// They do NOT participate in crossover or mutation.
type SpawnPoint struct {
	Policy CapitalPolicy `json:"policy"`
	Risk   RiskBounds    `json:"risk"`
}

// CapitalPolicy defines capital allocation rules.
type CapitalPolicy struct {
	InitialCapital float64 `json:"initial_capital"` // Initial capital in CNY, default 100000
	MonthlyInject  float64 `json:"monthly_inject"`  // Monthly injection in CNY, default 5000
	DeadLineMonths int     `json:"deadline_months"` // Max months before forced DCA, default 24
}

// RiskBounds defines risk management parameters.
type RiskBounds struct {
	FeeRate        float64 `json:"fee_rate"`         // Trading fee rate, default 0.0003 (0.03%)
	Slippage       float64 `json:"slippage"`         // Slippage, default 0.0001 (0.01%)
	GlobalStopLoss float64 `json:"global_stop_loss"` // Global stop loss threshold, default 0.30 (30%)
	LotStep        float64 `json:"lot_step"`         // Minimum trading unit (A-share: 100), default 100
	LotMin         float64 `json:"lot_min"`          // Minimum trading quantity, default 100
}

// DefaultSpawnPoint returns the default capital policy and risk bounds.
var DefaultSpawnPoint = SpawnPoint{
	Policy: CapitalPolicy{
		InitialCapital: 100000,
		MonthlyInject:  5000,
		DeadLineMonths: 24,
	},
	Risk: RiskBounds{
		FeeRate:        0.0003,
		Slippage:       0.0001,
		GlobalStopLoss: 0.30,
		LotStep:        100,
		LotMin:         100,
	},
}

// ─── ParamPack ──────────────────────────────────────────────

// ParamPack wraps Chromosome + SpawnPoint for JSON serialization.
type ParamPack struct {
	StrategyID string    `json:"strategy_id"`
	Chromosome Chromosome `json:"chromosome"`
	SpawnPoint SpawnPoint `json:"spawn_point"`
}

// EncodeParamPack serializes a ParamPack to JSON.
func EncodeParamPack(pp ParamPack) ([]byte, error) {
	return json.Marshal(pp)
}

// DecodeParamPack deserializes a ParamPack from JSON.
// Returns DefaultSeedChromosome + DefaultSpawnPoint if raw is empty or invalid.
func DecodeParamPack(raw []byte) ParamPack {
	if len(raw) == 0 {
		return ParamPack{
			StrategyID: "golden_cross",
			Chromosome: DefaultSeedChromosome,
			SpawnPoint: DefaultSpawnPoint,
		}
	}
	var pp ParamPack
	if err := json.Unmarshal(raw, &pp); err != nil {
		return ParamPack{
			StrategyID: "golden_cross",
			Chromosome: DefaultSeedChromosome,
			SpawnPoint: DefaultSpawnPoint,
		}
	}
	if pp.StrategyID == "" {
		pp.StrategyID = "golden_cross"
	}
	return pp
}

// GeneStep returns per-field mutation step sizes for Gaussian mutation.
func GeneStep() map[string]float64 {
	return map[string]float64{
		"A": 0.1, "B": 0.1, "C": 0.1,
		"Beta": 0.1, "Gamma": 0.05,
		"SigmaFloor": 0.001,
		"MacroAccelThreshold": 0.01,
		"MacroAccelMultiplier": 0.1,
		"MaxSoftReleasePct": 0.05,
		"DeadHoldTarget": 0.05,
		"MicroReservePct": 0.05,
	}
}
