// Package api provides HTTP handlers for the QuantSaaS REST API.
package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/chenthewho/ma-cross-strategy/internal/saas/epoch"
	"github.com/chenthewho/ma-cross-strategy/internal/saas/store"
)

// RedisChampionKey is the Redis cache key for the active champion gene.
const RedisChampionKey = "champion:gene:current"

// EvolutionHandler holds dependencies for evolution endpoints.
type EvolutionHandler struct {
	db          *store.DB
	redis       *store.RedisClient
	epochSvc    *epoch.EpochService
	logger      *zap.Logger
}

// NewEvolutionHandler creates a new EvolutionHandler.
func NewEvolutionHandler(db *store.DB, redis *store.RedisClient, epochSvc *epoch.EpochService, logger *zap.Logger) *EvolutionHandler {
	return &EvolutionHandler{
		db:       db,
		redis:    redis,
		epochSvc: epochSvc,
		logger:   logger,
	}
}

// RegisterRoutes registers evolution routes on the given gin.RouterGroup.
// These endpoints are only accessible in lab/dev mode (enforced by middleware).
func (h *EvolutionHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/tasks", h.CreateTask)
	rg.GET("/tasks", h.ListTasks)
	rg.POST("/tasks/:id/promote", h.PromoteTask)
}

// RegisterChampionRoute registers the champion genome endpoint.
func (h *EvolutionHandler) RegisterChampionRoute(rg *gin.RouterGroup) {
	rg.GET("/champion", h.GetChampion)
}

// ── Handlers ────────────────────────────────────────────────

// CreateTask handles POST /api/v1/evolution/tasks.
// Creates a new evolution task and starts the GA run.
func (h *EvolutionHandler) CreateTask(c *gin.Context) {
	var req epoch.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.epochSvc.CreateAndRunTask(c.Request.Context(), req)
	if err != nil {
		if err == epoch.ErrTaskAlreadyRunning {
			c.JSON(http.StatusConflict, gin.H{"error": "evolution task already running"})
			return
		}
		h.logger.Error("create evolution task failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create task"})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// ListTasks handles GET /api/v1/evolution/tasks.
// Returns the list of evolution tasks along with challenger gene records.
func (h *EvolutionHandler) ListTasks(c *gin.Context) {
	ctx := c.Request.Context()

	// Fetch all evolution tasks, ordered by most recent first
	var tasks []store.EvolutionTask
	if err := h.db.WithContext(ctx).Order("created_at DESC").Find(&tasks).Error; err != nil {
		h.logger.Error("list evolution tasks failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tasks"})
		return
	}

	// Fetch all challenger gene records
	var challengers []store.GeneRecord
	if err := h.db.WithContext(ctx).Where("role = ?", store.GeneChallenger).Order("created_at DESC").Find(&challengers).Error; err != nil {
		h.logger.Error("list challengers failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list challengers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks":       tasks,
		"challengers": challengers,
	})
}

// PromoteTask handles POST /api/v1/evolution/tasks/:id/promote.
// Performs the promote transaction: current champion → retired,
// selected challenger → champion, then deletes the Redis champion cache.
func (h *EvolutionHandler) PromoteTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	ctx := context.Background()

	// Execute the promote transaction
	err = h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Find the challenger gene by ID
		var challenger store.GeneRecord
		if err := tx.Where("id = ? AND role = ?", id, store.GeneChallenger).First(&challenger).Error; err != nil {
			return fmt.Errorf("challenger not found: %w", err)
		}

		// 2. Demote current champion → retired
		now := time.Now()
		if err := tx.Model(&store.GeneRecord{}).
			Where("role = ?", store.GeneChampion).
			Updates(map[string]any{
				"role": store.GeneRetired,
			}).Error; err != nil {
			return fmt.Errorf("demote champion: %w", err)
		}

		// 3. Promote challenger → champion
		if err := tx.Model(&challenger).Updates(map[string]any{
			"role":         store.GeneChampion,
			"activated_at": &now,
		}).Error; err != nil {
			return fmt.Errorf("promote challenger: %w", err)
		}

		return nil
	})

	if err != nil {
		h.logger.Error("promote failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 4. Delete Redis champion cache to force reload
	if delErr := h.redis.Del(ctx, RedisChampionKey); delErr != nil {
		h.logger.Warn("failed to delete champion cache", zap.Error(delErr))
	}

	c.JSON(http.StatusOK, gin.H{"message": "promoted successfully"})
}

// GetChampion handles GET /api/v1/genome/champion.
// Reads from Redis cache first, falls back to the database.
func (h *EvolutionHandler) GetChampion(c *gin.Context) {
	ctx := c.Request.Context()

	// Try Redis cache first
	var geneRecord store.GeneRecord
	found, err := h.redis.GetJSON(ctx, RedisChampionKey, &geneRecord)
	if err != nil {
		h.logger.Warn("redis get champion failed, falling back to DB", zap.Error(err))
	}

	if found {
		c.JSON(http.StatusOK, geneRecord)
		return
	}

	// Cache miss — query database
	if err := h.db.WithContext(ctx).
		Where("role = ?", store.GeneChampion).
		Order("activated_at DESC").
		First(&geneRecord).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no champion found"})
		return
	}

	// Populate Redis cache (no TTL — stays until promote invalidates it)
	if setErr := h.redis.SetJSON(ctx, RedisChampionKey, geneRecord, 0); setErr != nil {
		h.logger.Warn("failed to cache champion in redis", zap.Error(setErr))
	}

	c.JSON(http.StatusOK, geneRecord)
}
