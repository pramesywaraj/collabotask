package realtime

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

const sendBufferSize = 16

// Conn is one WebSocket connection (one browser tab).
// The writePump is the sole writer to s — satisfying coder/websocket's
// single-writer constraint. The readPump detects closure and triggers teardown.
type Conn struct {
	connID uuid.UUID
	userID uuid.UUID
	s      socket

	send chan []byte
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
		rooms:  make(map[uuid.UUID]struct{}),
	}
}

// writePump drains the send channel and writes to the socket.
// It is the only goroutine that calls s.Write.
func (c *Conn) writePump(ctx context.Context, onClose func(reason string)) {
	for msg := range c.send {
		if err := c.s.Write(ctx, msg); err != nil {
			break
		}
	}
	onClose("")
}

// readPump reads from the socket; any error signals closure.
// Part A discards message bytes — parsing is Part B.
func (c *Conn) readPump(ctx context.Context, onClose func(reason string)) {
	for {
		if _, err := c.s.Read(ctx); err != nil {
			break
		}
	}
	onClose("")
}

// teardown is idempotent: once closes send, closes the socket.
// hub must have already unregistered the conn before calling this.
func (c *Conn) teardown(reason string) {
	c.once.Do(func() {
		close(c.send)
		_ = c.s.Close(reason)
	})
}
