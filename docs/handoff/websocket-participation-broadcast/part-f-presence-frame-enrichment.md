# Part F — Presence Frame Enrichment (post-④ refinement)

**Date:** 2026-08-22
**Status:** **Planning handoff — nothing built yet.** Contract docs done (ADR-014, SRS §5.2/§5.3/§4.5); code pending.
**Design source:** [ADR-014](../../architecture/adr/adr-014-presence-frame-enrichment.md) (decision + rationale), SRS §5.2/§5.3/§4.5 (the wire contract). Born from the **④ two-axis `/code-review`** (base `dda8215`) — Spec-axis finding #5 (presence frames diverged from §5.2 + the SRS contradicted itself).

> **This does not reopen ④'s completion.** Parts A–E stand as built. This is a small, self-contained refinement of the presence payload that the ④ review surfaced. It reopens only the Part-B presence seam (frame shapes + the `handleJoin`/`onPresence` build sites), not the hub, eviction, or the `Broadcaster` port.

## Why

Presence includes users who are **not in the board roster the client fetched**: workspace-visible boards let any workspace member watch without joining, and private boards let workspace admins **break-glass** view without joining. Neither is in `board_members`, so a thin `USER_JOINED`/`ACTIVE_USERS` carrying only a UUID references a user the frontend has **no profile for** — blank avatars for exactly the "someone else is watching this board" case presence exists to show. Fix: **enrich the frames server-side** with `{id,name,avatar_url}`. Full options (rich vs. batch-endpoint vs. roster-only) in ADR-014.

## Scope

**In:** rich `ACTIVE_USERS` + `USER_JOINED`; enrichment in the WS delivery handler; wire structs; Wire rewiring; tests.
**Out:** the batch `GET /users?ids=` endpoint (ADR-014 option B, rejected); any hub change (hub stays byte-dumb); presence `joined_at`/`timestamp` (dropped); `USER_LEFT` shape (**stays thin**). The `ACCESS_REVOKED.reason` enum (#6) is a **doc-only** pin — already applied in SRS §5.2; no code change (the values already live in `common.EvictReason`).

## Locked decisions — do not re-litigate (all from ADR-014)

1. **Rich = `{id,name,avatar_url}`**; `USER_LEFT` stays `{board_id, user_id}` (removal keys on ID — no profile needed).
2. **Drop `timestamp`/`joined_at`** — the hub doesn't track connect-time; the client doesn't need it.
3. **Enrich in the delivery layer (`ws_handler`), not the hub.** The hub keeps shipping opaque `[]byte`; only the handler learns about users.
4. **`ACTIVE_USERS` enrichment piggybacks** `handleJoin`'s existing `CheckViewAccess` DB round-trip. **`USER_JOINED` adds one bounded, best-effort `GetByIds`** per 0→1 edge in `onPresence` (it's a hub-wide callback — can't reuse a per-request profile). Honest added cost; low-frequency.
5. **Best-effort throughout:** a lookup failure **logs + skips enrichment** (swallow, like `WriteActivity`/`Broadcast`) — never crash a pump, never block the presence broadcast.
6. **Reuse `UserRepository.GetByIds(ctx, ids) (map[uuid.UUID]*entity.User, error)`** — already on the interface + implemented (`WHERE id = ANY($1::uuid[])`). No new repo method.
7. **Keep presence edge-detection in the hub.** Do **not** move it into the handler to "fetch once" — the Part-B "don't split the presence seam" decision stands; only add the lookup.

## Files to touch

| File | Change |
|---|---|
| [`realtime/message.go`](../../../collabotask-backend/internal/realtime/message.go) | Add `PresenceUser{ID,Name,AvatarURL}` (json `id`/`name`/`avatar_url`). `ActiveUsersFrame.Users []PresenceUser` (drop `UserIDs`). Split the shared `UserPresenceFrame`: a rich `UserJoinedFrame{Type,BoardID,User PresenceUser}` and a thin left-frame `{Type,BoardID,UserID}`. |
| [`delivery/http/handler/ws_handler.go`](../../../collabotask-backend/internal/delivery/http/handler/ws_handler.go) | Add `userRepo repository.UserRepository` field + ctor param. `handleJoin`: after `hub.Join`, `GetByIds(ctx, snapshot)` → map to `[]PresenceUser` → send rich `ACTIVE_USERS`. `onPresence`: on `PresenceJoined`, bounded-ctx `GetByIds([userID])` → rich `USER_JOINED`; on `PresenceLeft`, thin frame (no lookup). |
| [`injection/providers.go`](../../../collabotask-backend/internal/injection/providers.go) | `ProvideWSHandler` gains `userRepo` (already provided by `ProvideUserRepository`). |
| [`injection/wire.go`](../../../collabotask-backend/internal/injection/wire.go) / `wire_gen.go` | `go generate ./...` (regen Wire). |
| [`delivery/http/handler/ws_handler_test.go`](../../../collabotask-backend/internal/delivery/http/handler/ws_handler_test.go) | Inject a mock `UserRepository`; update presence assertions to the rich shapes; add the new cases below. |

### Signature sketch
```go
func NewWSHandler(hub *realtime.Hub, access common.BoardAccessChecker,
    userRepo repository.UserRepository, originPatterns []string) *WSHandler

// onPresence (join branch): bounded, best-effort
ctx, cancel := context.WithTimeout(context.Background(), dbJoinCheckTimeout)
defer cancel()
users, err := h.userRepo.GetByIds(ctx, []uuid.UUID{userID})
// err or missing → log + send thin/skip; never panic
```

## Test checklist (`go test -race`)

- [ ] `ACTIVE_USERS` carries `users:[{id,name,avatar_url}]` for a multi-user room snapshot (incl. self).
- [ ] `USER_JOINED` carries `user:{id,name,avatar_url}`.
- [ ] `USER_LEFT` still carries **only** `user_id` (regression guard — no profile lookup on leave).
- [ ] **Watcher-not-in-roster** — a connected user who is *not* a board member still gets an enriched frame (the whole point of this change).
- [ ] **`GetByIds` returns error** → frame is still sent best-effort (enrichment skipped/degraded), no panic, a swallow-log is emitted.
- [ ] **`GetByIds` returns a partial map** (an id missing) → that entry is handled gracefully (skipped, not a nil-deref).
- [ ] Existing presence + eviction tests stay green; `-race` clean; `gofmt -l` clean.

## Build status & definition of done

- **Docs:** ADR-014 + SRS §5.2/§5.3/§4.5 ✅ 2026-08-22. `ACCESS_REVOKED.reason` enum pinned (#6) ✅.
- **Code:** ⏳ not started.
- **Done when:** rich frames shipped per the checklist, `go build`/`go vet`/`go test -race ./...` green, `gofmt -l` clean, then CLAUDE.md `## Now` + memory `[[websocket-participation-broadcast-design]]` updated to note the enrichment.
