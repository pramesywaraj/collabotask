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

type columnRepository struct {
	base
}

func NewColumnRepository(pool *pgxpool.Pool) repository.ColumnRepository {
	return &columnRepository{base: base{pool: pool}}
}

const columnsCap = 16

func (cr *columnRepository) Create(ctx context.Context, column *entity.Column) error {
	err := cr.exec(ctx).QueryRow(
		ctx,
		createColumnQuery,
		column.BoardID,
		column.Title,
		column.Position,
	).Scan(
		&column.ID,
		&column.BoardID,
		&column.Title,
		&column.Position,
		&column.CreatedAt,
		&column.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return domain.ErrConstraintViolation
			}
		}
		return fmt.Errorf("failed to create column: %w", err)
	}

	return nil
}

func (cr *columnRepository) CreateMany(ctx context.Context, columns []*entity.Column) error {
	if len(columns) == 0 {
		return nil
	}

	return cr.tx(ctx, func(ctx context.Context) error {
		for _, column := range columns {
			_, err := cr.exec(ctx).Exec(
				ctx,
				createColumnQuery,
				column.BoardID,
				column.Title,
				column.Position,
			)
			if err != nil {
				var pgErr *pgconn.PgError
				if errors.As(err, &pgErr) && pgErr.Code == "23505" {
					return domain.ErrConstraintViolation
				}
				return fmt.Errorf("failed to create column: %w", err)
			}
		}
		return nil
	})
}

func (cr *columnRepository) GetByID(ctx context.Context, columnID uuid.UUID) (*entity.Column, error) {
	column := &entity.Column{}

	err := cr.exec(ctx).QueryRow(
		ctx,
		getColumnByIDQuery,
		columnID,
	).Scan(
		&column.ID,
		&column.BoardID,
		&column.Title,
		&column.Position,
		&column.CreatedAt,
		&column.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrColumnNotFound
		}
		return nil, fmt.Errorf("failed to get column by id: %w", err)
	}

	return column, nil
}

func (cr *columnRepository) GetColumnsByBoard(ctx context.Context, boardID uuid.UUID) ([]*entity.Column, error) {
	rows, err := cr.exec(ctx).Query(
		ctx,
		listColumnByBoardIDQuery,
		boardID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns by board id: %w", err)
	}
	defer rows.Close()

	columns := make([]*entity.Column, 0, columnsCap)
	for rows.Next() {
		column := &entity.Column{}

		err := rows.Scan(
			&column.ID,
			&column.BoardID,
			&column.Title,
			&column.Position,
			&column.CreatedAt,
			&column.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}

		columns = append(columns, column)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating columns in board: %w", err)
	}

	return columns, nil
}

func (cr *columnRepository) GetMaxPosition(ctx context.Context, boardID uuid.UUID) (float64, error) {
	var position float64

	err := cr.exec(ctx).QueryRow(
		ctx,
		getColumnMaxPositionQuery,
		boardID,
	).Scan(
		&position,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to get column max position: %w", err)
	}

	return position, nil
}

func (cr *columnRepository) Update(ctx context.Context, column *entity.Column) error {
	var title *string
	if column.Title != "" {
		title = &column.Title
	}

	updatedAt := time.Now()

	err := cr.exec(ctx).QueryRow(
		ctx,
		updateColumnQuery,
		title,
		updatedAt,
		column.ID,
	).Scan(
		&column.ID,
		&column.BoardID,
		&column.Title,
		&column.Position,
		&column.CreatedAt,
		&column.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrColumnNotFound
		}
		return fmt.Errorf("failed to update column: %w", err)
	}

	return nil
}

func (cr *columnRepository) UpdatePosition(ctx context.Context, columnID uuid.UUID, position float64) (float64, error) {
	var newPos float64
	err := cr.tx(ctx, func(ctx context.Context) error {
		var col entity.Column
		err := cr.exec(ctx).QueryRow(ctx, updateColumnPositionQuery, position, columnID).Scan(
			&col.ID,
			&col.BoardID,
			&col.Title,
			&col.Position,
			&col.CreatedAt,
			&col.UpdatedAt,
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.ErrColumnNotFound
			}
			return fmt.Errorf("failed to update column position: %w", err)
		}

		// exec(ctx) returns the ambient pgx.Tx inside tx(ctx, fn) — safe to assert.
		p, err := rebalanceIfNeeded(ctx, cr.exec(ctx).(pgx.Tx), "columns", "board_id", col.BoardID, col.ID, col.Position)
		if err != nil {
			return fmt.Errorf("failed to rebalance columns: %w", err)
		}
		newPos = p
		return nil
	})
	return newPos, err
}

func (cr *columnRepository) Delete(ctx context.Context, columnID uuid.UUID) error {
	result, err := cr.exec(ctx).Exec(
		ctx,
		deleteColumnQuery,
		columnID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete column: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrColumnNotFound
	}

	return nil
}
