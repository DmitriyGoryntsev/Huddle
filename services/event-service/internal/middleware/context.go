package middleware

import (
	"context"

	"go.uber.org/zap"
)

type contextKey string

const (
	loggerKey     contextKey = "logger"
	requestIDKey  contextKey = "request_id"
	UserClaimsKey contextKey = "user_claims"
)

func GetLoggerFromCtx(ctx context.Context) *zap.SugaredLogger {
	if log, ok := ctx.Value(loggerKey).(*zap.SugaredLogger); ok {
		return log
	}
	return zap.NewNop().Sugar()
}

func GetRequestIDFromCtx(ctx context.Context) string {
	if reqID, ok := ctx.Value(requestIDKey).(string); ok {
		return reqID
	}
	return ""
}
