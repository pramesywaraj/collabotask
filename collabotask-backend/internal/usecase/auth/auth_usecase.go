package auth

import (
	"regexp"

	"collabotask/internal/config"
	"collabotask/internal/domain/repository"
)

const (
	minPasswordLen = 8
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

type AuthUseCaseImpl struct {
	userRepo repository.UserRepository
	authCfg  *config.AuthConfig
}

func NewAuthUseCase(userRepo repository.UserRepository, authCfg *config.AuthConfig) AuthUseCase {
	return &AuthUseCaseImpl{
		userRepo: userRepo,
		authCfg:  authCfg,
	}
}
