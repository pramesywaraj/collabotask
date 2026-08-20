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
	"github.com/stretchr/testify/assert"
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

	board := &entity.Board{ID: boardID}
	column := &entity.Column{ID: columnID, BoardID: boardID}

	input := card.CreateCardInput{
		BoardID:     boardID,
		ColumnID:    columnID,
		Title:       "Fix login bug",
		RequesterID: requesterID,
	}

	setupUntilCreate := func(d cardTestDeps) {
		d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(column, nil)
		d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).
			Return(&common.BoardAccess{Board: board}, nil)
		d.cardRepo.EXPECT().GetMaxPosition(mock.Anything, columnID).Return(float64(1000), nil)
		d.cardRepo.EXPECT().Create(mock.Anything, mock.Anything).
			Run(func(_ context.Context, c *entity.Card) { c.ID = cardID }).Return(nil)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Return(nil)
	}

	t.Run("happy path — CARD_CREATED broadcast with card + nil assignee", func(t *testing.T) {
		d := newDeps(t)
		setupUntilCreate(d)
		d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
			ev, ok := e.(common.CardCreated)
			return ok && ev.Card.ID == cardID && ev.Assignee == nil
		}))

		_, err := d.uc.CreateCard(context.Background(), input)
		require.NoError(t, err)
	})

	t.Run("happy path — CARD_CREATED broadcast with resolved assignee", func(t *testing.T) {
		assigneeID := uuid.New()
		d := newDeps(t)
		inputWithAssignee := card.CreateCardInput{
			BoardID:     boardID,
			ColumnID:    columnID,
			Title:       "Fix login bug",
			RequesterID: requesterID,
			AssignedTo:  ptr(assigneeID),
		}
		assignee := &entity.User{ID: assigneeID, Name: "Alice"}

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

		_, err := d.uc.CreateCard(context.Background(), inputWithAssignee)
		require.NoError(t, err)
	})

	t.Run("best-effort — Broadcast does not return and usecase succeeds regardless", func(t *testing.T) {
		d := newDeps(t)
		setupUntilCreate(d)
		// Broadcast has no return value — just assert it is called (port contract)
		d.broadcaster.EXPECT().Broadcast(boardID, mock.Anything)

		_, err := d.uc.CreateCard(context.Background(), input)
		require.NoError(t, err)
	})
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

	setupBase := func(d cardTestDeps) {
		d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(existingCard, nil)
		d.columnRepo.EXPECT().GetByID(mock.Anything, fromColID).Return(fromColumn, nil)
		d.columnRepo.EXPECT().GetByID(mock.Anything, toColID).Return(toColumn, nil)
		d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).
			Return(&common.BoardAccess{Board: board}, nil)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Maybe().Return(nil)
	}

	input := card.MoveCardInput{
		BoardID:      boardID,
		CardID:       cardID,
		FromColumnID: fromColID,
		ToColumnID:   toColID,
		ToPosition:   2000,
		RequesterID:  requesterID,
	}

	t.Run("happy path — CARD_MOVED broadcast when column changes", func(t *testing.T) {
		d := newDeps(t)
		setupBase(d)
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

		_, err := d.uc.MoveCard(context.Background(), input)
		require.NoError(t, err)
	})

	t.Run("no-op silence — same column and same position → Broadcast NOT called", func(t *testing.T) {
		d := newDeps(t)
		setupBase(d)
		// same column, same position as existingCard
		movedCard := &entity.Card{ID: cardID, ColumnID: fromColID, Position: 1000}
		d.cardRepo.EXPECT().Move(mock.Anything, cardID, fromColID, toColID, float64(1000)).Return(movedCard, nil)
		// broadcaster.EXPECT not set → unexpected call would fail test

		_, err := d.uc.MoveCard(context.Background(), card.MoveCardInput{
			BoardID:      boardID,
			CardID:       cardID,
			FromColumnID: fromColID,
			ToColumnID:   toColID,
			ToPosition:   1000,
			RequesterID:  requesterID,
		})
		require.NoError(t, err)
	})
}

// --- UpdateCard ---

