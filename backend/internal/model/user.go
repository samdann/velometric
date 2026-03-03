package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	FTP       *int      `json:"ftp,omitempty"`
	MaxHR     *int      `json:"max_hr,omitempty"`
	Weight    *float64  `json:"weight,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type HRZone struct {
	ZoneNumber    int      `json:"zone_number"`
	Name          string   `json:"name"`
	MinPercentage float64  `json:"min_percentage"`
	MaxPercentage *float64 `json:"max_percentage"`
	Color         string   `json:"color"`
}

type PowerZone struct {
	ZoneNumber    int      `json:"zone_number"`
	Name          string   `json:"name"`
	MinPercentage float64  `json:"min_percentage"`
	MaxPercentage *float64 `json:"max_percentage"`
	Color         string   `json:"color"`
}
