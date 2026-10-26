# Handoff — step ⑤ integrity pass: findings to action

**Date:** 2026-10-26
**Repo:** `collabotask` (backend: `collabotask-backend/`)
**Branch:** `test/deferred-sql-tests` · HEAD `38b11a6 test(integrity-pass): add the six deferred repository-layer SQL tests`

## What just happened

The six deferred repository-layer integration tests of build step ⑤ were written test-first
(`docs/handoff/handoff-deferred-sql-tests.md`) and committed as `38b11a6` — 16 test functions,
tests only, no production code touched. Writing them surfaced **two defects neither of which
belongs to that branch**, plus one housekeeping note. **No fixes have been applied.**

Finding #1 is a product defect that exists in `main` today. Finding #2 is why
`go test ./... -race` fails intermittently — it is a test bug, not a product one. Neither
blocks step ⑤ from closing.

## Verified clean (do not re-investigate)

- `tests/integration` — 26 tests, green, **5 consecutive `-race` runs**, no race.
- Whole backend suite green without `-race`; `go build ./...` clean.
- `38b11a6` touches only the five files under `tests/integration/`; nothing in `internal/`.
- Each new test was checked by deliberately breaking the behaviour it asserts and confirming
  it failed. The strongest: removing both `DISTINCT`s from `getUserBoardsInWorkspace` made the
  board list report **15 members and 6 cards** where the fixture has 3 and 2.

---

## Findings to action

### 1. New boards get default columns at positions 0, 1, 2 instead of STEP multiples — FIX (correctness, present now, low severity)

- File: `internal/repository/postgres/board.go:81-88` (`CreateWithOwner`).
- The loop index is written in as the position:

  ```go
  for position, colTitle := range domain.DefaultNewBoardColumnTitles {
      _, err = br.exec(ctx).Exec(ctx, createColumnQuery, board.ID, colTitle, position)
  ```

  `position` is Go's range counter over the three titles — `0`, `1`, `2`.
- Three other places all space by `domain.PositionStep` (1000):
  - `migrations/000006_fractional_positioning.up.sql` reseeds every existing column and card to
    `rn * 1000`, with a comment stating the wide gaps exist "so the initial gaps are wide enough
    for many inserts before a rebalance is triggered";
  - `internal/usecase/column/create_column.go:29` — `maxPos + domain.PositionStep`;
  - `internal/usecase/card/create_card.go:65` — same for cards.

  So a board created **before** migration 000006 has columns at 1000/2000/3000; a board created
  **after** it gets 0/1/2. This loop is the one seeding path that was not updated when ADR-004
  landed.
- Observed (throwaway probe against the real PG16 container, since deleted):

  ```
  default column "To Do"       position=0
  default column "In Progress" position=1
  default column "Done"        position=2
  next appended column would get position=1002  (maxPos 2 + STEP 1000)
  ```

- Impact: the gap available for a drop between two columns is 1 instead of 1000. The drop
  position is supplied by the client (`internal/usecase/column/update_column_position.go` passes
  `input.Position` through), and the natural value is the midpoint between neighbours, so on a
  fresh board:
  - first drag between `To Do` (0) and `In Progress` (1) → `0.5`; the minimum gap is exactly
    `PositionRebalanceThreshold`, and the check is `>=`, so **no** rebalance;
  - the next drag into that slot → `0.25` → full-partition rebalance.

  ADR-004's intent is dozens of drags before a rebalance ever fires. Nothing breaks and it is
  self-healing — after the first rebalance the columns sit at 1000/2000/3000 and behave normally.
  Cosmetically, until then a user-added column sits at 1002 beside defaults at 0, 1, 2.
  No data corruption, no wrong ordering, nothing user-visible.
- Fix direction: one line — space the defaults like everything else.

  ```go
  float64(position+1) * domain.PositionStep
  ```

