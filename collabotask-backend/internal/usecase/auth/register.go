package auth

import (
	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/dto"
	infraauth "collabotask/internal/infrastructure/auth"
	"context"
	"errors"
	"fmt"
	"strings"
)

func (u *AuthUseCaseImpl) Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	if err := ValidateRegisterInput(input); err != nil {
		return nil, err
	}

	exists, err := u.userRepo.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if exists {
		return nil, domain.ErrEmailAlreadyExists
	}

	hash, err := infraauth.HashPassword(u.authCfg, input.Password)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		Email:        strings.TrimSpace(strings.ToLower(input.Email)),
		Name:         strings.TrimSpace(input.Name),
		PasswordHash: hash,
		SystemRole:   entity.SystemRoleUser,
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			return nil, domain.ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	token, err := infraauth.GenerateToken(u.authCfg, user.ID, string(user.SystemRole))
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &RegisterOutput{
		User:  dto.UserToDTO(user),
		Token: token,
	}, nil
}
