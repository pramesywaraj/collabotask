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

func TestUpdateBoard(t *testing.T) {
	workspaceID := uuid.New()
	boardID := uuid.New()
	requesterID := uuid.New()

	existingBoard := &entity.Board{
		ID:              boardID,
		WorkspaceID:     workspaceID,
		Title:           "Old Title",
		CreatedBy:       requesterID,
		BackgroundColor: "#0079BF",
	}
	adminMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleAdmin}
	ownerBoardMember := &entity.BoardMember{BoardID: boardID, UserID: requesterID, Role: entity.BoardRoleOwner}
	regularMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleMember}

	validInput := board.UpdateBoardInput{
		RequesterID: requesterID,
		BoardID:     boardID,
		Title:       strPtr("New Title"),
	}

	tests := []struct {
		name       string
		input      board.UpdateBoardInput
		setupMocks func(d boardTestDeps)
		wantErr    error
		wantErrMsg string
		checkOut   func(t *testing.T, out *board.UpdateBoardOutput)
	}{
		{
			name:       "no fields provided → ErrAtLeastOneProvided",
			input:      board.UpdateBoardInput{RequesterID: requesterID, BoardID: boardID},
			setupMocks: func(d boardTestDeps) {},
			wantErr:    domain.ErrAtLeastOneProvided,
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
			name:  "unexpected error fetching board → wrapped error",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to fetch board detail",
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
			input: board.UpdateBoardInput{RequesterID: requesterID, BoardID: boardID, Title: strPtr("New Title")},
			setupMocks: func(d boardTestDeps) {
				otherBoard := &entity.Board{ID: boardID, WorkspaceID: workspaceID, CreatedBy: uuid.New()}
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(otherBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(regularMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
			},
			wantErr: domain.ErrBoardPermissionDenied,
		},
		{
			name:  "boardRepo.Update fails → wrapped error",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
				d.boardRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			wantErrMsg: "failed to update the board",
		},
		{
			name:  "success — workspace admin updates without board membership",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				b := *existingBoard
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(&b, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
				d.boardRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
			},
			checkOut: func(t *testing.T, out *board.UpdateBoardOutput) {
				assert.Equal(t, "New Title", out.Board.Title)
			},
		},
		{
			name:  "success — board owner updates title",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				b := *existingBoard
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(&b, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(regularMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(ownerBoardMember, nil)
				d.boardRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
			},
			checkOut: func(t *testing.T, out *board.UpdateBoardOutput) {
				assert.Equal(t, "New Title", out.Board.Title)
			},
		},
		{
			name: "success — description cleared (DescriptionPresent=true, Description=nil)",
			input: board.UpdateBoardInput{
				RequesterID:        requesterID,
				BoardID:            boardID,
				DescriptionPresent: true,
				Description:        nil,
			},
			setupMocks: func(d boardTestDeps) {
				desc := "existing description"
				b := *existingBoard
				b.Description = &desc
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(&b, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
				d.boardRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
			},
			checkOut: func(t *testing.T, out *board.UpdateBoardOutput) {
				assert.Nil(t, out.Board.Description)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := newDeps(t)
			tt.setupMocks(d)

			out, err := d.uc.UpdateBoard(context.Background(), tt.input)

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
