package handler

// White-box tests: package handler (not handler_test) so we can call the
// unexported handleMessage / handleJoin / handleLeave / onPresence methods.

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"collabotask/internal/mocks"
	"collabotask/internal/realtime"
	"collabotask/internal/usecase/common"

	"github.com/google/uuid"
	mock "github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// stubSocket — satisfies the unexported realtime.socket seam from this package.
// Go's structural typing allows this even though the interface is unexported.
// ---------------------------------------------------------------------------

type stubSocket struct {
	mu      sync.Mutex
	written [][]byte
	readErr chan error
}

func newStubSocket() *stubSocket {
	return &stubSocket{readErr: make(chan error, 1)}
}

func (s *stubSocket) Write(_ context.Context, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := make([]byte, len(data))
	copy(cp, data)
	s.written = append(s.written, cp)
	return nil
}

func (s *stubSocket) Read(_ context.Context) ([]byte, error) {
	return nil, <-s.readErr
}

func (s *stubSocket) Ping(_ context.Context) error { return nil }
func (s *stubSocket) Close(_ string) error          { return nil }

// disconnect unblocks the readPump (simulates remote close).
func (s *stubSocket) disconnect() {
	select {
	case s.readErr <- io.EOF:
	default:
	}
}

func (s *stubSocket) msgs() [][]byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([][]byte, len(s.written))
	copy(out, s.written)
	return out
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func newTestWSHandler(t *testing.T) (*WSHandler, *realtime.Hub, *mocks.MockBoardAccessChecker) {
	t.Helper()
	hub := realtime.NewHub()
	access := mocks.NewMockBoardAccessChecker(t)
	h := NewWSHandler(hub, access, nil)
	return h, hub, access
}

// registerTestConn registers a conn whose readPump calls h.handleMessage for
// every inbound frame.
func registerTestConn(ctx context.Context, hub *realtime.Hub, h *WSHandler, userID uuid.UUID) (*stubSocket, *realtime.Conn) {
	ss := newStubSocket()
	conn := hub.Register(ctx, userID, ss, h.handleMessage)
	return ss, conn
}

// listenerConn registers a bystander in the board so we can observe Broadcast
// output. NOTE: hub.Join fires the edge-triggered presence callback, so the
// listener receives its own USER_JOINED before any explicit onPresence call.
func listenerConn(ctx context.Context, hub *realtime.Hub, boardID uuid.UUID) (*stubSocket, func()) {
	ss := newStubSocket()
	conn := hub.Register(ctx, uuid.New(), ss, nil)
	hub.Join(boardID, conn)
	return ss, ss.disconnect
}

// findFrame unmarshals each raw msg looking for one whose Type field equals wantType.
func findFrameByType(msgs [][]byte, wantType realtime.FrameType) []byte {
	for _, raw := range msgs {
		var m map[string]any
		if json.Unmarshal(raw, &m) == nil {
			if m["type"] == string(wantType) {
				return raw
			}
		}
	}
	return nil
}

