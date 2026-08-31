package workspace

import (
	"collabotask/internal/domain/repository"
	"collabotask/internal/usecase/common"
)

type WorkspaceUseCase struct {
	workspaceRepo       repository.WorkspaceRepository
	workspaceMemberRepo repository.WorkspaceMemberRepository
	boardRepo           repository.BoardRepository
	userRepo            repository.UserRepository
	activityRepo        repository.ActivityRepository
	broadcaster         common.Broadcaster
	tx                  common.Transactor
}

func NewWorkspaceUseCase(
	wRepo repository.WorkspaceRepository,
	wmRepo repository.WorkspaceMemberRepository,
	bRepo repository.BoardRepository,
	uRepo repository.UserRepository,
	activityRepo repository.ActivityRepository,
	broadcaster common.Broadcaster,
	tx common.Transactor,
) *WorkspaceUseCase {
	return &WorkspaceUseCase{
		workspaceRepo:       wRepo,
		workspaceMemberRepo: wmRepo,
		boardRepo:           bRepo,
		userRepo:            uRepo,
		activityRepo:        activityRepo,
		broadcaster:         broadcaster,
		tx:                  tx,
	}
}
