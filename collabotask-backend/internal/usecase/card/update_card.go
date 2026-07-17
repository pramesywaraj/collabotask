package card

import (
	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/usecase/common"
	"collabotask/pkg/validator"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func (cru *CardUseCase) UpdateCard(ctx context.Context, input UpdateCardInput) (*UpdateCardOutput, error) {
	if err := validator.Struct(input); err != nil {
		return nil, fmt.Errorf("failed to validate update card input: %w", err)
	}

	var assignee *entity.User

	atLeastOne := validator.AtLeastOneProvided(input.Title) || input.DescriptionPresent || input.AssignedToPresent || input.DueDatePresent
	if !atLeastOne {
		return nil, domain.ErrAtLeastOneProvided
	}
	if input.AssignedToPresent && input.AssignedTo != nil && *input.AssignedTo == uuid.Nil {
		return nil, domain.ErrInvalidAssigneeID
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

	_, err = cru.boardAccessChecker.CheckMutateAccess(ctx, column.BoardID, input.RequesterID)
	if err != nil {
		return nil, err
	}

	var changedFields []string
	if input.Title != nil && *input.Title != card.Title {
		changedFields = append(changedFields, "title")
	}
	if input.DescriptionPresent {
		changedFields = append(changedFields, "description")
	}
	if input.AssignedToPresent {
		changedFields = append(changedFields, "assigned_to")
	}
	if input.DueDatePresent {
		changedFields = append(changedFields, "due_date")
	}

	if input.Title != nil {
		card.Title = *input.Title
	}
	if input.DescriptionPresent {
		if input.Description == nil {
			card.Description = nil
		} else if strings.TrimSpace(*input.Description) == "" {
			card.Description = nil
		} else {
			s := strings.TrimSpace(*input.Description)
			card.Description = &s
		}
	}
	if input.AssignedToPresent {
		if input.AssignedTo != nil {
			if err := cru.requireBoardMember(ctx, input.BoardID, *input.AssignedTo); err != nil {
				return nil, err
			}
		}
		card.AssignedTo = input.AssignedTo
	}
	if input.DueDatePresent {
		card.DueDate = input.DueDate
	}

	// Resolve the assignee for the response payload only. The membership gate
	// above already validated a newly-set assignee; this block is display-only
	// so a stale existing assignee never blocks an unrelated edit (Decision B).
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

	err = cru.cardRepo.Update(ctx, card)
	if err != nil {
		return nil, err
	}

	if len(changedFields) > 0 {
		requesterID := input.RequesterID
		common.WriteActivity(ctx, cru.activityRepo, &entity.Activity{
			BoardID:    column.BoardID,
			UserID:     &requesterID,
			ActionType: entity.ActivityActionUpdated,
			EntityType: entity.ActivityEntityCard,
			EntityID:   input.CardID,
			Metadata:   map[string]any{"card_title": card.Title, "changed_fields": changedFields},
		})
	}

	return &UpdateCardOutput{
		Card:     card,
		Assignee: assignee,
	}, nil
}
