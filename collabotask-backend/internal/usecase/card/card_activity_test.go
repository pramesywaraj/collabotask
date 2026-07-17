package card_test

// Activity log contract tests for the card use case (ADR-007 / SRS §4.6).
// Three contract points per mutation: (1) happy path — Log called with correct args,
// (2) resilience — Log error does not fail the mutation, (3) no-op silence — no-op
// mutations do NOT call Log.

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

func TestCreateCardActivityLog(t *testing.T) {
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
	}

	t.Run("happy path — CARD/CREATED logged with board_id, requester, card_title", func(t *testing.T) {
		d := newDeps(t)
		setupUntilCreate(d)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.MatchedBy(func(a *entity.Activity) bool {
			cardTitle, _ := a.Metadata[entity.ActivityMetaCardTitle].(string)
			return a.BoardID == boardID &&
				a.ActionType == entity.ActivityActionCreated &&
				a.EntityType == entity.ActivityEntityCard &&
				a.EntityID == cardID &&
				a.UserID != nil && *a.UserID == requesterID &&
				cardTitle == "Fix login bug"
		})).Return(nil)

		out, err := d.uc.CreateCard(context.Background(), input)
		require.NoError(t, err)
		assert.Equal(t, cardID, out.Card.ID)
	})

	t.Run("resilience — Log error is swallowed, mutation still succeeds", func(t *testing.T) {
		d := newDeps(t)
		setupUntilCreate(d)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Return(errors.New("db down"))

		_, err := d.uc.CreateCard(context.Background(), input)
		require.NoError(t, err)
	})
}

// --- DeleteCard ---

func TestDeleteCardActivityLog(t *testing.T) {
	boardID := uuid.New()
	columnID := uuid.New()
	cardID := uuid.New()
	requesterID := uuid.New()

	board := &entity.Board{ID: boardID}
	existingCard := &entity.Card{ID: cardID, ColumnID: columnID, Title: "Fix login bug"}
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
	}

	t.Run("happy path — CARD/DELETED logged with card_title snapshot", func(t *testing.T) {
		d := newDeps(t)
		setupUntilDelete(d)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.MatchedBy(func(a *entity.Activity) bool {
			cardTitle, _ := a.Metadata[entity.ActivityMetaCardTitle].(string)
			return a.BoardID == boardID &&
				a.ActionType == entity.ActivityActionDeleted &&
				a.EntityType == entity.ActivityEntityCard &&
				a.EntityID == cardID &&
				a.UserID != nil && *a.UserID == requesterID &&
				cardTitle == "Fix login bug"
		})).Return(nil)

		err := d.uc.DeleteCard(context.Background(), input)
		require.NoError(t, err)
	})

	t.Run("resilience — Log error is swallowed, deletion still succeeds", func(t *testing.T) {
		d := newDeps(t)
		setupUntilDelete(d)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Return(errors.New("db down"))

		err := d.uc.DeleteCard(context.Background(), input)
		require.NoError(t, err)
	})
}

// --- MoveCard ---

