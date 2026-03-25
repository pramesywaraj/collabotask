package card

import (
	"collabotask/internal/domain"
	"collabotask/internal/dto"
	"collabotask/internal/infrastructure/validator"
	"context"
	"errors"
	"fmt"
)

func (cru *CardUseCaseImpl) UpdateCard(ctx context.Context, input UpdateCardInput) (*UpdateCardOutput, error) {
	if err := validator.Struct(input); err != nil {
		return nil, fmt.Errorf("failed to validate update card input: %w", err)
	}

	atLeastOne := validator.AtLeastOneProvided(input.Title, input.Description, input.AssignedTo, input.DueDate)
	if !atLeastOne {
		return nil, domain.ErrAtLeastOneProvided
	}

	card, err := cru.cardRepo.GetByID(ctx, input.CardID)
	if err != nil {
		if errors.Is(err, domain.ErrCardNotFound) {
			return nil, domain.ErrCardNotFound
		}
		return nil, fmt.Errorf("failed to fetch card: %w", err)
	}

	column, err := cru.columnRepo.GetByID(ctx, card.ColumnID)
	if err != nil {
		return nil, err
	}

	_, err = cru.boardAccessChecker.Check(ctx, column.BoardID, input.RequesterID)
	if err != nil {
		return nil, err
	}

	if input.Title != nil {
		card.Title = *input.Title
	}
	if input.Description != nil && *input.Description != "" {
		card.Description = input.Description
	}
	if input.AssignedTo != nil {
		card.AssignedTo = input.AssignedTo
	}
	if input.DueDate != nil {
		card.DueDate = input.DueDate
	}

	err = cru.cardRepo.Update(ctx, card)
	if err != nil {
		return nil, err
	}

	return &UpdateCardOutput{
		Card: dto.CardToDTO(card),
	}, nil
}
