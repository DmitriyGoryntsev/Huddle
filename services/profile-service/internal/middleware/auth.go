package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func AuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Nginx прислал нам это в заголовке
			userID := c.Request().Header.Get("X-User-ID")

			if userID == "" {
				// Значит запрос пришел в обход Gateway
				return c.JSON(http.StatusForbidden, echo.Map{"error": "direct access forbidden"})
			}

			// Кладем ID пользователя в контекст для хендлеров
			c.Set("user_id", userID)
			return next(c)
		}
	}
}
