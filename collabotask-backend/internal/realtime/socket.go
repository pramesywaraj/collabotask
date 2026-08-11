package realtime

import "context"

// socket is the minimal transport the pumps drive. Part B wraps
// *coder/websocket.Conn; tests use fakeSocket.
// Shaped to match coder/websocket so the Part B adapter is trivial.
type socket interface {
	Write(ctx context.Context, data []byte) error
	// Read blocks until a message arrives or the connection closes.
	// A non-nil error signals closure; the error value is discarded by
	// Part A (closure detection only — parsing lands in Part B).
	Read(ctx context.Context) ([]byte, error)
	// Ping sends a ping and blocks until the pong arrives or ctx is cancelled.
	// Used by writePump to detect half-open connections.
	Ping(ctx context.Context) error
	Close(reason string) error
}
