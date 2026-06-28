package auth

import (
	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"context"
	"errors"
	"fmt"
	"strings"
)

func (u *AuthUseCase) Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	if err := ValidateRegisterInput(input); err != nil {
		return nil, err
	}

	email := normalizeEmail(input.Email)

	exists, err := u.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if exists {
		return nil, domain.ErrEmailAlreadyExists
	}

	hash, err := u.hasher.Hash(input.Password)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		Email:        email,
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

	token, err := u.tokens.Generate(user.ID, string(user.SystemRole))
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &RegisterOutput{
		User:  userToProfile(user),
		Token: token,
	}, nil
}
