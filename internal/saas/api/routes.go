// Package api provides HTTP handlers and route setup for QuantSaaS.
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/chenthewho/ma-cross-strategy/internal/quant"
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

	// Health check (no auth)
	r.GET("/api/v1/health", handleHealth)

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
	api.GET("/instances/:id/price-chart", handlePriceChart(db))

	// Dashboard
	api.GET("/dashboard", handleDashboard(db))
	api.GET("/dashboard/equity-snapshots", handleEquitySnapshots(db))

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

	// ── CORS preflight catch-all via NoRoute ──
	r.NoRoute(func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
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

// instanceWithPortfolio merges StrategyInstance with PortfolioState fields for API output.
type instanceWithPortfolio struct {
	ID         uint      `json:"id"`
	UserID     uint      `json:"user_id"`
	TemplateID string    `json:"template_id"`
	Name       string    `json:"name"`
	Symbol     string    `json:"symbol"`
	Status     string    `json:"status"`
	ParamPack  string    `json:"param_pack"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	// Portfolio fields (joined from portfolio_states)
	TotalEquity    float64 `json:"total_equity"`
	CNYBalance     float64 `json:"cny_balance"`
	DeadHold       float64 `json:"dead_hold"`
	FloatHold      float64 `json:"float_hold"`
	ColdSealedHold float64 `json:"cold_sealed_hold"`
	InitialCapital float64 `json:"initial_capital"`
}

func handleListInstances(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		var results []instanceWithPortfolio
		db.Raw(`SELECT si.id, si.user_id, si.template_id, si.name, si.symbol,
			si.status, si.param_pack,
			COALESCE(si.created_at, now()) as created_at,
			COALESCE(si.updated_at, now()) as updated_at,
			COALESCE(ps.total_equity, 0) as total_equity,
			COALESCE(ps.cny_balance, 0) as cny_balance,
			COALESCE(ps.dead_hold, 0) as dead_hold,
			COALESCE(ps.float_hold, 0) as float_hold,
			COALESCE(ps.cold_sealed_hold, 0) as cold_sealed_hold,
			COALESCE(ps.initial_capital, 0) as initial_capital
		FROM strategy_instances si
		LEFT JOIN portfolio_states ps ON ps.instance_id = si.id
		WHERE si.user_id = ? AND si.status != ?
		ORDER BY si.id DESC`, userID, store.InstanceDeleted).Scan(&results)
		c.JSON(http.StatusOK, gin.H{"instances": results})
	}
}

func handleCreateInstance(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		var req struct {
			TemplateID        string  `json:"template_id" binding:"required"`
			Name              string  `json:"name" binding:"required"`
			Symbol            string  `json:"symbol"`
			InitialCapital    float64 `json:"initial_capital"`
			MonthlyInject     float64 `json:"monthly_inject"`
			MacroIntervalDays int     `json:"macro_interval_days"`
			ColdSealedAmount  float64 `json:"cold_sealed_amount"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.Symbol == "" {
			req.Symbol = "BTCUSDT"
		}
		if req.InitialCapital == 0 {
			req.InitialCapital = 100000
		}
		if req.MonthlyInject == 0 {
			req.MonthlyInject = 5000
		}
		if req.MacroIntervalDays == 0 {
			req.MacroIntervalDays = 30
		}

		// Build ParamPack with custom DCA interval
		chromo := quant.DefaultSeedChromosome
		chromo.MacroIntervalDays = req.MacroIntervalDays
		spawn := quant.DefaultSpawnPoint
		spawn.Policy.MonthlyInject = req.MonthlyInject
		spawn.Policy.InitialCapital = req.InitialCapital
		pp := quant.ParamPack{
			StrategyID: req.TemplateID,
			Chromosome: chromo,
			SpawnPoint: spawn,
		}
		ppBytes, _ := json.Marshal(pp)

		inst := store.StrategyInstance{
			UserID: userID, TemplateID: req.TemplateID,
			Name: req.Name, Symbol: req.Symbol, Status: store.InstanceStopped,
			ParamPack: string(ppBytes),
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
			InitialCapital:  req.InitialCapital,
			ColdSealedHold: req.ColdSealedAmount,
		}
		db.Create(&ps)
		c.JSON(http.StatusCreated, gin.H{"instance": inst})
	}
}

