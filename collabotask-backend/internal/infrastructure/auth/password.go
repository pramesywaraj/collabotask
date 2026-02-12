package auth

import (
	"collabotask/internal/config"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(cfg *config.AuthConfig, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cfg.BcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hash), nil
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
