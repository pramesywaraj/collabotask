package realtime

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/coder/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// wsTestPair creates a matched pair with NO background server reader.
// Good for TestWSSocket_Write (server calls Read explicitly) and
// TestWSSocket_Read (server calls Write, client calls Read).
func wsTestPair(t *testing.T) (*wsSocket, *websocket.Conn) {
	t.Helper()
	serverCh := make(chan *websocket.Conn, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Logf("server accept: %v", err)
			return
		}
		serverCh <- c
	}))
	t.Cleanup(srv.Close)

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	clientConn, _, err := websocket.Dial(context.Background(), wsURL, nil)
	require.NoError(t, err)
	t.Cleanup(func() { clientConn.CloseNow() })

	server := <-serverCh
	t.Cleanup(func() { server.CloseNow() })
	return &wsSocket{conn: clientConn}, server
}

func TestWSSocket_Write(t *testing.T) {
	client, server := wsTestPair(t)

	require.NoError(t, client.Write(context.Background(), []byte("hello")))

	_, got, err := server.Read(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte("hello"), got)
}

func TestWSSocket_Read(t *testing.T) {
	client, server := wsTestPair(t)

	require.NoError(t, server.Write(context.Background(), websocket.MessageText, []byte("world")))

	got, err := client.Read(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte("world"), got)
}

// TestWSSocket_Ping: coder/websocket.Ping docs say "Ping must be called
// concurrently with Reader as it does not read from the connection but instead
// waits for a Reader call to read the pong." CloseRead is that concurrent
// reader on both sides.
func TestWSSocket_Ping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		<-c.CloseRead(context.Background()).Done()
	}))
	t.Cleanup(srv.Close)

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	clientConn, _, err := websocket.Dial(context.Background(), wsURL, nil)
	require.NoError(t, err)
	t.Cleanup(func() { clientConn.CloseNow() })

	// Client also needs a concurrent reader to receive the pong frame.
	clientConn.CloseRead(context.Background())

	err = (&wsSocket{conn: clientConn}).Ping(context.Background())
	assert.NoError(t, err)
}

// TestWSSocket_Close_NormalClosure: conn.Close does a 2-way close handshake;
// the server must read the CLOSE frame and respond. CloseRead on the server
// handles this automatically.
func TestWSSocket_Close_NormalClosure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		<-c.CloseRead(context.Background()).Done()
	}))
	t.Cleanup(srv.Close)

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	clientConn, _, err := websocket.Dial(context.Background(), wsURL, nil)
	require.NoError(t, err)
	t.Cleanup(func() { clientConn.CloseNow() })

	client := &wsSocket{conn: clientConn}
	require.NoError(t, client.Close("test done"))

	// After a graceful close the connection is done; subsequent reads must error.
	_, err = client.Read(context.Background())
	assert.Error(t, err, "Read on a closed connection must error")
}
