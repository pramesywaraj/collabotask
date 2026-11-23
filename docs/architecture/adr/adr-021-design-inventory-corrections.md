# ADR-021: Four Corrections from Walking the Design Inventory

- **Status:** Accepted — **§5 extended by [ADR-034](adr-034-side-menu.md)** (2026-11-21): a fifth presentation kind, *navigation*, for the sidebar on 2.4 and the board. **§6's *one component with two settings* narrowed by [ADR-038](adr-038-board-states-group-d.md)** (2026-11-23): 2.8 is a centred stack (a padlock drawing, the sentence and *Join board*), not the Banner part over a bare canvas. §6's data rules, roster rule and button stand. **§5's *Confirmations* row widened by [ADR-040](adr-040-menus-and-the-confirmation.md)** (2026-11-23): five actions ask first, *Archive board* and *Leave board* as well as the three deletes. The row's rule, a centred dialog at every width, is unchanged. §1's *Undo* after archiving is kept over ADR-030 §3. Everything else stands as written.
- **Date:** 2026-11-19
- **Scope:** Four gaps found by walking [design-file-inventory.md](../design-file-inventory.md) against the running API rather than against the other ADRs. **Supersedes nothing.** It *corrects one factual finding* in [ADR-018](adr-018-state-vocabulary.md)'s *Corrections to ADR-017* table, *extends* [ADR-019](adr-019-responsive-touch-scope.md) §6 from cards to columns, *closes* a question [ADR-019](adr-019-responsive-touch-scope.md) §3 left open, and *completes* ADR-018 §10's treatment of one state. ADR-017's tokens, ADR-018's channel ledger and ADR-020's Penpot structure are untouched. Unlike ADR-017/018/019/020 this ADR is **not** frontend-only: §1 requires one backend change, and the *Downstream edits* section records it.

## Context

ADR-017, ADR-018 and ADR-019 were each written against the ones before it. The design inventory was the first document to walk the whole app zone against the **API surface** — the router, the request and response DTOs, the board access checker and the realtime frame list. Four things that read as settled turned out not to be.

### 1. Unarchive is reachable. ADR-018 said it was not, and the SRS has said it was all along.

ADR-018's *Corrections to ADR-017* struck the archive icon on this finding:

> *Archived boards are filtered out of the list (`AND b.is_archived = FALSE`) and every access path 404s them. **Unreachable in Phase 1, and so is Unarchive***

The first half is correct. The second half is not, and it contradicts a line the SRS has carried since it was written — §3.3: *"archive/unarchive **is in Phase 1** (endpoint already exists in code)."*

The mechanism is a single asymmetry. Every board **read** path resolves through `BoardAccessChecker.resolve()`, which returns `ErrBoardNotFound` the moment `board.IsArchived` is true ([board_access_check.go](../../../collabotask-backend/internal/usecase/common/board_access_check.go)). `SetArchived` **does not use the checker.** It fetches the board directly with `boardRepo.GetByID` — whose query carries no `is_archived` predicate — and runs its own `canAdministerBoard` check ([set_archived.go](../../../collabotask-backend/internal/usecase/board/set_archived.go)).

So `POST …/board/:id/archive {"is_archived": false}` against an archived board **succeeds today.** It writes an `ActivityActionUnarchived` row and broadcasts `BOARD_UNARCHIVED` — a frame `common.FrameBoardUnarchived` already defines and [frontend-architecture.md](../frontend-architecture.md)'s event router already handles.

**What is missing is discovery, not capability.** `getUserBoardsInWorkspace` hardcodes `AND b.is_archived = FALSE` with no parameter ([board_queries.go](../../../collabotask-backend/internal/repository/postgres/board_queries.go)). There is no list in which an archived board appears, so there is nothing to press Unarchive *on*.

That is the whole gap: a P1 promise in US-12c, a working implementation in the backend, a contract line in the SRS — and one SQL predicate standing between them.

### 2. Column reorder is a dragging movement with no alternative

ADR-019 §6 requires a non-drag path for card moves on **WCAG 2.2 SC 2.5.7**, and states in terms that *"removing the picker at any breakpoint is a conformance regression, not a simplification. Anyone proposing it must supersede this ADR."*

