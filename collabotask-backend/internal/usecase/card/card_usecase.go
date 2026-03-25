package card

import (
	"collabotask/internal/domain/repository"
	"collabotask/internal/usecase/common"
)

type CardUseCaseImpl struct {
	cardRepo           repository.CardRepository
	columnRepo         repository.ColumnRepository
	boardAccessChecker common.BoardAccessChecker
}

func NewCardUseCase(
	cardRepo repository.CardRepository,
	columnRepo repository.ColumnRepository,
	boardAccessChecker common.BoardAccessChecker,
) CardUseCase {
	return &CardUseCaseImpl{
		cardRepo:           cardRepo,
		columnRepo:         columnRepo,
		boardAccessChecker: boardAccessChecker,
	}
}
