// Package ws provides the WebSocket Hub for SaaS-to-Agent communication.
package ws

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/chenthewho/ma-cross-strategy/internal/saas/auth"
)

// Message types exchanged between SaaS Hub and Agents.
const (
	MsgTypeAuth        = "auth"
	MsgTypeAuthResult  = "auth_result"
	MsgTypeHeartbeat   = "heartbeat"
	MsgTypeDeltaReport = "delta_report"
	MsgTypeTradeCmd    = "trade_command"
)

// AuthTimeout is the maximum time to wait for an auth message after connect.
const AuthTimeout = 10 * time.Second

// WriteTimeout is the deadline for each websocket write.
const WriteTimeout = 5 * time.Second

// AgentConn represents a connected agent with its WebSocket and metadata.
type AgentConn struct {
	UserID   uint
	Conn     *websocket.Conn
	LastSeen time.Time
}

// Hub manages all connected agent WebSockets and routes messages.
type Hub struct {
	mu      sync.RWMutex
	agents  map[uint]*AgentConn // userID -> conn
	authSvc *auth.TokenService
}

// NewHub creates a new Hub backed by the given TokenService for JWT validation.
func NewHub(authSvc *auth.TokenService) *Hub {
	return &Hub{agents: make(map[uint]*AgentConn), authSvc: authSvc}
}

// SendToAgent sends an arbitrary JSON-marshalable command to the agent with the given userID.
func (h *Hub) SendToAgent(userID uint, cmd any) error {
	h.mu.RLock()
	defer h.mu.RUnlock()
	agent, ok := h.agents[userID]
	if !ok {
		return fmt.Errorf("agent not connected")
	}
	return agent.Conn.WriteJSON(cmd)
}

// SendCommand sends a trade command (as raw JSON bytes) to a specific agent.
func (h *Hub) SendCommand(userID uint, tradeCommand json.RawMessage) error {
	msg := map[string]any{
		"type":    MsgTypeTradeCmd,
		"payload": tradeCommand,
	}
	return h.SendToAgent(userID, msg)
}

// Register stores a new agent connection keyed by userID.
func (h *Hub) Register(userID uint, conn *websocket.Conn) {
	h.mu.Lock()
	h.agents[userID] = &AgentConn{UserID: userID, Conn: conn, LastSeen: time.Now()}
	h.mu.Unlock()
}

// Unregister removes the agent connection for the given userID and closes the socket.
func (h *Hub) Unregister(userID uint) {
	h.mu.Lock()
	agent, ok := h.agents[userID]
	if ok {
		delete(h.agents, userID)
	}
	h.mu.Unlock()

	if ok && agent != nil {
		agent.Conn.Close()
	}
}

// HandleConnection performs the full lifecycle of an agent WebSocket:
//  1. Wait up to AuthTimeout for an auth message containing a JWT.
//  2. Validate the JWT, register the connection, and reply with auth_result.
//  3. Enter a read loop dispatching heartbeat and delta_report messages.
//  4. On any error or disconnect, unregister the agent.
func (h *Hub) HandleConnection(conn *websocket.Conn) {
	defer conn.Close()

	// ---- Phase 1: authentication ----
	authMsg, err := h.readAuthMessage(conn)
	if err != nil {
		h.sendAuthResult(conn, false, 0, err.Error())
		return
	}

	claims, err := h.authSvc.ParseToken(authMsg.Token)
	if err != nil {
		h.sendAuthResult(conn, false, 0, fmt.Sprintf("invalid token: %v", err))
		return
	}

	h.Register(claims.UserID, conn)
	defer h.Unregister(claims.UserID)

	h.sendAuthResult(conn, true, claims.UserID, "authenticated")

	// ---- Phase 2: message loop ----
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return // client disconnected; defer Unregister handles cleanup
		}

		var msg map[string]json.RawMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue // skip malformed message
		}

		typeRaw, ok := msg["type"]
		if !ok {
			continue
		}
		var msgType string
		if err := json.Unmarshal(typeRaw, &msgType); err != nil {
			continue
		}

		switch msgType {
		case MsgTypeHeartbeat:
			h.mu.Lock()
			if agent, ok := h.agents[claims.UserID]; ok {
				agent.LastSeen = time.Now()
			}
			h.mu.Unlock()
			// Optionally respond with pong
			conn.SetWriteDeadline(time.Now().Add(WriteTimeout))
			conn.WriteJSON(map[string]string{"type": "heartbeat_ack"})

		case MsgTypeDeltaReport:
			// Delta reports are received and can be forwarded to the
			// instance manager or GA engine as needed by upper layers.
			// For now we simply acknowledge.
			conn.SetWriteDeadline(time.Now().Add(WriteTimeout))
			conn.WriteJSON(map[string]string{"type": "delta_ack"})

		default:
			// Unknown message type; silently ignore
		}
	}
}

// ---- internal helpers ----

type authMessage struct {
	Type  string `json:"type"`
	Token string `json:"token"`
}

// readAuthMessage blocks until a websocket text message is received or the
// AuthTimeout expires.  Only messages with type "auth" are accepted.
func (h *Hub) readAuthMessage(conn *websocket.Conn) (*authMessage, error) {
	conn.SetReadDeadline(time.Now().Add(AuthTimeout))

	_, raw, err := conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("read auth message: %w", err)
	}

	var msg authMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		return nil, fmt.Errorf("parse auth message: %w", err)
	}
	if msg.Type != MsgTypeAuth {
		return nil, fmt.Errorf("expected auth message, got %q", msg.Type)
	}
	if msg.Token == "" {
		return nil, fmt.Errorf("auth message missing token")
	}

	// Clear the read deadline after successful auth so the main loop doesn't time out.
	conn.SetReadDeadline(time.Time{})

	return &msg, nil
}

// sendAuthResult writes the authentication outcome back to the client.
func (h *Hub) sendAuthResult(conn *websocket.Conn, success bool, userID uint, message string) {
	conn.SetWriteDeadline(time.Now().Add(WriteTimeout))
	conn.WriteJSON(map[string]any{
		"type":    MsgTypeAuthResult,
		"success": success,
		"user_id": userID,
		"message": message,
	})
}
