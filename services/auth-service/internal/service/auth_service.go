package service

import (
	"auth-service/internal/models"
	"auth-service/internal/repository"
	"auth-service/internal/utils"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AuthService — бизнес-логика аутентификации
type AuthService interface {
	RegisterUser(ctx context.Context, req models.UserRegister) (*models.User, error)
	Login(ctx context.Context, req models.UserLogin) (*models.TokenPair, error)
	RefreshToken(ctx context.Context, req models.RefreshTokenRequest) (*models.TokenPair, error)
	RevokeByJTI(ctx context.Context, jti string) error
}

// authService — реализация
type authService struct {
	userRepo   repository.UserRepository
	outboxRepo repository.OutboxRepository
	tokenSvc   utils.TokenService
	logger     *zap.SugaredLogger
}

// NewAuthService — конструктор
func NewAuthService(
	userRepo repository.UserRepository,
	outboxRepo repository.OutboxRepository,
	tokenSvc utils.TokenService,
	logger *zap.SugaredLogger,
) AuthService {
	return &authService{
		userRepo:   userRepo,
		outboxRepo: outboxRepo,
		tokenSvc:   tokenSvc,
		logger:     logger,
	}
}

// RegisterUser — регистрация нового пользователя с Outbox Pattern
func (s *authService) RegisterUser(ctx context.Context, req models.UserRegister) (*models.User, error) {
	log := s.logger.With("email", req.Email)

	// Проверка: email уже занят?
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		log.Errorw("Database error: email check failed", "error", err)
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if exists {
		log.Warnw("Registration failed: email already in use")
		return nil, fmt.Errorf("email already in use")
	}

	// НАЧИНАЕМ ТРАНЗАКЦИЮ
	tx, err := s.userRepo.BeginTx(ctx)
	if err != nil {
		log.Errorw("Failed to begin transaction", "error", err)
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				log.Errorw("Failed to rollback transaction", "error", rollbackErr)
			}
		}
	}()

	// Хэшируем пароль
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		log.Errorw("Password hashing failed", "error", err)
		return nil, fmt.Errorf("password hashing failed: %w", err)
	}

	// Создаём пользователя
	user := &models.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         "user",
		Status:       "active",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// СОХРАНЯЕМ ПОЛЬЗОВАТЕЛЯ в транзакции
	if err = s.userRepo.CreateTx(ctx, tx, user); err != nil {
		log.Errorw("Failed to create user in database", "error", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// СОЗДАЁМ СОБЫТИЕ UserRegistered
	userRegisteredEvent := models.UserRegistered{
		UserID:    user.ID,
		Email:     user.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      user.Role,
		CreatedAt: time.Now(),
	}

	eventPayload, err := json.Marshal(userRegisteredEvent)
	if err != nil {
		log.Errorw("Failed to marshal UserRegistered event", "error", err)
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	outboxEvent := &models.OutboxEvent{
		ID:        uuid.New(),
		EventType: "UserRegistered",
		Payload:   eventPayload,
	}

	// СОХРАНЯЕМ СОБЫТИЕ в Outbox
	if err = s.outboxRepo.InsertTx(ctx, tx, outboxEvent); err != nil {
		log.Errorw("Failed to insert outbox event", "error", err)
		return nil, fmt.Errorf("failed to create outbox event: %w", err)
	}

	// COMMIT
	if err = tx.Commit(ctx); err != nil {
		log.Errorw("Failed to commit transaction", "error", err)
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Infow("User registered successfully with outbox event",
		"user_id", user.ID,
		"outbox_event_id", outboxEvent.ID,
		"event_type", outboxEvent.EventType)

	return user, nil
}

// Login — вход пользователя
func (s *authService) Login(ctx context.Context, req models.UserLogin) (*models.TokenPair, error) {
	log := s.logger.With("email", req.Email)

	// Находим пользователя
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if err.Error() == "not found" {
			log.Infow("Login failed: user not found")
			return nil, fmt.Errorf("invalid email or password")
		}
		log.Errorw("Database error during login", "error", err)
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Проверяем пароль
	if err := utils.ComparePassword(user.PasswordHash, req.Password); err != nil {
		log.Infow("Login failed: invalid password")
		return nil, fmt.Errorf("invalid email or password")
	}

	// Обновляем last_login_at
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		log.Warnw("Failed to update last login", "user_id", user.ID, "error", err)
	}

	// Генерируем токены
	pair, err := s.tokenSvc.GeneratePair(ctx, user.ID, user.Email, user.Role)
	if err != nil {
		log.Errorw("Token generation failed", "user_id", user.ID, "error", err)
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	log.Infow("Login successful", "user_id", user.ID)
	return pair, nil
}

// RefreshToken — ротация токенов
func (s *authService) RefreshToken(ctx context.Context, req models.RefreshTokenRequest) (*models.TokenPair, error) {
	log := s.logger

	// Парсим refresh токен
	claims, err := s.tokenSvc.ParseRefresh(req.RefreshToken)
	if err != nil {
		log.Warnw("Invalid refresh token", "error", err)
		return nil, fmt.Errorf("invalid or revoked refresh token: %w", err)
	}

	userID, err := uuid.Parse(claims.Sub)
	if err != nil {
		log.Warnw("Invalid user ID in token", "sub", claims.Sub)
		return nil, fmt.Errorf("invalid user ID in token")
	}

	// Проверяем пользователя
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err.Error() == "not found" {
			log.Warnw("User not found during refresh", "user_id", userID)
			return nil, fmt.Errorf("user not found")
		}
		log.Errorw("Database error during refresh", "error", err)
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Ротируем токены
	newPair, err := s.tokenSvc.RotateRefresh(ctx, req.RefreshToken, user.ID, user.Email, user.Role)
	if err != nil {
		log.Errorw("Token rotation failed", "old_jti", claims.JTI, "error", err)
		return nil, fmt.Errorf("token rotation failed: %w", err)
	}

	log.Infow("Token refreshed", "user_id", user.ID, "old_jti", claims.JTI)
	return newPair, nil
}

// RevokeByJTI — отзыв токена по jti
func (s *authService) RevokeByJTI(ctx context.Context, jti string) error {
	log := s.logger.With("jti", jti)

	if err := s.tokenSvc.RevokeRefresh(ctx, jti); err != nil {
		log.Errorw("Failed to revoke token", "error", err)
		return fmt.Errorf("revocation failed: %w", err)
	}

	log.Infow("Token revoked successfully")
	return nil
}
