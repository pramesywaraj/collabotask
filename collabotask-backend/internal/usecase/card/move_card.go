package card

import (
	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/dto"
	"collabotask/internal/infrastructure/validator"
	"context"
	"errors"
	"fmt"
)

func (cru *CardUseCaseImpl) MoveCard(ctx context.Context, input MoveCardInput) (*MoveCardOutput, error) {
	if err := validator.Struct(input); err != nil {
		return nil, fmt.Errorf("failed to validate move card input: %w", err)
	}

	card, err := cru.cardRepo.GetByID(ctx, input.CardID)
	if err != nil {
		if errors.Is(err, domain.ErrCardNotFound) {
			return nil, domain.ErrCardNotFound
		}
		return nil, fmt.Errorf("failed to fetch card: %w", err)
	}
	if !card.BelongsToColumn(input.FromColumnID) {
		return nil, domain.ErrCardNotInColumn
	}

	fromColumn, err := cru.columnRepo.GetByID(ctx, input.FromColumnID)
	if err != nil {
		if errors.Is(err, domain.ErrColumnNotFound) {
			return nil, domain.ErrColumnNotFound
		}
		return nil, fmt.Errorf("failed to fetch 'from' column: %w", err)
	}

	toColumn, err := cru.columnRepo.GetByID(ctx, input.ToColumnID)
	if err != nil {
		if errors.Is(err, domain.ErrColumnNotFound) {
			return nil, domain.ErrColumnNotFound
		}
		return nil, fmt.Errorf("failed to fetch 'to' column: %w", err)
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

	var assignee *entity.User
	if movedCard.AssignedTo != nil {
		user, err := cru.userRepo.GetById(ctx, *movedCard.AssignedTo)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch assignee: %w", err)
		}
		if user == nil || user.IsEmpty() {
			return nil, fmt.Errorf("failed to fetch assignee: %w", domain.ErrUserNotFound)
		}

		assignee = user
	}

	return &MoveCardOutput{
		Card: dto.CardWithAssigneeToDTO(movedCard, assignee),
	}, nil
}
