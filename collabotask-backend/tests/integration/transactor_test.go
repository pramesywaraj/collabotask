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

// TestTransactor_TransferOwnershipLogFailureRollsBack covers the second transaction
// shape in the codebase (transfer_ownership.go): boardMemberRepo.TransferOwnership
// — itself a nested tx(ctx) that joins the ambient one — followed by LogActivity.
// When the activity write fails, ownership must not move.
func TestTransactor_TransferOwnershipLogFailureRollsBack(t *testing.T) {
	t.Cleanup(func() { testutil.TruncateAll(t, testPool) })

	oldOwner := fx.CreateTestUser(t)
	ws := fx.CreateTestWorkspace(t, oldOwner.ID)
	board := fx.CreateTestBoard(t, ws.ID, oldOwner.ID) // oldOwner is BOARD_OWNER
	target := fx.CreateTestUser(t)
	fx.AddBoardMember(t, board.ID, target.ID) // target is BOARD_MEMBER

	tx := pgRepos.NewTransactor(testPool)
	memberRepo := pgRepos.NewBoardMemberRepository(testPool)
	activityRepo := pgRepos.NewActivityRepository(testPool)

	err := tx.WithinTransaction(context.Background(), func(ctx context.Context) error {
		fromUserID, txErr := memberRepo.TransferOwnership(ctx, board.ID, target.ID)
		if txErr != nil {
			return txErr
		}
		require.NotNil(t, fromUserID)

		// Non-existent board_id → FK violation in activities → Log fails.
		return activityRepo.Log(ctx, &entity.Activity{
			BoardID:    uuid.New(),
			UserID:     &oldOwner.ID,
			ActionType: entity.ActivityActionOwnershipTransferred,
			EntityType: entity.ActivityEntityBoard,
			EntityID:   board.ID,
			Metadata: map[string]any{
				entity.ActivityMetaFromUserID: oldOwner.ID.String(),
				entity.ActivityMetaToUserID:   target.ID.String(),
			},
		})
	})
	require.Error(t, err, "Log FK failure must propagate from WithinTransaction")

	// Both role writes (demote + promote) must be rolled back.
	oldOwnerBM, getErr := memberRepo.GetMemberByBoardAndUser(context.Background(), board.ID, oldOwner.ID)
	require.NoError(t, getErr)
	assert.Equal(t, entity.BoardRoleOwner, oldOwnerBM.Role, "old owner must still be BOARD_OWNER")

	targetBM, getErr := memberRepo.GetMemberByBoardAndUser(context.Background(), board.ID, target.ID)
	require.NoError(t, getErr)
	assert.Equal(t, entity.BoardRoleMember, targetBM.Role, "target must still be BOARD_MEMBER")

	var activityCount int
	require.NoError(t, testPool.QueryRow(context.Background(),
		`SELECT count(*) FROM activities WHERE board_id = $1`, board.ID).Scan(&activityCount))
	assert.Equal(t, 0, activityCount, "no activity row may survive the rollback")
}