US-13d (P1) reorders columns by dragging. SC 2.5.7 does not distinguish what is being dragged. Columns were given no equivalent in ADR-019, in ADR-018, or in the inventory — so the project built an explicit conformance argument for one draggable object and silently skipped the other.

A second hole sits in the same place and is unrelated to conformance. ADR-019 §3 makes the phone board a scroll-snapped horizontal strip, and §5 **freezes that strip during a card drag** specifically to delete the snap-versus-autoscroll problem. Dragging a *column* travels along that same frozen axis. Nothing decided whether it is possible, let alone how.

### 3. Fifteen overlays, one decided

ADR-019 §3 commits **only the card modal** to becoming a full-height sheet below 640px. The inventory lists roughly fifteen overlays and records the cost of leaving the rest open: *"discovering it per-overlay while drawing is the expensive order."*

One structural fact decides most of it, and it was not in view when ADR-019 §3 was written: **several overlays open on top of an overlay that is already full-height.** The destructive confirmation is reached from inside the card drawer. So are the date picker and the assignee picker. A full-height sheet over a full-height sheet leaves the user two layers deep with no sense of what they are confirming and two things to dismiss.

### 4. The read-only board has exactly one population, and every one of them can leave it in one click

ADR-018 §10 renders state 24 as *"mutating controls absent + a persistent, non-dismissible banner"*, inheriting ADR-019 #3's reasoning that on touch the banner is the only remaining signal. Both are right. Neither asked **who sees it**.

Read off the access checker: `CheckMutateAccess` refuses a non-joined plain member, while `CheckViewAccess` admits them. A workspace **admin** never lands here — they hold mutate access on WORKSPACE boards without joining, and are bounced to the join gate (25) on PRIVATE ones. So state 24 has exactly one population: **a plain workspace member, on a workspace-visible board, who has not joined.**

And `SelfJoinBoard` permits precisely that person to join: *"a plain member may only join a WORKSPACE-visible one"* ([self_join_board.go](../../../collabotask-backend/internal/usecase/board/self_join_board.go)).

The banner is therefore not a warning about a permanent condition. It is the prompt for a one-click action — and as specified it does not carry one.

---

## Decision

### 1. Archive stays two-way. The board list gains an `archived` filter.

US-12c is honoured as written. Three things follow.

**The backend change.** `GET /api/v1/workspace/:workspace_id/board` takes an optional query parameter:

| | |
|---|---|
| `?archived=false` (default, and the absent case) | Current behaviour exactly — `AND b.is_archived = FALSE` |
| `?archived=true` | `AND b.is_archived = TRUE`, same visibility clause, same meta columns |

Only the predicate changes. The visibility rules, the counts and the ordering are untouched, so an archived board is visible to exactly the people who could see it before it was archived. This is the single smallest edit that closes the gap, and it is why this option was chosen over amending US-12c.

**The interface.** An **Active / Archived** switch above the board grid on the workspace screen. **Not a separate route** — the archived list is the same grid with a different filter, so it inherits the tile, the empty state and the loading state already being drawn. An archived tile's only action is **Unarchive**; opening one is still a 404 by `BoardAccessChecker`, and that is left alone.

> **The switch's look, the archived tile and the empty states are settled in [ADR-033](adr-033-workspace-screens.md)** (2026-11-21). The switch is two `Segment` instances; the current one has plum text and a 1px plum edge. An archived tile keeps its role and counts, has no hover, and shows an Outline *Unarchive* only to the board owner or a workspace admin. The archived list's empty state carries the open-box illustration.

**Archiving the board you are looking at navigates you out.** The actor returns to the workspace with a confirmation message carrying **Undo**. Undo costs nothing — it is the same endpoint with `is_archived: false` — and it is the cheapest way to teach that the Archived filter exists. Everyone *else* viewing the board gets the gone state (23), which is what the `BOARD_ARCHIVED` broadcast already delivers.

### 2. The archive icon decision is **not** reopened.

ADR-018 §9 gives the icon channel exactly one Phase-1 meaning — the filled lock on a private board — and struck archive from it. That striking **survives this ADR intact**, by a better argument than the one it was made on.

An archive icon marks a row as archived *among rows that are not*. In the Active list no archived board appears; in the Archived list **every** row is archived, so the icon marks nothing the filter has not already said. The channel's single meaning is preserved, and ADR-018 §9 needs no amendment — only its stated *reason* changes from "unreachable" to "redundant".

