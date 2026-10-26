# Handoff — step ⑤ deferred SQL tests

**Branch:** `test/deferred-sql-tests` (exists, pushed, even with `main`)
**Date:** 2026-10-26
**Scope:** write the six remaining repository-layer integration tests of build step ⑤. No production code changes expected.

Read first: `collabotask-backend/TESTING.md` (test approach), `docs/architecture/adr/adr-015-integration-test-harness.md` (harness design), `docs/architecture/adr/adr-016-transactor-atomic-activity-writes.md` (atomicity contract).

---

## State of the branch

Clean tree, one commit ahead of `main`: `d25b4a1 docs: correct stale WriteActivity comments, trim ⑤ carry-over` — three comment rewordings plus a `## Now` correction in `CLAUDE.md`. Docs-only, `go build ./...` passes, no interaction with the test work. **Do not commit without the user asking** — this project never auto-commits; leave work staged for review.

## What is already verified (do not redo)

The testcontainers suite was run end to end for the first time on 2026-10-26 and is sound:

- 13 tests pass against a real `postgres:16-alpine` (PG 16.12), database `collabotask_test`, migrations applied to version 9, `dirty=false`.
- No `t.Skip`, no env guards, no `-short` bail-outs — a pass is a real pass.
- Verified independently via Docker container events and a throwaway row dump.

So if a new test fails, **suspect the test or the code, not the harness.**

## Running the tests

Docker must be running (`docker version` should answer). Then:

```bash
cd collabotask-backend
go test ./tests/integration/... -v -count=1
```

First run pulls the Postgres image. The harness starts its own container on a random port, runs every migration, and terminates the container when the run ends. There is **no env var to point it at an existing database** — don't look for one, and don't add one.

## Harness API

In `package integration_test`, `TestMain` (in `tests/integration/board_repo_test.go`) provides two package-level values:

- `testPool *pgxpool.Pool` — the live pool
- `fx *testutil.Fixtures` — fixtures

Every test must start with:

```go
t.Cleanup(func() { testutil.TruncateAll(t, testPool) })
```

Fixture signatures (`tests/integration/testutil/fixtures.go`):

```go
fx.CreateTestUser(t)                                  *entity.User
fx.CreateTestWorkspace(t, ownerID)                    *entity.Workspace
fx.CreateTestBoard(t, workspaceID, ownerID)           *entity.Board
fx.CreateTestColumn(t, boardID)                       *entity.Column
fx.CreateTestCard(t, columnID, boardID, createdBy)    *entity.Card
fx.AddBoardMember(t, boardID, userID)
```

For migration-level behaviour there is `testutil.NewTestDBAtVersion(t, targetVersion)`, which starts a *separate* container migrated only to `targetVersion` and hands back the migrate handle so you can seed pre-migration state and then run `m.Up()`. Used by the existing migration-cleanup test.

### ⚠️ Fixture gotcha

**`fx.CreateTestBoard` seeds the three default columns — `To Do`, `In Progress`, `Done`.** A fixture board is not empty. A board plus one `fx.CreateTestColumn` yields **4** column rows, not 1. This will bite the rebalance and `card_count` tests specifically. Assert on relative counts or on specific ids, not on "the board has 1 column".

## Follow the existing shape

Mirror `tests/integration/transactor_test.go` and `tests/integration/assignee_fk_test.go`. The established pattern for atomicity tests: open `tx.WithinTransaction`, drive the **repository** calls inside it, force a failure, then assert the first write is gone. These tests compose repositories directly rather than driving the full use case — stay with that, don't restructure it.

The trick the existing test uses to force a realistic failure: write an activity with a **non-existent `board_id`**, which trips a foreign-key violation inside `activities`, making `Log` fail. Reuse it.

---

## The six tests

### 1 & 2 — atomicity variants (highest priority)

