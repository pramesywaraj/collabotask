package realtime_test

import (
	"context"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"collabotask/internal/realtime"
)

// ---------------------------------------------------------------------------
// fakeSocket — injectable transport for tests
// ---------------------------------------------------------------------------

type fakeSocket struct {
	mu       sync.Mutex
	written  [][]byte
	closed   []string // reasons passed to Close
	readErr  chan error
	closedCh chan struct{}

	// writeGate, if non-nil, blocks Write until the channel is closed.
	// Used to freeze the writePump without blocking Close (avoids self-deadlock
	// when Broadcast → teardown → Close is called on the same goroutine).
	writeGate chan struct{}
}

func newFakeSocket() *fakeSocket {
	return &fakeSocket{
		readErr:  make(chan error, 1),
		closedCh: make(chan struct{}),
	}
}

// disconnect makes the next Read return io.EOF (simulates remote close).
func (f *fakeSocket) disconnect() {
	select {
	case f.readErr <- io.EOF:
	default:
	}
}

func (f *fakeSocket) Write(_ context.Context, data []byte) error {
	if f.writeGate != nil {
		<-f.writeGate
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := make([]byte, len(data))
	copy(cp, data)
	f.written = append(f.written, cp)
	return nil
}

func (f *fakeSocket) Read(_ context.Context) ([]byte, error) {
	err := <-f.readErr
	return nil, err
}

func (f *fakeSocket) Ping(_ context.Context) error {
	return nil
}

func (f *fakeSocket) Close(reason string) error {
	f.mu.Lock()
	f.closed = append(f.closed, reason)
	f.mu.Unlock()
	select {
	case <-f.closedCh:
	default:
		close(f.closedCh)
	}
	return nil
}

func (f *fakeSocket) writtenMsgs() [][]byte {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([][]byte, len(f.written))
	copy(out, f.written)
	return out
}

func (f *fakeSocket) closedReasons() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]string, len(f.closed))
	copy(out, f.closed)
	return out
}

// waitClosed blocks until Close is called (or the test times out).
func (f *fakeSocket) waitClosed(t *testing.T) {
	t.Helper()
	select {
	case <-f.closedCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for socket Close")
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

type presenceEvent struct {
	boardID uuid.UUID
	userID  uuid.UUID
	kind    realtime.PresenceKind
}

type presenceRecorder struct {
	mu     sync.Mutex
	events []presenceEvent
}

func (r *presenceRecorder) record(boardID, userID uuid.UUID, kind realtime.PresenceKind) {
	r.mu.Lock()
	r.events = append(r.events, presenceEvent{boardID, userID, kind})
	r.mu.Unlock()
}

func (r *presenceRecorder) snapshot() []presenceEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]presenceEvent, len(r.events))
	copy(out, r.events)
	return out
}

func (r *presenceRecorder) reset() {
	r.mu.Lock()
	r.events = nil
	r.mu.Unlock()
}

func newHub(t *testing.T) (*realtime.Hub, *presenceRecorder) {
	t.Helper()
	rec := &presenceRecorder{}
	h := realtime.NewHub()
	h.SetPresence(rec.record)
	return h, rec
}

// registerConn registers a conn via a plain (non-blocking) fakeSocket.
// Always pair with defer fs.disconnect() to unblock the readPump on test exit.
func registerConn(ctx context.Context, h *realtime.Hub, userID uuid.UUID) (*fakeSocket, *realtime.Conn) {
	fs := newFakeSocket()
	c := h.Register(ctx, userID, fs, nil)
	return fs, c
}

// ---------------------------------------------------------------------------
// Slice 1 — Join / Leave / ActiveUsers
// ---------------------------------------------------------------------------

func TestJoin_ReturnsActiveUsersIncludingSelf(t *testing.T) {
	ctx := context.Background()
	h, _ := newHub(t)
	boardID := uuid.New()
	userID := uuid.New()

	fs, c := registerConn(ctx, h, userID)
	defer fs.disconnect()

	snapshot := h.Join(boardID, c)

	require.Len(t, snapshot, 1)
	assert.Equal(t, userID, snapshot[0])
}

