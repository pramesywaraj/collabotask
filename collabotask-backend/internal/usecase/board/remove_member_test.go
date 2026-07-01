package board_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/usecase/board"
)

func TestRemoveMemberBoard(t *testing.T) {
	workspaceID := uuid.New()
	boardID := uuid.New()
	requesterID := uuid.New()
	targetID := uuid.New()

	existingBoard := &entity.Board{
		ID:          boardID,
		WorkspaceID: workspaceID,
		CreatedBy:   uuid.New(),
	}
	adminMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleAdmin}
	regularMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleMember}

	validInput := board.RemoveMemberInput{
		RequesterID: requesterID,
		WorkspaceID: workspaceID,
		BoardID:     boardID,
		UserID:      targetID,
	}

	tests := []struct {
		name       string
		input      board.RemoveMemberInput
		setupMocks func(d boardTestDeps)
		wantErr    error
		wantErrMsg string
	}{
		{
			name:       "invalid input → validation error",
			input:      board.RemoveMemberInput{RequesterID: requesterID, WorkspaceID: workspaceID, BoardID: boardID},
			setupMocks: func(d boardTestDeps) {},
			wantErrMsg: "validation",
		},
		{
			name:       "RequesterID == UserID → ErrCannotRemoveYourself",
			input:      board.RemoveMemberInput{RequesterID: requesterID, WorkspaceID: workspaceID, BoardID: boardID, UserID: requesterID},
			setupMocks: func(d boardTestDeps) {},
			wantErr:    domain.ErrCannotRemoveYourself,
		},
		{
			name:  "board not found (repo ErrBoardNotFound) → ErrBoardNotFound",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(nil, domain.ErrBoardNotFound)
			},
			wantErr: domain.ErrBoardNotFound,
		},
		{
			name:  "unexpected error fetching board → wrapped error",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to verify board existence",
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
			name:  "board archived → ErrBoardNotFound",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				b := *existingBoard
				b.IsArchived = true
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(&b, nil)
			},
			wantErr: domain.ErrBoardNotFound,
		},
		{
			name:  "board belongs to different workspace → ErrBoardNotFound",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				b := *existingBoard
				b.WorkspaceID = uuid.New()
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(&b, nil)
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
			name:  "unexpected error fetching requester's board member → wrapped error",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to check requester permission",
		},
		{
			name:  "not authorized → ErrBoardPermissionDenied",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(regularMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
			},
			wantErr: domain.ErrBoardPermissionDenied,
		},
		{
			name:  "target member not found → ErrBoardMemberNotFound",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
				d.boardMbrRepo.EXPECT().Delete(mock.Anything, boardID, targetID).Return(domain.ErrBoardMemberNotFound)
			},
			wantErr: domain.ErrBoardMemberNotFound,
		},
		{
			name:  "Delete fails with unexpected error → wrapped error",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
				d.boardMbrRepo.EXPECT().Delete(mock.Anything, boardID, targetID).Return(errors.New("db error"))
			},
			wantErrMsg: "failed to remove member from board",
		},
		{
			name:  "success",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
				d.boardMbrRepo.EXPECT().Delete(mock.Anything, boardID, targetID).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := newDeps(t)
			tt.setupMocks(d)

			err := d.uc.RemoveMember(context.Background(), tt.input)

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
		})
	}
}
