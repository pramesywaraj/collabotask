package auth_test

import (
	"context"
	"errors"
	"strings"
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

func TestRegister(t *testing.T) {
	validEmail := "alice@example.com"
	validName := "Alice"
	validPassword := "secret123"
	userID := uuid.New()

	// expectCreate asserts the entity handed to Create is built correctly — the
	// password is the hasher's output, the email/name are normalized, and the
	// system role is USER. It also populates the DB-generated ID the way the real
	// repository does, so token generation and the returned profile can be checked
	// against a concrete ID instead of uuid.Nil.
	expectCreate := func(u *mocks.MockUserRepository, wantEmail, wantName string) {
		u.EXPECT().Create(mock.Anything, mock.MatchedBy(func(usr *entity.User) bool {
			return usr.Email == wantEmail &&
				usr.Name == wantName &&
				usr.PasswordHash == "$hashed$" &&
				usr.SystemRole == entity.SystemRoleUser
		})).Run(func(_ context.Context, usr *entity.User) {
			usr.ID = userID
		}).Return(nil)
	}

	tests := []struct {
		name       string
		input      auth.RegisterInput
		setupMocks func(userRepo *mocks.MockUserRepository, hasher *mocks.MockPasswordHasher, tokens *mocks.MockTokenGenerator)
		wantErr    error
		wantErrMsg string
		wantEmail  string
		wantName   string
	}{
		{
			name:  "invalid email format → validation error",
			input: auth.RegisterInput{Email: "notanemail", Name: validName, Password: validPassword},
			setupMocks: func(u *mocks.MockUserRepository, h *mocks.MockPasswordHasher, tk *mocks.MockTokenGenerator) {
			},
			wantErrMsg: "invalid email format",
		},
		{
			name:  "empty email → validation error",
			input: auth.RegisterInput{Email: "", Name: validName, Password: validPassword},
			setupMocks: func(u *mocks.MockUserRepository, h *mocks.MockPasswordHasher, tk *mocks.MockTokenGenerator) {
			},
			wantErrMsg: "email is required",
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
			name:  "name longer than 255 chars → validation error",
			input: auth.RegisterInput{Email: validEmail, Name: strings.Repeat("a", 256), Password: validPassword},
			setupMocks: func(u *mocks.MockUserRepository, h *mocks.MockPasswordHasher, tk *mocks.MockTokenGenerator) {
			},
			wantErrMsg: "at most 255 characters",
		},
		{
			name:  "ExistsByEmail returns DB error → wrapped error",
			input: auth.RegisterInput{Email: validEmail, Name: validName, Password: validPassword},
			setupMocks: func(u *mocks.MockUserRepository, h *mocks.MockPasswordHasher, tk *mocks.MockTokenGenerator) {
				u.EXPECT().ExistsByEmail(mock.Anything, validEmail).Return(false, errors.New("db timeout"))
			},
			wantErrMsg: "failed to check email",
		},
		{
			name:  "email already exists → ErrEmailAlreadyExists",
			input: auth.RegisterInput{Email: validEmail, Name: validName, Password: validPassword},
			setupMocks: func(u *mocks.MockUserRepository, h *mocks.MockPasswordHasher, tk *mocks.MockTokenGenerator) {
				u.EXPECT().ExistsByEmail(mock.Anything, validEmail).Return(true, nil)
			},
			wantErr: domain.ErrEmailAlreadyExists,
		},
		{
			name:  "hasher.Hash fails → error",
			input: auth.RegisterInput{Email: validEmail, Name: validName, Password: validPassword},
			setupMocks: func(u *mocks.MockUserRepository, h *mocks.MockPasswordHasher, tk *mocks.MockTokenGenerator) {
				u.EXPECT().ExistsByEmail(mock.Anything, validEmail).Return(false, nil)
				h.EXPECT().Hash(validPassword).Return("", errors.New("bcrypt error"))
			},
			wantErrMsg: "bcrypt error",
		},
		{
			name:  "userRepo.Create fails with ErrEmailAlreadyExists (race condition) → ErrEmailAlreadyExists",
			input: auth.RegisterInput{Email: validEmail, Name: validName, Password: validPassword},
			setupMocks: func(u *mocks.MockUserRepository, h *mocks.MockPasswordHasher, tk *mocks.MockTokenGenerator) {
				u.EXPECT().ExistsByEmail(mock.Anything, validEmail).Return(false, nil)
				h.EXPECT().Hash(validPassword).Return("$hashed$", nil)
				u.EXPECT().Create(mock.Anything, mock.Anything).Return(domain.ErrEmailAlreadyExists)
			},
			wantErr: domain.ErrEmailAlreadyExists,
		},
		{
			name:  "userRepo.Create fails with unexpected error → wrapped error",
			input: auth.RegisterInput{Email: validEmail, Name: validName, Password: validPassword},
			setupMocks: func(u *mocks.MockUserRepository, h *mocks.MockPasswordHasher, tk *mocks.MockTokenGenerator) {
				u.EXPECT().ExistsByEmail(mock.Anything, validEmail).Return(false, nil)
				h.EXPECT().Hash(validPassword).Return("$hashed$", nil)
				u.EXPECT().Create(mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			wantErrMsg: "failed to create user",
		},
		{
			name:  "tokens.Generate fails → wrapped error",
			input: auth.RegisterInput{Email: validEmail, Name: validName, Password: validPassword},
			setupMocks: func(u *mocks.MockUserRepository, h *mocks.MockPasswordHasher, tk *mocks.MockTokenGenerator) {
				u.EXPECT().ExistsByEmail(mock.Anything, validEmail).Return(false, nil)
				h.EXPECT().Hash(validPassword).Return("$hashed$", nil)
				expectCreate(u, validEmail, validName)
				tk.EXPECT().Generate(userID, "USER").Return("", errors.New("signing error"))
			},
			wantErrMsg: "failed to generate token",
		},
		{
			name:  "success → returns UserProfile + token",
			input: auth.RegisterInput{Email: validEmail, Name: validName, Password: validPassword},
			setupMocks: func(u *mocks.MockUserRepository, h *mocks.MockPasswordHasher, tk *mocks.MockTokenGenerator) {
				u.EXPECT().ExistsByEmail(mock.Anything, validEmail).Return(false, nil)
				h.EXPECT().Hash(validPassword).Return("$hashed$", nil)
				expectCreate(u, validEmail, validName)
				tk.EXPECT().Generate(userID, "USER").Return("jwt.token.here", nil)
			},
			wantEmail: validEmail,
			wantName:  validName,
		},
		{
			name:  "success → normalizes email case and trims name before storing",
			input: auth.RegisterInput{Email: "Alice@Example.COM", Name: "  Alice  ", Password: validPassword},
			setupMocks: func(u *mocks.MockUserRepository, h *mocks.MockPasswordHasher, tk *mocks.MockTokenGenerator) {
				// The existence check and the stored email must both use the normalized
				// form, so a differently-cased duplicate cannot slip past the check.
				u.EXPECT().ExistsByEmail(mock.Anything, "alice@example.com").Return(false, nil)
				h.EXPECT().Hash(validPassword).Return("$hashed$", nil)
				expectCreate(u, "alice@example.com", "Alice")
				tk.EXPECT().Generate(userID, "USER").Return("jwt.token.here", nil)
			},
			wantEmail: "alice@example.com",
			wantName:  "Alice",
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
				assert.Equal(t, userID, out.User.ID)
				assert.Equal(t, tt.wantEmail, out.User.Email)
				assert.Equal(t, tt.wantName, out.User.Name)
				assert.Equal(t, "USER", out.User.SystemRole)
				assert.Equal(t, "jwt.token.here", out.Token)
			}
		})
	}
}
