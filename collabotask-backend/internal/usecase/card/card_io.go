package card

import (
	"collabotask/internal/domain/entity"
	"time"

	"github.com/google/uuid"
)

// CardResult carries the persisted card together with its (optional) assignee.
// Use cases return domain entities directly; the HTTP layer maps them to a
// response shape. There is no intermediate DTO copy.
type CardResult struct {
	Card     *entity.Card
	Assignee *entity.User
}

type CreateCardInput struct {
	BoardID     uuid.UUID  `validate:"required"`
	ColumnID    uuid.UUID  `validate:"required"`
	Title       string     `validate:"required,min=1,max=500"`
	RequesterID uuid.UUID  `validate:"required"`
	Description *string    `validate:"omitempty,max=2000"`
	AssignedTo  *uuid.UUID `validate:"omitempty,uuid"`
	DueDate     *time.Time `validate:"omitempty"`
}

type CreateCardOutput = CardResult

type UpdateCardInput struct {
	BoardID            uuid.UUID `validate:"required"`
	ColumnID           uuid.UUID `validate:"required"`
	CardID             uuid.UUID `validate:"required"`
	RequesterID        uuid.UUID `validate:"required"`
	Title              *string   `validate:"omitempty,min=1,max=500"`
	Description        *string   `validate:"omitempty,max=2000"`
	DescriptionPresent bool
	AssignedTo         *uuid.UUID `validate:"omitempty,uuid"`
	AssignedToPresent  bool
	DueDate            *time.Time `validate:"omitempty"`
	DueDatePresent     bool
}

type UpdateCardOutput = CardResult

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
	// ToPosition is a fractional coordinate computed client-side. Any finite
	// float64 is valid (negative values are legal for head-inserts). Presence
	// is enforced at the HTTP layer via a *float64 pointer field on the request.
	ToPosition  float64
	RequesterID uuid.UUID `validate:"required"`
}

type MoveCardOutput = CardResult
