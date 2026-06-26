package card

import (
	"collabotask/internal/domain/repository"
	"collabotask/internal/usecase/common"
)

type CardUseCase struct {
	cardRepo           repository.CardRepository
	columnRepo         repository.ColumnRepository
	userRepo           repository.UserRepository
	boardAccessChecker common.BoardAccessChecker
}

func NewCardUseCase(
	cardRepo repository.CardRepository,
	columnRepo repository.ColumnRepository,
	userRepo repository.UserRepository,
	boardAccessChecker common.BoardAccessChecker,
) *CardUseCase {
	return &CardUseCase{
		cardRepo:           cardRepo,
		columnRepo:         columnRepo,
		userRepo:           userRepo,
		boardAccessChecker: boardAccessChecker,
	}
}
