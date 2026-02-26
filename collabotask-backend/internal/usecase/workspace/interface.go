package workspace

import (
	"collabotask/internal/domain/entity"
	"context"
	"time"

	"github.com/google/uuid"
)

type WorkspaceDTO struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	OwnerID     uuid.UUID `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type WorkspaceWithMetaDTO struct {
	WorkspaceDTO

	MemberCount uint                 `json:"member_count"`
	BoardCount  uint                 `json:"board_count"`
	Role        entity.WorkspaceRole `json:"role"`
}

type WorkspaceMemberDTO struct {
	UserID    uuid.UUID            `json:"id"`
	Email     string               `json:"email"`
	Name      string               `json:"name"`
	AvatarURL *string              `json:"avatar_url,omitempty"`
	Role      entity.WorkspaceRole `json:"role"`
	JoinedAt  time.Time            `json:"joined_at"`
}

type CreateWorkspaceInput struct {
	OwnerID     uuid.UUID `validate:"required"`
	Name        string    `validate:"required,min=2,max=255"`
	Description *string   `validate:"omitempty,max=1000"`
}

type CreateWorkspaceOutput struct {
	Workspace WorkspaceDTO
}

type InviteMemberInput struct {
	RequesterID uuid.UUID `validate:"required"`
	WorkspaceID uuid.UUID `validate:"required"`
	Emails      []string  `validate:"required,min=1,dive,email"`
}

type InviteMemberOutput struct {
	Message string `json:"message"`
}

type ListWorkspacesInput struct {
	UserID uuid.UUID
}

type ListWorkspacesOutput struct {
	Workspaces []WorkspaceWithMetaDTO
}

type RemoveMemberInput struct {
	RequesterID uuid.UUID
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
}

type WorkspaceUseCase interface {
	CreateWorkspace(ctx context.Context, input CreateWorkspaceInput) (*CreateWorkspaceOutput, error)
	InviteMember(ctx context.Context, input InviteMemberInput) (*InviteMemberOutput, error)
	ListWorkspaces(ctx context.Context, input ListWorkspacesInput) (*ListWorkspacesOutput, error)
	RemoveMember(ctx context.Context, input RemoveMemberInput) error
}
