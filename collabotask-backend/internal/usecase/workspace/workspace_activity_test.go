package workspace_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"collabotask/internal/domain/entity"
	"collabotask/internal/domain/repository"
	"collabotask/internal/mocks"
	"collabotask/internal/usecase/workspace"
)

// --- LeaveWorkspace ---
// Workspace cascade writes one MEMBER/LEFT activity per affected board.
// source="workspace" distinguishes it from board-scoped left events.

func TestLeaveWorkspaceActivityLog(t *testing.T) {
	requesterID := uuid.New()
	workspaceID := uuid.New()
	boardAID := uuid.New()
	boardBID := uuid.New()

	regularMember := &entity.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      requesterID,
		Role:        entity.WorkspaceRoleMember,
	}
	ws := &entity.Workspace{ID: workspaceID, OwnerID: uuid.New()}

	validInput := workspace.LeaveWorkspaceInput{
		RequesterID: requesterID,
		WorkspaceID: workspaceID,
	}

	t.Run("happy path — one MEMBER/LEFT per affected board", func(t *testing.T) {
		wsRepo := mocks.NewMockWorkspaceRepository(t)
		wsMemberRepo := mocks.NewMockWorkspaceMemberRepository(t)
		userRepo := mocks.NewMockUserRepository(t)
		activityRepo := mocks.NewMockActivityRepository(t)

		wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(regularMember, nil)
		wsRepo.EXPECT().GetByID(mock.Anything, workspaceID).Return(ws, nil)
		wsMemberRepo.EXPECT().RemoveWithParticipationCascade(mock.Anything, workspaceID, requesterID).Return(
			repository.WorkspaceCascadeResult{
				AffectedBoardIDs: []uuid.UUID{boardAID, boardBID},
			}, nil,
		)

		activityRepo.EXPECT().Log(mock.Anything, mock.MatchedBy(func(a *entity.Activity) bool {
			source, _ := a.Metadata[entity.ActivityMetaSource].(string)
			return a.BoardID == boardAID &&
				a.ActionType == entity.ActivityActionLeft &&
				a.EntityType == entity.ActivityEntityMember &&
				a.EntityID == requesterID &&
				a.UserID != nil && *a.UserID == requesterID &&
				source == "workspace"
		})).Return(nil).Once()
		activityRepo.EXPECT().Log(mock.Anything, mock.MatchedBy(func(a *entity.Activity) bool {
			source, _ := a.Metadata[entity.ActivityMetaSource].(string)
			return a.BoardID == boardBID &&
				a.ActionType == entity.ActivityActionLeft &&
				a.EntityType == entity.ActivityEntityMember &&
				source == "workspace"
		})).Return(nil).Once()

		uc := workspace.NewWorkspaceUseCase(wsRepo, wsMemberRepo, userRepo, activityRepo)
		err := uc.LeaveWorkspace(context.Background(), validInput)
		require.NoError(t, err)
	})

	t.Run("no affected boards → Log NOT called", func(t *testing.T) {
		wsRepo := mocks.NewMockWorkspaceRepository(t)
		wsMemberRepo := mocks.NewMockWorkspaceMemberRepository(t)
		userRepo := mocks.NewMockUserRepository(t)
		activityRepo := mocks.NewMockActivityRepository(t)

		wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(regularMember, nil)
		wsRepo.EXPECT().GetByID(mock.Anything, workspaceID).Return(ws, nil)
		wsMemberRepo.EXPECT().RemoveWithParticipationCascade(mock.Anything, workspaceID, requesterID).Return(
			repository.WorkspaceCascadeResult{AffectedBoardIDs: []uuid.UUID{}}, nil,
		)
		// No Log expectation

		uc := workspace.NewWorkspaceUseCase(wsRepo, wsMemberRepo, userRepo, activityRepo)
		err := uc.LeaveWorkspace(context.Background(), validInput)
		require.NoError(t, err)
	})

	t.Run("resilience — Log error per board is swallowed, leave still succeeds", func(t *testing.T) {
		wsRepo := mocks.NewMockWorkspaceRepository(t)
		wsMemberRepo := mocks.NewMockWorkspaceMemberRepository(t)
		userRepo := mocks.NewMockUserRepository(t)
		activityRepo := mocks.NewMockActivityRepository(t)

		wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(regularMember, nil)
		wsRepo.EXPECT().GetByID(mock.Anything, workspaceID).Return(ws, nil)
		wsMemberRepo.EXPECT().RemoveWithParticipationCascade(mock.Anything, workspaceID, requesterID).Return(
			repository.WorkspaceCascadeResult{
				AffectedBoardIDs: []uuid.UUID{boardAID},
			}, nil,
		)
		activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Return(errors.New("db down"))

		uc := workspace.NewWorkspaceUseCase(wsRepo, wsMemberRepo, userRepo, activityRepo)
		err := uc.LeaveWorkspace(context.Background(), validInput)
		require.NoError(t, err)
	})
}

