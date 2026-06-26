package entity

import (
	"time"

	"github.com/google/uuid"
)

type Card struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	ColumnID    uuid.UUID  `json:"column_id" db:"column_id"`
	Title       string     `json:"title" db:"title"`
	Description *string    `json:"description" db:"description"`
	Position    int        `json:"position" db:"position"`
	AssignedTo  *uuid.UUID `json:"assigned_to" db:"assigned_to"`
	DueDate     *time.Time `json:"due_date" db:"due_date"`
	CreatedBy   uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

func (Card) TableName() string {
	return "cards"
}

func (c *Card) IsEmpty() bool {
	return c.ID == uuid.Nil
}

func (c *Card) BelongsToColumn(columnID uuid.UUID) bool {
	return c.ColumnID == columnID
}
