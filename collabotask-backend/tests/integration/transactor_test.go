package integration_test

// Integration tests for the Transactor / atomic activity-writes invariant (ADR-016).
//
// Three contract points:
//  1. WithinTransaction rolls back on error — any INSERT inside the closure is absent after rollback.
//  2. Real Log inside a rolled-back tx leaves no activity row (footgun guard: Log uses exec(ctx),
//     which joins the ambient tx, so it sees the same rollback).
//  3. End-to-end: if Log fails (FK violation on a non-existent board_id), the closure error
//     propagates and the mutation row (card) is also absent — proving strict atomicity.

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	pgRepos "collabotask/internal/repository/postgres"
	"collabotask/tests/integration/testutil"
)

// TestTransactor_RollbackOnError verifies that a card INSERT inside a transaction
// that returns an error is never committed.
func TestTransactor_RollbackOnError(t *testing.T) {
	t.Cleanup(func() { testutil.TruncateAll(t, testPool) })

	user := fx.CreateTestUser(t)
	ws := fx.CreateTestWorkspace(t, user.ID)
	board := fx.CreateTestBoard(t, ws.ID, user.ID)
	col := fx.CreateTestColumn(t, board.ID)

	tx := pgRepos.NewTransactor(testPool)
	cardRepo := pgRepos.NewCardRepository(testPool)

	card := &entity.Card{
		ColumnID:  col.ID,
		BoardID:   board.ID,
		Title:     "Should Not Persist",
		Position:  1000,
		CreatedBy: user.ID,
	}

	err := tx.WithinTransaction(context.Background(), func(ctx context.Context) error {
		if err := cardRepo.Create(ctx, card); err != nil {
			return err
		}
		return errors.New("forced rollback")
	})
	require.Error(t, err)
	require.ErrorContains(t, err, "forced rollback")

	_, fetchErr := cardRepo.GetByID(context.Background(), card.ID)
	assert.ErrorIs(t, fetchErr, domain.ErrCardNotFound, "card must not exist after rollback")
}

// TestTransactor_LogHonorsTx verifies that activityRepo.Log uses the ambient tx from
// ctx (via exec(ctx)) — so a successful Log inside a rolled-back tx leaves no row.
func TestTransactor_LogHonorsTx(t *testing.T) {
	t.Cleanup(func() { testutil.TruncateAll(t, testPool) })

	user := fx.CreateTestUser(t)
	ws := fx.CreateTestWorkspace(t, user.ID)
	board := fx.CreateTestBoard(t, ws.ID, user.ID)

	tx := pgRepos.NewTransactor(testPool)
	activityRepo := pgRepos.NewActivityRepository(testPool)

	activity := &entity.Activity{
		BoardID:    board.ID,
		UserID:     &user.ID,
		ActionType: entity.ActivityActionCreated,
		EntityType: entity.ActivityEntityCard,
		EntityID:   uuid.New(),
		Metadata:   map[string]any{"board_title": board.Title},
	}

	err := tx.WithinTransaction(context.Background(), func(ctx context.Context) error {
		if err := activityRepo.Log(ctx, activity); err != nil {
			return err
		}
		return errors.New("force rollback after successful Log")
	})
	require.Error(t, err)

	// The Log call succeeded inside the tx, but the tx was rolled back.
	// The activity row must be absent.
	var count int
	scanErr := testPool.QueryRow(context.Background(),
		`SELECT count(*) FROM activities WHERE id = $1`, activity.ID).Scan(&count)
	require.NoError(t, scanErr)
	assert.Equal(t, 0, count, "activity row must be absent after tx rollback")
}

// TestTransactor_LogFailureRollsMutationBack verifies end-to-end atomicity (ADR-016):
// when Log fails (FK violation on non-existent board_id), the closure error propagates
// AND the card INSERT in the same tx is rolled back.
func TestTransactor_LogFailureRollsMutationBack(t *testing.T) {
	t.Cleanup(func() { testutil.TruncateAll(t, testPool) })

	user := fx.CreateTestUser(t)
	ws := fx.CreateTestWorkspace(t, user.ID)
	board := fx.CreateTestBoard(t, ws.ID, user.ID)
	col := fx.CreateTestColumn(t, board.ID)

	tx := pgRepos.NewTransactor(testPool)
	cardRepo := pgRepos.NewCardRepository(testPool)
	activityRepo := pgRepos.NewActivityRepository(testPool)

	card := &entity.Card{
		ColumnID:  col.ID,
		BoardID:   board.ID,
		Title:     "Atomic Card",
		Position:  1000,
		CreatedBy: user.ID,
	}

	// Passing a non-existent board_id triggers a FK violation in the activities table,
	// causing Log to fail and the closure to return an error.
	badBoardID := uuid.New()
	activity := &entity.Activity{
		BoardID:    badBoardID, // non-existent → FK violation → Log fails
		UserID:     &user.ID,
		ActionType: entity.ActivityActionCreated,
		EntityType: entity.ActivityEntityCard,
		EntityID:   uuid.New(),
		Metadata:   map[string]any{"card_title": card.Title},
	}

	err := tx.WithinTransaction(context.Background(), func(ctx context.Context) error {
		if err := cardRepo.Create(ctx, card); err != nil {
			return err
		}
		return activityRepo.Log(ctx, activity)
	})
	require.Error(t, err, "Log FK failure must propagate from WithinTransaction")

	// The card INSERT must be rolled back: Log failure rolls back the entire tx.
	_, fetchErr := cardRepo.GetByID(context.Background(), card.ID)
	assert.ErrorIs(t, fetchErr, domain.ErrCardNotFound, "card must not exist when Log fails")
}
