package board

import "collabotask/internal/domain/repository"

type BoardUseCaseImpl struct {
	boardRepo           repository.BoardRepository
	boardMemberRepo     repository.BoardMemberRepository
	workspaceRepo       repository.WorkspaceRepository
	workspaceMemberRepo repository.WorkspaceMemberRepository
	userRepo            repository.UserRepository
	columnRepo          repository.ColumnRepository
	cardRepo            repository.CardRepository
}

func NewBoardUseCase(
	boardRepo repository.BoardRepository,
	boardMemberRepo repository.BoardMemberRepository,
	workspaceRepo repository.WorkspaceRepository,
	workspaceMemberRepo repository.WorkspaceMemberRepository,
	userRepo repository.UserRepository,
	columnRepo repository.ColumnRepository,
	cardRepo repository.CardRepository,
) BoardUseCase {
	return &BoardUseCaseImpl{
		boardRepo:           boardRepo,
		boardMemberRepo:     boardMemberRepo,
		workspaceRepo:       workspaceRepo,
		workspaceMemberRepo: workspaceMemberRepo,
		userRepo:            userRepo,
		columnRepo:          columnRepo,
		cardRepo:            cardRepo,
	}
}
