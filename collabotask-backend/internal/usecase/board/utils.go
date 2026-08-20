package board

import (
	"context"

	"collabotask/internal/domain/entity"

	"github.com/google/uuid"
)

// canAdministerBoard reports whether the requester may administer the board:
// they hold the BOARD_OWNER role, or they are a workspace ADMIN. Per spec §2.3
// a workspace admin's authority does not depend on board membership — they
// administer without joining. board_members.role is the sole owner authority
// (created_by is a historical trace only — UC-12e, ADR-006).
func canAdministerBoard(boardMember *entity.BoardMember, workspaceMember *entity.WorkspaceMember) bool {
	return (boardMember != nil && !boardMember.IsEmpty() && boardMember.IsOwner()) ||
		(workspaceMember != nil && !workspaceMember.IsEmpty() && workspaceMember.IsAdmin())
}

// evictNonAllowed evicts connected users who are not board members or workspace
// admins after a WORKSPACE→PRIVATE visibility flip (§4.5 UC-19b). Best-effort:
// on repo failure it swallows and skips eviction rather than failing the mutation.
func (bu *BoardUseCase) evictNonAllowed(ctx context.Context, board *entity.Board) {
	members, err := bu.boardMemberRepo.GetMembersByBoard(ctx, board.ID)
	if err != nil {
		return
	}
	wsMembers, err := bu.workspaceMemberRepo.GetMembersByWorkspace(ctx, board.WorkspaceID)
	if err != nil {
		return
	}

	allowed := make([]uuid.UUID, 0, len(members)+len(wsMembers))
	for _, m := range members {
		allowed = append(allowed, m.UserID)
	}
	for _, m := range wsMembers {
		if m.IsAdmin() {
			allowed = append(allowed, m.UserID)
		}
	}
	bu.broadcaster.EvictExcept(board.ID, allowed, "board_made_private")
}
