package repository

import (
	"context"
	"errors"
	"event-service/internal/models"
	"event-service/pkg/db/postgres"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// EventRepository — интерфейс для работы с событиями
type EventRepository interface {
	Create(ctx context.Context, event *models.Event) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Event, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Event, error)
	ListNearby(ctx context.Context, filter models.EventFilter) ([]*models.Event, error)
	Delete(ctx context.Context, eventID, creatorID uuid.UUID) error
	UpdateStatus(ctx context.Context, eventID uuid.UUID, status models.EventStatus) error
	MarkExpiredAsFinished(ctx context.Context) (int64, error)

	// Участники
	AddParticipant(ctx context.Context, eventID, userID uuid.UUID, status models.ParticipantStatus) error
	GetParticipant(ctx context.Context, eventID, userID uuid.UUID) (*models.EventParticipant, error)
	UpdateParticipantStatus(ctx context.Context, eventID, userID uuid.UUID, status models.ParticipantStatus) error
	RemoveParticipant(ctx context.Context, eventID, userID uuid.UUID) error
	CountAcceptedParticipants(ctx context.Context, eventID uuid.UUID) (int, error)
	ListParticipants(ctx context.Context, eventID uuid.UUID) ([]models.EventParticipant, error)
	GetUserEventIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)

	// Вспомогательные
	CategoryExists(ctx context.Context, categoryID int) (bool, error)
}

type eventRepository struct {
	db     *postgres.DB
	logger *zap.SugaredLogger
}

func NewEventRepository(db *postgres.DB, logger *zap.SugaredLogger) EventRepository {
	return &eventRepository{db: db, logger: logger}
}

