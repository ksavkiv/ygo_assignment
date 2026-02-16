package destination

import (
	"encoding/json"
	"time"
)

type Destination struct {
	ID        int64           `json:"id"`
	City      string          `json:"city"`
	Country   string          `json:"country"`
	Latitude  float64         `json:"latitude"`
	Longitude float64         `json:"longitude"`
	Metadata  json.RawMessage `json:"metadata"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
