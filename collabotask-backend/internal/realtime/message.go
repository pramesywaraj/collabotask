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

// IncomingMessage is the minimal envelope for client → server frames.
// Unmarshal only; validate BoardID before acting.
type IncomingMessage struct {
	Type    MsgType   `json:"type"`
	BoardID uuid.UUID `json:"board_id"`
}

// FrameType identifies an outgoing server → client frame. Named for the same
// compiler-safety reason as MsgType.
type FrameType string

// Outgoing presence frame types (server → client).
const (
	FrameTypeActiveUsers FrameType = "ACTIVE_USERS"
	FrameTypeUserJoined  FrameType = "USER_JOINED"
	FrameTypeUserLeft    FrameType = "USER_LEFT"
)

// ActiveUsersFrame is sent to the joining conn only on JOIN_BOARD success.
type ActiveUsersFrame struct {
	Type    FrameType   `json:"type"`
	BoardID uuid.UUID   `json:"board_id"`
	UserIDs []uuid.UUID `json:"user_ids"`
}

// UserPresenceFrame is broadcast to the room on a 0→1 (USER_JOINED) or 1→0 (USER_LEFT) edge.
type UserPresenceFrame struct {
	Type    FrameType `json:"type"`
	BoardID uuid.UUID `json:"board_id"`
	UserID  uuid.UUID `json:"user_id"`
}
