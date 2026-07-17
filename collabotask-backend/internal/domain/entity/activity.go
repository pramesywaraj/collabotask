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

// ActivityMeta* are the canonical metadata keys for Activity.Metadata (SRS §4.6).
// Using constants prevents key-name drift across call sites and tests.
const (
	ActivityMetaCardTitle       = "card_title"
	ActivityMetaColumnTitle     = "column_title"
	ActivityMetaBoardTitle      = "board_title"
	ActivityMetaChangedFields   = "changed_fields"
	ActivityMetaFromColumnID    = "from_column_id"
	ActivityMetaFromColumnTitle = "from_column_title"
	ActivityMetaToColumnID      = "to_column_id"
	ActivityMetaToColumnTitle   = "to_column_title"
	ActivityMetaBreakGlass      = "break_glass"
	ActivityMetaRole            = "role"
	ActivityMetaSource          = "source"
	ActivityMetaFromUserID      = "from_user_id"
	ActivityMetaToUserID        = "to_user_id"
	ActivityMetaVisibilityFrom  = "visibility_from"
	ActivityMetaVisibilityTo    = "visibility_to"
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
