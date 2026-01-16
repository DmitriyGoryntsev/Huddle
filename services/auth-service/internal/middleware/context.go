package middleware

import (
	"auth-service/internal/models"
	"context"

	"go.uber.org/zap"
)

// contextKey - тип для ключей контекста (защита от коллизий)
type contextKey string

const (
	loggerKey     contextKey = "logger"
	requestIDKey  contextKey = "request_id"
	userKey       contextKey = "user"
	UserClaimsKey contextKey = "user_claims"
)

// GetLoggerFromCtx - получение логгера из контекста
func GetLoggerFromCtx(ctx context.Context) *zap.SugaredLogger {
	if log, ok := ctx.Value(loggerKey).(*zap.SugaredLogger); ok {
		return log
	}
	return zap.NewNop().Sugar()
}

// GetRequestIDFromCtx - получение request_id
func GetRequestIDFromCtx(ctx context.Context) string {
	if reqID, ok := ctx.Value(requestIDKey).(string); ok {
		return reqID
	}
	return ""
}

func GetUserClaims(ctx context.Context) (*models.AccessTokenClaims, bool) {
	if claims, ok := ctx.Value(userKey).(*models.AccessTokenClaims); ok {
		return claims, true
	}
	return nil, false
}
