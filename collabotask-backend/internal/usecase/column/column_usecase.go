package column

import (
	"collabotask/internal/domain/repository"
	"collabotask/internal/usecase/common"
)

type ColumnUseCaseImpl struct {
	columnRepo         repository.ColumnRepository
	boardAccessChecker common.BoardAccessChecker
}

func NewColumnUseCase(
	columnRepo repository.ColumnRepository,
	boardAccessChecker common.BoardAccessChecker,
) ColumnUseCase {
	return &ColumnUseCaseImpl{
		columnRepo:         columnRepo,
		boardAccessChecker: boardAccessChecker,
	}
}
