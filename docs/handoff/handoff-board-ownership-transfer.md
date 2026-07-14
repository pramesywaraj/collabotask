# Handoff — Board Ownership Transfer (step ③, third sub-step): implement

**Date:** 2026-07-14
**Repo:** `collabotask` (backend: `collabotask-backend/`)
**Branch:** `feat/board-ownership-transfer` · nothing written yet — this is a **planning handoff** from a `/grill-me` session. No code changed.

## What just happened
A grilling session worked out the full design for **UC-12e Board Ownership Transfer** — spec §9.1 build-order step ③, the sub-step *after* workspace ops. Every decision below is settled and **approved by the product owner**, and the rationale is recorded in **ADR-006** (`docs/architecture/adr/adr-006-board-ownership-transfer.md`). The contract is written into **SRS §4.3 UC-12e** and the §9.1 build-order bullet. **Nothing is implemented.** Next session should `/implement` this, **test-first**.

> Read **ADR-006** first — it carries the *why* for every choice below (especially the break-glass-vs-governance reasoning, which was re-litigated several times before landing).

## Next session focus
Build ownership transfer per the decisions below. Per root `CLAUDE.md` › "Working With Me": **non-trivial change → propose a short TodoWrite plan, then build.** Approach is **test-first (TDD)**: turn each guard branch into a table-driven use-case test (mock-backed), then implement to green.

Scope is **exactly**: the `created_by`-as-owner **proxy removal** · UC-12e transfer/appoint use case + endpoint · one new transactional repo method · one new sentinel error + mapper entry · request DTO / handler / route / swagger · regenerated mocks · tests. **Nothing else** — no WebSocket broadcast, no activity-log write, no richer roles.

---

## The design (settled — see ADR-006)

### 0. No migration
Uses existing tables/columns only (`board_members.role`). The single-owner invariant is enforced by the **operation**, not the schema.

### 1. Proxy removal — DO THIS FIRST (its own commit)
`created_by` stops being an owner proxy; `board_members.role` becomes the sole authority. Two edits, both flagged in-code today with "moves in UC-12e" comments:

- **`internal/usecase/board/utils.go` — `canAdministerBoard`**: drop the `createdBy` / `requesterID` params. New shape:
  ```go
  func canAdministerBoard(boardMember *entity.BoardMember, workspaceMember *entity.WorkspaceMember) bool {
      return (boardMember != nil && !boardMember.IsEmpty() && boardMember.IsOwner()) ||
             (workspaceMember != nil && !workspaceMember.IsEmpty() && workspaceMember.IsAdmin())
  }
  ```
  Then fix **every caller** — `grep -rn "canAdministerBoard" internal/usecase/board` (expect: `update_board.go`, `set_archived.go`, `invite_member.go`, `remove_member.go`, `get_workspace_invitees_for_board.go`). Compile will drive this.
- **`internal/usecase/board/leave_board.go` — owner guard**: replace `if board.CreatedBy == input.RequesterID { return ErrBoardOwnerCannotLeave }` (lines ~27–34) with a **role** check. Fetch the requester's board-member row first, then `if boardMember.IsOwner() { return domain.ErrBoardOwnerCannotLeave }`. (The row is already fetched a few lines down — reorder so the guard uses it.) Delete the big "DEFERRED (UC-12e)" comment block.

**Why first:** leaving the proxy in produces two concrete bugs the moment a transfer happens — the old owner keeps admin power via `created_by`, and the real new owner can leave without transferring (the guard checks the wrong person). ADR-006 › Context has the full trace. Lock both shut with regression tests (see Tests §12–13).

