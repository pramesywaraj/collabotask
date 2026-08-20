package card_test

// Broadcast contract tests for the card use case (Part C / ADR-009 / SRS §5.2).
// Three contract points per mutation: (1) happy path — Broadcast called with the
// right event, (2) no-op silence — guards that prevent the write also prevent the
// broadcast, (3) best-effort — Broadcast returns nothing and the usecase doesn't
// depend on it succeeding.

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"collabotask/internal/domain/entity"
	"collabotask/internal/usecase/card"
	"collabotask/internal/usecase/common"
)

// --- CreateCard ---

func TestCreateCardBroadcast(t *testing.T) {
	boardID := uuid.New()
	columnID := uuid.New()
	cardID := uuid.New()
	requesterID := uuid.New()
	assigneeID := uuid.New()

	board := &entity.Board{ID: boardID}
	column := &entity.Column{ID: columnID, BoardID: boardID}
	assignee := &entity.User{ID: assigneeID, Name: "Alice"}

	baseInput := card.CreateCardInput{
		BoardID:     boardID,
		ColumnID:    columnID,
		Title:       "Fix login bug",
		RequesterID: requesterID,
	}

	// wireCreate wires the create path with no assignee resolution.
	wireCreate := func(d cardTestDeps) {
		d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(column, nil)
		d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).
			Return(&common.BoardAccess{Board: board}, nil)
		d.cardRepo.EXPECT().GetMaxPosition(mock.Anything, columnID).Return(float64(1000), nil)
		d.cardRepo.EXPECT().Create(mock.Anything, mock.Anything).
			Run(func(_ context.Context, c *entity.Card) { c.ID = cardID }).Return(nil)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Return(nil)
	}

	tests := []struct {
		name       string
		setupMocks func(d cardTestDeps)
		input      card.CreateCardInput
	}{
		{
			name: "happy path — CARD_CREATED broadcast with card + nil assignee",
			setupMocks: func(d cardTestDeps) {
				wireCreate(d)
				d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
					ev, ok := e.(common.CardCreated)
					return ok && ev.Card.ID == cardID && ev.Assignee == nil
				}))
			},
			input: baseInput,
		},
		{
			name: "happy path — CARD_CREATED broadcast with resolved assignee",
			setupMocks: func(d cardTestDeps) {
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(column, nil)
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).
					Return(&common.BoardAccess{Board: board}, nil)
				d.boardMemberRepo.EXPECT().IsUserExists(mock.Anything, boardID, assigneeID).Return(true, nil)
				d.userRepo.EXPECT().GetById(mock.Anything, assigneeID).Return(assignee, nil)
				d.cardRepo.EXPECT().GetMaxPosition(mock.Anything, columnID).Return(float64(1000), nil)
				d.cardRepo.EXPECT().Create(mock.Anything, mock.Anything).
					Run(func(_ context.Context, c *entity.Card) { c.ID = cardID }).Return(nil)
				d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Return(nil)
				d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
					ev, ok := e.(common.CardCreated)
					return ok && ev.Card.ID == cardID && ev.Assignee != nil && ev.Assignee.ID == assigneeID
				}))
			},
			input: card.CreateCardInput{
				BoardID:     boardID,
				ColumnID:    columnID,
				Title:       "Fix login bug",
				RequesterID: requesterID,
				AssignedTo:  ptr(assigneeID),
			},
		},
		{
			name: "best-effort — Broadcast does not return and usecase succeeds regardless",
			setupMocks: func(d cardTestDeps) {
				wireCreate(d)
				d.broadcaster.EXPECT().Broadcast(boardID, mock.Anything)
			},
			input: baseInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newDeps(t)
			tt.setupMocks(d)

			_, err := d.uc.CreateCard(context.Background(), tt.input)
			require.NoError(t, err)
		})
	}
}

// --- MoveCard ---

