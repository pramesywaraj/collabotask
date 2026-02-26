package workspace

import (
	"collabotask/internal/domain/entity"
	"collabotask/internal/infrastructure/validator"
	"context"
	"fmt"
	"strings"
)

func (wu *WorkspaceUseCaseImpl) InviteMember(ctx context.Context, input InviteMemberInput) (*InviteMemberOutput, error) {
	if err := validator.Struct(input); err != nil {
		return nil, fmt.Errorf("invite member validation failed: %w", err)
	}

	requesterMember, err := wu.workspaceMemberRepo.GetByWorkspaceAndUser(ctx, input.WorkspaceID, input.RequesterID)
	if err != nil || requesterMember == nil || !requesterMember.IsAdmin() {
		return nil, ErrNotWorkspaceAdmin
	}

	for _, email := range input.Emails {
		trimmedEmail := strings.TrimSpace(strings.ToLower(email))
		if trimmedEmail == "" {
			continue
		}

		user, err := wu.userRepo.GetByEmail(ctx, trimmedEmail)
		if err != nil || user == nil {
			return nil, ErrUserNotFound
		}

		existsInWorkspace, err := wu.workspaceMemberRepo.IsUserExists(ctx, input.WorkspaceID, user.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to check member existence: %w", err)
		}
		if existsInWorkspace {
			return nil, ErrAlreadyMember
		}

		member := &entity.WorkspaceMember{
			WorkspaceID: input.WorkspaceID,
			UserID:      user.ID,
			Role:        entity.WorkspaceRoleMember,
		}
		if err := wu.workspaceMemberRepo.Create(ctx, member); err != nil {
			return nil, fmt.Errorf("failed to add member to workspace: %w", err)
		}
	}

	return &InviteMemberOutput{
		Message: "Users have been added to the workspace",
	}, nil
}