### 3. Column reorder gains **Move left / Move right**, in the column `⋯` menu.

Beside Rename and Delete. Available at every width, on every pointer type, and keyboard-reachable — which is what SC 2.5.7 requires and what ADR-019 §6 established as non-negotiable.

**It is deliberately not a picker.** ADR-019 §6 built the card picker as *column + neighbour* because a column is unbounded — "put it third" is unavailable when the API takes a float position and the column may hold two hundred cards. Columns are different in kind: a board seeds three, and the whole set is on screen or one swipe away. A relative nudge names the destination exactly as well as a list would, costs two menu items instead of an overlay, and adds nothing to §5's presentation rule.

Both entries compute a midpoint client-side and call the existing `PATCH …/columns/:id/position`, exactly as a drop does. At either end of the board the inapplicable entry is **absent**, per ADR-018 §5 — the leftmost column has no *Move left*.

> **The `⋯` that opens this menu is shown at rest, at every pointer.** ADR-019 §6 hides the *card's* `⋯` until hover at a fine pointer, and this section did not say whether the column's follows it. It does not: see **[ADR-027](adr-027-column-header.md)** §4 (2026-11-20).

### 4. Column dragging exists at **640px and up only**.

Below 640px, columns are reordered through §3 and not by gesture. This is the same trade ADR-019 §5 made for card reach: *"Freezing the board deletes that entire class of problem."* Dragging a column along the axis the strip scrolls on reopens the snap-versus-drag arbitration that §5 was written to avoid, and it would do so for the one object on the board that has a complete, equally fast alternative one tap away.

Card dragging on phones is **unaffected** — ADR-019 §4 and §5 stand as written.

A consequence for the design file: the drag board draws **two** placeholders. The card placeholder is ADR-018 §4's 2px dashed `--drop-placeholder` at card geometry; the column placeholder is the same token at **column-panel geometry** — full panel height, panel width. Both are desktop/tablet states; only the card one has a phone equivalent.

### 5. Overlay presentation below 640px — four kinds, one rule each.

ADR-019 §3's card-modal commitment stands and generalises into the first row.

| Kind | Members | Below 640px | 640px and up |
|---|---|---|---|
| **Forms and lists** | Card drawer, create workspace, create board, board settings, board members, add people, transfer ownership, Move card to…, the workspace team panel, presence overflow | **Full-height sheet** | Centred dialog or drawer |
| **Menus** | Board `⋯`, column `⋯`, card `⋯`, user menu | **Bottom action sheet** | Anchored popover |
| **Confirmations** (38) | Delete column, delete card, delete workspace | **Centred dialog — unchanged at every width** | Centred dialog |
| **Inline pickers** | Due date, assignee | **Expand in place inside the card drawer** | Expand in place, or an anchored popover |

**Confirmations and pickers are the rule's whole point.** Both are routinely reached from inside something that is already a full-height sheet — Delete from the card drawer, the date picker from the card drawer. Promoting either to a second sheet stacks two full-height layers, which is the state in which a person cannot see what they are about to delete. A centred dialog over a sheet reads correctly; a sheet over a sheet does not. The confirmation's proportions carry the same argument from the other side: a two-line question occupying 844px overstates itself next to the form that created the thing.

**Menus become a bottom sheet rather than a popover** for the reason ADR-019 §6 already gives the `⋯` its persistent treatment: the trigger is a 13px mark in a 44px box, and an anchored popover hanging off it near a screen edge is the one presentation that reliably lands under a thumb in the wrong place.

> **A fifth kind, *navigation*.** [ADR-034](adr-034-side-menu.md) (2026-11-21) adds a sidebar on 2.4 and the board. Below 1024 it slides over the content from the left, with the content dimmed. From 1024 up it sits beside the content. Its breakpoint is 1024, not 640, because beside the content it would leave a tablet in portrait one column. It never stacks on another layer: it opens only from the app bar's button or its shortcut, and its workspace switcher opens inside it, the rule the inline pickers above follow.

### 6. The read-only banner carries **Join**.

