// Package ws provides DeltaReport processing for the WebSocket Hub.
package ws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/chenthewho/ma-cross-strategy/internal/saas/store"
	"go.uber.org/zap"
)

// DeltaReportHandler processes incoming DeltaReports from agents,
// updating PortfolioState, SpotExecution, TradeRecord, and AuditLog.
type DeltaReportHandler struct {
	db     *store.DB
	logger *zap.Logger
}

// NewDeltaReportHandler creates a handler backed by the given DB and logger.
func NewDeltaReportHandler(db *store.DB, logger *zap.Logger) *DeltaReportHandler {
	return &DeltaReportHandler{db: db, logger: logger}
}

// DeltaReportPayload is the expected payload of a delta_report message.
type DeltaReportPayload struct {
	ClientOrderID string          `json:"client_order_id,omitempty"`
	Balances      []BalanceItem   `json:"balances"`
	Execution     *ExecutionItem  `json:"execution,omitempty"`
}

// BalanceItem represents a single asset balance from the broker.
type BalanceItem struct {
	Asset  string  `json:"asset"`
	Free   float64 `json:"free"`
	Frozen float64 `json:"frozen"`
}

// ExecutionItem represents a completed trade execution from the broker.
type ExecutionItem struct {
	OrderID     string  `json:"order_id"`
	FilledQty   float64 `json:"filled_qty"`
	FilledPrice float64 `json:"filled_price"`
	Fee         float64 `json:"fee"`
	Status      string  `json:"status"`
}

// Process processes a delta_report from an agent.
// If ClientOrderID is present, it updates the corresponding SpotExecution and creates a TradeRecord.
// If ClientOrderID is empty (initial snapshot on reconnect), only balances are updated.
func (h *DeltaReportHandler) Process(ctx context.Context, instanceID uint, payload json.RawMessage) error {
	var report DeltaReportPayload
	if err := json.Unmarshal(payload, &report); err != nil {
		return fmt.Errorf("unmarshal delta_report: %w", err)
	}

	// Always update balances from the agent's real account snapshot
	if len(report.Balances) > 0 {
		if err := h.updateBalances(ctx, instanceID, report.Balances); err != nil {
			h.logger.Warn("failed to update balances from delta_report",
				zap.Uint("instance_id", instanceID),
				zap.Error(err),
			)
		}
	}

	// If ClientOrderID is present, match to pending SpotExecution
	if report.ClientOrderID != "" && report.Execution != nil {
		if err := h.processExecution(ctx, instanceID, report); err != nil {
			h.logger.Error("failed to process execution from delta_report",
				zap.String("client_order_id", report.ClientOrderID),
				zap.Error(err),
			)
			return err
		}
	}

	return nil
}

func (h *DeltaReportHandler) updateBalances(ctx context.Context, instanceID uint, balances []BalanceItem) error {
	var ps store.PortfolioState
	if err := h.db.WithContext(ctx).Where("instance_id = ?", instanceID).First(&ps).Error; err != nil {
		return fmt.Errorf("find portfolio state: %w", err)
	}

	for _, b := range balances {
		switch b.Asset {
		case "CNY":
			ps.CNYBalance = b.Free + b.Frozen
		}
	}
	// TotalEquity simplified: CNYBalance + holdings at current value
	ps.TotalEquity = ps.CNYBalance + (ps.DeadHold+ps.FloatHold+ps.ColdSealedHold)*0 // placeholder

	if err := h.db.WithContext(ctx).Save(&ps).Error; err != nil {
		return fmt.Errorf("save portfolio state: %w", err)
	}

	// Audit log
	h.db.WithContext(ctx).Create(&store.AuditLog{
		InstanceID: instanceID,
		EventType:  "balance_update",
		Payload:    fmt.Sprintf(`{"cny_balance": %f}`, ps.CNYBalance),
	})

	return nil
}

func (h *DeltaReportHandler) processExecution(ctx context.Context, instanceID uint, report DeltaReportPayload) error {
	exec := report.Execution

	// Find the pending SpotExecution by ClientOrderID
	var se store.SpotExecution
	if err := h.db.WithContext(ctx).Where("client_order_id = ?", report.ClientOrderID).First(&se).Error; err != nil {
		return fmt.Errorf("find spot execution: %w", err)
	}

	// Update SpotExecution
	se.Status = store.ExecFilled
	se.FilledQty = exec.FilledQty
	se.FilledPrice = exec.FilledPrice
	se.Fee = exec.Fee
	if err := h.db.WithContext(ctx).Save(&se).Error; err != nil {
		return fmt.Errorf("save spot execution: %w", err)
	}

	// Create TradeRecord
	tr := store.TradeRecord{
		InstanceID:    instanceID,
		ClientOrderID: report.ClientOrderID,
		Action:        se.Action,
		Engine:        se.Engine,
		Symbol:        se.Symbol,
		FilledQty:     exec.FilledQty,
		FilledPrice:   exec.FilledPrice,
		Fee:           exec.Fee,
		LotType:       se.LotType,
	}
	if err := h.db.WithContext(ctx).Create(&tr).Error; err != nil {
		return fmt.Errorf("create trade record: %w", err)
	}

	// Update PortfolioState based on LotType
	var ps store.PortfolioState
	if err := h.db.WithContext(ctx).Where("instance_id = ?", instanceID).First(&ps).Error; err != nil {
		return fmt.Errorf("find portfolio state: %w", err)
	}

	switch se.LotType {
	case string(store.LotDeadStack):
		if se.Action == "BUY" {
			ps.DeadHold += exec.FilledQty
		}
	case string(store.LotFloating):
		if se.Action == "BUY" {
			ps.FloatHold += exec.FilledQty
		} else {
			ps.FloatHold -= exec.FilledQty
		}
	}
	if err := h.db.WithContext(ctx).Save(&ps).Error; err != nil {
		return fmt.Errorf("save portfolio state: %w", err)
	}

	// Audit log
	h.db.WithContext(ctx).Create(&store.AuditLog{
		InstanceID: instanceID,
		EventType:  "execution_filled",
		Payload:    fmt.Sprintf(`{"client_order_id": "%s", "qty": %f, "price": %f}`, report.ClientOrderID, exec.FilledQty, exec.FilledPrice),
	})

	return nil
}
