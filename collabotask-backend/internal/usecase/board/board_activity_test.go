package board_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"collabotask/internal/domain/entity"
	"collabotask/internal/domain/repository"
	"collabotask/internal/usecase/board"
	"collabotask/internal/usecase/common"
)

// --- UpdateBoard ---

func TestUpdateBoardActivityLog(t *testing.T) {
	workspaceID := uuid.New()
	boardID := uuid.New()
	requesterID := uuid.New()

	newBoard := func() *entity.Board { return &entity.Board{ID: boardID, WorkspaceID: workspaceID, Title: "Old Title"} }
	wsMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleAdmin}
	boardMember := &entity.BoardMember{BoardID: boardID, UserID: requesterID, Role: entity.BoardRoleOwner}

	setupBase := func(d boardTestDeps) {
		d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(newBoard(), nil)
		d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(wsMember, nil)
		d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(boardMember, nil)
		d.boardRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
	}

	t.Run("happy path — BOARD/UPDATED logged with changed title in changed_fields", func(t *testing.T) {
		d := newDeps(t)
		setupBase(d)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.MatchedBy(func(a *entity.Activity) bool {
			changedFields, _ := a.Metadata[entity.ActivityMetaChangedFields].([]string)
			boardTitle, _ := a.Metadata[entity.ActivityMetaBoardTitle].(string)
			return a.BoardID == boardID &&
				a.ActionType == entity.ActivityActionUpdated &&
				a.EntityType == entity.ActivityEntityBoard &&
				a.EntityID == boardID &&
				a.UserID != nil && *a.UserID == requesterID &&
				boardTitle == "New Title" &&
				len(changedFields) == 1 && changedFields[0] == "title"
		})).Return(nil)

		_, err := d.uc.UpdateBoard(context.Background(), board.UpdateBoardInput{
			RequesterID: requesterID,
			BoardID:     boardID,
			Title:       strPtr("New Title"),
		})
		require.NoError(t, err)
	})

	t.Run("no-op silence — description present but unchanged → Log NOT called", func(t *testing.T) {
		d := newDeps(t)
		desc := "existing description"
		sameBoard := &entity.Board{ID: boardID, WorkspaceID: workspaceID, Title: "Old Title", Description: &desc}
		d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(sameBoard, nil)
		d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(wsMember, nil)
		d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(boardMember, nil)
		d.boardRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
		// DescriptionPresent=true but value identical → no changedFields → Log must NOT be called

		_, err := d.uc.UpdateBoard(context.Background(), board.UpdateBoardInput{
			RequesterID:         requesterID,
			BoardID:             boardID,
			DescriptionPresent:  true,
			Description:         strPtr("existing description"),
		})
		require.NoError(t, err)
	})

	t.Run("no-op silence — title unchanged → Log NOT called", func(t *testing.T) {
		d := newDeps(t)
		sameBoard := &entity.Board{ID: boardID, WorkspaceID: workspaceID, Title: "Same Title"}
		d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(sameBoard, nil)
		d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(wsMember, nil)
		d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(boardMember, nil)
		d.boardRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
		// No activityRepo.Log expectation

		_, err := d.uc.UpdateBoard(context.Background(), board.UpdateBoardInput{
			RequesterID: requesterID,
			BoardID:     boardID,
			Title:       strPtr("Same Title"),
		})
		require.NoError(t, err)
	})

	t.Run("resilience — Log error is swallowed, update still succeeds", func(t *testing.T) {
		d := newDeps(t)
		setupBase(d)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Return(errors.New("db down"))

		_, err := d.uc.UpdateBoard(context.Background(), board.UpdateBoardInput{
			RequesterID: requesterID,
			BoardID:     boardID,
			Title:       strPtr("New Title"),
		})
		require.NoError(t, err)
	})

	t.Run("happy path — visibility flip logged with visibility_from and visibility_to", func(t *testing.T) {
		d := newDeps(t)
		visBoard := &entity.Board{ID: boardID, WorkspaceID: workspaceID, Title: "My Board", Visibility: entity.BoardVisibilityWorkspace}
		d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(visBoard, nil)
		d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(wsMember, nil)
		d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(boardMember, nil)
		d.boardRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)

		d.activityRepo.EXPECT().Log(mock.Anything, mock.MatchedBy(func(a *entity.Activity) bool {
			changedFields, _ := a.Metadata[entity.ActivityMetaChangedFields].([]string)
			visFrom, _ := a.Metadata[entity.ActivityMetaVisibilityFrom].(string)
			visTo, _ := a.Metadata[entity.ActivityMetaVisibilityTo].(string)
			return a.BoardID == boardID &&
				a.ActionType == entity.ActivityActionUpdated &&
				a.EntityType == entity.ActivityEntityBoard &&
				a.UserID != nil && *a.UserID == requesterID &&
				len(changedFields) == 1 && changedFields[0] == "visibility" &&
				visFrom == string(entity.BoardVisibilityWorkspace) &&
				visTo == string(entity.BoardVisibilityPrivate)
		})).Return(nil)

		_, err := d.uc.UpdateBoard(context.Background(), board.UpdateBoardInput{
			RequesterID: requesterID,
			BoardID:     boardID,
			Visibility:  strPtr(string(entity.BoardVisibilityPrivate)),
		})
		require.NoError(t, err)
	})
}

