package repository

import (
	"collabotask/internal/domain/entity"
	"context"

	"github.com/google/uuid"
)

// AffectedCard carries the card and column IDs cleared by a participation cascade.
// Returned even when nothing broadcasts yet so step ④ is a pure add (wire broadcast, no repo change).
type AffectedCard struct {
	CardID   uuid.UUID
	ColumnID uuid.UUID
}

// WorkspaceCascadeResult holds the entities affected by a workspace-member
// participation cascade, returned now so step ④ is a pure wire-up (no repo change needed).
type WorkspaceCascadeResult struct {
	AffectedCards    []AffectedCard
	AffectedBoardIDs []uuid.UUID
}

type WorkspaceMemberRepository interface {
	Create(ctx context.Context, member *entity.WorkspaceMember) error
	Delete(ctx context.Context, workspaceID, userID uuid.UUID) error
	GetByWorkspaceAndUser(ctx context.Context, workspaceID, userID uuid.UUID) (*entity.WorkspaceMember, error)
	GetMembersByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entity.WorkspaceMember, error)
	IsUserExists(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error)
	UpdateRole(ctx context.Context, workspaceID, userID uuid.UUID, role entity.WorkspaceRole) (*entity.WorkspaceMember, error)
	CountAdmins(ctx context.Context, workspaceID uuid.UUID) (int, error)
	// RemoveWithParticipationCascade removes the user from the workspace and all its
	// boards, clears card assignments, and returns the affected cards and board IDs.
	// AffectedBoardIDs includes every board the user was a member of (even those with
	// zero assigned cards) so the caller can write one MEMBER/REMOVED activity row per board.
	RemoveWithParticipationCascade(ctx context.Context, workspaceID, userID uuid.UUID) (WorkspaceCascadeResult, error)
}
