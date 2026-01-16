package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func RecoverMiddleware(logger *zap.SugaredLogger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					logger.Errorw("Panic recovered",
						"panic", r,
						"stack", string(debug.Stack()),
						"path", c.Request().URL.Path,
					)
					c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal server error"})
				}
			}()
			return next(c)
		}
	}
}
