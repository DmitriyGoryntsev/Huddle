package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"event-service/internal/config"
	"event-service/internal/repository"
	"event-service/internal/routes"
	"event-service/internal/service"
	http_transport "event-service/internal/transport/http"
	"event-service/internal/transport/http/handlers"
	"event-service/pkg/db/postgres"
	"event-service/pkg/db/redis"
	"event-service/pkg/logger"

	"go.uber.org/zap"
)

func main() {
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

	log.Infow("Starting profile-service", "env", cfg.Env)

	// Подключение к PostgreSQL
	pg, err := postgres.NewPostgres(&cfg.Postgres, log.SugaredLogger)
	if err != nil {
		log.Fatal("Postgres connection failed: ", err)
	}
	defer pg.Close()

	// Подключение к Redis
	redisClient, err := redis.NewClient(&cfg.Redis, log.SugaredLogger)
	if err != nil {
		log.Warnw("Redis connection failed (optional): ", "error", err)
	} else {
		defer redisClient.Close()
	}

	// Инициализация репозиториев
	profileRepo := repository.NewProfileRepository(pg, log.SugaredLogger)

	// Инициализация сервисов
	profileSvc := service.NewProfileService(profileRepo, log.SugaredLogger)

	// Инициализация обработчиков (Handlers)
	profileHandler := handlers.NewProfileHandler(profileSvc, log.SugaredLogger)

	// Настройка HTTP транспорта и Middleware
	routerCfg := http_transport.NewRouterConfig(cfg)
	router := http_transport.NewRouter(routerCfg, log)
	// Примечание: Здесь можно добавить AuthMiddleware, если профилю нужно валидировать токены от Auth-service

	// Регистрация маршрутов
	routes.SetupProfileRoutes(router.Echo(), profileHandler)

	// Запуск HTTP сервера в отдельной горутине
	go runServerWithRetry(router, cfg, log.SugaredLogger)

	// Ожидание сигналов завершения (Graceful Shutdown)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Infow("Shutting down profile-service...")

	// Контекст для корректного завершения работы
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	// Остановка HTTP сервера
	if err := router.ShuttingDown(shutdownCtx); err != nil {
		log.Errorw("Server forced to shutdown", "error", err)
	}

	log.Infow("Profile service stopped")
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
