# Software Requirements Specification — CollaboTask (Phase 1)

> **Status:** Refined base for development. This document is the source of truth for Phase 1 scope, product model, and technical contracts. It is aligned with the existing backend codebase (`collabotask-backend/`) and explicitly flags where the current code deviates from the agreed design (see **§9 Audit & Deviations**).

---

## 1. Introduction

### 1.1 Purpose
CollaboTask is a Trello-like task management application: a simpler, focused Kanban tool where users create workspaces, organize work into boards/columns/cards, collaborate in real time, and (in later phases) communicate around cards.

### 1.2 Product Principle (guiding decision)
This is a **learning/portfolio project built to production best-practice standards** — "as if it were a real product." Concretely:

- We do **not** take architecturally-wrong shortcuts (permission model, positioning strategy, API contracts are done properly).
- We **may defer large features**, but deferral must be *clean*: the architecture is not locked out of adding them later, and every "not in Phase 1" decision is recorded explicitly.
- Every "we are not doing X in Phase 1" is a conscious decision, not a forgotten gap.

### 1.3 Scope (Phase 1)
Users can:
- Register / log in (JWT auth).
- Create and manage **workspaces** and invite/manage members.
- Create and manage **boards** (with visibility), columns, and cards.
- Organize cards via drag-and-drop (move within/between columns), reorder columns.
- Collaborate in **real time** over WebSocket (presence + live board mutations).

Out of Phase 1 (see **§8 Scope Matrix**): comments, labels, attachments, activity-log UI, public boards, avatar upload, account deletion/deactivation.

### 1.4 Definitions
- **Workspace**: Top-level container that groups boards for an organization/team. (Note: earlier drafts said "Workplace" — the canonical term is **Workspace**.)
- **Board**: A container of columns/cards for a single project. Has a **visibility** setting.
- **Column**: An ordered lane within a board defining a stage of progress (e.g. To Do / In Progress / Done).
- **Card** / **Ticket**: A work item that holds task info and moves across columns.
- **Workspace Admin (`ADMIN`)**: Full control over a workspace and all its boards. Administrative superuser.
- **Workspace Member (`MEMBER`)**: Can participate in a workspace with limited administrative privileges.
- **Board Owner (`BOARD_OWNER`)**: The single accountable owner of a board (its creator, or whoever ownership was transferred to).
- **Board Member (`BOARD_MEMBER`)**: A user involved in a board (works on cards/columns) but cannot change board settings.

---

## 2. Product Model & Key Decisions

This section captures the decisions that everything else derives from. **Read this first.**

### 2.1 Hierarchy & Membership
```
Workspace ──< Board ──< Column ──< Card
   │             │
   └─ workspace_members (ADMIN/MEMBER)
                 └─ board_members (BOARD_OWNER/BOARD_MEMBER)
```
- "Being a **board member**" means *involvement* (presence, notifications, the board shows in "My Boards"). It is **separate** from *visibility* — a user may be able to *see* a WORKSPACE-visible board without being a member of it.

### 2.2 Board Visibility (PRIVATE / WORKSPACE)
Each board has `visibility ∈ {PRIVATE, WORKSPACE}`, default `WORKSPACE`.

| Visibility | Who can SEE & open the board |
| --- | --- |
| `WORKSPACE` | Every workspace member (and admin) can see metadata, open it, and self-join. |
| `PRIVATE` | Only board members + workspace admins. Admins see **metadata** always, but must **Join** to open the content (break-glass, logged). |

- `PUBLIC` (anyone with a link, unauthenticated) is **Phase 2** — intentionally deferred because it adds share-link + anonymous-access complexity.
- **Any workspace member may create a board** (Trello default). The structural hook for "admins restrict who can create boards" exists conceptually but is not implemented in Phase 1.

### 2.3 Roles & Permissions (the layered model — CRITICAL)
There are two role layers. **Effective permission is computed from BOTH, never from board role alone:**

```
can_administer_board(user, board) =
      user is BOARD_OWNER of that board
   OR user is ADMIN of that board's workspace
```

- A **workspace ADMIN has full power over every board** in the workspace, **even without being a board member**. The `board_members` row for an admin (when they join) only marks *involvement* — it is **not** the source of their authority. Their authority always derives from `workspace_members.role = 'ADMIN'`.
- This is why a workspace admin who joins a PRIVATE board as `BOARD_MEMBER` can still change board settings: the `BOARD_MEMBER` row does not downgrade the power they already hold at the workspace layer.

**Break-glass rule:** A workspace admin can always *see the metadata* of any board (incl. PRIVATE). To act on the **content** of a PRIVATE board, the admin **Joins** first (one click, recorded in the activity log). For WORKSPACE-visible boards, no join is needed. This keeps "admin touched a private board" auditable.

### 2.4 Permission Matrix
"✅(admin)" = allowed because of workspace-admin authority (layered model), independent of board role.

| Action | BOARD_OWNER | BOARD_MEMBER | Workspace ADMIN (even if not board member) | Workspace MEMBER (not on board) |
| --- | --- | --- | --- | --- |
| See board metadata (incl. PRIVATE) | ✅ | ✅ | ✅ | ✅ if WORKSPACE-visible only |
| Open board content (PRIVATE) | ✅ | ✅ | ✅ (after Join, logged) | ❌ |
| Open board content (WORKSPACE) | ✅ | ✅ | ✅ | ✅ |
| Edit board settings / visibility | ✅ | ❌ | ✅ | ❌ |
| Archive / unarchive board | ✅ | ❌ | ✅ | ❌ |
| Delete board | ✅ | ❌ | ✅ | ❌ |
| Invite / remove board member | ✅ | ❌ | ✅ | ❌ |
| Transfer board ownership | ✅ | ❌ | ✅ | ❌ |
| Create / edit / move / delete card & column | ✅ | ✅ | ✅ | ❌ |
| Self-join a board | ✅ | — | ✅ | ✅ if WORKSPACE-visible |

Workspace-level actions:

| Action | Workspace ADMIN | Workspace MEMBER |
| --- | --- | --- |
| Create board | ✅ | ✅ |
| Invite / remove workspace member | ✅ | ❌ |
| Promote/demote member (MEMBER↔ADMIN) | ✅ | ❌ |
| Delete workspace | ✅ (owner only — see 2.6) | ❌ |
| Leave workspace | ✅ (unless last admin/owner) | ✅ |

### 2.5 Board Ownership (single owner + transfer)
- **One board has exactly one `BOARD_OWNER`** (its creator by default). The owner title is single.
- The set of people who can administer a board is still potentially many: `BOARD_OWNER` **plus all workspace admins** (parallel authority via the layered model).
- **Transfer ownership** = the title *moves*: `X (BOARD_OWNER) → BOARD_MEMBER`, `Y (BOARD_MEMBER) → BOARD_OWNER`. It is not "add a co-owner." Target Y must already be a board member. (Phase 1 feature.)
- If X happens to also be a workspace admin, X keeps administrative power after transfer — via the admin layer, not via the (now lost) owner title.

### 2.6 Orphan boards & owner departure
- If a board owner leaves/is removed **without transferring**, the board becomes **owner-less but safe**: workspace admins still fully administer it (layered model). This is *not* an emergency.
- Appointing a new owner is **always a manual action** (no auto-promote). A workspace admin can appoint a new owner (themselves or a board member). For PRIVATE boards, the admin Joins first (logged), then appoints.
- Workspace must always have an owner (`workspaces.owner_id`, `ON DELETE RESTRICT`).

### 2.7 Real-time principle
Real-time is a **layer on every state-changing board action**, not a separate feature.
> **Every Phase-1 action that mutates board state broadcasts its event in Phase 1.**
Actions outside a board view (register, create workspace, promote/demote workspace member) are plain REST — there is no board room watching them. Board membership changes (invite/remove/join/leave/transfer) **do** broadcast, because users viewing the board should see them.

### 2.8 Governance vs Participation (assignment follows participation)
Two distinct capabilities, deliberately separated (aligns with Trello/Jira/ClickUp):
- **Governance** — controlling a board (settings, archive, delete, manage members). Workspace admins hold this via their role **without joining**.
- **Participation** — doing the work: being a card **assignee**, presence, "my boards". Requires **board membership**.

Consequences:
- A card's `assigned_to` **must be a board member** (not merely a workspace member). To assign someone who isn't on the board, add them first (UC-11) — **no auto-add**. The assignee picker lists board members only; the API check is a safety net.
- When a user loses participation — **removed from / leaves a board** (that board only), or **removed from / leaves the workspace** (all its boards) — their assignments on the affected board(s) are cleared (`assigned_to`→NULL) **in the same transaction** as the membership change. **Admins are not exempt:** an admin who leaves a board is unassigned, retaining only governance.
- Each cleared card broadcasts `CARD_UPDATED` (`assigned_to: null`) to that board's live room **after commit**. The membership change is the only activity-log entry; per-card unassigns are not logged.
- The clear-assignments-and-broadcast logic is a **single shared helper** used by UC-06, UC-06c, UC-10, UC-12d.

---

## 3. Technical Decisions

### 3.1 Stack (locked — reflects existing codebase)
**Backend** (`collabotask-backend/`, Go 1.25, clean architecture: `domain → usecase → repository → delivery`):
- HTTP: **Gin** (`gin-gonic/gin`)
- DB access: **pgx/v5** (raw SQL, no ORM) — aligns with the explicit SQL in this doc
- Migrations: **golang-migrate**
- DI: **google/wire**
- Validation: **go-playground/validator**
- Auth: **golang-jwt/v5**
- Logging: **zerolog**
- API docs: **swaggo/swag** (Swagger)
- **Realtime: `coder/websocket`** — *to be added; not yet in `go.mod`.*

