package board_test

// Broadcast contract tests for the board use case (Part C / ADR-009 / SRS §5.2).
// Three contract points per mutation: (1) happy path — Broadcast called with the
// right event type and board_id, (2) no-op silence — guards that prevent the write
// also prevent the broadcast, (3) best-effort — Broadcast has no return value.

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"collabotask/internal/domain/entity"
	"collabotask/internal/usecase/board"
	"collabotask/internal/usecase/common"
)

// --- UpdateBoard ---

func TestUpdateBoardBroadcast(t *testing.T) {
	workspaceID := uuid.New()
	boardID := uuid.New()
	requesterID := uuid.New()

	wsMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleAdmin}
	boardMember := &entity.BoardMember{BoardID: boardID, UserID: requesterID, Role: entity.BoardRoleOwner}

	setupBase := func(d boardTestDeps, b *entity.Board) {
		d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(b, nil)
		d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(wsMember, nil)
		d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(boardMember, nil)
		d.boardRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Maybe().Return(nil)
	}

	t.Run("happy path — BOARD_UPDATED broadcast with changed_fields=[title]", func(t *testing.T) {
		d := newDeps(t)
		b := &entity.Board{ID: boardID, WorkspaceID: workspaceID, Title: "Old Title"}
		setupBase(d, b)
		d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
			ev, ok := e.(common.BoardUpdated)
			return ok &&
				ev.Board.ID == boardID &&
				len(ev.ChangedFields) == 1 && ev.ChangedFields[0] == "title"
		}))

		_, err := d.uc.UpdateBoard(context.Background(), board.UpdateBoardInput{
			RequesterID: requesterID,
			BoardID:     boardID,
			Title:       strPtr("New Title"),
		})
		require.NoError(t, err)
	})

	t.Run("no-op silence — title unchanged → Broadcast NOT called", func(t *testing.T) {
		d := newDeps(t)
		b := &entity.Board{ID: boardID, WorkspaceID: workspaceID, Title: "Same Title"}
		setupBase(d, b)
		// broadcaster.EXPECT not set

		_, err := d.uc.UpdateBoard(context.Background(), board.UpdateBoardInput{
			RequesterID: requesterID,
			BoardID:     boardID,
			Title:       strPtr("Same Title"),
		})
		require.NoError(t, err)
	})
}

// --- SetArchived ---

func TestSetArchivedBroadcast(t *testing.T) {
	workspaceID := uuid.New()
	boardID := uuid.New()
	requesterID := uuid.New()

	wsMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleAdmin}
	boardMember := &entity.BoardMember{BoardID: boardID, UserID: requesterID, Role: entity.BoardRoleOwner}

	setupBase := func(d boardTestDeps, b *entity.Board) {
		d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(b, nil)
		d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(wsMember, nil)
		d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(boardMember, nil)
		d.boardRepo.EXPECT().SetArchived(mock.Anything, boardID, mock.AnythingOfType("bool")).Return(nil)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Maybe().Return(nil)
	}

	t.Run("happy path — BOARD_ARCHIVED broadcast when archiving", func(t *testing.T) {
		d := newDeps(t)
		b := &entity.Board{ID: boardID, WorkspaceID: workspaceID, IsArchived: false}
		setupBase(d, b)
		d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
			ev, ok := e.(common.BoardArchivedSet)
			return ok && ev.BoardID == boardID && ev.Archived == true
		}))

		_, err := d.uc.SetArchived(context.Background(), board.SetArchivedInput{
			RequesterID: requesterID,
			BoardID:     boardID,
			IsArchived:  boolPtr(true),
		})
		require.NoError(t, err)
	})

	t.Run("happy path — BOARD_UNARCHIVED broadcast when unarchiving", func(t *testing.T) {
		d := newDeps(t)
		b := &entity.Board{ID: boardID, WorkspaceID: workspaceID, IsArchived: true}
		setupBase(d, b)
		d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
			ev, ok := e.(common.BoardArchivedSet)
			return ok && ev.BoardID == boardID && ev.Archived == false
		}))

		_, err := d.uc.SetArchived(context.Background(), board.SetArchivedInput{
			RequesterID: requesterID,
			BoardID:     boardID,
			IsArchived:  boolPtr(false),
		})
		require.NoError(t, err)
	})

	t.Run("no-op silence — already archived → Broadcast NOT called", func(t *testing.T) {
		d := newDeps(t)
		b := &entity.Board{ID: boardID, WorkspaceID: workspaceID, IsArchived: true}
		d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(b, nil)
		d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(wsMember, nil)
		d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(boardMember, nil)
		// broadcaster.EXPECT not set

		_, err := d.uc.SetArchived(context.Background(), board.SetArchivedInput{
			RequesterID: requesterID,
			BoardID:     boardID,
			IsArchived:  boolPtr(true),
		})
		require.NoError(t, err)
	})
}

// --- TransferOwnership ---

