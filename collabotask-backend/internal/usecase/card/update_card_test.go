package card_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/usecase/card"
)

func TestUpdateCard(t *testing.T) {
	boardID := uuid.New()
	columnID := uuid.New()
	cardID := uuid.New()
	requesterID := uuid.New()
	assigneeID := uuid.New()

	board := &entity.Board{ID: boardID}
	newColumn := func() *entity.Column { return &entity.Column{ID: columnID, BoardID: boardID} }

	// UpdateCard mutates the card returned by GetByID, so each case gets a fresh
	// one. Starts with a description (so "clear" is observable) and no assignee.
	newCard := func() *entity.Card {
		return &entity.Card{ID: cardID, ColumnID: columnID, Title: "Old Title", Description: ptr("existing")}
	}

	base := func() card.UpdateCardInput {
		return card.UpdateCardInput{BoardID: boardID, ColumnID: columnID, CardID: cardID, RequesterID: requesterID}
	}
	titleInput := base()
	titleInput.Title = ptr("New Title")

	assignInput := base()
	assignInput.AssignedTo = ptr(assigneeID)
	assignInput.AssignedToPresent = true

	tests := []struct {
		name       string
		input      card.UpdateCardInput
		setupMocks func(d cardTestDeps)
		wantErr    error
		wantErrMsg string
		checkOut   func(t *testing.T, out *card.UpdateCardOutput)
	}{
		{
			name:       "invalid input → validation error",
			input:      card.UpdateCardInput{BoardID: boardID, ColumnID: columnID, CardID: cardID, Title: ptr("New Title")}, // missing RequesterID
			setupMocks: func(d cardTestDeps) {},
			wantErrMsg: "validation",
		},
		{
			name:       "no fields provided → ErrAtLeastOneProvided",
			input:      base(),
			setupMocks: func(d cardTestDeps) {},
			wantErr:    domain.ErrAtLeastOneProvided,
		},
		{
			name: "AssignedTo is uuid.Nil → ErrInvalidAssigneeID",
			input: func() card.UpdateCardInput {
				in := base()
				in.AssignedTo = ptr(uuid.Nil)
				in.AssignedToPresent = true
				return in
			}(),
			setupMocks: func(d cardTestDeps) {},
			wantErr:    domain.ErrInvalidAssigneeID,
		},
		{
			name:  "card not found → ErrCardNotFound",
			input: titleInput,
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(nil, domain.ErrCardNotFound)
			},
			wantErr: domain.ErrCardNotFound,
		},
		{
			name:  "cardRepo.GetByID returns unexpected error → wrapped error",
			input: titleInput,
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to fetch card",
		},
		{
			name:  "card belongs to different column → ErrCardNotInColumn",
			input: titleInput,
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(
					&entity.Card{ID: cardID, ColumnID: uuid.New()}, nil,
				)
			},
			wantErr: domain.ErrCardNotInColumn,
		},
		{
			name:  "column not found → ErrColumnNotFound",
			input: titleInput,
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(newCard(), nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(nil, domain.ErrColumnNotFound)
			},
			wantErr: domain.ErrColumnNotFound,
		},
		{
			name:  "columnRepo.GetByID returns unexpected error → wrapped error",
			input: titleInput,
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(newCard(), nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to fetch column",
		},
		{
			name:  "column belongs to different board → ErrColumnNotInBoard",
			input: titleInput,
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(newCard(), nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(
					&entity.Column{ID: columnID, BoardID: uuid.New()}, nil,
				)
			},
			wantErr: domain.ErrColumnNotInBoard,
		},
		{
			name:  "boardAccessChecker.Check fails → error propagated",
			input: titleInput,
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(newCard(), nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(newColumn(), nil)
				d.checker.EXPECT().Check(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardAccessDenied)
			},
			wantErr: domain.ErrBoardAccessDenied,
		},
		{
			name:  "cardRepo.Update fails → error",
			input: titleInput,
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(newCard(), nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(newColumn(), nil)
				d.checker.EXPECT().Check(mock.Anything, boardID, requesterID).Return(board, nil)
				d.cardRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(c *entity.Card) bool {
					return c.Title == "New Title"
				})).Return(errors.New("db error"))
			},
			wantErrMsg: "db error",
		},
		{
			// Assignee is resolved BEFORE the write, so a resolution failure must
			// leave nothing committed — Update is intentionally not expected here.
			name:  "assignee resolution fails → wrapped error (card not updated)",
			input: assignInput,
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(newCard(), nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(newColumn(), nil)
				d.checker.EXPECT().Check(mock.Anything, boardID, requesterID).Return(board, nil)
				d.userRepo.EXPECT().GetById(mock.Anything, assigneeID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to fetch assignee",
		},
		{
			name:  "assignee resolution nil/empty → ErrUserNotFound (card not updated)",
			input: assignInput,
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(newCard(), nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(newColumn(), nil)
				d.checker.EXPECT().Check(mock.Anything, boardID, requesterID).Return(board, nil)
				d.userRepo.EXPECT().GetById(mock.Anything, assigneeID).Return(&entity.User{}, nil)
			},
			wantErr: domain.ErrUserNotFound,
		},
		{
			name: "success — description cleared (DescriptionPresent=true, Description=nil)",
			input: func() card.UpdateCardInput {
				in := base()
				in.DescriptionPresent = true
				in.Description = nil
				return in
			}(),
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(newCard(), nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(newColumn(), nil)
				d.checker.EXPECT().Check(mock.Anything, boardID, requesterID).Return(board, nil)
				d.cardRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(c *entity.Card) bool {
					return c.Description == nil
				})).Return(nil)
			},
			checkOut: func(t *testing.T, out *card.UpdateCardOutput) {
				assert.Nil(t, out.Card.Description)
				assert.Nil(t, out.Assignee)
			},
		},
		{
			name: "success — description cleared when whitespace-only string",
			input: func() card.UpdateCardInput {
				in := base()
				in.DescriptionPresent = true
				in.Description = ptr("   ")
				return in
			}(),
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(newCard(), nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(newColumn(), nil)
				d.checker.EXPECT().Check(mock.Anything, boardID, requesterID).Return(board, nil)
				d.cardRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(c *entity.Card) bool {
					return c.Description == nil
				})).Return(nil)
			},
			checkOut: func(t *testing.T, out *card.UpdateCardOutput) {
				assert.Nil(t, out.Card.Description)
			},
		},
		{
			name: "success — sets description to a non-empty (trimmed) value",
			input: func() card.UpdateCardInput {
				in := base()
				in.DescriptionPresent = true
				in.Description = ptr("  hello world  ")
				return in
			}(),
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(newCard(), nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(newColumn(), nil)
				d.checker.EXPECT().Check(mock.Anything, boardID, requesterID).Return(board, nil)
				d.cardRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(c *entity.Card) bool {
					return c.Description != nil && *c.Description == "hello world"
				})).Return(nil)
			},
			checkOut: func(t *testing.T, out *card.UpdateCardOutput) {
				require.NotNil(t, out.Card.Description)
				assert.Equal(t, "hello world", *out.Card.Description)
			},
		},
		{
			name: "success — sets due date (DueDatePresent)",
			input: func() card.UpdateCardInput {
				in := base()
				in.DueDatePresent = true
				in.DueDate = ptr(time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC))
				return in
			}(),
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(newCard(), nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(newColumn(), nil)
				d.checker.EXPECT().Check(mock.Anything, boardID, requesterID).Return(board, nil)
				d.cardRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(c *entity.Card) bool {
					return c.DueDate != nil && c.DueDate.Equal(time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC))
				})).Return(nil)
			},
			checkOut: func(t *testing.T, out *card.UpdateCardOutput) {
				require.NotNil(t, out.Card.DueDate)
				assert.True(t, out.Card.DueDate.Equal(time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)))
			},
		},
		{
			name: "success — clears assignee (AssignedToPresent=true, AssignedTo=nil) → output has no assignee",
			input: func() card.UpdateCardInput {
				in := base()
				in.AssignedToPresent = true
				in.AssignedTo = nil
				return in
			}(),
			setupMocks: func(d cardTestDeps) {
				// Card starts assigned; clearing must drop it and skip the assignee fetch.
				existing := uuid.New()
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(
					&entity.Card{ID: cardID, ColumnID: columnID, Title: "Old", AssignedTo: &existing}, nil,
				)
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(newColumn(), nil)
				d.checker.EXPECT().Check(mock.Anything, boardID, requesterID).Return(board, nil)
				d.cardRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(c *entity.Card) bool {
					return c.AssignedTo == nil
				})).Return(nil)
			},
			checkOut: func(t *testing.T, out *card.UpdateCardOutput) {
				assert.Nil(t, out.Card.AssignedTo)
				assert.Nil(t, out.Assignee)
			},
		},
		{
			name:  "success — with assignee",
			input: assignInput,
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(newCard(), nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(newColumn(), nil)
				d.checker.EXPECT().Check(mock.Anything, boardID, requesterID).Return(board, nil)
				d.cardRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(c *entity.Card) bool {
					return c.AssignedTo != nil && *c.AssignedTo == assigneeID
				})).Return(nil)
				d.userRepo.EXPECT().GetById(mock.Anything, assigneeID).Return(
					&entity.User{ID: assigneeID, Name: "Assignee"}, nil,
				)
			},
			checkOut: func(t *testing.T, out *card.UpdateCardOutput) {
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

			out, err := d.uc.UpdateCard(context.Background(), tt.input)

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