**Frontend:** Two-zone split (**[ADR-011](../architecture/adr/adr-011-frontend-framework.md)**):
- *Public zone* — **Astro** (SSG, ADR-013): landing page, login, register. Login/register forms are React islands (`client:load`). No Next.js used in the frontend.
- *App zone* — **React + Vite SPA** (`app.*`, CSR/SPA, auth-gated): the Kanban app. No authenticated SSR in Phase 1 (host-only `__Host-token` cookie cannot be forwarded browser → Next.js server → Go API — ADR-008).
- **WebSocket connects directly to the Go backend**, never via Next.js. Auth via httpOnly cookie at handshake (ADR-008 / §3.5); no `?token=` query param.
- **State management** (**[ADR-012](../architecture/adr/adr-012-frontend-state-management.md)**): **TanStack Query** (server/REST cache) + **Zustand** (WS connection + presence + UI state). WS events patch TanStack Query cache directly — no refetch on board mutations.
- **Router:** React Router v7. **HTTP client:** axios (CSRF header via request interceptor; global 401 redirect via response interceptor). **UI components:** shadcn/ui + Tailwind.
- Drag-and-drop: `dnd-kit`.

**Database:** PostgreSQL.

### 3.2 Positioning strategy — Fractional NUMERIC + rebalancing
Card and column ordering uses **fractional positions** stored as `NUMERIC`, not integer-shift.
- Insert between A and B: `position = (A + B) / 2`. Insert at ends: `first - STEP` / `last + STEP` (e.g. STEP=1000, seed columns at 1000, 2000, 3000…).
- **One UPDATE per move** — other rows are never touched. Concurrency-friendly (no whole-column lock), broadcast payload is tiny ("card X is now at 1500").
- **Rebalancing:** when the gap between two neighbors falls below a threshold (precision exhaustion), rewrite that column's positions back to evenly-spaced values in one transaction. Rare under normal use.
- *Migration required:* change `columns.position` and `cards.position` from `INTEGER` → `NUMERIC`. (Lexorank was considered and rejected as overkill for Phase 1.)
- **Deterministic order:** fetch queries sort by `position ASC, id ASC`. The `id` tiebreak makes a transient position tie resolve identically on every read.
- **Known limitations (accepted for Phase 1):**
  - *Append is not atomic.* Card/column create reads `MAX(position)` then writes `MAX + STEP` in two steps, so two concurrent appends to the same column/board can land on the **same** position. Order stays stable via the `position, id` tiebreak, and the tie self-heals on the next adjacent move (gap 0 < threshold → rebalance). Not worth a lock at Phase-1 scale.
  - *Rebalance can deadlock under contention.* Rebalancing rewrites the whole partition (`UPDATE … WHERE column_id = …`); two threshold-crossing moves in the **same** column at the same instant can deadlock — Postgres aborts one (`40P01`) and the client retries. Rare (rebalancing itself is rare) and recoverable; a retry-on-deadlock wrapper is later hardening, not Phase-1.

### 3.3 Delete strategy
- Workspace and board deletion = **hard delete relying on `ON DELETE CASCADE`** (boards→columns→cards→…). UI must show a confirmation dialog.
- `is_archived` stays as a column + filter (`WHERE is_archived = FALSE`); archive/unarchive **is in Phase 1** (endpoint already exists in code).

### 3.4 FK on-delete policy (per-column, by meaning)
| FK | Policy | Reason |
| --- | --- | --- |
| `workspaces.owner_id` → users | `RESTRICT` | Workspace must always have an owner; deleting such a user is refused. |
| `boards.created_by` → users | `SET NULL` + column **NULLABLE** | `created_by` is only the "original creator" trace; active owner lives in `board_members`. |
| `cards.created_by` → users | `SET NULL` + column **NULLABLE** | **Fixes existing bug** (see §9): current schema declares `SET NULL` on a `NOT NULL` column → would error on delete. |
| `cards.assigned_to` → users | `SET NULL` | Card survives; assignment clears. |
| `workspace_members.user_id`, `board_members.user_id` | `CASCADE` | Membership disappears with the user. |

### 3.5 Auth specifics
- JWT, **24h expiry, no refresh token in Phase 1** (conscious decision; refresh-token rotation is a Phase 2 candidate). *(Matches code default `AUTH_JWT_EXPIRATION=24h`; the cookie `Max-Age` tracks whatever `cfg.JWTExpiration` is, so delivery is expiration-agnostic.)*
- **JWT delivery: httpOnly cookie** (build step ③.5 — **settled, [ADR-008](../architecture/adr/adr-008-auth-httponly-cookie.md)**; supersedes the former `Authorization: Bearer` header). The cookie carries the **same signed JWT** (stateless — session/refresh model is Phase 2); only transport changes.
  - **Cookie:** `__Host-token=<jwt>; HttpOnly; Secure; SameSite=Lax; Path=/; Max-Age=<JWT expiry>` (host-only — no `Domain`; `Secure` + `__Host-` prefix dropped only for non-`localhost` http dev via `COOKIE_SECURE=false`).
  - **Hard cutover:** middleware reads the **cookie only**; login/register **no longer return `token` in the body** (no `Bearer` path). Achieves the XSS goal by construction — no token is exposed to browser JS.
  - **Topology:** cross-origin **sibling subdomains** (`app.` ↔ `api.`); single-origin reverse-proxy (SSR auth + no CORS) is a future enhancement. **No authenticated SSR** in Phase 1 (host-only cookie).
  - **CSRF:** a custom header **`X-CSRF-Protection`** is **required on all state-changing requests** (`POST/PUT/PATCH/DELETE`, incl. `login`/`register`/`logout`) → 403 if absent; the header forces a CORS preflight gated by the origin allowlist (closes the same-site sibling-subdomain vector `SameSite=Lax` leaves open). No CSRF token/cookie needed.
  - **CORS:** explicit origin allowlist (**no `*`** with credentials — startup fails on that combination), `AllowCredentials: true`, methods `GET,POST,PATCH,DELETE,OPTIONS`, headers `Content-Type,X-CSRF-Protection`, `Vary: Origin`. Shared by REST + WebSocket (ADR-009).
- Password hashed with **bcrypt**; min length **8**.
- **No account deletion or deactivation in Phase 1** (see §8 / §10). Therefore the old "Account disabled → 403" error case is **removed** from UC-02.

### 3.6 Conventions (from existing code — the API contract follows these)
- Base path: `/api/v1`. Resource nesting: `/workspace/:workspace_id/board/:board_id/columns/:column_id/cards/:card_id`.
- **Standard response envelope** (all endpoints):
```json
{
  "success": true,
  "status_code": 200,
  "message": "Human readable message",
  "data": { },
  "error": null
}
```
Error shape:
```json
{
  "success": false,
  "status_code": 409,
  "message": "Email already registered",
  "error": { "code": "CONFLICT", "message": "Email already registered", "details": [] }
}
```
- All timestamps are **UTC, ISO-8601** (e.g. `2024-01-30T11:30:00Z`).

---

## 4. Use Cases

> Each use case lists **API** (method + path + auth + permission), request/response (fields match the existing DTOs), error cases, the **SQL** sketch, and the **WS event** (if any). UCs marked **NEW** are Phase 1 but not yet implemented in code.

### 4.1 Authentication

#### UC-01: User Registration
- **API:** `POST /api/v1/auth/register` · Public · requires `X-CSRF-Protection` header
- **Request:** `{ "email", "name", "password" }`
- **Response 201:** `{ "user": { "id", "email", "name", "avatar_url" } }` + **`Set-Cookie: __Host-token=<jwt>; …`** (auto-login; **no `token` in body** — ADR-008 hard cutover). *Auto-login is provisional: email verification (Phase 2+) supersedes it.*
- **Flow:** validate email format (client + server) → normalize email (lowercase + trim) → check email not taken (case-insensitive; the normalized email is both checked and stored) → bcrypt hash → insert user → issue JWT (7d) → **set the auth cookie** → redirect to workspace list.
- **Errors:** Email exists → 409 `CONFLICT`; invalid body → 400 `VALIDATION_ERROR`; missing CSRF header → 403 `FORBIDDEN`.
```sql
SELECT id FROM users WHERE email = $1;
INSERT INTO users (email, password_hash, name) VALUES ($1, $2, $3)
RETURNING id, email, name, avatar_url, created_at;
```

#### UC-02: User Login
- **API:** `POST /api/v1/auth/login` · Public · requires `X-CSRF-Protection` header
- **Request:** `{ "email", "password" }`
- **Response 200:** `{ "user": {…} }` + **`Set-Cookie: __Host-token=<jwt>; …`** (**no `token` in body** — ADR-008 hard cutover).
- **Flow:** normalize email (lowercase + trim) → look up user → verify bcrypt password → issue JWT → **set the auth cookie**. Lookup is case-insensitive (same normalization as registration).
- **Errors:** Email not found → 401 `UNAUTHORIZED`; wrong password → 401 `UNAUTHORIZED`; missing CSRF header → 403 `FORBIDDEN`. *(No "account disabled" case in Phase 1.)*
```sql
SELECT id, email, password_hash, name, avatar_url FROM users WHERE email = $1;
```

