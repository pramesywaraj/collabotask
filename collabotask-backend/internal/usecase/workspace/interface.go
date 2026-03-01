package workspace

import (
	"collabotask/internal/domain/entity"
	"context"
	"time"

	"github.com/google/uuid"
)

type WorkspaceDTO struct {
	ID          uuid.UUID
	Name        string
	Description *string
	OwnerID     uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type WorkspaceWithMetaDTO struct {
	WorkspaceDTO

	MemberCount uint
	BoardCount  uint
	Role        entity.WorkspaceRole
}

type WorkspaceMemberDTO struct {
	UserID    uuid.UUID
	Email     string
	Name      string
	AvatarURL *string
	Role      entity.WorkspaceRole
	JoinedAt  time.Time
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
	Message string
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
