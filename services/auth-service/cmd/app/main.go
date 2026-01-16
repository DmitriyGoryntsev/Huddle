// cmd/main.go
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"auth-service/internal/config"
	"auth-service/internal/middleware"
	"auth-service/internal/repository"
	"auth-service/internal/routes"
	"auth-service/internal/service"
	http_transport "auth-service/internal/transport/http"
	"auth-service/internal/transport/http/handlers"
	"auth-service/internal/utils"
	"auth-service/pkg/db/postgres"
	"auth-service/pkg/db/redis"
	"auth-service/pkg/logger"

	"github.com/labstack/echo/v4"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

func main() {
	// 1. Загрузка конфигурации
	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	// 2. Инициализация логгера
	log, err := logger.New(cfg.Logger)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	log.Infow("Starting auth-service", "version", "1.0.0", "env", cfg.Env)

	// 3. PostgreSQL
	pg, err := postgres.NewPostgres(&cfg.Postgres, log.SugaredLogger)
	if err != nil {
		log.Fatal("Postgres connection failed: ", err)
	}
	defer pg.Close()

	// 4. Redis
	redisClient, err := redis.NewClient(&cfg.Redis, log.SugaredLogger)
	if err != nil {
		log.Fatal("Redis connection failed: ", err)
	}
	defer redisClient.Close()

	// 5. TokenService
	tokenSvc, err := utils.NewTokenService(utils.TokenServiceConfig{
		SigningKey: cfg.JWT.Secret,
		AccessTTL:  cfg.JWT.TokenExpiry,
		RefreshTTL: cfg.JWT.RefreshExpiry,
		Issuer:     "auth-service",
		Redis:      redisClient.Inner(),
		Logger:     log.SugaredLogger,
	})
	if err != nil {
		log.Fatal("TokenService initialization failed: ", err)
	}

	// 6. Репозитории
	userRepo := repository.NewUserRepository(pg, log.SugaredLogger)
	outboxRepo := repository.NewOutboxRepository(pg)

	// 7. AuthService с Outbox поддержкой
	authSvc := service.NewAuthService(
		userRepo,
		outboxRepo,
		tokenSvc,
		log.SugaredLogger,
	)

	// 8. Kafka Producer для Outbox Publisher
	kafkaWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers:   cfg.Kafka.Brokers,
		Topic:     cfg.Kafka.Topic,
		BatchSize: cfg.Kafka.BatchSize,
		// Асинхронная запись с подтверждением
		Async:        false,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	})

	// 9. Outbox Publisher
	outboxPublisher := service.NewOutboxPublisher(
		outboxRepo,
		kafkaWriter,
		log.SugaredLogger,
	)

	// 10. Запускаем Outbox Publisher
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		log.Infow("Starting Outbox Publisher")
		outboxPublisher.Start(ctx)
	}()

	// 11. Graceful shutdown для Kafka Writer
	go func() {
		<-ctx.Done()
		log.Infow("Shutting down Kafka writer")
		if err := kafkaWriter.Close(); err != nil {
			log.Errorw("Failed to close Kafka writer", "error", err)
		}
	}()

	// 12. AuthHandler
	authHandler := handlers.NewAuthHandler(authSvc, log.SugaredLogger)

	// 13. HTTP Router
	routerCfg := http_transport.NewRouterConfig(cfg)
	router := http_transport.NewRouter(routerCfg, log)

	// 14. Middleware
	router.Echo().Use(middleware.AuthMiddleware(tokenSvc, log.SugaredLogger))

	// 15. Routes
	routes.SetupAuthRoutes(router.Echo(), authHandler)

	// 16. Health Check с информацией о Outbox
	router.Echo().GET("/health", func(c echo.Context) error {
		// Проверяем количество pending событий
		pendingEvents, err := outboxRepo.GetPendingEvents(c.Request().Context(), 1)
		if err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
				"status":  "degraded",
				"time":    time.Now().Format(time.RFC3339),
				"details": "outbox health check failed",
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":         "ok",
			"time":           time.Now().Format(time.RFC3339),
			"pending_events": len(pendingEvents),
			"kafka_brokers":  cfg.Kafka.Brokers,
			"kafka_topic":    cfg.Kafka.Topic,
		})
	})

	// 18. Запуск сервера с retry
	go runServerWithRetry(router, cfg, log.SugaredLogger)

	// 19. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Infow("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	// 20. Останавливаем Outbox Publisher
	cancel()

	// 21. Останавливаем HTTP сервер
	if err := router.ShuttingDown(shutdownCtx); err != nil {
		log.Errorw("Server forced to shutdown", "error", err)
	} else {
		log.Infow("Server gracefully stopped")
	}

	log.Infow("Auth service stopped")
}

// runServerWithRetry — запуск HTTP-сервера с retry
func runServerWithRetry(router *http_transport.Router, cfg *config.Config, log *zap.SugaredLogger) {
	maxRetries := cfg.HTTPServer.MaxRetries
	retryDelay := time.Duration(cfg.HTTPServer.RetryDelay) * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Infow("Starting HTTP server",
			"port", cfg.HTTPServer.Port,
			"attempt", attempt,
			"max_retries", maxRetries)

		if err := router.Run(); err != nil && err != http.ErrServerClosed {
			log.Errorw("Server failed", "attempt", attempt, "error", err)
			if attempt < maxRetries {
				log.Infow("Retrying in", "delay_seconds", retryDelay.Seconds())
				time.Sleep(retryDelay)
				continue
			}
			log.Fatalf("Server failed after all retries: %v", err)
		}
		break
	}
}
