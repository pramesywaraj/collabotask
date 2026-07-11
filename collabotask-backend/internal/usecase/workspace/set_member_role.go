package workspace

import (
	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/pkg/validator"
	"context"
	"errors"
	"fmt"
)

func (wu *WorkspaceUseCase) SetMemberRole(ctx context.Context, input SetMemberRoleInput) (SetMemberRoleOutput, error) {
	if err := validator.Struct(input); err != nil {
		return SetMemberRoleOutput{}, fmt.Errorf("validation: %w", err)
	}

	requester, err := wu.workspaceMemberRepo.GetByWorkspaceAndUser(ctx, input.WorkspaceID, input.RequesterID)
	if err != nil {
		if errors.Is(err, domain.ErrMemberNotFound) {
			return SetMemberRoleOutput{}, domain.ErrNotWorkspaceAdmin
		}
		return SetMemberRoleOutput{}, fmt.Errorf("failed to get requester: %w", err)
	}
	if !requester.IsAdmin() {
		return SetMemberRoleOutput{}, domain.ErrNotWorkspaceAdmin
	}

	target, err := wu.workspaceMemberRepo.GetByWorkspaceAndUser(ctx, input.WorkspaceID, input.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrMemberNotFound) {
			return SetMemberRoleOutput{}, domain.ErrMemberNotFound
		}
		return SetMemberRoleOutput{}, fmt.Errorf("failed to get target member: %w", err)
	}

	// Idempotent: already has the requested role.
	if target.Role == input.Role {
		return SetMemberRoleOutput{Member: target}, nil
	}

	// Demote guards — only apply when downgrading from ADMIN to MEMBER.
	if target.IsAdmin() && input.Role == entity.WorkspaceRoleMember {
		ws, err := wu.workspaceRepo.GetByID(ctx, input.WorkspaceID)
		if err != nil {
			return SetMemberRoleOutput{}, fmt.Errorf("failed to get workspace: %w", err)
		}

		if ws.OwnerID == input.UserID {
			return SetMemberRoleOutput{}, domain.ErrCannotDemoteOwner
		}

		last, err := wu.isLastAdmin(ctx, input.WorkspaceID)
		if err != nil {
			return SetMemberRoleOutput{}, err
		}
		if last {
			return SetMemberRoleOutput{}, domain.ErrCannotDemoteLastAdmin
		}
	}

	updated, err := wu.workspaceMemberRepo.UpdateRole(ctx, input.WorkspaceID, input.UserID, input.Role)
	if err != nil {
		return SetMemberRoleOutput{}, fmt.Errorf("failed to update member role: %w", err)
	}

	return SetMemberRoleOutput{Member: updated}, nil
}
