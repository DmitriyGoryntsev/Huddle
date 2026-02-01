package models

import (
	"time"
)

// CreateEventRequest - то, что мы ждем от фронтенда
type CreateEventRequest struct {
	CategoryID       int       `json:"category_id" validate:"required"`
	Title            string    `json:"title" validate:"required,min=3,max=100"`
	Description      string    `json:"description"`
	Latitude         float64   `json:"lat" validate:"required"`
	Longitude        float64   `json:"lon" validate:"required"`
	StartTime        time.Time `json:"start_time" validate:"required"`
	MaxParticipants  int       `json:"max_participants" validate:"required,min=2"`
	Price            float64   `json:"price"`
	RequiresApproval bool      `json:"requires_approval"`
}

// EventFilter - параметры для поиска на карте
type EventFilter struct {
	Latitude     float64 `query:"lat"`
	Longitude    float64 `query:"lon"`
	RadiusMeters float64 `query:"radius"`
	CategorySlug string  `query:"category"`
}

// UpdateParticipantStatusRequest - для аппрува участника
type UpdateParticipantStatusRequest struct {
	Status ParticipantStatus `json:"status" validate:"required,oneof=accepted rejected"`
}
