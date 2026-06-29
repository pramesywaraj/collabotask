package workspace

import (
	"collabotask/internal/domain"
	"collabotask/internal/domain/entity"
	"collabotask/pkg/validator"
	"context"
	"fmt"
	"strings"
)

func (wu *WorkspaceUseCase) InviteMember(ctx context.Context, input InviteMemberInput) (*InviteMemberOutput, error) {
	if err := validator.Struct(input); err != nil {
		return nil, fmt.Errorf("invite member validation failed: %w", err)
	}

	requesterMember, err := wu.workspaceMemberRepo.GetByWorkspaceAndUser(ctx, input.WorkspaceID, input.RequesterID)
	if err != nil || requesterMember == nil || !requesterMember.IsAdmin() {
		return nil, domain.ErrNotWorkspaceAdmin
	}

	results := make([]InviteResult, 0, len(input.Emails))
	for _, email := range input.Emails {
		trimmedEmail := strings.TrimSpace(strings.ToLower(email))
		if trimmedEmail == "" {
			continue
		}

		user, err := wu.userRepo.GetByEmail(ctx, trimmedEmail)
		if err != nil || user == nil {
			results = append(results, InviteResult{Email: trimmedEmail, Success: false, ErrorCode: "NOT_FOUND"})
			continue
		}

		existsInWorkspace, err := wu.workspaceMemberRepo.IsUserExists(ctx, input.WorkspaceID, user.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to check member existence: %w", err)
		}
		if existsInWorkspace {
			results = append(results, InviteResult{Email: trimmedEmail, Success: false, ErrorCode: "CONFLICT"})
			continue
		}

		member := &entity.WorkspaceMember{
			WorkspaceID: input.WorkspaceID,
			UserID:      user.ID,
			Role:        entity.WorkspaceRoleMember,
		}
		if err := wu.workspaceMemberRepo.Create(ctx, member); err != nil {
			return nil, fmt.Errorf("failed to add member to workspace: %w", err)
		}

		results = append(results, InviteResult{Email: trimmedEmail, Success: true, ErrorCode: ""})
	}

	return &InviteMemberOutput{Results: results}, nil
}
