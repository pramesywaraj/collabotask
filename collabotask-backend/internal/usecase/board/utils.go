package board

import (
	"collabotask/internal/domain/entity"

	"github.com/google/uuid"
)

// canAdministerBoard reports whether the requester may administer the board:
// they are the BOARD_OWNER, or a workspace ADMIN. Per spec §2.3 a workspace
// admin's authority does not depend on being a board member — they administer
// without joining.
//
// The `createdBy == requesterID` clause is an owner proxy, the same deferred
// pattern as leave_board's owner guard: it moves to board_members.role in UC-12e
// (ownership transfer). It is redundant in Phase 1 (the creator is always seeded
// as a BOARD_OWNER row, so IsOwner() already covers it) and correct until then
// (created_by == owner holds). ADR-005 removed created_by only from access-grant
// and role-display logic, not from this owner authority check.
func canAdministerBoard(createdBy, requesterID uuid.UUID, boardMember *entity.BoardMember, workspaceMember *entity.WorkspaceMember) bool {
	return createdBy == requesterID ||
		(boardMember != nil && !boardMember.IsEmpty() && boardMember.IsOwner()) ||
		(workspaceMember != nil && !workspaceMember.IsEmpty() && workspaceMember.IsAdmin())
}
