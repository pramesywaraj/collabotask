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

func TestUpdateColumn(t *testing.T) {
	boardID := uuid.New()
	columnID := uuid.New()
	requesterID := uuid.New()

	// Each case builds its own column: UpdateColumn mutates the entity returned by
	// GetByID (column.Title = input.Title), so a shared pointer would race across
	// the parallel subtests that reach that mutation.
	newColumn := func() *entity.Column {
		return &entity.Column{ID: columnID, BoardID: boardID, Title: "Old Title"}
	}

	validInput := column.UpdateColumnInput{
		BoardID:     boardID,
		ColumnID:    columnID,
		Title:       "New Title",
		RequesterID: requesterID,
	}

	tests := []struct {
		name       string
		input      column.UpdateColumnInput
		setupMocks func(d columnTestDeps)
		wantErr    error
		wantErrMsg string
		checkOut   func(t *testing.T, out *column.UpdateColumnOutput)
	}{
		{
			name:       "invalid input → validation error",
			input:      column.UpdateColumnInput{BoardID: boardID, ColumnID: columnID, Title: "", RequesterID: requesterID},
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
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(newColumn(), nil)
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardAccessDenied)
			},
			wantErr: domain.ErrBoardAccessDenied,
		},
		{
			name:  "columnRepo.Update fails → error",
			input: validInput,
			setupMocks: func(d columnTestDeps) {
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(newColumn(), nil)
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{Board: &entity.Board{ID: boardID}}, nil)
				d.columnRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(c *entity.Column) bool {
					return c.Title == "New Title"
				})).Return(errors.New("db error"))
			},
			wantErrMsg: "db error",
		},
		{
			name:  "success",
			input: validInput,
			setupMocks: func(d columnTestDeps) {
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(newColumn(), nil)
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{Board: &entity.Board{ID: boardID}}, nil)
				d.columnRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(c *entity.Column) bool {
					return c.Title == "New Title"
				})).Return(nil)
			},
			checkOut: func(t *testing.T, out *column.UpdateColumnOutput) {
				assert.Equal(t, columnID, out.Column.ID)
				assert.Equal(t, "New Title", out.Column.Title)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := newDeps(t)
			d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Maybe().Return(nil)
			tt.setupMocks(d)

			out, err := d.uc.UpdateColumn(context.Background(), tt.input)

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
