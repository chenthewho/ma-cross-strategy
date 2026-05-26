// Package epoch provides the evolution task service.
package epoch

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/chenthewho/ma-cross-strategy/internal/saas/ga"
	"github.com/chenthewho/ma-cross-strategy/internal/saas/store"
)

// CreateTaskRequest is the payload for POST /api/v1/evolution/tasks.
type CreateTaskRequest struct {
	StrategyID     string          `json:"strategy_id" binding:"required"`
	PopSize        int             `json:"pop_size"`
	MaxGenerations int             `json:"max_generations"`
	SpawnMode      string          `json:"spawn_mode"`
	SpawnPoint     json.RawMessage `json:"spawn_point"`
	TestMode       bool            `json:"test_mode"`
}

// EpochService manages evolution task lifecycle.
type EpochService struct {
	db          *store.DB
	engine      *ga.EvolutionEngine
	logger      *zap.Logger
	mu          sync.Mutex
	currentTask *store.EvolutionTask
}

// NewEpochService creates a new EpochService.
func NewEpochService(db *store.DB, engine *ga.EvolutionEngine, logger *zap.Logger) *EpochService {
	return &EpochService{db: db, engine: engine, logger: logger}
}

// CreateAndRunTask creates a new EvolutionTask and starts GA in background.
func (s *EpochService) CreateAndRunTask(ctx context.Context, req CreateTaskRequest) (*store.EvolutionTask, error) {
	s.mu.Lock()
	if s.currentTask != nil {
		s.mu.Unlock()
		return nil, ErrTaskAlreadyRunning
	}
	if req.PopSize == 0 { req.PopSize = 300 }
	if req.MaxGenerations == 0 { req.MaxGenerations = 25 }
	if req.SpawnMode == "" { req.SpawnMode = "inherit" }
	if req.TestMode { req.PopSize, req.MaxGenerations = 10, 3 }

	configJSON, _ := json.Marshal(map[string]any{
		"pop_size": req.PopSize, "max_generations": req.MaxGenerations,
		"spawn_mode": req.SpawnMode, "test_mode": req.TestMode,
	})

	task := &store.EvolutionTask{
		StrategyID: req.StrategyID, Status: store.EvoRunning,
		Config: string(configJSON), StartedAt: time.Now(),
	}
	if err := s.db.WithContext(ctx).Create(task).Error; err != nil {
		s.mu.Unlock()
		return nil, fmt.Errorf("create evolution task: %w", err)
	}
	s.currentTask = task
	s.mu.Unlock()

	go s.runEpoch(task, req)
	return task, nil
}

func (s *EpochService) IsRunning() bool {
	s.mu.Lock(); defer s.mu.Unlock()
	return s.currentTask != nil
}

var ErrTaskAlreadyRunning = fmt.Errorf("evolution task already running")

func (s *EpochService) runEpoch(task *store.EvolutionTask, req CreateTaskRequest) {
	ctx := context.Background()
	s.logger.Info("epoch started",
		zap.Uint("task_id", task.ID), zap.String("strategy_id", req.StrategyID),
		zap.Int("pop_size", req.PopSize), zap.Int("max_generations", req.MaxGenerations),
	)

	// Build EvaluablePlan from DB
	plan := ga.EvaluablePlan{Symbol: "510300.SH", TemplateName: req.StrategyID}
	cfg := ga.EpochConfig{PopSize: req.PopSize, MaxGenerations: req.MaxGenerations}

	result, err := s.engine.RunEpoch(ctx, cfg, plan)
	now := time.Now()

	if err != nil {
		s.logger.Error("epoch failed", zap.Uint("task_id", task.ID), zap.Error(err))
		s.db.WithContext(context.Background()).Model(task).Updates(map[string]any{
			"status": store.EvoFailed, "completed_at": &now,
			"result": fmt.Sprintf(`{"error": %q}`, err.Error()),
		})
	} else {
		championJSON := s.engine.EncodeResult(result.Champion, req.SpawnPoint)
		s.logger.Info("epoch completed",
			zap.Uint("task_id", task.ID), zap.Float64("score_total", result.ScoreTotal),
		)
		resultJSON, _ := json.Marshal(result)
		s.db.WithContext(context.Background()).Model(task).Updates(map[string]any{
			"status": store.EvoCompleted, "result": string(resultJSON), "completed_at": &now,
		})
		geneRecord := &store.GeneRecord{
			StrategyID: req.StrategyID, Role: store.GeneChallenger,
			ParamPack: string(championJSON), ScoreTotal: result.ScoreTotal, MaxDrawdown: result.MaxDD,
		}
		if len(result.Scores) >= 4 {
			geneRecord.Score6M, geneRecord.Score2Y = result.Scores[0], result.Scores[1]
			geneRecord.Score5Y, geneRecord.ScoreFull = result.Scores[2], result.Scores[3]
		}
		if err := s.db.WithContext(context.Background()).Create(geneRecord).Error; err != nil {
			s.logger.Error("create challenger gene failed", zap.Error(err))
		}
	}
	s.mu.Lock()
	s.currentTask = nil
	s.mu.Unlock()
}
