package broadcast

import (
	"context"
	"encoding/json"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"collabotask/internal/realtime"
	"collabotask/internal/usecase/common"
)

// stubSocket is a minimal realtime socket for adapter-level eviction tests: it
// records writes and blocks Read until disconnect(). Mirrors the hub's fakeSocket
// (which lives in the realtime_test package and is not importable here).
type stubSocket struct {
	mu      sync.Mutex
	written [][]byte
	readErr chan error
}

func newStubSocket() *stubSocket {
	return &stubSocket{readErr: make(chan error, 1)}
}

func (s *stubSocket) disconnect() {
	select {
	case s.readErr <- io.EOF:
	default:
	}
}

func (s *stubSocket) Write(_ context.Context, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := make([]byte, len(data))
	copy(cp, data)
	s.written = append(s.written, cp)
	return nil
}

func (s *stubSocket) Read(_ context.Context) ([]byte, error) { return nil, <-s.readErr }
func (s *stubSocket) Ping(_ context.Context) error           { return nil }
func (s *stubSocket) Close(string) error                     { return nil }

func (s *stubSocket) writtenCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.written)
}

func (s *stubSocket) frameTypes(t *testing.T) []string {
	t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	types := make([]string, 0, len(s.written))
	for _, raw := range s.written {
		var env struct {
			Type string `json:"type"`
		}
		require.NoError(t, json.Unmarshal(raw, &env))
		types = append(types, env.Type)
	}
	return types
}

// A non-silent reason sends one targeted ACCESS_REVOKED per board the user is in,
// before the hub tears the connection down.
func TestEvictUserFromRooms_NonSilent_SendsAccessRevokedPerBoard(t *testing.T) {
	ctx := context.Background()
	h := realtime.NewHub()
	b := NewHubBroadcaster(h)

	board1, board2 := uuid.New(), uuid.New()
	userID := uuid.New()

	s := newStubSocket()
	defer s.disconnect()
	c := h.Register(ctx, userID, s, nil)
	h.Join(board1, c)
	h.Join(board2, c)

	b.EvictUserFromRooms(userID, []uuid.UUID{board1, board2}, common.EvictReasonRemovedFromWorkspace)

	require.Eventually(t, func() bool {
		return s.writtenCount() == 2
	}, 2*time.Second, 5*time.Millisecond, "expected one ACCESS_REVOKED per board")

	for _, ft := range s.frameTypes(t) {
		assert.Equal(t, "ACCESS_REVOKED", ft)
	}
}

// EvictReasonSilent (voluntary leave) evicts without any ACCESS_REVOKED frame.
func TestEvictUserFromRooms_Silent_NoAccessRevoked(t *testing.T) {
	ctx := context.Background()
	h := realtime.NewHub()
	b := NewHubBroadcaster(h)

	board1, board2 := uuid.New(), uuid.New()
	userID := uuid.New()

	s := newStubSocket()
	defer s.disconnect()
	c := h.Register(ctx, userID, s, nil)
	h.Join(board1, c)
	h.Join(board2, c)

	b.EvictUserFromRooms(userID, []uuid.UUID{board1, board2}, common.EvictReasonSilent)

	// Wait for teardown; a silent eviction still closes the connection.
	select {
	case <-c.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for connection teardown")
	}
	assert.Equal(t, 0, s.writtenCount(), "silent eviction must not send ACCESS_REVOKED")
}
