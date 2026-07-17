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

func (bu *BoardUseCase) RemoveMember(ctx context.Context, input RemoveMemberInput) error {
	if err := validator.Struct(input); err != nil {
		return fmt.Errorf("failed to validate remove member input: %w", err)
	}

	if input.RequesterID == input.UserID {
		return domain.ErrCannotRemoveYourself
	}

	board, err := bu.boardRepo.GetByID(ctx, input.BoardID)
	if err != nil {
		if errors.Is(err, domain.ErrBoardNotFound) {
			return domain.ErrBoardNotFound
		}
		return fmt.Errorf("failed to verify board existence: %w", err)
	}
	if board == nil || board.IsEmpty() || board.IsArchived || board.WorkspaceID != input.WorkspaceID {
		return domain.ErrBoardNotFound
	}

	workspaceMember, err := bu.workspaceMemberRepo.GetByWorkspaceAndUser(ctx, input.WorkspaceID, input.RequesterID)
	if err != nil || workspaceMember == nil || workspaceMember.IsEmpty() {
		return domain.ErrUserNotInWorkspace
	}

	boardMember, err := bu.boardMemberRepo.GetMemberByBoardAndUser(ctx, input.BoardID, input.RequesterID)
	if err != nil && !errors.Is(err, domain.ErrBoardMemberNotFound) {
		return fmt.Errorf("failed to check requester permission: %w", err)
	}

	canManage := canAdministerBoard(boardMember, workspaceMember)
	if !canManage {
		return domain.ErrBoardPermissionDenied
	}

	// TODO(ws, step ④): broadcast CARD_UPDATED for the returned affected cards.
	_, err = bu.boardMemberRepo.RemoveWithParticipationCascade(ctx, input.BoardID, input.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrBoardMemberNotFound) {
			return domain.ErrBoardMemberNotFound
		}
		return fmt.Errorf("failed to remove member from board: %w", err)
	}

	common.WriteActivity(ctx, bu.activityRepo, input.RequesterID, &entity.Activity{
		BoardID:    input.BoardID,
		ActionType: entity.ActivityActionRemoved,
		EntityType: entity.ActivityEntityMember,
		EntityID:   input.UserID,
		Metadata:   map[string]any{entity.ActivityMetaSource: "board"},
	})

	return nil
}
