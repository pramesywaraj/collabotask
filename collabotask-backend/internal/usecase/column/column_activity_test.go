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

// --- CreateColumn ---

func TestCreateColumnActivityLog(t *testing.T) {
	boardID := uuid.New()
	columnID := uuid.New()
	requesterID := uuid.New()
	board := &entity.Board{ID: boardID}

	input := column.CreateColumnInput{
		BoardID:     boardID,
		Title:       "Backlog",
		RequesterID: requesterID,
	}

	setupUntilCreate := func(d columnTestDeps) {
		d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).
			Return(&common.BoardAccess{Board: board}, nil)
		d.columnRepo.EXPECT().GetMaxPosition(mock.Anything, boardID).Return(float64(1000), nil)
		d.columnRepo.EXPECT().Create(mock.Anything, mock.Anything).
			Run(func(_ context.Context, c *entity.Column) { c.ID = columnID }).Return(nil)
	}

	t.Run("happy path — COLUMN/CREATED logged with board_id, requester, column_title", func(t *testing.T) {
		d := newDeps(t)
		setupUntilCreate(d)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.MatchedBy(func(a *entity.Activity) bool {
			colTitle, _ := a.Metadata[entity.ActivityMetaColumnTitle].(string)
			return a.BoardID == boardID &&
				a.ActionType == entity.ActivityActionCreated &&
				a.EntityType == entity.ActivityEntityColumn &&
				a.EntityID == columnID &&
				a.UserID != nil && *a.UserID == requesterID &&
				colTitle == "Backlog"
		})).Return(nil)

		out, err := d.uc.CreateColumn(context.Background(), input)
		require.NoError(t, err)
		assert.Equal(t, columnID, out.Column.ID)
	})

	t.Run("resilience — Log error is swallowed, creation still succeeds", func(t *testing.T) {
		d := newDeps(t)
		setupUntilCreate(d)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Return(errors.New("db down"))

		_, err := d.uc.CreateColumn(context.Background(), input)
		require.NoError(t, err)
	})
}

// --- DeleteColumn ---

func TestDeleteColumnActivityLog(t *testing.T) {
	boardID := uuid.New()
	columnID := uuid.New()
	requesterID := uuid.New()
	board := &entity.Board{ID: boardID}
	col := &entity.Column{ID: columnID, BoardID: boardID, Title: "Backlog"}

	input := column.DeleteColumnInput{
		BoardID:     boardID,
		ColumnID:    columnID,
		RequesterID: requesterID,
	}

	setupUntilDelete := func(d columnTestDeps) {
		d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(col, nil)
		d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).
			Return(&common.BoardAccess{Board: board}, nil)
		d.columnRepo.EXPECT().Delete(mock.Anything, columnID).Return(nil)
	}

	t.Run("happy path — COLUMN/DELETED logged with column_title snapshot", func(t *testing.T) {
		d := newDeps(t)
		setupUntilDelete(d)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.MatchedBy(func(a *entity.Activity) bool {
			colTitle, _ := a.Metadata[entity.ActivityMetaColumnTitle].(string)
			return a.BoardID == boardID &&
				a.ActionType == entity.ActivityActionDeleted &&
				a.EntityType == entity.ActivityEntityColumn &&
				a.EntityID == columnID &&
				a.UserID != nil && *a.UserID == requesterID &&
				colTitle == "Backlog"
		})).Return(nil)

		err := d.uc.DeleteColumn(context.Background(), input)
		require.NoError(t, err)
	})

	t.Run("resilience — Log error is swallowed, deletion still succeeds", func(t *testing.T) {
		d := newDeps(t)
		setupUntilDelete(d)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Return(errors.New("db down"))

		err := d.uc.DeleteColumn(context.Background(), input)
		require.NoError(t, err)
	})
}

// --- UpdateColumn ---

