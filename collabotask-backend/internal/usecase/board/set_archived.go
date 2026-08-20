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

func (bu *BoardUseCase) SetArchived(ctx context.Context, input SetArchivedInput) (*SetArchivedOutput, error) {
	if err := validator.Struct(input); err != nil {
		return nil, fmt.Errorf("failed to validate set archived status in board input: %w", err)
	}

	board, err := bu.boardRepo.GetByID(ctx, input.BoardID)
	if err != nil {
		if errors.Is(err, domain.ErrBoardNotFound) {
			return nil, domain.ErrBoardNotFound
		}
		return nil, fmt.Errorf("failed to fetch board detail: %w", err)
	}
	if board == nil || board.IsEmpty() {
		return nil, domain.ErrBoardNotFound
	}

	workspaceMember, err := bu.workspaceMemberRepo.GetByWorkspaceAndUser(ctx, board.WorkspaceID, input.RequesterID)
	if err != nil || workspaceMember == nil || workspaceMember.IsEmpty() {
		return nil, domain.ErrUserNotInWorkspace
	}

	boardMember, err := bu.boardMemberRepo.GetMemberByBoardAndUser(ctx, input.BoardID, input.RequesterID)
	if err != nil && !errors.Is(err, domain.ErrBoardMemberNotFound) {
		return nil, fmt.Errorf("failed to fetch board membership: %w", err)
	}

	if !canAdministerBoard(boardMember, workspaceMember) {
		return nil, domain.ErrBoardPermissionDenied
	}

	if board.IsArchived == *input.IsArchived {
		return &SetArchivedOutput{Board: board}, nil
	}

	err = bu.boardRepo.SetArchived(ctx, input.BoardID, *input.IsArchived)
	if err != nil {
		return nil, fmt.Errorf("failed to set archived status for board: %w", err)
	}

	board.IsArchived = *input.IsArchived

	actionType := entity.ActivityActionArchived
	if !board.IsArchived {
		actionType = entity.ActivityActionUnarchived
	}
	common.WriteActivity(ctx, bu.activityRepo, input.RequesterID, &entity.Activity{
		BoardID:    board.ID,
		ActionType: actionType,
		EntityType: entity.ActivityEntityBoard,
		EntityID:   board.ID,
		Metadata:   map[string]any{},
	})

	bu.broadcaster.Broadcast(board.ID, common.BoardArchivedSet{
		BoardID:  board.ID,
		Archived: board.IsArchived,
	})

	return &SetArchivedOutput{
		Board: board,
	}, nil
}