### 2. UC-12e — Transfer / Appoint (the feature)
- **Route:** `POST /api/v1/workspace/:workspace_id/board/:board_id/transfer-ownership` · Auth · **can_administer_board**.
- **Request:** `{ "to_user_id": "uuid" }` (`binding:"required"`).
- **Usecase `TransferOwnership(ctx, TransferOwnershipInput)`** — guard order:
  1. `validator.Struct(input)`.
  2. **`access, err := bu.boardAccessChecker.CheckMutateAccess(ctx, input.BoardID, input.RequesterID)`** — this single call already does archived→404, board-not-found→404, PRIVATE-hide→404, workspace-membership→403, and **break-glass** (non-joined admin on PRIVATE → `ErrBoardJoinRequired` 403). `BoardUseCase` already has `boardAccessChecker` wired — no new DI.
  3. `if access.Board.WorkspaceID != input.WorkspaceID { return domain.ErrBoardNotFound }` (path/board consistency, as the other board UCs do).
  4. **`if !canAdministerBoard(access.BoardMember, access.WorkspaceMember) { return domain.ErrBoardPermissionDenied }`** — narrows "any joined member" (which `CheckMutateAccess` admits) down to **owner-or-admin**. *This is the key step:* `CheckMutateAccess` alone is intentionally too permissive (a plain board member passes it); `canAdministerBoard` is what rejects them with 403.
  5. Fetch the **target's** board-member row: `GetMemberByBoardAndUser(BoardID, ToUserID)`. `ErrBoardMemberNotFound` → **`ErrTransferTargetNotBoardMember`** (400). Any other error → wrap 500.
  6. **Idempotency:** `if targetBM.IsOwner() { return nil }` — target already owns it → 200 no-op (before touching the write path).
  7. `bu.boardMemberRepo.TransferOwnership(ctx, input.BoardID, input.ToUserID)`.
- **Returns** 200 + `null` data (mirrors remove/leave). Handler message e.g. `"Board ownership transferred"`.

> **Access matrix this produces** (verify against ADR-006 / SRS UC-12 — it's `CheckMutateAccess` ∘ `canAdministerBoard`):
> owner → ✅ · plain board member → 403 · admin joined → ✅ · admin not-joined + WORKSPACE → ✅ (no join) · admin not-joined + PRIVATE → 403 `BOARD_JOIN_REQUIRED` · ws-member not on board + WORKSPACE → 403 · ws-member not on board + PRIVATE → 404 · archived/not-found → 404.

### 3. The atomic repo method
Add to **`BoardMemberRepository`** (`internal/domain/repository/board_member_repository.go`):
```go
// TransferOwnership atomically moves the sole BOARD_OWNER title to newOwnerID:
// demote-by-role (0 rows on an orphan board) + promote-by-id, in one tx.
// Returns the demoted owner's id (nil on an orphan board) for the future
// OWNERSHIP_TRANSFERRED broadcast — return it now so step ④ is a pure wire-up.
TransferOwnership(ctx context.Context, boardID, newOwnerID uuid.UUID) (fromUserID *uuid.UUID, err error)
```
Impl in `internal/repository/postgres/board_member.go` (+ queries in `board_member_queries.go`), one transaction (follow the `RemoveWithParticipationCascade` pattern in `workspace_member.go`):
```sql
-- 1) demote current owner BY ROLE; capture who it was (0 or 1 rows):
UPDATE board_members SET role='BOARD_MEMBER'
  WHERE board_id=$1 AND role='BOARD_OWNER' RETURNING user_id;   -- pgx.ErrNoRows on orphan → fromUserID=nil
-- 2) promote target BY ID; assert exactly 1 row (else target vanished mid-tx → error):
UPDATE board_members SET role='BOARD_OWNER'
  WHERE board_id=$1 AND user_id=$2;                             -- RowsAffected()==1 else domain.ErrBoardMemberNotFound
```
Handle the demote's `pgx.ErrNoRows` as "no current owner" (orphan) → `fromUserID = nil`, **not** an error. Assert the promote affected exactly 1 row.

### 4. New error + mapper
- `internal/domain/sentinel_errors.go`: `ErrTransferTargetNotBoardMember = errors.New("transfer target must be a board member")`.
- `internal/delivery/http/errors/domain_mapper.go`: add it to the **`ErrCodeValidation` (400)** case, right next to `ErrInvalidAssigneeID` — same rule shape (assignment = participation, §2.8).

### 5. Delivery layer (mirror `RemoveMemberFromBoard`)
- `internal/delivery/http/request/board_request.go`: `TransferOwnershipRequest { ToUserID uuid.UUID \`json:"to_user_id" binding:"required"\` }`.
- `internal/delivery/http/handler/board_handler.go`: `TransferOwnership` handler — `GetAndCheckUserID` → `parseBoardPathParams` → bind → build `board.TransferOwnershipInput` → call → `GenerateSuccessResponse(..., nil)`. Add the swagger `// @Router .../transfer-ownership [post]` block (403/400/404/500 failures).
- `internal/delivery/http/router/router.go`: `boards.POST("/:board_id/transfer-ownership", cfg.BoardHandler.TransferOwnership)` on the boards group.
- `internal/usecase/board/board_io.go`: `TransferOwnershipInput { RequesterID, WorkspaceID, BoardID, ToUserID uuid.UUID \`validate:"required"\` }`.

### 6. Mocks
The new `BoardMemberRepository.TransferOwnership` method **breaks the generated mock** → tests won't compile until regenerated. Regenerate the repository mocks the same way the workspace-ops step did (`go generate ./...` or the project's mockery step — check the existing mock file header for the command). Confirm `BoardMemberRepository` mock now has `TransferOwnership` before writing usecase tests.

