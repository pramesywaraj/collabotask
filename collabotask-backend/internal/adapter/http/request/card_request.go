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
	// Do Not Delete, maybe will use later
	// FromColumnID uuid.UUID `json:"from_column_id" binding:"required"`
	ToColumnID uuid.UUID `json:"to_column_id" binding:"required"`
	ToPosition int       `json:"to_position" binding:"min=0"`
}
