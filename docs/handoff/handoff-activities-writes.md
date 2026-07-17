# Handoff — Activity Logging: `activities` table + writes (step ③, fifth/final sub-step): implement

**Date:** 2026-07-17
**Repo:** `collabotask` (backend: `collabotask-backend/`)
**Branch:** the assignee-validation-cascade sub-step is already committed on `feat/assignee-validation-cascade`. **Cut a fresh branch** for this sub-step, e.g. `feat/activity-logging`. **Nothing written yet — this is a planning handoff** from a `/grill-with-docs` session. No code changed.

## What just happened
A grilling session locked the **full design** for the `activities` table + writes — SRS §9.1 build-order step ③, the **final** sub-step (after assignee-validation-cascade, before step ④ WebSocket). Every decision is settled and **approved by the product owner**. This is the last piece of REST/data work before the realtime layer; writing the log here means step ④ only wires *broadcast* calls and never reopens mutation code to add a log write.

**The contract is fully written down — do not re-derive it:**
- **[ADR-007](../architecture/adr/adr-007-activity-logging-writes.md)** — the decision record (write model A vs B vs C, firing rule, vocabulary, entity_id, snapshot metadata, scope) + rationale + deferrals.
- **SRS §4.6 Activity Logging** — the authoritative write contract, incl. the **full mutation → activity map table**. §7 has the table/index; UC-09 and UC-12e point back to §4.6.
- Memory `[[activities-logging-best-effort-then-atomic]]`, `[[post-phase1-integration-tests]]`.

This handoff turns that contract into a build plan. **Read ADR-007 + SRS §4.6 first; this doc assumes them.**

## Next session focus
Per root `CLAUDE.md` › "Working With Me": **non-trivial change → propose a short TodoWrite plan, then build**, **test-first (TDD)**. Turn each logged mutation into use-case tests (mock-backed), then implement to green.

Scope is **exactly**:
1. **Migration `000008_add_activities`** — the table (SRS §7) + index.
2. **`entity.Activity` + `repository.ActivityRepository` (`Log`)** + a postgres impl.
3. **One shared write helper** that logs+swallows (centralizes the Option-A policy).
4. **Inject `activityRepo` into the four logging usecases** (Card, Column, Board, Workspace) + Wire/mocks regen.
5. **Add a `Log(...)` call to each state-changing mutation** per the §4.6 map, gated by the fire-on-state-change rule.
6. **Extend the two workspace-cascade repo methods to `RETURNING board_id`** (so the per-affected-board rows can be written).
7. Tests for all of the above.

**Nothing else** — no WebSocket broadcast (step ④), no composite-FK / schema-for-invariant (ADR-008 candidate, pass ⑤), no activity **reader**/feed (UC-22, Phase 2), no Transactor/atomic write (Option C, pass ⑤).

---

## The design (settled) — build notes

### 1. Migration `000008_add_activities`
`migrations/000008_add_activities.up.sql` (+ `.down.sql`), following the SRS §7 shape and the FK-by-meaning table (§3.4):
```sql
CREATE TABLE activities (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    board_id    UUID NOT NULL REFERENCES boards(id)  ON DELETE CASCADE,
    user_id     UUID          REFERENCES users(id)   ON DELETE SET NULL,   -- actor; never NULL in Phase 1
    action_type VARCHAR NOT NULL,     -- app-enforced enum (no CHECK — ADR-007)
    entity_type VARCHAR NOT NULL,     -- app-enforced enum
    entity_id   UUID NOT NULL,
    metadata    JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_activities_board_created ON activities (board_id, created_at DESC);
```
`.down.sql`: `DROP TABLE IF EXISTS activities;` (index drops with it). Match the timestamp type used by existing tables (check `000004`; use `TIMESTAMPTZ`/`timestamp` consistently). **No `CHECK`** on `action_type`/`entity_type` — deliberately app-enforced (ADR-007).

