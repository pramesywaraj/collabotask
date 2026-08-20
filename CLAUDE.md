# CollaboTask

Collaborative task management app (Trello-style). This is a monorepo.

The backend (Go) is implemented and lives in `collabotask-backend/`.
**All technical detail — architecture, conventions, commands, the "add an endpoint" recipe — is in** `collabotask-backend/README.md` **(conventions + recipe) and** `collabotask-backend/TESTING.md` **(test approach). Read those for any code work.**
The frontend architecture is settled and documented (`docs/architecture/frontend-architecture.md`, ADR-011/012). App zone: React + Vite SPA. Public zone: Next.js (SSG). Frontend implementation is a separate effort.

## Now



Phase 1 · CRUD done: auth, workspace, board, column, card

- **Unit-test coverage complete ✅** (275/275 cases; `collabotask-backend/temp_unit-test-checklist.md`). Deferred test cases noted in checklist for when `visibility` column lands (step ③).
- **Fractional NUMERIC positioning complete ✅** (build-order step ②, spec §3.2, ADR-004, migration 000006). Single UPDATE per move; repository-layer rebalance; `domain.PositionStep`/`PositionRebalanceThreshold` as single source of truth. Repo-layer rebalance tests deferred (no DB harness yet).
- **Board** `visibility` **complete ✅** (build-order step ③ first sub-step; spec §2.2–2.4, ADR-005, migration 000007). Three-method access checker (metadata/view/mutate) + break-glass, 404-hide vs 403-reveal, idempotent self-join, thin roster, `created_by` removed from access logic. SQL filter/`card_count` tested at usecase layer only (repo-layer deferred to post-Phase-1 integration pass).
- **Workspace ops complete ✅** (build-order step ③ second sub-step; spec §4.2 UC-06b/c/d). Promote/demote with demote guards (owner→403, last-admin→409, self-demotion OK, idempotent no-op); leave workspace (owner→403, last-admin→409, workspace-scoped participation cascade); delete workspace (owner-only). UC-06 RemoveMember retrofitted with the same cascade (`RemoveWithParticipationCascade` — clears board_members + unassigns cards in one TX). Mocks regenerated; 341 tests pass.
- **Board ownership transfer complete ✅** (build-order step ③ third sub-step; spec §4.3 UC-12e, ADR-006). `created_by` proxy removed from `canAdministerBoard()` + `leave_board` guard (role-based now). Atomic demote-by-role + promote-by-id in one TX; break-glass via `CheckMutateAccess` ∘ `canAdministerBoard`; `ErrTransferTargetNotBoardMember` (400); idempotent no-op; orphan-safe. Mocks regenerated; 361 tests pass. Activity write + `OWNERSHIP_TRANSFERRED` broadcast deferred (step ④).
- **Assignee validation + board-scoped unassign cascade complete ✅** (build-order step ③ fourth sub-step; spec §2.8, UC-14/UC-16, UC-10/UC-12d). Write-time gate: `boardMemberRepo.IsUserExists` → `ErrAssigneeNotBoardMember` (400) on `create_card`/`update_card`, collapsing the old 404 path (Decision C); `update_card` validates a *newly-set* assignee only, never a stale one (Decision B). Atomic `RemoveWithParticipationCascade` (delete membership + `UPDATE cards SET assigned_to=NULL … RETURNING` in one TX) wired into `leave_board` + board `remove_member`; owner-cannot-leave guard stays ahead of it. Mocks + Wire regenerated; 368 tests pass. `CARD_UPDATED` broadcast for cleared cards deferred (step ④; cascade already returns `[]AffectedCard`).
- **Activities table + writes complete ✅** (build-order step ③ fifth/final sub-step; spec §4.6, ADR-007, migration 000008). Best-effort synchronous after-commit writes (`common.WriteActivity` swallows errors, never fails the mutation); log-only-on-state-change; verb×entity vocabulary; snapshot-title metadata; board-scoped incl. per-affected-board rows on workspace cascade (`AffectedBoardIDs`). 17 call sites across card/column/board/workspace. 163 new activity-contract test cases (438 total passing). `CARD_UPDATED` broadcast for cascade-cleared cards and Transactor atomicity remain deferred to step ④ / ⑤.
- **Building → REST feature completion** (build-order step ③ per spec §9.1): board `visibility` ✅ → workspace ops ✅ → board ownership transfer (UC-12e) ✅ → assignee validation + unassign cascade (UC-14/UC-16) ✅ → activities table + writes ✅ → **step ③ COMPLETE**.
- **WebSocket design settled ✅ (not built)** — `/grill-with-docs` locked the full ④ design: `Broadcaster` port in the **usecase layer** (best-effort after-commit, mirrors activities); RWMutex hub + per-conn read/write pumps; **multi-connection edge-triggered presence** (corrects the SRS single-conn sketch); **continuous access enforcement** (instant eviction + new `ACCESS_REVOKED` event, UC-19b); REST-in/broadcast-out; participation-cascade fan-out. Recorded: ADR-009, SRS §4.5/§5, handoff `docs/handoff/websocket-participation-broadcast/index.md`, memory [[websocket-participation-broadcast-design]]. Built in parts A→B→C→D→E. **Post-③.5 drift recheck DONE ✅ 2026-07-29** (`/grill-with-docs`; 4 points, no reversal — reuse `middleware.Auth`, origin scheme→host + fail-fast, CSRF-GET invariant, handshake-only-auth limitation → Phase-2 session model; handoff §Post-③.5 drift recheck). **Parts A/B/C built + reviewed ✅** — C (Broadcaster port + 12 mutation broadcasts) had a 3-way `/code-review` aligned + fixes applied (P4 full-roster-row `MEMBER_ADDED.user`, table-driven tests, P1/P2 silence cases), 569 tests pass, **in PR 2026-08-20**. **Part D next** (continuous enforcement: `MEMBER_REMOVED`+`EvictUser`, `→PRIVATE` `EvictExcept`, board-leave evict, `ACCESS_REVOKED`).
- **Auth ③.5 complete ✅** — `/grill-with-docs` locked the full `Bearer` **→ httpOnly cookie** design (8 decisions, all with recommended-answer grilling): **JWT-in-cookie/stateless** (session/refresh → Phase 2); **topology A** cross-origin sibling subdomains `app.`↔`api.` (single-origin C2 reverse-proxy → future); **host-only cookie** ⇒ **no authenticated SSR** in Phase 1; `__Host-token; HttpOnly; Secure; SameSite=Lax; Path=/; Max-Age=<cfg.JWTExpiration>`; **CSRF** = required `X-CSRF-Protection` header on all mutations incl. login/register/logout (forces CORS-preflight gate; closes sibling-subdomain vector); **hard cutover** (cookie-only, login/register drop body `token`, no `Bearer`); **unconditional logout** `POST /auth/logout`; **CORS corrected** (explicit allowlist, add `PATCH`, no `*`+creds → fail-fast). Register auto-login kept provisionally (email-verify supersedes it, Phase 2). Recorded: ADR-008 (full), SRS §3.5/§4.1 (UC-01/02 + new UC-03 logout), handoff `docs/handoff/handoff-auth-httponly-cookie.md`, memory [[auth-cookie-migration-prereq]]. **Built + reviewed 2026-07-27** (473 tests pass; two-axis `/code-review` — 0 hard standards violations, 0 spec gaps; P1 stale-`Bearer`-prose sweep + P2 clarifying comments applied). **④ now unblocked; its post-③.5 drift recheck is DONE ✅ 2026-07-29** — the WS handshake reuses this step's cookie-based `middleware.Auth` (no inline read), shares the CORS allowlist with `OriginPatterns` (scheme→host translation), and the GET handshake is a deliberate CSRF-header exemption defended by `OriginPatterns`. Phase-2 session model will also close ④'s handshake-only-auth gap (evict-on-revoke).
- **Queue (in order):** ① P0 fixes ✅ → ② fractional positioning ✅ → ③ REST features ✅ → ③.5 auth httpOnly-cookie ✅ → **④ WebSocket + participation broadcast** *(A/B/C built + reviewed ✅ — C in PR 2026-08-20; **Part D next**: continuous enforcement + `ACCESS_REVOKED`, then E cascade fan-out)* → ⑤ post-Phase-1 integrity pass *(opens once ④ is done: stand up the DB/repo test harness → **composite-FK assignee invariant / ADR-010 first** → then the deferred SQL tests: cascade, rebalance, visibility filter, card_count)*.
- **Integrity pass (⑤) rationale:** the composite-FK invariant (`cards(board_id, assigned_to) → board_members … ON DELETE SET NULL (assigned_to)`, PG16) would make "assignee ∈ board members" a DB guarantee and auto-cascade unassignment on every membership-loss path — superseding the app-layer checks + `RemoveWithParticipationCascade` helpers and the §2.8 shared-helper design. Deferred out of Phase 1 (needs the harness; reopens shipped cascade code). Do it **FK-first** so you don't test cascade SQL you're about to delete. Full write-up: `docs/handoff/handoff-assignee-validation-cascade.md`; memory [[post-phase1-integration-tests]].
- **Frontend architecture settled ✅** — `/grill-with-docs` locked the full frontend design (10 decisions): **Vite SPA** (app zone) + **Astro** (public zone) split (ADR-011, ADR-013); **TanStack Query** (server/REST state) + **Zustand** (WS push + UI state), WS-patches-cache pattern (ADR-012); React Router v7; URL-based card modal; `_authenticated` layout route as auth middleware; feature-based folder structure; shadcn/ui + Tailwind; axios with CSRF interceptor; WS connects once on app mount. Recorded: ADR-011, ADR-012, ADR-013, `docs/architecture/frontend-architecture.md`, SRS §3.1 updated.
- **Don't touch:** Phase-2 deferrals (comments, labels, attachments, public boards, refresh tokens, account deletion) and the frontend implementation (architecture is settled, code not started).



