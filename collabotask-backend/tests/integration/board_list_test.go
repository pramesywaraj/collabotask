package integration_test

// Integration tests for getUserBoardsInWorkspace — the board-list query behind
// boardRepo.GetUserBoardsInWorkspace. Two things are proven here:
//
//   - the visibility filter (ADR-005): board_members is the sole source of access truth,
//     created_by is never consulted, and a plain member must not even learn that a PRIVATE
//     board exists;
//   - member_count / card_count: four stacked LEFT JOINs (bm, bm2, columns, cards) that
//     only COUNT(DISTINCT …) keeps from multiplying each other.

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"collabotask/internal/domain/entity"
	pgRepos "collabotask/internal/repository/postgres"
	"collabotask/tests/integration/testutil"
)

// boardIDs returns the ids of the listed boards, for presence/absence assertions.
func boardIDs(items []*entity.BoardListItem) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(items))
	for _, b := range items {
		ids = append(ids, b.ID)
	}
	return ids
}

// TestGetUserBoardsInWorkspace_AdminSeesUnjoinedPrivateBoard — a workspace admin outranks
// board authority, so a PRIVATE board they never joined is listed, flagged CAN_JOIN and
// carrying no board role: admins govern without joining, but must join to open contents.
func TestGetUserBoardsInWorkspace_AdminSeesUnjoinedPrivateBoard(t *testing.T) {
	t.Cleanup(func() { testutil.TruncateAll(t, testPool) })

	owner := fx.CreateTestUser(t)
	ws := fx.CreateTestWorkspace(t, owner.ID)
	fx.AddWorkspaceMember(t, ws.ID, owner.ID, entity.WorkspaceRoleAdmin)
	private := fx.CreateTestBoardWithVisibility(t, ws.ID, owner.ID, entity.BoardVisibilityPrivate)

	admin := fx.CreateTestUser(t)
	fx.AddWorkspaceMember(t, ws.ID, admin.ID, entity.WorkspaceRoleAdmin)

	boards, err := pgRepos.NewBoardRepository(testPool).
		GetUserBoardsInWorkspace(context.Background(), ws.ID, admin.ID)
	require.NoError(t, err)

	require.Equal(t, []uuid.UUID{private.ID}, boardIDs(boards), "the admin must see the private board")
	assert.Equal(t, entity.BoardCanJoin, boards[0].AccessStatus, "an unjoined board is CAN_JOIN")
	assert.Equal(t, entity.BoardRole(""), boards[0].UserRole, "a non-member has no board role")
}

// TestGetUserBoardsInWorkspace_PlainMemberCannotSeePrivateBoard — the same fixture read
// by a plain member. The board must be absent from the list entirely, not merely flagged
// unjoinable: a private board's existence is itself not disclosed (ADR-005).
func TestGetUserBoardsInWorkspace_PlainMemberCannotSeePrivateBoard(t *testing.T) {
	t.Cleanup(func() { testutil.TruncateAll(t, testPool) })

	owner := fx.CreateTestUser(t)
	ws := fx.CreateTestWorkspace(t, owner.ID)
	fx.AddWorkspaceMember(t, ws.ID, owner.ID, entity.WorkspaceRoleAdmin)
	private := fx.CreateTestBoardWithVisibility(t, ws.ID, owner.ID, entity.BoardVisibilityPrivate)

	member := fx.CreateTestUser(t)
	fx.AddWorkspaceMember(t, ws.ID, member.ID, entity.WorkspaceRoleMember)

	boardRepo := pgRepos.NewBoardRepository(testPool)

	boards, err := boardRepo.GetUserBoardsInWorkspace(context.Background(), ws.ID, member.ID)
	require.NoError(t, err)
	assert.Empty(t, boardIDs(boards), "a plain member must not see an unjoined private board")

	// Joining flips it: board_members, not visibility, is what grants access.
	fx.AddBoardMember(t, private.ID, member.ID)

	boards, err = boardRepo.GetUserBoardsInWorkspace(context.Background(), ws.ID, member.ID)
	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{private.ID}, boardIDs(boards), "a joined private board is listed")
	assert.Equal(t, entity.BoardJoined, boards[0].AccessStatus)
	assert.Equal(t, entity.BoardRoleMember, boards[0].UserRole, "the real board role is reported")
}

