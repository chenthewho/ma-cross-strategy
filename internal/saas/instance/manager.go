// Package instance manages strategy instance lifecycle and tick execution.
package instance

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chenthewho/ma-cross-strategy/internal/quant"
	"github.com/chenthewho/ma-cross-strategy/internal/saas/store"
	gc "github.com/chenthewho/ma-cross-strategy/internal/strategies/golden_cross"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Manager handles instance lifecycle (START/STOP) and tick execution.
type Manager struct {
	db     *store.DB
	logger *zap.Logger
	// sendCommand is called to deliver TradeCommands to agents via WebSocket.
	// nil means no agent connectivity (lab/dev mode or agent offline).
	sendCommand func(userID uint, cmd map[string]any) error
}

// NewManager creates an instance manager.
func NewManager(db *store.DB, logger *zap.Logger) *Manager {
	return &Manager{db: db, logger: logger}
}

// SetCommandSender sets the function used to deliver TradeCommands to agents.
func (m *Manager) SetCommandSender(fn func(userID uint, cmd map[string]any) error) {
	m.sendCommand = fn
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

// Tick runs a single tick for a RUNNING instance, implementing the full 10-step pipeline:
//
//	1. Idempotent bucket dedup — skip if bar already processed
//	2. Load PortfolioState + RuntimeState from DB
//	3. Load champion param pack (DB or Redis)
//	4. ACL outer ring — OHLCV strip to closes[] + timestamps[]
//	5. Build StrategyInput
//	6. Call Step() — single source of truth
//	7. Persist RuntimeState
//	8. Handle dead release intents — SaaS-side ledger only, write AuditLog
//	9. Translate intents to TradeCommands, write pending SpotExecutions, send via WS
//	10. Update LastProcessedBarTime
func (m *Manager) Tick(ctx context.Context, instance store.StrategyInstance) error {
	log := m.logger.With(zap.Uint("instance_id", instance.ID))

	// ── 1. Idempotent bucket dedup ──
	// Pull latest completed bar from KLine table for this symbol+interval.
	// If that bar timestamp <= LastProcessedBarTime, skip (same bucket already processed).
	var latestBar store.KLine
	err := m.db.WithContext(ctx).
		Where("symbol = ? AND interval = ?", instance.Symbol, "1h").
		Order("open_time DESC").First(&latestBar).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("fetch latest bar: %w", err)
	}

	var ps store.PortfolioState
	if err := m.db.WithContext(ctx).Where("instance_id = ?", instance.ID).First(&ps).Error; err != nil {
		return fmt.Errorf("load portfolio: %w", err)
	}

	if err == nil && latestBar.OpenTime <= ps.LastProcessedBarTime {
		log.Debug("tick skipped — bar already processed",
			zap.Int64("bar_time", latestBar.OpenTime),
			zap.Int64("last_processed", ps.LastProcessedBarTime),
		)
		return nil
	}

	// ── 2. Load RuntimeState ──
	var rs store.RuntimeState
	if err := m.db.WithContext(ctx).Where("instance_id = ?", instance.ID).First(&rs).Error; err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("load runtime: %w", err)
	}

	var runtime quant.RuntimeState
	if rs.StateJSON != "" {
		json.Unmarshal([]byte(rs.StateJSON), &runtime)
	}

	// ── 3. Load champion params ──
	params := gc.Params{
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

	// ── 4. ACL outer ring — OHLCV strip ──
	var bars []quant.Bar
	if err == nil {
		var klines []store.KLine
		m.db.WithContext(ctx).
			Where("symbol = ? AND interval = ?", instance.Symbol, "1h").
			Order("open_time ASC").Limit(500).Find(&klines)
		for _, k := range klines {
			bars = append(bars, quant.Bar{
				OpenTime: k.OpenTime,
				Open:     k.Open, High: k.High, Low: k.Low,
				Close: k.Close, Volume: k.Volume,
			})
		}
	}

	closes := quant.ExtractCloses(bars)
	timestamps := quant.ExtractTimestamps(bars)
	currentPrice := 0.0
	if len(closes) > 0 {
		currentPrice = closes[len(closes)-1]
	}

	// ── 5. Build StrategyInput ──
	input := quant.StrategyInput{
		Closes:       closes,
		Timestamps:   timestamps,
		CurrentPrice: currentPrice,
		Portfolio: quant.PortfolioSnapshot{
			CNYBalance:     ps.CNYBalance,
			DeadHold:       ps.DeadHold,
			FloatHold:      ps.FloatHold,
			ColdSealedHold: ps.ColdSealedHold,
			TotalEquity:    ps.TotalEquity,
		},
		Runtime: runtime,
	}

	// ── 6. Call Step() — single source of truth, same as backtest ──
	output := gc.Step(input, params)

	// ── 7. Persist RuntimeState ──
	output.NewRuntime.LastProcessedBar = ps.LastProcessedBarTime
	runtimeJSON, _ := json.Marshal(output.NewRuntime)
	if rs.ID == 0 {
		m.db.WithContext(ctx).Create(&store.RuntimeState{
			InstanceID: instance.ID,
			StateJSON:  string(runtimeJSON),
		})
	} else {
		m.db.WithContext(ctx).Model(&rs).Updates(map[string]any{
			"state_json": string(runtimeJSON),
			"updated_at": time.Now(),
		})
	}

	// ── 8. Handle dead release intents — SaaS-side only ──
	if output.ReleaseIntent != nil {
		log.Info("dead release executed",
			zap.String("type", output.ReleaseIntent.ReleaseType),
			zap.Float64("amount", output.ReleaseIntent.ReleaseAmount),
		)
		m.db.WithContext(ctx).Create(&store.AuditLog{
			InstanceID: instance.ID,
			EventType:  "dead_release",
			Payload: fmt.Sprintf(`{"type": "%s", "amount": %f, "reason": "%s"}`,
				output.ReleaseIntent.ReleaseType,
				output.ReleaseIntent.ReleaseAmount,
				output.ReleaseIntent.Reason,
			),
		})
	}

	// ── 9. Translate intents to TradeCommands + send via WebSocket ──
	ts := time.Now().Unix()
	userID := instance.UserID

	sendTradeCmd := func(intent *quant.MacroIntent, engine string) {
		if intent == nil || intent.AmountCNY == 0 {
			return
		}
		clientOrderID := fmt.Sprintf("inst%d-%s-%d", instance.ID, engine, ts)
		cmd := map[string]any{
			"type":            "command",
			"client_order_id": clientOrderID,
			"action":          intent.Action,
			"engine":          engine,
			"symbol":          instance.Symbol,
			"amount_cny":      fmt.Sprintf("%.2f", intent.AmountCNY),
			"lot_type":        intent.LotType,
		}

		// Write pending SpotExecution
		se := store.SpotExecution{
			InstanceID:    instance.ID,
			ClientOrderID: clientOrderID,
			Action:        intent.Action,
			Engine:        engine,
			Symbol:        instance.Symbol,
			AmountCNY:     intent.AmountCNY,
			LotType:       intent.LotType,
			Status:        store.ExecPending,
		}
		if err := m.db.WithContext(ctx).Create(&se).Error; err != nil {
			log.Error("failed to write pending execution", zap.Error(err))
			return
		}

		// Send via WebSocket Hub
		if m.sendCommand != nil {
			if err := m.sendCommand(userID, cmd); err != nil {
				log.Warn("agent not connected, command queued for next tick",
					zap.String("client_order_id", clientOrderID),
					zap.Error(err),
				)
			}
		} else {
			log.Warn("no command sender configured — agent offline",
				zap.String("client_order_id", clientOrderID),
			)
		}
	}

	sendTradeCmd(output.MacroIntent, "MACRO")
	sendTradeCmd(func() *quant.MacroIntent {
		if output.MicroIntent == nil {
			return nil
		}
		return &quant.MacroIntent{
			Action:    output.MicroIntent.Action,
			AmountCNY: output.MicroIntent.AmountCNY,
			Engine:    "MICRO",
			LotType:   output.MicroIntent.LotType,
		}
	}(), "MICRO")

	// ── 10. Update LastProcessedBarTime ──
	if err == nil {
		m.db.WithContext(ctx).Model(&ps).Updates(map[string]any{
			"last_processed_bar_time": latestBar.OpenTime,
			"updated_at":              time.Now(),
		})
	} else {
		m.db.WithContext(ctx).Model(&ps).Updates(map[string]any{
			"last_processed_bar_time": time.Now().UnixMilli(),
			"updated_at":              time.Now(),
		})
	}

	return nil
}
