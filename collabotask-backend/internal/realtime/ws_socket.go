package realtime

import (
	"context"

	"github.com/coder/websocket"
)

type wsSocket struct {
	conn *websocket.Conn
}

// NewWSSocket is exported so the handler package can build the adapter.
func NewWSSocket(conn *websocket.Conn) socket {
	return &wsSocket{conn: conn}
}

func (w *wsSocket) Write(ctx context.Context, data []byte) error {
	return w.conn.Write(ctx, websocket.MessageText, data)
}

func (w *wsSocket) Read(ctx context.Context) ([]byte, error) {
	_, data, err := w.conn.Read(ctx)
	return data, err
}

func (w *wsSocket) Ping(ctx context.Context) error {
	return w.conn.Ping(ctx)
}

func (w *wsSocket) Close(reason string) error {
	return w.conn.Close(websocket.StatusNormalClosure, reason)
}
