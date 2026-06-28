package auth

import (
	"collabotask/internal/domain"
	"context"
	"fmt"
)

func (u *AuthUseCase) Login(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	if err := ValidateLoginInput(input); err != nil {
		return nil, err
	}

	user, err := u.userRepo.GetByEmail(ctx, normalizeEmail(input.Email))
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	if user == nil {
		return nil, domain.ErrInvalidCredentials
	}

	if !u.hasher.Check(input.Password, user.PasswordHash) {
		return nil, domain.ErrInvalidCredentials
	}

	token, err := u.tokens.Generate(user.ID, string(user.SystemRole))
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginOutput{
		User:  userToProfile(user),
		Token: token,
	}, nil
}
