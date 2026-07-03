package column_test

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
	"collabotask/internal/usecase/column"
)

func TestUpdateColumnPosition(t *testing.T) {
	boardID := uuid.New()
	columnID := uuid.New()
	requesterID := uuid.New()

	board := &entity.Board{ID: boardID}
	// Used by error-path cases that don't reach the reorder step.
	existingColumn := &entity.Column{ID: columnID, BoardID: boardID, Position: 1}

	// Stable IDs for the sibling columns so the reorder assertions can pin the
	// exact resulting order. Fixed values (not uuid.New) where ID sort-order matters.
	otherA := uuid.New()
	otherB := uuid.New()

	validInput := column.UpdateColumnPositionInput{
		BoardID:     boardID,
		ColumnID:    columnID,
		Position:    2,
		RequesterID: requesterID,
	}

	// mkCol builds a column on boardID with the given position.
	mkCol := func(id uuid.UUID, pos int) *entity.Column {
		return &entity.Column{ID: id, BoardID: boardID, Position: pos, Title: "Col"}
	}

	// assertOrder matches the slice handed to ReorderPositions against the exact
	// expected ID order, requiring each column to be renumbered to its index
	// (contiguous 0..n-1). This verifies the whole sequence, not just the moved column.
	assertOrder := func(ids ...uuid.UUID) any {
		return mock.MatchedBy(func(cols []*entity.Column) bool {
			if len(cols) != len(ids) {
				return false
			}
			for i, c := range cols {
				if c.ID != ids[i] || c.Position != i {
					return false
				}
			}
			return true
		})
	}

	tests := []struct {
		name       string
		input      column.UpdateColumnPositionInput
		setupMocks func(d columnTestDeps)
		wantErr    error
		wantErrMsg string
		checkOut   func(t *testing.T, out *column.UpdateColumnPositionOutput)
	}{
		{
			name:       "invalid input (nil BoardID/ColumnID/RequesterID) → validation error",
			input:      column.UpdateColumnPositionInput{BoardID: boardID, ColumnID: columnID, Position: 1},
			setupMocks: func(d columnTestDeps) {},
			wantErrMsg: "validation",
		},
		{
			name:  "column not found (ErrColumnNotFound) → ErrColumnNotFound",
			input: validInput,
			setupMocks: func(d columnTestDeps) {
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(nil, domain.ErrColumnNotFound)
			},
			wantErr: domain.ErrColumnNotFound,
		},
		{
			name:  "columnRepo.GetByID returns unexpected error → wrapped error",
			input: validInput,
			setupMocks: func(d columnTestDeps) {
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to fetch column",
		},
		{
			name:  "column belongs to different board → ErrColumnNotInBoard",
			input: validInput,
			setupMocks: func(d columnTestDeps) {
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(
					&entity.Column{ID: columnID, BoardID: uuid.New()}, nil,
				)
			},
			wantErr: domain.ErrColumnNotInBoard,
		},
		{
			name:  "boardAccessChecker.Check fails → error propagated",
			input: validInput,
			setupMocks: func(d columnTestDeps) {
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(existingColumn, nil)
				d.checker.EXPECT().Check(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardAccessDenied)
			},
			wantErr: domain.ErrBoardAccessDenied,
		},
		{
			name:  "columnRepo.GetColumnsByBoard fails → wrapped error",
			input: validInput,
			setupMocks: func(d columnTestDeps) {
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(existingColumn, nil)
				d.checker.EXPECT().Check(mock.Anything, boardID, requesterID).Return(board, nil)
				d.columnRepo.EXPECT().GetColumnsByBoard(mock.Anything, boardID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to list columns for board",
		},
		{
			name:  "column missing from board list → ErrInconsistentState",
			input: validInput,
			setupMocks: func(d columnTestDeps) {
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(existingColumn, nil)
				d.checker.EXPECT().Check(mock.Anything, boardID, requesterID).Return(board, nil)
				// Return a list that does not contain columnID.
				d.columnRepo.EXPECT().GetColumnsByBoard(mock.Anything, boardID).Return(
					[]*entity.Column{mkCol(uuid.New(), 0)}, nil,
				)
			},
			wantErr: domain.ErrInconsistentState,
		},
		{
			// Target column is at index 1 of 3; target = 1 → oldIdx == newIdx, early return.
			// ReorderPositions must NOT be called (mock will fail the test if it is).
			name: "position == current (no-op) → returns unchanged column, no ReorderPositions call",
			input: column.UpdateColumnPositionInput{
				BoardID:     boardID,
				ColumnID:    columnID,
				Position:    1,
				RequesterID: requesterID,
			},
			setupMocks: func(d columnTestDeps) {
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(existingColumn, nil)
				d.checker.EXPECT().Check(mock.Anything, boardID, requesterID).Return(board, nil)
				d.columnRepo.EXPECT().GetColumnsByBoard(mock.Anything, boardID).Return([]*entity.Column{
					mkCol(uuid.New(), 0),
					{ID: columnID, BoardID: boardID, Position: 1, Title: "Col"},
					mkCol(uuid.New(), 2),
				}, nil)
			},
			checkOut: func(t *testing.T, out *column.UpdateColumnPositionOutput) {
				assert.Equal(t, columnID, out.Column.ID)
				assert.Equal(t, 1, out.Column.Position)
			},
		},
		{
			// Regression guard for the Position=0 bug fix: target column is at index 2 of 3,
			// moving to index 0 (first slot). Before the fix, Position=0 was rejected by
			// validate:"required" on the int field.
			name: "position = 0 → moves to front",
			input: column.UpdateColumnPositionInput{
				BoardID:     boardID,
				ColumnID:    columnID,
				Position:    0,
				RequesterID: requesterID,
			},
			setupMocks: func(d columnTestDeps) {
				// targetCol appears in both GetByID and GetColumnsByBoard so the
				// reorder loop updates its Position in-place.
				targetCol := &entity.Column{ID: columnID, BoardID: boardID, Position: 2, Title: "Col"}
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(targetCol, nil)
				d.checker.EXPECT().Check(mock.Anything, boardID, requesterID).Return(board, nil)
				d.columnRepo.EXPECT().GetColumnsByBoard(mock.Anything, boardID).Return([]*entity.Column{
					mkCol(otherA, 0),
					mkCol(otherB, 1),
					targetCol,
				}, nil)
				// [A, B, target] → move target to front → [target, A, B]
				d.columnRepo.EXPECT().ReorderPositions(mock.Anything, assertOrder(columnID, otherA, otherB)).Return(nil)
			},
			checkOut: func(t *testing.T, out *column.UpdateColumnPositionOutput) {
				assert.Equal(t, columnID, out.Column.ID)
				assert.Equal(t, 0, out.Column.Position)
			},
		},
		{
			// Target column is at index 2 of 3; target=-1 is clamped to 0.
			// Column must start off-edge (not at index 0) so clamping produces a real move.
			name: "negative position clamped to 0",
			input: column.UpdateColumnPositionInput{
				BoardID:     boardID,
				ColumnID:    columnID,
				Position:    -1,
				RequesterID: requesterID,
			},
			setupMocks: func(d columnTestDeps) {
				targetCol := &entity.Column{ID: columnID, BoardID: boardID, Position: 2, Title: "Col"}
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(targetCol, nil)
				d.checker.EXPECT().Check(mock.Anything, boardID, requesterID).Return(board, nil)
				d.columnRepo.EXPECT().GetColumnsByBoard(mock.Anything, boardID).Return([]*entity.Column{
					mkCol(otherA, 0),
					mkCol(otherB, 1),
					targetCol,
				}, nil)
				// [A, B, target] → clamp -1 to 0 → [target, A, B]
				d.columnRepo.EXPECT().ReorderPositions(mock.Anything, assertOrder(columnID, otherA, otherB)).Return(nil)
			},
			checkOut: func(t *testing.T, out *column.UpdateColumnPositionOutput) {
				assert.Equal(t, columnID, out.Column.ID)
				assert.Equal(t, 0, out.Column.Position)
			},
		},
		{
			// Target column is at index 0 of 3; target=99 is clamped to last index (2).
			// Column must start off-edge (not at index 2) so clamping produces a real move.
			name: "position > last index clamped to last",
			input: column.UpdateColumnPositionInput{
				BoardID:     boardID,
				ColumnID:    columnID,
				Position:    99,
				RequesterID: requesterID,
			},
			setupMocks: func(d columnTestDeps) {
				targetCol := &entity.Column{ID: columnID, BoardID: boardID, Position: 0, Title: "Col"}
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(targetCol, nil)
				d.checker.EXPECT().Check(mock.Anything, boardID, requesterID).Return(board, nil)
				d.columnRepo.EXPECT().GetColumnsByBoard(mock.Anything, boardID).Return([]*entity.Column{
					targetCol,
					mkCol(otherA, 1),
					mkCol(otherB, 2),
				}, nil)
				// [target, A, B] → clamp 99 to last → [A, B, target]
				d.columnRepo.EXPECT().ReorderPositions(mock.Anything, assertOrder(otherA, otherB, columnID)).Return(nil)
			},
			checkOut: func(t *testing.T, out *column.UpdateColumnPositionOutput) {
				assert.Equal(t, columnID, out.Column.ID)
				assert.Equal(t, 2, out.Column.Position)
			},
		},
		{
			name: "columnRepo.ReorderPositions fails → error",
			input: column.UpdateColumnPositionInput{
				BoardID:     boardID,
				ColumnID:    columnID,
				Position:    0,
				RequesterID: requesterID,
			},
			setupMocks: func(d columnTestDeps) {
				targetCol := &entity.Column{ID: columnID, BoardID: boardID, Position: 1, Title: "Col"}
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(targetCol, nil)
				d.checker.EXPECT().Check(mock.Anything, boardID, requesterID).Return(board, nil)
				d.columnRepo.EXPECT().GetColumnsByBoard(mock.Anything, boardID).Return([]*entity.Column{
					mkCol(uuid.New(), 0),
					targetCol,
				}, nil)
				d.columnRepo.EXPECT().ReorderPositions(mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			wantErrMsg: "db error",
		},
		{
			// Target column at index 0 of 3; move to index 2 (forward).
			name: "success — move forward (higher index)",
			input: column.UpdateColumnPositionInput{
				BoardID:     boardID,
				ColumnID:    columnID,
				Position:    2,
				RequesterID: requesterID,
			},
			setupMocks: func(d columnTestDeps) {
				targetCol := &entity.Column{ID: columnID, BoardID: boardID, Position: 0, Title: "Col"}
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(targetCol, nil)
				d.checker.EXPECT().Check(mock.Anything, boardID, requesterID).Return(board, nil)
				d.columnRepo.EXPECT().GetColumnsByBoard(mock.Anything, boardID).Return([]*entity.Column{
					targetCol,
					mkCol(otherA, 1),
					mkCol(otherB, 2),
				}, nil)
				// [target, A, B] → move target to index 2 → [A, B, target]
				d.columnRepo.EXPECT().ReorderPositions(mock.Anything, assertOrder(otherA, otherB, columnID)).Return(nil)
			},
			checkOut: func(t *testing.T, out *column.UpdateColumnPositionOutput) {
				assert.Equal(t, columnID, out.Column.ID)
				assert.Equal(t, 2, out.Column.Position)
			},
		},
		{
			// Target column at index 2 of 3; move to index 1 (backward).
			name: "success — move backward (lower index)",
			input: column.UpdateColumnPositionInput{
				BoardID:     boardID,
				ColumnID:    columnID,
				Position:    1,
				RequesterID: requesterID,
			},
			setupMocks: func(d columnTestDeps) {
				targetCol := &entity.Column{ID: columnID, BoardID: boardID, Position: 2, Title: "Col"}
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(targetCol, nil)
				d.checker.EXPECT().Check(mock.Anything, boardID, requesterID).Return(board, nil)
				d.columnRepo.EXPECT().GetColumnsByBoard(mock.Anything, boardID).Return([]*entity.Column{
					mkCol(otherA, 0),
					mkCol(otherB, 1),
					targetCol,
				}, nil)
				// [A, B, target] → move target to index 1 → [A, target, B]
				d.columnRepo.EXPECT().ReorderPositions(mock.Anything, assertOrder(otherA, columnID, otherB)).Return(nil)
			},
			checkOut: func(t *testing.T, out *column.UpdateColumnPositionOutput) {
				assert.Equal(t, columnID, out.Column.ID)
				assert.Equal(t, 1, out.Column.Position)
			},
		},
		{
			// GetColumnsByBoard returns rows out of order AND with a duplicate position
			// (tieLow & tieHigh both at 0). The use case must sort by position, breaking
			// ties by ID, before locating the target. tieLow's ID sorts before tieHigh's,
			// so sorted order is [tieLow(0), tieHigh(0), target(2)] and the target sits at
			// index 2; moving it to 0 yields [target, tieLow, tieHigh]. Without the sort,
			// the target would be found at input index 0, target position 0 would be a
			// no-op, and ReorderPositions would never be called — failing this test.
			name: "success — sorts unsorted input and breaks position ties by ID",
			input: column.UpdateColumnPositionInput{
				BoardID:     boardID,
				ColumnID:    columnID,
				Position:    0,
				RequesterID: requesterID,
			},
			setupMocks: func(d columnTestDeps) {
				tieLow := uuid.MustParse("00000000-0000-0000-0000-0000000000aa")
				tieHigh := uuid.MustParse("00000000-0000-0000-0000-0000000000bb")
				targetCol := &entity.Column{ID: columnID, BoardID: boardID, Position: 2, Title: "Col"}
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(targetCol, nil)
				d.checker.EXPECT().Check(mock.Anything, boardID, requesterID).Return(board, nil)
				d.columnRepo.EXPECT().GetColumnsByBoard(mock.Anything, boardID).Return([]*entity.Column{
					targetCol,          // pos 2, first in the unsorted input
					mkCol(tieHigh, 0),  // pos 0, ID sorts after tieLow
					mkCol(tieLow, 0),   // pos 0, ID sorts first
				}, nil)
				d.columnRepo.EXPECT().ReorderPositions(mock.Anything, assertOrder(columnID, tieLow, tieHigh)).Return(nil)
			},
			checkOut: func(t *testing.T, out *column.UpdateColumnPositionOutput) {
				assert.Equal(t, columnID, out.Column.ID)
				assert.Equal(t, 0, out.Column.Position)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := newDeps(t)
			tt.setupMocks(d)

			out, err := d.uc.UpdateColumnPosition(context.Background(), tt.input)

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