// findPresenceFrameByUser scans msgs for a UserPresenceFrame whose UserID == want.
func findPresenceFrameByUser(msgs [][]byte, want uuid.UUID) *realtime.UserPresenceFrame {
	for _, raw := range msgs {
		var f realtime.UserPresenceFrame
		if json.Unmarshal(raw, &f) == nil && f.UserID == want {
			return &f
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// handleMessage — routing
// ---------------------------------------------------------------------------

func TestHandleMessage_MalformedJSON_Discards(t *testing.T) {
	ctx := context.Background()
	h, hub, _ := newTestWSHandler(t)
	ss, conn := registerTestConn(ctx, hub, h, uuid.New())
	defer ss.disconnect()

	// Malformed frames are discarded synchronously inside handleMessage (no enqueue,
	// no goroutine hop), so we can assert immediately — nothing can appear later.
	h.handleMessage(ctx, conn, []byte("not-json{{{"))

	assert.Empty(t, ss.msgs(), "malformed frame must not trigger any write")
}

func TestHandleMessage_UnknownType_Discards(t *testing.T) {
	ctx := context.Background()
	h, hub, _ := newTestWSHandler(t)
	ss, conn := registerTestConn(ctx, hub, h, uuid.New())
	defer ss.disconnect()

	msg, _ := json.Marshal(map[string]any{
		"type":     "UNKNOWN_OP",
		"board_id": uuid.New().String(),
	})
	// Unknown type falls through the switch with no enqueue — synchronous discard.
	h.handleMessage(ctx, conn, msg)

	assert.Empty(t, ss.msgs(), "unknown message type must not trigger any write")
}

// ---------------------------------------------------------------------------
// handleJoin
// ---------------------------------------------------------------------------

func TestHandleJoin_AccessDenied_Silent(t *testing.T) {
	ctx := context.Background()
	h, hub, access := newTestWSHandler(t)
	userID := uuid.New()
	boardID := uuid.New()

	ss, conn := registerTestConn(ctx, hub, h, userID)
	defer ss.disconnect()

	access.EXPECT().CheckViewAccess(mock.Anything, boardID, userID).
		Return(nil, errors.New("denied")).Once()

	// Denial returns before any Join/enqueue, all in this goroutine — assert now.
	h.handleJoin(ctx, conn, boardID)

	assert.Empty(t, ss.msgs(), "no frame expected on denial")
	assert.Empty(t, hub.ActiveUsers(boardID), "conn must not be in the room after denial")
}

func TestHandleJoin_AccessGranted_SendsActiveUsers(t *testing.T) {
	ctx := context.Background()
	h, hub, access := newTestWSHandler(t)
	userID := uuid.New()
	boardID := uuid.New()

	ss, conn := registerTestConn(ctx, hub, h, userID)
	defer ss.disconnect()

	access.EXPECT().CheckViewAccess(mock.Anything, boardID, userID).
		Return(&common.BoardAccess{}, nil).Once()

	h.handleJoin(ctx, conn, boardID)

	// The joining conn receives two frames: USER_JOINED (from the 0→1 presence
	// edge inside hub.Join) and then ACTIVE_USERS (from handleJoin itself).
	require.Eventually(t, func() bool { return len(ss.msgs()) >= 2 }, 2*time.Second, 10*time.Millisecond)

	raw := findFrameByType(ss.msgs(), realtime.FrameTypeActiveUsers)
	require.NotNil(t, raw, "ACTIVE_USERS frame not found among sent messages")

	var frame realtime.ActiveUsersFrame
	require.NoError(t, json.Unmarshal(raw, &frame))
	assert.Equal(t, boardID, frame.BoardID)
	assert.Contains(t, frame.UserIDs, userID)
}

// TestHandleJoin_Timeout_Silent verifies that a cancelled context results in
// silent denial — no ACTIVE_USERS frame, user not added to the room.
func TestHandleJoin_Timeout_Silent(t *testing.T) {
	ctx := context.Background()
	h, hub, access := newTestWSHandler(t)
	userID := uuid.New()
	boardID := uuid.New()

	ss, conn := registerTestConn(ctx, hub, h, userID)
	defer ss.disconnect()

	access.EXPECT().CheckViewAccess(mock.Anything, boardID, userID).
		RunAndReturn(func(ctx context.Context, _, _ uuid.UUID) (*common.BoardAccess, error) {
			<-ctx.Done() // block until the bounded context expires
			return nil, ctx.Err()
		}).Once()

	// A pre-cancelled context: context.WithTimeout(cancelledCtx, 5s) is also
	// immediately done, so CheckViewAccess returns without blocking and handleJoin
	// returns before any enqueue — assert immediately.
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	h.handleJoin(cancelled, conn, boardID)

	assert.Empty(t, ss.msgs(), "no ACTIVE_USERS expected when the context times out")
}

// ---------------------------------------------------------------------------
// handleLeave
// ---------------------------------------------------------------------------

func TestHandleLeave_RemovesFromRoom(t *testing.T) {
	ctx := context.Background()
	h, hub, access := newTestWSHandler(t)
	userID := uuid.New()
	boardID := uuid.New()

	ss, conn := registerTestConn(ctx, hub, h, userID)
	defer ss.disconnect()

	access.EXPECT().CheckViewAccess(mock.Anything, boardID, userID).
		Return(&common.BoardAccess{}, nil).Once()
	h.handleJoin(ctx, conn, boardID)
	// After join: USER_JOINED (presence edge) + ACTIVE_USERS from handleJoin.
	require.Eventually(t, func() bool { return len(ss.msgs()) >= 2 }, 2*time.Second, 10*time.Millisecond)

	h.handleLeave(conn, boardID)

	assert.Empty(t, hub.ActiveUsers(boardID), "conn must be removed from the room after leave")
}

// ---------------------------------------------------------------------------
// onPresence — wire-frame broadcasting
// ---------------------------------------------------------------------------

func TestOnPresence_Joined_BroadcastsUserJoined(t *testing.T) {
	ctx := context.Background()
	h, hub, _ := newTestWSHandler(t)
	boardID := uuid.New()
	userID := uuid.New()

	// hub.Join inside listenerConn fires the listener's own USER_JOINED edge.
	// Wait for that initial message so we start from a known baseline.
	lss, cleanup := listenerConn(ctx, hub, boardID)
	defer cleanup()
	require.Eventually(t, func() bool { return len(lss.msgs()) >= 1 }, 2*time.Second, 10*time.Millisecond)

	h.onPresence(boardID, userID, realtime.PresenceJoined)

	require.Eventually(t, func() bool { return len(lss.msgs()) >= 2 }, 2*time.Second, 10*time.Millisecond)

	f := findPresenceFrameByUser(lss.msgs(), userID)
	require.NotNil(t, f, "USER_JOINED frame for userID not found")
	assert.Equal(t, realtime.FrameTypeUserJoined, f.Type)
	assert.Equal(t, boardID, f.BoardID)
}

func TestOnPresence_Left_BroadcastsUserLeft(t *testing.T) {
	ctx := context.Background()
	h, hub, _ := newTestWSHandler(t)
	boardID := uuid.New()
	userID := uuid.New()

	lss, cleanup := listenerConn(ctx, hub, boardID)
	defer cleanup()
	require.Eventually(t, func() bool { return len(lss.msgs()) >= 1 }, 2*time.Second, 10*time.Millisecond)

	h.onPresence(boardID, userID, realtime.PresenceLeft)

	require.Eventually(t, func() bool { return len(lss.msgs()) >= 2 }, 2*time.Second, 10*time.Millisecond)

	f := findPresenceFrameByUser(lss.msgs(), userID)
	require.NotNil(t, f, "USER_LEFT frame for userID not found")
	assert.Equal(t, realtime.FrameTypeUserLeft, f.Type)
	assert.Equal(t, boardID, f.BoardID)
}

// ---------------------------------------------------------------------------
// disconnect / readPump-EOF (Code review #2, R1)
//
// The one end-to-end path nothing else exercises: a socket dying (NOT a
// LEAVE_BOARD frame) drives readPump EOF → onClose → unregisterConn →
// teardown → close(done), and the leaver's last conn trips the 1→0 presence
// edge, emitting USER_LEFT to the surviving room.
// ---------------------------------------------------------------------------

func TestServeWS_DisconnectEOF_RemovesConnAndBroadcastsUserLeft(t *testing.T) {
	ctx := context.Background()
	h, hub, access := newTestWSHandler(t)
	boardID := uuid.New()
	leaver, watcher := uuid.New(), uuid.New()

	access.EXPECT().CheckViewAccess(mock.Anything, boardID, leaver).
		Return(&common.BoardAccess{}, nil).Once()
	access.EXPECT().CheckViewAccess(mock.Anything, boardID, watcher).
		Return(&common.BoardAccess{}, nil).Once()

	// Leaver joins first: its 0→1 edge broadcasts USER_JOINED{leaver} only to the
	// room-of-one (itself), so the watcher never sees a leaver-keyed JOIN frame —
	// the only leaver presence frame the watcher can observe is the USER_LEFT below.
	leaverSock, leaverConn := registerTestConn(ctx, hub, h, leaver)
	watcherSock, watcherConn := registerTestConn(ctx, hub, h, watcher)
	defer watcherSock.disconnect()
	h.handleJoin(ctx, leaverConn, boardID)
	h.handleJoin(ctx, watcherConn, boardID)

	require.Eventually(t, func() bool {
		return len(hub.ActiveUsers(boardID)) == 2
	}, 2*time.Second, 10*time.Millisecond, "both users must be in the room before disconnect")

	leaverSock.disconnect() // readPump Read returns io.EOF — no LEAVE_BOARD frame

	// 1. Leaver removed from the room; only the watcher remains.
	require.Eventually(t, func() bool {
		return len(hub.ActiveUsers(boardID)) == 1
	}, 2*time.Second, 10*time.Millisecond, "leaver must be removed from the room on disconnect")
	assert.ElementsMatch(t, []uuid.UUID{watcher}, hub.ActiveUsers(boardID))

	// 2. USER_LEFT{leaver} broadcast to the room — the watcher observes it.
	require.Eventually(t, func() bool {
		return findPresenceFrameByUser(watcherSock.msgs(), leaver) != nil
	}, 2*time.Second, 10*time.Millisecond, "watcher must receive a leaver presence frame")
	f := findPresenceFrameByUser(watcherSock.msgs(), leaver)
	require.NotNil(t, f)
	assert.Equal(t, realtime.FrameTypeUserLeft, f.Type, "leaver's presence frame must be USER_LEFT")
	assert.Equal(t, boardID, f.BoardID)

	// 3. Leaver's Done() is closed — the contract that unblocks ServeWS.
	select {
	case <-leaverConn.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("leaver conn.Done() not closed after disconnect")
	}
}

func TestOnPresence_UnknownKind_NoBroadcast(t *testing.T) {
	ctx := context.Background()
	h, hub, _ := newTestWSHandler(t)
	boardID := uuid.New()
	userID := uuid.New()

	lss, cleanup := listenerConn(ctx, hub, boardID)
	defer cleanup()
	require.Eventually(t, func() bool { return len(lss.msgs()) >= 1 }, 2*time.Second, 10*time.Millisecond)

	initialCount := len(lss.msgs())
	h.onPresence(boardID, userID, realtime.PresenceKind(99))

	time.Sleep(40 * time.Millisecond)
	assert.Len(t, lss.msgs(), initialCount, "unknown PresenceKind must not broadcast any frame")
}
