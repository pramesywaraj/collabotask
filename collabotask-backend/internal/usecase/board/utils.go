package board

import (
	"collabotask/internal/domain/entity"
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
