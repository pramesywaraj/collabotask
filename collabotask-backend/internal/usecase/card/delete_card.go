package card

import (
	"collabotask/internal/domain"
	"collabotask/internal/infrastructure/validator"
	"context"
	"errors"
	"fmt"
)

func (cru *CardUseCaseImpl) DeleteCard(ctx context.Context, input DeleteCardInput) error {
	if err := validator.Struct(input); err != nil {
		return fmt.Errorf("failed to validate delete card input: %w", err)
	}

	card, err := cru.cardRepo.GetByID(ctx, input.CardID)
	if err != nil {
		if errors.Is(err, domain.ErrCardNotFound) {
			return domain.ErrCardNotFound
		}
		return fmt.Errorf("failed to fetch card: %w", err)
	}
	if !card.BelongsToColumn(input.ColumnID) {
		return domain.ErrCardNotInColumn
	}

	column, err := cru.columnRepo.GetByID(ctx, card.ColumnID)
	if err != nil {
		if errors.Is(err, domain.ErrColumnNotFound) {
			return domain.ErrColumnNotFound
		}
		return fmt.Errorf("failed to fetch column: %w", err)
	}

	_, err = cru.boardAccessChecker.Check(ctx, column.BoardID, input.RequesterID)
	if err != nil {
		return err
	}

	err = cru.cardRepo.DeleteWithReorder(ctx, input.CardID)
	if err != nil {
		return err
	}

	return nil
}