#### UC-03: User Logout · **NEW**
- **API:** `POST /api/v1/auth/logout` · Public (no `Auth` middleware) · requires `X-CSRF-Protection` header
- **Request:** empty.
- **Response 200:** `{ }` + **`Set-Cookie: __Host-token=; Max-Age=0; …`** (clears the cookie with matching attributes).
- **Flow:** unconditional / idempotent — clears the auth cookie regardless of session validity (so an expired session can still log out cleanly). Stateless: this removes the browser's copy but does **not** invalidate the JWT server-side (true revocation / "log out all devices" is the Phase-2 session model — ADR-008 §7).
- **Errors:** missing CSRF header → 403 `FORBIDDEN`.

#### UC-02b: Get Profile
- **API:** `GET /api/v1/user/profile` · Auth
- **Response 200:** `{ "id", "email", "name", "avatar_url", "system_role" }`

### 4.2 Workspace Management

#### UC-03: Create Workspace
- **API:** `POST /api/v1/workspace` · Auth
- **Request:** `{ "name", "description"? }` — name: not empty, max 255.
- **Response 201:** workspace object `{ id, name, description, owner_id, role, member_count, board_count, created_at, updated_at }`
- **Flow (1 transaction):** insert workspace (owner=user) → insert `workspace_members(role=ADMIN)`.
```sql
BEGIN;
INSERT INTO workspaces (name, description, owner_id) VALUES ($1,$2,$3) RETURNING *;
INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1,$2,'ADMIN');
COMMIT;
```

#### UC-04: Invite Member(s) to Workspace
- **API:** `POST /api/v1/workspace/:workspace_id/member/invite` · Auth · **Workspace ADMIN**
- **Request:** `{ "emails": ["a@x.com", "b@y.com"] }` (batch supported)
- **Response 200:** per-email results array — each email is processed independently:
  ```json
  { "results": [
      { "email": "a@x.com", "success": true,  "error_code": "" },
      { "email": "b@y.com", "success": false, "error_code": "NOT_FOUND" },
      { "email": "c@z.com", "success": false, "error_code": "CONFLICT" }
  ]}
  ```
  HTTP status is always **200** when the batch runs. Domain errors (`NOT_FOUND`, `CONFLICT`) are surfaced per-email in the payload, not as top-level HTTP errors.
- **Flow:** verify requester is ADMIN (abort with 403 if not) → for each email independently: resolve to user → check workspace membership → insert as `MEMBER` → collect result. Infrastructure errors (DB failures) abort the whole batch with a top-level 5xx.
- **Errors:** not admin → 403 `FORBIDDEN` (top-level, before loop); DB/infra failure → 500 (top-level, aborts batch); per-email domain errors are in the results payload (`NOT_FOUND` / `CONFLICT`), not HTTP status codes.
- **`error_code` values:** `""` = success · `"NOT_FOUND"` = no account for that email · `"CONFLICT"` = already a workspace member.
```sql
SELECT 1 FROM workspace_members WHERE workspace_id=$1 AND user_id=$2 AND role='ADMIN';
-- per email:
SELECT id, email, name FROM users WHERE email = $1;
SELECT 1 FROM workspace_members WHERE workspace_id=$1 AND user_id=$2; -- explicit check (not ON CONFLICT) so CONFLICT is detectable per-email
INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1,$2,'MEMBER');
```

#### UC-05: List User's Workspaces
- **API:** `GET /api/v1/workspace` · Auth
- **Response 200:** array of workspaces with `role` (your role), `member_count`, `board_count` (non-archived), sorted most recent.
```sql
SELECT w.id, w.name, w.description, w.owner_id, wm.role AS role, w.created_at,
  COUNT(DISTINCT b.id) FILTER (WHERE b.is_archived = FALSE) AS board_count,
  COUNT(DISTINCT wm2.user_id) AS member_count
FROM workspaces w
INNER JOIN workspace_members wm ON w.id=wm.workspace_id
LEFT JOIN boards b ON w.id=b.workspace_id
LEFT JOIN workspace_members wm2 ON w.id=wm2.workspace_id
WHERE wm.user_id=$1
GROUP BY w.id, wm.role
ORDER BY w.created_at DESC;
```

#### UC-05b: Get Workspace Detail
- **API:** `GET /api/v1/workspace/:workspace_id` · Auth · workspace member
- **Response 200:** workspace + member list.

#### UC-06: Remove Member from Workspace
- **API:** `DELETE /api/v1/workspace/:workspace_id/member/remove/:user_id` · Auth · **Workspace ADMIN**
- **Flow (1 transaction):** verify requester admin; cannot remove self here → remove from `workspace_members` → cascade-remove from all `board_members` in this workspace → **clear the user's card assignments on every board in this workspace** (assignment = participation, §2.8), capturing affected card ids.
- **Note:** if the removed user was a `BOARD_OWNER`, the board becomes owner-less-but-safe (§2.6).
- **WS:** after commit, `MEMBER_REMOVED` per affected board, plus `CARD_UPDATED` (`assigned_to: null`) per cleared card to that board's live room.
```sql
BEGIN;
DELETE FROM workspace_members WHERE workspace_id=$1 AND user_id=$2;
DELETE FROM board_members bm USING boards b
  WHERE bm.board_id=b.id AND b.workspace_id=$1 AND bm.user_id=$2;
UPDATE cards c SET assigned_to=NULL
  FROM columns col JOIN boards b ON col.board_id=b.id
  WHERE c.column_id=col.id AND b.workspace_id=$1 AND c.assigned_to=$2
  RETURNING c.id, c.column_id;   -- broadcast list (CARD_UPDATED per row)
COMMIT;
```

#### UC-06b: Promote / Demote Workspace Member — **NEW (Phase 1)**
- **API:** `PATCH /api/v1/workspace/:workspace_id/member/:user_id/role` · Auth · **Workspace ADMIN**
- **Request:** `{ "role": "ADMIN" | "MEMBER" }` (validate `oneof=ADMIN MEMBER`)
- **Response 200:** updated member with new role.
- **Rationale:** the layered model relies on "admin as fallback"; we need a way to create more admins.
- **Guards (demote — only when new role = MEMBER and target currently ADMIN):**
  - Requester is not workspace admin → **403** `FORBIDDEN`.
  - Target is not a member → **404** `NOT_FOUND`.
  - Target is the workspace **owner** → **403** `FORBIDDEN` (`ErrCannotDemoteOwner`).
  - Target is the **last** admin → **409** `CONFLICT` (`ErrCannotDemoteLastAdmin`; resolvable: promote another first).
- **Self-demotion is ALLOWED** — an admin may demote themselves, gated by the same two guards above.
- **Idempotent:** setting a role the member already has → 200 no-op, returns the member.
```sql
SELECT COUNT(*) FROM workspace_members WHERE workspace_id=$1 AND role='ADMIN';  -- last-admin guard
UPDATE workspace_members SET role=$3 WHERE workspace_id=$1 AND user_id=$2 RETURNING *;
```

#### UC-06c: Leave Workspace — **NEW (Phase 1)**
- **API:** `POST /api/v1/workspace/:workspace_id/leave` · Auth · member
- **Response 200:** `null` data.
- **Guards:**
  - Requester is not a member → **404** `NOT_FOUND`.
  - Requester is the workspace **owner** → **403** `FORBIDDEN` (`ErrWorkspaceOwnerCannotLeave`; must transfer or delete).
  - Requester is the **last** admin → **409** `CONFLICT` (`ErrLastAdminCannotLeave`; promote another first).
- **Flow:** same cascade as UC-06 for the leaving user — remove `workspace_members`, cascade-remove from all `board_members` in this workspace, and **clear card assignments on every board in this workspace** (§2.8), one transaction. If the leaver was a `BOARD_OWNER`, the board becomes owner-less-but-safe (§2.6).
- **WS (step ④):** `MEMBER_REMOVED` + `CARD_UPDATED` per cleared card after commit.

#### UC-06d: Delete Workspace — **NEW (Phase 1)**
- **API:** `DELETE /api/v1/workspace/:workspace_id` · Auth · **owner only**
- **Response 200:** `null` data.
- **Guard:** requester is not the workspace owner → **403** `FORBIDDEN` (`ErrNotWorkspaceOwner`).
- **Behavior:** hard delete; DB cascades to boards/columns/cards/members. UI confirmation is a frontend concern — backend exposes plain `DELETE`.

### 4.3 Board Management

#### UC-07: Create Board
- **API:** `POST /api/v1/workspace/:workspace_id/board` · Auth · **any workspace member**
- **Request:** `{ "title", "description"?, "background_color"?, "visibility"? }` — `visibility` default `WORKSPACE`. **(`visibility` is NEW — not in code yet.)**
- **Response 201:** board object (see DTO §6).
- **Flow (1 transaction):** verify workspace member → insert board (`created_by`=user) → insert `board_members(role=BOARD_OWNER)` → seed default columns To Do/In Progress/Done with fractional positions (1000/2000/3000).
- **Errors:** not workspace member → 403; empty title → 400.
```sql
SELECT 1 FROM workspace_members WHERE workspace_id=$1 AND user_id=$2;
BEGIN;
INSERT INTO boards (workspace_id, title, description, background_color, visibility, created_by)
  VALUES ($1,$2,$3,$4,$5,$6) RETURNING *;
INSERT INTO board_members (board_id, user_id, role) VALUES ($1,$2,'BOARD_OWNER');
INSERT INTO columns (board_id, title, position) VALUES
  ($1,'To Do',1000),($1,'In Progress',2000),($1,'Done',3000) RETURNING *;
COMMIT;
```

