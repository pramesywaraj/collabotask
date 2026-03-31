package response

import (
	"collabotask/internal/dto"
	"time"

	"github.com/google/uuid"
)

type CardResponse struct {
	ID          uuid.UUID           `json:"id"`
	ColumnID    uuid.UUID           `json:"column_id"`
	Title       string              `json:"title"`
	Description *string             `json:"description"`
	Position    int                 `json:"position"`
	AssignedTo  *AssignedToResponse `json:"assigned_to"`
	DueDate     *time.Time          `json:"due_date"`
	CreatedBy   uuid.UUID           `json:"created_by"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

type AssignedToResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	AvatarURL *string   `json:"avatar_url"`
}

func CardDTOToResponse(card dto.CardWithAssigneeDTO) CardResponse {
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
