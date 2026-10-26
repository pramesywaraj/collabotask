# Build Log

Completed work in build order, newest step last. Each entry is an **index over** the ADR and handoff that hold the full story — it records what changed and which decisions are load-bearing, not the blow-by-blow.

The roadmap is spec §9.1. The pointer to current work is `## Now` in `CLAUDE.md`.

> Git history begins 2026-07-21; steps completed before then have no date.

| Step | What | Status | Records |
| --- | --- | --- | --- |
| — | Unit-test coverage (275/275) | ✅ | ADR-003 |
| ① | P0 fixes | ✅ | — |
| ② | Fractional NUMERIC positioning | ✅ | ADR-004, migration 000006 |
| ③ | REST feature completion (5 sub-steps) | ✅ | ADR-005/006/007, migrations 000007–000008 |
| ③.5 | Auth: Bearer → httpOnly cookie | ✅ 2026-07-27 | ADR-008 |
| ④ | WebSocket + participation broadcast (A–F) | ✅ 2026-08-22 | ADR-009/014 |
| ⑤ | Post-Phase-1 integrity pass | in progress (Transactor merged, PR #14) | ADR-010/015/016, migration 000009 |
| — | Frontend architecture | ✅ settled, not built | ADR-011/012/013 |

---

## ② Fractional NUMERIC positioning ✅

Spec §3.2 · [ADR-004](architecture/adr/adr-004-fractional-positioning.md) · migration 000006

Single `UPDATE` per move; rebalance lives in the repository layer, atomic with the move that triggers it. `domain.PositionStep` / `domain.PositionRebalanceThreshold` are the single source of truth.

Repo-layer rebalance tests deferred to ⑤ (no DB harness at the time).

---

## ③ REST feature completion ✅

Five sub-steps, in order.

### Board `visibility`

Spec §2.2–2.4 · [ADR-005](architecture/adr/adr-005-board-visibility-access.md) · migration 000007

Three-method access checker (metadata / view / mutate) plus break-glass. 404-hide vs 403-reveal; idempotent self-join; thin roster. `created_by` removed from access logic entirely.

### Workspace ops

Spec §4.2 UC-06b/c/d · [handoff](handoff/handoff-workspace-ops.md)

Promote/demote with guards (owner → 403, last-admin → 409, self-demotion allowed, idempotent no-op); leave workspace; delete workspace (owner-only). UC-06 `RemoveMember` retrofitted with the same cascade — `RemoveWithParticipationCascade` clears `board_members` and unassigns cards in one transaction.

### Board ownership transfer

Spec §4.3 UC-12e · [ADR-006](architecture/adr/adr-006-board-ownership-transfer.md)

Atomic demote-by-role + promote-by-id in one transaction. `created_by` proxy removed from `canAdministerBoard()` and the `leave_board` guard — both are role-based now. `ErrTransferTargetNotBoardMember` → 400. Idempotent, orphan-safe.

### Assignee validation + board-scoped unassign cascade

Spec §2.8, UC-14/UC-16, UC-10/UC-12d · [handoff](handoff/handoff-assignee-validation-cascade.md)

Write-time gate: `boardMemberRepo.IsUserExists` → `ErrAssigneeNotBoardMember` (400) on `create_card` / `update_card`, collapsing the old 404 path.

**Load-bearing:** `update_card` validates a *newly-set* assignee only — never a stale one. Re-validating on every update would reject edits to cards whose assignee left the board.

### Activities table + writes

Spec §4.6 · [ADR-007](architecture/adr/adr-007-activity-logging-writes.md) · migration 000008

Verb × entity vocabulary; log-only-on-state-change; snapshot-title metadata; board-scoped, including per-affected-board rows on workspace cascade (`AffectedBoardIDs`). 17 call sites across card / column / board / workspace.

Originally best-effort after-commit (errors swallowed). **Superseded in ⑤ by ADR-016** — writes are now atomic with the mutation.

---

## ③.5 Auth: Bearer → httpOnly cookie ✅ 2026-07-27

[ADR-008](architecture/adr/adr-008-auth-httponly-cookie.md) · SRS §3.5, §4.1 · [handoff](handoff/handoff-auth-httponly-cookie.md) · memory `[[auth-cookie-migration-prereq]]`

Eight decisions, hard cutover. JWT-in-cookie, stateless. Topology A — cross-origin sibling subdomains `app.` ↔ `api.`.

`__Host-token; HttpOnly; Secure; SameSite=Lax; Path=/; Max-Age=<cfg.JWTExpiration>`

**Load-bearing consequences:**

- The cookie is **host-only** ⇒ **no authenticated SSR in Phase 1**. This is the constraint that drives the whole frontend framework split (ADR-011).
- **CSRF** = a required `X-CSRF-Protection` header on *all* mutations, including login / register / logout. The point is to force a CORS preflight, which closes the sibling-subdomain vector.
- CORS uses an explicit allowlist. `*` plus credentials fails fast at startup.

Session and refresh tokens, single-origin reverse proxy, and email verification are all Phase 2. Register auto-login is kept provisionally until email verification supersedes it.

---

## ④ WebSocket + participation broadcast ✅ 2026-08-22

[ADR-009](architecture/adr/adr-009-websocket-realtime-layer.md) · [ADR-014](architecture/adr/adr-014-presence-frame-enrichment.md) · SRS §4.5, §5 · [handoff index](handoff/websocket-participation-broadcast/index.md) · memory `[[websocket-participation-broadcast-design]]`

Built in six parts, A → F. Final state: 596 tests, `-race` clean.

**Shape:** `Broadcaster` port in the **usecase** layer, mirroring activities. RWMutex hub with per-connection read/write pumps. REST in, broadcast out.

**Decisions that corrected the SRS sketch:**

- **Multi-connection edge-triggered presence** — the SRS assumed one connection per user. It is not.
- **Continuous access enforcement** — instant eviction plus an `ACCESS_REVOKED` event (UC-19b), rather than checking only at handshake.

**Part E** (workspace cascade fan-out) added `GetBoardIDsByWorkspace`, `AffectedCard.BoardID` routing and `EvictReasonRemovedFromWorkspace`, and closed a watcher leak for workspace-visible boards via `EvictUserFromRooms`.

**Part F** (presence frame enrichment, ADR-014) — thin `user_ids`-only frames couldn't render workspace-visible watchers or break-glass admins, who are connected but *not* in the board roster. `ACTIVE_USERS` / `USER_JOINED` are now enriched server-side to `{id, name, avatar_url}` in the `ws_handler`; the hub stays byte-dumb. `USER_LEFT` stays thin. The `ACCESS_REVOKED.reason` enum was pinned here.

**Known limitation:** auth is checked at handshake only. The Phase-2 session model closes this (evict-on-revoke).

A post-③.5 drift recheck (2026-07-29) confirmed no reversal: the handshake reuses the cookie-based `middleware.Auth`, shares the CORS allowlist with `OriginPatterns` (scheme → host translation), and the GET handshake is a deliberate CSRF-header exemption defended by `OriginPatterns`.

---

## ⑤ Post-Phase-1 integrity pass — in progress

Order: harness ✅ → composite-FK ✅ → Transactor ✅ → **deferred SQL tests ◀ next**

### Integration test harness ✅ 2026-08-23

[ADR-015](architecture/adr/adr-015-integration-test-harness.md)

testcontainers-go (`postgres:16-alpine`, auto-start on `go test`); `TruncateAll` in `t.Cleanup`; one container per package via `TestMain`; golang-migrate driven programmatically; a `Fixtures` struct for composable seed helpers. Tests live in `tests/integration/`.

Tracer bullet: `TestBoardRepository_CreateWithOwner`.

### Composite-FK assignee invariant ✅ 2026-08-24, hardened 2026-08-27

[ADR-010](architecture/adr/adr-010-composite-fk-assignee-invariant.md) · migration 000009 · [handoff](handoff/handoff-composite-fk-assignee-invariant.md)

`cards(board_id, assigned_to) → board_members … ON DELETE SET NULL (assigned_to)` is now a DB-level invariant.

**Load-bearing:** this is **defense in depth only**. Both app-layer guards stay. The FK *cannot* replace the application cascade, because the FK nulls the rows silently and the broadcast list would come back empty.

**Ordering gotcha:** both cascades unassign-and-collect **before** deleting the membership row. Reverse that and `ON DELETE SET NULL` empties the list first.

Migration 000009 adds `cards.board_id`, backfills it from `columns.board_id`, nulls pre-existing violators, then adds the composite FK. `23503` is mapped to `ErrAssigneeNotBoardMember` (400), scoped by constraint name via `isAssigneeFKViolation`.

### Transactor / atomic activity writes ✅ built 2026-08-31, merged (PR #14)

[ADR-016](architecture/adr/adr-016-transactor-atomic-activity-writes.md) · [handoff](handoff/handoff-transactor-atomic-activity-writes.md) — discharges ADR-007's deferred Option C

Two phases. **Phase 1, behavior-preserving:** all 52 pool-direct sites → `r.exec(ctx)`; 9 hand-rolled transaction methods now join an ambient transaction; the `Transactor` port is wired into every usecase via DI. **Phase 2, atomicity flip:** all 17 `WriteActivity` call sites → `WithinTransaction` + `LogActivity`; `WriteActivity` deleted.

Activity writes are no longer best-effort — a failed log now rolls back its mutation. The "resilience" unit tests became "propagation" tests across card / column / board / workspace, plus 3 integration tests (rollback, a `Log`-honors-tx footgun guard, and an end-to-end forced-`Log`-failure via FK violation).

598 unit tests green, `-race` clean. The integration tests **compile** but the testcontainers suite had not been run when this merged.

Two-axis review (2026-08-31): faithful to ADR-016, 0 correctness defects. Merged with three review items still open — tracked as carry-over in `CLAUDE.md` `## Now`:

- 3 stale comments naming the deleted `WriteActivity` (`activity_repository.go:10` is actively wrong).
- Integration test #3 covers only the `CreateCard` shape; ADR-016's bar also names `TransferOwnership` and break-glass `self_join_board`.
- The integration suite itself has not been executed.

### Deferred SQL tests — next

Repo-layer tests deferred from earlier steps, now that the harness exists:

- board unassign cascade
- position rebalance (from ②)
- visibility filter (from ③)
- `card_count` (from ③)

---

## Frontend architecture ✅ settled, not built

[ADR-011](architecture/adr/adr-011-frontend-framework.md) · [ADR-012](architecture/adr/adr-012-frontend-state-management.md) · [ADR-013](architecture/adr/adr-013-public-zone-framework.md) · [frontend-architecture.md](architecture/frontend-architecture.md) · SRS §3.1 · memory `[[frontend-architecture-settled]]`

Zone split: **Vite SPA** for the app zone (`app.collabotask.com`), **Astro** SSG for the public zone (`collabotask.com`, **landing only** — ADR-013 moved login and register into the app zone to keep the CORS allowlist to a single origin). ADR-013 supersedes ADR-011's public-zone choice; Next.js is not used anywhere in the frontend.

State: **TanStack Query** for server/REST state, **Zustand** for WS push and UI state, with WebSocket patching the query cache (ADR-012).

Also: React Router v7; URL-based card modal; `_authenticated` layout route as the auth guard; feature-based folders; shadcn/ui + Tailwind; axios with a CSRF interceptor; WS connects once on app mount.