// Create сохраняет событие. location в PostGIS: POINT(lon, lat)
func (r *eventRepository) Create(ctx context.Context, event *models.Event) error {
	query := `
		INSERT INTO events (creator_id, category_id, title, description, location, start_time, max_participants, price, requires_approval, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, ST_SetSRID(ST_MakePoint($5, $6), 4326)::geography, $7, $8, $9, $10, $11, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query,
		event.CreatorID, event.CategoryID, event.Title, event.Description,
		event.Longitude, event.Latitude, // PostGIS: lon, lat
		event.StartTime, event.MaxParticipants, event.Price,
		event.RequiresApproval, event.Status,
	).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt)
	if err != nil {
		r.logger.Errorw("Failed to create event", "creator_id", event.CreatorID, "error", err)
		return fmt.Errorf("failed to create event: %w", err)
	}

	// Creator автоматически первый участник (accepted)
	if err := r.AddParticipant(ctx, event.ID, event.CreatorID, models.ParticipantStatusAccepted); err != nil {
		r.logger.Errorw("Failed to add creator as participant", "event_id", event.ID, "error", err)
		return fmt.Errorf("failed to add creator: %w", err)
	}

	return nil
}

func (r *eventRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Event, error) {
	query := `
		SELECT id, creator_id, category_id, title, description,
		       ST_Y(location::geometry) as lat, ST_X(location::geometry) as lon,
		       start_time, max_participants, price, requires_approval, status, created_at, updated_at
		FROM events
		WHERE id = $1
	`
	event := &models.Event{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&event.ID, &event.CreatorID, &event.CategoryID, &event.Title, &event.Description,
		&event.Latitude, &event.Longitude,
		&event.StartTime, &event.MaxParticipants, &event.Price,
		&event.RequiresApproval, &event.Status, &event.CreatedAt, &event.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		r.logger.Errorw("Failed to get event", "event_id", id, "error", err)
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	count, _ := r.CountAcceptedParticipants(ctx, id)
	event.CurrentParticipants = count
	return event, nil
}

func (r *eventRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Event, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	query := `
		SELECT id, creator_id, category_id, title, description,
		       ST_Y(location::geometry) as lat, ST_X(location::geometry) as lon,
		       start_time, max_participants, price, requires_approval, status, created_at, updated_at
		FROM events WHERE id = ANY($1)
	`
	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("get events by ids: %w", err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		e := &models.Event{}
		if err := rows.Scan(
			&e.ID, &e.CreatorID, &e.CategoryID, &e.Title, &e.Description,
			&e.Latitude, &e.Longitude,
			&e.StartTime, &e.MaxParticipants, &e.Price,
			&e.RequiresApproval, &e.Status, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, err
		}
		count, _ := r.CountAcceptedParticipants(ctx, e.ID)
		e.CurrentParticipants = count
		events = append(events, e)
	}
	return events, rows.Err()
}

func (r *eventRepository) ListNearby(ctx context.Context, filter models.EventFilter) ([]*models.Event, error) {
	// Если не задан радиус — используем дефолтный (например 5 км)
	radiusM := filter.RadiusMeters
	if radiusM <= 0 {
		radiusM = 5000
	}

	// Только активные: open, не истекшие. Гео-фильтр — только если заданы координаты (не 0,0)
	query := `
		SELECT id, creator_id, category_id, title, description,
		       ST_Y(location::geometry) as lat, ST_X(location::geometry) as lon,
		       start_time, max_participants, price, requires_approval, status, created_at, updated_at
		FROM events
		WHERE status = 'open'
		  AND start_time > NOW()
		  AND (($1 = 0 AND $2 = 0) OR ST_DWithin(location, ST_SetSRID(ST_MakePoint($2, $1), 4326)::geography, $3))
		  AND ($4 = '' OR category_id IN (SELECT id FROM categories WHERE slug = $4))
		ORDER BY start_time ASC
	`
	rows, err := r.db.Query(ctx, query, filter.Latitude, filter.Longitude, radiusM, filter.CategorySlug)
	if err != nil {
		r.logger.Errorw("Failed to list events", "error", err)
		return nil, fmt.Errorf("failed to list events: %w", err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		e := &models.Event{}
		err := rows.Scan(
			&e.ID, &e.CreatorID, &e.CategoryID, &e.Title, &e.Description,
			&e.Latitude, &e.Longitude,
			&e.StartTime, &e.MaxParticipants, &e.Price,
			&e.RequiresApproval, &e.Status, &e.CreatedAt, &e.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		count, _ := r.CountAcceptedParticipants(ctx, e.ID)
		e.CurrentParticipants = count
		events = append(events, e)
	}
	return events, rows.Err()
}

func (r *eventRepository) Delete(ctx context.Context, eventID, creatorID uuid.UUID) error {
	result, err := r.db.Pool.Exec(ctx, `DELETE FROM events WHERE id = $1 AND creator_id = $2`, eventID, creatorID)
	if err != nil {
		r.logger.Errorw("Failed to delete event", "event_id", eventID, "error", err)
		return fmt.Errorf("failed to delete event: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("event not found or not authorized to delete")
	}
	return nil
}

func (r *eventRepository) UpdateStatus(ctx context.Context, eventID uuid.UUID, status models.EventStatus) error {
	err := r.db.Exec(ctx, `UPDATE events SET status = $1, updated_at = NOW() WHERE id = $2`, status, eventID)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	return nil
}

func (r *eventRepository) MarkExpiredAsFinished(ctx context.Context) (int64, error) {
	result, err := r.db.Pool.Exec(ctx,
		`UPDATE events SET status = 'finished', updated_at = NOW() WHERE status IN ('open', 'full') AND start_time < NOW()`,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

func (r *eventRepository) AddParticipant(ctx context.Context, eventID, userID uuid.UUID, status models.ParticipantStatus) error {
	err := r.db.Exec(ctx,
		`INSERT INTO event_participants (event_id, user_id, status) VALUES ($1, $2, $3)
		 ON CONFLICT (event_id, user_id) DO UPDATE SET status = $3`,
		eventID, userID, status,
	)
	if err != nil {
		return fmt.Errorf("add participant: %w", err)
	}
	return nil
}

func (r *eventRepository) GetParticipant(ctx context.Context, eventID, userID uuid.UUID) (*models.EventParticipant, error) {
	var p models.EventParticipant
	err := r.db.QueryRow(ctx,
		`SELECT event_id, user_id, status, joined_at FROM event_participants WHERE event_id = $1 AND user_id = $2`,
		eventID, userID,
	).Scan(&p.EventID, &p.UserID, &p.Status, &p.JoinedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *eventRepository) UpdateParticipantStatus(ctx context.Context, eventID, userID uuid.UUID, status models.ParticipantStatus) error {
	err := r.db.Exec(ctx,
		`UPDATE event_participants SET status = $1 WHERE event_id = $2 AND user_id = $3`,
		status, eventID, userID,
	)
	if err != nil {
		return fmt.Errorf("update participant: %w", err)
	}
	return nil
}

func (r *eventRepository) RemoveParticipant(ctx context.Context, eventID, userID uuid.UUID) error {
	err := r.db.Exec(ctx, `DELETE FROM event_participants WHERE event_id = $1 AND user_id = $2`, eventID, userID)
	if err != nil {
		return fmt.Errorf("remove participant: %w", err)
	}
	return nil
}

func (r *eventRepository) CountAcceptedParticipants(ctx context.Context, eventID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM event_participants WHERE event_id = $1 AND status = 'accepted'`,
		eventID,
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *eventRepository) ListParticipants(ctx context.Context, eventID uuid.UUID) ([]models.EventParticipant, error) {
	rows, err := r.db.Query(ctx,
		`SELECT event_id, user_id, status, joined_at FROM event_participants WHERE event_id = $1 ORDER BY joined_at ASC`,
		eventID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.EventParticipant
	for rows.Next() {
		var p models.EventParticipant
		if err := rows.Scan(&p.EventID, &p.UserID, &p.Status, &p.JoinedAt); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

func (r *eventRepository) GetUserEventIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.db.Query(ctx,
		`SELECT event_id FROM event_participants WHERE user_id = $1 AND status = 'accepted'`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *eventRepository) CategoryExists(ctx context.Context, categoryID int) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1)`, categoryID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
