package column

import (
	"cmp"
	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/dto"
	"collabotask/internal/infrastructure/validator"
	"context"
	"fmt"
	"slices"
)

func (cu *ColumnUseCaseImpl) UpdateColumnPosition(ctx context.Context, input UpdateColumnPositionInput) (*UpdateColumnPositionOutput, error) {
	if err := validator.Struct(input); err != nil {
		return nil, fmt.Errorf("failed to validate update column position input: %w", err)
	}

	column, err := cu.columnRepo.GetByID(ctx, input.ColumnID)
	if err != nil {
		return nil, err
	}

	_, err = cu.boardAccessChecker.Check(ctx, column.BoardID, input.RequesterID)
	if err != nil {
		return nil, err
	}

	columns, err := cu.columnRepo.ListByBoard(ctx, column.BoardID)
	if err != nil {
		return nil, fmt.Errorf("failed to list columns for board: %w", err)
	}

	slices.SortStableFunc(columns, func(a, b *entity.Column) int {
		if c := cmp.Compare(a.Position, b.Position); c != 0 {
			return c
		}
		return cmp.Compare(a.ID.String(), b.ID.String())
	})

	oldIdx := -1
	for i, col := range columns {
		if col.ID == input.ColumnID {
			oldIdx = i
			break
		}
	}
	if oldIdx < 0 {
		return nil, fmt.Errorf("column missing from board list: %w", domain.ErrInconsistentState)
	}

	newIdx := input.Position
	if newIdx < 0 {
		newIdx = 0
	}
	if newIdx >= len(columns) {
		newIdx = len(columns) - 1
	}

	if oldIdx == newIdx {
		return &UpdateColumnPositionOutput{
			Column: dto.ColumnToDTO(columns[oldIdx]),
		}, nil
	}

	moved := columns[oldIdx]
	without := slices.Delete(slices.Clone(columns), oldIdx, oldIdx+1)
	reordered := slices.Insert(without, newIdx, moved)

	for i, column := range reordered {
		column.Position = i
	}

	if err := cu.columnRepo.ReorderPositions(ctx, reordered); err != nil {
		return nil, err
	}

	var output *entity.Column
	for _, col := range reordered {
		if col.ID == input.ColumnID {
			output = col
			break
		}
	}
	if output == nil {
		return nil, fmt.Errorf("column missing after reorder: %w", domain.ErrInconsistentState)
	}

	return &UpdateColumnPositionOutput{
		Column: dto.ColumnToDTO(output),
	}, nil
}
