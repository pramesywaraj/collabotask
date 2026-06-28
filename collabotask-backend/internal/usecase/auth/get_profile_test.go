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

func TestGetProfile(t *testing.T) {
	userID := uuid.New()
	errDB := errors.New("db error")

	aUser := func() *entity.User {
		return &entity.User{
			ID:         userID,
			Email:      "alice@example.com",
			Name:       "Alice",
			SystemRole: entity.SystemRoleUser,
		}
	}

	tests := []struct {
		name       string
		userID     uuid.UUID
		setupMocks func(userRepo *mocks.MockUserRepository)
		wantErr    error
	}{
		{
			name:   "GetById returns error → error propagated",
			userID: userID,
			setupMocks: func(u *mocks.MockUserRepository) {
				u.EXPECT().GetById(mock.Anything, userID).Return(nil, errDB)
			},
			wantErr: errDB,
		},
		{
			name:   "GetById returns nil user → ErrUserNotFound",
			userID: userID,
			setupMocks: func(u *mocks.MockUserRepository) {
				u.EXPECT().GetById(mock.Anything, userID).Return(nil, nil)
			},
			wantErr: domain.ErrUserNotFound,
		},
		{
			name:   "success → returns UserProfile",
			userID: userID,
			setupMocks: func(u *mocks.MockUserRepository) {
				u.EXPECT().GetById(mock.Anything, userID).Return(aUser(), nil)
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

			tt.setupMocks(userRepo)

			uc := auth.NewAuthUseCase(userRepo, hasher, tokens)
			profile, err := uc.GetProfile(context.Background(), tt.userID)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, profile)
			} else {
				require.NoError(t, err)
				require.NotNil(t, profile)
				assert.Equal(t, userID, profile.ID)
				assert.Equal(t, "alice@example.com", profile.Email)
				assert.Equal(t, "Alice", profile.Name)
				assert.Equal(t, "USER", profile.SystemRole)
			}
		})
	}
}
