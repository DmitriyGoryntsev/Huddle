package models

import (
	"time"

	"github.com/google/uuid"
)

type EventStatus string

const (
	EventStatusOpen      EventStatus = "open"
	EventStatusFull      EventStatus = "full"
	EventStatusStarted   EventStatus = "started"
	EventStatusFinished  EventStatus = "finished"
	EventStatusCancelled EventStatus = "cancelled"
)

type Event struct {
	ID          uuid.UUID `json:"id" db:"id"`
	CreatorID   uuid.UUID `json:"creator_id" db:"creator_id"`
	CategoryID  int       `json:"category_id" db:"category_id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`

	// Координаты для API
	Latitude  float64 `json:"lat" db:"lat"`
	Longitude float64 `json:"lon" db:"lon"`

	StartTime           time.Time `json:"start_time" db:"start_time"`
	MaxParticipants     int       `json:"max_participants" db:"max_participants"`
	CurrentParticipants int       `json:"current_participants" db:"current_participants"` // Вычисляемое поле
	Price               float64   `json:"price" db:"price"`

	RequiresApproval bool        `json:"requires_approval" db:"requires_approval"`
	Status           EventStatus `json:"status" db:"status"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
