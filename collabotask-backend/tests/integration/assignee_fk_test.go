package integration_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	pgRepos "collabotask/internal/repository/postgres"
	"collabotask/tests/integration/testutil"
)

// TestAssigneeFKInvariant_OnWrite verifies that the composite FK
// fk_cards_assignee_board_member enforces the invariant at the DB level.
//   - assigned_to = non-member → FK rejects with 23503 on our named constraint
//   - assigned_to = member     → succeeds
//   - assigned_to = NULL       → succeeds (unconstrained)
func TestAssigneeFKInvariant_OnWrite(t *testing.T) {
	t.Cleanup(func() { testutil.TruncateAll(t, testPool) })

	owner := fx.CreateTestUser(t)
	ws := fx.CreateTestWorkspace(t, owner.ID)
	board := fx.CreateTestBoard(t, ws.ID, owner.ID)
	col := fx.CreateTestColumn(t, board.ID)

	nonMember := fx.CreateTestUser(t)
	member := fx.CreateTestUser(t)
	fx.AddBoardMember(t, board.ID, member.ID)

	cardRepo := pgRepos.NewCardRepository(testPool)

	t.Run("assigned_to = non-member → ErrAssigneeNotBoardMember (FK 23503)", func(t *testing.T) {
		card := &entity.Card{
			ColumnID:   col.ID,
			BoardID:    board.ID,
			Title:      "Assigned to non-member",
			Position:   1000,
			AssignedTo: &nonMember.ID,
			CreatedBy:  owner.ID,
		}
		err := cardRepo.Create(context.Background(), card)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrAssigneeNotBoardMember)
	})

	t.Run("assigned_to = member → succeeds", func(t *testing.T) {
		card := &entity.Card{
			ColumnID:   col.ID,
			BoardID:    board.ID,
			Title:      "Assigned to member",
			Position:   2000,
			AssignedTo: &member.ID,
			CreatedBy:  owner.ID,
		}
		err := cardRepo.Create(context.Background(), card)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, card.ID)
		assert.Equal(t, &member.ID, card.AssignedTo)
	})

	t.Run("assigned_to = NULL → succeeds", func(t *testing.T) {
		card := &entity.Card{
			ColumnID:  col.ID,
			BoardID:   board.ID,
			Title:     "Unassigned card",
			Position:  3000,
			CreatedBy: owner.ID,
		}
		err := cardRepo.Create(context.Background(), card)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, card.ID)
		assert.Nil(t, card.AssignedTo)
	})
}

// TestAssigneeFKInvariant_DBAutoCascade verifies that the FK's ON DELETE SET NULL
// fires when a board_member row is deleted directly (bypassing the app cascade).
func TestAssigneeFKInvariant_DBAutoCascade(t *testing.T) {
	t.Cleanup(func() { testutil.TruncateAll(t, testPool) })

	owner := fx.CreateTestUser(t)
	ws := fx.CreateTestWorkspace(t, owner.ID)
	board := fx.CreateTestBoard(t, ws.ID, owner.ID)
	col := fx.CreateTestColumn(t, board.ID)

	assignee := fx.CreateTestUser(t)
	fx.AddBoardMember(t, board.ID, assignee.ID)

	card := fx.CreateTestCard(t, col.ID, board.ID, owner.ID)

	// assign the member via direct UPDATE (bypassing app guard, which is fine here)
	_, err := testPool.Exec(context.Background(),
		`UPDATE cards SET assigned_to = $1 WHERE id = $2`,
		assignee.ID, card.ID,
	)
	require.NoError(t, err)

	// delete the board_member row directly — the FK should silently null assigned_to
	_, err = testPool.Exec(context.Background(),
		`DELETE FROM board_members WHERE board_id = $1 AND user_id = $2`,
		board.ID, assignee.ID,
	)
	require.NoError(t, err)

	// verify cards.assigned_to is now NULL
	var assignedTo *uuid.UUID
	err = testPool.QueryRow(context.Background(),
		`SELECT assigned_to FROM cards WHERE id = $1`, card.ID,
	).Scan(&assignedTo)
	require.NoError(t, err)
	assert.Nil(t, assignedTo, "FK ON DELETE SET NULL should have cleared assigned_to")
}

