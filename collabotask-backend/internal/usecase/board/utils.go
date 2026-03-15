package board

import (
	"collabotask/internal/domain/entity"

	"github.com/google/uuid"
)

func canManageBoardMembers(createdBy, requesterID uuid.UUID, boardMember *entity.BoardMember, workspaceMember *entity.WorkspaceMember) bool {
	return createdBy == requesterID ||
		(boardMember != nil && !boardMember.IsEmpty() && boardMember.IsOwner()) ||
		(workspaceMember.IsAdmin() && boardMember != nil && !boardMember.IsEmpty())
}
