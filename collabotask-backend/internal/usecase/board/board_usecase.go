package board

import (
	"collabotask/internal/domain/repository"
	"collabotask/internal/usecase/common"
)

type BoardUseCase struct {
	boardAccessChecker  common.BoardAccessChecker
	boardRepo           repository.BoardRepository
	boardMemberRepo     repository.BoardMemberRepository
	workspaceMemberRepo repository.WorkspaceMemberRepository
	userRepo            repository.UserRepository
	columnRepo          repository.ColumnRepository
	cardRepo            repository.CardRepository
}

func NewBoardUseCase(
	boardAccessChecker common.BoardAccessChecker,
	boardRepo repository.BoardRepository,
	boardMemberRepo repository.BoardMemberRepository,
	workspaceMemberRepo repository.WorkspaceMemberRepository,
	userRepo repository.UserRepository,
	columnRepo repository.ColumnRepository,
	cardRepo repository.CardRepository,
) *BoardUseCase {
	return &BoardUseCase{
		boardAccessChecker:  boardAccessChecker,
		boardRepo:           boardRepo,
		boardMemberRepo:     boardMemberRepo,
		workspaceMemberRepo: workspaceMemberRepo,
		userRepo:            userRepo,
		columnRepo:          columnRepo,
		cardRepo:            cardRepo,
	}
}
