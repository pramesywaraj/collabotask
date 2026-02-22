package repository

import (
	"collabotask/internal/domain/entity"
	"context"

	"github.com/google/uuid"
)

type WorkspaceMemberRepository interface {
	Create(ctx context.Context, member *entity.WorkspaceMember) error
	Delete(ctx context.Context, workspaceID, userID uuid.UUID) error

	GetByWorkspaceAndUser(ctx context.Context, workspaceID, userID uuid.UUID) (*entity.WorkspaceMember, error)
	ListMemberByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entity.WorkspaceMember, error)

	IsUserExists(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error)
}
