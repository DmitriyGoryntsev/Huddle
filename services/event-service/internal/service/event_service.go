package service

import (
	"context"
	"event-service/internal/models"
	"event-service/internal/repository"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// EventService — бизнес-логика событий Huddle
type EventService interface {
	Create(ctx context.Context, event *models.Event) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Event, error)
	List(ctx context.Context, filter models.EventFilter) ([]*models.Event, error)
	Delete(ctx context.Context, eventID, userID uuid.UUID) error
	Join(ctx context.Context, eventID, userID uuid.UUID) error
	Leave(ctx context.Context, eventID, userID uuid.UUID) error
	UpdateParticipantStatus(ctx context.Context, eventID, targetUserID, creatorID uuid.UUID, status models.ParticipantStatus) error
	GetEventParticipants(ctx context.Context, eventID uuid.UUID) ([]models.EventParticipant, error)
	GetUsersEvents(ctx context.Context, userID uuid.UUID) ([]*models.Event, error)
}

type eventService struct {
	repo  repository.EventRepository
	logger *zap.SugaredLogger
}

func NewEventService(repo repository.EventRepository, logger *zap.SugaredLogger) EventService {
	return &eventService{repo: repo, logger: logger}
}

func (s *eventService) Create(ctx context.Context, event *models.Event) error {
	ok, err := s.repo.CategoryExists(ctx, event.CategoryID)
	if err != nil {
		s.logger.Errorw("Failed to check category", "category_id", event.CategoryID, "error", err)
		return fmt.Errorf("failed to create event: %w", err)
	}
	if !ok {
		return fmt.Errorf("category %d does not exist", event.CategoryID)
	}

	event.Status = models.EventStatusOpen
	if err := s.repo.Create(ctx, event); err != nil {
		s.logger.Errorw("Failed to create event", "creator_id", event.CreatorID, "error", err)
		return err
	}
	s.logger.Infow("Event created", "event_id", event.ID, "creator_id", event.CreatorID)
	return nil
}

func (s *eventService) GetByID(ctx context.Context, id uuid.UUID) (*models.Event, error) {
	event, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, fmt.Errorf("event not found")
	}

	return event, nil
}

func (s *eventService) List(ctx context.Context, filter models.EventFilter) ([]*models.Event, error) {
	events, err := s.repo.ListNearby(ctx, filter)
	if err != nil {
		s.logger.Errorw("Failed to list events", "error", err)
		return nil, err
	}
	return events, nil
}

func (s *eventService) Delete(ctx context.Context, eventID, userID uuid.UUID) error {
	if err := s.repo.Delete(ctx, eventID, userID); err != nil {
		if err.Error() == "event not found or not authorized to delete" {
			return fmt.Errorf("only event creator can perform this action")
		}
		return err
	}
	s.logger.Infow("Event deleted", "event_id", eventID, "user_id", userID)
	return nil
}

func (s *eventService) Join(ctx context.Context, eventID, userID uuid.UUID) error {
	event, err := s.repo.GetByID(ctx, eventID)
	if err != nil || event == nil {
		return fmt.Errorf("event not found")
	}
	if event.Status != models.EventStatusOpen {
		return fmt.Errorf("event has expired")
	}

	participant, _ := s.repo.GetParticipant(ctx, eventID, userID)
	if participant != nil {
		if participant.Status == models.ParticipantStatusAccepted {
			return fmt.Errorf("already a participant")
		}
		if participant.Status == models.ParticipantStatusPending {
			return fmt.Errorf("already requested, waiting for approval")
		}
		// rejected — можно попробовать снова, обновим запись
	}

	count, _ := s.repo.CountAcceptedParticipants(ctx, eventID)
	if count >= event.MaxParticipants {
		return fmt.Errorf("event is full")
	}

	status := models.ParticipantStatusAccepted
	if event.RequiresApproval {
		status = models.ParticipantStatusPending
	}

	if err := s.repo.AddParticipant(ctx, eventID, userID, status); err != nil {
		s.logger.Errorw("Failed to add participant", "event_id", eventID, "user_id", userID, "error", err)
		return err
	}

	s.logger.Infow("User joined event", "event_id", eventID, "user_id", userID, "requires_approval", event.RequiresApproval)
	return nil
}

func (s *eventService) Leave(ctx context.Context, eventID, userID uuid.UUID) error {
	event, err := s.repo.GetByID(ctx, eventID)
	if err != nil || event == nil {
		return fmt.Errorf("event not found")
	}
	if event.CreatorID == userID {
		return fmt.Errorf("creator cannot leave, use delete instead")
	}

	participant, _ := s.repo.GetParticipant(ctx, eventID, userID)
	if participant == nil {
		return fmt.Errorf("not a participant")
	}

	if err := s.repo.RemoveParticipant(ctx, eventID, userID); err != nil {
		s.logger.Errorw("Failed to remove participant", "event_id", eventID, "user_id", userID, "error", err)
		return err
	}
	s.logger.Infow("User left event", "event_id", eventID, "user_id", userID)
	return nil
}

func (s *eventService) UpdateParticipantStatus(ctx context.Context, eventID, targetUserID, creatorID uuid.UUID, status models.ParticipantStatus) error {
	event, err := s.repo.GetByID(ctx, eventID)
	if err != nil || event == nil {
		return fmt.Errorf("event not found")
	}
	if event.CreatorID != creatorID {
		return fmt.Errorf("only event creator can perform this action")
	}
	if !event.RequiresApproval {
		return fmt.Errorf("event does not require approval")
	}

	participant, _ := s.repo.GetParticipant(ctx, eventID, targetUserID)
	if participant == nil || participant.Status != models.ParticipantStatusPending {
		return fmt.Errorf("no pending request from this user")
	}

	if err := s.repo.UpdateParticipantStatus(ctx, eventID, targetUserID, status); err != nil {
		return err
	}

	if status == models.ParticipantStatusAccepted {
		count, _ := s.repo.CountAcceptedParticipants(ctx, eventID)
		if count >= event.MaxParticipants {
			_ = s.repo.UpdateStatus(ctx, eventID, models.EventStatusFull)
			s.logger.Infow("Event is full", "event_id", eventID)
		}
	}

	s.logger.Infow("Participant status updated", "event_id", eventID, "target_user", targetUserID, "status", status)
	return nil
}

func (s *eventService) GetEventParticipants(ctx context.Context, eventID uuid.UUID) ([]models.EventParticipant, error) {
	event, err := s.repo.GetByID(ctx, eventID)
	if err != nil || event == nil {
		return nil, fmt.Errorf("event not found")
	}
	return s.repo.ListParticipants(ctx, eventID)
}

func (s *eventService) GetUsersEvents(ctx context.Context, userID uuid.UUID) ([]*models.Event, error) {
	ids, err := s.repo.GetUserEventIDs(ctx, userID)
	if err != nil {
		s.logger.Errorw("Failed to get user event ids", "user_id", userID, "error", err)
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}
	events, err := s.repo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	// Фильтруем только активные (open/full, не истекшие)
	var active []*models.Event
	for _, e := range events {
		if e.Status == models.EventStatusOpen || e.Status == models.EventStatusFull {
			active = append(active, e)
		}
	}
	return active, nil
}