func TestMoveCardBroadcast(t *testing.T) {
	boardID := uuid.New()
	fromColID := uuid.New()
	toColID := uuid.New()
	cardID := uuid.New()
	requesterID := uuid.New()

	board := &entity.Board{ID: boardID}
	existingCard := &entity.Card{ID: cardID, ColumnID: fromColID, Title: "Task", Position: 1000}
	fromColumn := &entity.Column{ID: fromColID, BoardID: boardID, Title: "Todo"}
	toColumn := &entity.Column{ID: toColID, BoardID: boardID, Title: "Done"}

	wireBase := func(d cardTestDeps) {
		d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(existingCard, nil)
		d.columnRepo.EXPECT().GetByID(mock.Anything, fromColID).Return(fromColumn, nil)
		d.columnRepo.EXPECT().GetByID(mock.Anything, toColID).Return(toColumn, nil)
		d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).
			Return(&common.BoardAccess{Board: board}, nil)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Maybe().Return(nil)
	}

	tests := []struct {
		name       string
		setupMocks func(d cardTestDeps)
		toPosition float64
	}{
		{
			name: "happy path — CARD_MOVED broadcast when column changes",
			setupMocks: func(d cardTestDeps) {
				wireBase(d)
				movedCard := &entity.Card{ID: cardID, ColumnID: toColID, Position: 2000}
				d.cardRepo.EXPECT().Move(mock.Anything, cardID, fromColID, toColID, float64(2000)).Return(movedCard, nil)
				d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
					ev, ok := e.(common.CardMoved)
					return ok &&
						ev.CardID == cardID &&
						ev.FromColumnID == fromColID &&
						ev.ToColumnID == toColID &&
						ev.Position == 2000
				}))
			},
			toPosition: 2000,
		},
		{
			name: "no-op silence — same column and same position → Broadcast NOT called",
			setupMocks: func(d cardTestDeps) {
				wireBase(d)
				movedCard := &entity.Card{ID: cardID, ColumnID: fromColID, Position: 1000}
				d.cardRepo.EXPECT().Move(mock.Anything, cardID, fromColID, toColID, float64(1000)).Return(movedCard, nil)
				// broadcaster.EXPECT not set → unexpected call would fail test
			},
			toPosition: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newDeps(t)
			tt.setupMocks(d)

			_, err := d.uc.MoveCard(context.Background(), card.MoveCardInput{
				BoardID:      boardID,
				CardID:       cardID,
				FromColumnID: fromColID,
				ToColumnID:   toColID,
				ToPosition:   tt.toPosition,
				RequesterID:  requesterID,
			})
			require.NoError(t, err)
		})
	}
}

// --- UpdateCard ---

