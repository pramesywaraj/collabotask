# ADR-009: WebSocket Realtime Layer — Broadcast Architecture, Presence & Continuous Access Enforcement

- **Status:** Accepted (design; implementation pending — build step ④, after ③.5 auth-cookie)
- **Date:** 2026-07-21
- **Scope:** How Phase-1 realtime is delivered (SRS §2.7, §4.5, §5, §9.1 step ④; US-18/19/20). Covers: where a board mutation's broadcast originates (layer + port); the event payload shape; the in-memory hub (registry + concurrency); the `/ws` endpoint (auth, server write-timeout, origin); the presence model (multi-connection, edge-triggered); **continuous access enforcement** (instant eviction on access loss) and the new `ACCESS_REVOKED` event; and the participation-cascade broadcast fan-out. **Depends on ADR-008** (auth migrates to an httpOnly cookie *before* this lands). Build plan: `docs/handoff/websocket-participation-broadcast/index.md`.

## Context

Step ④ is the last Phase-1 build step: stand up the WebSocket layer (UC-18/19) and wire realtime broadcasts onto the already-correct REST mutations, including the `CARD_UPDATED` broadcast for the participation-unassign cascade completed in step ③. The mutation surface is complete and tested (438 tests); activities already fire a **best-effort, after-commit, usecase-orchestrated** side effect (`common.WriteActivity`, ADR-007 Option A) at exactly the point a broadcast must fire. Realtime is defined as *"a layer on every state-changing board action"* (SRS §2.7), and the server is the source of truth (clients are optimistic and reconcile via REST).

Several SRS §4.5/§5 sketches are illustrative rather than binding and needed hardening before implementation — notably the presence registry (`map[board_id]→map[user_id]→connection`, which is single-connection-per-user and mishandles multi-tab and reconnect), and the token-in-query handshake (superseded by the ADR-008 cookie decision). This ADR records the decisions that make the layer robust.

## Options Considered

### Where the broadcast originates — usecase-layer port vs delivery-layer vs hybrid
- **Delivery/handler layer (rejected)** — keeps usecases free of realtime, but the handler lacks the data the events need: `changed_fields` (computed in the usecase), and the cascade's `AffectedBoardIDs` / `[]AffectedCard`. It would require re-plumbing that through output DTOs purely to fan out, and would split "what broadcasts" away from the state-change predicate that decides *whether* to broadcast.
- **Hybrid — cascade in usecase, simple events in handler (rejected)** — two places to look for "what broadcasts"; inconsistent.
- **Usecase layer via a `Broadcaster` port (chosen)** — mirrors `WriteActivity` exactly (best-effort, after-commit, side-by-side with the activity write); the usecase already holds all required data and the fire-on-state-change predicate. Clean-architecture is preserved: the usecase depends on a **consumer-side interface** (like repositories / `BoardAccessChecker`), never on `coder/websocket`. Broadcasting is a domain concern (SRS §2.7), not a transport detail; the transport lives behind the port.

### Event payload shape — typed structs vs `map[string]any`
- **`map[string]any` (rejected)** — maximally consistent with the activities `metadata` blob, but §5.2 is a **fixed, closed contract with a live consumer** (the frontend). Stringly-typed keys are only caught by tests; a typo is a silent runtime bug.
- **Typed event structs (chosen)** — one struct per §5.2 event implementing a small `Event` interface; the **adapter** marshals the `{type, payload}` envelope, so the usecase stays transport-agnostic. Compile-time safety on the field set; the structs double as living documentation of the contract. Deliberate divergence from activities, justified by "fixed contract + live consumer" (vs schemaless audit blob).

### Presence registry — single connection per user vs a set of connections
- **`map[user_id]→connection` (SRS §4.5 literal — rejected as the implementation)** — a second tab **overwrites** the first (orphaning it — it stops receiving broadcasts silently), and closing **either** tab emits `USER_LEFT` while the user is still present in the other. The UC-18 auto-reconnect makes it worse: the new connection briefly coexists with the not-yet-reaped old one → spurious `USER_LEFT`/`USER_JOINED` churn, requiring fragile pointer-identity cleanup.
- **`map[user_id]→{connID→conn}` set, edge-triggered (chosen)** — `USER_JOINED` fires only on the user's **0→1** edge, `USER_LEFT` only on **1→0**; `ACTIVE_USERS` lists distinct user_ids; broadcasts fan out to all of a user's tabs. Multi-tab and reconnect are correct by construction. The SRS `map` is read as illustrative shorthand; §5.2's stated intent ("everyone sees me join/leave") is served *better* by this.

