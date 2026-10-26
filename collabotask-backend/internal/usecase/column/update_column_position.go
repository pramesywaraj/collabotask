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
	var newPos float64
	if err := cu.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		var txErr error
		newPos, txErr = cu.columnRepo.UpdatePosition(ctx, column.ID, input.Position)
		if txErr != nil {
			return txErr
		}
		if newPos != oldPosition {
			return common.LogActivity(ctx, cu.activityRepo, input.RequesterID, &entity.Activity{
				BoardID:    column.BoardID,
				ActionType: entity.ActivityActionMoved,
				EntityType: entity.ActivityEntityColumn,
				EntityID:   column.ID,
				Metadata:   map[string]any{entity.ActivityMetaColumnTitle: column.Title},
			})
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// Use the repository's returned position, not input.Position: a rebalance may
	// have rewritten this column to a different value within the same transaction.
	column.Position = newPos

	if newPos != oldPosition {
		cu.broadcaster.Broadcast(column.BoardID, common.ColumnMoved{
			ColumnID: column.ID,
			Position: newPos,
		})
	}

	return &UpdateColumnPositionOutput{Column: column}, nil
}
