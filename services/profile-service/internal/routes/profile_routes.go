package routes

import (
	"profile-service/internal/transport/http/handlers"

	"github.com/labstack/echo/v4"
)

func SetupProfileRoutes(router *echo.Echo, profileHandler *handlers.ProfileHandler) {
	// Базовая группа API с версией
	api := router.Group("/api/v1")

	// Группа для авторизации
	profile := api.Group("/profile")
	{
		profile.GET("", profileHandler.GetProfile)
	}

}
