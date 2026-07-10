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

func TestGetBoardKanban(t *testing.T) {
	boardID := uuid.New()
	requesterID := uuid.New()
	columnID := uuid.New()
	assigneeID := uuid.New()

	existingBoard := &entity.Board{ID: boardID}

	validInput := board.GetBoardKanbanInput{
		RequesterID: requesterID,
		BoardID:     boardID,
	}

	tests := []struct {
		name       string
		input      board.GetBoardKanbanInput
		setupMocks func(d boardTestDeps)
		wantErr    error
		wantErrMsg string
		checkOut   func(t *testing.T, out *board.GetBoardKanbanOutput)
	}{
		{
			name:       "invalid input → validation error",
			input:      board.GetBoardKanbanInput{RequesterID: requesterID},
			setupMocks: func(d boardTestDeps) {},
			wantErrMsg: "validation",
		},
		{
			name:  "boardAccessChecker.Check fails → error propagated",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.checker.EXPECT().CheckViewAccess(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardNotFound)
			},
			wantErr: domain.ErrBoardNotFound,
		},
		{
			name:  "columnRepo.GetColumnsByBoard fails → wrapped error",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.checker.EXPECT().CheckViewAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{Board: existingBoard}, nil)
				d.columnRepo.EXPECT().GetColumnsByBoard(mock.Anything, boardID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to fetch list of columns",
		},
		{
			name:  "cardRepo.GetCardsByColumn fails → wrapped error",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				col := &entity.Column{ID: columnID, BoardID: boardID, Title: "To Do"}
				d.checker.EXPECT().CheckViewAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{Board: existingBoard}, nil)
				d.columnRepo.EXPECT().GetColumnsByBoard(mock.Anything, boardID).Return([]*entity.Column{col}, nil)
				d.cardRepo.EXPECT().GetCardsByColumn(mock.Anything, columnID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to list cards",
		},
		{
			name:  "userRepo.GetByIds fails → wrapped error",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				col := &entity.Column{ID: columnID, BoardID: boardID, Title: "To Do"}
				card := &entity.Card{ID: uuid.New(), ColumnID: columnID, Title: "Card", AssignedTo: &assigneeID}
				d.checker.EXPECT().CheckViewAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{Board: existingBoard}, nil)
				d.columnRepo.EXPECT().GetColumnsByBoard(mock.Anything, boardID).Return([]*entity.Column{col}, nil)
				d.cardRepo.EXPECT().GetCardsByColumn(mock.Anything, columnID).Return([]*entity.Card{card}, nil)
				d.userRepo.EXPECT().GetByIds(mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to fetch assignees",
		},
		{
			name:  "success — empty board (no columns)",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				d.checker.EXPECT().CheckViewAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{Board: existingBoard}, nil)
				d.columnRepo.EXPECT().GetColumnsByBoard(mock.Anything, boardID).Return([]*entity.Column{}, nil)
				d.userRepo.EXPECT().GetByIds(mock.Anything, mock.Anything).Return(map[uuid.UUID]*entity.User{}, nil)
			},
			checkOut: func(t *testing.T, out *board.GetBoardKanbanOutput) {
				assert.Empty(t, out.Columns)
			},
		},
		{
			name:  "success — with columns, cards, and assignees (deduplicates assignee IDs)",
			input: validInput,
			setupMocks: func(d boardTestDeps) {
				col := &entity.Column{ID: columnID, BoardID: boardID, Title: "To Do"}
				card1 := &entity.Card{ID: uuid.New(), ColumnID: columnID, Title: "Card 1", AssignedTo: &assigneeID}
				card2 := &entity.Card{ID: uuid.New(), ColumnID: columnID, Title: "Card 2", AssignedTo: &assigneeID}
				assignee := &entity.User{ID: assigneeID, Name: "Alice"}

				d.checker.EXPECT().CheckViewAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{Board: existingBoard}, nil)
				d.columnRepo.EXPECT().GetColumnsByBoard(mock.Anything, boardID).Return([]*entity.Column{col}, nil)
				d.cardRepo.EXPECT().GetCardsByColumn(mock.Anything, columnID).Return([]*entity.Card{card1, card2}, nil)
				// assigneeID appears twice but GetByIds must be called with it only once
				d.userRepo.EXPECT().GetByIds(mock.Anything, mock.MatchedBy(func(ids []uuid.UUID) bool {
					return len(ids) == 1 && ids[0] == assigneeID
				})).Return(map[uuid.UUID]*entity.User{assigneeID: assignee}, nil)
			},
			checkOut: func(t *testing.T, out *board.GetBoardKanbanOutput) {
				require.Len(t, out.Columns, 1)
				require.Len(t, out.Columns[0].Cards, 2)
				assert.Equal(t, "Alice", out.Columns[0].Cards[0].Assignee.Name)
				assert.Equal(t, "Alice", out.Columns[0].Cards[1].Assignee.Name)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := newDeps(t)
			tt.setupMocks(d)

			out, err := d.uc.GetBoardKanban(context.Background(), tt.input)

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
