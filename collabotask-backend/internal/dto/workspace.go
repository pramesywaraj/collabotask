package dto

import (
	"collabotask/internal/domain/entity"
	"time"

	"github.com/google/uuid"
)

type WorkspaceDTO struct {
	ID          uuid.UUID
	Name        string
	Description *string
	OwnerID     uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type WorkspaceWithMetaDTO struct {
	WorkspaceDTO

	MemberCount uint
	BoardCount  uint
	Role        entity.WorkspaceRole
}

type WorkspaceMemberDTO struct {
	UserID    uuid.UUID
	Email     string
	Name      string
	AvatarURL *string
	Role      entity.WorkspaceRole
	JoinedAt  time.Time
}

type WorkspaceDetailDTO struct {
	WorkspaceDTO

	UserRole entity.WorkspaceRole
	Members  []WorkspaceMemberDTO
}

func WorkspaceToDTO(workspace *entity.Workspace) WorkspaceDTO {
	return WorkspaceDTO{
		ID:          workspace.ID,
		Name:        workspace.Name,
		Description: workspace.Description,
		OwnerID:     workspace.OwnerID,
		CreatedAt:   workspace.CreatedAt,
		UpdatedAt:   workspace.UpdatedAt,
	}
}

func WorkspaceListItemToDTO(item *entity.WorkspaceListItem) WorkspaceWithMetaDTO {
	return WorkspaceWithMetaDTO{
		WorkspaceDTO: WorkspaceToDTO(&item.Workspace),
		MemberCount:  item.MemberCount,
		BoardCount:   item.BoardCount,
		Role:         item.Role,
	}
}

func WorkspaceMemberToDTO(member *entity.WorkspaceMember, user *entity.User) WorkspaceMemberDTO {
	return WorkspaceMemberDTO{
		UserID:    user.ID,
		Email:     user.Email,
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
		Role:      member.Role,
		JoinedAt:  member.JoinedAt,
	}
}
