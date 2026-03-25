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
	ListByBoard(ctx context.Context, boardID uuid.UUID) ([]*entity.Column, error)
	GetMaxPosition(ctx context.Context, boardID uuid.UUID) (int, error)
	Update(ctx context.Context, column *entity.Column) error
	ReorderPositions(ctx context.Context, columns []*entity.Column) error
	Delete(ctx context.Context, columnID uuid.UUID) error
}
