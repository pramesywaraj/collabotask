# ADR-011: Frontend Framework Split — Vite SPA (app zone) + Next.js (public zone)

- **Status:** Accepted — **public zone framework superseded by [ADR-013](adr-013-public-zone-framework.md)** (Astro replaces Next.js for the public zone, 2026-07-27). App zone decision (Vite SPA) is unchanged.
- **Date:** 2026-07-25
- **Scope:** Framework and rendering strategy for the CollaboTask frontend. Covers the zone split, why the app zone is a plain CSR SPA rather than a Next.js zone, and the security/DX/performance reasoning. Full frontend architecture: `docs/architecture/frontend-architecture.md`. Depends on ADR-008 (httpOnly cookie auth — the decision that blocks authenticated SSR) and ADR-009 (WebSocket design — the decision that makes the app zone's hard problems clear).

## Context

The frontend was noted in the SRS as "Next.js in two zones" — a public SSG zone and an app CSR zone. Before building, this decision was grilled in full. Two facts dominate:

1. **No authenticated SSR in Phase 1.** ADR-008 moves JWT delivery to a host-only `__Host-token` cookie (`SameSite=Lax; Path=/; no Domain`). A host-only cookie cannot be forwarded from the browser to a Next.js server to the Go API — the cross-subdomain SSR pattern is blocked by construction. The app zone will be CSR-only regardless of which framework is chosen.

2. **The app zone's hard problems have nothing to do with rendering strategy.** The three problems that determine whether this app is good or painful — WebSocket state, optimistic drag-and-drop, and concurrent-user reconciliation — are all client-side state problems that must be solved identically whether the framework is Next.js or Vite. Rendering strategy gets you to the page; it does not help once you're there.

Given these two facts, the question becomes: does Next.js add enough value in the app zone (where SSR is blocked and RSC is impractical alongside WebSocket) to justify its overhead?

## Options Considered

### Option A — Next.js two-zone (both zones as Next.js)

One framework for the whole frontend. The app zone uses Next.js purely as a CSR shell — file-based routing, React, and the build system — without SSR or RSC.

**Strengths:** one framework, one deployment pipeline, one `package.json` convention, component sharing is trivial.

**Weaknesses:**

- **RSC friction:** Next.js App Router's design direction is "mostly server, occasionally client." The board view requires `'use client'` on nearly every component (WS connection, drag-and-drop, presence, optimistic mutations). You pay the RSC mental model and compilation overhead for the opposite pattern — mostly client, never server.
- **Slower HMR:** Next.js processes files through a server-side compilation pipeline even in CSR mode. HMR is typically 500ms–2s. Vite uses native ES modules and resolves changed modules in ~50ms. Iterating on drag-and-drop and real-time interactions all day, the difference compounds.
- **Hydration complexity:** Next.js generates server HTML that React hydrates on the client. Hydration mismatches are a debugging class that does not exist in a pure CSR SPA. Since there is no SSR benefit in the app zone, the hydration step is pure overhead.
- **Strict CSP blocked:** Next.js injects inline scripts for its runtime and hydration. A strict `Content-Security-Policy: script-src 'self'` is impossible without per-request nonces (extra complexity). This weakens the last line of XSS defence.
- **Server attack surface:** Even in CSR mode, a Next.js deployment runs a Node.js server. A Vite SPA in production is static files on a CDN — no server process, smaller attack surface.

### Option B — Vite SPA (app zone) + Next.js (public zone) — chosen

The public zone (landing, login, register) is a standard Next.js SSG app. The app zone is a React + Vite SPA: `index.html` → React renders everything client-side.

**Strengths:**

- **One rule:** everything in the app zone is client-side. No `'use client'` directives, no hydration, no server/client boundary to reason about.
- **~50ms HMR:** Vite's native ESM resolution. No server-side compilation for the module that changed.
- **Strict CSP possible:** no inline scripts, so `script-src 'self'` is enforceable. XSS mitigation by construction.
- **No server in the app zone:** production is static files. No Node.js process to compromise.
- **Clean WebSocket integration:** a WS connection is a long-lived client-side object. In a pure CSR SPA this is natural. In Next.js it requires `'use client'` + `useEffect` + careful hydration suppression.

**Weaknesses:**

- Two separate projects (two `package.json`, two deployment configs). For a solo developer this is a real overhead, not a theoretical one. Shared components between zones require either a monorepo shared package or copy-paste.

**Why the weakness is acceptable:** the app zone and public zone share almost no components in Phase 1. The public zone is a static marketing shell (landing + auth forms). The app zone is the entire Kanban product. The overlap is minimal, and the DX + security gains in the zone where the hard work happens outweigh the two-project overhead.

## Decision

**Vite SPA for the app zone, Next.js for the public zone.**

The app zone will never use what makes Next.js valuable (SSR, RSC, API routes) in Phase 1, and the two features Next.js does bring to a CSR app (file-based routing, component conventions) are fully covered by React Router v7. The Vite SPA makes the hard problems — WS integration, optimistic D&D, strict CSP — straightforward. Next.js would make them harder without offering any compensating benefit.

### Supporting decisions (settled alongside this ADR)

- **Router:** React Router v7 (battle-tested, `clientLoader` for TanStack Query prefetch integration, nested routes for the URL-modal card pattern).
- **Auth guard:** `_authenticated.tsx` layout route with `clientLoader` — calls `queryClient.ensureQueryData(['profile'])`, redirects cross-origin to the Next.js public zone on 401. Frontend auth middleware equivalent.
- **Card detail:** URL-based modal — `/workspaces/:wsId/boards/:boardId/cards/:cardId` renders the board underneath + card drawer on top. Deep linking + browser back-button UX at no extra fetch cost (data comes from the `['kanban', boardId]` cache).
- **UI components:** shadcn/ui + Tailwind — Radix UI primitives (accessible), code ownership (no version lock-in), neutral aesthetic.
- **HTTP client:** axios — interceptors for automatic CSRF header (`X-CSRF-Protection: 1` on all mutations per SRS §3.5) and global 401 redirect. Phase 2 avatar upload benefits from `onUploadProgress`.
- **Folder structure:** feature-based (`features/workspace/`, `features/board/`, `features/card/`) — related code co-located, scales with the feature count.

### Out of scope (deferred, recorded so they aren't lost)

- **Authenticated SSR:** possible only with a same-origin reverse-proxy topology (single domain, Next.js server forwards cookie to Go API). Requires changing the deployment topology and moving the cookie off the `__Host-` prefix. Phase 2+ candidate; ADR-008 §7 notes it.
- **Monorepo shared package:** if the public and app zones grow to share significant UI, extract shared components into a workspace package. Phase 1 overlap is too small to justify it.
- **SSR for the app zone:** no path to this in Phase 1 (host-only cookie). Revisit if the auth topology changes.
