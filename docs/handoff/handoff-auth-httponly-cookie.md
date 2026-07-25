# Handoff — Auth: `Bearer` → httpOnly cookie (build step ③.5)

**Date:** 2026-07-23
**Repo:** `collabotask` (backend: `collabotask-backend/`)
**Status:** **Planning handoff** from a `/grill-with-docs` session. Design settled and **approved by the product owner**. **Nothing built yet.**

> This is the durable build guide for step ③.5. The *why/decision* record is **[ADR-008](../architecture/adr/adr-008-auth-httponly-cookie.md)**; the API/contract changes are in **SRS §3.5 / §4.1**. This file is the *how-to-build*: sequencing, files to touch, TDD checklist, tooling, and gotchas.

## For the implementing agent — start here
**Read in this order:** (1) **this handoff** — the *Build plan* table below is your task list, work it top-to-bottom, TDD; (2) **[ADR-008](../architecture/adr/adr-008-auth-httponly-cookie.md)** — the *why* behind any decision you're unsure about (8 decisions + rejected alternatives); (3) **SRS §3.5 / §4.1** — the exact API contract (cookie, CSRF header, CORS, UC-01/02 responses, UC-03 logout); (4) **`collabotask-backend/README.md`** (conventions + the "add an endpoint" recipe) and **`TESTING.md`** (test approach) — match the existing handler/middleware style, error envelope, and test layout.

**Cold-start clarifications:**
- **No Wire/DI regeneration expected** — this adds no new injected dependency: the cookie helper is a plain function, the CSRF middleware takes `*config.Config`, and `Logout` is a new method on the existing `AuthHandler`. Middlewares are wired by hand in `router.go`, not via Wire. (Only regenerate if you introduce a provider.)
- **No mock regeneration** — this is **HTTP-layer only**. The auth **usecase is unchanged** — it still generates and returns the JWT in its output; only the *handler* changes (JWT → `Set-Cookie` instead of response body). Usecase interfaces/mocks and the existing 438 usecase tests are untouched.
- **Middleware order in `router.go`:** `CORS` → `CSRF` (global `routes.Use`, so it also covers `/auth/*`) → `Auth` (per protected group only). CORS already aborts `OPTIONS` with `204`, so preflight never reaches CSRF; CSRF must still skip safe methods (`GET/HEAD/OPTIONS`).
- **CSRF applies to `/auth/*` too:** login/register/logout require `X-CSRF-Protection` — in flow tests, send the header on the **login** request (before any cookie exists).

**Definition of done:** all 6 build-plan steps green; `go build ./...` and `go vet ./...` clean; `go test ./...` passes **including the existing 438 usecase tests**; new middleware/handler/flow tests added per §Testing; `swag init` regenerated; **nothing committed — leave changes staged for the product owner's review** (no auto-commit).

## Sequencing (read first)
```
③        REST features ✅ (done)
③.5      Auth: Bearer → httpOnly cookie   ← THIS — built before ④
④        WebSocket + participation broadcast (ADR-009) — build AFTER ③.5, then recheck for drift
```
- **④ depends on ③.5.** The WS handshake authenticates by reading the **JWT cookie** → `ValidateToken` → reject before upgrade (no `?token=`, no subprotocol). The **explicit CORS origin allowlist** this step establishes is shared by REST + WS (ADR-009 §origin feeds it into `OriginPatterns`).
- **After ③.5 lands, recheck the WS handoff** (`docs/handoff/websocket-participation-broadcast/index.md`) for drift — it assumes the auth middleware/handlers this step rewrites.

## Settled design (condensed — full rationale in ADR-008)
1. **Cookie carries the JWT, stateless.** `GenerateToken`/`ValidateToken`/claims/expiry **unchanged** — only transport + delivery change. (Session/refresh model → Phase 2.)
2. **Topology A** — cross-origin sibling subdomains (`app.` ↔ `api.`). C2 single-origin → future enhancement.
3. **Host-only cookie** (no `Domain=`) — no authenticated SSR in Phase 1.
4. **Cookie:** `__Host-token=<jwt>; HttpOnly; Secure; SameSite=Lax; Path=/; Max-Age=<cfg.JWTExpiration>`.
5. **CSRF:** required `X-CSRF-Protection` header on **all** state-changing methods (incl. `login`/`register`/`logout`), enforced by new middleware, backed by strict CORS + `SameSite=Lax`.
6. **Hard cutover** — cookie-only; login/register **drop `token` from the body**; no `Bearer` path.
7. **Logout** — unconditional, CSRF-protected `POST /auth/logout`; clears the cookie; not behind `Auth`.
8. **CORS corrected** — explicit allowlist, reflect-if-allowlisted, credentials on, add `PATCH`, enumerate headers, `Vary: Origin`, fail-fast on `*`+credentials.

