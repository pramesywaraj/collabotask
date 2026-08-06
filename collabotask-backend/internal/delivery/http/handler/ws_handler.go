package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"collabotask/internal/delivery/http/middleware"
	"collabotask/internal/realtime"
	"collabotask/internal/usecase/common"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// WSHandler upgrades HTTP → WebSocket and drives the room lifecycle.
// The Hub is INJECTED (standalone ProvideHub singleton). The handler supplies the
// onPresence + message callbacks and wires the former via hub.SetPresence.
type WSHandler struct {
	hub            *realtime.Hub
	access         common.BoardAccessChecker
	originPatterns []string
}

func NewWSHandler(hub *realtime.Hub, access common.BoardAccessChecker, originPatterns []string) *WSHandler {
	h := &WSHandler{
		hub:            hub,
		access:         access,
		originPatterns: originPatterns,
	}
	hub.SetPresence(h.onPresence) // once, at startup, before any conn registers
	return h
}

// ServeWS upgrades the connection, registers it with the hub, and blocks until
// the connection closes. Auth is already enforced by middleware.Auth upstream.
func (h *WSHandler) ServeWS(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// Clear this connection's write deadline — the global WriteTimeout would
	// kill a long-lived idle socket. REST connections keep the 30s deadline.
	rc := http.NewResponseController(c.Writer)
	_ = rc.SetWriteDeadline(time.Time{})
	_ = rc.SetReadDeadline(time.Time{})

	wsConn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		OriginPatterns: h.originPatterns,
	})
	if err != nil {
		// websocket.Accept has already written the HTTP error response.
		return
	}
	// Safety net only: teardown does a graceful Close(StatusNormalClosure); this
	// CloseNow is the belt-and-suspenders backstop if we return on an unexpected path.
	// Calling CloseNow after a graceful Close is a no-op — not a double-close bug.
	defer wsConn.CloseNow()

	s := realtime.NewWSSocket(wsConn)
	conn := h.hub.Register(c.Request.Context(), userID, s, h.handleMessage)

	<-conn.Done() // block until pumps finish + teardown
}

// dbJoinCheckTimeout bounds the JOIN_BOARD access check (3 DB round-trips) so a
// slow/hung DB can't strand this connection's readPump.
const dbJoinCheckTimeout = 5 * time.Second

// handleMessage is the readPump callback for JOIN_BOARD and LEAVE_BOARD frames.
// ctx is the pump's context (used to bound the access check).
func (h *WSHandler) handleMessage(ctx context.Context, conn *realtime.Conn, data []byte) {
	var msg realtime.IncomingMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return // discard malformed frames
	}
	switch msg.Type {
	case realtime.MsgTypeJoinBoard:
		h.handleJoin(ctx, conn, msg.BoardID)
	case realtime.MsgTypeLeaveBoard:
		h.handleLeave(conn, msg.BoardID)
	}
}

// handleJoin re-validates access and admits the conn to the room.
//
// On denial (or timeout) it returns SILENTLY — no ACCESS_REVOKED, no error frame.
// This is deliberate: the authoritative access decision + user-facing feedback
// is the REST kanban fetch (same CheckViewAccess). The client MUST gate the board
// view on that REST response, never on receiving ACTIVE_USERS. Silence also honours
// the 404-hide invariant (an error frame would leak that a private board exists).
func (h *WSHandler) handleJoin(ctx context.Context, conn *realtime.Conn, boardID uuid.UUID) {
	ctx, cancel := context.WithTimeout(ctx, dbJoinCheckTimeout)
	defer cancel()

	if _, err := h.access.CheckViewAccess(ctx, boardID, conn.UserID()); err != nil {
		return // denied or timed out — silently ignore
	}
	snapshot := h.hub.Join(boardID, conn)
	frame, _ := json.Marshal(realtime.ActiveUsersFrame{
		Type:    realtime.FrameTypeActiveUsers,
		BoardID: boardID,
		UserIDs: snapshot,
	})
	conn.Send(frame) // send ACTIVE_USERS to the joining conn only
}

// handleLeave removes the conn from the room. The 1→0 presence edge (if it fires)
// will broadcast USER_LEFT via onPresence.
func (h *WSHandler) handleLeave(conn *realtime.Conn, boardID uuid.UUID) {
	h.hub.Leave(boardID, conn)
}

// onPresence is the Hub's edge callback. Called outside the hub lock.
// Broadcasts USER_JOINED or USER_LEFT to the room (including the joining user's other tabs).
//
// Ordering note: on a 0→1 join this fires INSIDE hub.Join, so the joiner receives
// USER_JOINED{self} on its own channel BEFORE handleJoin sends ACTIVE_USERS. That's
// fine by contract — the client treats ACTIVE_USERS as an authoritative snapshot
// (replace) and USER_JOINED/USER_LEFT as idempotent deltas, so a self-echo before
// the snapshot is a no-op.
func (h *WSHandler) onPresence(boardID, userID uuid.UUID, kind realtime.PresenceKind) {
	var frameType string
	switch kind {
	case realtime.PresenceJoined:
		frameType = realtime.FrameTypeUserJoined
	case realtime.PresenceLeft:
		frameType = realtime.FrameTypeUserLeft
	default:
		return
	}
	frame, _ := json.Marshal(realtime.UserPresenceFrame{
		Type:    frameType,
		BoardID: boardID,
		UserID:  userID,
	})
	h.hub.Broadcast(boardID, frame)
}
