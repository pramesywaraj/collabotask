package board

import (
	"collabotask/internal/domain/entity"

	"github.com/google/uuid"
)

// canAdministerBoard reports whether the requester may administer the board:
// they are the BOARD_OWNER, or a workspace ADMIN. Per spec §2.3 a workspace
// admin's authority does not depend on being a board member — they administer
// without joining.
func canAdministerBoard(createdBy, requesterID uuid.UUID, boardMember *entity.BoardMember, workspaceMember *entity.WorkspaceMember) bool {
	return createdBy == requesterID ||
		(boardMember != nil && !boardMember.IsEmpty() && boardMember.IsOwner()) ||
		(workspaceMember != nil && !workspaceMember.IsEmpty() && workspaceMember.IsAdmin())
}
