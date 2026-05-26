// Package ws provides the LocalAgent WebSocket client for connecting to the SaaS.
// It handles login, connection lifecycle, message routing, and auto-reconnect
// with exponential backoff.
package ws

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/chenthewho/ma-cross-strategy/internal/agent/broker"
	"github.com/chenthewho/ma-cross-strategy/internal/agent/config"
)

const (
	// Heartbeat interval as defined in the system topology spec.
	heartbeatInterval = 30 * time.Second

	// Reconnect backoff parameters.
	initialBackoff = 1 * time.Second
	maxBackoff     = 5 * time.Minute
	backoffFactor  = 2.0

	// WebSocket endpoint path.
	wsAgentPath = "/ws/agent"

	// REST login endpoint path.
	loginPath = "/api/v1/auth/login"
)

// ── JSON message wrappers ────────────────────────────────────

// wsMessage is a generic JSON envelope for all WebSocket messages.
type wsMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`

	// Inline fields used by specific message types (denormalised for simplicity).
	Token         string           `json:"token,omitempty"`
	Success       *bool            `json:"success,omitempty"`
	Error         string           `json:"error,omitempty"`
	ClientOrderID string           `json:"client_order_id,omitempty"`
	Balances      []broker.Balance `json:"balances,omitempty"`
	Execution     *broker.Execution `json:"execution,omitempty"`

	// Command fields (from SaaS → Agent).
	Action    string  `json:"action,omitempty"`
	Engine    string  `json:"engine,omitempty"`
	Symbol    string  `json:"symbol,omitempty"`
	AmountCNY float64 `json:"amount_cny,omitempty"`
	QtyAsset  float64 `json:"qty_asset,omitempty"`
	LotType   string  `json:"lot_type,omitempty"`
}

// ── AgentClient ──────────────────────────────────────────────

// AgentClient manages a WebSocket connection to the SaaS server.
// It handles login, authentication, message routing, heartbeats,
// and automatic reconnection with exponential backoff.
type AgentClient struct {
	cfg    *config.AgentConfig
	conn   *websocket.Conn
	mu     sync.Mutex
	done   chan struct{}
	jwt    string
	logger *log.Logger
}

// NewAgentClient creates a new AgentClient with the given configuration.
func NewAgentClient(cfg *config.AgentConfig) *AgentClient {
	return &AgentClient{
		cfg:    cfg,
		done:   make(chan struct{}),
		logger: log.Default(),
	}
}

// Run starts the agent's main loop: login, connect, serve.
// Blocks until Close() is called or an unrecoverable error occurs.
func (c *AgentClient) Run() {
	for {
		select {
		case <-c.done:
			return
		default:
		}

		c.runSession()

		// Session ended — wait before reconnecting.
		c.waitReconnect()

		select {
		case <-c.done:
			return
		default:
		}
	}
}

// Close signals the agent to stop and closes the WebSocket connection.
func (c *AgentClient) Close() {
	select {
	case <-c.done:
		return
	default:
		close(c.done)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
}

// ── Session lifecycle ────────────────────────────────────────

// runSession performs a full login+connect+loop cycle.
func (c *AgentClient) runSession() {
	// 1. Login via REST to get JWT.
	token, err := c.login()
	if err != nil {
		c.logger.Printf("login failed: %v", err)
		return
	}
	c.jwt = token

	// 2. Establish WebSocket connection.
	conn, err := c.dial()
	if err != nil {
		c.logger.Printf("websocket dial failed: %v", err)
		return
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		if c.conn != nil {
			_ = c.conn.Close()
			c.conn = nil
		}
		c.mu.Unlock()
	}()

	// 3. Authenticate over WebSocket.
	if err := c.authenticate(); err != nil {
		c.logger.Printf("auth failed: %v", err)
		return
	}

	// 4. Send initial DeltaReport (balances only, no client_order_id).
	if err := c.sendInitialSnapshot(); err != nil {
		c.logger.Printf("initial snapshot failed: %v", err)
		return
	}

	// 5. Run the message loop.
	c.messageLoop()
}

// login calls POST /api/v1/auth/login and returns a JWT token.
func (c *AgentClient) login() (string, error) {
	loginURL := c.cfg.SaaSURL + loginPath

	body := map[string]string{
		"email":    c.cfg.Email,
		"password": c.cfg.Password,
	}
	bodyBytes, _ := json.Marshal(body)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("build login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login returned status %d", resp.StatusCode)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode login response: %w", err)
	}

	if result.Token == "" {
		return "", fmt.Errorf("login response missing token")
	}

	return result.Token, nil
}

// dial connects to the WebSocket endpoint.
func (c *AgentClient) dial() (*websocket.Conn, error) {
	u, err := url.Parse(c.cfg.SaaSURL)
	if err != nil {
		return nil, fmt.Errorf("parse saas_url: %w", err)
	}

	wsScheme := "ws"
	if u.Scheme == "https" {
		wsScheme = "wss"
	}

	wsURL := fmt.Sprintf("%s://%s%s", wsScheme, u.Host, wsAgentPath)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", wsURL, err)
	}

	return conn, nil
}

// authenticate sends the auth message and waits for auth_result.
func (c *AgentClient) authenticate() error {
	authMsg := wsMessage{
		Type:  "auth",
		Token: c.jwt,
	}

	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return fmt.Errorf("connection lost")
	}

	if err := conn.WriteJSON(authMsg); err != nil {
		return fmt.Errorf("write auth: %w", err)
	}

	// Read auth_result with a timeout.
	_ = conn.SetReadDeadline(time.Now().Add(15 * time.Second))

	_, raw, err := conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("read auth_result: %w", err)
	}

	var resp wsMessage
	if err := json.Unmarshal(raw, &resp); err != nil {
		return fmt.Errorf("unmarshal auth_result: %w", err)
	}

	if resp.Type != "auth_result" {
		return fmt.Errorf("expected auth_result, got %q", resp.Type)
	}
	if resp.Success == nil || !*resp.Success {
		errStr := resp.Error
		if errStr == "" {
			errStr = "auth rejected"
		}
		return fmt.Errorf("auth failed: %s", errStr)
	}

	return nil
}

// sendInitialSnapshot sends balances as a delta_report with no client_order_id.
func (c *AgentClient) sendInitialSnapshot() error {
	balances, err := broker.GetBalances(c.cfg.Broker)
	if err != nil {
		return fmt.Errorf("get balances: %w", err)
	}

	msg := wsMessage{
		Type:     "delta_report",
		Balances: balances,
		// No client_order_id and no execution for the initial snapshot.
	}

	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return fmt.Errorf("connection lost")
	}

	return conn.WriteJSON(msg)
}

// ── Message loop ─────────────────────────────────────────────

// messageLoop reads messages from the WebSocket and dispatches them.
func (c *AgentClient) messageLoop() {
	// Heartbeat ticker.
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	// Read messages in a goroutine, write heartbeats via ticker.
	errCh := make(chan error, 1)

	go func() {
		errCh <- c.readLoop()
	}()

	for {
		select {
		case <-c.done:
			return

		case err := <-errCh:
			if err != nil {
				c.logger.Printf("read loop ended: %v", err)
			}
			return

		case <-ticker.C:
			if err := c.sendHeartbeat(); err != nil {
				c.logger.Printf("heartbeat send failed: %v", err)
				return
			}
		}
	}
}

// readLoop continuously reads messages from the WebSocket.
func (c *AgentClient) readLoop() error {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return fmt.Errorf("no connection")
	}

	for {
		select {
		case <-c.done:
			return nil
		default:
		}

		// Reset read deadline for each message (no persistent deadline).
		conn.SetReadDeadline(time.Time{})

		_, raw, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("read message: %w", err)
		}

		var msg wsMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			c.logger.Printf("bad message: %v", err)
			continue
		}

		switch msg.Type {
		case "command":
			go c.handleCommand(msg)

		case "heartbeat_ack":
			// Ignore — heartbeat keep-alive is handled by the sender side.
			// No action needed.

		case "report_ack":
			// Acknowledged by SaaS — nothing to do on the agent side.

		case "auth_result":
			// Should not receive auth_result during normal operation; ignore.

		default:
			c.logger.Printf("unknown message type: %q", msg.Type)
		}
	}
}

// handleCommand processes a TradeCommand from the SaaS:
// 1. Send command_ack immediately.
// 2. Execute the order via broker.PlaceOrder in a goroutine.
// 3. Send delta_report with the execution details.
func (c *AgentClient) handleCommand(msg wsMessage) {
	clientOrderID := msg.ClientOrderID

	// 1. Send command_ack immediately (don't wait for execution).
	err := c.sendCommandAck(clientOrderID)
	if err != nil {
		c.logger.Printf("command_ack failed for %s: %v", clientOrderID, err)
		return
	}

	// 2. Execute via broker.
	cmd := broker.TradeCommand{
		ClientOrderID: clientOrderID,
		Action:        msg.Action,
		Engine:        msg.Engine,
		Symbol:        msg.Symbol,
		AmountCNY:     msg.AmountCNY,
		QtyAsset:      msg.QtyAsset,
		LotType:       msg.LotType,
	}

	exec, err := broker.PlaceOrder(cmd, c.cfg.Broker)
	if err != nil {
		c.logger.Printf("place order failed for %s: %v", clientOrderID, err)
		// Still attempt to report balances, even if execution failed.
	}

	// 3. Fetch latest balances and send delta_report.
	balances, balErr := broker.GetBalances(c.cfg.Broker)
	if balErr != nil {
		c.logger.Printf("get balances failed: %v", balErr)
		balances = nil
	}

	if exec == nil {
		exec = &broker.Execution{
			OrderID: clientOrderID,
			Status:  "failed",
		}
	}

	report := wsMessage{
		Type:          "delta_report",
		ClientOrderID: clientOrderID,
		Balances:      balances,
		Execution:     exec,
	}

	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		c.logger.Printf("connection lost, cannot send delta_report for %s", clientOrderID)
		return
	}

	if err := conn.WriteJSON(report); err != nil {
		c.logger.Printf("delta_report send failed for %s: %v", clientOrderID, err)
	}
}

// sendCommandAck sends an immediate acknowledgement for a received command.
func (c *AgentClient) sendCommandAck(clientOrderID string) error {
	msg := wsMessage{
		Type:          "command_ack",
		ClientOrderID: clientOrderID,
	}

	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return fmt.Errorf("connection lost")
	}

	return conn.WriteJSON(msg)
}

// sendHeartbeat sends a heartbeat message to the SaaS.
func (c *AgentClient) sendHeartbeat() error {
	msg := wsMessage{Type: "heartbeat"}

	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return fmt.Errorf("connection lost")
	}

	return conn.WriteJSON(msg)
}

// ── Reconnection ─────────────────────────────────────────────

// waitReconnect blocks with exponential backoff before the next reconnect attempt.
// Connection is handled by runSession(); this only manages the delay.
func (c *AgentClient) waitReconnect() {
	c.logger.Printf("disconnected, will reconnect with exponential backoff...")

	backoff := initialBackoff

	for {
		sleepDuration := backoff
		c.logger.Printf("reconnecting in %v...", sleepDuration)

		timer := time.NewTimer(sleepDuration)

		select {
		case <-c.done:
			timer.Stop()
			return
		case <-timer.C:
		}

		// Try to log in and connect.
		token, err := c.login()
		if err != nil {
			c.logger.Printf("reconnect login failed: %v", err)
			backoff = c.nextBackoff(backoff)
			continue
		}
		c.jwt = token

		conn, err := c.dial()
		if err != nil {
			c.logger.Printf("reconnect dial failed: %v", err)
			backoff = c.nextBackoff(backoff)
			continue
		}

		c.mu.Lock()
		c.conn = conn
		c.mu.Unlock()

		if err := c.authenticate(); err != nil {
			c.logger.Printf("reconnect auth failed: %v", err)
			c.mu.Lock()
			_ = c.conn.Close()
			c.conn = nil
			c.mu.Unlock()
			backoff = c.nextBackoff(backoff)
			continue
		}

		if err := c.sendInitialSnapshot(); err != nil {
			c.logger.Printf("reconnect initial snapshot failed: %v", err)
			c.mu.Lock()
			_ = c.conn.Close()
			c.conn = nil
			c.mu.Unlock()
			backoff = c.nextBackoff(backoff)
			continue
		}

		// Successfully reconnected — transition back to the message loop.
		// Close the connection so runSession() can do a clean login+dial cycle.
		_ = c.conn.Close()
		c.mu.Lock()
		c.conn = nil
		c.mu.Unlock()
		c.logger.Printf("reconnected successfully, resuming message loop")
		return
	}
}

func (c *AgentClient) nextBackoff(current time.Duration) time.Duration {
	next := time.Duration(math.Ceil(float64(current) * backoffFactor))
	if next > maxBackoff {
		return maxBackoff
	}
	return next
}