// --- SetArchived ---

func TestSetArchivedActivityLog(t *testing.T) {
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
	}

	t.Run("happy path — BOARD/ARCHIVED logged when archiving", func(t *testing.T) {
		d := newDeps(t)
		b := &entity.Board{ID: boardID, WorkspaceID: workspaceID, IsArchived: false}
		setupBase(d, b)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.MatchedBy(func(a *entity.Activity) bool {
			return a.BoardID == boardID &&
				a.ActionType == entity.ActivityActionArchived &&
				a.EntityType == entity.ActivityEntityBoard &&
				a.EntityID == boardID &&
				a.UserID != nil && *a.UserID == requesterID
		})).Return(nil)

		_, err := d.uc.SetArchived(context.Background(), board.SetArchivedInput{
			RequesterID: requesterID,
			BoardID:     boardID,
			IsArchived:  boolPtr(true),
		})
		require.NoError(t, err)
	})

	t.Run("happy path — BOARD/UNARCHIVED logged when unarchiving", func(t *testing.T) {
		d := newDeps(t)
		b := &entity.Board{ID: boardID, WorkspaceID: workspaceID, IsArchived: true}
		setupBase(d, b)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.MatchedBy(func(a *entity.Activity) bool {
			return a.BoardID == boardID &&
				a.ActionType == entity.ActivityActionUnarchived &&
				a.EntityType == entity.ActivityEntityBoard
		})).Return(nil)

		_, err := d.uc.SetArchived(context.Background(), board.SetArchivedInput{
			RequesterID: requesterID,
			BoardID:     boardID,
			IsArchived:  boolPtr(false),
		})
		require.NoError(t, err)
	})

	t.Run("no-op silence — already archived → Log NOT called", func(t *testing.T) {
		d := newDeps(t)
		b := &entity.Board{ID: boardID, WorkspaceID: workspaceID, IsArchived: true}
		d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(b, nil)
		d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(wsMember, nil)
		d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(boardMember, nil)
		// SetArchived repo may or may not be called; Log must NOT be called

		_, err := d.uc.SetArchived(context.Background(), board.SetArchivedInput{
			RequesterID: requesterID,
			BoardID:     boardID,
			IsArchived:  boolPtr(true),
		})
		require.NoError(t, err)
	})

	t.Run("resilience — Log error is swallowed, archive still succeeds", func(t *testing.T) {
		d := newDeps(t)
		b := &entity.Board{ID: boardID, WorkspaceID: workspaceID, IsArchived: false}
		setupBase(d, b)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Return(errors.New("db down"))

		_, err := d.uc.SetArchived(context.Background(), board.SetArchivedInput{
			RequesterID: requesterID,
			BoardID:     boardID,
			IsArchived:  boolPtr(true),
		})
		require.NoError(t, err)
	})
}

// --- TransferOwnership ---

func TestTransferOwnershipActivityLog(t *testing.T) {
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
	}

	t.Run("happy path — BOARD/OWNERSHIP_TRANSFERRED logged with from/to user IDs", func(t *testing.T) {
		d := newDeps(t)
		setupUntilTransfer(d)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.MatchedBy(func(a *entity.Activity) bool {
			fromID, _ := a.Metadata[entity.ActivityMetaFromUserID].(string)
			toID, _ := a.Metadata[entity.ActivityMetaToUserID].(string)
			return a.BoardID == boardID &&
				a.ActionType == entity.ActivityActionOwnershipTransferred &&
				a.EntityType == entity.ActivityEntityBoard &&
				a.EntityID == boardID &&
				a.UserID != nil && *a.UserID == requesterID &&
				fromID == fromUserID.String() &&
				toID == toUserID.String()
		})).Return(nil)

		err := d.uc.TransferOwnership(context.Background(), input)
		require.NoError(t, err)
	})

	t.Run("no-op — target already owner → Log NOT called", func(t *testing.T) {
		d := newDeps(t)
		d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).
			Return(&common.BoardAccess{
				Board:           b,
				WorkspaceMember: wsMember,
				BoardMember:     ownerBoardMember,
			}, nil)
		d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, toUserID).Return(alreadyOwnerTarget, nil)
		// early return → Log must NOT be called

		err := d.uc.TransferOwnership(context.Background(), input)
		require.NoError(t, err)
	})

	t.Run("resilience — Log error is swallowed, transfer still succeeds", func(t *testing.T) {
		d := newDeps(t)
		setupUntilTransfer(d)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Return(errors.New("db down"))

		err := d.uc.TransferOwnership(context.Background(), input)
		require.NoError(t, err)
	})
}

