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

func TestCreateColumn(t *testing.T) {
	boardID := uuid.New()
	requesterID := uuid.New()
	board := &entity.Board{ID: boardID}

	validInput := column.CreateColumnInput{
		BoardID:     boardID,
		Title:       "My Column",
		RequesterID: requesterID,
	}

	tests := []struct {
		name       string
		input      column.CreateColumnInput
		setupMocks func(d columnTestDeps)
		wantErr    error
		wantErrMsg string
		checkOut   func(t *testing.T, out *column.CreateColumnOutput)
	}{
		{
			name:       "invalid input → validation error",
			input:      column.CreateColumnInput{BoardID: boardID, Title: "", RequesterID: requesterID},
			setupMocks: func(d columnTestDeps) {},
			wantErrMsg: "validation",
		},
		{
			name:  "boardAccessChecker.Check fails → error propagated",
			input: validInput,
			setupMocks: func(d columnTestDeps) {
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardNotFound)
			},
			wantErr: domain.ErrBoardNotFound,
		},
		{
			name:  "columnRepo.GetMaxPosition fails → wrapped error",
			input: validInput,
			setupMocks: func(d columnTestDeps) {
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{Board: board}, nil)
				d.columnRepo.EXPECT().GetMaxPosition(mock.Anything, boardID).Return(0, errors.New("db error"))
			},
			wantErrMsg: "failed to get max position",
		},
		{
			name:  "columnRepo.Create fails → error",
			input: validInput,
			setupMocks: func(d columnTestDeps) {
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{Board: board}, nil)
				d.columnRepo.EXPECT().GetMaxPosition(mock.Anything, boardID).Return(float64(2000), nil)
				d.columnRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(c *entity.Column) bool {
					return c.Position == 3000
				})).Return(errors.New("db error"))
			},
			wantErrMsg: "db error",
		},
		{
			name:  "success — first column in empty board (maxPos=0 → position=1000)",
			input: validInput,
			setupMocks: func(d columnTestDeps) {
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{Board: board}, nil)
				d.columnRepo.EXPECT().GetMaxPosition(mock.Anything, boardID).Return(float64(0), nil)
				d.columnRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(c *entity.Column) bool {
					return c.BoardID == boardID && c.Title == "My Column" && c.Position == 1000
				})).Return(nil)
			},
			checkOut: func(t *testing.T, out *column.CreateColumnOutput) {
				assert.Equal(t, boardID, out.Column.BoardID)
				assert.Equal(t, "My Column", out.Column.Title)
				assert.Equal(t, float64(1000), out.Column.Position)
			},
		},
		{
			name:  "success — appends after existing column (maxPos=4000 → position=5000)",
			input: validInput,
			setupMocks: func(d columnTestDeps) {
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).Return(&common.BoardAccess{Board: board}, nil)
				d.columnRepo.EXPECT().GetMaxPosition(mock.Anything, boardID).Return(float64(4000), nil)
				d.columnRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(c *entity.Column) bool {
					return c.BoardID == boardID && c.Title == "My Column" && c.Position == 5000
				})).Return(nil)
			},
			checkOut: func(t *testing.T, out *column.CreateColumnOutput) {
				assert.Equal(t, boardID, out.Column.BoardID)
				assert.Equal(t, float64(5000), out.Column.Position)
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

			out, err := d.uc.CreateColumn(context.Background(), tt.input)

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
