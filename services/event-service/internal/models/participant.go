package models

import (
	"time"

	"github.com/google/uuid"
)

type ParticipantStatus string

const (
	ParticipantStatusPending  ParticipantStatus = "pending"
	ParticipantStatusAccepted ParticipantStatus = "accepted"
	ParticipantStatusRejected ParticipantStatus = "rejected"
)

type EventParticipant struct {
	EventID  uuid.UUID         `json:"event_id" db:"event_id"`
	UserID   uuid.UUID         `json:"user_id" db:"user_id"`
	Status   ParticipantStatus `json:"status" db:"status"`
	JoinedAt time.Time         `json:"joined_at" db:"joined_at"`
}
