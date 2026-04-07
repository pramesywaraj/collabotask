package workspace

import (
	"collabotask/internal/dto"
	"context"

	"github.com/google/uuid"
)

type WorkspaceUseCase interface {
	CreateWorkspace(ctx context.Context, input CreateWorkspaceInput) (*CreateWorkspaceOutput, error)
	GetWorkspaceDetail(ctx context.Context, input GetWorkspaceDetailInput) (*GetWorkspaceDetailOutput, error)
	InviteMember(ctx context.Context, input InviteMemberInput) (*InviteMemberOutput, error)
	GetWorkspaces(ctx context.Context, input GetWorkspacesInput) (*GetWorkspacesOutput, error)
	RemoveMember(ctx context.Context, input RemoveMemberInput) error
}

type CreateWorkspaceInput struct {
	OwnerID     uuid.UUID `validate:"required"`
	Name        string    `validate:"required,min=2,max=255"`
	Description *string   `validate:"omitempty,max=1000"`
}

type CreateWorkspaceOutput struct {
	Workspace dto.WorkspaceDTO
}

type GetWorkspaceDetailInput struct {
	RequesterID uuid.UUID `validate:"required"`
	WorkspaceID uuid.UUID `validate:"required"`
}

type GetWorkspaceDetailOutput struct {
	Workspace dto.WorkspaceDetailDTO
}

type InviteMemberInput struct {
	RequesterID uuid.UUID `validate:"required"`
	WorkspaceID uuid.UUID `validate:"required"`
	Emails      []string  `validate:"required,min=1,dive,email"`
}

type InviteMemberOutput struct {
	Message string
}

type GetWorkspacesInput struct {
	UserID uuid.UUID
}

type GetWorkspacesOutput struct {
	Workspaces []dto.WorkspaceWithMetaDTO
}

type RemoveMemberInput struct {
	RequesterID uuid.UUID `validate:"required"`
	WorkspaceID uuid.UUID `validate:"required"`
	UserID      uuid.UUID `validate:"required"`
}