func TestUpdateCardBroadcast(t *testing.T) {
	boardID := uuid.New()
	columnID := uuid.New()
	cardID := uuid.New()
	requesterID := uuid.New()

	board := &entity.Board{ID: boardID}
	column := &entity.Column{ID: columnID, BoardID: boardID}
	newTitle := "New Title"

	setupUntilUpdate := func(d cardTestDeps, c *entity.Card) {
		d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(c, nil)
		d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(column, nil)
		d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).
			Return(&common.BoardAccess{Board: board}, nil)
		d.cardRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Maybe().Return(nil)
	}

	t.Run("happy path — CARD_UPDATED with changed_fields=[title]", func(t *testing.T) {
		d := newDeps(t)
		setupUntilUpdate(d, &entity.Card{ID: cardID, ColumnID: columnID, Title: "Old Title"})

		d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
			ev, ok := e.(common.CardUpdated)
			return ok &&
				ev.Card.ID == cardID &&
				len(ev.ChangedFields) == 1 && ev.ChangedFields[0] == "title"
		}))

		_, err := d.uc.UpdateCard(context.Background(), card.UpdateCardInput{
			BoardID:     boardID,
			ColumnID:    columnID,
			CardID:      cardID,
			RequesterID: requesterID,
			Title:       &newTitle,
		})
		require.NoError(t, err)
	})

	t.Run("no-op silence — title unchanged → Broadcast NOT called", func(t *testing.T) {
		sameTitle := "Old Title"
		d := newDeps(t)
		setupUntilUpdate(d, &entity.Card{ID: cardID, ColumnID: columnID, Title: "Old Title"})
		// broadcaster.EXPECT not set

		_, err := d.uc.UpdateCard(context.Background(), card.UpdateCardInput{
			BoardID:     boardID,
			ColumnID:    columnID,
			CardID:      cardID,
			RequesterID: requesterID,
			Title:       &sameTitle,
		})
		require.NoError(t, err)
	})

	t.Run("assigned_to change carries the assignee in the event", func(t *testing.T) {
		assigneeID := uuid.New()
		assignee := &entity.User{ID: assigneeID, Name: "Bob"}
		d := newDeps(t)
		d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(
			&entity.Card{ID: cardID, ColumnID: columnID, Title: "Task"}, nil,
		)
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

		_, err := d.uc.UpdateCard(context.Background(), card.UpdateCardInput{
			BoardID:          boardID,
			ColumnID:         columnID,
			CardID:           cardID,
			RequesterID:      requesterID,
			AssignedTo:       &assigneeID,
			AssignedToPresent: true,
		})
		require.NoError(t, err)
	})
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

	setupUntilDelete := func(d cardTestDeps) {
		d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(existingCard, nil)
		d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(column, nil)
		d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).
			Return(&common.BoardAccess{Board: board}, nil)
		d.cardRepo.EXPECT().Delete(mock.Anything, cardID).Return(nil)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Return(nil)
	}

	t.Run("happy path — CARD_DELETED broadcast with card_id and column_id", func(t *testing.T) {
		d := newDeps(t)
		setupUntilDelete(d)

		d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
			ev, ok := e.(common.CardDeleted)
			return ok && ev.CardID == cardID && ev.ColumnID == columnID
		}))

		err := d.uc.DeleteCard(context.Background(), input)
		require.NoError(t, err)
	})

	t.Run("best-effort — delete still succeeds; Broadcast has no return value", func(t *testing.T) {
		d := newDeps(t)
		setupUntilDelete(d)
		d.broadcaster.EXPECT().Broadcast(boardID, mock.Anything)

		err := d.uc.DeleteCard(context.Background(), input)
		require.NoError(t, err)
	})

	t.Run("no card write → no broadcast (cardRepo.Delete fails)", func(t *testing.T) {
		d := newDeps(t)
		d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(existingCard, nil)
		d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(column, nil)
		d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).
			Return(&common.BoardAccess{Board: board}, nil)
		d.cardRepo.EXPECT().Delete(mock.Anything, cardID).Return(errors.New("db down"))
		// broadcaster.EXPECT NOT set → unexpected call would fail

		err := d.uc.DeleteCard(context.Background(), input)
		assert.Error(t, err)
	})
}
