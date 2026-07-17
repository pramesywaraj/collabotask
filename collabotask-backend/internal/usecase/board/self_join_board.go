package board

import (
	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/usecase/common"
	"collabotask/pkg/validator"
	"context"
	"errors"
	"fmt"
)

func (bu *BoardUseCase) SelfJoinBoard(ctx context.Context, input SelfJoinBoardInput) (*SelfJoinBoardOutput, error) {
	if err := validator.Struct(input); err != nil {
		return nil, fmt.Errorf("failed to validate self join board input: %w", err)
	}

	board, err := bu.boardRepo.GetByID(ctx, input.BoardID)
	if err != nil {
		if errors.Is(err, domain.ErrBoardNotFound) {
			return nil, domain.ErrBoardNotFound
		}
		return nil, fmt.Errorf("failed to verify board existence: %w", err)
	}
	if board == nil || board.IsEmpty() || board.IsArchived || board.WorkspaceID != input.WorkspaceID {
		return nil, domain.ErrBoardNotFound
	}

	workspaceMember, err := bu.workspaceMemberRepo.GetByWorkspaceAndUser(ctx, input.WorkspaceID, input.RequesterID)
	if err != nil {
		if errors.Is(err, domain.ErrMemberNotFound) {
			return nil, domain.ErrUserNotInWorkspace
		}
		return nil, fmt.Errorf("failed to fetch workspace membership for requester: %w", err)
	}
	if workspaceMember == nil || workspaceMember.IsEmpty() {
		return nil, domain.ErrUserNotInWorkspace
	}

	boardMember, err := bu.boardMemberRepo.GetMemberByBoardAndUser(ctx, input.BoardID, input.RequesterID)
	if err != nil {
		if !errors.Is(err, domain.ErrBoardMemberNotFound) {
			return nil, fmt.Errorf("error occurred when fetching board membership: %w", err)
		}
	}

	// Already a member → idempotent no-op. This is checked before the
	// eligibility gate so an existing member of a board later flipped to PRIVATE
	// still gets a clean 200, not a 404.
	if boardMember != nil && !boardMember.IsEmpty() {
		return &SelfJoinBoardOutput{Joined: false}, nil
	}

	// Eligibility (ADR-005): an admin may join any board; a plain member may
	// only join a WORKSPACE-visible one. A plain member on a PRIVATE board must
	// not even learn it exists, so it is hidden (404) rather than refused (403).
	if !workspaceMember.IsAdmin() && board.Visibility != entity.BoardVisibilityWorkspace {
		return nil, domain.ErrBoardNotFound
	}

	newBoardMember := &entity.BoardMember{
		BoardID: input.BoardID,
		UserID:  input.RequesterID,
		Role:    entity.BoardRoleMember,
	}

	// ON CONFLICT DO NOTHING RETURNING: guards the race where a concurrent join
	// slipped in between the check above and here. A no-op insert reports
	// Joined=false so the activity log / broadcast stays silent.
	inserted, err := bu.boardMemberRepo.CreateIfAbsent(ctx, newBoardMember)
	if err != nil {
		return nil, fmt.Errorf("failed join to the board: %w", err)
	}

	if inserted {
		breakGlass := workspaceMember.IsAdmin() && board.Visibility == entity.BoardVisibilityPrivate
		requesterID := input.RequesterID
		common.WriteActivity(ctx, bu.activityRepo, &entity.Activity{
			BoardID:    input.BoardID,
			UserID:     &requesterID,
			ActionType: entity.ActivityActionJoined,
			EntityType: entity.ActivityEntityMember,
			EntityID:   input.RequesterID,
			Metadata:   map[string]any{"role": string(entity.BoardRoleMember), "break_glass": breakGlass},
		})
	}

	return &SelfJoinBoardOutput{Joined: inserted}, nil
}
