package workspace

import (
	"collabotask/internal/domain/repository"
)

type WorkspaceUseCaseImpl struct {
	workspaceRepo       repository.WorkspaceRepository
	workspaceMemberRepo repository.WorkspaceMemberRepository
	userRepo            repository.UserRepository
}

func NewWorkspaceUseCase(
	wRepo repository.WorkspaceRepository,
	wmRepo repository.WorkspaceMemberRepository,
	uRepo repository.UserRepository,
) WorkspaceUseCase {
	return &WorkspaceUseCaseImpl{
		workspaceRepo:       wRepo,
		workspaceMemberRepo: wmRepo,
		userRepo:            uRepo,
	}
}
