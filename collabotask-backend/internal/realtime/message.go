package realtime

import "github.com/google/uuid"

// MsgType identifies an incoming client → server frame. A named type (over a bare
// string) lets the compiler catch a wrong constant at the JOIN/LEAVE switch.
type MsgType string

// Incoming message types (client → server).
const (
	MsgTypeJoinBoard  MsgType = "JOIN_BOARD"
	MsgTypeLeaveBoard MsgType = "LEAVE_BOARD"
)

// FrameType identifies an outgoing server → client frame. Named for the same
// compiler-safety reason as MsgType.
type FrameType string

// Outgoing presence frame types (server → client).
const (
	FrameTypeActiveUsers FrameType = "ACTIVE_USERS"
	FrameTypeUserJoined  FrameType = "USER_JOINED"
	FrameTypeUserLeft    FrameType = "USER_LEFT"
)

// PresenceUser carries the profile fields embedded in presence frames (ADR-014).
// AvatarURL is nil when the user has no avatar set.
type PresenceUser struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	AvatarURL *string   `json:"avatar_url"`
}

// IncomingMessage is the minimal envelope for client → server frames.
// Unmarshal only; validate BoardID before acting.
type IncomingMessage struct {
	Type    MsgType   `json:"type"`
	BoardID uuid.UUID `json:"board_id"`
}

// ActiveUsersFrame is sent to the joining conn only on JOIN_BOARD success.
// Users carries the enriched profiles of every connected user in the room.
type ActiveUsersFrame struct {
	Type    FrameType      `json:"type"`
	BoardID uuid.UUID      `json:"board_id"`
	Users   []PresenceUser `json:"users"`
}

// UserJoinedFrame is broadcast to the room on a 0→1 presence edge.
// User carries the enriched profile of the joining user (ADR-014).
type UserJoinedFrame struct {
	Type    FrameType    `json:"type"`
	BoardID uuid.UUID    `json:"board_id"`
	User    PresenceUser `json:"user"`
}

// UserLeftFrame is broadcast to the room on a 1→0 presence edge.
// Stays thin — removal keys on user_id; no profile needed to drop an avatar.
type UserLeftFrame struct {
	Type    FrameType `json:"type"`
	BoardID uuid.UUID `json:"board_id"`
	UserID  uuid.UUID `json:"user_id"`
}
