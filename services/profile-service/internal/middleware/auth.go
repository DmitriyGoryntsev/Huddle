package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func AuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID := c.Request().Header.Get("X-User-ID")

			if userID == "" {
				return c.JSON(http.StatusForbidden, echo.Map{"error": "header X-User-ID is empty"})
			}

			c.Set("user_id", userID)
			return next(c)
		}
	}
}
