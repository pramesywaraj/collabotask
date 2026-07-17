package column_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/usecase/column"
	"collabotask/internal/usecase/common"
)

func TestDeleteColumn(t *testing.T) {
	boardID := uuid.New()
	columnID := uuid.New()
	requesterID := uuid.New()

	existingColumn := &entity.Column{ID: columnID, BoardID: boardID, Title: "My Column"}

	validInput := column.DeleteColumnInput{
		BoardID:     boardID,
		ColumnID:    columnID,
		RequesterID: requesterID,
	}

	tests := []struct {
		name       string
		input      column.DeleteColumnInput
		setupMocks func(d columnTestDeps)
		wantErr    error
		wantErrMsg string
	}{
		{
			name:       "invalid input → validation error",
			input:      column.DeleteColumnInput{BoardID: boardID, ColumnID: columnID},
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
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardAccessDenied)
			},
			wantErr: domain.ErrBoardAccessDenied,
		},
		{
			name:  "columnRepo.Delete fails → error",
			input: validInput,
			setupMocks: func(d columnTestDeps) {
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(existingColumn, nil)
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{Board: &entity.Board{ID: boardID}}, nil)
				d.columnRepo.EXPECT().Delete(mock.Anything, columnID).Return(errors.New("db error"))
			},
			wantErrMsg: "db error",
		},
		{
			name:  "success",
			input: validInput,
			setupMocks: func(d columnTestDeps) {
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(existingColumn, nil)
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{Board: &entity.Board{ID: boardID}}, nil)
				d.columnRepo.EXPECT().Delete(mock.Anything, columnID).Return(nil)
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

			err := d.uc.DeleteColumn(context.Background(), tt.input)

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
