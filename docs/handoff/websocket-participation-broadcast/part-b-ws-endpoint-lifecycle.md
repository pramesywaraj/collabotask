# Part B — `/ws` endpoint + lifecycle (step ④)

**Written:** 2026-07-31, just-in-time before building Part B (per [index.md](./index.md) line 110).
**Re-read against current code:** `internal/realtime/` (Part A: hub.go, conn.go, socket.go, hub_test.go), `internal/delivery/http/middleware/auth.go` + `csrf.go` + `cors.go`, `internal/config/config.go`, `internal/delivery/http/router/router.go`, `internal/injection/wire.go`, `internal/server/server.go`. `coder/websocket` is **not yet in go.mod**.
**Design source:** [index.md §3 (the `/ws` endpoint)](./index.md), [§4 (presence)](./index.md); [ADR-009](../../architecture/adr/adr-009-websocket-realtime-layer.md); post-③.5 drift recheck B1–B3 in [index.md §Post-③.5 drift recheck](./index.md).

> **Drift review applied 2026-07-31** (`/grill-with-docs`, before building). This doc was checked against the *merged* Part A code + ③.5 middleware and corrected. Key outcomes: dependency path `github.com/coder/websocket` (not `coder.com/…`); `CheckViewAccess` returns `(*BoardAccess, error)`; **Hub is now a standalone `ProvideHub` singleton** (was hidden inside `WSHandler`) so Parts C–E inject it cleanly; ping uses a **bounded** per-ping context; `JOIN_BOARD` access check is **timeout-bounded**; join-denied stays **silent** with the view gating on REST; presence ordering resolved as a **documented client contract**. All decisions carried into the sections below.

---

## Scope — what Part B owns vs. defers (read first)

Part B is the **HTTP + wire-protocol layer**. It wires the real transport to Part A's pure in-memory hub and implements the four client-visible lifecycle events.

**Part B builds:**
- `coder/websocket` dependency + real `socket` adapter (`ws_socket.go`).
- `WSHandler` in the HTTP handler package: upgrade, cookie auth via `middleware.Auth`, `ResponseController` timeout fix, `OriginPatterns` derived from `CORS.AllowedOrigins`.
- Incoming client message parsing: `JOIN_BOARD` / `LEAVE_BOARD` (JSON frames).
- Outgoing presence frames: `ACTIVE_USERS` (joiner-only snapshot), `USER_JOINED` / `USER_LEFT` (room broadcast on 0→1/1→0 edges).
- `JOIN_BOARD` access re-validation via `CheckViewAccess` before admitting the conn to the room.
- Ping/pong liveness (server-driven pings in the `writePump`, each with a **bounded** context).
- CSRF-GET invariant comment in `csrf.go` (drift recheck B3).
- `config.Validate()` fail-fast on unparseable CORS origins (drift recheck B2).
- Wire registration of `Hub` (standalone `ProvideHub` singleton) and `WSHandler`.

