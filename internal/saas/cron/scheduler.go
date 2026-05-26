// Package cron provides the tick scheduler for running strategy instances.
package cron

import (
	"context"
	"time"

	"github.com/chenthewho/ma-cross-strategy/internal/saas/instance"
	"github.com/chenthewho/ma-cross-strategy/internal/saas/store"
	"go.uber.org/zap"
)

// Scheduler periodically scans RUNNING instances and executes their ticks.
type Scheduler struct {
	db      *store.DB
	manager *instance.Manager
	logger  *zap.Logger
	interval time.Duration
}

func NewScheduler(db *store.DB, manager *instance.Manager, logger *zap.Logger) *Scheduler {
	return &Scheduler{
		db: db, manager: manager, logger: logger,
		interval: 1 * time.Minute,
	}
}

// Start begins the cron loop. Blocks until ctx is cancelled.
func (s *Scheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	s.logger.Info("cron scheduler started", zap.Duration("interval", s.interval))

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("cron scheduler stopped")
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) {
	var instances []store.StrategyInstance
	if err := s.db.WithContext(ctx).
		Where("status = ?", store.InstanceRunning).
		Find(&instances).Error; err != nil {
		s.logger.Error("cron scan failed", zap.Error(err))
		return
	}

	for _, inst := range instances {
		go func(i store.StrategyInstance) {
			if err := s.manager.Tick(ctx, i); err != nil {
				s.logger.Error("tick failed",
					zap.Uint("instance_id", i.ID),
					zap.Error(err),
				)
			}
		}(inst)
	}
}
