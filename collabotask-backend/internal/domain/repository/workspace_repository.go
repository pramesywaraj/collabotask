package repository

import (
	"collabotask/internal/domain/entity"
	"context"

	"github.com/google/uuid"
)

type WorkspaceRepository interface {
	Create(ctx context.Context, workspace *entity.Workspace) error
	Update(ctx context.Context, workspace *entity.Workspace) error
	Delete(ctx context.Context, workspaceID uuid.UUID) error

	GetByID(ctx context.Context, workspaceID uuid.UUID) (*entity.Workspace, error)
	GetUserWorkspaces(ctx context.Context, userID uuid.UUID) ([]*entity.Workspace, error)
}
