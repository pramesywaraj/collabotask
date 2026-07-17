package common_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"collabotask/internal/domain/entity"
	"collabotask/internal/mocks"
	"collabotask/internal/usecase/common"
)

func TestWriteActivity(t *testing.T) {
	boardID := uuid.New()
	entityID := uuid.New()
	userID := uuid.New()

	activity := &entity.Activity{
		BoardID:    boardID,
		UserID:     &userID,
		ActionType: entity.ActivityActionCreated,
		EntityType: entity.ActivityEntityCard,
		EntityID:   entityID,
		Metadata:   map[string]any{"card_title": "Test"},
	}

	t.Run("success — Log nil → no error, no panic", func(t *testing.T) {
		repo := mocks.NewMockActivityRepository(t)
		repo.EXPECT().Log(context.Background(), activity).Return(nil)
		require.NotPanics(t, func() {
			common.WriteActivity(context.Background(), repo, activity)
		})
	})

	t.Run("resilience — Log error is swallowed (no panic, no return value)", func(t *testing.T) {
		repo := mocks.NewMockActivityRepository(t)
		repo.EXPECT().Log(context.Background(), activity).Return(errors.New("db down"))
		require.NotPanics(t, func() {
			common.WriteActivity(context.Background(), repo, activity)
		})
	})
}