### Hub concurrency — RWMutex + per-conn pumps vs actor
- **Actor / channel-owned rooms (rejected)** — lock-free on the map, but the actor goroutine must never block, so it *still* needs the per-connection `send` channel + writePump for socket writes — you pay that machinery anyway, plus a command channel and a potential single-goroutine bottleneck. More parts, no gain at single-instance scale.
- **`sync.RWMutex` hub + per-connection read/write pumps (chosen)** — tiny critical sections (map lookups/appends), RLock broadcasts run in parallel, and the **per-connection buffered `send` channel with one `writePump`** is the sole socket writer (satisfies coder/websocket's single-writer rule) and decouples fan-out from slow clients. The lock is never held during a socket write. Robustness policies: bounded buffer → **drop the slow connection** (it reconnects + resyncs); **server-driven ping/pong** (server timers aren't background-throttled; the browser auto-PONGs) to reap half-open sockets; **`sync.Once` teardown**.

### `/ws` handshake — auth, server write-timeout, origin
- **Token in query / subprotocol (rejected)** — needed only because a browser `WebSocket` can't set headers. ADR-008 moves the JWT into an **httpOnly cookie**, which rides the handshake automatically; the WS `Accept` reads the cookie → `ValidateToken` → reject `401` before upgrade. No token in URLs.
- **Server `WriteTimeout`** — the shared `http.Server` applies a 30s write deadline that would kill a long-lived idle WS. Global-zero (strips REST slowloris protection) and a second listener/port (extra infra) are rejected in favour of **`http.ResponseController` clearing this connection's deadlines only**; REST keeps its 30s, and WS liveness is ping/pong.
- **Origin** — cookie auth forbids the `*` wildcard (credentialed requests). The explicit `CORS_ALLOWED_ORIGINS` is shared by REST CORS **and** coder/websocket's `OriginPatterns` (CSWSH protection); never `InsecureSkipVerify`.

