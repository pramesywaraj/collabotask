# ADR-003: Unit Testing Strategy — testify + mockery

- **Status:** Accepted
- **Date:** 2026-06-26
- **Scope:** Unit testing approach for `collabotask-backend`. Does not cover integration tests or end-to-end tests.

## Context

ADR-001 noted that zero unit tests existed in `internal/usecase` at the time of the clean architecture refactor, and deliberately deferred writing them — testability was an *enabled benefit*, not part of that scope. Now that the architecture is stable, we are beginning to write use case tests.

`go.uber.org/mock` (gomock) was already present in `go.mod` as a tentative dependency — added without a deliberate evaluation. Before writing any tests, we lock in the toolchain.

The testable surface is almost entirely in `internal/usecase/`. Those packages hold all business logic and depend exclusively on repository interfaces (defined in `internal/domain/repository/`) and `common.BoardAccessChecker`. This design means we can substitute the Postgres implementations with mocks in tests — no database required.

## Options Considered

### Option A: gomock (`go.uber.org/mock`)

- Mocks generated per-interface via `mockgen`
- Strict by default: unexpected calls fail the test immediately
- Verbose `EXPECT()` / `gomock.Any()` / `gomock.InOrder()` API
- Assertions require a separate library (typically testify anyway)
- Originally from Google, now maintained by Uber

### Option B: testify/mock + mockery (chosen)

- `github.com/stretchr/testify` covers both **assertions** (`assert`, `require`) and **mocks** (`mock`)
- `mockery` is a code generator: one `.mockery.yaml` config file, one `mockery` command generates all mocks
- Readable `.On("Method", args).Return(values)` API
- `AssertExpectations(t)` gives explicit verification that all expected calls happened
- `with-expecter: true` adds a type-safe `.EXPECT()` layer, catching method-name typos at compile time
- De facto community standard; most Go testing examples, blogs, and open-source projects use it

## Decision

**Use testify + mockery.** Remove `go.uber.org/mock`.

The decisive factor: testify covers both assertions and mocks in one ecosystem. In practice, teams using gomock end up importing testify for `assert`/`require` regardless, creating two testing libraries with overlapping concerns. A single ecosystem reduces cognitive overhead and keeps the test setup uniform.

mockery's project-level config (`with-expecter: true`, single `mockery` command) also outperforms `mockgen`'s per-file generation for a codebase with many repository interfaces.

## Consequences

**Positive**
- One import namespace for all testing utilities (`testify/assert`, `testify/require`, `testify/mock`)
- All mocks regenerated with a single `mockery` command at the project root
- Readable test setup that new contributors recognize immediately
- Type-safe mock expectations with `with-expecter: true`

**Negative / risks**
- Not strict by default: mock methods not configured with `.On(...)` return zero values instead of failing. Mitigation: always call `mock.AssertExpectations(t)` at the end of each test — this enforces that every `.On(...)` you declared was actually called
- `go.uber.org/mock` must be explicitly removed from `go.mod` to avoid confusion

## Practical scope

See `collabotask-backend/TESTING.md` for the full playbook: directory layout, mockery config, test structure, and worked examples.
