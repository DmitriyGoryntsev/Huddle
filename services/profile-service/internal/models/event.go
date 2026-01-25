package models

import (
	"time"

	"github.com/google/uuid"
)

// UserRegistered — событие регистрации пользователя
type UserRegistered struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}