func TestJoin_MultipleUsers_ActiveUsersContainsAll(t *testing.T) {
	ctx := context.Background()
	h, _ := newHub(t)
	boardID := uuid.New()
	u1, u2 := uuid.New(), uuid.New()

	fs1, c1 := registerConn(ctx, h, u1)
	fs2, c2 := registerConn(ctx, h, u2)
	defer fs1.disconnect()
	defer fs2.disconnect()

	h.Join(boardID, c1)
	h.Join(boardID, c2)

	assert.ElementsMatch(t, []uuid.UUID{u1, u2}, h.ActiveUsers(boardID))
}

func TestLeave_RemovesUserFromRoom(t *testing.T) {
	ctx := context.Background()
	h, _ := newHub(t)
	boardID := uuid.New()
	userID := uuid.New()

	fs, c := registerConn(ctx, h, userID)
	defer fs.disconnect()

	h.Join(boardID, c)
	h.Leave(boardID, c)

	assert.Empty(t, h.ActiveUsers(boardID))
}

func TestLeave_OtherRoomsUnaffected(t *testing.T) {
	ctx := context.Background()
	h, _ := newHub(t)
	board1, board2 := uuid.New(), uuid.New()
	userID := uuid.New()

	fs, c := registerConn(ctx, h, userID)
	defer fs.disconnect()

	h.Join(board1, c)
	h.Join(board2, c)
	h.Leave(board1, c)

	assert.Empty(t, h.ActiveUsers(board1))
	assert.ElementsMatch(t, []uuid.UUID{userID}, h.ActiveUsers(board2))
}

// ---------------------------------------------------------------------------
// Slice 2 — Edge-triggered presence
// ---------------------------------------------------------------------------

func TestPresence_JoinFires_OnFirstConnOnly(t *testing.T) {
	ctx := context.Background()
	h, rec := newHub(t)
	boardID := uuid.New()
	userID := uuid.New()

	fs1, c1 := registerConn(ctx, h, userID)
	fs2, c2 := registerConn(ctx, h, userID)
	defer fs1.disconnect()
	defer fs2.disconnect()

	h.Join(boardID, c1) // 0→1: must fire Joined
	h.Join(boardID, c2) // 1→2: multi-tab, must NOT fire

	events := rec.snapshot()
	require.Len(t, events, 1)
	assert.Equal(t, boardID, events[0].boardID)
	assert.Equal(t, userID, events[0].userID)
	assert.Equal(t, realtime.PresenceJoined, events[0].kind)
}

func TestPresence_LeaveFires_OnLastConnOnly(t *testing.T) {
	ctx := context.Background()
	h, rec := newHub(t)
	boardID := uuid.New()
	userID := uuid.New()

	fs1, c1 := registerConn(ctx, h, userID)
	fs2, c2 := registerConn(ctx, h, userID)
	defer fs1.disconnect()
	defer fs2.disconnect()

	h.Join(boardID, c1)
	h.Join(boardID, c2)
	rec.reset()

	h.Leave(boardID, c1) // 2→1: must NOT fire
	h.Leave(boardID, c2) // 1→0: must fire Left

	events := rec.snapshot()
	require.Len(t, events, 1)
	assert.Equal(t, realtime.PresenceLeft, events[0].kind)
	assert.Equal(t, userID, events[0].userID)
}

func TestActiveUsers_MultiTabUserCountedOnce(t *testing.T) {
	ctx := context.Background()
	h, _ := newHub(t)
	boardID := uuid.New()
	userID := uuid.New()

	fs1, c1 := registerConn(ctx, h, userID)
	fs2, c2 := registerConn(ctx, h, userID)
	defer fs1.disconnect()
	defer fs2.disconnect()

	h.Join(boardID, c1)
	h.Join(boardID, c2)

	active := h.ActiveUsers(boardID)
	assert.Len(t, active, 1, "multi-tab user must appear exactly once")
	assert.Equal(t, userID, active[0])
}

