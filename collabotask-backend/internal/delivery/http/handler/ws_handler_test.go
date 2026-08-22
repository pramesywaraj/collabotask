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

	"collabotask/internal/domain/entity"
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
func (s *stubSocket) Close(_ string) error         { return nil }

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

// newTestWSHandler creates a handler with a real hub, mock access checker, and
// mock user repository. No default GetByIds stub is registered — each test sets
// up exactly the expectations it needs. testify/mock panics on any unexpected
// call, which is the correct behaviour for catching unintended DB access.
func newTestWSHandler(t *testing.T) (*WSHandler, *realtime.Hub, *mocks.MockBoardAccessChecker, *mocks.MockUserRepository) {
	t.Helper()
	hub := realtime.NewHub()
	access := mocks.NewMockBoardAccessChecker(t)
	userRepo := mocks.NewMockUserRepository(t)
	h := NewWSHandler(hub, access, userRepo, nil)
	return h, hub, access, userRepo
}

// allowAnyGetByIds registers a broad, once-only stub that absorbs one GetByIds
// call regardless of args and returns an empty map. Useful for "don't care"
// calls such as the presence enrichment triggered by listenerConn's own join.
// Must be called BEFORE the hub.Join that triggers the enrichment.
func allowAnyGetByIds(userRepo *mocks.MockUserRepository) {
	userRepo.On("GetByIds", mock.Anything, mock.Anything).
		Return(map[uuid.UUID]*entity.User{}, nil).Once()
}

// registerTestConn registers a conn whose readPump calls h.handleMessage for
// every inbound frame.
func registerTestConn(ctx context.Context, hub *realtime.Hub, h *WSHandler, userID uuid.UUID) (*stubSocket, *realtime.Conn) {
	ss := newStubSocket()
	conn := hub.Register(ctx, userID, ss, h.handleMessage)
	return ss, conn
}

// listenerConn registers a bystander in the board so we can observe Broadcast
// output. NOTE: hub.Join fires the edge-triggered presence callback synchronously
// (onPresence → GetByIds), so callers must set up a GetByIds stub BEFORE calling
// this function.
func listenerConn(ctx context.Context, hub *realtime.Hub, boardID uuid.UUID) (*stubSocket, func()) {
	ss := newStubSocket()
	conn := hub.Register(ctx, uuid.New(), ss, nil)
	hub.Join(boardID, conn)
	return ss, ss.disconnect
}

// findFrameByType unmarshals each raw msg looking for one whose Type field equals wantType.
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

// findUserLeftFrameByUserID scans msgs for a UserLeftFrame whose UserID equals want.
func findUserLeftFrameByUserID(msgs [][]byte, want uuid.UUID) *realtime.UserLeftFrame {
	for _, raw := range msgs {
		var f realtime.UserLeftFrame
		if json.Unmarshal(raw, &f) == nil && f.UserID == want {
			return &f
		}
	}
	return nil
}

