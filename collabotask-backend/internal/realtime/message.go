package realtime

import "github.com/google/uuid"

// Incoming message types (client → server).
const (
	MsgTypeJoinBoard  = "JOIN_BOARD"
	MsgTypeLeaveBoard = "LEAVE_BOARD"
)

// IncomingMessage is the minimal envelope for client → server frames.
// Unmarshal only; validate BoardID before acting.
type IncomingMessage struct {
	Type    string    `json:"type"`
	BoardID uuid.UUID `json:"board_id"`
}

// Outgoing presence frame types (server → client).
const (
	FrameTypeActiveUsers = "ACTIVE_USERS"
	FrameTypeUserJoined  = "USER_JOINED"
	FrameTypeUserLeft    = "USER_LEFT"
)

// ActiveUsersFrame is sent to the joining conn only on JOIN_BOARD success.
type ActiveUsersFrame struct {
	Type    string      `json:"type"`
	BoardID uuid.UUID   `json:"board_id"`
	UserIDs []uuid.UUID `json:"user_ids"`
}

// UserPresenceFrame is broadcast to the room on a 0→1 (USER_JOINED) or 1→0 (USER_LEFT) edge.
type UserPresenceFrame struct {
	Type    string    `json:"type"`
	BoardID uuid.UUID `json:"board_id"`
	UserID  uuid.UUID `json:"user_id"`
}
