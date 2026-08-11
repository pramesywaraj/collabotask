package realtime

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	sendBufferSize = 16
	pingInterval   = 30 * time.Second // how often the server pings an idle socket
	pongWait       = 10 * time.Second // per-ping deadline: no pong in this window ⇒ half-open ⇒ drop
)

// Conn is one WebSocket connection (one browser tab).
// The writePump is the sole writer to s — satisfying coder/websocket's
// single-writer constraint. The readPump detects closure and triggers teardown.
type Conn struct {
	connID uuid.UUID
	userID uuid.UUID
	s      socket

	send chan []byte
	done chan struct{} // closed by teardown
	once sync.Once

	// rooms this conn has joined; guarded by the hub's Lock (never modified alone).
	rooms map[uuid.UUID]struct{}
}

func newConn(userID uuid.UUID, s socket) *Conn {
	return &Conn{
		connID: uuid.New(),
		userID: userID,
		s:      s,
		send:   make(chan []byte, sendBufferSize),
		done:   make(chan struct{}),
		rooms:  make(map[uuid.UUID]struct{}),
	}
}

// UserID exposes the authenticated user for access checks in the handler.
func (c *Conn) UserID() uuid.UUID { return c.userID }

// Done returns a channel that is closed when the connection is torn down.
func (c *Conn) Done() <-chan struct{} { return c.done }

// Send queues msg on the send channel (non-blocking; drops if buffer full).
// The writePump is the sole goroutine that actually writes to the socket.
func (c *Conn) Send(msg []byte) {
	select {
	case c.send <- msg:
	default: // treat same as slow-consumer; hub will drop the conn if Broadcast fills it
	}
}

// MessageHandler is called by readPump for each inbound message. It receives the
// pump's ctx so a handler that does I/O (e.g. the JOIN_BOARD access check) can bound
// it. The handler runs inside the readPump goroutine, so it must not block
// *unboundedly* — one slow message stalls only THIS connection's inbound processing,
// but it must still finish (hence the bounded ctx in handleJoin).
type MessageHandler func(ctx context.Context, c *Conn, data []byte)

// writePump drains the send channel and writes to the socket.
// It is the only goroutine that calls s.Write. Drives server-side pings.
func (c *Conn) writePump(ctx context.Context, onClose func(reason string)) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()
	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				// channel closed by teardown
				onClose("")
				return
			}
			if err := c.s.Write(ctx, msg); err != nil {
				onClose("")
				return
			}
		case <-ticker.C:
			// Bound each ping: coder/websocket's Ping blocks until the pong arrives
			// or ctx is done. The socket's read/write deadlines were cleared in
			// ServeWS, so WITHOUT this timeout a half-open socket would block the
			// writePump until the request ctx dies — defeating liveness on a quiet
			// board (no broadcasts to trip the slow-consumer drop).
			pingCtx, cancel := context.WithTimeout(ctx, pongWait)
			err := c.s.Ping(pingCtx)
			cancel()
			if err != nil {
				onClose("")
				return
			}
		}
	}
}

// readPump reads from the socket; any error signals closure.
// onMsg (if non-nil) is called for each inbound message; ctx is the pump's
// context (always non-nil — callers must not pass a nil context to Register).
func (c *Conn) readPump(ctx context.Context, onMsg MessageHandler, onClose func(string)) {
	for {
		data, err := c.s.Read(ctx)
		if err != nil {
			break
		}
		if onMsg != nil {
			onMsg(ctx, c, data)
		}
	}
	onClose("")
}

// teardown is idempotent: once closes send, closes the socket, closes done.
// hub must have already unregistered the conn before calling this.
func (c *Conn) teardown(reason string) {
	c.once.Do(func() {
		close(c.send)
		_ = c.s.Close(reason)
		close(c.done)
	})
}
