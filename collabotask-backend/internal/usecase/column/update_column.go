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

func (cu *ColumnUseCase) UpdateColumn(ctx context.Context, input UpdateColumnInput) (*UpdateColumnOutput, error) {
	if err := validator.Struct(input); err != nil {
		return nil, fmt.Errorf("failed to validate update column input: %w", err)
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

	titleChanged := column.Title != input.Title
	column.Title = input.Title

	err = cu.columnRepo.Update(ctx, column)
	if err != nil {
		return nil, err
	}

	if titleChanged {
		common.WriteActivity(ctx, cu.activityRepo, input.RequesterID, &entity.Activity{
			BoardID:    column.BoardID,
			ActionType: entity.ActivityActionUpdated,
			EntityType: entity.ActivityEntityColumn,
			EntityID:   column.ID,
			Metadata:   map[string]any{entity.ActivityMetaColumnTitle: column.Title},
		})
	}

	return &UpdateColumnOutput{
		Column: column,
	}, nil
}
