# ADR-016: Transactor / Unit-of-Work — Atomic Activity Writes

- **Status:** Accepted
- **Date:** 2026-08-31
- **Scope:** Discharges the **Option C** deferral parked by [ADR-007](adr-007-activity-logging-writes.md) ("Atomicity via Unit-of-Work / Transactor → integrity pass ⑤", [adr-007:64](adr-007-activity-logging-writes.md)). Supersedes ADR-007's **write model** only (best-effort, after-commit → transactional). Everything else ADR-007 decided — the firing rule (log on state change), the verb×entity vocabulary, `entity_id` semantics, the metadata snapshot principle, board-scoped-with-per-board-cascade-rows — is **unchanged**; this ADR governs *when* the activity row commits relative to its mutation, not *what* the row contains. Part of build-order step ⑤ (post-Phase-1 integrity pass), sequenced **after** the integration harness ([ADR-015](adr-015-integration-test-harness.md)) and the composite-FK invariant ([ADR-010](adr-010-composite-fk-assignee-invariant.md)).

## Context

ADR-007 shipped Phase-1 activity logging as **Option A**: after a mutation commits, the usecase calls `activityRepo.Log(ctx, …)` as a separate round-trip and, on error, logs-and-swallows — a logging failure must never fail the user's mutation ([common/activity.go:15-24](../../collabotask-backend/internal/usecase/common/activity.go)). This was explicitly a *placeholder for the correct end-state*, accepted because (a) the codebase had no shared transaction abstraction, and (b) there was no DB harness to verify any atomicity guarantee. Both blockers are now cleared: the composite-FK step reworked the cascade tx code that a Transactor would wrap ([ADR-010](adr-010-composite-fk-assignee-invariant.md)), and the integration harness stands up a real Postgres ([ADR-015](adr-015-integration-test-harness.md)) so an atomic rollback can actually be tested.

The gap Option A accepts is real: a mutation can commit while its activity row is lost in the ~ms window after commit (process death, or an independent `activities` INSERT failure). It is logged, not silent — but the two audit-sensitive rows (break-glass admin-join per §2.3, and `OWNERSHIP_TRANSFERRED` per ADR-006) are exactly the ones whose loss stings.

The load-bearing fact, restated from ADR-007 and re-verified against the code: **each repository owns a `*pgxpool.Pool` directly**; 52 call sites run statements straight on the pool, and **9 hand-rolled transaction methods** (`card.Move`, `boardMember.TransferOwnership`, `boardMember.RemoveWithParticipationCascade`, `workspaceMember.RemoveWithParticipationCascade`, `boardMember.CreateMany`, `column.CreateMany`, `column.UpdatePosition`, `board.CreateWithOwner`, `workspace.CreateWithOwner`) each open their own `db.Begin(ctx)` inside a single method. There is no way today for a usecase to compose "mutation + activity INSERT" in one transaction. This ADR builds that composition.

## Options Considered

### Q1 — Atomicity scope: strict-for-all vs. split (high-stakes only)

- **Split — strict only for `TransferOwnership` + break-glass join, best-effort for the other 15 (rejected).** Protects routine card/column mutations from ever failing because of an audit-log hiccup. But it entrenches a two-speed codebase permanently, and the line it draws — "was the mutation multi-statement?" — is an implementation detail, not a policy. Worse, the high-stakes paths *are* the multi-statement cascade paths, so the split adds complexity precisely where it least helps.
- **Strict for all 17 call sites (chosen).** Every mutation + its `Log` share one transaction. The feared failure — an `activities` INSERT failing while the mutation INSERT succeeds — requires the *same Postgres, same connection* to selectively reject one INSERT and accept another, i.e. a genuine DB fault (full disk, OOM, hardware) under which the mutation is already unreliable. The "audit tail wags the mutation dog" concern is valid for an *external* audit sink (Kafka, second DB, HTTP) whose availability is independent — it does not hold for a same-DB, same-connection INSERT. Uniform, one pattern, no policy seam to maintain.

### Q2 — Nested-transaction handling: how the 9 hand-rolled methods compose

Strict-for-all forces every one of the 17 usecases to wrap `mutation + Log` in one transaction. Trivial for single-statement mutations (`CreateCard`); the problem is the 9 methods that **already** own a transaction. `TransferOwnership` opens `bmr.db.Begin(ctx)` on a fresh pooled connection and **commits before the usecase reaches the `Log`** — so an outer usecase transaction and the method's inner transaction run on *different connections* and are never atomic with each other. Those 9 methods must stop owning their connection and run on whatever transaction the usecase provides.

