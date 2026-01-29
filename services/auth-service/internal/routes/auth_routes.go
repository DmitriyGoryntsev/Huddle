package routes

import (
	"auth-service/internal/middleware"
	"auth-service/internal/transport/http/handlers"
	"auth-service/internal/utils"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func SetupAuthRoutes(router *echo.Echo, authHandler *handlers.AuthHandler, tokenSvc utils.TokenService, logger *zap.SugaredLogger) {
	// Базовая группа API с версией
	api := router.Group("/api/v1")

	// Группа для авторизации
	auth := api.Group("/auth")
	{
		// --- ПУБЛИЧНЫЕ ЭНДПОИНТЫ ---
		// Здесь middleware не применяется
		auth.POST("/register", authHandler.RegisterUser)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh-token", authHandler.RefreshToken)

		// --- ЗАЩИЩЕННЫЕ ЭНДПОИНТЫ ---
		// Создаем подгруппу, к которой применяем AuthMiddleware
		protected := auth.Group("")
		protected.Use(middleware.AuthMiddleware(tokenSvc, logger))
		{
			// POST /api/v1/auth/logout -> Требует валидный Access Token
			protected.POST("/logout", authHandler.Logout)

			// GET /api/v1/auth/validate -> Эндпоинт для Nginx (auth_request)
			protected.GET("/validate", authHandler.Validate)
		}
	}
}