// --- SelfJoinBoard ---

func TestSelfJoinBoardActivityLog(t *testing.T) {
	workspaceID := uuid.New()
	boardID := uuid.New()
	requesterID := uuid.New()

	workspaceBoard := &entity.Board{ID: boardID, WorkspaceID: workspaceID, Visibility: entity.BoardVisibilityWorkspace}
	privateBoard := &entity.Board{ID: boardID, WorkspaceID: workspaceID, Visibility: entity.BoardVisibilityPrivate}
	adminMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleAdmin}
	regularMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleMember}

	input := board.SelfJoinBoardInput{
		RequesterID: workspaceID,
		BoardID:     boardID,
		WorkspaceID: workspaceID,
	}
	input.RequesterID = requesterID

	t.Run("happy path — MEMBER/JOINED logged when inserted, break_glass=false (workspace board)", func(t *testing.T) {
		d := newDeps(t)
		d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(workspaceBoard, nil)
		d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(regularMember, nil)
		d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, nil)
		newMember := &entity.BoardMember{BoardID: boardID, UserID: requesterID, Role: entity.BoardRoleMember}
		d.boardMbrRepo.EXPECT().CreateIfAbsent(mock.Anything, newMember).Return(true, nil)

		d.activityRepo.EXPECT().Log(mock.Anything, mock.MatchedBy(func(a *entity.Activity) bool {
			breakGlass, _ := a.Metadata[entity.ActivityMetaBreakGlass].(bool)
			role, _ := a.Metadata[entity.ActivityMetaRole].(string)
			return a.BoardID == boardID &&
				a.ActionType == entity.ActivityActionJoined &&
				a.EntityType == entity.ActivityEntityMember &&
				a.EntityID == requesterID &&
				a.UserID != nil && *a.UserID == requesterID &&
				role == string(entity.BoardRoleMember) &&
				!breakGlass
		})).Return(nil)

		out, err := d.uc.SelfJoinBoard(context.Background(), input)
		require.NoError(t, err)
		require.True(t, out.Joined)
	})

	t.Run("happy path — MEMBER/JOINED logged with break_glass=true when admin joins PRIVATE board", func(t *testing.T) {
		d := newDeps(t)
		d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(privateBoard, nil)
		d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
		d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, nil)
		newMember := &entity.BoardMember{BoardID: boardID, UserID: requesterID, Role: entity.BoardRoleMember}
		d.boardMbrRepo.EXPECT().CreateIfAbsent(mock.Anything, newMember).Return(true, nil)

		d.activityRepo.EXPECT().Log(mock.Anything, mock.MatchedBy(func(a *entity.Activity) bool {
			breakGlass, _ := a.Metadata[entity.ActivityMetaBreakGlass].(bool)
			role, _ := a.Metadata[entity.ActivityMetaRole].(string)
			return a.BoardID == boardID &&
				a.ActionType == entity.ActivityActionJoined &&
				a.EntityType == entity.ActivityEntityMember &&
				role == string(entity.BoardRoleMember) &&
				breakGlass
		})).Return(nil)

		out, err := d.uc.SelfJoinBoard(context.Background(), input)
		require.NoError(t, err)
		require.True(t, out.Joined)
	})

	t.Run("no-op silence — already a member (idempotent re-join) → Log NOT called", func(t *testing.T) {
		d := newDeps(t)
		d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(workspaceBoard, nil)
		d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(regularMember, nil)
		existingMember := &entity.BoardMember{BoardID: boardID, UserID: requesterID, Role: entity.BoardRoleMember}
		d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(existingMember, nil)
		// Early return — Log must NOT be called

		out, err := d.uc.SelfJoinBoard(context.Background(), input)
		require.NoError(t, err)
		require.False(t, out.Joined)
	})

	t.Run("no-op silence — race: CreateIfAbsent returns inserted=false → Log NOT called", func(t *testing.T) {
		d := newDeps(t)
		d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(workspaceBoard, nil)
		d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(regularMember, nil)
		d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, nil)
		newMember := &entity.BoardMember{BoardID: boardID, UserID: requesterID, Role: entity.BoardRoleMember}
		d.boardMbrRepo.EXPECT().CreateIfAbsent(mock.Anything, newMember).Return(false, nil)
		// Log must NOT be called

		out, err := d.uc.SelfJoinBoard(context.Background(), input)
		require.NoError(t, err)
		require.False(t, out.Joined)
	})

	t.Run("resilience — Log error is swallowed, join still succeeds", func(t *testing.T) {
		d := newDeps(t)
		d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(workspaceBoard, nil)
		d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(regularMember, nil)
		d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, nil)
		newMember := &entity.BoardMember{BoardID: boardID, UserID: requesterID, Role: entity.BoardRoleMember}
		d.boardMbrRepo.EXPECT().CreateIfAbsent(mock.Anything, newMember).Return(true, nil)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Return(errors.New("db down"))

		out, err := d.uc.SelfJoinBoard(context.Background(), input)
		require.NoError(t, err)
		require.True(t, out.Joined)
	})
}

