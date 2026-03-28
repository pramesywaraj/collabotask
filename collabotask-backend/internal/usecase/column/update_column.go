package column

import (
	"collabotask/internal/domain"
	"collabotask/internal/dto"
	"collabotask/internal/infrastructure/validator"
	"context"
	"errors"
	"fmt"
)

func (cu *ColumnUseCaseImpl) UpdateColumn(ctx context.Context, input UpdateColumnInput) (*UpdateColumnOutput, error) {
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

	_, err = cu.boardAccessChecker.Check(ctx, column.BoardID, input.RequesterID)
	if err != nil {
		return nil, err
	}

	column.Title = input.Title

	err = cu.columnRepo.Update(ctx, column)
	if err != nil {
		return nil, err
	}

	return &UpdateColumnOutput{
		Column: dto.ColumnToDTO(column),
	}, nil
}
