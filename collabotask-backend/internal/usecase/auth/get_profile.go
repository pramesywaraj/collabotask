package auth

import (
	"collabotask/internal/domain"
	"collabotask/internal/dto"
	"context"

	"github.com/google/uuid"
)

func (u *AuthUseCaseImpl) GetProfile(ctx context.Context, userID uuid.UUID) (*dto.UserDTO, error) {
	user, err := u.userRepo.GetById(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	result := dto.UserToDTO(user)

	return &result, nil
}