// --- InviteMember (board) ---

func TestBoardInviteMemberActivityLog(t *testing.T) {
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
	}

	t.Run("happy path — MEMBER/ADDED logged once per invitee", func(t *testing.T) {
		d := newDeps(t)
		setupBase(d)
		d.userRepo.EXPECT().GetByIds(mock.Anything, []uuid.UUID{userAID}).
			Return(map[uuid.UUID]*entity.User{userAID: {ID: userAID}}, nil)
		d.wsMbrRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, userAID).Return(true, nil)
		d.boardMbrRepo.EXPECT().IsUserExists(mock.Anything, boardID, userAID).Return(false, nil)
		d.boardMbrRepo.EXPECT().CreateMany(mock.Anything, mock.Anything).Return(nil)

		d.activityRepo.EXPECT().Log(mock.Anything, mock.MatchedBy(func(a *entity.Activity) bool {
			role, _ := a.Metadata[entity.ActivityMetaRole].(string)
			return a.BoardID == boardID &&
				a.ActionType == entity.ActivityActionAdded &&
				a.EntityType == entity.ActivityEntityMember &&
				a.EntityID == userAID &&
				a.UserID != nil && *a.UserID == requesterID &&
				role == string(entity.BoardRoleMember)
		})).Return(nil).Once()

		err := d.uc.InviteMember(context.Background(), board.InviteMemberInput{
			RequesterID: requesterID,
			WorkspaceID: workspaceID,
			BoardID:     boardID,
			UserIDs:     []uuid.UUID{userAID},
		})
		require.NoError(t, err)
	})

	t.Run("happy path — MEMBER/ADDED logged once per each of multiple invitees", func(t *testing.T) {
		d := newDeps(t)
		setupBase(d)
		d.userRepo.EXPECT().GetByIds(mock.Anything, []uuid.UUID{userAID, userBID}).
			Return(map[uuid.UUID]*entity.User{
				userAID: {ID: userAID},
				userBID: {ID: userBID},
			}, nil)
		d.wsMbrRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, userAID).Return(true, nil)
		d.boardMbrRepo.EXPECT().IsUserExists(mock.Anything, boardID, userAID).Return(false, nil)
		d.wsMbrRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, userBID).Return(true, nil)
		d.boardMbrRepo.EXPECT().IsUserExists(mock.Anything, boardID, userBID).Return(false, nil)
		d.boardMbrRepo.EXPECT().CreateMany(mock.Anything, mock.Anything).Return(nil)

		d.activityRepo.EXPECT().Log(mock.Anything, mock.MatchedBy(func(a *entity.Activity) bool {
			return a.ActionType == entity.ActivityActionAdded && a.EntityType == entity.ActivityEntityMember &&
				a.EntityID == userAID
		})).Return(nil).Once()
		d.activityRepo.EXPECT().Log(mock.Anything, mock.MatchedBy(func(a *entity.Activity) bool {
			return a.ActionType == entity.ActivityActionAdded && a.EntityType == entity.ActivityEntityMember &&
				a.EntityID == userBID
		})).Return(nil).Once()

		err := d.uc.InviteMember(context.Background(), board.InviteMemberInput{
			RequesterID: requesterID,
			WorkspaceID: workspaceID,
			BoardID:     boardID,
			UserIDs:     []uuid.UUID{userAID, userBID},
		})
		require.NoError(t, err)
	})

	t.Run("resilience — Log error per invitee is swallowed, invite still succeeds", func(t *testing.T) {
		d := newDeps(t)
		setupBase(d)
		d.userRepo.EXPECT().GetByIds(mock.Anything, []uuid.UUID{userAID}).
			Return(map[uuid.UUID]*entity.User{userAID: {ID: userAID}}, nil)
		d.wsMbrRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, userAID).Return(true, nil)
		d.boardMbrRepo.EXPECT().IsUserExists(mock.Anything, boardID, userAID).Return(false, nil)
		d.boardMbrRepo.EXPECT().CreateMany(mock.Anything, mock.Anything).Return(nil)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Return(errors.New("db down"))

		err := d.uc.InviteMember(context.Background(), board.InviteMemberInput{
			RequesterID: requesterID,
			WorkspaceID: workspaceID,
			BoardID:     boardID,
			UserIDs:     []uuid.UUID{userAID},
		})
		require.NoError(t, err)
	})
}

