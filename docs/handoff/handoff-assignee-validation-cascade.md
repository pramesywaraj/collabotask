# Handoff — Assignee Board-Member Validation + Board-Scoped Unassign Cascade (step ③, fourth sub-step): implement

**Date:** 2026-07-15
**Repo:** `collabotask` (backend: `collabotask-backend/`)
**Branch:** currently `feat/board-ownership-transfer` (ownership work already merged-in). Cut a fresh branch for this sub-step, e.g. `feat/assignee-validation-cascade`. **Nothing written yet — this is a planning handoff** from a `/grill-with-docs` session. No code changed.

## What just happened
A grilling session locked the full design for **assignee = board-member validation (UC-14/UC-16)** and the **board-scoped unassign cascade (UC-10/UC-12d)** — spec §9.1 build-order step ③, the sub-step *after* board ownership transfer, *before* activities. This is the **"data-correction only"** slice: it makes the data correct, but wires **no** WebSocket broadcast (that is step ④). Every decision below is settled and **approved by the product owner**.

The contract already largely exists in **SRS §2.8 (Governance vs Participation)**, **UC-14/UC-16**, **UC-10/UC-12d**, and the **§9 audit items** (#7 / the P1 "assignee not validated" bullet). This handoff records the *four decisions the grill added on top* and the *one architectural idea we deliberately deferred*. **Nothing is implemented.** Next session should `/implement` this, **test-first**.

## Next session focus
Build the two halves per the decisions below. Per root `CLAUDE.md` › "Working With Me": **non-trivial change → propose a short TodoWrite plan, then build.** Approach is **test-first (TDD)**: turn each branch into a table-driven use-case test (mock-backed), then implement to green.

Scope is **exactly**:
- **Half 1 — validation:** wire `boardMemberRepo` into `CardUseCase`; add the board-membership gate to `create_card` and `update_card`; one new sentinel error + mapper entry.
- **Half 2 — cascade:** one new atomic repo method on `BoardMemberRepository`; swap `leave_board` and `remove_member` onto it; regenerated mocks.
- Tests for both halves; Wire regen; mock regen.

**Nothing else** — no `CARD_UPDATED` broadcast, no activity-log write, no schema/migration, no composite-FK (see *Deferred* below).

---

## The design (settled)

### 0. No migration
Both halves use existing tables/columns only. The invariant "assignee ∈ board members" is enforced at the **application layer** this step (a DB-level enforcement is the deferred composite-FK idea below — explicitly **not** done here).

### 1. Wiring — `CardUseCase` gains `boardMemberRepo` (do this first)
`CardUseCase` today has `cardRepo, columnRepo, userRepo, boardAccessChecker` only (`internal/usecase/card/card_usecase.go`). Add `boardMemberRepo repository.BoardMemberRepository` to the struct + `NewCardUseCase` params + the assignment block.
- `internal/injection/` — update **`ProvideCardUseCase`** signature to take the board-member repo, then pass it at the call site. `boardMemberRepository` is **already constructed** in `wire_gen.go:38` (`ProvideBoardMemberRepository(db)`); the call at `wire_gen.go:46` just needs the extra arg. Regenerate Wire (`go generate ./internal/injection/...`) — do **not** hand-edit `wire_gen.go` as the final state; let regen produce it.

> Use `boardMemberRepo.IsUserExists(ctx, boardID, userID)` — the exact membership predicate. **Do NOT** use `boardAccessChecker` for the assignee: the access checker grants non-joined workspace admins access (governance), but §2.8 gives assignment **no admin exception** — a non-member admin is *not* a valid assignee. Different rule, different mechanism (see Decision D below).

### 2. Half 1 — validation

**`create_card.go`** (currently `internal/usecase/card/create_card.go:35-49`): the assignee block today does `userRepo.GetById` and 404s if the user doesn't exist globally. Replace the gate:
1. nil-UUID → `domain.ErrInvalidAssigneeID` (400) — keep as-is.
2. `isMember, err := cru.boardMemberRepo.IsUserExists(ctx, input.BoardID, *input.AssignedTo)` → `!isMember` → `domain.ErrAssigneeNotBoardMember` (400).
3. Then `userRepo.GetById` **only for the response payload** (`Assignee` name/avatar). After membership is confirmed the user necessarily exists (FK `board_members.user_id → users ON DELETE CASCADE`), so this is display-only; keep its existing error wrap defensively.

**`update_card.go`** (currently `internal/usecase/card/update_card.go:70-90`): **this is the subtle one — see Decision B.** Put the membership gate **inside the `if input.AssignedToPresent` branch**, next to line 70, and only when `input.AssignedTo != nil`:
```go
if input.AssignedToPresent {
    if input.AssignedTo != nil {
        isMember, err := cru.boardMemberRepo.IsUserExists(ctx, input.BoardID, *input.AssignedTo)
        if err != nil {
            return nil, fmt.Errorf("failed to verify assignee board membership: %w", err)
        }
        if !isMember {
            return nil, domain.ErrAssigneeNotBoardMember
        }
    }
    card.AssignedTo = input.AssignedTo
}
```
- **Delete the comment at `update_card.go:79`** ("When the board-member rule (§2.8) is added, it belongs here too") — the check deliberately does **not** go in the display-resolution block (lines 80-90). That block stays a pure `userRepo.GetById` for the payload, **with no membership gate**, so re-resolving an *unchanged* (possibly stale) assignee never blocks an edit.
- The nil-UUID guard already exists at `update_card.go:26-28` (`ErrInvalidAssigneeID`).

Net behavior (both files): set a non-member assignee → **400**; clear to `null` → allowed; edit title/description/due-date with an untouched assignee → **never blocked**, even if that assignee is stale.

### 3. New error + mapper
- `internal/domain/sentinel_errors.go`: `ErrAssigneeNotBoardMember = errors.New("assignee must be a board member")` (place next to `ErrInvalidAssigneeID`).
- `internal/delivery/http/errors/domain_mapper.go`: add it to the **`ErrCodeValidation` (400)** case, right next to `ErrInvalidAssigneeID` / `ErrTransferTargetNotBoardMember` — same rule shape (assignment = participation, §2.8).

### 4. Half 2 — the atomic cascade repo method
Add to **`BoardMemberRepository`** (`internal/domain/repository/board_member_repository.go`), mirroring `workspaceMemberRepository.RemoveWithParticipationCascade` (`internal/repository/postgres/workspace_member.go:169`):
```go
// RemoveWithParticipationCascade deletes the board_members row and clears that
// user's card assignments on THIS board, in one tx. Returns the unassigned cards
// (for the future CARD_UPDATED broadcast — returned now so step ④ is a pure
// wire-up). ErrBoardMemberNotFound when the row is absent.
RemoveWithParticipationCascade(ctx context.Context, boardID, userID uuid.UUID) ([]AffectedCard, error)
```
Impl in `internal/repository/postgres/board_member.go` (+ query in `board_member_queries.go`), one transaction:
```sql
-- 1) delete membership; 0 rows → domain.ErrBoardMemberNotFound
DELETE FROM board_members WHERE board_id=$1 AND user_id=$2;
-- 2) unassign that user's cards on THIS board (columns.board_id means no boards join needed)
UPDATE cards c SET assigned_to = NULL
  FROM columns col
  WHERE c.column_id = col.id AND col.board_id = $1 AND c.assigned_to = $2
  RETURNING c.id, c.column_id;
```
Reuse the existing `repository.AffectedCard` struct (defined in `workspace_member_repository.go:12`). Follow the workspace method's tx/rollback/scan/commit shape exactly.

### 5. Half 2 — usecase edits
- **`internal/usecase/board/leave_board.go:42`**: swap `boardMemberRepo.Delete(...)` → `RemoveWithParticipationCascade(...)`. The **owner-cannot-leave guard at lines 38-40 stays before it**, so a leaving owner never reaches the cascade. Discard the returned slice for now (`_, err =`).
- **`internal/usecase/board/remove_member.go:46`**: swap `Delete(...)` → `RemoveWithParticipationCascade(...)`. **No new guards** — an owner can still be removed (orphan-safe, §2.6); the cascade simply unassigns whoever is removed. Discard the returned slice for now.
- Both keep their existing `ErrBoardMemberNotFound` handling (the cascade preserves that sentinel on a 0-row delete).

### 6. Mocks + Wire
- **Mock regen (mandatory):** the new `BoardMemberRepository.RemoveWithParticipationCascade` breaks the generated `MockBoardMemberRepository` in `internal/mocks/repository_mocks.go` → tests won't compile until regenerated. Use mockery (`github.com/vektra/mockery` — same tool that produced the file; check the project's mockery step / config). The workspace equivalent (`MockWorkspaceMemberRepository_...RemoveWithParticipationCascade...` at `repository_mocks.go:2853+`) is the shape to expect.
- **Wire regen (mandatory):** the `ProvideCardUseCase` signature change won't compile until Wire is regenerated.

