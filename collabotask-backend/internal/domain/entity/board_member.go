package entity

import (
	"time"

	"github.com/google/uuid"
)

type BoardRole string

const (
	BoardRoleOwner  BoardRole = "BOARD_OWNER"
	BoardRoleMember BoardRole = "BOARD_MEMBER"
)

type BoardMember struct {
	BoardID  uuid.UUID `json:"board_id" db:"board_id"`
	UserID   uuid.UUID `json:"user_id" db:"user_id"`
	Role     BoardRole `json:"role" db:"role"`
	JoinedAt time.Time `json:"joined_at" db:"joined_at"`
}

func (BoardMember) TableName() string {
	return "board_members"
}

func (bm *BoardMember) IsEmpty() bool {
	return bm.BoardID == uuid.Nil && bm.UserID == uuid.Nil
}

func (bm *BoardMember) IsOwner() bool {
	return bm.Role == BoardRoleOwner
}

func (bm *BoardMember) IsMember() bool {
	return bm.Role == BoardRoleMember
}
