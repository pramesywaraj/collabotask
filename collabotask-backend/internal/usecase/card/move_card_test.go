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

func TestMoveCard(t *testing.T) {
	boardID := uuid.New()
	fromColumnID := uuid.New()
	toColumnID := uuid.New()
	cardID := uuid.New()
	requesterID := uuid.New()
	assigneeID := uuid.New()

	board := &entity.Board{ID: boardID}
	existingCard := &entity.Card{ID: cardID, ColumnID: fromColumnID}
	fromColumn := &entity.Column{ID: fromColumnID, BoardID: boardID}
	toColumn := &entity.Column{ID: toColumnID, BoardID: boardID}

	validInput := card.MoveCardInput{
		BoardID:      boardID,
		CardID:       cardID,
		FromColumnID: fromColumnID,
		ToColumnID:   toColumnID,
		ToPosition:   1,
		RequesterID:  requesterID,
	}
	withPosition := func(pos int) card.MoveCardInput {
		in := validInput
		in.ToPosition = pos
		return in
	}

	// happy-path mocks up to (but not including) the Move call.
	setupUntilMove := func(d cardTestDeps) {
		d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(existingCard, nil)
		d.columnRepo.EXPECT().GetByID(mock.Anything, fromColumnID).Return(fromColumn, nil)
		d.columnRepo.EXPECT().GetByID(mock.Anything, toColumnID).Return(toColumn, nil)
		d.checker.EXPECT().Check(mock.Anything, boardID, requesterID).Return(board, nil)
	}

	tests := []struct {
		name       string
		input      card.MoveCardInput
		setupMocks func(d cardTestDeps)
		wantErr    error
		wantErrMsg string
		checkOut   func(t *testing.T, out *card.MoveCardOutput)
	}{
		{
			name:       "invalid input → validation error",
			input:      card.MoveCardInput{BoardID: boardID, CardID: cardID, FromColumnID: fromColumnID, ToColumnID: toColumnID, ToPosition: 1}, // missing RequesterID
			setupMocks: func(d cardTestDeps) {},
			wantErrMsg: "validation",
		},
		{
			name:  "card not found → ErrCardNotFound",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(nil, domain.ErrCardNotFound)
			},
			wantErr: domain.ErrCardNotFound,
		},
		{
			name:  "cardRepo.GetByID returns unexpected error → wrapped error",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to fetch card",
		},
		{
			name:  "card not in fromColumn → ErrCardNotInColumn",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(
					&entity.Card{ID: cardID, ColumnID: uuid.New()}, nil,
				)
			},
			wantErr: domain.ErrCardNotInColumn,
		},
		{
			name:  "fromColumn not found → ErrColumnNotFound",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(existingCard, nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, fromColumnID).Return(nil, domain.ErrColumnNotFound)
			},
			wantErr: domain.ErrColumnNotFound,
		},
		{
			name:  "fromColumn GetByID returns unexpected error → wrapped error",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(existingCard, nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, fromColumnID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to fetch 'from' column",
		},
		{
			name:  "fromColumn belongs to different board → ErrColumnNotInBoard",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(existingCard, nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, fromColumnID).Return(
					&entity.Column{ID: fromColumnID, BoardID: uuid.New()}, nil,
				)
			},
			wantErr: domain.ErrColumnNotInBoard,
		},
		{
			name:  "toColumn not found → ErrColumnNotFound",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(existingCard, nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, fromColumnID).Return(fromColumn, nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, toColumnID).Return(nil, domain.ErrColumnNotFound)
			},
			wantErr: domain.ErrColumnNotFound,
		},
		{
			name:  "toColumn GetByID returns unexpected error → wrapped error",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(existingCard, nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, fromColumnID).Return(fromColumn, nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, toColumnID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to fetch 'to' column",
		},
		{
			name:  "fromColumn.BoardID != toColumn.BoardID → ErrInconsistentState",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(existingCard, nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, fromColumnID).Return(fromColumn, nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, toColumnID).Return(
					&entity.Column{ID: toColumnID, BoardID: uuid.New()}, nil,
				)
			},
			wantErr: domain.ErrInconsistentState,
		},
		{
			name:  "boardAccessChecker.Check fails → error propagated",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(existingCard, nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, fromColumnID).Return(fromColumn, nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, toColumnID).Return(toColumn, nil)
				d.checker.EXPECT().Check(mock.Anything, boardID, requesterID).Return(nil, domain.ErrBoardAccessDenied)
			},
			wantErr: domain.ErrBoardAccessDenied,
		},
		{
			name:  "cardRepo.GetMaxPosition fails → error",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				setupUntilMove(d)
				d.cardRepo.EXPECT().GetMaxPosition(mock.Anything, toColumnID).Return(-1, errors.New("db error"))
			},
			wantErrMsg: "db error",
		},
		{
			name:  "cardRepo.Move fails → error",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				setupUntilMove(d)
				d.cardRepo.EXPECT().GetMaxPosition(mock.Anything, toColumnID).Return(5, nil)
				d.cardRepo.EXPECT().Move(mock.Anything, cardID, fromColumnID, toColumnID, 1).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "db error",
		},
		{
			name:  "assignee fetch fails after move → wrapped error",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				setupUntilMove(d)
				d.cardRepo.EXPECT().GetMaxPosition(mock.Anything, toColumnID).Return(5, nil)
				d.cardRepo.EXPECT().Move(mock.Anything, cardID, fromColumnID, toColumnID, 1).Return(
					&entity.Card{ID: cardID, ColumnID: toColumnID, AssignedTo: ptr(assigneeID)}, nil,
				)
				d.userRepo.EXPECT().GetById(mock.Anything, assigneeID).Return(nil, errors.New("db error"))
			},
			wantErrMsg: "failed to fetch assignee",
		},
		{
			name:  "assignee nil/empty after move → ErrUserNotFound",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				setupUntilMove(d)
				d.cardRepo.EXPECT().GetMaxPosition(mock.Anything, toColumnID).Return(5, nil)
				d.cardRepo.EXPECT().Move(mock.Anything, cardID, fromColumnID, toColumnID, 1).Return(
					&entity.Card{ID: cardID, ColumnID: toColumnID, AssignedTo: ptr(assigneeID)}, nil,
				)
				d.userRepo.EXPECT().GetById(mock.Anything, assigneeID).Return(&entity.User{}, nil)
			},
			wantErr: domain.ErrUserNotFound,
		},
		{
			name:  "success — ToPosition = 0 moves to front (regression guard: 0 must not be rejected)",
			input: withPosition(0),
			setupMocks: func(d cardTestDeps) {
				setupUntilMove(d)
				d.cardRepo.EXPECT().GetMaxPosition(mock.Anything, toColumnID).Return(5, nil)
				d.cardRepo.EXPECT().Move(mock.Anything, cardID, fromColumnID, toColumnID, 0).Return(
					&entity.Card{ID: cardID, ColumnID: toColumnID, Position: 0}, nil,
				)
			},
			checkOut: func(t *testing.T, out *card.MoveCardOutput) {
				assert.Equal(t, 0, out.Card.Position)
				assert.Nil(t, out.Assignee)
			},
		},
		{
			name:  "success — ToPosition within range",
			input: withPosition(3),
			setupMocks: func(d cardTestDeps) {
				setupUntilMove(d)
				d.cardRepo.EXPECT().GetMaxPosition(mock.Anything, toColumnID).Return(5, nil)
				d.cardRepo.EXPECT().Move(mock.Anything, cardID, fromColumnID, toColumnID, 3).Return(
					&entity.Card{ID: cardID, ColumnID: toColumnID, Position: 3}, nil,
				)
			},
			checkOut: func(t *testing.T, out *card.MoveCardOutput) {
				assert.Equal(t, 3, out.Card.Position)
			},
		},
		{
			name:  "success — ToPosition clamped to max+1",
			input: withPosition(100),
			setupMocks: func(d cardTestDeps) {
				setupUntilMove(d)
				d.cardRepo.EXPECT().GetMaxPosition(mock.Anything, toColumnID).Return(5, nil)
				d.cardRepo.EXPECT().Move(mock.Anything, cardID, fromColumnID, toColumnID, 6).Return(
					&entity.Card{ID: cardID, ColumnID: toColumnID, Position: 6}, nil,
				)
			},
			checkOut: func(t *testing.T, out *card.MoveCardOutput) {
				assert.Equal(t, 6, out.Card.Position)
			},
		},
		{
			name:  "success — with assignee",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				setupUntilMove(d)
				d.cardRepo.EXPECT().GetMaxPosition(mock.Anything, toColumnID).Return(5, nil)
				d.cardRepo.EXPECT().Move(mock.Anything, cardID, fromColumnID, toColumnID, 1).Return(
					&entity.Card{ID: cardID, ColumnID: toColumnID, Position: 1, AssignedTo: ptr(assigneeID)}, nil,
				)
				d.userRepo.EXPECT().GetById(mock.Anything, assigneeID).Return(
					&entity.User{ID: assigneeID, Name: "Assignee"}, nil,
				)
			},
			checkOut: func(t *testing.T, out *card.MoveCardOutput) {
				assert.Equal(t, 1, out.Card.Position)
				require.NotNil(t, out.Assignee)
				assert.Equal(t, assigneeID, out.Assignee.ID)
			},
		},
		{
			name:  "success — card without assignee",
			input: validInput,
			setupMocks: func(d cardTestDeps) {
				setupUntilMove(d)
				d.cardRepo.EXPECT().GetMaxPosition(mock.Anything, toColumnID).Return(5, nil)
				d.cardRepo.EXPECT().Move(mock.Anything, cardID, fromColumnID, toColumnID, 1).Return(
					&entity.Card{ID: cardID, ColumnID: toColumnID, Position: 1, AssignedTo: nil}, nil,
				)
			},
			checkOut: func(t *testing.T, out *card.MoveCardOutput) {
				assert.Nil(t, out.Card.AssignedTo)
				assert.Nil(t, out.Assignee)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := newDeps(t)
			tt.setupMocks(d)

			out, err := d.uc.MoveCard(context.Background(), tt.input)

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
