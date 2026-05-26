package quant

// ─── K-Line & Data Extraction ───────────────────────────────

// Bar represents a single candlestick (K-line).
type Bar struct {
	OpenTime int64   // ms timestamp
	Open     float64
	High     float64
	Low      float64
	Close    float64
	Volume   float64
}

// ─── Portfolio ──────────────────────────────────────────────

// PortfolioSnapshot is an account snapshot.
type PortfolioSnapshot struct {
	CNYBalance     float64   // CNY cash balance
	DeadHold       float64   // macro bottom position (shares)
	FloatHold      float64   // micro floating position (shares)
	ColdSealedHold float64   // permanently locked position (shares)
	TotalEquity    float64   // total equity in CNY
	DeadHoldLots   []SpotLot // bottom-position lot details
}

// SpotLot tracks an individual position lot with cost basis.
type SpotLot struct {
	LotType      LotType
	Amount       float64 // shares/units
	CostPrice    float64 // average cost per unit
	CreatedAt    int64   // creation timestamp (ms)
	IsColdSealed bool    // never releasable if true
}

// LotType represents the three-state position classification.
type LotType string

const (
	LotDeadStack  LotType = "DEAD_STACK"
	LotFloating   LotType = "FLOATING"
	LotColdSealed LotType = "COLD_SEALED"
)

// ─── Runtime State ──────────────────────────────────────────

// RuntimeState persists strategy internal state across ticks.
type RuntimeState struct {
	LastProcessedBar   int64   // timestamp of last processed bar (ms)
	LastMacroAction    int64   // timestamp of last macro operation (ms)
	AccumulatedReserve float64 // accumulated reserved funds
}

// ─── AI Signal (optional) ───────────────────────────────────

// AISignalVector carries optional AI-generated signals.
// During backtesting, all values are 0.
type AISignalVector struct {
	SentimentScore float64 // normalized -1 to +1
	NewsImpact     float64 // normalized -1 to +1
	MacroBias      float64 // normalized -1 to +1
}

// ─── Strategy Input / Output ────────────────────────────────

// StrategyInput is the complete input snapshot fed to Step().
type StrategyInput struct {
	Closes         []float64         // close price series (ACL-degraded, no OHLCV)
	Timestamps     []int64           // timestamp series (ms)
	CurrentPrice   float64           // latest price
	Portfolio      PortfolioSnapshot // account snapshot
	Runtime        RuntimeState      // strategy runtime state
	AISignalVector AISignalVector    // AI signals (optional; [0,0,0] in backtest)
}

// StrategyOutput is the complete output from Step().
type StrategyOutput struct {
	MacroIntent   *MacroIntent
	MicroIntent   *MicroIntent
	ReleaseIntent *ReleaseIntent
	MarketState   MarketState
	NewRuntime    RuntimeState
}

// MacroIntent represents a macro-engine buy decision.
type MacroIntent struct {
	Action    string  // always "BUY"
	AmountCNY float64 // buy amount in CNY
	Engine    string  // "MACRO"
	LotType   string  // "DEAD_STACK"
}

// MicroIntent represents a micro-engine buy/sell decision.
type MicroIntent struct {
	Action    string  // "BUY" | "SELL"
	AmountCNY float64 // positive = buy, negative = sell
	Engine    string  // "MICRO"
	LotType   string  // "FLOATING"
}

// ReleaseIntent represents a dead-hold release decision.
type ReleaseIntent struct {
	ReleaseType   string  // "SOFT" | "HARD"
	ReleaseAmount float64 // shares to release
	Reason        string  // audit log description
}

// ─── Market State ───────────────────────────────────────────

// MarketState describes the current market regime.
type MarketState struct {
	State                  string  // "quiet" | "bull" | "bear" | "panic"
	TimeDilationMultiplier float64 // scale factor for macro engine time windows
	BetaMultiplier         float64 // scale factor for micro Sigmoid aggressiveness
	IsQuiet                bool    // true = suppress micro dust orders
}

// Market state constants.
const (
	MarketQuiet = "quiet"
	MarketBull  = "bull"
	MarketBear  = "bear"
	MarketPanic = "panic"
)

// ─── Macro Engine Types ─────────────────────────────────────

// MacroDecisionInput carries all inputs for the macro engine.
type MacroDecisionInput struct {
	TotalEquity          float64
	SpendableCNY         float64
	CurrentPrice         float64
	DeadHold             float64 // current dead hold (shares)
	DeadHoldValue        float64 // current dead hold (CNY)
	MonthlyInject        float64 // monthly injection amount (from SpawnPoint)
	TimeDilation         float64 // from market state
	PriceDeviation       float64 // (Price - EMA_long) / EMA_long
	DaysSinceLastMacro   int     // days since last macro action
	MacroIntervalDays    int     // base DCA interval days (from chromosome)
	MacroAccelThreshold  float64 // price deviation trigger threshold
	MacroAccelMultiplier float64 // acceleration multiplier
}

// ─── Micro Engine Types ─────────────────────────────────────

// MicroDecisionInput carries all inputs for the micro engine.
type MicroDecisionInput struct {
	Closes             []float64
	CurrentPrice       float64
	TotalEquity        float64
	CurrentMicroWeight float64
	IsQuiet            bool
	BetaMultiplier     float64
	// Chromosome parameters
	A          float64 // X1 coefficient
	B          float64 // X2 coefficient
	C          float64 // X3 coefficient
	Beta       float64 // Sigmoid aggressiveness
	Gamma      float64 // inventory bias coefficient
	SigmaFloor float64 // volatility floor
	EMAShort   int     // short EMA bars
	EMALong    int     // long EMA bars
}

// MicroDecisionOutput carries the micro engine's output.
type MicroDecisionOutput struct {
	TargetWeight    float64 // target portfolio weight ∈ [0,1]
	Signal          float64 // raw signal for debugging
	TheoreticalCNY  float64 // theoretical order size before filtering
	OrderCNY        float64 // actual order size after wedge filtering
	VolatilityRatio float64 // volatility ratio for debugging
}
