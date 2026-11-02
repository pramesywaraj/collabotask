package integration_test

// Integration tests for the fractional-positioning rebalance (ADR-004, migration
// 000006_fractional_positioning). rebalanceIfNeeded is a repository-internal helper, so
// it is exercised through the two public seams that call it — cardRepo.Move (partitioned
// by column_id) and columnRepo.UpdatePosition (partitioned by board_id).
//
// The contract: a move inspects its two neighbour gaps; if the smaller one drops below
// domain.PositionRebalanceThreshold the whole partition is rewritten to evenly spaced
// domain.PositionStep multiples, preserving the order the move produced, and the moved
// item's *actual* new position is returned — never the position that was requested.

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"collabotask/internal/domain"
	pgRepos "collabotask/internal/repository/postgres"
	"collabotask/tests/integration/testutil"
)

// cardOrder returns the ids and positions of a column's cards in sort order.
func cardOrder(t *testing.T, columnID uuid.UUID) ([]uuid.UUID, []float64) {
	t.Helper()
	cards, err := pgRepos.NewCardRepository(testPool).GetCardsByColumn(context.Background(), columnID)
	require.NoError(t, err)
	ids := make([]uuid.UUID, 0, len(cards))
	positions := make([]float64, 0, len(cards))
	for _, c := range cards {
		ids = append(ids, c.ID)
		positions = append(positions, c.Position)
	}
	return ids, positions
}

// columnOrder returns the ids and positions of a board's columns in sort order.
func columnOrder(t *testing.T, boardID uuid.UUID) ([]uuid.UUID, []float64) {
	t.Helper()
	cols, err := pgRepos.NewColumnRepository(testPool).GetColumnsByBoard(context.Background(), boardID)
	require.NoError(t, err)
	ids := make([]uuid.UUID, 0, len(cols))
	positions := make([]float64, 0, len(cols))
	for _, c := range cols {
		ids = append(ids, c.ID)
		positions = append(positions, c.Position)
	}
	return ids, positions
}

// TestCardMove_RebalancesWhenGapTooSmall drops a card 0.2 above its new predecessor —
// under the 0.5 threshold — and expects the column rewritten to clean STEP multiples in
// the order the move produced, with Move reporting the position the DB actually holds.
func TestCardMove_RebalancesWhenGapTooSmall(t *testing.T) {
	t.Cleanup(func() { testutil.TruncateAll(t, testPool) })

	user := fx.CreateTestUser(t)
	ws := fx.CreateTestWorkspace(t, user.ID)
	board := fx.CreateTestBoard(t, ws.ID, user.ID)
	col := fx.CreateTestColumn(t, board.ID)

	first := fx.CreateTestCardAt(t, col.ID, board.ID, user.ID, "first", 1000)
	second := fx.CreateTestCardAt(t, col.ID, board.ID, user.ID, "second", 2000)
	moved := fx.CreateTestCardAt(t, col.ID, board.ID, user.ID, "moved", 3000)

	cardRepo := pgRepos.NewCardRepository(testPool)

	// 1000.2 sits 0.2 after `first` and 999.8 before `second` → min gap 0.2 < 0.5.
	got, err := cardRepo.Move(context.Background(), moved.ID, col.ID, col.ID, 1000.2)
	require.NoError(t, err)

	// Order after the move: first, moved, second → positions 1 × STEP, 2 × STEP, 3 × STEP.
	assert.Equal(t, 2*domain.PositionStep, got.Position,
		"Move must report the rebalanced position, not the requested 1000.2")

	stored, err := cardRepo.GetByID(context.Background(), moved.ID)
	require.NoError(t, err)
	assert.Equal(t, got.Position, stored.Position, "the reported position must be the one the DB holds")

	ids, positions := cardOrder(t, col.ID)
	assert.Equal(t, []uuid.UUID{first.ID, moved.ID, second.ID}, ids, "the move order must survive the rewrite")
	assert.Equal(t, []float64{domain.PositionStep, 2 * domain.PositionStep, 3 * domain.PositionStep}, positions,
		"the whole partition must be rewritten to evenly spaced STEP multiples")
}

