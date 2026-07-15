package card

import (
	"collabotask/internal/domain"
	"collabotask/internal/domain/repository"
	"collabotask/internal/usecase/common"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type CardUseCase struct {
	cardRepo           repository.CardRepository
	columnRepo         repository.ColumnRepository
	userRepo           repository.UserRepository
	boardAccessChecker common.BoardAccessChecker
	boardMemberRepo    repository.BoardMemberRepository
}

func NewCardUseCase(
	cardRepo repository.CardRepository,
	columnRepo repository.ColumnRepository,
	userRepo repository.UserRepository,
	boardAccessChecker common.BoardAccessChecker,
	boardMemberRepo repository.BoardMemberRepository,
) *CardUseCase {
	return &CardUseCase{
		cardRepo:           cardRepo,
		columnRepo:         columnRepo,
		userRepo:           userRepo,
		boardAccessChecker: boardAccessChecker,
		boardMemberRepo:    boardMemberRepo,
	}
}

// requireBoardMember returns ErrAssigneeNotBoardMember (400) when userID has no
// board_members row for boardID, collapsing the "unknown user" and "not a member"
// cases into one sentinel (Decision C, §2.8).
func (cru *CardUseCase) requireBoardMember(ctx context.Context, boardID, userID uuid.UUID) error {
	isMember, err := cru.boardMemberRepo.IsUserExists(ctx, boardID, userID)
	if err != nil {
		return fmt.Errorf("failed to verify assignee board membership: %w", err)
	}
	if !isMember {
		return domain.ErrAssigneeNotBoardMember
	}
	return nil
}
