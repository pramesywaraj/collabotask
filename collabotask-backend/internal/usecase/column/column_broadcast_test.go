package column_test

// Broadcast contract tests for the column use case (Part C / ADR-009 / SRS §5.2).
// Three contract points per mutation: (1) happy path — Broadcast called with the
// right event type and board_id, (2) no-op silence — guards that prevent the write
// also prevent the broadcast, (3) best-effort — Broadcast has no return value.

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/usecase/column"
	"collabotask/internal/usecase/common"
)

// --- CreateColumn ---

func TestCreateColumnBroadcast(t *testing.T) {
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
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Return(nil)
	}

	t.Run("happy path — COLUMN_CREATED broadcast with column entity", func(t *testing.T) {
		d := newDeps(t)
		setupUntilCreate(d)
		d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
			ev, ok := e.(common.ColumnCreated)
			return ok && ev.Column.ID == columnID
		}))

		_, err := d.uc.CreateColumn(context.Background(), input)
		require.NoError(t, err)
	})

	t.Run("best-effort — Broadcast has no return value; usecase succeeds", func(t *testing.T) {
		d := newDeps(t)
		setupUntilCreate(d)
		d.broadcaster.EXPECT().Broadcast(boardID, mock.Anything)

		_, err := d.uc.CreateColumn(context.Background(), input)
		require.NoError(t, err)
	})
}

// --- UpdateColumn ---

func TestUpdateColumnBroadcast(t *testing.T) {
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
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Maybe().Return(nil)
	}

	t.Run("happy path — COLUMN_UPDATED broadcast with column_id and new title", func(t *testing.T) {
		d := newDeps(t)
		setupUntilUpdate(d, "Old Title", "New Title")
		d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
			ev, ok := e.(common.ColumnUpdated)
			return ok && ev.ColumnID == columnID && ev.Title == "New Title"
		}))

		_, err := d.uc.UpdateColumn(context.Background(), column.UpdateColumnInput{
			BoardID:     boardID,
			ColumnID:    columnID,
			Title:       "New Title",
			RequesterID: requesterID,
		})
		require.NoError(t, err)
	})

	t.Run("no-op silence — same title → Broadcast NOT called", func(t *testing.T) {
		d := newDeps(t)
		setupUntilUpdate(d, "Same Title", "Same Title")
		// broadcaster.EXPECT not set — unexpected call would fail

		_, err := d.uc.UpdateColumn(context.Background(), column.UpdateColumnInput{
			BoardID:     boardID,
			ColumnID:    columnID,
			Title:       "Same Title",
			RequesterID: requesterID,
		})
		require.NoError(t, err)
	})
}

// --- DeleteColumn ---

func TestDeleteColumnBroadcast(t *testing.T) {
	boardID := uuid.New()
	columnID := uuid.New()
	requesterID := uuid.New()
	board := &entity.Board{ID: boardID}
	col := &entity.Column{ID: columnID, BoardID: boardID, Title: "Done"}

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
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Return(nil)
	}

	t.Run("happy path — COLUMN_DELETED broadcast with column_id", func(t *testing.T) {
		d := newDeps(t)
		setupUntilDelete(d)
		d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
			ev, ok := e.(common.ColumnDeleted)
			return ok && ev.ColumnID == columnID
		}))

		err := d.uc.DeleteColumn(context.Background(), input)
		require.NoError(t, err)
	})

	t.Run("no write → no broadcast (Delete fails)", func(t *testing.T) {
		d := newDeps(t)
		d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(col, nil)
		d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).
			Return(&common.BoardAccess{Board: board}, nil)
		d.columnRepo.EXPECT().Delete(mock.Anything, columnID).Return(domain.ErrColumnNotFound)
		// broadcaster.EXPECT NOT set

		err := d.uc.DeleteColumn(context.Background(), input)
		require.Error(t, err)
	})
}

// --- UpdateColumnPosition ---

func TestUpdateColumnPositionBroadcast(t *testing.T) {
	boardID := uuid.New()
	columnID := uuid.New()
	requesterID := uuid.New()
	board := &entity.Board{ID: boardID}
	newCol := func() *entity.Column { return &entity.Column{ID: columnID, BoardID: boardID, Position: 1000} }

	setupBase := func(d columnTestDeps) {
		d.columnRepo.EXPECT().GetByID(mock.Anything, columnID).Return(newCol(), nil)
		d.checker.EXPECT().CheckMutateAccess(mock.Anything, boardID, requesterID).
			Return(&common.BoardAccess{Board: board}, nil)
		d.activityRepo.EXPECT().Log(mock.Anything, mock.Anything).Maybe().Return(nil)
	}

	t.Run("happy path — COLUMN_MOVED broadcast with column_id and new position", func(t *testing.T) {
		d := newDeps(t)
		setupBase(d)
		d.columnRepo.EXPECT().UpdatePosition(mock.Anything, columnID, float64(2000)).Return(float64(2000), nil)
		d.broadcaster.EXPECT().Broadcast(boardID, mock.MatchedBy(func(e common.Event) bool {
			ev, ok := e.(common.ColumnMoved)
			return ok && ev.ColumnID == columnID && ev.Position == 2000
		}))

		_, err := d.uc.UpdateColumnPosition(context.Background(), column.UpdateColumnPositionInput{
			BoardID:     boardID,
			ColumnID:    columnID,
			Position:    2000,
			RequesterID: requesterID,
		})
		require.NoError(t, err)
	})

	t.Run("no-op silence — returned position equals old position → Broadcast NOT called", func(t *testing.T) {
		d := newDeps(t)
		setupBase(d)
		d.columnRepo.EXPECT().UpdatePosition(mock.Anything, columnID, float64(1000)).Return(float64(1000), nil)
		// broadcaster.EXPECT not set

		_, err := d.uc.UpdateColumnPosition(context.Background(), column.UpdateColumnPositionInput{
			BoardID:     boardID,
			ColumnID:    columnID,
			Position:    1000,
			RequesterID: requesterID,
		})
		require.NoError(t, err)
	})
}
