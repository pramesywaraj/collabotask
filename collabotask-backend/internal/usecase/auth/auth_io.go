package auth

import (
	"collabotask/internal/domain/entity"

	"github.com/google/uuid"
)

// UserProfile is the safe, outward-facing view of a user. It deliberately
// OMITS entity.User.PasswordHash (and timestamps) so the credential hash can
// never leak out of the use case. This is a genuine transformation of the
// entity, which is exactly why it is kept (unlike the 1:1 DTOs that were
// removed from other modules).
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
