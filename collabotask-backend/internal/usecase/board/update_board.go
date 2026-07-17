package board

import (
	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/usecase/common"
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

	if !canAdministerBoard(boardMember, workspaceMember) {
		return nil, domain.ErrBoardPermissionDenied
	}

	var changedFields []string
	if input.Title != nil && *input.Title != board.Title {
		changedFields = append(changedFields, "title")
	}
	if input.DescriptionPresent {
		var newDesc *string
		if input.Description != nil && strings.TrimSpace(*input.Description) != "" {
			s := strings.TrimSpace(*input.Description)
			newDesc = &s
		}
		oldNil := board.Description == nil
		newNil := newDesc == nil
		if oldNil != newNil || (!oldNil && !newNil && *board.Description != *newDesc) {
			changedFields = append(changedFields, "description")
		}
	}
	if input.BackgroundColor != nil && *input.BackgroundColor != board.BackgroundColor {
		changedFields = append(changedFields, "background_color")
	}
	var oldVisibility entity.BoardVisibility
	if input.Visibility != nil && string(board.Visibility) != *input.Visibility {
		changedFields = append(changedFields, "visibility")
		oldVisibility = board.Visibility
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

	if len(changedFields) > 0 {
		meta := map[string]any{
			entity.ActivityMetaBoardTitle:    board.Title,
			entity.ActivityMetaChangedFields: changedFields,
		}
		if oldVisibility != "" {
			meta[entity.ActivityMetaVisibilityFrom] = string(oldVisibility)
			meta[entity.ActivityMetaVisibilityTo] = string(board.Visibility)
		}
		common.WriteActivity(ctx, bu.activityRepo, input.RequesterID, &entity.Activity{
			BoardID:    board.ID,
			ActionType: entity.ActivityActionUpdated,
			EntityType: entity.ActivityEntityBoard,
			EntityID:   board.ID,
			Metadata:   meta,
		})
	}

	return &UpdateBoardOutput{
		Board: board,
	}, nil
}
