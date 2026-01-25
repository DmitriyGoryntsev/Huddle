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

	userID, ok := c.Get("userID").(uuid.UUID)
	if !ok {
		log.Error("failed to extract userID from context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}

	profile, err := h.service.GetProfile(c.Request().Context(), userID)
	if err != nil {
		log.Errorw("Error retrieving profile", "UserID", userID, "error", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to get profile"})
	}

	return c.JSON(http.StatusOK, profile)
}
