package common

import "github.com/google/uuid"

// FrameType is the outgoing "type" tag of a server→client mutation frame (§5.2).
// Presence frame types (USER_JOINED/USER_LEFT/ACTIVE_USERS) live in the realtime
// package; they are emitted by the hub, not through this port.
type FrameType string

const (
	FrameCardCreated  FrameType = "CARD_CREATED"
	FrameCardMoved    FrameType = "CARD_MOVED"
	FrameCardUpdated  FrameType = "CARD_UPDATED"
	FrameCardDeleted  FrameType = "CARD_DELETED"
	FrameColumnCreated FrameType = "COLUMN_CREATED"
	FrameColumnUpdated FrameType = "COLUMN_UPDATED"
	FrameColumnDeleted FrameType = "COLUMN_DELETED"
	FrameColumnMoved   FrameType = "COLUMN_MOVED"
	FrameMemberAdded          FrameType = "MEMBER_ADDED"
	FrameMemberRemoved        FrameType = "MEMBER_REMOVED"
	FrameOwnershipTransferred FrameType = "OWNERSHIP_TRANSFERRED"
	FrameBoardUpdated         FrameType = "BOARD_UPDATED"
	FrameBoardArchived        FrameType = "BOARD_ARCHIVED"
	FrameBoardUnarchived      FrameType = "BOARD_UNARCHIVED"
	// FrameAccessRevoked is sent only to the evicted connection (not broadcast to the room).
	// A non-empty reason in EvictUser/EvictExcept triggers this. Voluntary leaves pass "".
	FrameAccessRevoked FrameType = "ACCESS_REVOKED"
)

// Event is a typed realtime mutation event. Concrete structs live in
// broadcast_events.go and carry *entity.* — the transport (envelope + payload
// shape) is the adapter's job, so the usecase never learns the wire format.
type Event interface {
	FrameType() FrameType
}

// Broadcaster is the usecase-layer port over the realtime hub. Best-effort,
// after-commit, swallow-on-error — identical contract to WriteActivity: a failed
// broadcast NEVER fails or rolls back the mutation. The concrete adapter lives in
// internal/realtime/broadcast.
//
// Part C uses only Broadcast. EvictUser/EvictExcept/EvictUserFromRooms are the
// settled surface (index §1) and the hub already implements them (Part A); they
// are declared here now so Parts D/E add call sites without re-touching the port
// or the adapter. See Decision C4.
type Broadcaster interface {
	Broadcast(boardID uuid.UUID, event Event)
	EvictUser(boardID, userID uuid.UUID, reason string)
	EvictExcept(boardID uuid.UUID, allowed []uuid.UUID, reason string)
	EvictUserFromRooms(userID uuid.UUID, boardIDs []uuid.UUID, reason string)
}
