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
	"collabotask/internal/domain/entity"
	"collabotask/internal/mocks"
	"collabotask/internal/usecase/auth"
)

func TestLogin(t *testing.T) {
	userID := uuid.New()
	email := "alice@example.com"
	password := "secret123"
	hashedPassword := "$2a$hashed$"

	aUser := func() *entity.User {
		return &entity.User{
			ID:           userID,
			Email:        email,
			Name:         "Alice",
			PasswordHash: hashedPassword,
			SystemRole:   entity.SystemRoleUser,
		}
	}

	tests := []struct {
		name       string
		input      auth.LoginInput
		setupMocks func(userRepo *mocks.MockUserRepository, hasher *mocks.MockPasswordHasher, tokens *mocks.MockTokenGenerator)
		wantErr    error
		wantErrMsg string
	}{
		{
			name:  "empty email → validation error",
			input: auth.LoginInput{Email: "", Password: password},
			setupMocks: func(u *mocks.MockUserRepository, h *mocks.MockPasswordHasher, tk *mocks.MockTokenGenerator) {
			},
			wantErrMsg: "email is required",
		},
		{
			name:  "empty password → validation error",
			input: auth.LoginInput{Email: email, Password: ""},
			setupMocks: func(u *mocks.MockUserRepository, h *mocks.MockPasswordHasher, tk *mocks.MockTokenGenerator) {
			},
			wantErrMsg: "password is required",
		},
		{
			name:  "GetByEmail returns error → ErrInvalidCredentials",
			input: auth.LoginInput{Email: email, Password: password},
			setupMocks: func(u *mocks.MockUserRepository, h *mocks.MockPasswordHasher, tk *mocks.MockTokenGenerator) {
				u.EXPECT().GetByEmail(mock.Anything, email).Return(nil, errors.New("not found"))
			},
			wantErr: domain.ErrInvalidCredentials,
		},
		{
			name:  "GetByEmail returns nil user → ErrInvalidCredentials",
			input: auth.LoginInput{Email: email, Password: password},
			setupMocks: func(u *mocks.MockUserRepository, h *mocks.MockPasswordHasher, tk *mocks.MockTokenGenerator) {
				u.EXPECT().GetByEmail(mock.Anything, email).Return(nil, nil)
			},
			wantErr: domain.ErrInvalidCredentials,
		},
		{
			name:  "wrong password → ErrInvalidCredentials",
			input: auth.LoginInput{Email: email, Password: password},
			setupMocks: func(u *mocks.MockUserRepository, h *mocks.MockPasswordHasher, tk *mocks.MockTokenGenerator) {
				u.EXPECT().GetByEmail(mock.Anything, email).Return(aUser(), nil)
				h.EXPECT().Check(password, hashedPassword).Return(false)
			},
			wantErr: domain.ErrInvalidCredentials,
		},
		{
			name:  "tokens.Generate fails → wrapped error",
			input: auth.LoginInput{Email: email, Password: password},
			setupMocks: func(u *mocks.MockUserRepository, h *mocks.MockPasswordHasher, tk *mocks.MockTokenGenerator) {
				u.EXPECT().GetByEmail(mock.Anything, email).Return(aUser(), nil)
				h.EXPECT().Check(password, hashedPassword).Return(true)
				tk.EXPECT().Generate(userID, "USER").Return("", errors.New("signing error"))
			},
			wantErrMsg: "failed to generate token",
		},
		{
			name:  "success → returns UserProfile + token",
			input: auth.LoginInput{Email: email, Password: password},
			setupMocks: func(u *mocks.MockUserRepository, h *mocks.MockPasswordHasher, tk *mocks.MockTokenGenerator) {
				u.EXPECT().GetByEmail(mock.Anything, email).Return(aUser(), nil)
				h.EXPECT().Check(password, hashedPassword).Return(true)
				tk.EXPECT().Generate(userID, "USER").Return("jwt.token.here", nil)
			},
		},
		{
			name:  "success → normalizes email before lookup",
			input: auth.LoginInput{Email: "Alice@Example.COM", Password: password},
			setupMocks: func(u *mocks.MockUserRepository, h *mocks.MockPasswordHasher, tk *mocks.MockTokenGenerator) {
				// Login lowercases/trims the email before the lookup, so the repo must
				// be queried with the normalized form regardless of input casing.
				u.EXPECT().GetByEmail(mock.Anything, email).Return(aUser(), nil)
				h.EXPECT().Check(password, hashedPassword).Return(true)
				tk.EXPECT().Generate(userID, "USER").Return("jwt.token.here", nil)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			userRepo := mocks.NewMockUserRepository(t)
			hasher := mocks.NewMockPasswordHasher(t)
			tokens := mocks.NewMockTokenGenerator(t)

			tt.setupMocks(userRepo, hasher, tokens)

			uc := auth.NewAuthUseCase(userRepo, hasher, tokens)
			out, err := uc.Login(context.Background(), tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, out)
			} else if tt.wantErrMsg != "" {
				require.ErrorContains(t, err, tt.wantErrMsg)
				assert.Nil(t, out)
			} else {
				require.NoError(t, err)
				require.NotNil(t, out)
				assert.Equal(t, userID, out.User.ID)
				assert.Equal(t, email, out.User.Email)
				assert.Equal(t, "Alice", out.User.Name)
				assert.Equal(t, "USER", out.User.SystemRole)
				assert.Equal(t, "jwt.token.here", out.Token)
			}
		})
	}
}
