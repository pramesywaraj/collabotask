# Handoff — Workspace Ops (step ③, second sub-step): implement

**Date:** 2026-07-11
**Repo:** `collabotask` (backend: `collabotask-backend/`)
**Branch:** `main` · nothing written yet — this is a **planning handoff** from a `/grill-me` session. No code changed.

## What just happened
A grilling session worked out the full design for the **workspace operations** batch — spec §9.1
build-order step ③, the sub-step *after* board visibility: **promote/demote (UC-06b), leave
(UC-06c), delete (UC-06d)**, plus a **retrofit of the already-shipped UC-06 (remove member)** to
carry the participation cascade it currently lacks. Every decision below is settled and approved by
the product owner. **Nothing is implemented.** Next session should update two spec sections, then
`/implement` this, test-first.

## Next session focus
Build workspace ops per the decisions below. Per root `CLAUDE.md` › "Working With Me":
**non-trivial change → propose a short TodoWrite plan, then build.** Approach is **test-first (TDD)**:
turn each guard branch into a table-driven use-case test (mock-backed), then implement to green.

Scope is **exactly**: UC-06d delete · UC-06b promote/demote · UC-06c leave · UC-06 cascade retrofit ·
the shared **workspace-scoped** participation cascade · new sentinel errors + mapper entries ·
routes/handlers/DTOs. **Nothing else** — no WebSocket, no board-scoped cascade, no ownership transfer.

---

## The design (settled)

### 0. No migration
All three ops use existing tables/columns (`workspace_members.role`, `workspaces.owner_id`,
`board_members`, `cards.assigned_to`). The DB already cascades `workspaces → boards → columns → cards`
and both membership tables (migrations 000002–000004, all `ON DELETE CASCADE`).

### 1. UC-06d — Delete Workspace (smallest)
- **Route:** `DELETE /api/v1/workspace/:workspace_id` · Auth · **owner only**.
- **Usecase `DeleteWorkspace`:** load workspace → guard `workspace.OwnerID == requester` else
  `ErrNotWorkspaceOwner` (**403**) → call existing `WorkspaceRepository.Delete` (DB cascades boards/
  columns/cards/members).
- **"UI confirmation" is a frontend concern** — backend exposes plain `DELETE`, no token/param.
- **Returns** 200 + `null` data.

### 2. UC-06b — Promote / Demote (`PATCH .../member/:user_id/role`)
- **Route:** `PATCH /api/v1/workspace/:workspace_id/member/:user_id/role` · Auth · **workspace ADMIN**.
- **Request:** `{ "role": "ADMIN" | "MEMBER" }` (validate `oneof=ADMIN MEMBER`).
- **Usecase `SetMemberRole`:**
  - requester must be admin → `ErrNotWorkspaceAdmin` (**403**, exists).
  - target must be a member → `ErrMemberNotFound` (**404**, exists).
  - **demote guards** (only when new role = MEMBER and target currently ADMIN):
    - target is the workspace **owner** → `ErrCannotDemoteOwner` (**403**).
    - target is the **last** admin → `ErrCannotDemoteLastAdmin` (**409**).
  - **Self-demotion is ALLOWED** — an admin may demote *themselves*, gated by the same two guards
    (industry norm: Trello/Jira/Slack/Notion/GitHub all allow self-step-down while ≥1 admin remains).
  - **Idempotent:** setting a role the member already has → 200 no-op, returns the member.
- **Repo:** new `WorkspaceMemberRepository.UpdateRole(ctx, wsID, userID, role) (*WorkspaceMember, error)`
  (`UPDATE ... RETURNING *`) and `CountAdmins(ctx, wsID) (int, error)` for the last-admin guard.
- **Returns** the updated member (role). **Plain REST — no broadcast** (spec §2.7).

