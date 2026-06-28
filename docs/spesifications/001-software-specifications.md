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

**Frontend:** **Next.js** in two zones:
- *Public zone* (SSG/SSR, SEO): landing page, marketing, login/register.
- *App zone* (`/app/*`, CSR/SPA, auth-gated): the Kanban app.
- **WebSocket connects directly to the Go backend**, never via Next.js API routes. (Next on serverless cannot host persistent WS; keeping WS in Go avoids that entirely.) Drag-and-drop: `dnd-kit`.

**Database:** PostgreSQL.

### 3.2 Positioning strategy — Fractional NUMERIC + rebalancing
Card and column ordering uses **fractional positions** stored as `NUMERIC`, not integer-shift.
- Insert between A and B: `position = (A + B) / 2`. Insert at ends: `first - STEP` / `last + STEP` (e.g. STEP=1000, seed columns at 1000, 2000, 3000…).
- **One UPDATE per move** — other rows are never touched. Concurrency-friendly (no whole-column lock), broadcast payload is tiny ("card X is now at 1500").
- **Rebalancing:** when the gap between two neighbors falls below a threshold (precision exhaustion), rewrite that column's positions back to evenly-spaced values in one transaction. Rare under normal use.
- *Migration required:* change `columns.position` and `cards.position` from `INTEGER` → `NUMERIC`. (Lexorank was considered and rejected as overkill for Phase 1.)

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
- JWT, **7-day expiry, no refresh token in Phase 1** (conscious decision; refresh-token rotation is a Phase 2 candidate).
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
- **API:** `POST /api/v1/auth/register` · Public
- **Request:** `{ "email", "name", "password" }`
- **Response 201:** `{ "user": { "id", "email", "name", "avatar_url" }, "token": "<jwt>" }`
- **Flow:** validate email format (client + server) → normalize email (lowercase + trim) → check email not taken (case-insensitive; the normalized email is both checked and stored) → bcrypt hash → insert user → issue JWT (7d) → redirect to workspace list.
- **Errors:** Email exists → 409 `CONFLICT`; invalid body → 400 `VALIDATION_ERROR`.
```sql
SELECT id FROM users WHERE email = $1;
INSERT INTO users (email, password_hash, name) VALUES ($1, $2, $3)
RETURNING id, email, name, avatar_url, created_at;
```

#### UC-02: User Login
- **API:** `POST /api/v1/auth/login` · Public
- **Request:** `{ "email", "password" }`
- **Response 200:** `{ "user": {…}, "token": "<jwt>" }`
- **Flow:** normalize email (lowercase + trim) → look up user → verify bcrypt password → issue JWT. Lookup is case-insensitive (same normalization as registration).
- **Errors:** Email not found → 401 `UNAUTHORIZED`; wrong password → 401 `UNAUTHORIZED`. *(No "account disabled" case in Phase 1.)*
```sql
SELECT id, email, password_hash, name, avatar_url FROM users WHERE email = $1;
```

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
- **Request:** `{ "role": "ADMIN" | "MEMBER" }`
- **Rationale:** the layered model relies on "admin as fallback"; we need a way to create more admins. Guard: cannot demote the **last** admin / the workspace owner below ADMIN.
```sql
UPDATE workspace_members SET role=$3 WHERE workspace_id=$1 AND user_id=$2 RETURNING *;
```

#### UC-06c: Leave Workspace — **NEW (Phase 1)**
- **API:** `POST /api/v1/workspace/:workspace_id/leave` · Auth · member
- **Guard:** workspace owner cannot leave (must transfer/delete first); last admin cannot leave.
- **Flow:** same cascade as UC-06 for the leaving user — remove `board_members` across the workspace **and clear their card assignments on every board in this workspace** (§2.8), one transaction; broadcast `MEMBER_REMOVED` + `CARD_UPDATED` per cleared card after commit.

#### UC-06d: Delete Workspace — **NEW (Phase 1)**
- **API:** `DELETE /api/v1/workspace/:workspace_id` · Auth · **owner only**
- **Behavior:** hard delete; cascades to boards/columns/cards/members. UI confirmation required.

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
- **Visibility rule:** admins see all boards; members see WORKSPACE-visible boards + PRIVATE boards they belong to.
```sql
-- conceptually: WHERE b.workspace_id=$ws AND b.is_archived=FALSE AND (
--   wm.role='ADMIN' OR b.visibility='WORKSPACE'
--   OR b.created_by=$user OR bm.user_id IS NOT NULL )
```