---

## Out of scope / DEFERRED (do NOT do here — recorded so they aren't lost)
- **Activity-log write** for the transfer → the `activities` sub-step of step ③. Leave a marked `// TODO(activities, UC-12e)` at the point after a successful transfer.
- **`OWNERSHIP_TRANSFERRED` broadcast** (`{ board_id, from_user_id, to_user_id }`) → **step ④**. The repo method already returns `fromUserID` so step ④ is a pure wire-up. Leave a marked `// TODO(ws, step ④)`.
- **Richer board/workspace roles + owner delegation** → **Phase 2+** (memory: `role-model-enrichment-deferred`). Two-roles-per-layer is deliberate; "owner unavailable" is already covered by the layered model + orphan-safe boards + this endpoint.
- **A "list board members (admin view)" endpoint** for non-joined admins to discover targets → frontend-era, shared with `remove-member` (ADR-006 › Consequences). Not needed for the backend endpoint (target is validated server-side).

## Docs to update
**Before code — ALREADY DONE (this handoff's session):** ADR-006 written; SRS §4.3 UC-12e fleshed out; SRS §9.1 build-order bullet annotated. **Do not redo these.**

**After code (audit closure — same as the workspace-ops commit):**
- SRS **§9** "🟡 P1 — Missing Phase-1 features": strike `ownership transfer (UC-12e)` → ✅ (ADR-006).
- SRS **§9** "🔵 P2 — Two sources of truth for board ownership": mark **✅ resolved** (ADR-006 — `board_members.role` is now the sole owner authority; `created_by` is a trace).
- SRS **§9.2** map: flip the `Missing: ownership transfer` and `Two sources of truth for board ownership` rows to ✅.
- Root `CLAUDE.md` › `## Now` — move the pointer past ownership transfer to "assignee board-member validation + unassign cascade (UC-14/UC-16, UC-10/UC-12d)".
- `collabotask-backend/temp_unit-test-checklist.md` — add the UC-12e cases + the deferred repo-layer test note.

**NOT touching, and why:** `CONTEXT.md` (no new/renamed domain *term* — "board owner" already exists); **PRD / user-stories** (US-12e already covers it, no scope change).

## Tests
**Test-first, use-case layer, mock-backed.** Cover every branch:

**Transfer (`transfer_ownership_test.go`):**
1. Owner → board member: old owner becomes `BOARD_MEMBER`, target becomes sole `BOARD_OWNER` (assert `TransferOwnership` called).
2. Workspace admin transfers between two members on a **WORKSPACE** board → success, no join.
3. Non-joined admin on a **PRIVATE** board → `403 BOARD_JOIN_REQUIRED` (assert `TransferOwnership` **not** called).
4. Admin who *has* joined the private board → success.
5. Target not a board member → `400 ErrTransferTargetNotBoardMember`.
6. Target already owns it → `200` no-op (assert `TransferOwnership` **not** called).
7. Orphan board (no owner) + admin appoints a target member → success, `fromUserID` nil (mock returns nil).
8. Plain board member (non-owner, non-admin) attempts → `403 ErrBoardPermissionDenied`; and a ws-member not on a **PRIVATE** board → `404`.
9. Archived board → `404 ErrBoardNotFound`.
10. Board not found / workspace mismatch → `404`.
11. Missing/malformed `to_user_id` → `400` (validation).

**Proxy-removal regression (guards the two bugs — update existing test files):**
12. `update_board` / `set_archived`: the old creator who is **not** a workspace admin and holds a `BOARD_MEMBER` row (simulating post-transfer) → **denied** 403 (proves authority no longer leaks via `created_by`).
13. `leave_board`: a `BOARD_OWNER` → blocked (`ErrBoardOwnerCannotLeave`); a plain `BOARD_MEMBER` who happens to match `created_by` → **allowed** to leave (proves the guard is role-based now).
14. Update all existing `update_board` / `set_archived` / `invite_member` / `remove_member` / `get_workspace_invitees_for_board` tests for the new `canAdministerBoard(boardMember, workspaceMember)` signature (the `createdBy` arg is gone).

**Deferred (no DB harness — [[post-phase1-integration-tests]]):** the actual `TransferOwnership` SQL (demote-by-role RETURNING + promote row-count assert, incl. the orphan 0-row path) — test the *usecase* with mocks; add a note to `temp_unit-test-checklist.md`.

## ⚠️ Build-time risks to verify
- **Mock regen is mandatory** — the new interface method won't compile against the old mock. Regenerate before running tests.
- **`canAdministerBoard` signature change ripples** to ~5 callers + their tests. Compile-driven; work through each. Don't leave the `createdBy` arg passed anywhere.
- **Router boot** — `/:board_id/transfer-ownership` is a *static* suffix under the `:board_id` param, alongside `/:board_id/{invite,member,leave,join,invitees}`. Should be fine (all static suffixes), but Gin panics at startup on wildcard/child conflicts — `go test ./internal/delivery/http/...` (or run the server) to confirm it still boots.

## Context / reference (do not re-read unless needed)
- **Rationale:** `docs/architecture/adr/adr-006-board-ownership-transfer.md` (read first).
- **Contract:** `docs/spesifications/001-software-specifications.md` §2.5, §2.6, §2.3, §2.8, **UC-12e**, §9.1/§9.2; sibling **ADR-005** (visibility/break-glass, thin roster).
- **Story:** `docs/product/001-user-stories.md` **US-12e**.
- Backend conventions & the "add an endpoint" recipe: `collabotask-backend/CLAUDE.md`.
- **Current code touchpoints:**
  - `internal/usecase/board/utils.go` (`canAdministerBoard` signature) · `leave_board.go` (owner guard) · callers of `canAdministerBoard`
  - new `internal/usecase/board/transfer_ownership.go` (+ `_test.go`) · `board_io.go` (input) · `board_usecase.go` (has `boardAccessChecker`, `boardMemberRepo` already)
  - `internal/usecase/common/board_access_check.go` (`CheckMutateAccess` — reuse, don't reimplement)
  - `internal/domain/repository/board_member_repository.go` (new method) · `internal/repository/postgres/{board_member,board_member_queries}.go`
  - `internal/domain/sentinel_errors.go` · `internal/delivery/http/errors/domain_mapper.go`
  - `internal/delivery/http/handler/board_handler.go` · `.../request/board_request.go` · `.../router/router.go`
  - `internal/domain/entity/board_member.go` (`IsOwner()`, `IsEmpty()`) · the generated `BoardMemberRepository` mock

## Suggested skills
- **tdd** — guard-branch tables first (red → green).
- **code-review** (ultra / xhigh) — after implementation, against the working tree; verify the access matrix and that no `created_by` owner-proxy remains.

## Quick verification commands (from `collabotask-backend/`)
- `go generate ./...` (mocks) — or the project's mockery step
- `go build ./... && go vet ./...`
- `go test ./internal/usecase/... ./internal/delivery/...`
- Router boot check: `go test ./internal/delivery/http/...` or run the server.
