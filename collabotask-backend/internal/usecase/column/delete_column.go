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

func (cu *ColumnUseCase) DeleteColumn(ctx context.Context, input DeleteColumnInput) error {
	if err := validator.Struct(input); err != nil {
		return fmt.Errorf("failed to validate delete column input: %w", err)
	}

	column, err := cu.columnRepo.GetByID(ctx, input.ColumnID)
	if err != nil {
		if errors.Is(err, domain.ErrColumnNotFound) {
			return domain.ErrColumnNotFound
		}
		return fmt.Errorf("failed to fetch column: %w", err)
	}
	if !column.BelongsToBoard(input.BoardID) {
		return domain.ErrColumnNotInBoard
	}

	_, err = cu.boardAccessChecker.CheckMutateAccess(ctx, column.BoardID, input.RequesterID)
	if err != nil {
		return err
	}

	colTitle := column.Title
	if err = cu.columnRepo.Delete(ctx, column.ID); err != nil {
		return err
	}

	common.WriteActivity(ctx, cu.activityRepo, input.RequesterID, &entity.Activity{
		BoardID:    column.BoardID,
		ActionType: entity.ActivityActionDeleted,
		EntityType: entity.ActivityEntityColumn,
		EntityID:   column.ID,
		Metadata:   map[string]any{entity.ActivityMetaColumnTitle: colTitle},
	})
	return nil
}
