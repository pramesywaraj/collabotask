# ADR-008: Authentication — migrate JWT delivery from `Bearer` header to httpOnly cookie

- **Status:** **Proposed — design pending** (reserved number; to be written in full after its own `/grill-with-docs` session, before build step ③.5). This file reserves ADR-008 and records the *intent + why-now*; it is **not** the finished decision record.
- **Date:** 2026-07-21 (reservation)
- **Scope:** How the Phase-1 JWT reaches the server across REST **and** WebSocket **and** SSR — moving from the current `Authorization: Bearer` header (JS-held token) to an **httpOnly, Secure, SameSite cookie**. New build step **③.5**, sequenced **before** step ④ (WebSocket, [ADR-009](./adr-009-websocket-realtime-layer.md)).

## Why this exists now (context)

Grilling the WebSocket layer (ADR-009) surfaced that "how does the JWT reach the WS handshake" is a symptom of a deeper decision: **where the token lives on the client.** The current `Bearer`-in-JS model has three problems for the target app (Next.js, public SSR zone + auth-gated app zone, realtime over WS):

1. **WebSocket:** a browser's native `WebSocket` cannot set an `Authorization` header → forces a `?token=` query param or a `Sec-WebSocket-Protocol` hack.
2. **SSR:** `localStorage`/JS-memory is invisible to the Next.js **server**, so it cannot make authenticated calls during server render — SSR-authenticated data fetching is impossible without a cookie.
3. **XSS:** a JS-readable token is exfiltratable — the #1 practical token-theft vector.

An **httpOnly cookie** solves all three at once: it rides the WS handshake and every REST/SSR request automatically, and JS cannot read it. The product owner chose to **do this properly and first**, as a fundamental prerequisite of the realtime layer, rather than smuggle a token into the WebSocket. It forces **explicit CORS origins** (no `*` with credentials), which ADR-009 shares between REST and WS.

## Decisions to settle in the full grilling (not yet decided)

- **Cookie attributes** — `httpOnly` + `Secure` + `SameSite` (Lax vs Strict), `Path`, `Max-Age` vs the JWT lifetime; the local-`http` dev wrinkle (`Secure` + `localhost` exemption; docker/staging need HTTPS).
- **Cookie contents** — keep the **JWT itself** (stateless, smallest change — move it from the login response body into `Set-Cookie`) vs a session id backed by a table.
- **CSRF** — now mandatory (cookies auto-send): `SameSite=Lax` baseline + double-submit token or required custom header for state-changing requests. (WS gets CSWSH protection from the shared origin check — ADR-009 §origin.)
- **Migration ergonomics** — dual-read (cookie *and* `Bearer`) so Swagger/tests/tooling keep working, vs a hard cutover.
- **Logout** — `Bearer` had no server-side logout; cookie auth needs an endpoint that clears the cookie.
- **Origins** — flip `CORS_ALLOWED_ORIGINS` from `*` to an explicit list; `AllowCredentials: true` (already set).

## Consequences (intent-level)

- **Unblocks ADR-009** with a clean WS handshake (read cookie → `ValidateToken` → reject before upgrade); removes the `?token=`/subprotocol workarounds; enables SSR auth; hardens against XSS.
- Touches the auth middleware, login/register handlers (`Set-Cookie` instead of/alongside a body token), CORS config, and adds CSRF + logout — a REST-auth change, **kept separate from the WS work** so both stay reviewable.
- **This stub must be replaced** by the full ADR (Options Considered / Decision / Consequences) after the auth grilling, before ③.5 is built.