### 3. UC-06c — Leave (`POST .../leave`) + UC-06 retrofit — THE CASCADE
This is the crux of the batch. **Decision: build the workspace-scoped participation cascade NOW**
(don't defer to the later §9.1 cascade step) — deferring would strand a leaver as a `board_member`
of boards in a workspace they've left, still assignable. §9.1's own note: the cascade is a plain
`UPDATE/DELETE` with **no WebSocket dependency; only the broadcast waits for step ④**.

- **Shared repo method (one transaction)** — used by BOTH UC-06c and the UC-06 retrofit, since the
  cascade is identical (same user, same workspace); only the *usecase guards* differ:
  ```sql
  BEGIN;
  DELETE FROM workspace_members WHERE workspace_id=$1 AND user_id=$2;
  DELETE FROM board_members bm USING boards b
    WHERE bm.board_id=b.id AND b.workspace_id=$1 AND bm.user_id=$2;
  UPDATE cards c SET assigned_to=NULL
    FROM columns col JOIN boards b ON col.board_id=b.id
    WHERE c.column_id=col.id AND b.workspace_id=$1 AND c.assigned_to=$2
    RETURNING c.id, c.column_id;
  COMMIT;
  ```
  Suggested name: `WorkspaceMemberRepository.RemoveWithParticipationCascade(ctx, wsID, userID)
  ([]AffectedCard, error)`. **Return the affected card ids/columns** even though nothing consumes
  them yet — this makes step ④ a pure add (wire the broadcast), no repo change. If the
  `workspace_members` row is absent, return `ErrMemberNotFound` (**404**).
- **UC-06c `LeaveWorkspace`** (`POST /api/v1/workspace/:workspace_id/leave` · Auth · member):
  - member must exist → `ErrMemberNotFound` (**404**).
  - **owner** cannot leave → `ErrWorkspaceOwnerCannotLeave` (**403**) — must transfer/delete first
    (mirrors existing `ErrBoardOwnerCannotLeave` → 403).
  - **last admin** cannot leave → reuse `ErrCannotDemoteLastAdmin`? **No** — add a distinct
    `ErrLastAdminCannotLeave` (**409**) for a clear client message.
  - then call the shared cascade. **Returns** 200 + `null`.
- **UC-06 retrofit** (`RemoveMember`, `internal/usecase/workspace/remove_member.go`): keep its existing
  guards (admin-only 403; `ErrCannotRemoveYourself` 400) but **replace the single
  `workspaceMemberRepo.Delete` call with the shared cascade** so removal also clears board membership
  + assignments. Guards unchanged.
- **Owned boards survive** a leave/remove — the cascade drops the `board_members` row (incl. any
  `BOARD_OWNER`), leaving the board **owner-less-but-safe** (§2.6); the board and its `created_by` are
  untouched. Not an error.

### 4. Status-code policy (the one real judgment call — SETTLED)
Split, because it makes **both** existing mapper precedents consistent at once:
| Guard | Error | Code | Why |
|---|---|---|---|
| not admin (promote/demote/remove) | `ErrNotWorkspaceAdmin` | **403** | exists |
| not owner (delete) | `ErrNotWorkspaceOwner` (new) | **403** | authority rule |
| owner can't leave | `ErrWorkspaceOwnerCannotLeave` (new) | **403** | mirrors `ErrBoardOwnerCannotLeave` |
| demote owner | `ErrCannotDemoteOwner` (new) | **403** | permanent structural rule |
| demote last admin | `ErrCannotDemoteLastAdmin` (new) | **409** | *resolvable* state conflict (promote another admin first) — matches how mapper already uses 409 (`ErrInconsistentState`, `ErrAlreadyMember`) |
| last admin can't leave | `ErrLastAdminCannotLeave` (new) | **409** | same |
| target / self not a member | `ErrMemberNotFound` | **404** | exists |

**Rule of thumb:** permanent authority/structural = **403**; "fix the state and retry" = **409**.

---

## Out of scope / DEFERRED (do NOT do here — recorded so they aren't lost)
- **All WebSocket broadcasts** — `MEMBER_REMOVED` + per-card `CARD_UPDATED` on leave/remove; nothing
  for promote/demote (plain REST). → **step ④**. The cascade repo method returns affected card ids now
  so step ④ is a pure wire-up.
- **Board-scoped participation cascade** (UC-10 leave board, UC-12d remove-from-board) — same concept,
  single-board `WHERE` clause. → lands with **board leave/remove** work, not now.
- **Board ownership transfer (UC-12e)** — the **next** queue item after this batch.

## Docs to update
**Before code (technical contract — flow spec→code):**
- **SRS `001-software-specifications.md` §4.2** — add status codes to UC-06b/c/d inline (owner→403,
  last-admin→409, target-not-member→404), self-demotion allowed + idempotent no-op (UC-06b),
  not-a-member→404 (UC-06c), non-owner→403 (UC-06d).
- **SRS §9.1 build order** — record the deviation: the **workspace-scoped** participation cascade
  (UC-06 + UC-06c) lands in *this* step, not the later cascade step; board-scoped (UC-10/12d) still
  later; broadcast still step ④.

**After code (audit closure — same as the visibility commit):**
- Flip §9.2 audit rows + §9 "🟡 P1" list entries for promote/demote, leave, delete to ✅.
- Root `CLAUDE.md` › `## Now` — move the pointer past workspace ops.

**NOT touching, and why:** `CONTEXT.md` (no new/renamed domain *term*); **no new ADR** (refinements of
already-decided architecture — §2.8 cascade, §2.6 owner-less-but-safe, ADR-005 role authority — not a
new hard-to-reverse choice); **PRD** (no scope change). **Optional:** a US-06b self-demotion scenario.

## Tests
- **Test-first, use-case layer, mock-backed.** Cover every guard branch:
  - UC-06b: not-admin→403, target-not-member→404, demote-owner→403, demote-last-admin→409,
    self-demote allowed (not-owner, not-last), idempotent no-op, successful promote/demote.
  - UC-06c: not-a-member→404, owner→403, last-admin→409, successful leave (cascade invoked).
  - UC-06d: not-owner→403, successful delete.
  - UC-06 retrofit: existing guards still hold; cascade method now invoked (assert the call).
- **Deferred (no DB harness — [[post-phase1-integration-tests]]):** the actual cascade SQL
  (cross-workspace `board_members` delete + `cards` unassign) and `CountAdmins` — test the *usecase*
  with mocks; add a note to `collabotask-backend/temp_unit-test-checklist.md`. Repo-layer/integration
  tests are the dedicated post-Phase-1 pass.

## ⚠️ Build-time risk to verify
Gin route registration: the spec path `member/:user_id/role` joins existing
`member/invite` and `member/remove/:user_id` — a **static/param mix under `/member/`**. Gin's router
can panic with "wildcard conflicts with existing children" on such trees. **Verify the router still
boots** after adding the PATCH route (`go run`/tests will panic at startup if it conflicts). If it
panics, raise it before reshaping the path — don't silently deviate from the spec's URL.

## Context / reference (do not re-read unless needed)
- Contract: `docs/spesifications/001-software-specifications.md` §2.6, §2.7, §2.8, UC-06/06b/06c/06d, §9.1/§9.2.
- Stories: `docs/product/001-user-stories.md` US-06/06b/06c/06d.
- Backend conventions & the "add an endpoint" recipe: `collabotask-backend/CLAUDE.md`.
- Current code touchpoints:
  - `internal/usecase/workspace/{remove_member,workspace_usecase,workspace_io}.go` (retrofit + shared plumbing)
  - new `internal/usecase/workspace/{set_member_role,leave_workspace,delete_workspace}.go` (+ `_test.go`)
  - `internal/domain/repository/{workspace_repository,workspace_member_repository}.go` (new methods)
  - `internal/repository/postgres/{workspace_member,workspace_member_queries}.go` (UpdateRole, CountAdmins, cascade)
  - `internal/domain/sentinel_errors.go` (new errors) · `internal/delivery/http/errors/domain_mapper.go` (mapper)
  - `internal/delivery/http/handler/workspace_handler.go` · `.../request/workspace_request.go` (SetRole DTO)
  - `internal/delivery/http/router/router.go` (3 new routes on the `workspaces` group)
  - `internal/domain/entity/workspace_member.go` (has `IsAdmin()`); `workspaces.owner_id` on `entity.Workspace`

## Suggested skills
- **tdd** — guard-branch tables first (red → green).
- **code-review** (ultra / xhigh) — after implementation, against the working tree.

## Quick verification commands (from `collabotask-backend/`)
- `go build ./... && go vet ./...`
- `go test ./internal/usecase/... ./internal/delivery/...`
- Router boot check (catches the Gin wildcard risk): `go test ./internal/delivery/http/...` or run the server.
