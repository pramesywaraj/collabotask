package repository

import (
	"collabotask/internal/domain/entity"
	"context"

	"github.com/google/uuid"
)

type ColumnRepository interface {
	Create(ctx context.Context, column *entity.Column) error
	CreateMany(ctx context.Context, columns []*entity.Column) error
	GetByID(ctx context.Context, columnID uuid.UUID) (*entity.Column, error)
	GetColumnsByBoard(ctx context.Context, boardID uuid.UUID) ([]*entity.Column, error)
	GetMaxPosition(ctx context.Context, boardID uuid.UUID) (float64, error)
	Update(ctx context.Context, column *entity.Column) error
	// UpdatePosition writes the column's new position and returns the position it
	// actually holds afterward. These differ when the write triggers a rebalance,
	// so the caller must use the returned value rather than the requested one.
	UpdatePosition(ctx context.Context, columnID uuid.UUID, position float64) (float64, error)
	Delete(ctx context.Context, columnID uuid.UUID) error
}
