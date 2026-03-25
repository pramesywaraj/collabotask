package column

import (
	"collabotask/internal/domain/entity"
	"collabotask/internal/dto"
	"collabotask/internal/infrastructure/validator"
	"context"
	"fmt"
)

func (cu *ColumnUseCaseImpl) CreateColumn(ctx context.Context, input CreateColumnInput) (*CreateColumnOutput, error) {
	if err := validator.Struct(input); err != nil {
		return nil, fmt.Errorf("failed to validate create column input: %w", err)
	}

	board, err := cu.boardAccessChecker.Check(ctx, input.BoardID, input.RequesterID)
	if err != nil {
		return nil, err
	}

	nextPosition, err := cu.columnRepo.GetMaxPosition(ctx, board.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get max position for the column: %w", err)
	}

	column := &entity.Column{
		BoardID:  board.ID,
		Title:    input.Title,
		Position: nextPosition + 1,
	}
	err = cu.columnRepo.Create(ctx, column)
	if err != nil {
		return nil, err
	}

	return &CreateColumnOutput{
		Column: dto.ColumnToDTO(column),
	}, nil
}
