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
	"collabotask/internal/usecase/common"
	"collabotask/internal/usecase/workspace"
)

// Broadcast-contract tests for the workspace participation cascade (Part E /
// ADR-009 / SRS §4.5 UC-19b / §5.2). The fan-out evicts the user from EVERY
// workspace board (superset — closes the workspace-visible-watcher leak), then
// broadcasts MEMBER_REMOVED per affected board and CARD_UPDATED per cleared card
// routed to that card's own board.

type wsBroadcastDeps struct {
	wsRepo       *mocks.MockWorkspaceRepository
	wsMemberRepo *mocks.MockWorkspaceMemberRepository
	boardRepo    *mocks.MockBoardRepository
	broadcaster  *mocks.MockBroadcaster
	uc           *workspace.WorkspaceUseCase
}

func newWSBroadcastDeps(t *testing.T) wsBroadcastDeps {
	wsRepo := mocks.NewMockWorkspaceRepository(t)
	wsMemberRepo := mocks.NewMockWorkspaceMemberRepository(t)
	boardRepo := mocks.NewMockBoardRepository(t)
	userRepo := mocks.NewMockUserRepository(t)
	activityRepo := mocks.NewMockActivityRepository(t)
	broadcaster := mocks.NewMockBroadcaster(t)
	activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Maybe().Return(nil)

	return wsBroadcastDeps{
		wsRepo:       wsRepo,
		wsMemberRepo: wsMemberRepo,
		boardRepo:    boardRepo,
		broadcaster:  broadcaster,
		uc:           workspace.NewWorkspaceUseCase(wsRepo, wsMemberRepo, boardRepo, userRepo, activityRepo, broadcaster),
	}
}

// --- RemoveMember (UC-06, involuntary → ACCESS_REVOKED) ---

func TestRemoveMemberBroadcast(t *testing.T) {
	requesterID := uuid.New()
	targetID := uuid.New()
	workspaceID := uuid.New()
	boardA, boardB, boardC := uuid.New(), uuid.New(), uuid.New()
	cardID, colID := uuid.New(), uuid.New()

	adminMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleAdmin}
	input := workspace.RemoveMemberInput{RequesterID: requesterID, WorkspaceID: workspaceID, UserID: targetID}

	t.Run("happy path — evict all workspace boards, MEMBER_REMOVED + CARD_UPDATED per affected", func(t *testing.T) {
		d := newWSBroadcastDeps(t)
		d.wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
		d.wsMemberRepo.EXPECT().RemoveWithParticipationCascade(mock.Anything, workspaceID, targetID).Return(
			repository.WorkspaceCascadeResult{
				AffectedBoardIDs: []uuid.UUID{boardA, boardB},
				AffectedCards:    []repository.AffectedCard{{CardID: cardID, ColumnID: colID, BoardID: boardA}},
			}, nil,
		)

		// boardC is a WORKSPACE-visible board the target watched without joining —
		// present in the eviction superset but not in AffectedBoardIDs.
		d.boardRepo.EXPECT().GetBoardIDsByWorkspace(mock.Anything, workspaceID).
			Return([]uuid.UUID{boardA, boardB, boardC}, nil)

		// 1. Evict from ALL workspace boards, non-silent reason (→ ACCESS_REVOKED per board).
		d.broadcaster.EXPECT().EvictUserFromRooms(targetID, []uuid.UUID{boardA, boardB, boardC}, common.EvictReasonRemovedFromWorkspace)
		// 2. MEMBER_REMOVED per affected board.
		d.broadcaster.EXPECT().Broadcast(boardA, mock.MatchedBy(func(e common.Event) bool {
			ev, ok := e.(common.MemberRemoved)
			return ok && ev.BoardID == boardA && ev.UserID == targetID
		}))
		d.broadcaster.EXPECT().Broadcast(boardB, mock.MatchedBy(func(e common.Event) bool {
			ev, ok := e.(common.MemberRemoved)
			return ok && ev.BoardID == boardB && ev.UserID == targetID
		}))
		// 3. CARD_UPDATED routed to the cleared card's own board (boardA).
		d.broadcaster.EXPECT().Broadcast(boardA, mock.MatchedBy(func(e common.Event) bool {
			ev, ok := e.(common.CardUpdated)
			return ok && ev.Card.ID == cardID &&
				len(ev.ChangedFields) == 1 && ev.ChangedFields[0] == "assigned_to" &&
				ev.Assignee == nil
		}))

		require.NoError(t, d.uc.RemoveMember(context.Background(), input))
	})

	t.Run("board-list error → eviction falls back to AffectedBoardIDs, broadcasts still emitted", func(t *testing.T) {
		d := newWSBroadcastDeps(t)
		d.wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
		d.wsMemberRepo.EXPECT().RemoveWithParticipationCascade(mock.Anything, workspaceID, targetID).Return(
			repository.WorkspaceCascadeResult{AffectedBoardIDs: []uuid.UUID{boardA}}, nil,
		)
		d.boardRepo.EXPECT().GetBoardIDsByWorkspace(mock.Anything, workspaceID).
			Return(nil, errors.New("db down"))

		// Fallback: evict over AffectedBoardIDs (still tears down every member-board).
		d.broadcaster.EXPECT().EvictUserFromRooms(targetID, []uuid.UUID{boardA}, common.EvictReasonRemovedFromWorkspace)
		d.broadcaster.EXPECT().Broadcast(boardA, mock.MatchedBy(func(e common.Event) bool {
			_, ok := e.(common.MemberRemoved)
			return ok
		}))

		require.NoError(t, d.uc.RemoveMember(context.Background(), input))
	})
}

