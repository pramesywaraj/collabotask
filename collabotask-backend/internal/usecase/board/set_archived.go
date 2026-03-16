package board

import (
	"collabotask/internal/domain"
	"collabotask/internal/dto"
	"collabotask/internal/infrastructure/validator"
	"context"
	"errors"
	"fmt"
)

func (bu *BoardUseCaseImpl) SetArchived(ctx context.Context, input SetArchivedInput) (*SetArchivedOutput, error) {
	if err := validator.Struct(input); err != nil {
		return nil, fmt.Errorf("failed to validate set archived on board input: %w", err)
	}

	board, err := bu.boardRepo.GetByID(ctx, input.BoardID)
	if err != nil {
		if errors.Is(err, domain.ErrBoardNotFound) {
			return nil, domain.ErrBoardNotFound
		}
		return nil, fmt.Errorf("failed to fetch board detail: %w", err)
	}
	if board == nil || board.IsEmpty() {
		return nil, domain.ErrBoardNotFound
	}

	boardMember, err := bu.boardMemberRepo.GetMemberByBoardAndUser(ctx, input.BoardID, input.RequesterID)
	if err != nil {
		if errors.Is(err, domain.ErrBoardMemberNotFound) {
			return nil, domain.ErrBoardMemberNotFound
		}
		return nil, fmt.Errorf("failed to fetch board membership: %w", err)
	}
	if boardMember == nil || !boardMember.IsOwner() {
		return nil, domain.ErrBoardPermissionDenied
	}

	err = bu.boardRepo.SetArchived(ctx, input.BoardID, *input.IsArchived)
	if err != nil {
		return nil, fmt.Errorf("failed to set archived the board: %w", err)
	}

	board.IsArchived = *input.IsArchived

	return &SetArchivedOutput{
		Board: dto.BoardToDTO(board),
	}, nil
}
