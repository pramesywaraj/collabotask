package workspace

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"collabotask/internal/domain/entity"
	"collabotask/internal/domain/repository"
	"collabotask/internal/usecase/common"
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

// fanOutParticipationCascade broadcasts the realtime consequences of a
// workspace-member participation cascade (UC-06 / UC-06c). Evict-first ordering
// (§4.5 UC-19b): the user is evicted from every workspace board room — including
// WORKSPACE-visible boards they watched without joining (the AffectedBoardIDs
// leak, index §5) — then each affected board sees MEMBER_REMOVED, then each
// cleared card fires CARD_UPDATED{assigned_to:null} routed by its own board.
//
// A non-silent reason sends a targeted ACCESS_REVOKED per evicted board (removal);
// EvictReasonSilent stays quiet (voluntary leave — the room learns via USER_LEFT).
// Best-effort throughout: called after commit, never fails the mutation.
func (wu *WorkspaceUseCase) fanOutParticipationCascade(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	result repository.WorkspaceCascadeResult,
	reason common.EvictReason,
) {
	boardIDs, err := wu.boardRepo.GetBoardIDsByWorkspace(ctx, workspaceID)
	if err != nil {
		// Fallback: still evicts every member-board; only the rare
		// workspace-visible-watcher edge is missed. Never fail the mutation.
		boardIDs = result.AffectedBoardIDs
	}
	wu.broadcaster.EvictUserFromRooms(userID, boardIDs, reason)

	for _, boardID := range result.AffectedBoardIDs {
		wu.broadcaster.Broadcast(boardID, common.MemberRemoved{
			BoardID: boardID,
			UserID:  userID,
		})
	}
	for _, card := range result.AffectedCards {
		wu.broadcaster.Broadcast(card.BoardID, common.CardUpdated{
			Card:          &entity.Card{ID: card.CardID},
			ChangedFields: []string{"assigned_to"},
		})
	}
}
