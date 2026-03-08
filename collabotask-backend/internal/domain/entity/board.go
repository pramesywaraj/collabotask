package entity

import (
	"time"

	"github.com/google/uuid"
)

type BoardAccessStatus string

const (
	BoardJoined  BoardAccessStatus = "JOINED"
	BoardCanJoin BoardAccessStatus = "CAN_JOIN"
)

type Board struct {
	ID              uuid.UUID `json:"id" db:"id"`
	WorkspaceID     uuid.UUID `json:"workspace_id" db:"workspace_id"`
	Title           string    `json:"title" db:"title"`
	Description     *string   `json:"description" db:"description"`
	CreatedBy       uuid.UUID `json:"created_by" db:"created_by"`
	IsArchived      bool      `json:"is_archived" db:"is_archived"`
	BackgroundColor string    `json:"background_color" db:"background_color"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type BoardListItem struct {
	Board

	UserRole     BoardRole         `json:"user_role"`
	AccessStatus BoardAccessStatus `json:"access_status"`
	MemberCount  uint              `json:"member_count"`
}

func (Board) TableName() string {
	return "boards"
}

func (b *Board) IsEmpty() bool {
	return b.ID == uuid.Nil
}
