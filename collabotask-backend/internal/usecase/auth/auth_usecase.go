package auth

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"collabotask/internal/config"
	"collabotask/internal/domain/entity"
	"collabotask/internal/domain/repository"
	infraauth "collabotask/internal/infrastructure/auth"

	"github.com/google/uuid"
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

func (u *AuthUseCaseImpl) Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	if err := validateRegisterInput(input); err != nil {
		return nil, err
	}

	exists, err := u.userRepo.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("email already exists")
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
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	token, err := infraauth.GenerateToken(u.authCfg, user.ID, string(user.SystemRole))
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &RegisterOutput{
		User:  userToDTO(user),
		Token: token,
	}, nil
}

func (u *AuthUseCaseImpl) Login(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	if err := validateLoginInput(input); err != nil {
		return nil, err
	}

	user, err := u.userRepo.GetByEmail(ctx, strings.TrimSpace(strings.ToLower(input.Email)))
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	if !infraauth.CheckPassword(input.Password, user.PasswordHash) {
		return nil, fmt.Errorf("invalid email or password")
	}

	token, err := infraauth.GenerateToken(u.authCfg, user.ID, string(user.SystemRole))
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginOutput{
		User:  userToDTO(user),
		Token: token,
	}, nil
}

func (u *AuthUseCaseImpl) GetProfile(ctx context.Context, userID uuid.UUID) (*UserDTO, error) {
	user, err := u.userRepo.GetById(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	result := userToDTO(user)

	return &result, nil
}

func validateRegisterInput(input RegisterInput) error {
	if input.Email == "" {
		return fmt.Errorf("email is required")
	}
	if !emailRegex.MatchString(input.Email) {
		return fmt.Errorf("invalid email format")
	}
	email := strings.TrimSpace(strings.ToLower(input.Email))
	if email == "" {
		return fmt.Errorf("email is required")
	}
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return fmt.Errorf("name is required")
	}
	if len(name) > 255 {
		return fmt.Errorf("name must be at most 255 characters")
	}
	if input.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(input.Password) < minPasswordLen {
		return fmt.Errorf("password must be at least %d characters", minPasswordLen)
	}
	return nil
}

func validateLoginInput(input LoginInput) error {
	if input.Email == "" {
		return fmt.Errorf("email is required")
	}
	if input.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}

func userToDTO(user *entity.User) UserDTO {
	return UserDTO{
		ID:         user.ID,
		Email:      user.Email,
		Name:       user.Name,
		AvatarURL:  user.AvatarURL,
		SystemRole: string(user.SystemRole),
	}
}
