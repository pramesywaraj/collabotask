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

func (wu *WorkspaceUseCase) RemoveMember(ctx context.Context, input RemoveMemberInput) error {
	if err := validator.Struct(input); err != nil {
		return fmt.Errorf("failed to validate input when removing member: %w", err)
	}

	requesterMember, err := wu.workspaceMemberRepo.GetByWorkspaceAndUser(ctx, input.WorkspaceID, input.RequesterID)
	if err != nil {
		if errors.Is(err, domain.ErrMemberNotFound) {
			return domain.ErrNotWorkspaceAdmin
		}
		return fmt.Errorf("failed to get requester: %w", err)
	}
	if !requesterMember.IsAdmin() {
		return domain.ErrNotWorkspaceAdmin
	}
	if input.RequesterID == input.UserID {
		return domain.ErrCannotRemoveYourself
	}

	result, err := wu.workspaceMemberRepo.RemoveWithParticipationCascade(ctx, input.WorkspaceID, input.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrMemberNotFound) {
			return domain.ErrMemberNotFound
		}
		return fmt.Errorf("failed to remove member from the workspace: %w", err)
	}

	for _, boardID := range result.AffectedBoardIDs {
		common.WriteActivity(ctx, wu.activityRepo, input.RequesterID, &entity.Activity{
			BoardID:    boardID,
			ActionType: entity.ActivityActionRemoved,
			EntityType: entity.ActivityEntityMember,
			EntityID:   input.UserID,
			Metadata:   map[string]any{entity.ActivityMetaSource: "workspace"},
		})
	}

	// Involuntary removal → ACCESS_REVOKED per board (UC-06, §4.5 UC-19b).
	wu.fanOutParticipationCascade(ctx, input.WorkspaceID, input.UserID, result, common.EvictReasonRemovedFromWorkspace)

	return nil
}
