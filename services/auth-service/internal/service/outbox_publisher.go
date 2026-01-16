package service

import (
	"context"
	"time"

	"auth-service/internal/repository"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type OutboxPublisher struct {
	repo   repository.OutboxRepository
	writer *kafka.Writer
	logger *zap.SugaredLogger
	ticker *time.Ticker
}

func NewOutboxPublisher(
	repo repository.OutboxRepository,
	writer *kafka.Writer,
	logger *zap.SugaredLogger,
) *OutboxPublisher {
	return &OutboxPublisher{
		repo:   repo,
		writer: writer,
		logger: logger,
		ticker: time.NewTicker(5 * time.Second),
	}
}

func (p *OutboxPublisher) Start(ctx context.Context) {
	go p.run(ctx)
}

func (p *OutboxPublisher) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			p.ticker.Stop()
			return
		case <-p.ticker.C:
			if err := p.publishBatch(ctx); err != nil {
				p.logger.Errorw("Failed to publish batch", "error", err)
			}
		}
	}
}

func (p *OutboxPublisher) publishBatch(ctx context.Context) error {
	events, err := p.repo.GetPendingEvents(ctx, 100)
	if err != nil {
		return err
	}
	if len(events) == 0 {
		return nil
	}

	for _, event := range events {
		msg := kafka.Message{
			Key:   []byte(event.ID.String()),
			Topic: "user-events",
			Value: event.Payload,
		}

		if err := p.writer.WriteMessages(ctx, msg); err != nil {
			p.logger.Errorw("Failed to publish event", "event_id", event.ID, "error", err)
			continue
		}

		if err := p.repo.MarkAsPublished(ctx, event.ID); err != nil {
			p.logger.Errorw("Failed to mark event as published", "event_id", event.ID, "error", err)
		}

		p.logger.Infow("Event published successfully",
			"event_id", event.ID,
			"event_type", event.EventType)
	}

	return nil
}
