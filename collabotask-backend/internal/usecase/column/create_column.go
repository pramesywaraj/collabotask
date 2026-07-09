package column

import (
	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/pkg/validator"
	"context"
	"fmt"
)

func (cu *ColumnUseCase) CreateColumn(ctx context.Context, input CreateColumnInput) (*CreateColumnOutput, error) {
	if err := validator.Struct(input); err != nil {
		return nil, fmt.Errorf("failed to validate create column input: %w", err)
	}

	board, err := cu.boardAccessChecker.Check(ctx, input.BoardID, input.RequesterID)
	if err != nil {
		return nil, err
	}

	maxPos, err := cu.columnRepo.GetMaxPosition(ctx, board.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get max position for the column: %w", err)
	}

	column := &entity.Column{
		BoardID:  board.ID,
		Title:    input.Title,
		Position: maxPos + domain.PositionStep,
	}
	err = cu.columnRepo.Create(ctx, column)
	if err != nil {
		return nil, err
	}

	return &CreateColumnOutput{
		Column: column,
	}, nil
}