### 2. Domain — `entity.Activity` + `repository.ActivityRepository`
`internal/domain/entity/activity.go` — mirror the `entity.Board`/`entity.BoardVisibility` style (typed string enums + `TableName()`):
```go
type ActivityEntityType string
const ( ActivityEntityBoard  ActivityEntityType = "BOARD"
        ActivityEntityColumn ActivityEntityType = "COLUMN"
        ActivityEntityCard   ActivityEntityType = "CARD"
        ActivityEntityMember ActivityEntityType = "MEMBER" )

type ActivityActionType string
const ( ActivityCreated, ActivityUpdated, ActivityDeleted, ActivityMoved,
        ActivityArchived, ActivityUnarchived, ActivityJoined, ActivityLeft,
        ActivityAdded, ActivityRemoved, ActivityOwnershipTransferred ActivityActionType = /* "CREATED" … */ )

type Activity struct {
    ID         uuid.UUID
    BoardID    uuid.UUID
    UserID     *uuid.UUID              // actor (SET NULL-able; never nil in Phase 1)
    ActionType ActivityActionType
    EntityType ActivityEntityType
    EntityID   uuid.UUID
    Metadata   map[string]any          // marshaled to jsonb
    CreatedAt  time.Time
}
```
`internal/domain/repository/activity_repository.go`:
```go
type ActivityRepository interface {
    Log(ctx context.Context, a *entity.Activity) error
    // Optional: LogMany for the workspace-cascade N-rows path (else loop Log).
}
```
Postgres impl in `internal/repository/postgres/activity.go` (+ `activity_queries.go`), same `db *pgxpool.Pool` shape as the others; `Log` = one `INSERT` (marshal `Metadata` → jsonb). **A `Log` failure returns an error to the caller — the *swallow* happens in the helper (§3), not here.**

