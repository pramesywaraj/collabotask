package card

import (
	"collabotask/internal/dto"
	"context"
	"time"

	"github.com/google/uuid"
)

type CardUseCase interface {
	CreateCard(ctx context.Context, input CreateCardInput) (*CreateCardOutput, error)
	UpdateCard(ctx context.Context, input UpdateCardInput) (*UpdateCardOutput, error)
	DeleteCard(ctx context.Context, input DeleteCardInput) error
	MoveCard(ctx context.Context, input MoveCardInput) (*MoveCardOutput, error)
}

type CreateCardInput struct {
	BoardID     uuid.UUID  `validate:"required"`
	ColumnID    uuid.UUID  `validate:"required"`
	Title       string     `validate:"required,min=1,max=500"`
	RequesterID uuid.UUID  `validate:"required"`
	Description *string    `validate:"omitempty"`
	AssignedTo  *uuid.UUID `validate:"omitempty"`
	DueDate     *time.Time `validate:"omitempty"`
}

type CreateCardOutput struct {
	Card dto.CardWithAssigneeDTO
}

type UpdateCardInput struct {
	BoardID     uuid.UUID  `validate:"required"`
	ColumnID    uuid.UUID  `validate:"required"`
	CardID      uuid.UUID  `validate:"required"`
	RequesterID uuid.UUID  `validate:"required"`
	Title       *string    `validate:"omitempty,min=1,max=500"`
	Description *string    `validate:"omitempty"`
	AssignedTo  *uuid.UUID `validate:"omitempty"`
	DueDate     *time.Time `validate:"omitempty"`
}

type UpdateCardOutput struct {
	Card dto.CardWithAssigneeDTO
}

type DeleteCardInput struct {
	BoardID     uuid.UUID `validate:"required"`
	ColumnID    uuid.UUID `validate:"required"`
	CardID      uuid.UUID `validate:"required"`
	RequesterID uuid.UUID `validate:"required"`
}

type MoveCardInput struct {
	BoardID      uuid.UUID `validate:"required"`
	CardID       uuid.UUID `validate:"required"`
	FromColumnID uuid.UUID `validate:"required"`
	ToColumnID   uuid.UUID `validate:"required"`
	ToPosition   int       `validate:"required,min=0"`
	RequesterID  uuid.UUID `validate:"required"`
}

type MoveCardOutput struct {
	Card dto.CardWithAssigneeDTO
}
