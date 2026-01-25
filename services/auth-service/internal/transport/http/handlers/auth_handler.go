// internal/handlers/auth_handler.go
package handlers

import (
	"auth-service/internal/middleware"
	"auth-service/internal/models"
	"auth-service/internal/service"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// AuthHandler — HTTP-обёртка над AuthService
type AuthHandler struct {
	service   service.AuthService
	logger    *zap.SugaredLogger
	validator *validator.Validate
}

// NewAuthHandler — конструктор
func NewAuthHandler(service service.AuthService, logger *zap.SugaredLogger) *AuthHandler {
	return &AuthHandler{
		service:   service,
		logger:    logger,
		validator: validator.New(),
	}
}

// RegisterUser — регистрация
func (h *AuthHandler) RegisterUser(c echo.Context) error {
	var req models.UserRegister
	log := middleware.GetLoggerFromCtx(c.Request().Context())

	if err := c.Bind(&req); err != nil {
		log.Warnw("Bind failed", "error", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request body"})
	}

	if err := h.validator.Struct(req); err != nil {
		log.Warnw("Validation failed", "error", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	user, err := h.service.RegisterUser(c.Request().Context(), req)
	if err != nil {
		log.Errorw("Register failed", "email", req.Email, "error", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "registration failed"})
	}

	log.Infow("User registered", "user_id", user.ID, "email", user.Email)
	return c.JSON(http.StatusCreated, echo.Map{
		"message": "user registered successfully",
		"user":    user,
	})
}

// Login — вход
func (h *AuthHandler) Login(c echo.Context) error {
	var req models.UserLogin
	log := middleware.GetLoggerFromCtx(c.Request().Context())

	if err := c.Bind(&req); err != nil {
		log.Warnw("Bind failed", "error", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request body"})
	}

	if err := h.validator.Struct(req); err != nil {
		log.Warnw("Validation failed", "error", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	tokens, err := h.service.Login(c.Request().Context(), req)
	if err != nil {
		log.Infow("Login failed", "email", req.Email, "error", err)
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid credentials"})
	}

	log.Infow("Login successful", "email", req.Email)
	return c.JSON(http.StatusOK, echo.Map{
		"message": "login successful",
		"tokens":  tokens,
	})
}

// RefreshToken — обновление пары токенов
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req models.RefreshTokenRequest
	log := middleware.GetLoggerFromCtx(c.Request().Context())

	if err := c.Bind(&req); err != nil {
		log.Warnw("Bind failed", "error", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request body"})
	}

	if err := h.validator.Struct(req); err != nil {
		log.Warnw("Validation failed", "error", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	newTokens, err := h.service.RefreshToken(c.Request().Context(), req)
	if err != nil {
		log.Warnw("Refresh failed", "error", err)
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid or revoked refresh token"})
	}

	log.Infow("Token refreshed successfully")
	return c.JSON(http.StatusOK, echo.Map{
		"message": "token refreshed",
		"tokens":  newTokens,
	})
}

// Logout — отзыв текущего refresh токена (по jti из access токена)
func (h *AuthHandler) Logout(c echo.Context) error {
	claims, ok := middleware.GetUserClaims(c.Request().Context())
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}

	log := middleware.GetLoggerFromCtx(c.Request().Context())

	// Отзываем по jti из access токена
	if err := h.service.RevokeByJTI(c.Request().Context(), claims.JTI); err != nil {
		log.Errorw("Logout failed", "jti", claims.JTI, "user_id", claims.Sub, "error", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "logout failed"})
	}

	log.Infow("User logged out", "user_id", claims.Sub, "jti", claims.JTI)
	return c.JSON(http.StatusOK, echo.Map{"message": "logged out successfully"})
}

func (h *AuthHandler) Validate(c echo.Context) error {
	// 1. Достаем claims из контекста ЗАПРОСА (так как middleware положил их туда через context.WithValue)
	// Важно: тип должен точно совпадать с тем, что возвращает tokenSvc.ParseAccess
	claims, ok := c.Request().Context().Value(middleware.UserClaimsKey).(*models.AccessTokenClaims)

	if !ok {
		// Если claims нет, значит что-то пошло не так в цепочке middleware
		// Nginx получит 401 и заблокирует запрос
		return c.NoContent(http.StatusUnauthorized)
	}

	// 2. Устанавливаем заголовок, который Nginx перехватит и пробросит в profile-service
	// В твоем middleware используется claims.Sub (судя по логам),
	// убедись, что в модели это поле называется так же.
	c.Response().Header().Set("X-User-ID", claims.Sub)

	// 3. Возвращаем 200 OK. Для Nginx это сигнал: "Пропускай!"
	return c.NoContent(http.StatusOK)
}
