package dto

import (
	"collabotask/internal/domain/entity"

	"github.com/google/uuid"
)

type UserDTO struct {
	ID         uuid.UUID `json:"id"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	AvatarURL  *string   `json:"avatar_url"`
	SystemRole string    `json:"system_role"`
}

func UserToDTO(user *entity.User) UserDTO {
	return UserDTO{
		ID:         user.ID,
		Email:      user.Email,
		Name:       user.Name,
		AvatarURL:  user.AvatarURL,
		SystemRole: string(user.SystemRole),
	}
}