// TestGetUserBoardsInWorkspace_PlainMemberSeesWorkspaceBoard — a WORKSPACE-visible board
// is listed for any workspace member as CAN_JOIN, and reports their real role once joined.
func TestGetUserBoardsInWorkspace_PlainMemberSeesWorkspaceBoard(t *testing.T) {
	t.Cleanup(func() { testutil.TruncateAll(t, testPool) })

	owner := fx.CreateTestUser(t)
	ws := fx.CreateTestWorkspace(t, owner.ID)
	fx.AddWorkspaceMember(t, ws.ID, owner.ID, entity.WorkspaceRoleAdmin)
	board := fx.CreateTestBoard(t, ws.ID, owner.ID) // WORKSPACE-visible

	member := fx.CreateTestUser(t)
	fx.AddWorkspaceMember(t, ws.ID, member.ID, entity.WorkspaceRoleMember)

	boardRepo := pgRepos.NewBoardRepository(testPool)

	boards, err := boardRepo.GetUserBoardsInWorkspace(context.Background(), ws.ID, member.ID)
	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{board.ID}, boardIDs(boards))
	assert.Equal(t, entity.BoardCanJoin, boards[0].AccessStatus, "visible but not joined → CAN_JOIN")
	assert.Equal(t, entity.BoardRole(""), boards[0].UserRole)

	fx.AddBoardMember(t, board.ID, member.ID)

	boards, err = boardRepo.GetUserBoardsInWorkspace(context.Background(), ws.ID, member.ID)
	require.NoError(t, err)
	require.Len(t, boards, 1, "joining must not duplicate the board row")
	assert.Equal(t, entity.BoardJoined, boards[0].AccessStatus)
	assert.Equal(t, entity.BoardRoleMember, boards[0].UserRole)
}

// TestGetUserBoardsInWorkspace_ExcludesArchivedBoards — archived boards drop out of the
// list (is_archived = FALSE), even for the owner who is a joined BOARD_OWNER.
func TestGetUserBoardsInWorkspace_ExcludesArchivedBoards(t *testing.T) {
	t.Cleanup(func() { testutil.TruncateAll(t, testPool) })

	owner := fx.CreateTestUser(t)
	ws := fx.CreateTestWorkspace(t, owner.ID)
	fx.AddWorkspaceMember(t, ws.ID, owner.ID, entity.WorkspaceRoleAdmin)
	kept := fx.CreateTestBoard(t, ws.ID, owner.ID)
	archived := fx.CreateTestBoard(t, ws.ID, owner.ID)

	boardRepo := pgRepos.NewBoardRepository(testPool)
	require.NoError(t, boardRepo.SetArchived(context.Background(), archived.ID, true))

	boards, err := boardRepo.GetUserBoardsInWorkspace(context.Background(), ws.ID, owner.ID)
	require.NoError(t, err)
	assert.Equal(t, []uuid.UUID{kept.ID}, boardIDs(boards), "an archived board must not be listed")
}

// TestGetUserBoardsInWorkspace_NonWorkspaceMemberSeesNothing — the INNER JOIN on
// workspace_members is the outer gate: without a workspace membership nothing is listed,
// not even a WORKSPACE-visible board, and not even a board the user created.
func TestGetUserBoardsInWorkspace_NonWorkspaceMemberSeesNothing(t *testing.T) {
	t.Cleanup(func() { testutil.TruncateAll(t, testPool) })

	owner := fx.CreateTestUser(t)
	ws := fx.CreateTestWorkspace(t, owner.ID)
	fx.AddWorkspaceMember(t, ws.ID, owner.ID, entity.WorkspaceRoleAdmin)
	fx.CreateTestBoard(t, ws.ID, owner.ID)

	// Created the board below, and is even a board member — but never joined the workspace.
	outsider := fx.CreateTestUser(t)
	outsiderBoard := fx.CreateTestBoard(t, ws.ID, outsider.ID)

	boards, err := pgRepos.NewBoardRepository(testPool).
		GetUserBoardsInWorkspace(context.Background(), ws.ID, outsider.ID)
	require.NoError(t, err)
	assert.Empty(t, boardIDs(boards),
		"a non-workspace-member sees nothing, including the board they created")

	// Sanity: that board does exist and does list for a workspace admin, so the empty
	// result above is the workspace gate, not a missing fixture.
	adminView, err := pgRepos.NewBoardRepository(testPool).
		GetUserBoardsInWorkspace(context.Background(), ws.ID, owner.ID)
	require.NoError(t, err)
	assert.Contains(t, boardIDs(adminView), outsiderBoard.ID)
}

