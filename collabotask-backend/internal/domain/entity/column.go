package entity

import (
	"time"

	"github.com/google/uuid"
)

type Column struct {
	ID        uuid.UUID `json:"id" db:"id"`
	BoardID   uuid.UUID `json:"board_id" db:"board_id"`
	Title     string    `json:"title" db:"title"`
	Position  int       `json:"position" db:"position"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func (Column) TableName() string {
	return "columns"
}

func (c *Column) IsEmpty() bool {
	return c.ID == uuid.Nil
}
