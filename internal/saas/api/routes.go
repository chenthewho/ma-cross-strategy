// Package api provides HTTP handlers and route setup for QuantSaaS.
package api

import (
	"net/http"
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
	// Public
	r.POST("/api/v1/auth/register", handleRegister(db))
	r.POST("/api/v1/auth/login", handleLogin(db, tokenSvc))

	// Protected (JWT middleware)
	api := r.Group("/api/v1")
	api.Use(AuthMiddleware(tokenSvc))

	api.GET("/strategies", handleListStrategies)
	api.GET("/instances", handleListInstances(db))
	api.POST("/instances", handleCreateInstance(db))
	api.POST("/instances/:id/start", handleStartInstance(db))
	api.POST("/instances/:id/stop", handleStopInstance(db))
	api.DELETE("/instances/:id", handleDeleteInstance(db))
	api.GET("/dashboard", handleDashboard(db))
	api.GET("/agents/status", handleAgentStatus(hub))

	// Genome (protected via JWT)
	api.GET("/genome/champion", handleGetChampion(db))

	// Lab-only routes
	lab := api.Group("")
	lab.Use(LabOnlyMiddleware(appRole))
	lab.POST("/evolution/tasks", handleCreateEvolutionTask(db))
	lab.GET("/evolution/tasks", handleListEvolutionTasks(db))
	lab.POST("/evolution/tasks/:id/promote", handlePromoteTask(db))

	// WebSocket (public, auth via first message)
	r.GET("/ws/agent", func(c *gin.Context) {
		conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		hub.HandleConnection(conn)
	})
}

// ── Middleware ────────────────────────────────────────────────

// AuthMiddleware extracts and validates the JWT from the Authorization header.
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

// LabOnlyMiddleware blocks requests if app_role is not "lab" or "dev".
func LabOnlyMiddleware(appRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if appRole != "lab" && appRole != "dev" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "lab endpoints not available in this deployment"})
			return
		}
		c.Next()
	}
}

// ── Placeholder Handlers ─────────────────────────────────────

func handleRegister(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"status": "ok", "message": "register placeholder"})
	}
}

func handleLogin(db *store.DB, tokenSvc *auth.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "login placeholder"})
	}
}

func handleListStrategies(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "strategies": []any{}})
}

func handleListInstances(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "instances": []any{}})
	}
}

func handleCreateInstance(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"status": "ok", "message": "create instance placeholder"})
	}
}

func handleStartInstance(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "start instance placeholder"})
	}
}

func handleStopInstance(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "stop instance placeholder"})
	}
}

func handleDeleteInstance(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "delete instance placeholder"})
	}
}

func handleDashboard(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "dashboard placeholder"})
	}
}

func handleAgentStatus(hub *ws.Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "agent status placeholder"})
	}
}

func handleGetChampion(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "champion placeholder"})
	}
}

func handleCreateEvolutionTask(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"status": "ok", "message": "create evolution task placeholder"})
	}
}

func handleListEvolutionTasks(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "tasks": []any{}, "challengers": []any{}})
	}
}

func handlePromoteTask(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "promote task placeholder"})
	}
}
