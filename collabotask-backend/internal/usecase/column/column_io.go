package column

import (
	"collabotask/internal/domain/entity"

	"github.com/google/uuid"
)

// ColumnResult carries the persisted column. Use cases return the domain entity
// directly; the HTTP layer maps it to a response shape. No intermediate DTO copy.
type ColumnResult struct {
	Column *entity.Column
}

type CreateColumnInput struct {
	BoardID     uuid.UUID `validate:"required"`
	Title       string    `validate:"required,min=1,max=255"`
	RequesterID uuid.UUID `validate:"required"`
}

type CreateColumnOutput = ColumnResult

type UpdateColumnInput struct {
	BoardID     uuid.UUID `validate:"required"`
	ColumnID    uuid.UUID `validate:"required"`
	Title       string    `validate:"required,min=1,max=255"`
	RequesterID uuid.UUID `validate:"required"`
}

type UpdateColumnOutput = ColumnResult

type DeleteColumnInput struct {
	BoardID     uuid.UUID `validate:"required"`
	ColumnID    uuid.UUID `validate:"required"`
	RequesterID uuid.UUID `validate:"required"`
}

type UpdateColumnPositionInput struct {
	BoardID     uuid.UUID `validate:"required"`
	ColumnID    uuid.UUID `validate:"required"`
	// Position is a 0-based target index; out-of-range values are clamped in the
	// use case, so it carries no range constraint here (0 and negatives are valid input).
	Position    int
	RequesterID uuid.UUID `validate:"required"`
}

type UpdateColumnPositionOutput = ColumnResult
