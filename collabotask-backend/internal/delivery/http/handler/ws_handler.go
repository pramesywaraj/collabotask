package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"collabotask/internal/delivery/http/middleware"
	"collabotask/internal/domain/entity"
	"collabotask/internal/domain/repository"
	"collabotask/internal/realtime"
	"collabotask/internal/usecase/common"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// WSHandler upgrades HTTP → WebSocket and drives the room lifecycle.
// The Hub is INJECTED (standalone ProvideHub singleton). The handler supplies the
// onPresence + message callbacks and wires the former via hub.SetPresence.
type WSHandler struct {
	hub            *realtime.Hub
	access         common.BoardAccessChecker
	userRepo       repository.UserRepository
	originPatterns []string
}

func NewWSHandler(hub *realtime.Hub, access common.BoardAccessChecker, userRepo repository.UserRepository, originPatterns []string) *WSHandler {
	h := &WSHandler{
		hub:            hub,
		access:         access,
		userRepo:       userRepo,
		originPatterns: originPatterns,
	}
	hub.SetPresence(h.onPresence) // once, at startup, before any conn registers
	return h
}

// ServeWS godoc
// @Summary      Open a WebSocket connection
// @Description  Upgrades the HTTP connection to WebSocket. Requires an authenticated httpOnly cookie. Origin is validated against the CORS allowlist (OriginPatterns). The CSRF header is not required for this GET — CSWSH defense is provided by OriginPatterns + SameSite=Lax.
// @Tags         realtime
// @Success      101 {string} string "Switching Protocols"
// @Failure      401 {string} string "Unauthorized"
// @Router       /ws [get]
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
	// Enrich the snapshot with profiles; piggybacks the CheckViewAccess round-trip
	// already in this bounded context window (ADR-014).
	users := h.enrichUsers(ctx, snapshot)
	frame, _ := json.Marshal(realtime.ActiveUsersFrame{
		Type:    realtime.FrameTypeActiveUsers,
		BoardID: boardID,
		Users:   users,
	})
	conn.Send(frame) // send ACTIVE_USERS to the joining conn only
}

// handleLeave removes the conn from the room. The 1→0 presence edge (if it fires)
// will broadcast USER_LEFT via onPresence.
func (h *WSHandler) handleLeave(conn *realtime.Conn, boardID uuid.UUID) {
	h.hub.Leave(boardID, conn)
}

// onPresence is the Hub's edge callback. Called outside the hub lock.
// Broadcasts USER_JOINED (rich) or USER_LEFT (thin) to the room.
//
// Ordering note: on a 0→1 join this fires INSIDE hub.Join, so the joiner receives
// USER_JOINED{self} on its own channel BEFORE handleJoin sends ACTIVE_USERS. That's
// fine by contract — the client treats ACTIVE_USERS as an authoritative snapshot
// (replace) and USER_JOINED/USER_LEFT as idempotent deltas, so a self-echo before
// the snapshot is a no-op.
func (h *WSHandler) onPresence(boardID, userID uuid.UUID, kind realtime.PresenceKind) {
	switch kind {
	case realtime.PresenceJoined:
		// One bounded, best-effort GetByIds per 0→1 edge (ADR-014: honest added cost,
		// low-frequency). Can't reuse handleJoin's context — this is a hub-wide callback.
		ctx, cancel := context.WithTimeout(context.Background(), dbJoinCheckTimeout)
		defer cancel()
		user := h.enrichSingleUser(ctx, userID)
		frame, _ := json.Marshal(realtime.UserJoinedFrame{
			Type:    realtime.FrameTypeUserJoined,
			BoardID: boardID,
			User:    user,
		})
		h.hub.Broadcast(boardID, frame)

	case realtime.PresenceLeft:
		// USER_LEFT stays thin — removal keys on user_id; no profile needed (ADR-014).
		frame, _ := json.Marshal(realtime.UserLeftFrame{
			Type:    realtime.FrameTypeUserLeft,
			BoardID: boardID,
			UserID:  userID,
		})
		h.hub.Broadcast(boardID, frame)
	}
}

// enrichUsers maps a slice of user IDs to PresenceUser profiles via a single
// GetByIds call. Best-effort: on lookup error it logs, swallows, and returns
// degraded entries (ID only). Missing IDs in a partial map are skipped silently.
func (h *WSHandler) enrichUsers(ctx context.Context, ids []uuid.UUID) []realtime.PresenceUser {
	if len(ids) == 0 {
		return nil
	}
	userMap, err := h.userRepo.GetByIds(ctx, ids)
	if err != nil {
		log.Error().Err(err).Msg("presence ACTIVE_USERS enrichment failed (swallowed)")
		out := make([]realtime.PresenceUser, len(ids))
		for i, id := range ids {
			out[i] = realtime.PresenceUser{ID: id}
		}
		return out
	}
	out := make([]realtime.PresenceUser, 0, len(ids))
	for _, id := range ids {
		u, ok := userMap[id]
		if !ok {
			continue // partial map: skip missing entry, not a nil-deref
		}
		out = append(out, presenceUserFrom(u))
	}
	return out
}

// enrichSingleUser fetches the profile for one user ID. Best-effort: on lookup
// error or missing result it logs and returns a degraded entry (ID only).
func (h *WSHandler) enrichSingleUser(ctx context.Context, userID uuid.UUID) realtime.PresenceUser {
	userMap, err := h.userRepo.GetByIds(ctx, []uuid.UUID{userID})
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).
			Msg("presence USER_JOINED enrichment failed (swallowed)")
		return realtime.PresenceUser{ID: userID}
	}
	u, ok := userMap[userID]
	if !ok {
		return realtime.PresenceUser{ID: userID}
	}
	return presenceUserFrom(u)
}

func presenceUserFrom(u *entity.User) realtime.PresenceUser {
	return realtime.PresenceUser{
		ID:        u.ID,
		Name:      u.Name,
		AvatarURL: u.AvatarURL,
	}
}
