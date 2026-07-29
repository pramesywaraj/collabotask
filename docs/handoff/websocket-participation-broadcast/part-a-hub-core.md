# Part A — Hub core (step ④)

**Written:** 2026-07-29, just-in-time before building Part A (per [index.md](./index.md) line 110).
**Re-read against current code:** `internal/usecase/common/*`, `internal/domain/entity/*` (IDs are `uuid.UUID`), test conventions (`package X_test`, testify `require`/`mock`, table-driven with a `newXDeps(t)` helper). No `internal/realtime` package exists yet.
**Design source:** [index.md §2 (the hub)](./index.md), [§4 (presence)](./index.md), [§5 (eviction)](./index.md); [ADR-009](../../architecture/adr/adr-009-websocket-realtime-layer.md).

---

## Scope — what Part A owns vs. defers (read first)

Part A is the **pure in-memory concurrency core**. No HTTP, no wire protocol, no `coder/websocket` import. Everything is unit-testable under `go test -race` with a fake socket.

**Part A builds:**
- The registry `map[boardID]map[userID]map[connID]*Conn` + `sync.RWMutex`.
- `Conn`: buffered `send` channel, the **`writePump`** (sole socket writer), a **`readPump`** that only detects socket closure → triggers teardown (no message parsing yet), `sync.Once` teardown.
- Room mechanics: `Join`, `Leave`, disconnect-removes-from-all-rooms.
- **Edge-triggered presence** computation (0→1 / 1→0 detection, multi-tab suppression) surfaced through the `onPresence` seam + `ActiveUsers` snapshot.
- Low-level fan-out `Broadcast(boardID, []byte)` with **slow-consumer drop**.
- The three **eviction primitives** — `EvictUser`, `EvictExcept`, `EvictUserFromRooms` — as *mechanics only* (unregister + teardown, carrying a `reason`).

