package entity

import (
	"time"

	"github.com/google/uuid"
)

type Workspace struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" validate:"required,min=2,max=255" db:"name"`
	Description *string   `json:"description" validate:"max=1000" db:"description"`
	OwnerID     uuid.UUID `json:"owner_id" db:"owner_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

func (Workspace) TableName() string {
	return "workspaces"
}

func (w *Workspace) IsEmpty() bool {
	return w.ID == uuid.Nil
}