// findUserJoinedFrameByUserID scans msgs for a UserJoinedFrame whose User.ID equals want.
func findUserJoinedFrameByUserID(msgs [][]byte, want uuid.UUID) *realtime.UserJoinedFrame {
	for _, raw := range msgs {
		var f realtime.UserJoinedFrame
		if json.Unmarshal(raw, &f) == nil && f.Type == realtime.FrameTypeUserJoined && f.User.ID == want {
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
	h, hub, _, _ := newTestWSHandler(t)
	ss, conn := registerTestConn(ctx, hub, h, uuid.New())
	defer ss.disconnect()

	// Malformed frames are discarded synchronously inside handleMessage (no enqueue,
	// no goroutine hop), so we can assert immediately — nothing can appear later.
	h.handleMessage(ctx, conn, []byte("not-json{{{"))

	assert.Empty(t, ss.msgs(), "malformed frame must not trigger any write")
}

func TestHandleMessage_UnknownType_Discards(t *testing.T) {
	ctx := context.Background()
	h, hub, _, _ := newTestWSHandler(t)
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
	h, hub, access, _ := newTestWSHandler(t)
	userID := uuid.New()
	boardID := uuid.New()

	ss, conn := registerTestConn(ctx, hub, h, userID)
	defer ss.disconnect()

	access.EXPECT().CheckViewAccess(mock.Anything, boardID, userID).
		Return(nil, errors.New("denied")).Once()

	// Denial returns before any Join/enqueue — assert immediately, no GetByIds.
	h.handleJoin(ctx, conn, boardID)

	assert.Empty(t, ss.msgs(), "no frame expected on denial")
	assert.Empty(t, hub.ActiveUsers(boardID), "conn must not be in the room after denial")
}

// TestHandleJoin_AccessGranted_SendsRichActiveUsers verifies that on a successful
// JOIN_BOARD the joining conn receives an ACTIVE_USERS frame with enriched profiles.
// hub.Join fires onPresence(PresenceJoined) synchronously → GetByIds call #1 (USER_JOINED);
// enrichUsers runs after → GetByIds call #2 (ACTIVE_USERS). Both return Alice's profile.
func TestHandleJoin_AccessGranted_SendsRichActiveUsers(t *testing.T) {
	ctx := context.Background()
	h, hub, access, userRepo := newTestWSHandler(t)
	userID := uuid.New()
	boardID := uuid.New()
	avatarURL := "https://example.com/avatar.png"

	ss, conn := registerTestConn(ctx, hub, h, userID)
	defer ss.disconnect()

	access.EXPECT().CheckViewAccess(mock.Anything, boardID, userID).
		Return(&common.BoardAccess{}, nil).Once()

	// hub.Join fires onPresence → enrichSingleUser → GetByIds (call #1, USER_JOINED).
	// enrichUsers after hub.Join → GetByIds (call #2, ACTIVE_USERS).
	// Both calls use []uuid.UUID{userID}; .Times(2) returns Alice's profile for both.
	userRepo.On("GetByIds", mock.Anything, []uuid.UUID{userID}).
		Return(map[uuid.UUID]*entity.User{
			userID: {ID: userID, Name: "Alice", AvatarURL: &avatarURL},
		}, nil).Times(2)

	h.handleJoin(ctx, conn, boardID)

	// The joining conn receives two frames: USER_JOINED (from the 0→1 presence
	// edge inside hub.Join) and then ACTIVE_USERS (from handleJoin itself).
	require.Eventually(t, func() bool { return len(ss.msgs()) >= 2 }, 2*time.Second, 10*time.Millisecond)

	raw := findFrameByType(ss.msgs(), realtime.FrameTypeActiveUsers)
	require.NotNil(t, raw, "ACTIVE_USERS frame not found among sent messages")

	var frame realtime.ActiveUsersFrame
	require.NoError(t, json.Unmarshal(raw, &frame))
	assert.Equal(t, boardID, frame.BoardID)
	require.Len(t, frame.Users, 1)
	assert.Equal(t, userID, frame.Users[0].ID)
	assert.Equal(t, "Alice", frame.Users[0].Name)
	require.NotNil(t, frame.Users[0].AvatarURL)
	assert.Equal(t, avatarURL, *frame.Users[0].AvatarURL)
}

// TestHandleJoin_MultiUserRoom_RichActiveUsers verifies that ACTIVE_USERS carries
// enriched profiles for all connected users in the room (incl. self).
func TestHandleJoin_MultiUserRoom_RichActiveUsers(t *testing.T) {
	ctx := context.Background()
	h, hub, access, userRepo := newTestWSHandler(t)
	boardID := uuid.New()
	aliceID := uuid.New()
	bobID := uuid.New()
	avatarURL := "https://example.com/bob.png"

	// hub.Join(boardID, bobConn) fires onPresence(PresenceJoined, bobID) synchronously.
	// Set up enrichment stubs BEFORE any hub.Join call.
	// Call #1: Bob's USER_JOINED enrichment (don't care about content).
	userRepo.On("GetByIds", mock.Anything, []uuid.UUID{bobID}).
		Return(map[uuid.UUID]*entity.User{}, nil).Once()
	// Call #2: Alice's USER_JOINED enrichment (inside handleJoin → hub.Join → onPresence).
	userRepo.On("GetByIds", mock.Anything, []uuid.UUID{aliceID}).
		Return(map[uuid.UUID]*entity.User{}, nil).Once()
	// Call #3: Alice's ACTIVE_USERS enrichment (2-user snapshot).
	userRepo.On("GetByIds", mock.Anything, mock.MatchedBy(func(ids []uuid.UUID) bool {
		return len(ids) == 2
	})).Return(map[uuid.UUID]*entity.User{
		aliceID: {ID: aliceID, Name: "Alice", AvatarURL: nil},
		bobID:   {ID: bobID, Name: "Bob", AvatarURL: &avatarURL},
	}, nil).Once()

	// Bob joins first (direct hub.Join — fires GetByIds([bobID]) synchronously).
	bobSock, bobConn := registerTestConn(ctx, hub, h, bobID)
	defer bobSock.disconnect()
	hub.Join(boardID, bobConn)

	// Alice joins via handleJoin (fires GetByIds([aliceID]) then GetByIds([aliceID, bobID])).
	aliceSock, aliceConn := registerTestConn(ctx, hub, h, aliceID)
	defer aliceSock.disconnect()

	access.EXPECT().CheckViewAccess(mock.Anything, boardID, aliceID).
		Return(&common.BoardAccess{}, nil).Once()

	h.handleJoin(ctx, aliceConn, boardID)

	require.Eventually(t, func() bool { return len(aliceSock.msgs()) >= 2 }, 2*time.Second, 10*time.Millisecond)

	raw := findFrameByType(aliceSock.msgs(), realtime.FrameTypeActiveUsers)
	require.NotNil(t, raw, "ACTIVE_USERS frame not found")

	var frame realtime.ActiveUsersFrame
	require.NoError(t, json.Unmarshal(raw, &frame))
	assert.Equal(t, boardID, frame.BoardID)
	assert.Len(t, frame.Users, 2, "ACTIVE_USERS must include all connected users")

	byID := make(map[uuid.UUID]realtime.PresenceUser)
	for _, u := range frame.Users {
		byID[u.ID] = u
	}
	assert.Equal(t, "Alice", byID[aliceID].Name)
	assert.Nil(t, byID[aliceID].AvatarURL)
	assert.Equal(t, "Bob", byID[bobID].Name)
	require.NotNil(t, byID[bobID].AvatarURL)
	assert.Equal(t, avatarURL, *byID[bobID].AvatarURL)
}

// TestHandleJoin_Timeout_Silent verifies that a cancelled context results in
// silent denial — no ACTIVE_USERS frame, user not added to the room.
func TestHandleJoin_Timeout_Silent(t *testing.T) {
	ctx := context.Background()
	h, hub, access, _ := newTestWSHandler(t)
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
	// returns before any enqueue — assert immediately, no GetByIds.
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
	h, hub, access, userRepo := newTestWSHandler(t)
	userID := uuid.New()
	boardID := uuid.New()

	ss, conn := registerTestConn(ctx, hub, h, userID)
	defer ss.disconnect()

	access.EXPECT().CheckViewAccess(mock.Anything, boardID, userID).
		Return(&common.BoardAccess{}, nil).Once()
	// GetByIds is called twice: once for USER_JOINED enrichment, once for ACTIVE_USERS.
	userRepo.On("GetByIds", mock.Anything, []uuid.UUID{userID}).
		Return(map[uuid.UUID]*entity.User{userID: {ID: userID, Name: "Alice"}}, nil).Times(2)

	h.handleJoin(ctx, conn, boardID)
	// After join: USER_JOINED (presence edge) + ACTIVE_USERS from handleJoin.
	require.Eventually(t, func() bool { return len(ss.msgs()) >= 2 }, 2*time.Second, 10*time.Millisecond)

	h.handleLeave(conn, boardID)

	assert.Empty(t, hub.ActiveUsers(boardID), "conn must be removed from the room after leave")
}

// ---------------------------------------------------------------------------
// onPresence — wire-frame broadcasting
// ---------------------------------------------------------------------------

// TestOnPresence_Joined_BroadcastsRichUserJoined verifies that a 0→1 presence
// edge emits a USER_JOINED frame carrying the full {id,name,avatar_url} profile.
func TestOnPresence_Joined_BroadcastsRichUserJoined(t *testing.T) {
	ctx := context.Background()
	h, hub, _, userRepo := newTestWSHandler(t)
	boardID := uuid.New()
	userID := uuid.New()
	avatarURL := "https://example.com/avatar.png"

	// listenerConn fires onPresence(PresenceJoined, listenerID) synchronously — allow it.
	// Must be registered BEFORE listenerConn is called.
	allowAnyGetByIds(userRepo)

	// hub.Join inside listenerConn fires the listener's own USER_JOINED edge.
	// Wait for that initial message so we start from a known baseline.
	lss, cleanup := listenerConn(ctx, hub, boardID)
	defer cleanup()
	require.Eventually(t, func() bool { return len(lss.msgs()) >= 1 }, 2*time.Second, 10*time.Millisecond)

	// Set up enrichment for the explicit onPresence call below.
	userRepo.On("GetByIds", mock.Anything, []uuid.UUID{userID}).
		Return(map[uuid.UUID]*entity.User{
			userID: {ID: userID, Name: "Bob", AvatarURL: &avatarURL},
		}, nil).Once()

	h.onPresence(boardID, userID, realtime.PresenceJoined)

	require.Eventually(t, func() bool { return len(lss.msgs()) >= 2 }, 2*time.Second, 10*time.Millisecond)

	f := findUserJoinedFrameByUserID(lss.msgs(), userID)
	require.NotNil(t, f, "USER_JOINED frame for userID not found")
	assert.Equal(t, realtime.FrameTypeUserJoined, f.Type)
	assert.Equal(t, boardID, f.BoardID)
	assert.Equal(t, userID, f.User.ID)
	assert.Equal(t, "Bob", f.User.Name)
	require.NotNil(t, f.User.AvatarURL)
	assert.Equal(t, avatarURL, *f.User.AvatarURL)
}

// TestOnPresence_Left_BroadcastsThinUserLeft verifies that a 1→0 presence edge
// emits a USER_LEFT frame carrying only {type, board_id, user_id} — no profile lookup.
// Any unexpected GetByIds call would cause testify/mock to panic, acting as the
// regression guard that USER_LEFT stays thin.
func TestOnPresence_Left_BroadcastsThinUserLeft(t *testing.T) {
	ctx := context.Background()
	h, hub, _, userRepo := newTestWSHandler(t)
	boardID := uuid.New()
	userID := uuid.New()

	// listenerConn fires onPresence(PresenceJoined) for listener's own join — allow it.
	allowAnyGetByIds(userRepo)

	lss, cleanup := listenerConn(ctx, hub, boardID)
	defer cleanup()
	require.Eventually(t, func() bool { return len(lss.msgs()) >= 1 }, 2*time.Second, 10*time.Millisecond)

	// onPresence(PresenceLeft) must NOT call GetByIds.
	// No additional stub registered — any unexpected call panics the test.
	h.onPresence(boardID, userID, realtime.PresenceLeft)

	require.Eventually(t, func() bool { return len(lss.msgs()) >= 2 }, 2*time.Second, 10*time.Millisecond)

	f := findUserLeftFrameByUserID(lss.msgs(), userID)
	require.NotNil(t, f, "USER_LEFT frame for userID not found")
	assert.Equal(t, realtime.FrameTypeUserLeft, f.Type)
	assert.Equal(t, boardID, f.BoardID)
	assert.Equal(t, userID, f.UserID, "USER_LEFT must carry only user_id (thin frame)")
}

// ---------------------------------------------------------------------------
// Watcher-not-in-roster enrichment (the whole point of ADR-014)
// ---------------------------------------------------------------------------

// TestHandleJoin_WatcherNotInRoster_GetsEnrichedFrame verifies that a user who
// is NOT a board member (workspace-visible watcher or break-glass admin) still
// gets an enriched ACTIVE_USERS frame. CheckViewAccess grants access even for
// non-members; enrichment goes via userRepo independently of board membership.
func TestHandleJoin_WatcherNotInRoster_GetsEnrichedFrame(t *testing.T) {
	ctx := context.Background()
	h, hub, access, userRepo := newTestWSHandler(t)
	watcherID := uuid.New() // not a board member
	boardID := uuid.New()

	ss, conn := registerTestConn(ctx, hub, h, watcherID)
	defer ss.disconnect()

	// Access is granted via workspace-visible / break-glass path (not board membership).
	access.EXPECT().CheckViewAccess(mock.Anything, boardID, watcherID).
		Return(&common.BoardAccess{}, nil).Once()

	// GetByIds called twice: USER_JOINED enrichment + ACTIVE_USERS enrichment.
	userRepo.On("GetByIds", mock.Anything, []uuid.UUID{watcherID}).
		Return(map[uuid.UUID]*entity.User{
			watcherID: {ID: watcherID, Name: "Watcher", AvatarURL: nil},
		}, nil).Times(2)

	h.handleJoin(ctx, conn, boardID)

	require.Eventually(t, func() bool { return len(ss.msgs()) >= 2 }, 2*time.Second, 10*time.Millisecond)

	raw := findFrameByType(ss.msgs(), realtime.FrameTypeActiveUsers)
	require.NotNil(t, raw)

	var frame realtime.ActiveUsersFrame
	require.NoError(t, json.Unmarshal(raw, &frame))
	require.Len(t, frame.Users, 1)
	assert.Equal(t, watcherID, frame.Users[0].ID)
	assert.Equal(t, "Watcher", frame.Users[0].Name)
}

// ---------------------------------------------------------------------------
// Best-effort error handling (GetByIds failures)
// ---------------------------------------------------------------------------

// TestEnrichUsers_GetByIdsError_FrameStillSent verifies that when GetByIds returns
// an error during ACTIVE_USERS enrichment, the frame is still sent (best-effort /
// degraded with ID-only entries) and no panic occurs.
func TestEnrichUsers_GetByIdsError_FrameStillSent(t *testing.T) {
	ctx := context.Background()
	h, hub, access, userRepo := newTestWSHandler(t)
	userID := uuid.New()
	boardID := uuid.New()

	ss, conn := registerTestConn(ctx, hub, h, userID)
	defer ss.disconnect()

	access.EXPECT().CheckViewAccess(mock.Anything, boardID, userID).
		Return(&common.BoardAccess{}, nil).Once()

	// hub.Join fires onPresence → enrichSingleUser → GetByIds (call #1, USER_JOINED).
	// handleJoin then calls enrichUsers → GetByIds (call #2, ACTIVE_USERS).
	// FIFO: first Once → success (call #1); second Once → error (call #2).
	userRepo.On("GetByIds", mock.Anything, mock.Anything).
		Return(map[uuid.UUID]*entity.User{}, nil).Once()
	userRepo.On("GetByIds", mock.Anything, mock.Anything).
		Return(nil, errors.New("db error")).Once()

	h.handleJoin(ctx, conn, boardID)

	// Frame must still arrive even on enrichment error (best-effort).
	require.Eventually(t, func() bool { return len(ss.msgs()) >= 2 }, 2*time.Second, 10*time.Millisecond)

	raw := findFrameByType(ss.msgs(), realtime.FrameTypeActiveUsers)
	require.NotNil(t, raw, "ACTIVE_USERS frame must be sent even when enrichment fails")

	var frame realtime.ActiveUsersFrame
	require.NoError(t, json.Unmarshal(raw, &frame))
	assert.Equal(t, boardID, frame.BoardID)
	// Degraded: frame still contains the user, with ID only (name="" avatar=nil).
	require.Len(t, frame.Users, 1)
	assert.Equal(t, userID, frame.Users[0].ID)
}

// TestEnrichSingleUser_GetByIdsError_JoinFrameStillSent verifies that when GetByIds
// returns an error for USER_JOINED enrichment in onPresence, the frame is still
// broadcast (degraded with ID only) and no panic occurs.
func TestEnrichSingleUser_GetByIdsError_JoinFrameStillSent(t *testing.T) {
	ctx := context.Background()
	h, hub, _, userRepo := newTestWSHandler(t)
	boardID := uuid.New()
	userID := uuid.New()

	// listenerConn fires onPresence(PresenceJoined, listenerID) synchronously.
	allowAnyGetByIds(userRepo)

	lss, cleanup := listenerConn(ctx, hub, boardID)
	defer cleanup()
	require.Eventually(t, func() bool { return len(lss.msgs()) >= 1 }, 2*time.Second, 10*time.Millisecond)

	// onPresence(PresenceJoined, userID) will call GetByIds — simulate error.
	userRepo.On("GetByIds", mock.Anything, []uuid.UUID{userID}).
		Return(nil, errors.New("db error")).Once()

	h.onPresence(boardID, userID, realtime.PresenceJoined)

	// Frame must still arrive even on enrichment error (best-effort, degraded).
	require.Eventually(t, func() bool { return len(lss.msgs()) >= 2 }, 2*time.Second, 10*time.Millisecond)

	f := findUserJoinedFrameByUserID(lss.msgs(), userID)
	require.NotNil(t, f, "USER_JOINED frame must be sent even when enrichment fails")
	assert.Equal(t, userID, f.User.ID)
	assert.Equal(t, "", f.User.Name, "degraded frame has empty name")
	assert.Nil(t, f.User.AvatarURL, "degraded frame has nil avatar")
}

// TestEnrichUsers_PartialMap_MissingEntrySkipped verifies that when GetByIds returns
// a partial map (some IDs missing), the missing entries are skipped gracefully without
// a nil-deref or panic.
func TestEnrichUsers_PartialMap_MissingEntrySkipped(t *testing.T) {
	ctx := context.Background()
	h, hub, access, userRepo := newTestWSHandler(t)
	aliceID := uuid.New()
	bobID := uuid.New() // will be missing from the returned map
	boardID := uuid.New()

	// hub.Join(boardID, bobConn) fires onPresence(PresenceJoined, bobID) synchronously.
	// Set up stub BEFORE hub.Join.
	userRepo.On("GetByIds", mock.Anything, []uuid.UUID{bobID}).
		Return(map[uuid.UUID]*entity.User{}, nil).Once()

	// Register Bob as an existing conn in the room (direct hub path, no access check).
	bobSock, bobConn := registerTestConn(ctx, hub, h, bobID)
	defer bobSock.disconnect()
	hub.Join(boardID, bobConn) // fires GetByIds([bobID]) synchronously

	aliceSock, aliceConn := registerTestConn(ctx, hub, h, aliceID)
	defer aliceSock.disconnect()

	access.EXPECT().CheckViewAccess(mock.Anything, boardID, aliceID).
		Return(&common.BoardAccess{}, nil).Once()

	// Alice's USER_JOINED enrichment (inside handleJoin → hub.Join → onPresence).
	userRepo.On("GetByIds", mock.Anything, []uuid.UUID{aliceID}).
		Return(map[uuid.UUID]*entity.User{}, nil).Once()
	// ACTIVE_USERS enrichment: Bob's ID is intentionally missing from the returned map.
	userRepo.On("GetByIds", mock.Anything, mock.MatchedBy(func(ids []uuid.UUID) bool {
		return len(ids) == 2
	})).Return(map[uuid.UUID]*entity.User{
		aliceID: {ID: aliceID, Name: "Alice"},
		// bobID intentionally omitted — partial map
	}, nil).Once()

	h.handleJoin(ctx, aliceConn, boardID)

	require.Eventually(t, func() bool { return len(aliceSock.msgs()) >= 2 }, 2*time.Second, 10*time.Millisecond)

	raw := findFrameByType(aliceSock.msgs(), realtime.FrameTypeActiveUsers)
	require.NotNil(t, raw)

	var frame realtime.ActiveUsersFrame
	require.NoError(t, json.Unmarshal(raw, &frame))
	// Bob's entry is skipped (missing from map) — only Alice is returned.
	assert.Len(t, frame.Users, 1)
	assert.Equal(t, aliceID, frame.Users[0].ID)
}

// ---------------------------------------------------------------------------
// disconnect / readPump-EOF
//
// The one end-to-end path nothing else exercises: a socket dying (NOT a
// LEAVE_BOARD frame) drives readPump EOF → onClose → unregisterConn →
// teardown → close(done), and the leaver's last conn trips the 1→0 presence
// edge, emitting USER_LEFT to the surviving room.
// ---------------------------------------------------------------------------

func TestServeWS_DisconnectEOF_RemovesConnAndBroadcastsUserLeft(t *testing.T) {
	ctx := context.Background()
	h, hub, access, userRepo := newTestWSHandler(t)
	boardID := uuid.New()
	leaver, watcher := uuid.New(), uuid.New()

	access.EXPECT().CheckViewAccess(mock.Anything, boardID, leaver).
		Return(&common.BoardAccess{}, nil).Once()
	access.EXPECT().CheckViewAccess(mock.Anything, boardID, watcher).
		Return(&common.BoardAccess{}, nil).Once()

	// Four GetByIds calls during the two joins (don't care about content):
	// 1. Leaver's USER_JOINED enrichment (hub.Join → onPresence → enrichSingleUser)
	// 2. Leaver's ACTIVE_USERS enrichment (enrichUsers after hub.Join)
	// 3. Watcher's USER_JOINED enrichment (hub.Join → onPresence → enrichSingleUser)
	// 4. Watcher's ACTIVE_USERS enrichment (enrichUsers after hub.Join, 2-user snapshot)
	// USER_LEFT on disconnect must NOT trigger a 5th call (regression guard).
	for range 4 {
		userRepo.On("GetByIds", mock.Anything, mock.Anything).
			Return(map[uuid.UUID]*entity.User{}, nil).Once()
	}

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
		return findUserLeftFrameByUserID(watcherSock.msgs(), leaver) != nil
	}, 2*time.Second, 10*time.Millisecond, "watcher must receive a leaver presence frame")
	f := findUserLeftFrameByUserID(watcherSock.msgs(), leaver)
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
	h, hub, _, userRepo := newTestWSHandler(t)
	boardID := uuid.New()
	userID := uuid.New()

	// listenerConn fires onPresence(PresenceJoined) for listener's own join.
	allowAnyGetByIds(userRepo)

	lss, cleanup := listenerConn(ctx, hub, boardID)
	defer cleanup()
	require.Eventually(t, func() bool { return len(lss.msgs()) >= 1 }, 2*time.Second, 10*time.Millisecond)

	initialCount := len(lss.msgs())
	h.onPresence(boardID, userID, realtime.PresenceKind(99))

	time.Sleep(40 * time.Millisecond)
	assert.Len(t, lss.msgs(), initialCount, "unknown PresenceKind must not broadcast any frame")
}
