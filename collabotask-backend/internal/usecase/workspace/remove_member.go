package workspace

import (
	"context"
	"fmt"
)

func (wu *WorkspaceUseCaseImpl) RemoveMember(ctx context.Context, input RemoveMemberInput) error {
	requesterMember, err := wu.workspaceMemberRepo.GetByWorkspaceAndUser(ctx, input.WorkspaceID, input.RequesterID)
	if err != nil || requesterMember == nil || !requesterMember.IsAdmin() {
		return ErrNotWorkspaceAdmin
	}
	if input.RequesterID == input.UserID {
		return fmt.Errorf("cannot remove yourself")
	}

	err = wu.workspaceMemberRepo.Delete(ctx, input.WorkspaceID, input.UserID)
	if err != nil {
		return fmt.Errorf("failed to remove member from the workspace: %w", err)
	}

	return nil
}