// ---------------------------------------------------------------------------
// Slice 3 — Broadcast fan-out + slow-consumer drop
// ---------------------------------------------------------------------------

func TestBroadcast_DeliversToConn(t *testing.T) {
	ctx := context.Background()
	h, _ := newHub(t)
	boardID := uuid.New()
	userID := uuid.New()

	fs, c := registerConn(ctx, h, userID)
	defer fs.disconnect()

	h.Join(boardID, c)
	h.Broadcast(boardID, []byte(`{"type":"TEST"}`))

	require.Eventually(t, func() bool {
		return len(fs.writtenMsgs()) == 1
	}, 2*time.Second, 10*time.Millisecond)

	assert.Equal(t, []byte(`{"type":"TEST"}`), fs.writtenMsgs()[0])
}

func TestBroadcast_DoesNotReachOtherRooms(t *testing.T) {
	ctx := context.Background()
	h, _ := newHub(t)
	board1, board2 := uuid.New(), uuid.New()
	u1, u2 := uuid.New(), uuid.New()

	fs1, c1 := registerConn(ctx, h, u1)
	fs2, c2 := registerConn(ctx, h, u2)
	defer fs1.disconnect()
	defer fs2.disconnect()

	h.Join(board1, c1)
	h.Join(board2, c2)

	h.Broadcast(board1, []byte(`{"type":"PING"}`))

	time.Sleep(50 * time.Millisecond)
	assert.Empty(t, fs2.writtenMsgs(), "conn in board2 must not receive board1 broadcast")
}

func TestBroadcast_IncludesSender(t *testing.T) {
	ctx := context.Background()
	h, _ := newHub(t)
	boardID := uuid.New()
	userID := uuid.New()

	fs, c := registerConn(ctx, h, userID)
	defer fs.disconnect()

	h.Join(boardID, c)
	h.Broadcast(boardID, []byte(`{"type":"CARD_MOVED"}`))

	require.Eventually(t, func() bool {
		return len(fs.writtenMsgs()) == 1
	}, 2*time.Second, 10*time.Millisecond)
}

func TestBroadcast_SlowConsumer_IsDropped(t *testing.T) {
	ctx := context.Background()
	h, _ := newHub(t)
	boardID := uuid.New()
	userID := uuid.New()

	// writeGate blocks Write without blocking Close, avoiding the self-deadlock
	// that occurs when Broadcast → teardown → Close is called on the same
	// goroutine as the one holding a mutex that Close also needs.
	writeGate := make(chan struct{})
	fs := newFakeSocket()
	fs.writeGate = writeGate
	// Cleanup: unblock the writePump and readPump so goroutines exit cleanly.
	t.Cleanup(func() {
		close(writeGate)
		fs.disconnect()
	})

	slowConn := h.Register(ctx, userID, fs, nil)
	h.Join(boardID, slowConn)

	// Flood with more messages than the buffer can hold (sendBufferSize=16).
	// The writePump is frozen on Write (blocked on writeGate), so the send
	// buffer fills up and the overflow triggers the slow-consumer drop.
	for i := 0; i < 20; i++ {
		h.Broadcast(boardID, []byte(`x`))
	}

	fs.waitClosed(t)
	assert.Empty(t, h.ActiveUsers(boardID))
}

// ---------------------------------------------------------------------------
// Slice 4 — Teardown on disconnect
// ---------------------------------------------------------------------------

func TestTeardown_DisconnectRemovesFromAllRooms(t *testing.T) {
	ctx := context.Background()
	h, _ := newHub(t)
	board1, board2 := uuid.New(), uuid.New()
	userID := uuid.New()

	fs, c := registerConn(ctx, h, userID)

	h.Join(board1, c)
	h.Join(board2, c)

	fs.disconnect()
	fs.waitClosed(t)

	require.Eventually(t, func() bool {
		return len(h.ActiveUsers(board1)) == 0 && len(h.ActiveUsers(board2)) == 0
	}, 2*time.Second, 10*time.Millisecond)
}

