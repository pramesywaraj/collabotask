# ADR-012: Frontend State Management — TanStack Query + Zustand

- **Status:** Accepted
- **Date:** 2026-07-25
- **Scope:** How client-side state is managed in the React + Vite SPA (app zone). Covers the three-category state split, tool selection, state ownership boundaries, the WS-patches-cache pattern, and the optimistic drag-and-drop flow. Full frontend architecture: `docs/architecture/frontend-architecture.md`. Depends on ADR-011 (Vite SPA decision), ADR-008 (httpOnly cookie — affects auth state model), ADR-009 (WebSocket design — defines the event contract the frontend must handle).

## Context

A real-time Kanban app has three categories of state that are fundamentally different in nature:

| Category | Examples | Characteristics |
|---|---|---|
| **Server state** | Board data, columns, cards, members, profile | Lives on the server; fetched via REST; can go stale; benefits from caching and background refresh |
| **WebSocket push state** | Presence (who's viewing), WS connection status, board rooms | Arrives via push events only; no REST endpoint to `GET`; not cacheable in the request/response sense |
| **UI state** | Which card modal is open, drag in progress, toasts | Ephemeral; local to the browser session; no server representation |

A single state tool cannot serve all three well. The session's authoritative board state (server state) must be reconcilable with out-of-band WS pushes — this is the central architectural challenge. Picking the wrong tool for either category makes the reconciliation either impossible or produces refetch flashes that break the "board is alive" promise.

**Auth state specifically:** with ADR-008's httpOnly cookie, there is no token in JavaScript. "Is the user logged in?" is answered by whether `GET /profile` returns 200 or 401. The user profile is server state — it belongs in TanStack Query, not in a separate auth store. `isAuthenticated` is derived from the query's success status, not stored separately.

## Options Considered

### Redux Toolkit

Full state manager with reducers, slices, RTK Query for server state.

**Rejected.** RTK Query is a capable server-state layer, but the Redux model adds boilerplate (actions, reducers, selectors) for problems TanStack Query solves with a hook. The optimistic update pattern in RTK Query (`onQueryStarted` + `updateQueryData`) is less ergonomic than TanStack Query's `onMutate` / `onError` / `onSettled` lifecycle. No meaningful advantage over TanStack Query + Zustand for this problem shape; significant additional complexity.

### Zustand only

Use Zustand for all three categories — server state stored in Zustand slices, WS events update the same slices.

**Rejected.** Zustand has no built-in concept of "this data can go stale and needs background refresh," "show loading state while fetching," or "deduplicate concurrent requests for the same key." You would reinvent TanStack Query's cache, badly, while building features. The server state category is large and nuanced enough to justify a dedicated tool.

### TanStack Query only

Use TanStack Query for server state; attempt to model WS push state and UI state as queries too.

**Rejected.** TanStack Query is designed for request/response patterns — it needs a `queryFn` that fetches data from somewhere. Presence data (`ACTIVE_USERS`, `USER_JOINED`, `USER_LEFT`) arrives via WS push events with no REST equivalent. Modeling it as a query either requires a fake queryFn or a manual `setQueryData` that bypasses the cache's intended use. WS connection state (connected/disconnected/reconnecting) is not fetchable data at all. Forcing these into TanStack Query produces awkward, unidiomatic code.

### TanStack Query + Zustand — chosen

Use each tool for what it is designed for:

- **TanStack Query** → server state (REST cache). Fetches, caches, deduplicates, background-refreshes.
- **Zustand** → WS push state (no REST equivalent) + UI state (no server representation).

WS events that touch data with a REST equivalent (cards moved, members added, board updated) patch the TanStack Query cache directly via `queryClient.setQueryData`. This means the board state has one container — TanStack Query's cache — that is populated initially by REST and kept live by WS events. No reconciliation required; both paths write to the same store.

## Decision

**TanStack Query for server state. Zustand for WebSocket push state and UI state.**

### State ownership boundary

```
TanStack Query cache              Zustand
────────────────────────          ─────────────────────────────
['profile']                       wsStore.socket
['workspaces']                    wsStore.status
['workspace', wsId]               wsStore.joinedRooms: Set<boardId>
['boards', wsId]                  wsStore.presence: Map<boardId, User[]>
['board', boardId]                uiStore.openCardId
['kanban', boardId]               uiStore.dragging
['board-members', boardId]        uiStore.toasts
['invitees', boardId]
```

**Decision rule:**
- Has a REST `GET` endpoint → TanStack Query.
- Arrives via WS push only, no REST equivalent → Zustand.
- Local to the session, no server representation → Zustand.

### WS-patches-cache pattern

WS events that mutate board state update the TanStack Query cache directly — they never trigger a refetch:

```ts
// In the Zustand WS store event handler
case 'CARD_MOVED': {
  queryClient.setQueryData(['kanban', boardId], old =>
    applyCardMove(old, event.payload)
  )
  break
}
```

`applyEvent` functions (e.g. `applyCardMove`, `applyCardUpdate`) are **pure functions** — same input, same output. They are reused by both the optimistic update path and the WS event handler path, so the board state is never updated by two different code paths.

This avoids the refetch flash (no loading spinner on every board change) and is what makes the board feel live rather than polling.

### Optimistic drag-and-drop flow

```ts
const moveCard = useMutation({
  mutationFn: ({ cardId, toColumnId, toPosition }) =>
    api.post(`/columns/${toColumnId}/cards/${cardId}/move`, { to_column_id: toColumnId, to_position: toPosition }),

  onMutate: async ({ cardId, fromColumnId, toColumnId, toPosition }) => {
    await queryClient.cancelQueries({ queryKey: ['kanban', boardId] })
    const snapshot = queryClient.getQueryData(['kanban', boardId])
    queryClient.setQueryData(['kanban', boardId], old =>
      applyCardMove(old, { cardId, fromColumnId, toColumnId, toPosition })
    )
    return { snapshot }
  },

  onError: (_err, _vars, context) => {
    queryClient.setQueryData(['kanban', boardId], context.snapshot)
    toast.error('Failed to move card — reverted')
  },
  // onSettled: no invalidation — WS CARD_MOVED confirms the authoritative position
})
```

**Sequence:**
1. `onMutate` fires synchronously — snapshot + optimistic update — card moves in UI instantly.
2. REST request goes out (CSRF header added automatically by axios interceptor).
3a. Server accepts → WS `CARD_MOVED` arrives → `setQueryData` with authoritative position (may differ from optimistic if rebalancing occurred).
3b. Server rejects → `onError` restores snapshot → card snaps back → toast shown.

### Concurrent-user reconciliation

When two users act on the same entity simultaneously, WS events are the ground truth for **position and column**; REST responses are the ground truth for **field-level data** (title, description, assignee). Merge, don't replace: a REST response for a PATCH on the card title should not overwrite the column change a WS event has already applied to the cache.

### AUTH state — no separate auth store

With httpOnly cookie auth (ADR-008), there is no token to store in JS. Auth state is:

```ts
const { data: currentUser, isSuccess: isAuthenticated } = useQuery({
  queryKey: ['profile'],
  queryFn: fetchProfile,
  retry: false,   // don't retry on 401 — it means "not logged in"
  staleTime: 5 * 60 * 1000,
})
```

`isAuthenticated = isSuccess`. No Zustand auth slice. No Redux auth reducer. The `_authenticated.tsx` layout route calls `queryClient.ensureQueryData(['profile'])` — this is both the auth gate and the profile prefetch, in one network call that is cache-served on subsequent navigations.

### Out of scope (deferred, recorded so they aren't lost)

- **Offline support / persistence:** TanStack Query's `@tanstack/query-persist-client` can persist the cache to localStorage across page refreshes. Not needed in Phase 1; add when "fast cold start" becomes a stated goal.
- **Optimistic conflict resolution:** the current model is last-write-wins via WS event. If two users move the same card simultaneously, the last WS event wins. A finer-grained conflict model (e.g. CRDT positions) is not needed at Phase 1 scale.
- **Refresh token rotation:** Phase 1 uses 7-day JWT. When Phase 2 adds refresh tokens, the axios response interceptor is the correct place to intercept 401s, call the refresh endpoint, and retry the original request — the interceptor is already wired for the global 401 redirect.
