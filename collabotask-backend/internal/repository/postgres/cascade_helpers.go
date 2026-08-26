package postgres

import (
	"collabotask/internal/domain/repository"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// scanAffectedCards collects the cards returned by an unassign-cascade query.
// If boardID is non-nil the board is a fixed query parameter (board path) and
// each card's BoardID is taken from it; otherwise BoardID is scanned as the
// third column (workspace path, which spans multiple boards).
//
// Draining rows here also frees the transaction's connection for the next query
// on the same tx — the cascades reorder this collection ahead of the membership
// delete so the FK's ON DELETE SET NULL doesn't empty the broadcast list first.
func scanAffectedCards(rows pgx.Rows, boardID *uuid.UUID) ([]repository.AffectedCard, error) {
	var affected []repository.AffectedCard
	for rows.Next() {
		var card repository.AffectedCard
		var err error
		if boardID != nil {
			err = rows.Scan(&card.CardID, &card.ColumnID)
			card.BoardID = *boardID
		} else {
			err = rows.Scan(&card.CardID, &card.ColumnID, &card.BoardID)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to scan affected card: %w", err)
		}
		affected = append(affected, card)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating affected cards: %w", err)
	}
	return affected, nil
}
