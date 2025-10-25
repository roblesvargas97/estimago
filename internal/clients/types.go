package clients

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Client struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Phone     *string    `json:"phone,omitempty"`
	Email     *string    `email,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type CreateClientInput struct {
	Name  string           `json:"name"`
	Email *string          `json:"email,omitempty"`
	Phone *string          `json:"phone,omitempty"`
	Meta  *json.RawMessage `json:"meta,omitempty"`
}