### 3. The shared write helper (centralizes Option A)
Put the swallow+log policy in **one** place so all ~17 call sites are one-liners and the resilience guarantee lives once. Suggested `internal/usecase/common/activity.go`:
```go
// WriteActivity performs the best-effort, after-commit activity write (ADR-007 Option A):
// on error it logs and swallows — a logging failure must never fail the caller's mutation.
func WriteActivity(ctx context.Context, repo repository.ActivityRepository, a *entity.Activity) {
    if err := repo.Log(ctx, a); err != nil {
        log.Error().Err(err).Str("board_id", a.BoardID.String()).
            Str("action", string(a.ActionType)).Str("entity", string(a.EntityType)).
            Msg("activity log write failed (swallowed)")
    }
}
```
Call sites become: `common.WriteActivity(ctx, uc.activityRepo, &entity.Activity{ ... })` **after the mutation succeeds**, guarded by the fire rule. (Keeping metadata assembly in the usecase is deliberate — it's what makes the eventual Option-C move a pure relocation.)

### 4. Wiring — inject `activityRepo` into four usecases
Add `ProvideActivityRepository` to `internal/injection/providers.go` (mirror `ProvideBoardMemberRepository`, `providers.go:51`). Thread `repository.ActivityRepository` into **all four** logging usecases' structs + `New…UseCase` params + providers, then **regenerate Wire** (`go generate ./internal/injection/...` — do not hand-finalize `wire_gen.go`):
- `ProvideCardUseCase` (`providers.go:92`) → `CardUseCase` (`internal/usecase/card/card_usecase.go`)
- `ProvideColumnUseCase` (`providers.go:86`) → `ColumnUseCase`
- `ProvideBoardUseCase` (`providers.go:75`) → `BoardUseCase`
- `ProvideWorkspaceUseCase` (`providers.go:67`) → `WorkspaceUseCase` *(for the per-board cascade rows only)*

### 5. Call sites — one `WriteActivity` per state-changing mutation (§4.6 map)
Add after the mutation commits, **only when state actually changed** (fire rule). Map (see SRS §4.6 for exact metadata shape):

| File | Row | Fire gate |
| --- | --- | --- |
| `card/create_card.go` | CARD/CREATED `{card_title}` | always |
| `card/move_card.go` | CARD/MOVED `{card_title, from/to_column_id+title}` | column or position changed |
| `card/update_card.go` | CARD/UPDATED `{card_title, changed_fields}` | `changed_fields ≠ ∅` |
| `card/delete_card.go` | CARD/DELETED `{card_title}` | always |
| `column/create_column.go` | COLUMN/CREATED `{column_title}` | always |
| `column/update_column.go` | COLUMN/UPDATED `{column_title, changed_fields}` | changed |
| `column/delete_column.go` | COLUMN/DELETED `{column_title}` | always |
| `column/update_column_position.go` | COLUMN/MOVED `{column_title}` | position changed |
| `board/update_board.go` | BOARD/UPDATED `{changed_fields, visibility_from/to?}` | changed |
| `board/set_archived.go` | BOARD/ARCHIVED\|UNARCHIVED `{}` | state changed |
| `board/transfer_ownership.go` | BOARD/OWNERSHIP_TRANSFERRED `{from_user_id,to_user_id}` | real transfer (guard at `transfer_ownership.go:37` already returns on no-op) |
| `board/self_join_board.go` | MEMBER/JOINED `{role, break_glass?}` | `joined==true` only (`self_join_board.go:68`) |
| `board/invite_member.go` | MEMBER/ADDED `{role}` ×invitee | per newly-added |
| `board/leave_board.go` | MEMBER/LEFT `{source:"board"}` | always |
| `board/remove_member.go` | MEMBER/REMOVED `{source:"board"}` | always |
| `workspace/remove_member.go` | MEMBER/REMOVED `{source:"workspace"}` ×affected board | per affected board |
| `workspace/leave_workspace.go` | MEMBER/LEFT `{source:"workspace"}` ×affected board | per affected board |

Notes that will bite if ignored:
- **`entity_id` for MEMBER events = the *target* user** (joiner/invitee/leaver/removed), **actor stays in `UserID`** (ADR-007 / §4.6). For self-join & leave they coincide.
- **`break_glass`** = the joiner is a workspace **admin who was not already a member**, joining a **PRIVATE** board. `self_join_board.go` already resolves visibility + eligibility — compute the flag there; the existing TODO is at `self_join_board.go:68`.
- **Replace the two placeholder TODOs** already in the tree: `transfer_ownership.go:41` and `self_join_board.go:68` (and the note in `board_io.go:126`).
- **`changed_fields`/state-diff:** the update usecases must compare old→new to both *decide whether to log* and *populate `changed_fields`*. If a usecase doesn't currently load the pre-image, add the fetch (cheap) — or derive from the input DTO's presence flags (e.g. card update already tracks `AssignedToPresent`).
- **`move_card`** needs `from/to` column **titles** — it already resolves columns for validation; reuse them for metadata rather than adding queries.

### 6. Workspace-cascade repo change — return the affected board ids
The per-affected-board rows (UC-06/UC-06c) need the set of boards the user was removed from. Today `WorkspaceMemberRepository.RemoveWithParticipationCascade` (`internal/repository/postgres/workspace_member.go:169`) returns **only** `[]AffectedCard` — insufficient (`AffectedCard` has no `board_id`, and a board the user was on with **zero** assigned cards wouldn't appear at all but still needs a row).

**Change:** have the cascade also capture the deleted `board_members` rows' board ids. The existing `DELETE FROM board_members … WHERE b.workspace_id=$1 AND bm.user_id=$2` (§4.2 UC-06 SQL) becomes `… RETURNING board_id`. Extend the return, e.g. `([]AffectedCard, []uuid.UUID, error)` (or a small result struct). This is the **same "return it now, broadcast/log later" pattern** already used for `AffectedCard`. Wire both `workspace/remove_member.go` and `workspace/leave_workspace.go` to loop those board ids → one `WriteActivity` each (actor = admin for remove, self for leave; `source:"workspace"`). The board-scoped cascade (`BoardMemberRepository.RemoveWithParticipationCascade`) needs **no** return change — the usecase already knows the single `board_id`.

### 7. Mocks + Wire
- **Mock regen (mandatory):** add `ActivityRepository: {}` under `collabotask/internal/domain/repository` in **`.mockery.yaml`** (next to `BoardMemberRepository`), then run mockery → `internal/mocks/repository_mocks.go` gains `MockActivityRepository`. The `WorkspaceMemberRepository` signature change also regenerates its mock.
- **Wire regen (mandatory):** the four `Provide…UseCase` signature changes won't compile until regenerated.

---

## Out of scope / DEFERRED (do NOT do here — recorded so they aren't lost)
- **WebSocket broadcasts** (`CARD_UPDATED`, `USER_JOINED`, `OWNERSHIP_TRANSFERRED`, `MEMBER_REMOVED`, …) → **step ④**. Activity write and broadcast are independent; the cascade methods already return `[]AffectedCard` for the future `CARD_UPDATED`.
- **Atomic activity writes (Option C / Unit-of-Work / Transactor)** → **integrity pass ⑤**, after the DB harness and the composite-FK (**ADR-008**). Resolved ⑤ order in `[[post-phase1-integration-tests]]`. Keep metadata assembly in the usecase so this becomes a pure relocation into a `WithinTransaction(...)` closure.
- **Composite-FK assignee invariant** (ADR-008 candidate) → pass ⑤. Unrelated to this sub-step beyond the shared ⑤ home.
- **Activity reader / UC-22 feed** → Phase 2.
- **Old→new value diffs** on UPDATE (we store `changed_fields` names only) → Phase-2 enhancement if the feed needs it.
- **Async write** — rejected (ADR-007): the write is synchronous. (Request `ctx` is canceled on handler return; async needs `context.WithoutCancel` + panic `recover()` for no Phase-1 benefit.)
- **Repo-layer SQL tests** for the `Log` INSERT and the `RETURNING board_id` → post-Phase-1 integration pass (no DB harness — `[[post-phase1-integration-tests]]`).

## Docs to update **after** code (audit closure — same rhythm as prior sub-steps)
- SRS **§9 P1** "WebSocket layer entirely … `activities` table + writes" → strike the `activities` half to ✅ (WebSocket remains for step ④).
- SRS **§9.1 step ③** activities bullet → mark ✅ DONE; note best-effort/Option-A, per-affected-board cascade rows, migration `000008`.
- SRS **§9.2** map → annotate the `activities` half of the WebSocket row as ✅.
- SRS **UC-09 / UC-12e** deferred-log notes → strike "silent until then" for the activity write (broadcast still waits for ④).
- Root `CLAUDE.md` › `## Now` → mark the activities sub-step ✅; move the queue pointer to **step ④ (WebSocket + participation broadcast)**; note step ③ (REST features) is now complete.
- `collabotask-backend/temp_unit-test-checklist.md` → add the activity-write cases (below) + the deferred repo-layer `Log`/`RETURNING` SQL note.
- **NOT touching, and why:** `CONTEXT.md` (no new/renamed domain *term* — "activity" already in the model); **PRD / user-stories** (US-09 already says the join is "recorded in the activity history"; no scope change). ADR-007 is already written (this session).

## Tests (test-first, use-case layer, mock-backed)
Per logged mutation, the **three contract points** (ADR-007 / Q7):
1. **Happy path** — on a successful state-changing mutation, `activityRepo.Log` is called **once** with the exact expected row (`board_id`, actor `user_id`, `action_type`, `entity_type`, `entity_id`, `metadata`). Assert the argument (incl. `entity_id` = target for MEMBER events).
2. **Resilience** — when `Log` returns an error, the mutation still returns **success** (the Option-A guarantee; test at representative sites + once on `common.WriteActivity`).
3. **No-op silence** — a state-unchanged request (no-op update/move/archive, idempotent re-join, no-op transfer) does **not** call `Log`.

Plus targeted cases:
- **self_join:** `break_glass:true` when non-member admin joins a PRIVATE board; **absent/false** for a normal member join; **no** `Log` when already a member (`joined:false`).
- **transfer_ownership:** logged on real transfer; **not** logged on the idempotent no-op (target already owner); `from_user_id=null` on an orphan-board appointment.
- **update_card / update_board / update_column:** `changed_fields` reflects exactly the changed fields; empty change → no `Log`.
- **workspace remove_member / leave_workspace:** **one** `MEMBER/REMOVED`(or `LEFT`) per affected board id returned by the cascade (`source:"workspace"`); a board the user was on with zero assigned cards **still** gets a row.
- **board leave/remove:** one `MEMBER/LEFT`|`REMOVED` `{source:"board"}` on that board.

**Deferred (no DB harness — `[[post-phase1-integration-tests]]`):** the actual `INSERT`/jsonb round-trip and the `RETURNING board_id` SQL — test the *usecase* with mocks; add a checklist note.

## ⚠️ Build-time risks to verify
- **Wire regen** — four `Provide…UseCase` signature changes; won't compile until regenerated.
- **Mock regen** — new `ActivityRepository` + changed `WorkspaceMemberRepository` break the mock package; regenerate before running tests.
- **Ripple into existing use-case tests** — adding `activityRepo` to four `New…UseCase` constructors touches **every** card/column/board/workspace use-case test's setup (add the new mock; for happy-path tests, stub `Log → nil`). Compile-driven — expect a wide but mechanical test-setup sweep.
- **`WorkspaceMemberRepository.RemoveWithParticipationCascade` return change** ripples into its existing UC-06/UC-06c tests and the board-ownership/participation tests that mock it.
- **Metadata jsonb marshaling** — ensure `map[string]any` marshals deterministically enough for assertions (assert decoded map, not raw bytes).

## Context / reference (do not re-read unless needed)
- **Decision record:** `docs/architecture/adr/adr-007-activity-logging-writes.md`.
- **Contract:** SRS `docs/spesifications/001-software-specifications.md` **§4.6** (map + rules), §7 (table/index), §2.7, §2.8, UC-09/UC-12e.
- **Story:** `docs/product/001-user-stories.md` US-09 (the only user-facing mention — "recorded in the activity history").
- **Model / deferrals:** `[[activities-logging-best-effort-then-atomic]]`, `[[post-phase1-integration-tests]]` (⑤ order; Option C; composite-FK = ADR-008).
- **Patterns to mirror:** repo `internal/repository/postgres/board_member.go` (+ `_queries.go`); DI `internal/injection/providers.go`; entity `internal/domain/entity/board.go` (typed enums + `TableName()`); the prior sub-step's shape `docs/handoff/handoff-assignee-validation-cascade.md`.
- Backend conventions & the "add an endpoint" recipe: `collabotask-backend/README.md` (Conventions) + `TESTING.md`.
- **Current placeholder TODOs to replace:** `internal/usecase/board/transfer_ownership.go:41`, `internal/usecase/board/self_join_board.go:68`, `internal/usecase/board/board_io.go:126`.

## Suggested skills
- **tdd** — the three contract points per mutation, red→green.
- **code-review** (ultra / xhigh) — after implementation; verify: swallow policy centralized in one helper; fire-on-state-change actually gated; `entity_id`=target for MEMBER events; per-affected-board cascade rows (incl. the zero-assigned-card board); no WebSocket/broadcast crept in.

## Quick verification commands (from `collabotask-backend/`)
- `go generate ./internal/injection/...` (Wire) + the mockery step (regen mocks)
- `migrate` up to `000008` against a scratch DB (confirm the migration applies + rolls back)
- `go build ./... && go vet ./...`
- `go test ./internal/usecase/... ./internal/delivery/...`
