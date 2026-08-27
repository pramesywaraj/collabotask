package postgres

import (
	"collabotask/internal/domain/repository"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// scanAffectedCards drains an unassign-cascade query into the affected-card list
// the caller broadcasts. The query must RETURN (card_id, column_id, board_id);
// both cascade paths read cards.board_id directly (denormalized in migration 000009),
// so they share one row shape.
func scanAffectedCards(rows pgx.Rows) ([]repository.AffectedCard, error) {
	var affected []repository.AffectedCard
	for rows.Next() {
		var card repository.AffectedCard
		if err := rows.Scan(&card.CardID, &card.ColumnID, &card.BoardID); err != nil {
			return nil, fmt.Errorf("failed to scan affected card: %w", err)
		}
		affected = append(affected, card)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating affected cards: %w", err)
	}
	return affected, nil
}
