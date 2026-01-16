package middleware

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

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

func LoggingMiddleware(logger *zap.SugaredLogger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			requestID := GetRequestIDFromCtx(c.Request().Context())

			requestLogger := logger.With(
				"request_id", requestID,
				"method", c.Request().Method,
				"path", c.Request().URL.Path,
				"remote_ip", c.RealIP(),
			)

			ctx := context.WithValue(c.Request().Context(), loggerKey, requestLogger)
			c.SetRequest(c.Request().WithContext(ctx))

			requestLogger.Infow("Request started")

			err := next(c)

			duration := time.Since(start)
			status := c.Response().Status

			logFields := []interface{}{
				"status", status,
				"latency_ms", duration.Milliseconds(),
				"bytes_in", c.Request().ContentLength,
				"bytes_out", c.Response().Size,
			}

			if err != nil {
				requestLogger.Errorw("Request failed", append(logFields, "error", err)...)
			} else {
				level := "info"
				if status >= 500 {
					level = "error"
				} else if status >= 400 {
					level = "warn"
				}

				// Динамический уровень логирования в зависимости от статуса
				switch level {
				case "error":
					requestLogger.Errorw("Request completed", logFields...)
				case "warn":
					requestLogger.Warnw("Request completed", logFields...)
				default:
					requestLogger.Infow("Request completed", logFields...)
				}
			}

			return err
		}
	}
}
