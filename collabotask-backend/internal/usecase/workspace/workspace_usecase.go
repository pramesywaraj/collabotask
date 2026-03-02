package workspace

import (
	"collabotask/internal/domain/entity"
	"collabotask/internal/domain/repository"
)

type WorkspaceUseCaseImpl struct {
	workspaceRepo       repository.WorkspaceRepository
	workspaceMemberRepo repository.WorkspaceMemberRepository
	userRepo            repository.UserRepository
}

func NewWorkspaceUseCase(
	wRepo repository.WorkspaceRepository,
	wmRepo repository.WorkspaceMemberRepository,
	uRepo repository.UserRepository,
) WorkspaceUseCase {
	return &WorkspaceUseCaseImpl{
		workspaceRepo:       wRepo,
		workspaceMemberRepo: wmRepo,
		userRepo:            uRepo,
	}
}

func workspaceToDTO(workspace *entity.Workspace) WorkspaceDTO {
	return WorkspaceDTO{
		ID:          workspace.ID,
		Name:        workspace.Name,
		Description: workspace.Description,
		OwnerID:     workspace.OwnerID,
		CreatedAt:   workspace.CreatedAt,
		UpdatedAt:   workspace.UpdatedAt,
	}
}

func workspaceListItemToDTO(item *entity.WorkspaceListItem) WorkspaceWithMetaDTO {
	return WorkspaceWithMetaDTO{
		WorkspaceDTO: workspaceToDTO(&item.Workspace),
		MemberCount:  item.MemberCount,
		BoardCount:   item.BoardCount,
		Role:         item.Role,
	}
}

func workspaceMemberToDTO(member *entity.WorkspaceMember, user *entity.User) WorkspaceMemberDTO {
	return WorkspaceMemberDTO{
		UserID:    user.ID,
		Email:     user.Email,
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
		Role:      member.Role,
		JoinedAt:  member.JoinedAt,
	}
}
