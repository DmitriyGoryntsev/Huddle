package utils

import (
	"auth-service/internal/models"
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// TokenService — безопасное управление JWT-токенами с отзывом через Redis
type TokenService interface {
	GeneratePair(ctx context.Context, userID uuid.UUID, email, role string) (*models.TokenPair, error)
	ParseAccess(tokenStr string) (*models.AccessTokenClaims, error)
	ParseRefresh(tokenStr string) (*models.RefreshTokenClaims, error)
	RevokeRefresh(ctx context.Context, jti string) error
	IsRevoked(ctx context.Context, jti string) (bool, error)
	RotateRefresh(ctx context.Context, oldRefreshToken string, userID uuid.UUID, email, role string) (*models.TokenPair, error)
}

// tokenService — реализация
type tokenService struct {
	signingKey []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
	issuer     string
	redis      *redis.Client
	logger     *zap.SugaredLogger
}

// TokenServiceConfig — конфигурация
type TokenServiceConfig struct {
	SigningKey string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	Issuer     string
	Redis      *redis.Client
	Logger     *zap.SugaredLogger
}

// NewTokenService — конструктор с валидацией
func NewTokenService(cfg TokenServiceConfig) (TokenService, error) {
	if cfg.SigningKey == "" {
		return nil, fmt.Errorf("signing key is required")
	}
	if cfg.AccessTTL <= 0 || cfg.RefreshTTL <= 0 {
		return nil, fmt.Errorf("token TTLs must be positive")
	}
	if cfg.Redis == nil {
		return nil, fmt.Errorf("redis client is required for token revocation")
	}
	if cfg.Logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	if cfg.Issuer == "" {
		cfg.Issuer = "auth-service"
	}

	return &tokenService{
		signingKey: []byte(cfg.SigningKey),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
		issuer:     cfg.Issuer,
		redis:      cfg.Redis,
		logger:     cfg.Logger,
	}, nil
}

// keyFunc — возвращает ключ для подписи/проверки
func (s *tokenService) keyFunc(*jwt.Token) (interface{}, error) {
	return s.signingKey, nil
}

// GeneratePair — создаёт пару токенов + сохраняет jti в Redis
func (s *tokenService) GeneratePair(ctx context.Context, userID uuid.UUID, email, role string) (*models.TokenPair, error) {
	jti := uuid.New().String()
	userIDStr := userID.String()

	s.logger.Infow("Generating token pair",
		"user_id", userIDStr,
		"email", email,
		"jti", jti,
	)

	// Access Token
	accessClaims := models.NewAccessTokenClaims(userIDStr, email, role, jti, s.issuer, s.accessTTL)
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	signedAccess, err := accessToken.SignedString(s.signingKey)
	if err != nil {
		s.logger.Errorw("Failed to sign access token", "error", err)
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	// Refresh Token
	refreshClaims := models.NewRefreshTokenClaims(userIDStr, jti, s.issuer, s.refreshTTL)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefresh, err := refreshToken.SignedString(s.signingKey)
	if err != nil {
		s.logger.Errorw("Failed to sign refresh token", "error", err)
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	// Сохраняем jti в Redis (white list)
	if err := s.redis.Set(ctx, "jti:"+jti, userIDStr, s.refreshTTL).Err(); err != nil {
		s.logger.Errorw("Failed to store jti in Redis", "jti", jti, "error", err)
		return nil, fmt.Errorf("failed to store jti in redis: %w", err)
	}

	exp := time.Now().Add(s.accessTTL).Unix()

	s.logger.Infow("Token pair generated successfully", "jti", jti, "expires_in", exp)

	return &models.TokenPair{
		AccessToken:  signedAccess,
		RefreshToken: signedRefresh,
		ExpiresIn:    exp,
		TokenType:    "Bearer",
	}, nil
}

// ParseAccess — парсит и валидирует access токен
func (s *tokenService) ParseAccess(tokenStr string) (*models.AccessTokenClaims, error) {
	claims := &models.AccessTokenClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, s.keyFunc,
		jwt.WithIssuer(s.issuer),
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithLeeway(5*time.Second),
	)
	if err != nil {
		s.logger.Warnw("Invalid access token", "error", err)
		return nil, fmt.Errorf("invalid access token: %w", err)
	}
	return claims, nil
}

// ParseRefresh — парсит refresh токен + проверяет отзыв
func (s *tokenService) ParseRefresh(tokenStr string) (*models.RefreshTokenClaims, error) {
	claims := &models.RefreshTokenClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, s.keyFunc,
		jwt.WithIssuer(s.issuer),
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithLeeway(5*time.Second),
	)
	if err != nil {
		s.logger.Warnw("Failed to parse refresh token", "error", err)
		return nil, fmt.Errorf("parse error: %w", err)
	}
	if !token.Valid {
		s.logger.Warnw("Refresh token is not valid", "jti", claims.JTI)
		return nil, fmt.Errorf("token is not valid")
	}

	// Проверка отзыва
	revoked, err := s.IsRevoked(context.Background(), claims.JTI)
	if err != nil {
		s.logger.Errorw("Redis error during revocation check", "jti", claims.JTI, "error", err)
		return nil, fmt.Errorf("revocation check failed: %w", err)
	}
	if revoked {
		s.logger.Infow("Refresh token is revoked", "jti", claims.JTI, "user_id", claims.Sub)
		return nil, fmt.Errorf("refresh token revoked")
	}

	s.logger.Debugw("Refresh token valid", "jti", claims.JTI, "user_id", claims.Sub)
	return claims, nil
}

// RevokeRefresh — добавляет jti в blacklist
func (s *tokenService) RevokeRefresh(ctx context.Context, jti string) error {
	if err := s.redis.Set(ctx, "revoked:"+jti, "1", 30*24*time.Hour).Err(); err != nil {
		s.logger.Errorw("Failed to revoke token in Redis", "jti", jti, "error", err)
		return err
	}
	s.logger.Infow("Token revoked successfully", "jti", jti)
	return nil
}

// IsRevoked — проверяет, отозван ли токен
func (s *tokenService) IsRevoked(ctx context.Context, jti string) (bool, error) {
	_, err := s.redis.Get(ctx, "revoked:"+jti).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		s.logger.Errorw("Redis error on IsRevoked", "jti", jti, "error", err)
		return false, err
	}
	return true, nil
}

// RotateRefresh — ротация: старый → отозван, новый → выдан
func (s *tokenService) RotateRefresh(ctx context.Context, oldRefreshToken string, userID uuid.UUID, email, role string) (*models.TokenPair, error) {
	// 1. Парсим старый
	oldClaims, err := s.ParseRefresh(oldRefreshToken)
	if err != nil {
		return nil, err
	}

	// 2. Отзываем старый
	if err := s.RevokeRefresh(ctx, oldClaims.JTI); err != nil {
		s.logger.Warnw("Failed to revoke old token during rotation", "old_jti", oldClaims.JTI, "error", err)
		// Не фатально — продолжаем
	}

	// 3. Генерируем новый
	newPair, err := s.GeneratePair(ctx, userID, email, role)
	if err != nil {
		return nil, err
	}

	s.logger.Infow("Token rotation successful",
		"old_jti", oldClaims.JTI,
		"new_jti", newPair.AccessToken[:16]+"...", // маскируем
		"user_id", userID.String(),
	)

	return newPair, nil
}
