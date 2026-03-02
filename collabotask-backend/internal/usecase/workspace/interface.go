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

type WorkspaceDetailDTO struct {
	WorkspaceDTO

	UserRole entity.WorkspaceRole
	Members  []WorkspaceMemberDTO
}

type CreateWorkspaceInput struct {
	OwnerID     uuid.UUID `validate:"required"`
	Name        string    `validate:"required,min=2,max=255"`
	Description *string   `validate:"omitempty,max=1000"`
}

type CreateWorkspaceOutput struct {
	Workspace WorkspaceDTO
}

type WorkspaceDetailInput struct {
	RequesterID uuid.UUID `validate:"required"`
	WorkspaceID uuid.UUID `validate:"required"`
}

type WorkspaceDetailOutput struct {
	Workspace WorkspaceDetailDTO
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
	RequesterID uuid.UUID `validate:"required"`
	WorkspaceID uuid.UUID `validate:"required"`
	UserID      uuid.UUID `validate:"required"`
}

type WorkspaceUseCase interface {
	CreateWorkspace(ctx context.Context, input CreateWorkspaceInput) (*CreateWorkspaceOutput, error)
	WorkspaceDetail(ctx context.Context, input WorkspaceDetailInput) (*WorkspaceDetailOutput, error)
	InviteMember(ctx context.Context, input InviteMemberInput) (*InviteMemberOutput, error)
	ListWorkspaces(ctx context.Context, input ListWorkspacesInput) (*ListWorkspacesOutput, error)
	RemoveMember(ctx context.Context, input RemoveMemberInput) error
}