### Access on a live connection — validate-at-join only vs continuous enforcement
- **Validate-at-join only (rejected)** — an access loss mid-session (removal, or a WORKSPACE→PRIVATE flip) leaves the connection watching a board it may no longer see until its socket happens to drop. On a PRIVATE board that is a real leak, and relying on the client to self-evict is a weak security boundary.
- **Continuous enforcement / instant eviction (chosen)** — the server enforces access continuously, not just at join. Access loss triggers immediate eviction via one of three hub primitives: `EvictUser(board, user)` (board removal/leave), `EvictExcept(board, allowed)` (WORKSPACE→PRIVATE flip — evict everyone not a member/admin), `EvictUserFromRooms(user, boardIDs)` (workspace removal/leave — **also closes the workspace-visible-watcher leak**: a user watching a WORKSPACE board without membership isn't in `AffectedBoardIDs`). `JOIN_BOARD` re-validates with **`CheckViewAccess`** (room visibility == kanban visibility). `PRIVATE→WORKSPACE` needs no eviction (access only widens).

### Eviction signal — dedicated event vs reuse
- **Reuse `MEMBER_REMOVED` / `BOARD_UPDATED` (rejected)** — forces the client to *infer* "am I still allowed?" from side-channel events; inference for a security-adjacent behavior is the fragility we avoid elsewhere.
- **Dedicated `ACCESS_REVOKED{board_id, reason}` (chosen; new to §5.2)** — one unambiguous "leave this board view" signal, uniform across removal and flip. Sent only on **involuntary** loss; **voluntary** leaves (UC-10/UC-06c) evict **silently** (`USER_LEFT` covers the room). **Evict-first ordering** in cascades keeps the room broadcasts (`MEMBER_REMOVED`, per-card `CARD_UPDATED`) clean of the departing user.

### Mutation transport — REST-in/broadcast-out vs WS-in; sender echo
- **REST-in, server-broadcasts-out (chosen; SRS §5.1)** — reuse the complete, tested REST handlers; the WS layer adds one `Broadcast(...)` next to each `WriteActivity(...)`. WS client→server messages are only `JOIN_BOARD` / `LEAVE_BOARD`.
- **Include the sender in broadcasts (chosen)** — the only clean option: a REST mutation carries no WS connection identity to exclude, and the sender's *other* tabs must update. Clients dedupe against their authoritative REST response.

## Decision

- **Broadcasts originate in the usecase layer via a `Broadcaster` port**, best-effort and after commit (swallow-on-failure), co-located with `WriteActivity`. Port surface: `Broadcast`, `EvictUser`, `EvictExcept`, `EvictUserFromRooms`. The concrete in-memory **hub** lives in a new realtime package and is swappable to Redis pub/sub behind the port (SRS §10).
- **Usecases emit typed event structs**; the adapter marshals the `{type, payload}` §5.2 envelope.
- **Hub registry `map[boardID]→map[userID]→{connID→conn}`**, guarded by `sync.RWMutex`; each connection has a buffered `send` channel + single `writePump` (sole writer) and a `readPump`. Robustness: bounded-buffer drop, server-driven ping/pong, `sync.Once` teardown.
- **Presence is edge-triggered** (`USER_JOINED` on 0→1, `USER_LEFT` on 1→0; `ACTIVE_USERS` = distinct connected user_ids to the joiner). This **supersedes** the SRS §4.5 single-connection registry sketch.
- **`/ws` endpoint:** authenticate from the httpOnly JWT cookie before upgrade; clear this connection's read/write deadlines via `http.ResponseController` (REST keeps 30s); share the explicit `CORS_ALLOWED_ORIGINS` with `OriginPatterns`.
- **Continuous access enforcement:** instant eviction on access loss via the three primitives; `JOIN_BOARD` re-validates via `CheckViewAccess`; **new `ACCESS_REVOKED` event** on involuntary loss, silent evict on voluntary; evict-first cascade ordering.
- **REST-in / server-broadcasts-out**, sender included; `WorkspaceUseCase` gains the `Broadcaster` for its UC-06/06c board cascades (consuming the existing `AffectedBoardIDs` + `[]AffectedCard`).

### Out of scope (deferred, recorded so they aren't lost)
- **Auth-cookie migration itself** → **ADR-008 / build step ③.5** (its own grilling session; this ADR only *assumes* the cookie exists).
- **Multi-instance broadcast (Redis pub/sub)** → SRS §10; behind the port, not Phase 1.
- **Growth-driven hardening** → tuning constants to config (ping interval, `send` buffer, backoff); WS observability (connection/room gauges, dropped-connection counter, broadcast latency); WS abuse guards (`JOIN_BOARD` rate limit, max message size). Recorded in the handoff's optimization notes.
- **Proactive disconnect-on-hidden-tab; compression/binary transport** → premature; only if usage/payloads demand.
- **Transactor/atomic side effects** → unchanged from ADR-007 (integrity pass ⑤); broadcasts stay best-effort like activity writes.

## Consequences

**Positive**
- Realtime rides the **existing, tested** REST surface — the WS layer adds broadcast taps, not business logic. Activities + broadcasts sit side-by-side at one lifecycle point.
- The usecase stays framework-free and unit-testable against a mock `Broadcaster`; the hub is unit-testable in-memory under `-race`. No DB harness required.
- Presence and reconnect are correct under multi-tab/device by construction; the SRS registry bug never ships.
- Access is enforced continuously — losing participation stops board visibility immediately, honouring "the server is the source of truth"; the workspace-visible-watcher leak is closed.

**Negative / notes**
- **The usecase layer takes one more dependency** (the `Broadcaster` port) on every mutating usecase, and `WorkspaceUseCase` — a workspace-layer usecase — now reaches board rooms for its cascades. Accepted: it mirrors the existing `activityRepo` injection, and the cascade genuinely has board-scoped consequences watchers must see.
- **This ADR overrides two SRS §4.5/§5 sketches** (single-connection presence registry; query-param token). SRS is updated to match; the sketches are read as illustrative, not binding.
- **`ACCESS_REVOKED` extends the §5.2 contract** — a conscious addition the frontend must handle.
- **The single-instance in-memory hub is the scaling ceiling** (SRS §10). The port abstraction is the only thing that must be respected so the Redis swap stays a drop-in.
- **Depends on ADR-008 landing first.** If auth-cookie is deferred, the handshake would fall back to a query-param/subprotocol token and the origin list could keep `*` — both explicitly rejected here; re-open this ADR if that sequencing changes.
