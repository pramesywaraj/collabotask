package board

import (
	"collabotask/internal/domain"
	"collabotask/internal/infrastructure/validator"
	"context"
	"errors"
	"fmt"
)

func (bu *BoardUseCaseImpl) LeaveBoard(ctx context.Context, input LeaveBoardInput) error {
	if err := validator.Struct(input); err != nil {
		return fmt.Errorf("failed to validate leave board input: %w", err)
	}

	board, err := bu.boardRepo.GetByID(ctx, input.BoardID)
	if err != nil {
		if errors.Is(err, domain.ErrBoardNotFound) {
			return domain.ErrBoardNotFound
		}
		return fmt.Errorf("failed to fetch board detail: %w", err)
	}
	if board == nil || board.IsEmpty() || board.IsArchived {
		return domain.ErrBoardNotFound
	}

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