> The Part A code-review action items (`hub.go` comment reword + naming the `presenceEdge` struct) are **already merged** ([hub.go:126](../../../collabotask-backend/internal/realtime/hub.go#L126), [hub.go:164](../../../collabotask-backend/internal/realtime/hub.go#L164)) — nothing to do.

**Part B explicitly defers:**
| Deferred to | What |
|---|---|
| **Part C** | `Broadcaster` port in `usecase/common`; typed event structs; `{type, payload}` envelope; mutation broadcasts (~16 events). Part B's `hub.Broadcast(boardID, []byte)` is the raw fan-out Part C's adapter will call. |
| **Part D** | `ACCESS_REVOKED` wire frame delivered before eviction teardown. The `reason` param is already threaded through all eviction paths — Part D adds only the frame send, no signature churn. |
| **Part E** | Workspace-cascade fan-out (`WorkspaceUseCase` gains `Broadcaster`). |

---

## Dependency to add

```bash
cd collabotask-backend
go get github.com/coder/websocket@latest
```

The module path is `github.com/coder/websocket` (latest `v1.8.x`). Import it as `"github.com/coder/websocket"`. **Not** `coder.com/…` — that path does not exist.

---

## Files to create or modify

```
internal/realtime/
  socket.go           ← MODIFY: add Ping(ctx) to the interface; update doc comment
  ws_socket.go        ← NEW: *websocket.Conn → socket adapter
  conn.go             ← MODIFY: readPump onMessage callback (with ctx); writePump ping ticker
                               (bounded ctx); Conn gains done chan struct{} + Done(); add "time" import
  hub.go              ← MODIFY: Register gains onMsg param; NewHub() drops the callback arg
                               and gains SetPresence() (standalone-singleton wiring)
  message.go          ← NEW: incoming + outgoing wire frame types

internal/delivery/http/handler/
  ws_handler.go       ← NEW: WSHandler + ServeWS

internal/delivery/http/middleware/
  csrf.go             ← MODIFY: add CSRF-GET invariant comment only (no logic change)

internal/delivery/http/router/
  router.go           ← MODIFY: add WSHandler to Config + register GET /ws route

internal/config/
  config.go           ← MODIFY: WSOriginPatterns() helper + Validate() fail-fast

internal/injection/
  providers.go        ← MODIFY: add ProvideHub + ProvideWSHandler funcs; thread WSHandler into ProvideRouter
  wire.go             ← MODIFY: add ProvideHub + ProvideWSHandler to the sets (function bodies live in providers.go)
  wire_gen.go         ← REGENERATE (run wire ./internal/injection/)
```

---

## Key changes to existing Part A types

### `socket` interface (`socket.go`)

Add `Ping`:

```go
type socket interface {
    Write(ctx context.Context, data []byte) error
    // Read blocks until a message arrives or the connection closes.
    Read(ctx context.Context) ([]byte, error)
    // Ping sends a ping and blocks until the pong arrives or ctx is cancelled.
    // Used by writePump to detect half-open connections.
    Ping(ctx context.Context) error
    Close(reason string) error
}
```

The `fakeSocket` in `hub_test.go` gains a `Ping` stub (returns `nil`). Existing tests are unaffected.

### `Conn` (`conn.go`)

Add a `done` channel so `ServeWS` can block until teardown without polling:

```go
type Conn struct {
    connID uuid.UUID
    userID uuid.UUID
    s      socket

    send chan []byte
    done chan struct{}  // ← new; closed by teardown
    once sync.Once

    rooms map[uuid.UUID]struct{}
}

func newConn(userID uuid.UUID, s socket) *Conn {
    return &Conn{
        connID: uuid.New(),
        userID: userID,
        s:      s,
        send:   make(chan []byte, sendBufferSize),
        done:   make(chan struct{}),          // ← new
        rooms:  make(map[uuid.UUID]struct{}),
    }
}

// Done returns a channel that is closed when the connection is torn down.
func (c *Conn) Done() <-chan struct{} { return c.done }

func (c *Conn) teardown(reason string) {
    c.once.Do(func() {
        close(c.send)
        _ = c.s.Close(reason)
        close(c.done)    // ← new
    })
}
```

**`readPump`** — add `onMsg` callback (called with each inbound message before closure):

```go
// MessageHandler is called by readPump for each inbound message. It receives the
// pump's ctx so a handler that does I/O (e.g. the JOIN_BOARD access check) can bound
// it. The handler runs inside the readPump goroutine, so it must not block
// *unboundedly* — one slow message stalls only THIS connection's inbound processing,
// but it must still finish (hence the bounded ctx in handleJoin). See Q5, index.md.
type MessageHandler func(ctx context.Context, c *Conn, data []byte)

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
```

**`writePump`** — add ping ticker:

`sendBufferSize` already exists in `conn.go`; **add only the new constants** (and a `"time"` import to `conn.go`). Do **not** re-declare `sendBufferSize` — that's a compile error.

```go
const (
    pingInterval = 30 * time.Second // how often the server pings an idle socket
    pongWait     = 10 * time.Second // per-ping deadline: no pong in this window ⇒ half-open ⇒ drop
)

func (c *Conn) writePump(ctx context.Context, onClose func(string)) {
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
            // board (no broadcasts to trip the slow-consumer drop). See Q4, index.md.
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
```

### `Hub` construction (`hub.go`) — standalone singleton (Q1)

Part A's `NewHub(onPresence)` bakes the presence callback into the constructor, which forces the hub to be built *inside* `WSHandler` (the callback is a handler method). That hides the hub from Wire, and Parts C/D/E all need to inject it. **Decouple construction from the callback** so the hub can be a first-class `ProvideHub` singleton:

```go
func NewHub() *Hub {
    return &Hub{rooms: make(map[uuid.UUID]map[uuid.UUID]map[uuid.UUID]*Conn)}
}

// SetPresence wires the edge callback. Called once at startup (in ProvideWSHandler),
// before any conn registers — so no lock is needed and there is no race.
func (h *Hub) SetPresence(fn func(boardID, userID uuid.UUID, kind PresenceKind)) {
    h.onPresence = fn
}
```

This is safe because `onPresence` only ever calls `h.hub.Broadcast(...)` (nothing else on the handler), and it's set once before traffic. `WSHandler` now *receives* the hub and calls `hub.SetPresence(h.onPresence)` — see `NewWSHandler` below.

### `Hub.Register` (`hub.go`)

Add the `onMsg` parameter and pass `c` directly so the closure is capture-safe:

```go
func (h *Hub) Register(ctx context.Context, userID uuid.UUID, s socket, onMsg MessageHandler) *Conn {
    c := newConn(userID, s)
    teardown := func(reason string) { h.unregisterConn(c, reason) }
    go c.readPump(ctx, onMsg, teardown)
    go c.writePump(ctx, teardown)
    return c
}
```

`c` is created before either goroutine starts, so the closure is safe — `onMsg(ctx, c, data)` in `readPump` always sees the correct `c`.

Update the existing `hub_test.go`: change `newHub` from `realtime.NewHub(rec.record)` to `h := realtime.NewHub(); h.SetPresence(rec.record)`, and add `nil` as the `onMsg` argument to every `Register` call. All 19 tests remain green (nil onMsg is a no-op; `SetPresence` before any Register is equivalent to the old constructor wiring).

---

## New files

### `internal/realtime/ws_socket.go`

Thin adapter so `*websocket.Conn` satisfies the `socket` seam:

```go
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
```

### `internal/realtime/message.go`

Incoming client frames and outgoing presence frames. Part C's event frames (`CARD_MOVED`, etc.) are not here — they live in `usecase/common` as typed structs marshaled by the Broadcaster adapter.

```go
package realtime

import "github.com/google/uuid"

// Incoming message types (client → server).
const (
    MsgTypeJoinBoard  = "JOIN_BOARD"
    MsgTypeLeaveBoard = "LEAVE_BOARD"
)

// IncomingMessage is the minimal envelope for client → server frames.
// Unmarshal only; validate BoardID before acting.
type IncomingMessage struct {
    Type    string    `json:"type"`
    BoardID uuid.UUID `json:"board_id"`
}

// Outgoing presence frame types (server → client).
const (
    FrameTypeActiveUsers = "ACTIVE_USERS"
    FrameTypeUserJoined  = "USER_JOINED"
    FrameTypeUserLeft    = "USER_LEFT"
)

// ActiveUsersFrame is sent to the joining conn only on JOIN_BOARD success.
type ActiveUsersFrame struct {
    Type    string      `json:"type"`
    BoardID uuid.UUID   `json:"board_id"`
    UserIDs []uuid.UUID `json:"user_ids"`
}

// UserPresenceFrame is broadcast to the room on a 0→1 (USER_JOINED) or 1→0 (USER_LEFT) edge.
type UserPresenceFrame struct {
    Type    string    `json:"type"`
    BoardID uuid.UUID `json:"board_id"`
    UserID  uuid.UUID `json:"user_id"`
}
```

### `internal/delivery/http/handler/ws_handler.go`

```go
package handler

import (
    "context"
    "encoding/json"
    "net/http"
    "time"

    "collabotask/internal/delivery/http/middleware"
    "collabotask/internal/realtime"
    "collabotask/internal/usecase/common"

    "github.com/coder/websocket"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

// WSHandler upgrades HTTP → WebSocket and drives the room lifecycle.
// The Hub is INJECTED (standalone ProvideHub singleton). The handler supplies the
// onPresence + message callbacks and wires the former via hub.SetPresence.
type WSHandler struct {
    hub            *realtime.Hub
    access         common.BoardAccessChecker
    originPatterns []string
}

func NewWSHandler(hub *realtime.Hub, access common.BoardAccessChecker, originPatterns []string) *WSHandler {
    h := &WSHandler{
        hub:            hub,
        access:         access,
        originPatterns: originPatterns,
    }
    hub.SetPresence(h.onPresence) // once, at startup, before any conn registers
    return h
}

// No Hub() accessor: Parts C/D/E inject the same *realtime.Hub straight from
// ProvideHub, so the hub is never reached through the handler.

// ServeWS upgrades the connection, registers it with the hub, and blocks until
// the connection closes. Auth is already enforced by middleware.Auth upstream.
func (h *WSHandler) ServeWS(c *gin.Context) {
    userID, ok := middleware.GetUserID(c)
    if !ok {
        c.AbortWithStatus(http.StatusUnauthorized)
        return
    }

    // Clear this connection's write deadline — the global WriteTimeout would
    // kill a long-lived idle socket. REST connections keep the 30s deadline.
    rc := http.NewResponseController(c.Writer)
    _ = rc.SetWriteDeadline(time.Time{})
    _ = rc.SetReadDeadline(time.Time{})

    wsConn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
        OriginPatterns: h.originPatterns,
    })
    if err != nil {
        // websocket.Accept has already written the HTTP error response.
        return
    }
    // Safety net only: teardown does a graceful Close(StatusNormalClosure); this
    // CloseNow is the belt-and-suspenders backstop if we return on an unexpected path.
    // Calling CloseNow after a graceful Close is a no-op — not a double-close bug.
    defer wsConn.CloseNow()

    s := realtime.NewWSSocket(wsConn)
    conn := h.hub.Register(c.Request.Context(), userID, s, h.handleMessage)

    <-conn.Done() // block until pumps finish + teardown
}

// dbJoinCheckTimeout bounds the JOIN_BOARD access check (3 DB round-trips) so a
// slow/hung DB can't strand this connection's readPump. See Q5, index.md.
const dbJoinCheckTimeout = 5 * time.Second

// handleMessage is the readPump callback for JOIN_BOARD and LEAVE_BOARD frames.
// ctx is the pump's context (used to bound the access check).
func (h *WSHandler) handleMessage(ctx context.Context, conn *realtime.Conn, data []byte) {
    var msg realtime.IncomingMessage
    if err := json.Unmarshal(data, &msg); err != nil {
        return // discard malformed frames
    }
    switch msg.Type {
    case realtime.MsgTypeJoinBoard:
        h.handleJoin(ctx, conn, msg.BoardID)
    case realtime.MsgTypeLeaveBoard:
        h.handleLeave(conn, msg.BoardID)
    }
}

// handleJoin re-validates access and admits the conn to the room.
//
// On denial (or timeout) it returns SILENTLY — no ACCESS_REVOKED, no error frame.
// This is deliberate (Q2): the authoritative access decision + user-facing feedback
// is the REST kanban fetch (same CheckViewAccess). The client MUST gate the board
// view on that REST response, never on receiving ACTIVE_USERS. Silence also honours
// the 404-hide invariant (an error frame would leak that a private board exists).
func (h *WSHandler) handleJoin(ctx context.Context, conn *realtime.Conn, boardID uuid.UUID) {
    ctx, cancel := context.WithTimeout(ctx, dbJoinCheckTimeout)
    defer cancel()

    // Re-validate: room visibility == kanban visibility; break-glass falls out.
    if _, err := h.access.CheckViewAccess(ctx, boardID, conn.UserID()); err != nil {
        return // denied or timed out — silently ignore (see doc comment above)
    }
    snapshot := h.hub.Join(boardID, conn)
    frame, _ := json.Marshal(realtime.ActiveUsersFrame{
        Type:    realtime.FrameTypeActiveUsers,
        BoardID: boardID,
        UserIDs: snapshot,
    })
    conn.Send(frame) // send ACTIVE_USERS to the joining conn only
}

// handleLeave removes the conn from the room. The 1→0 presence edge (if it fires)
// will broadcast USER_LEFT via onPresence.
func (h *WSHandler) handleLeave(conn *realtime.Conn, boardID uuid.UUID) {
    h.hub.Leave(boardID, conn)
}

// onPresence is the Hub's edge callback. Called outside the hub lock.
// Broadcasts USER_JOINED or USER_LEFT to the room (including the joining user's other tabs).
//
// Ordering note (Q3): on a 0→1 join this fires INSIDE hub.Join, so the joiner receives
// USER_JOINED{self} on its own channel BEFORE handleJoin sends ACTIVE_USERS. That's
// fine by contract — the client treats ACTIVE_USERS as an authoritative snapshot
// (replace) and USER_JOINED/USER_LEFT as idempotent deltas, so a self-echo before the
// snapshot is a no-op. We do NOT refactor to suppress it (would split the presence seam
// and weaken Join's invariant); see the "Client wire contract" section below.
func (h *WSHandler) onPresence(boardID, userID uuid.UUID, kind realtime.PresenceKind) {
    var frameType string
    switch kind {
    case realtime.PresenceJoined:
        frameType = realtime.FrameTypeUserJoined
    case realtime.PresenceLeft:
        frameType = realtime.FrameTypeUserLeft
    default:
        return
    }
    frame, _ := json.Marshal(realtime.UserPresenceFrame{
        Type:    frameType,
        BoardID: boardID,
        UserID:  userID,
    })
    h.hub.Broadcast(boardID, frame)
}
```

**Additional exported items needed on `Conn`** (add to `conn.go`):
```go
// UserID exposes the authenticated user for access checks in the handler.
func (c *Conn) UserID() uuid.UUID { return c.userID }

// Send queues msg on the send channel (non-blocking; drops if buffer full).
// The writePump is the sole goroutine that actually writes to the socket.
func (c *Conn) Send(msg []byte) {
    select {
    case c.send <- msg:
    default: // treat same as slow-consumer; hub will drop the conn if Broadcast fills it
    }
}
```

---

## Changes to existing files

### `internal/delivery/http/middleware/csrf.go`

Add the invariant comment after the existing `csrfSafeMethods` comment block (no logic change):

```go
// CSRF requires the X-CSRF-Protection header on all state-changing requests.
// Paired with SameSite=Lax cookies and an explicit CORS allowlist, this closes
// the cross-origin mutation vector for sibling-subdomain topologies.
//
// INVARIANT: do not extend CSRF checks to GET/HEAD/OPTIONS.
// The WebSocket handshake is a GET; the browser WebSocket API cannot send
// X-CSRF-Protection. Handshake CSWSH defense lives in OriginPatterns +
// SameSite=Lax, not in this middleware. See ws_handler.go + index.md §Post-③.5.
func CSRF() gin.HandlerFunc {
```

### `internal/config/config.go`

Add a `WSOriginPatterns()` helper + a fail-fast validation guard (drift recheck B2). Keep `CORS.AllowedOrigins` as the single source of truth — never add a second WS-specific env var.

```go
import "net/url"

// WSOriginPatterns derives host-only patterns for coder/websocket's OriginPatterns
// from the CORS allowlist. CORS stores full origins (with scheme) for reflection;
// websocket.Accept's OriginPatterns matches only the host part.
// Called once at startup after Validate() ensures every origin is parseable.
func (c *Config) WSOriginPatterns() []string {
    patterns := make([]string, 0, len(c.CORS.AllowedOrigins))
    for _, origin := range c.CORS.AllowedOrigins {
        if origin == "*" {
            continue // "*" has no host; cookie auth forbids it anyway (guarded in Validate)
        }
        u, _ := url.Parse(origin) // validated in Validate(); no error possible here
        patterns = append(patterns, u.Host)
    }
    return patterns
}
```

Add to `Validate()`, after the `*`+credentials guard:

```go
// Fail-fast: unparseable CORS origin → WS origin pattern derivation would silently
// reject every cross-origin handshake. Mirrors the *+credentials guard above.
for _, origin := range c.CORS.AllowedOrigins {
    if origin == "*" {
        continue // already guarded above
    }
    u, err := url.Parse(origin)
    if err != nil || u.Host == "" {
        return fmt.Errorf("CORS_ALLOWED_ORIGINS: %q is not a valid origin URL (needs scheme + host)", origin)
    }
}
```

### `internal/delivery/http/router/router.go`

Add `WSHandler` to the `Config` struct and register the `/ws` route. The route must be **outside** the `/api/v1` group (it's not a REST endpoint) and mounted **under `middleware.Auth`** (abort-before-upgrade is the goal):

```go
type Config struct {
    Cfg              *config.Config
    Log              *logger.Logger
    AuthHandler      *handler.AuthHandler
    UserHandler      *handler.UserHandler
    WorkspaceHandler *handler.WorkspaceHandler
    BoardHandler     *handler.BoardHandler
    ColumnHandler    *handler.ColumnHandler
    CardHandler      *handler.CardHandler
    WSHandler        *handler.WSHandler        // ← new
}

func New(cfg Config) *gin.Engine {
    // ... existing setup ...

    // WebSocket — not under /api/v1; auth middleware aborts with 401 before upgrade
    // on missing/invalid cookie, so the handler always has a valid userID.
    ws := routes.Group("/ws")
    ws.Use(middleware.Auth(&cfg.Cfg.Auth))
    {
        ws.GET("", cfg.WSHandler.ServeWS)
    }

    // ... existing v1Routes ...
    return routes
}
```

### `internal/injection/` — `providers.go` (functions) + `wire.go` (sets)

Provider **functions** go in [`providers.go`](../../../collabotask-backend/internal/injection/providers.go) (alongside `ProvideRouter`, `ProvideBoardAccessChecker`, …); only **set membership** goes in [`wire.go`](../../../collabotask-backend/internal/injection/wire.go). `ProvideHub` **is** a standalone provider now (Q1) — a singleton both `WSHandler` and Part C's Broadcaster adapter consume. `providers.go` must import `collabotask/internal/realtime`.

```go
// providers.go — new functions

// ProvideHub is the single in-memory hub for the process (one registry, one lock).
// Its presence callback is wired by ProvideWSHandler via hub.SetPresence.
func ProvideHub() *realtime.Hub {
    return realtime.NewHub()
}

func ProvideWSHandler(
    cfg *config.Config,
    hub *realtime.Hub,
    access common.BoardAccessChecker, // NOT usecase.BoardAccessChecker — it lives in usecase/common
) *handler.WSHandler {
    return handler.NewWSHandler(hub, access, cfg.WSOriginPatterns())
}

// ProvideRouter gains the wsH param + threads it into router.Config:
func ProvideRouter(
    cfg *config.Config, log *logger.Logger,
    authHandler *handler.AuthHandler, userHandler *handler.UserHandler,
    workspaceHandler *handler.WorkspaceHandler, boardHandler *handler.BoardHandler,
    columnHandler *handler.ColumnHandler, cardHandler *handler.CardHandler,
    wsHandler *handler.WSHandler, // ← new
) *gin.Engine {
    return router.New(router.Config{
        Cfg: cfg, Log: log,
        AuthHandler: authHandler, UserHandler: userHandler,
        WorkspaceHandler: workspaceHandler, BoardHandler: boardHandler,
        ColumnHandler: columnHandler, CardHandler: cardHandler,
        WSHandler: wsHandler, // ← new
    })
}
```

```go
// wire.go — add to the sets (ProvideHub before ProvideWSHandler, which depends on it):
HandlerSet = wire.NewSet(
    ProvideAuthHandler,
    ProvideUserHandler,
    ProvideWorkspaceHandler,
    ProvideBoardHandler,
    ProvideColumnHandler,
    ProvideCardHandler,
    ProvideHub,       // ← new (singleton; Part C injects the same *realtime.Hub)
    ProvideWSHandler, // ← new
)
```

Regenerate after edits: `wire ./internal/injection/` (Wire resolves the `wsHandler` param of `ProvideRouter` automatically once `ProvideWSHandler` is in the graph).

---

## Client wire contract (Q2/Q3 — documentation, must also land in SRS §4.5/§5)

These are **doc obligations, not code.** They protect the silent-drop (Q2) and
edge-triggered-presence (Q3) decisions from a naive frontend. Carry them into SRS
§4.5/§5 and the frontend contract:

1. **The board view gates on the REST kanban fetch, never on `ACTIVE_USERS`.** A denied
   `JOIN_BOARD` is silent; the authoritative 404/403 comes from `GET .../kanban` (same
   `CheckViewAccess`). Treat `ACTIVE_USERS` as a presence bonus on an already-granted
   view — do **not** show a spinner waiting for it.
2. **`ACTIVE_USERS` is an authoritative full snapshot — apply as REPLACE**, even if you
   already hold presence state. Never "skip the snapshot because I already have data."
3. **`USER_JOINED` / `USER_LEFT` are idempotent deltas.** Adding an already-present user
   or removing an absent one is a no-op. Ordering relative to the snapshot is not
   guaranteed (3rd-party joins race) — the snapshot is the source of truth.
4. **You may receive `USER_JOINED` for yourself** on your own join (0→1 self-echo). It's
   an idempotent no-op; it only matters for join *toasts*, where you suppress `userID == self`.

---

## Design decisions inherited from index.md / drift recheck (do not re-litigate)

| Decision | Why locked |
|---|---|
| Mount `/ws` under `middleware.Auth`; handler reads `GetUserID(c)` — **no inline cookie read** | Drift recheck B1. One auth path shared with REST. Phase-2 session swap changes one place. |
| Derive `OriginPatterns` from `CORS.AllowedOrigins` via `url.Parse().Host`; fail-fast in `Validate()` | Drift recheck B2. `OriginPatterns` matches host only; CORS stores full origins. Single source of truth, no second env var. |
| CSRF middleware passes the GET handshake — **do not change this** | Drift recheck B3. Browser `WebSocket` API cannot send `X-CSRF-Protection`. Handshake CSWSH defense = `OriginPatterns` + `SameSite=Lax`. |
| `http.ResponseController` clears **this connection's** deadlines only | Index.md §3. REST keeps its 30s. WS liveness = ping/pong. |
| `JOIN_BOARD` re-validates via `CheckViewAccess` before admitting to room | Index.md §4. Room visibility == kanban visibility; break-glass (workspace admin) falls out. |
| `ACTIVE_USERS` sent to joiner only (snapshot of distinct user_ids incl. self) | Index.md §4. Never queries DB — built from the in-memory room. |
| `USER_JOINED`/`USER_LEFT` broadcast to room (incl. sender's other tabs) | Index.md §4 + §6. Include sender — REST mutations carry no WS identity to exclude. |
| `ACCESS_REVOKED` frame is **Part D** — not Part B | The `reason` param is already threaded; Part D adds only the frame delivery. |
| **Hub is a standalone `ProvideHub` singleton** (2026-07-31 review, Q1) | Was hidden in `WSHandler`; Wire can't inject a method-exposed hub into Parts C–E usecases. A singleton keeps the graph clean; presence callback wired via `hub.SetPresence`. |
| **Presence ordering/self-echo: document the client contract, don't refactor** (Q3) | Snapshot-replace + idempotent-delta is required for 3rd-party join races anyway, so it absorbs the self-echo for free. Refactoring to suppress would split the presence seam + weaken `Join`'s invariant. |
| **`JOIN_BOARD` denied → silent** (Q2) | The REST kanban fetch is the authoritative access signal; the view gates on it, not `ACTIVE_USERS`. Silence also preserves the 404-hide invariant. |
| **Ping + `JOIN_BOARD` access check are timeout-bounded** (Q4/Q5) | Cleared socket deadlines + `context.Background()` would let a half-open socket / hung DB pin a pump. Bounded `ctx` (`pongWait` / `dbJoinCheckTimeout`) restores liveness. |

---

## Test checklist (`go test -race ./internal/realtime/...` + handler tests)

### `hub_test.go` — update existing tests

- [ ] `newHub` helper: `realtime.NewHub(rec.record)` → `h := realtime.NewHub(); h.SetPresence(rec.record)`.
- [ ] Update all `hub.Register(ctx, userID, s)` calls to `hub.Register(ctx, userID, s, nil)` (new `onMsg` param).
- [ ] `fakeSocket` gains a `Ping(ctx) error` stub returning `nil`.
- [ ] All 19 existing tests pass (SetPresence-before-Register ≡ old constructor wiring; nil onMsg is a no-op).

### New: `ws_socket_test.go` (or inline in handler test)

- [ ] `wsSocket.Write` calls `conn.Write(ctx, MessageText, data)`.
- [ ] `wsSocket.Read` unwraps `conn.Read(ctx)` and returns `data, err`.
- [ ] `wsSocket.Ping` delegates to `conn.Ping(ctx)`.
- [ ] `wsSocket.Close` calls `conn.Close(StatusNormalClosure, reason)`.

### New: presence wire-frame tests (`realtime_test` package or `handler_test`)

Using `fakeSocket` + a mock `BoardAccessChecker`. `handleMessage` now takes `ctx` — pass `context.Background()` (or a test ctx) in tests:

- [ ] **`JOIN_BOARD` — access granted:** `handleMessage` with valid JOIN_BOARD → `CheckViewAccess` called → `hub.Join` called → `ACTIVE_USERS` queued on the joining conn's send channel → if 0→1 edge, `USER_JOINED` broadcast to room.
- [ ] **`JOIN_BOARD` — access denied:** `CheckViewAccess` returns error → **silent**: no hub.Join, no ACTIVE_USERS, no broadcast, no error frame.
- [ ] **`JOIN_BOARD` — access check times out:** mock `CheckViewAccess` returns `context.DeadlineExceeded` → treated exactly like denial (silent, no join). Confirms the `dbJoinCheckTimeout` path is a no-op on the room.
- [ ] **`JOIN_BOARD` — duplicate (multi-tab):** second conn for same user → hub.Join called (second tab enters room) → 0→1 edge **not** fired (user already present) → no `USER_JOINED` broadcast; ACTIVE_USERS sent with correct snapshot (user appears once).
- [ ] **`LEAVE_BOARD`:** `handleLeave` → `hub.Leave` → if 1→0 edge, `USER_LEFT` broadcast to room.
- [ ] **`LEAVE_BOARD` — multi-tab, not last:** hub.Leave called → 1→0 edge not fired → no `USER_LEFT` broadcast.
- [ ] **Disconnect (readPump EOF):** conn removed from all its rooms; 1→0 edges fire `USER_LEFT` for each; `conn.Done()` closed.
- [ ] **Malformed frame:** `handleMessage` with non-JSON → discarded silently, no panic.
- [ ] **Unknown frame type:** `handleMessage` with `{"type":"UNKNOWN"}` → discarded, no panic.

> **Deferred (recorded gap):** ping/pong liveness (`pingInterval`/`pongWait`) is **not** unit-tested — a real test would need an injectable clock (30s ticker). Covered by the live-client checkpoint below. Revisit if/when tuning constants move to config (index.md optimization notes).

### `config_test.go` (extend existing)

- [ ] `WSOriginPatterns()` on `{"https://app.collabotask.com","http://localhost:3000"}` → `{"app.collabotask.com","localhost:3000"}`.
- [ ] `WSOriginPatterns()` skips `"*"` → `{"*"}` yields `{}` (no empty-host pattern).
- [ ] `Validate()` rejects `"not-a-url"` in `CORS_ALLOWED_ORIGINS`.
- [ ] `Validate()` rejects an origin with no host (e.g., `"https://"`).
- [ ] `Validate()` accepts `"http://localhost:3000"` (port included in host).

---

## Checkpoint (done-when)

A real client can connect to `/ws` with a valid auth cookie, send a `JOIN_BOARD` frame, see an `ACTIVE_USERS` response, and see `USER_JOINED`/`USER_LEFT` events as other tabs join and leave. The connection survives a reconnect (new socket, JOIN_BOARD re-sent). Auth guard rejects `401` before upgrade on missing/invalid cookie. All tests green under `go test -race ./...`. → unblocks **Part C** (Broadcaster port + mutation broadcasts).

---

## Note for Part C

Part C injects the **same `*realtime.Hub`** straight from `ProvideHub` (Q1 — there is no `WSHandler.Hub()` accessor). The Broadcaster port surface (`Broadcast(boardID, Event)`, `EvictUser`, `EvictExcept`, `EvictUserFromRooms`) wraps `Hub` methods directly — the adapter owns the marshaling from typed event structs to `[]byte`. Part B's `hub.Broadcast(boardID, []byte)` is the underlying fan-out that the adapter calls.
