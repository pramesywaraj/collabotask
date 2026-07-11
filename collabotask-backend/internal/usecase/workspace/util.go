package workspace

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func normalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}

func (wu *WorkspaceUseCase) isLastAdmin(ctx context.Context, workspaceID uuid.UUID) (bool, error) {
	count, err := wu.workspaceMemberRepo.CountAdmins(ctx, workspaceID)
	if err != nil {
		return false, fmt.Errorf("failed to count admins: %w", err)
	}
	return count <= 1, nil
}
