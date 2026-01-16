package repository

import (
	"context"
	"fmt"

	"auth-service/internal/models"
	"auth-service/pkg/db/postgres"

	"github.com/google/uuid"
)

type OutboxRepository interface {
	InsertTx(ctx context.Context, tx Tx, event *models.OutboxEvent) error
	GetPendingEvents(ctx context.Context, limit int) ([]*models.OutboxEvent, error)
	MarkAsPublished(ctx context.Context, eventID uuid.UUID) error
}

type outboxRepository struct {
	db *postgres.DB
}

func NewOutboxRepository(db *postgres.DB) OutboxRepository {
	return &outboxRepository{db: db}
}

func (r *outboxRepository) InsertTx(ctx context.Context, tx Tx, event *models.OutboxEvent) error {
	pgxTx, ok := tx.(*pgxTx)
	if !ok {
		return fmt.Errorf("invalid transaction type")
	}

	query := `
        INSERT INTO auth.outbox (id, event_type, payload) 
        VALUES ($1, $2, $3)
    `
	return pgxTx.Exec(ctx, query, event.ID, event.EventType, event.Payload)
}

func (r *outboxRepository) GetPendingEvents(ctx context.Context, limit int) ([]*models.OutboxEvent, error) {
	query := `
        SELECT id, event_type, payload, created_at, attempts 
        FROM auth.outbox 
        WHERE status = 'pending' 
        ORDER BY created_at ASC 
        LIMIT $1
        FOR UPDATE SKIP LOCKED
    `

	rows, err := r.db.Pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending events: %w", err)
	}
	defer rows.Close()

	var events []*models.OutboxEvent
	for rows.Next() {
		var event models.OutboxEvent
		if err := rows.Scan(
			&event.ID, &event.EventType, &event.Payload,
			&event.CreatedAt, &event.Attempts,
		); err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return events, nil
}

func (r *outboxRepository) MarkAsPublished(ctx context.Context, eventID uuid.UUID) error {
	query := `
        UPDATE auth.outbox 
        SET status = 'published', 
            published_at = NOW(), 
            attempts = attempts + 1 
        WHERE id = $1 AND status = 'pending'
    `
	_, err := r.db.Pool.Exec(ctx, query, eventID)
	return err
}
