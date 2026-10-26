package postgres

import (
	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/domain/repository"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type cardRepository struct {
	base
}

func NewCardRepository(pool *pgxpool.Pool) repository.CardRepository {
	return &cardRepository{base: base{pool: pool}}
}

const cardCaps = 16

// isAssigneeFKViolation reports whether err is the composite-FK violation
// (23503 on fk_cards_assignee_board_member), meaning the assignee is not a
// member of the card's board. Shared by Create and Update, which both map it to
// domain.ErrAssigneeNotBoardMember (400).
func isAssigneeFKViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) &&
		pgErr.Code == "23503" &&
		pgErr.ConstraintName == "fk_cards_assignee_board_member"
}

func (cdr *cardRepository) Create(ctx context.Context, card *entity.Card) error {
	err := cdr.exec(ctx).QueryRow(
		ctx,
		createCardQuery,
		card.ColumnID,
		card.BoardID,
		card.Title,
		card.Description,
		card.Position,
		card.AssignedTo,
		card.DueDate,
		card.CreatedBy,
	).Scan(
		&card.ID,
		&card.ColumnID,
		&card.BoardID,
		&card.Title,
		&card.Description,
		&card.Position,
		&card.AssignedTo,
		&card.DueDate,
		&card.CreatedBy,
		&card.CreatedAt,
		&card.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrConstraintViolation
		}
		if isAssigneeFKViolation(err) {
			return domain.ErrAssigneeNotBoardMember
		}
		return fmt.Errorf("failed to create card: %w", err)
	}

	return nil
}

func (cdr *cardRepository) Update(ctx context.Context, card *entity.Card) error {
	var title *string
	if card.Title != "" {
		title = &card.Title
	}

	updatedAt := time.Now()

	err := cdr.exec(ctx).QueryRow(
		ctx,
		updateCardQuery,
		title,
		card.Description,
		card.AssignedTo,
		card.DueDate,
		updatedAt,
		card.ID,
	).Scan(
		&card.ID,
		&card.ColumnID,
		&card.BoardID,
		&card.Title,
		&card.Description,
		&card.Position,
		&card.AssignedTo,
		&card.DueDate,
		&card.CreatedBy,
		&card.CreatedAt,
		&card.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrCardNotFound
		}
		if isAssigneeFKViolation(err) {
			return domain.ErrAssigneeNotBoardMember
		}
		return fmt.Errorf("failed to update card: %w", err)
	}

	return nil
}

func (cdr *cardRepository) Delete(ctx context.Context, cardID uuid.UUID) error {
	result, err := cdr.exec(ctx).Exec(
		ctx,
		deleteCardQuery,
		cardID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete the card: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrCardNotFound
	}

	return nil
}

func (cdr *cardRepository) GetByID(ctx context.Context, cardID uuid.UUID) (*entity.Card, error) {
	card := &entity.Card{}

	err := cdr.exec(ctx).QueryRow(
		ctx,
		getCardByIDQuery,
		cardID,
	).Scan(
		&card.ID,
		&card.ColumnID,
		&card.BoardID,
		&card.Title,
		&card.Description,
		&card.Position,
		&card.AssignedTo,
		&card.DueDate,
		&card.CreatedBy,
		&card.CreatedAt,
		&card.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCardNotFound
		}
		return nil, fmt.Errorf("failed to get card by id: %w", err)
	}

	return card, nil
}

func (cdr *cardRepository) GetCardsByColumn(ctx context.Context, columnID uuid.UUID) ([]*entity.Card, error) {
	rows, err := cdr.exec(ctx).Query(
		ctx,
		listCardByColumnQuery,
		columnID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query cards by column id: %w", err)
	}
	defer rows.Close()

	cards := make([]*entity.Card, 0, cardCaps)
	for rows.Next() {
		card := &entity.Card{}

		err := rows.Scan(
			&card.ID,
			&card.ColumnID,
			&card.BoardID,
			&card.Title,
			&card.Description,
			&card.Position,
			&card.AssignedTo,
			&card.DueDate,
			&card.CreatedBy,
			&card.CreatedAt,
			&card.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}

		cards = append(cards, card)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cards in column: %w", err)
	}

	return cards, nil
}

func (cdr *cardRepository) GetMaxPosition(ctx context.Context, columnID uuid.UUID) (float64, error) {
	var position float64

	err := cdr.exec(ctx).QueryRow(
		ctx,
		getMaxCardPositionQuery,
		columnID,
	).Scan(
		&position,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to get card max position: %w", err)
	}

	return position, nil
}

func (cdr *cardRepository) Move(ctx context.Context, cardID, fromColumnID, toColumnID uuid.UUID, toPosition float64) (*entity.Card, error) {
	var moved *entity.Card
	err := cdr.tx(ctx, func(ctx context.Context) error {
		var actualColumnID uuid.UUID
		err := cdr.exec(ctx).QueryRow(ctx, lockCardForMoveQuery, cardID).Scan(&actualColumnID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.ErrCardNotFound
			}
			return fmt.Errorf("failed to lock card: %w", err)
		}
		if fromColumnID != actualColumnID {
			return domain.ErrInconsistentState
		}

		card := &entity.Card{}
		err = cdr.exec(ctx).QueryRow(ctx, moveCardQuery, toColumnID, toPosition, cardID).Scan(
			&card.ID,
			&card.ColumnID,
			&card.BoardID,
			&card.Title,
			&card.Description,
			&card.Position,
			&card.AssignedTo,
			&card.DueDate,
			&card.CreatedBy,
			&card.CreatedAt,
			&card.UpdatedAt,
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.ErrCardNotFound
			}
			return fmt.Errorf("failed to move card: %w", err)
		}

		// exec(ctx) returns the ambient pgx.Tx inside tx(ctx, fn) — safe to assert.
		newPos, err := rebalanceIfNeeded(ctx, cdr.exec(ctx).(pgx.Tx), "cards", "column_id", toColumnID, card.ID, card.Position)
		if err != nil {
			return fmt.Errorf("failed to rebalance cards: %w", err)
		}
		card.Position = newPos
		moved = card
		return nil
	})
	return moved, err
}
