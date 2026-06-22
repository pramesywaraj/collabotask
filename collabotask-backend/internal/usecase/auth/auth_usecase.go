package auth

import (
	"regexp"

	"collabotask/internal/domain/repository"

	"github.com/google/uuid"
)

const (
	minPasswordLen = 8
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Check(password, hash string) bool
}

type TokenGenerator interface {
	Generate(userID uuid.UUID, role string) (string, error)
}

type AuthUseCase struct {
	userRepo repository.UserRepository
	hasher   PasswordHasher
	tokens   TokenGenerator
}

func NewAuthUseCase(
	userRepo repository.UserRepository,
	hasher PasswordHasher,
	tokens TokenGenerator,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo: userRepo,
		hasher:   hasher,
		tokens:   tokens,
	}
}
