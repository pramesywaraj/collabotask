package column

import (
	"collabotask/internal/domain/repository"
	"collabotask/internal/usecase/common"
)

type ColumnUseCase struct {
	columnRepo         repository.ColumnRepository
	boardAccessChecker common.BoardAccessChecker
	activityRepo       repository.ActivityRepository
	broadcaster        common.Broadcaster
}

func NewColumnUseCase(
	columnRepo repository.ColumnRepository,
	boardAccessChecker common.BoardAccessChecker,
	activityRepo repository.ActivityRepository,
	broadcaster common.Broadcaster,
) *ColumnUseCase {
	return &ColumnUseCase{
		columnRepo:         columnRepo,
		boardAccessChecker: boardAccessChecker,
		activityRepo:       activityRepo,
		broadcaster:        broadcaster,
	}
}
