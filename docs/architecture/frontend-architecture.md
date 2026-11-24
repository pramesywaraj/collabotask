# CollaboTask Frontend Architecture

> **Scope:** the React + Vite SPA (`collabotask-frontend/` — to be created). The public zone (landing only) is a separate Astro (SSG) site. This document covers the app zone only.

## Guiding principle

> **The hard problems in a real-time Kanban app are WebSocket state, optimistic drag-and-drop, and concurrent-user reconciliation. None of these are solved by rendering strategy.** Every architectural decision below exists to solve one of those three problems cleanly.

The server is always the source of truth. Clients are optimistic and reconcile via the authoritative WS broadcast.

---

## Zone split

| Zone | Technology | Responsibility |
|---|---|---|
| Public | **Astro** (SSG) | Landing only — SEO-indexed, no auth, no API calls |
| App | React + Vite SPA | Login, register + the Kanban app — auth-gated, real-time |

No authenticated SSR in Phase 1 — the httpOnly `__Host-token` cookie is host-only and cannot be forwarded from any server (`collabotask.com` or otherwise) to the Go API at `api.collabotask.com` (ADR-008). Both zones are effectively CSR-only for auth. The app zone is CSR by design; the public zone is static by design.

---

## Stack

| Concern | Decision | Why |
|---|---|---|
| Build tool | Vite | Native ESM HMR (~50ms), no server-side compilation in dev, strict CSP possible (no inline scripts) |
| Framework | React 18 | — |
| Router | React Router v7 | Battle-tested, nested routes, `clientLoader` for prefetch |
| Server state | TanStack Query | REST cache; WS events patch it directly; no refetch flash |
| WS + UI state | Zustand | Push-only presence (no REST equivalent); WS connection lifecycle; ephemeral UI |
| HTTP client | axios | Interceptors for CSRF header + global 401 handling; upload progress for Phase 2 |
| UI components | shadcn/ui + Tailwind | Radix primitives (accessible), you own the code, no lock-in |
| Drag-and-drop | dnd-kit | — |

---

## Folder structure

```
src/
  features/
    workspace/
      components/        ← WorkspaceCard, MemberList, InviteForm
      hooks/             ← useWorkspaces, useWorkspaceDetail
      api.ts             ← fetchWorkspaces, inviteMember, ...
    board/
      components/        ← BoardCard, BoardSettings, KanbanView
      hooks/             ← useBoards, useKanban
      api.ts
    card/
      components/        ← CardDetail, CardForm (modal content)
      hooks/             ← useCard, useMoveCard
      api.ts
    auth/
      components/        ← redirect logic only (login/register live in routes/)
      hooks/             ← useCurrentUser
  shared/
    components/          ← Button, Modal, Avatar, Toast, Spinner (shadcn/ui wrappers)
    hooks/               ← useDebounce, useClickOutside
  routes/                ← React Router v7 file-based route modules
  lib/
    api.ts               ← axios instance (base URL, credentials, CSRF interceptor, 401 → session-ended flag)
    queryClient.ts       ← TanStack Query client (stale times, retry config)
    wsStore.ts           ← Zustand: WS connection + room management + presence
    uiStore.ts           ← Zustand: modals, drag state, toasts
```

---

## Route tree

```
routes/
  login.tsx                      → /login
  register.tsx                   → /register
  _authenticated.tsx             ← layout: auth check + WS connect
    _index.tsx                   → /  — redirects to the remembered workspace, else /workspaces
    workspaces._index.tsx        → /workspaces  (2.3 — no sidebar)
    workspaces.$wsId.tsx         ← layout: the sidebar + <Outlet/>; remembers this workspace
      workspaces.$wsId._index.tsx                   → /workspaces/:wsId  (2.4)
      workspaces.$wsId.boards.$boardId.tsx          → /workspaces/:wsId/boards/:boardId
      workspaces.$wsId.boards.$boardId.cards.$cardId.tsx  → card detail modal over board
```

**The sidebar's layout route** (`workspaces.$wsId.tsx`, [ADR-034](adr/adr-034-side-menu.md)) wraps 2.4 and the board, so the sidebar stays mounted as you move between them. When it opens, it writes the remembered workspace to `localStorage`, keyed by user id. 2.3 is `workspaces._index.tsx`, so it sits outside that layout and has no sidebar.

