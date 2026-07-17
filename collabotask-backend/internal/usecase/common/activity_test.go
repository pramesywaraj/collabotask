package common_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"collabotask/internal/domain/entity"
	"collabotask/internal/mocks"
	"collabotask/internal/usecase/common"
)

func TestWriteActivity(t *testing.T) {
	boardID := uuid.New()
	entityID := uuid.New()
	actorID := uuid.New()

	newActivity := func() *entity.Activity {
		return &entity.Activity{
			BoardID:    boardID,
			ActionType: entity.ActivityActionCreated,
			EntityType: entity.ActivityEntityCard,
			EntityID:   entityID,
			Metadata:   map[string]any{entity.ActivityMetaCardTitle: "Test"},
		}
	}

	t.Run("success — Log nil, UserID set on activity", func(t *testing.T) {
		repo := mocks.NewMockActivityRepository(t)
		a := newActivity()
		repo.EXPECT().Log(context.Background(), mock.MatchedBy(func(got *entity.Activity) bool {
			return got.UserID != nil && *got.UserID == actorID
		})).Return(nil)
		require.NotPanics(t, func() {
			common.WriteActivity(context.Background(), repo, actorID, a)
		})
	})

	t.Run("resilience — Log error is swallowed (no panic, no return value)", func(t *testing.T) {
		repo := mocks.NewMockActivityRepository(t)
		repo.EXPECT().Log(context.Background(), mock.Anything).Return(errors.New("db down"))
		require.NotPanics(t, func() {
			common.WriteActivity(context.Background(), repo, actorID, newActivity())
		})
	})
}
