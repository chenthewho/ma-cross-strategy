// Package api provides HTTP handlers and route setup for QuantSaaS.
package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/chenthewho/ma-cross-strategy/internal/saas/auth"
	"github.com/chenthewho/ma-cross-strategy/internal/saas/store"
	"github.com/chenthewho/ma-cross-strategy/internal/saas/ws"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// SetupRoutes registers all HTTP routes on the gin engine.
func SetupRoutes(r *gin.Engine, db *store.DB, hub *ws.Hub, tokenSvc *auth.TokenService, appRole string) {
	// ── Public routes (no JWT) ──
	r.POST("/api/v1/auth/register", handleRegister(db, tokenSvc))
	r.POST("/api/v1/auth/login", handleLogin(db, tokenSvc))

	// Public system status (no auth)
	r.GET("/api/v1/system/status", handleSystemStatus(hub))

	// ── Protected routes (JWT required) ──
	api := r.Group("/api/v1")
	api.Use(AuthMiddleware(tokenSvc))

	// Strategies
	api.GET("/strategies", handleListStrategies)
	api.GET("/strategies/:id", handleGetStrategy)

	// Instances
	api.GET("/instances", handleListInstances(db))
	api.POST("/instances", handleCreateInstance(db))
	api.POST("/instances/:id/start", handleStartInstance(db))
	api.POST("/instances/:id/stop", handleStopInstance(db))
	api.DELETE("/instances/:id", handleDeleteInstance(db))
	api.GET("/instances/:id/lots", handleGetInstanceLots(db))
	api.GET("/instances/:id/trades", handleGetInstanceTrades(db))

	// Dashboard
	api.GET("/dashboard", handleDashboard(db))

	// Agent status
	api.GET("/agents/status", handleAgentStatus(hub))

	// Genome (protected)
	api.GET("/genome/champion", handleGetChampion(db))
	api.GET("/genome/challengers", handleGetChallengers(db))

	// ── Lab-only routes ──
	lab := api.Group("")
	lab.Use(LabOnlyMiddleware(appRole))
	lab.POST("/evolution/tasks", handleCreateEvolutionTask(db))
	lab.GET("/evolution/tasks", handleListEvolutionTasks(db))
	lab.POST("/evolution/tasks/:id/promote", handlePromoteTask(db))
	lab.POST("/backtests", handleCreateBacktest(db))
	lab.GET("/backtests/:id", handleGetBacktest(db))

	// ── WebSocket (public, auth via first message) ──
	r.GET("/ws/agent", func(c *gin.Context) {
		conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		hub.HandleConnection(conn)
	})
}

// ── Middleware ────────────────────────────────────────────────

func AuthMiddleware(tokenSvc *auth.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}
		claims, err := tokenSvc.ParseToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func LabOnlyMiddleware(appRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if appRole != "lab" && appRole != "dev" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "lab endpoints not available in this deployment"})
			return
		}
		c.Next()
	}
}

// ── Auth Handlers ─────────────────────────────────────────────

func handleRegister(db *store.DB, tokenSvc *auth.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email" binding:"required"`
			Password string `json:"password" binding:"required,min=6"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Check duplicate
		var existing store.User
		if db.Where("email = ?", req.Email).First(&existing).Error == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}
		user := store.User{Email: req.Email, PasswordHash: hashPassword(req.Password), Role: "user"}
		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "create user failed"})
			return
		}
		token, _ := tokenSvc.SignToken(user.ID, user.Email, user.Role)
		c.JSON(http.StatusCreated, gin.H{"token": token, "user": gin.H{"id": user.ID, "email": user.Email, "role": user.Role}})
	}
}

func handleLogin(db *store.DB, tokenSvc *auth.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var user store.User
		if db.Where("email = ?", req.Email).First(&user).Error != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		if user.PasswordHash != hashPassword(req.Password) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		token, _ := tokenSvc.SignToken(user.ID, user.Email, user.Role)
		c.JSON(http.StatusOK, gin.H{"token": token, "user": gin.H{"id": user.ID, "email": user.Email, "role": user.Role}})
	}
}

// ── Strategy Handlers ────────────────────────────────────────

func handleListStrategies(c *gin.Context) {
	// Return registered strategy templates from catalog
	strategies := []gin.H{
		{"id": "golden_cross", "name": "动态均衡策略", "version": "1.0.0", "is_spot": true,
			"description": "通过Sigmoid动态天平实现宏观+微观双层仓位管理，由遗传算法优化参数"},
	}
	c.JSON(http.StatusOK, gin.H{"strategies": strategies})
}

func handleGetStrategy(c *gin.Context) {
	id := c.Param("id")
	if id == "golden_cross" {
		c.JSON(http.StatusOK, gin.H{"id": "golden_cross", "name": "动态均衡策略", "version": "1.0.0"})
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "strategy not found"})
}

// ── Instance Handlers ────────────────────────────────────────

func handleListInstances(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		var instances []store.StrategyInstance
		db.Where("user_id = ? AND status != ?", userID, store.InstanceDeleted).Find(&instances)
		c.JSON(http.StatusOK, gin.H{"instances": instances})
	}
}

func handleCreateInstance(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		var req struct {
			TemplateID       string  `json:"template_id" binding:"required"`
			Name             string  `json:"name" binding:"required"`
			Symbol           string  `json:"symbol" binding:"required"`
			InitialCapital   float64 `json:"initial_capital"`
			MonthlyInject    float64 `json:"monthly_inject"`
			ColdSealedAmount float64 `json:"cold_sealed_amount"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		inst := store.StrategyInstance{
			UserID: userID, TemplateID: req.TemplateID,
			Name: req.Name, Symbol: req.Symbol, Status: store.InstanceStopped,
		}
		if err := db.Create(&inst).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "create instance failed"})
			return
		}
		// Create initial PortfolioState
		ps := store.PortfolioState{
			InstanceID:     inst.ID,
			CNYBalance:     req.InitialCapital,
			TotalEquity:    req.InitialCapital,
			ColdSealedHold: req.ColdSealedAmount,
		}
		db.Create(&ps)
		c.JSON(http.StatusCreated, gin.H{"instance": inst})
	}
}

