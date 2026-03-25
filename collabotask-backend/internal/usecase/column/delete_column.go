package column

import (
	"collabotask/internal/infrastructure/validator"
	"context"
	"fmt"
)

func (cu *ColumnUseCaseImpl) DeleteColumn(ctx context.Context, input DeleteColumnInput) error {
	if err := validator.Struct(input); err != nil {
		return fmt.Errorf("failed to validate delete column input: %w", err)
	}

	column, err := cu.columnRepo.GetByID(ctx, input.ColumnID)
	if err != nil {
		return err
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