## Build plan (ordered, TDD — red → green per step)

Each step is a natural checkpoint. Middleware/handler tests need **no DB harness** (gin test context + `httptest`); the existing 438 usecase tests are untouched by this work.

| Step | Builds | Done when |
|---|---|---|
| **1 — Config + CORS** | `CORSConfig` flip to explicit allowlist; add `PATCH`; enumerate headers incl. `X-CSRF-Protection`; `Vary: Origin`; fail-fast validation; new `AuthConfig` fields `CookieSecure`, `CookieName`; `.env.example` | Config tests: `*`+credentials → startup error; CORS reflects only allowlisted origin, omits `ACAO` otherwise; preflight advertises `PATCH` + `X-CSRF-Protection` |
| **2 — Cookie write helpers + login/register** | A small `authcookie` helper (`Set(c, jwt)`, `Clear(c)`) that honors `CookieSecure`/`CookieName`/`__Host-` prefix + `Max-Age = cfg.JWTExpiration`; login/register call `Set` and **stop returning `token`**; drop `Token` from `AuthResponse` | Login/register set the right `Set-Cookie` (attribute-assertion test); body has no `token` |
| **3 — Auth middleware reads cookie** | Rewrite `middleware.Auth` to read `__Host-token`/`CookieName` cookie → same `ValidateToken` → same context set | Middleware unit tests: valid cookie → `userID` in context; absent/invalid → 401. Protected routes still work end-to-end via cookie |
| **4 — CSRF middleware** | New `middleware.CSRF`: on `POST/PUT/PATCH/DELETE`, require `X-CSRF-Protection` header, else 403; safe methods pass through; wire globally (before `Auth`, after `CORS`) | Unit tests: mutation without header → 403; with header → passes; GET without header → passes; applies to `/auth/*` too |
| **5 — Logout endpoint** | `POST /api/v1/auth/logout` handler → `authcookie.Clear`; route in the public `auth` group under CSRF mw but **not** `Auth` mw; `200` | Flow test: login → logout → cookie cleared (`Max-Age=0`); logout with expired/no cookie still 200 |
| **6 — Swagger + docs regen** | Replace the `Bearer` security definition with cookie auth note + `@securityDefinitions.apikey CSRF (header: X-CSRF-Protection)`; annotate mutations `@Security CSRF`; add UC-03 logout; `swag init` | Swagger UI: log in via try-it-out, Authorize CSRF=`1`, mutations succeed |

## Files to touch
- [`internal/config/config.go`](../../collabotask-backend/internal/config/config.go) — explicit CORS defaults (no `*`), add `PATCH`, enumerate `AllowedHeaders`; new `AuthConfig.CookieSecure` (default `true`), `AuthConfig.CookieName` (default `__Host-token`); `Validate()` fail-fast on `AllowCredentials && origins∋"*"`.
- [`internal/delivery/http/middleware/cors.go`](../../collabotask-backend/internal/delivery/http/middleware/cors.go) — reflect-only-if-allowlisted, drop the `AllowedOrigins[0]` fallback, add `Vary: Origin`.
- **NEW** `internal/delivery/http/middleware/csrf.go` — the required-header check.
- [`internal/delivery/http/middleware/auth.go`](../../collabotask-backend/internal/delivery/http/middleware/auth.go) — read cookie instead of `Authorization`; same `ValidateToken` + context.
- **NEW** `internal/delivery/http/helper/authcookie.go` (or similar) — `Set`/`Clear` cookie helpers (single source for attributes; keeps handlers thin).
- [`internal/delivery/http/handler/auth_handler.go`](../../collabotask-backend/internal/delivery/http/handler/auth_handler.go) — login/register `Set` the cookie; **new `Logout` handler**.
- [`internal/delivery/http/response/auth_response.go`](../../collabotask-backend/internal/delivery/http/response/auth_response.go) — remove `Token` from `AuthResponse`.
- [`internal/delivery/http/router/router.go`](../../collabotask-backend/internal/delivery/http/router/router.go) — add `POST /auth/logout`; register the CSRF middleware (global, after CORS).
- `.env.example` — `CORS_ALLOWED_ORIGINS` explicit example, `COOKIE_SECURE`, `COOKIE_NAME`.
- Swagger annotations across mutation handlers + `docs/` regen.