func TestUpdateCardBroadcast(t *testing.T) {
	boardID := uuid.New()
	columnID := uuid.New()
	cardID := uuid.New()
	requesterID := uuid.New()
	assigneeID := uuid.New()

	board := &entity.Board{ID: boardID}
	column := &entity.Column{ID: columnID, BoardID: boardID}
	assignee := &entity.User{ID: assigneeID, Name: "Bob"}

	tests := []struct {
		name       string
		setupMocks func(d cardTestDeps)
		input      card.UpdateCardInput
	}{
		{
			name: "happy path — CARD_UPDATED with changed_fields=[title]",
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).
					Return(&entity.Card{ID: cardID, ColumnID: columnID, Title: "Old Title"}, nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(column, nil)
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).
					Return(&common.BoardAccess{Board: board}, nil)
				d.cardRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
				d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Maybe().Return(nil)
				d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
					ev, ok := e.(common.CardUpdated)
					return ok &&
						ev.Card.ID == cardID &&
						len(ev.ChangedFields) == 1 && ev.ChangedFields[0] == "title"
				}))
			},
			input: card.UpdateCardInput{
				BoardID:     boardID,
				ColumnID:    columnID,
				CardID:      cardID,
				RequesterID: requesterID,
				Title:       ptr("New Title"),
			},
		},
		{
			name: "no-op silence — title unchanged → Broadcast NOT called",
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).
					Return(&entity.Card{ID: cardID, ColumnID: columnID, Title: "Old Title"}, nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(column, nil)
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).
					Return(&common.BoardAccess{Board: board}, nil)
				d.cardRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
				d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Maybe().Return(nil)
				// broadcaster.EXPECT not set
			},
			input: card.UpdateCardInput{
				BoardID:     boardID,
				ColumnID:    columnID,
				CardID:      cardID,
				RequesterID: requesterID,
				Title:       ptr("Old Title"),
			},
		},
		{
			name: "assigned_to change carries the assignee in the event",
			setupMocks: func(d cardTestDeps) {
				d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).
					Return(&entity.Card{ID: cardID, ColumnID: columnID, Title: "Task"}, nil)
				d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(column, nil)
				d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).
					Return(&common.BoardAccess{Board: board}, nil)
				d.boardMemberRepo.EXPECT().IsUserExists(mock.Anything, boardID, assigneeID).Return(true, nil)
				d.userRepo.EXPECT().GetById(mock.Anything, assigneeID).Return(assignee, nil)
				d.cardRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
				d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Return(nil)
				d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
					ev, ok := e.(common.CardUpdated)
					return ok &&
						ev.Card.ID == cardID &&
						ev.Assignee != nil && ev.Assignee.ID == assigneeID
				}))
			},
			input: card.UpdateCardInput{
				BoardID:           boardID,
				ColumnID:          columnID,
				CardID:            cardID,
				RequesterID:       requesterID,
				AssignedTo:        &assigneeID,
				AssignedToPresent: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newDeps(t)
			tt.setupMocks(d)

			_, err := d.uc.UpdateCard(context.Background(), tt.input)
			require.NoError(t, err)
		})
	}
}

// --- DeleteCard ---

func TestDeleteCardBroadcast(t *testing.T) {
	boardID := uuid.New()
	columnID := uuid.New()
	cardID := uuid.New()
	requesterID := uuid.New()

	board := &entity.Board{ID: boardID}
	existingCard := &entity.Card{ID: cardID, ColumnID: columnID, Title: "Fix bug"}
	column := &entity.Column{ID: columnID, BoardID: boardID}

	input := card.DeleteCardInput{
		BoardID:     boardID,
		ColumnID:    columnID,
		CardID:      cardID,
		RequesterID: requesterID,
	}

	wireRead := func(d cardTestDeps) {
		d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(existingCard, nil)
		d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(column, nil)
		d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).
			Return(&common.BoardAccess{Board: board}, nil)
	}

	tests := []struct {
		name       string
		setupMocks func(d cardTestDeps)
		wantErr    bool
	}{
		{
			name: "happy path — CARD_DELETED broadcast with card_id and column_id",
			setupMocks: func(d cardTestDeps) {
				wireRead(d)
				d.cardRepo.EXPECT().Delete(mock.Anything, cardID).Return(nil)
				d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Return(nil)
				d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
					ev, ok := e.(common.CardDeleted)
					return ok && ev.CardID == cardID && ev.ColumnID == columnID
				}))
			},
		},
		{
			name: "best-effort — delete still succeeds; Broadcast has no return value",
			setupMocks: func(d cardTestDeps) {
				wireRead(d)
				d.cardRepo.EXPECT().Delete(mock.Anything, cardID).Return(nil)
				d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Return(nil)
				d.broadcaster.EXPECT().Broadcast(boardID, mock.Anything)
			},
		},
		{
			name: "no card write → no broadcast (cardRepo.Delete fails)",
			setupMocks: func(d cardTestDeps) {
				wireRead(d)
				d.cardRepo.EXPECT().Delete(mock.Anything, cardID).Return(errors.New("db down"))
				// broadcaster.EXPECT NOT set → unexpected call would fail
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := newDeps(t)
			tt.setupMocks(d)

			err := d.uc.DeleteCard(context.Background(), input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