State 24's banner reads as an invitation, not a refusal, and holds the Join action. ADR-018 §10's *persistent and non-dismissible* is unchanged — it is now persistent *because it is the exit*, which is a stronger reason than the one it had.

**24 and 25 become one component with two settings**, not two independent screens:

| | Board 2.7 — join to edit | Board 2.8 — join to look |
|---|---|---|
| Visibility | WORKSPACE | PRIVATE |
| Who | Plain member, not joined | Workspace admin, not joined |
| Contents | Fully visible, not editable | Hidden |
| Roster | Visible | **Empty** — the server withholds it pre-join |
| Action | Join, in the banner | Join, the only action |
| Joining is recorded as | An ordinary join | Break-glass |

The roster row is not a design choice: `GetBoardDetail` returns `Members: []` for a non-joined requester on a PRIVATE board, by design. The gate can state a member *count* — the board tile carries one — but cannot name anybody.

> **So the banner part has one appearance, and it always carries Join.** Inventory §4.2 listed an *informational* banner as well. By this section's own Context §4, nobody could ever see one, and it is the banner rejected below (*Join on the board tile only*), so it was not drawn. The banner's surface, padding and wording are in **[ADR-028](adr-028-banner.md)** (2026-11-20).

---

## Options Considered

### Archive as one-way for Phase 1, amending US-12c

The zero-backend-change option, and the one the inventory listed first. Rejected on what it orphans: a working endpoint, an `ActivityActionUnarchived` verb, a `BOARD_UNARCHIVED` frame in the realtime contract, a handler in the frontend architecture's event router, and an SRS line explicitly placing unarchive in Phase 1. Amending a P1 story to match a **UI omission** inverts the project's own documented flow (PRD → stories → SRS → code). The cost it avoids is one query parameter.

### Dropping archive from Phase 1 entirely

The cleanest design outcome — one fewer state, one fewer list, no icon question at all. Rejected on churn in the wrong direction: PRD §5.1 names archive in the Phase-1 board list, the backend has shipped it, and removing a built feature to simplify a design file is the tail wagging the dog.

### Two-way archive with the archived list deferred to Phase 2

Keeps the Archive control, ships no exit, and records the gap honestly rather than amending the story. Rejected because it is the trapdoor the inventory names: a control whose consequence cannot be undone by any route the user can find. Honesty in a document does not reach the person pressing the button.

### A separate `/workspaces/:id/archived` route

Considered for the archived list. Rejected: it is the same grid with one predicate changed, so a route buys a bookmarkable URL at the cost of a route module, a loader, an auth path, a back-navigation decision and a second empty state. A filter on the grid inherits all of those from the grid.

### A full "Move column to…" picker, matching cards

One pattern for both draggable objects, and it reads as more consistent. Rejected on fit: the card picker exists because a column's contents are unbounded and a neighbour must be named from a list. A board's columns are few and visible, so the picker's list would be a longer way of saying *left* or *right*. It would also add a fifteenth overlay and therefore another row under §5 — cost in two places for no added reach.

### Column dragging on the phone strip

Rejected for ADR-019 §5's reason, applied to a different object. It requires disabling scroll-snap on drag start and restoring it on drop, on cancel, and on a drag interrupted by an incoming WebSocket update — the exact list §5 called "the hardest behaviour in the plan to verify" — to duplicate, worse, what two menu entries already do at every width.

### Every overlay becoming a full-height sheet below 640px

One rule, nothing to remember, and no per-overlay judgement while drawing. Rejected on the stacking case: the two overlays most often opened *from inside* a sheet are the confirmation and the pickers, and both would become a second sheet. It also spends 844px on a two-line question.

### Keeping ADR-019 §3 as written — only the card modal becomes a sheet

The minimal reading, and defensible as "no decision is better than a wrong one". Rejected because it is not actually neutral: it leaves a form in a centred dialog at 390px, where the on-screen keyboard covers roughly half of it, for every one of the ten form overlays. The absence of a rule is itself a rule, and a worse one.

### Join on the board tile only, leaving the read-only banner informational

The status quo. Rejected because the board is openable without joining by design — `CheckViewAccess` admits exactly this person — so "join before you open" describes a path the product does not enforce. Anyone who opens first is then told what they cannot do and given nowhere to go, on a screen ADR-019's Context fact 3 classifies as routine browsing.

