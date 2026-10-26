package integration_test

// Integration tests for the board_members removal paths. The sentinel these assert on is
// load-bearing: leave_board.go and remove_member.go both branch on
// domain.ErrBoardMemberNotFound to turn a no-op removal into a 404 instead of reporting a
// success (and broadcasting a departure) that never happened.

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	pgRepos "collabotask/internal/repository/postgres"
	"collabotask/tests/integration/testutil"
)

// TestBoardMemberRepository_RemoveNonMemberReturnsSentinel — a delete that affects zero
// rows must surface as ErrBoardMemberNotFound, not as a silent success. Both removal paths
// are covered: Delete and the participation cascade.
func TestBoardMemberRepository_RemoveNonMemberReturnsSentinel(t *testing.T) {
	t.Cleanup(func() { testutil.TruncateAll(t, testPool) })

	owner := fx.CreateTestUser(t)
	ws := fx.CreateTestWorkspace(t, owner.ID)
	board := fx.CreateTestBoard(t, ws.ID, owner.ID)

	stranger := fx.CreateTestUser(t) // exists, but was never added to the board
	memberRepo := pgRepos.NewBoardMemberRepository(testPool)

	err := memberRepo.Delete(context.Background(), board.ID, stranger.ID)
	assert.ErrorIs(t, err, domain.ErrBoardMemberNotFound, "Delete of a non-member must not report success")

	affected, err := memberRepo.RemoveWithParticipationCascade(context.Background(), board.ID, stranger.ID)
	assert.ErrorIs(t, err, domain.ErrBoardMemberNotFound, "the cascade must not report success either")
	assert.Empty(t, affected, "no cards can be affected by a removal that did not happen")

	// Positive control: the same calls on a real member succeed, so the sentinel above is
	// the zero-row branch and not an unconditional error.
	member := fx.CreateTestUser(t)
	fx.AddBoardMember(t, board.ID, member.ID)
	require.NoError(t, memberRepo.Delete(context.Background(), board.ID, member.ID))

	fx.AddBoardMember(t, board.ID, member.ID)
	_, err = memberRepo.RemoveWithParticipationCascade(context.Background(), board.ID, member.ID)
	require.NoError(t, err)
}

// TestBoardMemberRepository_TransferToNonMemberReturnsSentinel — the same sentinel guards
// the promote half of a transfer: ownership may only move to an existing board member, so
// a promote that matches no row must fail rather than leave the board without an owner.
func TestBoardMemberRepository_TransferToNonMemberReturnsSentinel(t *testing.T) {
	t.Cleanup(func() { testutil.TruncateAll(t, testPool) })

	owner := fx.CreateTestUser(t)
	ws := fx.CreateTestWorkspace(t, owner.ID)
	board := fx.CreateTestBoard(t, ws.ID, owner.ID)
	stranger := fx.CreateTestUser(t)

	memberRepo := pgRepos.NewBoardMemberRepository(testPool)

	_, err := memberRepo.TransferOwnership(context.Background(), board.ID, stranger.ID)
	assert.ErrorIs(t, err, domain.ErrBoardMemberNotFound)

	// The failed promote must not have left the demote applied.
	ownerBM, getErr := memberRepo.GetMemberByBoardAndUser(context.Background(), board.ID, owner.ID)
	require.NoError(t, getErr)
	assert.Equal(t, entity.BoardRoleOwner, ownerBM.Role, "the board must still have its owner")
}
