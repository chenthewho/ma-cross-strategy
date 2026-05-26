// Package main is the Agent entry point.
// The agent connects to the SaaS server via WebSocket, handles trade commands,
// and reports execution results and balances back to the SaaS platform.
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	agentconfig "github.com/chenthewho/ma-cross-strategy/internal/agent/config"
	agentws "github.com/chenthewho/ma-cross-strategy/internal/agent/ws"
)

func main() {
	// ── 1. Load config.agent.yaml ─────────────────────────────
	configPath := "config.agent.yaml"
	if v := os.Getenv("AGENT_CONFIG_PATH"); v != "" {
		configPath = v
	}

	cfg, err := agentconfig.LoadAgentConfig(configPath)
	if err != nil {
		log.Fatalf("failed to load agent config: %v", err)
	}

	// ── 2. Init logger ────────────────────────────────────────
	logger := newAgentLogger()
	logger.Info("agent starting", zap.String("saas_url", cfg.SaaSURL))

	// ── 3. Init agent client ──────────────────────────────────
	client := agentws.NewAgentClient(cfg)

	// ── 4. Start agent main loop (login → connect → serve) ────
	go client.Run()

	// ── 5. Graceful exit on SIGTERM ───────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.Info("shutting down", zap.String("signal", sig.String()))
	client.Close()
	logger.Info("agent exited cleanly")
}

// newAgentLogger creates a zap development logger for the agent.
func newAgentLogger() *zap.Logger {
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, err := cfg.Build()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	return logger
}
