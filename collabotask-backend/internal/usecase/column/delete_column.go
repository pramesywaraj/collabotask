package column

import (
	"collabotask/internal/domain"
	"collabotask/internal/infrastructure/validator"
	"context"
	"errors"
	"fmt"
)

func (cu *ColumnUseCaseImpl) DeleteColumn(ctx context.Context, input DeleteColumnInput) error {
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

	_, err = cu.boardAccessChecker.Check(ctx, column.BoardID, input.RequesterID)
	if err != nil {
		return err
	}

	err = cu.columnRepo.Delete(ctx, column.ID)
	if err != nil {
		return err
	}

	return nil
}
