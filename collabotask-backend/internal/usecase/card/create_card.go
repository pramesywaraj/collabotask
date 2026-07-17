package card

import (
	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/usecase/common"
	"collabotask/pkg/validator"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

func (cru *CardUseCase) CreateCard(ctx context.Context, input CreateCardInput) (*CreateCardOutput, error) {
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
	if !column.BelongsToBoard(input.BoardID) {
		return nil, domain.ErrColumnNotInBoard
	}

	_, err = cru.boardAccessChecker.CheckMutateAccess(ctx, column.BoardID, input.RequesterID)
	if err != nil {
		return nil, err
	}

	var assignee *entity.User
	if input.AssignedTo != nil {
		if *input.AssignedTo == uuid.Nil {
			return nil, domain.ErrInvalidAssigneeID
		}
		if err := cru.requireBoardMember(ctx, input.BoardID, *input.AssignedTo); err != nil {
			return nil, err
		}
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

	card := &entity.Card{
		ColumnID:    column.ID,
		Title:       input.Title,
		Description: input.Description,
		Position:    maxPos + domain.PositionStep,
		AssignedTo:  input.AssignedTo,
		DueDate:     input.DueDate,
		CreatedBy:   input.RequesterID,
	}
	err = cru.cardRepo.Create(ctx, card)
	if err != nil {
		return nil, fmt.Errorf("failed to create card: %w", err)
	}

	common.WriteActivity(ctx, cru.activityRepo, input.RequesterID, &entity.Activity{
		BoardID:    column.BoardID,
		ActionType: entity.ActivityActionCreated,
		EntityType: entity.ActivityEntityCard,
		EntityID:   card.ID,
		Metadata:   map[string]any{entity.ActivityMetaCardTitle: card.Title},
	})

	return &CreateCardOutput{
		Card:     card,
		Assignee: assignee,
	}, nil
}
