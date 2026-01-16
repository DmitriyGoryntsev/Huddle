package middleware

import (
	"auth-service/internal/utils"
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// AuthMiddleware — проверяет access токен и кладёт claims в контекст
func AuthMiddleware(tokenSvc utils.TokenService, logger *zap.SugaredLogger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 1. Извлекаем токен
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				logger.Debug("Missing Authorization header")
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "missing authorization header"})
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				logger.Debugw("Invalid Authorization header format", "header", authHeader)
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid authorization header format"})
			}

			tokenStr := parts[1]

			// 2. Парсим access токен
			claims, err := tokenSvc.ParseAccess(tokenStr)
			if err != nil {
				logger.Infow("Invalid access token", "error", err, "token_prefix", tokenStr[:10]+"...")
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid or expired token"})
			}

			// 3. Логируем успешную аутентификацию
			logger.Infow("User authenticated",
				"user_id", claims.Sub,
				"email", claims.Email,
				"role", claims.Role,
				"jti", claims.JTI,
			)

			// 4. Кладём claims и logger в контекст
			ctx := context.WithValue(c.Request().Context(), UserClaimsKey, claims)
			ctx = context.WithValue(ctx, loggerKey, logger)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}