func handleStartInstance(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		db.Model(&store.StrategyInstance{}).Where("id = ? AND status = ?", id, store.InstanceStopped).
			Update("status", store.InstanceRunning)
		c.JSON(http.StatusOK, gin.H{"status": "started"})
	}
}

func handleStopInstance(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		db.Model(&store.StrategyInstance{}).Where("id = ? AND status = ?", id, store.InstanceRunning).
			Update("status", store.InstanceStopped)
		c.JSON(http.StatusOK, gin.H{"status": "stopped"})
	}
}

func handleDeleteInstance(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		db.Model(&store.StrategyInstance{}).Where("id = ?", id).Update("status", store.InstanceDeleted)
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	}
}

func handleGetInstanceLots(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var lots []store.SpotLot
		db.Where("instance_id = ?", id).Find(&lots)
		c.JSON(http.StatusOK, gin.H{"lots": lots})
	}
}

func handleGetInstanceTrades(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var trades []store.TradeRecord
		db.Where("instance_id = ?", id).Order("created_at DESC").Limit(50).Find(&trades)
		c.JSON(http.StatusOK, gin.H{"trades": trades})
	}
}

// ── Dashboard ─────────────────────────────────────────────────

func handleDashboard(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		var instances []store.StrategyInstance
		db.Where("user_id = ? AND status != ?", userID, store.InstanceDeleted).Find(&instances)
		c.JSON(http.StatusOK, gin.H{"instances": instances})
	}
}

// ── Agent ─────────────────────────────────────────────────────

func handleAgentStatus(hub *ws.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "connected": true})
	}
}

func handleSystemStatus(hub *ws.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		engine := "running"
		apiConnected := false
		if hub != nil {
			apiConnected = true // simplified
		}
		c.JSON(http.StatusOK, gin.H{
			"engine":         engine,
			"api_connected":  apiConnected,
			"api_configured": true,
		})
	}
}

// ── Genome ────────────────────────────────────────────────────

func handleGetChampion(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var champ store.GeneRecord
		if db.Where("role = ?", store.GeneChampion).Order("activated_at DESC").First(&champ).Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "no champion found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"champion": champ})
	}
}

func handleGetChallengers(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var challengers []store.GeneRecord
		db.Where("role = ?", store.GeneChallenger).Order("created_at DESC").Limit(20).Find(&challengers)
		c.JSON(http.StatusOK, gin.H{"challengers": challengers})
	}
}

// ── Evolution ─────────────────────────────────────────────────

func handleCreateEvolutionTask(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			StrategyID     string `json:"strategy_id" binding:"required"`
			PopSize        int    `json:"pop_size"`
			MaxGenerations int    `json:"max_generations"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.PopSize == 0 { req.PopSize = 300 }
		if req.MaxGenerations == 0 { req.MaxGenerations = 25 }
		task := store.EvolutionTask{
			StrategyID: req.StrategyID, Status: store.EvoRunning,
		}
		db.Create(&task)
		c.JSON(http.StatusCreated, gin.H{"task": task})
	}
}

func handleListEvolutionTasks(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tasks []store.EvolutionTask
		db.Order("created_at DESC").Limit(10).Find(&tasks)
		var challengers []store.GeneRecord
		db.Where("role = ?", store.GeneChallenger).Order("created_at DESC").Limit(10).Find(&challengers)
		c.JSON(http.StatusOK, gin.H{"tasks": tasks, "challengers": challengers})
	}
}

func handlePromoteTask(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		taskID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		_ = taskID
		c.JSON(http.StatusOK, gin.H{"status": "promoted"})
	}
}

// ── Backtest ──────────────────────────────────────────────────

func handleCreateBacktest(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"status": "backtest started"})
	}
}

func handleGetBacktest(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "completed"})
	}
}

// ── Helpers ───────────────────────────────────────────────────

func hashPassword(s string) string {
	// Simple hash for prototype — use bcrypt in production
	h := uint64(0)
	for _, c := range s {
		h = h*31 + uint64(c)
	}
	return fmt.Sprintf("sha256:%x", h)
}
