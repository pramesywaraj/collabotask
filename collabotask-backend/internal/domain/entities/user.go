package entities

import (
	"time"

	"github.com/google/uuid"
)

type SystemRole string

const (
	SystemRoleSuperAdmin SystemRole = "SUPER_ADMIN"
	SystemRoleUser       SystemRole = "USER"
)

type User struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Email        string     `json:"email" db:"email" validate:"required,email"`
	Name         string     `json:"name" db:"name" validate:"required,min=1,max=255"`
	PasswordHash string     `json:"-" db:"password_hash" validate:"required"`
	AvatarURL    string     `json:"avatar_url" db:"avatar_url"`
	SystemRole   SystemRole `json:"system_role" db:"system_role" validate:"required,oneof=SUPER_ADMIN USER"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) IsEmpty() bool {
	return u.ID == uuid.Nil
}

func (u *User) IsSuperAdmin() bool {
	return u.SystemRole == SystemRoleSuperAdmin
}

func (u *User) IsRegularUser() bool {
	return u.SystemRole == SystemRoleUser
}
