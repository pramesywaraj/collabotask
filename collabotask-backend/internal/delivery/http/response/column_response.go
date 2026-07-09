package response

import (
	"collabotask/internal/domain/entity"
	"time"

	"github.com/google/uuid"
)

type ColumnResponse struct {
	ID        uuid.UUID `json:"id"`
	BoardID   uuid.UUID `json:"board_id"`
	Title     string    `json:"title"`
	Position  float64   `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ColumnWithCardsResponse struct {
	ColumnResponse
	Cards []CardResponse `json:"cards"`
}

// ColumnToResponse maps a domain column to the HTTP response shape.
func ColumnToResponse(column *entity.Column) ColumnResponse {
	return ColumnResponse{
		ID:        column.ID,
		BoardID:   column.BoardID,
		Title:     column.Title,
		Position:  column.Position,
		CreatedAt: column.CreatedAt,
		UpdatedAt: column.UpdatedAt,
	}
}