**The entry route** (`_index.tsx`, `/`) reads the remembered workspace and redirects there, or to `/workspaces` if there is none. Sign-in lands on `/`, unless it carries a `next` path from 2.13 (*axios instance*, below). `/workspaces` itself never redirects, so *All workspaces* always reaches 2.3. If the remembered workspace's detail is refused (403 or 404), the value is cleared and the app goes to `/workspaces`. After `removed_from_workspace` it is cleared too, if it names that workspace.

The `_authenticated.tsx` layout route:
1. Runs `clientLoader` on every navigation into the authenticated subtree.
2. Calls `queryClient.ensureQueryData(['profile'])` — cache hit is instant; network only on miss/stale.
3. On 401 → `throw redirect('/login')` (same-origin; login lives in the app zone).
4. On success → connects the WS via `wsStore.connect()`.

Card detail (`$cardId`) is a **URL-based modal**: the board renders underneath, the card opens as a drawer on top. Browser back closes it. Card data comes from the `['kanban', boardId]` cache — no extra fetch.

---

## State ownership

```
TanStack Query cache              Zustand
────────────────────────          ─────────────────────────────
['profile']                       wsStore.socket
['workspaces']                    wsStore.status (idle|connecting|connected|…)
['workspace', wsId]               wsStore.joinedRooms: Set<boardId>
['boards', wsId]                  wsStore.presence: Map<boardId, User[]>
['board', boardId]                uiStore.openCardId
['kanban', boardId]  ←── patched  uiStore.dragging { cardId, sourceColumnId }
['board-members', boardId]        uiStore.toasts
['invitees', boardId]             uiStore.sidebarOpen  ← persisted (localStorage)
                                  uiStore.sessionEnded ← set by the 401 handler (2.13)
```

**`['boards', wsId]` has two readers.** 2.4's Active grid shows all of it. The sidebar shows only the boards whose `access_status` is `JOINED`, filtered in a selector, so the request is made once ([ADR-034](adr/adr-034-side-menu.md) §8). It is refetched after you join, leave, create, archive or unarchive a board, and patched when the open board's `BOARD_UPDATED`, `BOARD_ARCHIVED` or `ACCESS_REVOKED` arrives.

**Two per-device values live in browser storage, not in either column:** the last workspace (`localStorage`, keyed by user id) and `uiStore.sidebarOpen`, which only applies from 1024 up. Both are conveniences ([ADR-034](adr/adr-034-side-menu.md) §3, §6). Read them inside `try`/`catch`, and fall back to the defaults when they're missing: the sidebar open, and 2.3.

**Rule:** if it has a REST endpoint you can `GET` → TanStack Query. If it arrives only via WS push with no REST equivalent → Zustand.

WS events that touch data with a REST equivalent (board members, kanban) patch the TanStack Query cache directly — they do not move the data to Zustand.

---

## axios instance (`lib/api.ts`)

```ts
const client = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
  withCredentials: true,                   // sends httpOnly cookie on every request
})

// CSRF header on all mutations — required by SRS §3.5; missing it → 403
client.interceptors.request.use(config => {
  if (!['get', 'head', 'options'].includes(config.method ?? '')) {
    config.headers['X-CSRF-Protection'] = '1'
  }
  return config
})

// Global 401 handler — mid-session expiry only.
// Profile calls are excluded: _authenticated.tsx clientLoader owns first-load 401s
// so they go through React Router's redirect, not a hard navigation.
// It does not navigate: it raises a flag, and the authenticated layout shows 2.13 (ADR-038 §13).
client.interceptors.response.use(
  res => res,
  err => {
    const isProfileCall = err.config?.url?.includes('/user/profile')
    if (err.response?.status === 401 && !isProfileCall) {
      useUiStore.getState().setSessionEnded(true)
    }
    return Promise.reject(err)
  }
)
```

