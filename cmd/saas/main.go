// Package main is the SaaS HTTP server entry point.
// It starts the QuantSaaS platform: gin server, DB, Redis, cron scheduler, GA engine.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/chenthewho/ma-cross-strategy/internal/saas/api"
	"github.com/chenthewho/ma-cross-strategy/internal/saas/auth"
	"github.com/chenthewho/ma-cross-strategy/internal/saas/config"
	"github.com/chenthewho/ma-cross-strategy/internal/saas/cron"
	"github.com/chenthewho/ma-cross-strategy/internal/saas/ga"
	gcEvolvable "github.com/chenthewho/ma-cross-strategy/internal/saas/ga"
	"github.com/chenthewho/ma-cross-strategy/internal/saas/instance"
	"github.com/chenthewho/ma-cross-strategy/internal/saas/store"
	"github.com/chenthewho/ma-cross-strategy/internal/saas/ws"
)

func main() {
	// ── 1. Load config ────────────────────────────────────────
	configPath := "config.yaml"
	if v := os.Getenv("CONFIG_PATH"); v != "" {
		configPath = v
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// ── 2. Init logger ────────────────────────────────────────
	logger := newLogger(cfg.Server.Mode)

	// ── 3. Connect DB + AutoMigrate ───────────────────────────
	db, err := store.NewDB(cfg.Database)
	if err != nil {
		logger.Fatal("failed to connect database", zap.Error(err))
	}
	logger.Info("database connected and migrated")

	// ── 4. Connect Redis ──────────────────────────────────────
	redis, err := store.NewRedisClient(cfg.Redis.Addr, cfg.Redis.DB)
	if err != nil {
		logger.Warn("redis connection failed — continuing without cache", zap.Error(err))
	}
	if redis != nil {
		defer redis.Close()
		logger.Info("redis connected")
	}

	// ── 5. Init token service (needed for hub) ────────────────
	tokenSvc := auth.NewTokenService(cfg.JWT.Secret, cfg.JWT.ExpireHours)

	// ── 6. Init WebSocket Hub ─────────────────────────────────
	hub := ws.NewHub(tokenSvc)

	// ── 7. Init instance manager + recover RUNNING instances ──
	mgr := instance.NewManager(db, logger)
	recoverRunningInstances(db, mgr, logger)

	// ── 8. Init GA engine (lab/dev mode only) ─────────────────
	if cfg.AppRole == "lab" || cfg.AppRole == "dev" {
		initGAEngine(logger)
	}

	// ── 9. Setup gin server ───────────────────────────────────
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Recovery())

	api.SetupRoutes(router, db, hub, tokenSvc, cfg.AppRole)

	// ── 9a. Serve frontend static files (SPA fallback) ────────
	router.Use(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") || strings.HasPrefix(c.Request.URL.Path, "/ws/") {
			c.Next()
			return
		}
		filePath := "./web-frontend/dist" + c.Request.URL.Path
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			c.File("./web-frontend/dist/index.html")
			c.Abort()
			return
		}
		c.File(filePath)
		c.Abort()
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	// ── 10. Start cron scheduler ──────────────────────────────
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	scheduler := cron.NewScheduler(db, mgr, logger)
	go scheduler.Start(ctx)

	// ── 11. Start gin server ──────────────────────────────────
	go func() {
		logger.Info("saas server starting", zap.Int("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	// ── 12. Graceful shutdown on SIGTERM ──────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.Info("shutting down", zap.String("signal", sig.String()))

	cancel() // stop cron scheduler

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server exited cleanly")
}

// newLogger creates a zap logger with appropriate level based on mode.
func newLogger(mode string) *zap.Logger {
	var cfg zap.Config
	if mode == "release" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	logger, err := cfg.Build()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	return logger
}

// recoverRunningInstances finds all RUNNING instances and resets them to STOPPED
// (they will be restarted by the user explicitly after a server restart).
func recoverRunningInstances(db *store.DB, mgr *instance.Manager, logger *zap.Logger) {
	var instances []store.StrategyInstance
	if err := db.Where("status = ?", store.InstanceRunning).Find(&instances).Error; err != nil {
		logger.Warn("failed to scan running instances for recovery", zap.Error(err))
		return
	}
	if len(instances) == 0 {
		return
	}
	logger.Info("recovering running instances after restart", zap.Int("count", len(instances)))
	for _, inst := range instances {
		if err := mgr.StopInstance(context.Background(), inst.ID); err != nil {
			logger.Warn("failed to stop instance during recovery",
				zap.Uint("instance_id", inst.ID), zap.Error(err))
		}
	}
}

// initGAEngine initializes the GA evolution engine for lab/dev mode.
func initGAEngine(logger *zap.Logger) {
	evolvable := &gcEvolvable.GoldenCrossEvolvable{}
	_ = ga.NewEngine(evolvable, time.Now().UnixNano())
	logger.Info("GA engine initialized (lab/dev mode)")
}
