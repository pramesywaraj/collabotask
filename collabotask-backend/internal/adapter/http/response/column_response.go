package response

import (
	"collabotask/internal/dto"
	"time"

	"github.com/google/uuid"
)

type ColumnResponse struct {
	ID        uuid.UUID `json:"id"`
	BoardID   uuid.UUID `json:"board_id"`
	Title     string    `json:"title"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ColumnWithCardsResponse struct {
	ColumnResponse
	Cards []CardResponse `json:"cards"`
}

func ColumnDTOToResponse(column dto.ColumnDTO) ColumnResponse {
	return ColumnResponse{
		ID:        column.ID,
		BoardID:   column.BoardID,
		Title:     column.Title,
		Position:  column.Position,
		CreatedAt: column.CreatedAt,
		UpdatedAt: column.UpdatedAt,
	}
}

func ColumnWithCardsDTOToResponse(col dto.ColumnWithCardsDTO) ColumnWithCardsResponse {
	cards := make([]CardResponse, 0, len(col.Cards))
	for _, c := range col.Cards {
		cards = append(cards, CardDTOToResponse(c))
	}

	return ColumnWithCardsResponse{
		ColumnResponse: ColumnDTOToResponse(col.ColumnDTO),
		Cards:          cards,
	}
}
