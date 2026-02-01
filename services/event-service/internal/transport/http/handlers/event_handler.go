package handlers

import (
	"net/http"
	"strconv"

	"event-service/internal/middleware"
	"event-service/internal/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type EventService interface {
	Create(event *models.Event) error
	GetByID(id uuid.UUID) (*models.Event, error)
	List(filter models.EventFilter) ([]models.Event, error)
	Delete(eventID, userID uuid.UUID) error
	Join(eventID, userID uuid.UUID) error
	Leave(eventID, userID uuid.UUID) error
	UpdateParticipantStatus(eventID, targetUserID, creatorID uuid.UUID, status models.ParticipantStatus) error
	GetUsersEvents(userID uuid.UUID) ([]models.Event, error)
}

type EventHandler struct {
	service EventService
}

func NewEventHandler(service EventService) *EventHandler {
	return &EventHandler{service: service}
}

// 1. Создание события
func (h *EventHandler) CreateEvent(c echo.Context) error {
	log := middleware.GetLoggerFromCtx(c.Request().Context())
	userID := uuid.MustParse(c.Request().Header.Get("X-User-ID"))

	log.Info("Creating new event", "user_id", userID)

	var req models.CreateEventRequest
	if err := c.Bind(&req); err != nil {
		log.Error("Failed to bind create event request", "error", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request body"})
	}

	event := &models.Event{
		CreatorID:        userID,
		Title:            req.Title,
		Description:      req.Description,
		Latitude:         req.Latitude,
		Longitude:        req.Longitude,
		CategoryID:       req.CategoryID,
		StartTime:        req.StartTime,
		MaxParticipants:  req.MaxParticipants,
		Price:            req.Price,
		RequiresApproval: req.RequiresApproval,
		Status:           models.EventStatusOpen,
	}

	if err := h.service.Create(event); err != nil {
		log.Error("Failed to create event in service", "error", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to create event"})
	}

	log.Info("Event created successfully", "event_id", event.ID)
	return c.JSON(http.StatusCreated, event)
}

// 2. Список событий (карта/фильтры)
func (h *EventHandler) ListEvents(c echo.Context) error {
	log := middleware.GetLoggerFromCtx(c.Request().Context())

	lat, _ := strconv.ParseFloat(c.QueryParam("lat"), 64)
	lon, _ := strconv.ParseFloat(c.QueryParam("lon"), 64)
	radius, _ := strconv.ParseFloat(c.QueryParam("radius"), 64)

	log.Info("Listing events with filters", "lat", lat, "lon", lon, "radius", radius)

	filter := models.EventFilter{
		Latitude:     lat,
		Longitude:    lon,
		RadiusMeters: radius,
		CategorySlug: c.QueryParam("category"),
	}

	events, err := h.service.List(filter)
	if err != nil {
		log.Error("Failed to list events", "error", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to fetch events"})
	}

	return c.JSON(http.StatusOK, events)
}

// 3. Получить одно событие
func (h *EventHandler) GetEvent(c echo.Context) error {
	log := middleware.GetLoggerFromCtx(c.Request().Context())
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid event id"})
	}

	log.Info("Fetching event details", "event_id", id)

	event, err := h.service.GetByID(id)
	if err != nil {
		log.Warn("Event not found", "event_id", id)
		return c.JSON(http.StatusNotFound, echo.Map{"error": "event not found"})
	}

	return c.JSON(http.StatusOK, event)
}

// 4. Удалить событие
func (h *EventHandler) DeleteEvent(c echo.Context) error {
	log := middleware.GetLoggerFromCtx(c.Request().Context())
	eventID := uuid.MustParse(c.Param("id"))
	userID := uuid.MustParse(c.Request().Header.Get("X-User-ID"))

	log.Info("Attempting to delete event", "event_id", eventID, "user_id", userID)

	if err := h.service.Delete(eventID, userID); err != nil {
		log.Error("Failed to delete event", "error", err)
		return c.JSON(http.StatusForbidden, echo.Map{"error": "could not delete event"})
	}

	return c.NoContent(http.StatusNoContent)
}

// 5. Присоединиться к событию
func (h *EventHandler) JoinEvent(c echo.Context) error {
	log := middleware.GetLoggerFromCtx(c.Request().Context())
	eventID := uuid.MustParse(c.Param("id"))
	userID := uuid.MustParse(c.Request().Header.Get("X-User-ID"))

	log.Info("User joining event", "event_id", eventID, "user_id", userID)

	if err := h.service.Join(eventID, userID); err != nil {
		log.Error("Failed to join event", "error", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "request sent"})
}

// 6. Покинуть событие
func (h *EventHandler) LeaveEvent(c echo.Context) error {
	log := middleware.GetLoggerFromCtx(c.Request().Context())
	eventID := uuid.MustParse(c.Param("id"))
	userID := uuid.MustParse(c.Request().Header.Get("X-User-ID"))

	log.Info("User leaving event", "event_id", eventID, "user_id", userID)

	if err := h.service.Leave(eventID, userID); err != nil {
		log.Error("Failed to leave event", "error", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to leave event"})
	}

	return c.NoContent(http.StatusNoContent)
}

// 7. Одобрить/отклонить участника
func (h *EventHandler) UpdateParticipantStatus(c echo.Context) error {
	log := middleware.GetLoggerFromCtx(c.Request().Context())
	eventID := uuid.MustParse(c.Param("id"))
	targetUserID := uuid.MustParse(c.Param("user_id"))
	creatorID := uuid.MustParse(c.Request().Header.Get("X-User-ID"))

	var req models.UpdateParticipantStatusRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid status"})
	}

	log.Info("Updating participant status", "event_id", eventID, "target_user", targetUserID, "status", req.Status)

	if err := h.service.UpdateParticipantStatus(eventID, targetUserID, creatorID, req.Status); err != nil {
		log.Error("Failed to update status", "error", err)
		return c.JSON(http.StatusForbidden, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"status": "updated"})
}

// 8. Мои события
func (h *EventHandler) GetMyEvents(c echo.Context) error {
	log := middleware.GetLoggerFromCtx(c.Request().Context())
	userID := uuid.MustParse(c.Request().Header.Get("X-User-ID"))

	log.Info("Fetching user's events", "user_id", userID)

	events, err := h.service.GetUsersEvents(userID)
	if err != nil {
		log.Error("Failed to fetch user events", "error", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to fetch events"})
	}

	return c.JSON(http.StatusOK, events)
}
