// Package instance manages strategy instance lifecycle and tick execution.
package instance

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chenthewho/ma-cross-strategy/internal/quant"
	"github.com/chenthewho/ma-cross-strategy/internal/saas/store"
	"github.com/chenthewho/ma-cross-strategy/internal/strategies/golden_cross"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Manager handles instance lifecycle (START/STOP) and tick execution.
type Manager struct {
	db     *store.DB
	logger *zap.Logger
}

func NewManager(db *store.DB, logger *zap.Logger) *Manager {
	return &Manager{db: db, logger: logger}
}

// StartInstance transitions an instance to RUNNING state.
func (m *Manager) StartInstance(ctx context.Context, instanceID uint) error {
	return m.db.WithContext(ctx).Model(&store.StrategyInstance{}).
		Where("id = ? AND status = ?", instanceID, store.InstanceStopped).
		Update("status", store.InstanceRunning).Error
}

// StopInstance transitions an instance to STOPPED state.
func (m *Manager) StopInstance(ctx context.Context, instanceID uint) error {
	return m.db.WithContext(ctx).Model(&store.StrategyInstance{}).
		Where("id = ? AND status = ?", instanceID, store.InstanceRunning).
		Update("status", store.InstanceStopped).Error
}

// Tick runs a single tick for a RUNNING instance.
// Called by the cron scheduler every minute.
func (m *Manager) Tick(ctx context.Context, instance store.StrategyInstance) error {
	// 1. Load portfolio state
	var ps store.PortfolioState
	if err := m.db.WithContext(ctx).Where("instance_id = ?", instance.ID).First(&ps).Error; err != nil {
		return fmt.Errorf("load portfolio: %w", err)
	}

	// 2. Load runtime state
	var rs store.RuntimeState
	err := m.db.WithContext(ctx).Where("instance_id = ?", instance.ID).First(&rs).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("load runtime: %w", err)
	}

	// 3. Load current champion params
	params := golden_cross.Params{
		Chromosome: quant.DefaultSeedChromosome,
		SpawnPoint: quant.DefaultSpawnPoint,
	}
	var champ store.GeneRecord
	if err := m.db.WithContext(ctx).Where("strategy_id = ? AND role = ?",
		instance.TemplateID, store.GeneChampion).First(&champ).Error; err == nil {
		pp := quant.DecodeParamPack([]byte(champ.ParamPack))
		params.Chromosome = pp.Chromosome
		params.SpawnPoint = pp.SpawnPoint
	}

	// 4. Build StrategyInput (simplified — full impl would pull K-lines from DB)
	input := quant.StrategyInput{
		Closes:     []float64{0},
		Timestamps: []int64{time.Now().UnixMilli()},
		Portfolio: quant.PortfolioSnapshot{
			CNYBalance:     ps.CNYBalance,
			DeadHold:       ps.DeadHold,
			FloatHold:      ps.FloatHold,
			ColdSealedHold: ps.ColdSealedHold,
			TotalEquity:    ps.TotalEquity,
		},
		Runtime: quant.RuntimeState{LastProcessedBar: ps.LastProcessedBarTime},
	}

	// 5. Execute Step()
	_ = golden_cross.Step(input, params)

	// 6. Persist updated state
	var runtimeJSON []byte
	if rs.StateJSON == "" {
		runtimeJSON, _ = json.Marshal(map[string]int64{"last_bar": time.Now().UnixMilli()})
	} else {
		runtimeJSON = []byte(rs.StateJSON)
	}
	m.db.WithContext(ctx).Model(&rs).Updates(map[string]any{
		"state_json": string(runtimeJSON),
		"updated_at": time.Now(),
	})
	m.db.WithContext(ctx).Model(&ps).Updates(map[string]any{
		"last_processed_bar_time": time.Now().UnixMilli(),
		"updated_at":             time.Now(),
	})

	return nil
}
