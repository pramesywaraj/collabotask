# ADR-015: Integration Test Harness — Repository Layer

- **Status:** Accepted
- **Date:** 2026-08-23
- **Scope:** Integration testing approach for `internal/repository/postgres/`. Complements ADR-003 (unit testing strategy), which explicitly deferred this layer.

## Context

ADR-003 established testify + mockery for use-case unit tests and explicitly excluded the repository layer: "Talks to Postgres; needs a real DB. Use integration tests." Since then, several SQL-level behaviours have been deferred — repo-layer rebalance, visibility filter, `card_count`, and board unassign-cascade — because no test harness existed.

Step ⑤ of the build queue (post-Phase-1 integrity pass) opens with standing up this harness so those deferred tests can be written. The composite-FK assignee invariant (ADR-010 candidate) also depends on it.

The decisions below were locked via a `/grill-with-docs` session on 2026-08-23.

## Decisions

### 1. Postgres infrastructure — testcontainers-go

**Options considered:**
- **A — testcontainers-go**: each `go test` invocation spins up a fresh `postgres:16-alpine` container, runs migrations, executes tests, tears down. Zero manual setup; CI-friendly.
- **B — assume local Postgres running** (dev `docker-compose up postgres`): connect via env var DSN. Requires the dev DB to be up; flaky in CI.
- **C — dedicated docker-compose test service** on a separate port: same fragility as B, just isolated to a different port.

**Decision: A (testcontainers-go).** Self-contained for both local and CI — Docker is already a requirement for the dev workflow. `postgres:16-alpine` matches the existing dev compose image, so no version drift. Pays a ~3–5s container startup cost per package binary, not per test function.

### 2. Test isolation — truncate all tables in `t.Cleanup`

**Options considered:**
- **A — per-test transaction rollback**: wrap each test in a TX, defer rollback. Zero overhead but incompatible with repos that open their own transactions (`CreateWithOwner`, `RemoveWithParticipationCascade`, etc.).
- **B — truncate all tables in `t.Cleanup`**: a helper runs `TRUNCATE … RESTART IDENTITY CASCADE` after each test. Works with any repo code regardless of internal transactions.
- **C — fresh container per test function**: perfect isolation but pays startup cost per test — prohibitively slow.

**Decision: B (truncate in `t.Cleanup`).** The repo layer makes heavy use of internal transactions, ruling out option A. Truncation is a single SQL statement (milliseconds) and gives the same clean-slate guarantee.

### 3. Test file location — separate `tests/integration/` directory

**Options considered:**
- **A — inline in `internal/repository/postgres/`**, gated by `//go:build integration` build tag.
- **B — separate `tests/integration/` directory**: integration tests live outside `internal/`; `go test ./internal/...` never touches them without any flag.

**Decision: B (separate directory).** Physical separation is unconditional — `go test ./internal/...` always runs unit tests only; `go test ./tests/integration/...` always runs integration tests. No build tag discipline required from contributors. The tradeoff (test files are farther from implementation) is acceptable given the small number of repo packages.

### 4. Schema setup — golang-migrate programmatically

**Options considered:**
- **A — `golang-migrate` called in Go** (`m.Up()` pointing at `migrations/`): already in `go.mod`, consistent with production usage.
- **B — embed SQL files with `//go:embed`**: no path fragility, SQL travels with the binary.

**Decision: A (golang-migrate).** Already a dependency; no new tooling. Path is anchored to the module root via `runtime.Caller(0)` in the harness to avoid relative-path fragility.

### 5. Container lifecycle — one container per package via `TestMain`

**Options considered:**
- **A — `TestMain` starts one container for the whole package**: pays the startup cost once; tests clean up with `TruncateAll`.
- **B — one container per test function**: perfect isolation, but multiplies startup cost by the number of tests.

**Decision: A (`TestMain`).** Decision 2 (truncate) already provides per-test isolation. One container per package is the standard testcontainers-go pattern for repo suites and keeps total test time acceptable.

### 6. Test fixtures — Go helper functions

**Options considered:**
- **A — Go helper functions** (`createTestUser`, `createTestWorkspace`, etc.) in `testutil`: type-safe, composable, fail at compile time on schema changes.
- **B — Raw SQL fixture files**: flexible for large static datasets but brittle on schema changes.

**Decision: A (Go helpers).** The repo suite is transactional CRUD — fixtures need to compose dynamically (e.g. "a workspace with 3 boards of mixed visibility"). SQL files don't compose well. Helpers stay in sync with entity types automatically.

## Resulting Structure

```
collabotask-backend/
└── tests/
    └── integration/
        ├── testutil/
        │   ├── harness.go       ← NewTestDB (TestMain helper), TruncateAll
        │   └── fixtures.go      ← createTestUser, createTestWorkspace, …
        ├── board_repo_test.go
        ├── card_repo_test.go
        ├── column_repo_test.go
        ├── workspace_repo_test.go
        └── …
```

## Run Commands

```bash
# Unit tests only (fast, no Docker required)
go test ./internal/...

# Integration tests only (requires Docker)
go test ./tests/integration/...

# Both
go test ./...
```

## Consequences

**Positive**
- Repository layer is now testable against a real PG16 database, unblocking all deferred SQL-level tests.
- Harness is self-contained — no manual setup step for contributors or CI.
- Clean separation between fast unit tests and slow integration tests by directory path alone.
- `TestMain` + `TruncateAll` pattern keeps test runtime proportional to the number of repo packages, not test cases.

**Negative / risks**
- Requires Docker running locally to execute integration tests (same requirement as the existing dev workflow).
- Migration path (`runtime.Caller(0)` anchor) must be kept correct if the harness file is ever moved.
- Tests in `tests/integration/` must import repo packages explicitly — slightly more verbose than inline tests.

## Relationship to Other ADRs

- **ADR-003** — established the unit test strategy this ADR complements. ADR-003's explicit exclusion of the repository layer is now resolved.
- **ADR-010** (planned) — composite-FK assignee invariant; depends on this harness being in place before it can be tested.
- **ADR-007** — activity logging Transactor atomicity (Option C) is also deferred to ⑤ and will use this same harness.
