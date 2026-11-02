# CollaboTask

Collaborative task management app (Trello-style). This is a monorepo.

The backend (Go) is implemented and lives in `collabotask-backend/`.
**All technical detail — architecture, conventions, commands, the "add an endpoint" recipe — is in** `collabotask-backend/README.md` **(conventions + recipe) and** `collabotask-backend/TESTING.md` **(test approach). Read those for any code work.**
The frontend architecture is settled and documented (`docs/architecture/frontend-architecture.md`, ADR-011/012/013). App zone: React + Vite SPA — login, register and the Kanban app. Public zone: Astro (SSG) — landing only. Frontend implementation is a separate effort.

## Now

**Phase 1 backend is complete. Frontend implementation — not started.**

Build order (spec §9.1): ① P0 fixes ✅ → ② positioning ✅ → ③ REST features ✅ → ③.5 cookie auth ✅ → ④ WebSocket ✅ → ⑤ integrity pass ✅ — **all closed.**

Step ⑤ closed 2026-10-26 on branch `test/deferred-sql-tests` — **not yet merged.** The deferred repository-layer SQL tests are written (31 integration tests), the suite has been run end to end, and the Transactor carry-over items are all discharged. `-race` is clean after fixing a pre-existing shared-fixture race in `set_archived_test.go` (unit tree re-verified 8/8 on 2026-10-27; `tests/integration` verified 5/5 on 2026-10-26).

**Two open backend items**, both pre-existing, both detailed in `docs/handoff/handoff-integrity-pass-findings.md`:
- New boards seed their default columns at 0/1/2 instead of `PositionStep` multiples (`internal/repository/postgres/board.go:81`). Harmless today, but a client computing drop midpoints lands on the rebalance threshold on its very first drag — **fix before frontend work starts.** Three column cases in `positioning_test.go` encode today's spacing and move with it.
- Five files fail `gofmt` (own branch, no rush — nothing enforces formatting).

Frontend: architecture is settled (ADR-011/012/013, `docs/architecture/frontend-architecture.md`); **no code written yet.** First decision when it starts is what to scaffold — the Vite app zone comes before the Astro landing.

**Don't touch:** Phase-2 deferrals — comments, labels, attachments, public boards, refresh tokens, account deletion.

Per-step history, decisions and dates: `docs/build-log.md`.

## Core Concepts

- Hierarchy: **Workspace → Board → Column → Card**
- **Workspace Admin authority outranks Board authority.**
- Board visibility: `workspace-visible` or `private`.
- Private boards are visible to workspace admins but require **joining before opening contents**.
- Realtime: board changes (card create/update/delete/move, column reorder, member join/leave) are broadcast to connected clients. Clients are optimistic; **the server is the source of truth.**



## Tech Stack

- Backend: Go, Gin, PostgreSQL (pgx), Google Wire, JWT, WebSocket
- Frontend (app zone): React + Vite SPA, TypeScript, React Router v7, TanStack Query, Zustand, axios, shadcn/ui, Tailwind, dnd-kit
- Frontend (public zone): Astro (SSG) — landing only; React islands where interactivity is needed



## Docs

- Build log — `docs/build-log.md` (what's been built, in build order, with links to each ADR/handoff)
- PRD — `docs/product/001-PRD.md`
- User stories — `docs/product/001-user-stories.md`
- Software spec — `docs/spesifications/001-software-specifications.md`
- Backend technical — `collabotask-backend/README.md` (conventions + "add an endpoint" recipe), `collabotask-backend/TESTING.md` (test approach)
- Architecture — `docs/architecture/backend-architecture.md` (backend), `docs/architecture/frontend-architecture.md` (frontend), `docs/architecture/adr/`



### Updating docs (flow downstream: PRD → user-stories → SRS → code)

- `## Now` **above** — every time I switch focus (the only frequently-edited thing). Keep it a pointer, not a log.
- `docs/build-log.md` — append an entry when a build-order step completes; move the detail out of `## Now`.
- `CONTEXT.md` — only when a domain *term* is added/renamed/retired. Not for features.
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