package auth

import (
	"collabotask/internal/domain/entity"

	"github.com/google/uuid"
)

type UserProfile struct {
	ID         uuid.UUID
	Email      string
	Name       string
	AvatarURL  *string
	SystemRole string
}

func userToProfile(user *entity.User) UserProfile {
	return UserProfile{
		ID:         user.ID,
		Email:      user.Email,
		Name:       user.Name,
		AvatarURL:  user.AvatarURL,
		SystemRole: string(user.SystemRole),
	}
}

type RegisterInput struct {
	Email    string
	Name     string
	Password string
}

type RegisterOutput struct {
	User  UserProfile
	Token string
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	User  UserProfile
	Token string
}
