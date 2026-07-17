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

func TestSetArchived(t *testing.T) {
	workspaceID := uuid.New()
	boardID := uuid.New()
	requesterID := uuid.New()

	existingBoard := &entity.Board{
		ID:          boardID,
		WorkspaceID: workspaceID,
		Title:       "My Board",
		CreatedBy:   requesterID,
		IsArchived:  false,
	}
	adminMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleAdmin}
	regularMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleMember}

	validInput := board.SetArchivedInput{
		RequesterID: requesterID,
		BoardID:     boardID,
		IsArchived:  boolPtr(true),
	}

	tests := []struct {
		name       string
		input      board.SetArchivedInput
		setupMocks func(d boardTestDeps)
		wantErr    error
		wantErrMsg string
		checkOut   func(t *testing.T, out *board.SetArchivedOutput)
	}{
		{
			name:       "invalid input (IsArchived nil) → validation error",
			input:      board.SetArchivedInput{RequesterID: requesterID, BoardID: boardID, IsArchived: nil},
			setupMocks: func(d boardTestDeps) {},
			wantErrMsg: "validation",
		},
		{
			name:  "board not found (repo error) → ErrBoardNotFound",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(nil, domain.ErrBoardNotFound)
			},
			wantErr: domain.ErrBoardNotFound,
		},
		{
			name:  "board nil/empty → ErrBoardNotFound",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(&entity.Board{}, nil)
			},
			wantErr: domain.ErrBoardNotFound,
		},
		{
			name:  "requester not in workspace → ErrUserNotInWorkspace",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(nil, nil)
			},
			wantErr: domain.ErrUserNotInWorkspace,
		},
		{
			name:  "unexpected error fetching board member → wrapped error",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to fetch board membership",
		},
		{
			name:  "not authorized → ErrBoardPermissionDenied",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				otherBoard := &entity.Board{ID: boardID, WorkspaceID: workspaceID, CreatedBy: uuid.New()}
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(otherBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(regularMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
			},
			wantErr: domain.ErrBoardPermissionDenied,
		},
		{
			// Regression (UC-12e proxy removal): old board creator demoted to BOARD_MEMBER
			// (simulating post-transfer state) must be denied — authority no longer leaks via created_by.
			name:  "old creator with BOARD_MEMBER role (post-transfer) → ErrBoardPermissionDenied",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				demotedCreator := &entity.BoardMember{BoardID: boardID, UserID: requesterID, Role: entity.BoardRoleMember}
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil) // existingBoard.CreatedBy == requesterID
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(regularMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(demotedCreator, nil)
			},
			wantErr: domain.ErrBoardPermissionDenied,
		},
		{
			name:  "boardRepo.SetArchived fails → wrapped error",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
				d.boardRepo.EXPECT().SetArchived(mock.Anything, boardID, true).Return(errors.New("db error"))
			},
			wantErrMsg: "failed to set archived status",
		},
		{
			name:  "success — archive board",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				b := *existingBoard
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(&b, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
				d.boardRepo.EXPECT().SetArchived(mock.Anything, boardID, true).Return(nil)
			},
			checkOut: func(t *testing.T, out *board.SetArchivedOutput) {
				assert.True(t, out.Board.IsArchived)
			},
		},
		{
			name:  "success — unarchive board",
			input: board.SetArchivedInput{RequesterID: requesterID, BoardID: boardID, IsArchived: boolPtr(false)},
			setupMocks: func(d boardTestDeps) {
				b := *existingBoard
				b.IsArchived = true
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(&b, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
				d.boardRepo.EXPECT().SetArchived(mock.Anything, boardID, false).Return(nil)
			},
			checkOut: func(t *testing.T, out *board.SetArchivedOutput) {
				assert.False(t, out.Board.IsArchived)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := newDeps(t)
			d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Maybe().Return(nil)
			tt.setupMocks(d)

			out, err := d.uc.SetArchived(context.Background(), tt.input)

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