#### UC-09: Admin Joins Board (self-service / break-glass)
- **API:** `POST /api/v1/workspace/:workspace_id/board/:board_id/join` · Auth
- **Who:** workspace admin joining any board, or any workspace member joining a WORKSPACE-visible board.
- **Flow:** validate eligibility → insert `board_members(role=BOARD_MEMBER)` (idempotent) → **log activity** → broadcast `USER_JOINED`.
```sql
INSERT INTO board_members (board_id, user_id, role) VALUES ($1,$2,'BOARD_MEMBER')
ON CONFLICT (board_id, user_id) DO NOTHING RETURNING *;
INSERT INTO activities (board_id,user_id,action_type,entity_type,entity_id,metadata)
VALUES ($1,$2,'MEMBER_ADDED','MEMBER',$2, jsonb_build_object('self_joined',true));
```
- **WS:** `USER_JOINED`.

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
- **Access:** layered model (§2.3). For PRIVATE boards a non-member admin must have Joined.
- **Then:** client opens WS, sends `JOIN_BOARD`, receives `ACTIVE_USERS`.
```sql
-- access check (layered):
SELECT 1 FROM boards b
INNER JOIN workspace_members wm ON b.workspace_id=wm.workspace_id AND wm.user_id=$user
LEFT JOIN board_members bm ON b.id=bm.board_id AND bm.user_id=$user
WHERE b.id=$board AND (
  wm.role='ADMIN' OR b.visibility='WORKSPACE'
  OR b.created_by=$user OR bm.user_id IS NOT NULL);
-- then fetch board + columns + cards ORDER BY col.position, c.position
```

#### UC-12b: Update Board Settings
- **API:** `PATCH /api/v1/workspace/:workspace_id/board/:board_id` · Auth · **can_administer_board**
- **Request:** `{ "title"?, "description"?, "background_color"?, "visibility"? }`
- **WS:** `BOARD_UPDATED`.

#### UC-12c: Archive / Unarchive Board (Phase 1 — already in code)
- **API:** `POST /api/v1/workspace/:workspace_id/board/:board_id/archive` · Auth · **can_administer_board**
- **Request:** `{ "is_archived": true|false }`
- **WS:** `BOARD_ARCHIVED` / `BOARD_UNARCHIVED`.

#### UC-12d: Remove Board Member
- **API:** `DELETE /api/v1/workspace/:workspace_id/board/:board_id/member` · Auth · **can_administer_board**
- **Request:** `{ "user_id": "uuid" }`
- **Flow (1 transaction):** delete the `board_members` row → **clear the removed user's card assignments on this board** (§2.8), capturing affected card ids.
- **WS:** after commit, `MEMBER_REMOVED` + `CARD_UPDATED` (`assigned_to: null`) per cleared card.

#### UC-12e: Transfer Board Ownership — **NEW (Phase 1)**
- **API:** `POST /api/v1/workspace/:workspace_id/board/:board_id/transfer-ownership` · Auth · **current BOARD_OWNER or workspace ADMIN**
- **Request:** `{ "to_user_id": "uuid" }` — target must already be a board member.
- **Effect:** demote current owner to `BOARD_MEMBER`, promote target to `BOARD_OWNER` (single owner invariant), in one transaction. Log activity.
- **WS:** `OWNERSHIP_TRANSFERRED`.

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

#### UC-13d: Reorder Column — **(in code, but see §9)**
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
- **API:** `PATCH /…/cards/:card_id` · `{ "title"?, "description"?, "assigned_to"?, "due_date"? }` (assigned_to nullable to unassign; if set, **must be a board member** — §2.8). **WS:** `CARD_UPDATED`.

#### UC-17: Delete Card
- **API:** `DELETE /…/cards/:card_id` — confirm in UI. With fractional positions, no neighbor reorder needed. **WS:** `CARD_DELETED`.

### 4.5 Real-time Collaboration

