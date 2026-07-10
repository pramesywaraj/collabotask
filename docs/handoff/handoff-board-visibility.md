# Handoff — Board Visibility (step ③, first sub-step): implement

**Date:** 2026-07-10
**Repo:** `collabotask` (backend: `collabotask-backend/`)
**Branch:** `main` · nothing written yet — this is a **planning handoff** from a `/grill-me` session. No code changed.

## What just happened
A grilling session worked out the full design for **board `visibility`** (spec §9.1 build-order
step ③, the *first* sub-step: cross-cutting, do before workspace ops / transfer / assignee /
activities). Every decision below is settled and approved by the product owner. **Nothing is
implemented.** Next session should `/implement` this, test-first.

## Next session focus
Build board visibility per the decisions below. Per root `CLAUDE.md` › "Working With Me":
**non-trivial change → propose a short TodoWrite plan, then build.** Approach is **test-first (TDD)**:
turn the access truth-table into table-driven use-case tests, then implement to green.

Scope is **exactly**: migration 000007 · the 3-method access checker · create/list/detail/kanban/
self-join/update-settings · card & column mutations switch to the mutate-checker. **Nothing else.**

---

## The design (settled)

### 1. Migration `000007` (bundled, up + down)
- Add `boards.visibility varchar(20) NOT NULL DEFAULT 'WORKSPACE' CHECK (visibility IN ('PRIVATE','WORKSPACE'))`.
- Change `boards.created_by`: **drop NOT NULL** + change FK to **`ON DELETE SET NULL`** (was `RESTRICT`).
  Schema hygiene / forward-compat only — it can **never be NULL in Phase 1** (no account deletion).

### 2. Entity / DTO
- `entity.Board` gets `Visibility BoardVisibility` (new type + consts `BoardVisibilityWorkspace="WORKSPACE"`, `BoardVisibilityPrivate="PRIVATE"`).
- `entity.BoardListItem` gets `CardCount uint`.
- `Board.CreatedBy` **stays `uuid.UUID`** (do NOT make it a pointer — see Deferred).
- Board response DTO + create/update inputs get `visibility`.

### 3. Access model — 3-method checker (replaces `Check`/`Resolve`)
`common/board_access_check.go` exposes three intent methods over a private `resolve()` that
gathers facts (board + workspace member + board member + visibility):
- `CheckMetadataAccess` → board **detail**
- `CheckViewAccess` → **kanban**
- `CheckMutateAccess` → card & column mutations

**Truth table (the test table):**

| Requester | Metadata (detail) | View (kanban) | Mutate (card/column) |
|---|---|---|---|
| Board owner / member | ✅ | ✅ | ✅ |
| Workspace **admin**, not joined, **WORKSPACE** | ✅ | ✅ | ✅ |
| Workspace **admin**, not joined, **PRIVATE** | ✅ `CAN_JOIN`, **roster hidden** | 🔓 403 `BOARD_JOIN_REQUIRED` | 🔓 403 `BOARD_JOIN_REQUIRED` |
| Workspace **member**, not on board, **WORKSPACE** | ✅ | ✅ | ❌ 403 (reveal) |
| Workspace **member**, not on board, **PRIVATE** | ❌ 404 (hide existence) | ❌ 404 | ❌ 404 |
| Not a workspace member | 404 / not-in-workspace | same | same |

