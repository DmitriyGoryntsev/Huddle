package routes

import (
	"event-service/internal/middleware"
	"event-service/internal/transport/http/handlers"

	"github.com/labstack/echo/v4"
)

func SetupEventRoutes(router *echo.Echo, eventHandler *handlers.EventHandler, categoryHandler *handlers.CategoryHandler) {
	// Группа с авторизацией (Nginx уже проверил JWT, нам нужно просто вытащить ID)
	api := router.Group("/api/v1", middleware.AuthMiddleware())

	events := api.Group("/events")
	{
		// Создать 
		events.POST("", eventHandler.CreateEvent)

		// Поиск на карте: GET /api/v1/events?lat=55.75&lon=37.61&radius=1000
		events.GET("", eventHandler.ListEvents)

		events.GET("/:id", eventHandler.GetEvent)
		events.DELETE("/:id", eventHandler.DeleteEvent)

		// Работа с участниками
		participation := events.Group("/:id/participants")
		{
			participation.GET("", eventHandler.GetEventParticipants)
			participation.POST("", eventHandler.JoinEvent)
			participation.DELETE("", eventHandler.LeaveEvent)

			// PATCH /api/v1/events/123/participants/456
			participation.PATCH("/:user_id", eventHandler.UpdateParticipantStatus)
		}
	}

	userEvents := api.Group("/my-events")
	{
		userEvents.GET("", eventHandler.GetMyEvents)
	}

	categories := api.Group("/categories")
	{
		categories.GET("", categoryHandler.ListCategories)
	}
}