// TestTransactor_TransferOwnershipSuccessCommitsBothWrites is the committing half of
// the same shape: roles swap, the activity row lands, and TransferOwnership reports the
// demoted owner — the value transfer_ownership.go hands to the OwnershipTransferred
// broadcast, so a wrong or nil result would silently mis-address the frame.
func TestTransactor_TransferOwnershipSuccessCommitsBothWrites(t *testing.T) {
	t.Cleanup(func() { testutil.TruncateAll(t, testPool) })

	oldOwner := fx.CreateTestUser(t)
	ws := fx.CreateTestWorkspace(t, oldOwner.ID)
	board := fx.CreateTestBoard(t, ws.ID, oldOwner.ID)
	target := fx.CreateTestUser(t)
	fx.AddBoardMember(t, board.ID, target.ID)

	tx := pgRepos.NewTransactor(testPool)
	memberRepo := pgRepos.NewBoardMemberRepository(testPool)
	activityRepo := pgRepos.NewActivityRepository(testPool)

	var fromUserID *uuid.UUID
	err := tx.WithinTransaction(context.Background(), func(ctx context.Context) error {
		var txErr error
		fromUserID, txErr = memberRepo.TransferOwnership(ctx, board.ID, target.ID)
		if txErr != nil {
			return txErr
		}
		return activityRepo.Log(ctx, &entity.Activity{
			BoardID:    board.ID,
			UserID:     &oldOwner.ID,
			ActionType: entity.ActivityActionOwnershipTransferred,
			EntityType: entity.ActivityEntityBoard,
			EntityID:   board.ID,
			Metadata: map[string]any{
				entity.ActivityMetaFromUserID: oldOwner.ID.String(),
				entity.ActivityMetaToUserID:   target.ID.String(),
			},
		})
	})
	require.NoError(t, err)

	require.NotNil(t, fromUserID, "the demoted owner must be reported, not nil")
	assert.Equal(t, oldOwner.ID, *fromUserID, "fromUserID must be the demoted owner")
	assert.NotEqual(t, target.ID, *fromUserID, "fromUserID must not be the new owner")

	oldOwnerBM, getErr := memberRepo.GetMemberByBoardAndUser(context.Background(), board.ID, oldOwner.ID)
	require.NoError(t, getErr)
	assert.Equal(t, entity.BoardRoleMember, oldOwnerBM.Role, "old owner must be demoted to BOARD_MEMBER")

	targetBM, getErr := memberRepo.GetMemberByBoardAndUser(context.Background(), board.ID, target.ID)
	require.NoError(t, getErr)
	assert.Equal(t, entity.BoardRoleOwner, targetBM.Role, "target must be promoted to BOARD_OWNER")

	var activityCount int
	require.NoError(t, testPool.QueryRow(context.Background(),
		`SELECT count(*) FROM activities
		 WHERE board_id = $1 AND action_type = $2`,
		board.ID, entity.ActivityActionOwnershipTransferred).Scan(&activityCount))
	assert.Equal(t, 1, activityCount, "exactly one OWNERSHIP_TRANSFERRED row must be committed")
}

// TestTransactor_SelfJoinBreakGlassLogFailureRollsBack covers the third transaction
// shape (self_join_board.go): boardMemberRepo.CreateIfAbsent followed — only when the
// insert actually happened — by a LogActivity carrying break_glass: true. When the
// activity write fails, the board_members row must not be left behind: a half-applied
// break-glass join would hand an admin silent access to a private board's contents.
func TestTransactor_SelfJoinBreakGlassLogFailureRollsBack(t *testing.T) {
	t.Cleanup(func() { testutil.TruncateAll(t, testPool) })

	owner := fx.CreateTestUser(t)
	ws := fx.CreateTestWorkspace(t, owner.ID)
	board := fx.CreateTestBoardWithVisibility(t, ws.ID, owner.ID, entity.BoardVisibilityPrivate)

	admin := fx.CreateTestUser(t)
	fx.AddWorkspaceMember(t, ws.ID, admin.ID, entity.WorkspaceRoleAdmin)

	tx := pgRepos.NewTransactor(testPool)
	memberRepo := pgRepos.NewBoardMemberRepository(testPool)
	activityRepo := pgRepos.NewActivityRepository(testPool)

	var inserted bool
	err := tx.WithinTransaction(context.Background(), func(ctx context.Context) error {
		var txErr error
		inserted, txErr = memberRepo.CreateIfAbsent(ctx, &entity.BoardMember{
			BoardID: board.ID,
			UserID:  admin.ID,
			Role:    entity.BoardRoleMember,
		})
		if txErr != nil {
			return txErr
		}
		require.True(t, inserted, "a first-time join must report inserted")

		// Non-existent board_id → FK violation in activities → Log fails.
		return activityRepo.Log(ctx, &entity.Activity{
			BoardID:    uuid.New(),
			UserID:     &admin.ID,
			ActionType: entity.ActivityActionJoined,
			EntityType: entity.ActivityEntityMember,
			EntityID:   admin.ID,
			Metadata: map[string]any{
				entity.ActivityMetaRole:       string(entity.BoardRoleMember),
				entity.ActivityMetaBreakGlass: true,
			},
		})
	})
	require.Error(t, err, "Log FK failure must propagate from WithinTransaction")

	_, getErr := memberRepo.GetMemberByBoardAndUser(context.Background(), board.ID, admin.ID)
	assert.ErrorIs(t, getErr, domain.ErrBoardMemberNotFound,
		"the board_members row must not survive a failed activity write")

	var activityCount int
	require.NoError(t, testPool.QueryRow(context.Background(),
		`SELECT count(*) FROM activities WHERE board_id = $1`, board.ID).Scan(&activityCount))
	assert.Equal(t, 0, activityCount, "no activity row may survive the rollback")
}