---

## The four grilled decisions (rationale — so they aren't re-litigated)

**A. Refuse, no auto-add.** Assigning a non-member → 400, not auto-join. Keeps "participation = a deliberate join" and avoids a silent content-access grant on PRIVATE boards. (SRS §2.8 line 127; the picker lists only members, so the API check is a safety net.)

**B. Validate the *newly-set* assignee only — never re-validate an unchanged one.** The gate fires only on `AssignedToPresent && != nil`. Rationale: `update_card` re-loads the current assignee *for the response payload*; if the membership check attached there, a **stale existing assignee would block an unrelated edit** (e.g. Bob fixing a typo on a card still assigned to since-departed Alice would 400). The cascade (Half 2) is what keeps assignees honest; the write-time check only polices *new* assignments.

**C. Collapse to a single 400.** `IsUserExists` returns false for both "real user, not on this board" and "not a user at all", so there is **one** gate → `ErrAssigneeNotBoardMember` (400). Drop the old 404-`ErrUserNotFound` path for the assignee. Simpler, and more private (a board mutator can't probe global user existence). Consistent with the mirror rule `ErrTransferTargetNotBoardMember` (also 400).

**D. No admin exception; "joined" restricts the *assignee*, not the *assigner* (Reading 1).**
- *Actor axis* (who can assign): unchanged — anyone with `CheckMutateAccess`, i.e. any board member **and** non-joined workspace admins (governance). Assigning is a participation-level action open to any board member, **not** reserved for owners/admins.
- *Target axis* (who can be assigned): must have a `board_members` row — checked via `IsUserExists`, **no admin exception**. A non-joined admin can assign *members* but cannot assign **themselves** until they Join (for a PRIVATE board, that Join is the logged break-glass).

---

## DEFERRED — the composite-FK invariant (record, do NOT build here)
The grill surfaced a stronger design than the app-layer check + explicit cascade: enforce the whole invariant in the **schema**. On this exact stack it is available and elegant:
- `board_members` PK is **`(board_id, user_id)`** — a ready-made composite-FK target.
- Postgres **16** (docker-compose) supports `ON DELETE SET NULL (column_list)` (PG 15+).
- No cross-board card move in Phase 1 → a card's board is effectively immutable.

**The idea:** add `board_id` to `cards` (backfill, NOT NULL), then
`cards(board_id, assigned_to) → board_members(board_id, user_id) ON DELETE SET NULL (assigned_to)`.
That single migration would: (1) make assignee-validity a **DB guarantee** on every write path and **close the check→write TOCTOU race** for free; (2) make the unassign-cascade **automatic** on *every* membership-loss path (board leave/remove, workspace leave/remove, user deletion) in the same tx as the delete — subsuming the data-mutation half of both the workspace cascade (already shipped) and this board cascade.

**Why deferred, not now:** it's a schema/architecture change beyond "data-correction only"; it **reopens shipped, tested** workspace-cascade code; it doesn't fully replace the explicit `UPDATE … RETURNING` (the FK nulls rows *silently*, so the step-④ broadcast still needs the affected-card list); and it wants **SQL-level tests** the project has deliberately deferred to the post-Phase-1 integration pass. It also **supersedes the §2.8 "shared clear-assignments helper"** approach, so it deserves its own ADR when adopted.

**Where it goes:** this is the headline candidate for the **post-Phase-1 integrity pass** — write it up as **`adr-008`** *at that time* (not now; it's a candidate, not a decision). *(Renumbered from `adr-007`, which is now activity logging — ADR-007.)* See `[[post-phase1-integration-tests]]`.

## Out of scope / DEFERRED (do NOT do here — recorded so they aren't lost)
- **`CARD_UPDATED` (`assigned_to: null`) broadcasts** for the cleared cards → **step ④**. The cascade method already returns `[]AffectedCard` so step ④ is a pure wire-up. Leave a marked `// TODO(ws, step ④)` where the slice is discarded in `leave_board.go` / `remove_member.go`.
- **Activity-log writes** → the `activities` sub-step of step ③ (membership change is the only log entry; per-card unassigns are never logged — §2.8 line 129).
- **The TOCTOU window** on the app-layer check (assignee leaves between check and write): **accepted** for Phase 1 — §2.8 calls the API check "a safety net," not a guarantee. Closing it properly is the composite-FK / `FOR SHARE` work above, deferred to the integrity pass.
- **Repo-layer SQL tests** for the cascade `UPDATE` → post-Phase-1 integration pass (no DB harness — `[[post-phase1-integration-tests]]`).
- **`move_card` re-validation** — none. Move doesn't change the assignee; its assignee-resolution at `move_card.go:63` stays display-only.

## Docs to update
**Before code:** nothing required — the contract is already in SRS §2.8 / UC-14 / UC-16 / UC-10 / UC-12d / §9. (This handoff is the only upfront artifact.)

**After code (audit closure — same rhythm as the ownership-transfer commit):**
- SRS **§9** "🟡 P1 — Assignee not validated as a board member" (`create_card.go`/`update_card.go`) → strike to **✅**, note `ErrAssigneeNotBoardMember` + the Decision-B "newly-set only" rule.
- SRS **§9.2** map: flip `Assignee must be a board member` and `Unassign on lost participation (board-scoped: UC-10, UC-12d)` rows to ✅.
- SRS **UC-16**: add the one-line Decision-B clarification (validate only a newly-set assignee; a stale existing assignee never blocks an unrelated edit).
- SRS **§9 deferred / integrity-pass note**: record the composite-FK invariant as the `adr-008` candidate.
- Root `CLAUDE.md` › `## Now`: mark this sub-step ✅; move the pointer to the **activities** sub-step; note the board-scoped cascade broadcast still waits for step ④.
- `collabotask-backend/temp_unit-test-checklist.md`: add the UC-14/16 + UC-10/12d cases + the deferred repo-layer cascade-SQL note.

**NOT touching, and why:** `CONTEXT.md` (no new/renamed domain *term* — "assignee"/"board member" already exist); **PRD / user-stories** (US-14/16, US-10/12d already cover it, no scope change); **no ADR now** (composite-FK is a deferred candidate, not an adopted decision).

## Tests (test-first, use-case layer, mock-backed)
**Validation — `create_card_test.go` / `update_card_test.go`:**
1. Assignee is a board member → success (assert `IsUserExists` called; card written).
2. Assignee is **not** a board member → `400 ErrAssigneeNotBoardMember` (assert **no** card write).
3. Assignee is a nil UUID → `400 ErrInvalidAssigneeID`.
4. `create`: no assignee (`AssignedTo == nil`) → success, `IsUserExists` **not** called.
5. `update`: clear assignee (`AssignedToPresent && AssignedTo == nil`) → success, `IsUserExists` **not** called.
6. `update` **Decision B**: edit *title only* on a card whose existing assignee is a non-member (stale) → **success** (assert `IsUserExists` **not** called; the edit is not blocked).
7. `update`: set a new non-member assignee while also editing the title → `400` (assert **no** write).

**Cascade — `leave_board_test.go` / `remove_member_test.go`:**
8. Member leaves → `RemoveWithParticipationCascade` called (not `Delete`); success.
9. **Owner** leaves → `ErrBoardOwnerCannotLeave`, cascade **not** called (guard precedes it).
10. Admin removes a member → cascade called; success.
11. Remove/leave a user with **no** membership row → `ErrBoardMemberNotFound` (idempotent not-found preserved).
12. Existing permission/guard tests for `remove_member` (requester authority) still pass with the swapped repo call.

**Deferred (no DB harness — `[[post-phase1-integration-tests]]`):** the actual cascade `UPDATE cards … RETURNING` SQL and the `board_members` `DELETE` row-count semantics — test the *usecase* with mocks; add a note to `temp_unit-test-checklist.md`.

## ⚠️ Build-time risks to verify
- **Wire regen** — `ProvideCardUseCase` signature change won't compile until regenerated. `boardMemberRepository` is already in scope in `wire_gen.go:38`; regen just threads it into `wire_gen.go:46`.
- **Mock regen** — the new interface method breaks `MockBoardMemberRepository`. Regenerate before running tests.
- **Existing card tests** construct `CardUseCase` / its mock set — adding `boardMemberRepo` to `NewCardUseCase` ripples into every card use-case test's setup. Compile-driven; update the constructors (add the new mock, and for existing assignee-happy-path tests, stub `IsUserExists → true`).
- **`create_card` output** still needs the `Assignee` entity for the response — don't drop the `userRepo.GetById` display fetch when replacing the *gate*.

## Context / reference (do not re-read unless needed)
- **Contract:** `docs/spesifications/001-software-specifications.md` §2.8, UC-14, UC-16, UC-10, UC-12d, §9 (#7 / P1 "assignee not validated"), §9.2 map.
- **Story:** `docs/product/001-user-stories.md` US-14, US-16, US-10, US-12d.
- **Model:** memory `[[assignment-participation-model]]` (the grilled 2026-06-25 decisions this implements) and `[[post-phase1-integration-tests]]`.
- **Pattern to mirror:** `internal/repository/postgres/workspace_member.go:169` (`RemoveWithParticipationCascade`) + `workspace_member_queries.go`.
- Backend conventions & the "add an endpoint" recipe: `collabotask-backend/CLAUDE.md`.
- **Current code touchpoints:**
  - `internal/usecase/card/{card_usecase.go, create_card.go, update_card.go}` (+ `_test.go`)
  - `internal/usecase/board/{leave_board.go, remove_member.go}` (+ `_test.go`)
  - `internal/domain/repository/board_member_repository.go` (new method) · `internal/repository/postgres/{board_member.go, board_member_queries.go}`
  - `internal/domain/sentinel_errors.go` · `internal/delivery/http/errors/domain_mapper.go`
  - `internal/injection/{wire.go, wire_gen.go}` (`ProvideCardUseCase`) · the generated `MockBoardMemberRepository`

## Suggested skills
- **tdd** — branch tables first (red → green), especially Decision-B test #6.
- **code-review** (ultra / xhigh) — after implementation; verify the two axes (actor vs assignee), that the cascade is one transaction, and that no membership gate leaked into the display-resolution block.

## Quick verification commands (from `collabotask-backend/`)
- `go generate ./internal/injection/...` (Wire) — and the mockery step for mocks
- `go build ./... && go vet ./...`
- `go test ./internal/usecase/... ./internal/delivery/...`

---

## 📋 Code-review findings (2026-07-15, post-implementation)

A two-axis `/code-review` (Standards + Spec) ran against the working-tree diff vs `HEAD` (7d7c5ca). **Verdict: clean, on-spec, no correctness defects and no scope creep.** All four grilled decisions (A–D), the single-tx cascade, `ErrBoardMemberNotFound` preservation, the owner-guard ordering, and the `// TODO(ws, step ④)` markers are all implemented correctly. The 7 validation + 5 cascade test cases all map to the diff.

The items below are **the only things worth adjusting** — all are minor/optional; nothing blocks shipping.

### Actionable (do this one)
1. **Duplicated Code — extract the membership gate.** The identical 6-line block appears in both `internal/usecase/card/create_card.go` and `internal/usecase/card/update_card.go`:
   ```go
   isMember, err := cru.boardMemberRepo.IsUserExists(ctx, input.BoardID, *input.AssignedTo)
   if err != nil {
       return nil, fmt.Errorf("failed to verify assignee board membership: %w", err)
   }
   if !isMember {
       return nil, domain.ErrAssigneeNotBoardMember
   }
   ```
   Pull it into a private `CardUseCase` helper, e.g. `func (cru *CardUseCase) requireBoardMember(ctx context.Context, boardID, userID uuid.UUID) error`, and call it from both sites. Keep create's separate `uuid.Nil` guard where it is — only the membership predicate is shared. Judgement call, but worth doing.

### Optional (nice-to-know, no action required)
2. **Weak assertion in cascade test #9** (`leave_board_test.go`, "owner leaves → cascade not called"). It currently proves the negative only indirectly via mockery strict-mock (an unexpected call would fail `NewMock...(t)`). That's acceptable, but an explicit `AssertNotCalled(t, "RemoveWithParticipationCascade", ...)` would make the intent obvious to a future reader. Optional.
3. **Cosmetic rollback-style divergence.** The new `internal/repository/postgres/board_member.go` cascade uses `defer func() { _ = tx.Rollback(ctx) }()`, whereas the sibling it mirrors (`workspace_member.go`) uses bare `defer tx.Rollback(ctx)`. Both are correct; the new form is arguably cleaner. Only change it if the repo values uniformity with the sibling. No functional impact.

### Not a code issue, but flag it
4. **Broken doc pointer:** several places (root `CLAUDE.md`, this handoff line 176, the "add an endpoint" references) point at `collabotask-backend/CLAUDE.md`, but that file does **not** exist on disk — the actual backend standards live in `collabotask-backend/README.md` ("Conventions"), `TESTING.md`, and the ADRs. Not part of this sub-step, but worth correcting the pointer separately so future agents find the real standards.

### Still recommended before commit
- The review was **static** (read-only). Run `go build ./... && go test ./internal/usecase/... ./internal/delivery/...` to confirm the regenerated Wire + mocks actually compile and the new test cases pass.
