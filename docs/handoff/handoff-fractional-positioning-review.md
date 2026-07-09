# Handoff — Fractional NUMERIC positioning: act on code-review findings

**Date:** 2026-07-08
**Repo:** `collabotask` (backend: `collabotask-backend/`)
**Branch:** `main` · changes are **uncommitted in the working tree** (not staged/committed)

## What just happened
A code review (`/code-review`, xhigh effort) was run against the in-progress
**fractional NUMERIC positioning** change (build-order step ① in root `CLAUDE.md` › `## Now`).
The implementation was judged faithful to the ADR overall; build, `go vet`, and 97 use-case
tests pass. The review produced 5 findings + 2 minor notes. **No fixes have been applied yet.**
This session ended after producing a findings summary for the implementer.

## Next session focus
Address the review findings below. Per project convention (root `CLAUDE.md` › "Working With Me"):
**non-trivial changes → propose a short plan and wait for approval before coding.** Findings #1
and #5 are the clear-cut fixes; #2 is a test gap; #3/#4 are acceptable-to-document.

## Context / reference (do not re-read unless needed)
- Decision: `docs/architecture/adr/adr-004-fractional-positioning.md`
- Contract: `docs/spesifications/001-software-specifications.md` §3.2, UC-13d, UC-15
- Backend conventions & recipes: `collabotask-backend/CLAUDE.md`
- The change under review (working tree): `git diff HEAD` + new untracked files:
  - `internal/repository/postgres/positioning.go` (new — rebalance helper)
  - `migrations/000006_fractional_positioning.{up,down}.sql` (new)

## Findings to action

### 1. Stale returned position after a rebalance — FIX (correctness, present now)
- Files: `internal/repository/postgres/card.go:228-255` (`Move`),
  `internal/repository/postgres/column.go:199-231` (`UpdatePosition`),
  `internal/usecase/column/update_column_position.go:36`.
- `moved`/`col` is scanned with the requested `position`, then `rebalanceIfNeeded` may rewrite
  that row, but the returned entity's `Position` is never refreshed (and the use case overwrites
  the column with `input.Position`). Response returns a position the DB no longer holds when a
  rebalance fires.
- Repro: neighbors `1000` & `1000.4`, drop at `1000.2` → gap `0.2 < 0.5` → rebalance rewrites
  column to `1000, 2000, 3000…`; response still returns `1000.2`.
- Fix direction: have `rebalanceIfNeeded` return the moved item's new position (or re-select the
  row post-rebalance) and propagate it into the returned entity in both paths.
- Note: ADR deferred only the *bulk* "many rows changed" broadcast, not the moved row's own value.

### 2. Repository positioning logic has zero test coverage — ADD TESTS
- `internal/repository/postgres/positioning.go` has no tests (no `*_test.go` in that package).
- Untested: ±Inf open-end handling, `min(gap) < threshold` decision, tie-break neighbor selection,
  whole-partition `ROW_NUMBER()` rewrite. Use-case tests only mock the repo.
- Fix direction: DB-backed tests (testcontainers / test DB). Cover: no-rebalance path; rebalance on
  tight gap; head/tail inserts (no neighbor → ±Inf); exact-tie; post-rebalance spacing + order.

### 3. Append path allows a permanent position tie under concurrency — DOCUMENT or optional fix
- `internal/usecase/card/create_card.go:53-67`, `internal/usecase/column/create_column.go:22-32`.
- Non-atomic `GetMaxPosition` → `Create(max + STEP)`; concurrent creates collide at `max+1000`.
  Order stays correct via `id` tiebreak, but create never rebalances so the tie persists.

### 4. Potential deadlock on concurrent same-column rebalances — DOCUMENT (rare, recoverable)
- `internal/repository/postgres/card.go` `Move`: holds `FOR UPDATE` on its card, then rebalance
  does `UPDATE cards … WHERE column_id=$2` over the whole partition. Two simultaneous threshold-
  crossing moves in one column can deadlock (`40P01`); Postgres aborts one → 500, retry-recoverable.

### 5. `positionStep = 1000` duplicated 3× + migration literal — FIX (altitude/drift)
- `internal/usecase/card/create_card.go:14`, `internal/usecase/column/create_column.go:10`,
  `internal/repository/postgres/positioning.go:14`, and hardcoded `1000` in
  `migrations/000006_fractional_positioning.up.sql`.
- ADR-004 called for a single domain-level `STEP` (+ `RebalanceThreshold`). Seed spacing (use-case)
  and rebalance spacing (repository) MUST stay equal; scattering invites silent drift.
- Fix direction: define `STEP`/`RebalanceThreshold` once at the domain level; reference everywhere.
  (Migration literal can't import Go consts — keep it, but add a comment tying it to the domain STEP.)

### Minor notes (no action required)
- Rebalance runs two neighbor `SELECT`s on every move regardless of proximity to threshold.
- Neighbor lookups use the `column_id`-only index (`idx_card_column_id`) + in-memory sort; a
  composite `(column_id, position)` index would serve them directly. Fine at Phase-1 scale.

## Verified-correct (don't re-investigate)
- `*float64` presence-required validation (0 and negatives accepted; NaN/±Inf rejected by JSON decoder).
- `ORDER BY position ASC, id ASC` deterministic tiebreak in list queries.
- Migration: `cards.position` never had a CHECK; `columns_position_check` correctly dropped on `up` /
  restored on `down`; reseed via `ROW_NUMBER()` is order-preserving and self-healing.
- Integer-shift methods fully removed (`IncrementPositionsFrom`, `DecrementPositionsAfter`,
  `ReorderPositions`, `DeleteWithReorder`, the `max+1` clamp) — no dangling refs; build/vet/tests green.

## Suggested skills
- **tdd** — for finding #2: write the DB-backed rebalance tests test-first (red → green).
- **code-review** — re-run after applying #1/#5 to confirm no regressions in the touched paths.
- **simplify** — optional, when consolidating the `positionStep` constant (#5) to keep the refactor tight.

## Quick verification commands (from `collabotask-backend/`)
- `go build ./... && go vet ./...`
- `go test ./internal/usecase/card/... ./internal/usecase/column/...`
