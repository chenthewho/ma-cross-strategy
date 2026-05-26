// Package broker provides a mock broker API client for the LocalAgent.
// In production, this package is replaced with a real broker SDK adapter.
package broker

import "github.com/chenthewho/ma-cross-strategy/internal/agent/config"

// Execution represents the result of a placed order.
type Execution struct {
	OrderID     string  `json:"order_id"`
	FilledQty   float64 `json:"filled_qty"`
	FilledPrice float64 `json:"filled_price"`
	Fee         float64 `json:"fee"`
	Status      string  `json:"status"`
}

// Balance represents a single asset balance entry.
type Balance struct {
	Asset  string  `json:"asset"`
	Free   float64 `json:"free"`
	Frozen float64 `json:"frozen"`
}

// TradeCommand mirrors the SaaS-side trade command fields.
type TradeCommand struct {
	ClientOrderID string  `json:"client_order_id"`
	Action        string  `json:"action"` // BUY | SELL
	Engine        string  `json:"engine"` // MACRO | MICRO
	Symbol        string  `json:"symbol"`
	AmountCNY     float64 `json:"amount_cny"`
	QtyAsset      float64 `json:"qty_asset"`
	LotType       string  `json:"lot_type"` // DEAD_STACK | FLOATING
}

// PlaceOrder simulates calling the broker API. Returns a mock Execution.
func PlaceOrder(cmd TradeCommand, _ config.BrokerConfig) (*Execution, error) {
	return &Execution{
		OrderID:     cmd.ClientOrderID,
		FilledQty:   1,
		FilledPrice: 1,
		Fee:         0,
		Status:      "filled",
	}, nil
}

// GetBalances simulates calling the broker API to fetch current balances.
func GetBalances(_ config.BrokerConfig) ([]Balance, error) {
	return []Balance{
		{Asset: "CNY", Free: 100000, Frozen: 0},
	}, nil
}
