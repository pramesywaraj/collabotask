package workspace

import (
	"collabotask/internal/domain/entity"
	"collabotask/internal/infrastructure/validator"
	"context"
	"fmt"
)

func (wu *WorkspaceUseCaseImpl) CreateWorkspace(ctx context.Context, input CreateWorkspaceInput) (*CreateWorkspaceOutput, error) {
	if err := validator.Struct(input); err != nil {
		return nil, fmt.Errorf("create workspace validation failed: %w", err)
	}

	var description *string
	if input.Description != nil && *input.Description != "" {
		description = input.Description
	}

	workspace := &entity.Workspace{
		Name:        input.Name,
		Description: description,
		OwnerID:     input.OwnerID,
	}

	err := wu.workspaceRepo.CreateWithOwner(ctx, workspace, input.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	return &CreateWorkspaceOutput{
		Workspace: workspaceToDTO(workspace),
	}, nil
}
