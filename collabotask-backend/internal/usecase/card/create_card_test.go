package card_test

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
	"collabotask/internal/usecase/card"
)

func TestCreateCard(t *testing.T) {
	boardID := uuid.New()
	columnID := uuid.New()
	requesterID := uuid.New()
	assigneeID := uuid.New()

	board := &entity.Board{ID: boardID}
	newColumn := func() *entity.Column { return &entity.Column{ID: columnID, BoardID: boardID} }

	validInput := card.CreateCardInput{
		BoardID:     boardID,
		ColumnID:    columnID,
		Title:       "My Card",
		RequesterID: requesterID,
	}
	assigneeInput := card.CreateCardInput{
		BoardID:     boardID,
		ColumnID:    columnID,
		Title:       "My Card",
		RequesterID: requesterID,
		AssignedTo:  ptr(assigneeID),
	}

	tests := []struct {
		name       string
		input      card.CreateCardInput
		setupMocks func(d cardTestDeps)
		wantErr    error
		wantErrMsg string
		checkOut   func(t *testing.T, out *card.CreateCardOutput)
	}{
		{
			name:       "invalid input → validation error",
			input:      card.CreateCardInput{BoardID: boardID, ColumnID: columnID, Title: "", RequesterID: requesterID},
			setupMocks: func(d cardTestDeps) {},
			wantErrMsg: "validation",
		},
		{
			name:  "column not found → ErrColumnNotFound",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(nil, domain.ErrColumnNotFound)
			},
			wantErr: domain.ErrColumnNotFound,
		},
		{
			name:  "columnRepo.GetByID returns unexpected error → wrapped error",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to fetch column",
		},
		{
			name:  "column belongs to different board → ErrColumnNotInBoard",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(
					&entity.Column{ID: columnID, BoardID: uuid.New()}, nil,
				)
			},
			wantErr: domain.ErrColumnNotInBoard,
		},
		{
			name:  "boardAccessChecker.Check fails → error propagated",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(newColumn(), nil)
				d.checker.EXPECT().Check(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardAccessDenied)
			},
			wantErr: domain.ErrBoardAccessDenied,
		},
		{
			name:  "AssignedTo is uuid.Nil → ErrInvalidAssigneeID",
			input: card.CreateCardInput{BoardID: boardID, ColumnID: columnID, Title: "My Card", RequesterID: requesterID, AssignedTo: ptr(uuid.Nil)},
			setupMocks: func(d cardTestDeps) {
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(newColumn(), nil)
				d.checker.EXPECT().Check(mock.Anything, boardID, requesterID).Return(board, nil)
			},
			wantErr: domain.ErrInvalidAssigneeID,
		},
		{
			name:  "assignee GetById returns error → wrapped error",
			input: assigneeInput,
			setupMocks: func(d cardTestDeps) {
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(newColumn(), nil)
				d.checker.EXPECT().Check(mock.Anything, boardID, requesterID).Return(board, nil)
				d.userRepo.EXPECT().GetById(mock.Anything, assigneeID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to fetch assignee",
		},
		{
			name:  "assignee nil/empty → ErrUserNotFound",
			input: assigneeInput,
			setupMocks: func(d cardTestDeps) {
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(newColumn(), nil)
				d.checker.EXPECT().Check(mock.Anything, boardID, requesterID).Return(board, nil)
				d.userRepo.EXPECT().GetById(mock.Anything, assigneeID).Return(nil, nil)
			},
			wantErr: domain.ErrUserNotFound,
		},
		{
			name:  "cardRepo.GetMaxPosition fails → wrapped error",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(newColumn(), nil)
				d.checker.EXPECT().Check(mock.Anything, boardID, requesterID).Return(board, nil)
				d.cardRepo.EXPECT().GetMaxPosition(mock.Anything, columnID).Return(-1, errors.New("db error"))
			},
			wantErrMsg: "failed to get cards max position",
		},
		{
			name:  "cardRepo.Create fails → wrapped error",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(newColumn(), nil)
				d.checker.EXPECT().Check(mock.Anything, boardID, requesterID).Return(board, nil)
				d.cardRepo.EXPECT().GetMaxPosition(mock.Anything, columnID).Return(3, nil)
				d.cardRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			wantErrMsg: "failed to create card",
		},
		{
			name:  "success — without assignee (position = maxPos + 1)",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(newColumn(), nil)
				d.checker.EXPECT().Check(mock.Anything, boardID, requesterID).Return(board, nil)
				d.cardRepo.EXPECT().GetMaxPosition(mock.Anything, columnID).Return(4, nil)
				d.cardRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(c *entity.Card) bool {
					return c.ColumnID == columnID && c.Title == "My Card" && c.Position == 5 &&
						c.AssignedTo == nil && c.CreatedBy == requesterID
				})).Return(nil)
			},
			checkOut: func(t *testing.T, out *card.CreateCardOutput) {
				assert.Equal(t, columnID, out.Card.ColumnID)
				assert.Equal(t, 5, out.Card.Position)
				assert.Nil(t, out.Card.AssignedTo)
				assert.Nil(t, out.Assignee)
			},
		},
		{
			name:  "success — with assignee",
			input: assigneeInput,
			setupMocks: func(d cardTestDeps) {
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(newColumn(), nil)
				d.checker.EXPECT().Check(mock.Anything, boardID, requesterID).Return(board, nil)
				d.userRepo.EXPECT().GetById(mock.Anything, assigneeID).Return(
					&entity.User{ID: assigneeID, Name: "Assignee"}, nil,
				)
				d.cardRepo.EXPECT().GetMaxPosition(mock.Anything, columnID).Return(0, nil)
				d.cardRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(c *entity.Card) bool {
					return c.Position == 1 && c.AssignedTo != nil && *c.AssignedTo == assigneeID
				})).Return(nil)
			},
			checkOut: func(t *testing.T, out *card.CreateCardOutput) {
				assert.Equal(t, 1, out.Card.Position)
				require.NotNil(t, out.Assignee)
				assert.Equal(t, assigneeID, out.Assignee.ID)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := newDeps(t)
			tt.setupMocks(d)

			out, err := d.uc.CreateCard(context.Background(), tt.input)

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
