# Handoff — Composite-FK Assignee Invariant (⑤ integrity pass, step 2): implement

**Date:** 2026-08-24
**Repo:** `collabotask` (backend: `collabotask-backend/`)
**Status:** Design **settled + approved** via `/grill-with-docs` (2026-08-24). Recorded as **ADR-010** (Accepted). **Nothing implemented yet.** Next session should `/implement` this test-first, using the ADR-015 harness for the SQL-level tests.

This is the **second** item of the ⑤ post-Phase-1 integrity pass: harness ✅ → **composite-FK assignee invariant (this)** → ADR-007 Transactor/Option-C → remaining deferred SQL tests. Do it before Transactor so C composes the post-reorder cascade code.

---

## The one-paragraph summary

Enforce "a card's assignee must be a board member" **in the schema**, as defense-in-depth on top of the two app-layer guards that already ship. The composite FK `cards(board_id, assigned_to) → board_members(board_id, user_id) ON DELETE SET NULL (assigned_to)` makes a dangling assignee structurally impossible and closes the check→write race. **It does NOT delete any app code** — guard #2's explicit `UPDATE … RETURNING` must stay because ④'s `CARD_UPDATED` broadcast needs the affected-card list, which the FK's silent null can't provide. See ADR-010 for the full rationale (including why this reverses the original 2026-07-15 "FK replaces the cascade" framing).

## Settled decisions (do not re-litigate — see ADR-010)

1. **Adopt the FK** as a DB invariant + backstop; keep guard #1 (write-time check) and guard #2 (participation cascade).
2. **Denormalize `board_id` onto `cards`**, set once at creation from the validated `column.BoardID`; no trigger (board is immutable).
3. **Reorder both cascades** — unassign-and-collect the cards **before** deleting the membership — or the FK silently empties the broadcast list.
4. **Migration cleanup**: auto-null pre-existing violators before `ADD CONSTRAINT`, `RAISE NOTICE` the count. No-op on clean DBs.
5. **Error mapping**: repo-layer translate `23503` on constraint `fk_cards_assignee_board_member` → existing `domain.ErrAssigneeNotBoardMember` (already 400). Other `23503`s stay 500.
6. Keep the existing `assigned_to → users` FK.

---

## Build plan (test-first)

### A. Migration `000009_add_cards_board_id_and_assignee_fk`

**up.sql** (order matters):
```sql
-- 1. add the denormalized column, nullable for backfill
ALTER TABLE cards ADD COLUMN board_id UUID;

-- 2. backfill from each card's column's board
UPDATE cards c SET board_id = col.board_id
  FROM columns col WHERE c.column_id = col.id;

-- 3. lock it down
ALTER TABLE cards ALTER COLUMN board_id SET NOT NULL;
ALTER TABLE cards
  ADD CONSTRAINT fk_cards_board FOREIGN KEY (board_id)
  REFERENCES boards(id) ON DELETE CASCADE;

-- 4. clean up any assignee that isn't a member of the card's board, BEFORE the lock
WITH violators AS (
  UPDATE cards c SET assigned_to = NULL
  WHERE c.assigned_to IS NOT NULL
    AND NOT EXISTS (
      SELECT 1 FROM board_members bm
      WHERE bm.board_id = c.board_id AND bm.user_id = c.assigned_to
    )
  RETURNING 1
)
-- surface the count in migration output
DO $$ ... RAISE NOTICE ... $$;   -- see note below on how to emit the count

-- 5. the invariant
ALTER TABLE cards
  ADD CONSTRAINT fk_cards_assignee_board_member
  FOREIGN KEY (board_id, assigned_to)
  REFERENCES board_members (board_id, user_id)
  ON DELETE SET NULL (assigned_to);
```
> Note: golang-migrate runs plain SQL; the cleanup UPDATE can be a bare statement, and the `RAISE NOTICE` count needs a `DO $$ ... $$` block or a `\echo`-free approach. Simplest: a `DO $$ DECLARE n int; BEGIN ... GET DIAGNOSTICS n = ROW_COUNT; RAISE NOTICE 'nulled % violating assignees', n; END $$;` wrapping the cleanup UPDATE. Pick whatever the harness's migrate step surfaces.

**down.sql:** drop the two constraints, drop the column. (The nulled assignments are not restorable — acceptable; they were invalid.)

