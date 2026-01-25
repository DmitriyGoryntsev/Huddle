package events

import (
	"context"
	"encoding/json"
	"fmt"
	"profile-service/internal/models"
	"profile-service/internal/service"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type UserConsumer struct {
	reader  *kafka.Reader
	service service.ProfileService
	logger  *zap.SugaredLogger
}

func NewUserConsumer(brokers []string, topic string, service service.ProfileService, logger *zap.SugaredLogger) *UserConsumer {
	return &UserConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  "profile-service-group", // Важно: ID группы для отслеживания офсетов
			MinBytes: 10e3,                    // 10KB
			MaxBytes: 10e6,                    // 10MB
		}),
		service: service,
		logger:  logger,
	}
}

// Start запускает цикл прослушивания сообщений
func (c *UserConsumer) Start(ctx context.Context) {
	c.logger.Infow("Kafka consumer started", "topic", c.reader.Config().Topic)

	for {
		// Читаем сообщение
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return // Контекст отменен, выходим
			}
			c.logger.Errorw("Failed to read message from Kafka", "error", err)
			continue
		}

		// Обрабатываем сообщение
		if err := c.processEvent(ctx, msg.Value); err != nil {
			c.logger.Errorw("Failed to process event", "error", err)
			// TODO Retry или отправку в DLQ (Dead Letter Queue)
		}
	}
}

func (c *UserConsumer) processEvent(ctx context.Context, payload []byte) error {
	// Распаковываем JSON в структуру события
	var event models.UserRegistered
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	c.logger.Infow("Received UserRegistered event", "user_id", event.UserID)

	// Вызываем наш сервис для создания профиля
	return c.service.CreateProfile(ctx, event)
}

func (c *UserConsumer) Close() error {
	return c.reader.Close()
}
