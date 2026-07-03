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
	// ToPosition is a 0-based target index; the top slot is 0. Only the upper
	// bound is clamped in the use case, so negatives are rejected here (min=0),
	// but 0 must stay valid — hence no `required` (which would reject the zero value).
	ToPosition  int       `validate:"min=0"`
	RequesterID uuid.UUID `validate:"required"`
}

type MoveCardOutput = CardResult
