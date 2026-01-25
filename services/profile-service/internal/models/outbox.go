package models

import (
	"time"

	"github.com/google/uuid"
)

type OutboxEvent struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	EventType   string     `db:"event_type" json:"event_type"`
	Payload     []byte     `db:"payload" json:"-"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	PublishedAt *time.Time `db:"published_at" json:"published_at"`
	Attempts    int        `db:"attempts" json:"attempts"`
	Status      string     `db:"status" json:"status"`
}
