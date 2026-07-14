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

func TestLeaveBoard(t *testing.T) {
	boardID := uuid.New()
	requesterID := uuid.New()
	ownerID := uuid.New()

	existingBoard := &entity.Board{
		ID:        boardID,
		CreatedBy: ownerID,
	}
	boardMember := &entity.BoardMember{BoardID: boardID, UserID: requesterID, Role: entity.BoardRoleMember}
	ownerBoardMember := &entity.BoardMember{BoardID: boardID, UserID: ownerID, Role: entity.BoardRoleOwner}

	validInput := board.LeaveBoardInput{
		RequesterID: requesterID,
		BoardID:     boardID,
	}

	tests := []struct {
		name       string
		input      board.LeaveBoardInput
		setupMocks func(d boardTestDeps)
		wantErr    error
		wantErrMsg string
	}{
		{
			name:       "invalid input → validation error",
			input:      board.LeaveBoardInput{RequesterID: requesterID},
			setupMocks: func(d boardTestDeps) {},
			wantErrMsg: "validation",
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
			name:  "board is nil → ErrBoardNotFound",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(nil, nil)
			},
			wantErr: domain.ErrBoardNotFound,
		},
		{
			name:  "board is empty → ErrBoardNotFound",
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
			// Regression (UC-12e proxy removal): guard is now role-based, not created_by-based.
			name:  "BOARD_OWNER tries to leave → ErrBoardOwnerCannotLeave",
			input: board.LeaveBoardInput{RequesterID: ownerID, BoardID: boardID},
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, ownerID).Return(ownerBoardMember, nil)
			},
			wantErr: domain.ErrBoardOwnerCannotLeave,
		},
		{
			// Regression (UC-12e proxy removal): a BOARD_MEMBER who happens to match
			// created_by must be allowed to leave — the guard is role-based, not creator-based.
			name:  "BOARD_MEMBER who matches created_by → allowed to leave",
			input: board.LeaveBoardInput{RequesterID: ownerID, BoardID: boardID},
			setupMocks: func(d boardTestDeps) {
				creatorAsMember := &entity.BoardMember{BoardID: boardID, UserID: ownerID, Role: entity.BoardRoleMember}
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, ownerID).Return(creatorAsMember, nil)
				d.boardMbrRepo.EXPECT().Delete(mock.Anything, boardID, ownerID).Return(nil)
			},
		},
		{
			name:  "requester not a board member → ErrBoardMemberNotFound",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardMemberNotFound)
			},
			wantErr: domain.ErrBoardMemberNotFound,
		},
		{
			name:  "unexpected error fetching board membership → wrapped error",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to check requester membership in board",
		},
		{
			name:  "boardMemberRepo.Delete fails → wrapped error",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(boardMember, nil)
				d.boardMbrRepo.EXPECT().Delete(mock.Anything, boardID, requesterID).Return(errors.New("db error"))
			},
			wantErrMsg: "failed to remove member from board",
		},
		{
			name:  "success",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.boardRepo.EXPECT().GetByID(mock.Anything, boardID).Return(existingBoard, nil)
				d.boardMbrRepo.EXPECT().GetMemberByBoardAndUser(mock.Anything, boardID, requesterID).Return(boardMember, nil)
				d.boardMbrRepo.EXPECT().Delete(mock.Anything, boardID, requesterID).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := newDeps(t)
			tt.setupMocks(d)

			err := d.uc.LeaveBoard(context.Background(), tt.input)

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