#### UC-18: WebSocket Lifecycle
- **Endpoint:** `GET /api/v1/ws?token=<jwt>` (or `Authorization` header) — **NEW, not in code.**
- **Connect:** validate JWT → register connection → client sends `JOIN_BOARD{board_id}` → server validates board access → add to in-memory board room → broadcast `USER_JOINED` → send `ACTIVE_USERS` to the joiner.
- **Disconnect:** remove from room + pool → broadcast `USER_LEFT`.
- **Reconnect:** client exponential backoff (immediate, 1s, 2s, 4s, … max 30s) → re-`JOIN_BOARD` → client refetches board state via REST and reconciles.

#### UC-19: Presence
- Server keeps in-memory rooms: `map[board_id] -> map[user_id] -> connection`. On join, the joiner receives `ACTIVE_USERS`.

#### UC-20: Optimistic UI
- Client applies the change immediately, sends to server; on success other clients update, on rejection the originating client reverts and shows an error.

---

## 5. WebSocket Message Contract

All messages share the envelope `{ "type": "<TYPE>", "payload": { … } }`. Timestamps UTC ISO-8601.

### 5.1 Client → Server
| Type | Payload | Notes |
| --- | --- | --- |
| `JOIN_BOARD` | `{ board_id }` | Join a board room (access re-validated). |
| `LEAVE_BOARD` | `{ board_id }` | Optional explicit leave. |

> Card/column mutations may go over REST **or** WS depending on implementation; the **authoritative broadcast** below is what every client must handle. (Recommended: mutations via REST, server broadcasts the resulting event — simpler and reuses existing handlers.)

### 5.2 Server → Clients (broadcast to board room)
| Type | Payload (key fields) | Emitted by |
| --- | --- | --- |
| `USER_JOINED` | `{ board_id, user{id,name,avatar_url}, timestamp }` | UC-09, UC-18 |
| `USER_LEFT` | `{ board_id, user_id, timestamp }` | UC-10, UC-18 |
| `ACTIVE_USERS` | `{ board_id, users:[{id,name,avatar_url,joined_at}] }` | sent to joiner (UC-19) |
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

---

## 6. API Response DTOs (match existing code)

| Entity | JSON fields |
| --- | --- |
| **User** | `id, email, name, avatar_url, system_role` |
| **AuthResult** | `user{…}, token` |
| **Workspace** | `id, name, description, owner_id, role, member_count, board_count, created_at, updated_at` |
| **Board** | `id, workspace_id, title, description, created_by, is_archived, background_color, visibility, user_role, access_status, member_count, created_at, updated_at` *(visibility = NEW)* |
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

**activities** **[NEW — written in Phase 1, UI in Phase 2]**: `id uuid pk`, `board_id → boards cascade not null`, `user_id → users on delete SET NULL`, `action_type varchar not null`, `entity_type varchar not null`, `entity_id uuid not null`, `metadata jsonb default '{}'`, `created_at`.

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

**🟡 P1 — Missing Phase-1 features (need to be built):** board `visibility` column + logic; ownership transfer (UC-12e); promote/demote (UC-06b); leave workspace (UC-06c); delete workspace (UC-06d); **WebSocket layer entirely** (UC-18/19); `activities` table + writes.

**🟡 P1 — Positioning is INTEGER-shift today:**
- Card move: `internal/repository/postgres/card.go:282-341` shifts neighbor positions in a transaction with `FOR UPDATE`.
- Column reorder: `internal/usecase/column/update_column_position.go` reassigns `0,1,2,…` sequentially (thrashes under concurrency).
- Migrate to fractional NUMERIC (§3.2).

**🟡 P1 — Assignee not validated as a board member:** `create_card.go:40`, `update_card.go:83` check the user exists but not that they are a **board member** of the card's board (assignment = participation, §2.8). Add the board-membership check (UC-14/UC-16). *(Supersedes the earlier "validate against workspace" framing — board membership is the stricter, correct rule.)*

**🔵 P2 — Two sources of truth for board ownership:** owner is implied by both `boards.created_by` and a `board_members(BOARD_OWNER)` row. Treat `board_members.role` as authoritative for *current* owner; `created_by` is only the historical creator. Document/enforce this so transfer logic is unambiguous.

