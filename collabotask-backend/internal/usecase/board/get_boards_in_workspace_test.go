package board_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/usecase/board"
)

func TestGetBoardsInWorkspace(t *testing.T) {
	workspaceID := uuid.New()
	requesterID := uuid.New()
	boardID := uuid.New()

	wsMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleMember}

	validInput := board.GetBoardsInput{
		WorkspaceID: workspaceID,
		RequesterID: requesterID,
	}

	tests := []struct {
		name       string
		input      board.GetBoardsInput
		setupMocks func(d boardTestDeps)
		wantErr    error
		wantErrMsg string
		checkOut   func(t *testing.T, out *board.GetBoardsOutput)
	}{
		{
			name:       "invalid input → validation error",
			input:      board.GetBoardsInput{WorkspaceID: workspaceID},
			setupMocks: func(d boardTestDeps) {},
			wantErrMsg: "validation",
		},
		{
			name:  "workspaceMemberRepo.GetByWorkspaceAndUser returns DB error → wrapped error",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to check user existence in workspace",
		},
		{
			name:  "requester not in workspace → ErrUserNotInWorkspace",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(nil, nil)
			},
			wantErr: domain.ErrUserNotInWorkspace,
		},
		{
			name:  "boardRepo.GetUserBoardsInWorkspace fails → wrapped error",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(wsMember, nil)
				d.boardRepo.EXPECT().GetUserBoardsInWorkspace(mock.Anything, workspaceID, requesterID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to fetch user boards in workspace",
		},
		{
			name:  "success — empty list",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(wsMember, nil)
				d.boardRepo.EXPECT().GetUserBoardsInWorkspace(mock.Anything, workspaceID, requesterID).Return([]*entity.BoardListItem{}, nil)
			},
			checkOut: func(t *testing.T, out *board.GetBoardsOutput) {
				assert.Empty(t, out.Boards)
			},
		},
		{
			name:  "success — with boards (UserRole mapped correctly)",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(wsMember, nil)
				d.boardRepo.EXPECT().GetUserBoardsInWorkspace(mock.Anything, workspaceID, requesterID).Return([]*entity.BoardListItem{
					{
						Board:        entity.Board{ID: boardID, WorkspaceID: workspaceID, Title: "Board A"},
						UserRole:     entity.BoardRoleMember,
						AccessStatus: entity.BoardJoined,
						MemberCount:  3,
					},
					{
						Board:        entity.Board{ID: uuid.New(), WorkspaceID: workspaceID, Title: "Board B"},
						UserRole:     "",
						AccessStatus: entity.BoardCanJoin,
						MemberCount:  1,
					},
				}, nil)
			},
			checkOut: func(t *testing.T, out *board.GetBoardsOutput) {
				require.Len(t, out.Boards, 2)

				first := out.Boards[0]
				assert.Equal(t, boardID, first.Board.ID)
				require.NotNil(t, first.UserRole)
				assert.Equal(t, entity.BoardRoleMember, *first.UserRole)
				assert.Equal(t, entity.BoardJoined, first.AccessStatus)
				assert.Equal(t, uint(3), first.MemberCount)

				second := out.Boards[1]
				assert.Nil(t, second.UserRole)
				assert.Equal(t, entity.BoardCanJoin, second.AccessStatus)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := newDeps(t)
			tt.setupMocks(d)

			out, err := d.uc.GetBoardsInWorkspace(context.Background(), tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, out)
				return
			}
			if tt.wantErrMsg != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErrMsg)
				assert.Nil(t, out)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, out)
			if tt.checkOut != nil {
				tt.checkOut(t, out)
			}
		})
	}
}
