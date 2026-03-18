package board

import (
	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/dto"
	"collabotask/internal/infrastructure/validator"
	"context"
	"fmt"
)

const defaultBackgroundColor = "#0079BF"

func (bu *BoardUseCaseImpl) CreateBoard(ctx context.Context, input CreateBoardInput) (*CreateBoardOutput, error) {
	if err := validator.Struct(input); err != nil {
		return nil, fmt.Errorf("create board validation failed: %w", err)
	}

	// For now we just check whether the user exist
	// in the workspace or not, maybe later we will add
	// workspace admin or member check
	exists, err := bu.workspaceMemberRepo.IsUserExists(ctx, input.WorkspaceID, input.RequesterID)
	if err != nil {
		return nil, fmt.Errorf("failed to check workspace membership: %w", err)
	}
	if !exists {
		return nil, domain.ErrUserNotInWorkspace
	}

	var description *string
	if input.Description != nil && *input.Description != "" {
		description = input.Description
	}
	backgroundColor := defaultBackgroundColor
	if input.BackgroundColor != nil && *input.BackgroundColor != "" {
		backgroundColor = *input.BackgroundColor
	}

	board := &entity.Board{
		WorkspaceID:     input.WorkspaceID,
		Title:           input.Title,
		Description:     description,
		CreatedBy:       input.RequesterID,
		BackgroundColor: backgroundColor,
	}

	err = bu.boardRepo.CreateWithOwner(ctx, board, input.RequesterID)
	if err != nil {
		return nil, fmt.Errorf("failed to create board: %w", err)
	}

	return &CreateBoardOutput{
		Board: dto.BoardToDTO(board),
	}, nil
}