- **(A) Ambient-tx, repo self-wraps ("join-or-begin") — chosen.** Repos resolve their executor from context (`exec(ctx)`): the ambient tx if one exists, else the pool. The 9 methods swap `db.Begin()` for a `joinOrBegin` helper — *if a tx is in the context, run the body on it; else open one, run, commit.* Each method is atomic both standalone and composed; no caller must remember to wrap.
- **(B) Push all Begin/Commit up to the usecase (rejected).** The 9 methods become pure executor-consumers that never begin; every standalone caller must wrap them in `WithinTransaction`. Purest Unit-of-Work and explicit at every call site — but a larger diff, and a standing footgun: calling a stripped method without a wrap compiles and silently splits its statements into separate auto-commits (a `TransferOwnership` that demotes without promoting).
- **(C) Only wrap single-statement mutations (rejected).** Leaves the 9 as-is, best-effort. Directly contradicts Q1 and leaves the high-stakes cascade rows — the ones the feature exists to protect — unprotected. Worst of both worlds.

**The codebase evidence decides A.** Of the ~10 hand-rolled-tx call sites, **8 already pair with an activity write** (both `RemoveWithParticipationCascade` sites, `TransferOwnership`, `CreateMany`/invite, `Move`, `UpdatePosition`). The only 2 standalone callers — `create_board` and `create_workspace` — write **no activity at all** ([create_board.go](../../collabotask-backend/internal/usecase/board/create_board.go) has no `WriteActivity` and no broadcast). Under **A** those 2 paths change by *zero lines* (no ambient tx ⇒ they self-begin exactly as today) and keep their existing internal atomicity for free; under **B** they would have to be force-wrapped purely to preserve atomicity they already had, for a feature they don't participate in. A also generalizes an idiom the code already speaks: `rebalanceIfNeeded`/`neighborPosition` ([positioning.go:21](../../collabotask-backend/internal/repository/postgres/positioning.go)) already take an explicit `pgx.Tx` and run on a caller-provided transaction — A lifts that from "explicit param" to "resolved from context."

