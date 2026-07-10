# Collabotask Architecture Guide

This backend follows **Clean Architecture**. There is one rule that matters more
than any folder name:

> **Dependency rule — dependencies point inward only.**
> Business code (`domain`, `usecase`) must not know about HTTP, the database, or
> any framework. Outer layers depend on inner layers, never the reverse.

Everything else below is a consequence of that rule, not a separate "pattern".

> Note on layout: folders are organized **by component** (`delivery`,
> `repository`, `usecase`, `domain`) rather than by an `adapter/` umbrella. The
> *names* differ from the textbook's "Interface Adapters", but the rule is the
> same — what matters is the direction of dependencies, not the folder names.

## Layers

```
            HTTP request
                 │
                 ▼
   delivery/http  ─────────────┐   (handlers, request/response — driving side)
                 │             │
                 ▼             │
   usecase       │            depends on
                 │             │
                 ▼             │
   domain    ◄───────────────┘   (entities + repository interfaces)
                 ▲
                 │ implements interfaces
   repository/postgres           (SQL implementations — driven side)
   infrastructure                (DB pool, JWT, bcrypt)
```

| Path | Role |
|------|------|
| `domain/entity/` | Business entities |
| `domain/repository/` | Repository **interfaces** (defined here, on the consumer side) |
| `domain/sentinel_errors.go` | Domain errors |
| `usecase/<feature>/` | Application logic (one concrete `<Feature>UseCase` struct per feature) |
| `usecase/common/` | Shared use case helpers (e.g. `BoardAccessChecker`) |
| `delivery/http/` | REST API: handler, request, response, middleware, router |
| `repository/postgres/` | Repository implementations (SQL) |
| `infrastructure/` | Technical concerns: DB pool, JWT, bcrypt |
| `injection/` | Google Wire dependency injection (composition root) |

**Key invariant:** the repository *interface* stays in `domain/repository/`
(consumer side), while its *implementation* lives in `repository/postgres/`.
The implementation depends on the domain interface — never the reverse.

## Request flow

```
router → handler → usecase → repository (interface) → postgres impl → SQL
                ↓
         request / response JSON
```

Example — `POST /api/v1/workspace/:id/board`:

1. `router/router.go` — route registration
2. `handler/board_handler.go` — bind JSON, call use case
3. `usecase/board/create_board.go` — validation + business rules
4. `domain/repository/board_repository.go` — interface (consumer side)
5. `repository/postgres/board.go` — SQL implementation

## Conventions a contributor must follow

These are the decisions that are easy to get wrong. They are enforced by review,
not by the compiler.

### Use cases return entities, not DTOs

Use cases return domain entities (`*entity.X`) directly. There is **no global
`dto` package**. Only create a dedicated result type when the output genuinely
*transforms* the data (joins, aggregations, or hiding sensitive fields) — and
put it next to the use case, in `usecase/<feature>/<feature>_io.go`.

Examples of justified result types: `board.BoardWithMeta`, `workspace.WorkspaceDetail`,
`auth.UserProfile` (the last one hides `entity.User.PasswordHash`).

Rule of thumb: *transforms the data → make a type; copies the entity 1:1 → return
the entity.* See [ADR-001](./architecture/adr-001-simplify-clean-architecture.md).

### Interfaces only where there is a reason

- **Repositories** use interfaces (in `domain/repository/`) — they have real
  implementors and are mocked in tests. Keep them.
- **Use cases** are concrete structs, not interfaces (`board.BoardUseCase`, not
  an interface). If a handler test ever needs a mock, define a small interface
  **in the handler package** (consumer side), not in the use case package.
- Rule: *accept interfaces, return structs.* Don't add an interface "just in
  case" — Go's structural typing lets you add one later for free.

### Data shapes

| Layer | Purpose |
|-------|---------|
| `entity` | Database / domain shape |
| `request` | Incoming JSON |
| `response` | Outgoing JSON |

Keep sensitive fields (e.g. password hash) out of result types and responses.

### Shared board access

Use `usecase/common.BoardAccessChecker` for board access checks. It exposes
three intent methods over a shared `resolve()` (board + workspace/board
membership + visibility), each enforcing the ADR-005 permission matrix and
returning a `*BoardAccess` (board + membership context):

- `CheckMetadataAccess()` — board detail (thin-roster on PRIVATE + not-joined)
- `CheckViewAccess()` — kanban view (PRIVATE break-glass for non-joined admins)
- `CheckMutateAccess()` — card/column mutations (break-glass + member 403/404)

Pick the method that matches the caller's intent. Do not duplicate access logic
in individual use cases.

### HTTP errors

- Handlers call `helper.HandleUseCaseError(ctx, err)` for use case failures.
- Domain error → HTTP status mapping lives in
  `delivery/http/errors/domain_mapper.go` (`MapDomainError`).

### Validation

- HTTP body validation: Gin binding + `response.HandleValidationError`
- Business input validation: `pkg/validator`, inside use cases

## Adding a new endpoint

Work **inside-out** (innermost layer first):

1. `domain/` — entity or repository method if needed
2. `repository/postgres/` — SQL + implementation
3. `usecase/<feature>/<feature>_io.go` — Input / Output result types
4. `usecase/<feature>/<operation>.go` — business logic (a method on the concrete
   `<Feature>UseCase` struct)
5. `delivery/http/request/` and `response/` — HTTP shapes (if needed)
6. `delivery/http/handler/` — handler method
7. `delivery/http/router/router.go` — route
8. `injection/providers.go` — wire dependencies, then regenerate Wire

## Regenerate Wire

After changing `injection/providers.go`:

```bash
cd internal/injection && go generate
```

## Seeing the available methods of a use case

There is no interface listing them. Use the toolchain:

```bash
go doc ./internal/usecase/board BoardUseCase
```

Or rely on the one-file-per-operation convention: the file names in
`usecase/<feature>/` are the list of operations.

---

For the **why** behind the structure and the simplification history, see:

- [adr-001-simplify-clean-architecture.md](./architecture/adr-001-simplify-clean-architecture.md) — buang DTO/interface seremonial
- [adr-002-by-component-layout-and-ports.md](./architecture/adr-002-by-component-layout-and-ports.md) — layout by-component, auth ports, validator→pkg
- [refactor-learning-notes.md](./architecture/refactor-learning-notes.md) — explanations & examples
- [refactor-playbook.md](./architecture/refactor-playbook.md) — the per-module refactor record
