// internal/models/user.go
package models

import (
	"time"

	"github.com/google/uuid"
)

// User — основная сущность в БД
type User struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	Email             string     `json:"email" db:"email"`
	PasswordHash      string     `json:"-" db:"password_hash"`
	IsVerified        bool       `json:"isVerified" db:"is_verified"`
	EmailVerifyToken  *string    `json:"-" db:"email_verify_token"`
	EmailVerifySentAt *time.Time `json:"-" db:"email_verify_sent_at"`
	Role              string     `json:"role" db:"role"`     // user, admin
	Status            string     `json:"status" db:"status"` // active, blocked
	LastLoginAt       *time.Time `json:"lastLoginAt,omitempty" db:"last_login_at"`
	CreatedAt         time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt         time.Time  `json:"updatedAt" db:"updated_at"`
}
