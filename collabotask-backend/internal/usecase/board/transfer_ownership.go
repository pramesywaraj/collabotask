package board

import (
	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/usecase/common"
	"collabotask/pkg/validator"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
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

	var fromUserID *uuid.UUID
	if err := bu.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		var txErr error
		fromUserID, txErr = bu.boardMemberRepo.TransferOwnership(ctx, input.BoardID, input.ToUserID)
		if txErr != nil {
			return fmt.Errorf("failed to transfer ownership: %w", txErr)
		}
		meta := map[string]any{entity.ActivityMetaToUserID: input.ToUserID.String()}
		if fromUserID != nil {
			meta[entity.ActivityMetaFromUserID] = fromUserID.String()
		}
		return common.LogActivity(ctx, bu.activityRepo, input.RequesterID, &entity.Activity{
			BoardID:    input.BoardID,
			ActionType: entity.ActivityActionOwnershipTransferred,
			EntityType: entity.ActivityEntityBoard,
			EntityID:   input.BoardID,
			Metadata:   meta,
		})
	}); err != nil {
		return err
	}

	bu.broadcaster.Broadcast(input.BoardID, common.OwnershipTransferred{
		BoardID:    input.BoardID,
		FromUserID: fromUserID,
		ToUserID:   input.ToUserID,
	})

	return nil
}
