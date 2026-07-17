package workspace

import (
	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/usecase/common"
	"collabotask/pkg/validator"
	"context"
	"errors"
	"fmt"
)

func (wu *WorkspaceUseCase) LeaveWorkspace(ctx context.Context, input LeaveWorkspaceInput) error {
	if err := validator.Struct(input); err != nil {
		return fmt.Errorf("validation: %w", err)
	}

	member, err := wu.workspaceMemberRepo.GetByWorkspaceAndUser(ctx, input.WorkspaceID, input.RequesterID)
	if err != nil {
		if errors.Is(err, domain.ErrMemberNotFound) {
			return domain.ErrMemberNotFound
		}
		return fmt.Errorf("failed to get member: %w", err)
	}

	ws, err := wu.workspaceRepo.GetByID(ctx, input.WorkspaceID)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	if ws.OwnerID == input.RequesterID {
		return domain.ErrWorkspaceOwnerCannotLeave
	}

	if member.Role == entity.WorkspaceRoleAdmin {
		last, err := wu.isLastAdmin(ctx, input.WorkspaceID)
		if err != nil {
			return err
		}
		if last {
			return domain.ErrLastAdminCannotLeave
		}
	}

	result, err := wu.workspaceMemberRepo.RemoveWithParticipationCascade(ctx, input.WorkspaceID, input.RequesterID)
	if err != nil {
		return fmt.Errorf("failed to leave workspace: %w", err)
	}

	requesterID := input.RequesterID
	for _, boardID := range result.AffectedBoardIDs {
		common.WriteActivity(ctx, wu.activityRepo, &entity.Activity{
			BoardID:    boardID,
			UserID:     &requesterID,
			ActionType: entity.ActivityActionLeft,
			EntityType: entity.ActivityEntityMember,
			EntityID:   input.RequesterID,
			Metadata:   map[string]any{"source": "workspace"},
		})
	}

	return nil
}
