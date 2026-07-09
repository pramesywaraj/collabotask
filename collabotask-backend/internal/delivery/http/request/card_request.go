package request

import (
	"time"

	"github.com/google/uuid"
)

type CreateCardRequest struct {
	Title       string     `json:"title" binding:"required,min=1,max=500"`
	Description *string    `json:"description" binding:"omitempty"`
	AssignedTo  *uuid.UUID `json:"assigned_to" binding:"omitempty"`
	DueDate     *time.Time `json:"due_date" binding:"omitempty"`
}

type UpdateCardRequest struct {
	Title       *string                  `json:"title" binding:"omitempty,min=1,max=500"`
	Description OptionalPatch[string]    `json:"description"`
	AssignedTo  OptionalPatch[uuid.UUID] `json:"assigned_to"`
	DueDate     OptionalPatch[time.Time] `json:"due_date"`
}

type MoveCardRequest struct {
	ToColumnID uuid.UUID `json:"to_column_id" binding:"required"`
	// Pointer enforces presence (omitted or null → 400). Any finite float64 value
	// is valid; the JSON decoder rejects NaN/±Inf before binding runs.
	ToPosition *float64 `json:"to_position" binding:"required"`
}
