# CollaboTask

Collaborative task management app (Trello-style). This is a monorepo.

The backend (Go) is implemented and lives in `collabotask-backend/`.
**All technical detail — architecture, conventions, commands, the "add an endpoint" recipe — is in** `collabotask-backend/README.md` **(conventions + recipe) and** `collabotask-backend/TESTING.md` **(test approach). Read those for any code work.**
The frontend architecture is settled and documented (`docs/architecture/frontend-architecture.md`, ADR-011/012/013). App zone: React + Vite SPA — login, register and the Kanban app. Public zone: Astro (SSG) — landing only. Frontend implementation is a separate effort.

## Now

**Phase 1 backend is complete. Frontend implementation — not started.**

Build order (spec §9.1): ① P0 fixes ✅ → ② positioning ✅ → ③ REST features ✅ → ③.5 cookie auth ✅ → ④ WebSocket ✅ → ⑤ integrity pass ✅ — **all closed.**

Step ⑤ closed 2026-10-26 on branch `test/deferred-sql-tests` — **not yet merged.** The deferred repository-layer SQL tests are written (31 integration tests), the suite has been run end to end, and the Transactor carry-over items are all discharged. `-race` is clean after fixing a pre-existing shared-fixture race in `set_archived_test.go` (unit tree re-verified 8/8 on 2026-10-27; `tests/integration` verified 5/5 on 2026-10-26).

**One open backend item:** five files fail `gofmt` — own branch, no rush, nothing enforces formatting. See `docs/handoff/handoff-integrity-pass-findings.md` › #3.

The default-column spacing defect (boards seeded at 0/1/2 instead of `PositionStep` multiples) was **fixed 2026-11-02** — new boards now seed at 1000/2000/3000, so the frontend's drag-and-drop starts with the gaps ADR-004 intends.

Frontend: architecture is settled (ADR-011/012/013, `docs/architecture/frontend-architecture.md`), the **visual baseline is settled — [ADR-017](docs/architecture/adr/adr-017-ui-foundations.md)** (supersedes the uncommitted `docs/architecture/design-tokens.md`), the **responsive + touch scope is settled — [ADR-019](docs/architecture/adr/adr-019-responsive-touch-scope.md)** (supersedes ADR-017 §8: phones and tablets are in scope for Phase 1, card dragging works on touch), and the **state vocabulary is settled — [ADR-018](docs/architecture/adr/adr-018-state-vocabulary.md)** (2026-11-18, written after ADR-019): 39 interaction states enumerated before any channel was spent, every value measured, motion re-pointed from server-correction to "a teammate did this", Phosphor icons. The **design tool is settled — [ADR-020](docs/architecture/adr/adr-020-design-tool-penpot.md)** (2026-11-18, supersedes ADR-018 §12 only): the design file is **Penpot**, named `Collabotask App`. The Figma account is starter tier on both teams, which gates the two-mode variable collections §12 required; Penpot token sets + theme groups deliver the Fine/Coarse flip for free — verified empirically, so §12's scope is restored rather than cut. Still **no code written**. The design file is **started, not empty** — verified against the live Penpot file on 2026-11-19: the five token sets (116 tokens as of 2026-11-20), both theme groups (`Theme`: Light/Dark, `Density`: Fine/Coarse) and the Card's five-state variant container all exist in `Collabotask App`. The **foundations board is drawn** (2026-11-20, page `Foundations`; binding check clean, both flips looked at). Unstarted: nine more components, the 15 whole-screen boards, the ~16 overlays.

