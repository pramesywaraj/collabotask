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
| ⑤ | Post-Phase-1 integrity pass | ✅ 2026-10-26 | ADR-010/015/016, migration 000009 |
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

## ⑤ Post-Phase-1 integrity pass ✅ 2026-10-26

Order: harness ✅ → composite-FK ✅ → Transactor ✅ → deferred SQL tests ✅

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

### Deferred SQL tests ✅ 2026-10-26

[handoff](handoff/handoff-deferred-sql-tests.md) · [review findings](handoff/handoff-integrity-pass-findings.md)

Repo-layer tests deferred from earlier steps, written now that the harness exists. Branch `test/deferred-sql-tests`. **31 integration tests pass.** Almost entirely tests — the one production change is the default-column spacing fix below, which writing these tests surfaced. The suite was also run end to end for the first time here, which closes the third Transactor carry-over item.

**Default-column spacing fix ✅ 2026-11-02.** `CreateWithOwner` seeded a new board's three default columns from the `range` index — 0, 1, 2 — instead of `PositionStep` multiples, leaving neighbour gaps of 1 where ADR-004 intends 1000. It was the one seeding path never updated when ADR-004 landed; `create_column.go`, `create_card.go` and migration 000006 all space by STEP. Nothing user-visible and self-healing after one rebalance, but a client dropping at the midpoint between two columns would have hit the rebalance threshold on its second drag — so it was fixed before frontend work rather than after. `testutil.CreateTestColumn` moved with it: its hard-coded `Position: 1000` would have collided with the new `To Do`, so it now appends one STEP past the last default. Three column cases in `positioning_test.go` encoded the old spacing and were retargeted; the comfortable-gap case got *simpler*, since the defaults now provide a landable gap and its workaround fourth column was no longer needed. Full suite green 3/3 with `-race`: 632 tests.

`-race` on the *whole* suite was not clean when this was first written: `TestSetArchived` shared one mutable `*entity.Board` across `t.Parallel()` subtests while `set_archived.go:47` wrote to it — a pre-existing fault, unrelated to this branch, that failed 4 of 4 full runs. Each subtest now takes its own copy; verified 30/30 on the test and 8/8 across `internal/`. Found by writing these tests, not by them.

- **Atomicity variants** (`transactor_test.go`) — ADR-016's bar asked for more than the `CreateCard` shape. `TransferOwnership` now has a rollback test and a success test (the latter pins `fromUserID`, which the broadcast depends on), and break-glass `self_join_board` has both a rollback test and a join-twice test proving `CreateIfAbsent` reports `inserted=false` and writes no second activity row. This closes the last open carry-over from the Transactor review.
- **Position rebalance** (`positioning_test.go`) — `rebalanceIfNeeded` through both of its seams: cards partitioned by `column_id` and columns by `board_id`. Each gets all four cases: a sub-threshold gap rewrites the whole partition to `STEP` multiples in the order the move produced, the returned position is the one the DB actually holds (this was a real bug), a comfortable gap rewrites nothing, and an open end (`±Inf` neighbour) is not a packed one.
- **Visibility filter and `card_count`** (`board_list_test.go`) — the five visibility cases of `getUserBoardsInWorkspace`, plus the `COUNT(DISTINCT …)` guard built as one board holding 3 members *and* 2 cards at once, so a weakened `COUNT()` reports 6 and fails. A fixture with one member or one card would have passed either way.
- **`ErrBoardMemberNotFound`** (`board_member_repo_test.go`) — a 0-row remove and a 0-row ownership transfer both return the sentinel the use cases branch on, not a silent success.
- **Board unassign cascade** — already covered board- and workspace-level in `assignee_fk_test.go` from the composite-FK step; nothing to add.

**Where the proof stops:** the two atomicity tests rebuild the use-case logic by hand inside `WithinTransaction` (the established pattern in this file), so they prove the Transactor and the repositories roll back together but do **not** exercise the use-case wiring itself — delete the `if inserted` guard in `self_join_board.go` and `TestTransactor_SelfJoinTwiceLogsOnce` still passes.

Fixture de-duplication done at the same time: `CreateTestBoard` and `CreateTestCard` now delegate to the parameterised variants, so the seeded defaults live in exactly one place.

