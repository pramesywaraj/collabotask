# ADR-005: Board Visibility — Access Enforcement (Metadata / View / Mutate + Break-Glass)

- **Status:** Accepted
- **Date:** 2026-07-10
- **Scope:** How board content access is authorized once boards have `visibility ∈ {PRIVATE, WORKSPACE}` (SRS §2.2–2.4, UC-07/08/09/12/12b). Covers the access checker's shape, break-glass semantics, hidden-vs-revealed responses, self-join idempotency, and the board-detail roster. Implements the P1 "missing board visibility" item in SRS §9. Aligns with SRS §2.3 (layered permission) and §2.8 (governance vs participation).

## Context

Before visibility, every board was treated identically: `boardAccessChecker` granted content access to any workspace admin, the board creator, or any board member — with **one** rule used for both *viewing* (kanban/detail) and *mutating* (cards/columns). Adding `visibility` breaks that single rule apart, because the SRS permission matrix (§2.4) grants different access to the same actor depending on **what they're doing** and **the board's visibility**:

- A **plain workspace member** may *view* a WORKSPACE-visible board's content but may **not** *mutate* it (§2.4: open content ✅, create/edit card ❌).
- A **workspace admin** governs every board, but §2.3's **break-glass** rule says they must **Join** (a logged action) before acting on a **PRIVATE** board's content — while a WORKSPACE-visible board needs no join.

So a single boolean "has access" is no longer expressible: view and mutate diverge, and the divergence depends on visibility. The checker must encode a small matrix, and several of its cells were genuine decisions with defensible alternatives — recorded below so they aren't silently re-litigated.

The full authorization matrix this ADR implements:

| Requester | Metadata (detail) | View (kanban) | Mutate (card/column) |
|---|---|---|---|
| Board owner / member | ✅ | ✅ | ✅ |
| Workspace **admin**, not joined, **WORKSPACE** | ✅ | ✅ | ✅ |
| Workspace **admin**, not joined, **PRIVATE** | ✅ `CAN_JOIN`, roster hidden | 🔓 `BOARD_JOIN_REQUIRED` | 🔓 `BOARD_JOIN_REQUIRED` |
| Workspace **member**, not on board, **WORKSPACE** | ✅ | ✅ | ❌ 403 |
| Workspace **member**, not on board, **PRIVATE** | ❌ 404 | ❌ 404 | ❌ 404 |

## Options Considered

### Shape of the checker — one rule, an intent flag, or three methods

- **One rule (rejected)** — no longer expressible; view and mutate genuinely differ for the plain-member/WORKSPACE cell.
- **One method + `intent` enum (rejected)** — collapses three policies behind a parameter every call site must pass and every reader must decode. Easy to misread; not self-documenting.
- **Three methods (chosen):** `CheckMetadataAccess` (detail), `CheckViewAccess` (kanban), `CheckMutateAccess` (card/column). A private `resolve()` gathers the facts (board + workspace membership + board membership + visibility) once; each public method applies its policy tail. Call sites read as their intent. The three levels are genuinely different policies, so three named methods beat a flag.

### Does break-glass cover mutations, or only viewing?

This is the sharpest tension in the SRS. §2.4's matrix lists card/column mutation as **✅ for "Workspace ADMIN (even if not board member)"** — read literally, an admin could mutate a PRIVATE board *without joining*. But §2.3 says: "**To act on the content of a PRIVATE board, the admin Joins first** (one click, recorded in the activity log)."

- **Literal matrix reading (rejected)** — admins mutate PRIVATE boards without ever joining. Leaves **no audit trail** for that touch: the whole point of break-glass is to make "an admin reached into a private board" recorded. A backdoor that skips the kanban and POSTs directly to a card would bypass the logged join.
- **Break-glass covers mutations too (chosen)** — any content access to a PRIVATE board by a non-joined admin returns `BOARD_JOIN_REQUIRED`, **view and mutate alike**. §2.4's "✅(admin)" is read as the **post-join steady state**: once the admin has joined (which the gate forces), their mutation authority comes from the admin role, not the `BOARD_MEMBER` row (consistent with §2.3's explanation of why a joined admin keeps governance). On WORKSPACE-visible boards there is no privacy boundary to breach, so no join is required — matching the mainstream convention in Trello/Jira/Asana/ClickUp/Linear, where a join-first ceremony (when it exists) is reserved for private content.