#### UC-08: List Boards in Workspace
- **API:** `GET /api/v1/workspace/:workspace_id/board` · Auth · workspace member
- **Response 200:** for each board: details + `user_role` + `access_status` (`JOINED` | `CAN_JOIN`) + `member_count` + `card_count`.
- **Visibility rule:** admins see all boards; members see WORKSPACE-visible boards + PRIVATE boards they belong to. `access_status` is `CAN_JOIN` for a non-member who may still join (admin on any board, member on a WORKSPACE board); `created_by` is **not** consulted — `board_members` is the sole source of role/access truth (ADR-005).
```sql
-- conceptually: WHERE b.workspace_id=$ws AND b.is_archived=FALSE AND (
--   wm.role='ADMIN' OR b.visibility='WORKSPACE' OR bm.user_id IS NOT NULL )
-- card_count via LEFT JOIN columns → cards, COUNT(DISTINCT c.id)
```

#### UC-09: Admin Joins Board (self-service / break-glass)
- **API:** `POST /api/v1/workspace/:workspace_id/board/:board_id/join` · Auth
- **Who:** workspace admin joining any board, or any workspace member joining a WORKSPACE-visible board.
- **Flow:** validate eligibility → insert `board_members(role=BOARD_MEMBER)` (idempotent) → *(deferred: log activity `MEMBER/JOINED` with `break_glass?` flag per §4.6 → broadcast `USER_JOINED`)*. Logged only on a newly-joined row (idempotent no-op stays silent).
- **Response 200 (idempotent):** `{ "joined": true }` when newly added, or `{ "joined": false, "message": "You are already a member of this board" }` when already a member. A returned row from `ON CONFLICT … DO NOTHING RETURNING` means newly joined; no row means already a member (which also suppresses the future activity/broadcast).
- **Errors:** ineligible → **403** (`ErrBoardCannotJoin`, remapped from 409); a **plain member on a PRIVATE board** → **404** `BOARD_NOT_FOUND` (existence hidden, matching UC-08's filter — not a 403, to avoid leaking existence).
```sql
INSERT INTO board_members (board_id, user_id, role) VALUES ($1,$2,'BOARD_MEMBER')
ON CONFLICT (board_id, user_id) DO NOTHING RETURNING *;
-- (activities insert + USER_JOINED broadcast deferred to the activities/WebSocket sub-steps)
```
- **WS (deferred):** `USER_JOINED`. Self-join is intentionally silent until step ④.

#### UC-10: Leave Board
- **API:** `POST /api/v1/workspace/:workspace_id/board/:board_id/leave` · Auth · board member
- **Guard:** a `BOARD_OWNER` cannot leave without transferring first (board would lose its owner title intentionally → instead require transfer). Admin leaving just drops involvement; their **governance** authority remains.
- **Flow (1 transaction):** delete the `board_members` row → **clear the user's card assignments on this board** (§2.8 — applies to admins too: they lose participation, keep governance), capturing affected card ids.
- **WS:** after commit, `USER_LEFT` + `CARD_UPDATED` (`assigned_to: null`) per cleared card.

#### UC-11: Invite Member(s) to Board
- **API:** `POST /api/v1/workspace/:workspace_id/board/:board_id/invite` · Auth · **can_administer_board** (BOARD_OWNER OR workspace ADMIN)
- **Request:** `{ "user_ids": ["uuid", …] }`
- **Validation:** requester can administer; each invitee is a workspace member; not already a board member.
- **Errors:** not authorized → 403; not workspace member → 400; already board member → 409.
- **WS:** `MEMBER_ADDED` per invitee.

#### UC-11b: Get Invitable Workspace Members
- **API:** `GET /api/v1/workspace/:workspace_id/board/:board_id/invitees` · Auth · can_administer_board
- **Response:** workspace members not yet on the board.

#### UC-12: View Board (Open / Kanban)
- **API (detail):** `GET /api/v1/workspace/:workspace_id/board/:board_id` · Auth
- **API (kanban):** `GET /api/v1/workspace/:workspace_id/board/:board_id/kanban` · Auth — nested columns+cards.
- **Access:** three-method checker (ADR-005) over a shared `resolve()`: `CheckMetadataAccess` (detail), `CheckViewAccess` (kanban), `CheckMutateAccess` (card/column). `created_by` is **not** consulted; `board_members` + workspace-admin status are the sole authority.
- **Break-glass (PRIVATE only, covers view *and* mutate):** a non-joined workspace admin gets **403** `BOARD_JOIN_REQUIRED` on kanban and on any card/column mutation — they must Join first. On WORKSPACE boards admins act with no join.
- **Denied responses (404 hide vs 403 reveal):** an ineligible **plain member** on a **PRIVATE** board gets **404** everywhere (existence hidden, matches the UC-08 list). A plain member denied *mutation* on a **WORKSPACE** board gets **403** (they legitimately see the board). The non-joined admin's PRIVATE denial is **403** `BOARD_JOIN_REQUIRED` (they can already see it in their list).
- **Detail thin roster:** `GET …/:board_id` returns board fields + `access_status`, but **omits the `members` roster** when the board is **PRIVATE and the requester is not joined** (the only such actor is the non-joined admin — plain members 404 here). Joined viewers and everyone on WORKSPACE boards get the full roster.
- **Then:** client opens WS, sends `JOIN_BOARD`, receives `ACTIVE_USERS`.
```sql
-- access check (layered; created_by NOT used):
SELECT 1 FROM boards b
INNER JOIN workspace_members wm ON b.workspace_id=wm.workspace_id AND wm.user_id=$user
LEFT JOIN board_members bm ON b.id=bm.board_id AND bm.user_id=$user
WHERE b.id=$board AND (
  wm.role='ADMIN' OR b.visibility='WORKSPACE' OR bm.user_id IS NOT NULL);
-- then fetch board + columns + cards ORDER BY col.position, c.position
```

#### UC-12b: Update Board Settings
- **API:** `PATCH /api/v1/workspace/:workspace_id/board/:board_id` · Auth · **can_administer_board**
- **Request:** `{ "title"?, "description"?, "background_color"?, "visibility"? }` (`visibility ∈ {PRIVATE, WORKSPACE}`).
- **Visibility flip has no data cascade:** members stay and assignments already require membership. `WORKSPACE→PRIVATE` only affects *future* access.
- **WS (deferred):** `BOARD_UPDATED`.

#### UC-12c: Archive / Unarchive Board (Phase 1 — already in code)
- **API:** `POST /api/v1/workspace/:workspace_id/board/:board_id/archive` · Auth · **can_administer_board**
- **Request:** `{ "is_archived": true|false }`
- **WS:** `BOARD_ARCHIVED` / `BOARD_UNARCHIVED`.

#### UC-12d: Remove Board Member
- **API:** `DELETE /api/v1/workspace/:workspace_id/board/:board_id/member` · Auth · **can_administer_board**
- **Request:** `{ "user_id": "uuid" }`
- **Flow (1 transaction):** delete the `board_members` row → **clear the removed user's card assignments on this board** (§2.8), capturing affected card ids.
- **WS:** after commit, `MEMBER_REMOVED` + `CARD_UPDATED` (`assigned_to: null`) per cleared card.

#### UC-12e: Transfer Board Ownership — **NEW (Phase 1)** — *(see ADR-006)*
- **API:** `POST /api/v1/workspace/:workspace_id/board/:board_id/transfer-ownership` · Auth · **can_administer_board** (current BOARD_OWNER **or** workspace ADMIN)
- **Request:** `{ "to_user_id": "uuid" }` — target must already be a board member.
- **Effect (one operation = "set sole owner"):** demote whoever currently holds `BOARD_OWNER` → `BOARD_MEMBER`, promote the target → `BOARD_OWNER`, in **one transaction**. Preserves the single-owner invariant (§2.5). **Also covers orphan-board appointment (§2.6):** on an owner-less board the demote step matches **0 rows** and the target is simply promoted — no separate "appoint" endpoint. The demote targets the owner **by role**, so the operation is identical whether the requester is the outgoing owner, an admin transferring between two others, or an admin appointing an owner to an orphan board.
- **Ownership source of truth:** `board_members.role` is authoritative for the *current* owner; `created_by` is a historical trace only and is **not** consulted (closes §9 P2; completes the ADR-005 cleanup). This UC also removes the last two `created_by`-as-owner proxies — the `canAdministerBoard()` clause and the `leave_board` owner guard — both switch to the `BOARD_OWNER` board-member row.
- **Break-glass (PRIVATE only):** a non-joined workspace admin gets **403** `BOARD_JOIN_REQUIRED` — they must **Join** first (which also gives them the roster to choose a target). The owner path and WORKSPACE-visible boards need no join (ADR-006; consistent with §2.3, ADR-005).
- **Errors:**
  - Requester cannot administer the board → **403** `ErrBoardPermissionDenied` (plain member who can see the board); a plain member on a **PRIVATE** board they're not on → **404** (existence hidden, per UC-12).
  - Non-joined admin on a **PRIVATE** board → **403** `BOARD_JOIN_REQUIRED`.
  - Target is **not a board member** → **400** `ErrTransferTargetNotBoardMember` (new; mirrors the assignee-must-be-a-board-member rule, §2.8 — "add/join them first").
  - Board archived / not found / workspace mismatch → **404** `ErrBoardNotFound` (same as the membership-mutation family).
  - Missing / malformed `to_user_id` → **400** `VALIDATION_ERROR`.
- **Idempotent:** target already holds `BOARD_OWNER` → **200** no-op (mirrors UC-06b), returns success without touching the repo.
```sql
BEGIN;
-- demote current owner by ROLE (0 rows on an orphan board — appointment case):
UPDATE board_members SET role='BOARD_MEMBER' WHERE board_id=$1 AND role='BOARD_OWNER';
-- promote target by id (asserted to affect exactly 1 row — else target vanished):
UPDATE board_members SET role='BOARD_OWNER'  WHERE board_id=$1 AND user_id=$2;
COMMIT;
```
- **Log activity** → `BOARD/OWNERSHIP_TRANSFERRED`, `entity_id = board_id`, `metadata {from_user_id (null if orphan), to_user_id}` (§4.6). *(Design settled — ADR-007; write lands in the `activities` sub-step of step ③, silent until then.)*
- **WS (deferred to step ④):** `OWNERSHIP_TRANSFERRED` `{ board_id, from_user_id, to_user_id }` (`from_user_id` is the demoted owner, which may differ from the requester; null on an orphan appointment).

### 4.4 Column & Card Management

> All actions in this section require board **content access** (BOARD_OWNER, BOARD_MEMBER, or workspace ADMIN) via the shared `boardAccessChecker`.

#### UC-13: Create Column
- **API:** `POST /api/v1/workspace/:ws/board/:board_id/columns` · Auth
- **Request:** `{ "title" }` — position computed as `last + STEP` (fractional).
- **WS:** `COLUMN_CREATED`.

#### UC-13b: Update Column — **(in code)**
- **API:** `PATCH /…/columns/:column_id` · `{ "title" }` · **WS:** `COLUMN_UPDATED`.

#### UC-13c: Delete Column — **(in code)**
- **API:** `DELETE /…/columns/:column_id` — cascades to its cards. Requires confirmation in UI. **WS:** `COLUMN_DELETED`.

#### UC-13d: Reorder Column
- **API:** `PATCH /…/columns/:column_id/position` · `{ "position" }`
- **Target model:** fractional — compute midpoint between neighbors; 1 UPDATE. **WS:** `COLUMN_MOVED`.

#### UC-14: Create Card
- **API:** `POST /…/columns/:column_id/cards` · Auth
- **Request:** `{ "title", "description"?, "assigned_to"?, "due_date"? }` — title required, max 500. `assigned_to` **must be a board member** (assignment = participation, §2.8; see §9 #7).
- **WS:** `CARD_CREATED`.

#### UC-15: Move Card (Drag & Drop)
- **API:** `POST /…/columns/:column_id/cards/:card_id/move` · Auth
- **Request:** `{ "to_column_id", "to_position" }` — `to_position` is the **fractional** target (midpoint of the drop neighbors, computed client-side or server-side).
- **Flow:** optimistic UI → REST/WS move → server validates + updates **one row** → broadcast `CARD_MOVED` → on reject, client reverts.
```sql
-- fractional move: single-row update, no neighbor shifting
UPDATE cards SET column_id=$to_col, position=$new_pos, updated_at=now()
WHERE id=$card RETURNING *;
-- (rebalance the target column only if neighbor gap < threshold)
```
- **WS:** `CARD_MOVED`.

#### UC-16: Update Card
- **API:** `PATCH /…/cards/:card_id` · `{ "title"?, "description"?, "assigned_to"?, "due_date"? }` (assigned_to nullable to unassign; if set, **must be a board member** — §2.8, else 400 `ErrAssigneeNotBoardMember`). **WS:** `CARD_UPDATED`.
- **Assignee validation (Decision B):** the board-membership gate fires **only on a newly-set assignee** (`assigned_to` present and non-null). An *unchanged* existing assignee is never re-validated — it is re-resolved display-only for the response — so a stale assignee (one who since left the board) never blocks an unrelated edit (e.g. a title fix). Honesty of existing assignments is maintained by the unassign cascade (UC-10/UC-12d), not the write-time check.

#### UC-17: Delete Card
- **API:** `DELETE /…/cards/:card_id` — confirm in UI. With fractional positions, no neighbor reorder needed. **WS:** `CARD_DELETED`.

### 4.5 Real-time Collaboration

> **Design settled — [ADR-009](../architecture/adr/adr-009-websocket-realtime-layer.md); build step ④** (after ③.5 auth-cookie). Broadcasts originate in the **usecase layer** via a `Broadcaster` port, best-effort + after-commit (swallow-on-failure, exactly like `WriteActivity`/ADR-007). Mutations stay **REST-in, server-broadcasts-out** (§5.1); the WS layer adds broadcast taps. Full build plan: `docs/handoff/websocket-participation-broadcast/index.md`.

#### UC-18: WebSocket Lifecycle
- **Endpoint:** `GET /api/v1/ws` · Auth — **NEW, not in code.** Authenticated from the **httpOnly JWT cookie** (ADR-008 / §3.5): `Accept` reads the cookie → `ValidateToken` → reject `401` **before** upgrade. (No `?token=` query param, no subprotocol.)
- **Server write-timeout:** the shared `http.Server`'s 30s `WriteTimeout` (§3.1) would kill a long-lived idle WS — cleared for *this connection only* via `http.ResponseController`; REST keeps its 30s. WS liveness is **server-driven ping/pong**, not a fixed deadline.
- **Origin (CSWSH):** coder/websocket's `OriginPatterns` is fed the **explicit** `CORS_ALLOWED_ORIGINS` shared with REST CORS (cookie auth forbids `*`). Never `InsecureSkipVerify`.
- **Connect:** authenticate → register connection → client sends `JOIN_BOARD{board_id}` → server re-validates board access via **`CheckViewAccess`** (room visibility == kanban visibility; break-glass falls out) → add to in-memory room → broadcast `USER_JOINED` (on the user's 0→1 edge) → send `ACTIVE_USERS` to the joiner.
- **Disconnect:** remove connection from all its rooms → broadcast `USER_LEFT` (on the user's 1→0 edge).
- **Reconnect:** client exponential backoff (immediate, 1s, 2s, 4s, … max 30s) → re-`JOIN_BOARD` → client refetches board state via REST and reconciles.

#### UC-19: Presence (multi-connection, edge-triggered)
- Server keeps in-memory rooms **keyed by user with a *set* of connections**: `map[board_id] → map[user_id] → {conn_id → connection}` (supersedes the earlier single-connection sketch — correct under multi-tab/device and reconnect; ADR-009).
- Presence is **edge-triggered**: `USER_JOINED` fires only on a user's **0→1** connection edge in a room, `USER_LEFT` only on **1→0**; broadcasts fan out to **all** of a user's connections. `ACTIVE_USERS` (sent to the joiner) lists **distinct connected user_ids** — built from the in-memory room, never the DB roster (presence = who's connected, not who's a member).

#### UC-19b: Continuous Access Enforcement (instant eviction)
- Access is enforced **continuously, not just at join.** When a connection loses access it is **evicted immediately** from the affected room(s) via one of three hub primitives:
  - `EvictUser(board, X)` — removed from / leaves a board (UC-10, UC-12d).
  - `EvictExcept(board, members+admins)` — board flips **WORKSPACE→PRIVATE** (UC-12b): evict everyone not a member/admin. (`PRIVATE→WORKSPACE` needs no eviction — access only widens.)
  - `EvictUserFromRooms(X, workspace's boards)` — removed from / leaves the workspace (UC-06, UC-06c); also closes the **workspace-visible-watcher leak** (X watching a WORKSPACE board without membership → not in `AffectedBoardIDs`).
- **Involuntary** loss (removal, flip) sends the evicted connection a new **`ACCESS_REVOKED{board_id, reason}`** event (§5.2) — its single "leave this board view" signal. **Voluntary** leaves (UC-10, UC-06c) evict **silently** (`USER_LEFT` covers the room). Cascades use **evict-first ordering** (evict → `MEMBER_REMOVED` → per-card `CARD_UPDATED`) so room broadcasts exclude the departing user.

#### UC-20: Optimistic UI
- Client applies the change immediately, sends to server; on success other clients update, on rejection the originating client reverts and shows an error. Broadcasts **include the sender** (§5.2) — a REST mutation carries no WS connection identity to exclude, and the sender's other tabs must update; clients dedupe against their authoritative REST response.

### 4.6 Activity Logging (writes — Phase 1; UI = UC-22, Phase 2)

Every board-scoped, state-changing mutation writes **one `activities` row** (SRS §7). This is the authoritative contract for those writes; the individual UCs above reference it rather than repeating it. **Full rationale + alternatives: [ADR-007](../architecture/adr/adr-007-activity-logging-writes.md).** The activity-log **reader** (UC-22 feed) is Phase 2 — this section governs **writes only**.

**Write model (ADR-007, Option A):** best-effort, **synchronous**, **after the mutation commits**, orchestrated in the usecase layer. On write failure → log (zerolog) and **swallow** — a logging failure must never fail or roll back the user's action. It is **not atomic** with the mutation (the codebase has no shared transaction abstraction); true atomicity via a Unit-of-Work/Transactor is deferred to the post-Phase-1 integrity pass ⑤ (see §9 + [[activities-logging-best-effort-then-atomic]]). Accepts a rare, *logged* gap where a mutation commits but its activity row is lost.

**Firing rule:** **log only on an actual state change.** No-op update/move/archive, an idempotent re-join (`joined:false`), and a no-op ownership transfer write **nothing** (and therefore fire no `break_glass`).

**Vocabulary (app-enforced enum — no DB `CHECK`):**
- `entity_type ∈ {BOARD, COLUMN, CARD, MEMBER}`
- `action_type ∈ {CREATED, UPDATED, DELETED, MOVED, ARCHIVED, UNARCHIVED, JOINED, LEFT, ADDED, REMOVED, OWNERSHIP_TRANSFERRED}`
- This is a **verb × entity split** (why §7 has two columns), kept deliberately separate from the §5.2 WebSocket event names.

**`entity_id`** = the id of the named entity: `CARD`→card, `COLUMN`→column, `BOARD`→board, `MEMBER`→the **target** user (the actor is always `user_id`). Ownership transfer is a **`BOARD`** event (not `MEMBER`), with both users in metadata.

**Metadata (snapshot principle):** snapshot the human-readable label of any entity that can be **deleted or renamed** in Phase 1 — i.e. `card_title` / `column_title` (`activities` has no FK to cards/columns, so a hard-deleted entity leaves a *dangling* `entity_id`; the snapshot also keeps the row audit-faithful to its moment). Users are **ids-only** (never deleted or renamed in Phase 1 → resolve at read). `UPDATE` stores `changed_fields` **names only**, not old→new diffs. Two audit-enrichment flags: `break_glass: true` on `MEMBER/JOINED` when a non-member workspace admin opens a **PRIVATE** board (UC-09, §2.3); `source ∈ {board, workspace}` on `MEMBER/LEFT|REMOVED` to separate a direct board action from a workspace-cascade departure.

**Mutation → activity map:**

| UC | Row(s) | `entity_type`/`action_type` | `entity_id` | metadata | Fires when |
| --- | --- | --- | --- | --- | --- |
| UC-14 | Create card | CARD/CREATED | card | `{card_title}` | always |
| UC-15 | Move card | CARD/MOVED | card | `{card_title, from_column_id, from_column_title, to_column_id, to_column_title}` | column or position changed |
| UC-16 | Update card | CARD/UPDATED | card | `{card_title, changed_fields}` | `changed_fields ≠ ∅` |
| UC-17 | Delete card | CARD/DELETED | card | `{card_title}` | always |
| UC-13 | Create column | COLUMN/CREATED | column | `{column_title}` | always |
| UC-13b | Update column | COLUMN/UPDATED | column | `{column_title, changed_fields}` | changed |
| UC-13c | Delete column | COLUMN/DELETED | column | `{column_title}` | always |
| UC-13d | Reorder column | COLUMN/MOVED | column | `{column_title}` | position changed |
| UC-12b | Update board | BOARD/UPDATED | board | `{board_title, changed_fields, visibility_from?, visibility_to?}` | changed |
| UC-12c | Archive / unarchive | BOARD/ARCHIVED \| UNARCHIVED | board | `{}` | state changed |
| UC-12e | Transfer ownership | BOARD/OWNERSHIP_TRANSFERRED | board | `{from_user_id (null if orphan), to_user_id}` | real transfer (not no-op) |
| UC-09 | Join / break-glass | MEMBER/JOINED | joiner | `{role, break_glass?}` | newly joined only |
| UC-11 | Invite to board | MEMBER/ADDED ×invitee | invitee | `{role}` | per newly-added |
| UC-10 | Leave board | MEMBER/LEFT | leaver | `{source:"board"}` | always |
| UC-12d | Remove board member | MEMBER/REMOVED | removed | `{source:"board"}` | always |
| UC-06 | Remove workspace member | MEMBER/REMOVED **×affected board** | removed | `{source:"workspace"}` | per affected board |
| UC-06c | Leave workspace | MEMBER/LEFT **×affected board** | leaver | `{source:"workspace"}` | per affected board |

**Scope — board-scoped only.** Workspace-entity actions (UC-03/04/05/06b/06d) are **not** logged (`activities.board_id` is `NOT NULL` — they have no board). Their **board cascades** (UC-06/UC-06c) still emit one board-scoped row **per affected board**, which requires the workspace-cascade repo methods to also `RETURNING board_id`.

**Not logged:** the **per-card unassigns** from any participation cascade (§2.8 — those are a step-④ `CARD_UPDATED` *broadcast*, never an activity row); no-op requests (firing rule); all workspace-entity actions.

**Index:** `CREATE INDEX ... ON activities (board_id, created_at DESC)` ships in the write-phase migration (`000008`) — the Phase-2 read pattern (`WHERE board_id=$1 ORDER BY created_at DESC`) is fully known.

---

## 5. WebSocket Message Contract

All messages share the envelope `{ "type": "<TYPE>", "payload": { … } }`. Timestamps UTC ISO-8601.

> **Handshake auth:** the connection is authenticated from the **httpOnly JWT cookie** at upgrade time (ADR-008 / §4.5 UC-18) — there is no auth message. Client→server messages are **only** the two below; all mutations go over REST (§5.1 note).

### 5.1 Client → Server
| Type | Payload | Notes |
| --- | --- | --- |
| `JOIN_BOARD` | `{ board_id }` | Join a board room; access re-validated via `CheckViewAccess`. |
| `LEAVE_BOARD` | `{ board_id }` | Optional explicit leave. |

> Card/column mutations may go over REST **or** WS depending on implementation; the **authoritative broadcast** below is what every client must handle. (Recommended: mutations via REST, server broadcasts the resulting event — simpler and reuses existing handlers.)

### 5.2 Server → Clients (broadcast to board room)
| Type | Payload (key fields) | Emitted by |
| --- | --- | --- |
| `USER_JOINED` | `{ board_id, user{id,name,avatar_url}, timestamp }` | UC-09, UC-18 |
| `USER_LEFT` | `{ board_id, user_id, timestamp }` | UC-10, UC-18 |
| `ACTIVE_USERS` | `{ board_id, users:[{id,name,avatar_url,joined_at}] }` | sent to joiner (UC-19) |
| `ACCESS_REVOKED` | `{ board_id, reason }` | sent to an **involuntarily** evicted connection (UC-19b) — removal or WORKSPACE→PRIVATE flip. Voluntary leaves evict silently. |
| `MEMBER_ADDED` | `{ board_id, user{…}, role }` | UC-11 |
| `MEMBER_REMOVED` | `{ board_id, user_id }` | UC-12d |
| `OWNERSHIP_TRANSFERRED` | `{ board_id, from_user_id, to_user_id }` | UC-12e |
| `BOARD_UPDATED` | `{ board_id, fields:{…} }` | UC-12b |
| `BOARD_ARCHIVED` / `BOARD_UNARCHIVED` | `{ board_id }` | UC-12c |
| `COLUMN_CREATED` | `{ column{ id,board_id,title,position } }` | UC-13 |
| `COLUMN_UPDATED` | `{ column_id, title }` | UC-13b |
| `COLUMN_DELETED` | `{ column_id }` | UC-13c |
| `COLUMN_MOVED` | `{ column_id, position }` | UC-13d |
| `CARD_CREATED` | `{ card{…} }` | UC-14 |
| `CARD_MOVED` | `{ card_id, from_column_id, to_column_id, position }` | UC-15 |
| `CARD_UPDATED` | `{ card_id, fields:{…} }` | UC-16 |
| `CARD_DELETED` | `{ card_id, column_id }` | UC-17 |

**Broadcast rule:** by default broadcast to all room members **including** the sender (so optimistic clients reconcile against the authoritative payload), except `ACTIVE_USERS` which is sent only to the joiner.

### 5.3 Client handling (presence & join frames)

The §5.2 Broadcast rule and UC-20 govern how a client reconciles **mutation** frames (dedupe against the authoritative REST response). The **presence / join** frames carry their own client obligations — settled during the step-④ build (ADR-009) and recorded here so the frontend treats them correctly:

- **Gate the board view on the REST kanban fetch, never on `ACTIVE_USERS`.** A denied `JOIN_BOARD` is **silent** — no error frame (this preserves the 404-hide invariant, §2.2). The authoritative 403/404 comes from `GET .../kanban` (same `CheckViewAccess`, §4.5 UC-18). `ACTIVE_USERS` is a presence bonus on an already-granted view — never show a spinner waiting for it.
- **`ACTIVE_USERS` is an authoritative full snapshot — apply as REPLACE**, even when presence state is already held. Never skip the snapshot because data already exists.
- **`USER_JOINED` / `USER_LEFT` are idempotent deltas.** Adding an already-present user or removing an absent one is a no-op. Their ordering relative to the snapshot is **not** guaranteed (third-party joins race) — the `ACTIVE_USERS` snapshot is the source of truth.
- **A client may receive `USER_JOINED` for itself** on its own 0→1 join (self-echo). It is an idempotent no-op — relevant only to suppress a self "joined" toast (`user.id == self`).

---

## 6. API Response DTOs (match existing code)

| Entity | JSON fields |
| --- | --- |
| **User** | `id, email, name, avatar_url, system_role` |
| **AuthResult** | `user{…}, token` |
| **Workspace** | `id, name, description, owner_id, role, member_count, board_count, created_at, updated_at` |
| **Board** | `id, workspace_id, title, description, created_by, is_archived, background_color, visibility, user_role, access_status, member_count, card_count, created_at, updated_at` *(visibility, card_count = NEW; list view omits `members`, and detail omits `members` when PRIVATE + not-joined)* |
| **Self-join result** | `joined` (bool), `message?` (present only when already a member) |
| **Column** | `id, board_id, title, position, created_at, updated_at` |
| **Card** | `id, column_id, title, description, position, assigned_to{id,name,avatar_url}, due_date, created_by, created_at, updated_at` |

---

## 7. Database Schema (target state)

> Reflects current migrations + the Phase-1 changes agreed here. Changes vs current code are flagged **[CHANGE]** / **[NEW]**.

**users**: `id uuid pk`, `email varchar(255) unique not null`, `password_hash varchar(255) not null`, `name varchar(255) not null`, `avatar_url varchar(500) null`, `system_role varchar(50) not null default 'USER' check (USER|SUPER_ADMIN)`, `created_at`, `updated_at`.

**workspaces**: `id uuid pk`, `name varchar(255) not null`, `description text null`, `owner_id uuid not null references users on delete RESTRICT`, `created_at`, `updated_at`.

**workspace_members**: pk `(workspace_id, user_id)`, `workspace_id → workspaces cascade`, `user_id → users cascade`, `role varchar(50) not null default 'MEMBER' check (ADMIN|MEMBER)`, `joined_at`.

**boards**: `id uuid pk`, `workspace_id → workspaces cascade not null`, `title varchar(255) not null`, `description text null`, `background_color varchar(8) not null default '#0079BF'`, `is_archived boolean not null default false`, **[NEW]** `visibility varchar(20) not null default 'WORKSPACE' check (PRIVATE|WORKSPACE)`, `created_by uuid references users on delete SET NULL` **[CHANGE: NULLABLE + SET NULL]**, `created_at`, `updated_at`.

**board_members**: pk `(board_id, user_id)`, cascades, `role varchar(50) not null default 'BOARD_MEMBER' check (BOARD_OWNER|BOARD_MEMBER)`, `joined_at`.

**columns**: `id uuid pk`, `board_id → boards cascade not null`, `title varchar(255) not null`, **[CHANGE]** `position NUMERIC not null` (was INTEGER), `created_at`, `updated_at`.

**cards**: `id uuid pk`, `column_id → columns cascade not null`, `title varchar(500) not null`, `description text null`, **[CHANGE]** `position NUMERIC not null` (was INTEGER), `assigned_to uuid null references users on delete SET NULL`, `due_date timestamp null`, `created_by uuid null references users on delete SET NULL` (NULLABLE — applied in migration 000005), `created_at`, `updated_at`.

**activities** **[NEW — written in Phase 1, UI in Phase 2; migration `000008`]**: `id uuid pk`, `board_id → boards cascade not null`, `user_id → users on delete SET NULL`, `action_type varchar not null`, `entity_type varchar not null`, `entity_id uuid not null`, `metadata jsonb default '{}'`, `created_at`. Plus index `(board_id, created_at DESC)` for the Phase-2 feed. `action_type`/`entity_type` are an **app-enforced** verb×entity vocabulary (no DB `CHECK`); `metadata` is app-shaped jsonb governed by the **snapshot principle** (snapshot card/column titles; users ids-only). Full write contract in **§4.6**; rationale in **[ADR-007](../architecture/adr/adr-007-activity-logging-writes.md)**.

**Phase-2 tables (create later, kept here for schema awareness):** `comments`, `labels`, `card_labels`, `attachments`.

---

## 8. Scope Matrix

| Feature | Phase 1 | Phase 2 / Later |
| --- | --- | --- |
| Auth (register/login/profile, JWT 7d, no refresh) | ✅ | refresh-token rotation |
| Workspace: create/list/detail/invite/remove member | ✅ | |
| Workspace: promote/demote, leave, delete | ✅ (NEW) | |
| Board: create (+visibility WORKSPACE/PRIVATE)/list/view/kanban | ✅ | board `PUBLIC` via link |
| Board: update/archive/invite/remove/join/leave | ✅ | |
| Board: transfer ownership | ✅ (NEW) | |
| Column: create/update/delete/reorder | ✅ | |
| Card: create/move/update/delete | ✅ | |
| Fractional positioning | ✅ (CHANGE) | lexorank (only if scale demands) |
| Realtime WS (presence + all board mutations) | ✅ (NEW) | multi-instance via Redis pub/sub |
| Activity logging (write) | ✅ | activity-log **UI** (UC-22) |
| Comments | ❌ | ✅ |
| Labels / card labels | ❌ | ✅ |
| Attachments | ❌ | ✅ (needs file storage) |
| Avatar upload | ❌ | ✅ (needs file storage) |
| Account delete / deactivate | ❌ | ✅ (see §10) |

---

## 9. Audit & Deviations (current code vs this spec)

> Findings from a code audit of `collabotask-backend/`. These are the items to fix so the implementation matches the agreed design. **Priority ordered.**

**✅ FIXED (P0) — Layered permission not enforced for board *management*** (board *access* was already correct):
- `utils.go` — renamed `canManageBoardMembers()` → **`canAdministerBoard()`**; admin clause is now `workspaceMember.IsAdmin()` alone (no longer requires a board-member row). Fixes invite/remove/invitees, which route through it.
- `update_board.go` + `set_archived.go` — now fetch the workspace member, tolerate a non-member admin (`ErrBoardMemberNotFound` → nil), and gate on `canAdministerBoard()`. *(Discovery: `set_archived.go` had the identical owner-only bug; fixed for consistency per this item's note.)* Transfer-ownership (UC-12e) will use the same helper when built.

**✅ FIXED (P0) — `cards.created_by` constraint bug:** was `NOT NULL … ON DELETE SET NULL` (contradictory). Resolved by **migration `000005_make_cards_created_by_nullable`** (`ALTER … DROP NOT NULL`) — original `000004` left intact since migrations are immutable once applied.

**🟡 P1 — Missing Phase-1 features (need to be built):** ~~board `visibility` column + logic~~ **✅ DONE** (ADR-005, migration 000007 — three-method checker, break-glass, thin roster, idempotent self-join); ~~promote/demote (UC-06b); leave workspace (UC-06c); delete workspace (UC-06d)~~ **✅ DONE** (workspace-scoped participation cascade included — board-scoped deferred); ~~ownership transfer (UC-12e)~~ **✅ DONE** (ADR-006 — `created_by` proxy removed, atomic demote+promote repo method, break-glass, idempotent no-op, `ErrTransferTargetNotBoardMember`); **WebSocket layer entirely** (UC-18/19); ~~`activities` table + writes~~ **✅ DONE** (ADR-007, migration 000008 — 17 call sites, Option A best-effort, 438 tests pass).

**✅ FIXED (P1) — Positioning migrated off INTEGER-shift:** card move and column reorder now use fractional NUMERIC coordinates with repository-layer rebalancing (**ADR-004**, migration **000006_fractional_positioning**). One UPDATE per move; the integer-shift methods (`IncrementPositionsFrom`, `DecrementPositionsAfter`, `ReorderPositions`, `DeleteWithReorder`) and the `max+1` clamp are removed. See §3.2 for the contract + accepted limitations. *Repository-layer rebalance tests deferred — no DB-backed test harness exists yet.*

**✅ FIXED (P1) — Assignee not validated as a board member:** ~~`create_card.go`, `update_card.go` check the user exists but not that they are a **board member** of the card's board~~ — both now gate `assigned_to` through `boardMemberRepo.IsUserExists` and return **400 `ErrAssigneeNotBoardMember`** for a non-member (assignment = participation, §2.8). Collapses the old "user exists" 404 path into a single validation error (Decision C). **Decision B:** `update_card` validates only a *newly-set* assignee (`AssignedToPresent && != nil`); a stale existing assignee is re-resolved display-only and never blocks an unrelated edit — the board-scoped unassign cascade (UC-10/UC-12d) keeps assignees honest instead. TOCTOU window accepted for Phase 1 (§2.8 "safety net"); DB-level enforcement is the deferred composite-FK candidate (see integrity-pass note below). *(Superseded the earlier "validate against workspace" framing — board membership is the stricter, correct rule.)*

**🔵 Deferred (integrity pass) — composite-FK assignee invariant (ADR-010 candidate):** the "assignee ∈ board members" rule is enforced at the **application layer** today (write-time `IsUserExists` check + explicit `RemoveWithParticipationCascade`). A stronger design enforces it in the **schema**: add `board_id` to `cards` (backfill, NOT NULL), then `cards(board_id, assigned_to) → board_members(board_id, user_id) ON DELETE SET NULL (assigned_to)` (PG16). That would make assignee-validity a DB guarantee on every write path (closing the TOCTOU race) and make the unassign-cascade automatic on *every* membership-loss path — subsuming both the workspace and board app-layer cascades. **Deferred out of Phase 1:** it's a schema/architecture change beyond "data-correction only", reopens shipped cascade code, still needs the explicit `RETURNING` list for the step-④ broadcast, and wants SQL-level tests the project has deferred to the post-Phase-1 integrity pass. Headline candidate for that pass; write up as `adr-010` **when adopted** (candidate, not a decision). *(Renumbered 008→010: ADR-008 is now **auth httpOnly-cookie** and ADR-009 is the **WebSocket realtime layer** — both Phase-1 build steps that landed ahead of this deferred candidate.)* See [[post-phase1-integration-tests]].

**✅ RESOLVED (P2) — Two sources of truth for board ownership:** ~~owner is implied by both `boards.created_by` and a `board_members(BOARD_OWNER)` row~~ — `board_members.role` is now the **sole** ownership authority (ADR-006). `created_by` is a historical-creator trace only; the `canAdministerBoard()` proxy check and the `leave_board` creator guard have been replaced with role-based checks as part of UC-12e.

**🔵 P2 — Transaction boundaries:** ensure create-board and create-workspace wrap their multi-insert in a single transaction (avoid orphaned rows on partial failure).

**🔵 P2 — REST naming nit:** `DELETE /workspace/:id/member/remove/:user_id` embeds a verb. Prefer `DELETE /workspace/:id/members/:user_id`. (Cosmetic; align when convenient.)

**🔵 P2 — Self-join validation:** confirm `self_join_board.go` verifies the joiner is a workspace member and the board is WORKSPACE-visible (or the joiner is an admin) before inserting.

### 9.1 Suggested build order
Correctness/security first, then features. (Story IDs reference [001-user-stories.md](./001-user-stories.md).)
1. ~~**P0 fixes:** layered permission for board management + `cards.created_by` NULLABLE bug.~~ **✅ DONE** (`canAdministerBoard` helper across update/archive/invite/remove; migration 000005). Tests pending — approach TBD.
2. ~~**Fractional positioning:** migrate card-move and column-reorder off integer-shift to NUMERIC (§3.2).~~ **✅ DONE** (ADR-004, migration 000006). Repository-layer rebalance tests deferred (no DB harness yet).
3. **REST features — complete the Phase 1 surface before realtime.** Build order within this step: visibility first (cross-cutting), then workspace ops, then board ops, then assignee validation + cascade, then activities.
   - ~~Board `visibility` column + access logic (UC-07, UC-08, UC-09, UC-12).~~ **✅ DONE** (ADR-005, migration 000007). Three-method access checker (metadata/view/mutate) + PRIVATE break-glass, 404-hide vs 403-reveal, idempotent self-join (`joined` flag), thin roster, `created_by` removed from access/role logic. SQL filter + `card_count` tested at the usecase layer only (repo-layer deferred to post-Phase-1 integration pass).
   - ~~Promote/demote workspace member (UC-06b), leave workspace (UC-06c), delete workspace (UC-06d).~~ **✅ DONE.** **Deviation from original plan:** the **workspace-scoped** participation cascade (remove from all `board_members` + clear card assignments across the workspace) was built in this sub-step alongside UC-06c and UC-06 retrofit, rather than deferring to the cascade sub-step below. The cascade has no WebSocket dependency; only the *broadcast* waits for step ④. Board-scoped cascade (UC-10, UC-12d) still lands with board leave/remove work.
   - ~~Ownership transfer (UC-12e)~~ **✅ DONE** (ADR-006). `created_by` proxy removed from `canAdministerBoard()` + `leave_board`; atomic demote-by-role + promote-by-id in one TX; break-glass via `CheckMutateAccess` ∘ `canAdministerBoard`; `ErrTransferTargetNotBoardMember` (400); idempotent no-op when target already owns; orphan-safe. Activity write + `OWNERSHIP_TRANSFERRED` broadcast **deferred** (activities sub-step / step ④).
   - ~~Assignee = board member validation (UC-14/UC-16) + unassign-on-lost-participation data correction (UC-10, UC-12d — board-scoped).~~ **✅ DONE.** Write-time gate via `boardMemberRepo.IsUserExists` → `ErrAssigneeNotBoardMember` (400), newly-set assignee only (Decision B); atomic `RemoveWithParticipationCascade` on `BoardMemberRepository` (delete membership + `UPDATE cards SET assigned_to = NULL … RETURNING` in one TX), wired into `leave_board` + board `remove_member`. No WebSocket dependency — the `CARD_UPDATED` broadcast for the cleared cards waits for step ④ (cascade already returns `[]AffectedCard`).
   - ~~`activities` table migration + write an activity row on every (board-scoped) mutation.~~ **✅ DONE** (ADR-007, migration 000008). Best-effort synchronous after-commit write (Option A — `common.WriteActivity` swallows errors, never fails the mutation); log-only-on-state-change (no-op requests stay silent); verb×entity vocabulary (§4.6); snapshot-title metadata; board-scoped incl. per-affected-board rows on the workspace cascade (`AffectedBoardIDs` from `RemoveWithParticipationCascade`). 17 call sites across card/column/board/workspace usecases; 163 new activity-contract test cases (438 total passing). `CARD_UPDATED` broadcast for cascade-cleared cards and Transactor atomicity remain deferred to step ④ / ⑤.
3.5. ~~**Auth: `Bearer` → httpOnly cookie — NEW prerequisite, before ④.**~~ **✅ DONE** (2026-07-27; full record **[ADR-008](../architecture/adr/adr-008-auth-httponly-cookie.md)**, contract SRS §3.5/§4.1). Moved JWT delivery from the JS-held `Authorization: Bearer` header to an **httpOnly, Secure, SameSite=Lax `__Host-` cookie**, unifying auth across REST + WebSocket + SSR and forcing explicit CORS origins. Hard cutover (login/register drop body `token`); cookie-reading `Auth` middleware; new `X-CSRF-Protection` CSRF gate on all mutations incl. `/auth/*`; unconditional `POST /auth/logout`; CORS hardening (explicit allowlist, `PATCH`, `Vary: Origin`, fail-fast on `*`+credentials). Surfaced while grilling ④ (a browser `WebSocket` can't set headers; SSR can't read `localStorage`; JS-readable tokens are XSS-exposed). 473 tests pass; two-axis `/code-review` clean (0 hard standards violations, 0 spec gaps); P1/P2 doc-string + comment follow-ups applied. **④ now unblocked — recheck its handoff for drift.**
4. **WebSocket layer + participation broadcast — design settled ([ADR-009](../architecture/adr/adr-009-websocket-realtime-layer.md); handoff `docs/handoff/websocket-participation-broadcast/`).** Full realtime layer (UC-18/19); `Broadcaster` port (usecase-layer, best-effort after-commit like activities); wire broadcast calls (`CARD_UPDATED`, `CARD_MOVED`, `COLUMN_MOVED`, etc.) onto the already-correct mutations; multi-connection edge-triggered presence; **continuous access enforcement** (instant eviction + `ACCESS_REVOKED`, UC-19b); add the `CARD_UPDATED` broadcast for the unassign cascade completed in step ③. Built in parts **A→B→C→D→E** (hub → endpoint → mutation broadcasts → enforcement → cascade fan-out). **Depends on ③.5; recheck for drift after it lands.**

> **Why ③ and ④ were arranged this way:** REST features first means the WebSocket layer wraps a complete, correct surface from day one — no retrofit pass. Activities writes belong in step ③ because they are plain DB inserts tied to each mutation; adding them alongside the feature avoids reopening those code paths in step ④. The participation *cascade* (data correction) goes in step ③ too — it has no WebSocket dependency; only its *broadcast* waits for step ④.

### 9.2 Audit → use-case → story map
Which use cases an audit item touches, and the story that asserts the fixed behavior.

| Audit item | Priority | Use case(s) | Story |
| --- | --- | --- | --- |
| Layered permission for board management | P0 | UC-11, UC-12b (and archive/transfer/remove) | US-11, US-12b |
| `cards.created_by` constraint bug | P0 | UC-17 (card delete path) | US-17 |
| Self-join eligibility validation ✅ | P2 | UC-09 | US-09 |
| Missing: board `visibility` ✅ (ADR-005, mig. 000007) | P1 | UC-07, UC-08, UC-12 | US-07, US-08, US-12 |
| ~~Missing: ownership transfer~~ ✅ (ADR-006) | P1 | UC-12e | US-12e |
| ~~Missing: promote/demote, leave, delete workspace~~ ✅ | P1 | UC-06b, UC-06c, UC-06d | US-06b, US-06c, US-06d |
| Missing: WebSocket layer + `activities` (activities design settled — ADR-007, §4.6) | P1 | UC-18, UC-19, §4.6 | US-18, US-19, US-09 |
| Positioning is INTEGER-shift | P1 | UC-13d, UC-15 | US-13d, US-15 |
| ~~Assignee must be a board member (§2.8)~~ ✅ | P1 | UC-14, UC-16 | US-14, US-16 |
| ~~Unassign on lost participation (workspace: UC-06, UC-06c)~~ ✅ | P1 | UC-06, UC-06c | US-06, US-06c |
| ~~Unassign on lost participation (board-scoped: UC-10, UC-12d)~~ ✅ | P1 | UC-10, UC-12d | US-10, US-12d |
| ~~Two sources of truth for board ownership~~ ✅ (ADR-006) | P2 | UC-12e | US-12e |
| Transaction boundaries (create workspace/board) | P2 | UC-03, UC-07 | US-03, US-07 |

---

## 10. Open Notes / Future Considerations
- **Account deletion / deactivation:** *not in Phase 1.* Real-product/GDPR concern for later. Schema is forward-compatible (FK policy in §3.4); adding it later needs no migration rewrite. When added, also re-introduce a "disabled account → 403" path in login. *(This note exists so the decision is not forgotten.)*
- **Refresh tokens:** Phase 1 uses a 24h JWT only; refresh rotation is a Phase 2 candidate.
- **Multi-instance realtime:** Phase 1 uses single-instance in-memory rooms. Horizontal scaling (Redis pub/sub for cross-instance broadcast) is deferred; the room abstraction should be written so it can be swapped to a pub/sub backend later.
- **Non-functional (Phase 1 targets, indicative):** pagination default `limit 50` on list endpoints; basic rate limiting on auth endpoints; WS reconnect backoff as in UC-18.

---

## 11. Phase 2 Use Cases (reference)
- **UC-21 Add Comment** (`comments` table) — broadcast `COMMENT_ADDED`.
- **UC-22 View Activity Log** — paginated read of `activities` (data already written in Phase 1).
- Labels, attachments, avatar upload, public boards, account deletion.
