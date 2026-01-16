package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenPair — пара токенов, возвращаемая клиенту
type TokenPair struct {
	AccessToken  string `json:"access_token"`  // JWT access token
	RefreshToken string `json:"refresh_token"` // JWT refresh token
	ExpiresIn    int64  `json:"expires_in"`    // Время жизни access в секундах (Unix)
	TokenType    string `json:"token_type"`    // Всегда "Bearer"
}

// AccessTokenClaims — claims для access токена (минималистичный, но информативный)
type AccessTokenClaims struct {
	jwt.RegisteredClaims

	// Обязательные поля по RFC 7519
	Sub string `json:"sub"` // Subject — ID пользователя
	JTI string `json:"jti"` // JWT ID — уникальный ID токена

	// Прикладные поля
	Email string `json:"email,omitempty"` // Опционально
	Role  string `json:"role,omitempty"`  // Опционально
}

// RefreshTokenClaims — claims для refresh токена (минималистичный!)
type RefreshTokenClaims struct {
	jwt.RegisteredClaims

	// Только то, что нужно для безопасного обновления
	Sub string `json:"sub"` // Subject — ID пользователя
	JTI string `json:"jti"` // JWT ID — для отзыва
}

// RefreshTokenRequest — запрос на обновление токена
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,jwt"`
}

// RevokeTokenRequest — запрос на отзыв токена (для логаута)
type RevokeTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,jwt"`
}

// NewAccessTokenClaims создаёт claims для access токена
func NewAccessTokenClaims(userID, email, role, jti string, issuer string, ttl time.Duration) AccessTokenClaims {
	now := time.Now()
	exp := now.Add(ttl)

	return AccessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Subject:   userID,
			Issuer:    issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			NotBefore: jwt.NewNumericDate(now),
		},
		JTI:   jti,
		Sub:   userID,
		Email: email,
		Role:  role,
	}
}

// NewRefreshTokenClaims создаёт claims для refresh токена
func NewRefreshTokenClaims(userID, jti string, issuer string, ttl time.Duration) RefreshTokenClaims {
	now := time.Now()
	exp := now.Add(ttl)

	return RefreshTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Subject:   userID,
			Issuer:    issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			NotBefore: jwt.NewNumericDate(now),
		},
		JTI: jti,
		Sub: userID,
	}
}
