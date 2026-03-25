package card

import (
	"collabotask/internal/domain/entity"
	"collabotask/internal/dto"
	"collabotask/internal/infrastructure/validator"
	"context"
	"fmt"
)

func (cru *CardUseCaseImpl) CreateCard(ctx context.Context, input CreateCardInput) (*CreateCardOutput, error) {
	if err := validator.Struct(input); err != nil {
		return nil, fmt.Errorf("failed to validate create card input: %w", err)
	}

	column, err := cru.columnRepo.GetByID(ctx, input.ColumnID)
	if err != nil {
		return nil, err
	}

	_, err = cru.boardAccessChecker.Check(ctx, column.BoardID, input.RequesterID)
	if err != nil {
		return nil, err
	}

	maxPos, err := cru.cardRepo.GetMaxPosition(ctx, input.ColumnID)
	if err != nil {
		return nil, err
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
		Card: dto.CardToDTO(card),
	}, nil
}
