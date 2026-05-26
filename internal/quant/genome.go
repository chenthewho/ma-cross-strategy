package quant

// ─── Hard Bounds ────────────────────────────────────────────────────────────

// HardBounds defines the absolute minima and maxima for every chromosome field.
// No chromosome may ever exceed these bounds after clamping.
var HardBounds = struct {
	AMin, AMax                   float64
	BMin, BMax                   float64
	CMin, CMax                   float64
	BetaMin, BetaMax             float64
	GammaMin, GammaMax           float64
	SigmaFloorMin, SigmaFloorMax float64
	MacroIntervalDaysMin, MacroIntervalDaysMax       int
	MacroAccelThresholdMin, MacroAccelThresholdMax    float64
	MacroAccelMultiplierMin, MacroAccelMultiplierMax  float64
	SoftReleaseMonthsMin, SoftReleaseMonthsMax        int
	MaxSoftReleasePctMin, MaxSoftReleasePctMax        float64
	DeadHoldTargetMin, DeadHoldTargetMax              float64
	MicroReservePctMin, MicroReservePctMax            float64
	EMAShortBarsMin, EMAShortBarsMax                 int
	EMALongBarsMin, EMALongBarsMax                   int
}{
	AMin: -3, AMax: 3,
	BMin: -3, BMax: 3,
	CMin: -3, CMax: 3,
	BetaMin: 0.1, BetaMax: 5.0,
	GammaMin: 0, GammaMax: 2.0,
	SigmaFloorMin: 0.001, SigmaFloorMax: 0.05,
	MacroIntervalDaysMin: 7, MacroIntervalDaysMax: 90,
	MacroAccelThresholdMin: 0.01, MacroAccelThresholdMax: 0.20,
	MacroAccelMultiplierMin: 1.0, MacroAccelMultiplierMax: 5.0,
	SoftReleaseMonthsMin: 1, SoftReleaseMonthsMax: 24,
	MaxSoftReleasePctMin: 0.05, MaxSoftReleasePctMax: 0.50,
	DeadHoldTargetMin: 0.10, DeadHoldTargetMax: 0.90,
	MicroReservePctMin: 0.05, MicroReservePctMax: 0.50,
	EMAShortBarsMin: 5, EMAShortBarsMax: 55,
	EMALongBarsMin: 20, EMALongBarsMax: 200,
}

// ─── Chromosome ─────────────────────────────────────────────────────────────

// Chromosome encodes the full set of evolvable parameters for the golden_cross
// strategy. Every field is bounded by HardBounds and additionally constrained
// by structural rules enforced by ClampChromosome.
type Chromosome struct {
	A  float64 `json:"a"`
	B  float64 `json:"b"`
	C  float64 `json:"c"`
	Beta       float64 `json:"beta"`
	Gamma      float64 `json:"gamma"`
	SigmaFloor float64 `json:"sigma_floor"`

	MacroIntervalDays    int     `json:"macro_interval_days"`
	MacroAccelThreshold  float64 `json:"macro_accel_threshold"`
	MacroAccelMultiplier float64 `json:"macro_accel_multiplier"`

	SoftReleaseMonths int     `json:"soft_release_months"`
	MaxSoftReleasePct float64 `json:"max_soft_release_pct"`

	DeadHoldTarget  float64 `json:"dead_hold_target"`
	MicroReservePct float64 `json:"micro_reserve_pct"`

	EMAShortBars int `json:"ema_short_bars"`
	EMALongBars  int `json:"ema_long_bars"`
}

