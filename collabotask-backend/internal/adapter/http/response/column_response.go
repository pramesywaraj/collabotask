package response

import (
	"collabotask/internal/domain/entity"
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

// ColumnWithCardsDTOToResponse bridges the still-DTO-based kanban path (board)
// to the response shape. TODO: remove once the board module is migrated off
// internal/dto, same as the card and column modules were.
func ColumnWithCardsDTOToResponse(col dto.ColumnWithCardsDTO) ColumnWithCardsResponse {
	cards := make([]CardResponse, 0, len(col.Cards))
	for _, c := range col.Cards {
		cards = append(cards, cardWithAssigneeDTOToResponse(c))
	}

	return ColumnWithCardsResponse{
		ColumnResponse: ColumnResponse{
			ID:        col.ID,
			BoardID:   col.BoardID,
			Title:     col.Title,
			Position:  col.Position,
			CreatedAt: col.CreatedAt,
			UpdatedAt: col.UpdatedAt,
		},
		Cards: cards,
	}
}

// cardWithAssigneeDTOToResponse bridges the still-DTO-based kanban path
// (board/column) to CardResponse. TODO: remove once the column/board modules
// are migrated off internal/dto, same as the card module was.
func cardWithAssigneeDTOToResponse(card dto.CardWithAssigneeDTO) CardResponse {
	var assignedTo *AssignedToResponse
	if card.AssignedTo != nil {
		assignedTo = &AssignedToResponse{
			ID:        *card.AssignedTo,
			AvatarURL: card.AssigneeAvatarURL,
		}
		if card.AssigneeName != nil {
			assignedTo.Name = *card.AssigneeName
		}
	}

	return CardResponse{
		ID:          card.ID,
		ColumnID:    card.ColumnID,
		Title:       card.Title,
		Description: card.Description,
		Position:    card.Position,
		AssignedTo:  assignedTo,
		DueDate:     card.DueDate,
		CreatedBy:   card.CreatedBy,
		CreatedAt:   card.CreatedAt,
		UpdatedAt:   card.UpdatedAt,
	}
}