### Denied response — 404 (hide) vs 403 (reveal)

For an **ineligible plain member** on a **PRIVATE** board (metadata, view, or mutate):

- **403 for all (rejected)** — tells the member "a board exists here you can't see," leaking existence and contradicting the board **list** (UC-08), which already filters PRIVATE boards the member can't see out of the result entirely.
- **404, hide existence (chosen)** — the board is uniformly invisible to that member: absent from the list, `404` on direct hit. Mirrors GitHub's private-repo behavior and blocks board-ID enumeration.

The **non-joined admin** is a *different* actor and keeps **403 `BOARD_JOIN_REQUIRED`** — they can already see the board exists (it's in their list), so the honest response is a prompt to join, not a lie that it's absent. On a **WORKSPACE** board, a plain member denied *mutation* gets **403**, not 404: they legitimately see the board, they just can't edit it. So the rule is precise — 404 only where the board should be invisible (PRIVATE + ineligible), 403 everywhere the actor may legitimately know it exists.

### Self-join when already a member — error, silent success, or informative success

UC-09's SQL is `ON CONFLICT DO NOTHING` (idempotent), but the existing code returned an error on re-join, and the error was overloaded to mean both "ineligible" and "already a member."

- **409 / error on re-join (rejected)** — surfaces a *satisfied intent* as a failure. Self-join is an intent-expressing action ("make me a member") that gets re-fired by double-clicks, retries after a dropped response, and the optimistic/reconnect model (§2.7). Modeling the success end-state as a 409 forces every optimistic client to special-case "409-already-member actually means success."
- **Silent idempotent success (rejected)** — 200 but no signal; the client can't tell "just joined" from "already in," and can't show useful feedback.
- **Informative idempotent success (chosen)** — **200** with body `{ "joined": true }` (newly added) or `{ "joined": false, "message": … }` (already a member). Captures the "tell the user" intent without turning a good outcome into an error. The `joined` flag is also needed **internally**: a re-join must **not** re-log the activity or re-broadcast `USER_JOINED` — `ON CONFLICT DO NOTHING RETURNING *` returns no row on a no-op, which is exactly that signal. Ineligibility (a plain member on a PRIVATE board) is now the checker's `404`, so it never reaches this path.

### Board detail before joining — full metadata, break-glass, or thin metadata

`GET /board/:id` returns not just board fields but the **full member roster including every member's email**. So "admin sees PRIVATE metadata" (§2.4) collides with "roster + emails leak before the logged join."

- **Full metadata incl. roster (rejected)** — a non-joined admin pulls a PRIVATE board's entire roster + emails *without* the break-glass join, partly defeating the audit boundary.
- **Break-glass the detail endpoint too (rejected)** — stricter, but a behavior regression: §2.4 explicitly grants admins PRIVATE *metadata*, and the frontend needs board fields to render a "Join to open" wall.
- **Thin metadata pre-join (chosen)** — detail returns board fields + `access_status`, but **omits the `Members` roster** when the board is **PRIVATE and the requester is not joined**. Preserves "see metadata," keeps the roster behind the join. Because a plain member can *never* reach a PRIVATE board's detail (404), the only actor this affects is the **non-joined admin** — precisely the break-glass actor. On WORKSPACE boards, where a non-joined viewer can already open the kanban, the roster is shown (hiding it there would serve no privacy purpose).

### `created_by` in access/role logic

`board.created_by` was used as an access grant (checker) and as a `user_role`/`access_status` fallback (list + detail).

- **Keep it (rejected)** — redundant and fragile: the creator is *always* seeded as a `board_members(BOARD_OWNER)` row, so `board_members` yields the same answer; and once `created_by` becomes nullable + `SET NULL` (migration 000007) it's an unreliable authority.
- **Remove from all access/role logic (chosen)** — `board_members` (+ workspace-admin status) is the sole source. `created_by` survives only as a stored/returned historical trace. Advances the SRS §9 P2 "two sources of truth for board ownership" item. (Its last logic use, the `leave_board` owner guard, is deferred to UC-12e — see below.)

## Decision

- **Three-method access checker** over a private `resolve()`: `CheckMetadataAccess`, `CheckViewAccess`, `CheckMutateAccess`, enforcing the matrix in Context.
- **Break-glass is PRIVATE-only and covers view *and* mutate:** a non-joined admin on a PRIVATE board gets `ErrBoardJoinRequired` → **403 `BOARD_JOIN_REQUIRED`**. On WORKSPACE-visible boards admins act freely, no join.
- **Denied responses:** ineligible plain member on a PRIVATE board → **404** (`ErrBoardNotFound`, existence hidden); plain member denied mutation on a WORKSPACE board → **403**.
- **Self-join:** eligibility — admin → any board; member → WORKSPACE-visible only; ineligible → **403** (`ErrBoardCannotJoin` remapped 409→403). Already-member → **idempotent 200** + `{ joined: bool, message? }` via `ON CONFLICT DO NOTHING RETURNING *`.
- **Board detail:** thin metadata pre-join — omit the `Members` roster when **PRIVATE + not joined**.
- **`created_by` removed** from the checker, the list query filter + `user_role`/`access_status` fallbacks, and the detail fallback.
- **Mutation use cases** (card create/update/delete/move, column create/update/delete/reorder) switch from the old `Check` to `CheckMutateAccess`.

### Out of scope (deferred, recorded so they aren't lost)

- **Self-join activity log** (UC-09) → the `activities` sub-step of build step ③. **WebSocket broadcasts** (`USER_JOINED`, `BOARD_UPDATED`) → build step ④. Self-join and settings-change are intentionally silent for now.
- **`Board.CreatedBy` → `*uuid.UUID`** → Phase 2, with account deletion (SRS §10) — the only feature that can make it NULL. Kept as `uuid.UUID` now.
- **`leave_board` owner guard → `board_members.role`** → UC-12e (ownership transfer), where `created_by` stops implying the current owner. Correct until then.

## Consequences

**Positive**

- The permission matrix is enforced in **one place**, by three intent-named methods, instead of scattered `if admin || member` checks across ten use cases.
- Break-glass is airtight: **every** content touch on a PRIVATE board by a non-member admin is gated by the (soon-to-be-logged) join — no view/mutate asymmetry to exploit.
- PRIVATE boards are uniformly invisible to ineligible members (list *and* direct hit both hide them); no ID-enumeration leak.
- Optimistic/reconnect clients can re-fire self-join safely; the `joined` flag drives both UX and the (later) log/broadcast suppression.
- `board_members` is the single source of role/access truth; `created_by` is demoted to a trace, advancing the §9 P2 cleanup.

**Negative / notes**

- **Three access levels must be kept straight.** A future endpoint must pick the right method; using `CheckViewAccess` where it means to mutate would wrongly admit a plain member on a WORKSPACE board. The method names are the guardrail.
- **404-vs-403 is deliberate, not incidental.** Any refactor that "simplifies" all denials to one status will reintroduce the PRIVATE-board existence leak. The split (404 hide / 403 reveal) is a security property.
- **Break-glass is only auditable once `activities` and WebSocket land.** Until then the join is enforced but not yet *logged/broadcast* — the audit value is designed-in but not fully realized within this step.
- **Thin-roster relies on the 404 gate.** It's safe to hide the roster only for the non-joined admin *because* plain members can't reach PRIVATE detail at all. If that 404 ever weakens to 403-with-body, the roster-hiding must be re-examined.
