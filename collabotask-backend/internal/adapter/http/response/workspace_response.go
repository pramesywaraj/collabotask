package response

import (
	"collabotask/internal/domain/entity"
	"collabotask/internal/dto"
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

func WorkspaceDTOToResponse(d dto.WorkspaceDTO) WorkspaceResponse {
	return WorkspaceResponse{
		ID:          d.ID,
		Name:        d.Name,
		Description: d.Description,
		OwnerID:     d.OwnerID,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

func WorkspaceWithMetaDTOToResponse(d dto.WorkspaceWithMetaDTO) WorkspaceWithMetaResponse {
	return WorkspaceWithMetaResponse{
		WorkspaceResponse: WorkspaceDTOToResponse(d.WorkspaceDTO),
		MemberCount:       d.MemberCount,
		BoardCount:        d.BoardCount,
		Role:              d.Role,
	}
}

func WorkspaceMemberDTOToResponse(d dto.WorkspaceMemberDTO) WorkspaceMemberResponse {
	return WorkspaceMemberResponse{
		UserID:    d.UserID,
		Email:     d.Email,
		Name:      d.Name,
		AvatarURL: d.AvatarURL,
		Role:      d.Role,
		JoinedAt:  d.JoinedAt,
	}
}

func WorkspaceDetailDTOToResponse(d dto.WorkspaceDetailDTO) WorkspaceDetailResponse {
	members := make([]WorkspaceMemberResponse, 0, len(d.Members))
	for _, member := range d.Members {
		members = append(members, WorkspaceMemberDTOToResponse(member))
	}

	return WorkspaceDetailResponse{
		WorkspaceResponse: WorkspaceDTOToResponse(d.WorkspaceDTO),
		UserRole:          d.UserRole,
		Members:           members,
	}
}
