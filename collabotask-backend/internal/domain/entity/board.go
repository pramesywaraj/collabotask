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

type BoardVisibility string

const (
	BoardVisibilityWorkspace BoardVisibility = "WORKSPACE"
	BoardVisibilityPrivate   BoardVisibility = "PRIVATE"
)

type Board struct {
	ID          uuid.UUID `json:"id" db:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id" db:"workspace_id"`
	Title       string    `json:"title" db:"title"`
	Description *string   `json:"description" db:"description"`
	// CreatedBy stays uuid.UUID in Phase 1: it can never be NULL (no account
	// deletion). It becomes *uuid.UUID in Phase 2 alongside account deletion
	// (SRS §10) — the only feature that can null it (migration 000007 already
	// made the column nullable + ON DELETE SET NULL for forward-compat). It is
	// a stored/returned trace only; board_members is the source of ownership.
	CreatedBy       uuid.UUID       `json:"created_by" db:"created_by"`
	IsArchived      bool            `json:"is_archived" db:"is_archived"`
	BackgroundColor string          `json:"background_color" db:"background_color"`
	Visibility      BoardVisibility `json:"visibility" db:"visibility"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

type BoardListItem struct {
	Board

	UserRole     BoardRole         `json:"user_role"`
	AccessStatus BoardAccessStatus `json:"access_status"`
	MemberCount  uint              `json:"member_count"`
	CardCount    uint              `json:"card_count"`
}

func (Board) TableName() string {
	return "boards"
}

func (b *Board) IsEmpty() bool {
	return b.ID == uuid.Nil
}
