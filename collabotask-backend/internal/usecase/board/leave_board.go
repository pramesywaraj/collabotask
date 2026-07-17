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

func (bu *BoardUseCase) LeaveBoard(ctx context.Context, input LeaveBoardInput) error {
	if err := validator.Struct(input); err != nil {
		return fmt.Errorf("failed to validate leave board input: %w", err)
	}

	board, err := bu.boardRepo.GetByID(ctx, input.BoardID)
	if err != nil {
		if errors.Is(err, domain.ErrBoardNotFound) {
			return domain.ErrBoardNotFound
		}
		return fmt.Errorf("failed to verify board existence: %w", err)
	}
	if board == nil || board.IsEmpty() || board.IsArchived {
		return domain.ErrBoardNotFound
	}

	boardMember, err := bu.boardMemberRepo.GetMemberByBoardAndUser(ctx, input.BoardID, input.RequesterID)
	if err != nil {
		if errors.Is(err, domain.ErrBoardMemberNotFound) {
			return domain.ErrBoardMemberNotFound
		}
		return fmt.Errorf("failed to check requester membership in board: %w", err)
	}
	if boardMember.IsEmpty() {
		return domain.ErrBoardMemberNotFound
	}

	if boardMember.IsOwner() {
		return domain.ErrBoardOwnerCannotLeave
	}

	// TODO(ws, step ④): broadcast CARD_UPDATED for the returned affected cards.
	_, err = bu.boardMemberRepo.RemoveWithParticipationCascade(ctx, input.BoardID, input.RequesterID)
	if err != nil {
		if errors.Is(err, domain.ErrBoardMemberNotFound) {
			return domain.ErrBoardMemberNotFound
		}
		return fmt.Errorf("failed to remove member from board: %w", err)
	}

	requesterID := input.RequesterID
	common.WriteActivity(ctx, bu.activityRepo, &entity.Activity{
		BoardID:    input.BoardID,
		UserID:     &requesterID,
		ActionType: entity.ActivityActionLeft,
		EntityType: entity.ActivityEntityMember,
		EntityID:   input.RequesterID,
		Metadata:   map[string]any{"source": "board"},
	})

	return nil
}
