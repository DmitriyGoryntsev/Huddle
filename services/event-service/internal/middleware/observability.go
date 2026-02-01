package middleware

import (
	"context"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// RequestIDMiddleware генерирует или пробрасывает ID запроса
func RequestIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			reqID := c.Request().Header.Get("X-Request-ID")
			if reqID == "" {
				reqID = uuid.New().String()
			}
			c.Response().Header().Set("X-Request-ID", reqID)

			ctx := context.WithValue(c.Request().Context(), requestIDKey, reqID)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

// LoggingMiddleware логирует детали запроса
func LoggingMiddleware(logger *zap.SugaredLogger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			reqID := GetRequestIDFromCtx(c.Request().Context())

			reqLogger := logger.With("request_id", reqID, "method", c.Request().Method, "path", c.Request().URL.Path)

			ctx := context.WithValue(c.Request().Context(), loggerKey, reqLogger)
			c.SetRequest(c.Request().WithContext(ctx))

			err := next(c)

			duration := time.Since(start)
			status := c.Response().Status

			reqLogger.Infow("Request finished",
				"status", status,
				"duration_ms", duration.Milliseconds(),
			)
			return err
		}
	}
}

// RecoverMiddleware ловит паники, чтобы сервис не упал целиком
func RecoverMiddleware(logger *zap.SugaredLogger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					logger.Errorw("Panic recovered",
						"panic", r,
						"stack", string(debug.Stack()),
					)
					_ = c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal error"})
				}
			}()
			return next(c)
		}
	}
}
