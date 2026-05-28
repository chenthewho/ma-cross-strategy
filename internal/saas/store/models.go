// Package store provides database models and access for QuantSaaS.
// All database schema is managed via GORM AutoMigrate — no SQL files.
package store

import (
	"time"

	"gorm.io/gorm"
)

// User represents a registered user with subscription plan.
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	Role         string    `gorm:"default:'user'" json:"role"` // "user" | "admin"
	Plan         string    `gorm:"default:'free'" json:"plan"` // "free" | "pro"
	MaxInstances int       `gorm:"default:3" json:"max_instances"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// StrategyTemplate is the master strategy definition (blueprint).
type StrategyTemplate struct {
	ID          string    `gorm:"primaryKey" json:"id"` // e.g. "golden_cross"
	Name        string    `gorm:"not null" json:"name"`
	Version     string    `gorm:"not null" json:"version"`
	IsSpot      bool      `gorm:"default:true" json:"is_spot"` // true = spot (A-stock/gold ETF)
	Description string    `json:"description"`
	Manifest    string    `gorm:"type:text" json:"manifest"` // JSON blob for strategy metadata
	CreatedAt   time.Time `json:"created_at"`
}

// InstanceStatus represents the lifecycle state of a strategy instance.
type InstanceStatus string

const (
	InstanceRunning InstanceStatus = "RUNNING"
	InstanceStopped InstanceStatus = "STOPPED"
	InstanceError   InstanceStatus = "ERROR"
	InstanceDeleted InstanceStatus = "DELETED"
)

// StrategyInstance binds a strategy template to a user with capital allocation.
type StrategyInstance struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `gorm:"index;not null" json:"user_id"`
	User         User           `gorm:"foreignKey:UserID" json:"-"`
	TemplateID   string         `gorm:"index;not null" json:"template_id"`
	Template     StrategyTemplate `gorm:"foreignKey:TemplateID" json:"-"`
	Name         string         `gorm:"not null" json:"name"`
	Symbol       string         `gorm:"not null" json:"symbol"`    // e.g. "510300.SH"
	Status       InstanceStatus `gorm:"default:'STOPPED';index" json:"status"`
	ParamPack    string         `gorm:"type:text" json:"param_pack"` // JSON: Chromosome + SpawnPoint
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// PortfolioState is the current account snapshot for an instance.
type PortfolioState struct {
	ID                   uint      `gorm:"primaryKey" json:"id"`
	InstanceID           uint      `gorm:"uniqueIndex;not null" json:"instance_id"`
	CNYBalance           float64   `gorm:"not null;default:0" json:"cny_balance"`
	DeadHold             float64   `gorm:"not null;default:0" json:"dead_hold"`
	FloatHold            float64   `gorm:"not null;default:0" json:"float_hold"`
	ColdSealedHold       float64   `gorm:"not null;default:0" json:"cold_sealed_hold"`
	TotalEquity          float64   `gorm:"not null;default:0" json:"total_equity"`
	FloatUnits           float64   `gorm:"not null;default:0" json:"float_units"`   // units held in floating position
	RealizedPnL          float64   `gorm:"not null;default:0" json:"realized_pnl"`  // cumulative realized PnL
	InitialCapital       float64   `gorm:"not null;default:0" json:"initial_capital"`
	LastProcessedBarTime int64     `gorm:"default:0" json:"last_processed_bar_time"` // ms timestamp
	UpdatedAt            time.Time `json:"updated_at"`
}

// RuntimeState persists strategy internal state across ticks.
type RuntimeState struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	InstanceID uint      `gorm:"uniqueIndex;not null" json:"instance_id"`
	StateJSON  string    `gorm:"type:text" json:"state_json"` // JSON blob produced by Step()
	UpdatedAt  time.Time `json:"updated_at"`
}

// LotType represents the three-state position classification.
type LotType string

const (
	LotDeadStack   LotType = "DEAD_STACK"
	LotFloating    LotType = "FLOATING"
	LotColdSealed  LotType = "COLD_SEALED"
)

// SpotLot tracks individual position lots with cost basis.
type SpotLot struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	InstanceID   uint      `gorm:"index;not null" json:"instance_id"`
	LotType      LotType   `gorm:"not null;index" json:"lot_type"`
	Amount       float64   `gorm:"not null" json:"amount"`       // shares/units
	CostPrice    float64   `gorm:"not null" json:"cost_price"`   // avg cost per unit
	IsColdSealed bool      `gorm:"default:false" json:"is_cold_sealed"`
	CreatedAt    time.Time `json:"created_at"`
}

// TradeRecord stores completed trade records.
type TradeRecord struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	InstanceID    uint      `gorm:"index;not null" json:"instance_id"`
	ClientOrderID string    `gorm:"uniqueIndex;not null" json:"client_order_id"`
	Action        string    `gorm:"not null" json:"action"`  // "BUY" | "SELL"
	Engine        string    `gorm:"not null" json:"engine"`  // "MACRO" | "MICRO"
	Symbol        string    `gorm:"not null" json:"symbol"`
	FilledQty     float64   `gorm:"not null" json:"filled_qty"`
	FilledPrice   float64   `gorm:"not null" json:"filled_price"`
	CostBasis     float64   `gorm:"default:0" json:"cost_basis"` // USD cost basis (SELL only, same unit as filled_price)
	Fee           float64   `gorm:"default:0" json:"fee"`
	LotType       string    `gorm:"not null" json:"lot_type"` // "DEAD_STACK" | "FLOATING"
	CreatedAt     time.Time `json:"created_at"`
}

// ExecutionStatus tracks the lifecycle of a trade execution.
type ExecutionStatus string

const (
	ExecPending ExecutionStatus = "pending"
	ExecFilled  ExecutionStatus = "filled"
	ExecFailed  ExecutionStatus = "failed"
)

// SpotExecution tracks raw execution from command issuance to fill.
type SpotExecution struct {
	ID            uint            `gorm:"primaryKey" json:"id"`
	InstanceID    uint            `gorm:"index;not null" json:"instance_id"`
	ClientOrderID string          `gorm:"uniqueIndex;not null" json:"client_order_id"`
	Action        string          `gorm:"not null" json:"action"`
	Engine        string          `gorm:"not null" json:"engine"`
	Symbol        string          `gorm:"not null" json:"symbol"`
	AmountCNY     float64         `json:"amount_cny"`  // for BUY orders
	QtyAsset      float64         `json:"qty_asset"`   // for SELL orders
	LotType       string          `gorm:"not null" json:"lot_type"`
	Status        ExecutionStatus `gorm:"default:'pending';index" json:"status"`
	FilledQty     float64         `gorm:"default:0" json:"filled_qty"`
	FilledPrice   float64         `gorm:"default:0" json:"filled_price"`
	Fee           float64         `gorm:"default:0" json:"fee"`
	ErrorMsg      string          `json:"error_msg"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// AuditLog stores immutable audit trail events.
type AuditLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	InstanceID uint     `gorm:"index" json:"instance_id"`
	EventType string    `gorm:"not null;index" json:"event_type"`
	Payload   string    `gorm:"type:text" json:"payload"` // JSON blob
	CreatedAt time.Time `json:"created_at"`
}

// GeneRole represents the three-state gene lifecycle.
type GeneRole string

const (
	GeneChallenger GeneRole = "challenger"
	GeneChampion   GeneRole = "champion"
	GeneRetired    GeneRole = "retired"
)

// GeneRecord stores evolved gene packages.
type GeneRecord struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	StrategyID   string    `gorm:"index;not null" json:"strategy_id"`
	Role         GeneRole  `gorm:"not null;index" json:"role"`
	ParamPack    string    `gorm:"type:text;not null" json:"param_pack"` // JSON: Chromosome + SpawnPoint
	ScoreTotal   float64   `json:"score_total"`
	MaxDrawdown  float64   `json:"max_drawdown"`
	Score6M      float64   `json:"score_6m"`
	Score2Y      float64   `json:"score_2y"`
	Score5Y      float64   `json:"score_5y"`
	ScoreFull    float64   `json:"score_full"`
	ActivatedAt  *time.Time `json:"activated_at"` // when promoted to champion
	CreatedAt    time.Time  `json:"created_at"`
}

// EvolutionTaskStatus represents the state of an evolution run.
type EvolutionTaskStatus string

const (
	EvoRunning   EvolutionTaskStatus = "running"
	EvoCompleted EvolutionTaskStatus = "completed"
	EvoFailed    EvolutionTaskStatus = "failed"
)

// EvolutionTask tracks GA evolution jobs.
type EvolutionTask struct {
	ID          uint                `gorm:"primaryKey" json:"id"`
	StrategyID  string              `gorm:"index;not null" json:"strategy_id"`
	Status      EvolutionTaskStatus `gorm:"default:'running';index" json:"status"`
	Progress    string              `gorm:"type:text" json:"progress"` // JSON: current gen, best score, etc.
	Config      string              `gorm:"type:text" json:"config"`   // JSON: pop_size, max_generations, spawn_mode
	Result      string              `gorm:"type:text" json:"result"`   // JSON: final EpochResult
	StartedAt   time.Time           `json:"started_at"`
	CompletedAt *time.Time          `json:"completed_at"`
	CreatedAt   time.Time           `json:"created_at"`
}

// KLine stores historical candlestick data.
type KLine struct {
	ID        uint    `gorm:"primaryKey" json:"id"`
	Symbol    string  `gorm:"uniqueIndex:idx_kline_symbol_interval_time;not null" json:"symbol"`
	Interval  string  `gorm:"uniqueIndex:idx_kline_symbol_interval_time;not null" json:"interval"` // "1m", "1h", "1d"
	OpenTime  int64   `gorm:"uniqueIndex:idx_kline_symbol_interval_time;not null" json:"open_time"` // ms
	Open      float64 `gorm:"not null" json:"open"`
	High      float64 `gorm:"not null" json:"high"`
	Low       float64 `gorm:"not null" json:"low"`
	Close     float64 `gorm:"not null" json:"close"`
	Volume    float64 `gorm:"not null" json:"volume"`
	CreatedAt time.Time `json:"created_at"`
}

// EquitySnapshot stores periodic NAV snapshots for charting.
type EquitySnapshot struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	InstanceID uint      `gorm:"index;not null" json:"instance_id"`
	TotalEquity float64  `gorm:"not null" json:"total_equity"`
	RecordedAt time.Time `gorm:"index;not null" json:"recorded_at"`
}

// AutoMigrateAll runs GORM AutoMigrate for all models.
func AutoMigrateAll(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&StrategyTemplate{},
		&StrategyInstance{},
		&PortfolioState{},
		&RuntimeState{},
		&SpotLot{},
		&TradeRecord{},
		&SpotExecution{},
		&AuditLog{},
		&GeneRecord{},
		&EvolutionTask{},
		&KLine{},
		&EquitySnapshot{},
	)
}