// TestCardMove_NoRebalanceOnComfortableGap is the other side of the threshold: a move
// that lands in the middle of a wide gap keeps the fractional position it asked for and
// leaves its neighbours alone — rebalancing is the rare path, not the default one.
func TestCardMove_NoRebalanceOnComfortableGap(t *testing.T) {
	t.Cleanup(func() { testutil.TruncateAll(t, testPool) })

	user := fx.CreateTestUser(t)
	ws := fx.CreateTestWorkspace(t, user.ID)
	board := fx.CreateTestBoard(t, ws.ID, user.ID)
	col := fx.CreateTestColumn(t, board.ID)

	first := fx.CreateTestCardAt(t, col.ID, board.ID, user.ID, "first", 1000)
	second := fx.CreateTestCardAt(t, col.ID, board.ID, user.ID, "second", 2000)
	moved := fx.CreateTestCardAt(t, col.ID, board.ID, user.ID, "moved", 3000)

	cardRepo := pgRepos.NewCardRepository(testPool)

	// 1500 sits 500 from either neighbour → both gaps clear the 0.5 threshold.
	got, err := cardRepo.Move(context.Background(), moved.ID, col.ID, col.ID, 1500)
	require.NoError(t, err)
	assert.Equal(t, 1500.0, got.Position, "a comfortable gap must leave the requested position intact")

	ids, positions := cardOrder(t, col.ID)
	assert.Equal(t, []uuid.UUID{first.ID, moved.ID, second.ID}, ids)
	assert.Equal(t, []float64{1000, 1500, 2000}, positions, "no rewrite: the neighbours keep their positions")
}

// TestCardMove_NoRebalanceAtOpenEnd covers the open end of a partition. With no next
// neighbour the gap is +Inf, so min(gapBefore, +Inf) must be judged on the real gap
// alone — an open end is not a packed one.
func TestCardMove_NoRebalanceAtOpenEnd(t *testing.T) {
	t.Cleanup(func() { testutil.TruncateAll(t, testPool) })

	user := fx.CreateTestUser(t)
	ws := fx.CreateTestWorkspace(t, user.ID)
	board := fx.CreateTestBoard(t, ws.ID, user.ID)
	col := fx.CreateTestColumn(t, board.ID)

	first := fx.CreateTestCardAt(t, col.ID, board.ID, user.ID, "first", 1000)
	moved := fx.CreateTestCardAt(t, col.ID, board.ID, user.ID, "moved", 2000)

	cardRepo := pgRepos.NewCardRepository(testPool)

	// Append past the end: no next neighbour (+Inf), prev gap 4000.
	got, err := cardRepo.Move(context.Background(), moved.ID, col.ID, col.ID, 5000)
	require.NoError(t, err)
	assert.Equal(t, 5000.0, got.Position, "an open end must not trigger a rebalance")

	ids, positions := cardOrder(t, col.ID)
	assert.Equal(t, []uuid.UUID{first.ID, moved.ID}, ids)
	assert.Equal(t, []float64{1000, 5000}, positions)

	// The other open end: first position in the partition, no previous neighbour (-Inf).
	got, err = cardRepo.Move(context.Background(), moved.ID, col.ID, col.ID, 10)
	require.NoError(t, err)
	assert.Equal(t, 10.0, got.Position, "the leading open end must not trigger a rebalance either")

	ids, positions = cardOrder(t, col.ID)
	assert.Equal(t, []uuid.UUID{moved.ID, first.ID}, ids)
	assert.Equal(t, []float64{10, 1000}, positions)
}

// TestColumnUpdatePosition_RebalancesWhenGapTooSmall exercises the second partition the
// same helper serves — columns, partitioned by board_id. A fixture board arrives with the
// three default columns one STEP apart; dragging the last one 0.2 past the first packs the
// gap and must rewrite the board's columns to clean STEP multiples.
func TestColumnUpdatePosition_RebalancesWhenGapTooSmall(t *testing.T) {
	t.Cleanup(func() { testutil.TruncateAll(t, testPool) })

	user := fx.CreateTestUser(t)
	ws := fx.CreateTestWorkspace(t, user.ID)
	board := fx.CreateTestBoard(t, ws.ID, user.ID)

	columnRepo := pgRepos.NewColumnRepository(testPool)

	cols, err := columnRepo.GetColumnsByBoard(context.Background(), board.ID)
	require.NoError(t, err)
	require.Len(t, cols, len(domain.DefaultNewBoardColumnTitles), "a fixture board starts with its default columns")
	toDo, inProgress, done := cols[0], cols[1], cols[2]

	// STEP+0.2 sits 0.2 after `toDo` (STEP) and nearly a full STEP before `inProgress`
	// (2*STEP) → min gap 0.2 < 0.5.
	packed := domain.PositionStep + 0.2
	newPos, err := columnRepo.UpdatePosition(context.Background(), done.ID, packed)
	require.NoError(t, err)
	assert.Equal(t, 2*domain.PositionStep, newPos,
		"UpdatePosition must report the rebalanced position, not the requested one")

	stored, err := columnRepo.GetByID(context.Background(), done.ID)
	require.NoError(t, err)
	assert.Equal(t, newPos, stored.Position, "the reported position must be the one the DB holds")

	ids, positions := columnOrder(t, board.ID)
	assert.Equal(t, []uuid.UUID{toDo.ID, done.ID, inProgress.ID}, ids, "the drag order must survive the rewrite")
	assert.Equal(t, []float64{domain.PositionStep, 2 * domain.PositionStep, 3 * domain.PositionStep}, positions,
		"the whole board partition must be rewritten to evenly spaced STEP multiples")
}