Key rules:
- **Break-glass is PRIVATE-only** and covers **mutations too** — a non-joined admin cannot view *or*
  mutate a PRIVATE board; they get `BOARD_JOIN_REQUIRED` and must self-join first. On WORKSPACE-visible
  boards an admin acts freely with no join. (§2.4's "✅(admin)" for mutation = the *post-join* steady state.)
- **404 hides / 403 reveals:** an ineligible **plain member** on a PRIVATE board gets **404** (existence
  hidden, matches the list filtering it out, blocks ID enumeration). A **non-joined admin** gets **403
  `BOARD_JOIN_REQUIRED`** (they can already see it in their list → prompt to join). On a WORKSPACE board,
  a plain member denied *mutation* gets **403** (they legitimately see the board), not 404.
- **`created_by` is removed from ALL access/role logic** — the checker, the list query's filter +
  `user_role`/`access_status` fallbacks, and the detail `user_role` fallback. `board_members` is the sole
  source (the creator is always seeded as `BOARD_OWNER`). `created_by` survives only as a stored/returned trace.

### 4. Endpoint behavior
- **UC-07 Create:** accept `visibility` (validate `oneof=PRIVATE WORKSPACE`, default `WORKSPACE` when omitted).
- **UC-08 List** (`getUserBoardsInWorkspace`): filter `wm.role='ADMIN' OR b.visibility='WORKSPACE' OR bm.user_id IS NOT NULL`
  (drop `created_by`); add `card_count`; `access_status` = `CAN_JOIN` for a non-member on a WORKSPACE board.
- **UC-12 Detail** (`CheckMetadataAccess`): **thin metadata pre-join** — return board fields + `access_status`,
  but **omit the `Members` roster when PRIVATE + not-joined** (rule "3b": only ever the non-joined admin).
  Joined viewers, and everyone on WORKSPACE boards, get the full roster.
- **UC-12 Kanban** (`CheckViewAccess`): break-glass on PRIVATE.
- **UC-09 Self-join:** eligibility — admin → any board; member → WORKSPACE-visible only; ineligible → **403**
  (remap `ErrBoardCannotJoin` 409→403). **Idempotent** already-member → **200** + body `{ "joined": false, "message": "You are already a member of this board" }`; newly joined → `{ "joined": true }`. Use
  `INSERT … ON CONFLICT (board_id,user_id) DO NOTHING RETURNING *` — a returned row = newly joined; no row = already member.
- **UC-12b Update:** accept `visibility` (gated by `can_administer_board`). **No data cascade** on a flip —
  members stay, assignments already require membership. WORKSPACE→PRIVATE only affects *future* access.
- **Card & column mutations** (`create/update/delete/move` card, `create/update/delete/reorder` column):
  swap `boardAccessChecker.Check` → `CheckMutateAccess`.

### 5. Errors / mapper
- New `domain.ErrBoardJoinRequired` → **403** with error code **`BOARD_JOIN_REQUIRED`**.
- Remap `domain.ErrBoardCannotJoin`: **409 → 403** (after idempotency, it only ever means "ineligible").
- Ineligible plain member on PRIVATE → `domain.ErrBoardNotFound` (**404**).

---

## Out of scope / DEFERRED (do NOT do here — recorded so they aren't lost)
- **Self-join activity log** (UC-09 "log activity") → lands in the **`activities` sub-step** of step ③.
- **WebSocket broadcasts** (`USER_JOINED` on self-join, `BOARD_UPDATED` on visibility flip) → **step ④**.
  Self-join and settings-change are intentionally "silent" for now.
- **`Board.CreatedBy` → `*uuid.UUID`** → **Phase 2**, with account deletion/deactivation (§10) — the only
  feature that can make it NULL. Add a breadcrumb comment at the entity field.
- **`leave_board` owner guard → role-based** (`internal/usecase/board/leave_board.go:27`, currently
  `board.CreatedBy == requester`) → **UC-12e** (ownership transfer), where `board_members.role` becomes the
  authoritative owner check. Correct until then (`created_by == owner` holds). Add a breadcrumb comment.

## Docs to update alongside the code
- **New `docs/architecture/adr/adr-005-board-visibility-access.md`** — rationale for the 3-level checker,
  break-glass-on-mutation, 404-hide vs 403-reveal, idempotent self-join, thin roster (rejected alts:
  literal-matrix mutation reading, 403-reveal, error-on-rejoin).
- **SRS `001-software-specifications.md`:** add `BOARD_JOIN_REQUIRED`; the 404-vs-403 rule; break-glass
  covers mutations; idempotent self-join response + ineligible→403; thin-roster (UC-12); note migration
  000007 and that `created_by` is no longer used in access logic. Mark §9.1 step-③ visibility ✅ **after** build.
- **User stories** (optional): US-09 idempotent re-join scenario; US-12 PRIVATE-hidden-from-member (404).
- **Root `CLAUDE.md` › `## Now`** (optional): note visibility in progress.

## Tests
- **Test-first, use-case layer, mock-backed** — the truth table above IS the test table. Cover: all three
  checker methods × the matrix; self-join eligibility + idempotent already-member; create-board visibility
  default; thin-roster-on-PRIVATE; 404-vs-403 split. Includes the two pre-marked cases in
  `collabotask-backend/temp_unit-test-checklist.md:419-421`.
- **Deferred (no DB harness):** the list query's SQL visibility filter + `card_count` — test the *usecase*
  with mocks; add a checklist note. Repo-layer/integration tests are a dedicated pass after Phase 1.

## Context / reference (do not re-read unless needed)
- Contract: `docs/spesifications/001-software-specifications.md` §2.2–2.4, UC-07/08/09/12/12b, §6, §7, §9.1.
- Stories: `docs/product/001-user-stories.md` US-07/08/09/12/12b.
- Backend conventions & the "add an endpoint" recipe: `collabotask-backend/CLAUDE.md`.
- Current code touchpoints:
  - `internal/usecase/common/board_access_check.go` (the checker — becomes 3 methods)
  - `internal/usecase/board/{create_board,get_boards_in_workspace,get_board_detail,get_board_kanban,self_join_board,update_board}.go`
  - `internal/usecase/{card,column}/*` (swap to `CheckMutateAccess`)
  - `internal/repository/postgres/{board,board_queries,board_member_queries}.go`
  - `internal/delivery/http/errors/domain_mapper.go`, `internal/domain/sentinel_errors.go`
  - `internal/domain/entity/board.go`, `internal/delivery/http/response/board_response.go`
  - `migrations/000007_*` (new)

## Suggested skills
- **tdd** — write the access-matrix table tests first (red → green).
- **code-review** (xhigh) — after implementation, against the working tree.

## Quick verification commands (from `collabotask-backend/`)
- `go build ./... && go vet ./...`
- `go test ./internal/usecase/... ./internal/delivery/...`
