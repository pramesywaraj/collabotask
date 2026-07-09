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
	db *pgxpool.Pool
}

func NewCardRepository(db *pgxpool.Pool) repository.CardRepository {
	return &cardRepository{db: db}
}

const cardCaps = 16

func (cdr *cardRepository) Create(ctx context.Context, card *entity.Card) error {
	err := cdr.db.QueryRow(
		ctx,
		createCardQuery,
		card.ColumnID,
		card.Title,
		card.Description,
		card.Position,
		card.AssignedTo,
		card.DueDate,
		card.CreatedBy,
	).Scan(
		&card.ID,
		&card.ColumnID,
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

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return domain.ErrConstraintViolation
			}
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

	err := cdr.db.QueryRow(
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
		return fmt.Errorf("failed to update card: %w", err)
	}

	return nil
}

func (cdr *cardRepository) Delete(ctx context.Context, cardID uuid.UUID) error {
	result, err := cdr.db.Exec(
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

	err := cdr.db.QueryRow(
		ctx,
		getCardByIDQuery,
		cardID,
	).Scan(
		&card.ID,
		&card.ColumnID,
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
	rows, err := cdr.db.Query(
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

	err := cdr.db.QueryRow(
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
	tx, err := cdr.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin move card transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var actualColumnID uuid.UUID
	err = tx.QueryRow(ctx, lockCardForMoveQuery, cardID).Scan(&actualColumnID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCardNotFound
		}
		return nil, fmt.Errorf("failed to lock card: %w", err)
	}
	if fromColumnID != actualColumnID {
		return nil, domain.ErrInconsistentState
	}

	moved := &entity.Card{}
	err = tx.QueryRow(ctx, moveCardQuery, toColumnID, toPosition, cardID).Scan(
		&moved.ID,
		&moved.ColumnID,
		&moved.Title,
		&moved.Description,
		&moved.Position,
		&moved.AssignedTo,
		&moved.DueDate,
		&moved.CreatedBy,
		&moved.CreatedAt,
		&moved.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCardNotFound
		}
		return nil, fmt.Errorf("failed to move card: %w", err)
	}

	newPos, err := rebalanceIfNeeded(ctx, tx, "cards", "column_id", toColumnID, moved.ID, moved.Position)
	if err != nil {
		return nil, fmt.Errorf("failed to rebalance cards: %w", err)
	}
	moved.Position = newPos

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit move card transaction: %w", err)
	}

	return moved, nil
}
