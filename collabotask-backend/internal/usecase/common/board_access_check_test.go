package common_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/mocks"
	"collabotask/internal/usecase/common"
)

type checkerDeps struct {
	boardRepo    *mocks.MockBoardRepository
	boardMbrRepo *mocks.MockBoardMemberRepository
	wsMbrRepo    *mocks.MockWorkspaceMemberRepository
	checker      common.BoardAccessChecker
}

func newCheckerDeps(t *testing.T) checkerDeps {
	t.Helper()
	d := checkerDeps{
		boardRepo:    mocks.NewMockBoardRepository(t),
		boardMbrRepo: mocks.NewMockBoardMemberRepository(t),
		wsMbrRepo:    mocks.NewMockWorkspaceMemberRepository(t),
	}
	d.checker = common.NewBoardAccessChecker(d.boardRepo, d.boardMbrRepo, d.wsMbrRepo)
	return d
}

// TestBoardAccessChecker exercises Resolve (Check delegates to it).
// Success cases assert the returned BoardAccess fields, which callers like
// GetBoardDetail depend on (e.g. nil BoardMember → accessStatus=CAN_JOIN).
func TestBoardAccessChecker(t *testing.T) {
	boardID := uuid.New()
	workspaceID := uuid.New()
	requesterID := uuid.New()
	creatorID := uuid.New() // always distinct from requesterID

	existingBoard := &entity.Board{
		ID:          boardID,
		WorkspaceID: workspaceID,
		CreatedBy:   creatorID,
	}
	adminMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleAdmin}
	regularMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleMember}
	boardMemberRow := &entity.BoardMember{BoardID: boardID, UserID: requesterID, Role: entity.BoardRoleMember}

	tests := []struct {
		name         string
		setupMocks   func(d checkerDeps)
		wantErr      error
		wantErrMsg   string
		assertAccess func(t *testing.T, a *common.BoardAccess)
	}{
		{
			name: "board not found (repo ErrBoardNotFound) → ErrBoardNotFound",
			setupMocks: func(d checkerDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(nil, domain.ErrBoardNotFound)
			},
			wantErr: domain.ErrBoardNotFound,
		},
		{
			name: "board nil → ErrBoardNotFound",
			setupMocks: func(d checkerDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(nil, nil)
			},
			wantErr: domain.ErrBoardNotFound,
		},
		{
			name: "board empty → ErrBoardNotFound",
			setupMocks: func(d checkerDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(&entity.Board{}, nil)
			},
			wantErr: domain.ErrBoardNotFound,
		},
		{
			name: "board archived → ErrBoardNotFound",
			setupMocks: func(d checkerDeps) {
				b := *existingBoard
				b.IsArchived = true
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(&b, nil)
			},
			wantErr: domain.ErrBoardNotFound,
		},
		{
			name: "requester not in workspace (ErrMemberNotFound) → ErrUserNotInWorkspace",
			setupMocks: func(d checkerDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(nil, domain.ErrMemberNotFound)
			},
			wantErr: domain.ErrUserNotInWorkspace,
		},
		{
			name: "workspace member nil (no error from repo) → ErrUserNotInWorkspace",
			setupMocks: func(d checkerDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(nil, nil)
			},
			wantErr: domain.ErrUserNotInWorkspace,
		},
		{
			name: "workspaceMemberRepo returns unexpected DB error → wrapped error",
			setupMocks: func(d checkerDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to fetch workspace membership",
		},
		{
			name: "unexpected error fetching board membership → wrapped error",
			setupMocks: func(d checkerDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to fetch board membership",
		},
		{
			name: "success — workspace admin (no board member entry needed)",
			setupMocks: func(d checkerDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
			},
			assertAccess: func(t *testing.T, a *common.BoardAccess) {
				t.Helper()
				require.Equal(t, boardID, a.Board.ID)
				require.NotNil(t, a.WorkspaceMember)
				require.True(t, a.WorkspaceMember.IsAdmin())
				// nil BoardMember signals accessStatus=CAN_JOIN to GetBoardDetail
				require.Nil(t, a.BoardMember)
			},
		},
		{
			name: "success — board creator (no board member entry)",
			setupMocks: func(d checkerDeps) {
				b := *existingBoard
				b.CreatedBy = requesterID
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(&b, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(regularMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
			},
			assertAccess: func(t *testing.T, a *common.BoardAccess) {
				t.Helper()
				require.Equal(t, requesterID, a.Board.CreatedBy)
				require.NotNil(t, a.WorkspaceMember)
				require.False(t, a.WorkspaceMember.IsAdmin())
				require.Nil(t, a.BoardMember)
			},
		},
		{
			name: "success — board member",
			setupMocks: func(d checkerDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(regularMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(boardMemberRow, nil)
			},
			assertAccess: func(t *testing.T, a *common.BoardAccess) {
				t.Helper()
				require.NotNil(t, a.Board)
				require.NotNil(t, a.WorkspaceMember)
				require.NotNil(t, a.BoardMember)
				require.Equal(t, requesterID, a.BoardMember.UserID)
				require.Equal(t, entity.BoardRoleMember, a.BoardMember.Role)
			},
		},
		{
			name: "regular workspace member, not creator, not board member → ErrBoardAccessDenied",
			setupMocks: func(d checkerDeps) {
				// existingBoard.CreatedBy = creatorID ≠ requesterID
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(regularMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
			},
			wantErr: domain.ErrBoardAccessDenied,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := newCheckerDeps(t)
			tt.setupMocks(d)

			access, err := d.checker.Resolve(context.Background(), boardID, requesterID)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			if tt.wantErrMsg != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.wantErrMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, access)
			if tt.assertAccess != nil {
				tt.assertAccess(t, access)
			}
		})
	}
}
