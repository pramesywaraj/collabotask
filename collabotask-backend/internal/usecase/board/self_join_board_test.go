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

func TestSelfJoinBoard(t *testing.T) {
	workspaceID := uuid.New()
	boardID := uuid.New()
	requesterID := uuid.New()

	workspaceBoard := &entity.Board{
		ID:          boardID,
		WorkspaceID: workspaceID,
		CreatedBy:   uuid.New(),
		Visibility:  entity.BoardVisibilityWorkspace,
	}
	privateBoard := &entity.Board{
		ID:          boardID,
		WorkspaceID: workspaceID,
		CreatedBy:   uuid.New(),
		Visibility:  entity.BoardVisibilityPrivate,
	}
	adminMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleAdmin}
	regularMember := &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: requesterID, Role: entity.WorkspaceRoleMember}
	existingBoardMember := &entity.BoardMember{BoardID: boardID, UserID: requesterID, Role: entity.BoardRoleMember}

	validInput := board.SelfJoinBoardInput{
		RequesterID: requesterID,
		BoardID:     boardID,
		WorkspaceID: workspaceID,
	}

	tests := []struct {
		name       string
		input      board.SelfJoinBoardInput
		setupMocks func(d boardTestDeps)
		wantErr    error
		wantErrMsg string
		wantJoined bool
	}{
		{
			name:       "invalid input → validation error",
			input:      board.SelfJoinBoardInput{RequesterID: requesterID, BoardID: boardID},
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
			name:  "board archived → ErrBoardNotFound",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				b := *workspaceBoard
				b.IsArchived = true
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(&b, nil)
			},
			wantErr: domain.ErrBoardNotFound,
		},
		{
			name:  "board belongs to different workspace → ErrBoardNotFound",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				b := *workspaceBoard
				b.WorkspaceID = uuid.New()
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(&b, nil)
			},
			wantErr: domain.ErrBoardNotFound,
		},
		{
			name:  "requester not in workspace (ErrMemberNotFound) → ErrUserNotInWorkspace",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(workspaceBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(nil, domain.ErrMemberNotFound)
			},
			wantErr: domain.ErrUserNotInWorkspace,
		},
		{
			name:  "unexpected error fetching board membership → wrapped error",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(workspaceBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "error occurred when fetching board membership",
		},
		{
			name:  "already a member (admin) → idempotent, Joined=false, no insert",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(privateBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(existingBoardMember, nil)
			},
			wantJoined: false,
		},
		{
			name:  "plain member on PRIVATE board → ErrBoardNotFound (existence hidden)",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(privateBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(regularMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
			},
			wantErr: domain.ErrBoardNotFound,
		},
		{
			name:  "plain member on WORKSPACE board → joins, Joined=true",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(workspaceBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(regularMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
				d.boardMbrRepo.EXPECT().CreateIfAbsent(mock.Anything, mock.MatchedBy(func(m *entity.BoardMember) bool {
					return m.BoardID == boardID && m.UserID == requesterID && m.Role == entity.BoardRoleMember
				})).Return(true, nil)
			},
			wantJoined: true,
		},
		{
			name:  "admin on PRIVATE board → break-glass join, Joined=true",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(privateBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
				d.boardMbrRepo.EXPECT().CreateIfAbsent(mock.Anything, mock.Anything).Return(true, nil)
			},
			wantJoined: true,
		},
		{
			name:  "eligible but concurrent insert won the race → Joined=false",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(workspaceBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
				d.boardMbrRepo.EXPECT().CreateIfAbsent(mock.Anything, mock.Anything).Return(false, nil)
			},
			wantJoined: false,
		},
		{
			name:  "CreateIfAbsent fails → wrapped error",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(workspaceBoard, nil)
				d.wsMbrRepo.EXPECT().GetByWorkspaceAndUser(mock.Anything, workspaceID, requesterID).Return(adminMember, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
				d.boardMbrRepo.EXPECT().CreateIfAbsent(mock.Anything, mock.Anything).Return(false, errors.New("db error"))
			},
			wantErrMsg: "failed join to the board",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := newDeps(t)
			tt.setupMocks(d)

			out, err := d.uc.SelfJoinBoard(context.Background(), tt.input)

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
			require.NotNil(t, out)
			require.Equal(t, tt.wantJoined, out.Joined)
		})
	}
}
