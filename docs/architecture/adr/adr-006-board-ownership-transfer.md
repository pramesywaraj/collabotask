# ADR-006: Board Ownership — Single Source of Truth & Transfer/Appoint Access Model

- **Status:** Accepted
- **Date:** 2026-07-14
- **Scope:** How the *current* board owner is determined, and how ownership is transferred or appointed, once UC-12e lands (SRS §2.5, §2.6, §4.3 UC-12e). Covers: retiring the `created_by`-as-owner proxy in favour of `board_members.role` as the sole ownership authority; the transfer access model (`canAdministerBoard` + break-glass on PRIVATE boards); the single-owner invariant + orphan-board appointment; idempotency; target validation. **Completes the `created_by` cleanup begun in ADR-005** and closes the SRS §9 P2 "two sources of truth for board ownership" item. Aligns with SRS §2.3 (layered permission), §2.5 (single owner + transfer), §2.6 (orphan boards), §2.8 (governance vs participation), and ADR-005 (visibility / break-glass).

## Context

ADR-005 removed `created_by` from board **access-grant** and **role-display** logic, but explicitly **deferred two remaining `created_by`-as-owner uses to UC-12e** (ADR-005 › "Out of scope"):

1. `canAdministerBoard()` — its `createdBy == requesterID` clause (gates update-settings / archive / invite / remove, and will gate transfer itself).
2. `leave_board` — its "owner cannot leave" guard, written as `board.CreatedBy == requesterID`.

Both use `created_by` as a **proxy for the current owner**. That proxy is only correct while `created_by == owner` — i.e. while no transfer exists. **UC-12e is the feature that breaks that equivalence**, so the proxy must die here. Concretely, leaving it in produces two bugs the instant a transfer happens (old owner O keeps `created_by = O`; new owner N holds the `BOARD_OWNER` row):

- **O keeps powers they no longer hold** — O calls update-settings/archive; `createdBy == requesterID` is still true → O administers a board they've handed off (unless they're independently a workspace admin, in which case they'd keep power *via the admin layer* anyway).
- **The real owner can walk out** — N tries to leave; `board.CreatedBy(O) == N` is false → the "owner cannot leave without transferring" guard never fires → the board's actual sole owner leaves without handing off, exactly the state the guard exists to prevent.

Transfer also surfaces the **sharpest permission tension** in the SRS: whether a workspace admin who is **not** a board member must **Join** a PRIVATE board before transferring ownership on it. §2.4's matrix lists "Transfer board ownership" in the *governance* group (✅ for admin even if not a board member), alongside edit-settings / archive / delete — none of which require a join. But §2.6 says: *"A workspace admin can appoint a new owner… **For PRIVATE boards, the admin Joins first (logged), then appoints.**"* These read in opposite directions and must be reconciled.

## Options Considered

### Owner source of truth — `created_by` proxy vs `board_members.role`

- **Keep the `created_by` proxy (rejected)** — breaks on the first transfer (the two bugs above). `created_by` is also becoming an unreliable authority: migration 000007 already made it nullable + `ON DELETE SET NULL` for Phase-2 account deletion.
- **`board_members.role` is the sole authority (chosen)** — the `BOARD_OWNER` row *is* the current owner; `created_by` is demoted to a stored historical trace only. `canAdministerBoard()` drops its `createdBy` parameter and derives owner-authority from the board-member row (`IsOwner()`) plus workspace-admin status; `leave_board`'s guard looks up the `BOARD_OWNER` row instead of `created_by`. This finishes what ADR-005 started and closes SRS §9 P2.

### Transfer access on PRIVATE boards — governance (no join) vs break-glass (join first)

- **Pure governance / no join (rejected)** — consistent with delete/remove, *but* on a PRIVATE board the thin-roster rule (ADR-005) strips the member list out of the board-detail response for a non-joined admin. So the admin has **no way to see who to hand ownership to**. Making governance usable would then require *either* a new "list board members (admin)" endpoint *or* repealing thin-roster — both worse than the alternative below.
- **Repeal thin-roster for admins (rejected)** — hands a non-joined admin the full PRIVATE roster (incl. emails) with no logged join, reversing a deliberate ADR-005 decision and weakening the break-glass audit boundary. Too large a product-model change to ride in on a feature commit.
- **Break-glass (chosen)** — a non-joined workspace admin acting on a PRIVATE board's ownership gets **403 `BOARD_JOIN_REQUIRED`** and must Join first. Joining hands them the full roster *through the existing joined-viewer path* — **no new endpoint, thin-roster untouched** — and satisfies §2.6's "Joins first, then appoints." It reuses ADR-005's principle (any PRIVATE-content touch by a non-member admin is join-gated) rather than carving out an exception. The single cost: the admin becomes a `BOARD_MEMBER` on that board (reasonable — they're administering it — and they can leave afterward). Owner-initiated transfers and all WORKSPACE-visible boards need **no** join.

  §2.4's "✅(admin)" for transfer is therefore read the same way ADR-005 reads the mutation row: as the **post-join steady state**, not a licence to skip the join.

### One endpoint — transfer only, or transfer + orphan appointment

