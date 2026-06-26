package entity

import (
	"time"

	"github.com/google/uuid"
)

type WorkspaceRole string

const (
	WorkspaceRoleAdmin  WorkspaceRole = "ADMIN"
	WorkspaceRoleMember WorkspaceRole = "MEMBER"
)

type WorkspaceMember struct {
	WorkspaceID uuid.UUID     `json:"workspace_id" db:"workspace_id"`
	UserID      uuid.UUID     `json:"user_id" db:"user_id"`
	Role        WorkspaceRole `json:"role" db:"role" validate:"required,oneof=ADMIN MEMBER"`
	JoinedAt    time.Time     `json:"joined_at" db:"joined_at"`
}

func (WorkspaceMember) TableName() string {
	return "workspace_members"
}

func (wm *WorkspaceMember) IsEmpty() bool {
	return wm.WorkspaceID == uuid.Nil && wm.UserID == uuid.Nil
}

func (wm *WorkspaceMember) IsAdmin() bool {
	return wm.Role == WorkspaceRoleAdmin
}

func (wm *WorkspaceMember) IsMember() bool {
	return wm.Role == WorkspaceRoleMember
}
