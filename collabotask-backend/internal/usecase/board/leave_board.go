package board

import (
	"collabotask/internal/domain"
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

	// DEFERRED (UC-12e, ownership transfer): this owner guard uses created_by as
	// a proxy for the current owner. board_members.role becomes the authoritative
	// owner check then. Correct until UC-12e — created_by == owner holds in
	// Phase 1 (no transfer yet). Not migrated with board visibility (ADR-005),
	// which only removed created_by from access-grant / role-display logic.
	if board.CreatedBy == input.RequesterID {
		return domain.ErrBoardOwnerCannotLeave
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

	err = bu.boardMemberRepo.Delete(ctx, input.BoardID, input.RequesterID)
	if err != nil {
		if errors.Is(err, domain.ErrBoardMemberNotFound) {
			return domain.ErrBoardMemberNotFound
		}
		return fmt.Errorf("failed to remove member from board: %w", err)
	}

	return nil
}
