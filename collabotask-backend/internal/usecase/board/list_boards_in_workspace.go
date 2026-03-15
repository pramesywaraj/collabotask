package board

import (
	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/internal/dto"
	"collabotask/internal/infrastructure/validator"
	"context"
	"fmt"
)

func (bu *BoardUseCaseImpl) ListBoardsInWorkspace(ctx context.Context, input ListBoardsInput) (*ListBoardsOutput, error) {
	if err := validator.Struct(input); err != nil {
		return nil, fmt.Errorf("failed to validate list boards in workspace input: %w", err)
	}

	workspaceMember, err := bu.workspaceMemberRepo.GetByWorkspaceAndUser(ctx, input.WorkspaceID, input.RequesterID)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence in workspace: %w", err)
	}
	if workspaceMember == nil || workspaceMember.IsEmpty() {
		return nil, domain.ErrUserNotInWorkspace
	}

	boards, err := bu.boardRepo.GetUserBoardsInWorkspace(ctx, input.WorkspaceID, input.RequesterID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user boards in workspace: %w", err)
	}

	resultBoards := make([]dto.BoardWithMetaDTO, 0, len(boards))
	for _, board := range boards {
		var userRole *entity.BoardRole
		if board.UserRole != "" {
			role := board.UserRole
			userRole = &role
		}

		resultBoards = append(resultBoards, dto.BoardWithMetaDTO{
			BoardDTO:     dto.BoardToDTO(&board.Board),
			UserRole:     userRole,
			AccessStatus: board.AccessStatus,
			MemberCount:  board.MemberCount,
		})
	}

	return &ListBoardsOutput{
		Boards: resultBoards,
	}, nil
}
