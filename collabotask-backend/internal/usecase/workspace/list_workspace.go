package workspace

import (
	"context"
	"fmt"
)

func (wu *WorkspaceUseCaseImpl) ListWorkspaces(ctx context.Context, input ListWorkspacesInput) (*ListWorkspacesOutput, error) {
	userWorkspaces, err := wu.workspaceRepo.GetUserWorkspaces(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user workspaces: %w", err)
	}

	workspaces := make([]WorkspaceWithMetaDTO, 0, len(userWorkspaces))
	for _, item := range userWorkspaces {
		workspaces = append(workspaces, workspaceListItemToDTO(item))
	}

	return &ListWorkspacesOutput{
		Workspaces: workspaces,
	}, nil
}