// TestGetUserBoardsInWorkspace_CountsDistinctMembersAndCards guards the COUNT(DISTINCT …)
// pair. The query joins bm2 (members) and columns→cards (cards) in the same row set, so
// the two fan out against each other: a board with 3 members and 2 cards produces 3 × 2
// rows, and plain COUNT(…) would report 6 for both counts. Both counts therefore have to
// exceed one in the same fixture — a board with a single member or a single card passes
// either way and proves nothing.
//
// Three boards are listed at once, which also pins the counts as per-board (GROUP BY b.id)
// rather than workspace-wide.
func TestGetUserBoardsInWorkspace_CountsDistinctMembersAndCards(t *testing.T) {
	t.Cleanup(func() { testutil.TruncateAll(t, testPool) })

	owner := fx.CreateTestUser(t)
	ws := fx.CreateTestWorkspace(t, owner.ID)
	fx.AddWorkspaceMember(t, ws.ID, owner.ID, entity.WorkspaceRoleAdmin)

	// Board A: 3 members (owner + 2) and 2 cards in one column → the fan-out case.
	boardA := fx.CreateTestBoard(t, ws.ID, owner.ID)
	for i := 0; i < 2; i++ {
		extra := fx.CreateTestUser(t)
		fx.AddBoardMember(t, boardA.ID, extra.ID)
	}
	colA := fx.CreateTestColumn(t, boardA.ID)
	fx.CreateTestCard(t, colA.ID, boardA.ID, owner.ID)
	fx.CreateTestCardAt(t, colA.ID, boardA.ID, owner.ID, "second card", 2000)

	// Board B: columns (the three defaults) but no cards at all.
	boardB := fx.CreateTestBoard(t, ws.ID, owner.ID)

	// Board C: 3 cards spread across two columns — each counted once, summed per board.
	boardC := fx.CreateTestBoard(t, ws.ID, owner.ID)
	colC1 := fx.CreateTestColumn(t, boardC.ID)
	colC2 := fx.CreateTestColumn(t, boardC.ID)
	fx.CreateTestCard(t, colC1.ID, boardC.ID, owner.ID)
	fx.CreateTestCardAt(t, colC1.ID, boardC.ID, owner.ID, "second card", 2000)
	fx.CreateTestCard(t, colC2.ID, boardC.ID, owner.ID)

	boards, err := pgRepos.NewBoardRepository(testPool).
		GetUserBoardsInWorkspace(context.Background(), ws.ID, owner.ID)
	require.NoError(t, err)
	require.Len(t, boards, 3)

	byID := make(map[uuid.UUID]*entity.BoardListItem, len(boards))
	for _, b := range boards {
		byID[b.ID] = b
	}

	require.Contains(t, byID, boardA.ID)
	assert.Equal(t, uint(3), byID[boardA.ID].MemberCount, "3 members, not 3 × the card count")
	assert.Equal(t, uint(2), byID[boardA.ID].CardCount, "2 cards, not 2 × the member count")

	require.Contains(t, byID, boardB.ID)
	assert.Equal(t, uint(1), byID[boardB.ID].MemberCount)
	assert.Equal(t, uint(0), byID[boardB.ID].CardCount, "columns without cards count zero, not one")

	require.Contains(t, byID, boardC.ID)
	assert.Equal(t, uint(1), byID[boardC.ID].MemberCount)
	assert.Equal(t, uint(3), byID[boardC.ID].CardCount, "cards across several columns are summed once each")
}