func TestTransferOwnershipBroadcast(t *testing.T) {
	workspaceID := uuid.New()
	boardID := uuid.New()
	requesterID := uuid.New()
	toUserID := uuid.New()
	fromUserID := uuid.New()

	b := &entity.Board{ID: boardID, WorkspaceID: workspaceID}
	wsMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleAdmin}
	ownerBoardMember := &entity.BoardMember{BoardID: boardID, UserID: requesterID, Role: entity.BoardRoleOwner}
	targetMember := &entity.BoardMember{BoardID: boardID, UserID: toUserID, Role: entity.BoardRoleMember}
	alreadyOwnerTarget := &entity.BoardMember{BoardID: boardID, UserID: toUserID, Role: entity.BoardRoleOwner}

	input := board.TransferOwnershipInput{
		RequesterID: requesterID,
		WorkspaceID: workspaceID,
		BoardID:     boardID,
		ToUserID:    toUserID,
	}

	setupUntilTransfer := func(d boardTestDeps) {
		d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).
			Return(&common.BoardAccess{
				Board:           b,
				WorkspaceMember: wsMember,
				BoardMember:     ownerBoardMember,
			}, nil)
		d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, toUserID).Return(targetMember, nil)
		d.boardMbrRepo.EXPECT().TransferOwnership(mock.Anything, boardID, toUserID).Return(&fromUserID, nil)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Maybe().Return(nil)
	}

	t.Run("happy path — OWNERSHIP_TRANSFERRED broadcast with from/to user IDs", func(t *testing.T) {
		d := newDeps(t)
		setupUntilTransfer(d)
		d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
			ev, ok := e.(common.OwnershipTransferred)
			return ok &&
				ev.BoardID == boardID &&
				ev.ToUserID == toUserID &&
				ev.FromUserID != nil && *ev.FromUserID == fromUserID
		}))

		err := d.uc.TransferOwnership(context.Background(), input)
		require.NoError(t, err)
	})

	t.Run("no-op — target already owner → Broadcast NOT called", func(t *testing.T) {
		d := newDeps(t)
		d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).
			Return(&common.BoardAccess{
				Board:           b,
				WorkspaceMember: wsMember,
				BoardMember:     ownerBoardMember,
			}, nil)
		d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, toUserID).Return(alreadyOwnerTarget, nil)
		// broadcaster.EXPECT not set

		err := d.uc.TransferOwnership(context.Background(), input)
		require.NoError(t, err)
	})
}

// --- InviteMember ---

func TestInviteMemberBroadcast(t *testing.T) {
	workspaceID := uuid.New()
	boardID := uuid.New()
	requesterID := uuid.New()
	userAID := uuid.New()
	userBID := uuid.New()

	b := &entity.Board{ID: boardID, WorkspaceID: workspaceID}
	wsMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleAdmin}
	ownerBoardMember := &entity.BoardMember{BoardID: boardID, UserID: requesterID, Role: entity.BoardRoleOwner}

	setupBase := func(d boardTestDeps) {
		d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(b, nil)
		d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(wsMember, nil)
		d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(ownerBoardMember, nil)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Maybe().Return(nil)
	}

	t.Run("happy path — MEMBER_ADDED broadcast once per invitee", func(t *testing.T) {
		d := newDeps(t)
		setupBase(d)
		userA := &entity.User{ID: userAID, Name: "Alice"}
		d.userRepo.EXPECT().GetByIds(mock.Anything, []uuid.UUID{userAID}).
			Return(map[uuid.UUID]*entity.User{userAID: userA}, nil)
		d.wsMbrRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, userAID).Return(true, nil)
		d.boardMbrRepo.EXPECT().IsUserExists(mock.Anything, boardID, userAID).Return(false, nil)
		d.boardMbrRepo.EXPECT().CreateMany(mock.Anything, mock.Anything).Return(nil)

		d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
			ev, ok := e.(common.MemberAdded)
			return ok && ev.BoardID == boardID && ev.User.ID == userAID && ev.Role == entity.BoardRoleMember
		}))

		err := d.uc.InviteMember(context.Background(), board.InviteMemberInput{
			RequesterID: requesterID,
			WorkspaceID: workspaceID,
			BoardID:     boardID,
			UserIDs:     []uuid.UUID{userAID},
		})
		require.NoError(t, err)
	})

	t.Run("happy path — MEMBER_ADDED broadcast once per each of multiple invitees", func(t *testing.T) {
		d := newDeps(t)
		setupBase(d)
		userA := &entity.User{ID: userAID, Name: "Alice"}
		userB := &entity.User{ID: userBID, Name: "Bob"}
		d.userRepo.EXPECT().GetByIds(mock.Anything, []uuid.UUID{userAID, userBID}).
			Return(map[uuid.UUID]*entity.User{userAID: userA, userBID: userB}, nil)
		d.wsMbrRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, userAID).Return(true, nil)
		d.boardMbrRepo.EXPECT().IsUserExists(mock.Anything, boardID, userAID).Return(false, nil)
		d.wsMbrRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, userBID).Return(true, nil)
		d.boardMbrRepo.EXPECT().IsUserExists(mock.Anything, boardID, userBID).Return(false, nil)
		d.boardMbrRepo.EXPECT().CreateMany(mock.Anything, mock.Anything).Return(nil)

		d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
			ev, ok := e.(common.MemberAdded)
			return ok && ev.User.ID == userAID
		})).Once()
		d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
			ev, ok := e.(common.MemberAdded)
			return ok && ev.User.ID == userBID
		})).Once()

		err := d.uc.InviteMember(context.Background(), board.InviteMemberInput{
			RequesterID: requesterID,
			WorkspaceID: workspaceID,
			BoardID:     boardID,
			UserIDs:     []uuid.UUID{userAID, userBID},
		})
		require.NoError(t, err)
	})
}
