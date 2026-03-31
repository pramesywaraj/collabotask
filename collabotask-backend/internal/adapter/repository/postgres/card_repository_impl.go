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

type CardRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewCardRepository(db *pgxpool.Pool) repository.CardRepository {
	return &CardRepositoryImpl{db: db}
}

const cardCaps = 16

func (cdr *CardRepositoryImpl) Create(ctx context.Context, card *entity.Card) error {
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

func (cdr *CardRepositoryImpl) Update(ctx context.Context, card *entity.Card) error {
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

func (cdr *CardRepositoryImpl) Delete(ctx context.Context, cardID uuid.UUID) error {
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

func (cdr *CardRepositoryImpl) DeleteWithReorder(ctx context.Context, cardID uuid.UUID) error {
	tx, err := cdr.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin delete card with reorder transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var columnID uuid.UUID
	var position int

	err = tx.QueryRow(
		ctx,
		lockCardQuery,
		cardID,
	).Scan(&columnID, &position)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrCardNotFound
		}
		return fmt.Errorf("failed to lock card when delete with reorder: %w", err)
	}

	result, err := tx.Exec(
		ctx,
		deleteCardQuery,
		cardID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete card: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrCardNotFound
	}

	// Reorder remaining column's cards
	if _, err := tx.Exec(ctx, decrementPositionCardAfterQuery, columnID, position); err != nil {
		return fmt.Errorf("failed to reorder cards after delete: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit delete card transaction: %w", err)
	}

	return nil
}

func (cdr *CardRepositoryImpl) GetByID(ctx context.Context, cardID uuid.UUID) (*entity.Card, error) {
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

func (cdr *CardRepositoryImpl) ListByColumn(ctx context.Context, columnID uuid.UUID) ([]*entity.Card, error) {
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

func (cdr *CardRepositoryImpl) GetMaxPosition(ctx context.Context, columnID uuid.UUID) (int, error) {
	var position int

	err := cdr.db.QueryRow(
		ctx,
		getMaxCardPositionQuery,
		columnID,
	).Scan(
		&position,
	)
	if err != nil {
		return -1, fmt.Errorf("failed to get card max position: %w", err)
	}

	return position, nil
}

func (cdr *CardRepositoryImpl) IncrementPositionsFrom(ctx context.Context, columnID uuid.UUID, position int) error {
	_, err := cdr.db.Exec(
		ctx,
		incrementPositionCardFromQuery,
		columnID,
		position,
	)
	if err != nil {
		return fmt.Errorf("failed to increment card positions from: %w", err)
	}

	return nil
}

func (cdr *CardRepositoryImpl) DecrementPositionsAfter(ctx context.Context, columnID uuid.UUID, position int) error {
	_, err := cdr.db.Exec(
		ctx,
		decrementPositionCardAfterQuery,
		columnID,
		position,
	)
	if err != nil {
		return fmt.Errorf("failed to decrement card positions after: %w", err)
	}

	return nil
}

func (cdr *CardRepositoryImpl) Move(ctx context.Context, cardID, fromColumnID, toColumnID uuid.UUID, toPosition int) (*entity.Card, error) {
	tx, err := cdr.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin move card transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var actualColumnID uuid.UUID
	var oldPosition int

	err = tx.QueryRow(ctx, lockCardQuery, cardID).Scan(&actualColumnID, &oldPosition)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCardNotFound
		}
		return nil, fmt.Errorf("failed to lock card: %w", err)
	}
	if fromColumnID != actualColumnID {
		return nil, domain.ErrInconsistentState
	}

	if _, err := tx.Exec(ctx, decrementPositionCardAfterQuery, actualColumnID, oldPosition); err != nil {
		return nil, fmt.Errorf("failed to decrement card position: %w", err)
	}
	if _, err := tx.Exec(ctx, incrementPositionCardFromQuery, toColumnID, toPosition); err != nil {
		return nil, fmt.Errorf("failed to increment card position: %w", err)
	}

	moved := &entity.Card{}
	err = tx.QueryRow(
		ctx,
		moveCardQuery,
		toColumnID,
		toPosition,
		cardID,
	).Scan(
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

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit move card transaction: %w", err)
	}

	return moved, nil
}