// ClampChromosome first clips every field to its HardBounds range, then
// enforces structural constraints that must hold between related fields.
func ClampChromosome(c *Chromosome) {
	// Per-field hard bounds.
	c.A = ClipFloat64(c.A, HardBounds.AMin, HardBounds.AMax)
	c.B = ClipFloat64(c.B, HardBounds.BMin, HardBounds.BMax)
	c.C = ClipFloat64(c.C, HardBounds.CMin, HardBounds.CMax)
	c.Beta = ClipFloat64(c.Beta, HardBounds.BetaMin, HardBounds.BetaMax)
	c.Gamma = ClipFloat64(c.Gamma, HardBounds.GammaMin, HardBounds.GammaMax)
	c.SigmaFloor = ClipFloat64(c.SigmaFloor, HardBounds.SigmaFloorMin, HardBounds.SigmaFloorMax)

	c.MacroIntervalDays = clampInt(c.MacroIntervalDays, HardBounds.MacroIntervalDaysMin, HardBounds.MacroIntervalDaysMax)
	c.MacroAccelThreshold = ClipFloat64(c.MacroAccelThreshold, HardBounds.MacroAccelThresholdMin, HardBounds.MacroAccelThresholdMax)
	c.MacroAccelMultiplier = ClipFloat64(c.MacroAccelMultiplier, HardBounds.MacroAccelMultiplierMin, HardBounds.MacroAccelMultiplierMax)

	c.SoftReleaseMonths = clampInt(c.SoftReleaseMonths, HardBounds.SoftReleaseMonthsMin, HardBounds.SoftReleaseMonthsMax)
	c.MaxSoftReleasePct = ClipFloat64(c.MaxSoftReleasePct, HardBounds.MaxSoftReleasePctMin, HardBounds.MaxSoftReleasePctMax)

	c.DeadHoldTarget = ClipFloat64(c.DeadHoldTarget, HardBounds.DeadHoldTargetMin, HardBounds.DeadHoldTargetMax)
	c.MicroReservePct = ClipFloat64(c.MicroReservePct, HardBounds.MicroReservePctMin, HardBounds.MicroReservePctMax)

	c.EMAShortBars = clampInt(c.EMAShortBars, HardBounds.EMAShortBarsMin, HardBounds.EMAShortBarsMax)
	c.EMALongBars = clampInt(c.EMALongBars, HardBounds.EMALongBarsMin, HardBounds.EMALongBarsMax)

	// ── Structural constraints ──────────────────────────────────────────

	// EMA period ordering: short must be strictly less than long.
	if c.EMAShortBars >= c.EMALongBars {
		c.EMAShortBars = c.EMALongBars - 1
		if c.EMAShortBars < HardBounds.EMAShortBarsMin {
			c.EMAShortBars = HardBounds.EMAShortBarsMin
			c.EMALongBars = c.EMAShortBars + 1
		}
	}

	// Combined reserve must not exceed 95% of total equity.
	if c.DeadHoldTarget+c.MicroReservePct > 0.95 {
		// Scale both proportionally down to 0.95.
		scale := 0.95 / (c.DeadHoldTarget + c.MicroReservePct)
		c.DeadHoldTarget *= scale
		c.MicroReservePct *= scale
	}
}

// clampInt is a helper that clamps an integer to [lo, hi].
func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// ─── Default Seed ───────────────────────────────────────────────────────────

// DefaultSeedChromosome is the reference starting point for evolution.
// Values are taken from the golden_cross strategy specification.
var DefaultSeedChromosome = Chromosome{
	A: 0.0, B: 0.3, C: 0.1,
	Beta: 1.0, Gamma: 0.3,
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

// ─── Spawn Point ────────────────────────────────────────────────────────────

// CapitalPolicy defines the capital injection rules for a strategy epoch.
// These values are frozen for the duration of an epoch and are not evolved.
type CapitalPolicy struct {
	InitialCapital float64 `json:"initial_capital"`
	MonthlyInject  float64 `json:"monthly_inject"`
	DeadLineMonths float64 `json:"dead_line_months"`
}

// RiskBounds defines the risk-control parameters for a strategy epoch.
// These values are frozen for the duration of an epoch and are not evolved.
type RiskBounds struct {
	FeeRate        float64 `json:"fee_rate"`
	Slippage       float64 `json:"slippage"`
	GlobalStopLoss float64 `json:"global_stop_loss"`
	LotStep        int     `json:"lot_step"`
	LotMin         int     `json:"lot_min"`
}

// SpawnPoint bundles the capital policy and risk bounds that define the
// starting conditions for a strategy epoch. These are epoch-level frozen
// parameters — they do not participate in the genome.
type SpawnPoint struct {
	Policy CapitalPolicy `json:"policy"`
	Risk   RiskBounds    `json:"risk"`
}

// ─── Gene Step Sizes ────────────────────────────────────────────────────────

// GeneStep returns the mutation step size for each chromosome field.
// These values control how much a gene can shift per mutation event.
func GeneStep() map[string]float64 {
	return map[string]float64{
		"a":                      0.05,
		"b":                      0.05,
		"c":                      0.05,
		"beta":                   0.10,
		"gamma":                  0.05,
		"sigma_floor":            0.001,
		"macro_interval_days":    1.0,
		"macro_accel_threshold":  0.005,
		"macro_accel_multiplier": 0.10,
		"soft_release_months":    1.0,
		"max_soft_release_pct":   0.01,
		"dead_hold_target":       0.01,
		"micro_reserve_pct":      0.01,
		"ema_short_bars":         1.0,
		"ema_long_bars":          2.0,
	}
}