**Phase 1 backend is closed. Frontend implementation is next.**

---

## Frontend architecture ✅ settled, not built

[ADR-011](architecture/adr/adr-011-frontend-framework.md) · [ADR-012](architecture/adr/adr-012-frontend-state-management.md) · [ADR-013](architecture/adr/adr-013-public-zone-framework.md) · [frontend-architecture.md](architecture/frontend-architecture.md) · SRS §3.1 · memory `[[frontend-architecture-settled]]`

Zone split: **Vite SPA** for the app zone (`app.collabotask.com`), **Astro** SSG for the public zone (`collabotask.com`, **landing only** — ADR-013 moved login and register into the app zone to keep the CORS allowlist to a single origin). ADR-013 supersedes ADR-011's public-zone choice; Next.js is not used anywhere in the frontend.

State: **TanStack Query** for server/REST state, **Zustand** for WS push and UI state, with WebSocket patching the query cache (ADR-012).

Also: React Router v7; URL-based card modal; `_authenticated` layout route as the auth guard; feature-based folders; shadcn/ui + Tailwind; axios with a CSRF interceptor; WS connects once on app mount.

---

## UI foundations ✅ settled, not built

[ADR-017](architecture/adr/adr-017-ui-foundations.md) · supersedes the uncommitted `architecture/design-tokens.md` proposal

The visual baseline for the app zone, decided by grilling then by rendering. Settled: density (2-line card clamp, 8 cards at 1440×900 — verified 612px of cards against 704px of visible column), **contained column panels** as the board structure, plum `#6B4E9E` as the primary (6.54:1 on white, so one value serves fill, ring and link text), warm stone neutrals, an eight-hue identity palette whose roles swap between modes, **no shadow on a resting card** (shadow means "temporarily lifted", nothing else), a 4/6/8 radius ladder, Poppins at a 14px default across five steps, dark-ready-but-not-dark-shipped, and a pointer-only ≥1024px scope. **The scope clause was superseded on 2026-11-17 by [ADR-019](architecture/adr/adr-019-responsive-touch-scope.md)**; everything else in ADR-017 stands.

Chosen by looking, not arguing: three structural directions were rendered side by side in both themes before anything was locked. Direction C ("bare") was eliminated by measurement — a card with no surface step needs a 3:1 border to identify itself, so the least-drawn direction ends up with the heaviest edges.

Seven defects found and fixed along the way, four of them in the superseded draft: `--drag-over` at 1.02:1 and `--drop-placeholder` at 1.37:1 on the P0 drag path, `--success` shipped unaudited at 2.54:1, `--column-header` identical to `--column`, plus three found while deriving the dark values — `--muted-foreground` fails on a column panel at stone-500, `--primary-foreground` inverts between modes, and every identity rail fails in dark unless the palette swaps roles.

ADR-017 deliberately spends **no** channel on interaction state. The channels left unspent — opacity, motion, dashed border, cursor, icon, column-panel tint — are listed in the ADR as ADR-018's budget. ADR-019 later struck **cursor** from that budget at coarse pointer (a touchscreen has no cursor) and added three constraints of its own.

---

## Responsive scope and touch interaction ✅ settled, not built

[ADR-019](architecture/adr/adr-019-responsive-touch-scope.md) · supersedes [ADR-017](architecture/adr/adr-017-ui-foundations.md) §8 only

Phones and tablets move **into** Phase 1, with card dragging working on touch. Settled by grilling: two adaptation axes (width drives layout, pointer drives density, touch sensors always registered), **two board layouts with the crossover at 640px** — chosen over 768px because an iPad mini in portrait is 744px and would otherwise render one ~670px column — a **horizontal snap strip** below that with the next column peeking by exactly 44px, **long-press pickup** at 200–250ms with no grip handle, **adjacent-only drag reach** on the strip, a **coarse-pointer density scale** (6 cards at 390×844; the smallest possible card is exactly 44px tall), and presence capped at 3 avatars plus an overflow chip.

ADR-017 §8's third argument did not survive: *"a drag must cross columns the user cannot see"* assumed a fixed 280px column, which is a premise rather than a property of phones. Its other two arguments — that a second density scale is real cost, and that a phone board deserves design — both stand and are paid rather than dodged.

