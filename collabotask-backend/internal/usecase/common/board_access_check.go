package common

import (
	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/domain/repository"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type BoardAccessChecker interface {
	Check(ctx context.Context, boardID, requesterID uuid.UUID) (*entity.Board, error)
}

type BoardAccessCheckerImpl struct {
	boardRepo           repository.BoardRepository
	boardMemberRepo     repository.BoardMemberRepository
	workspaceMemberRepo repository.WorkspaceMemberRepository
}

func NewBoardAccessChecker(
	boardRepo repository.BoardRepository,
	boardMemberRepo repository.BoardMemberRepository,
	workspaceMemberRepo repository.WorkspaceMemberRepository,
) BoardAccessChecker {
	return &BoardAccessCheckerImpl{
		boardRepo:           boardRepo,
		boardMemberRepo:     boardMemberRepo,
		workspaceMemberRepo: workspaceMemberRepo,
	}
}

func (ba *BoardAccessCheckerImpl) Check(ctx context.Context, boardID, requesterID uuid.UUID) (*entity.Board, error) {
	board, err := ba.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		if errors.Is(err, domain.ErrBoardNotFound) {
			return nil, domain.ErrBoardNotFound
		}
		return nil, fmt.Errorf("failed to fetch board: %w", err)
	}
	if board == nil || board.IsEmpty() || board.IsArchived {
		return nil, domain.ErrBoardNotFound
	}

	workspaceMembership, err := ba.workspaceMemberRepo.GetByWorkspaceAndUser(ctx, board.WorkspaceID, requesterID)
	if err != nil || workspaceMembership == nil || workspaceMembership.IsEmpty() {
		return nil, domain.ErrUserNotInWorkspace
	}

	boardMembership, _ := ba.boardMemberRepo.GetMemberByBoardAndUser(ctx, boardID, requesterID)
	hasAccess := workspaceMembership.IsAdmin() || board.CreatedBy == requesterID || (boardMembership != nil && !boardMembership.IsEmpty())
	if !hasAccess {
		return nil, domain.ErrBoardAccessDenied
	}

	return board, nil
}
