package board

import (
	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/usecase/common"
	"collabotask/pkg/validator"
	"context"
	"errors"
	"fmt"
)

func (bu *BoardUseCase) TransferOwnership(ctx context.Context, input TransferOwnershipInput) error {
	if err := validator.Struct(input); err != nil {
		return fmt.Errorf("failed to validate transfer ownership input: %w", err)
	}

	access, err := bu.boardAccessChecker.CheckMutateAccess(ctx, input.BoardID, input.RequesterID)
	if err != nil {
		return err
	}

	if access.Board.WorkspaceID != input.WorkspaceID {
		return domain.ErrBoardNotFound
	}

	if !canAdministerBoard(access.BoardMember, access.WorkspaceMember) {
		return domain.ErrBoardPermissionDenied
	}

	targetBM, err := bu.boardMemberRepo.GetMemberByBoardAndUser(ctx, input.BoardID, input.ToUserID)
	if err != nil {
		if errors.Is(err, domain.ErrBoardMemberNotFound) {
			return domain.ErrTransferTargetNotBoardMember
		}
		return fmt.Errorf("failed to fetch target member: %w", err)
	}

	if targetBM.IsOwner() {
		return nil
	}

	// TODO(ws, step ④): broadcast OWNERSHIP_TRANSFERRED { board_id, from_user_id, to_user_id } via WebSocket.
	fromUserID, err := bu.boardMemberRepo.TransferOwnership(ctx, input.BoardID, input.ToUserID)
	if err != nil {
		return fmt.Errorf("failed to transfer ownership: %w", err)
	}

	requesterID := input.RequesterID
	meta := map[string]any{"to_user_id": input.ToUserID.String()}
	if fromUserID != nil {
		meta["from_user_id"] = fromUserID.String()
	}
	common.WriteActivity(ctx, bu.activityRepo, &entity.Activity{
		BoardID:    input.BoardID,
		UserID:     &requesterID,
		ActionType: entity.ActivityActionOwnershipTransferred,
		EntityType: entity.ActivityEntityBoard,
		EntityID:   input.BoardID,
		Metadata:   meta,
	})

	return nil
}
