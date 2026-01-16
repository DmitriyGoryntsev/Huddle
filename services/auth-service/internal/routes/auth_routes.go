package routes

import (
	"auth-service/internal/transport/http/handlers"

	"github.com/labstack/echo/v4"
)

func SetupAuthRoutes(router *echo.Echo, authHandler *handlers.AuthHandler) {
	// Базовая группа API с версией
	api := router.Group("/api/v1")

	// Группа для авторизации
	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.RegisterUser)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh-token", authHandler.RefreshToken)
		auth.POST("/logout", authHandler.Logout)
	}

}
