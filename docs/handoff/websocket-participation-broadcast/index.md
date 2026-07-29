# Handoff — WebSocket + Participation Broadcast (step ④): design index

**Date:** 2026-07-21
**Repo:** `collabotask` (backend: `collabotask-backend/`)
**Status:** **Planning handoff** from a `/grill-with-docs` session. Design settled and **approved by the product owner**. **Nothing built yet.**

> This is the **durable index** for step ④. It holds the settled architecture, the decisions, and the A–E implementation map with checkpoints. The **per-part detail files** (`part-a-*.md` … `part-e-*.md`) are written **just-in-time**, each right before that part is built — so they reflect the real code state *after* the ③.5 auth-cookie work lands, not today's. Don't pre-write them.

## Sequencing (important — read first)
```
NOW      docs capture (this handoff, ADR-009, SRS §4.5/§5, CLAUDE.md, memory)
③.5      Auth: Bearer → httpOnly cookie   ← grilled + built BEFORE ④ (its own session; ADR-008)
④        WebSocket + participation broadcast (this design), built in parts A→B→C→D→E
```
- **④ depends on ③.5.** The WS handshake authenticates from the **JWT cookie** (no `?token=`, no subprotocol). Cookie auth also forces **explicit CORS origins** (no `*`) shared by REST + WS.
- **Drift recheck after ③.5: DONE (2026-07-29, `/grill-with-docs`).** See **[§ Post-③.5 drift recheck — resolved](#post-35-drift-recheck--resolved-2026-07-29)** below. Four points resolved (middleware reuse, origin-format translation, CSRF-GET invariant, handshake-only-auth limitation); **no design reversal** — mechanics refined and one limitation made explicit. §3 below already reflects the corrected mechanics.
- Full rationale for the WS decisions: **[ADR-009](../../architecture/adr/adr-009-websocket-realtime-layer.md)**. Auth-cookie decision + why-before-WS: **ADR-008** (pending its grilling). SRS **§4.5 / §5** hold the settled realtime contract.

---

## Post-③.5 drift recheck — resolved (2026-07-29)

`/grill-with-docs` session comparing this design against what ③.5 (auth httpOnly cookie) **actually built** ([`middleware/auth.go`](../../../collabotask-backend/internal/delivery/http/middleware/auth.go), [`csrf.go`](../../../collabotask-backend/internal/delivery/http/middleware/csrf.go), [`cors.go`](../../../collabotask-backend/internal/delivery/http/middleware/cors.go), [`handler/auth_handler.go`](../../../collabotask-backend/internal/delivery/http/handler/auth_handler.go), [`config.go`](../../../collabotask-backend/internal/config/config.go)). **Headline: the big-ticket item was NOT drift** — this design pre-committed to cookie auth on 2026-07-21 and ③.5 built exactly that transport. Drift was confined to finer mechanics that ③.5's implementation now pins down. Four points, all resolved; §3 above is updated to match.

| # | Drift | Resolution | Lands in |
|---|---|---|---|
| **1** | Design read as "the `/ws` endpoint reads the cookie itself" — a Bearer-era hangover (the old middleware read `Authorization`, unusable for WS). | **Reuse `middleware.Auth`.** ③.5 made it cookie-based, so `/ws` mounts under `middleware.Auth(&cfg.Auth)`; handler reads `GetUserID(c)`. Abort-before-handler = reject-before-upgrade. Zero new auth surface; one validation path shared with REST (so the Phase-2 session swap changes one place). | Part B |
| **2** | "Feed `CORS_ALLOWED_ORIGINS` into `OriginPatterns`" glossed a **format mismatch**: CORS stores full origins *with scheme* (for reflection); `OriginPatterns` matches *host only*. Feeding full origins → every cross-origin handshake silently rejected. | **Derive host-patterns** via `url.Parse(origin).Host` over `CORS.AllowedOrigins`. Single source of truth (no second WS origin var). **Fail-fast at startup** on any unparseable origin (mirrors the existing `*`+credentials guard in `config.Validate()`). | Part B + config validation |
| **3** | Global CSRF passes the GET handshake — **correct, but an undocumented invariant** a future dev could "harden" into a WS outage. | **Mechanism correct; document it.** Handshake CSWSH defense = `OriginPatterns` + `SameSite=Lax`, *not* the header (browser WS can't send `X-CSRF-Protection`). Add a comment to `csrf.go`; the "never gate GET in CSRF" invariant is now in §3. | Part B (comment only) |
| **4** | Design has continuous **authorization** enforcement (evict on membership loss) but only handshake-time **authentication**. Logout / token-expiry do **not** evict a live socket (③.5 logout is stateless — can't revoke a JWT). | **Accept handshake-only auth** as a conscious Phase-1 limitation; consistent with ADR-008 §7 (stateless logout). Residual exposure is low — a still-valid member who logged out keeps an already-open socket until it drops; anyone who *loses access* is still evicted. **Closed in Phase 2** (see ledger). Not a ④ prerequisite. | Note only (see ledger A) |

### Note-for-later ledger (do not miss)

**A. Phase-2 session/revocation model — the real fix for drift #4.**
- Add a server-side auth anchor: opaque session-id + `sessions` table, **or** ADR-008's likely hybrid (short-lived access JWT + revocable refresh token). Gives the server a real "off switch": logout → revoke session (dead immediately); "log out all devices" → revoke all.
- **WS closure is additive, not rework:** bind each socket to its session at handshake; a revocation event evicts bound sockets through **④'s existing `EvictUser` / `EvictUserFromRooms` / `ACCESS_REVOKED` primitives**. Closes **both** the logout and token-expiry gaps at once. Choosing handshake-only now does **not** paint us into a corner — Phase 2 wires a new *trigger* into the same eviction surface.
- **Triggers to start it:** (1) real need for instant revocation / "log out all devices" (compromised account, password-change kills sessions); (2) wanting short-lived access tokens (pulls in refresh); (3) "remember me" UX; (4) device/session management UI. **Bundle with refresh tokens** — same infrastructure. Already recorded as deferred in ADR-008 §1 + out-of-scope.

**B. Part-B implementation reminders (drift #1–3), so Part B doesn't re-derive them:**
- Mount `/ws` under `middleware.Auth(&cfg.Auth)`; handler uses `GetUserID(c)`; **no inline cookie read.**
- Build `OriginPatterns` from `CORS.AllowedOrigins` via `url.Parse().Host`; **fail-fast** on unparseable origin.
- Add the CSRF-GET invariant comment to [`csrf.go`](../../../collabotask-backend/internal/delivery/http/middleware/csrf.go); handshake CSWSH defense lives in `OriginPatterns` + `SameSite=Lax`.

---

## The settled design (condensed)

### 1. Layering & artifacts
- Broadcast originates in the **usecase layer** via an injected **`Broadcaster` port** (consumer-side interface, like `common.BoardAccessChecker`). Concrete **hub** in a new realtime package; swappable to Redis pub/sub later behind the port.
- Usecases emit **typed event structs** (`realtime.CardMoved{…}`, one per §5.2 event, implementing a small `Event` interface). The **adapter** marshals to the `{type, payload}` wire envelope — the usecase never learns the transport format. (Deliberate divergence from the activities `map[string]any`: §5.2 is a fixed, live-consumer contract, so typed structs buy compile-time safety + double as documentation.)
- **Usecase-facing port surface:** `Broadcast(boardID, Event)`, `EvictUser(boardID, userID, reason)`, `EvictExcept(boardID, allowed[], reason)`, `EvictUserFromRooms(userID, boardIDs[], reason)`. Presence/room join-leave is driven by the **WS delivery handler** against the hub directly, not through this port.

### 2. The hub
- **Registry:** `map[boardID]map[userID]map[connID]*Conn` — a **set of connections per user** (multi-tab / multi-device correct).
- **Concurrency:** `sync.RWMutex` guards the map (RLock broadcast, Lock join/leave/evict). Each `Conn` has a **buffered `send` channel + one `writePump`** (the sole socket writer — satisfies coder/websocket's single-writer rule) and a **`readPump`** (handles `JOIN_BOARD` / `LEAVE_BOARD`). The lock is **never held during a socket write**.
- **Three robustness policies:** bounded `send` buffer → **drop the slow connection** (it reconnects + resyncs via REST); **server-driven ping/pong** liveness (reaps half-open sockets); **`sync.Once` teardown** (unregister from all rooms → close `send` → close socket).

### 3. The `/ws` endpoint
- **Auth:** **reuse the existing `middleware.Auth`** (③.5 made it cookie-based). Mount `/ws` under `middleware.Auth(&cfg.Auth)`; on invalid/absent cookie it aborts `401` *before the handler runs* = before upgrade. The handler reads `middleware.GetUserID(c)` then upgrades — **no inline cookie read, no second auth path.** (Corrected 2026-07-29: the original "endpoint reads the cookie itself" phrasing was a Bearer-era hangover from when the middleware read `Authorization` and couldn't be reused for WS. See drift recheck below.)
- **Timeout:** clear this connection's read/write deadlines via `http.ResponseController` so the server's 30s `WriteTimeout` doesn't kill a long-lived WS; **REST keeps its 30s**. WS liveness = ping/pong, not a fixed deadline.
- **Origin:** feed the explicit `CORS_ALLOWED_ORIGINS` into coder/websocket's `OriginPatterns` (CSWSH protection) — **but translate first:** `OriginPatterns` matches the Origin header's **host** (`app.collabotask.com`, no scheme), while `CORS_ALLOWED_ORIGINS` holds full origins (`https://app.collabotask.com`, needed for CORS reflection). Derive patterns via `url.Parse(origin).Host` per configured origin; **fail-fast at startup** on any unparseable origin. Single source of truth (no second WS-only origin var). Never `InsecureSkipVerify`.
- **CSRF on the handshake:** the handshake is a **GET**, so the global `middleware.CSRF` passes it through (GET is RFC-7231-safe) — this is *required*, because the browser `WebSocket` API cannot send `X-CSRF-Protection`. The handshake's cross-site defense is **`OriginPatterns` + `SameSite=Lax`**, not the header. **Invariant: never extend CSRF to gate GET, or every WS handshake breaks.**
- **Topology:** one `/ws` per tab; a connection may hold a **set** of rooms; on disconnect it's removed from all.

### 4. Presence (edge-triggered)
- `JOIN_BOARD` → re-validate via **`CheckViewAccess`** (room visibility == kanban visibility; break-glass falls out). Add to the set; **`USER_JOINED` only on the user's 0→1 edge**; send **`ACTIVE_USERS`** (snapshot of distinct user_ids incl. self) to the **joiner only**.
- `LEAVE_BOARD` / disconnect → remove from set; **`USER_LEFT` only on 1→0**. Reconnect churn is silent by construction.
- `ACTIVE_USERS` is built from the **in-memory room** (who is connected now), never the DB roster.

### 5. Continuous access enforcement
The server enforces access **continuously, not just at join.** Any loss of access → **instant eviction**:

| Trigger | Primitive | Signal to evicted |
|---|---|---|
| Removed from board (UC-12d) | `EvictUser(board, X)` | `ACCESS_REVOKED` (involuntary) |
| Removed from workspace (UC-06) | `EvictUserFromRooms(X, workspace's boards)` | `ACCESS_REVOKED` |
| WORKSPACE→PRIVATE flip (UC-12b) | `EvictExcept(board, members+admins)` | `ACCESS_REVOKED` |
| Leaves board (UC-10) | `EvictUser(board, X)` | silent (`USER_LEFT` covers room) |
| Leaves workspace (UC-06c) | `EvictUserFromRooms(X, workspace's boards)` | silent |
| PRIVATE→WORKSPACE flip | — | — (access only widens) |

- **New event `ACCESS_REVOKED{board_id, reason}`** — a conscious addition to SRS §5.2; the client's single unambiguous "leave this board view" signal for **involuntary** loss. Voluntary leaves evict silently.
- **`EvictUserFromRooms`** closes the workspace-visible-watcher leak (X watching a WORKSPACE board they weren't a member of, then removed from the workspace — not in `AffectedBoardIDs`).

### 6. Broadcast wiring
- **REST-in, server-broadcasts-out** (SRS §5.1): existing REST handlers stay the write path; add one `Broadcast(...)` next to each `WriteActivity(...)`. The §5.2 table is the checklist (~16 events). WS client→server messages are only `JOIN_BOARD` / `LEAVE_BOARD`.
- **Best-effort, after commit, swallow** — identical contract to `WriteActivity`; a failed broadcast never fails/rolls back the mutation.
- **Include the sender** — the only clean option: a REST mutation has no WS connection identity to exclude, and multi-tab means the sender's *other* tabs must update. Clients dedupe against their REST response.

### 7. Participation cascade fan-out
- `WorkspaceUseCase` **gains the `Broadcaster`** (its UC-06/06c cascades touch board rooms, even though plain workspace actions never broadcast — SRS §2.7).
- Consumes the existing `AffectedBoardIDs` + `[]AffectedCard`. **Per affected board, evict-first ordering:** `EvictUser/EvictUserFromRooms` → `MEMBER_REMOVED` → `CARD_UPDATED{assigned_to:null}` per cleared card. Evicting first keeps the room broadcasts clean of the departing user.

### 8. Testing
- **Hub:** real unit tests, pure in-memory, under `go test -race` — join/leave/edge-triggered presence, multi-tab, the eviction family, slow-consumer drop, teardown.
- **Broadcast/evict emission:** asserted via a **mock `Broadcaster`** in usecase tests, mirroring the activity-contract tests. No DB harness required.

---

## Implementation map — build in parts, review at each checkpoint

| Part | Builds | Done when (checkpoint) | Depends on |
|---|---|---|---|
| **A — Hub core** | Registry + RWMutex, `Conn` with send-channel + read/write pumps, edge-triggered presence, the 3 eviction primitives — pure in-memory, no HTTP | **Built + reviewed ✅ 2026-07-29** — 19/19 tests green under `-race`; two-axis `/code-review` clean (0 hard-standards, 0 spec gaps); 2 cosmetic action items → [part-a](./part-a-hub-core.md#code-review--part-a-2026-07-29) | ③.5 auth exists |
| **B — `/ws` endpoint + lifecycle** | HTTP handler → hub; cookie auth, origin, `ResponseController` timeout fix; `JOIN_BOARD`/`LEAVE_BOARD`; `ACTIVE_USERS`/`USER_JOINED`/`USER_LEFT`; ping/pong | A real client connects, joins a board, sees live presence, survives reconnect | A |
| **C — Broadcaster port + mutation broadcasts** | Port + typed events + envelope adapter; inject into card/column/board usecases; `Broadcast` next to each `WriteActivity` (~16 events); best-effort | Card/column/board mutations appear live in a joined room; usecase tests assert emission via mock | A, B |
| **D — Continuous enforcement** | `ACCESS_REVOKED`; `EvictUser` on board-remove (UC-12d), `EvictExcept` on →PRIVATE flip (UC-12b), evict on board-leave (UC-10) | Board-scoped access loss evicts the right connections instantly | A, C |
| **E — Participation cascade fan-out** | `WorkspaceUseCase` gains `Broadcaster`; UC-06/06c per-affected-board fan-out (evict → `MEMBER_REMOVED` → `CARD_UPDATED`) + `EvictUserFromRooms` leak closure | Removing/leaving a workspace member fans out + evicts across all affected boards | A, C, D |

Each part is a natural PR-sized checkpoint. Write its `part-x-*.md` detail file (files to touch, signatures, test checklist) **when you start that part**, after re-reading the then-current code.

---

## Optimization notes — problems most likely to appear as usage / the app grows
Not Phase-1 work. Recorded so they aren't forgotten; each is feasible behind the current design.

- **Tuning constants → config** (ping interval, `send` buffer size, reconnect backoff). The #1 source of "flaky WebSocket" reports: buffer too small drops healthy clients; wrong ping interval causes false disconnects or lets intermediaries reap idle connections. Surface to config and tune from observed behavior.
- **Observability** (active-connection & room gauges, dropped-connection counter, broadcast latency). The first time users "randomly disconnect" in production, you are blind without these.
- **WS-level abuse guards** (cap `JOIN_BOARD` churn per connection, max inbound message size). One misbehaving/malicious client can otherwise spam the hub; no guard exists today.
- **Redis pub/sub multi-instance** (SRS §10). The hard ceiling: the moment a second instance runs, in-memory rooms don't share. Swap the `Broadcaster` implementation behind the port — no usecase changes.

Deferred as premature (note only): proactive disconnect-on-hidden-tab (only at large idle-tab scale); `permessage-deflate`/binary transport (only if payloads grow).

---

## References
- **[ADR-009](../../architecture/adr/adr-009-websocket-realtime-layer.md)** — WebSocket realtime layer decisions + rationale.
- **ADR-008** — auth-cookie migration (③.5), written after its grilling session.
- **SRS §4.5 / §5** — the settled realtime + message contract (incl. `ACCESS_REVOKED`, eviction rules, multi-connection presence).
- Memory: `[[websocket-participation-broadcast-design]]`, `[[auth-cookie-migration-prereq]]`, `[[activities-logging-best-effort-then-atomic]]`.
