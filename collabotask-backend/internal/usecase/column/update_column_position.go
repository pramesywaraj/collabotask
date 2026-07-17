package column

import (
	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/usecase/common"
	"collabotask/pkg/validator"
	"context"
	"errors"
	"fmt"
)

func (cu *ColumnUseCase) UpdateColumnPosition(ctx context.Context, input UpdateColumnPositionInput) (*UpdateColumnPositionOutput, error) {
	if err := validator.Struct(input); err != nil {
		return nil, fmt.Errorf("failed to validate update column position input: %w", err)
	}

	column, err := cu.columnRepo.GetByID(ctx, input.ColumnID)
	if err != nil {
		if errors.Is(err, domain.ErrColumnNotFound) {
			return nil, domain.ErrColumnNotFound
		}
		return nil, fmt.Errorf("failed to fetch column: %w", err)
	}
	if !column.BelongsToBoard(input.BoardID) {
		return nil, domain.ErrColumnNotInBoard
	}

	_, err = cu.boardAccessChecker.CheckMutateAccess(ctx, column.BoardID, input.RequesterID)
	if err != nil {
		return nil, err
	}

	oldPosition := column.Position
	newPos, err := cu.columnRepo.UpdatePosition(ctx, column.ID, input.Position)
	if err != nil {
		return nil, err
	}

	// Use the repository's returned position, not input.Position: a rebalance may
	// have rewritten this column to a different value within the same transaction.
	column.Position = newPos

	if newPos != oldPosition {
		requesterID := input.RequesterID
		common.WriteActivity(ctx, cu.activityRepo, &entity.Activity{
			BoardID:    column.BoardID,
			UserID:     &requesterID,
			ActionType: entity.ActivityActionMoved,
			EntityType: entity.ActivityEntityColumn,
			EntityID:   column.ID,
			Metadata:   map[string]any{"column_title": column.Title},
		})
	}

	return &UpdateColumnPositionOutput{Column: column}, nil
}
