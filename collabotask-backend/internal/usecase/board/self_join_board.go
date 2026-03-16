package board

import (
	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/infrastructure/validator"
	"context"
	"errors"
	"fmt"
)

func (bu *BoardUseCaseImpl) SelfJoinBoard(ctx context.Context, input SelfJoinBoardInput) error {
	if err := validator.Struct(input); err != nil {
		return fmt.Errorf("failed to validate self join board input: %w", err)
	}

	board, err := bu.boardRepo.GetByID(ctx, input.BoardID)
	if err != nil {
		if errors.Is(err, domain.ErrBoardNotFound) {
			return domain.ErrBoardNotFound
		}
		return fmt.Errorf("failed to fetch board detail: %w", err)
	}
	if board == nil || board.IsEmpty() || board.IsArchived || board.WorkspaceID != input.WorkspaceID {
		return domain.ErrBoardNotFound
	}

	workspaceMember, err := bu.workspaceMemberRepo.GetByWorkspaceAndUser(ctx, input.WorkspaceID, input.RequesterID)
	if err != nil {
		if errors.Is(err, domain.ErrMemberNotFound) {
			return domain.ErrUserNotInWorkspace
		}
		return fmt.Errorf("failed to fetch workspace membership for requester: %w", err)
	}
	if workspaceMember == nil || workspaceMember.IsEmpty() {
		return domain.ErrUserNotInWorkspace
	}

	boardMember, err := bu.boardMemberRepo.GetMemberByBoardAndUser(ctx, input.BoardID, input.RequesterID)
	if err != nil {
		if !errors.Is(err, domain.ErrBoardMemberNotFound) {
			return fmt.Errorf("error occurred when fetching board membership: %w", err)
		}
	}

	canJoin := workspaceMember.IsAdmin() && (boardMember == nil || boardMember.IsEmpty())
	if !canJoin {
		return domain.ErrBoardCannotJoin
	}

	newBoardMember := &entity.BoardMember{
		BoardID: input.BoardID,
		UserID:  input.RequesterID,
		Role:    entity.BoardRoleMember,
	}

	if err := bu.boardMemberRepo.Create(ctx, newBoardMember); err != nil {
		return fmt.Errorf("failed to add member to board: %w", err)
	}

	return nil
}