// TestAssigneeFKInvariant_BoardCascadeBroadcastList verifies that after the cascade
// reorder, RemoveWithParticipationCascade (board path) still returns the correct
// non-empty []AffectedCard. This is the regression guard for the reorder.
func TestAssigneeFKInvariant_BoardCascadeBroadcastList(t *testing.T) {
	t.Cleanup(func() { testutil.TruncateAll(t, testPool) })

	owner := fx.CreateTestUser(t)
	ws := fx.CreateTestWorkspace(t, owner.ID)
	board := fx.CreateTestBoard(t, ws.ID, owner.ID)
	col := fx.CreateTestColumn(t, board.ID)

	assignee := fx.CreateTestUser(t)
	fx.AddBoardMember(t, board.ID, assignee.ID)

	// Create two cards assigned to assignee so we can assert both come back.
	card1 := fx.CreateTestCard(t, col.ID, board.ID, owner.ID)
	card2 := fx.CreateTestCard(t, col.ID, board.ID, owner.ID)

	_, err := testPool.Exec(context.Background(),
		`UPDATE cards SET assigned_to = $1 WHERE id = ANY($2)`,
		assignee.ID, []uuid.UUID{card1.ID, card2.ID},
	)
	require.NoError(t, err)

	boardMemberRepo := pgRepos.NewBoardMemberRepository(testPool)
	affected, err := boardMemberRepo.RemoveWithParticipationCascade(
		context.Background(), board.ID, assignee.ID,
	)
	require.NoError(t, err)
	require.Len(t, affected, 2, "both assigned cards should be in the broadcast list")

	affectedIDs := make([]uuid.UUID, len(affected))
	for i, a := range affected {
		assert.Equal(t, board.ID, a.BoardID)
		affectedIDs[i] = a.CardID
	}
	assert.ElementsMatch(t, []uuid.UUID{card1.ID, card2.ID}, affectedIDs)

	// DB should reflect NULL assigned_to
	var count int
	err = testPool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM cards WHERE id = ANY($1) AND assigned_to IS NULL`,
		[]uuid.UUID{card1.ID, card2.ID},
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

// TestAssigneeFKInvariant_WorkspaceCascadeBroadcastList verifies that after the
// cascade reorder, RemoveWithParticipationCascade (workspace path) still returns
// the correct non-empty []AffectedCard.
func TestAssigneeFKInvariant_WorkspaceCascadeBroadcastList(t *testing.T) {
	t.Cleanup(func() { testutil.TruncateAll(t, testPool) })

	owner := fx.CreateTestUser(t)
	ws := fx.CreateTestWorkspace(t, owner.ID)
	board := fx.CreateTestBoard(t, ws.ID, owner.ID)
	col := fx.CreateTestColumn(t, board.ID)

	// Add assignee to workspace and board
	assignee := fx.CreateTestUser(t)
	wsMemberRepo := pgRepos.NewWorkspaceMemberRepository(testPool)
	err := wsMemberRepo.Create(context.Background(), &entity.WorkspaceMember{
		WorkspaceID: ws.ID,
		UserID:      assignee.ID,
		Role:        entity.WorkspaceRoleMember,
	})
	require.NoError(t, err)
	fx.AddBoardMember(t, board.ID, assignee.ID)

	card1 := fx.CreateTestCard(t, col.ID, board.ID, owner.ID)
	card2 := fx.CreateTestCard(t, col.ID, board.ID, owner.ID)

	_, err = testPool.Exec(context.Background(),
		`UPDATE cards SET assigned_to = $1 WHERE id = ANY($2)`,
		assignee.ID, []uuid.UUID{card1.ID, card2.ID},
	)
	require.NoError(t, err)

	result, err := wsMemberRepo.RemoveWithParticipationCascade(
		context.Background(), ws.ID, assignee.ID,
	)
	require.NoError(t, err)
	require.Len(t, result.AffectedCards, 2, "both assigned cards should be in the broadcast list")
	assert.Len(t, result.AffectedBoardIDs, 1)
	assert.Equal(t, board.ID, result.AffectedBoardIDs[0])

	affectedIDs := make([]uuid.UUID, len(result.AffectedCards))
	for i, a := range result.AffectedCards {
		assert.Equal(t, board.ID, a.BoardID)
		affectedIDs[i] = a.CardID
	}
	assert.ElementsMatch(t, []uuid.UUID{card1.ID, card2.ID}, affectedIDs)
}

// TestAssigneeFKInvariant_ConstraintName verifies that the DB FK constraint is
// named exactly fk_cards_assignee_board_member, so our error-mapping code works.
func TestAssigneeFKInvariant_ConstraintName(t *testing.T) {
	t.Cleanup(func() { testutil.TruncateAll(t, testPool) })

	owner := fx.CreateTestUser(t)
	ws := fx.CreateTestWorkspace(t, owner.ID)
	board := fx.CreateTestBoard(t, ws.ID, owner.ID)
	col := fx.CreateTestColumn(t, board.ID)

	nonMember := fx.CreateTestUser(t)

	_, err := testPool.Exec(context.Background(),
		`INSERT INTO cards (column_id, board_id, title, position, assigned_to, created_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		col.ID, board.ID, "Test", 1000.0, nonMember.ID, owner.ID,
	)
	require.Error(t, err)

	var pgErr *pgconn.PgError
	require.ErrorAs(t, err, &pgErr)
	assert.Equal(t, "23503", pgErr.Code)
	assert.Equal(t, "fk_cards_assignee_board_member", pgErr.ConstraintName)
}
