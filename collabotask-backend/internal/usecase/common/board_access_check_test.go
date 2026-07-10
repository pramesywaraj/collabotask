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

const (
	boardIntent = iota
	metadataIntent
	viewIntent
	mutateIntent
)

func invoke(ctx context.Context, c common.BoardAccessChecker, intent int, boardID, requesterID uuid.UUID) (*common.BoardAccess, error) {
	switch intent {
	case metadataIntent:
		return c.CheckMetadataAccess(ctx, boardID, requesterID)
	case viewIntent:
		return c.CheckViewAccess(ctx, boardID, requesterID)
	default:
		return c.CheckMutateAccess(ctx, boardID, requesterID)
	}
}

// TestBoardAccessChecker_Matrix drives the ADR-005 authorization matrix:
// each actor/visibility row asserts the expected outcome for all three intents.
func TestBoardAccessChecker_Matrix(t *testing.T) {
	boardID := uuid.New()
	workspaceID := uuid.New()
	requesterID := uuid.New()
	creatorID := uuid.New()

	newBoard := func(v entity.BoardVisibility) *entity.Board {
		return &entity.Board{ID: boardID, WorkspaceID: workspaceID, CreatedBy: creatorID, Visibility: v}
	}
	adminMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleAdmin}
	regularMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleMember}
	boardMemberRow := &entity.BoardMember{BoardID: boardID, UserID: requesterID, Role: entity.BoardRoleMember}

	// setupActor wires the resolve() facts (board, workspace membership, board
	// membership) for one row. Called fresh for each intent.
	type actor struct {
		board  *entity.Board
		wsMbr  *entity.WorkspaceMember
		joined bool
	}

	setup := func(d checkerDeps, a actor) {
		d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(a.board, nil)
		d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(a.wsMbr, nil)
		if a.joined {
			d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(boardMemberRow, nil)
		} else {
			d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
		}
	}

	rows := []struct {
		name     string
		actor    actor
		metadata error // nil = granted
		view     error
		mutate   error
	}{
		{
			name:     "board member (joined) — all granted",
			actor:    actor{board: newBoard(entity.BoardVisibilityPrivate), wsMbr: regularMember, joined: true},
			metadata: nil, view: nil, mutate: nil,
		},
		{
			name:     "admin, not joined, WORKSPACE — all granted",
			actor:    actor{board: newBoard(entity.BoardVisibilityWorkspace), wsMbr: adminMember, joined: false},
			metadata: nil, view: nil, mutate: nil,
		},
		{
			name:     "admin, not joined, PRIVATE — metadata ok; view/mutate break-glass",
			actor:    actor{board: newBoard(entity.BoardVisibilityPrivate), wsMbr: adminMember, joined: false},
			metadata: nil, view: domain.ErrBoardJoinRequired, mutate: domain.ErrBoardJoinRequired,
		},
		{
			name:     "member, not joined, WORKSPACE — metadata/view ok; mutate 403",
			actor:    actor{board: newBoard(entity.BoardVisibilityWorkspace), wsMbr: regularMember, joined: false},
			metadata: nil, view: nil, mutate: domain.ErrBoardAccessDenied,
		},
		{
			name:     "member, not joined, PRIVATE — all hidden (404)",
			actor:    actor{board: newBoard(entity.BoardVisibilityPrivate), wsMbr: regularMember, joined: false},
			metadata: domain.ErrBoardNotFound, view: domain.ErrBoardNotFound, mutate: domain.ErrBoardNotFound,
		},
	}

	intents := []struct {
		name string
		code int
	}{
		{"metadata", metadataIntent},
		{"view", viewIntent},
		{"mutate", mutateIntent},
	}

	for _, row := range rows {
		row := row
		for _, in := range intents {
			in := in
			t.Run(row.name+"/"+in.name, func(t *testing.T) {
				t.Parallel()
				d := newCheckerDeps(t)
				setup(d, row.actor)

				access, err := invoke(context.Background(), d.checker, in.code, boardID, requesterID)

				var wantErr error
				switch in.code {
				case metadataIntent:
					wantErr = row.metadata
				case viewIntent:
					wantErr = row.view
				case mutateIntent:
					wantErr = row.mutate
				}

				if wantErr != nil {
					require.ErrorIs(t, err, wantErr)
					require.Nil(t, access)
					return
				}
				require.NoError(t, err)
				require.NotNil(t, access)
				require.Equal(t, boardID, access.Board.ID)
			})
		}
	}
}

// TestBoardAccessChecker_ResolveErrors covers the shared resolve() failure
// modes once (via CheckViewAccess); they are identical across the three methods.
func TestBoardAccessChecker_ResolveErrors(t *testing.T) {
	boardID := uuid.New()
	workspaceID := uuid.New()
	requesterID := uuid.New()

	existingBoard := &entity.Board{ID: boardID, WorkspaceID: workspaceID, CreatedBy: uuid.New(), Visibility: entity.BoardVisibilityWorkspace}
	adminMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleAdmin}

	tests := []struct {
		name       string
		setupMocks func(d checkerDeps)
		wantErr    error
		wantErrMsg string
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
			name: "workspace member nil → ErrUserNotInWorkspace",
			setupMocks: func(d checkerDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(nil, nil)
			},
			wantErr: domain.ErrUserNotInWorkspace,
		},
		{
			name: "workspaceMemberRepo unexpected DB error → wrapped error",
			setupMocks: func(d checkerDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to fetch workspace membership",
		},
		{
			name: "boardMemberRepo unexpected DB error → wrapped error",
			setupMocks: func(d checkerDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to fetch board membership",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := newCheckerDeps(t)
			tt.setupMocks(d)

			access, err := d.checker.CheckViewAccess(context.Background(), boardID, requesterID)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, access)
				return
			}
			require.Error(t, err)
			require.ErrorContains(t, err, tt.wantErrMsg)
			require.Nil(t, access)
		})
	}
}
