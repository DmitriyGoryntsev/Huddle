// internal/models/auth.go
package models

// === Аутентификация ===

type UserLogin struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type UserRegister struct {
	Email     string  `json:"email" validate:"required,email"`
	Password  string  `json:"password" validate:"required,min=6"`
	FirstName string  `json:"firstName" validate:"required,min=1"`
	LastName  string  `json:"lastName" validate:"required,min=1"`
	Phone     *string `json:"phone,omitempty" validate:"omitempty,e164"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" validate:"required,min=6"`
	NewPassword string `json:"newPassword" validate:"required,min=6"`
}

type ResetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}
