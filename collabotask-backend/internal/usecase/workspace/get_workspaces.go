package workspace

import (
	"context"
	"fmt"
)

func (wu *WorkspaceUseCase) GetWorkspaces(ctx context.Context, input GetWorkspacesInput) (*GetWorkspacesOutput, error) {
	userWorkspaces, err := wu.workspaceRepo.GetUserWorkspaces(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user workspaces: %w", err)
	}

	workspaces := make([]WorkspaceWithMeta, 0, len(userWorkspaces))
	for _, item := range userWorkspaces {
		ws := item.Workspace
		workspaces = append(workspaces, WorkspaceWithMeta{
			Workspace:   &ws,
			MemberCount: item.MemberCount,
			BoardCount:  item.BoardCount,
			Role:        item.Role,
		})
	}

	return &GetWorkspacesOutput{
		Workspaces: workspaces,
	}, nil
}