func TestUpdateColumnActivityLog(t *testing.T) {
	boardID := uuid.New()
	columnID := uuid.New()
	requesterID := uuid.New()
	board := &entity.Board{ID: boardID}

	setupUntilUpdate := func(d columnTestDeps, oldTitle, newTitle string) {
		col := &entity.Column{ID: columnID, BoardID: boardID, Title: oldTitle}
		d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(col, nil)
		d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).
			Return(&common.BoardAccess{Board: board}, nil)
		d.columnRepo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil)
	}

	t.Run("happy path — COLUMN/UPDATED logged when title changes", func(t *testing.T) {
		d := newDeps(t)
		setupUntilUpdate(d, "Old Title", "New Title")
		d.activityRepo.EXPECT().Log(mock.Anything, mock.MatchedBy(func(a *entity.Activity) bool {
			colTitle, _ := a.Metadata[entity.ActivityMetaColumnTitle].(string)
			return a.BoardID == boardID &&
				a.ActionType == entity.ActivityActionUpdated &&
				a.EntityType == entity.ActivityEntityColumn &&
				a.EntityID == columnID &&
				a.UserID != nil && *a.UserID == requesterID &&
				colTitle == "New Title"
		})).Return(nil)

		out, err := d.uc.UpdateColumn(context.Background(), column.UpdateColumnInput{
			BoardID:     boardID,
			ColumnID:    columnID,
			Title:       "New Title",
			RequesterID: requesterID,
		})
		require.NoError(t, err)
		assert.Equal(t, "New Title", out.Column.Title)
	})

	t.Run("no-op silence — same title → Log NOT called", func(t *testing.T) {
		d := newDeps(t)
		setupUntilUpdate(d, "Same Title", "Same Title")
		// No expectation on activityRepo.Log → unexpected call fails the test

		_, err := d.uc.UpdateColumn(context.Background(), column.UpdateColumnInput{
			BoardID:     boardID,
			ColumnID:    columnID,
			Title:       "Same Title",
			RequesterID: requesterID,
		})
		require.NoError(t, err)
	})

	t.Run("resilience — Log error is swallowed, update still succeeds", func(t *testing.T) {
		d := newDeps(t)
		setupUntilUpdate(d, "Old Title", "New Title")
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Return(errors.New("db down"))

		_, err := d.uc.UpdateColumn(context.Background(), column.UpdateColumnInput{
			BoardID:     boardID,
			ColumnID:    columnID,
			Title:       "New Title",
			RequesterID: requesterID,
		})
		require.NoError(t, err)
	})
}

// --- UpdateColumnPosition ---

func TestUpdateColumnPositionActivityLog(t *testing.T) {
	boardID := uuid.New()
	columnID := uuid.New()
	requesterID := uuid.New()
	board := &entity.Board{ID: boardID}
	newCol := func() *entity.Column {
		return &entity.Column{ID: columnID, BoardID: boardID, Title: "Backlog", Position: 1000}
	}

	setupBase := func(d columnTestDeps) {
		d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(newCol(), nil)
		d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).
			Return(&common.BoardAccess{Board: board}, nil)
	}

	t.Run("happy path — COLUMN/MOVED logged when position changes", func(t *testing.T) {
		d := newDeps(t)
		setupBase(d)
		d.columnRepo.EXPECT().UpdatePosition(mock.Anything, columnID, float64(2000)).Return(float64(2000), nil)

		d.activityRepo.EXPECT().Log(mock.Anything, mock.MatchedBy(func(a *entity.Activity) bool {
			colTitle, _ := a.Metadata[entity.ActivityMetaColumnTitle].(string)
			return a.BoardID == boardID &&
				a.ActionType == entity.ActivityActionMoved &&
				a.EntityType == entity.ActivityEntityColumn &&
				a.EntityID == columnID &&
				a.UserID != nil && *a.UserID == requesterID &&
				colTitle == "Backlog"
		})).Return(nil)

		out, err := d.uc.UpdateColumnPosition(context.Background(), column.UpdateColumnPositionInput{
			BoardID:     boardID,
			ColumnID:    columnID,
			Position:    2000,
			RequesterID: requesterID,
		})
		require.NoError(t, err)
		assert.Equal(t, float64(2000), out.Column.Position)
	})

	t.Run("no-op silence — returned position equals old position → Log NOT called", func(t *testing.T) {
		d := newDeps(t)
		setupBase(d)
		// UpdatePosition returns the SAME value as the column's current position.
		d.columnRepo.EXPECT().UpdatePosition(mock.Anything, columnID, float64(1000)).Return(float64(1000), nil)
		// activityRepo.Log must NOT be called

		_, err := d.uc.UpdateColumnPosition(context.Background(), column.UpdateColumnPositionInput{
			BoardID:     boardID,
			ColumnID:    columnID,
			Position:    1000,
			RequesterID: requesterID,
		})
		require.NoError(t, err)
	})

	t.Run("resilience — Log error is swallowed, move still succeeds", func(t *testing.T) {
		d := newDeps(t)
		setupBase(d)
		d.columnRepo.EXPECT().UpdatePosition(mock.Anything, columnID, float64(2000)).Return(float64(2000), nil)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Return(errors.New("db down"))

		_, err := d.uc.UpdateColumnPosition(context.Background(), column.UpdateColumnPositionInput{
			BoardID:     boardID,
			ColumnID:    columnID,
			Position:    2000,
			RequesterID: requesterID,
		})
		require.NoError(t, err)
	})

	t.Run("UpdatePosition repo error does not call Log", func(t *testing.T) {
		d := newDeps(t)
		setupBase(d)
		d.columnRepo.EXPECT().UpdatePosition(mock.Anything, columnID, float64(2000)).Return(float64(0), domain.ErrColumnNotFound)
		// No Log expectation

		_, err := d.uc.UpdateColumnPosition(context.Background(), column.UpdateColumnPositionInput{
			BoardID:     boardID,
			ColumnID:    columnID,
			Position:    2000,
			RequesterID: requesterID,
		})
		require.Error(t, err)
	})
}
