package board

import (
	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/pkg/validator"
	"context"
	"errors"
	"fmt"
	"strings"
)

func (bu *BoardUseCase) UpdateBoard(ctx context.Context, input UpdateBoardInput) (*UpdateBoardOutput, error) {
	if err := validator.Struct(input); err != nil {
		return nil, fmt.Errorf("failed to validate update board input: %w", err)
	}

	atLeastOne := validator.AtLeastOneProvided(input.Title, input.BackgroundColor, input.Visibility) || input.DescriptionPresent
	if !atLeastOne {
		return nil, domain.ErrAtLeastOneProvided
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

	if !canAdministerBoard(board.CreatedBy, input.RequesterID, boardMember, workspaceMember) {
		return nil, domain.ErrBoardPermissionDenied
	}

	if input.Title != nil {
		board.Title = *input.Title
	}
	if input.DescriptionPresent {
		if input.Description == nil {
			board.Description = nil
		} else if strings.TrimSpace(*input.Description) == "" {
			board.Description = nil
		} else {
			s := strings.TrimSpace(*input.Description)
			board.Description = &s
		}
	}
	if input.BackgroundColor != nil {
		board.BackgroundColor = *input.BackgroundColor
	}
	// Flipping visibility has no data cascade: members stay and assignments
	// already require membership. WORKSPACE→PRIVATE only affects future access.
	if input.Visibility != nil {
		board.Visibility = entity.BoardVisibility(*input.Visibility)
	}

	err = bu.boardRepo.Update(ctx, board)
	if err != nil {
		return nil, fmt.Errorf("failed to update the board: %w", err)
	}

	return &UpdateBoardOutput{
		Board: board,
	}, nil
}
