package board

import "collabotask/internal/domain/repository"

type BoardUseCaseImpl struct {
	boardRepo           repository.BoardRepository
	boardMemberRepo     repository.BoardMemberRepository
	workspaceRepo       repository.WorkspaceRepository
	workspaceMemberRepo repository.WorkspaceMemberRepository
	userRepo            repository.UserRepository
}

func NewBoardUseCase(
	boardRepo repository.BoardRepository,
	boardMemberRepo repository.BoardMemberRepository,
	workspaceRepo repository.WorkspaceRepository,
	workspaceMemberRepo repository.WorkspaceMemberRepository,
	userRepo repository.UserRepository,
) BoardUseCase {
	return &BoardUseCaseImpl{
		boardRepo:           boardRepo,
		boardMemberRepo:     boardMemberRepo,
		workspaceRepo:       workspaceRepo,
		workspaceMemberRepo: workspaceMemberRepo,
		userRepo:            userRepo,
	}
}