The accepted cost of A: the transaction is **implicit in the context** (idiomatic Go for this pattern). The related "forgot `exec(ctx)`" risk is closed structurally — the `base` wrapper (Decision #4) makes the wrong call not compile — rather than left to review discipline.

### Q3 — `common.WriteActivity`'s fate

`WriteActivity` does two things: sets `a.UserID = &actorID` for caller convenience, and **swallows** the `Log` error. Strict-for-all makes the swallow wrong — the error must propagate to roll back the tx.

- **Delete it; inline `return activityRepo.Log(ctx, a)` (rejected).** Simplest, but drops the `a.UserID = &actorID` convenience used at all 17 sites, re-scattering `act.UserID = &actorID` boilerplate.
- **Repurpose to a returning helper `LogActivity(ctx, repo, actor, a) error` (chosen).** Keeps the actor-assignment ergonomics, drops only the swallow, makes propagation explicit. A rename-and-unswallow, not a delete.

### Q4 — `exec(ctx)` conversion scope

- **All 52 pool-direct sites (chosen).** `exec(ctx)` is a **behavior-preserving superset** of `r.db`: with no ambient tx it returns the pool (byte-identical to today); with one, it joins. Converting a read that only ever runs outside a transaction is therefore a zero-behavior-change edit — so uniform conversion carries no risk and permanently removes the "unconverted method silently escapes a tx" footgun.
- **Mutation methods only (rejected).** Smaller diff, but leaves a latent read-your-writes bug: a future read-modify-write inside `WithinTransaction` calling an unconverted `GetByID` reads a *different* connection and can't see the tx's own uncommitted writes.

## Decision

1. **Write model = transactional, strict, for all 17 activity call sites** (Q1-strict). Mutation and `Log` commit or roll back together; a `Log` error fails the mutation.
2. **Composition = ambient-tx join-or-begin** (Q2-A). Repos resolve their executor from context; the 9 hand-rolled methods become `joinOrBegin` (self-begin standalone, join when composed).
3. **Placement (mirrors the `Broadcaster` port/impl split):**
   - `Transactor` **port** in `internal/usecase/common/` — `WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error`.
   - `DBTX` interface, unexported `txKey`, the Transactor **implementation**, and an embedded **`base`** wrapper in `internal/repository/postgres/` — the ctx-key must be unexported and shared between the impl (writes it) and every repo (reads it), so all of it sits in one package.
4. **Compile-time footgun closure — the `base` wrapper (see #2 below).** Repos **do not** hold a raw `*pgxpool.Pool db` field. They embed a `base` struct that exposes only `exec(ctx) DBTX` (ambient tx or pool) and `tx(ctx, fn)` (join-or-begin), keeping the pool as an unexported `base.pool`. Statements become `r.exec(ctx).QueryRow(...)`; the old habit `r.db.QueryRow(...)` **no longer compiles**. The raw pool survives only as `base.pool`, reachable only inside the `postgres` package — a single narrow surface guarded by a `forbidigo` lint rule in CI (`\.pool\.(Query|QueryRow|Exec|Begin)` forbidden outside `base.go`).
5. **Usecase shape:** the closure wraps **only** the mutation + `Log`; access-checks, validation, and pre-reads stay **before** it; **broadcast moves to post-commit** (never broadcast a mutation that may roll back).
6. **`common.WriteActivity` → `common.LogActivity(...) error`** (Q3): still sets `a.UserID = &actorID`, returns the error, no swallow. A comment at `LogActivity` pins the propagation to this ADR so it is not "fixed" back to swallowing.
7. **Convert all 52 pool-direct sites to `r.exec(ctx)`** (Q4), and the 9 hand-rolled methods to `r.tx(ctx, fn)`.
8. **DI:** add `ProvideTransactor(db.Pool)`; inject the `Transactor` into the four usecases that write activities (board, card, column, workspace). Repo constructor **signatures are unchanged** (still take `*pgxpool.Pool`); each now wraps it in the embedded `base`.

### Rollout — two phases, each independently green (mitigates the large-diff risk)

The change ships in two sequenced passes so the huge mechanical diff and the small behavioral flip are reviewed — and bisectable — separately:

- **Phase 1 — behavior-preserving refactor.** Introduce `Transactor`, `base`, `exec`/`tx`, the `DBTX` interface and the `forbidigo` rule; convert all 52 sites; rewrite the 9 hand-rolled methods to `r.tx(ctx, fn)`; wire DI. **No usecase changes** — activity writes still run through the old swallowing `WriteActivity`. The full existing suite (598 unit + 10 integration) must stay green with **zero behavior change**. This is the large, boring diff.
- **Phase 2 — the behavioral flip.** Move the 17 call sites inside `WithinTransaction`, swap `WriteActivity` → `LogActivity`, move broadcasts post-commit, add the atomicity tests. Small, focused; this is the only pass where semantics change.

### Testing bar

- **Integration (real Postgres, via [ADR-015](adr-015-integration-test-harness.md)):**
  1. **Transactor rolls back** — `WithinTransaction` doing a real INSERT then returning an error ⇒ row absent.
  2. **Real `Log` honors the ambient tx** — a real `activityRepo.Log` inside a rolled-back `WithinTransaction` leaves no row. This is the footgun guard: it goes red if `Log` (or its repo) still uses `r.db` instead of `exec(ctx)`.
  3. **End-to-end composition** — real DB + real Transactor + real mutation repo with a forced `Log` failure ⇒ the usecase errors **and the mutation row is absent.** Written for `CreateCard` (single-statement) and repeated for the two high-stakes multi-statement paths `TransferOwnership` and break-glass `self_join_board`. This test also **locks the behavior reversal** (#3 in Consequences): reverting `LogActivity` to a swallow turns it red.
- **Unit (mock `Transactor`):** each of the 17 usecases asserts it routes mutation + `Log` through `WithinTransaction` — cheap insurance against future drift.

### Out of scope (unchanged from ADR-007 / deferred)

- The activity **row contents** — firing rule, vocabulary, `entity_id`, metadata snapshot, scope — are ADR-007 and are not reopened here.
- **`activities` reads / UC-22 feed** → Phase 2.
- **Savepoints / nested `WithinTransaction`** — not needed: the port is not nested (repos use the low-level `joinOrBegin`, not the port), and no usecase calls another usecase's `WithinTransaction`. Implementation verifies this holds; if a nesting need appears later, `joinOrBegin` is the extension point.
- **No schema change / no migration** — this is a pure refactor.

## How it works (mechanism + worked example)

### The pieces

```go
// internal/usecase/common/transactor.go — the port (like Broadcaster)
type Transactor interface {
    WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// internal/repository/postgres/dbtx.go — infra-only
type DBTX interface {                       // *pgxpool.Pool AND pgx.Tx both satisfy this
    QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
    Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
    Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}
type txKey struct{}

// internal/repository/postgres/base.go — embedded by every repo; the ONLY hop to the DB
type base struct { pool *pgxpool.Pool }                         // pool is unexported, package-only
func (b base) exec(ctx context.Context) DBTX { /* ambient tx from ctx, else b.pool */ }
func (b base) tx(ctx context.Context, fn func(context.Context) error) error {
    // ambient tx ⇒ run fn(ctx) on it; else b.pool.Begin → fn(ctx') → Commit/Rollback
}

type cardRepository struct { base }         // no `db` field — r.db.QueryRow(...) won't compile
```

- Every repo statement: `r.db.QueryRow(...)` → `r.exec(ctx).QueryRow(...)` (all 52 sites).
- Every hand-rolled method: `r.db.Begin()` → `r.tx(ctx, func(ctx) error { … })`.
- The Transactor impl `Begin`s on the pool, stashes the tx under `txKey` in the returned context, runs `fn`, then `Commit` on success / `Rollback` on error — it shares `txKey` with `base` so `r.exec(ctx)`/`r.tx(ctx, …)` pick the ambient tx up.

### Worked example — `TransferOwnership` (high-stakes, multi-statement)

**Before** ([transfer_ownership.go:43-64](../../collabotask-backend/internal/usecase/board/transfer_ownership.go)) — the repo's own tx commits, then a *separate, swallowed* `Log`; the two are not atomic:

```go
fromUserID, err := bu.boardMemberRepo.TransferOwnership(ctx, …)   // own tx — commits here
if err != nil { return … }
common.WriteActivity(ctx, bu.activityRepo, requesterID, act)       // separate write, swallowed
bu.broadcaster.Broadcast(…)
```

**After** — access-checks stay outside; mutation + `Log` share one tx; broadcast is post-commit:

```go
// … CheckMutateAccess, GetMemberByBoardAndUser, owner/no-op guards: unchanged, OUTSIDE the tx …
var fromUserID *uuid.UUID
err := bu.tx.WithinTransaction(ctx, func(ctx context.Context) error {
    var err error
    fromUserID, err = bu.boardMemberRepo.TransferOwnership(ctx, input.BoardID, input.ToUserID) // JOINS the ambient tx
    if err != nil { return err }
    return common.LogActivity(ctx, bu.activityRepo, input.RequesterID, ownershipActivity(fromUserID, input)) // same tx
})
if err != nil { return fmt.Errorf("failed to transfer ownership: %w", err) }
bu.broadcaster.Broadcast(input.BoardID, common.OwnershipTransferred{ … })   // post-commit, best-effort — unchanged
return nil
```

`TransferOwnership`'s `r.tx(ctx, …)` sees the ambient tx and runs the demote+promote **on it**; `LogActivity` runs on it too. If the `Log` fails, the whole transfer rolls back. When some future caller invokes `TransferOwnership` standalone, `r.tx` opens its own tx — atomicity preserved with no call-site change.

### The single-statement case — `CreateCard`

The mutation runs first in the closure, so an orphaned `Log` (row written, mutation failed) is structurally impossible:

```go
err := cru.tx.WithinTransaction(ctx, func(ctx context.Context) error {
    if err := cru.cardRepo.Create(ctx, card); err != nil { return err }   // exec(ctx) ⇒ ambient tx
    return common.LogActivity(ctx, cru.activityRepo, input.RequesterID, createActivity(card))
})
if err != nil { return nil, err }
cru.broadcaster.Broadcast(column.BoardID, common.CardCreated{Card: card, Assignee: assignee})   // post-commit
```

## Consequences

**Positive**
- The audit trail becomes **truly atomic** with its mutation; break-glass and ownership-transfer rows can no longer be lost while the mutation survives — closing the one gap ADR-007 knowingly accepted.
- One uniform pattern (no two-speed codebase); `exec(ctx)` makes every repo method tx-composable, and the guarantee is now **verifiable** against a real DB.
- The existing hand-rolled cascades gain a reusable transaction seam (`joinOrBegin`) instead of nine bespoke `db.Begin()` blocks.
- The two standalone creators (`create_board`, `create_workspace`) are untouched; the change concentrates where activity writes actually happen.

**Negative / notes** (each with its strengthened mitigation)

| # | Cost | Mitigation |
|---|------|------------|
| 1 | **Implicit transaction in context** — a repo method can't show "am I in a tx?" from its signature; the fact lives in `ctx`. | Left as a known trade-off (engineering it away = the rejected Option B). Uniformity makes "there may be an ambient tx" always-true-everywhere; a package `doc.go` states the contract. |
| 2 | **Footgun: bypassing the ambient tx** by hitting the pool directly in a new/edited repo method — silently escapes the transaction and survives a rollback. | **Closed structurally, not by review:** the `base` wrapper (Decision #4) removes the `db` field, so `r.db.QueryRow(...)` **won't compile**; the residual `base.pool` surface is forbidden outside `base.go` by a `forbidigo` CI rule; integration test #2 catches any leak at runtime. |
| 3 | **Behavior reversal at 17 call sites** — a `Log` failure now fails the user's mutation (was swallowed). Intended (this *is* Option C), but a real semantics change. | **Locked, not just noted:** integration test #3 goes red if it's reverted to a swallow; a comment at `LogActivity` records the intent. Defended by Q1's same-DB reasoning. |
| 4 | **Large mechanical diff** (52 conversions + 9 method rewrites + DI). | **Two-phase rollout** (Decision, Rollout): a behavior-preserving refactor (full suite green, zero behavior change) then a small behavioral flip — each reviewed and `git bisect`-able on its own, on top of the existing 598 unit + 10 integration tests. |
