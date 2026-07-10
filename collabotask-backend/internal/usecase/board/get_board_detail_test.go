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
	"collabotask/internal/usecase/common"
)

func TestGetBoardDetail(t *testing.T) {
	boardID := uuid.New()
	requesterID := uuid.New()
	memberUserID := uuid.New()

	workspaceBoard := &entity.Board{
		ID:         boardID,
		Title:      "My Board",
		CreatedBy:  uuid.New(),
		Visibility: entity.BoardVisibilityWorkspace,
	}
	privateBoard := &entity.Board{
		ID:         boardID,
		Title:      "My Board",
		CreatedBy:  uuid.New(),
		Visibility: entity.BoardVisibilityPrivate,
	}

	validInput := board.GetBoardDetailInput{
		RequesterID: requesterID,
		BoardID:     boardID,
	}

	tests := []struct {
		name       string
		input      board.GetBoardDetailInput
		setupMocks func(d boardTestDeps)
		wantErr    error
		wantErrMsg string
		checkOut   func(t *testing.T, out *board.GetBoardDetailOutput)
	}{
		{
			name:       "invalid input → validation error",
			input:      board.GetBoardDetailInput{RequesterID: requesterID},
			setupMocks: func(d boardTestDeps) {},
			wantErrMsg: "validation",
		},
		{
			name:  "checker denies (plain member on PRIVATE) → ErrBoardNotFound propagated",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.checker.EXPECT().CheckMetadataAccess(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardNotFound)
			},
			wantErr: domain.ErrBoardNotFound,
		},
		{
			name:  "boardMemberRepo.GetMembersByBoard fails → wrapped error",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.checker.EXPECT().CheckMetadataAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{Board: workspaceBoard}, nil)
				d.boardMbrRepo.EXPECT().GetMembersByBoard(mock.Anything, boardID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to fetch board members",
		},
		{
			name:  "userRepo.GetByIds fails → wrapped error",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				member := &entity.BoardMember{BoardID: boardID, UserID: memberUserID, Role: entity.BoardRoleMember}
				d.checker.EXPECT().CheckMetadataAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{Board: workspaceBoard}, nil)
				d.boardMbrRepo.EXPECT().GetMembersByBoard(mock.Anything, boardID).Return([]*entity.BoardMember{member}, nil)
				d.userRepo.EXPECT().GetByIds(mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to fetch members details",
		},
		{
			name:  "user missing from users map → ErrUserNotFound",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				member := &entity.BoardMember{BoardID: boardID, UserID: memberUserID, Role: entity.BoardRoleMember}
				d.checker.EXPECT().CheckMetadataAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{Board: workspaceBoard}, nil)
				d.boardMbrRepo.EXPECT().GetMembersByBoard(mock.Anything, boardID).Return([]*entity.BoardMember{member}, nil)
				d.userRepo.EXPECT().GetByIds(mock.Anything, mock.Anything).Return(map[uuid.UUID]*entity.User{}, nil)
			},
			wantErr: domain.ErrUserNotFound,
		},
		{
			name:  "success — requester is board member (accessStatus=JOINED, role set, full roster)",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				bm := &entity.BoardMember{BoardID: boardID, UserID: requesterID, Role: entity.BoardRoleMember}
				access := &common.BoardAccess{Board: privateBoard, BoardMember: bm}
				d.checker.EXPECT().CheckMetadataAccess(mock.Anything, boardID, requesterID).Return(access, nil)
				d.boardMbrRepo.EXPECT().GetMembersByBoard(mock.Anything, boardID).Return([]*entity.BoardMember{bm}, nil)
				d.userRepo.EXPECT().GetByIds(mock.Anything, mock.Anything).Return(map[uuid.UUID]*entity.User{
					requesterID: {ID: requesterID, Email: "r@example.com", Name: "Requester"},
				}, nil)
			},
			checkOut: func(t *testing.T, out *board.GetBoardDetailOutput) {
				assert.Equal(t, entity.BoardJoined, out.Board.AccessStatus)
				require.NotNil(t, out.Board.UserRole)
				assert.Equal(t, entity.BoardRoleMember, *out.Board.UserRole)
				assert.Len(t, out.Board.Members, 1)
			},
		},
		{
			name:  "success — non-joined admin on WORKSPACE board (CAN_JOIN, nil role, roster shown)",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				access := &common.BoardAccess{Board: workspaceBoard, BoardMember: nil}
				d.checker.EXPECT().CheckMetadataAccess(mock.Anything, boardID, requesterID).Return(access, nil)
				d.boardMbrRepo.EXPECT().GetMembersByBoard(mock.Anything, boardID).Return([]*entity.BoardMember{}, nil)
				d.userRepo.EXPECT().GetByIds(mock.Anything, mock.Anything).Return(map[uuid.UUID]*entity.User{}, nil)
			},
			checkOut: func(t *testing.T, out *board.GetBoardDetailOutput) {
				assert.Equal(t, entity.BoardCanJoin, out.Board.AccessStatus)
				assert.Nil(t, out.Board.UserRole)
			},
		},
		{
			name:  "success — non-joined admin on PRIVATE board (thin roster: Members empty, no roster fetch)",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				access := &common.BoardAccess{Board: privateBoard, BoardMember: nil}
				d.checker.EXPECT().CheckMetadataAccess(mock.Anything, boardID, requesterID).Return(access, nil)
				// GetMembersByBoard / GetByIds must NOT be called — roster is hidden.
			},
			checkOut: func(t *testing.T, out *board.GetBoardDetailOutput) {
				assert.Equal(t, entity.BoardCanJoin, out.Board.AccessStatus)
				assert.Nil(t, out.Board.UserRole)
				assert.Empty(t, out.Board.Members)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := newDeps(t)
			tt.setupMocks(d)

			out, err := d.uc.GetBoardDetail(context.Background(), tt.input)

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
