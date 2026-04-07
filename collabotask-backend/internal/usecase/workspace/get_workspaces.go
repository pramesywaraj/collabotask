package workspace

import (
	"context"
	"fmt"
)

func (wu *WorkspaceUseCaseImpl) GetWorkspaces(ctx context.Context, input GetWorkspacesInput) (*GetWorkspacesOutput, error) {
	userWorkspaces, err := wu.workspaceRepo.GetUserWorkspaces(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user workspaces: %w", err)
	}

	workspaces := make([]WorkspaceWithMetaDTO, 0, len(userWorkspaces))
	for _, item := range userWorkspaces {
		workspaces = append(workspaces, workspaceListItemToDTO(item))
	}

	return &GetWorkspacesOutput{
		Workspaces: workspaces,
	}, nil
}