// TestColumnUpdatePosition_NoRebalanceOnComfortableGap is the column mirror of the card
// comfortable-gap case. The three default columns sit a full STEP apart, so the midpoint
// between two of them is a comfortable landing spot — dropping `Done` there keeps the
// requested position and leaves every other column exactly where it was.
func TestColumnUpdatePosition_NoRebalanceOnComfortableGap(t *testing.T) {
	t.Cleanup(func() { testutil.TruncateAll(t, testPool) })

	user := fx.CreateTestUser(t)
	ws := fx.CreateTestWorkspace(t, user.ID)
	board := fx.CreateTestBoard(t, ws.ID, user.ID)

	columnRepo := pgRepos.NewColumnRepository(testPool)

	cols, err := columnRepo.GetColumnsByBoard(context.Background(), board.ID)
	require.NoError(t, err)
	require.Len(t, cols, len(domain.DefaultNewBoardColumnTitles), "a fixture board starts with its default columns")
	toDo, inProgress, done := cols[0], cols[1], cols[2]

	// The midpoint of `toDo` (STEP) and `inProgress` (2*STEP): both gaps are STEP/2,
	// comfortably clear of the 0.5 threshold.
	midpoint := 1.5 * domain.PositionStep
	newPos, err := columnRepo.UpdatePosition(context.Background(), done.ID, midpoint)
	require.NoError(t, err)
	assert.Equal(t, midpoint, newPos, "a comfortable gap must leave the requested position intact")

	ids, positions := columnOrder(t, board.ID)
	assert.Equal(t, []uuid.UUID{toDo.ID, done.ID, inProgress.ID}, ids)
	assert.Equal(t, []float64{domain.PositionStep, midpoint, 2 * domain.PositionStep}, positions,
		"no rewrite: the other columns keep their seeded STEP multiples")
}

// TestColumnUpdatePosition_NoRebalanceAtOpenEnd covers both open ends of the board
// partition. With no neighbour on one side that gap is ±Inf, so min(gap, ±Inf) must be
// judged on the real gap alone — an open end is not a packed one.
func TestColumnUpdatePosition_NoRebalanceAtOpenEnd(t *testing.T) {
	t.Cleanup(func() { testutil.TruncateAll(t, testPool) })

	user := fx.CreateTestUser(t)
	ws := fx.CreateTestWorkspace(t, user.ID)
	board := fx.CreateTestBoard(t, ws.ID, user.ID)

	columnRepo := pgRepos.NewColumnRepository(testPool)

	cols, err := columnRepo.GetColumnsByBoard(context.Background(), board.ID)
	require.NoError(t, err)
	require.Len(t, cols, len(domain.DefaultNewBoardColumnTitles), "a fixture board starts with its default columns")
	toDo, inProgress, done := cols[0], cols[1], cols[2]

	// Drag `toDo` past the end: no next neighbour (+Inf), prev gap 5000-3*STEP = 2000.
	newPos, err := columnRepo.UpdatePosition(context.Background(), toDo.ID, 5000)
	require.NoError(t, err)
	assert.Equal(t, 5000.0, newPos, "a trailing open end must not trigger a rebalance")

	ids, positions := columnOrder(t, board.ID)
	assert.Equal(t, []uuid.UUID{inProgress.ID, done.ID, toDo.ID}, ids)
	assert.Equal(t, []float64{2 * domain.PositionStep, 3 * domain.PositionStep, 5000}, positions)

	// The other open end: first position in the partition, no previous neighbour (-Inf),
	// next gap 3*STEP - 0.25.
	newPos, err = columnRepo.UpdatePosition(context.Background(), inProgress.ID, 0.25)
	require.NoError(t, err)
	assert.Equal(t, 0.25, newPos, "the leading open end must not trigger a rebalance either")

	ids, positions = columnOrder(t, board.ID)
	assert.Equal(t, []uuid.UUID{inProgress.ID, done.ID, toDo.ID}, ids)
	assert.Equal(t, []float64{0.25, 3 * domain.PositionStep, 5000}, positions)
}
