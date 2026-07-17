package workspace

import (
	"collabotask/internal/domain/repository"
)

type WorkspaceUseCase struct {
	workspaceRepo       repository.WorkspaceRepository
	workspaceMemberRepo repository.WorkspaceMemberRepository
	userRepo            repository.UserRepository
	activityRepo        repository.ActivityRepository
}

func NewWorkspaceUseCase(
	wRepo repository.WorkspaceRepository,
	wmRepo repository.WorkspaceMemberRepository,
	uRepo repository.UserRepository,
	activityRepo repository.ActivityRepository,
) *WorkspaceUseCase {
	return &WorkspaceUseCase{
		workspaceRepo:       wRepo,
		workspaceMemberRepo: wmRepo,
		userRepo:            uRepo,
		activityRepo:        activityRepo,
	}
}