- **Test impact:** `tests/integration/positioning_test.go` ›
  `TestColumnUpdatePosition_RebalancesWhenGapTooSmall` deliberately documents today's `0, 1, 2`
  behaviour (it drags `Done` to `0.2` to pack the gap). Fixing this means updating that test's
  drag target — e.g. to `1000.2` — so it still lands under the threshold. Update both together.
- Frontend note: the app zone is not built yet, so no client depends on the current spacing.

### 2. Flaky data race in `TestSetArchived` — FIX (test-only; breaks `go test ./... -race`)

- Files: `internal/usecase/board/set_archived_test.go:23, 73, 82, 106, 116, 127, 141`;
  the racing write is `internal/usecase/board/set_archived.go:47`.
- Mechanism: one mutable fixture shared across parallel subtests.
  - line 23 — `existingBoard := &entity.Board{…}`, created once for the whole table;
  - every subtest runs `t.Parallel()`;
  - lines 73, 82, 106, 116 hand **that same pointer** to the use case:
    `d.boardRepo.EXPECT().GetByID(…).Return(existingBoard, nil)`;
  - the use case writes to the board it was handed — `set_archived.go:47`,
    `board.IsArchived = *input.IsArchived`;
  - lines 127 and 141 read the same struct to copy it — `b := *existingBoard`.

  One subtest writes `IsArchived` on the shared fixture while another reads it. The race report
  names exactly this pair: write at `set_archived.go:47`, previous read at
  `set_archived_test.go:141`.
- Reproduction: `go test ./internal/usecase/board/ -race -count=1 -run TestSetArchived`, run
  repeatedly. Measured **3 races in 8 runs**. A single clean run proves nothing; loop it.
- Not a product defect. Production never shares one board struct across concurrent requests —
  each request fetches its own. `set_archived.go` is correct; only the fixture sharing is wrong.
- Fix direction: give each subtest its own copy, the way lines 127 and 141 already do. A factory
  beside the fixture keeps it to four edits:

  ```go
  newExistingBoard := func() *entity.Board {
      b := *existingBoard
      return &b
  }
  ```

  …then the four `Return(existingBoard, nil)` sites become `Return(newExistingBoard(), nil)`.
  Dropping `t.Parallel()` would also silence it, but that hides the sharing instead of fixing it.
- **Sweep while in there:** a shared fixture pointer plus `t.Parallel()` is an easy pattern to
  repeat across the sibling use-case tests. The others may simply be getting lucky — grep for
  `t.Parallel()` alongside a shared `&entity.` fixture and confirm with a looped `-race` run.
- Verify the fix with 20–30 looped runs, not one.

### 3. Five files fail `gofmt` — HOUSEKEEPING (pre-existing, unrelated)

```
internal/delivery/http/response/workspace_response.go
internal/domain/entity/activity.go
internal/domain/sentinel_errors.go
internal/usecase/board/transfer_ownership_test.go
tests/integration/testutil/harness.go
```

The first four are named in `handoff-deferred-sql-tests.md`; `harness.go` is a fifth it did not
list. All pre-existing on `main`, all untouched by `38b11a6`. Keep the formatting sweep on its own
branch rather than mixing it into feature or test work.

---

## Suggested order

1. **#2 first** — it is what makes `go test ./... -race` fail, so it is blocking anyone who runs
   the full suite with the race detector. Test-only, four edits, no product risk.
2. **#1 next**, together with its one-line test update. Small and safe, but it is product code,
   so per root `CLAUDE.md` › "Working With Me" it wants a short plan and approval first.
3. **#3 whenever** — its own branch.

Neither #1 nor #2 reopens ADR-016, step ④, or any Phase-2 deferral.

## Still open on step ⑤ itself

- `/code-review` on `38b11a6` (two-axis, per project convention).
- `docs/build-log.md` — append the step-⑤ entry.
- `CLAUDE.md` › `## Now` — move the pointer on to frontend implementation.
