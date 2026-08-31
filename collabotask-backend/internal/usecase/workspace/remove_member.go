package workspace

import (
	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/domain/repository"
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

	var result repository.WorkspaceCascadeResult
	if err := wu.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		var txErr error
		result, txErr = wu.workspaceMemberRepo.RemoveWithParticipationCascade(ctx, input.WorkspaceID, input.UserID)
		if txErr != nil {
			if errors.Is(txErr, domain.ErrMemberNotFound) {
				return domain.ErrMemberNotFound
			}
			return fmt.Errorf("failed to remove member from the workspace: %w", txErr)
		}
		for _, boardID := range result.AffectedBoardIDs {
			if err := common.LogActivity(ctx, wu.activityRepo, input.RequesterID, &entity.Activity{
				BoardID:    boardID,
				ActionType: entity.ActivityActionRemoved,
				EntityType: entity.ActivityEntityMember,
				EntityID:   input.UserID,
				Metadata:   map[string]any{entity.ActivityMetaSource: "workspace"},
			}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	// Involuntary removal → ACCESS_REVOKED per board (UC-06, §4.5 UC-19b).
	wu.fanOutParticipationCascade(ctx, input.WorkspaceID, input.UserID, result, common.EvictReasonRemovedFromWorkspace)

	return nil
}
