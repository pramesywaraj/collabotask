# Design File Inventory — What To Draw in Penpot

> **What this doc is:** the authoring checklist for the Penpot file `Collabotask App` — every screen, overlay, whole-screen board and variant set the app zone needs, and what sits on each one.
> **What this doc is *not*:** a decision record. It makes **no decisions about the product** — every row in §§0–6 is derived from an accepted ADR, a user story or the route tree, and cites its source. **One exception, stated so it is not mistaken for a derivation:** §7's drawing order is this document's own working plan. It follows from no ADR, and it is freely revisable — unlike everything above it.
>
> **Status:** Living document. **Owner:** pramesywaraj. **Date:** 2026-11-19.
> **Audience:** whoever authors the design file — today that's the solo dev; written so a future teammate can pick it up without re-reading five ADRs.
>
> **Starting a drawing session? Read [§7](#7-the-order-to-draw-in-and-what-is-already-drawn) first.** It holds the order to draw in, what is already in the file, and a query you can paste into Penpot to see the real state rather than this document's guess at it.
>
> **2026-11-19 — §6 is closed.** The five open questions were grilled against the running API rather than against the other ADRs, which turned three of them into facts and found four gaps §6 never asked about. Everything settled there is now recorded in **[ADR-021](adr/adr-021-design-inventory-corrections.md)**, and this document is again purely derived. §6 below keeps the answers for reference; nothing in it is still open.

---

## 0. Scope and sources

The **app zone** only — the React + Vite SPA ([ADR-011](adr/adr-011-frontend-framework.md)). The Astro landing page is explicitly outside this file: [ADR-020](adr/adr-020-design-tool-penpot.md) excludes it, and [ADR-017](adr/adr-017-ui-foundations.md) §3 leaves it free to carry more colour than a dense board. It needs its own, later effort.

Eleven documents fix everything here; this one only arranges it.

| Source | Fixes |
|---|---|
| [ADR-017](adr/adr-017-ui-foundations.md) | Tokens and values, fine density, contained columns, type scale, radius, elevation, colour modes |
| [ADR-019](adr/adr-019-responsive-touch-scope.md) | Two layouts (crossover 640px), touch drag, the coarse density scale |
| [ADR-018](adr/adr-018-state-vocabulary.md) | All 39 interaction states and the channel each one spends |
| [ADR-020](adr/adr-020-design-tool-penpot.md) | Penpot as the medium; token sets, theme groups, variants, which boards exist |
| [ADR-021](adr/adr-021-design-inventory-corrections.md) | Archive is two-way and has a list; column reorder without dragging; how overlays present below 640px; the read-only board's exit |
| [ADR-022](adr/adr-022-form-control-geometry.md) | Form-control geometry — control height, horizontal padding, the gap between stacked fields, form width |
| [ADR-023](adr/adr-023-spacing-between-objects.md) | Spacing between objects — the screen's side inset, the column gap, no vertical canvas padding, the section break; the rail height, the two border widths, the initials' type step |
| [ADR-024](adr/adr-024-control-and-field-internals.md) | Inside one field and one control — the label gap, icon size and gap, the focus ring as a shadow, the input's fill, pressed on a filled button |
| [ADR-025](adr/adr-025-identity-hues-as-tokens.md) | The identity hues as design-file tokens (`subtle` / `strong`, flipping with Theme); `size.avatar`; the Avatar's ten appearances, including the no-name glyph |
| [ADR-026](adr/adr-026-count-chip.md) | The count chip — its own fill (`color.chip`, not `color.border`), its 20px height and minimum width, its 6px side padding; one appearance that holds any text |
| [ADR-027](adr/adr-027-column-header.md) | The inside of the column header — its 12px inset (the card titles' left edge), `spacing.inline-gap` 8 between title, count and `⋯`, the `⋯` as Phosphor `DotsThree` shown at rest; one appearance, no axis |

Screen boundaries come from the route tree in [frontend-architecture.md](frontend-architecture.md) §*Route tree*; screen *contents* come from [001-user-stories.md](../product/001-user-stories.md).

**The density rule, because it governs the whole file.** Fine and Coarse are a *theme group*, not a second set of frames ([ADR-020](adr/adr-020-design-tool-penpot.md) §2). Draw every board once at Fine and let the flip carry it to Coarse. Layout is the exception: no mode flip turns a fixed 280px column into a viewport-relative snap strip, so the phone board is drawn separately (§2.15).

---

## 1. The six screens

Six routes, one of which is an overlay over another. Every state named in a row is an [ADR-018](adr/adr-018-state-vocabulary.md) state number.

### 1.1 `/login`

| On it | Source |
|---|---|
| Heading at 28/36 — login and register own the only use of that step | ADR-017 §6 |
| Email and password fields with the `--input` stone-500 `#78716C` boundary, deliberately heavier than most shipped systems | ADR-017 §3.4 |
| Plum primary submit; inline plum link across to register | ADR-017 §3.1 |
| Variants: field invalid (36), busy (17 — `--primary-subtle` fill, spinner, label becomes the progressive verb), form not satisfied (16 — `--muted` fill), focus-visible (14 — 2px `--ring`, 2px offset) | ADR-018 §5, §6 |
| **No "forgot password" link.** There is no reset endpoint, and the product cannot send email at all — invites resolve an address that already has an account. The absence is now a recorded deferral rather than an omission, so **don't draw a link that would have nowhere to go** | PRD §5.2; ADR-021 *Out of scope* |

No separate narrow board. A two-field form reflows at any width, and the density flip supplies the 44px targets.

### 1.2 `/register`

Same furniture plus a name field. Two additions worth drawing:

- A **duplicate-email conflict renders as field-invalid, not as a toast** — it points at a field, so it belongs on the field.
- **The password minimum is 8 characters and it is the only limit a real person will hit here.** Say it before they submit, not after. (Name 1–255 and a valid email address are the other two; neither needs stating.)

### 1.3 `/workspaces`

| On it | Source |
|---|---|
| App bar, 56px — product mark, user menu | ADR-017 §1 derivation |
| Workspace tiles: name, **my role**, member count, active board count | US-05 |
| "Create workspace" primary action | US-03 |
| Variants: cold load (19 — static skeleton, no shimmer), empty (21, flavour 1), load failed (22) | ADR-018 §7 |

**Creating a workspace puts you inside it.** The panel closes and the app navigates to the new workspace — it does not return you to a list with one more tile on it. Same rule one level down (§1.4), and together they are the whole of the product's onboarding: PRD §4.1's "about a minute" is a consequence of these two navigations, not of a welcome tour. There is no welcome tour (ADR-021 *Downstream edits*; PRD §4.1).

The **empty state here is the first screen a new account ever sees**, so it carries the weight a welcome screen otherwise would: one sentence and one button.

### 1.4 `/workspaces/:wsId`

One route carries both the boards and the team. **Composition settled by [ADR-021](adr/adr-021-design-inventory-corrections.md): the board grid fills the screen; the team opens over it from a member-count button.** Boards are why you come here almost every time; the team is managed once, at the start. That keeps this as **one** board in the file rather than two, and the team list becomes an overlay (§3) that inherits §3's presentation rule for free.

| On it | Source |
|---|---|
| Workspace header: name, **description**, my role, and a **"N members" button** that opens the team panel | US-05b |
| An **Active / Archived** switch above the board grid. Not a route — the same grid, one filter | US-12c; ADR-021 §1 |
| Board tiles: title, **identity swatch** from `board.background_color`, my role, joined / can-join, member count, card count | ADR-017 §3.3; US-08 |
| A **filled** Phosphor lock on private boards — the icon channel's only Phase-1 meaning, and the only place fill is permitted | ADR-018 §9 |
| A Join affordance on a board I may join but haven't | US-09 |
| Create board; delete workspace behind a destructive confirmation (38) | US-07, US-06d |
| Variants: cold load (19), empty boards (21, flavour 2), **empty archived list (21, flavour 3)**, load failed (22), and the non-member case which is refused rather than rendered | ADR-018 §7, §10 |

**Archived boards.** Under *Active* they are absent — not greyed, not iconned (US-08). Under *Archived* every tile is archived, so the archive icon would mark nothing the filter has not already said: **ADR-018 §9's single-meaning icon channel is untouched by this filter** (ADR-021 §2). An archived tile's only action is **Unarchive**; it cannot be opened, and that is deliberate — an archived board can be found, never browsed.

**Delete workspace is owner-only, so for an admin who is not the owner the control is _absent_, not refused** (US-06d; ADR-018 §5).

**Creating a board puts you on it**, columns and all — same rule as §1.3, and the second half of PRD §4.1's minute.

**The swatch must fall back on read.** It is constrained to the eight identity hues, but `background_color` validates only `min=4,max=8` and is not a hex validator, so the string `"pink"` reaches the client (ADR-017 §3.3). There is a second field of the same species on the card itself — see §1.5.

**The workspace's name and description cannot be edited** — there is no update endpoint, so what was typed at creation is what this header shows forever (PRD §5.1). Don't draw an edit affordance on them.

### 1.5 `/workspaces/:wsId/boards/:boardId` — the board

The product. Two layouts, crossing over at 640px (ADR-019 §2); from 640px up, ADR-017 §1–§2 apply unchanged.

| Region | Contents | Fine → Coarse |
|---|---|---|
| App bar, 56px | Board name 20/28; **presence cluster**; board `⋯`; add-column action | — |
| Board accent | A 4px identity rail at the top, from the eight hues. **Not** a canvas colour — the canvas is always neutral `--board-canvas` | — |
| Canvas | `--board-canvas` `#E7E5E4`, horizontal scroll. 16px side inset (`spacing.screen-x`) and **no top or bottom padding**: the column panels run from under the app bar to the bottom of the screen (ADR-023 §2.3) | column gap 8 → 12 (`spacing.column-gap`) |
| Column panel | Own `--board-column` fill, 8px radius, 8px padding, 280px wide | 280px → `min(100vw − 72px, 420px)` below 640 |
| Column header | Shares the panel fill; `border-bottom: 1px solid var(--border)` is **load-bearing** because columns are unbounded and cards scroll under it. Title, count chip, `⋯` (rename, **move left**, **move right**, delete). The `⋯` is **shown at rest at every pointer**, unlike the card's ([ADR-027](adr/adr-027-column-header.md) §4), and the title starts where the card titles start | 36px → 44px; `⋯` area 24 → 44 |
| Card | 14/20 title clamped to **2 lines** (mandatory — titles accept 500 characters), 1px `--border`, **no resting shadow**, conditional meta row, `⋯` top-right | padding 8/12 → 12/12; avatar 20 → 24; meta row 24 → 28; gap 8 → 12; title right padding 24 → 32 |
| Meta row | Rendered **only** when the card has an assignee or a due date. Initials avatar, due date in `tabular-nums`, amber when due soon, rose when overdue | — |
| Avatars, everywhere | **Always initials in Phase 1.** Uploads are deferred (PRD §5.2), so `avatar_url` is never set — don't draw a photo variant | PRD §5.2 |
| Column footer | Add card | — |

Per-card height is `8 + lines×20 + (24 if meta) + 8` at Fine and `12 + lines×20 + (28 if meta) + 12` at Coarse. The smallest card is 36px at Fine and exactly 44px at Coarse, so cards clear the target floor by construction rather than by special-casing.

**The initials avatar needs a fallback, for the same reason the colour swatch does.** An assigned card always carries an assignee object, but its `name` is filled only if the server's user lookup succeeded — so `{"id": "…", "name": "", "avatar_url": null}` is representable. Initials derived from an empty string produce an empty circle. Draw the fallback (a neutral `--muted` circle, or the Phosphor user glyph) rather than discovering it as a blank on screen. **Settled by [ADR-025](adr/adr-025-identity-hues-as-tokens.md) §4: the glyph, on a muted circle** — an empty muted circle is 1.09:1 against a white card and invisible on a dark one. Same species as `background_color: "pink"` in §1.4: a field the API's validation does not actually guarantee.

The card's `⋯` is **hover-revealed at fine pointer and persistent at coarse** (ADR-019 §6), and it does not go in the meta row — that row is conditional, and putting `⋯` there would force it onto every card.

**There is no board-level "add card".** A card belongs to a column, so a top-bar Add card would have to ask which one first. The column footer already answers that by standing in the column. Removed from this table on 2026-11-19.

#### The board `⋯` menu

Never enumerated before 2026-11-19, and it is the only home five P1 stories have. Entries the viewer may not use are **absent, not greyed** (ADR-018 §5), so this menu has three shapes:

| Entry | Opens | Who sees it | Source |
|---|---|---|---|
| Board settings… | Overlay (§3) — title, description, colour, visibility | Board owner or workspace admin | US-12b |
| Members… | Overlay (§3) — add, remove, transfer ownership | Anyone who can read the board; add/remove/transfer inside it are owner-or-admin only | US-11, US-12d, US-12e |
| Archive board | Destructive-shaped confirmation (38), then **navigates out** | Board owner or workspace admin | US-12c |
| Leave board | Confirmation (38) | Anyone on the board **except its owner** | US-10 |

**The owner sees no Leave.** US-10 refuses it until ownership is handed over, and ADR-018 §5 says a refused action is absent rather than present-and-blocked. The consequence has to be drawn for: Members… must make "You own this board" and **Transfer ownership** visible enough that the owner finds the route, because the menu gives them no hint that leaving exists at all.

**Archiving navigates the person who did it back to the workspace**, with a confirmation message carrying **Undo** — the same endpoint with the flag flipped. Everyone else holding the board open gets the gone state (board 2.12). Without this, the actor would land on the screen that says their board no longer exists, which reads as a failure rather than as the thing they just asked for (ADR-021 §1).

#### Column drag has two homes, cards have one

**Columns are reordered by dragging only at 640px and up.** Below that, the `⋯` menu's Move left / Move right is the only route — dragging a column along the axis the snap strip scrolls on is the problem ADR-019 §5 froze the board to avoid (ADR-021 §4). Card dragging is unaffected at every width.

At either end of the board the inapplicable entry is **absent**: the leftmost column has no Move left.

### 1.6 `.../cards/:cardId` — card detail

A URL-based overlay: the board renders underneath and browser back closes it. A drawer at 640px and above, a **full-height sheet** below (ADR-019 §3).

| On it | Source |
|---|---|
| Title at 20/28, editable inline (37 — the `--input` boundary doing the editing job) | ADR-017 §6; ADR-018 §10 |
| Description at 16/24 — the **only** use of that step in the app | ADR-017 §6 |
| Assignee picker, limited to board members, **and clearable** — US-16 requires leaving a card unassigned | US-16; assignment-is-participation |
| Due-date picker; Phosphor calendar, Regular, 16px | ADR-018 §9 |
| **"Move to…"** — required here at every width | ADR-019 §6 |
| Delete, behind a destructive confirmation (38) | US-17 |
| Created / updated timestamps, `tabular-nums` | ADR-017 §6 |
| Variant: remote deletion of the open card (32) — the drawer closes with a toast | ADR-018 §10 |

**Below 640px this drawer is the thing everything else opens on top of**, which is what shapes §3's presentation rule. The date picker, the assignee picker and the delete confirmation are all reached from inside it, so none of them may become a second full-height sheet: the two pickers **expand in place inside the drawer**, and the confirmation stays a centred dialog at every width (ADR-021 §5).

---

## 2. Whole-screen boards

[ADR-020](adr/adr-020-design-tool-penpot.md) §5 fixes this list: a board only where **the whole screen differs**. Fifteen, numbered as they'll be named in the file.

| # | Board | Why it is whole-screen |
|---|---|---|
| 2.1 | Login | §1.1 |
| 2.2 | Register | §1.2 |
| 2.3 | Workspaces | §1.3 |
| 2.4 | Workspace detail | §1.4 |
| 2.5 | Board — at rest | §1.5. The reference frame everything else varies from |
| 2.6 | Card drawer over the board | §1.6, at ≥640 |
| 2.7 | **Read-only board** (24) — *join to edit* | Mutating controls are **absent**, plus a persistent, non-dismissible banner **that carries the Join action**. On touch the banner is the only remaining signal — a read-only user presses and holds and nothing happens. Routine browsing, not an edge case |
| 2.8 | **Join-required gate** (25) — *join to look* | A private board an admin hasn't joined: they see it exists, not its contents. One Join action, and joining is recorded as break-glass |
| 2.9 | **Cold load** (19) | Static skeleton blocks at the card and column geometry, `--muted` fill, **no shimmer** — shimmer is movement, and movement means a teammate acted |
| 2.10 | **Empty board** (21) | No columns. New boards seed To Do / In Progress / Done, so this is the deleted-them-all case |
| 2.11 | **Load failed** (22) | Message plus a retry control |
| 2.12 | **Gone** (23) | **Archived while you had it open**, or its workspace deleted while you had it open. *Not* "deleted" — there is no delete-a-board action in Phase 1 and no `BOARD_DELETED` event. *Not* "access lost" either: that is state 26, which navigates you out with a toast rather than rendering a screen |
| 2.13 | **Session expired** (39) | A 7-day JWT dying mid-session, drawn *before* the hard navigation to `/login`. Without it, expiry looks exactly like a crash |
| 2.14 | **Drag** | Four states at once, which is why it earns a board: source at 60% opacity (6), overlay with `--shadow-dragged` and `scale(1.02)` (7), a 2px dashed `--drop-placeholder` (8), and the column **body** tinted `--drag-over` with the header left untinted (9). **Draw two placeholders** — see the note below |
| 2.15 | **Phone board, 390×844** | Density is a variable; layout is a breakpoint. Also the only place ADR-019 §7's "6 cards visible" gets checked by eye — the one number in its table resting on an assumption |

**On 2.14, the two things easiest to draw wrong.** The tint covers the column body only: on either tint the header's `--border` rule sits at about 1.19:1 and vanishes, exactly when it is most needed. And the placeholder is `#57534E`, not `--input` — measured against the tint rather than against white, `--input` thins from 4.80:1 to 3.22:1.

**And a third, added 2026-11-19: 2.14 carries _two_ placeholders, not one.** Cards and columns are both draggable, and their drop placeholders are different shapes drawn from the same token:

| Dragging | Placeholder | Where it exists |
|---|---|---|
| A card | 2px dashed `--drop-placeholder` at **card** geometry | Every width |
| A column | 2px dashed `--drop-placeholder` at **column-panel** geometry — full panel height, panel width | **640px and up only** |

Below 640px columns are reordered from the `⋯` menu instead, so the column placeholder has no phone state to draw (ADR-021 §4). Card drag is unchanged at every width.

**On 2.15**, the strip is `min(100vw − 72px, 420px)`, and the 72px reservation is not arbitrary: it leaves the next column peeking by exactly **44px**, one tap target, which is what makes it a legal drop target. Below 640px the board does **not** scroll during a drag — valid targets are the current column and that peek.

**On 2.7 and 2.8 — draw them as one component with two settings, not two screens.** They are the same question asked at two visibility levels, and the differences are mechanical:

| | 2.7 — join to edit | 2.8 — join to look |
|---|---|---|
| Board visibility | WORKSPACE | PRIVATE |
| Who is standing here | A plain workspace member who hasn't joined | A workspace admin who hasn't joined |
| Board contents | Fully visible, nothing editable | Hidden |
| Roster | Visible | **Empty** — the server withholds it until you join |
| The action | Join, in the banner | Join, the only thing on screen |
| Recorded as | An ordinary join | Break-glass |

The empty roster on 2.8 is not a drawing choice: `GetBoardDetail` deliberately returns no members to a non-joined requester on a private board. The gate can state a member *count* — the board tile on §1.4 has one — but cannot name anybody. And 2.7 has exactly one possible occupant, who is always allowed to join, which is why its banner is an invitation rather than a refusal (ADR-021 §6).

---

## 3. Overlays

Drawn as component variant sets rather than boards, and placed on one or two holding boards. The destructive confirmation is **one component with several copies of its copy**, not several components.

| Overlay | Notable content | Source |
|---|---|---|
| **Workspace team panel** | The team list: name, email, role, promote / demote, remove, leave. Opened from the "N members" button in §1.4's header | US-05b, US-06, US-06b, US-06c |
| Create workspace | Name (**2+ characters**) and an optional description. **Both are permanent** — no rename exists, so this panel is the only chance to get them right. Worth a word of warning in the copy | US-03 |
| Invite people | Email batch plus **per-row results** — the mixed batch is the normal case, not the error case. Exactly two failure messages to write: the address has no account, and they're already in the workspace | US-04 |
| Create board | Title (**3+ characters** — the only minimum here that will trip anyone), visibility, colour picker restricted to the eight hues | US-07 |
| Board settings | Title, description, colour, visibility | US-12b |
| Board members | Add, remove (**which unassigns their cards on this board**), transfer ownership, leave | US-11, US-12d, US-12e, US-10 |
| Add people — candidate list | **Settled by fact, not by choice: the endpoint ships.** It returns *every* workspace member flagged as already-on-the-board, so draw the whole roster with the already-added rows marked and unselectable | US-11b |
| Transfer ownership | Target must **already** be a board member; a non-member is refused rather than auto-added | US-12e |
| **Move card to…** | A **column** plus a **neighbour** — top of column, after «card», bottom of column — because position is a float, not an index. Three entry points: card drawer, card `⋯`, and `M` on a focused card | ADR-019 §6 |
| Destructive confirmation (38) | Rose confirm. Its unsatisfied state is drawn in `--muted` tokens, **never** in transparency | ADR-018 §5 |
| **Board `⋯`** | Board settings, Members, Archive board, Leave board — enumerated in §1.5, and shaped by who is looking | US-10, US-11, US-12b, US-12c |
| Column `⋯` and inline rename (37) | Rename, **move left, move right**, delete | US-13b, US-13c, US-13d |
| Card `⋯` | Open, Move, Delete | US-15, US-17 |
| User menu | Name, email, initials — and logout. **This is the profile surface**; nothing in it is editable, because no endpoint changes any of it | US-02b; ADR-021 |
| Presence overflow sheet | The `+N` chip's full list, below 640px | ADR-019 §8 |
| Toast | Carries several states' messages; its own design is deferred. One content requirement that is *not* deferred: **access-revoked has four distinct reasons** and needs four wordings — removed from this board, board made private, removed from the workspace, and the silent voluntary case which shows nothing | ADR-018 *Out of scope* |
| Date picker, assignee picker | Phosphor Regular 16px throughout. **Not separate layers** — see the rule below | ADR-018 §9 |

**Two batch behaviours that look alike and are not.** Inviting to a *workspace* is per-row: unknown addresses and existing members are reported individually and the valid invites still land. Adding people to a *board* is **all-or-nothing** — one person already on it fails the entire batch with a single error. So the board panel must **not** be drawn with per-row result rows. It doesn't need them: the candidate list already flags who's on the board, which is what prevents the failure in the first place.

### Presentation below 640px — four kinds, one rule each

ADR-019 §3 committed only the card modal. [ADR-021](adr/adr-021-design-inventory-corrections.md) §5 settles the rest in one pass, because the deciding fact is structural rather than per-overlay: **several of these open on top of something that is already a full-height sheet.**

| Kind | Members | Below 640px | 640px and up |
|---|---|---|---|
| **Forms and lists** | Card drawer, create workspace, create board, board settings, board members, add people, transfer ownership, Move card to…, the workspace team panel, presence overflow | **Full-height sheet** | Centred dialog or drawer |
| **Menus** | Board `⋯`, column `⋯`, card `⋯`, user menu | **Bottom action sheet** | Anchored popover |
| **Confirmations** (38) | Delete column, delete card, delete workspace, archive board, leave board | **Centred dialog — unchanged at every width** | Centred dialog |
| **Inline pickers** | Due date, assignee | **Expand in place inside the card drawer** | Expand in place, or an anchored popover |

The last two rows are the rule's whole point. Delete is reached from inside the card drawer; so is the date picker. Promoting either to a second sheet puts two full-height layers on a phone, which is the state in which nobody can see what they're about to delete.

---

## 4. Not screens, but drawn before them

### 4.1 Foundations board

| Holds |
|---|
| The token catalogue, as sets: `base`, `theme-light`, `theme-dark`, `density-fine`, `density-coarse` (ADR-020 §2) |
| The eight identity hues in **both role assignments** — pale fill with deep ink at Light, deep fill with pale ink at Dark, and the rail flipping to pale at Dark because every deep value fails 3:1 against a dark card |
| The elevation ladder: canvas → column → card, in both modes, with dark as a *compressed* version of the same ladder, never an inverted one |
| The radius ladder: 4px chips, 6px buttons and inputs, 8px cards, panels and overlays |
| The five type steps, and the **14px-against-16px specimen** ADR-017 names as the first thing on the page — it exists so the default can be revisited by looking rather than re-argued |

### 4.2 Variant containers

Per [ADR-020](adr/adr-020-design-tool-penpot.md) §4, every state is a variant. The card's `State` axis is `Rest · Hover · Pressed · InFlight · DragSource` — five variants across two densities and two themes is **one** component, not twenty frames.

| Component | Axis |
|---|---|
| Card | the five above |
| Button | rest, hover, pressed, focus-visible, busy, not-satisfied |
| Input | rest, focus, invalid, inline-editing |
| Avatar | eight hues, **plus Muted** (initials, for the presence cluster's reconnecting state) **and Unknown** (the no-name fallback — initials from an empty string render an empty circle, §1.5). The two role assignments are not an axis: the hues are tokens that follow the Theme switch ([ADR-025](adr/adr-025-identity-hues-as-tokens.md)) |
| Column header | — . **One appearance.** Drag-over leaves the header **untinted** (§2.14), so it changes nothing on the part. The part's board shows the header twice instead, above an untinted body and above a `color.drag-over` one, so the rule can be seen against the tint ([ADR-027](adr/adr-027-column-header.md) §5) |
| Banner | informational, **and with an action** — the read-only banner carries Join (§2, board 2.7) |
| Count chip | — . One appearance that holds any text: `12` in a column header, `12 cards` and `3 members` on a board tile, because two bare numbers on one tile cannot be told apart ([ADR-026](adr/adr-026-count-chip.md) §4) |
| Toast, skeleton block | — |
| Presence cluster | live, **reconnecting** — hues drop to flat `--muted` with a "Reconnecting…" label beside them, spending no hue |

Hover, pressed and the border ladder are three consecutive stone steps on the **edge**, never a fill: shadcn hovers to `--accent`, which is `--board-column`'s own value, so a card hovering to accent dissolves into the panel behind it.

### 4.3 What cannot be drawn

**Motion.** Penpot has no duration or easing token type, and the channel's meaning — "a teammate did this" — is a multi-client behaviour no static file can show. It stays code-only: 120ms chrome, 200ms board objects, 260ms containers, 110ms reduced-motion cross-fade.

**Equal-width digits.** ADR-017 §6 puts `tabular-nums` on dates and counts. Penpot has no token type for it and no setting for a font's alternate figures, so the file shows Poppins's default digits, which are proportional: a `1` is 4px wide and an `8` is 8px ([ADR-026](adr/adr-026-count-chip.md) §5). A count or a date with a `1`, `2` or `7` in it can therefore draw slightly narrower in the file than in the app.

**Who gets which hue.** In code the palette is a lookup indexed by `hash(id) % 8`, living in a TypeScript constant with both role assignments (ADR-017 §9). The file cannot show that assignment; each drawn avatar simply picks a hue. The colours themselves *are* tokens in the file since [ADR-025](adr/adr-025-identity-hues-as-tokens.md), so they follow the Theme switch.

---

## 5. What the file totals

| | Count |
|---|---|
| Whole-screen boards | **15** (§2) |
| Overlay variant sets | ~**16** (§3) |
| Foundations board | **1** |
| Component variant containers | ~**10** (§4.2) |

**The board count did not move, and that was the point of §6.1.** Putting the team behind a button rather than in tabs keeps the workspace screen as one board; the Active/Archived switch is a filter on the same grid rather than a route; and 2.7 and 2.8 collapse into one component. The one addition is on the overlay side — the workspace team panel, which inherits §3's presentation rule for free.

Verification is a **predicate, not an eyeball** (ADR-020 §6). The check itself, as runnable code, is in **§7.5** — deliberately not restated here, so the two cannot drift.

---

## 6. The questions that were open — all five settled 2026-11-19

Settled by grilling the inventory against the running API rather than against the other ADRs, which turned three of these into facts. All five are recorded in **[ADR-021](adr/adr-021-design-inventory-corrections.md)**; kept here so the reasoning is next to the rows it changed.

### 6.1 Workspace detail composition — **settled: boards fill the screen, team opens over it**

The board grid gets the whole screen with the Active/Archived switch above it; a **"N members" button** in the header opens the team as a panel. Chosen on frequency: you come to this screen for a board nearly every visit and manage the team once, at the start. Tabs would have made Team a peer of Boards and cost a second board in the file; a permanent side panel costs width on every visit and still needs a separate phone answer. **The board count did not move** (§5).

### 6.2 There is no profile surface — **settled: the user menu, and nothing is editable**

`GET /user/profile` has no counterpart — no endpoint changes a name, an email or an avatar, and uploads are Phase 2. A profile *page* would therefore be a page of three read-only facts. The user menu carries them, which is all US-02b asks for. A real profile page arrives with Phase-2 editing; this is a deferral, not a rejection (PRD §5.1).

### 6.3 Archive has no exit — **settled: it had one all along, and now it has a list**

The premise was wrong, and the SRS had said so since it was written. ADR-018 struck the archive icon on the finding that archived boards 404 everywhere *"and so is Unarchive"* — but `SetArchived` is the one board path that **does not** go through `BoardAccessChecker`. It fetches the board directly and runs its own permission check, so unarchiving an archived board **succeeds today**, writes an `unarchived` activity row, and broadcasts a `BOARD_UNARCHIVED` frame the realtime contract already defines.

What was missing was **discovery**: the board list hardcodes `AND b.is_archived = FALSE` with no parameter, so no list contains an archived board and there is nothing to press Unarchive on.

So the resolution is the smallest of the three the old §6.3 offered: **keep archive two-way, add an `archived` parameter to the board list, and draw an Active / Archived switch on the grid.** US-12c is honoured as written rather than amended to match a UI omission.

**It does not reopen the archive icon.** In the Active list no archived board appears; in the Archived list every row is archived, so the icon marks nothing the filter has not already said. ADR-018 §9's single-meaning icon channel stands — only its stated reason changes from *unreachable* to *redundant*.

**This is the one item here with a backend cost**, and it reopens a backend `CLAUDE.md` records as complete. One query parameter, one predicate.

### 6.4 One sheet rule, or fourteen decisions — **settled: four kinds, one rule each**

See the table in §3. The deciding fact is structural rather than per-overlay: the confirmation and the two pickers are routinely opened from *inside* the card drawer, which is already full-height on a phone. So they cannot be sheets, and everything else can.

### 6.5 The five empty flavours are not enumerated — **settled: there are six, and one on the old list was impossible**

| # | Surface | Where |
|---|---|---|
| 1 | No workspaces | Board 2.3 — the first screen a new account ever sees |
| 2 | No boards in this workspace | Board 2.4, Active |
| 3 | **Nothing archived** | Board 2.4, Archived — new, from §6.3 |
| 4 | No columns on a board | Board 2.10 |
| 5 | No cards in a column | In-column surface |
| 6 | Everyone in this workspace is already on this board | The add-people panel (§3) |

**One was struck: "empty team".** §1.4 used to list it, and it can never happen — you are always a member of the workspace you are looking at, so that list always contains at least you.

**Six is why ADR-018's count is now wrong**, and harmlessly so: it said *"empty (five flavours)"* and deferred the copy as not a vocabulary decision. The copy is still not a vocabulary decision; there is just one more line of it to write.

---

## 7. The order to draw in, and what is already drawn

**Why this section exists.** This file is about forty separate pieces of work spread over many sittings, and **Penpot is not stored in git** — so nothing in this repository can tell you what has actually been drawn. That gap has already caused one error: `CLAUDE.md` described the design file as untouched while the full set of colour and spacing values, and the whole Card, were already sitting in it.

Two habits prevent a repeat:

1. **Draw in the order in §7.3.** It is not a preference. Each stage uses what the stage before it produced.
2. **Ask Penpot what exists; do not trust a written list** (§7.5). The checklist in §7.6 is a plan, not a record.

### 7.1 The five words this section leans on

Everything else in this document is product language. These five are tool language, so they are defined once, here.

| Word | What it means, plainly |
|---|---|
| **Token** | A named value — `color.card`, `spacing.card-padding-x` — used in place of typing a colour code or a number. Change the token once and everything using it changes. |
| **Token set** | A bundle of tokens switched on or off together. There are five: one shared, one each for light and dark, one each for the two spacing scales. |
| **Theme** | A saved choice of which sets are switched on. Four exist, in two independent groups: Light / Dark, and Fine / Coarse. |
| **Variant container** | One component holding all of a thing's appearances. The Card holds five — resting, hovered, pressed, `InFlight` (the moment after you act, before the server answers), and being dragged — so it is **one** item in the file instead of five near-identical copies. §7.3 calls these **parts**, for short. |
| **Bound** | A shape's colour or spacing is *bound* when it points at a token instead of holding a typed-in value. An unbound value looks identical on screen and silently ignores the Light/Dark and Fine/Coarse switches. Finding unbound values is the entire job of §7.5's second check. |

### 7.2 What is already drawn

Read directly from the live file on 2026-11-20, not assumed.

| Thing | State |
|---|---|
| The five token sets | **Done** — `base` 30, `theme-light` 52, `theme-dark` 52, `density-fine` 12, `density-coarse` 12. **158 tokens**: 103 derived from ADR-017 §9 in full, plus the four form-control values from [ADR-022](adr/adr-022-form-control-geometry.md), the six spacing and border values from [ADR-023](adr/adr-023-spacing-between-objects.md) (the column gap written once per density set), the four from [ADR-024](adr/adr-024-control-and-field-internals.md) (the focus ring written once per theme set), the 16 identity tokens from [ADR-025](adr/adr-025-identity-hues-as-tokens.md) (written once per theme set), the three count-chip tokens from [ADR-026](adr/adr-026-count-chip.md) (`color.chip` written once per theme set), and `spacing.inline-gap` from [ADR-027](adr/adr-027-column-header.md). ADR-025 also renamed `size.card-avatar` to `size.avatar`. Every addition was flip-verified on 2026-11-20, and Stage 2 has tested the catalogue against a real screen |
| The four themes, in two groups | **Done** — `Theme`: Light, Dark. `Density`: Fine, Coarse. |
| Card variant container | **Done**, on page `Components` — all five appearances, with its fill, outline, four corners and four paddings all bound. **2026-11-20:** outline width bound to `border.width`, and the avatar corrected. It had been bound to `size.title-reserve` for width (24 → 32) with a typed 20px height, which made an oval; it is now `size.card-avatar` on both axes (20 → 24). The binding check cannot see sizes, which is why it never caught this. **Later on 2026-11-20** the placeholder was replaced by a real Avatar instance in all five variants ([ADR-025](adr/adr-025-identity-hues-as-tokens.md) §5); heights unchanged at 80 / 92 |
| Input variant container | **Done** 2026-11-20, on page `Components`, board `Input — states`. Four appearances: `Rest` (placeholder), `Focus` (the ring), `Invalid` (destructive edge and a message below), `InlineEditing` (the 20/28 title in the same box, no label). A field is the whole field: label, box and message, so a screen needs nothing loose around it |
| Button variant container | **Done** 2026-11-20, board `Button — states`. Six appearances: `Rest`, `Hover`, `Pressed` (the same fill as `Hover`, by decision: ADR-024 §3), `FocusVisible`, `Busy` (Phosphor `CircleNotch` at 16 and *Logging in…*), `NotSatisfied`. **Primary only**: other kinds become a second axis when a screen needs one |
| Avatar variant container | **Done** 2026-11-20, board `Avatar — states`. One axis, `Hue`, ten appearances: the eight hues (fill `color.identity.<hue>.subtle`, initials `.strong`), `Muted` (initials on `color.muted`), `Unknown` (Phosphor `User` on `color.muted`). Sized by `size.avatar` with an ellipse behind, so no radius to bind. Flips with Theme by itself |
| Count chip | **Done** 2026-11-20, board `Count chip — states`. **One component, no axis**: fill `color.chip`, height and minimum width `size.chip-height`, side padding `spacing.chip-padding-x`, corners `radius.badge`, text `type.meta` in `color.muted-foreground`; its width follows the text. Shown with `3`, `12` and `128` on a column-coloured surface and `12 cards` and `3 members` on a tile-coloured one. **Penpot does not draw the minimum width** on a board that follows its content, so a `1` chip draws 16 wide here and 20 in the app ([ADR-026](adr/adr-026-count-chip.md) *Operational note*) |
| Column header | **Done** 2026-11-20, board `Column header — states`. **One component, no axis**: height `size.column-header`, fill `color.board-column`, a bottom-only rule (`color.border`, `border.width` on `strokeWidthBottom`), side padding `spacing.card-padding-x`, gap `spacing.inline-gap`. Inside: `Lead` (title in `type.body` and `color.foreground`, then a `Count chip` instance) and `Menu` (`size.target-min`, centring Phosphor `DotsThree` at `size.icon` in `color.muted-foreground`). Shown as `In Progress` / `12` on two bound column panels, one above an untinted body and one above a `color.drag-over` body. **No export shows the rule**: Penpot's exporter drops a stroke whose sides differ, so check it on the canvas ([ADR-027](adr/adr-027-column-header.md) *Operational note*) |
| The other four parts | Not started |
| Foundations board | **Done** 2026-11-20, on page `Foundations`. Six sections, specimen first: the 14/16 specimen, type, the elevation ladder (real Card instances on a bound canvas and columns), radius, the identity hues in both assignments, and the catalogue as three tables generated from the sets' stored values. The binding check reports only the 48 `Hue /` shapes and the five Card `⋯` glyphs, both expected. It was switched to Dark and to Coarse and looked at, then restored to Light + Fine. **Updated 2026-11-20 with ADR-025:** the catalogue tables were regenerated to 153 rows. That also adds the five ADR-024 tokens, which Stage 2 had left off the board. The identity notes now say the hues are tokens; this section's two side-by-side pairings stay typed in. **Updated again 2026-11-20 with ADR-026:** three rows added, so the catalogue read 157. **And with ADR-027:** one row added to the `base` table, so it reads 158. Its two loose column headers in the elevation ladder (a title only) predate the Column header part and have not yet been replaced by instances |
| The 15 whole-screen boards | **2.1 Login done** 2026-11-20, on page `Page 1`, built from Input and Button instances plus its heading and the register line. Drawn as it first appears: empty fields, the email placeholder only, the button `NotSatisfied`. The binding check returns nothing, and all four Theme × Density combinations were looked at. The other fourteen are not started |
| The ~16 overlays | Not started |

The foundation is laid, six parts stand on it, and one screen has been built from them end to end. Everything below picks up from there.

### 7.3 The order

**Stage 1 — the foundations board** (§4.1).

Before any component, for a reason that is easy to miss: it carries the **14px-against-16px specimen**, which [ADR-017](adr/adr-017-ui-foundations.md) puts first on the page *so the default text size can be reconsidered by looking at it rather than by arguing about it*. If that size is ever going to move, it has to move before ten components are built on top of it.

#### What question 4 returned at Stage 1 — and what closed it

Stage 2's question 4 (below) was asked early, on 2026-11-20, of every value this board needs, before anything was drawn. Every colour, radius and type step was there. What was missing was a whole family: **the space between objects**. Every spacing token in the catalogue sat either inside one object (card, column, control) or between items within one (cards in a column, fields in a form). This board is the first thing in the file to put objects side by side on a page, and its elevation ladder needs the same canvas inset and column gap that board 2.5 does.

**Settled by [ADR-023](adr/adr-023-spacing-between-objects.md)** (2026-11-20):
- `spacing.screen-x` 16
- `spacing.column-gap` 8 / 12
- no top or bottom canvas padding
- `spacing.section-gap` 24
- three values missing for the same reason: `size.board-rail` 4, `border.width` 1 and `border.width-strong` 2, and avatar initials at `type.meta`

All of them are in the file and flip-verified. Question 4 has now fired twice in two stages, both times before a part was drawn, which is the whole case for asking it early.

**Stage 2 — the proof run: two parts, one screen, one full check**

The stage that pays for itself. Build the two smallest parts, build one real screen out of them, and check the whole chain **before** committing to the other seven.

**Why these are parts and not just drawings.** [ADR-020](adr/adr-020-design-tool-penpot.md) §4 settles it: every appearance is a **variant of one component**, never a separate frame. The Card is the worked example — five appearances across two densities and two themes is **one** item in the file rather than twenty copies of a rectangle. The difference shows up on the first change, not at the start: move the card's padding token once and every card on every board moves with it. Draw the twenty copies instead and the file is out of date the first time anything shifts — and a design file nobody trusts is one nobody opens.

In order:

1. **Input** — §1.1, §1.2 (the fields), §1.6 (inline title editing)
2. **Button** — §1.1, §1.2
3. **Board 2.1, Login** — §1.1, assembled from those two and nothing else
4. **The full check**, below

**The full check — four questions, in this order:**

| # | Ask | How |
|---|---|---|
| 1 | Does every part of the screen point at a shared value? | The binding check in §7.5 |
| 2 | Does the screen survive **Light → Dark**? | Switch the `Theme` group and look — background, text and borders together |
| 3 | Does it survive **Fine → Coarse**? | Switch the `Density` group and look for anything that now overflows or collides |
| 4 | **Did you have to type a number or a colour because no shared value existed?** | Only the person who just drew it can answer. Every *yes* is a gap in the catalogue — add the value **now** |

**Question 4 is the reason this stage exists.** The first three can be run at any time, and will be, at the end of every stage. The fourth can only be answered by somebody who has just tried to build a real screen out of the catalogue, and it is the one that gets more expensive the longer it waits.

#### What question 4 returned the first time — and what closed it

Run on 2026-11-20, before a single part was drawn. The catalogue's original 103 values were derived from [ADR-017](adr/adr-017-ui-foundations.md) §9, whose vocabulary is the **board**: cards, columns, canvas, avatars, meta rows. Login is a **form**, and no form geometry had ever been needed to describe a board — so control height, horizontal padding, field gap and form width were all absent, while every colour, radius and type step Login needs was present.

**Settled by [ADR-022](adr/adr-022-form-control-geometry.md)** (2026-11-20), which extends §9 with four values: `size.control-height` 36 / 44, `spacing.control-padding-x` 12, `spacing.field-gap` 16 / 20, `size.form-width` 400. They are in the file and flip-verified, so **Stage 2 is a drawing task again.**

**Run question 4 anyway.** It found this one before nine parts were built on top of it, which is the whole case for the order in §7.4.

#### What question 4 returned when Stage 2 ran — and what closed it

Asked again on 2026-11-20, when Stage 2 was actually run, of every value Input, Button and Login need, before drawing. ADR-022 had given a control its outside and ADR-023 the space between objects; neither needed **the inside of one field or one control**. Five values were missing:
- the gap between a field's label, box and message
- the focus ring's 2px offset, which a Penpot stroke cannot draw (ADR-023 *Out of scope* expected this)
- the spinner's size
- the gap between the spinner and the label
- a fill for *pressed* on the plum button, since ADR-018 §6's pressed values are for an edge

**Settled by [ADR-024](adr/adr-024-control-and-field-internals.md)** (2026-11-20):
- `spacing.field-inner-gap` 8 and `spacing.icon-gap` 8, kept apart by ADR-022 §4's test
- `size.icon` 16
- `shadow.focus-ring`, two stacked shadows in each theme set: the way Tailwind's `ring-offset` renders, so the file draws what the code ships
- pressed reuses `color.primary-hover`

It also records one thing found **while** drawing: an input's box must be filled (`color.background`), because Penpot casts an unfilled board's shadow from its children, and a transparent box drew the ring around its letters.

All five are in the file and flip-verified, and Stage 2 then drew clean on them. Question 4 has fired at three stages out of three, every time before a part was drawn.

**Why Login and not Register.** Register is Login plus a name field (§1.2), so drawing it here would add work without adding an answer. It goes with the other screens in Stage 4.

**If the check surfaces nothing, that is a result rather than a waste.** It means the catalogue holds, and Stage 3 runs without looking over its shoulder.

**Stage 3 — the remaining seven parts** (§4.2), in this order:

| # | Part | Draw it from | Why it sits here |
|---|---|---|---|
| 3 | **Avatar** | §1.5 (initials, and the empty-name fallback), §4.1 (eight hues in **both** role assignments) | Cards, board tiles and the presence cluster all need it. Always initials — uploads are Phase 2, so never draw a photo |
| 4 | **Count chip** | §1.5 (column header), §1.4 (board tiles) | Small, and the column header needs it to exist first |
| 5 | **Column header** | §1.5 (title, count, `⋯` with move left / move right), §2.14 (drag-over, header left **untinted**) | Needs Count chip |
| 6 | **Banner** | §2.7 (read-only — persistent, non-dismissible, **and it carries Join**) | One board uses it, but that board cannot be drawn without it |
| 7 | **Skeleton block** | §2.9 (card and column geometry, `--muted` fill, **no shimmer**) | Only board 2.9 — shimmer is movement, and movement means a teammate acted |
| 8 | **Toast** | §3 (its own design is deferred; the **four** access-revoked wordings are not) | Draw the container and the longest message it must hold |
| 9 | **Presence cluster** | §1.5 (app bar), §3 (the `+N` overflow sheet) | **Last.** It is a cluster *of Avatars* — the only hard dependency here |

Numbering carries on from Stage 2, because Input and Button are parts 1 and 2. The Card, drawn in Stage 0, is the pattern every row follows.

**Which appearances each one needs is in §4.2, and that table is the authority.** It is not repeated here on purpose: two lists of the same thing drift, and §4.2 is the one already cited by the §Appendix coverage check.

#### What question 4 returned for the Avatar — and what closed it

Asked on 2026-11-20 before drawing. **No number was missing**: size, initials, circle and the muted colours were all in the catalogue. It found a structural problem instead. ADR-017 §3.3 swaps the identity colours' roles between Light and Dark, and ADR-020 §3 kept those colours typed in. A typed-in colour does not follow the Theme switch, so every avatar and rail would need a hand-swapped Dark pairing, and the Light → Dark check could never see them.

The two options were compared side by side with a live mock board. **Settled by [ADR-025](adr/adr-025-identity-hues-as-tokens.md)** (2026-11-20):
- `color.identity.<hue>.subtle` and `.strong`, 32 entries across the two theme sets; the code keeps its TypeScript constant
- `size.card-avatar` renamed `size.avatar`, because the presence avatar is the same size
- `Unknown` is the Phosphor `User` glyph, not an empty circle
- a `Muted` appearance is added for the presence cluster's reconnecting state

Question 4 has now fired at four stages out of four. This is the first time it found a rule rather than a missing value.

#### What question 4 returned for the Count chip — and what closed it

Asked on 2026-11-20 before drawing. The number's colour, its type step and the 4px corners were all in the catalogue. Three values were not:
- **the fill.** ADR-017 §3.4 chose the muted text colour by measuring it against this chip, but §9 never gave the chip a variable. The only token holding both of its values was `color.border`, by coincidence, and the border's job is an edge.
- **the height**
- **the side padding.** No geometry ADR had needed a chip, since a chip is neither a control, a field nor a gap.

**Settled by [ADR-026](adr/adr-026-count-chip.md)** (2026-11-20):
- `color.chip`, `#E7E5E4` / `#44403C`, written as literals in each theme set rather than as `{color.border}`, so the edge ladder can move without moving the surface §3.4 measured
- `size.chip-height` 20, which also sets the minimum width so one digit is a square, in `base` because type does not grow at Coarse
- `spacing.chip-padding-x` 6, measured rather than chosen: `(20 − 8) ÷ 2`, where 8 is Poppins's widest digit at 12px
- the chip holds text, so a board tile can say `12 cards` and `3 members`

Two things it raised were not gaps. `tabular-nums` cannot be drawn (§4.3). ADR-019 §8's `+N` chip belongs to the presence cluster, because it opens a sheet.

Question 4 has now fired at five stages out of five.

#### What question 4 returned for the Column header — and what closed it

Asked on 2026-11-20 before drawing. Height, fill, rule, title type and colour, the chip, and the `⋯`'s colour and tap area were all in the catalogue. Two values were not:
- **the inset of the header's contents.** `spacing.column-padding` is the panel's own padding and is already spent outside the header: ADR-019 §7's *"16 padding"* counts the 8 above it.
- **the gap between the title, the chip and the `⋯`.** `spacing.icon-gap` holds the same 8 for a different job.

Three choices no ADR had made came with them: the `⋯`'s glyph, whether it is shown before hover, and whether drag-over changes the header at all.

**Settled by [ADR-027](adr/adr-027-column-header.md)** (2026-11-20):
- the inset binds `spacing.card-padding-x` (12), by rule: the column title starts where the card titles start, and at Fine the header's `⋯` sits directly above the cards'
- `spacing.inline-gap` 8, in `base`, for neighbours in one row that are not an icon and its label
- the `⋯` is Phosphor `DotsThree` at `size.icon` in a `size.target-min` area, all bound; the Card's typed 13px `⋯` is left as it is
- the column's `⋯` is shown at rest at every pointer, unlike the card's
- one appearance, no axis: drag-over leaves the header untinted, so the board shows it above a tinted body instead of as a second variant

Not gaps: the title's weight (`type.body`, from the 2026-11-18 weight decision), the menu itself (Stage 5) and truncating a long title (code only).

Question 4 has now fired at six stages out of six.

**Stage 4 — the remaining fourteen screens** (§2), in five groups:

| Group | Boards | Why here |
|---|---|---|
| a | 2.2 Register | Login's twin plus a name field — nearly free now, and it adds the field-invalid state |
| b | 2.3 Workspaces, 2.4 Workspace detail | Need Avatar and Count chip |
| c | **2.5 Board — at rest** | §2 calls it *"the reference frame everything else varies from"*. The hinge of the whole stage |
| d | 2.6 drawer, 2.7 + 2.8 (**one component, two settings**), 2.9 cold load, 2.10 empty, 2.11 load failed, 2.12 gone, 2.13 session expired | Each is a variation on 2.5, and cheap once 2.5 exists |
| e | 2.14 Drag, then 2.15 Phone | 2.14 needs 2.5 plus **both** placeholders. 2.15 is a different layout rather than a switch, so it is drawn by hand and goes last |

**Stage 5 — the ~16 overlays** (§3).

Last, and not arbitrarily. §3's entire rule for what happens below 640px is written *relative to the card drawer already filling the screen on a phone*. Until 2.6 exists there is nothing to judge that against.

**Stage 6 — the final check** (§7.5, and [ADR-020](adr/adr-020-design-tool-penpot.md) §6).

**The standing rule: run the check at the end of every stage, not only at the end of the file.** [ADR-020](adr/adr-020-design-tool-penpot.md) warns that a single typed-in number quietly turns the Light/Dark and Fine/Coarse switches from a test into decoration. Better to learn that after two parts than after forty.

### 7.4 The three things that fix this order

Everything else above could be rearranged. These three cannot:

1. **Values before parts.** A part built on typed-in numbers has to be rebuilt, not corrected — which is why the foundations board comes before everything else.
2. **Each part before the screens that use it.** The fifteen screens are assembled *from* the ten parts, so drawing a screen first means drawing its furniture twice: once loose, once properly. Stage 2's proof run is the one deliberate exception, and a narrow one — it builds exactly the two parts Login needs and not a third.
3. **2.5 before 2.6–2.14.** Nine of the fifteen boards are described in §2 as differences from 2.5. Without it there is no "from".

**Why not all nine parts first.** That order was considered and set aside, and it is recorded here so nobody quietly restores it. It does the same total work in the same number of steps; the only difference is that the first time a value reaches a *finished screen* is after nine parts rather than after two. A gap in the catalogue, or a flip that breaks across a whole screen, would then be found with nine parts already standing on it. Stage 2 trades one change of gear for that answer.

### 7.5 How to check progress, and how to check quality

**Progress — run this rather than trusting a list.** Paste into the Penpot tool's `execute_code`. It reports what the file contains right now:

```js
const cat = penpot.library.local.tokens;
return {
  tokenSets: cat.sets.map(s => `${s.name}: ${s.tokens.length}`),
  themes: cat.themes.map(t => `${t.group}/${t.name}`),
  components: penpot.library.local.components.map(c => c.name),
  pages: penpotUtils.getPages().map(p => ({
    page: p.name,
    boards: penpotUtils.getPageById(p.id).root.children.map(c => c.name),
  })),
};
```

**Quality — find typed-in values that should be tokens.** This is [ADR-020](adr/adr-020-design-tool-penpot.md) §6's check. Change the board name in the first line and paste the rest unchanged:

```js
const board = penpotUtils.findShape(s => s.name === "2.5 Board — at rest");
const loose = s => {
  const b = s.tokens || {}, out = [];
  if ((s.fills || []).some(f => f.fillColor) && !b.fill) out.push("fill");
  if ((s.strokes || []).length) {
    if (!b.strokeColor) out.push("strokeColor");
    if (!Object.keys(b).some(k => k.startsWith("strokeWidth"))) out.push("strokeWidth");
  }
  const L = s.flex || s.grid;
  if (L && !s.name.startsWith("Doc /")) {          // documentation geometry may be typed (below)
    for (const [k, v] of [["paddingTop", L.topPadding], ["paddingRight", L.rightPadding],
                          ["paddingBottom", L.bottomPadding], ["paddingLeft", L.leftPadding],
                          ["rowGap", L.rowGap], ["columnGap", L.columnGap]])
      if (v && !b[k]) out.push(`${k}=${v}`);
  }
  for (const k of ["borderRadiusTopLeft", "borderRadiusTopRight", "borderRadiusBottomRight", "borderRadiusBottomLeft"])
    if (s[k] && !b[k]) out.push(`${k}=${s[k]}`);
  if (s.type === "text" && !b.typography && !b.fontSize) out.push(`typography(${s.fontSize}px)`);
  if ((s.shadows || []).length && !b.shadow) out.push("shadow");
  if (s.opacity != null && s.opacity < 1 && !b.opacity) out.push(`opacity=${s.opacity}`);
  return out;
};
const all = penpotUtils.findShapes(() => true, board);       // excludes the board itself
const shapes = all.some(s => s.id === board.id) ? all : [board, ...all];
return shapes
  .filter(s => !(s.isVariantContainer && s.isVariantContainer()))  // Penpot's own frame, not design
  .map(s => [s.name, loose(s)])
  .filter(([, l]) => l.length)
  .map(([n, l]) => `${n} → ${l.join(", ")}`);
```

**Why it looks like this — replaced 2026-11-20.** The version first written here watched nine property names, and eight of them do not exist on Penpot shapes: colours are `fills[]` and `strokes[]`, and padding and gaps live on `board.flex`. So it could never report an unbound fill, padding or gap. The ninth, `borderRadiusTopLeft`, exists on every shape including text, so it reported every square corner, giving 41 false positives on the Card. Two traps the replacement avoids:
- **Only non-zero values count.** A zero has nothing to bind, which matters because the canvas's vertical padding is 0 by design (ADR-023 §2.3).
- **A bound stroke width can be any `strokeWidth*` key.** A column-header rule binds only `strokeWidthBottom` (ADR-023 *Operational note*). Likewise, an unset opacity reads as `null`, and `null < 1` is true in JavaScript, hence the guard.

**What it does not see.**
- **Width and height.** A size set by layout (fill, auto-height text) is indistinguishable here from a typed one.
- **Stroke style** (the placeholder's dashes).
- **Text content.**
- **Whether a shadow has a surface to cast from.** A shape with no fill casts its shadow from its children ([ADR-024](adr/adr-024-control-and-field-internals.md) *Operational note*). So `shadow.focus-ring` bound to an unfilled box passes this check and still draws the ring around the text.

Look at those by eye.

**What it is expected to report.** Anything not on this list is a defect:

| Shape | Reported | Why it stays |
|---|---|---|
| Anything named `Hue / …`, **on the Foundations board only** | `fill` | Its identity section shows both pairings side by side, which a token-bound shape cannot: it only shows the active theme. Everywhere else the hues are tokens (`color.identity.<hue>.subtle` / `.strong`, [ADR-025](adr/adr-025-identity-hues-as-tokens.md)), so **a `Hue /` shape anywhere else is a defect** |
| The Card's `Menu` | `typography(13px)` | The `⋯` glyph size has no token |

**Shapes that exist only to document.** A shape whose name starts `Doc /` has **geometry** — size, padding, gaps — that belongs to this file's documentation rather than to the product. Examples are a swatch's dimensions, the gap between a swatch and its label, and the gap between eight avatars in a row. That geometry is not a token (ADR-023 *Out of scope*). Borrowing a product token for it, such as the card gap or field gap, is the coupling ADR-023 rejected: the documentation would reflow on Fine → Coarse, and a change to the card gap would move it. **Everything else about a `Doc /` shape is bound like any other shape: fill, stroke colour and width, radius, type, opacity and shadow.** That way the Light/Dark flip still tests the documentation. The check skips padding and gaps on `Doc /` shapes and nothing else. It cannot see sizes on any shape.

Widened from sizes alone to all geometry on 2026-11-20, before the foundations board was drawn, once it was clear the board's own groups need inner gaps that no product token describes.

**By eye — only where looking genuinely beats checking.** Use `export_shape` on the phone board (2.15) and on the state boards as whole compositions. For everything else the two checks above are stricter than an eye.

**One thing no export shows: a stroke whose sides differ in width.** Penpot's PNG and SVG exports draw the column header's bottom-only rule as a zero-width outline, and so it is invisible in every export, on the Foundations board as on the part. Confirm it from the shape's stroke data and `penpot.generateStyle`, or on the canvas ([ADR-027](adr/adr-027-column-header.md) *Operational note*).

### 7.6 The checklist

A plan, not a record — §7.5 is the record. Tick a stage when it closes, and if the two ever disagree, re-derive rather than trust this.

**Stage 0 — foundations in place** ✅ *(2026-11-19)*

- [x] Five token sets, **116 tokens** (103 from ADR-017 §9 + 6 entries from [ADR-022](adr/adr-022-form-control-geometry.md) + 7 from [ADR-023](adr/adr-023-spacing-between-objects.md)) · [x] Four themes in two groups · [x] Card variant container · [x] Form-control geometry settled and flip-verified · [x] Spacing between objects and border widths settled and flip-verified · [x] §7.5's binding check rewritten and proven against the Card

**Stage 1 — foundations board** (1) ✅ *(2026-11-20)*

- [x] 4.1 Foundations board — token catalogue, eight identity hues in both assignments, elevation ladder, radius ladder, five type steps, 14px/16px specimen
- [x] Binding check clean (expected rows only) · [x] Light → Dark · [x] Fine → Coarse · [x] Question 4 asked before drawing — it fired, and [ADR-023](adr/adr-023-spacing-between-objects.md) closed it

**Stage 2 — the proof run** (2 parts, 1 screen, 1 check) ✅ *(2026-11-20)*

- [x] Input · [x] Button · [x] Board 2.1 Login
- [x] Binding check (no rows at all) · [x] Light → Dark · [x] Fine → Coarse · [x] **Every value you had to type in is now a token** — question 4 fired before drawing, and [ADR-024](adr/adr-024-control-and-field-internals.md) closed it

**Stage 3 — the remaining seven parts** (7)

- [x] Avatar *(2026-11-20; question 4 fired and [ADR-025](adr/adr-025-identity-hues-as-tokens.md) closed it)* · [x] Count chip *(2026-11-20; question 4 fired and [ADR-026](adr/adr-026-count-chip.md) closed it)* · [x] Column header *(2026-11-20; question 4 fired and [ADR-027](adr/adr-027-column-header.md) closed it)* · [ ] Banner
- [ ] Skeleton block · [ ] Toast · [ ] Presence cluster

**Stage 4 — the remaining fourteen screens** (14)

- [ ] 2.2 Register
- [ ] 2.3 Workspaces · [ ] 2.4 Workspace detail
- [ ] **2.5 Board — at rest**
- [ ] 2.6 Card drawer · [ ] 2.7 + 2.8 Read-only / join gate · [ ] 2.9 Cold load · [ ] 2.10 Empty board
- [ ] 2.11 Load failed · [ ] 2.12 Gone · [ ] 2.13 Session expired
- [ ] 2.14 Drag · [ ] 2.15 Phone board

**Stage 5 — overlays** (~16)

- [ ] Workspace team panel · [ ] Create workspace · [ ] Invite people · [ ] Create board
- [ ] Board settings · [ ] Board members · [ ] Add people · [ ] Transfer ownership
- [ ] Move card to… · [ ] Destructive confirmation · [ ] Board `⋯` · [ ] Column `⋯` + inline rename
- [ ] Card `⋯` · [ ] User menu · [ ] Presence overflow · [ ] Toast content · [ ] Date + assignee pickers

**Stage 6 — final check**

- [ ] Binding check clean on every board · [ ] Light/Dark flip reviewed · [ ] Fine/Coarse flip reviewed · [ ] 2.15 exported and the "6 cards visible" number confirmed by eye

### 7.7 What is deliberately not part of this order

- **The `archived` query parameter.** The one open backend item. It decides whether the Archived list can be *filled with real data*, never whether board 2.4 can be *drawn*. Build it whenever.
- **Scaffolding the React app.** A separate root, with no edge to this one in either direction. [ADR-020](adr/adr-020-design-tool-penpot.md) is explicit that the design file generates no code and exists as a reference, so neither effort ever waits on the other.
- **The Astro landing page.** Outside this document entirely (§0).

---

## Appendix — the 39 states, and where each is drawn

A completeness check against [ADR-018](adr/adr-018-state-vocabulary.md) §1. Two states are closed by inspection rather than drawn — 35 was a third until 2026-11-19.

| # | State | Drawn on |
|---|---|---|
| 1, 2 | Pending create / change | Card variant `InFlight` — 60% opacity, meta row promoted from `--muted-foreground` to `--foreground` |
| 3 | Server-corrected | Folds into the movement channel; no separate frame |
| 4 | Rejected | **Not animated** — instant return, toast, dim lifts |
| 5 | Drag-armed | Board 2.14; `--shadow-overlay` ramping over the hold, with `scale(1.01)` |
| 6, 7, 8, 9 | Drag source, overlay, placeholder, column tint | Board 2.14 |
| 10 | Drag-rejected | No channel — instant return |
| 11 | Draggable at rest | Cursor at fine; **nothing at coarse** — the one accepted gap in the whole vocabulary |
| 12, 15 | Hover, pressed | Card and Button variants — border colour steps, never a fill |
| 13 | Hover-equivalent at coarse | The persistent `⋯` (§1.5) |
| 14 | Focus-visible | Every focusable component variant |
| 16, 17 | Not satisfied, busy | Button variants (§4.2) |
| 18 | Selected | **Does not exist in Phase 1** — which is what released 2px border weight for the placeholder |
| 19 | Cold load | Board 2.9 |
| 20 | Background refresh | **Nothing** — what's on screen is correct |
| 21 | Empty, **six** flavours | Boards 2.3, 2.4 (Active **and** Archived), 2.10, the in-column surface, and the add-people panel (§6.5) |
| 22, 23 | Load failed, gone | Boards 2.11, 2.12 |
| 24, 25 | Read-only, join gate | Boards 2.7, 2.8 — **one component, two settings**, and 2.7's banner carries Join |
| 26 | Access revoked | **Navigates you off the board**, plus a toast naming the reason. Nothing reverts — this is not 27. Four reasons, four wordings |
| 27 | Action forbidden | Toast; the optimistic change reverts via the in-flight dim |
| 28 | Connecting | Part of cold load (2.9) |
| 29 | Reconnecting | Presence cluster variant (§4.2) |
| 30 | Live | **No treatment** — colour is an event, not a texture |
| 31 | Remote change arrived | The movement channel; not a static frame |
| 32 | Remote deletion of the open card | Card drawer variant (§1.6) |
| 33 | Presence change | Presence cluster |
| 34 | Private board | Filled Phosphor lock, on board tiles (§1.4) |
| 35 | Archived board | **Reachable after all** (§6.3) — but it still spends no channel. An archived board appears only in the Archived list, where every row is archived, so nothing on the tile needs to mark it. The filter is the signal |
| 36 | Field invalid | Input variant; boards 2.1, 2.2 |
| 37 | Inline editing | Input variant; column rename, card title |
| 38 | Destructive confirmation | Overlay (§3) |
| 39 | Session expired | Board 2.13 |

---

## Related documents

- **Visual baseline:** [ADR-017](adr/adr-017-ui-foundations.md) — tokens, density, structure, type, elevation, modes
- **Responsive and touch:** [ADR-019](adr/adr-019-responsive-touch-scope.md) — layouts, drag, the coarse scale
- **State vocabulary:** [ADR-018](adr/adr-018-state-vocabulary.md) — all 39 states and their measured values
- **The medium:** [ADR-020](adr/adr-020-design-tool-penpot.md) — Penpot sets, themes, variants, verification
- **Form-control geometry:** [ADR-022](adr/adr-022-form-control-geometry.md) — control height, padding, field gap, form width; extends ADR-017 §9
- **Spacing between objects:** [ADR-023](adr/adr-023-spacing-between-objects.md) — screen inset, column gap, no vertical canvas padding, section gap, rail height, border widths, initials; extends ADR-017 §9 and ADR-020 §3
- **Control and field internals:** [ADR-024](adr/adr-024-control-and-field-internals.md) — label gap, icon size and gap, the focus ring as a shadow, the input's fill, pressed on a filled button; extends ADR-017 §9
- **Identity hues in the design file:** [ADR-025](adr/adr-025-identity-hues-as-tokens.md) — the 16 identity colours as `subtle` / `strong` tokens that follow Theme, `size.avatar`, the Avatar's appearances; supersedes ADR-020 §3's identity exclusion
- **The count chip:** [ADR-026](adr/adr-026-count-chip.md) — `color.chip` as its own role, the 20px height and minimum width, the 6px side padding, one appearance holding any text; extends ADR-017 §9
- **The column header:** [ADR-027](adr/adr-027-column-header.md) — the inset bound to `spacing.card-padding-x`, `spacing.inline-gap`, the `⋯` as Phosphor `DotsThree` shown at rest, one appearance; extends ADR-017 §9
- **What this document's own open questions settled into:** [ADR-021](adr/adr-021-design-inventory-corrections.md) — archive, column reorder, overlay presentation, the read-only exit
- **Routes and data flow:** [frontend-architecture.md](frontend-architecture.md)
- **Product intent:** [001-PRD.md](../product/001-PRD.md) · **Acceptance per feature:** [001-user-stories.md](../product/001-user-stories.md)