func TestMoveCardActivityLog(t *testing.T) {
	boardID := uuid.New()
	fromColumnID := uuid.New()
	toColumnID := uuid.New()
	cardID := uuid.New()
	requesterID := uuid.New()

	board := &entity.Board{ID: boardID}
	existingCard := &entity.Card{ID: cardID, ColumnID: fromColumnID, Title: "Fix login bug", Position: 1000}
	fromColumn := &entity.Column{ID: fromColumnID, BoardID: boardID, Title: "To Do"}
	toColumn := &entity.Column{ID: toColumnID, BoardID: boardID, Title: "Done"}

	setupBase := func(d cardTestDeps) {
		d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(existingCard, nil)
		d.columnRepo.EXPECT().GetByID(mock.Anything, fromColumnID).Return(fromColumn, nil)
		d.columnRepo.EXPECT().GetByID(mock.Anything, toColumnID).Return(toColumn, nil)
		d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).
			Return(&common.BoardAccess{Board: board}, nil)
	}

	t.Run("happy path — CARD/MOVED logged when column changes", func(t *testing.T) {
		d := newDeps(t)
		setupBase(d)
		movedCard := &entity.Card{ID: cardID, ColumnID: toColumnID, Title: "Fix login bug", Position: 2000}
		d.cardRepo.EXPECT().Move(mock.Anything, cardID, fromColumnID, toColumnID, float64(2000)).Return(movedCard, nil)

		d.activityRepo.EXPECT().Log(mock.Anything, mock.MatchedBy(func(a *entity.Activity) bool {
			fromColID, _ := a.Metadata[entity.ActivityMetaFromColumnID].(string)
			fromColTitle, _ := a.Metadata[entity.ActivityMetaFromColumnTitle].(string)
			toColID, _ := a.Metadata[entity.ActivityMetaToColumnID].(string)
			toColTitle, _ := a.Metadata[entity.ActivityMetaToColumnTitle].(string)
			cardTitle, _ := a.Metadata[entity.ActivityMetaCardTitle].(string)
			return a.BoardID == boardID &&
				a.ActionType == entity.ActivityActionMoved &&
				a.EntityType == entity.ActivityEntityCard &&
				a.EntityID == cardID &&
				a.UserID != nil && *a.UserID == requesterID &&
				cardTitle == "Fix login bug" &&
				fromColID == fromColumnID.String() &&
				fromColTitle == "To Do" &&
				toColID == toColumnID.String() &&
				toColTitle == "Done"
		})).Return(nil)

		out, err := d.uc.MoveCard(context.Background(), card.MoveCardInput{
			BoardID:      boardID,
			CardID:       cardID,
			FromColumnID: fromColumnID,
			ToColumnID:   toColumnID,
			ToPosition:   2000,
			RequesterID:  requesterID,
		})
		require.NoError(t, err)
		assert.Equal(t, toColumnID, out.Card.ColumnID)
	})

	t.Run("happy path — CARD/MOVED logged when position changes in same column", func(t *testing.T) {
		d := newDeps(t)
		d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(existingCard, nil)
		d.columnRepo.EXPECT().GetByID(mock.Anything, fromColumnID).Return(fromColumn, nil)
		d.columnRepo.EXPECT().GetByID(mock.Anything, fromColumnID).Return(fromColumn, nil) // toColumnID = fromColumnID
		d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).
			Return(&common.BoardAccess{Board: board}, nil)

		sameColumnSameBoard := &entity.Column{ID: fromColumnID, BoardID: boardID, Title: "To Do"}
		_ = sameColumnSameBoard

		movedCard := &entity.Card{ID: cardID, ColumnID: fromColumnID, Title: "Fix login bug", Position: 1500}
		d.cardRepo.EXPECT().Move(mock.Anything, cardID, fromColumnID, fromColumnID, float64(1500)).Return(movedCard, nil)

		d.activityRepo.EXPECT().Log(mock.Anything, mock.MatchedBy(func(a *entity.Activity) bool {
			return a.ActionType == entity.ActivityActionMoved &&
				a.EntityType == entity.ActivityEntityCard
		})).Return(nil)

		_, err := d.uc.MoveCard(context.Background(), card.MoveCardInput{
			BoardID:      boardID,
			CardID:       cardID,
			FromColumnID: fromColumnID,
			ToColumnID:   fromColumnID,
			ToPosition:   1500,
			RequesterID:  requesterID,
		})
		require.NoError(t, err)
	})

	t.Run("no-op silence — same column AND same position → Log NOT called", func(t *testing.T) {
		d := newDeps(t)
		setupBase(d)
		// movedCard has same column and same position as existingCard
		movedCard := &entity.Card{ID: cardID, ColumnID: fromColumnID, Title: "Fix login bug", Position: 1000}
		d.cardRepo.EXPECT().Move(mock.Anything, cardID, fromColumnID, toColumnID, float64(1000)).Return(movedCard, nil)
		// activityRepo.Log must NOT be called — any unexpected call will fail the test

		_, err := d.uc.MoveCard(context.Background(), card.MoveCardInput{
			BoardID:      boardID,
			CardID:       cardID,
			FromColumnID: fromColumnID,
			ToColumnID:   toColumnID,
			ToPosition:   1000,
			RequesterID:  requesterID,
		})
		require.NoError(t, err)
	})

	t.Run("resilience — Log error is swallowed, move still succeeds", func(t *testing.T) {
		d := newDeps(t)
		setupBase(d)
		movedCard := &entity.Card{ID: cardID, ColumnID: toColumnID, Title: "Fix login bug", Position: 2000}
		d.cardRepo.EXPECT().Move(mock.Anything, cardID, fromColumnID, toColumnID, float64(2000)).Return(movedCard, nil)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Return(errors.New("db down"))

		_, err := d.uc.MoveCard(context.Background(), card.MoveCardInput{
			BoardID:      boardID,
			CardID:       cardID,
			FromColumnID: fromColumnID,
			ToColumnID:   toColumnID,
			ToPosition:   2000,
			RequesterID:  requesterID,
		})
		require.NoError(t, err)
	})
}