Consider an index on `cards(board_id, assigned_to)` only if the FK check / cascade needs it — Postgres does **not** auto-index the referencing side of an FK. Assess against the harness; add `idx_cards_board_assignee` if cascade deletes are slow. (Likely fine at Phase-1 scale; note it, don't prematurely add.)

### B. Entity + create path (populate `board_id`)

- `entity.Card`: add `BoardID uuid.UUID`.
- `create_card.go:60`: set `BoardID: column.BoardID` on the constructed card (authoritative, already validated at line 27).
- `cardRepo.Create` (INSERT query): add `board_id` to the column list + values.
- **No change to `update_card.go` / `move_card.go`** for `board_id` — the board never changes on those paths.

### C. Reorder guard #2 (behavior-preserving)

- **Board path** — `internal/repository/postgres/board_member.go` `RemoveWithParticipationCascade` (~L168-207): run `unassignBoardCardsForUserQuery` (collect `[]AffectedCard`) **before** `deleteBoardMemberForCascadeQuery`. Keep the 0-rows-→`ErrBoardMemberNotFound` semantics on the delete.
- **Workspace path** — `internal/repository/postgres/workspace_member.go` `RemoveWithParticipationCascade` (~L169-230): run `unassignCardsForUserQuery` (collect cards) **before** `deleteBoardMembershipsForUserQuery`. The workspace_member row delete (L176) and the board-IDs delete (L185, `RETURNING bm.board_id`) still both run; only the *card* unassign moves ahead of the board-membership delete.
- **Watch:** the workspace path needs `AffectedBoardIDs` from the board-membership delete AND `AffectedCards` from the card unassign. Both are independent selects on `(workspace_id, user_id)`; reordering the card-unassign earlier doesn't change either result set. Verify with the harness (integration test #3 below).

### D. Error mapping (repo layer)

In `cardRepo.Create` and `cardRepo.Update` (`internal/repository/postgres/card.go`), catch the pg error:
```go
var pgErr *pgconn.PgError
if errors.As(err, &pgErr) &&
   pgErr.Code == "23503" &&
   pgErr.ConstraintName == "fk_cards_assignee_board_member" {
    return domain.ErrAssigneeNotBoardMember
}
```
No new sentinel, no new mapper entry — `ErrAssigneeNotBoardMember` already maps to 400 (`domain_mapper.go:45`). Mirror the `23505` handling in `workspace_member.go:41-46`.

---

## Tests

**Integration (ADR-015 harness, `tests/integration/`) — the point of doing this in ⑤:**
1. **Invariant on write:** insert a card with `assigned_to` = a non-member → FK rejects (`23503`, our constraint). Insert with a member → succeeds. Insert with `assigned_to = NULL` → succeeds (unconstrained).
2. **FK auto-cascade:** assign a member, delete their `board_members` row directly → `cards.assigned_to` becomes NULL (the DB doing it, no app code).
3. **Guard #2 still broadcasts after reorder:** call board + workspace `RemoveWithParticipationCascade` → returns the correct non-empty `[]AffectedCard` (this is the regression guard for the reorder — it would come back empty if the reorder were wrong).
4. **Migration cleanup:** seed a violating row (assignee not a member), run migrate up on a fresh container, assert the row's `assigned_to` is NULL and the FK installs.
5. Also fold in the other **deferred SQL tests** that this pass owns if convenient (rebalance, visibility filter, `card_count`) — or leave to the pass's final test step.

**Unit (mock-backed, `internal/usecase/...`) — mostly unchanged:**
- Existing card/board/workspace use-case tests still pass. `create_card` now sets `BoardID`; update any fixture that asserts the constructed card's fields.
- Guard #1 tests (`ErrAssigneeNotBoardMember` on non-member) are unchanged — the app check still runs first.

## Docs to update (after code — audit closure)

- **SRS §2.8 / §9.2**: note assignee-validity is now a **DB-enforced** invariant (composite FK), not only app-layer; strike any "TOCTOU accepted for Phase 1" caveat.
- Root `CLAUDE.md` › `## Now`: flip the ⑤ pointer — FK **built ✅**, move to Transactor/Option-C. (The queue + rationale lines were already corrected when ADR-010 was written; update the status.)
- `collabotask-backend/temp_unit-test-checklist.md`: check off the deferred board/workspace cascade-SQL cases now covered by integration tests #2/#3.
- Memory `post-phase1-integration-tests`: mark FK **done**, cascade-SQL tests **done** (or partial).

## Risks / gotchas

- **The reorder is the sharp edge.** Get integration test #3 red-first: with the FK installed and the *old* order, the broadcast list is empty. That test is what proves the reorder is correct.
- **Regenerate nothing structural** — no new repo interface methods (only INSERT column + query reorders + entity field). No Wire/mock regen expected, but run `go generate` + `go build` to be sure the entity field didn't ripple.
- **`down.sql` data loss** is acceptable (nulled invalid assignees), but call it out in the migration comment.
- **Index decision** — measure before adding `idx_cards_board_assignee`; don't add speculatively.

## Quick verification (from `collabotask-backend/`)
- `go build ./... && go vet ./...`
- `go test ./internal/...` (unit — fast)
- `go test ./tests/integration/...` (harness — Docker)

## Suggested skills
- **tdd** — integration test #3 (reorder regression) red→green first.
- **code-review** (two-axis) — after: verify the reorder in both cascades, the constraint-name-scoped error mapping, and that no membership gate leaked anywhere new.

---

## Post-implementation review — 2026-08-25

Two-axis `/code-review` of the committed implementation (`af30ed4`, range `af30ed4~1...af30ed4`).

**Verdict:** Standards axis clean (0 hard documented-standard violations); Spec axis faithful — every settled decision (1–6 above) is present and correct, the reorder is verified in **both** cascade paths, the constraint-name-scoped error mapping is on both `Create` and `Update`, and **both app guards are kept**. The implementation makes sense and solves the problem. Findings below are tagged by action; only F1 and F2 are follow-ups.

### F1 — [Action: add before closing ⑤] Migration-cleanup path is untested (Test #4 dropped)

**Finding.** Of the five tests in §Tests, #4 (migration cleanup) was not implemented. `up.sql` step 4 (null pre-existing violators + `RAISE NOTICE` the count) has **no coverage**, because the harness always migrates a *clean* DB — the violator-nulling branch never runs. Tests #1 (invariant on write), #2 (FK auto-cascade), #3 (board+workspace broadcast-list after reorder) all exist, plus an extra `TestAssigneeFKInvariant_ConstraintName` that protects the error mapping (justified, not scope creep).

Not a correctness bug — the branch only fires on legacy pre-step-③ dirty data, and nulling an unreachable assignee is the correct end state — but it is the one real gap.

**Solution.** Add integration test #4: on a fresh container, seed a card whose `assigned_to` is a non-member of its board **before** migration 000009 runs, migrate up, then assert (a) that row's `assigned_to` is now `NULL` and (b) the composite FK installed successfully. Add before formally closing the FK step.

### F2 — [Action: apply now] Redundant explicit `rows.Close()` — YAGNI / Speculative Generality

**Finding.** Both cascades pair a `defer <rows>.Close()` (needed — closes the rows on the early error-return paths) with a **second, explicit** `<rows>.Close()` right before the next statement on the same `tx`:
- `board_member.go:197` — before `tx.Exec(deleteBoardMemberForCascadeQuery, …)`
- `workspace_member.go:205` — before `tx.Query(deleteBoardMembershipsForUserQuery, …)`

The explicit close is **functionally redundant today**: a pgx transaction runs on one connection with one active query at a time, and `rows.Next()` auto-closes the rows when the loop drains (returns `false`), freeing the connection before the next query. Neither loop has an early `break`/mid-iteration `return`, so auto-close always fires. The only justification for the explicit call is defensiveness against a *future* edit that adds a `break` (which would skip auto-close and make the next query fail with `conn busy`) — a hypothetical need, i.e. **YAGNI**. If that edit ever lands, it should carry its own `rows.Close()`.

Note `boardRows` in the same workspace function (`workspace_member.go:212`) already uses the **defer-only** pattern — so dropping the explicit closes makes all three row-iterations consistent.

**Solution (decided).** Drop the explicit `<rows>.Close()` at `board_member.go:197` and `workspace_member.go:205`; keep the `defer`. (Rejected alternative: keep it *with* a one-line comment documenting pgx's one-active-query-per-connection contract — a comment is what would earn a redundant line its place, but for now we drop rather than future-proof.)

### F3 — [Optional] Duplicated scan-into-`[]AffectedCard` loop

The `for rows.Next(){ Scan; append }` + `rows.Err()` shape is near-identical in `board_member.go` and `workspace_member.go`, differing only by the `BoardID` source (param vs scanned column). Pre-existing shape relocated by the reorder, not introduced here. Optional: extract a shared `scanAffectedCards` helper. Low value — fold into a future cascade-touching change (e.g. the ADR-007 Transactor work, which composes this code).

### F4 — [Optional] Duplicated FK-error mapping

The `errors.As(&pgErr) && pgErr.Code == "23503" && pgErr.ConstraintName == "fk_cards_assignee_board_member"` block repeats in `cardRepo.Create` and `cardRepo.Update`. Optional: a small `mapAssigneeFKError(err) error` helper. Two trivial sites; borderline — leave unless it grows a third caller.

### Action summary

| # | Action | When |
|---|---|---|
| F1 | Add integration test #4 (migration violator-cleanup) | Before closing the FK step |
| F2 | Drop the redundant explicit `rows.Close()` in both cascades (`board_member.go:197`, `workspace_member.go:205`); keep the `defer` | Now |
| F3 | Extract shared `scanAffectedCards` helper | Optional / future |
| F4 | Extract `mapAssigneeFKError` helper | Optional / future |

Code changes are **not yet applied** — this section records the findings and the agreed solutions.
