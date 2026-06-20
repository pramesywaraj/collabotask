package auth

import (
	"collabotask/internal/domain"
	infraauth "collabotask/internal/infrastructure/auth"
	"context"
	"fmt"
	"strings"
)

func (u *AuthUseCase) Login(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	if err := ValidateLoginInput(input); err != nil {
		return nil, err
	}

	user, err := u.userRepo.GetByEmail(ctx, strings.TrimSpace(strings.ToLower(input.Email)))
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if !infraauth.CheckPassword(input.Password, user.PasswordHash) {
		return nil, domain.ErrInvalidCredentials
	}

	token, err := infraauth.GenerateToken(u.authCfg, user.ID, string(user.SystemRole))
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginOutput{
		User:  userToProfile(user),
		Token: token,
	}, nil
}
