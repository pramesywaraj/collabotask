# Handoff — Transactor / Atomic Activity Writes (⑤ integrity pass, step 3): implement

**Date:** 2026-08-31
**Repo:** `collabotask` (backend: `collabotask-backend/`)
**Status:** Design **settled + approved** via `/grill-with-docs` (2026-08-31). Recorded as **ADR-016** (Accepted). **Nothing implemented yet.** Next session should `/implement` this test-first, using the ADR-015 harness for the atomicity tests.

This is the **third** item of the ⑤ post-Phase-1 integrity pass: harness ✅ → composite-FK assignee invariant ✅ → **Transactor / Option-C (this)** → remaining deferred SQL tests (cascade, rebalance, visibility filter, `card_count`).

> **ADR-016 is the source of truth.** This handoff is only the bridge + ordered build checklist. Read [`docs/architecture/adr/adr-016-transactor-atomic-activity-writes.md`](../architecture/adr/adr-016-transactor-atomic-activity-writes.md) first — it holds the full rationale, the mechanism, worked examples (`TransferOwnership`, `CreateCard`), and the Consequences/mitigations table. Do not re-litigate the decisions below.

---

## The one-paragraph summary

Discharge ADR-007's deferred **Option C**: make each board mutation and its activity-log write **atomic** — one transaction, both commit or both roll back — closing the Phase-1 gap where a mutation could commit while its activity row was lost. Repos gain a context-carried executor (ambient tx if the usecase opened one, else the pool); the 17 activity-writing usecases wrap `mutation + Log` in a `Transactor.WithinTransaction(...)` closure. This reverses the Phase-1 "swallow the log error" behavior: a `Log` failure now fails the mutation (defended by the same-DB reasoning in ADR-016 Q1).

## Settled decisions (do not re-litigate — see ADR-016)

1. **Strict atomicity for all 17 activity call sites** (not a high-stakes-only split).
2. **Composition = ambient-tx "join-or-begin"** (Option A): repos resolve the executor from `ctx`; the 9 hand-rolled tx methods join an ambient tx or open their own.
3. **Placement:** `Transactor` **port** in `internal/usecase/common/` (mirrors `Broadcaster`); `DBTX` iface, unexported `txKey`, the impl, and the embedded **`base`** wrapper in `internal/repository/postgres/`.
4. **`base` wrapper closes the footgun at compile time:** repos embed `base` (exposing only `exec(ctx)` and `tx(ctx, fn)`); there is **no `db` field**, so `r.db.QueryRow(...)` won't compile. Residual `base.pool` guarded by a `forbidigo` CI rule.
5. **`common.WriteActivity` → `common.LogActivity(...) error`** — sets `a.UserID`, returns the error, no swallow.
6. **Convert all 52 pool-direct sites to `r.exec(ctx)`**; the 9 hand-rolled methods to `r.tx(ctx, fn)`.
7. **Usecase shape:** closure wraps only mutation + `Log`; access-checks/reads before; **broadcast post-commit**.
8. **No schema change / no migration** — pure refactor.

---

## Build plan — two phases, each independently green (ADR-016 "Rollout")

### Phase 1 — behavior-preserving refactor (the large, boring diff; zero behavior change)

Land the plumbing with **no usecase changes** — activity writes still go through the old swallowing `WriteActivity`. The full existing suite (598 unit + 10 integration) must stay green.

1. **`internal/repository/postgres/dbtx.go`** — define `DBTX` (`QueryRow`/`Query`/`Exec`, satisfied by both `*pgxpool.Pool` and `pgx.Tx`), the unexported `txKey`, and the ctx get/put helpers.
2. **`internal/repository/postgres/base.go`** — `type base struct { pool *pgxpool.Pool }` with `exec(ctx) DBTX` (ambient tx or pool) and `tx(ctx, fn)` (join-or-begin: ambient ⇒ run `fn` on it; else `Begin → fn → Commit/Rollback`).
3. **Transactor impl** (same package) — `Begin`s on the pool, puts the tx under `txKey`, runs `fn`, `Commit`/`Rollback`. Shares `txKey` with `base`.
4. **Every repo** — embed `base`; drop the `db` field; constructors keep their `*pgxpool.Pool` signature and wrap it (`&cardRepository{base: base{pool: pool}}`).
5. **Convert all 52 sites** — `r.db.QueryRow(...)` → `r.exec(ctx).QueryRow(...)`. Rewrite the **9 hand-rolled methods** (`card.Move`, `boardMember.TransferOwnership`, both `RemoveWithParticipationCascade`, `boardMember.CreateMany`, `column.CreateMany`, `column.UpdatePosition`, `board.CreateWithOwner`, `workspace.CreateWithOwner`) from `db.Begin()` to `r.tx(ctx, func(ctx) error { … })`. Note `rebalanceIfNeeded`/`neighborPosition` already take an explicit `pgx.Tx` — feed them `r.exec(ctx).(pgx.Tx)` (or thread the tx through) inside the closure.
6. **`forbidigo` rule** in `.golangci.yml` — forbid `\.pool\.(Query|QueryRow|Exec|Begin)` outside `base.go`.
7. **Gate:** `go test ./...` + `go test -race ./...` + `go test ./tests/integration/...` all green, unchanged behavior.

