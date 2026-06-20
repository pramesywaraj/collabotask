package auth

import (
	"collabotask/internal/domain"
	"context"

	"github.com/google/uuid"
)

func (u *AuthUseCase) GetProfile(ctx context.Context, userID uuid.UUID) (*UserProfile, error) {
	user, err := u.userRepo.GetById(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	result := userToProfile(user)

	return &result, nil
}
