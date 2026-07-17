# ADR-007: Activity Logging — Best-Effort Writes, Row Shape & Deferred Atomicity

- **Status:** Accepted
- **Date:** 2026-07-17
- **Scope:** How Phase-1 board mutations write to the `activities` table (SRS §7, §2.7, §2.8, §4.6, build-order step ③ final sub-step). Covers: the **write model** (best-effort after-commit vs transactional) and its deferral of a Unit-of-Work/Transactor to the post-Phase-1 integrity pass ⑤; the **firing rule** (log only on state change); the **vocabulary** (`action_type` × `entity_type`); `entity_id` **semantics**; the **metadata snapshot principle** + two audit-enrichment flags; and **scope** (board-scoped only, incl. per-board rows on workspace cascades). The activity-log **UI** is Phase 2 (UC-22); this ADR governs **writes only**. Discharges the activity-log deferrals parked by ADR-005 (self-join) and ADR-006 (`OWNERSHIP_TRANSFERRED`, [adr-006:63](adr-006-board-ownership-transfer.md)). The composite-FK assignee invariant is a **separate** deferred decision → **ADR-008 candidate** (not this ADR).

## Context

SRS §7 defines an `activities` table `{ id, board_id, user_id, action_type, entity_type, entity_id, metadata jsonb, created_at }` — **written in Phase 1, read (UC-22 feed) in Phase 2.** SRS §9.1 step ③ places the write here so step ④ "only wires broadcast calls and never revisits mutation code just to add a log write." Two prior ADRs explicitly parked their audit writes for this step: ADR-005 (self-join stays silent until activities land) and ADR-006 (transfer is silent until then, [adr-006:63](adr-006-board-ownership-transfer.md)). UC-09's break-glass admin-join is required to be **auditable** (§2.3).

The load-bearing constraint is architectural: **the codebase has no shared transaction abstraction.** Each repository owns a `*pgxpool.Pool`; multi-statement operations (`RemoveWithParticipationCascade`, `TransferOwnership`, `CreateMany`) each open a tx *inside a single repo method*, and single-statement mutations run bare on the pool. There is no way today for a usecase to compose "mutation + activity insert" in one transaction without either leaking the audit concern into every repo method or building a Unit-of-Work. There is also **no DB-backed test harness** in Phase 1 (memory `[[post-phase1-integration-tests]]`) — so any cross-repo atomicity guarantee cannot be verified until pass ⑤ stands one up.

Because nothing *reads* activities until Phase 2, the design must be shaped now to be (a) correct-enough as an audit trail, (b) self-renderable later without reopening mutation code, and (c) cheap to make atomic later — without paying for atomicity now.

## Options Considered

### Write model — best-effort after-commit vs transactional

