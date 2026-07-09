package postgres

import (
	"collabotask/internal/domain"
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// rebalanceIfNeeded inspects the two immediate neighbor gaps around the just-moved
// item. If the minimum gap falls below domain.PositionRebalanceThreshold, it rewrites
// the entire partition to evenly-spaced domain.PositionStep multiples within the same
// transaction. It returns the moved item's position after the operation: unchanged
// when no rebalance fires, or its new STEP-multiple value when one does — callers must
// propagate this into the returned entity so the response never reports a position the
// DB no longer holds. table/partCol are package-internal constants, never user-supplied.
func rebalanceIfNeeded(
	ctx context.Context,
	tx pgx.Tx,
	table string,
	partCol string,
	partVal uuid.UUID,
	itemID uuid.UUID,
	position float64,
) (float64, error) {
	prevPos, err := neighborPosition(ctx, tx, table, partCol, partVal, itemID, position, true)
	if err != nil {
		return 0, err
	}
	nextPos, err := neighborPosition(ctx, tx, table, partCol, partVal, itemID, position, false)
	if err != nil {
		return 0, err
	}

	gapBefore := position - prevPos
	gapAfter := nextPos - position
	if math.Min(gapBefore, gapAfter) >= domain.PositionRebalanceThreshold {
		return position, nil
	}

	// Rewrite all positions in the partition to STEP multiples.
	// The format string is safe: table and partCol are always hardcoded by the caller.
	_, err = tx.Exec(ctx, fmt.Sprintf(`
		UPDATE %s
		SET position = rn.rn * $1
		FROM (
			SELECT id, ROW_NUMBER() OVER (ORDER BY position ASC, id ASC) AS rn
			FROM %s
			WHERE %s = $2
		) AS rn
		WHERE %s.id = rn.id AND %s.%s = $2
	`, table, table, partCol, table, table, partCol), domain.PositionStep, partVal)
	if err != nil {
		return 0, fmt.Errorf("rebalance %s: %w", table, err)
	}

	// The rewrite gave itemID a new STEP-multiple position; re-read it so the caller
	// can return the value the DB actually holds. Only runs on the rare rebalance path.
	var newPos float64
	err = tx.QueryRow(ctx, fmt.Sprintf(`SELECT position FROM %s WHERE id = $1`, table), itemID).Scan(&newPos)
	if err != nil {
		return 0, fmt.Errorf("read rebalanced position in %s: %w", table, err)
	}

	return newPos, nil
}

// neighborPosition returns the position of the nearest item before (prev=true)
// or after (prev=false) the item at (position, itemID) in the given partition.
// Returns ±Inf when no neighbor exists (open end — no rebalance needed there).
func neighborPosition(
	ctx context.Context,
	tx pgx.Tx,
	table, partCol string,
	partVal, itemID uuid.UUID,
	position float64,
	prev bool,
) (float64, error) {
	var query string
	if prev {
		// Largest position strictly before (position, itemID) in sort order.
		query = fmt.Sprintf(`
			SELECT position FROM %s
			WHERE %s = $1
			  AND (position < $2 OR (position = $2 AND id < $3))
			ORDER BY position DESC, id DESC
			LIMIT 1
		`, table, partCol)
	} else {
		// Smallest position strictly after (position, itemID) in sort order.
		query = fmt.Sprintf(`
			SELECT position FROM %s
			WHERE %s = $1
			  AND (position > $2 OR (position = $2 AND id > $3))
			ORDER BY position ASC, id ASC
			LIMIT 1
		`, table, partCol)
	}

	var pos float64
	err := tx.QueryRow(ctx, query, partVal, position, itemID).Scan(&pos)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			if prev {
				return math.Inf(-1), nil
			}
			return math.Inf(1), nil
		}
		return 0, fmt.Errorf("query neighbor in %s: %w", table, err)
	}

	return pos, nil
}
