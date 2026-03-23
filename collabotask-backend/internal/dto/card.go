package dto

import (
	"collabotask/internal/domain/entity"
	"time"

	"github.com/google/uuid"
)

type CardDTO struct {
	ID          uuid.UUID
	ColumnID    uuid.UUID
	Title       string
	Description *string
	Position    int
	AssignedTo  *uuid.UUID
	DueDate     *time.Time
	CreatedBy   uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CardWithAssigneeDTO struct {
	CardDTO

	AssigneeName      *string
	AssigneeAvatarURL *string
}

func CardToDTO(card *entity.Card) CardDTO {
	return CardDTO{
		ID:          card.ID,
		ColumnID:    card.ColumnID,
		Title:       card.Title,
		Description: card.Description,
		Position:    card.Position,
		AssignedTo:  card.AssignedTo,
		DueDate:     card.DueDate,
		CreatedBy:   card.CreatedBy,
		CreatedAt:   card.CreatedAt,
		UpdatedAt:   card.UpdatedAt,
	}
}

func CardWithAssigneeToDTO(card *entity.Card, assignee *entity.User) CardWithAssigneeDTO {
	result := CardWithAssigneeDTO{
		CardDTO: CardToDTO(card),
	}
	if assignee != nil {
		result.AssigneeName = &assignee.Name
		result.AssigneeAvatarURL = assignee.AvatarURL
	}
	return result
}
