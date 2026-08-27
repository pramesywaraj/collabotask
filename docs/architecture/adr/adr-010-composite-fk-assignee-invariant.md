# ADR-010: Composite-FK Assignee Invariant

- **Status:** Accepted
- **Date:** 2026-08-24
- **Scope:** Database-level enforcement of the "assignee ∈ board members" rule on `cards`. Complements — does **not** replace — the app-layer write-time check (step ③, SRS §2.8) and the participation cascades (step ③/④). Tested via the ADR-015 harness. Second item of the ⑤ post-Phase-1 integrity pass (after the harness, before the ADR-007 Transactor/Option-C work).

## Context

"A card's assignee must be a board member" is a core domain rule (`CONTEXT.md` › *Assignment = participation*; SRS §2.8; memory `assignment-participation-model`). Today it is enforced **only in the application layer**, by two guards:

- **Guard #1 (write-time):** `create_card` / `update_card` call `boardMemberRepo.IsUserExists` and return `ErrAssigneeNotBoardMember` (400) for a non-member. Validates a *newly-set* assignee only (Decision B, step ③).
- **Guard #2 (membership loss):** `RemoveWithParticipationCascade` (board path + workspace path) clears the departing user's assignments and returns the affected cards, which ④ broadcasts as `CARD_UPDATED{assigned_to:null}` via `common.BroadcastClearedCards` (3 sites).

Both shipped and are tested. The gap: the rule lives only in app code. Two failure modes remain:

1. A check→write race (TOCTOU): guard #1 passes, the user leaves, the write lands a dangling assignee.
2. Any future path that forgets guard #2 leaves stale assignments silently.

This was recorded 2026-07-15 as the "headline candidate" of the ⑤ integrity pass, on the premise that a composite FK would make the cascade *automatic* and let us **delete** guard #2. Re-grilled 2026-08-24.

## The finding that reframed the decision

Step ④ shipped between the original note and this grill, making the affected-card **list** load-bearing: guard #2's `UPDATE … RETURNING` feeds the `CARD_UPDATED` broadcasts. A composite FK's `ON DELETE SET NULL` nulls the cards **silently** — it produces no list — so **it cannot replace guard #2.** The FK's value therefore shifts from *simplification* to *defense-in-depth*: it deletes no code and changes no observable behavior; it only makes the invariant impossible to violate.

## Decision

Adopt the composite FK as a DB-level invariant + backstop. **Keep both app guards.**

### Schema (migration 000009)

- Add `cards.board_id UUID NOT NULL`, `REFERENCES boards(id) ON DELETE CASCADE`. Denormalized: a card reaches its board only transitively (via `columns`) today; the composite FK needs `board_id` on the row itself. Set once at creation from the card's already-validated `column.BoardID` (`create_card.go:27,31`); **never updated**, because a card's board is immutable — `move_card` rejects cross-board moves with `ErrInconsistentState` (`move_card.go:48`) and columns never change boards.
- Add the composite FK, **explicitly named**:
  ```sql
  ALTER TABLE cards
    ADD CONSTRAINT fk_cards_assignee_board_member
    FOREIGN KEY (board_id, assigned_to)
    REFERENCES board_members (board_id, user_id)
    ON DELETE SET NULL (assigned_to);
  ```
  `board_members` PK is `(board_id, user_id)` — a ready-made composite target. `assigned_to` is nullable ⇒ default MATCH SIMPLE skips the check when it is NULL, so unassigned cards are unconstrained. `ON DELETE SET NULL (assigned_to)` (PG15+) nulls only the assignee, not `board_id`.
- Keep the existing `assigned_to → users ON DELETE SET NULL` FK alongside (referential integrity for the user; removing it is risk for no gain).

### Required reorder of guard #2 (behavior-preserving)

Both cascades today **delete the membership first, then unassign the cards** (`board_member.go` delete→unassign; `workspace_member.go` board-membership-delete→unassign). With the FK in place, deleting the membership auto-nulls the cards, so the subsequent `UPDATE … RETURNING` matches nothing → the broadcast list comes back **empty** (a silent regression). Fix: **unassign-and-collect the cards first, then delete the membership**, in both paths. The card-unassign query does not reference `board_members`, so it is safe to run first.

### Migration cleanup

Pre-existing rows that already violate the rule would block `ADD CONSTRAINT`. The migration first nulls any violating `assigned_to` (with `RAISE NOTICE` of the count), then adds the FK. This is a **no-op** on any database that only went through the app since step ③; it only bites on legacy dirty data, and nulling an unreachable assignee is the correct end state.

### Error surface

A write that reaches the FK (the race case) raises PG `23503`. Translate it in the card repo — keyed on the **constraint name** `fk_cards_assignee_board_member`, not the bare code — to the existing `domain.ErrAssigneeNotBoardMember` (already maps to 400), mirroring the `23505 → ErrAlreadyMember` pattern in `workspace_member.go`. Guard #1 stays the primary friendly gate; the FK returns the *identical* 400 in the race. Any **other** `23503` on `cards` stays a 500 (genuinely unexpected).

## Considered options

- **A — Adopt as DB invariant (chosen).** A cheap, permanent guarantee: board-immutability makes `board_id` set-once (no trigger, no ongoing sync burden). Closes the TOCTOU race, backstops future paths, and encodes a core domain rule in the schema. The integrity pass is the right moment.
- **B — Defer to Phase 2.** Rejected. Since the app already enforces the rule, ⑤ would shrink to "write the deferred tests." But the guarantee is unusually cheap and permanent *here specifically*, and a core domain rule is worth encoding in the schema during a pass literally named "integrity." The trigger that would have revived it: cross-board card move, or a real correctness bug slipping past the guards.

## Consequences

**Positive**
- "Assignee ∈ board members" becomes a DB guarantee on every write path; the TOCTOU race is closed structurally.
- Any future/raw path that drops a membership auto-nulls assignments — a permanent floor under Phase-2 code.
- The schema now documents a core domain rule at the point it is enforced.

**Negative / risks**
- `cards.board_id` denormalizes the card→board link (a second source of truth), kept honest only by board-immutability. **If cross-board card move ever lands (Phase 2), it MUST update `board_id` and re-validate the assignee**, or the FK will reject the move — arguably the FK protecting correctness, but a sharp edge to remember.
- Reopens ④'s shipped, reviewed cascade code for the reorder (behavior-preserving, but a real edit — see risks in the handoff).
- **Contradicts the original 2026-07-15 framing**: the FK does **not** delete guard #2 and does **not** supersede the §2.8 shared-helper design. Both stay.

## Relationship to other ADRs / docs

- **ADR-015** — integration harness; the FK cascade behavior + reorder + migration cleanup are tested here.
- **ADR-009** — the ④ broadcast whose affected-card list is the reason guard #2 must stay.
- **ADR-007** — the Transactor / Option-C atomic-activity work, the next ⑤ item; it composes the post-reorder cascade code, so this FK lands first.
- **SRS §2.8**, memory `assignment-participation-model`, memory `post-phase1-integration-tests` — the app-layer contract this hardens.
- Supersedes the **"candidate"** status recorded in `docs/handoff/handoff-assignee-validation-cascade.md` (§ *Deferred*) and `post-phase1-integration-tests`.
