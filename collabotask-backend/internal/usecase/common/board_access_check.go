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

type BoardAccess struct {
	Board           *entity.Board
	WorkspaceMember *entity.WorkspaceMember
	BoardMember     *entity.BoardMember
}

type BoardAccessChecker interface {
	Check(ctx context.Context, boardID, requesterID uuid.UUID) (*entity.Board, error)
	Resolve(ctx context.Context, boardID, requesterID uuid.UUID) (*BoardAccess, error)
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
	access, err := ba.Resolve(ctx, boardID, requesterID)
	if err != nil {
		return nil, err
	}

	return access.Board, nil
}

func (ba *BoardAccessCheckerImpl) Resolve(ctx context.Context, boardID, requesterID uuid.UUID) (*BoardAccess, error) {
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

	workspaceMember, err := ba.workspaceMemberRepo.GetByWorkspaceAndUser(ctx, board.WorkspaceID, requesterID)
	if err != nil || workspaceMember == nil || workspaceMember.IsEmpty() {
		return nil, domain.ErrUserNotInWorkspace
	}

	boardMember, err := ba.boardMemberRepo.GetMemberByBoardAndUser(ctx, boardID, requesterID)
	if err != nil {
		if !errors.Is(err, domain.ErrBoardMemberNotFound) {
			return nil, fmt.Errorf("failed to fetch board membership: %w", err)
		}
		boardMember = nil
	}

	hasAccess := workspaceMember.IsAdmin() ||
		board.CreatedBy == requesterID ||
		(boardMember != nil && !boardMember.IsEmpty())
	if !hasAccess {
		return nil, domain.ErrBoardAccessDenied
	}

	return &BoardAccess{
		Board:           board,
		WorkspaceMember: workspaceMember,
		BoardMember:     boardMember,
	}, nil
}
