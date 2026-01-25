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

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

func main() {
	// Контекст для управления жизненным циклом приложения
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Загрузка конфигурации
	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	// Инициализация логгера
	log, err := logger.New(cfg.Logger)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	log.Infow("Starting auth-service", "env", cfg.Env)

	// Подключение к PostgreSQL
	pg, err := postgres.NewPostgres(&cfg.Postgres, log.SugaredLogger)
	if err != nil {
		log.Fatal("Postgres connection failed: ", err)
	}
	defer pg.Close()

	// Подключение к Redis
	redisClient, err := redis.NewClient(&cfg.Redis, log.SugaredLogger)
	if err != nil {
		log.Fatal("Redis connection failed: ", err)
	}
	defer redisClient.Close()

	// Инициализация TokenService
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

	// Инициализация репозиториев
	userRepo := repository.NewUserRepository(pg, log.SugaredLogger)
	outboxRepo := repository.NewOutboxRepository(pg)

	// Инициализация сервисов
	authSvc := service.NewAuthService(userRepo, outboxRepo, tokenSvc, log.SugaredLogger)

	// Настройка Kafka Writer
	kafkaWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      cfg.Kafka.Brokers,
		Topic:        cfg.Kafka.Topic,
		BatchSize:    cfg.Kafka.BatchSize,
		Async:        false,
		WriteTimeout: 10 * time.Second,
	})

	// Инициализация Outbox Publisher
	outboxPublisher := service.NewOutboxPublisher(outboxRepo, kafkaWriter, log.SugaredLogger)

	// Запуск фоновых задач
	go func() {
		log.Infow("Starting Outbox Publisher")
		outboxPublisher.Start(ctx)
	}()

	// Инициализация обработчиков (Handlers)
	authHandler := handlers.NewAuthHandler(authSvc, log.SugaredLogger)

	// Настройка HTTP транспорта и Middleware
	routerCfg := http_transport.NewRouterConfig(cfg)
	router := http_transport.NewRouter(routerCfg, log)
	router.Echo().Use(middleware.AuthMiddleware(tokenSvc, log.SugaredLogger))

	// Регистрация маршрутов
	routes.SetupAuthRoutes(router.Echo(), authHandler)

	// Запуск HTTP сервера в отдельной горутине
	go runServerWithRetry(router, cfg, log.SugaredLogger)

	// Ожидание сигналов завершения (Graceful Shutdown)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Infow("Shutting down auth-service...")

	// Контекст для корректного завершения работы
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	// Остановка фоновых процессов
	cancel()

	// Закрытие ресурсов Kafka
	if err := kafkaWriter.Close(); err != nil {
		log.Errorw("Failed to close Kafka writer", "error", err)
	}

	// Остановка HTTP сервера
	if err := router.ShuttingDown(shutdownCtx); err != nil {
		log.Errorw("Server forced to shutdown", "error", err)
	}

	log.Infow("Auth service stopped")
}

func runServerWithRetry(router *http_transport.Router, cfg *config.Config, log *zap.SugaredLogger) {
	maxRetries := cfg.HTTPServer.MaxRetries
	retryDelay := time.Duration(cfg.HTTPServer.RetryDelay) * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Infow("Starting HTTP server", "port", cfg.HTTPServer.Port, "attempt", attempt)

		if err := router.Run(); err != nil && err != http.ErrServerClosed {
			log.Errorw("Server failed", "attempt", attempt, "error", err)
			if attempt < maxRetries {
				time.Sleep(retryDelay)
				continue
			}
			log.Fatalf("Server failed after all retries")
		}
		break
	}
}
