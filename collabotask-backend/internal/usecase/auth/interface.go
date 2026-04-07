package auth

import (
	"collabotask/internal/dto"
	"context"

	"github.com/google/uuid"
)

type AuthUseCase interface {
	Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error)
	Login(ctx context.Context, input LoginInput) (*LoginOutput, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*dto.UserDTO, error)
}

type RegisterInput struct {
	Email    string
	Name     string
	Password string
}

type RegisterOutput struct {
	User  dto.UserDTO
	Token string
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	User  dto.UserDTO
	Token string
}
