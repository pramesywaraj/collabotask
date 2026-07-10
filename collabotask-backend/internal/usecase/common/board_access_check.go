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

// IsBoardMember reports whether the requester holds a board_members row
// (owner or member). A nil/empty row means "not joined".
func (a *BoardAccess) IsBoardMember() bool {
	return a.BoardMember != nil && !a.BoardMember.IsEmpty()
}

func (a *BoardAccess) isWorkspaceAdmin() bool {
	return a.WorkspaceMember != nil && !a.WorkspaceMember.IsEmpty() && a.WorkspaceMember.IsAdmin()
}

func (a *BoardAccess) isPrivate() bool {
	return a.Board != nil && a.Board.Visibility == entity.BoardVisibilityPrivate
}

// BoardAccessChecker enforces the board-visibility permission matrix (ADR-005).
// The three intent methods share a private resolve() that gathers the facts
// (board + workspace membership + board membership + visibility) once; each
// method then applies its own policy tail. board_members (plus workspace-admin
// status) is the sole source of role/access truth — created_by is never
// consulted here.
type BoardAccessChecker interface {
	// CheckMetadataAccess gates GET /board/:id (board detail). Only a plain
	// member on a PRIVATE board they haven't joined is denied (404, existence
	// hidden). Admins always pass (break-glass roster hiding is the caller's job).
	CheckMetadataAccess(ctx context.Context, boardID, requesterID uuid.UUID) (*BoardAccess, error)
	// CheckViewAccess gates the kanban view. Break-glass applies: a non-joined
	// admin on a PRIVATE board must Join first (403 BOARD_JOIN_REQUIRED).
	CheckViewAccess(ctx context.Context, boardID, requesterID uuid.UUID) (*BoardAccess, error)
	// CheckMutateAccess gates card/column mutations. Break-glass applies, and a
	// plain member is never allowed to mutate: 404 on PRIVATE (hidden), 403 on
	// WORKSPACE (they legitimately see the board but may not edit it).
	CheckMutateAccess(ctx context.Context, boardID, requesterID uuid.UUID) (*BoardAccess, error)
}

type boardAccessChecker struct {
	boardRepo           repository.BoardRepository
	boardMemberRepo     repository.BoardMemberRepository
	workspaceMemberRepo repository.WorkspaceMemberRepository
}

func NewBoardAccessChecker(
	boardRepo repository.BoardRepository,
	boardMemberRepo repository.BoardMemberRepository,
	workspaceMemberRepo repository.WorkspaceMemberRepository,
) BoardAccessChecker {
	return &boardAccessChecker{
		boardRepo:           boardRepo,
		boardMemberRepo:     boardMemberRepo,
		workspaceMemberRepo: workspaceMemberRepo,
	}
}

func (ba *boardAccessChecker) CheckMetadataAccess(ctx context.Context, boardID, requesterID uuid.UUID) (*BoardAccess, error) {
	access, err := ba.resolve(ctx, boardID, requesterID)
	if err != nil {
		return nil, err
	}

	// Joined members and workspace admins can always read metadata. Only a
	// plain member on a PRIVATE board is denied — and it is hidden (404), not
	// revealed, matching the board list which filters it out entirely.
	if access.IsBoardMember() || access.isWorkspaceAdmin() {
		return access, nil
	}
	if access.isPrivate() {
		return nil, domain.ErrBoardNotFound
	}

	return access, nil
}

func (ba *boardAccessChecker) CheckViewAccess(ctx context.Context, boardID, requesterID uuid.UUID) (*BoardAccess, error) {
	access, err := ba.resolve(ctx, boardID, requesterID)
	if err != nil {
		return nil, err
	}

	if access.IsBoardMember() {
		return access, nil
	}
	if granted, err := adminBreakGlass(access); err != nil {
		return nil, err
	} else if granted {
		return access, nil
	}

	// Non-joined plain member. WORKSPACE boards are viewable; PRIVATE ones are
	// hidden (404).
	if access.isPrivate() {
		return nil, domain.ErrBoardNotFound
	}

	return access, nil
}

func (ba *boardAccessChecker) CheckMutateAccess(ctx context.Context, boardID, requesterID uuid.UUID) (*BoardAccess, error) {
	access, err := ba.resolve(ctx, boardID, requesterID)
	if err != nil {
		return nil, err
	}

	if access.IsBoardMember() {
		return access, nil
	}
	if granted, err := adminBreakGlass(access); err != nil {
		return nil, err
	} else if granted {
		return access, nil
	}

	// Non-joined plain member: never allowed to mutate. Hide PRIVATE (404),
	// reveal WORKSPACE (403 — they can see the board, they just can't edit it).
	if access.isPrivate() {
		return nil, domain.ErrBoardNotFound
	}

	return nil, domain.ErrBoardAccessDenied
}

// adminBreakGlass decides the non-joined workspace admin's fate for the
// view/mutate policies. It returns:
//   - (false, nil)                      → not an admin; fall through to the plain-member policy.
//   - (true,  nil)                      → admin on a WORKSPACE board: granted, no join needed.
//   - (false, ErrBoardJoinRequired)     → admin on a PRIVATE board: must Join first (break-glass).
func adminBreakGlass(access *BoardAccess) (granted bool, err error) {
	if !access.isWorkspaceAdmin() {
		return false, nil
	}
	if access.isPrivate() {
		return false, domain.ErrBoardJoinRequired
	}
	return true, nil
}

func (ba *boardAccessChecker) resolve(ctx context.Context, boardID, requesterID uuid.UUID) (*BoardAccess, error) {
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
	if err != nil {
		if errors.Is(err, domain.ErrMemberNotFound) {
			return nil, domain.ErrUserNotInWorkspace
		}
		return nil, fmt.Errorf("failed to fetch workspace membership: %w", err)
	}
	if workspaceMember == nil || workspaceMember.IsEmpty() {
		return nil, domain.ErrUserNotInWorkspace
	}

	boardMember, err := ba.boardMemberRepo.GetMemberByBoardAndUser(ctx, boardID, requesterID)
	if err != nil {
		if !errors.Is(err, domain.ErrBoardMemberNotFound) {
			return nil, fmt.Errorf("failed to fetch board membership: %w", err)
		}
		boardMember = nil
	}

	return &BoardAccess{
		Board:           board,
		WorkspaceMember: workspaceMember,
		BoardMember:     boardMember,
	}, nil
}
