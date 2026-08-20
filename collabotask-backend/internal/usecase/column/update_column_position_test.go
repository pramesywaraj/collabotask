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
	"collabotask/internal/usecase/common"
)

func TestUpdateColumnPosition(t *testing.T) {
	boardID := uuid.New()
	columnID := uuid.New()
	requesterID := uuid.New()

	board := &entity.Board{ID: boardID}
	existingColumn := &entity.Column{ID: columnID, BoardID: boardID, Position: 2000}

	validInput := column.UpdateColumnPositionInput{
		BoardID:     boardID,
		ColumnID:    columnID,
		Position:    1500,
		RequesterID: requesterID,
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
			name:       "invalid input (missing RequesterID) → validation error",
			input:      column.UpdateColumnPositionInput{BoardID: boardID, ColumnID: columnID, Position: 1500},
			setupMocks: func(d columnTestDeps) {},
			wantErrMsg: "validation",
		},
		{
			name:  "column not found → ErrColumnNotFound",
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
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardAccessDenied)
			},
			wantErr: domain.ErrBoardAccessDenied,
		},
		{
			name:  "columnRepo.UpdatePosition fails → error propagated",
			input: validInput,
			setupMocks: func(d columnTestDeps) {
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(existingColumn, nil)
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{Board: board}, nil)
				d.columnRepo.EXPECT().UpdatePosition(mock.Anything, columnID, float64(1500)).Return(float64(0), errors.New("db error"))
			},
			wantErrMsg: "db error",
		},
		{
			name:  "success — midpoint between two columns",
			input: validInput,
			setupMocks: func(d columnTestDeps) {
				// Fresh instance per test: the use case mutates Position on the returned
				// pointer, so sharing a single struct across parallel tests is a data race.
				col := &entity.Column{ID: columnID, BoardID: boardID, Position: 2000}
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(col, nil)
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{Board: board}, nil)
				d.columnRepo.EXPECT().UpdatePosition(mock.Anything, columnID, float64(1500)).Return(float64(1500), nil)
			},
			checkOut: func(t *testing.T, out *column.UpdateColumnPositionOutput) {
				assert.Equal(t, columnID, out.Column.ID)
				assert.Equal(t, float64(1500), out.Column.Position)
			},
		},
		{
			// Finding #1 regression: when a rebalance rewrites this column inside the
			// repo transaction, UpdatePosition returns a position different from the
			// requested one. The use case must surface the repo's returned value, not
			// input.Position — otherwise the response reports a position the DB no
			// longer holds.
			name:  "success — rebalance rewrote position (returns repo value, not input)",
			input: validInput, // requests 1500
			setupMocks: func(d columnTestDeps) {
				col := &entity.Column{ID: columnID, BoardID: boardID, Position: 2000}
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(col, nil)
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{Board: board}, nil)
				// Requested 1500, but a rebalance rewrote this column to 2000.
				d.columnRepo.EXPECT().UpdatePosition(mock.Anything, columnID, float64(1500)).Return(float64(2000), nil)
			},
			checkOut: func(t *testing.T, out *column.UpdateColumnPositionOutput) {
				assert.Equal(t, columnID, out.Column.ID)
				assert.Equal(t, float64(2000), out.Column.Position)
			},
		},
		{
			name: "success — head-insert (negative position is valid)",
			input: column.UpdateColumnPositionInput{
				BoardID:     boardID,
				ColumnID:    columnID,
				Position:    -1000,
				RequesterID: requesterID,
			},
			setupMocks: func(d columnTestDeps) {
				col := &entity.Column{ID: columnID, BoardID: boardID, Position: 2000}
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(col, nil)
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{Board: board}, nil)
				d.columnRepo.EXPECT().UpdatePosition(mock.Anything, columnID, float64(-1000)).Return(float64(-1000), nil)
			},
			checkOut: func(t *testing.T, out *column.UpdateColumnPositionOutput) {
				assert.Equal(t, columnID, out.Column.ID)
				assert.Equal(t, float64(-1000), out.Column.Position)
			},
		},
		{
			// Regression guard: position 0 must not be rejected (was incorrectly treated
			// as "not provided" when the field was int with validate:"required").
			name: "success — position 0 is valid (not rejected as zero value)",
			input: column.UpdateColumnPositionInput{
				BoardID:     boardID,
				ColumnID:    columnID,
				Position:    0,
				RequesterID: requesterID,
			},
			setupMocks: func(d columnTestDeps) {
				col := &entity.Column{ID: columnID, BoardID: boardID, Position: 2000}
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(col, nil)
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{Board: board}, nil)
				d.columnRepo.EXPECT().UpdatePosition(mock.Anything, columnID, float64(0)).Return(float64(0), nil)
			},
			checkOut: func(t *testing.T, out *column.UpdateColumnPositionOutput) {
				assert.Equal(t, columnID, out.Column.ID)
				assert.Equal(t, float64(0), out.Column.Position)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := newDeps(t)
			d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Maybe().Return(nil)
			d.broadcaster.EXPECT().Broadcast(mock.Anything, mock.Anything).Maybe()
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