- **(A) Best-effort, after commit, usecase-orchestrated (chosen).** After the mutation commits, the usecase calls `activityRepo.Log(ctx, …)` as a **separate synchronous** call; on error → `zerolog`-error and **swallow** (a logging failure must never fail or roll back the user's action). Zero architectural change. Accepts a rare, *logged* (non-silent) gap where a mutation commits but its activity row is lost.
- **(B) Fold the activity INSERT into each mutation's repo tx (rejected).** Truly atomic, but leaks the audit concern into the repository layer: every mutation method grows an activity parameter and `card`/`column` repos gain a dependency on the `activities` table. That is exactly the "architecturally-wrong shortcut" SRS §1.2 forbids, and pass ⑤ would have to unwind it.
- **(C) Unit-of-Work / Transactor abstraction (deferred to ⑤, not rejected).** The correct end-state: repos resolve an executor from context; the usecase composes mutation + `Log` in one `WithinTransaction(...)` closure. Clean **and** atomic, and reusable for the existing hand-rolled cascades. But it is invasive (≈50 pool-direct call sites across all 7 repos switch to a context-carried executor; 9 tx methods re-plumb) and its one guarantee — atomic rollback — is **only testable with the DB harness ⑤ builds.** Doing it now = full cost, unverifiable, and ⑤ reworks the same tx code (the composite-FK / ADR-008 removes several cascade tx methods a Transactor would have wrapped).

Async fire-and-forget for the write was also considered and rejected: the usecases receive `ctx.Request.Context()`, which is **canceled when the handler returns**, so a naive goroutine reusing it loses the row precisely because it's async; the correct version needs `context.WithoutCancel` + a panic `recover()` (a bare-goroutine panic crashes the process) + nondeterministic tests — all to move a single local INSERT (~sub-ms) off the request path. Not worth it at Phase-1 scale.

### Firing rule — every request vs only on state change

- **Log every mutation request (rejected)** — pollutes the Phase-2 feed with "updated the card" entries that changed nothing, and would fire spurious `break_glass` rows on idempotent re-joins.
- **Log only when state actually changed (chosen)** — generalizes the silence already designed into the idempotent self-join and no-op transfer. Falls out for free from the diff the update usecases must compute anyway (to populate `changed_fields`).

### Vocabulary — verb × entity split vs fused WS-event names

- **Fused names mirroring the §5.2 WS contract (rejected)** — `CARD_CREATED`, `MEMBER_REMOVED`, … welds the audit vocabulary to the wire vocabulary and makes `entity_type` redundant.
- **Verb × entity split (chosen)** — `entity_type ∈ {BOARD, COLUMN, CARD, MEMBER}` × `action_type ∈ {CREATED, UPDATED, DELETED, MOVED, ARCHIVED, UNARCHIVED, JOINED, LEFT, ADDED, REMOVED, OWNERSHIP_TRANSFERRED}`. It is *why* §7 has two columns, gives the Phase-2 feed a two-axis filter, and keeps the wire vocabulary a separate concern. **App-enforced** (a Go enum), no DB `CHECK` — consistent with "code is source of truth, DB is safety net," and avoids a migration per new action type.

### `entity_id` for membership events — target user vs board

- **Board id (rejected)** — `board_id` is already its own column; using it for `entity_id` too loses "which member."
- **The target user's id (chosen)** — `entity_id` = the member affected (joiner/leaver/added/removed); the **actor** stays in `user_id`. Self-join/leave → `user_id == entity_id`; invite/remove → they differ, giving the Phase-2 feed the "Alice removed Bob" pair directly. **Ownership transfer is a `BOARD` event** (`entity_id = board_id`, `from_user_id`/`to_user_id` in metadata), matching the board-centric WS event — not a `MEMBER` event.

### Metadata — snapshot vs resolve-at-read

- **Resolve every label at read time via joins (rejected)** — breaks for hard-deleted cards/columns (`activities` has **no FK to cards/columns**, so their rows survive with a dangling `entity_id`) and is *audit-unfaithful* (a later rename would rewrite history).
- **Snapshot deletable/renamable labels; ids-only for users (chosen)** — snapshot `card_title` / `column_title` (+ move's `from/to_column_title`) into metadata at write time; store only ids for users, who are never deleted or renamed in Phase 1 (resolve-at-read is safe and stable). `UPDATE` stores `changed_fields` **names only**, not old→new diffs (no reader yet — don't speculate on Phase-2 UI). This makes each row self-renderable and faithful to its moment.
- **Two audit-enrichment flags (chosen):** `break_glass: true` on `MEMBER/JOINED` when a non-member workspace admin opens a PRIVATE board (the one security-relevant event, §2.3 — otherwise indistinguishable from a normal join); `source ∈ {board, workspace}` on `MEMBER/LEFT|REMOVED` to distinguish a direct board action from a workspace-cascade departure.

### Scope of the workspace cascade — per-affected-board rows vs nothing

- **Nothing at the board layer (rejected)** — a board's feed would show a member silently vanishing when they were removed from / left the *workspace*.
- **One board-scoped row per affected board (chosen)** — from a board's perspective "X is no longer here" is a real membership change regardless of trigger. Requires the workspace-cascade repo methods to also `RETURNING board_id` (the boards the user was removed from) — the same "return it now, broadcast later" pattern the codebase already uses for `AffectedCard`. Per-card unassigns remain **unlogged** (§2.8).

## Decision

- **Write model = best-effort, synchronous, after-commit, in the usecase** (Option A). On `activityRepo.Log` error → log + swallow; never fail the mutation. `action_type`/`entity_type`/`metadata` assembly lives in the **usecase** layer so the eventual move into a Transactor closure (C) is a pure relocation.
- **Fire only on actual state change** — no-op update/move/archive, idempotent re-join, and no-op transfer write nothing (and therefore fire no `break_glass`).
- **Vocabulary** = verb × entity, **app-enforced** (no DB `CHECK`): `entity_type ∈ {BOARD, COLUMN, CARD, MEMBER}`, `action_type ∈ {CREATED, UPDATED, DELETED, MOVED, ARCHIVED, UNARCHIVED, JOINED, LEFT, ADDED, REMOVED, OWNERSHIP_TRANSFERRED}`.
- **`entity_id`** = the named entity's id; for `MEMBER` events it is the **target** user (actor in `user_id`); ownership transfer is a **`BOARD`** event.
- **Metadata** follows the snapshot principle: snapshot `card_title`/`column_title` (+ move columns); users are ids-only; `UPDATE` carries `changed_fields`; plus `break_glass` (JOINED) and `source` (LEFT/REMOVED) flags. Full per-mutation map in **SRS §4.6**.
- **Scope** = board-scoped only. Workspace-entity actions are unlogged (no `board_id`); their **board cascades emit one row per affected board**, requiring the workspace-cascade methods to `RETURNING board_id`.
- **Index:** add `CREATE INDEX ... ON activities (board_id, created_at DESC)` in the write-phase migration (the Phase-2 read pattern is fully known; free on an empty table).
- **Migration:** new `000008_add_activities` (table per SRS §7 + the index). `activities.board_id → boards ON DELETE CASCADE`, `activities.user_id → users ON DELETE SET NULL`.

### Out of scope (deferred, recorded so they aren't lost)

- **Atomicity via Unit-of-Work / Transactor (Option C)** → **integrity pass ⑤**, after the DB harness and after the composite-FK step. Resolved ⑤ order: **harness → composite-FK (ADR-008) → Transactor/atomic activity writes → deferred SQL + atomicity tests.** The Phase-1 activity call sites move inside a `WithinTransaction(...)` closure with no logging-logic rework. See memory `[[activities-logging-best-effort-then-atomic]]`, `[[post-phase1-integration-tests]]`.
- **Old→new value diffs** on `UPDATE` (only `changed_fields` names now) → a genuine Phase-2 enhancement if the feed UI needs it.
- **`activities` reads / UC-22 activity-log feed** → Phase 2.
- **Actor/target user-name snapshots** → not needed until Phase-2 account deletion (SRS §10) can null `user_id`.
- **The WebSocket broadcasts** for these mutations → step ④ (independent of the activity write).

## How it works (mechanism + worked examples)

### Write flow — single mutation (best-effort, after commit)

```
HTTP handler  ── ctx = c.Request.Context() ──►  UseCase.MoveCard(ctx, input)
                                                     │
                                    (1) access check │
                                                     ▼
                                    (2) MUTATION ──► cardRepo.Move(...)  ═══► COMMITTED (durable)
                                                     │
                    (3) state actually changed? ─ no ─────────────► (nothing logged, return)
                                                     │ yes
                                    (4) build *entity.Activity{...}   ← metadata assembled in the usecase
                                                     │
                    (5) common.WriteActivity(ctx, activityRepo, act)
                             └─ activityRepo.Log(ctx, act) ─► INSERT INTO activities
                                     ├─ ok    → row written
                                     └─ error → zerolog.Error() + SWALLOW   ← mutation is NOT failed/rolled back
                                                     │
                                    (6) return output ─────────────► HTTP 200
                                        (identical response whether step 5 succeeded or not)
```

Four properties this encodes: the write is **after commit** (2 before 4), **gated on real change** (3), **synchronous** (before 6, on the request goroutine — the ctx is still alive), and **non-fatal** (5's error is swallowed). Atomicity (Option C) would later wrap steps 2 + 5 in one `WithinTransaction(...)`; the call site at (4)–(5) doesn't change.

### Write flow — workspace cascade (one request → N board rows)

```
WorkspaceUseCase.RemoveMember(ctx, {workspace, targetUser=Bob})
      │
      ├─ RemoveWithParticipationCascade(...) ── one tx ──►  DELETE workspace_members
      │       returns: affectedCards[], affectedBoardIDs[]   DELETE board_members … RETURNING board_id
      │                                                       UPDATE cards SET assigned_to=NULL … RETURNING
      │
      └─ for each boardID in affectedBoardIDs:               ← one MEMBER/REMOVED row PER board Bob was on
             WriteActivity(board_id=boardID, actor=admin, entity_id=Bob, {source:"workspace"})

   (the per-card unassigns are NOT logged — §2.8; they become step-④ CARD_UPDATED broadcasts)
```

### Worked examples — the exact rows written

**A) Alice drags card `c-123` "Fix login bug" from To Do → Done** (UC-15). Rich snapshot metadata; renders with **no** join to (possibly-deleted) cards/columns:
```json
{
  "board_id": "b-1", "user_id": "u-alice",
  "action_type": "MOVED", "entity_type": "CARD", "entity_id": "c-123",
  "metadata": {
    "card_title": "Fix login bug",
    "from_column_id": "col-todo", "from_column_title": "To Do",
    "to_column_id":   "col-done", "to_column_title":   "Done"
  },
  "created_at": "2026-07-17T09:00:00Z"
}
```
→ *"Alice moved 'Fix login bug' from To Do to Done."*

**B) Carol (workspace ADMIN, not a member) opens PRIVATE board `b-9`** (UC-09 break-glass). Note `user_id == entity_id` (actor == target on self-join) and the security flag:
```json
{
  "board_id": "b-9", "user_id": "u-carol",
  "action_type": "JOINED", "entity_type": "MEMBER", "entity_id": "u-carol",
  "metadata": { "role": "BOARD_MEMBER", "break_glass": true },
  "created_at": "2026-07-17T10:15:00Z"
}
```
→ *"Carol joined (admin break-glass into a private board)."* A normal member join to a WORKSPACE board is identical **minus** `break_glass`. Carol re-clicking Join later = idempotent no-op → **no row** (fire rule).

