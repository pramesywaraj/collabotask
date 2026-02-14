package auth

import (
	"context"

	"github.com/google/uuid"
)

type UserDTO struct {
	ID         uuid.UUID `json:"id"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	AvatarURL  *string   `json:"avatar_url"`
	SystemRole string    `json:"system_role"`
}

type RegisterInput struct {
	Email    string
	Name     string
	Password string
}

type RegisterOutput struct {
	User  UserDTO
	Token string
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	User  UserDTO
	Token string
}

type AuthUseCase interface {
	Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error)
	Login(ctx context.Context, input LoginInput) (*LoginOutput, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*UserDTO, error)
}
