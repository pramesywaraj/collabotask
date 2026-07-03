# CollaboTask

Collaborative task management app (Trello-style). This is a monorepo.

The backend (Go) is implemented and lives in `collabotask-backend/`.
**All technical detail — architecture, conventions, commands, the "add an endpoint" recipe — is in `collabotask-backend/CLAUDE.md`. Read that for any code work.**
The frontend (Next.js) is planned as a separate effort.

## Now
<!-- Update "Building" when focus shifts. Read every turn. Full sequence: spec §9.1. -->
Phase 1 · CRUD done: auth, workspace, board, column, card
- **Unit-test coverage complete ✅** (275/275 cases; `collabotask-backend/temp_unit-test-checklist.md`). Each layer surfaced & fixed correctness bugs — auth: case-insensitive email dedup, login nil-user guard, dead validation code; column: `position` validation rejected `0` (fixed to `*int` presence-required + clamped, commit `31d858d`); card: same `0`-rejected bug on `to_position` (fixed), `UpdateCard` assignee resolved *before* write; common: `board_access_check.go` workspace-member lookup swallowed unexpected DB errors as `ErrUserNotInWorkspace` (fixed to wrap and propagate; `ErrMemberNotFound` still maps to `ErrUserNotInWorkspace`). Deferred test cases noted in checklist for when `visibility` column lands (step ③).
- **Building → fractional NUMERIC positioning** (build-order step ①): migrate card-move and column-reorder off integer-shift to NUMERIC midpoint (spec §3.2). No realtime needed yet.
- **Queue (in order):** ① fractional NUMERIC positioning ← *here* → ② WebSocket layer + `activities` table → ③ new features: board `visibility`, ownership transfer, promote/demote, leave/delete workspace → ④ assignment=participation (board-member validation + unassign cascade; **needs ② first**).
- **Don't touch:** Phase-2 deferrals (comments, labels, attachments, public boards, refresh tokens, account deletion) and the planned Next.js frontend.

## Core Concepts
- Hierarchy: **Workspace → Board → Column → Card**
- **Workspace Admin authority outranks Board authority.**
- Board visibility: `workspace-visible` or `private`.
- Private boards are visible to workspace admins but require **joining before opening contents**.
- Realtime: board changes (card create/update/delete/move, column reorder, member join/leave) are broadcast to connected clients. Clients are optimistic; **the server is the source of truth.**

## Tech Stack
- Backend: Go, Gin, PostgreSQL (pgx), Google Wire, JWT, WebSocket
- Frontend (planned): Next.js, TypeScript, TanStack Query, Zustand, dnd-kit

## Docs
- PRD — `docs/product/001-PRD.md`
- User stories — `docs/product/001-user-stories.md`
- Software spec — `docs/spesifications/001-software-specifications.md`
- Backend technical — `collabotask-backend/CLAUDE.md`
- Architecture — `docs/architecture/backend-architecture.md`, `docs/architecture/adr/`

### Updating docs (flow downstream: PRD → user-stories → SRS → code)
- **`## Now` above** — every time I switch focus (the only frequently-edited thing).
- **`CONTEXT.md`** — only when a domain *term* is added/renamed/retired. Not for features.
- **PRD / user-stories** — scope or product-decision changes (edit first, upstream).
- **SRS** — technical-contract changes (API, schema, permission rules, closing §9 audit items).
- **ADRs** — never edit an existing one; write a new ADR that supersedes it.

## Working With Me
- **Don't scan the repo.** For a feature, read only the relevant user story + spec section.
- **Non-trivial change → I plan first and wait for your approval before coding.** One-liners I just do.
- **If a request conflicts with the spec or architecture, I flag it before building** — not after.
- Specific decisions (e.g. "admins bypass the member check") go to memory, not this file.
- **Task tracking is doc-driven + just-in-time:** spec §9.1 is the roadmap, `## Now` is the pointer. I break a step into a TodoWrite checklist only when you start it — no GitHub issues, no separate task files.

## Agent skills

### Issue tracker

Issues live in GitHub Issues; use the `gh` CLI. External PRs are not a triage surface. See `docs/agents/issue-tracker.md`.

### Triage labels

Default label vocabulary — no custom mappings. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context layout: `CONTEXT.md` at repo root + `docs/architecture/adr/`. See `docs/agents/domain.md`.
