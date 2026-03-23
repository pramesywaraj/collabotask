package column

import (
	"collabotask/internal/dto"
	"context"

	"github.com/google/uuid"
)

type ColumnUseCase interface {
	CreateColumn(ctx context.Context, input CreateColumnInput) (*CreateColumnOutput, error)
	UpdateColumn(ctx context.Context, input UpdateColumnInput) (*UpdateColumnOutput, error)
	DeleteColumn(ctx context.Context, input DeleteColumnInput) error
	UpdateColumnPosition(ctx context.Context, input UpdateColumnPositionInput) (*UpdateColumnPositionOutput, error)
}

type CreateColumnInput struct {
	BoardID     uuid.UUID `validate:"required"`
	Title       string    `validate:"required,min=1,max=255"`
	RequesterID uuid.UUID `validate:"required"`
}

type CreateColumnOutput struct {
	Column dto.ColumnDTO
}

type UpdateColumnInput struct {
	ColumnID    uuid.UUID `validate:"required"`
	Title       string    `validate:"required,min=1,max=255"`
	RequesterID uuid.UUID `validate:"required"`
}

type UpdateColumnOutput struct {
	Column dto.ColumnDTO
}

type DeleteColumnInput struct {
	ColumnID    uuid.UUID `validate:"required"`
	RequesterID uuid.UUID `validate:"required"`
}

type UpdateColumnPositionInput struct {
	ColumnID    uuid.UUID `validate:"required"`
	Position    int       `validate:"required,min=0"`
	RequesterID uuid.UUID `validate:"required"`
}

type UpdateColumnPositionOutput struct {
	Column dto.ColumnDTO
}
