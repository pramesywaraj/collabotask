package workspace

import (
	"collabotask/internal/domain"
	"collabotask/internal/infrastructure/validator"
	"context"
	"errors"
	"fmt"
)

func (wu *WorkspaceUseCaseImpl) RemoveMember(ctx context.Context, input RemoveMemberInput) error {
	if err := validator.Struct(input); err != nil {
		return fmt.Errorf("failed to validate input when removing member: %w", err)
	}

	requesterMember, err := wu.workspaceMemberRepo.GetByWorkspaceAndUser(ctx, input.WorkspaceID, input.RequesterID)
	if err != nil || requesterMember == nil || !requesterMember.IsAdmin() {
		return domain.ErrNotWorkspaceAdmin
	}
	if input.RequesterID == input.UserID {
		return domain.ErrCannotRemoveYourself
	}

	err = wu.workspaceMemberRepo.Delete(ctx, input.WorkspaceID, input.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrMemberNotFound) {
			return domain.ErrMemberNotFound
		}
		return fmt.Errorf("failed to remove member from the workspace: %w", err)
	}

	return nil
}
