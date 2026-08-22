# ADR-014: Presence Frame Enrichment — Rich Presence Payloads

- **Status:** Accepted (design; supersedes the presence-payload decision in [ADR-009](./adr-009-websocket-realtime-layer.md) only — the edge-triggered multi-connection model, hub, eviction, and `Broadcaster` port are unchanged)
- **Date:** 2026-08-22
- **Scope:** The wire shape of the three presence frames — `ACTIVE_USERS`, `USER_JOINED`, `USER_LEFT` (SRS §5.2/§5.3, §4.5 UC-19). Where the profile fields are populated (layer). Does **not** touch mutation frames, the hub registry, presence edge-triggering, or continuous enforcement — all of ADR-009 stands except the presence-payload bullet. **Builds on ADR-009.**

## Context

ADR-009 chose thin presence frames: `ACTIVE_USERS` = *"distinct connected user_ids"* and `USER_JOINED`/`USER_LEFT` carrying only a `user_id`, *"built from the in-memory room, never the DB roster"* — deliberately keeping presence a pure, DB-free, in-memory operation.

The ④ two-axis `/code-review` (base `dda8215`) surfaced a wire-contract gap the thin design can't cover. **Presence includes users who are not in the board roster the client fetched.** Room access is gated by `CheckViewAccess`, and two populations have view access *without* board membership:

- **Workspace-visible boards** — any workspace member can open and watch without joining.
- **Private boards** — workspace admins can **break-glass** view without joining.

Neither appears in `board_members`, so neither is in the roster returned by `GET .../kanban`. Because presence is explicitly *"who's connected, not who's a member"* (ADR-009), a `USER_JOINED`/`ACTIVE_USERS` can legitimately reference a userID the frontend has **no local profile for** — precisely the *"someone else is also looking at this board"* case presence exists to surface. A thin frame leaves those avatars blank.

SRS §5.2's original (spec-phase) shape was in fact rich (`user{id,name,avatar_url}`); ADR-009 thinned it for the no-DB rationale, and the §5.2 table was never reconciled — leaving the SRS internally contradictory (§4.5 said `user_ids`, §5.2 said `users:[{…}]`). This ADR resolves that in favour of enrichment.

## Options Considered

### A. Rich frames — carry the profile on the frame (chosen)
Enrich `ACTIVE_USERS`/`USER_JOINED` with `{id,name,avatar_url}` server-side; the client renders directly.
- The thin-frame *performance* rationale doesn't apply to presence: join/leave is **low-frequency** (a user opening/closing a board view), unlike the high-frequency `CARD_*` path where thin frames earn their keep.
- The join path **already hits the DB** — `handleJoin` calls `CheckViewAccess` before it can send `ACTIVE_USERS`, so enriching the snapshot there piggybacks on an existing round-trip rather than reintroducing DB access to a pure path.
- Self-contained frames spare the not-yet-built frontend a whole profile-cache + resolve-race subsystem, and are **consistent with `MEMBER_ADDED`**, which already carries full `user{…}`.

### B. Thin frames + a batch user-lookup endpoint (rejected)
Keep `user_ids`; frontend resolves unknown IDs via a new `GET /users?ids=…` with client-side caching.
- Closes the gap, but pushes a caching layer + a snapshot-vs-profile resolve race onto the client, and requires new REST infrastructure — more moving parts than the problem warrants at Phase-1 scale.

### C. Thin frames + resolve from the already-fetched roster (rejected)
Free, no backend work — but does **not** cover workspace-visible watchers or break-glass admins (the whole point). Blank avatars for exactly the interesting users.

## Decision

- **Presence frames carry the profile:**
  - `ACTIVE_USERS` → `{ board_id, users:[{ id, name, avatar_url }] }`
  - `USER_JOINED` → `{ board_id, user:{ id, name, avatar_url } }`
  - `USER_LEFT` → `{ board_id, user_id }` — **stays thin** (removal keys on ID; no profile needed to drop an avatar).
- **`timestamp` / `joined_at` are dropped** from §5.2's original guess — the hub doesn't track connection time and the client doesn't need it.
- **Enrichment lives in the delivery layer (`ws_handler`), not the hub.** The hub keeps shipping opaque `[]byte`; the handler maps `user_id → {id,name,avatar_url}` via the existing `UserRepository.GetByIds`:
  - `ACTIVE_USERS` — enriched in `handleJoin`, piggybacking the `CheckViewAccess` round-trip.
  - `USER_JOINED` — the hub's `onPresence` callback is hub-wide (fires for *any* user's 0→1 edge), so it can't reuse a per-request profile; it does **one bounded, best-effort `GetByIds`** per 0→1 edge. Presence edge-detection stays in the hub (the Part-B "don't split the presence seam" decision is unchanged).
- **Best-effort throughout:** a profile lookup failure logs-and-skips enrichment; it never crashes a pump or blocks the presence broadcast. Consistent with the swallow-on-error contract of `WriteActivity`/`Broadcast`.

This **supersedes only** ADR-009's bullet *"Presence is edge-triggered (… `ACTIVE_USERS` = distinct connected user_ids to the joiner)"* insofar as the **payload** goes; the edge-triggered multi-connection model itself is untouched.

## Consequences

**Positive**
- Closes the watcher/admin profile gap — presence renders for every connected user, including non-members, which is exactly what presence is for.
- Simplest correct frontend contract: render straight from the frame, no profile cache / batch endpoint / resolve race. A real win for the greenfield frontend.
- Reconciles the SRS self-contradiction (§4.5 vs §5.2) and aligns presence with the existing `MEMBER_ADDED` precedent.

**Negative / notes**
- **`USER_JOINED` adds one bounded DB read per 0→1 join edge** (in `onPresence`). Low-frequency and best-effort, but a real cost — stated honestly (it does *not* piggyback the way `ACTIVE_USERS` does).
- **`WSHandler` gains a `UserRepository` dependency** (wired via `ProvideWSHandler`). The hub stays DB-free; only the delivery layer learns about users.
- **Overrides ADR-009's presence-payload decision** and updates SRS §4.5/§5.2/§5.3. ADR-009 is left unedited (project convention); this ADR is the superseding record.
- Not applicable to multi-instance scale-out any more than ADR-009 was — the enrichment is a local repo read behind the same handler.