**Session expired, 2.13 ([ADR-038](adr/adr-038-board-states-group-d.md) §13).** Replaced 2026-11-23: the handler used to assign `window.location.href = '/login'`, which looked like a crash and lost the page you were on.
- While `uiStore.sessionEnded` is set, `_authenticated.tsx` renders the 2.13 page in place of its outlet and closes the WebSocket. Nothing moves until the person presses *Sign in*.
- *Sign in* goes to `/login?next=<path>`, where `<path>` is the current `location.pathname + location.search`, URL-encoded.
- After a successful sign-in, Login navigates to `next` **only if it is a same-site path**: a string that starts with `/` and not with `//` or `/\`. Anything else is ignored, and sign-in lands on `/` as before. This check stops an open redirect.
- First-load 401s are unchanged: the `clientLoader` still throws `redirect('/login')`.

---

## Data flow for every board mutation

```
1. User acts (drag, click, type)
2. useMutation.onMutate  → snapshot TQ cache → apply optimistic update → UI instant
3. axios                 → REST mutation (CSRF header automatic via interceptor)
4. useMutation.onError   → restore snapshot → UI reverts → toast error
5. WS event arrives      → queryClient.setQueryData → all clients converge on server truth
```

The WS event reaches the originating client too (SRS §5.2 broadcast rule includes sender). If the server adjusted the position (e.g. rebalanced fractional positions), this corrects the optimistic state silently.

**Confirmed actions that take you somewhere else are not optimistic** ([ADR-040](adr/adr-040-menus-and-the-confirmation.md) §13). Deleting a card or a column follows the flow above: the confirmation closes at once and the object is in flight. *Delete workspace*, *Archive board* and *Leave board* wait instead. The confirm button goes Busy until the response arrives, and then:
- **after *Delete workspace*:** navigate to `/workspaces`, with no toast. If the remembered workspace names this one, clear it first
- **after *Archive board*:** navigate to `/workspaces/:wsId`, with a toast that offers *Undo*. *Undo* is the same request with `is_archived: false` (US-12c; ADR-040 §5)
- **after *Leave board*:** navigate to `/workspaces/:wsId`, with no toast, since leaving by choice shows nothing (ADR-030)

On a failure the dialog stays open with its button back at rest, and the error is reported as any other mutation's.

**A card's due date is a day, not a moment** ([ADR-041](adr/adr-041-the-card-drawers-pickers.md) §12; SRS §3.6). Send the picked day as `` `${day}T00:00:00Z` `` and read it back as `due_date.slice(0, 10)`. Never pass it through `new Date()` for display, or a teammate in another time zone sees a different day. The date wording function (ADR-037 §5, ADR-038 §6) compares that day with the viewer's local today, written as `YYYY-MM-DD`. Clearing sends `due_date: null`.

**Move card to… is one component in three places** (ADR-041 §1–§3; ADR-019 §6). In the card drawer it expands in place under *Column*. From the card ⋯, picking *Move card to…* closes the `DropdownMenu` and opens a `Popover` on the same trigger from 640 up, or replaces the bottom `Sheet`'s content below 640. `M` on a focused card opens the `Popover` at the card. It asks for a column, then a place, and picking the place sends the move, with the target position computed exactly as a drop computes it (UC-15). After a move from the board, focus follows the card to its new place and a polite live region says where it went (*Moved to Done, at the bottom.*). From the drawer, the drawer stays open.

**Forms** ([ADR-042](adr/adr-042-the-forms.md)). One `FormFrame` component renders a shadcn `Dialog` from 640 up (448 wide, no close button, `onInteractOutside` prevented, so pressing outside does nothing) and a full-height `Sheet` below, with *Close* in its head and one full-width button. Escape, *Cancel* and *Close* close it without asking. Every form follows the same rules:
- **Trim every value before checking it and before sending it.** The server doesn't trim names or titles, so two spaces would pass a 2-character minimum.
- **A minimum length is checked as you type; a missing value on submit.** The Primary button's *not satisfied* look comes from the same check. Fields stop at 255 characters (names, titles) and 1,000 (descriptions), so a too-long error never shows.
- **A server failure is the `Form message`**, the last line above the buttons, and every value is kept. 403 after a role change has its own wording.
- **Creating puts you inside what you made**: after *Create workspace*, `/workspaces/:newId`; after *Create board*, the new board.
- **Board settings sends only what changed** (UC-12b's PATCH). Nothing changed closes the form with no request.

**A board's colour** ([ADR-042](adr/adr-042-the-forms.md) §9; SRS §4.3). `background_color` stores the hue's deep value (`#744539` … `#74395F`). `hueFromBoard(board)` reads it by case-insensitive exact match and otherwise falls back to the hue derived from the board id. Create a board starts on `firstFreeHue(boards)`, the first hue in ADR-017 §3.3's order that no active board in the workspace uses (terracotta when all are used), computed from the board list the page already holds.

**Inviting people** ([ADR-042](adr/adr-042-the-forms.md) §12–§13; UC-04). One malformed address makes the server fail the whole batch with a 400, so the address field splits input on commas, semicolons, spaces and new lines, lower-cases and trims each address, drops repeats, and checks each one as it becomes a chip. *Invite* is not satisfied while any chip is invalid or none exist. The 200 answer has a row per address. Added and already-a-member addresses leave the field, while addresses with no account stay as chips. A polite live region announces *2 of 4 added to Product Team.*

