package destination

import "time"

type Destination struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Country     string    `json:"country"`
	Description string    `json:"description"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
