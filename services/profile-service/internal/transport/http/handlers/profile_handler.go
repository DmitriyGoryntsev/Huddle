package handlers

import (
	"net/http"
	"profile-service/internal/middleware"
	"profile-service/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type ProfileHandler struct {
	service   service.ProfileService
	logger    *zap.SugaredLogger
	validator *validator.Validate
}

// NewProfileHandler — конструктор
func NewProfileHandler(service service.ProfileService, logger *zap.SugaredLogger) *ProfileHandler {
	return &ProfileHandler{
		service:   service,
		logger:    logger,
		validator: validator.New(),
	}
}

// GetProfile - получение профиля пользователя
func (h *ProfileHandler) GetProfile(c echo.Context) error {
	log := middleware.GetLoggerFromCtx(c.Request().Context())

	userIDStr, ok := c.Get("user_id").(string)
	if !ok {
		log.Error("failed to extract userID from context (not a string)")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Errorw("failed to parse userID string to UUID", "str", userIDStr, "error", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid user id format"})
	}

	profile, err := h.service.GetProfile(c.Request().Context(), userID)
	if err != nil {
		log.Errorw("Error retrieving profile", "UserID", userID, "error", err)
		return c.JSON(http.StatusNotFound, echo.Map{"error": "failed to get profile"})
	}

	return c.JSON(http.StatusOK, profile)
}
