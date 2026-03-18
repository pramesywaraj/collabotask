package response

import (
	"collabotask/internal/domain/entity"
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
