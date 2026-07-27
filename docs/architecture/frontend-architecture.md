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
    api.ts               ← axios instance (base URL, credentials, CSRF interceptor, 401 redirect)
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
    workspaces.tsx               → /workspaces
    workspaces.$wsId.tsx         → /workspaces/:wsId
    workspaces.$wsId.boards.$boardId.tsx          → /workspaces/:wsId/boards/:boardId
    workspaces.$wsId.boards.$boardId.cards.$cardId.tsx  → card detail modal over board
```

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
['invitees', boardId]
```

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
client.interceptors.response.use(
  res => res,
  err => {
    const isProfileCall = err.config?.url?.includes('/user/profile')
    if (err.response?.status === 401 && !isProfileCall) {
      window.location.href = '/login'
    }
    return Promise.reject(err)
  }
)
```

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