**The design inventory's five open questions are settled — [ADR-021](docs/architecture/adr/adr-021-design-inventory-corrections.md)** (2026-11-19). Grilled against the running API rather than the other ADRs, which turned three of them into facts and found four gaps §6 never asked about: archive is **two-way and always was** (`SetArchived` is the one board path that skips `BoardAccessChecker`, so unarchive works today — only a *list* was missing, so ADR-018's "Unarchive is unreachable" is corrected, and the SRS had said so all along); **column reorder gets Move left / Move right** in the column `⋯` menu, closing a WCAG 2.5.7 gap ADR-019 §6 built for cards and skipped for columns; **overlays below 640px follow four rules, not fifteen decisions**; and the **read-only banner carries Join** — its only possible occupant is always allowed to join. Also settled: workspace screen = boards full-screen + team behind a member-count button, profile = the user menu, six empty flavours not five, Create drops you inside what you made, and password reset is now a recorded Phase-2 deferral rather than an unnoticed hole.

**Form-control geometry is settled — [ADR-022](docs/architecture/adr/adr-022-form-control-geometry.md)** (2026-11-20, **extends** ADR-017 §9 — supersedes nothing). Found by running the design file's own Stage 2 question 4 against the live catalogue *before* anything was drawn: ADR-017 fixes what a control looks like (14/20 text, 6px radius, the `--input` boundary) but never how big one is, because §9's vocabulary is the board and Login is a form. Four values — `size.control-height` 36/44, `spacing.control-padding-x` 12, `spacing.field-gap` 16/20, `size.form-width` 400 — now in Penpot and flip-verified. **ADR-017 is annotated, not edited**, which is its own established pattern.

**Spacing between objects is settled — [ADR-023](docs/architecture/adr/adr-023-spacing-between-objects.md)** (2026-11-20, **extends** ADR-017 §9 and adds a `borderWidth` row to ADR-020 §3 — supersedes nothing). Stage 1's question 4 fired before drawing: every spacing token sat inside one object, and nothing named the space *between* two. Six of its eight values are derived from ADR-017/018/019 rather than chosen — `spacing.screen-x` 16, `spacing.column-gap` 8/12 (the only gap ADR-019 §2's width table reconciles at), **zero vertical canvas padding** (what both card-count derivations already assume), `size.board-rail` 4, `border.width` 1 / `-strong` 2. The two choices: `spacing.section-gap` 24 and initials at `type.meta`. In Penpot and flip-verified. The inventory's §7.5 binding check was also **rewritten**: as first written it could not see fills, padding or gaps at all.

**One backend item is open, and it is the only thing ADR-021 cannot deliver in docs:** `GET /workspace/:id/board` needs an `archived` query parameter (handler → usecase → a parameterised predicate in `getUserBoardsInWorkspace`, which currently hardcodes `AND b.is_archived = FALSE`). SRS UC-08 documents it. **This reopens a backend otherwise complete** — small, but say so rather than discover it.

Next: **two independent roots, no blocking edge between them.** ① **Scaffolding** — the Vite app zone before the Astro landing; ADR-018 §*Downstream edits* item 6 is the checklist (override shadcn's `disabled:opacity-50`, set `timeout: 10000` on axios, confirm Phosphor's import path and default status, confirm the Tailwind major version). ② The **Penpot design file** per ADR-020 §§2–6 — [`docs/architecture/design-file-inventory.md`](docs/architecture/design-file-inventory.md) is the authoring checklist, and **[§7](docs/architecture/design-file-inventory.md#7-the-order-to-draw-in-and-what-is-already-drawn) is where a drawing session starts**: the six-stage order (proof run first — two parts, one screen, one full check, before the other seven parts), what is already drawn, and a paste-in query that reports the file's real state instead of trusting a list. **Stage 1 is done; next up is Stage 2, the proof run** — Input, Button, board 2.1 Login, then the four-question check. Either root can go first. The `archived` parameter is only needed before the Archived list is *built*, not before it is drawn.

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
- Design file inventory — `docs/architecture/design-file-inventory.md` (what to draw in Penpot: the 15 whole-screen boards, ~16 overlays, a 39-state coverage check — derived from ADR-017/018/019/020/021, decides nothing. Its §6 is closed; the answers live in ADR-021)



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