These close the last open carry-over item from the Transactor review (ADR-016, PR #14). ADR-016 made all 17 activity writes atomic with their mutation, but only **one** shape is proven at integration level — `TestTransactor_LogFailureRollsMutationBack`, which covers `CreateCard`. The other two transaction shapes in the codebase are unit-tested only.

**1. `TransferOwnership`** — [transfer_ownership.go:46-62](../../collabotask-backend/internal/usecase/board/transfer_ownership.go#L46-L62). Inside `WithinTransaction`: `boardMemberRepo.TransferOwnership` then `LogActivity`. Prove that when the activity write fails, **ownership does not move** — the old owner is still `OWNER` and the target is still `MEMBER`. Note `TransferOwnership` returns `fromUserID`, which the broadcast depends on; worth asserting it comes back correct on the success path too.

**2. break-glass `self_join_board`** — [self_join_board.go:67-91](../../collabotask-backend/internal/usecase/board/self_join_board.go#L67-L91). Inside `WithinTransaction`: `boardMemberRepo.CreateIfAbsent` then, only when `inserted`, `LogActivity` carrying `break_glass: true` (a workspace admin joining a `PRIVATE` board). Two cases worth having:
- activity write fails → the `board_members` row is **not** left behind
- joining twice → `CreateIfAbsent` reports `inserted=false` via `ON CONFLICT DO NOTHING RETURNING`, and **no second activity row** is written

### 3 — position rebalance

`rebalanceIfNeeded` in [positioning.go:21](../../collabotask-backend/internal/repository/postgres/positioning.go#L21) (ADR-004, migration `000006_fractional_positioning`). Constants live in [internal/domain/positioning.go](../../collabotask-backend/internal/domain/positioning.go): `PositionStep = 1000`, `PositionRebalanceThreshold = 0.5`.

It inspects the two neighbour gaps around a just-moved item; if the smaller gap drops below the threshold it rewrites the **whole partition** to evenly spaced `PositionStep` multiples inside the same transaction, then re-reads and returns the moved item's new position.

Two partitions use it, test both:
- cards, partitioned by `column_id` — [card.go:264](../../collabotask-backend/internal/repository/postgres/card.go#L264)
- columns, partitioned by `board_id` — [column.go:214](../../collabotask-backend/internal/repository/postgres/column.go#L214)

Cases: pack neighbours so the gap falls under 0.5 and assert the partition comes back as clean `STEP` multiples **in the original order**; assert the returned position equals what the database actually holds (this was a real bug — the use case must not report the requested position); assert a comfortable gap does **not** trigger a rewrite; and check an open end (no neighbour → `±Inf`, no rebalance).

### 4 & 5 — visibility filter and `card_count`

Both live in one query, `getUserBoardsInWorkspace` — [board_queries.go:30](../../collabotask-backend/internal/repository/postgres/board_queries.go#L30), called from [board.go:183](../../collabotask-backend/internal/repository/postgres/board.go#L183). Per ADR-005, `board_members` is the sole source of access truth; `created_by` is deliberately not consulted.

**Visibility cases.** A board appears if the requester is a workspace `ADMIN`, or the board is `WORKSPACE`-visible, or they are already a member:
- workspace admin sees a `PRIVATE` board they have not joined, with `access_status = CAN_JOIN` and `user_role = NULL`
- plain member does **not** see a `PRIVATE` board they have not joined — it must be absent entirely, not merely flagged
- plain member sees a `WORKSPACE` board as `CAN_JOIN`; once joined, `JOINED` with their real role
- archived boards are excluded (`is_archived = FALSE`)
- a user who is not a workspace member gets nothing (the `INNER JOIN workspace_members` enforces this)

**`card_count` — the one that matters most.** The query stacks four `LEFT JOIN`s (`bm`, `bm2`, `columns`, `cards`) and leans entirely on `COUNT(DISTINCT …)` to stop them multiplying each other. Build a board with **several members and several cards at once** — say 3 members and 2 cards — and assert `member_count = 3` and `card_count = 2`. Without the `DISTINCT`s this returns 6 for both. A test with one member or one card passes either way and proves nothing, so make sure both counts are greater than one in the same fixture. Also cover: a board with columns but no cards → `card_count = 0`; cards spread across several columns are summed once each.

### 6 — `ErrBoardMemberNotFound` on a 0-row delete

`board_member.go` returns `domain.ErrBoardMemberNotFound` at [lines 82, 137, 185 and 210](../../collabotask-backend/internal/repository/postgres/board_member.go#L82); the cascade query is documented at [board_member_queries.go:53](../../collabotask-backend/internal/repository/postgres/board_member_queries.go#L53). Assert that removing a user who is **not** a member of the board yields the sentinel rather than a silent success — the use cases at `leave_board.go:32` and `remove_member.go:54` branch on exactly that sentinel, so it is load-bearing.

---

## Conventions to respect

- **Never unit-test the repository layer** (`internal/repository/postgres/`) — integration only. Conversely, the use-case layer is mock-tested and should not grow database tests.
- Build the tests **test-first**, one at a time. The `/tdd` skill is the house style for this.
- Finish with the `/code-review` skill, which checks the diff against both repo standards and the originating spec.
- `-race` is expected to stay clean.
- **Do not commit** unless the user asks. Leave work staged for review.

## Out of scope

Phase-2 deferrals, confirmed untouchable: comments, labels, attachments, public boards, refresh tokens, account deletion. Also don't reopen step ④ (WebSocket) or ADR-016 itself.

Four files already fail `gofmt` on `main` — `workspace_response.go`, `activity.go`, `sentinel_errors.go`, `transfer_ownership_test.go`. Pre-existing and unrelated; `/code-review` may flag them. Leave them to a separate branch rather than mixing a formatting sweep into this one.

## Done means

Six tests written and passing, suite green end to end with `-race`, `docs/build-log.md` updated with a step-⑤ entry, and `## Now` in `CLAUDE.md` moved on to the next step. Step ⑤ is the last thing before frontend implementation begins.