func TestTeardown_PresenceEdge_OnDisconnect(t *testing.T) {
	ctx := context.Background()
	h, rec := newHub(t)
	boardID := uuid.New()
	userID := uuid.New()

	fs, c := registerConn(ctx, h, userID)
	h.Join(boardID, c)
	rec.reset()

	fs.disconnect()
	fs.waitClosed(t)

	require.Eventually(t, func() bool {
		return len(rec.snapshot()) == 1
	}, 2*time.Second, 10*time.Millisecond)

	ev := rec.snapshot()[0]
	assert.Equal(t, realtime.PresenceLeft, ev.kind)
	assert.Equal(t, userID, ev.userID)
}

func TestTeardown_DoubleTeardown_IsNoop(t *testing.T) {
	ctx := context.Background()
	h, _ := newHub(t)
	boardID := uuid.New()
	userID := uuid.New()

	fs, c := registerConn(ctx, h, userID)
	h.Join(boardID, c)

	// Trigger two teardown paths concurrently: disconnect (readPump) + EvictUser.
	fs.disconnect()
	h.EvictUser(boardID, userID, "test")
	fs.waitClosed(t)

	// sync.Once must guarantee Close is called exactly once.
	time.Sleep(50 * time.Millisecond)
	assert.Len(t, fs.closedReasons(), 1)
}

// ---------------------------------------------------------------------------
// Slice 5 — Eviction primitives
// ---------------------------------------------------------------------------

func TestEvictUser_TearsDownUserConns_InBoard(t *testing.T) {
	ctx := context.Background()
	h, _ := newHub(t)
	boardID := uuid.New()
	victim, bystander := uuid.New(), uuid.New()

	fsV, cV := registerConn(ctx, h, victim)
	fsB, cB := registerConn(ctx, h, bystander)
	defer fsV.disconnect() // unblock readPump after eviction teardown
	defer fsB.disconnect()

	h.Join(boardID, cV)
	h.Join(boardID, cB)

	h.EvictUser(boardID, victim, "access revoked")
	fsV.waitClosed(t)

	assert.ElementsMatch(t, []uuid.UUID{bystander}, h.ActiveUsers(boardID))
	assert.Equal(t, []string{"access revoked"}, fsV.closedReasons())
	assert.Empty(t, fsB.closedReasons(), "bystander must not be closed")
}

func TestEvictUser_OtherBoardsUnaffected(t *testing.T) {
	ctx := context.Background()
	h, _ := newHub(t)
	board1, board2 := uuid.New(), uuid.New()
	userID := uuid.New()

	fs1, c1 := registerConn(ctx, h, userID)
	fs2, c2 := registerConn(ctx, h, userID)
	defer fs1.disconnect()
	defer fs2.disconnect()

	h.Join(board1, c1)
	h.Join(board2, c2)

	h.EvictUser(board1, userID, "removed")
	fs1.waitClosed(t)

	assert.ElementsMatch(t, []uuid.UUID{userID}, h.ActiveUsers(board2))
	assert.Empty(t, fs2.closedReasons(), "board2 conn must not be closed")
}

func TestEvictExcept_EvictsOnlyNonAllowed(t *testing.T) {
	ctx := context.Background()
	h, _ := newHub(t)
	boardID := uuid.New()
	allowed, evicted := uuid.New(), uuid.New()

	fsA, cA := registerConn(ctx, h, allowed)
	fsE, cE := registerConn(ctx, h, evicted)
	defer fsA.disconnect()
	defer fsE.disconnect()

	h.Join(boardID, cA)
	h.Join(boardID, cE)

	h.EvictExcept(boardID, []uuid.UUID{allowed}, "board went private")
	fsE.waitClosed(t)

	assert.ElementsMatch(t, []uuid.UUID{allowed}, h.ActiveUsers(boardID))
	assert.Empty(t, fsA.closedReasons(), "allowed conn must not be closed")
}

