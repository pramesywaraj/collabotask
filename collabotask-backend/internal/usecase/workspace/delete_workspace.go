package workspace

import (
	"collabotask/internal/domain"
	"collabotask/pkg/validator"
	"context"
	"errors"
	"fmt"
)

func (wu *WorkspaceUseCase) DeleteWorkspace(ctx context.Context, input DeleteWorkspaceInput) error {
	if err := validator.Struct(input); err != nil {
		return fmt.Errorf("validation: %w", err)
	}

	ws, err := wu.workspaceRepo.GetByID(ctx, input.WorkspaceID)
	if err != nil {
		if errors.Is(err, domain.ErrWorkspaceNotFound) {
			return domain.ErrWorkspaceNotFound
		}
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	if ws.OwnerID != input.RequesterID {
		return domain.ErrNotWorkspaceOwner
	}

	if err := wu.workspaceRepo.Delete(ctx, input.WorkspaceID); err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}

	return nil
}
