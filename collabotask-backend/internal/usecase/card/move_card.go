package card

import (
	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/usecase/common"
	"collabotask/pkg/validator"
	"context"
	"errors"
	"fmt"
)

func (cru *CardUseCase) MoveCard(ctx context.Context, input MoveCardInput) (*MoveCardOutput, error) {
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
	if !fromColumn.BelongsToBoard(input.BoardID) {
		return nil, domain.ErrColumnNotInBoard
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

	_, err = cru.boardAccessChecker.CheckMutateAccess(ctx, fromColumn.BoardID, input.RequesterID)
	if err != nil {
		return nil, err
	}

	oldColumnID := card.ColumnID
	oldPosition := card.Position

	movedCard, err := cru.cardRepo.Move(ctx, input.CardID, input.FromColumnID, input.ToColumnID, input.ToPosition)
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

	if movedCard.ColumnID != oldColumnID || movedCard.Position != oldPosition {
		common.WriteActivity(ctx, cru.activityRepo, input.RequesterID, &entity.Activity{
			BoardID:    fromColumn.BoardID,
			ActionType: entity.ActivityActionMoved,
			EntityType: entity.ActivityEntityCard,
			EntityID:   input.CardID,
			Metadata: map[string]any{
				entity.ActivityMetaCardTitle:       movedCard.Title,
				entity.ActivityMetaFromColumnID:    oldColumnID.String(),
				entity.ActivityMetaFromColumnTitle: fromColumn.Title,
				entity.ActivityMetaToColumnID:      movedCard.ColumnID.String(),
				entity.ActivityMetaToColumnTitle:   toColumn.Title,
			},
		})

		cru.broadcaster.Broadcast(fromColumn.BoardID, common.CardMoved{
			CardID:       input.CardID,
			FromColumnID: oldColumnID,
			ToColumnID:   movedCard.ColumnID,
			Position:     movedCard.Position,
		})
	}

	return &MoveCardOutput{
		Card:     movedCard,
		Assignee: assignee,
	}, nil
}