func handleStartInstance(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		res := db.Model(&store.StrategyInstance{}).Where("id = ? AND status = ?", id, store.InstanceStopped).
			Update("status", store.InstanceRunning)
		if res.RowsAffected == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "instance not found or already running"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "started"})
	}
}

func handleStopInstance(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		res := db.Model(&store.StrategyInstance{}).Where("id = ? AND status = ?", id, store.InstanceRunning).
			Update("status", store.InstanceStopped)
		if res.RowsAffected == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "instance not found or already stopped"})
			return
		}
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

func handlePriceChart(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

		// Get instance symbol
		var inst store.StrategyInstance
		if db.First(&inst, id).Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
			return
		}

		// Load K-line closes for the price line
		type KlinePoint struct {
			OpenTime int64   `json:"open_time"`
			Close    float64 `json:"close"`
		}
		var klines []KlinePoint
		db.Raw(`SELECT open_time, close FROM (
			SELECT open_time, close FROM k_lines 
			WHERE symbol = ? AND interval = '1h' 
			ORDER BY open_time DESC LIMIT 200
		) sub ORDER BY open_time ASC`, inst.Symbol).Scan(&klines)

		// Load trades for buy/sell markers
		type TradeMarker struct {
			CreatedAt string  `json:"created_at"`
			Price     float64 `json:"price"`
			Action    string  `json:"action"`
			Engine    string  `json:"engine"`
			Qty       float64 `json:"qty"`
		}
		var markers []TradeMarker
		db.Raw(`SELECT created_at, filled_price as price, action, engine, filled_qty as qty
			FROM trade_records 
			WHERE instance_id = ? 
			ORDER BY created_at ASC`, id).Scan(&markers)

		// Calculate average buy price
		type AvgResult struct{ AvgPrice float64 }
		var avg AvgResult
		db.Raw(`SELECT AVG(filled_price) as avg_price FROM trade_records 
			WHERE instance_id = ? AND action = 'BUY'`, id).Scan(&avg)

		c.JSON(http.StatusOK, gin.H{
			"klines":        klines,
			"trades":        markers,
			"avg_buy_price": avg.AvgPrice,
		})
	}
}

// ── Dashboard ─────────────────────────────────────────────────

func handleDashboard(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		var results []instanceWithPortfolio
		db.Raw(`SELECT si.id, si.user_id, si.template_id, si.name, si.symbol,
			si.status, si.param_pack,
			COALESCE(si.created_at, now()) as created_at,
			COALESCE(si.updated_at, now()) as updated_at,
			COALESCE(ps.total_equity, 0) as total_equity,
			COALESCE(ps.cny_balance, 0) as cny_balance,
			COALESCE(ps.dead_hold, 0) as dead_hold,
			COALESCE(ps.float_hold, 0) as float_hold,
			COALESCE(ps.float_units, 0) as float_units,
			COALESCE(ps.realized_pnl, 0) as realized_pnl,
			COALESCE(ps.cold_sealed_hold, 0) as cold_sealed_hold,
			COALESCE(ps.initial_capital, 0) as initial_capital
		FROM strategy_instances si
		LEFT JOIN portfolio_states ps ON ps.instance_id = si.id
		WHERE si.user_id = ? AND si.status != ?
		ORDER BY si.id DESC`, userID, store.InstanceDeleted).Scan(&results)
		c.JSON(http.StatusOK, gin.H{"instances": results})
	}
}

func handleEquitySnapshots(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		instanceID := c.Query("instance_id")
		if instanceID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "instance_id is required"})
			return
		}
		var snapshots []store.EquitySnapshot
		db.Where("instance_id = ?", instanceID).Order("recorded_at ASC").Limit(200).Find(&snapshots)
		c.JSON(http.StatusOK, gin.H{"snapshots": snapshots})
	}
}

// ── Agent ─────────────────────────────────────────────────────

func handleAgentStatus(hub *ws.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "connected": true})
	}
}

func handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
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