func TestEvictUserFromRooms_EvictsAcrossBoards(t *testing.T) {
	ctx := context.Background()
	h, _ := newHub(t)
	board1, board2, board3 := uuid.New(), uuid.New(), uuid.New()
	victim, bystander := uuid.New(), uuid.New()

	fsV1, cV1 := registerConn(ctx, h, victim)
	fsV2, cV2 := registerConn(ctx, h, victim)
	fsB, cB := registerConn(ctx, h, bystander)
	defer fsV1.disconnect()
	defer fsV2.disconnect()
	defer fsB.disconnect()

	h.Join(board1, cV1)
	h.Join(board2, cV2)
	h.Join(board3, cB)

	h.EvictUserFromRooms(victim, []uuid.UUID{board1, board2}, "workspace removed")
	fsV1.waitClosed(t)
	fsV2.waitClosed(t)

	assert.Empty(t, h.ActiveUsers(board1))
	assert.Empty(t, h.ActiveUsers(board2))
	assert.ElementsMatch(t, []uuid.UUID{bystander}, h.ActiveUsers(board3))
}

// ---------------------------------------------------------------------------
// Slice 6 — BroadcastToUser (targeted delivery before eviction)
// ---------------------------------------------------------------------------

func TestBroadcastToUser_DeliversOnlyToTargetUser(t *testing.T) {
	ctx := context.Background()
	h, _ := newHub(t)
	boardID := uuid.New()
	target, bystander := uuid.New(), uuid.New()

	fsT, cT := registerConn(ctx, h, target)
	fsB, cB := registerConn(ctx, h, bystander)
	defer fsT.disconnect()
	defer fsB.disconnect()

	h.Join(boardID, cT)
	h.Join(boardID, cB)

	msg := []byte(`{"type":"ACCESS_REVOKED"}`)
	h.BroadcastToUser(boardID, target, msg)

	require.Eventually(t, func() bool {
		return len(fsT.writtenMsgs()) == 1
	}, 2*time.Second, 10*time.Millisecond)

	assert.Equal(t, msg, fsT.writtenMsgs()[0])
	time.Sleep(20 * time.Millisecond)
	assert.Empty(t, fsB.writtenMsgs(), "bystander must not receive targeted message")
}

func TestBroadcastToUser_MultiTab_AllConnsReceive(t *testing.T) {
	ctx := context.Background()
	h, _ := newHub(t)
	boardID := uuid.New()
	userID := uuid.New()

	fs1, c1 := registerConn(ctx, h, userID)
	fs2, c2 := registerConn(ctx, h, userID)
	defer fs1.disconnect()
	defer fs2.disconnect()

	h.Join(boardID, c1)
	h.Join(boardID, c2)

	msg := []byte(`{"type":"ACCESS_REVOKED"}`)
	h.BroadcastToUser(boardID, userID, msg)

	require.Eventually(t, func() bool {
		return len(fs1.writtenMsgs()) == 1 && len(fs2.writtenMsgs()) == 1
	}, 2*time.Second, 10*time.Millisecond)
}

func TestBroadcastToUser_UnknownBoard_IsNoop(t *testing.T) {
	ctx := context.Background()
	h, _ := newHub(t)
	userID := uuid.New()

	fs, _ := registerConn(ctx, h, userID)
	defer fs.disconnect()

	// no Join → no room → should not panic
	h.BroadcastToUser(uuid.New(), userID, []byte(`{}`))
	time.Sleep(20 * time.Millisecond)
	assert.Empty(t, fs.writtenMsgs())
}

// ---------------------------------------------------------------------------
// Race sanity
// ---------------------------------------------------------------------------

func TestRace_ConcurrentJoinBroadcastEvict(t *testing.T) {
	ctx := context.Background()
	h, _ := newHub(t)
	boardID := uuid.New()

	const n = 20
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			userID := uuid.New()
			fs, c := registerConn(ctx, h, userID)
			h.Join(boardID, c)
			h.Broadcast(boardID, []byte(`{"type":"RACE"}`))
			h.Leave(boardID, c)
			fs.disconnect()
		}()
	}
	wg.Wait()
}