**Part A explicitly defers (boundary notes so later parts don't get re-litigated):**
| Deferred to | What |
|---|---|
| **Part B** | HTTP upgrade, cookie auth, `OriginPatterns`, `ResponseController` timeout; `coder/websocket` dependency + the real `socket` adapter; readPump **message parsing** (`JOIN_BOARD`/`LEAVE_BOARD`); turning presence edges into `USER_JOINED`/`USER_LEFT`/`ACTIVE_USERS` **wire frames**; ping/pong loop. |
| **Part C** | The `Broadcaster` **port** in `usecase/common`, typed event structs, `{type,payload}` envelope marshaling. Part A's `Broadcast([]byte)` is the raw fan-out the port will wrap. |
| **Part D** | The **`ACCESS_REVOKED` wire frame** delivered before teardown. Part A threads the `reason` param through eviction so Part D wires the frame **without changing signatures**. |

---

## Package & files

New package `internal/realtime` (concrete adapter; the `Broadcaster` **port** it will satisfy lives in `usecase/common`, added in Part C — keeps the dependency rule: usecase owns the interface, `realtime` implements it).

```
internal/realtime/
  hub.go            Hub: registry, RWMutex, Join/Leave/Broadcast/ActiveUsers, the 3 evict primitives
  conn.go           Conn: send chan, writePump, readPump (closure-detect), sync.Once teardown
  socket.go         the `socket` seam interface (+ doc; Part B provides the coder/websocket impl)
  hub_test.go       -race unit tests (package realtime_test) + a fakeSocket
```

---

## Key types & signatures

```go
package realtime

// socket is the minimal transport the pumps drive. Part B wraps
// *coder/websocket.Conn; Part A tests use fakeSocket. Shaped to match
// coder/websocket so the Part B adapter is trivial (Write always MessageText).
type socket interface {
    Write(ctx context.Context, data []byte) error
    Read(ctx context.Context) ([]byte, error) // Part A: only its error signals closure
    Close(reason string) error
}

// PresenceKind is the edge reported through the onPresence seam.
type PresenceKind int
const ( PresenceJoined PresenceKind = iota; PresenceLeft )

// Conn is one WebSocket connection (one browser tab). Not created directly in
// tests via the socket seam.
type Conn struct { /* connID uuid.UUID; userID uuid.UUID; send chan []byte; rooms set; once sync.Once; ... */ }

type Hub struct { /* mu sync.RWMutex; rooms map[uuid.UUID]map[uuid.UUID]map[uuid.UUID]*Conn; onPresence func(...) */ }

func NewHub(onPresence func(boardID, userID uuid.UUID, kind PresenceKind)) *Hub

// Registers a fresh connection and starts its pumps. connID minted internally (uuid.New()).
func (h *Hub) Register(userID uuid.UUID, s socket) *Conn

// Room membership. Join returns the ACTIVE_USERS snapshot (distinct user_ids in
// the room, incl. self) for the joiner. Both funnel 0→1 / 1→0 edges to onPresence.
func (h *Hub) Join(boardID uuid.UUID, c *Conn) (activeUsers []uuid.UUID)
func (h *Hub) Leave(boardID uuid.UUID, c *Conn)

// Snapshot of distinct connected user_ids in a room (in-memory, never the DB roster).
func (h *Hub) ActiveUsers(boardID uuid.UUID) []uuid.UUID

// Raw fan-out to a room. RLock; non-blocking send per Conn; a full buffer drops
// that Conn (teardown). Part C's Broadcaster.Broadcast wraps this after marshaling.
func (h *Hub) Broadcast(boardID uuid.UUID, msg []byte)

// Eviction — mechanics only (Part A). reason is threaded for Part D's ACCESS_REVOKED frame.
func (h *Hub) EvictUser(boardID, userID uuid.UUID, reason string)
func (h *Hub) EvictExcept(boardID uuid.UUID, allowed []uuid.UUID, reason string)
func (h *Hub) EvictUserFromRooms(userID uuid.UUID, boardIDs []uuid.UUID, reason string)
```

---

## Design decisions (the non-obvious calls — flag if you disagree)

1. **`socket` seam, no `coder/websocket` in Part A.** "Pure in-memory, no HTTP" + `-race` tests force a transport seam. The interface mirrors coder/websocket so Part B's adapter is a thin wrapper. Keeps Part A dependency-free and fast.
2. **`onPresence` callback instead of emitting wire frames.** Part A computes edges; the callback reports them. Tests inject a recorder to assert edge-triggering directly; Part B injects the real `USER_JOINED`/`USER_LEFT` broadcaster. Keeps Part A wire-free while still owning the edge logic (which is where the concurrency bugs live).
3. **Eviction = unregister + teardown + `reason`.** The `ACCESS_REVOKED` *frame* is Part D. Justified by the index map: Part D "**Builds:** `ACCESS_REVOKED`". Part A gives the mechanism; the `reason` param already in the signature means Part D changes zero signatures.
4. **readPump detects closure only.** Message parsing (`JOIN_BOARD`/`LEAVE_BOARD`) is Part B. In Part A the readPump reads-and-discards; a read error ⇒ teardown. `fakeSocket.Read` returning `io.EOF` simulates disconnect for teardown tests.
5. **Bounded `send` buffer, drop-on-full** (`const sendBufferSize = 16`). Slow connection is dropped (it reconnects + resyncs via REST — index §2). Buffer size → config later (index "Optimization notes").
6. **`sync.Once` teardown**: unregister from all rooms → close `send` → `socket.Close(reason)`. Idempotent; safe under concurrent evict + disconnect.

## Concurrency invariants (enforce + comment in code)
- **The lock is never held during a socket write.** `Broadcast` RLocks only to copy the target `send` channels / do non-blocking sends; the `writePump` (outside any lock) drains `send` → `socket.Write`.
- **One writer per socket** = the `writePump`, satisfying coder/websocket's single-writer rule (index §2).
- `RLock` for `Broadcast`/`ActiveUsers`; `Lock` for `Register`/`Join`/`Leave`/`Evict*`/teardown.
- `onPresence` is invoked **outside** the lock (it will broadcast, which re-locks).

---

## Test checklist (`go test -race ./internal/realtime/...`)

Fake socket (`fakeSocket`: records writes, programmable `Read` to block or return EOF, records `Close`).

- [ ] **Join / Leave** — single conn joins a room, `ActiveUsers` == {user}; leaves, room empties.
- [ ] **Edge-triggered presence** — `onPresence` fires `PresenceJoined` on 0→1 only; a **second tab** of the same user joining fires **nothing**; first tab leaving fires nothing; last tab leaving fires `PresenceLeft` (1→0).
- [ ] **ACTIVE_USERS snapshot** — `Join` returns distinct user_ids incl. self; multi-tab user appears once.
- [ ] **Broadcast fan-out** — message reaches every conn in the room; conns in other rooms get nothing; sender is **included** (index §6).
- [ ] **Slow-consumer drop** — a conn whose `send` buffer is full (writePump blocked) is dropped on the next `Broadcast`; it's unregistered from all its rooms; other conns unaffected.
- [ ] **Teardown on disconnect** — `fakeSocket.Read` returns EOF ⇒ conn removed from **all** its rooms; 1→0 edges reported; `Close` called once; double-teardown is a no-op (`sync.Once`).
- [ ] **EvictUser** — all of the target user's conns in that board are torn down (`reason` recorded on `Close`); other users and the user's conns in *other* boards untouched.
- [ ] **EvictExcept** — every conn whose user ∉ `allowed` is evicted; allowed users stay.
- [ ] **EvictUserFromRooms** — the user is evicted from exactly the listed boards; their conns elsewhere stay.
- [ ] **Race sanity** — concurrent `Join`/`Leave`/`Broadcast`/`Evict*` across goroutines is clean under `-race`.

---

## Checkpoint (done-when)
Hub unit tests green under `-race`: join/leave, multi-tab edge-triggered presence, the eviction family, slow-consumer drop, teardown. No HTTP, no `coder/websocket` dependency yet. → unblocks **Part B** (`/ws` endpoint + lifecycle), which supplies the real `socket` adapter and the wire protocol.

---

## Code review — Part A (2026-07-29)

Two-axis `/code-review` (Standards + Spec, parallel sub-agents) of the `internal/realtime/` package against this doc + [index.md](./index.md). **Outcome: clean, faithful implementation — nothing blocks Part B.**

**Grounding:** `go build` ✔ · `go vet` clean ✔ · `gofmt` clean ✔ · **19/19 tests pass under `-race`** ✔

### Standards axis — 0 hard violations
- **Concurrency invariants all upheld** (the highest-value checks):
  - *Lock never held during a socket write* — `Broadcast` does only non-blocking `c.send <- msg` under `RLock`; the real `s.Write` is in `writePump` outside any lock; `unregisterConn` releases `h.mu` before `teardown`.
  - *No send-on-closed-channel panic window.* `unregisterConn` removes `c` from every room under `Lock` **before** `teardown` closes `send`; `Broadcast` only selects on `c.send` (under `RLock`) for conns still in the room — mutually exclusive, so no concurrent send can hit a closed channel. High confidence.
  - *Double-teardown safe* — both pumps → `unregisterConn` → `sync.Once` `teardown`; `unregisterConn` is itself idempotent (second pass finds `c.rooms` empty). `TestTeardown_DoubleTeardown_IsNoop` covers it.
- **TESTING.md conformance:** external `package realtime_test`, `t.Run` slices, `require`/`assert` split, hand-rolled `fakeSocket` appropriate for the unexported `socket` seam (not in the mockery config).
- **Dependency rule:** imports only `context`, `sync`, `uuid` — no HTTP/DB/framework; `Broadcaster` port correctly deferred to Part C.
- **Reviewed and deliberately kept** (not smells to fix): the duplicated evict "collect-under-lock → tear-down-after-unlock" shape (each site differs in RLock/Lock + predicate), `EvictUserFromRooms` looping `EvictUser` (a spec-required primitive Parts D/E call directly), and the threaded `reason string` (decision #3 — Part D's `ACCESS_REVOKED` adds no signature churn).

### Spec axis — 0 missing, 0 scope creep, 0 wrong
- Every "Part A builds" bullet present; all 10 test-checklist items map to concrete tests.
- No scope leaked from Parts B/C/D — no HTTP, no `coder/websocket`, no message parsing, no `Broadcaster` port, no `ACCESS_REVOKED` frame. `reason` is threaded but never rendered as a frame — exactly per decision #3.
- Two signature deltas from the "Key types & signatures" block, **both in-spirit, accepted:**
  - `Register(ctx, userID, s)` adds `ctx` — the `socket` seam needs a context to drive `Read`/`Write`; threading it from `Register` is the natural source.
  - `Join` returns unnamed `[]uuid.UUID` vs the doc's named `(activeUsers []uuid.UUID)` — documentation-only; behaviour identical.

### Action items
Both are **cosmetic polish — non-blocking.** Do them at the start of Part B (or skip the optional one); neither gates Part A approval.

- [ ] **Reword the `// re-acquire properly via unregisterConn` comment** (`hub.go`, in `EvictUser`). It reads like a stale TODO — `EvictUser` already released the `Lock`, so nothing is "re-acquired" there. State the intent instead, e.g. `// tear down outside the lock (unregisterConn re-locks per conn)`.
- [ ] *(Optional)* **Name the anonymous `struct{ boardID uuid.UUID; kind PresenceKind }`** in `unregisterConn` as a `presenceEdge` local type, to remove the verbatim repetition between the slice element type and its literal. Trivial; skip if not worth the churn.

**Verdict:** Part A approved. Both axes converge — a clean, faithful hub core. → proceed to **Part B**.
