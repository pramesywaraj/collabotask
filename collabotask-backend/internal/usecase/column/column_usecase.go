package column

import (
	"collabotask/internal/domain/repository"
	"collabotask/internal/usecase/common"
)

type ColumnUseCase struct {
	columnRepo         repository.ColumnRepository
	boardAccessChecker common.BoardAccessChecker
}

func NewColumnUseCase(
	columnRepo repository.ColumnRepository,
	boardAccessChecker common.BoardAccessChecker,
) *ColumnUseCase {
	return &ColumnUseCase{
		columnRepo:         columnRepo,
		boardAccessChecker: boardAccessChecker,
	}
}
