// Package marketdata fetches real-time K-line data from Binance public API.
package marketdata

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/chenthewho/ma-cross-strategy/internal/saas/store"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"
)

const (
	binanceKlineURL = "https://api.binance.com/api/v3/klines"
	symbol          = "BTCUSDT"
	interval        = "1h"
	limit           = 500
	refreshSec      = 30
)

// Service fetches and caches market data from Binance.
type Service struct {
	db     *store.DB
	logger *zap.Logger

	mu      sync.RWMutex
	lastC   float64 // latest close price
	running bool
	stopCh  chan struct{}
}

// New creates a market data service.
func New(db *store.DB, logger *zap.Logger) *Service {
	return &Service{db: db, logger: logger, stopCh: make(chan struct{})}
}

// LatestPrice returns the most recent close price (thread-safe).
func (s *Service) LatestPrice() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastC
}

// Start initializes K-line data and begins periodic refresh.
func (s *Service) Start() error {
	if err := s.fetchAndStore(); err != nil {
		return fmt.Errorf("initial fetch: %w", err)
	}
	s.running = true
	go s.refreshLoop()
	s.logger.Info("market data service started", zap.Float64("latest_price", s.lastC))
	return nil
}

// Stop shuts down the refresh loop.
func (s *Service) Stop() {
	if !s.running {
		return
	}
	close(s.stopCh)
	s.running = false
}

func (s *Service) refreshLoop() {
	ticker := time.NewTicker(refreshSec * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			if err := s.fetchAndStore(); err != nil {
				s.logger.Warn("market data refresh failed", zap.Error(err))
			}
		}
	}
}

func (s *Service) fetchAndStore() error {
	url := fmt.Sprintf("%s?symbol=%s&interval=%s&limit=%d", binanceKlineURL, symbol, interval, limit)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	var raw [][]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	if len(raw) == 0 {
		return fmt.Errorf("empty response")
	}

	// Parse bars
	bars := make([]store.KLine, 0, len(raw))
	for _, r := range raw {
		toFloat := func(v any) float64 {
			switch val := v.(type) {
			case float64:
				return val
			case string:
				var f float64
				fmt.Sscanf(val, "%f", &f)
				return f
			}
			return 0
		}
		bar := store.KLine{
			Symbol:   symbol,
			Interval: interval,
			OpenTime: int64(toFloat(r[0])),
			Open:     toFloat(r[1]),
			High:     toFloat(r[2]),
			Low:      toFloat(r[3]),
			Close:    toFloat(r[4]),
			Volume:   toFloat(r[5]),
		}
		bars = append(bars, bar)
	}

	// Upsert into DB (on conflict update latest values)
	for _, bar := range bars {
		s.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "symbol"}, {Name: "interval"}, {Name: "open_time"}},
			DoUpdates: clause.AssignmentColumns([]string{"open", "high", "low", "close", "volume"}),
		}).Create(&bar)
	}

	// Update cached latest price
	latest := bars[len(bars)-1]
	s.mu.Lock()
	s.lastC = latest.Close
	s.mu.Unlock()

	s.logger.Debug("market data refreshed",
		zap.Int("bars", len(bars)),
		zap.Float64("latest", latest.Close),
		zap.Time("bar_time", time.UnixMilli(latest.OpenTime)),
	)
	return nil
}