// --- UpdateCard ---

func TestUpdateCardActivityLog(t *testing.T) {
	boardID := uuid.New()
	columnID := uuid.New()
	cardID := uuid.New()
	requesterID := uuid.New()

	board := &entity.Board{ID: boardID}
	column := &entity.Column{ID: columnID, BoardID: boardID}
	newCard := func() *entity.Card { return &entity.Card{ID: cardID, ColumnID: columnID, Title: "Old Title"} }
	newTitle := "New Title"

	setupUntilUpdate := func(d cardTestDeps, c *entity.Card) {
		d.cardRepo.EXPECT().GetByID(mock.Anything, cardID).Return(c, nil)
		d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(column, nil)
		d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).
			Return(&common.BoardAccess{Board: board}, nil)
		d.cardRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
	}

	t.Run("happy path — CARD/UPDATED with changed title in changed_fields", func(t *testing.T) {
		d := newDeps(t)
		setupUntilUpdate(d, newCard())

		d.activityRepo.EXPECT().Log(mock.Anything, mock.MatchedBy(func(a *entity.Activity) bool {
			changedFields, _ := a.Metadata[entity.ActivityMetaChangedFields].([]string)
			cardTitle, _ := a.Metadata[entity.ActivityMetaCardTitle].(string)
			return a.BoardID == boardID &&
				a.ActionType == entity.ActivityActionUpdated &&
				a.EntityType == entity.ActivityEntityCard &&
				a.EntityID == cardID &&
				a.UserID != nil && *a.UserID == requesterID &&
				cardTitle == "New Title" &&
				len(changedFields) == 1 && changedFields[0] == "title"
		})).Return(nil)

		_, err := d.uc.UpdateCard(context.Background(), card.UpdateCardInput{
			BoardID:     boardID,
			ColumnID:    columnID,
			CardID:      cardID,
			RequesterID: requesterID,
			Title:       &newTitle,
		})
		require.NoError(t, err)
	})

	t.Run("no-op silence — title unchanged → Log NOT called", func(t *testing.T) {
		sameTitle := "Old Title"
		d := newDeps(t)
		setupUntilUpdate(d, newCard())
		// No expectation on activityRepo.Log → unexpected call would fail test

		_, err := d.uc.UpdateCard(context.Background(), card.UpdateCardInput{
			BoardID:     boardID,
			ColumnID:    columnID,
			CardID:      cardID,
			RequesterID: requesterID,
			Title:       &sameTitle,
		})
		require.NoError(t, err)
	})

	t.Run("resilience — Log error is swallowed, update still succeeds", func(t *testing.T) {
		d := newDeps(t)
		setupUntilUpdate(d, newCard())
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Return(errors.New("db down"))

		_, err := d.uc.UpdateCard(context.Background(), card.UpdateCardInput{
			BoardID:     boardID,
			ColumnID:    columnID,
			CardID:      cardID,
			RequesterID: requesterID,
			Title:       &newTitle,
		})
		require.NoError(t, err)
	})
}