**C) Bob edits `c-123`'s title + due date** (UC-16). `changed_fields` names only, no old→new diffs:
```json
{
  "board_id": "b-1", "user_id": "u-bob",
  "action_type": "UPDATED", "entity_type": "CARD", "entity_id": "c-123",
  "metadata": { "card_title": "Fix login regression", "changed_fields": ["title", "due_date"] }
}
```
→ *"Bob updated the card (title, due date)."* A PATCH that changes nothing → **no row**.

**D) Admin Alice removes Bob from workspace `ws-1`; Bob was on boards `b-1` and `b-7`** (UC-06). One request → **two** board-scoped rows (even if `b-7` had zero cards assigned to Bob — membership alone earns a row):
```json
[
  { "board_id": "b-1", "user_id": "u-alice", "action_type": "REMOVED", "entity_type": "MEMBER", "entity_id": "u-bob", "metadata": { "source": "workspace" } },
  { "board_id": "b-7", "user_id": "u-alice", "action_type": "REMOVED", "entity_type": "MEMBER", "entity_id": "u-bob", "metadata": { "source": "workspace" } }
]
```
→ each board's feed: *"Alice removed Bob (left the workspace)."*

## Consequences

**Positive**
- Zero architectural change now; the primary mutation is never at risk from a logging failure; the write is deterministically testable at the usecase layer (mock `Log`) — including the defining "log error is swallowed" guarantee — with **no DB harness needed**.
- Rows are **self-renderable and audit-faithful** at Phase-2 read time without joining to possibly-deleted cards/columns.
- The break-glass admin-join and ownership transfer — the audit-sensitive events — become queryable (`break_glass`, board-scoped `OWNERSHIP_TRANSFERRED`), discharging the ADR-005/006 deferrals.
- The correct atomic design (C) is not locked out; it's scheduled where it's cheap and testable (⑤), and the usecase-layer call sites are already positioned for it.

**Negative / notes**
- **Not atomic in Phase 1.** A mutation can commit while its activity row is lost (process death or an independent INSERT failure in the ~ms gap). It is **logged, not silent**, and closed by Option C in ⑤. Accepted because the table has no Phase-1 reader and the only high-stakes rows (break-glass, transfer) fail to write only under genuine DB trouble.
- **Vocabulary is app-enforced**, so a code path could in principle write an out-of-enum string; a DB `CHECK` (or the Phase-2 read validating) is the later guard if needed.
- **The workspace-cascade per-board rows depend on a repo-signature change** (`RETURNING board_id`); until built, those methods return `AffectedCard` only.
- **Fire-on-state-change means the usecases must diff** old-vs-new to decide whether to log and to populate `changed_fields`; this is shared with work they already do, but any new mutation must remember the rule.
