package response

import (
	"collabotask/internal/usecase/auth"

	"github.com/google/uuid"
)

type UserResponse struct {
	ID         uuid.UUID `json:"id"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	AvatarURL  *string   `json:"avatar_url,omitempty"`
	SystemRole string    `json:"system_role"`
}

type AuthResponse struct {
	User UserResponse `json:"user"`
}

func UserToResponse(u auth.UserProfile) UserResponse {
	return UserResponse{
		ID:         u.ID,
		Email:      u.Email,
		Name:       u.Name,
		AvatarURL:  u.AvatarURL,
		SystemRole: u.SystemRole,
	}
}