**🔵 P2 — Transaction boundaries:** ensure create-board and create-workspace wrap their multi-insert in a single transaction (avoid orphaned rows on partial failure).

**🔵 P2 — REST naming nit:** `DELETE /workspace/:id/member/remove/:user_id` embeds a verb. Prefer `DELETE /workspace/:id/members/:user_id`. (Cosmetic; align when convenient.)

**🔵 P2 — Self-join validation:** confirm `self_join_board.go` verifies the joiner is a workspace member and the board is WORKSPACE-visible (or the joiner is an admin) before inserting.

### 9.1 Suggested build order
Correctness/security first, then features. (Story IDs reference [001-user-stories.md](./001-user-stories.md).)
1. ~~**P0 fixes:** layered permission for board management + `cards.created_by` NULLABLE bug.~~ **✅ DONE** (`canAdministerBoard` helper across update/archive/invite/remove; migration 000005). Tests pending — approach TBD.
2. **Fractional positioning:** migrate card-move and column-reorder off integer-shift to NUMERIC (§3.2).
3. **WebSocket layer:** the entire realtime layer (UC-18/19) + `activities` table & writes.
4. **NEW features:** board `visibility`, ownership transfer (UC-12e), promote/demote (UC-06b), leave workspace (UC-06c), delete workspace (UC-06d).
5. **Assignment = participation (§2.8):** tighten assignee validation to **board member** (UC-14/UC-16), and add the unassign-on-lost-participation cascade + `CARD_UPDATED` broadcast via one shared helper (UC-06, UC-06c, UC-10, UC-12d). Depends on the WebSocket layer (#3) for the broadcast.

### 9.2 Audit → use-case → story map
Which use cases an audit item touches, and the story that asserts the fixed behavior.

| Audit item | Priority | Use case(s) | Story |
| --- | --- | --- | --- |
| Layered permission for board management | P0 | UC-11, UC-12b (and archive/transfer/remove) | US-11, US-12b |
| `cards.created_by` constraint bug | P0 | UC-17 (card delete path) | US-17 |
| Self-join eligibility validation | P2 | UC-09 | US-09 |
| Missing: board `visibility` | P1 | UC-07, UC-08, UC-12 | US-07, US-08, US-12 |
| Missing: ownership transfer | P1 | UC-12e | US-12e |
| Missing: promote/demote, leave, delete workspace | P1 | UC-06b, UC-06c, UC-06d | US-06b, US-06c, US-06d |
| Missing: WebSocket layer + `activities` | P1 | UC-18, UC-19 | US-18, US-19 |
| Positioning is INTEGER-shift | P1 | UC-13d, UC-15 | US-13d, US-15 |
| Assignee must be a board member (§2.8) | P1 | UC-14, UC-16 | US-14, US-16 |
| Unassign on lost participation (member removal/leave) | P1 | UC-06, UC-06c, UC-10, UC-12d | US-06, US-06c, US-10, US-12d |
| Two sources of truth for board ownership | P2 | UC-12e | US-12e |
| Transaction boundaries (create workspace/board) | P2 | UC-03, UC-07 | US-03, US-07 |

---

## 10. Open Notes / Future Considerations
- **Account deletion / deactivation:** *not in Phase 1.* Real-product/GDPR concern for later. Schema is forward-compatible (FK policy in §3.4); adding it later needs no migration rewrite. When added, also re-introduce a "disabled account → 403" path in login. *(This note exists so the decision is not forgotten.)*
- **Refresh tokens:** Phase 1 uses 7-day JWT only; refresh rotation is a Phase 2 candidate.
- **Multi-instance realtime:** Phase 1 uses single-instance in-memory rooms. Horizontal scaling (Redis pub/sub for cross-instance broadcast) is deferred; the room abstraction should be written so it can be swapped to a pub/sub backend later.
- **Non-functional (Phase 1 targets, indicative):** pagination default `limit 50` on list endpoints; basic rate limiting on auth endpoints; WS reconnect backoff as in UC-18.

---

## 11. Phase 2 Use Cases (reference)
- **UC-21 Add Comment** (`comments` table) — broadcast `COMMENT_ADDED`.
- **UC-22 View Activity Log** — paginated read of `activities` (data already written in Phase 1).
- Labels, attachments, avatar upload, public boards, account deletion.