**Adding a column** ([ADR-042](adr/adr-042-the-forms.md) §14; UC-13). *Add column* appends a client-only draft column, scrolls the board to it, and focuses its header field. Enter, or blur with a name, calls UC-13, and the server's column replaces the draft. Escape, or blur when empty, drops it. No dialog is involved.

---

## WS event routing

Every event arriving on the socket is routed in the Zustand WS store's event handler:

```
CARD_CREATED / CARD_UPDATED / CARD_MOVED / CARD_DELETED
COLUMN_CREATED / COLUMN_UPDATED / COLUMN_MOVED / COLUMN_DELETED
BOARD_UPDATED / BOARD_ARCHIVED / BOARD_UNARCHIVED
MEMBER_ADDED / MEMBER_REMOVED / OWNERSHIP_TRANSFERRED
  → queryClient.setQueryData(['kanban' | 'board' | 'board-members', boardId], applyEvent)

USER_JOINED / USER_LEFT / ACTIVE_USERS
  → wsStore.presence update (Zustand only — no REST equivalent)

ACCESS_REVOKED
  → queryClient.removeQueries(['kanban', boardId])
  → navigate to /workspaces/:wsId  (leave the board view)
     — or to /workspaces when reason is removed_from_workspace (ADR-030 §6)
  → toast naming the reason (wordings: ADR-030 §5)
```

Event applier functions (`applyCardMove`, `applyCardUpdate`, etc.) are **pure functions** — same input, same output. Both the optimistic update and the WS handler reuse them.

---

## WebSocket lifecycle

```
_authenticated mounts    → wsStore.connect()  → wss://api.collabotask.com/api/v1/ws
                                                  (httpOnly cookie sent automatically)
Board view mounts        → wsStore.joinBoard(boardId)   → server: JOIN_BOARD
Board view unmounts      → wsStore.leaveBoard(boardId)  → server: LEAVE_BOARD
Connection drops         → wsStore reconnect: 0ms → 1s → 2s → 4s … max 30s
                           on reconnect: re-send JOIN_BOARD for all joinedRooms
                           board view: refetch ['kanban', boardId] to reconcile missed events
User logs out            → POST /auth/logout (CSRF-protected) → cookie cleared server-side
                           → queryClient.clear()             → cached user data wiped
                           → wsStore.disconnect()            → socket closed
                           → navigate('/login')
```

One connection per session, alive from authenticated-zone entry to logout. Navigation between boards uses `JOIN_BOARD` / `LEAVE_BOARD` — no reconnect cost.

---

## The three hard problems

| Problem | Solution |
|---|---|
| WS events update UI without refetch flash | `queryClient.setQueryData` with pure `applyEvent` functions — no invalidation, no loading state |
| Optimistic D&D with rollback | TanStack Query `onMutate` snapshots the cache; `onError` restores it; card snaps back on rejection |
| Concurrent user reconciliation | WS broadcast is ground truth for position/column; REST response is ground truth for field data. Merge, don't replace. |

---

## Adding a new feature

Work **outside-in**:

1. `features/<name>/api.ts` — add the axios call(s)
2. `features/<name>/hooks/` — add `useQuery` / `useMutation` with `queryKey`, optimistic `onMutate`, and rollback `onError`
3. `features/<name>/components/` — build the UI, consuming the hook
4. `routes/` — add or update the route module
5. If the feature involves WS events: add the event case to `wsStore` event handler + the `applyEvent` pure function

No new state primitives needed unless the data has no REST equivalent (presence-style push-only data).

---

## Related decisions

- [ADR-011](./adr/adr-011-frontend-framework.md) — why Vite SPA for the app zone (not Next.js two-zone)
- [ADR-012](./adr/adr-012-frontend-state-management.md) — why TanStack Query + Zustand (state ownership + WS-patches-cache pattern)
- [ADR-013](./adr/adr-013-public-zone-framework.md) — why Astro for the public zone (not Next.js)
- [ADR-008](./adr/adr-008-auth-httponly-cookie.md) — httpOnly cookie auth (why no authenticated SSR; CSRF header requirement)
- [ADR-009](./adr/adr-009-websocket-realtime-layer.md) — WebSocket backend design (Broadcaster port, hub, presence model)
- [ADR-034](./adr/adr-034-side-menu.md) — the sidebar on 2.4 and the board (the shared layout route, the `/` redirect to the last workspace, the persisted open state)
- [ADR-042](./adr/adr-042-the-forms.md) — the forms (one `FormFrame`, trimming and validation, the stored board colour, the invite batch, the draft column)
