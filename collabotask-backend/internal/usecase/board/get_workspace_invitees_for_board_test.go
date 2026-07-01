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

func TestGetWorkspaceInviteesForBoard(t *testing.T) {
	workspaceID := uuid.New()
	boardID := uuid.New()
	requesterID := uuid.New()
	memberAID := uuid.New()
	memberBID := uuid.New()

	existingBoard := &entity.Board{
		ID:          boardID,
		WorkspaceID: workspaceID,
		CreatedBy:   requesterID,
	}
	adminMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleAdmin}
	regularMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleMember}

	validInput := board.GetWorkspaceInviteesForBoardInput{
		RequesterID: requesterID,
		WorkspaceID: workspaceID,
		BoardID:     boardID,
	}

	tests := []struct {
		name       string
		input      board.GetWorkspaceInviteesForBoardInput
		setupMocks func(d boardTestDeps)
		wantErr    error
		wantErrMsg string
		checkOut   func(t *testing.T, out *board.GetWorkspaceInviteesForBoardOutput)
	}{
		{
			name:       "invalid input → validation error",
			input:      board.GetWorkspaceInviteesForBoardInput{RequesterID: requesterID, WorkspaceID: workspaceID},
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
				otherBoard := &entity.Board{ID: boardID, WorkspaceID: workspaceID, CreatedBy: uuid.New()}
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(otherBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(regularMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
			},
			wantErr: domain.ErrBoardPermissionDenied,
		},
		{
			name:  "GetMembersByWorkspace fails → wrapped error",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
				d.wsMbrRepo.EXPECT().GetMembersByWorkspace(mock.Anything, workspaceID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to list workspace members",
		},
		{
			name:  "GetMembersByBoard fails → wrapped error",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
				d.wsMbrRepo.EXPECT().GetMembersByWorkspace(mock.Anything, workspaceID).Return([]*entity.WorkspaceMember{}, nil)
				d.boardMbrRepo.EXPECT().GetMembersByBoard(mock.Anything, boardID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to list board members",
		},
		{
			name:  "userRepo.GetByIds fails → wrapped error",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				wsMbr := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: memberAID, Role: entity.WorkspaceRoleMember}
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
				d.wsMbrRepo.EXPECT().GetMembersByWorkspace(mock.Anything, workspaceID).Return([]*entity.WorkspaceMember{wsMbr}, nil)
				d.boardMbrRepo.EXPECT().GetMembersByBoard(mock.Anything, boardID).Return([]*entity.BoardMember{}, nil)
				d.userRepo.EXPECT().GetByIds(mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to fetch user details",
		},
		{
			name:  "success — IsBoardMember correctly true/false per member",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				wsMbrA := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: memberAID, Role: entity.WorkspaceRoleMember}
				wsMbrB := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: memberBID, Role: entity.WorkspaceRoleMember}
				boardMbrA := &entity.BoardMember{BoardID: boardID, UserID: memberAID, Role: entity.BoardRoleMember}

				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
				d.wsMbrRepo.EXPECT().GetMembersByWorkspace(mock.Anything, workspaceID).Return([]*entity.WorkspaceMember{wsMbrA, wsMbrB}, nil)
				d.boardMbrRepo.EXPECT().GetMembersByBoard(mock.Anything, boardID).Return([]*entity.BoardMember{boardMbrA}, nil)
				d.userRepo.EXPECT().GetByIds(mock.Anything, mock.Anything).Return(map[uuid.UUID]*entity.User{
					memberAID: {ID: memberAID, Email: "a@example.com", Name: "Alice"},
					memberBID: {ID: memberBID, Email: "b@example.com", Name: "Bob"},
				}, nil)
			},
			checkOut: func(t *testing.T, out *board.GetWorkspaceInviteesForBoardOutput) {
				require.Len(t, out.Members, 2)
				byID := make(map[uuid.UUID]board.BoardInvitee, len(out.Members))
				for _, m := range out.Members {
					byID[m.UserID] = m
				}
				assert.True(t, byID[memberAID].IsBoardMember)
				assert.False(t, byID[memberBID].IsBoardMember)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := newDeps(t)
			tt.setupMocks(d)

			out, err := d.uc.GetWorkspaceInviteesForBoard(context.Background(), tt.input)

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
