package board

import (
	"collabotask/internal/domain"
	"collabotask/internal/dto"
	"collabotask/internal/infrastructure/validator"
	"context"
	"errors"
	"fmt"
)

func (bu *BoardUseCaseImpl) UpdateBoard(ctx context.Context, input UpdateBoardInput) (*UpdateBoardOutput, error) {
	if err := validator.Struct(input); err != nil {
		return nil, fmt.Errorf("failed to validate update board input: %w", err)
	}

	atLeastOne := validator.AtLeastOneProvided(input.Title, input.Description, input.BackgroundColor)
	if !atLeastOne {
		return nil, domain.ErrAtLeastOneProvided
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

	if input.Title != nil {
		board.Title = *input.Title
	}
	if input.Description != nil {
		board.Description = input.Description
	}
	if input.BackgroundColor != nil {
		board.BackgroundColor = *input.BackgroundColor
	}

	err = bu.boardRepo.Update(ctx, board)
	if err != nil {
		return nil, fmt.Errorf("failed to update the board: %w", err)
	}

	return &UpdateBoardOutput{
		Board: dto.BoardToDTO(board),
	}, nil
}
