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

type ColumnRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewColumnRepository(db *pgxpool.Pool) repository.ColumnRepository {
	return &ColumnRepositoryImpl{db: db}
}

const columnsCap = 16

func (cr *ColumnRepositoryImpl) Create(ctx context.Context, column *entity.Column) error {
	err := cr.db.QueryRow(
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

func (cr *ColumnRepositoryImpl) CreateMany(ctx context.Context, columns []*entity.Column) error {
	if len(columns) == 0 {
		return nil
	}

	tx, err := cr.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction to create many columns: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, column := range columns {
		_, err := tx.Exec(
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

	return tx.Commit(ctx)
}

func (cr *ColumnRepositoryImpl) GetByID(ctx context.Context, columnID uuid.UUID) (*entity.Column, error) {
	column := &entity.Column{}

	err := cr.db.QueryRow(
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

func (cr *ColumnRepositoryImpl) ListByBoard(ctx context.Context, boardID uuid.UUID) ([]*entity.Column, error) {
	rows, err := cr.db.Query(
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

func (cr *ColumnRepositoryImpl) GetMaxPosition(ctx context.Context, boardID uuid.UUID) (int, error) {
	var position int

	err := cr.db.QueryRow(
		ctx,
		getColumnMaxPositionQuery,
		boardID,
	).Scan(
		&position,
	)
	if err != nil {
		return -1, fmt.Errorf("failed to get column max position: %w", err)
	}

	return position, nil
}

func (cr *ColumnRepositoryImpl) Update(ctx context.Context, column *entity.Column) error {
	var title *string
	if column.Title != "" {
		title = &column.Title
	}

	updatedAt := time.Now()

	err := cr.db.QueryRow(
		ctx,
		updateColumnQuery,
		title,
		column.Position,
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

func (cr *ColumnRepositoryImpl) Delete(ctx context.Context, columnID uuid.UUID) error {
	result, err := cr.db.Exec(
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
