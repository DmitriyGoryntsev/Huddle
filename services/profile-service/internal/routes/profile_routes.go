package routes

import (
	"profile-service/internal/middleware"
	"profile-service/internal/transport/http/handlers"

	"github.com/labstack/echo/v4"
)

func SetupProfileRoutes(router *echo.Echo, profileHandler *handlers.ProfileHandler) {
	// Базовая группа API с версией
	api := router.Group("/api/v1")

	// Группа для авторизации
	profiles := api.Group("/profiles", middleware.AuthMiddleware())
	{
		// GET /api/v1/profiles/me -> Получить СВОЙ профиль
		profiles.GET("/me", profileHandler.GetProfile)

		// GET /api/v1/profiles/:id -> Посмотреть ЧУЖОЙ профиль
		// profiles.GET("/:id", profileHandler.GetProfileByID)

		// PUT /api/v1/profiles/me -> Обновить свой профиль
		// profiles.PUT("/me", profileHandler.UpdateProfile)
	}

}