## Core Concepts

- Hierarchy: **Workspace → Board → Column → Card**
- **Workspace Admin authority outranks Board authority.**
- Board visibility: `workspace-visible` or `private`.
- Private boards are visible to workspace admins but require **joining before opening contents**.
- Realtime: board changes (card create/update/delete/move, column reorder, member join/leave) are broadcast to connected clients. Clients are optimistic; **the server is the source of truth.**



## Tech Stack

- Backend: Go, Gin, PostgreSQL (pgx), Google Wire, JWT, WebSocket
- Frontend (app zone): React + Vite SPA, TypeScript, React Router v7, TanStack Query, Zustand, axios, shadcn/ui, Tailwind, dnd-kit
- Frontend (public zone): Astro (SSG) — landing, login, register; React islands for forms



## Docs

- PRD — `docs/product/001-PRD.md`
- User stories — `docs/product/001-user-stories.md`
- Software spec — `docs/spesifications/001-software-specifications.md`
- Backend technical — `collabotask-backend/README.md` (conventions + "add an endpoint" recipe), `collabotask-backend/TESTING.md` (test approach)
- Architecture — `docs/architecture/backend-architecture.md` (backend), `docs/architecture/frontend-architecture.md` (frontend), `docs/architecture/adr/`



### Updating docs (flow downstream: PRD → user-stories → SRS → code)

- `## Now` **above** — every time I switch focus (the only frequently-edited thing).
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