// --- RemoveMember (workspace) ---
// Workspace cascade writes one MEMBER/REMOVED activity per affected board.
// source="workspace" distinguishes it from board-scoped removed events.

func TestWorkspaceRemoveMemberActivityLog(t *testing.T) {
	requesterID := uuid.New()
	targetID := uuid.New()
	workspaceID := uuid.New()
	boardAID := uuid.New()

	adminMember := &entity.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      requesterID,
		Role:        entity.WorkspaceRoleAdmin,
	}

	validInput := workspace.RemoveMemberInput{
		RequesterID: requesterID,
		WorkspaceID: workspaceID,
		UserID:      targetID,
	}

	t.Run("happy path — one MEMBER/REMOVED per affected board", func(t *testing.T) {
		wsRepo := mocks.NewMockWorkspaceRepository(t)
		wsMemberRepo := mocks.NewMockWorkspaceMemberRepository(t)
		userRepo := mocks.NewMockUserRepository(t)
		activityRepo := mocks.NewMockActivityRepository(t)

		wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
		wsMemberRepo.EXPECT().RemoveWithParticipationCascade(mock.Anything, workspaceID, targetID).Return(
			repository.WorkspaceCascadeResult{
				AffectedBoardIDs: []uuid.UUID{boardAID},
			}, nil,
		)

		activityRepo.EXPECT().Log(mock.Anything, mock.MatchedBy(func(a *entity.Activity) bool {
			source, _ := a.Metadata[entity.ActivityMetaSource].(string)
			return a.BoardID == boardAID &&
				a.ActionType == entity.ActivityActionRemoved &&
				a.EntityType == entity.ActivityEntityMember &&
				a.EntityID == targetID &&
				a.UserID != nil && *a.UserID == requesterID &&
				source == "workspace"
		})).Return(nil).Once()

		uc := workspace.NewWorkspaceUseCase(wsRepo, wsMemberRepo, userRepo, activityRepo)
		err := uc.RemoveMember(context.Background(), validInput)
		require.NoError(t, err)
	})

	t.Run("no affected boards → Log NOT called", func(t *testing.T) {
		wsRepo := mocks.NewMockWorkspaceRepository(t)
		wsMemberRepo := mocks.NewMockWorkspaceMemberRepository(t)
		userRepo := mocks.NewMockUserRepository(t)
		activityRepo := mocks.NewMockActivityRepository(t)

		wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
		wsMemberRepo.EXPECT().RemoveWithParticipationCascade(mock.Anything, workspaceID, targetID).Return(
			repository.WorkspaceCascadeResult{AffectedBoardIDs: []uuid.UUID{}}, nil,
		)
		// No Log expectation

		uc := workspace.NewWorkspaceUseCase(wsRepo, wsMemberRepo, userRepo, activityRepo)
		err := uc.RemoveMember(context.Background(), validInput)
		require.NoError(t, err)
	})

	t.Run("resilience — Log error is swallowed, remove still succeeds", func(t *testing.T) {
		wsRepo := mocks.NewMockWorkspaceRepository(t)
		wsMemberRepo := mocks.NewMockWorkspaceMemberRepository(t)
		userRepo := mocks.NewMockUserRepository(t)
		activityRepo := mocks.NewMockActivityRepository(t)

		wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
		wsMemberRepo.EXPECT().RemoveWithParticipationCascade(mock.Anything, workspaceID, targetID).Return(
			repository.WorkspaceCascadeResult{
				AffectedBoardIDs: []uuid.UUID{boardAID},
			}, nil,
		)
		activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Return(errors.New("db down"))

		uc := workspace.NewWorkspaceUseCase(wsRepo, wsMemberRepo, userRepo, activityRepo)
		err := uc.RemoveMember(context.Background(), validInput)
		require.NoError(t, err)
	})
}
