package card

import (
	"collabotask/internal/domain"
	"collabotask/internal/dto"
	"collabotask/internal/infrastructure/validator"
	"context"
	"fmt"
)

func (cru *CardUseCaseImpl) MoveCard(ctx context.Context, input MoveCardInput) (*MoveCardOutput, error) {
	if err := validator.Struct(input); err != nil {
		return nil, fmt.Errorf("failed to validate move card input: %w", err)
	}

	fromColumn, err := cru.columnRepo.GetByID(ctx, input.FromColumnID)
	if err != nil {
		return nil, err
	}

	toColumn, err := cru.columnRepo.GetByID(ctx, input.ToColumnID)
	if err != nil {
		return nil, err
	}

	if fromColumn.BoardID != toColumn.BoardID {
		return nil, domain.ErrInconsistentState
	}

	_, err = cru.boardAccessChecker.Check(ctx, fromColumn.BoardID, input.RequesterID)
	if err != nil {
		return nil, err
	}

	max, err := cru.cardRepo.GetMaxPosition(ctx, input.ToColumnID)
	if err != nil {
		return nil, err
	}

	newPos := input.ToPosition
	if newPos > max+1 {
		newPos = max + 1
	}

	movedCard, err := cru.cardRepo.Move(ctx, input.CardID, input.FromColumnID, input.ToColumnID, newPos)
	if err != nil {
		return nil, err
	}

	return &MoveCardOutput{
		Card: dto.CardToDTO(movedCard),
	}, nil
}