### Phase 2 — the behavioral flip (small, focused; the only pass where semantics change)

1. **`Transactor` port** in `internal/usecase/common/` + **`ProvideTransactor(db.Pool)`** in `internal/injection/providers.go`; inject into the 4 usecases that write activities (board, card, column, workspace) — add the field + constructor param, update `wire.go`/providers, run **`wire`** to regen `wire_gen.go`.
2. **`.mockery.yaml`** — add `Transactor: {}` under `collabotask/internal/usecase/common` (regens into `common_mocks.go`, like `Broadcaster`); run **`mockery`**. ⚠️ The mock's `WithinTransaction` **must invoke the passed `fn(ctx)`** (testify `.Run(...)`) so usecase unit tests actually exercise the closure body.
3. **`common.WriteActivity` → `LogActivity(...) error`** ([common/activity.go](../../collabotask-backend/internal/usecase/common/activity.go)) — return the error instead of swallowing; add a comment pinning the no-swallow to ADR-016.
4. **Flip the 17 call sites** — wrap `mutation + LogActivity` in `tx.WithinTransaction(...)`; move `Broadcast(...)` **after** it (post-commit). (`WriteActivity` sites list below.)
5. **Tests** (ADR-016 testing bar):
   - **Integration (real DB):** (1) `WithinTransaction` rolls back a real INSERT on error; (2) a real `Log` inside a rolled-back tx leaves no row (footgun guard); (3) forced `Log` failure ⇒ usecase errors **and mutation row absent** — for `CreateCard`, `TransferOwnership`, break-glass `self_join_board`. Test (3) also **locks the behavior reversal**.
   - **Unit (mock Transactor):** each of the 17 usecases asserts it routes mutation + `Log` through `WithinTransaction`.
6. **Gate:** full suite + `-race` green; then `/code-review` (two-axis) before commit.

---

## The 17 activity call sites (where the Phase-2 flip lands)

`card/{create,update,move,delete}_card.go` · `board/{update_board,leave_board,set_archived,invite_member,transfer_ownership,self_join_board,remove_member}.go` · `workspace/{leave_workspace,remove_member,util}.go` · `column/{create_column,update_column,update_column_position,delete_column}.go`
(Grep to confirm: `grep -rl WriteActivity internal/usecase --include='*.go'`.)

The two **standalone** hand-rolled callers that write **no** activity and therefore need **zero** call-site change under Option A: `board/create_board.go`, `workspace/create_workspace.go`.

## Gotchas

- **Broadcast must move post-commit** at every flipped site — broadcasting inside the closure would announce a mutation that may still roll back.
- **`joinOrBegin` re-entrancy:** the port is not nested (repos use `base.tx`, usecases use the port); confirm no usecase calls another usecase's `WithinTransaction`. If one appears later, `base.tx` is the extension point (savepoints).
- **Reads stay outside the closure** — access-checks and pre-reads (`GetByID`, `CheckMutateAccess`) run before `WithinTransaction`; only the write + `Log` go inside.
- The mutation runs **first** in every closure, so an orphaned `Log` (row written, mutation failed) is structurally impossible.

## Commands

```
cd collabotask-backend
go test ./...                       # unit
go test -race ./...                 # race
go test ./tests/integration/...     # integration (testcontainers postgres:16-alpine)
mockery                             # after editing .mockery.yaml
go run github.com/google/wire/cmd/wire ./internal/injection   # after DI changes
```

## After it lands

Update `## Now` in `CLAUDE.md` (⑤ Transactor ✅), memory [[activities-logging-best-effort-then-atomic]] and [[post-phase1-integration-tests]]. Then ⑤'s last item: the remaining deferred SQL tests (cascade, rebalance, visibility filter, `card_count`).
