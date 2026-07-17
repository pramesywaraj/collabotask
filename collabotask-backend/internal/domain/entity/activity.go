package entity

import (
	"time"

	"github.com/google/uuid"
)

type ActivityEntityType string

const (
	ActivityEntityBoard  ActivityEntityType = "BOARD"
	ActivityEntityColumn ActivityEntityType = "COLUMN"
	ActivityEntityCard   ActivityEntityType = "CARD"
	ActivityEntityMember ActivityEntityType = "MEMBER"
)

type ActivityActionType string

const (
	ActivityActionCreated              ActivityActionType = "CREATED"
	ActivityActionUpdated              ActivityActionType = "UPDATED"
	ActivityActionDeleted              ActivityActionType = "DELETED"
	ActivityActionMoved                ActivityActionType = "MOVED"
	ActivityActionArchived             ActivityActionType = "ARCHIVED"
	ActivityActionUnarchived           ActivityActionType = "UNARCHIVED"
	ActivityActionJoined               ActivityActionType = "JOINED"
	ActivityActionLeft                 ActivityActionType = "LEFT"
	ActivityActionAdded                ActivityActionType = "ADDED"
	ActivityActionRemoved              ActivityActionType = "REMOVED"
	ActivityActionOwnershipTransferred ActivityActionType = "OWNERSHIP_TRANSFERRED"
)

type Activity struct {
	ID         uuid.UUID
	BoardID    uuid.UUID
	UserID     *uuid.UUID     // actor; SET NULL-able on Phase-2 account deletion; never nil in Phase 1
	ActionType ActivityActionType
	EntityType ActivityEntityType
	EntityID   uuid.UUID
	Metadata   map[string]any // marshaled to jsonb
	CreatedAt  time.Time
}
