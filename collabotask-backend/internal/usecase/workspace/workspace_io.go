package workspace

import (
	"collabotask/internal/domain/entity"
	"time"

	"github.com/google/uuid"
)

// Result types for the workspace use cases. Plain workspaces are returned as
// *entity.Workspace directly; the types below carry genuine transformations
// (joined/aggregated data) and live next to the use case that produces them.

type WorkspaceWithMeta struct {
	Workspace   *entity.Workspace
	MemberCount uint
	BoardCount  uint
	Role        entity.WorkspaceRole
}

type WorkspaceMember struct {
	UserID    uuid.UUID
	Email     string
	Name      string
	AvatarURL *string
	Role      entity.WorkspaceRole
	JoinedAt  time.Time
}

type WorkspaceDetail struct {
	Workspace *entity.Workspace
	UserRole  entity.WorkspaceRole
	Members   []WorkspaceMember
}

type CreateWorkspaceInput struct {
	OwnerID     uuid.UUID `validate:"required"`
	Name        string    `validate:"required,min=2,max=255"`
	Description *string   `validate:"omitempty,max=1000"`
}

type CreateWorkspaceOutput struct {
	Workspace *entity.Workspace
}

type GetWorkspaceDetailInput struct {
	RequesterID uuid.UUID `validate:"required"`
	WorkspaceID uuid.UUID `validate:"required"`
}

type GetWorkspaceDetailOutput struct {
	Workspace WorkspaceDetail
}

type InviteResult struct {
	Email     string
	Success   bool
	ErrorCode string // "" | "NOT_FOUND" | "CONFLICT"
}

type InviteMemberInput struct {
	RequesterID uuid.UUID `validate:"required"`
	WorkspaceID uuid.UUID `validate:"required"`
	Emails      []string  `validate:"required,min=1,dive,email"`
}

type InviteMemberOutput struct {
	Results []InviteResult
}

type GetWorkspacesInput struct {
	UserID uuid.UUID
}

type GetWorkspacesOutput struct {
	Workspaces []WorkspaceWithMeta
}

type RemoveMemberInput struct {
	RequesterID uuid.UUID `validate:"required"`
	WorkspaceID uuid.UUID `validate:"required"`
	UserID      uuid.UUID `validate:"required"`
}

type SetMemberRoleInput struct {
	RequesterID uuid.UUID            `validate:"required"`
	WorkspaceID uuid.UUID            `validate:"required"`
	UserID      uuid.UUID            `validate:"required"`
	Role        entity.WorkspaceRole `validate:"required,oneof=ADMIN MEMBER"`
}

type SetMemberRoleOutput struct {
	Member *entity.WorkspaceMember
}

type LeaveWorkspaceInput struct {
	RequesterID uuid.UUID `validate:"required"`
	WorkspaceID uuid.UUID `validate:"required"`
}

type DeleteWorkspaceInput struct {
	RequesterID uuid.UUID `validate:"required"`
	WorkspaceID uuid.UUID `validate:"required"`
}
