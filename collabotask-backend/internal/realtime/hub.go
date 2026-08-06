package realtime

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

// PresenceKind signals the direction of a user's presence edge in a room.
type PresenceKind int

const (
	PresenceJoined PresenceKind = iota
	PresenceLeft
)

// Hub is the in-memory connection registry.
// rooms: boardID → userID → connID → *Conn
// Concurrency: RLock for reads/broadcasts; Lock for mutations.
// The lock is never held during a socket write.
type Hub struct {
	mu    sync.RWMutex
	rooms map[uuid.UUID]map[uuid.UUID]map[uuid.UUID]*Conn

	// onPresence is called outside the lock on every 0→1 (Joined) or 1→0 (Left) edge.
	onPresence func(boardID, userID uuid.UUID, kind PresenceKind)
}

func NewHub() *Hub {
	return &Hub{rooms: make(map[uuid.UUID]map[uuid.UUID]map[uuid.UUID]*Conn)}
}

// SetPresence wires the edge callback. Called once at startup (in ProvideWSHandler),
// before any conn registers — so no lock is needed and there is no race.
func (h *Hub) SetPresence(fn func(boardID, userID uuid.UUID, kind PresenceKind)) {
	h.onPresence = fn
}

// Register creates a Conn for the socket and starts its pumps.
// onMsg is called by readPump for each inbound message (may be nil).
func (h *Hub) Register(ctx context.Context, userID uuid.UUID, s socket, onMsg MessageHandler) *Conn {
	c := newConn(userID, s)
	teardown := func(reason string) { h.unregisterConn(c, reason) }
	go c.readPump(ctx, onMsg, teardown)
	go c.writePump(ctx, teardown)
	return c
}

// Join adds c to the boardID room and returns a snapshot of distinct connected
// user IDs in the room (including the joiner). Fires onPresence on the 0→1 edge.
func (h *Hub) Join(boardID uuid.UUID, c *Conn) []uuid.UUID {
	h.mu.Lock()
	if h.rooms[boardID] == nil {
		h.rooms[boardID] = make(map[uuid.UUID]map[uuid.UUID]*Conn)
	}
	userConns := h.rooms[boardID]
	wasAbsent := len(userConns[c.userID]) == 0
	if userConns[c.userID] == nil {
		userConns[c.userID] = make(map[uuid.UUID]*Conn)
	}
	userConns[c.userID][c.connID] = c
	c.rooms[boardID] = struct{}{}
	snapshot := activeUserIDs(userConns)
	h.mu.Unlock()

	if wasAbsent && h.onPresence != nil {
		h.onPresence(boardID, c.userID, PresenceJoined)
	}
	return snapshot
}

// Leave removes c from the boardID room. Fires onPresence on the 1→0 edge.
func (h *Hub) Leave(boardID uuid.UUID, c *Conn) {
	h.mu.Lock()
	fired, _ := h.removeConnFromRoom(boardID, c)
	h.mu.Unlock()

	if fired && h.onPresence != nil {
		h.onPresence(boardID, c.userID, PresenceLeft)
	}
}

// ActiveUsers returns distinct connected user IDs in the room. Never consults DB.
func (h *Hub) ActiveUsers(boardID uuid.UUID) []uuid.UUID {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return activeUserIDs(h.rooms[boardID])
}

// Broadcast sends msg to every Conn in boardID's room.
// A full send buffer drops that Conn (slow-consumer policy). The lock is
// released before any socket write; non-blocking sends only under RLock.
func (h *Hub) Broadcast(boardID uuid.UUID, msg []byte) {
	h.mu.RLock()
	room := h.rooms[boardID]
	var toDrop []*Conn
	for _, conns := range room {
		for _, c := range conns {
			select {
			case c.send <- msg:
			default:
				toDrop = append(toDrop, c)
			}
		}
	}
	h.mu.RUnlock()

	for _, c := range toDrop {
		h.unregisterConn(c, "slow consumer")
	}
}

// EvictUser tears down all conns for userID in boardID.
func (h *Hub) EvictUser(boardID, userID uuid.UUID, reason string) {
	h.mu.Lock()
	var victims []*Conn
	if room := h.rooms[boardID]; room != nil {
		for _, c := range room[userID] {
			victims = append(victims, c)
		}
	}
	h.mu.Unlock()

	// tear down outside the lock (unregisterConn re-locks per conn)
	for _, c := range victims {
		h.unregisterConn(c, reason)
	}
}

// EvictExcept tears down every conn in boardID whose user is not in allowed.
func (h *Hub) EvictExcept(boardID uuid.UUID, allowed []uuid.UUID, reason string) {
	allowSet := make(map[uuid.UUID]struct{}, len(allowed))
	for _, id := range allowed {
		allowSet[id] = struct{}{}
	}

	h.mu.RLock()
	var victims []*Conn
	for userID, conns := range h.rooms[boardID] {
		if _, ok := allowSet[userID]; !ok {
			for _, c := range conns {
				victims = append(victims, c)
			}
		}
	}
	h.mu.RUnlock()

	for _, c := range victims {
		h.unregisterConn(c, reason)
	}
}

// EvictUserFromRooms tears down all conns for userID in each of the given boards.
func (h *Hub) EvictUserFromRooms(userID uuid.UUID, boardIDs []uuid.UUID, reason string) {
	for _, boardID := range boardIDs {
		h.EvictUser(boardID, userID, reason)
	}
}

// unregisterConn removes c from all its rooms and tears it down.
// Safe to call multiple times (teardown is sync.Once).
type presenceEdge struct {
	boardID uuid.UUID
	kind    PresenceKind
}

func (h *Hub) unregisterConn(c *Conn, reason string) {
	h.mu.Lock()
	var edges []presenceEdge
	for boardID := range c.rooms {
		if fired, kind := h.removeConnFromRoom(boardID, c); fired {
			edges = append(edges, presenceEdge{boardID, kind})
		}
	}
	h.mu.Unlock()

	c.teardown(reason)

	if h.onPresence != nil {
		for _, e := range edges {
			h.onPresence(e.boardID, c.userID, e.kind)
		}
	}
}

// removeConnFromRoom removes c from a single room. Caller must hold h.mu (Lock).
// Returns (true, PresenceLeft) if this was the user's last conn (1→0 edge).
func (h *Hub) removeConnFromRoom(boardID uuid.UUID, c *Conn) (edgeFired bool, kind PresenceKind) {
	room := h.rooms[boardID]
	if room == nil {
		return false, 0
	}
	delete(room[c.userID], c.connID)
	delete(c.rooms, boardID)
	if len(room[c.userID]) == 0 {
		delete(room, c.userID)
		if len(room) == 0 {
			delete(h.rooms, boardID)
		}
		return true, PresenceLeft
	}
	return false, 0
}

// activeUserIDs returns distinct user IDs from a room's user→conns map.
// Caller must hold at least h.mu.RLock.
func activeUserIDs(room map[uuid.UUID]map[uuid.UUID]*Conn) []uuid.UUID {
	if len(room) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, 0, len(room))
	for userID := range room {
		ids = append(ids, userID)
	}
	return ids
}