// --- LeaveBoard ---

func TestLeaveBoardActivityLog(t *testing.T) {
	boardID := uuid.New()
	requesterID := uuid.New()

	b := &entity.Board{ID: boardID}
	regularMember := &entity.BoardMember{BoardID: boardID, UserID: requesterID, Role: entity.BoardRoleMember}

	setupBase := func(d boardTestDeps) {
		d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(b, nil)
		d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(regularMember, nil)
		d.boardMbrRepo.EXPECT().RemoveWithParticipationCascade(mock.Anything, boardID, requesterID).Return([]repository.AffectedCard{}, nil)
	}

	input := board.LeaveBoardInput{
		RequesterID: requesterID,
		BoardID:     boardID,
	}

	t.Run("happy path — MEMBER/LEFT logged with source=board", func(t *testing.T) {
		d := newDeps(t)
		setupBase(d)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.MatchedBy(func(a *entity.Activity) bool {
			source, _ := a.Metadata[entity.ActivityMetaSource].(string)
			return a.BoardID == boardID &&
				a.ActionType == entity.ActivityActionLeft &&
				a.EntityType == entity.ActivityEntityMember &&
				a.EntityID == requesterID &&
				a.UserID != nil && *a.UserID == requesterID &&
				source == "board"
		})).Return(nil)

		err := d.uc.LeaveBoard(context.Background(), input)
		require.NoError(t, err)
	})

	t.Run("resilience — Log error is swallowed, leave still succeeds", func(t *testing.T) {
		d := newDeps(t)
		setupBase(d)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Return(errors.New("db down"))

		err := d.uc.LeaveBoard(context.Background(), input)
		require.NoError(t, err)
	})
}

// --- RemoveMember (board) ---

func TestBoardRemoveMemberActivityLog(t *testing.T) {
	workspaceID := uuid.New()
	boardID := uuid.New()
	requesterID := uuid.New()
	targetID := uuid.New()

	b := &entity.Board{ID: boardID, WorkspaceID: workspaceID}
	wsMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleAdmin}
	ownerBoardMember := &entity.BoardMember{BoardID: boardID, UserID: requesterID, Role: entity.BoardRoleOwner}

	setupBase := func(d boardTestDeps) {
		d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(b, nil)
		d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(wsMember, nil)
		d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(ownerBoardMember, nil)
		d.boardMbrRepo.EXPECT().RemoveWithParticipationCascade(mock.Anything, boardID, targetID).Return([]repository.AffectedCard{}, nil)
	}

	input := board.RemoveMemberInput{
		RequesterID: requesterID,
		WorkspaceID: workspaceID,
		BoardID:     boardID,
		UserID:      targetID,
	}

	t.Run("happy path — MEMBER/REMOVED logged with source=board", func(t *testing.T) {
		d := newDeps(t)
		setupBase(d)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.MatchedBy(func(a *entity.Activity) bool {
			source, _ := a.Metadata[entity.ActivityMetaSource].(string)
			return a.BoardID == boardID &&
				a.ActionType == entity.ActivityActionRemoved &&
				a.EntityType == entity.ActivityEntityMember &&
				a.EntityID == targetID &&
				a.UserID != nil && *a.UserID == requesterID &&
				source == "board"
		})).Return(nil)

		err := d.uc.RemoveMember(context.Background(), input)
		require.NoError(t, err)
	})

	t.Run("resilience — Log error is swallowed, remove still succeeds", func(t *testing.T) {
		d := newDeps(t)
		setupBase(d)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Return(errors.New("db down"))

		err := d.uc.RemoveMember(context.Background(), input)
		require.NoError(t, err)
	})
}
