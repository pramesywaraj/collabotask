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

**Status 2026-11-02:** #1 and #2 are both **fixed and verified** — do not redo either. #3 is
still open, still housekeeping. The three step-⑤ follow-ups at the bottom of this doc are all
done. The branch is still unmerged.

## Verified clean (do not re-investigate)

- `tests/integration` — 26 tests, green, **5 consecutive `-race` runs**, no race.
- Whole backend suite green without `-race`; `go build ./...` clean.
- `38b11a6` touches only the five files under `tests/integration/`; nothing in `internal/`.
- Each new test was checked by deliberately breaking the behaviour it asserts and confirming
  it failed. The strongest: removing both `DISTINCT`s from `getUserBoardsInWorkspace` made the
  board list report **15 members and 6 cards** where the fixture has 3 and 2.

---

## Findings to action

### 1. New boards get default columns at positions 0, 1, 2 instead of STEP multiples — ✅ FIXED 2026-11-02

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
- **Applied 2026-11-02** — one line, as directed:

  ```go
  float64(position+1) * domain.PositionStep
  ```

  New boards now seed at 1000/2000/3000, matching `create_column.go`, `create_card.go` and
  migration 000006's `rn * 1000`. No migration needed: boards predating 000006 are already on
  STEP multiples, and any created since self-heal on their first rebalance.

  **One extra change the plan did not anticipate.** `testutil.CreateTestColumn` hard-coded
  `Position: 1000`, which after the fix collides exactly with the `To Do` default. It now seeds
  at `(len(DefaultNewBoardColumnTitles)+1) * PositionStep` — i.e. one STEP past the last
  default, the position `create_column.go` would really append it at. Eleven call sites across
  four test files use that fixture; none assert the column's own position, so nothing else
  moved. This is the fixture-wide risk the verification note warned about, and it is why the
  whole suite had to run rather than just `positioning_test.go`.

  **Verified:** 31 integration tests pass (same count as before — none lost or skipped), and
  the full suite is green three runs in a row with `-race`: 632 tests, 31 packages, zero
  `DATA RACE` reports.

- **Test impact — three tests, not one.** Re-checked 2026-10-27 against `242a34d`, which
  expanded `positioning_test.go` by 95 lines after this finding was first written. All three
  column cases in `tests/integration/positioning_test.go` hard-code today's `0, 1, 2` spacing,
  in their assertions *and* their comments. They were written against the defect knowingly, so
  fixing it moves all three together:
  - `TestColumnUpdatePosition_RebalancesWhenGapTooSmall` — drags `Done` to `0.2`. Retarget to
    `1000.2` so the gap still falls under `PositionRebalanceThreshold`; the
    `2*domain.PositionStep` assertion still holds.
  - `TestColumnUpdatePosition_NoRebalanceOnComfortableGap` — asserts
    `{0, 1, 500, domain.PositionStep}`, and seeds a **fourth** column at `STEP` purely to
    manufacture a gap wide enough to drop into, because the defaults are too packed to land
    between. With 1000-spaced defaults that workaround stops being necessary, so this test gets
    *simpler* — drop the extra column and drop into the real gap.
  - `TestColumnUpdatePosition_NoRebalanceAtOpenEnd` — asserts `{1, 2, 5000}`, then
    `{0.25, 2, 5000}` after the second leg.

  Comments move with them: around lines 153 and 186 they state "three default columns at
  positions 0, 1, 2" and "sit one apart at 0, 1, 2 — too packed to drop anything between them".
  The second sentence stops being true after the fix.
- **Verification needs Docker.** All three are integration tests, so nothing can be confirmed
  without the testcontainers harness (`docker version` must answer, or `TestMain` panics in
  `NewTestDB`). Beyond these three: changing the seeded positions changes what **every**
  `fx.CreateTestBoard` call produces, and all 31 integration tests build their boards through
  that fixture — so run the whole suite, not just `positioning_test.go`, then
  `go test ./... -race`.
- Frontend note: the app zone is not built yet, so no client depends on the current spacing.

### 2. Flaky data race in `TestSetArchived` — ✅ FIXED 2026-10-27 (test-only)

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
- Reproduction (before the fix): `go test ./internal/usecase/board/ -race -count=1 -run
  TestSetArchived`, run repeatedly. Measured **3 races in 8 runs** when first found, then **5 in
  12** and **4 of 4** whole-tree runs a day later — it got more reliable, not less. A single
  clean run proves nothing; loop it.
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
- **Sweep — done, nothing else found.** The worry was that sibling tests share the same
  pattern and are merely getting lucky. All 32 test files under `internal/usecase/` that call
  `t.Parallel()` were run six times over with `-race`: the only site that ever appeared was
  `set_archived.go:47`. Six runs is not proof of absence, but it was enough to not spend a day
  auditing 32 files.
- **Applied 2026-10-27**, as described above: a `newExistingBoard()` factory beside the fixture,
  the four sharing sites switched to it, and the two subtests that were already hand-copying
  (`b := *existingBoard`) folded onto the same factory so the file speaks one idiom.
  **Verified 30/30** on `TestSetArchived` and **8/8** across `internal/`, zero `DATA RACE`
  reports — against 4-of-4 failing beforehand.

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

1. ~~**#2 first**~~ — done 2026-10-27, see above.
2. ~~**#1**~~ — done 2026-11-02, see above. Approved as a plan first, then applied and
   verified against a real container.
3. **#3 whenever** — its own branch.

Neither #1 nor #2 reopens ADR-016, step ④, or any Phase-2 deferral.

## Step ⑤ follow-ups — all done

- ✅ `/code-review` ran on `38b11a6`; its findings produced `242a34d` (the missing
  column-rebalance cases).
- ✅ `docs/build-log.md` — step-⑤ entry appended, summary table row flipped to ✅.
- ✅ `CLAUDE.md` › `## Now` — moved on to frontend, and now carries #1 and #3 as the two open
  backend items rather than claiming there are none.

One thing to know if you read the first version of those docs: they stated the suite was
"`-race` clean" while #2 was still live and failing every run. That claim has been narrowed to
name which half was verified and when. Worth not reintroducing — the branch is still unmerged,
so "Phase 1 backend is complete" is true locally and not yet on `main`.
