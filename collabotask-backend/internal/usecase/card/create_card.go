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

func (cru *CardUseCaseImpl) CreateCard(ctx context.Context, input CreateCardInput) (*CreateCardOutput, error) {
	if err := validator.Struct(input); err != nil {
		return nil, fmt.Errorf("failed to validate create card input: %w", err)
	}

	column, err := cru.columnRepo.GetByID(ctx, input.ColumnID)
	if err != nil {
		if errors.Is(err, domain.ErrColumnNotFound) {
			return nil, domain.ErrColumnNotFound
		}
		return nil, fmt.Errorf("failed to fetch column: %w", err)
	}

	_, err = cru.boardAccessChecker.Check(ctx, column.BoardID, input.RequesterID)
	if err != nil {
		return nil, err
	}

	var assignee *entity.User
	if input.AssignedTo != nil {
		user, err := cru.userRepo.GetById(ctx, *input.AssignedTo)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch assignee: %w", err)
		}
		if user == nil || user.IsEmpty() {
			return nil, fmt.Errorf("failed to fetch assignee: %w", domain.ErrUserNotFound)
		}

		assignee = user
	}

	maxPos, err := cru.cardRepo.GetMaxPosition(ctx, input.ColumnID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cards max position in the column: %w", err)
	}

	nextPos := maxPos + 1
	card := &entity.Card{
		ColumnID:    column.ID,
		Title:       input.Title,
		Description: input.Description,
		Position:    nextPos,
		AssignedTo:  input.AssignedTo,
		DueDate:     input.DueDate,
		CreatedBy:   input.RequesterID,
	}
	err = cru.cardRepo.Create(ctx, card)
	if err != nil {
		return nil, fmt.Errorf("failed to create card: %w", err)
	}

	return &CreateCardOutput{
		Card: dto.CardWithAssigneeToDTO(card, assignee),
	}, nil
}
