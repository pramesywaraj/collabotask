package board

import (
	"collabotask/internal/domain"
	"collabotask/internal/infrastructure/validator"
	"context"
	"errors"
	"fmt"
)

func (bu *BoardUseCaseImpl) RemoveMember(ctx context.Context, input RemoveMemberInput) error {
	if err := validator.Struct(input); err != nil {
		return fmt.Errorf("failed to validate remove member input: %w", err)
	}

	if input.RequesterID == input.UserID {
		return domain.ErrCannotRemoveYourself
	}

	board, err := bu.boardRepo.GetByID(ctx, input.BoardID)
	if err != nil || board == nil || board.WorkspaceID != input.WorkspaceID {
		return domain.ErrBoardNotFound
	}

	workspaceMember, err := bu.workspaceMemberRepo.GetByWorkspaceAndUser(ctx, input.WorkspaceID, input.RequesterID)
	if err != nil || workspaceMember == nil || workspaceMember.IsEmpty() {
		return domain.ErrUserNotInWorkspace
	}

	boardMember, err := bu.boardMemberRepo.GetMemberByBoardAndUser(ctx, input.BoardID, input.RequesterID)
	if err != nil {
		return fmt.Errorf("failed to check requester permission: %w", err)
	}

	canManage := canManageBoardMembers(board.CreatedBy, input.RequesterID, boardMember, workspaceMember)
	if !canManage {
		return domain.ErrBoardPermissionDenied
	}

	err = bu.boardMemberRepo.Delete(ctx, input.BoardID, input.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrBoardMemberNotFound) {
			return domain.ErrBoardMemberNotFound
		}
		return fmt.Errorf("failed to remove member from board: %w", err)
	}

	return nil
}
