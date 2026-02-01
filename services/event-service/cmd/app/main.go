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
	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	log, err := logger.New(cfg.Logger)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	log.Infow("Starting event-service", "env", cfg.Env)

	pg, err := postgres.NewPostgres(&cfg.Postgres, log.SugaredLogger)
	if err != nil {
		log.Fatal("Postgres connection failed: ", err)
	}
	defer pg.Close()

	redisClient, err := redis.NewClient(&cfg.Redis, log.SugaredLogger)
	if err != nil {
		log.Warnw("Redis connection failed (optional): ", "error", err)
	} else {
		defer redisClient.Close()
	}

	eventRepo := repository.NewEventRepository(pg, log.SugaredLogger)
	categoryRepo := repository.NewCategoryRepository(pg, log.SugaredLogger)
	eventSvc := service.NewEventService(eventRepo, log.SugaredLogger)
	eventHandler := handlers.NewEventHandler(eventSvc)
	categoryHandler := handlers.NewCategoryHandler(categoryRepo)

	routerCfg := http_transport.NewRouterConfig(cfg)
	router := http_transport.NewRouter(routerCfg, log)

	routes.SetupEventRoutes(router.Echo(), eventHandler, categoryHandler)

	go runServerWithRetry(router, cfg, log.SugaredLogger)
	go runExpiredEventsWorker(eventRepo, log.SugaredLogger)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Infow("Shutting down event-service...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := router.ShuttingDown(shutdownCtx); err != nil {
		log.Errorw("Server forced to shutdown", "error", err)
	}

	log.Infow("Event service stopped")
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

func runExpiredEventsWorker(repo repository.EventRepository, log *zap.SugaredLogger) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		n, err := repo.MarkExpiredAsFinished(ctx)
		cancel()
		if err != nil {
			log.Errorw("Failed to mark expired events", "error", err)
			continue
		}
		if n > 0 {
			log.Infow("Marked expired events as finished", "count", n)
		}
	}
}