- **Separate "appoint owner" endpoint for orphan boards (rejected)** — duplicates ~90% of transfer (§2.6 allows admins to appoint an owner to an owner-less board; there is no separate UC for it).
- **One "set sole owner" operation (chosen)** — a single atomic op: **demote-by-role** (`WHERE role='BOARD_OWNER'`, matches 0 or 1 rows) + **promote-by-id** (the target). Orphan appointment falls out for free: on an owner-less board the demote step matches **0 rows** and we simply promote. No special branch, no second endpoint.

### Target is already the owner — reject vs idempotent

- **Reject 400/409 (rejected)** — punishes harmless retries (double-click, optimistic re-fire, admin race) and is stricter than the rest of the codebase.
- **Idempotent 200 no-op (chosen)** — mirrors workspace `SetMemberRole` ("already has the requested role → 200 no-op"). Checked in the usecase *before* touching the repo.

### Target is not a board member — reuse 404 vs dedicated 400

- **Reuse `ErrBoardMemberNotFound` → 404 (rejected)** — reads as "the board / endpoint doesn't exist," misleading for a bad body field.
- **Dedicated `ErrTransferTargetNotBoardMember` → 400 (chosen)** — mirrors the assignee-must-be-a-board-member rule (`ErrInvalidAssigneeID` → 400, §2.8). 400 correctly says "fix your input — add/join them to the board first" (US-12e: "they must join first").

## Decision

- **`board_members.role` is the sole source of truth for the current board owner.** `created_by` is a historical trace only. `canAdministerBoard()` drops its `createdBy` param (owner-authority = the `BOARD_OWNER` board-member row **or** workspace-admin status); `leave_board`'s owner guard looks up the `BOARD_OWNER` row. This removal is the **first commit** of UC-12e.
- **Permission = `canAdministerBoard`** (current board owner **or** workspace admin), **with break-glass on PRIVATE boards:** a non-joined admin → **403 `BOARD_JOIN_REQUIRED`** (must Join first, which is logged once activities land). Owner path and WORKSPACE-visible boards: no join.
- **One endpoint, one operation — "set sole owner":** atomic **demote-by-role + promote-by-id** in a single transaction; covers **both** transfer (owner exists) and **orphan appointment** (no owner) with no extra branching. The single-owner invariant is preserved (0 owners only on an orphan board).
- **Target already owns the board → idempotent 200 no-op** (guarded in the usecase before the repo call).
- **Target is not a board member → 400 `ErrTransferTargetNotBoardMember`** (new sentinel).
- **Archived board → 404 `ErrBoardNotFound`**, matching the whole board-membership-mutation family (invite / remove / leave / self-join).
- **Endpoint:** `POST /api/v1/workspace/:workspace_id/board/:board_id/transfer-ownership`, body `{ "to_user_id": "uuid" }`.

### Out of scope (deferred, recorded so they aren't lost)

- **Transfer activity-log write** (UC-12e) → the `activities` sub-step of build step ③. **`OWNERSHIP_TRANSFERRED` broadcast** → build step ④. Transfer is intentionally silent (no log/broadcast) until then, exactly like self-join in ADR-005.
- **Richer board/workspace roles + owner delegation** (e.g. an intermediate "board manager" role, an acting-owner concept) → **Phase 2+**. Phase-1's two-roles-per-layer is a *conscious* minimal choice; the "owner unavailable" case is already covered by the layered model (admins govern in parallel) + orphan-safe boards + this transfer/appoint endpoint. Not a Phase-1 gap.
- **`Board.CreatedBy` → `*uuid.UUID`** → Phase 2, with account deletion (SRS §10) — unchanged from ADR-005.

## Consequences

**Positive**

- The `created_by` cleanup is **fully complete**: `board_members.role` is now the single ownership authority across access, role-display, *and* owner-authority checks. SRS §9 P2 "two sources of truth for board ownership" closes.
- The two proxy bugs (old owner keeps power; real owner can leave) are eliminated by construction, with regression tests to lock them shut.
- Transfer break-glass is consistent with ADR-005's PRIVATE-content rule — **no new endpoint, no thin-roster repeal** — and honours §2.6.
- Orphan-board appointment needs **no special code path**; it's the same "set sole owner" op with a zero-row demote.

**Negative / notes**

- **Transfer is deliberately stricter than delete/remove on PRIVATE boards** (it requires a join; they don't). This is intentional: transfer is the one governance action that *needs the member roster to choose a target*, and the break-glass join is what supplies it — plus the audit trail. Recorded here so the asymmetry isn't "simplified" away later.
- **Break-glass is only fully auditable once `activities` + WebSocket land** — the join is *enforced* now but not yet *logged/broadcast* (same caveat as ADR-005).
- **Roster-discovery for a non-joined admin is a latent gap for `remove-member` too** (it also takes a `user_id` without exposing the roster). The eventual fix — a shared "list board members (admin view)" endpoint — is a frontend-era concern, not transfer's to carry.
- **The single-owner invariant is enforced by the operation, not the schema.** `board_members` permits any number of `BOARD_OWNER` rows; correctness relies on every ownership write going through the atomic demote-by-role + promote path. Any future code that inserts/updates a `BOARD_OWNER` row directly must preserve this.
