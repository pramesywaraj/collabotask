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

func (cru *CardUseCaseImpl) UpdateCard(ctx context.Context, input UpdateCardInput) (*UpdateCardOutput, error) {
	if err := validator.Struct(input); err != nil {
		return nil, fmt.Errorf("failed to validate update card input: %w", err)
	}

	var assignee *entity.User

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
	if !card.BelongsToColumn(input.ColumnID) {
		return nil, domain.ErrCardNotInColumn
	}

	column, err := cru.columnRepo.GetByID(ctx, card.ColumnID)
	if err != nil {
		if errors.Is(err, domain.ErrColumnNotFound) {
			return nil, domain.ErrColumnNotFound
		}
		return nil, fmt.Errorf("failed to fetch column: %w", err)
	}
	if !column.BelongsToBoard(input.BoardID) {
		return nil, domain.ErrColumnNotInBoard
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

	if card.AssignedTo != nil {
		user, err := cru.userRepo.GetById(ctx, *card.AssignedTo)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch assignee: %w", err)
		}
		if user == nil || user.IsEmpty() {
			return nil, fmt.Errorf("failed to fetch assignee: %w", domain.ErrUserNotFound)
		}

		assignee = user
	}

	return &UpdateCardOutput{
		Card: dto.CardWithAssigneeToDTO(card, assignee),
	}, nil
}
