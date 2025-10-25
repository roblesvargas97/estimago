package quotes

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type QuoteItem struct {
	Kind      string   `json:"kind"`
	Name      string   `json:"name"`
	Qty       float64  `json:"qty"`
	Unit      string   `json:"unit"`
	UnitPrice float64  `json:"unit_price"`
	LineTotal *float64 `json:"line_total,omitempty"`
}

type CreateQuoteIn struct {
	ClientID   *uuid.UUID  `json:"client_id"`
	Items      []QuoteItem `json:"items"`
	LaborHours float64     `json:"labor_hours"`
	LaborRate  float64     `json:"labor_rate"`
	MarginPct  float64     `json:"margin_pct"`
	TaxPct     float64     `json:"tax_pct"`
	Currency   string      `json:"currency"`
	Notes      *string     `json:"notes"`
}

type Quote struct {
	ID         uuid.UUID       `json:"id"`
	ClientID   *uuid.UUID      `json:"client_id"`
	Items      json.RawMessage `json:"items"`
	LaborHours float64         `json:"labor_hours"`
	LaborRate  float64         `json:"labor_rate"`
	MarginPct  float64         `json:"margin_pct"`
	TaxPct     string          `json:"tax_pct"`
	Subtotal   float64         `json:"subtotal"`
	Total      float64         `json:"total"`
	Currency   string          `json:"currency"`
	Notes      *string         `json:"notes"`
	PublicID   *string         `json:"public_id"`
	Status     string          `json:"status"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}
