package repository

import (
	"collabotask/internal/domain/entity"
	"context"

	"github.com/google/uuid"
)

type CardRepository interface {
	Create(ctx context.Context, card *entity.Card) error
	Update(ctx context.Context, card *entity.Card) error
	Delete(ctx context.Context, cardID uuid.UUID) error
	DeleteWithReorder(ctx context.Context, cardID uuid.UUID) error
	GetByID(ctx context.Context, cardID uuid.UUID) (*entity.Card, error)
	GetCardsByColumn(ctx context.Context, columnID uuid.UUID) ([]*entity.Card, error)
	GetMaxPosition(ctx context.Context, columnID uuid.UUID) (int, error)
	IncrementPositionsFrom(ctx context.Context, columnID uuid.UUID, position int) error
	DecrementPositionsAfter(ctx context.Context, columnID uuid.UUID, position int) error
	Move(ctx context.Context, cardID, fromColumnID, toColumnID uuid.UUID, toPosition int) (*entity.Card, error)
}
