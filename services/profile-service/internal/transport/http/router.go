package http

import (
	"context"
	"fmt"
	"profile-service/internal/config"
	"profile-service/internal/middleware"
	"profile-service/pkg/logger"

	"github.com/labstack/echo/v4"
)

type RouterConfig struct {
	Port string
}

type Router struct {
	config RouterConfig
	router *echo.Echo
}

func NewRouterConfig(cfg *config.Config) RouterConfig {
	return RouterConfig{
		Port: cfg.HTTPServer.Port,
	}
}

func NewRouter(rConfig RouterConfig, logger *logger.Logger) *Router {
	r := echo.New()

	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggingMiddleware(logger.SugaredLogger))
	r.Use(middleware.RecoverMiddleware(logger.SugaredLogger))

	return &Router{
		config: rConfig,
		router: r,
	}
}

func (r *Router) Run() error {
	return r.router.Start(fmt.Sprintf(":%s", r.config.Port))
}

func (r *Router) ShuttingDown(ctx context.Context) error {
	return r.router.Shutdown(ctx)
}

func (r *Router) Echo() *echo.Echo {
	return r.router
}
