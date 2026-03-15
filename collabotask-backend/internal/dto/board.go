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

	UserRole     *entity.BoardRole
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

	UserRole     *entity.BoardRole
	AccessStatus entity.BoardAccessStatus
	Members      []BoardMemberDTO
}

func BoardToDTO(board *entity.Board) BoardDTO {
	return BoardDTO{
		ID:              board.ID,
		WorkspaceID:     board.WorkspaceID,
		Title:           board.Title,
		Description:     board.Description,
		CreatedBy:       board.CreatedBy,
		IsArchived:      board.IsArchived,
		BackgroundColor: board.BackgroundColor,
		CreatedAt:       board.CreatedAt,
		UpdatedAt:       board.UpdatedAt,
	}
}

func BoardMemberToDTO(member *entity.BoardMember, user *entity.User) BoardMemberDTO {
	return BoardMemberDTO{
		UserID:    member.UserID,
		Email:     user.Email,
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
		Role:      member.Role,
		JoinedAt:  member.JoinedAt,
	}
}
