package auth

import (
	"fmt"
	"strings"
)

func ValidateLoginInput(input LoginInput) error {
	if input.Email == "" {
		return fmt.Errorf("email is required")
	}
	if input.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}

func ValidateRegisterInput(input RegisterInput) error {
	if input.Email == "" {
		return fmt.Errorf("email is required")
	}
	if !emailRegex.MatchString(input.Email) {
		return fmt.Errorf("invalid email format")
	}
	email := strings.TrimSpace(strings.ToLower(input.Email))
	if email == "" {
		return fmt.Errorf("email is required")
	}
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return fmt.Errorf("name is required")
	}
	if len(name) > 255 {
		return fmt.Errorf("name must be at most 255 characters")
	}
	if input.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(input.Password) < minPasswordLen {
		return fmt.Errorf("password must be at least %d characters", minPasswordLen)
	}
	return nil
}
