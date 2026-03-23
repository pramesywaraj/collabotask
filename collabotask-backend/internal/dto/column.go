package dto

import (
	"collabotask/internal/domain/entity"
	"time"

	"github.com/google/uuid"
)

type ColumnDTO struct {
	ID        uuid.UUID
	BoardID   uuid.UUID
	Title     string
	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ColumnWithCardsDTO struct {
	ColumnDTO

	Cards []CardWithAssigneeDTO
}

func ColumnToDTO(column *entity.Column) ColumnDTO {
	return ColumnDTO{
		ID:        column.ID,
		BoardID:   column.BoardID,
		Title:     column.Title,
		Position:  column.Position,
		CreatedAt: column.CreatedAt,
		UpdatedAt: column.UpdatedAt,
	}
}
