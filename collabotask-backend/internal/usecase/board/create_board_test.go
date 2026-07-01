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

func TestCreateBoard(t *testing.T) {
	workspaceID := uuid.New()
	requesterID := uuid.New()
	boardID := uuid.New()

	validInput := board.CreateBoardInput{
		WorkspaceID: workspaceID,
		Title:       "My Board",
		RequesterID: requesterID,
	}

	tests := []struct {
		name       string
		input      board.CreateBoardInput
		setupMocks func(d boardTestDeps)
		wantErr    error
		wantErrMsg string
		checkOut   func(t *testing.T, out *board.CreateBoardOutput)
	}{
		{
			name:       "invalid input → validation error",
			input:      board.CreateBoardInput{WorkspaceID: workspaceID, Title: "", RequesterID: requesterID},
			setupMocks: func(d boardTestDeps) {},
			wantErrMsg: "validation",
		},
		{
			name:  "workspaceMemberRepo.IsUserExists returns DB error → wrapped error",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.wsMbrRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, requesterID).Return(false, errors.New("db error"))
			},
			wantErrMsg: "failed to check workspace membership",
		},
		{
			name:  "user not in workspace → ErrUserNotInWorkspace",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.wsMbrRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, requesterID).Return(false, nil)
			},
			wantErr: domain.ErrUserNotInWorkspace,
		},
		{
			name:  "boardRepo.CreateWithOwner fails → wrapped error",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.wsMbrRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, requesterID).Return(true, nil)
				d.boardRepo.EXPECT().CreateWithOwner(mock.Anything, mock.Anything, requesterID).Return(errors.New("db error"))
			},
			wantErrMsg: "failed to create board",
		},
		{
			name:  "success — uses custom background color",
			input: board.CreateBoardInput{WorkspaceID: workspaceID, Title: "My Board", RequesterID: requesterID, BackgroundColor: strPtr("#ABC")},
			setupMocks: func(d boardTestDeps) {
				d.wsMbrRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, requesterID).Return(true, nil)
				d.boardRepo.EXPECT().CreateWithOwner(mock.Anything, mock.MatchedBy(func(b *entity.Board) bool {
					return b.BackgroundColor == "#ABC"
				}), requesterID).Run(func(_ context.Context, b *entity.Board, _ uuid.UUID) {
					b.ID = boardID
				}).Return(nil)
			},
			checkOut: func(t *testing.T, out *board.CreateBoardOutput) {
				assert.Equal(t, boardID, out.Board.ID)
				assert.Equal(t, "#ABC", out.Board.BackgroundColor)
			},
		},
		{
			name:  "success — falls back to default background color",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.wsMbrRepo.EXPECT().IsUserExists(mock.Anything, workspaceID, requesterID).Return(true, nil)
				d.boardRepo.EXPECT().CreateWithOwner(mock.Anything, mock.MatchedBy(func(b *entity.Board) bool {
					return b.BackgroundColor == "#0079BF"
				}), requesterID).Run(func(_ context.Context, b *entity.Board, _ uuid.UUID) {
					b.ID = boardID
				}).Return(nil)
			},
			checkOut: func(t *testing.T, out *board.CreateBoardOutput) {
				assert.Equal(t, boardID, out.Board.ID)
				assert.Equal(t, "#0079BF", out.Board.BackgroundColor)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := newDeps(t)
			tt.setupMocks(d)

			out, err := d.uc.CreateBoard(context.Background(), tt.input)

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