func TestRemoveMemberBroadcast_CascadeError(t *testing.T) {
	requesterID := uuid.New()
	targetID := uuid.New()
	workspaceID := uuid.New()

	adminMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleAdmin}
	input := workspace.RemoveMemberInput{RequesterID: requesterID, WorkspaceID: workspaceID, UserID: targetID}

	d := newWSBroadcastDeps(t)
	d.wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
	d.wsMemberRepo.EXPECT().RemoveWithParticipationCascade(mock.Anything, workspaceID, targetID).Return(
		repository.WorkspaceCascadeResult{}, errors.New("db error"),
	)
	// No GetBoardIDsByWorkspace, no eviction, no broadcast when the cascade errors.

	require.Error(t, d.uc.RemoveMember(context.Background(), input))
}

// --- LeaveWorkspace (UC-06c, voluntary → silent) ---

func TestLeaveWorkspaceBroadcast(t *testing.T) {
	requesterID := uuid.New()
	workspaceID := uuid.New()
	boardA, boardB := uuid.New(), uuid.New()
	cardID, colID := uuid.New(), uuid.New()

	regularMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleMember}
	ws := &entity.Workspace{ID: workspaceID, OwnerID: uuid.New()}
	input := workspace.LeaveWorkspaceInput{RequesterID: requesterID, WorkspaceID: workspaceID}

	t.Run("happy path — silent eviction, MEMBER_REMOVED + CARD_UPDATED per affected", func(t *testing.T) {
		d := newWSBroadcastDeps(t)
		d.wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(regularMember, nil)
		d.wsRepo.EXPECT().GetByID(mock.Anything, workspaceID).Return(ws, nil)
		d.wsMemberRepo.EXPECT().RemoveWithParticipationCascade(mock.Anything, workspaceID, requesterID).Return(
			repository.WorkspaceCascadeResult{
				AffectedBoardIDs: []uuid.UUID{boardA},
				AffectedCards:    []repository.AffectedCard{{CardID: cardID, ColumnID: colID, BoardID: boardA}},
			}, nil,
		)
		d.boardRepo.EXPECT().GetBoardIDsByWorkspace(mock.Anything, workspaceID).
			Return([]uuid.UUID{boardA, boardB}, nil)

		// Voluntary leave → silent (no ACCESS_REVOKED).
		d.broadcaster.EXPECT().EvictUserFromRooms(requesterID, []uuid.UUID{boardA, boardB}, common.EvictReasonSilent)
		d.broadcaster.EXPECT().Broadcast(boardA, mock.MatchedBy(func(e common.Event) bool {
			ev, ok := e.(common.MemberRemoved)
			return ok && ev.BoardID == boardA && ev.UserID == requesterID
		}))
		d.broadcaster.EXPECT().Broadcast(boardA, mock.MatchedBy(func(e common.Event) bool {
			ev, ok := e.(common.CardUpdated)
			return ok && ev.Card.ID == cardID && len(ev.ChangedFields) == 1 && ev.ChangedFields[0] == "assigned_to"
		}))

		require.NoError(t, d.uc.LeaveWorkspace(context.Background(), input))
	})
}

func TestLeaveWorkspaceBroadcast_CascadeError(t *testing.T) {
	requesterID := uuid.New()
	workspaceID := uuid.New()

	regularMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleMember}
	ws := &entity.Workspace{ID: workspaceID, OwnerID: uuid.New()}
	input := workspace.LeaveWorkspaceInput{RequesterID: requesterID, WorkspaceID: workspaceID}

	d := newWSBroadcastDeps(t)
	d.wsMemberRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(regularMember, nil)
	d.wsRepo.EXPECT().GetByID(mock.Anything, workspaceID).Return(ws, nil)
	d.wsMemberRepo.EXPECT().RemoveWithParticipationCascade(mock.Anything, workspaceID, requesterID).Return(
		repository.WorkspaceCascadeResult{}, errors.New("db error"),
	)
	// No eviction, no broadcast when the cascade errors.

	require.Error(t, d.uc.LeaveWorkspace(context.Background(), input))
}