## Config / `.env.example` deltas
```
# was: CORS_ALLOWED_ORIGINS=*    (invalid with credentials)
CORS_ALLOWED_ORIGINS=https://app.collabotask.com   # dev: http://localhost:3000
CORS_ALLOWED_METHODS=GET,POST,PATCH,DELETE,OPTIONS  # add PATCH
CORS_ALLOWED_HEADERS=Content-Type,X-CSRF-Protection # enumerate (no *)
CORS_ALLOW_CREDENTIALS=true
COOKIE_SECURE=true          # false only for non-localhost http dev/CI
COOKIE_NAME=__Host-token    # when COOKIE_SECURE=false, use a plain name (prefix requires Secure)
```

## Testing
- **Middleware unit tests** (gin test context, `httptest.ResponseRecorder`, stub next-handler; no DB): `Auth` — valid cookie → context userID; absent/invalid → 401. `CSRF` — mutation missing header → 403; present → pass; safe method → pass; applies to `/auth/*`.
- **Flow tests** (`httptest.NewServer(router)` + `net/http/cookiejar`, run with `COOKIE_SECURE=false`): login → jar captures cookie → authenticated mutation with `X-CSRF-Protection` succeeds; logout → cookie cleared. A `authedRequest` helper sets `Content-Type` + the CSRF header.
- **One attribute-assertion test** (independent of flow): inspect the raw `Set-Cookie` from login → asserts `HttpOnly`, `SameSite=Lax`, `Path=/`, and — when `COOKIE_SECURE=true` — `Secure` + `__Host-` prefix + no `Domain`. Keeps prod hardening covered even though flow tests run insecure.

## Tooling under hard cutover (for the PR description / dev README)
- **Swagger:** served same-origin → after `POST /auth/login` via try-it-out the browser holds the cookie; add the `CSRF` apiKey header security scheme so **Authorize → `1`** injects `X-CSRF-Protection` on mutations. Works over `http://localhost` even with `Secure` (browser localhost exemption).
- **Postman:** automatic per-domain cookie jar (login once); add `X-CSRF-Protection: 1` as a **collection-level header**; Authorization = No Auth. Use `https://` or `COOKIE_SECURE=false`.
- **curl:** `-c cookies.txt` on login, `-b cookies.txt -H 'X-CSRF-Protection: 1'` on mutations.

## Gotchas
- **`Secure` + plain `http`:** non-browser clients (curl/Postman/Go cookiejar) honor `Secure` literally and won't resend the cookie over `http://` — hence `COOKIE_SECURE=false` for local/CI (and drop the `__Host-` prefix then, since it *requires* `Secure`). Browsers special-case `localhost`, so Swagger is fine either way.
- **Deleting a `__Host-` cookie** requires re-issuing it with the **same** attributes (`Secure; Path=/`, no `Domain`) + `Max-Age=0` — a mismatch leaves it uncleared.
- **Middleware order:** `CORS` (handles preflight/OPTIONS) → `CSRF` (state-changing header gate, global incl. `/auth`) → `Auth` (cookie validation, protected groups only).
- **Max-Age vs JWT expiry:** always derive `Max-Age` from `cfg.JWTExpiration` — never hardcode. (Reconcile SRS §3.5 "7-day" vs code default 24h separately; design is expiration-agnostic.)
- **Register auto-login is provisional** — remove it when email verification lands (Phase 2+).

## References
- **[ADR-008](../architecture/adr/adr-008-auth-httponly-cookie.md)** — the decision record (8 decisions + deferrals).
- **SRS §3.5 / §4.1** — auth contract (cookie delivery, CSRF header, CORS, UC-01/02 responses, UC-03 logout).
- **[ADR-009](../architecture/adr/adr-009-websocket-realtime-layer.md)** + `docs/handoff/websocket-participation-broadcast/index.md` — the ④ work this unblocks; recheck for drift after ③.5.
- Memory: `[[auth-cookie-migration-prereq]]`, `[[websocket-participation-broadcast-design]]`.
