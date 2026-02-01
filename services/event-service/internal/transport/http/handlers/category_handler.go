package handlers

import (
	"net/http"

	"event-service/internal/middleware"
	"event-service/internal/repository"

	"github.com/labstack/echo/v4"
)

type CategoryHandler struct {
	repo repository.CategoryRepository
}

func NewCategoryHandler(repo repository.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{repo: repo}
}

func (h *CategoryHandler) ListCategories(c echo.Context) error {
	log := middleware.GetLoggerFromCtx(c.Request().Context())

	categories, err := h.repo.List(c.Request().Context())
	if err != nil {
		log.Errorw("Failed to list categories", "error", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to fetch categories"})
	}
	return c.JSON(http.StatusOK, categories)
}
