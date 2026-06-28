package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"collabotask/internal/domain"
	"collabotask/internal/mocks"
	"collabotask/internal/usecase/auth"
)

func TestRegister(t *testing.T) {
	validEmail := "alice@example.com"
	validName := "Alice"
	validPassword := "secret123"

	tests := []struct {
		name        string
		input       auth.RegisterInput
		setupMocks  func(userRepo *mocks.MockUserRepository, hasher *mocks.MockPasswordHasher, tokens *mocks.MockTokenGenerator)
		wantErr     error
		wantErrMsg  string
		wantSuccess bool
	}{
		{
			name:  "invalid email format → validation error",
			input: auth.RegisterInput{Email: "notanemail", Name: validName, Password: validPassword},
			setupMocks: func(u *mocks.MockUserRepository, h *mocks.MockPasswordHasher, tk *mocks.MockTokenGenerator) {
			},
			wantErrMsg: "invalid email format",
		},
		{
			name:  "password shorter than 8 chars → validation error",
			input: auth.RegisterInput{Email: validEmail, Name: validName, Password: "short"},
			setupMocks: func(u *mocks.MockUserRepository, h *mocks.MockPasswordHasher, tk *mocks.MockTokenGenerator) {
			},
			wantErrMsg: "at least 8 characters",
		},
		{
			name:  "empty name → validation error",
			input: auth.RegisterInput{Email: validEmail, Name: "", Password: validPassword},
			setupMocks: func(u *mocks.MockUserRepository, h *mocks.MockPasswordHasher, tk *mocks.MockTokenGenerator) {
			},
			wantErrMsg: "name is required",
		},
		{
			name:  "ExistsByEmail returns DB error → wrapped error",
			input: auth.RegisterInput{Email: validEmail, Name: validName, Password: validPassword},
			setupMocks: func(u *mocks.MockUserRepository, h *mocks.MockPasswordHasher, tk *mocks.MockTokenGenerator) {
				u.EXPECT().ExistsByEmail(context.Background(), validEmail).Return(false, errors.New("db timeout"))
			},
			wantErrMsg: "failed to check email",
		},
		{
			name:  "email already exists → ErrEmailAlreadyExists",
			input: auth.RegisterInput{Email: validEmail, Name: validName, Password: validPassword},
			setupMocks: func(u *mocks.MockUserRepository, h *mocks.MockPasswordHasher, tk *mocks.MockTokenGenerator) {
				u.EXPECT().ExistsByEmail(context.Background(), validEmail).Return(true, nil)
			},
			wantErr: domain.ErrEmailAlreadyExists,
		},
		{
			name:  "hasher.Hash fails → error",
			input: auth.RegisterInput{Email: validEmail, Name: validName, Password: validPassword},
			setupMocks: func(u *mocks.MockUserRepository, h *mocks.MockPasswordHasher, tk *mocks.MockTokenGenerator) {
				u.EXPECT().ExistsByEmail(context.Background(), validEmail).Return(false, nil)
				h.EXPECT().Hash(validPassword).Return("", errors.New("bcrypt error"))
			},
			wantErrMsg: "bcrypt error",
		},
		{
			name:  "userRepo.Create fails with ErrEmailAlreadyExists (race condition) → ErrEmailAlreadyExists",
			input: auth.RegisterInput{Email: validEmail, Name: validName, Password: validPassword},
			setupMocks: func(u *mocks.MockUserRepository, h *mocks.MockPasswordHasher, tk *mocks.MockTokenGenerator) {
				u.EXPECT().ExistsByEmail(context.Background(), validEmail).Return(false, nil)
				h.EXPECT().Hash(validPassword).Return("$hashed$", nil)
				u.EXPECT().Create(context.Background(), mock.Anything).Return(domain.ErrEmailAlreadyExists)
			},
			wantErr: domain.ErrEmailAlreadyExists,
		},
		{
			name:  "userRepo.Create fails with unexpected error → wrapped error",
			input: auth.RegisterInput{Email: validEmail, Name: validName, Password: validPassword},
			setupMocks: func(u *mocks.MockUserRepository, h *mocks.MockPasswordHasher, tk *mocks.MockTokenGenerator) {
				u.EXPECT().ExistsByEmail(context.Background(), validEmail).Return(false, nil)
				h.EXPECT().Hash(validPassword).Return("$hashed$", nil)
				u.EXPECT().Create(context.Background(), mock.Anything).Return(errors.New("db error"))
			},
			wantErrMsg: "failed to create user",
		},
		{
			name:  "tokens.Generate fails → wrapped error",
			input: auth.RegisterInput{Email: validEmail, Name: validName, Password: validPassword},
			setupMocks: func(u *mocks.MockUserRepository, h *mocks.MockPasswordHasher, tk *mocks.MockTokenGenerator) {
				u.EXPECT().ExistsByEmail(context.Background(), validEmail).Return(false, nil)
				h.EXPECT().Hash(validPassword).Return("$hashed$", nil)
				u.EXPECT().Create(context.Background(), mock.Anything).Return(nil)
				tk.EXPECT().Generate(uuid.Nil, "USER").Return("", errors.New("signing error"))
			},
			wantErrMsg: "failed to generate token",
		},
		{
			name:  "success → returns UserProfile + token",
			input: auth.RegisterInput{Email: validEmail, Name: validName, Password: validPassword},
			setupMocks: func(u *mocks.MockUserRepository, h *mocks.MockPasswordHasher, tk *mocks.MockTokenGenerator) {
				u.EXPECT().ExistsByEmail(context.Background(), validEmail).Return(false, nil)
				h.EXPECT().Hash(validPassword).Return("$hashed$", nil)
				u.EXPECT().Create(context.Background(), mock.Anything).Return(nil)
				tk.EXPECT().Generate(uuid.Nil, "USER").Return("jwt.token.here", nil)
			},
			wantSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := mocks.NewMockUserRepository(t)
			hasher := mocks.NewMockPasswordHasher(t)
			tokens := mocks.NewMockTokenGenerator(t)

			tt.setupMocks(userRepo, hasher, tokens)

			uc := auth.NewAuthUseCase(userRepo, hasher, tokens)
			out, err := uc.Register(context.Background(), tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, out)
			} else if tt.wantErrMsg != "" {
				require.ErrorContains(t, err, tt.wantErrMsg)
				assert.Nil(t, out)
			} else {
				require.NoError(t, err)
				require.NotNil(t, out)
				assert.Equal(t, validEmail, out.User.Email)
				assert.Equal(t, validName, out.User.Name)
				assert.Equal(t, "USER", out.User.SystemRole)
				assert.Equal(t, "jwt.token.here", out.Token)
			}
		})
	}
}