// TestTransactor_SelfJoinTwiceLogsOnce verifies the "only when inserted" arm of the
// same shape. CreateIfAbsent is an ON CONFLICT DO NOTHING RETURNING insert, so a repeat
// join reports inserted=false — which is what keeps self-join idempotent: no duplicate
// JOINED activity row, and the existing membership untouched.
func TestTransactor_SelfJoinTwiceLogsOnce(t *testing.T) {
	t.Cleanup(func() { testutil.TruncateAll(t, testPool) })

	owner := fx.CreateTestUser(t)
	ws := fx.CreateTestWorkspace(t, owner.ID)
	board := fx.CreateTestBoardWithVisibility(t, ws.ID, owner.ID, entity.BoardVisibilityPrivate)

	admin := fx.CreateTestUser(t)
	fx.AddWorkspaceMember(t, ws.ID, admin.ID, entity.WorkspaceRoleAdmin)

	tx := pgRepos.NewTransactor(testPool)
	memberRepo := pgRepos.NewBoardMemberRepository(testPool)
	activityRepo := pgRepos.NewActivityRepository(testPool)

	// join mirrors the self_join_board closure: log only when the insert happened.
	join := func() bool {
		var inserted bool
		err := tx.WithinTransaction(context.Background(), func(ctx context.Context) error {
			var txErr error
			inserted, txErr = memberRepo.CreateIfAbsent(ctx, &entity.BoardMember{
				BoardID: board.ID,
				UserID:  admin.ID,
				Role:    entity.BoardRoleMember,
			})
			if txErr != nil {
				return txErr
			}
			if inserted {
				return activityRepo.Log(ctx, &entity.Activity{
					BoardID:    board.ID,
					UserID:     &admin.ID,
					ActionType: entity.ActivityActionJoined,
					EntityType: entity.ActivityEntityMember,
					EntityID:   admin.ID,
					Metadata: map[string]any{
						entity.ActivityMetaRole:       string(entity.BoardRoleMember),
						entity.ActivityMetaBreakGlass: true,
					},
				})
			}
			return nil
		})
		require.NoError(t, err)
		return inserted
	}

	assert.True(t, join(), "first join must insert")
	assert.False(t, join(), "second join must report inserted=false, not an error")

	member, getErr := memberRepo.GetMemberByBoardAndUser(context.Background(), board.ID, admin.ID)
	require.NoError(t, getErr)
	assert.Equal(t, entity.BoardRoleMember, member.Role, "the existing membership must be untouched")

	var memberCount int
	require.NoError(t, testPool.QueryRow(context.Background(),
		`SELECT count(*) FROM board_members WHERE board_id = $1 AND user_id = $2`,
		board.ID, admin.ID).Scan(&memberCount))
	assert.Equal(t, 1, memberCount, "no duplicate board_members row")

	var activityCount int
	require.NoError(t, testPool.QueryRow(context.Background(),
		`SELECT count(*) FROM activities
		 WHERE board_id = $1 AND action_type = $2 AND entity_id = $3`,
		board.ID, entity.ActivityActionJoined, admin.ID).Scan(&activityCount))
	assert.Equal(t, 1, activityCount, "the repeat join must not write a second JOINED row")

	var breakGlass bool
	require.NoError(t, testPool.QueryRow(context.Background(),
		`SELECT (metadata->>'break_glass')::bool FROM activities
		 WHERE board_id = $1 AND action_type = $2`,
		board.ID, entity.ActivityActionJoined).Scan(&breakGlass))
	assert.True(t, breakGlass, "a workspace admin joining a PRIVATE board is a break-glass join")
}
