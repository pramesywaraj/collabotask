# Collabotask — Backend

A **Trello-like task management API**. Collabotask lets teams organize work on
Kanban-style boards: create workspaces, invite members, build boards with
columns, and move cards (tasks) across columns to track progress.

Built in Go with **Clean Architecture** — business logic is isolated from HTTP,
the database, and any framework, so it stays testable and easy to change.

> **Status: Phase 1 is built.** The REST API, httpOnly-cookie auth, and real-time
> collaboration — WebSocket broadcast, presence, and participation cascades — are
> all in place, along with the activity log that records what a team did. See the
> [PRD](../docs/product/001-PRD.md) and [user stories](../docs/product/001-user-stories.md)
> for what that covers, and the [build log](../docs/build-log.md) for how it got here.
>
> One API contract item is still open: `GET /workspace/:id/board` needs an
> `archived` query parameter, without which archived boards cannot be listed and
> so cannot be unarchived
> ([ADR-021](../docs/architecture/adr/adr-021-design-inventory-corrections.md)).
> The frontend is designed but not yet written.

---

## Features

- **Auth** — register / login with JWT; bcrypt-hashed passwords.
- **Workspaces** — create, list, view; invite & remove members.
- **Boards** — create, list, detail, archive; member management
  (invite, remove, self-join, leave); role-based access.
- **Columns** — create, update, reorder, delete within a board.
- **Cards** — create, update, delete, and move across columns (drag-and-drop
  ordering on the backend).
- **Kanban view** — fetch a whole board (columns + cards + assignees) in one call.
- **API docs** — Swagger UI served at `/swagger/index.html`.

## Tech stack

| Concern | Choice |
|---|---|
| Language | Go 1.25 |
| HTTP framework | [Gin](https://github.com/gin-gonic/gin) |
| Database | PostgreSQL via [pgx](https://github.com/jackc/pgx) |
| Migrations | [golang-migrate](https://github.com/golang-migrate/migrate) (run on startup) |
| Auth | JWT ([golang-jwt](https://github.com/golang-jwt/jwt)) + bcrypt |
| Validation | [go-playground/validator](https://github.com/go-playground/validator) |
| DI | [google/wire](https://github.com/google/wire) |
| Logging | [zerolog](https://github.com/rs/zerolog) |
| API docs | [swaggo](https://github.com/swaggo/swag) |

---

## Quick start

Requirements: **Go 1.25+** and **PostgreSQL** (or just Docker).

```bash
# 1. Configure
cp .env.example .env        # then fill in values (see "Configuration" below)

# 2a. Run everything with Docker (app + postgres)
docker compose up

# 2b. Or run locally against your own postgres
go run ./cmd/api
```

Migrations run **automatically on startup** — there is no separate migrate step.

- API base path: `http://localhost:<SERVER_PORT>/api/v1`
- Swagger UI: `http://localhost:<SERVER_PORT>/swagger/index.html`

### Configuration

All config is read from environment variables (full list with defaults in
`.env.example`). The ones you **must** set:

| Variable | Purpose |
|---|---|
| `AUTH_JWT_SECRET` | Secret for signing JWTs — **app won't start without it** |
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` | Database connection |

---

## API overview

All routes are under `/api/v1`. Protected routes require `Authorization: Bearer <token>`.

| Method | Path | Description |
|---|---|---|
| `POST` | `/auth/register` | Register a new user |
| `POST` | `/auth/login` | Log in, receive JWT |
| `GET` | `/user/profile` | Current user's profile |
| `POST` `GET` | `/workspace` | Create / list workspaces |
| `GET` | `/workspace/:id` | Workspace detail |
| `POST` `DELETE` | `/workspace/:id/member/...` | Invite / remove workspace member |
| `POST` `GET` | `/workspace/:id/board` | Create / list boards |
| `GET` | `/workspace/:id/board/:board_id/kanban` | Full kanban (columns + cards) |
| `POST` `PATCH` `DELETE` | `/.../columns/...` | Manage columns (incl. reorder) |
| `POST` `PATCH` `DELETE` | `/.../cards/...` | Manage cards |
| `POST` | `/.../cards/:card_id/move` | Move a card across columns |

Full, always-accurate reference: **Swagger UI** at `/swagger/index.html`.

---

## Project layout

```
cmd/api/              Application entry point
internal/
  domain/             Entities + repository interfaces + domain errors  (innermost)
    entity/
    repository/       Repository INTERFACES (the contracts)
  usecase/            Application logic, one package per feature
  delivery/http/      HTTP layer: handlers, request/response, middleware, router
  repository/postgres/ PostgreSQL implementations of the repository interfaces
  infrastructure/     Technical adapters: DB pool, JWT, bcrypt
  injection/          Google Wire dependency injection
pkg/                  Reusable, framework-free utilities (logger, validator)
migrations/           SQL migration files
```

The dependency rule: **everything points inward**. `usecase` and `domain` never
import HTTP, the database, or any framework. See
[docs/architecture/backend-architecture.md](../docs/architecture/backend-architecture.md) for the full picture.

---

## Development

```bash
go build ./...                       # compile everything
go test ./...                        # run tests (set AUTH_JWT_SECRET for config tests)
go vet ./...                         # static checks
gofmt -l internal/ pkg/              # list unformatted files (should be empty)
cd internal/injection && go generate # regenerate Wire after editing providers.go
go doc ./internal/usecase/board BoardUseCase   # list a use case's methods
```

### Conventions (short version)

- Use cases are **concrete structs** (no interface), return **entities** directly.
- A dedicated result type exists only when it **transforms** data (e.g.
  `auth.UserProfile` hides the password hash); otherwise return the entity.
- Interfaces live on the **consumer** side, only when there's a real reason
  (repositories, the shared board-access checker, the auth ports).

Full conventions and the "add an endpoint" recipe: [docs/architecture/backend-architecture.md](../docs/architecture/backend-architecture.md).

## Documentation

| Doc | What it covers |
|---|---|
| [docs/architecture/backend-architecture.md](../docs/architecture/backend-architecture.md) | Layers, conventions, how to add a feature |
| [docs/architecture/adr/adr-001-…](../docs/architecture/adr/adr-001-simplify-clean-architecture.md) | Why DTOs & ceremonial interfaces were removed |
| [docs/architecture/adr/adr-002-…](../docs/architecture/adr/adr-002-by-component-layout-and-ports.md) | Why by-component layout, auth ports, validator in `pkg/` |
| [docs/product/001-PRD.md](../docs/product/001-PRD.md) | Product requirements |
| [docs/product/001-user-stories.md](../docs/product/001-user-stories.md) | User stories & acceptance criteria |
| [docs/spesifications/001-software-specifications.md](../docs/spesifications/001-software-specifications.md) | API contracts, schema, permission rules |
