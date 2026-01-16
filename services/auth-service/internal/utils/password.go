package utils

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// hashPassword — bcrypt хэширование
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt failed: %w", err)
	}
	return string(hash), nil
}

// comparePassword — сравнение хэша
func ComparePassword(hash, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return fmt.Errorf("password mismatch")
	}
	return nil
}