A **"Move to…" picker ships at every breakpoint**, not just on phones: WCAG 2.2 SC 2.5.7 requires a no-drag alternative to every drag, so the picker is what makes drag-and-drop AA-conformant at *desktop* width too. It is also the keyboard move path, and it makes a cross-board phone move one activity row instead of three.

Two findings surfaced that ADR-018 inherits: the **cursor** channel is void on touch, and the **read-only banner must be persistent** — a non-joined member on a workspace-visible board gets a fully visible, entirely uneditable board (`CheckMutateAccess`), and on touch the banner is the only signal distinguishing "no permission" from "broken".

---

## State vocabulary ✅ settled, not built

[ADR-018](architecture/adr/adr-018-state-vocabulary.md) · spends the budget [ADR-017](architecture/adr/adr-017-ui-foundations.md) reserved, under [ADR-019](architecture/adr/adr-019-responsive-touch-scope.md)'s four constraints · supersedes nothing · PRD §5.4 SC 2.5.8 gap **closed**

What every interaction state looks like, with real values. Written after ADR-019 despite the lower number, which ADR-017 had reserved by filename.

**Enumeration came first, deliberately.** ADR-017 records that the draft it replaced failed by assigning every channel *before* counting what needed one, then running out. So ADR-018 lists **39 states** — grouped by cause, including the six that nearly got cut (connection health, other people's changes) and session-expiry, which was missed on the first pass — and assigns nothing until the list is closed. Every one is covered at the end, and no channel carries two meanings.

Of ADR-017's seven reserved channels: **four confirmed** (opacity, dashed border, column tint, lock icon), **one overturned**, **one demoted**, **one struck**.

**Motion was overturned by measurement.** ADR-017 reserved it for "server-corrected". That state needs **eleven consecutive drops into the same shrinking gap** to fire (`PositionStep` 1000 against a 0.5 threshold) and **preserves visible order when it does** — the rebalance rewrites positions with `ROW_NUMBER() OVER (ORDER BY position ASC, id ASC)`, so nothing on screen moves. Zero render sites, the same shape of defect as the draft's unaudited `--success`. Re-pointed to **"a teammate did this"**, which fires constantly and is PRD §4.2's literal promise. It needs no actor field — which the backend does not have — because movement is triggered by a position actually changing, and the broadcast for your own move is a no-op.

**Settled by looking, not arguing**, as ADR-017 did for board structure: two throwaway pages, one comparing four movement policies on live boards and one showing every state value on its real surfaces.

**Three defects found that arithmetic alone would have shipped**, all of which look fine on screen — plus one value that passes and still should not be used:

- **`opacity: 50%` for "saving" fails body text at 3.44:1.** The whole card is one compositing layer, so its text blends toward the *column*, not the card fill.
- **The meta row fails long before the title** — muted grey under a fade clears 4.5:1 only at 80%, where the card looks untouched. Resolved by promoting the meta row to `--foreground` while in flight: the card is de-emphasised as a whole, so its internal hierarchy has nothing left to express.
- **ADR-017's own `--input` is too thin for the drop placeholder** at 3.22:1 — it clears the 3:1 minimum by 0.22, having been chosen for a white field. The placeholder only ever appears inside a column you are already over, so it sits on the drag-over tint, never the bare panel; ADR-017 stated the 3:1 requirement without noticing it had already darkened that surface.
- **shadcn's `hover:bg-accent` dissolves a card into its column**, since ADR-017 sets `--accent` and `--board-column` to the same `#F5F5F4`. Hover moved to the card edge, walking three consecutive stone steps.

A fifth collision was headed off: shadcn's `disabled:opacity-50` would have put two meanings on the opacity channel on day one. **Nothing in the app is "disabled"** — the three situations split into *not allowed* (remove the control plus a persistent banner), *in flight* (plum-subtle, says "wait" not "no"), and *form not satisfied* (grey but still a button). **Dimming means one thing: your change hasn't landed** — and the object is genuinely inert, backed by a `timeout: 10000` on the axios client, which currently has none at all, so a hung request would otherwise freeze a card until reload.

Values, all computed: drag-over `#D6D3D1` at **1.37:1** against the panel (the draft's was 1.02:1) on the column **body only**, since the header's rule vanishes on a tint · placeholder `#57534E` 2px dashed at **5.12:1** · in-flight 60% · 24×24 hit areas at mouse density with both 20px avatars declared **not targets**, which closes PRD §5.4's first gap.

**Motion specifics**, which ADR-017 deferred here by name: 120ms chrome / 200ms board objects / 260ms containers, arriving decelerates and leaving accelerates. A rejected move is **not** animated — easing it back would be indistinguishable from a teammate putting it there. Nothing animates unless visible before *and* after, because an unbounded column can move a card thousands of pixels. Under `prefers-reduced-motion` movement becomes a **cross-fade, not nothing**: removing it outright would take the only explanation of a changing board away from the people least able to absorb one.

**Icons: Phosphor v2, Regular at 16px** — which renders at exactly 1.00px, matching ADR-017's 1px borders (Lucide would give 1.33px). **Regular is furniture, Fill is the channel**: a filled icon means a state, a Regular one labels an action. Costs roughly ten import rewrites, since shadcn's components ship Lucide.

**Two channels have no Phase-1 render site, both recorded rather than fixed:** the archive icon (archived boards are filtered out of the list and 404 on every access path, so Unarchive is unreachable too) and `--done`/emerald (cards have no status field; a column titled "Done" is free text).

Figma covers both densities without doubling: the state values turn out to be **density-independent** — every one is a colour, an opacity, a stroke weight or a duration — so a second two-mode collection carries only ADR-019 §7's seven numbers, states are component variants rather than frames, and the frame budget is the fine set plus exactly one 390×844 phone board, since layout is a breakpoint and cannot be produced by flipping a mode.

---

## Design file, Stages 1 to 4 ✅ drawn 2026-11-23

[ADR-020](architecture/adr/adr-020-design-tool-penpot.md) (the tool) · [ADR-021](architecture/adr/adr-021-design-inventory-corrections.md) to [ADR-039](architecture/adr/adr-039-drag-and-phone-board.md) · [design-file-inventory.md](architecture/design-file-inventory.md) §7 is the record · Penpot `Collabotask App`, which is **not in git**

The design file holds the foundations board, **19 parts** and **all fifteen whole-screen boards**, each built from the parts, with the extra boards some states need (2.3 and 2.4's states, 2.12's two wordings, 2.14 as a card drag and a column drag, 2.15's drag board). **210 tokens** in five sets and two theme groups, so every board flips Light ↔ Dark and Fine ↔ Coarse by switching a theme, never by redrawing.

**Question 4 was asked before every stage: "which values here have no token, and which choices were never made?"** It fired at seventeen stages out of seventeen, and each answer became an ADR before anything was drawn. With the inventory's corrections (ADR-021) and the side menu (ADR-034), that is nineteen ADRs, ADR-021 to ADR-039, most decided by the user on a rendered page with every option shown in both themes.

Things drawing found that reading would not have:
- **The skeleton was invisible** in Light: `--muted` is the column's own colour (ADR-029).
- **No one-sided rule had ever drawn.** Penpot saves per-side stroke widths and draws none of them, so 54 rules became inset shadows (ADR-036).
- **Late dates were told apart by colour alone**, failing WCAG 1.4.1, so they now carry words (ADR-037).
- **The phone's card count**, the one assumed number in ADR-019 §7, was checked by eye: 5 whole cards and 44px of the sixth (ADR-039).

Every board passes the binding check (inventory §7.5), which finds typed-in values that should be tokens, apart from the card `⋯` glyph, whose size has no token. Next is Stage 5, the ~18 overlays.

---

## Design file, Stage 5 ✅ drawn 2026-11-24

[ADR-040](architecture/adr/adr-040-menus-and-the-confirmation.md) to [ADR-044](architecture/adr/adr-044-the-sidebar-sliding-over-and-the-toast.md) · [design-file-inventory.md](architecture/design-file-inventory.md) §7 is the record · Penpot `Collabotask App`, which is **not in git**

Every overlay is drawn, in five groups, each settling its kind of container once: the menus and the confirmation (ADR-040), the card drawer's pickers (ADR-041), the forms (ADR-042), the lists (ADR-043), and the sidebar sliding over and the toast (ADR-044). A new page, `Overlays`, holds **47 screens**, each with one overlay open over a copy of a whole-screen board. **222 tokens** and **34 parts**: twelve token entries and fifteen parts came with Stage 5.

**Question 4 was asked before every group**, and each answer became an ADR before anything was drawn. Most choices were made by the user on a rendered page, with every option drawn in both themes; across the five groups the user took most recommendations and overrode a handful, each recorded in its ADR with its cost.

Things drawing and the backend code found that reading the design docs would not have:
- **A phone could have stacked a full-height sheet on a full-height sheet** (*Move card to…* from the drawer). Every picker now opens in place (ADR-041).
- **Nothing refused removing the workspace owner**, and **leaving a board told open lists nothing**. Two backend items, the SRS edited first (ADR-043).
- **Archiving evicts no one**, so the archive toast's *Undo* can bring teammates back from the "gone" screen, best effort (ADR-044).
- **A workspace role change has no live message**, so a demoted admin learns of it from a 403; the board now reloads after a refused change (ADR-044).

Every board passes the binding check (inventory §7.5) apart from the card `⋯` glyph and the calendar's today, whose sizes have no token. Next is Stage 6, the final check.

**Regrouped 2026-11-25, before Stage 6** ([ADR-045](architecture/adr/adr-045-design-file-pages-by-area.md)): a screen on `Page 1` and what opens over it on `Overlays` were drawn far apart, so a developer couldn't tell which flow to build. The 73 boards now sit on three pages by area, **Account**, **Workspaces** and **Board**, one row per screen: at rest, its states, what opens over it, its phone versions last. 31 boards moved by copy and paste, each checked against its original before the original was deleted.

---

## Design file, Stage 6 ✅ checked 2026-11-27

[ADR-046](architecture/adr/adr-046-stage-6-final-check.md) · [design-file-inventory.md](architecture/design-file-inventory.md) §7.5–§7.6 are the record · Penpot `Collabotask App`, which is **not in git**

The final check ran 2026-11-25 to 2026-11-27 over all five pages and all four Theme × Density combinations. The binding check is **clean on every board**: about 26,000 shapes, and every row it returned is on the expected list. The flips looked at every page for text at the wrong size, text that doesn't fit and text too faint to read, and at the screens the ADRs say change most. **No token (222), no part (34), no board added.** The file is left in Light + Fine.

Four findings, all taken as recommended:
- **Initials on a dimmed card, and the contents of a dragged column's place, fall under 4.5:1** (2.38 to 4.44 in Light). ADR-018 measured only the title and meta row; `Card` gained the avatar later. **Recorded as an exception**, since the object is inert and a full-strength copy or the settled card replaces it (ADR-046 §1).
- **Twelve documentation boards had no fill**, so their captions were 1.49:1 in Dark. Filled with `color.background`.
- **The Count chip board draws three chips clipped.** Five repairs did not hold; the real screens are correct. Left as a known fault, and noted in the inventory.
- **Four Penpot repairs**: the `List` head's padding mode, the `Form`'s clipped focus ring, one blank email chip, and the two `Footer` boards.

**Not looked at by eye**, and said so in the ADR: 17 of Board's 30 desktop screens in Dark + Fine, and every part but `Button` in Dark + Fine. The data checks covered all of it.

Things the check found that reading would not have:
- **A page that isn't open reads every text as 0 × 0**, so a fit check passes everything on it, and an export draws stale text (*Add cc* for *Add column*).
- **A flex board in *simple* padding mode ignores a different left padding.** The `List` head's 12 never applied.
- **A board that clips its content cuts a focus ring**, which is a shadow.

**This closes the design file.** What remains is ① scaffolding and the five backend items, in one pass before the frontend reaches the screens that need them.

**Illustrations resized 2026-11-27** ([ADR-047](architecture/adr/adr-047-illustration-size-192.md)), after Stage 6 closed: the user tried to change `size.illustration` from 128 to 256 and only the wrapper grew, because the shapes inside were pinned to the top-left corner. The user looked at 1.5× and 2× on test copies of *2.4 Workspace detail — empty* and **chose 192**. All eight looks now scale with their box, the line stays 2px, the nine screens that show a drawing grew by 64px and still fit, and the part's own board was re-laid out. The code needs `--illustration-size: 192px` and a stroke that does not scale. No token added (222), no part (34).
