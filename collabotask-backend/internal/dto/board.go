package dto

import (
	"collabotask/internal/domain/entity"
	"time"

	"github.com/google/uuid"
)

type BoardDTO struct {
	ID              uuid.UUID
	WorkspaceID     uuid.UUID
	Title           string
	Description     *string
	CreatedBy       uuid.UUID
	IsArchived      bool
	BackgroundColor string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type BoardWithMetaDTO struct {
	BoardDTO

	UserRole     entity.BoardRole
	AccessStatus entity.BoardAccessStatus
	MemberCount  uint
}

type BoardMemberDTO struct {
	UserID    uuid.UUID
	Email     string
	Name      string
	AvatarURL *string
	Role      entity.BoardRole
	JoinedAt  time.Time
}

type BoardDetailDTO struct {
	BoardDTO

	UserRole     entity.BoardRole
	AccessStatus entity.BoardAccessStatus
	Members      []BoardMemberDTO
}
