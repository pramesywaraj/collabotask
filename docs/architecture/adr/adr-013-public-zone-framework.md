# ADR-013: Public Zone Framework — Astro

- **Status:** Accepted — supersedes the public-zone framework mention in ADR-011 (which noted Next.js as the public zone choice before this grilling pass).
- **Date:** 2026-07-27
- **Scope:** Framework for the public zone only (`collabotask.com` — **landing page only**). Login and register have been moved to the app zone (Vite SPA at `app.collabotask.com`) to keep the CORS allowlist to a single origin and leverage the app zone's existing axios CSRF interceptor. The app zone framework (Vite SPA) is unchanged — ADR-011 still stands for that decision.

## Context

ADR-011 established the zone split: Vite SPA for the app zone, and noted Next.js for the public zone. Before building, the public zone framework was grilled separately.

The public zone's Phase 1 requirements are narrow:
- Landing page (static, SEO-indexed, minimal JS)

Login and register were initially scoped here but moved to the app zone after a cross-doc review surfaced two problems: (a) they would require a second CORS origin (`collabotask.com`) in the API allowlist alongside `app.collabotask.com`; (b) they have zero SEO value and do not benefit from SSG — so there was no reason to pay the CORS cost. The Astro public zone makes no API calls.

Three pages. Mostly static. A small amount of client-side JS for forms and the logged-in redirect check. No authenticated SSR (host-only cookie cannot be forwarded from `collabotask.com` server → Go API — same constraint as the app zone, ADR-008). No API routes needed.

## Options Considered

### Next.js (SSG / static export)

The originally noted choice. React throughout, known by the team, can add SSR or API routes if topology changes later.

**Why not chosen for the public zone:**
- Ships ~80KB Next.js runtime for pages that don't need it.
- Brings a Node.js server for 3 static pages.
- Next.js's headline features (SSR, RSC, API routes) are unused and inapplicable here.
- "Using a truck to carry a backpack" — works, but the fit is poor.
- React in the public zone is still available via Astro's React integration; the consistency argument does not require Next.js specifically.

### Gatsby

Rejected outright. Declining ecosystem since 2022, Netlify acquisition reduced team, GraphQL complexity not needed, slower builds. Nothing Gatsby does that Astro or Next.js doesn't do better in 2025.

### Astro — chosen

Built specifically for mostly-static sites with selective interactivity ("islands architecture"). Ships zero JavaScript by default; each interactive component opts in via a `client:*` directive.

```astro
---
// login.astro — static shell, zero framework runtime shipped for the page
---
<html>
  <body>
    <LoginForm client:load />   ← React island; only this component is hydrated
  </body>
</html>
```

**Why chosen:**

- **Right-sized for the use case.** Landing page = pure HTML/CSS, no JS framework overhead. Login/register = React islands for the form + auth check. Astro is built for exactly this pattern.
- **Smallest bundle on the landing page.** No framework runtime. Pure HTML/CSS ships to the browser. Excellent Lighthouse scores without effort.
- **React components work natively** via `@astrojs/react`. The login/register forms are standard React components with `useEffect` + `fetch` for the auth check — identical code to what would run in Next.js.
- **Growing ecosystem, clear future.** Astro has significant adoption momentum in 2025 for static/content sites. Better long-term bet than Gatsby; right-sized vs Next.js for this scope.
- **TypeScript-native.** No additional config needed.

**The auth redirect check in an Astro page:**

```tsx
// components/LoginForm.tsx — React island
export function LoginForm() {
  const [checking, setChecking] = useState(true)

  useEffect(() => {
    fetch('https://api.collabotask.com/api/v1/user/profile', {
      credentials: 'include',
    }).then(res => {
      if (res.ok) window.location.href = 'https://app.collabotask.com'
      else setChecking(false)
    }).catch(() => setChecking(false))
  }, [])

  if (checking) return <Spinner />
  return <form>…</form>
}
```

The SSR limitation is unchanged from ADR-008: the `__Host-token` cookie is bound to `api.collabotask.com`, so Astro's server (at `collabotask.com`) cannot see it either. The redirect check is client-side only — identical behavior to what Next.js would do.

**Weakness accepted:** Astro's `.astro` file syntax is new. It is minimal (frontmatter `---` block + HTML template) and learned in an hour. The React components themselves are standard React — the learning surface is just the `.astro` wrapper.

## Decision

**Astro for the public zone.** Next.js is not used in the frontend at all.

Updated zone split:

| Zone | Technology | Responsibility |
|---|---|---|
| Public | **Astro** (SSG) | Landing only — static, SEO-indexed, no API calls |
| App | React + Vite SPA | Login, register + the Kanban app — auth-gated, real-time (ADR-011) |

### Out of scope (deferred, recorded)

- **Marketing expansion (Phase 2+):** blog, pricing, changelog. Astro is excellent for content-heavy sites (MDX support, content collections). This choice pays forward.
- **Shared component library:** if the login/register form components grow complex enough to share with the app zone, extract to a monorepo package. Phase 1 overlap is minimal — the forms are simple.
