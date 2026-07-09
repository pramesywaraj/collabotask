package response

import (
	"collabotask/internal/domain/entity"
	"time"

	"github.com/google/uuid"
)

type CardResponse struct {
	ID          uuid.UUID           `json:"id"`
	ColumnID    uuid.UUID           `json:"column_id"`
	Title       string              `json:"title"`
	Description *string             `json:"description"`
	Position    float64             `json:"position"`
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

// CardToResponse maps a domain card (and its optional assignee) to the HTTP
// response shape. assignee may be nil when the card is unassigned.
func CardToResponse(card *entity.Card, assignee *entity.User) CardResponse {
	var assignedTo *AssignedToResponse

	if card.AssignedTo != nil {
		assignedTo = &AssignedToResponse{
			ID: *card.AssignedTo,
		}
		if assignee != nil {
			assignedTo.Name = assignee.Name
			assignedTo.AvatarURL = assignee.AvatarURL
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