### Join in the board `⋯` menu

Keeps the banner purely declarative. Rejected on ADR-018 §6's own rule: a function must be reachable without hunting, and the one action this person needs would sit behind a menu whose other entries are all unavailable to them.

---

## Out of scope (deferred, recorded)

- **Deleting a board.** There is no delete-board endpoint and no `BOARD_DELETED` frame; §3.3 of the SRS describes one that does not exist. Archive is the only removal in Phase 1. Recorded, not built.
- **Renaming a workspace.** No update endpoint exists, so a workspace's name and description are fixed at creation. Out of scope here; recorded in the PRD by this ADR's downstream edits.
- **Password reset.** No endpoint, and no email-sending capability anywhere in the product — workspace invites resolve an existing account by address and never send mail. Recorded as a Phase-2 deferral in the PRD rather than left as an unnoticed hole.
- **Profile editing.** `GET /user/profile` has no counterpart. The profile surface is the user menu; a profile *page* arrives with Phase-2 editing and avatar upload.
- **SC 4.1.3 — announcing changes that happen without user action.** Untouched. PRD §5.4 keeps it open and ADR-018 §8 still names the reduced-motion overlap.
- **Column width at fine pointer.** 280px remains ungrilled, as ADR-017, ADR-019 and ADR-018 all record.
- **Toast content.** Still scaffolding work, with one addition from this ADR: `ACCESS_REVOKED` carries four distinct reasons (`removed_from_board`, `board_made_private`, `removed_from_workspace`, and the silent voluntary case), so the toast needs four wordings, not one.

  > **The three spoken wordings are written; the fourth case stays silent.** **[ADR-030](adr-030-toast.md)** §5 (2026-11-20): *You were removed from “title”.* / *“title” is now private. Only its members can open it.* / *You were removed from the workspace “name”.* None names who did it, since the frame carries no actor. The toast's own design is in the same ADR.
- **Phase-2 deferrals** — comments, labels, attachments, public boards, refresh tokens, account deletion — untouched.

## Downstream edits this ADR requires

1. **Backend — `GetUserBoardsInWorkspace`.** An `archived` filter on the board list: a query parameter on the handler, through the usecase, to a parameterised predicate in `getUserBoardsInWorkspace`. The *only* backend change this ADR requires. **This reopens a backend CLAUDE.md records as complete**, which is stated here rather than discovered later.
2. **SRS §4.3 / UC-08** — document the `?archived=` parameter and its default.
3. **SRS §3.3** — correct *"Workspace and board deletion = hard delete"*: board deletion does not exist in Phase 1.
4. **SRS UC-11b** — correct *"workspace members not yet on the board"*: `GetWorkspaceInviteesForBoard` returns **every** workspace member with an `is_board_member` flag, which is what makes the "everyone is already on this board" empty state representable.
5. **PRD §5.2** — add password reset to the deferral table, with the reason (no email capability).
6. **PRD §5.1** — narrow *"basic profile"* to viewing; note that workspaces cannot be renamed after creation.
7. **US-12c** — add acceptance for where Unarchive is reached from.
8. **US-13d** — add acceptance for the non-drag move, matching US-15's picker.
9. **US-11b** — drop from P2; it is implemented and the inventory depends on it.
10. **`design-file-inventory.md`** — §6 replaced by the settled answers; §1.4, §1.5, §2, §3 and the appendix updated. It remains a derived document and still decides nothing.
11. **`CONTEXT.md`** — add **Archived**. Unlike "snap strip", "sheet" and "in flight", this *is* a domain term: a reversible board state the backend stores, logs with its own verbs and broadcasts, and the only removal a board has. Its absence from the glossary is what let two accepted documents disagree about it for a day.
12. **ADR-018** — no edit. Its *Corrections* table is corrected here by reference, per the project's never-edit-an-ADR rule. §9's icon decision stands unchanged (§2 above).
13. **ADR-017, ADR-019, ADR-020** — no change. ADR-019 §6 is extended, not amended; §3 is completed, not overturned.
14. **`CLAUDE.md` › `## Now`** — the inventory's five open questions are settled; the backend item in edit 1 is the one blocking thing before the archived list can be drawn.
15. **`docs/build-log.md`** — an entry when the work lands, not now.
