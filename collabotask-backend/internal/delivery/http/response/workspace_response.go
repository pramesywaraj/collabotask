package response

import (
	"collabotask/internal/domain/entity"
	"collabotask/internal/usecase/workspace"
	"time"

	"github.com/google/uuid"
)

type WorkspaceResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	OwnerID     uuid.UUID `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type WorkspaceWithMetaResponse struct {
	WorkspaceResponse

	MemberCount uint                 `json:"member_count"`
	BoardCount  uint                 `json:"board_count"`
	Role        entity.WorkspaceRole `json:"role"`
}

type WorkspaceMemberResponse struct {
	UserID    uuid.UUID            `json:"id"`
	Email     string               `json:"email"`
	Name      string               `json:"name"`
	AvatarURL *string              `json:"avatar_url"`
	Role      entity.WorkspaceRole `json:"role"`
	JoinedAt  time.Time            `json:"joined_at"`
}

type WorkspaceDetailResponse struct {
	WorkspaceResponse
	UserRole entity.WorkspaceRole      `json:"user_role"`
	Members  []WorkspaceMemberResponse `json:"members"`
}

func WorkspaceToResponse(w *entity.Workspace) WorkspaceResponse {
	return WorkspaceResponse{
		ID:          w.ID,
		Name:        w.Name,
		Description: w.Description,
		OwnerID:     w.OwnerID,
		CreatedAt:   w.CreatedAt,
		UpdatedAt:   w.UpdatedAt,
	}
}

func WorkspaceWithMetaToResponse(w workspace.WorkspaceWithMeta) WorkspaceWithMetaResponse {
	return WorkspaceWithMetaResponse{
		WorkspaceResponse: WorkspaceToResponse(w.Workspace),
		MemberCount:       w.MemberCount,
		BoardCount:        w.BoardCount,
		Role:              w.Role,
	}
}

func WorkspaceMemberToResponse(m workspace.WorkspaceMember) WorkspaceMemberResponse {
	return WorkspaceMemberResponse{
		UserID:    m.UserID,
		Email:     m.Email,
		Name:      m.Name,
		AvatarURL: m.AvatarURL,
		Role:      m.Role,
		JoinedAt:  m.JoinedAt,
	}
}

func WorkspaceDetailToResponse(d workspace.WorkspaceDetail) WorkspaceDetailResponse {
	members := make([]WorkspaceMemberResponse, 0, len(d.Members))
	for _, member := range d.Members {
		members = append(members, WorkspaceMemberToResponse(member))
	}

	return WorkspaceDetailResponse{
		WorkspaceResponse: WorkspaceToResponse(d.Workspace),
		UserRole:          d.UserRole,
		Members:           members,
	}
}
