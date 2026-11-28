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

Twenty-four documents fix everything here; this one only arranges it.

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
| [ADR-028](adr/adr-028-banner.md) | The read-only banner — on `color.background` with a bottom rule, `spacing.screen-x` at the sides, `spacing.banner-padding-y` 8 above and below, `spacing.inline-gap` before Join; one appearance, because nobody can see an informational one |
| [ADR-029](adr/adr-029-skeleton-block.md) | The cold-load skeleton block — its own fill (`color.skeleton`, because `color.muted` *is* the column in Light), `size.skeleton-height` 60 / 72, card-sized blocks inside real column panels, never on the bare canvas; one appearance, no outline |
| [ADR-030](adr/adr-030-toast.md) | The toast — inverted (`color.toast` / `color.toast-foreground`, because `color.popover` is 1.00:1 over a card in Light and a column in Dark), no edge, `spacing.toast-padding` 12, `size.toast-width` 358, a close button shown at rest; the three access-revoked wordings; `/workspaces` after `removed_from_workspace`; one appearance, no axis |
| [ADR-031](adr/adr-031-presence-cluster.md) | The presence cluster — faces overlapping by `spacing.avatar-overlap` −6 / −8 with a 2px `shadow.avatar-ring`, grey faces on `color.avatar-neutral` (because `color.muted` is 1.09:1 on the white app bar), `+N` as a round seat, three faces below 640px and five above with you first, *Reconnecting…* in the seat's place on a phone; State = Live · Reconnecting · FocusVisible |
| [ADR-032](adr/adr-032-auth-form-errors-and-password-rules.md) | Where Login and Register report errors — a `Form message` just above the button for a form-level failure, a taken email on its field; Register's four password rules as a live `Password rule` checklist (`spacing.list-gap` 4, met in `color.done`); a confirm-password field; *Sign in*; every wording on both forms; fifteen behaviour rules |
| [ADR-033](adr/adr-033-workspace-screens.md) | The workspace screens — both on `color.background`; `size.app-bar` 56; a roomy `Tile` (padding 16, rows 8 apart, a left rail in the hue's `.strong`); the grid's `spacing.tile-gap` 8 / 12 and `size.tile-min-width` 264; the Active / Archived `Segment` (plum text and edge on the current one); an Outline `Kind` on the Button; *Delete workspace* behind an owner-only `⋯`; centred empty states with a plum-tinted `Illustration`, and the rules for drawing the next one; *Owner · Admin · Member*; every wording on both screens |
| [ADR-034](adr/adr-034-side-menu.md) | The sidebar on 2.4 and the board — a workspace switcher, *All boards* and the boards you've joined; beside the content from 1024 up and sliding over it below (a fifth presentation kind, *navigation*); *CollaboTask* at its top and the user menu at its bottom; the app bar naming the current row; plum text and edge on the current row, a 12px colour square per board, 240 wide, `color.scrim`; 2.3 stays, and the app opens in your last workspace |
| [ADR-037](adr/adr-037-board-at-rest.md) | The board at rest — the first card a card gap below the column header; the rail as its own 4px strip under the app bar; the bar's name on the left and presence, *Add column* and ⋯ on the right; *Add card* as a quiet row after the last card that sticks to a long column's bottom; late dates in words (*Due today*, *Due tomorrow*, *Overdue · Nov 20*); *No cards yet.* in an empty column |
| [ADR-040](adr/adr-040-menus-and-the-confirmation.md) | The menus and the confirmation (Stage 5, group a) — one menu surface, a 224 popover from 640 up and a bottom sheet with a head below; a glyph per entry; rose for the deletes and *Leave board*; five actions that ask first; *Undo* after archiving kept; typing the name to delete a workspace; a 448 dialog, 24 inside; the rose button and its pale rose busy; inline rename filling the header; *Sign out*; every wording and where each action leaves you; the `Overlays` page |
| [ADR-041](adr/adr-041-the-card-drawers-pickers.md) | The card drawer's pickers (Stage 5, group b) — every picker in place under its field, *Move card to…* too, so a phone never shows a sheet over a sheet; from the board, Move in the card ⋯'s popover or sheet; a column, then a place; the menu surface laid in the flow; plum text and edge on the current choice; a caret on all three fields; a six-week calendar with today semibold; a due date as a day at midnight UTC; *Unassigned*, you, then A–Z, with a filter above eight members; the drawer at 390 |
| [ADR-042](adr/adr-042-the-forms.md) | The forms (Stage 5, group c) — group a's 448 dialog from 640 up and a full-height sheet below, no ×, pressing outside does nothing; a field of several lines that grows from one line, `spacing.control-padding-y` 8 / 12; a minimum length shown as you type; the colour as eight buttons holding their hue (20 / 28), the first unused hue first, stored as the deep hex value; who can see it as two choice rows; *Save changes* and a line before going private; *Invite people* on 2.4's heading, addresses as chips, results under the field; *Add column* as a new column on the board, not a form |
| [ADR-043](adr/adr-043-the-lists.md) | The lists (Stage 5, group d) — group c's dialog and sheet hold the team and a board's members, eight rows before the list scrolls; a person's row with the role and email on its second line; you first, then A–Z; a ⋯ per row; ten actions that ask first; *Leave workspace* in the workspace ⋯; adding people one at a time as the board list's second group; *Make owner* and a confirmation; the presence list as a popover; two backend items; the file's Product Team counts and Q3 Launch Plan's owner corrected |
| [ADR-044](adr/adr-044-the-sidebar-sliding-over-and-the-toast.md) | The sidebar sliding over and the toast (Stage 5, group e) — the slide-over drawn on the phone, focus on the current row, the user menu's sheet over it below 640; a toast 16 from the bottom and right, across the bottom on a phone; up to three stacked, newest at the bottom, 8 apart; no timer; *Undo* as the Outline button, taking you back to the board and teammates back from 2.12; *Couldn’t move “X”.* and five reasons; the board reloads after a 403 or 404 |
| [ADR-045](adr/adr-045-design-file-pages-by-area.md) | The design file's pages, by area of the app — Foundations, Components, Account, Workspaces, Board; `Page 1` renamed Workspaces and `Overlays` renamed Board, so 31 of 73 boards moved; one row per screen (at rest, its states, what opens over it, its phone versions last), each under a `Doc / Row label`; the phone's user menu stays on Board; no board renamed |
| [ADR-048](adr/adr-048-board-page-split-in-four.md) | The Board page split in four, after Stage 6 — Board (2.5 and its states, 2.7, 2.14), *Board — what opens over it*, Card (2.6) and *Board — phone* (2.15), each row moved whole; Board held 64% of the file's shapes. Supersedes ADR-045 §1, §5's page and §7's `Board` table only

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
| Wording: heading and button *Sign in* (busy: *Signing in…*); under the form, *Don’t have an account yet?* **Create one** | ADR-032 §7 |
| Variants: field invalid (36), busy (17 — `--primary-subtle` fill, spinner, label becomes the progressive verb), form not satisfied (16 — `--muted` fill), focus-visible (14 — 2px `--ring`, 2px offset) | ADR-018 §5, §6 |
| **A failed sign-in (401) is reported once, above the button**, by the `Form message` part: *Wrong email or password.* Never under a field: the 401 is the same for an unknown email and a wrong password, and US-02 forbids saying which. The same line carries *can’t reach* and *something went wrong* | ADR-032 §1, §7 |
| **No "forgot password" link.** There is no reset endpoint, and the product cannot send email at all — invites resolve an address that already has an account. The absence is now a recorded deferral rather than an omission, so **don't draw a link that would have nowhere to go** | PRD §5.2; ADR-021 *Out of scope* |

No separate narrow board. A two-field form reflows at any width, and the density flip supplies the 44px targets.

### 1.2 `/register`

Same furniture plus a name field and a confirm-password field. Three additions worth drawing:

- A **duplicate-email conflict renders as field-invalid, not as a toast** — it points at a field, so it belongs on the field (*An account with this email already exists.*).
- **Four password rules, said before they submit, not after**: at least 8 characters, an uppercase and a lowercase letter, a number, a symbol. They are a live checklist under the password field (the `Password rule` part). Rows tick as you type, and the missed ones turn rose only after the button is pressed. *2026-11-21: this line first read "the password minimum is 8 characters and it is the only limit". The four rules replaced that by the user's decision, taken against NIST SP 800-63B-4 ([ADR-032](adr/adr-032-auth-form-errors-and-password-rules.md) §3), and US-01 and SRS UC-01 changed with them.* (Name 1–255 and a valid email address need no stating.)
- **Confirm password**, checked when the button is pressed (*Passwords don’t match.*). It is compared in the browser and never sent.

Board 2.2 is drawn at the moment the email comes back taken, with everything else filled and all four rules met (ADR-032 §9).

### 1.3 `/workspaces`

| On it | Source |
|---|---|
| App bar, `size.app-bar` 56 — *CollaboTask* on the left in `type.title`, linking here; the user menu (an Avatar in a `size.target-min` box) on the right. A bottom rule, `spacing.screen-x` at the sides | ADR-017 §1 derivation; ADR-033 §2 |
| Workspace tiles: name, **my role**, member count, active board count — the `Tile` part with its rail hidden, a workspace having no colour. Roles read *Owner · Admin · Member* (Owner from `owner_id`); counts *5 members · 4 boards*. No faces: the list returns no names | US-05; ADR-033 §3, §11 |
| "Create workspace" primary action — the screen's one Primary button, and the empty state's one button too | US-03; ADR-033 §8 |
| Variants: cold load (19 — static skeleton, no shimmer; blocks at `size.tile-skeleton-height` in the tile grid, directly on the page, never inside a tile, [ADR-029](adr/adr-029-skeleton-block.md) §4), empty (21, flavour 1 — the sprout illustration and the sentence, centred in the grid's place), load failed (22 — a sentence and an Outline *Try again*) | ADR-018 §7; ADR-033 §8, §10 |

**Creating a workspace puts you inside it.** The panel closes and the app navigates to the new workspace — it does not return you to a list with one more tile on it. Same rule one level down (§1.4), and together they are the whole of the product's onboarding: PRD §4.1's "about a minute" is a consequence of these two navigations, not of a welcome tour. There is no welcome tour (ADR-021 *Downstream edits*; PRD §4.1).

The **empty state here is the first screen a new account ever sees**, so it carries the weight a welcome screen otherwise would: one sentence and one button. *2026-11-21:* and an illustration above the sentence, by the user's request ([ADR-033](adr/adr-033-workspace-screens.md) §8). The button is still the heading's *Create workspace*.

Both workspace screens sit on `color.background`, the page the forms use ([ADR-033](adr/adr-033-workspace-screens.md) §1).

**No sidebar here** ([ADR-034](adr/adr-034-side-menu.md) §2), so this app bar is unchanged. The app opens in the last workspace you used (ADR-034 §6), so 2.3 is now seen by a new account, from the switcher's *All workspaces* (§1.7), and when the remembered workspace can no longer be opened.

### 1.4 `/workspaces/:wsId`

One route carries both the boards and the team. **Composition settled by [ADR-021](adr/adr-021-design-inventory-corrections.md): the board grid fills the screen; the team opens over it from a member-count button.** Boards are why you come here almost every time; the team is managed once, at the start. That keeps this as **one** board in the file rather than two, and the team list becomes an overlay (§3) that inherits §3's presentation rule for free.

| On it | Source |
|---|---|
| **The sidebar** beside the screen from 1024 up (§1.7). The app bar's title is *All boards* in `type.title`, with the sidebar's show/hide button at its left while the sidebar is hidden or below 1024. *CollaboTask* and the user menu are in the sidebar, not in this bar. With the sidebar open, the grid fits 4 tiles across at 1440, 3 at 1280 and 2 at 1024 | [ADR-034](adr/adr-034-side-menu.md) §5, §11 |
| Workspace header: name, **description**, my role, and a **"N members" button** that opens the team panel. The name in `type.heading` with the role beside it, the description under it; *5 members* is an Outline button. **Since 2026-11-24, *Invite people* follows it**, an Outline button for workspace admins that opens the invite form ([ADR-042](adr/adr-042-the-forms.md) §12) | US-05b, US-04; ADR-033 §6 |
| An **Active / Archived** switch above the board grid. Not a route — the same grid, one filter. Two `Segment` instances: the current one has plum text and a 1px plum edge (plum's *where you are*, not state 18) | US-12c; ADR-021 §1; ADR-033 §5 |
| Board tiles: title, **identity swatch** from `board.background_color`, my role, joined / can-join, member count, card count. The `Tile` part: a left rail `size.board-rail` wide in the hue's `.strong`; *12 cards · 3 members*; roles *Owner · Member* | ADR-017 §3.3; US-08; ADR-033 §3 |
| A **filled** Phosphor lock on private boards — the icon channel's only Phase-1 meaning, and the only place fill is permitted | ADR-018 §9 |
| A Join affordance on a board I may join but haven't — an Outline *Join* in the tile's action row, with no role shown | US-09; ADR-033 §6, §11 |
| Create board; delete workspace behind a destructive confirmation (38). *Create board* is Primary; *Delete workspace* sits in a `⋯` at the end of the heading row, which only the owner sees | US-07, US-06d; ADR-033 §7 |
| Variants: cold load (19; blocks in the tile grid on the page, [ADR-029](adr/adr-029-skeleton-block.md) §4; the header waits for its data, the switch and *Create board* draw at once), empty boards (21, flavour 2 — the empty-board illustration), **empty archived list (21, flavour 3 — the open-box illustration)**, load failed (22), and the non-member case which is refused rather than rendered | ADR-018 §7, §10; ADR-033 §8, §10 |

**Archived boards.** Under *Active* they are absent — not greyed, not iconned (US-08). Under *Archived* every tile is archived, so the archive icon would mark nothing the filter has not already said: **ADR-018 §9's single-meaning icon channel is untouched by this filter** (ADR-021 §2). An archived tile's only action is **Unarchive**; it cannot be opened, and that is deliberate — an archived board can be found, never browsed.

**Delete workspace is owner-only, so for an admin who is not the owner the control is _absent_, not refused** (US-06d; ADR-018 §5).

**Creating a board puts you on it**, columns and all — same rule as §1.3, and the second half of PRD §4.1's minute.

**The swatch must fall back on read.** It is constrained to the eight identity hues, but `background_color` validates only `min=4,max=8` and is not a hex validator, so the string `"pink"` reaches the client (ADR-017 §3.3). There is a second field of the same species on the card itself — see §1.5.

**The workspace's name and description cannot be edited** — there is no update endpoint, so what was typed at creation is what this header shows forever (PRD §5.1). Don't draw an edit affordance on them.

### 1.5 `/workspaces/:wsId/boards/:boardId` — the board

The product. Two layouts, crossing over at 640px (ADR-019 §2); from 640px up, ADR-017 §1–§2 apply unchanged.

| Region | Contents | Fine → Coarse |
|---|---|---|
| App bar, 56px | Board name 20/28; **presence cluster**; board `⋯`; add-column action. The sidebar's show/hide button sits at its left while the sidebar is hidden or below 1024; with the sidebar open from 1024 up, the bar starts at the sidebar's edge (§1.7). **No user menu**: it is at the bottom of the sidebar ([ADR-034](adr/adr-034-side-menu.md) §5). **Arranged by [ADR-037](adr/adr-037-board-at-rest.md) §3:** the name takes the room that is left; then presence, *Add column* (Primary, the board's one plum button) and the ⋯ in `color.foreground`, `spacing.inline-gap` apart, ⋯ last. **Below 640 there is no room for *Add column*: it is the first entry in the board ⋯ menu** ([ADR-039](adr/adr-039-drag-and-phone-board.md) §7) | — |
| Board accent | A 4px identity rail, from the eight hues, **in its own strip under the app bar**, spanning the board area to the right of the sidebar ([ADR-037](adr/adr-037-board-at-rest.md) §2). **Not** a canvas colour — the canvas is always neutral `--board-canvas` | — |
| Canvas | `--board-canvas` `#E7E5E4`, horizontal scroll. 16px side inset (`spacing.screen-x`) and **no top or bottom padding**: the column panels start under the rail and are **as tall as their content, at most to the bottom of the screen** (ADR-023 §2.3, narrowed by [ADR-037](adr/adr-037-board-at-rest.md) §4) | column gap 8 → 12 (`spacing.column-gap`) |
| Column panel | Own `--board-column` fill, 8px radius, 8px padding, 280px wide. The first card sits `spacing.card-gap` below the header (ADR-037 §1) | 280px → `min(100vw − 72px, 420px)` below 640 |
| Column header | Shares the panel fill; `border-bottom: 1px solid var(--border)` is **load-bearing** because columns are unbounded and cards scroll under it. Title, count chip, `⋯` (rename, **move left**, **move right**, delete). The `⋯` is **shown at rest at every pointer**, unlike the card's ([ADR-027](adr/adr-027-column-header.md) §4), and the title starts where the card titles start | 36px → 44px; `⋯` area 24 → 44 |
| Card | 14/20 title clamped to **2 lines** (mandatory — titles accept 500 characters), 1px `--border`, **no resting shadow**, conditional meta row, `⋯` top-right | padding 8/12 → 12/12; avatar 20 → 24; meta row 24 → 28; gap 8 → 12; title right padding 24 → 32 |
| Meta row | Rendered **only** when the card has an assignee or a due date. Initials avatar on the right, due date in `tabular-nums` on the left. A late date says so in words: *Overdue · Nov 20* in rose before today, *Due today* or *Due tomorrow* in amber, the plain date after that (ADR-037 §5) | — |
| Avatars, everywhere | **Always initials in Phase 1.** Uploads are deferred (PRD §5.2), so `avatar_url` is never set — don't draw a photo variant | PRD §5.2 |
| Column footer | Add card: the `Add card` part, a quiet row right after the last card. In a column longer than the screen it sticks to the bottom, with a top rule on its `Footer` (ADR-037 §4). An empty column says *No cards yet.* above it (§6.5) | row 36 → 44 |

Per-card height is `8 + lines×20 + (24 if meta) + 8` at Fine and `12 + lines×20 + (28 if meta) + 12` at Coarse. The smallest card is 36px at Fine and exactly 44px at Coarse, so cards clear the target floor by construction rather than by special-casing.

**The initials avatar needs a fallback, for the same reason the colour swatch does.** An assigned card always carries an assignee object, but its `name` is filled only if the server's user lookup succeeded — so `{"id": "…", "name": "", "avatar_url": null}` is representable. Initials derived from an empty string produce an empty circle. Draw the fallback (a neutral `--muted` circle, or the Phosphor user glyph) rather than discovering it as a blank on screen. **Settled by [ADR-025](adr/adr-025-identity-hues-as-tokens.md) §4: the glyph, on a muted circle** — an empty muted circle is 1.09:1 against a white card and invisible on a dark one. Same species as `background_color: "pink"` in §1.4: a field the API's validation does not actually guarantee.

The card's `⋯` is **hover-revealed at fine pointer and persistent at coarse** (ADR-019 §6), and it does not go in the meta row — that row is conditional, and putting `⋯` there would force it onto every card.

**There is no board-level "add card".** A card belongs to a column, so a top-bar Add card would have to ask which one first. The column footer already answers that by standing in the column. Removed from this table on 2026-11-19.

#### The board `⋯` menu

Never enumerated before 2026-11-19, and it is the only home five P1 stories have. Entries the viewer may not use are **absent, not greyed** (ADR-018 §5), so this menu has three shapes:

| Entry | Opens | Who sees it | Source |
|---|---|---|---|
| Add column | What the bar's *Add column* does: since 2026-11-24 a new column at the end of the board, its header a field, not a form ([ADR-042](adr/adr-042-the-forms.md) §14). **Below 640 only**, first in the menu, because the phone's bar has no room for it | Board members and workspace admins (`CheckMutateAccess`), as the bar's button | UC-13; [ADR-039](adr/adr-039-drag-and-phone-board.md) §7 |
| Board settings… | Overlay (§3) — title, description, colour, visibility | Board owner or workspace admin | US-12b |
| Members… | Overlay (§3) — add, remove, transfer ownership | Anyone who can read the board; add/remove/transfer inside it are owner-or-admin only | US-11, US-12d, US-12e |
| Archive board | Confirmation (38) with a **Primary** confirm, since archiving can be undone; then **navigates out** with an *Undo* toast | Board owner or workspace admin | US-12c; [ADR-040](adr/adr-040-menus-and-the-confirmation.md) §3, §5 |
| Leave board | Confirmation (38), **rose** in the menu and on the confirm: leaving clears your assignments on the board for good | Anyone on the board **except its owner** | US-10; ADR-040 §3 |

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

**Settled and drawn by [ADR-038](adr/adr-038-board-states-group-d.md)** (2026-11-23), at 640 and up. The drawer covers the sidebar and the bar, full height on the right, over `color.scrim`. It is `size.drawer-width` (560) wide, in `color.popover` with `shadow.rule-left` and `shadow.overlay`, with `spacing.screen-x` inside. Labels sit on the left at `size.field-label` (96), and each value is a `Field`. *Column* opens *Move card to…*. *Delete card* is a quiet row in `color.destructive-text` at the foot. The date reads *Due tomorrow · Nov 24* (§1–§6).

**Its pickers are settled and drawn by [ADR-041](adr/adr-041-the-card-drawers-pickers.md)** (2026-11-23). Each field opens its picker in place, under its value from 640 up (432 wide) and across the row below 640. *Assignee* opens the members: *Unassigned*, you, then A–Z, with a filter above eight members. *Due date* opens a six-week calendar, today semibold, with *Remove due date* under it. *Column* opens *Move card to…*: a column, then a place. All three fields carry the caret, and the current choice is plum text and a plum edge. The drawer at 390 is drawn there too, as a full-height sheet between the browser's bars.

**Below 640px this drawer is the thing everything else opens on top of**, which is what shapes §3's presentation rule. The date picker, the assignee picker and the delete confirmation are all reached from inside it, so none of them may become a second full-height sheet: the pickers **expand in place inside the drawer** (*Move card to…* as well, since [ADR-041](adr/adr-041-the-card-drawers-pickers.md) §1), and the confirmation stays a centred dialog at every width (ADR-021 §5).

### 1.7 The sidebar — on §1.4, §1.5 and §1.6

Not a screen: a column beside two of them. Settled by [ADR-034](adr/adr-034-side-menu.md).

| On it | Source |
|---|---|
| A top row `size.app-bar` tall: the show/hide button (Phosphor `SidebarSimple` at `size.icon` in a `size.target-min` box), then *CollaboTask* in `type.title`. Its bottom rule continues the app bar's | ADR-034 §10.1, §10.7 |
| **The workspace switcher**: the workspace name in `type.card-title`, with Phosphor `CaretUpDown` at the right. It opens your workspaces **inside** the sidebar, the current one marked, ending with *All workspaces* (→ 2.3). Picking one goes to its 2.4. The workspace rows keep the empty 16px slot, so their names line up with the board names, and the selected one carries a **still plum dot** (`size.dot` 8) in its slot | ADR-034 §7; [ADR-035](adr/adr-035-sidebar-loading-failure-and-switcher-mark.md) §3 |
| ***All boards***, with Phosphor `SquaresFour`, → 2.4. Current while you're on 2.4 | ADR-034 §8 |
| ***Your boards*** (`type.meta`, `color.muted-foreground`, `spacing.section-gap` above it), then **only the boards you've joined**, newest first: a 12px square (`size.swatch`) in the hue's `.strong`, the title, and the filled lock if private | ADR-034 §8, §10.4 |
| **The user menu** at the bottom: an Avatar at `size.avatar` and your name, under a `color.border` rule. The menu it opens is drawn: your face, name and email, then *Sign out*, 224 wide, opening upward with its left edge on the row's | ADR-034 §5; [ADR-040](adr/adr-040-menus-and-the-confirmation.md) §11 |
| **The current row**: plum text and a 1px plum edge, as on the Active / Archived switch. The other rows step the edge on hover and pressed, and focus adds `shadow.focus-ring` | ADR-034 §10.3 |
| Empty (flavour 7, §6.5): *You haven’t joined a board yet.* in `type.body` and `color.muted-foreground`, with no illustration | ADR-034 §8 |
| Cold load (19): the switcher row and three board rows are **row-sized Skeleton blocks** (`size.control-height`, `radius.control`). The top row, *All boards*, the label and the user row draw at once | ADR-035 §1 |
| Load failed (22): ***Your boards didn’t load.*** in `color.destructive-text`, then a ***Try again*** row with Phosphor `ArrowClockwise`, uncoloured. The only rose a load failure spends anywhere | ADR-035 §2 |

**Where and when.** On 2.4 and the board, with all of the board's states, not on 2.3. From 1024 up it sits beside the content, `size.sidebar-width` (240) wide, on `color.background` with a `border.width` rule on its right. It is open by default, and the choice is remembered on this device. Below 1024 it slides over the content from the left, with the content dimmed by `color.scrim`, and starts closed. That version is an overlay, drawn 2026-11-24 on the phone, on what is now the *Board — phone* page's 2.15 row ([ADR-048](adr/adr-048-board-page-split-in-four.md); [ADR-044](adr/adr-044-the-sidebar-sliding-over-and-the-toast.md) §1–§3, and §3 here). Ctrl + \\ (⌘ + \\) shows and hides it at every width.

**Its values.** Rows are `size.control-height`, `spacing.control-padding-x`, `radius.control` and `type.body`, sitting `spacing.list-gap` apart. The list sits `spacing.sidebar-padding` (8) in from the sides, so at Fine a row's text lines up under the button's glyph.

**The app bar names the current row**: the board's name on a board, *All boards* on 2.4.

**What the board's states do to it.** On 2.7 and 2.8 no row is marked current, because the board isn't one you've joined. On 2.12 the board leaves the list. On 2.4's cold load and load failure the list shares the grid's request, so it shows the two rows above. On a board the list has its own request, so the same two looks appear whenever that request is cold or fails ([ADR-035](adr/adr-035-sidebar-loading-failure-and-switcher-mark.md)).

**The rest of its geometry** (ADR-035 §4–§6): *Your boards* sits `spacing.list-gap` above its rows. The user row's Avatar takes the slot's place, so your name starts 4 right of the board names at Fine and 8 at Coarse. *Your boards* and both sentences are inset by `spacing.control-padding-x`.

---

## 2. Whole-screen boards

[ADR-020](adr/adr-020-design-tool-penpot.md) §5 fixes this list: a board only where **the whole screen differs**. Fifteen, numbered as they'll be named in the file.

| # | Board | Why it is whole-screen |
|---|---|---|
| 2.1 | Login | §1.1 |
| 2.2 | Register | §1.2 |
| 2.3 | Workspaces | §1.3. Drawn 2026-11-21 with three state boards: cold load, empty, load failed ([ADR-033](adr/adr-033-workspace-screens.md) §13) |
| 2.4 | Workspace detail | §1.4. Drawn 2026-11-21 with five state boards: archived, cold load, empty, archived and empty, load failed (ADR-033 §13). **Redrawn** 2026-11-21, all six, with the sidebar beside them, the *All boards* bar and a 4-column grid ([ADR-035](adr/adr-035-sidebar-loading-failure-and-switcher-mark.md) §7) |
| 2.5 | Board — at rest | §1.5. The reference frame everything else varies from. The sidebar sits beside it, open by default at this width (§1.7). **Drawn 2026-11-23** ([ADR-037](adr/adr-037-board-at-rest.md) §8): *Q3 Launch Plan*, five columns (To Do longer than the screen, so it reaches the bottom and its *Add card* sticks; In Progress with a due-soon and an overdue card; Review empty; Done; the fifth peeking). The short columns end under their *Add card* row |
| 2.6 | Card drawer over the board | §1.6, at ≥640. **Drawn 2026-11-23** ([ADR-038](adr/adr-038-board-states-group-d.md) §1–§6) |
| 2.7 | **Read-only board** (24) — *join to edit* | Mutating controls are **absent**, plus a persistent, non-dismissible banner **that carries the Join action**. On touch the banner is the only remaining signal — a read-only user presses and holds and nothing happens. Routine browsing, not an edge case. In the sidebar no row is marked current, because the board isn't one you've joined (§1.7). **Drawn 2026-11-23**, the banner under the rail ([ADR-038](adr/adr-038-board-states-group-d.md) §7) |
| 2.8 | **Join-required gate** (25) — *join to look* | A private board an admin hasn't joined: they see it exists, not its contents. One Join action, and joining is recorded as break-glass. As on 2.7, no sidebar row is marked current. **Drawn 2026-11-23 as a centred stack**: a padlock drawing, the sentence and *Join board* ([ADR-038](adr/adr-038-board-states-group-d.md) §8) |
| 2.9 | **Cold load** (19) | Static skeleton blocks at the card geometry, inside real column panels, `color.skeleton` fill, **no shimmer** — shimmer is movement, and movement means a teammate acted. This row first said `--muted` fill, after ADR-018 §7. In Light that is the column's own colour, so the blocks could not be seen ([ADR-029](adr/adr-029-skeleton-block.md)). **Drawn 2026-11-23**: the name and rail at once, panels out to the right edge ([ADR-038](adr/adr-038-board-states-group-d.md) §9) |
| 2.10 | **Empty board** (21) | No columns. New boards seed To Do / In Progress / Done, so this is the deleted-them-all case. It carries an illustration (flavour 4), drawn to [ADR-033](adr/adr-033-workspace-screens.md) §9's rules and measured against the canvas. **Drawn 2026-11-23**, a blank whiteboard ([ADR-038](adr/adr-038-board-states-group-d.md) §10) |
| 2.11 | **Load failed** (22) | Message plus a retry control. **Drawn 2026-11-23**, a centred stack with a plug drawing ([ADR-038](adr/adr-038-board-states-group-d.md) §11) |
| 2.12 | **Gone** (23) | **Archived while you had it open**, or its workspace deleted while you had it open. *Not* "deleted" — there is no delete-a-board action in Phase 1 and no `BOARD_DELETED` event. *Not* "access lost" either: that is state 26, which navigates you out with a toast rather than rendering a screen. **Drawn 2026-11-23 twice**, for its two wordings ([ADR-038](adr/adr-038-board-states-group-d.md) §11) |
| 2.13 | **Session expired** (39) | A 7-day JWT dying mid-session, drawn *before* the hard navigation to `/login`. Without it, expiry looks exactly like a crash. **Drawn 2026-11-23**: Login's left-aligned block with an hourglass. There is no longer a hard navigation: it waits for *Sign in*, then brings you back to the same page ([ADR-038](adr/adr-038-board-states-group-d.md) §13) |
| 2.14 | **Drag** | Four states at once, which is why it earns a board: source at 60% opacity (6), overlay with `--shadow-dragged` and `scale(1.02)` (7), a 2px dashed `--drop-placeholder` (8), and the column **body** tinted `--drag-over` with the header left untinted (9). **Draw two placeholders** — see the note below. **Drawn 2026-11-23 as two boards** ([ADR-039](adr/adr-039-drag-and-phone-board.md) §11): *2.14 Drag — a card*, with *Final pricing sign-off* lifted from In Progress over Done, and *2.14 Drag — a column*, with In Progress lifted to between Review and Done. The scale is not drawn (§4.3) |
| 2.15 | **Phone board, 390×844** | Density is a variable; layout is a breakpoint. Also the only place ADR-019 §7's "6 cards visible" gets checked by eye — the one number in its table resting on an assumption. The sidebar's show/hide button takes the bar's 44px "menu" slot (ADR-019 §8, [ADR-034](adr/adr-034-side-menu.md) §5); the sidebar itself slides over, drawn on *2.15 + sidebar* and two more boards ([ADR-044](adr/adr-044-the-sidebar-sliding-over-and-the-toast.md)). **Drawn 2026-11-23** at Coarse, with the browser's 140px as two labelled bands, the bar without *Add column*, and one 318 column with 44px of the next ([ADR-039](adr/adr-039-drag-and-phone-board.md) §8). **Plus *2.15 Phone — dragging a card***: a card lifted out of To Do and held over the next column's 44px edge (§10) |

**On 2.14, the two things easiest to draw wrong.** The tint covers the column body only: on either tint the header's `--border` rule sits at about 1.19:1 and vanishes, exactly when it is most needed. And the placeholder is `#57534E`, not `--input` — measured against the tint rather than against white, `--input` thins from 4.80:1 to 3.22:1.

**And a third, added 2026-11-19: 2.14 carries _two_ placeholders, not one.** Cards and columns are both draggable, and their drop placeholders are different shapes drawn from the same token:

| Dragging | Placeholder | Where it exists |
|---|---|---|
| A card | 2px dashed `--drop-placeholder` at **card** geometry | Every width |
| A column | 2px dashed `--drop-placeholder` at **column-panel** geometry — full panel height, panel width | **640px and up only** |

Below 640px columns are reordered from the `⋯` menu instead, so the column placeholder has no phone state to draw (ADR-021 §4). Card drag is unchanged at every width.

**And a fourth, added 2026-11-23: a short column is a short target.** Since [ADR-037](adr/adr-037-board-at-rest.md) §4 a column is only as tall as its content. 2.14 must decide whether the drop area and its `--drag-over` tint stop at the panel's end or reach the bottom of the screen. On a phone this decides how tall the next column's 44px peek is as a target. Reaching the bottom keeps short columns easy to hit (ADR-037 *Out of scope*).

> **Settled 2026-11-23 by [ADR-039](adr/adr-039-drag-and-phone-board.md) §1–§2: the drop area stops at the panel.** Nothing changes height during a drag, which was the user's call over stretching every column to the bottom for the drag. The tint covers the cards' area only, from the header's rule to the last card, and *Add card* stays untinted just under it, so a long column's stuck footer keeps its rule as the header does. On a phone the next column's edge is a target only as tall as that column's cards; *2.15 Phone — dragging a card* shows it.

**On 2.15**, the strip is `min(100vw − 72px, 420px)`, and the 72px reservation is not arbitrary: it leaves the next column peeking by exactly **44px**, one tap target, which is what makes it a legal drop target. Below 640px the board does **not** scroll during a drag — valid targets are the current column and that peek.

> **Checked by eye 2026-11-23** ([ADR-039](adr/adr-039-drag-and-phone-board.md) §9): with the browser's 140px drawn, To Do shows **5 whole cards and 44px of the sixth**, against ADR-037's predicted 5 and 56px. The five average 82px against the assumed 80, so the count holds. The column's 318 is typed on this board, since only code can work out `min(100vw − 72px, 420px)` (§8).

**On 2.7 and 2.8 — draw them as one component with two settings, not two screens.** *Narrowed by [ADR-038](adr/adr-038-board-states-group-d.md) §8 (2026-11-23): they share the data rules below and the same Join button, but 2.8 is a centred stack, not the banner over a bare canvas.* They are the same question asked at two visibility levels, and the differences are mechanical:

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
| **Workspace team panel** | The team list: name, email, role, promote / demote, remove, leave. Opened from the "N members" button in §1.4's header. **Settled 2026-11-24** ([ADR-043](adr/adr-043-the-lists.md) §1–§8): group c's 448 dialog and the full-height sheet; each row an Avatar, the name over *Role · email*, and a ⋯ for an admin (*Make admin* or *Make member*, *Remove member*); the owner's row has none; you first, then A–Z; eight rows, then it scrolls, with a filter above eight people. *Leave* is in the workspace ⋯ for everyone but the owner | US-05b, US-06, US-06b, US-06c; ADR-043 |
| Create workspace | Name (**2+ characters**) and an optional description. **Both are permanent** — no rename exists, so this panel is the only chance to get them right. Worth a word of warning in the copy. **Drawn 2026-11-24** ([ADR-042](adr/adr-042-the-forms.md) §1, §7): the form dialog with *You can’t rename a workspace or change its description later.*; the description stops at 1,000 (board *2.3 + create workspace*) | US-03; ADR-042 |
| Invite people | Email batch plus **per-row results** — the mixed batch is the normal case, not the error case. Exactly two failure messages to write: the address has no account, and they're already in the workspace. **Drawn 2026-11-24** ([ADR-042](adr/adr-042-the-forms.md) §12–§13): opened from 2.4's heading; each address a chip, a rose one when it isn't an address; after sending, a row per address under the field, marked like the password checklist, and the field keeps the addresses with no account (boards *2.4 + invite people*, *2.4 + invite people, sent*, *Invite people at 390*) | US-04; ADR-042 |
| Create board | Title (**3+ characters** — the only minimum here that will trip anyone), visibility, colour picker restricted to the eight hues. **Drawn 2026-11-24** ([ADR-042](adr/adr-042-the-forms.md) §6, §8–§10): the minimum shows as you type; eight colour buttons, the first unused hue chosen; *Workspace* and *Private* as two choice rows (boards *2.4 + create board*, *2.4 + create board, a short title*, *Create a board at 390*) | US-07; ADR-042 |
| Board settings | Title, description, colour, visibility. **Drawn 2026-11-24** ([ADR-042](adr/adr-042-the-forms.md) §11): Create a board's form filled in, with *Save changes*; choosing *Private* adds a line saying who loses access (board *2.5 + board settings*) | US-12b; ADR-042 |
| Board members | Add, remove (**which unassigns their cards on this board**), transfer ownership, leave. **Settled 2026-11-24** ([ADR-043](adr/adr-043-the-lists.md) §9–§14): the team list's frame and rows, *Owner · Member* on the second line; for the owner or an admin a ⋯ with *Make owner* and *Remove from board*, the owner's row included for an admin (the board is then owner-less); the owner's sentence says how to leave. *Leave board* stays in the board ⋯ | US-11, US-12d, US-12e, US-10; ADR-043 |
| Add people — candidate list | **Settled by fact, not by choice: the endpoint ships.** It returns *every* workspace member flagged as already-on-the-board, so draw the whole roster with the already-added rows marked and unselectable. **Settled 2026-11-24 as no panel of its own** ([ADR-043](adr/adr-043-the-lists.md) §9–§10): the board list's second group, *Not on the board yet*, for the owner and admins, each person with an *Add* button that adds them alone; you are left out until you join. Flavour 6's line when it is empty | US-11b; ADR-043 |
| Transfer ownership | Target must **already** be a board member; a non-member is refused rather than auto-added. **Settled 2026-11-24 as a confirmation** ([ADR-043](adr/adr-043-the-lists.md) §11): *Make owner* in a person's ⋯ on the board list, then a Primary confirm | US-12e; ADR-043 |
| **Move card to…** | A **column** plus a **neighbour** — top of column, after «card», bottom of column — because position is a float, not an index. Three entry points: card drawer, card `⋯`, and `M` on a focused card. **Drawn 2026-11-23** ([ADR-041](adr/adr-041-the-card-drawers-pickers.md) §1–§4): a column, then a place, and the place moves the card. In the drawer it expands under *Column*; from the card ⋯ and `M` it takes the menu's place, a popover from 640 up and the bottom sheet below (boards *2.6 + moving, a column*, *2.5 + moving from the card ⋯*, *2.6 at 390 + moving, a place* and *2.15 + moving a card*) | ADR-019 §6; ADR-041 |
| Destructive confirmation (38) | Rose confirm. Its unsatisfied state is drawn in `--muted` tokens, **never** in transparency. **Drawn 2026-11-23** ([ADR-040](adr/adr-040-menus-and-the-confirmation.md) §6–§8): the `Confirmation` part, 448 wide (`size.dialog-width`), 24 inside, with five wordings. The confirm is rose for the three deletes and *Leave board*, and Primary for *Archive board*. *Delete workspace* asks you to type its name, and its button stays *not satisfied* until you do | ADR-018 §5 |
| **Workspace `⋯`** | *Delete workspace*, leading to confirmation 38. **The owner's only**: absent for everyone else. Opened from the `⋯` at the end of §1.4's heading row. **Drawn** 2026-11-23 (ADR-040, board *2.4 + workspace ⋯*) | US-06d; ADR-033 §7 |
| **Board `⋯`** | Board settings, Members, Archive board, Leave board — enumerated in §1.5, and shaped by who is looking. **Drawn** 2026-11-23, as a popover and, below 640, as a sheet with *Add column* first (ADR-040, boards *2.5 + board ⋯* and *2.15 + board ⋯*) | US-10, US-11, US-12b, US-12c |
| Column `⋯` and inline rename (37) | Rename, **move left, move right**, delete. **Drawn** 2026-11-23: *Rename* turns the whole header into the Input's *InlineEditing* field, the name where it was (ADR-040 §10, boards *2.5 + column ⋯* and *2.5 + renaming a column*) | US-13b, US-13c, US-13d |
| Card `⋯` | Open, Move, Delete: *Open card*, *Move card to…*, *Delete card*. **Drawn** 2026-11-23 (ADR-040, board *2.5 + card ⋯*) | US-15, US-17 |
| User menu | Name, email, initials — and *Sign out*. **This is the profile surface**; nothing in it is editable, because no endpoint changes any of it. Opened from the app bar on 2.3, and from the bottom of the sidebar on 2.4 and the board. **Drawn** 2026-11-23 from the sidebar (ADR-040 §11, board *2.4 + user menu*) | US-02b; ADR-021; [ADR-034](adr/adr-034-side-menu.md) §5 |
| **Create column** | A column's title (1–255 characters). Opened by the bar's *Add column* from 640 up, and by the board ⋯'s *Add column* below 640. **Added 2026-11-23**: §1.5 named *the create-column form* and this table never listed it. **Settled 2026-11-24 as no overlay at all** ([ADR-042](adr/adr-042-the-forms.md) §14): a new column at the end of the board, its header the rename field, focused; Enter adds it, Escape removes it (boards *2.5 + adding a column*, *2.15 + adding a column*) | UC-13; [ADR-039](adr/adr-039-drag-and-phone-board.md) §7; ADR-040 *Out of scope* |
| **The sidebar, sliding over** | Below 1024: the sidebar (§1.7) over the content from the left, `color.scrim` over the content behind it, `shadow.overlay` on the panel. For focus it is a modal dialog: Escape closes it, and focus returns to the button. **Drawn 2026-11-24** ([ADR-044](adr/adr-044-the-sidebar-sliding-over-and-the-toast.md) §1–§3): on the phone, closed and with the switcher open; focus opens on the current row; below 640 the user menu's bottom sheet opens over it, over a second dimming (boards *2.15 + sidebar*, *2.15 + sidebar, switching workspace*, *2.15 + sidebar, user menu*) | ADR-034 §4, §9, §10.8; ADR-044 |
| Presence overflow sheet | The full list of who is here, faces with names. The whole cluster opens it, not only the `+N` seat: a sheet below 640px, a dialog or drawer above (§3 below). **Settled 2026-11-24 as menu-kind** ([ADR-043](adr/adr-043-the-lists.md) §15–§16): a popover under the cluster from 640 up, a bottom sheet below, headed *8 here now*; you first, then the cluster's order, a face and a name per row | ADR-019 §8; [ADR-031](adr/adr-031-presence-cluster.md) §3; ADR-043 |
| Toast | Carries several states' messages. **Its own design is settled and drawn as a part** ([ADR-030](adr/adr-030-toast.md)): inverted, one appearance, a close button. **Access-revoked has four reasons, and three wordings**: *You were removed from “title”.* · *“title” is now private. Only its members can open it.* · *You were removed from the workspace “name”.* The voluntary case shows nothing. **The wordings for 4, 27 and 32, and where a toast sits, settled 2026-11-24** ([ADR-044](adr/adr-044-the-sidebar-sliding-over-and-the-toast.md)): 16 from the bottom and right, across the bottom on a phone; up to three stacked, newest at the bottom, 8 apart; no timer; failures read *Couldn’t move “X”.* and one of five reasons, and after a 403 or 404 the board reloads; 32 reads *“X” was deleted.* (seven boards: two on the Workspaces page, five on Board). **And one toast with an action:** after you archive the board you're on, the toast offers *Undo* (US-12c), which supersedes ADR-030 §3's *no action button* for that toast only ([ADR-040](adr/adr-040-menus-and-the-confirmation.md) §5). It is an Outline button; it takes you back to the board, and brings teammates back from 2.12 (ADR-044 §8–§11) | ADR-018 *Out of scope*; ADR-030 §5; ADR-044 |
| Date picker, assignee picker | Phosphor Regular 16px throughout. **Not separate layers** — see the rule below. **Drawn 2026-11-23** ([ADR-041](adr/adr-041-the-card-drawers-pickers.md) §5–§11): in place under the field, on the menu surface laid in the flow. The calendar draws six weeks, today semibold, with *Remove due date*; the members start with *Unassigned* and you (boards *2.6 + assignee*, *2.6 + due date*, *2.6 at 390 + due date*) | ADR-018 §9; ADR-041 |

**Two batch behaviours that look alike and are not.** Inviting to a *workspace* is per-row: unknown addresses and existing members are reported individually and the valid invites still land. Adding people to a *board* is **all-or-nothing** — one person already on it fails the entire batch with a single error. So the board panel must **not** be drawn with per-row result rows. It doesn't need them: the candidate list already flags who's on the board, which is what prevents the failure in the first place. **Since 2026-11-24 it adds one person per press** ([ADR-043](adr/adr-043-the-lists.md) §9), so the all-or-nothing batch can no longer fail on someone else's name.

### Presentation on narrow screens — five kinds, one rule each

ADR-019 §3 committed only the card modal. [ADR-021](adr/adr-021-design-inventory-corrections.md) §5 settles the rest in one pass, because the deciding fact is structural rather than per-overlay: **several of these open on top of something that is already a full-height sheet.**

| Kind | Members | Below 640px | 640px and up |
|---|---|---|---|
| **Forms and lists** | Card drawer, create workspace, create board, board settings, board members (with add people), the workspace team panel | **Full-height sheet** | Centred dialog or drawer |
| **Menus** | Workspace `⋯`, board `⋯`, column `⋯`, card `⋯`, user menu, *Move card to…* from the card `⋯` or `M`, a person's `⋯` in a list, and the presence list | **Bottom action sheet** | Anchored popover |
| **Confirmations** (38) | Delete column, delete card, delete workspace, archive board, leave board | **Centred dialog — unchanged at every width** | Centred dialog |
| **Inline pickers** | Due date, assignee, and *Move card to…* from the drawer | **Expand in place inside the card drawer**, across the row | Expand in place, under the value ([ADR-041](adr/adr-041-the-card-drawers-pickers.md) §5) |
| **Navigation** | The sidebar (§1.7) | **Slides over the content from the left**, content dimmed — and this kind switches at **1024**, not 640 | Beside the content **from 1024 up**; slides over it from 640 to 1023 |

**The fifth kind, added 2026-11-21 by [ADR-034](adr/adr-034-side-menu.md) §4.** Its breakpoint is 1024 because beside the content the sidebar would leave a tablet in portrait one column. It never stacks on another layer: it opens only from the app bar's button or its shortcut, and its switcher opens inside it.

**Forms settled on 2026-11-24** ([ADR-042](adr/adr-042-the-forms.md) §1–§3): from 640 up a form is group a's 448 dialog, not a drawer, with no × and nothing happening when you press outside it; below 640 the full-height sheet has *Close* in its head and one full-width button after the fields. *Create column* is no longer an overlay (§14 there).

**Navigation drawn on 2026-11-24** ([ADR-044](adr/adr-044-the-sidebar-sliding-over-and-the-toast.md) §1–§3): the panel on the phone, focus on the current row. Below 640 the user menu opens over it as the bottom sheet, over a second dimming, the one layer allowed over it (ADR-043 §5's rule).

**Lists settled on 2026-11-24** ([ADR-043](adr/adr-043-the-lists.md) §1–§2, §9, §11, §15): from 640 up a list is group c's 448 dialog with *Close*, and it stops at eight rows and scrolls; below 640 it is the full-height sheet. *Add people* and *transfer ownership* are no panels of their own: the second is a confirmation, and the first is the board list's second group. The presence list moved to *Menus*: it has nothing to do, so it is a popover under the cluster and a bottom sheet below.

**Move card to… changed rows on 2026-11-23** ([ADR-041](adr/adr-041-the-card-drawers-pickers.md) §1–§2), superseding ADR-021 §5's placement of it and nothing else in the table. In the card drawer it is an inline picker, expanding under *Column*, so the sheet over a sheet never happens. From the board it is menu-kind: the card ⋯ popover or bottom sheet turns into it in place. Its content is one component wherever it opens.

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

Per [ADR-020](adr/adr-020-design-tool-penpot.md) §4, every state is a variant. The card's `State` axis is `Rest · Hover · Pressed · InFlight · DragSource` — five variants across two densities and two themes is **one** component, not twenty frames. *Since 2026-11-23 it is seven, with `Armed` and `Lifted` ([ADR-039](adr/adr-039-drag-and-phone-board.md) §5).*

| Component | Axis |
|---|---|
| Card | the five above, **plus `Armed`** (state 5: Rest with `shadow.overlay`, where a held press on a touch screen ends) **and `Lifted`** (state 7: Rest with `shadow.dragged`, the card being dragged), since 2026-11-23 ([ADR-039](adr/adr-039-drag-and-phone-board.md) §5). `DragSource` turns the meta row to `color.foreground`, so a dimmed overdue date keeps its words and not its rose. Its meta row packs to its end, so a lone avatar stays on the right and a date fills the width the avatar leaves ([ADR-037](adr/adr-037-board-at-rest.md) §7) |
| Button | `State`: rest, hover, pressed, focus-visible, busy, not-satisfied. **`Kind`: Primary, Outline** (2026-11-21). Outline is `color.background` with a `color.border` edge that steps the ladder on hover and pressed, for *N members*, *Join*, *Unarchive* and *Try again*. Its Busy is Primary's Busy look, and it has no not-satisfied ([ADR-033](adr/adr-033-workspace-screens.md) §6). **`Kind` Destructive** (2026-11-23), the rose confirm: `color.destructive`, hover and pressed `color.destructive-hover`, focus with the plum ring, Busy in pale rose (`color.destructive-subtle`, `color.destructive-subtle-foreground`, the user's call), and Primary's not-satisfied look ([ADR-040](adr/adr-040-menus-and-the-confirmation.md) §8). 17 looks |
| Glyph | `Name` × `Tone`, **twenty-eight looks** since [ADR-043](adr/adr-043-the-lists.md) added `Crown`, `ShieldCheck`, `User`, `UserMinus` Muted and Destructive, and `DotsThree` Muted for a person's ⋯ and the list rows; twenty-two since [ADR-042](adr/adr-042-the-forms.md) added `Check` Done, `Check` Muted and `X` Destructive for the invite results; nineteen since ADR-041 added `X` Muted, `CaretLeft`, `CaretRight` and `MagnifyingGlass`. First fifteen: twelve Muted (`Plus`, `Gear`, `Users`, `Archive`, `DoorOpen`, `PencilSimple`, `ArrowLeft`, `ArrowRight`, `Trash`, `ArrowsOut`, `Kanban`, `SignOut`), `Trash` and `DoorOpen` Destructive, `X` Foreground. A 16 board (`size.icon`) holding Phosphor's Regular outline in `color.muted-foreground`, `color.destructive-text` or `color.foreground`. The first glyph part: the older parts keep their own paths (ADR-040 §14) |
| Menu row | `State`: **Rest**, **Hover**, **Pressed**, **FocusVisible** × `Tone`: **Default**, **Destructive**. The Sidebar row's geometry filled `color.popover`: a 16 `Slot` holding a `Glyph`, the label in `type.body`, `color.foreground` or `color.destructive-text`. No edge at rest; hover and pressed step the edge; focus adds `shadow.focus-ring` (ADR-040 §2) |
| Menu | `Of`: **Workspace**, **Board**, **Column**, **Card**, **User** × `Presentation`: **Popover**, **Sheet**. Popover: `size.menu-width` 224, `color.popover`, a `color.border` edge outside its box, `shadow.overlay`, `radius.surface`, and a `Panel` with `spacing.menu-padding` inside and rows `spacing.list-gap` apart. Sheet: 390 wide, top corners only, the `Panel` carrying `shadow.rule-top` and a `Head` (the name in `color.muted-foreground` and a Close). The User looks start with your face, name and email (ADR-040 §2, §9, §11). **Sixteen looks since 2026-11-24**: `Of` gains **Teammate** (*Make admin*, *Remove member*), **Board member** (*Make owner*, *Remove from board*) and **Workspace member** (*Leave workspace*), each Popover and Sheet; their sheets' heads show the person's face and name ([ADR-043](adr/adr-043-the-lists.md) §5, §7) |
| Confirmation | — . **One part, five wordings** (§3): `size.dialog-width`, `spacing.dialog-padding` inside, `spacing.section-gap` between blocks, the menu's surface. `Text` (title in `type.title`, sentence in `type.body` and `color.muted-foreground`, `spacing.field-inner-gap` apart), a hidden `Field` (an Input), and `Actions` (an Outline *Cancel* and the confirm, `spacing.inline-gap` apart, at the right). Its board shows *Delete column*, *Delete card*, *Delete workspace* empty and typed, *Archive board* and *Leave board* (ADR-040 §7). **Six more wordings since 2026-11-24**, in a second row: *Remove member*, *Remove from board*, *Remove owner* (all rose), *Make owner* and *Stop being an admin* (Primary), and *Leave workspace* (rose) ([ADR-043](adr/adr-043-the-lists.md) §19) |
| Picker row | `State`: **Rest**, **Hover**, **Pressed**, **FocusVisible**, **Current**. A picker's rows: the Menu row's geometry, a 16 `Slot` holding a `Glyph`, a hidden Avatar, a `Label` (the name, and a hidden *(you)* in `color.muted-foreground`, `spacing.inline-gap` apart), and a hidden Count chip at the end. Current is plum text and a 1px plum edge, as the sidebar's current row ([ADR-041](adr/adr-041-the-card-drawers-pickers.md) §7, §16) |
| Day | `State`: **Rest**, **Hover**, **Pressed**, **FocusVisible**, **Current** × `Today`: **No**, **Yes** (Rest and Current only), seven looks. A `size.control-height` square, `radius.control`, the number in `type.body`; the edge ladder; Current plum; Today semibold ([ADR-041](adr/adr-041-the-card-drawers-pickers.md) §9–§10) |
| Calendar | — . **One appearance**: November 2026, Monday first, six week rows of `Day` copies and blanks, the month's head with `CaretLeft` and `CaretRight`, the weekday initials in `type.meta` ([ADR-041](adr/adr-041-the-card-drawers-pickers.md) §9) |
| Picker | `Of`: **Assignee**, **Assignee filtered**, **Due date**, **Column**, **Place** × `Presentation`: **Inline**, **Popover**, **Sheet**, nine looks. Inline is the menu surface in the flow, 432 wide (Due date hugs its grid); Popover and Sheet are the Menu's, holding `Picker row` copies, with *Move to* over the columns ([ADR-041](adr/adr-041-the-card-drawers-pickers.md) §5–§6, §16) |
| Input | `State`: rest, focus, invalid, inline-editing × **`Content`: Text, Lines, Chips** (2026-11-24, [ADR-042](adr/adr-042-the-forms.md) §17). Lines (Rest, Focus) grows from one line, `spacing.control-padding-y` above and below, with *(optional)* in `color.muted-foreground` beside its label; Chips (Rest, Focus, Invalid) wraps `Address chip` copies `spacing.list-gap` apart. Nine looks |
| Swatch | `State`: **Rest**, **Hover**, **Pressed**, **FocusVisible**, **Current**. A colour button: a `size.control-height` square, `radius.control`, `color.popover`, padded `spacing.inline-gap`, holding a `Colour` square in the hue's `.strong` (20 at Fine, 28 at Coarse); the edge ladder; Current is a plum edge ([ADR-042](adr/adr-042-the-forms.md) §8) |
| Choice row | `State`: **Rest**, **Hover**, **Pressed**, **FocusVisible**, **Current**. A radio row 400 wide: `spacing.control-padding-y` / `-x`, a 16 circle and the name in a `Head`, the sentence in `type.meta` under it; Current is plum text, edge and circle with a `size.dot` dot ([ADR-042](adr/adr-042-the-forms.md) §10) |
| Address chip | `State`: **Rest**, **Invalid**. `size.target-min` tall, `color.chip`, `radius.badge`, the address in `type.body`, a `size.target-min` square holding an `X`; Invalid adds a `color.destructive` edge ([ADR-042](adr/adr-042-the-forms.md) §12) |
| Invite result | `Outcome`: **Added**, **NoAccount**, **AlreadyIn**. The glyph and the address, then the outcome in `type.meta`: an emerald check, a rose ×, a grey check ([ADR-042](adr/adr-042-the-forms.md) §13) |
| Form | `Form`: **Workspace**, **Board**, **Settings**, **Invite**, **Invite sent** × `Frame`: **Dialog**, **Sheet** (Board and Invite only), seven looks. The 448 dialog or the 390 sheet, holding `Input`, `Swatch`, `Choice row`, `Invite result` and `Button` copies and a hidden `Form message` ([ADR-042](adr/adr-042-the-forms.md) §1–§2, §17) |
| Person row | `Trail`: **Menu**, **Add**, **Adding**, **None**. A row `size.control-height` tall and `size.form-width` wide, `spacing.control-padding-x` on its left, `spacing.icon-gap`, `radius.control`, `color.popover`: an Avatar; `Who`, a `Label` (the name in `type.body`, then a hidden *(you)* in `color.muted-foreground`, `spacing.inline-gap` apart) over `Sub` (*Member · dimas@example.com* in `type.meta`, `color.muted-foreground`); then a `size.target-min` `Menu` holding `DotsThree`, or an Outline *Add* / *Adding…*. None has `spacing.control-padding-x` on its right too. No hover or pressed look: the row is not a button ([ADR-043](adr/adr-043-the-lists.md) §3–§4, §20) |
| List | `Of`: **Team**, **Members**, **Here** × `Frame`: **Dialog**, **Sheet**, **Popover** (Here only), six looks. Dialog: the `Form` part's 448 frame holding `Text`, a `Body` (the filter `Input`, then the rows, eight tall before they scroll) and `Actions` with *Close*. Sheet: 390, `spacing.screen-x` inside, a `Head` with *Close*, no `Actions`. Members draws its two groups, *On the board* and *Not on the board yet*, the second with *Add*. Here: the Menu's surface, a *8 here now* label in `type.meta`, and eight `Person row` copies with `Trail` None and no second line ([ADR-043](adr/adr-043-the-lists.md) §1–§2, §9, §15, §20). **Since 2026-11-24** both Members looks hold a hidden `Empty` line after their second group, flavour 6, which a copy shows in place of *Not on the board yet* |
| Avatar | eight hues, **plus Muted** (initials, for the presence cluster's reconnecting state) **and Unknown** (the no-name fallback — initials from an empty string render an empty circle, §1.5). The two role assignments are not an axis: the hues are tokens that follow the Theme switch ([ADR-025](adr/adr-025-identity-hues-as-tokens.md)). Muted and Unknown sit on `color.avatar-neutral`, because `color.muted` is 1.09:1 on the white app bar ([ADR-031](adr/adr-031-presence-cluster.md)) |
| Column header | — . **One appearance.** Drag-over leaves the header **untinted** (§2.14), so it changes nothing on the part. The part's board shows the header twice instead, above an untinted body and above a `color.drag-over` one, so the rule can be seen against the tint ([ADR-027](adr/adr-027-column-header.md) §5) |
| Banner | — . **One appearance, and it carries Join** (§2, board 2.7). This row first listed an *informational* banner as well. Nobody can see one: the read-only board's only visitors can all join, and no other state is a banner ([ADR-028](adr/adr-028-banner.md) §4). Join's own looks, *Joining…* included, are the Button's |
| Count chip | — . One appearance that holds any text: `12` in a column header, `12 cards` and `3 members` on a board tile, because two bare numbers on one tile cannot be told apart ([ADR-026](adr/adr-026-count-chip.md) §4) |
| Skeleton block | — . **One appearance**: a flat `color.skeleton` block at a card's corners and width, `size.skeleton-height` tall, with no outline. It sits on a column panel or on the page background, **never directly on the canvas**, where in Light it disappears ([ADR-029](adr/adr-029-skeleton-block.md) §4) |
| Toast | — . **One appearance**: an inverted box (`color.toast`, text and close glyph `color.toast-foreground`) with no edge, `size.toast-width` wide, `spacing.toast-padding` inside, a message that wraps and a Phosphor `X` close shown at rest. Every Phase-1 toast says one thing and offers nothing but closing, so there are no kinds and no action button ([ADR-030](adr/adr-030-toast.md) §3). **Since 2026-11-24 a hidden `Undo`** between the message and the close, an Outline `Button` copy that the archive toast's copies show; still no axis. Its board also shows a stack of three, 8 apart, and *Undoing…* ([ADR-044](adr/adr-044-the-sidebar-sliding-over-and-the-toast.md) §14) |
| Presence cluster | `State`: **Live**, **Reconnecting**, **FocusVisible**. Reconnecting turns the faces to the Avatar's Muted look on `color.avatar-neutral`, with a "Reconnecting…" label beside them, spending no hue. The whole cluster is one target, `size.target-min` tall; `+N` is a round seat in the stack. No hover or pressed look, since the cluster has no edge to step ([ADR-031](adr/adr-031-presence-cluster.md) §6) |
| Form message | — . **One appearance**: one line of `type.body` in `color.destructive-text`, filling the form’s width, inside the fields stack just above the button. A form-level failure only — a field’s own error stays on the field. Its board shows Login straight after a 401 ([ADR-032](adr/adr-032-auth-form-errors-and-password-rules.md) §1) |
| Tile | `State`: **Rest**, **Hover**, **Pressed**, **FocusVisible**. Padding `spacing.tile-padding`, rows `spacing.tile-row-gap` apart: title (and a hidden lock), counts (role and two Count chips), and an action row (a hidden Outline button). A left rail in the hue's `.strong`, `size.board-rail` wide; hidden on a workspace tile. Hover and pressed step the edge ([ADR-033](adr/adr-033-workspace-screens.md) §3) |
| Segment | `State`: **Current**, **Other**, **Hover**, **Pressed**, **FocusVisible**. The Active / Archived switch's two options. Current is plum text and a 1px plum edge; the others are muted text that steps the edge ladder (ADR-033 §5) |
| Illustration | `Subject`: **NoWorkspaces**, **NoBoards**, **NothingArchived**, and since 2026-11-23 **NoColumns**, **PrivateBoard**, **LoadFailed**, **BoardGone**, **SessionEnded** (ADR-038 §14). 128 × 128 (`size.illustration`), a 2px `color.illustration-ink` line, `color.illustration-tint` and `color.background` fills, drawn as paths. A new drawing is a new `Subject`, following ADR-033 §9's rules |
| Sidebar row | `State`: **Rest**, **Hover**, **Pressed**, **FocusVisible**, **Current** ([ADR-034](adr/adr-034-side-menu.md) §10.3). One part draws every row in the sidebar ([ADR-035](adr/adr-035-sidebar-loading-failure-and-switcher-mark.md) §7):<br>• the geometry is `size.control-height`, `spacing.control-padding-x`, `radius.control`, `type.body` and a `color.background` fill<br>• a 16px slot holds the board's square (`size.swatch`), and hides the `SquaresFour` glyph, the `ArrowClockwise` glyph and the dot (`size.dot`)<br>• a hidden Avatar for the user row, the title with a hidden lock after it, and a hidden `CaretUpDown` for the switcher<br>• Current is plum text, glyphs and a 1px plum edge, and the others step the edge ladder |
| Sidebar | `Switcher`: **Closed**, **Open** × `List`: **Rows**, **Loading**, **Failed**, **Empty**. Five variants, because Open is drawn with Rows only. The whole column:<br>• the top row, the switcher, *All boards*, *Your boards* and the rows, and the user row under a rule<br>• `size.sidebar-width` × 900, `spacing.sidebar-padding` inside, on `color.background` with a right-only rule<br>• its board also shows 2.4's app bar with the sidebar hidden (ADR-035 §7) |
| Add card | `State`: **Rest**, **Hover**, **Pressed**, **FocusVisible**. Phosphor `Plus` at `size.icon` and *Add card* in `type.body`, both `color.muted-foreground`; `size.control-height`, `spacing.control-padding-x`, `spacing.icon-gap`, `radius.control`, filled `color.board-column`. No edge at rest; hover and pressed step the edge, focus adds `shadow.focus-ring`. The rule it shows while stuck sits on a `Footer` around it, since the row already binds a shadow ([ADR-037](adr/adr-037-board-at-rest.md) §4) |
| Field | `State`: **Rest**, **Hover**, **Pressed**, **FocusVisible**. The card drawer's value rows. The Sidebar row's geometry filled `color.popover`: a 16 `Slot` holding Phosphor `Kanban` or `CalendarBlank`, a hidden Avatar, the value in `type.body`, and a `CaretUpDown`, shown on *Column* only until [ADR-041](adr/adr-041-the-card-drawers-pickers.md) §8 showed it on all three. No edge at rest; hover and pressed step the edge, focus adds `shadow.focus-ring` ([ADR-038](adr/adr-038-board-states-group-d.md) §16) |
| Password rule | `State`: **Unmet**, **Met**, **Missed** — Phosphor `Circle` / `Check` / `X` at `size.icon`, the rule in `type.meta`, both in `color.muted-foreground` / `color.done` / `color.destructive-text`. Rows sit `spacing.list-gap` apart, under the field’s box at `spacing.field-inner-gap`. The glyph’s shape carries the state, because emerald and rose read alike to a deuteranope ([ADR-032](adr/adr-032-auth-form-errors-and-password-rules.md) §4) |

Hover, pressed and the border ladder are three consecutive stone steps on the **edge**, never a fill: shadcn hovers to `--accent`, which is `--board-column`'s own value, so a card hovering to accent dissolves into the panel behind it.

### 4.3 What cannot be drawn

**Motion.** Penpot has no duration or easing token type, and the channel's meaning — "a teammate did this" — is a multi-client behaviour no static file can show. It stays code-only: 120ms chrome, 200ms board objects, 260ms containers, 110ms reduced-motion cross-fade.

**Equal-width digits.** ADR-017 §6 puts `tabular-nums` on dates and counts. Penpot has no token type for it and no setting for a font's alternate figures, so the file shows Poppins's default digits, which are proportional: a `1` is 4px wide and an `8` is 8px ([ADR-026](adr/adr-026-count-chip.md) §5). A count or a date with a `1`, `2` or `7` in it can therefore draw slightly narrower in the file than in the app.

**Growth.** A lifted card is 2% bigger (`scale(1.02)`) and a held press 1% (`scale(1.01)`, ADR-018 §4). Penpot cannot scale a shape together with its text and has no token type for it, so the `Card` part's `Lifted` and `Armed` looks are drawn at the card's real size, 2% and 1% smaller than the app draws them ([ADR-039](adr/adr-039-drag-and-phone-board.md) §6). Stretching the box instead would draw a wider card with the same 14px words, which the app never shows.

**Who gets which hue.** In code the palette is a lookup indexed by `hash(id) % 8`, living in a TypeScript constant with both role assignments (ADR-017 §9). The file cannot show that assignment; each drawn avatar simply picks a hue. The colours themselves *are* tokens in the file since [ADR-025](adr/adr-025-identity-hues-as-tokens.md), so they follow the Theme switch.

---

## 5. What the file totals

| | Count |
|---|---|
| Whole-screen boards | **15** (§2), all drawn by 2026-11-23. Some carry extra boards for their states or wordings: 2.3 and 2.4's state boards, 2.12's two wordings, 2.14 as a card drag and a column drag, and 2.15's drag board ([ADR-039](adr/adr-039-drag-and-phone-board.md)) |
| Overlay variant sets | ~**19** (§3) — the workspace `⋯` added 2026-11-21, the sidebar sliding over ([ADR-034](adr/adr-034-side-menu.md)) the same day, and *Create column* 2026-11-23. **Group a drawn** 2026-11-23: the five menus and the confirmation, as the `Menu` and `Confirmation` parts and nine boards on the `Overlays` page ([ADR-040](adr/adr-040-menus-and-the-confirmation.md)). **Group b drawn** later the same day: the three pickers, as the `Picker` part and eight more boards ([ADR-041](adr/adr-041-the-card-drawers-pickers.md)). **Group c drawn** 2026-11-24: four forms, as the `Form` part and ten more boards; *Create column* left the overlays, since it is a new column on the board ([ADR-042](adr/adr-042-the-forms.md)). **Group d drawn** 2026-11-24: the team, a board's members and who's here, as the `Person row` and `List` parts and ten more boards; *Add people* and *Transfer ownership* left the overlays, since one is the members list's second group and the other a confirmation ([ADR-043](adr/adr-043-the-lists.md)). **Group e drawn** 2026-11-24: the sidebar sliding over, and the toast on each kind of screen, ten more boards; **Stage 5 is closed** ([ADR-044](adr/adr-044-the-sidebar-sliding-over-and-the-toast.md)) |
| Foundations board | **1** |
| Component variant containers | **35** so far (§4.2): `Logo`, on the `Brand` page, the latest ([ADR-049](adr/adr-049-the-logo.md), 2026-11-28). Before it, 34, with `Person row` and `List` from Stage 5's group d ([ADR-043](adr/adr-043-the-lists.md)). Before them, 32, with `Swatch`, `Choice row`, `Address chip`, `Invite result` and `Form` from group c ([ADR-042](adr/adr-042-the-forms.md)). Before those, 27, with `Picker row`, `Day`, `Calendar` and `Picker` from group b ([ADR-041](adr/adr-041-the-card-drawers-pickers.md)). Before those, 23, with `Glyph`, `Menu row`, `Menu` and `Confirmation` from group a ([ADR-040](adr/adr-040-menus-and-the-confirmation.md)). Before those, 19, `Field` from 2.6 ([ADR-038](adr/adr-038-board-states-group-d.md)) the latest: ten through Stage 3, `Form message` and `Password rule` from 2.2, `Tile`, `Segment` and `Illustration` from 2.3 and 2.4, `Sidebar row` and `Sidebar` from the 2.4 redraw ([ADR-035](adr/adr-035-sidebar-loading-failure-and-switcher-mark.md)), and `Add card` from 2.5 ([ADR-037](adr/adr-037-board-at-rest.md)) |

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
| 5 | No cards in a column | In-column surface: *No cards yet.*, above *Add card*. Drawn on 2.5's Review column ([ADR-037](adr/adr-037-board-at-rest.md) §6) |
| 6 | Everyone in this workspace is already on this board | The board members list's second group, which replaced the add-people panel (§3, [ADR-043](adr/adr-043-the-lists.md) §9): *Everyone in Product Team is already on this board.* in place of *Not on the board yet*. Drawn 2026-11-24 as a copy on `List — states`, after Stage 5's closing check found it drawn nowhere |
| 7 | **No joined boards in this workspace** | The sidebar's board list (§1.7) — new, from [ADR-034](adr/adr-034-side-menu.md) §8. Drawn as the `Sidebar` part's `Empty` variant, and on board 2.4 — empty |

> **2026-11-21: flavours 1 to 4 carry an illustration** above their sentence, centred in the place of what is missing. Flavours 5 and 6 keep one line: a 128 drawing does not fit a 280 column or a panel list. The drawing rules are in [ADR-033](adr/adr-033-workspace-screens.md) §9.

> **2026-11-21: a seventh.** The sidebar lists only the boards you've joined, so someone who has joined none sees *You haven’t joined a board yet.* It keeps one line, with no illustration, like flavours 5 and 6: *All boards* just above it is the way forward ([ADR-034](adr/adr-034-side-menu.md) §8).

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
| The five token sets | **Done** — `base` 52, `theme-light` 70, `theme-dark` 70, `density-fine` 18, `density-coarse` 18. **228 tokens** since 2026-11-28, when [ADR-049](adr/adr-049-the-logo.md) added `color.logo.still`, `color.logo.moving` and `color.logo.name` (stored as `{color.logo.still}`) once per theme set, for the logo. **222 tokens** since 2026-11-24, when [ADR-042](adr/adr-042-the-forms.md) added `spacing.control-padding-y`, 8 and 12, once per density set, for a field that grows with its text. [ADR-041](adr/adr-041-the-card-drawers-pickers.md) (later still on 2026-11-23) added none. **220 tokens** since later on 2026-11-23, when [ADR-040](adr/adr-040-menus-and-the-confirmation.md) added `spacing.menu-padding`, `size.menu-width`, `spacing.dialog-padding` and `size.dialog-width` (stored as `{size.form-width} + {spacing.dialog-padding} * 2`) in `base`, and `color.destructive-hover`, `color.destructive-subtle` and `color.destructive-subtle-foreground` once per theme set. Before that, **210 tokens** since 2026-11-23, when [ADR-038](adr/adr-038-board-states-group-d.md) added `size.drawer-width` and `size.field-label` in `base` and `shadow.rule-left` once per theme set. Before that, **206 tokens** since 2026-11-22, when [ADR-036](adr/adr-036-one-sided-rules-as-inset-shadows.md) added `shadow.rule-top`, `shadow.rule-right` and `shadow.rule-bottom` once per theme set. They draw every one-sided rule, because Penpot never draws a per-side stroke width. Before that, **200**, since 2026-11-21. [ADR-034](adr/adr-034-side-menu.md) added five entries: `size.sidebar-width`, `spacing.sidebar-padding` and `size.swatch` in `base`, and `color.scrim` once per theme set, stored as `rgba()` because Penpot refuses the space-separated form. [ADR-035](adr/adr-035-sidebar-loading-failure-and-switcher-mark.md) added one, `size.dot` in `base`. Before that, **194**, when [ADR-033](adr/adr-033-workspace-screens.md) added fifteen entries: `size.app-bar`, `size.tile-min-width`, `spacing.tile-padding`, `spacing.tile-row-gap`, `size.illustration`, `spacing.illustration-gap` and `border.width-illustration` in `base`; `color.illustration-ink` and `color.illustration-tint` once per theme set; `spacing.tile-gap` and `size.tile-skeleton-height` once per density set. Before that, **179**, since [ADR-032](adr/adr-032-auth-form-errors-and-password-rules.md) added `spacing.list-gap` 4 to `base`. Before that, **178 tokens**: 103 derived from ADR-017 §9 in full, plus the four form-control values from [ADR-022](adr/adr-022-form-control-geometry.md), the six spacing and border values from [ADR-023](adr/adr-023-spacing-between-objects.md) (the column gap written once per density set), the four from [ADR-024](adr/adr-024-control-and-field-internals.md) (the focus ring written once per theme set), the 16 identity tokens from [ADR-025](adr/adr-025-identity-hues-as-tokens.md) (written once per theme set), the three count-chip tokens from [ADR-026](adr/adr-026-count-chip.md) (`color.chip` written once per theme set), `spacing.inline-gap` from [ADR-027](adr/adr-027-column-header.md), `spacing.banner-padding-y` from [ADR-028](adr/adr-028-banner.md), the four skeleton entries from [ADR-029](adr/adr-029-skeleton-block.md) (`color.skeleton` once per theme set, `size.skeleton-height` once per density set), and the six toast entries from [ADR-030](adr/adr-030-toast.md) (`color.toast` and `color.toast-foreground` once per theme set, `spacing.toast-padding` and `size.toast-width` in `base`), and the nine presence entries from [ADR-031](adr/adr-031-presence-cluster.md) (`spacing.avatar-overlap` and `spacing.seat-padding-x` once per density set, `shadow.avatar-ring` and `color.avatar-neutral` once per theme set, `radius.full` in `base`). ADR-025 also renamed `size.card-avatar` to `size.avatar`. Every addition was flip-verified on 2026-11-20, and Stage 2 has tested the catalogue against a real screen |
| The four themes, in two groups | **Done** — `Theme`: Light, Dark. `Density`: Fine, Coarse. |
| Card variant container | **Done**, on page `Components` — all five appearances, with its fill, outline, four corners and four paddings all bound. **2026-11-20:** outline width bound to `border.width`, and the avatar corrected. It had been bound to `size.title-reserve` for width (24 → 32) with a typed 20px height, which made an oval; it is now `size.card-avatar` on both axes (20 → 24). The binding check cannot see sizes, which is why it never caught this. **Later on 2026-11-20** the placeholder was replaced by a real Avatar instance in all five variants ([ADR-025](adr/adr-025-identity-hues-as-tokens.md) §5); heights unchanged at 80 / 92. **2026-11-23:** the meta row packs to its end, the date fills the width the avatar leaves, and `spacing.inline-gap` sits between them, so a card with an assignee and no date keeps its avatar on the right ([ADR-037](adr/adr-037-board-at-rest.md) §7); heights unchanged. **Later on 2026-11-23:** `Armed` (`shadow.overlay`) and `Lifted` (`shadow.dragged`) added at 1660 and 1980, copies of Rest, with their labels; the part holds seven looks ([ADR-039](adr/adr-039-drag-and-phone-board.md) §5) |
| Input variant container | **Done** 2026-11-20, on page `Components`, board `Input — states`. Four appearances: `Rest` (placeholder), `Focus` (the ring), `Invalid` (destructive edge and a message below), `InlineEditing` (the 20/28 title in the same box, no label). A field is the whole field: label, box and message, so a screen needs nothing loose around it. **2026-11-24** a second axis, `Content`, with five looks, 480 apart after the first four: `Rest` and `Focus` × `Lines`, and `Rest`, `Focus` and `Invalid` × `Chips`; the four it had are `Text`. Lines' `Box` sizes itself to its text, at least `size.control-height`, `spacing.control-padding-y` above and below, and its label is a `Label row` with *(optional)*. Chips' `Box` wraps four `Address chip` copies, `spacing.list-gap` from its edge and apart; Invalid keeps the `color.input` edge and turns one chip rose. The board is 4360 wide ([ADR-042](adr/adr-042-the-forms.md) §17) |
| Button variant container | **Done** 2026-11-20, board `Button — states`. Six appearances: `Rest`, `Hover`, `Pressed` (the same fill as `Hover`, by decision: ADR-024 §3), `FocusVisible`, `Busy` (Phosphor `CircleNotch` at 16 and *Signing in…*), `NotSatisfied`. The sample label reads *Sign in* since 2026-11-21 ([ADR-032](adr/adr-032-auth-form-errors-and-password-rules.md) §7). **Primary only**: other kinds become a second axis when a screen needs one. **2026-11-21:** that axis exists. `Kind` = Primary · Outline; five Outline variants (Rest, Hover, Pressed, FocusVisible, Busy) sit on a second row, labelled *Join* and *Joining…* ([ADR-033](adr/adr-033-workspace-screens.md) §6). **2026-11-23:** `Kind` Destructive, six looks on a third row at y 1010, labelled *Delete* and *Deleting…* ([ADR-040](adr/adr-040-menus-and-the-confirmation.md) §8). The board was widened to 1265 and every board below it moved down 130 |
| Avatar variant container | **Done** 2026-11-20, board `Avatar — states`. One axis, `Hue`, ten appearances: the eight hues (fill `color.identity.<hue>.subtle`, initials `.strong`), `Muted` (initials on `color.muted`), `Unknown` (Phosphor `User` on `color.muted`). Sized by `size.avatar` with an ellipse behind, so no radius to bind. Flips with Theme by itself. **Later on 2026-11-20** the Muted and Unknown circles were rebound to `color.avatar-neutral` ([ADR-031](adr/adr-031-presence-cluster.md)), and every instance followed |
| Count chip | **Done** 2026-11-20, board `Count chip — states`. **One component, no axis**: fill `color.chip`, height and minimum width `size.chip-height`, side padding `spacing.chip-padding-x`, corners `radius.badge`, text `type.meta` in `color.muted-foreground`; its width follows the text. Shown with `3`, `12` and `128` on a column-coloured surface and `12 cards` and `3 members` on a tile-coloured one. **Penpot does not draw the minimum width** on a board that follows its content, so a `1` chip draws 16 wide here and 20 in the app ([ADR-026](adr/adr-026-count-chip.md) *Operational note*). **Known fault, left on purpose (2026-11-27, [ADR-046](adr/adr-046-stage-6-final-check.md) §3):** the three copies *128*, *12 cards* and *3 members* draw clipped at 23 wide, and five repairs did not hold. The chips on the real screens are correct, so **do not read the cut-off as the design** |
| Column header | **Done** 2026-11-20, board `Column header — states`. **One component, no axis**: height `size.column-header`, fill `color.board-column`, a bottom-only rule (`color.border`, `border.width` on `strokeWidthBottom`), side padding `spacing.card-padding-x`, gap `spacing.inline-gap`. Inside: `Lead` (title in `type.body` and `color.foreground`, then a `Count chip` instance) and `Menu` (`size.target-min`, centring Phosphor `DotsThree` at `size.icon` in `color.muted-foreground`). Shown as `In Progress` / `12` on two bound column panels, one above an untinted body and one above a `color.drag-over` body. **No export shows the rule**: Penpot's exporter drops a stroke whose sides differ, so check it on the canvas ([ADR-027](adr/adr-027-column-header.md) *Operational note*) |
| Banner | **Done** 2026-11-20, board `Banner — states`. **One component, no axis**: fill `color.background`, a bottom-only rule (`color.border`, `border.width`), side padding `spacing.screen-x`, top and bottom `spacing.banner-padding-y`, gap `spacing.inline-gap`, height following the content (52 at Fine, 60 at Coarse). Inside: `Message` (*You’re viewing this board. Join it to add, edit and move cards.*, `type.body` in `color.foreground`, fills the row and wraps) and `Join` (a `Button` instance, `Rest`, labelled *Join board*). Shown at 1024 on one line and at 390 wrapped to two, each above a `Doc / Canvas` strip: in Dark the banner and the canvas are the same colour, and the rule is the only edge. No informational appearance ([ADR-028](adr/adr-028-banner.md) §4) |
| Skeleton block | **Done** 2026-11-20, board `Skeleton block — states`. **One component, no axis**: a board with no layout, fill `color.skeleton`, four corners `radius.surface`, height `size.skeleton-height` (60 at Fine, 72 at Coarse), width filling its column, no outline. Shown as the main and two instances in one `Doc / Surface / Column` (bound to `color.board-column`, `size.column-width`, `spacing.column-padding`, `radius.surface`, row gap `spacing.card-gap`) on a `Doc / Canvas` bound to `color.board-canvas`, so the Light check shows canvas, column and block together. ADR-018 §7's `--muted` fill is withdrawn: in Light it is the column's own colour ([ADR-029](adr/adr-029-skeleton-block.md)) |
| Toast | **Done** 2026-11-20, board `Toast — states`. **One component, no axis**: a flex row, fill `color.toast`, no edge, four corners `radius.surface`, `shadow.overlay`, padding `spacing.toast-padding` on all sides, gap `spacing.inline-gap`, width `size.toast-width`, height following the content (48 for one line and 64 for two at Fine, 68 at Coarse). Inside: `Message` (`type.body` in `color.toast-foreground`, fills the row and wraps) and `Close` (`size.target-min`, centring Phosphor `X` at `size.icon` in `color.toast-foreground`). The main holds the longest wording, *“Q3 Product Launch Plan” is now private. Only its members can open it.*, floating over three Card instances in a bound column on a 390 `Doc / Canvas`. One instance, with the shortest wording, sits on a `Doc / Page` bound to `color.background` ([ADR-030](adr/adr-030-toast.md)). **2026-11-24:** the main gains a hidden `Undo`, an Outline `Button` copy labelled *Undo*, between `Message` and `Close`; the board is widened to 1710 with `Doc / Page, 390 — a stack of three` (8 apart, newest at the bottom) and `Doc / Page, 390 — Undoing` (the Outline's Busy look); its three Card copies' `Meta row` gaps, typed 8 since ADR-037 §7, are bound to `spacing.inline-gap` ([ADR-044](adr/adr-044-the-sidebar-sliding-over-and-the-toast.md) §14) |
| Presence cluster | **Done** 2026-11-20, board `Presence cluster — states`. One axis, `State`: `Live`, `Reconnecting`, `FocusVisible`. The box is a flex row: height `size.target-min`, gap `spacing.inline-gap`, fill `color.background`, four corners `radius.control`, and `shadow.focus-ring` on FocusVisible. Inside, `Faces` holds five Avatar instances (AR sage, MK periwinkle, JT ochre, SL mauve, DN teal; Muted in Reconnecting), each with `shadow.avatar-ring`, overlapping by `spacing.avatar-overlap` on their left margins. The `Seat` (`+3`) overlaps the same way: `size.avatar` tall and as its minimum width, `spacing.seat-padding-x` at the sides, `color.avatar-neutral`, `radius.full`, the ring. Reconnecting adds `Label` (*Reconnecting…*, `type.meta` in `color.muted-foreground`). The board shows the variant container, two 375 app-bar strips (Live with three faces and `+5`, and Reconnecting with the seat hidden) and two 1024 strips (five faces, live and reconnecting), each on a `Doc / Canvas`. The strips' bottom rule is per-side, so no export shows it ([ADR-031](adr/adr-031-presence-cluster.md)) |
| Form message | **Done** 2026-11-21, board `Form message — states`. **One component, no axis**: a flex column, width `size.form-width` (filling the form), holding one `Message` in `type.body` and `color.destructive-text`. The main sits inside `Doc / Page, Login after a 401`, between Login's filled Email and Password and a `Rest` Sign in button: Login straight after a wrong email or password ([ADR-032](adr/adr-032-auth-form-errors-and-password-rules.md) §1) |
| Tile | **Done** 2026-11-21, board `Tile — states`. One axis, `State`: `Rest`, `Hover`, `Pressed`, `FocusVisible`. A flex column, padding `spacing.tile-padding`, row gap `spacing.tile-row-gap`, width and minimum width `size.tile-min-width`, fill `color.card`, a `color.border` edge, `radius.surface`, clipped. Inside: `Title row` (`Title` in `type.card-title`, a hidden filled `Lock` at `size.icon`, `spacing.icon-gap`), `Meta row` (`size.meta-row` tall, `spacing.inline-gap`: `Role` in `type.meta`, two Count chip instances *12 cards* and *3 members*), a hidden `Action row` holding an Outline Button, and an absolute `Rail` (`size.board-rail` wide, periwinkle `.strong`, pinned top and bottom so it follows the tile's height) ([ADR-033](adr/adr-033-workspace-screens.md) §3) |
| Segment | **Done** 2026-11-21, board `Segment — states`. One axis, `State`: `Current`, `Other`, `Hover`, `Pressed`, `FocusVisible`. A flex row: `size.control-height`, `spacing.control-padding-x`, `radius.control`, `color.background`; Current has a `color.primary` edge and label, the others a `color.muted-foreground` label and an edge only on Hover and Pressed (ADR-033 §5) |
| Illustration | **Done** 2026-11-21, board `Illustration — states`. One axis, `Subject`: `NoWorkspaces` (a sprout in a pot), `NoBoards` (an empty board), `NothingArchived` (an open, empty box). **2026-11-23:** `NoColumns` (a blank whiteboard), `PrivateBoard` (a padlock), `LoadFailed` (a plug pulled from its socket), `BoardGone` (a closed, taped box) and `SessionEnded` (an hourglass that has run out), all tinted, none with a sparkle (ADR-038 §14). Each is a 128 board (`size.illustration`) of paths: every stroke `color.illustration-ink` at `border.width-illustration`, fills `color.illustration-tint` or `color.background` (ADR-033 §8, §9). **Since 2026-11-27 each is a 192 board** ([ADR-047](adr/adr-047-illustration-size-192.md)): every path is set to **Scale** on both axes, so changing `size.illustration` resizes the drawing, and the strokes stay 2. The part's board is 2184 × 314 with the eight looks 260 apart, and the nine copies on screens grew by 64 |
| Sidebar row | **Done** 2026-11-21, board `Sidebar row — states`. One axis, `State`: `Rest`, `Hover`, `Pressed`, `FocusVisible`, `Current`. A flex row, 224 wide at rest:<br>• `size.control-height`, `spacing.control-padding-x`, `spacing.icon-gap`, `radius.control` and `color.background`, with an inner 1px edge on Hover, Pressed and Current<br>• `Slot` (`size.icon`), holding `Swatch` (`size.swatch`, `radius.badge`, periwinkle `.strong`) and the hidden `Glyph / All boards`, `Glyph / Retry` and `Dot` (`size.dot`, `color.primary`)<br>• a hidden Avatar, then `Name` (fills the row: `Title` in `type.body`, and a hidden filled `Lock`), then a hidden `Caret`<br>• the glyphs are Phosphor paths at 16 in `color.muted-foreground`, and `color.primary` on Current ([ADR-035](adr/adr-035-sidebar-loading-failure-and-switcher-mark.md) §7) |
| Sidebar | **Done** 2026-11-21, board `Sidebar — states`. Two axes, `Switcher` × `List`: Closed · Rows, Open · Rows, Closed · Loading, Closed · Failed, Closed · Empty.<br>• 240 × 900, `color.background`, a right-only rule<br>• `Top row` (`size.app-bar`, `spacing.screen-x`, `spacing.inline-gap`, a bottom rule): `Toggle` (`size.target-min`, Phosphor `SidebarSimple`) and *CollaboTask* in `type.title`. **Since 2026-11-28** a `Logo` copy, `Sidebar lockup`, in place of the text, in all five looks; all 48 copies followed ([ADR-049](adr/adr-049-the-logo.md))<br>• `List` (`spacing.sidebar-padding`, `spacing.section-gap`): `Nav` (the switcher, and *All boards* as Current) and `Boards` (*Your boards*, then four rows), each `spacing.list-gap` apart<br>• `User` (`spacing.sidebar-padding`, a top rule): *AP · Ayu Pratama*<br>• below the variants, `App bar — sidebar hidden`: the button, then *All boards* (ADR-035 §7) |
| Add card | **Done** 2026-11-23, board `Add card — states`, filled `color.board-column`. One axis, `State`: `Rest`, `Hover`, `Pressed`, `FocusVisible`. A flex row 264 wide at rest: `size.control-height`, `spacing.control-padding-x`, `spacing.icon-gap`, `radius.control`, filled `color.board-column`; an inner 1px edge on Hover (`color.border-hover`) and Pressed (`color.border-pressed`), `shadow.focus-ring` on FocusVisible. Inside: `Icon` (`size.icon`, Phosphor `Plus` 12 × 12 at offset 2, `color.muted-foreground`) and `Label` (*Add card*, `type.body`, `color.muted-foreground`) ([ADR-037](adr/adr-037-board-at-rest.md) §4) |
| Field | **Done** 2026-11-23, board `Field — states`, filled `color.popover`. One axis, `State`: `Rest`, `Hover`, `Pressed`, `FocusVisible`, 452 apart. Made from an unlinked Sidebar row: a flex row 432 wide, `size.control-height`, `spacing.control-padding-x`, `spacing.icon-gap`, `radius.control`, filled `color.popover`; an inner 1px edge on Hover and Pressed, `shadow.focus-ring` on FocusVisible. Inside: `Slot` (`Glyph / Column`, Phosphor `Kanban` 12 × 11 at 2, 3; the hidden `Glyph / Due date`, `CalendarBlank` 12 × 13 at 2, 1), a hidden Avatar, `Value` (`Text` in `type.body`) and `Caret` ([ADR-038](adr/adr-038-board-states-group-d.md) §16). **Later on 2026-11-23** the copies on 2.6 and *2.6 + delete card* show the caret on all three fields ([ADR-041](adr/adr-041-the-card-drawers-pickers.md) §8) |
| Password rule | **Done** 2026-11-21, board `Password rule — states`. One axis, `State`: `Unmet`, `Met`, `Missed`. A flex row aligned to the top, width `size.form-width`, gap `spacing.icon-gap`: a 16 `Icon` board (`size.icon`) holding Phosphor `Circle` / `Check` / `X`, and the rule's `Text` in `type.meta`. Glyph and text share `color.muted-foreground` / `color.done` / `color.destructive-text`. Below the variant container, two `Doc / Page` demos of the whole field: a `Password field` (Input + a `Password rules` list, rows `spacing.list-gap` apart) while typing `abcde1!`, and after pressing, where the field is `Invalid` with its `Message` hidden ([ADR-032](adr/adr-032-auth-form-errors-and-password-rules.md) §4) |
| Foundations board | **Done** 2026-11-20, on page `Foundations`. Six sections, specimen first: the 14/16 specimen, type, the elevation ladder (real Card instances on a bound canvas and columns), radius, the identity hues in both assignments, and the catalogue as three tables generated from the sets' stored values. The binding check reports only the 48 `Hue /` shapes and the five Card `⋯` glyphs, both expected. It was switched to Dark and to Coarse and looked at, then restored to Light + Fine. **Updated 2026-11-20 with ADR-025:** the catalogue tables were regenerated to 153 rows. That also adds the five ADR-024 tokens, which Stage 2 had left off the board. The identity notes now say the hues are tokens; this section's two side-by-side pairings stay typed in. **Updated again 2026-11-20 with ADR-026:** three rows added, so the catalogue read 157. **And with ADR-027:** one row added to the `base` table, so it reads 158. **And with ADR-028:** one more `base` row, so it reads 159. **And with ADR-029:** a `color.skeleton` row in the theme table (swatch bound) and a `size.skeleton-height` row in the density table, so it reads 163, with the theme table titled 53 each and the density table 13 each. **And with ADR-030:** two `base` rows and two theme rows (both swatches rebound), so it reads 169, with `base` titled 33 and the theme table 55 each. **And with ADR-031:** a `radius.full` row in `base`, a `shadow.avatar-ring` and a `color.avatar-neutral` row in the theme table (both swatches rebound), and two density rows, so it reads 178, with `base` titled 34, the theme table 57 each and the density table 15 each. Its radius ladder still shows ADR-017's three steps; `radius.full` is in the table only. Its two loose column headers in the elevation ladder (a title only) predate the Column header part and have not yet been replaced by instances. **And with ADR-032** (2026-11-21): a `spacing.list-gap` row in `base`, so it reads 179, with `base` titled 35. **And with ADR-033** (2026-11-21): seven `base` rows, two theme rows (both swatches rebound to their own tokens) and two density rows, so it reads 194, with `base` titled 42, the theme table 59 each and the density table 17 each; the *as of* date reads 2026-11-21. **And with ADR-034 and ADR-035** (2026-11-21): four `base` rows and a `color.scrim` row, whose swatch is bound and shows the alpha. It reads 200, with `base` titled 46 and the theme table 60 each. The binding check reports its expected 53 rows and nothing else. **And with ADR-036** (2026-11-22): three theme rows, so it reads 206. **And with ADR-038** (2026-11-23): two `base` rows and a `shadow.rule-left` row (swatch bound), so it reads 210, with `base` titled 48 and the theme table 64 each. Its five Card instances' meta rows had kept a typed gap of 8 after ADR-037 §7; they are now bound to `spacing.inline-gap`. **And with ADR-040** (later on 2026-11-23): four `base` rows and three theme rows (swatches bound), so it reads 220, with `base` titled 52 and the theme table 67 each; the *as of* date reads 2026-11-23. **Checked at Stage 6 (2026-11-27):** the board's title reads 222, `base` 52, the theme table 67 each and the density table 18 each, which is the count ADR-042 left behind. Every one of the 222 tokens has a row, and every row's stored value matches its token in the sets. **Later on 2026-11-27** the `size.illustration` row was changed from 128 to 192 by hand ([ADR-047](adr/adr-047-illustration-size-192.md)). **And with ADR-049** (2026-11-28): three theme rows for `color.logo.still`, `.moving` and `.name` (swatches bound), so it reads 228, with the theme table 70 each; the *as of* date reads 2026-11-28 |
| The `Brand` page | **Done** 2026-11-28 ([ADR-049](adr/adr-049-the-logo.md) §8), first in the file, 220 shapes in three rows, each under a `Doc / Row label`.<br>• **Logo** (y 0): the `Logo` part at 0, 140; `Doc / Logo — construction` (the mark at 256 on its 32 grid) and `Doc / Logo — sizes` (64, 32, 24, the straight 16, and the lockup)<br>• **Logo — fixed colours** (640): `Doc / Logo — Light and Dark`, `Doc / Logo — one colour` and `Doc / Logo — app icon` (the 512 master `Fixed / App icon`, and a rounded 120 preview)<br>• **Logo — rules** (1616): `Doc / Logo — space and size`<br>Everything named `Fixed / …` is a detached copy with typed colours, on purpose (§7.5) |
| The `Account` page | **Done** 2026-11-25 ([ADR-045](adr/adr-045-design-file-pages-by-area.md) §7). Four boards, one row each, at x 0. Every row on the screen pages starts with a `Doc / Row label` board 160 above its boards, and rows are 400 apart (ADR-045 §3–§4).<br>• **2.1 Login** (y 160): Input and Button copies, its heading and the register line, drawn as it first appears: empty fields, the email placeholder only, the button `NotSatisfied`, *Sign in* and *Don’t have an account yet? Create one* ([ADR-032](adr/adr-032-auth-form-errors-and-password-rules.md))<br>• **2.2 Register** (1460): a copy of 2.1, drawn as the email comes back taken: Name, Email `Invalid` with the 409 wording, a `Password field` with the four rules `Met`, Confirm password, the button at `Rest`, *Already have an account? Sign in* (ADR-032 §9)<br>• **2.13 Session expired** (2760): a copy of 2.1 ([ADR-038](adr/adr-038-board-states-group-d.md))<br>• **User menu** (4060): *2.4 + user menu* ([ADR-040](adr/adr-040-menus-and-the-confirmation.md)). Its row's note points to the phone's user menu, which is on *Board — phone* as *2.15 + sidebar, user menu*; the note names that page (ADR-045 §5, [ADR-048](adr/adr-048-board-page-split-in-four.md) §5)<br>**2026-11-28:** 2.1, 2.2 and 2.13 open their `Form` with a `Logo` copy (`Lockup`), `spacing.section-gap` above the heading. Register's form is 640 tall ([ADR-049](adr/adr-049-the-logo.md) §7) |
| The `Workspaces` page | **Done** 2026-11-25 (ADR-045 §7), from what was `Page 1`. 24 boards in three rows. 2.3 has no sidebar ([ADR-033](adr/adr-033-workspace-screens.md)). Each 2.4 board is a row: a `Sidebar` copy in the variant its state needs, then `Main`, the app bar reading *All boards* and a 4-column grid ([ADR-035](adr/adr-035-sidebar-loading-failure-and-switcher-mark.md) §7).<br>• **2.3 Workspaces** (y 160): 2.3, cold load, empty, load failed, then *2.3 + create workspace* and *2.3 + removed from a workspace*<br>• **2.4 Workspace detail** (1460): 2.4, archived, cold load, empty, archived and empty, load failed<br>• **2.4 — what opens over it** (2760): *2.4 + workspace ⋯*, *+ create board*, *+ create board, a short title*, *+ invite people*, *+ invite people, sent*, *+ members*, *+ removing a member*, *+ archived, with Undo*; at 3860, *Create a board at 390*, *Invite people at 390*, *Members at 390* and *Members at 390 + a person’s ⋯*<br>Each overlay is an absolute child of its screen, placed by position: the desktop ones for Fine, the phone ones for Coarse (ADR-040 §1). Each toast sits 16 from the bottom and right ([ADR-044](adr/adr-044-the-sidebar-sliding-over-and-the-toast.md) §14). Drawn by ADR-033, -035 and -040 to -044. **2026-11-28:** 2.3's six app bars hold a `Logo` copy (`Lockup`) on the left in place of *CollaboTask* ([ADR-049](adr/adr-049-the-logo.md) §7) |
| The `Board` page | **Done** 2026-11-25 (ADR-045 §7), from what was `Overlays`. **Since 2026-11-27, 11 boards in three rows** ([ADR-048](adr/adr-048-board-page-split-in-four.md) §2–§3); the other three rows are on the three pages below.<br>• **2.5 Board** (y 160): 2.5 Board — at rest, then 2.8 Join gate, 2.9 Cold load, 2.10 Empty board, 2.11 Load failed, 2.12 Gone — archived and 2.12 Gone — not found. 2.5 is built like 2.4: a `Sidebar` copy with *Q3 Launch Plan* current, then `Main`, holding the app bar, the `Rail` and the `Canvas` with five columns ([ADR-037](adr/adr-037-board-at-rest.md) §8, ADR-038)<br>• **2.7 Read-only board** (1460): 2.7, then *2.7 + an edit you can’t make*<br>• **2.14 Drag** (2760): *2.14 Drag — a card* and *2.14 Drag — a column*, copies of 2.5 ([ADR-039](adr/adr-039-drag-and-phone-board.md))<br>Overlays and toasts are placed as on Workspaces, here and on the three pages below. Drawn by ADR-037 to ADR-044 |
| The `Board — what opens over it` page | **Done** 2026-11-27 ([ADR-048](adr/adr-048-board-page-split-in-four.md)), moved whole from `Board`. 13 boards in one row, its label at 0, 0.<br>• **2.5 — what opens over it** (y 160): *2.5 + board ⋯*, *+ column ⋯*, *+ renaming a column*, *+ card ⋯*, *+ moving from the card ⋯*, *+ board settings*, *+ adding a column*; at 1260, *+ members*, *+ members, a person’s ⋯*, *+ make owner*, *+ who’s here*, *+ a move that didn’t stick*, *+ the open card deleted*. *Leave board* is hidden in the board ⋯, since you own Q3 Launch Plan ([ADR-043](adr/adr-043-the-lists.md)) |
| The `Card` page | **Done** 2026-11-27 ([ADR-048](adr/adr-048-board-page-split-in-four.md)), moved whole from `Board`. 9 boards in one row, its label at 0, 0.<br>• **2.6 Card drawer** (y 160): 2.6, *2.6 + assignee*, *+ due date*, *+ moving, a column*, *+ delete card*, *+ a change that didn’t save*, then from x 9840 *2.6 at 390*, *2.6 at 390 + due date* and *2.6 at 390 + moving, a place*. A drawer's picker sits in its `Fields` as a `Row / … picker`. The phone drawer is the 2.6 drawer resized to 390 × 704, with no shadow ([ADR-041](adr/adr-041-the-card-drawers-pickers.md) §16)<br>**2026-11-27:** 45 size bindings lost after the move were re-applied, and *2.6 + delete card*'s *Q3 Launch Plan* sidebar row, 44 tall in Fine since an earlier flip, is bound to `size.control-height` and 36 (ADR-048 §6) |
| The `Board — phone` page | **Done** 2026-11-27 ([ADR-048](adr/adr-048-board-page-split-in-four.md)), moved whole from `Board`. 12 boards in one row, its label at 0, 0.<br>• **2.15 Phone board** (y 160): *2.15 Phone board* and *2.15 Phone — dragging a card*, 390 × 844, built by hand (ADR-039), then *2.15 + board ⋯*, *+ delete column*, *+ adding a column*, *+ moving a card*, *Board members at 390*, *2.15 + who’s here*, *+ a move that didn’t stick*, *+ sidebar*, *+ sidebar, switching workspace* and *+ sidebar, user menu* |
| Glyph | **Done** 2026-11-23, board `Glyph — states` (y 7838), filled `color.popover`. `Name` × `Tone`, fifteen looks, 48 apart. Each a 16 board bound to `size.icon`, holding one path: Phosphor's Regular outline scaled from 256 to 16, no stroke, its fill bound to the tone's colour ([ADR-040](adr/adr-040-menus-and-the-confirmation.md) §14). **Later on 2026-11-23** four more, 36 apart: `X` Muted, `CaretLeft`, `CaretRight` and `MagnifyingGlass`; the board widened to 953 ([ADR-041](adr/adr-041-the-card-drawers-pickers.md) §16). **2026-11-24** three more: `Check` Done and `Check` Muted (the password checklist's check, its default black stroke removed) and `X` Destructive; the board widened to 1060 ([ADR-042](adr/adr-042-the-forms.md) §17). **2026-11-24** six more for the lists: `Crown`, `ShieldCheck`, `User`, `UserMinus` Muted and Destructive, and `DotsThree` Muted, at x 990–1170; the board widened to 1276, twenty-eight looks ([ADR-043](adr/adr-043-the-lists.md) §20) |
| Menu row | **Done** 2026-11-23, board `Menu row — states` (y 8074), filled `color.popover`. `State` × `Tone`, eight looks in two rows, *Rename* and *Delete column*. A flex row 216 wide: `size.control-height`, `spacing.control-padding-x`, `spacing.icon-gap`, `radius.control`, filled `color.popover`; an inner 1px edge on Hover and Pressed, `shadow.focus-ring` on FocusVisible. Inside: `Slot` (`size.icon`) holding a `Glyph`, and `Label` filling the row |
| Menu | **Done** 2026-11-23, board `Menu — states` (y 8388), filled `color.board-canvas`. `Of` × `Presentation`, ten looks: the five popovers in a row, the five sheets under them. Popover: `size.menu-width`, `color.popover`, an outer 1px `color.border` edge, `shadow.overlay`, `radius.surface`; `Panel` with `spacing.menu-padding` and `spacing.list-gap`. Sheet: 390 typed, top corners `radius.surface`, `shadow.overlay`; `Panel` filled `color.popover` with `shadow.rule-top`, and a `Head` (`size.control-height`, `spacing.control-padding-x` on its left, `spacing.inline-gap`): `Title` or `Identity`, and `Close` (`size.target-min`, a `Glyph` X). `Identity`: an Avatar, then `Name` (`type.card-title`) over `Email` (`type.meta`), `size.control-height` tall. **2026-11-24** six more: *Teammate*, *Board member* and *Workspace member*, each Popover and Sheet, the sheets' heads showing Dimas Nugroho; the board widened to 4192, sixteen looks ([ADR-043](adr/adr-043-the-lists.md) §20) |
| Confirmation | **Done** 2026-11-23, board `Confirmation — states` (y 9212), filled `color.board-canvas`. One component, no axis, worded for *Delete column*, then five copies: *Delete card*, *Delete workspace* (field shown, confirm *NotSatisfied*), the same typed (*Deleting…*), *Archive board* (confirm Primary) and *Leave board*. **2026-11-24** six more in a second row, 378 below: *Remove member*, *Remove from board*, *Remove owner*, *Make owner* (Primary), *Stop being an admin* (Primary) and *Leave workspace*; the board is 648 tall ([ADR-043](adr/adr-043-the-lists.md) §20) |
| Picker row | **Done** 2026-11-23, board `Picker row — states` (y 9680), filled `color.popover`. `State`: Rest, Hover, Pressed, FocusVisible, Current, 244 apart. A flex row 224 wide: `size.control-height`, `spacing.control-padding-x`, `spacing.icon-gap`, `radius.control`, filled `color.popover`; inner 1px edges on Hover, Pressed and Current (`color.primary`), `shadow.focus-ring` on FocusVisible. Inside: `Slot` (a `Glyph`, `ArrowLeft`), a hidden Avatar, `Label` (`Name` in `type.body`, a hidden `You`), a hidden Count chip ([ADR-041](adr/adr-041-the-card-drawers-pickers.md) §16) |
| Day | **Done** 2026-11-23, board `Day — states` (y 9944). `State` × `Today`, seven looks. A `size.control-height` square, `radius.control`, filled `color.popover`, centring `Number` in `type.body`. On Today the number binds `font.weight.semibold` and `font.family`, its 14/20 typed: §7.5 expects it ([ADR-041](adr/adr-041-the-card-drawers-pickers.md) *Operational note*) |
| Calendar | **Done** 2026-11-23, board `Calendar — states` (y 10208). One part: `Head` (`Month` in `type.card-title`, `Previous` and `Next` at `size.target-min`), `Weekdays`, and `Weeks`: six rows of seven, 30 `Day` copies and 12 `Blank` boards, all `spacing.list-gap` apart. 276 × 296 at Fine, 332 × 352 at Coarse ([ADR-041](adr/adr-041-the-card-drawers-pickers.md) §9) |
| Picker | **Done** 2026-11-23, board `Picker — states` (y 10720), filled `color.board-canvas`. `Of` × `Presentation`, nine looks: Inline (Assignee, Assignee filtered, Due date, Column, Place), Popover and Sheet (Column, Place). Inline: `color.popover`, an inner `color.border` edge, `radius.surface`, `spacing.menu-padding`, `spacing.list-gap`. Popover and Sheet: detached `Menu` Card looks with `Picker row` copies in their `Panel` ([ADR-041](adr/adr-041-the-card-drawers-pickers.md) §16) |
| Swatch | **Done** 2026-11-24, board `Swatch — states` (y 12000), filled `color.popover`. `State`: Rest, Hover, Pressed, FocusVisible, Current. A flex board, `size.control-height` square, `radius.control`, `color.popover`, padded `spacing.inline-gap` on every side, holding a `Colour` rectangle that fills it, `radius.badge`, bound to `color.identity.terracotta.strong`; copies re-bind it. Inner 1px edges on Hover, Pressed and Current (`color.primary`), `shadow.focus-ring` on FocusVisible ([ADR-042](adr/adr-042-the-forms.md) §17) |
| Choice row | **Done** 2026-11-24, board `Choice row — states` (y 12264). `State`: Rest, Hover, Pressed, FocusVisible, Current. A flex column 400 wide (`size.form-width`), `spacing.control-padding-y` / `-x`, `radius.control`, `color.popover`: `Head` (the `Radio`, a `size.icon` circle with a `color.input` edge and a hidden `size.dot` `Dot`, then `Name` in `type.body`, `spacing.icon-gap` apart) over `Sub` (an `Indent` the circle's width, then `Sentence` in `type.meta`). The edge ladder; Current plum ([ADR-042](adr/adr-042-the-forms.md) §17) |
| Address chip | **Done** 2026-11-24, board `Address chip — states` (y 12544). `State`: Rest, Invalid. A flex row, `size.target-min` tall, `color.chip`, `radius.badge`, `spacing.inline-gap` on the left: `Address` in `type.body`, then `Remove`, a `size.target-min` square holding a `Glyph` X Muted. Invalid adds a `color.destructive` edge ([ADR-042](adr/adr-042-the-forms.md) §17) |
| Invite result | **Done** 2026-11-24, board `Invite result — states` (y 12808). `Outcome`: Added, NoAccount, AlreadyIn. A flex column 400 wide: `Head` (a `Glyph`, then `Address`) over `Sub` (`Indent`, then `Outcome` in `type.meta`, in `color.done`, `color.destructive-text` or `color.muted-foreground`) ([ADR-042](adr/adr-042-the-forms.md) §17) |
| Form | **Done** 2026-11-24, board `Form — states` (y 13064). `Form` × `Frame`, seven looks: Workspace, Board, Settings, Invite and Invite sent as Dialog; Board and Invite as Sheet. A Dialog is the confirmation's frame (`size.dialog-width`, `spacing.dialog-padding`, `spacing.section-gap`, `color.popover`, a `color.border` edge, `radius.surface`, `shadow.overlay`) holding `Text`, `Fields` (`spacing.field-gap`) and `Actions`; a Sheet is 390 wide, `spacing.screen-x` inside, a `Head` with *Close*, and one full-width button. Invite/Sheet shows a rose chip and *Invite* not satisfied ([ADR-042](adr/adr-042-the-forms.md) §17) |
| Person row | **Done** 2026-11-24, board `Person row — states` (y 13964), filled `color.popover`. `Trail`: Menu, Add, Adding, None. A flex row `size.control-height` × `size.form-width`, `spacing.control-padding-x` on its left, `spacing.icon-gap`, `radius.control`, `color.popover`: an Avatar, `Who` (`Label`: `Name` in `type.body` and a hidden `You`, `spacing.inline-gap` apart; over `Sub` in `type.meta`, `color.muted-foreground`), then `Menu` (`size.target-min`, a `Glyph` `DotsThree` Muted) or an Outline *Add* / Busy *Adding…*. None pads its right too ([ADR-043](adr/adr-043-the-lists.md) §20) |
| List | **Done** 2026-11-24, board `List — states` (y 14228). `Of` × `Frame`, six looks: Team and Members as Dialog and Sheet, Here as Popover and Sheet. A Dialog is the `Form` part's 448 frame holding `Text`, a `Body` (the filter `Input`, then the rows, `spacing.list-gap` apart) and `Actions` with *Close*; Team shows eleven rows with the last three hidden, eight tall (380 at Coarse, 316 at Fine). Members is drawn scrolled to its end, so *Not on the board yet* shows with three *Add* rows. A Sheet is 390, `spacing.screen-x` inside, a `Head` with *Close*. Here is the Menu's surface with *8 here now* and eight `Person row` copies, `Trail` None, no second line ([ADR-043](adr/adr-043-the-lists.md) §20). **2026-11-24:** flavour 6 added. Both Members looks gained a hidden `Empty` (a row padded `spacing.control-padding-x` holding *Everyone in Product Team is already on this board.* in `type.body` and `color.muted-foreground`) after *Not on the board yet*; the board is widened to 3088 with `List / Members, everyone on the board`, a Members/Dialog copy that hides the second group, shows `Empty`, reads *11 people on Q3 Launch Plan.* and ends Lina, Maya, Rina, Sari |
| Logo | **Done** 2026-11-28, on page `Brand` (not `Components`), the part `Logo`. One axis, `Look`: `Mark`, `Mark small` (the straight version, below 24) and `Lockup`. The marks are 32 boards of three paths set to **Scale**, filled `color.logo.still` (the long and top cards) and `color.logo.moving`. `Lockup` is a flex row, `spacing.icon-gap` apart: a `Mark` copy, then `Name` (*CollaboTask*, `font.family`, `font.weight.semibold`, `color.logo.name`, 20/28 typed). **Later on 2026-11-28** a fourth look, `Sidebar lockup`: mark 28, gap 8, name 17.5, 147 × 28, for the sidebar's top row ([ADR-049](adr/adr-049-the-logo.md) §7, §8) |
| The ~19 overlays | **Group a done** 2026-11-23: the workspace, board, column and card ⋯ menus, the user menu, the inline rename and the confirmation. **Group b done** later the same day: the date and assignee pickers and *Move card to…*, in place in the drawer and floating from the board ([ADR-041](adr/adr-041-the-card-drawers-pickers.md)). **Group c done** 2026-11-24: create workspace, create board, board settings and invite people as the `Form` part, and *Add column* as a new column on the board ([ADR-042](adr/adr-042-the-forms.md)). **Group d done** 2026-11-24: the team and a board's members as the `List` part from *N members* and the board ⋯, a person's ⋯ with the `Menu` part, adding people as the members list's second group, *Make owner* and five more confirmations, and who's here from the presence cluster ([ADR-043](adr/adr-043-the-lists.md)). **Group e done** 2026-11-24: the sidebar sliding over on the phone, with the switcher open and with the user menu's sheet over it, and the toast on 2.3, 2.4, 2.5, 2.6, 2.7 and 2.15 ([ADR-044](adr/adr-044-the-sidebar-sliding-over-and-the-toast.md)). **All the overlays are drawn** |

The foundation is laid, thirty-four parts stand on it (Stage 3's ten, two that 2.2 needed, three that 2.3 and 2.4 needed, the two sidebar parts, `Add card` from 2.5, `Field` from 2.6, `Glyph`, `Menu row`, `Menu` and `Confirmation` from Stage 5's group a, `Picker row`, `Day`, `Calendar` and `Picker` from group b, `Swatch`, `Choice row`, `Address chip`, `Invite result` and `Form` from group c, and `Person row` and `List` from group d; since 2026-11-28 a thirty-fifth, `Logo`, on `Brand`, [ADR-049](adr/adr-049-the-logo.md)), and all fifteen screens have been built from them end to end, 2.5 the reference frame among them. **Stages 1 to 4 are closed** (2026-11-23, [ADR-039](adr/adr-039-drag-and-phone-board.md)). **Stage 5 is closed** (2026-11-24: [ADR-040](adr/adr-040-menus-and-the-confirmation.md), [ADR-041](adr/adr-041-the-card-drawers-pickers.md), [ADR-042](adr/adr-042-the-forms.md), [ADR-043](adr/adr-043-the-lists.md), [ADR-044](adr/adr-044-the-sidebar-sliding-over-and-the-toast.md)); **Stage 6, the final check, ran 2026-11-25 to 2026-11-27** ([ADR-046](adr/adr-046-stage-6-final-check.md)). All 34 boards on `Components` now have a bound `color.background` fill, the twelve oldest having gained one on 2026-11-27, so a caption never sits on Penpot's own canvas. Everything below is the record of how the file got there.

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
| 7 | **Skeleton block** | §2.9 (card-sized blocks in column panels, `color.skeleton` fill, **no shimmer**) | Only board 2.9 — shimmer is movement, and movement means a teammate acted |
| 8 | **Toast** | §3 (its design and the access-revoked wordings are settled by [ADR-030](adr/adr-030-toast.md); placement and the other wordings are not) | Draw the container and the longest message it must hold |
| 9 | **Presence cluster** | §1.5 (app bar), §3 (the `+N` overflow sheet; its geometry, the seat and the appearances are settled by [ADR-031](adr/adr-031-presence-cluster.md), and the sheet itself is Stage 5) | **Last.** It is a cluster *of Avatars* — the only hard dependency here |

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

#### What question 4 returned for the Banner — and what closed it

Asked on 2026-11-20 before drawing. The fill, the rule, the side padding, the message's type and colour, and the gap before Join were all in the catalogue, each bound by a rule rather than by a matching number. One value was not: **the space above and below the contents.** The banner holds Join at 36 / 44, and nothing said how much room surrounds it.

The check also asked whether §4.2's *informational* banner had anyone to show it to. It did not.

**Settled by [ADR-028](adr/adr-028-banner.md)** (2026-11-20):
- `spacing.banner-padding-y` 8, in `base`, because Join already grows at Coarse. Padding rather than a height, because the message wraps on a phone.
- the fill is `color.background`, because Join's focus ring draws its offset in that colour, with a bottom rule because Dark's background and canvas are the same `#0C0A09`
- side padding `spacing.screen-x`, so the message starts where the first column starts
- `spacing.inline-gap` before Join, taking ADR-027's offer
- one appearance, no axis: every visitor to a read-only board can join, and no other state is a banner
- no icon, and the wording *"You’re viewing this board. Join it to add, edit and move cards."* with a *Join board* button

Question 4 has now fired at seven stages out of seven.

#### What question 4 returned for the Skeleton block — and what closed it

Asked on 2026-11-20 before drawing. The corners, the gap between blocks and the column panel were all in the catalogue, each bound by the rule that a block stands where a card will. Two things were not usable:
- **the fill.** ADR-018 §7 named `--muted`, and in Light `color.board-column` is stored as `{color.muted}`, so a muted block inside a column is 1.00:1, the column's own colour. Dark worked only because `color.muted` holds the card's value there. A column-sized muted block on the canvas can be seen, but card-sized ones inside it cannot, in either theme.
- **the height.** A card's height comes from its title lines and its meta row (ADR-017 §1), and a skeleton holds neither.

**Settled by [ADR-029](adr/adr-029-skeleton-block.md)** (2026-11-20), which supersedes ADR-018 §7's fill and nothing else in it:
- `color.skeleton`, `#E7E5E4` / `#292524`, written as values in each theme set. It is 1.15:1 on the column in both themes, the only candidate that is the same in both. In Dark it is the card's own fill.
- `size.skeleton-height` 60 / 72, in the density sets: ADR-017 §1's formula for one title line and a meta row, the real card shape nearest §1's average of 69.5
- the column is drawn as its real panel, with card-sized blocks inside it, because one colour cannot draw both sizes nested
- a block sits on a column panel or on the page background, never directly on the canvas (1.00:1 in Light) and never inside a card-coloured surface (1.00:1 in Dark). The cold loads of §1.3 and §1.4 have to respect this when they are drawn at Stage 4.
- no outline, one height, no axis

**One consequence for the code:** shadcn's Skeleton ships `animate-pulse`, and it is removed in the same pass as ADR-018's `disabled:` override.

Question 4 has now fired at eight stages out of eight, and this is the first time it found a value that existed but could not be seen.

#### What question 4 returned for the Toast — and what closed it

Asked on 2026-11-20 before drawing. The corners, the shadow, the message's type, the gap before the close button and the close glyph's size and area were all in the catalogue, each bound by a rule. Three things were not usable:
- **the fill.** shadcn's toast is `color.popover`, and a toast floats over everything. Measured against the canvas, a column, a card and the page, it is 1.00:1 over a card and over the page in Light and 1.00:1 over a column in Dark. The `color.border` edge reaches 1.26 at best in Light.
- **the padding**
- **the width.** ADR-022 had said a toast needs no size token, and drawing one needs a width that is either typed or a token.

It also found that the access-revoked wordings did not exist anywhere. Only their reasons did. And after `removed_from_workspace`, ADR-018 §10's destination is a workspace you can no longer open.

**Settled by [ADR-030](adr/adr-030-toast.md)** (2026-11-20):
- an inverted grey toast: `color.toast` `#44403C` / `#E7E5E4` (stone-700 / stone-200, not the text colours) with `color.toast-foreground` `#FFFFFF` / `#0C0A09`, written as values. It is 8.18:1 or better over every surface in both themes, so it has no edge.
- `spacing.toast-padding` 12 on all sides, in `base`, the inset every other surface gives its text
- `size.toast-width` 358, which is `390 − 2 × spacing.screen-x`, so a message wraps the same way on the phone and on a desktop
- one appearance, no axis, no icon, no hue, and a Phosphor `X` close shown at rest at every pointer
- three wordings that name what you lost and not who did it, and `/workspaces` as the destination after `removed_from_workspace`

**One consequence for the code:** if the scaffold uses shadcn's Sonner wrapper, its `--normal-*` variables are re-pointed to the new ones.

Question 4 has now fired at nine stages out of nine.

#### What question 4 returned for the Presence cluster — and what closed it

Asked on 2026-11-20 before drawing. The faces were in the catalogue: `size.avatar`, the initials, the Avatar's Muted look, the label's colour and the gap before it. Three things were not usable:
- **how far the faces overlap.** ADR-019 §8 gives 8px with 24px faces on a phone, and nothing said what a 20px desktop face does. Penpot also refuses a negative gap.
- **anything between overlapping circles.** The worst neighbours (ochre, sage, teal) are 1.00–1.03 apart, and two grey faces are identical.
- **the grey circle.** `color.muted` is 1.09:1 on the white app bar, under the faintest live circle (1.11). In a stack, the face in front cuts off the initials behind at an edge nobody can see. In Dark it is 1.30, far under the live circles' 2.05.

Four choices no ADR had made came with them: what `+N` is, where *Reconnecting…* goes on a phone, how many faces show from 640px up, and whether you see yourself. The options were rendered side by side in Light and Dark. The user took the recommendations except one: `+N` is a round seat, not a Count chip. Question 4 was then asked again of the seat.

**Settled by [ADR-031](adr/adr-031-presence-cluster.md)** (2026-11-20):
- `spacing.avatar-overlap` −6 / −8 in the density sets, so a covered face shows the same share at both sizes. It is drawn as a negative left margin.
- `shadow.avatar-ring`, a 2px ring in `{color.background}`: Tailwind's `ring-2 ring-background`, which takes no room
- `color.avatar-neutral` `#E7E5E4` / `#57534E`, matched to the faintest live circle in each theme. It is now the Avatar's Muted and Unknown circle as well.
- `+N` as a round seat in the stack: `radius.full` 9999, `spacing.seat-padding-x` 2 / 4 so that a one-digit count is a circle, and the neutral fill. A phone's cluster is 72–75px, against ADR-019's 96.
- the whole cluster is the button, `size.target-min` tall (ADR-018 §6), so the seat needs no target of its own
- you first; three faces below 640px and five above; below 640px *Reconnecting…* takes the seat's place
- `State` = Live · Reconnecting · FocusVisible, with no hover or pressed look

**One addition to §7.5's check:** this is the first part to bind a margin, and the binding check could not see margins. It does now.

Question 4 has now fired at ten stages out of ten, and **Stage 3 is closed.**

**Stage 4 — the remaining fourteen screens** (§2), in five groups:

| Group | Boards | Why here |
|---|---|---|
| a | 2.2 Register ✅ | Login's twin plus a name field — nearly free now, and it adds the field-invalid state. *Drawn 2026-11-21. It turned out not to be free: it settled where errors go and brought the password rules, two parts and a token ([ADR-032](adr/adr-032-auth-form-errors-and-password-rules.md))* |
| b | 2.3 Workspaces ✅, 2.4 Workspace detail ✅ | Need Avatar and Count chip. *Drawn 2026-11-21, with eight state boards. They needed three parts, a Button kind and fifteen tokens, and brought the empty-state illustrations ([ADR-033](adr/adr-033-workspace-screens.md)). **Redrawn the same day**, all six boards, with the sidebar ([ADR-034](adr/adr-034-side-menu.md), [ADR-035](adr/adr-035-sidebar-loading-failure-and-switcher-mark.md))* |
| c | **2.5 Board — at rest** ✅, with the sidebar beside it | §2 calls it *"the reference frame everything else varies from"*. The hinge of the whole stage. *Drawn 2026-11-23. It needed no token, one part (`Add card`) and a fix to the Card's meta row, and settled the header gap, the rail, the app bar, Add card, late dates and the empty column ([ADR-037](adr/adr-037-board-at-rest.md))* |
| d | 2.6 drawer ✅, 2.7 + 2.8 ✅, 2.9 cold load ✅, 2.10 empty ✅, 2.11 load failed ✅, 2.12 gone ✅, 2.13 session expired ✅ | Each is a variation on 2.5, and cheap once 2.5 exists. *Drawn 2026-11-23. It needed four token entries, one part (`Field`) and five drawings, and settled the drawer, the centred stack and 2.13's way back ([ADR-038](adr/adr-038-board-states-group-d.md))* |
| e | 2.14 Drag ✅, then 2.15 Phone ✅ | 2.14 needs 2.5 plus **both** placeholders. 2.15 is a different layout rather than a switch, so it is drawn by hand and goes last. *Drawn 2026-11-23. It needed no token and no part, two looks on the `Card`, and a phone drag board, and settled the short column, Add card during a drag and Add column below 640 ([ADR-039](adr/adr-039-drag-and-phone-board.md)).* **Stage 4 is closed.** |

#### What question 4 returned for 2.2 Register — and what closed it

Asked on 2026-11-21 before drawing, in two rounds.

**Round 1 found no missing value**, for the first time in eleven stages. Every colour both forms use clears its floor on `color.background`, the only surface either has. What was missing:
- **a home for a form-level error.** Login's 401 is the same for an unknown email and a wrong password, and US-02 forbids saying which, so it cannot sit under a field. ADR-024 had left this open.
- **any wording.** No document had error copy, and the server's 409 text (*email already exists*) differs from the SRS's example and is not meant for people.
- **a slot for a hint.** The Input part had none.

Two measurements shaped the answers. In Light, the invalid edge is **1.02:1** against a resting edge, so it differs by hue alone. Error text is **1.21:1** against hint text, and 21 apart for a protanope, so a hint that turned into an error in place would change only its words.

**Round 2** came from the user's notes: four password rules, then a live checklist over a sentence, then a confirm field. The rules were checked first against NIST SP 800-63B-4, which forbids composition rules. That was said, and the user chose them anyway. The checklist needed one value the file lacked: the space between two rows. **Question 4 fired**, at eleven stages out of eleven, in round 2.

Both rounds were rendered side by side in Light and Dark with the file's values, and answered with a picker.

**Settled by [ADR-032](adr/adr-032-auth-form-errors-and-password-rules.md)** (2026-11-21):
- a form-level failure is a `Form message`, one line of `type.body` in `color.destructive-text` just above the button. A taken email stays on its field.
- the four rules, shown as a `Password rule` checklist: Unmet · Met · Missed as Phosphor `Circle` / `Check` / `X`, met in `color.done`. The glyph's shape carries the state.
- `spacing.list-gap` 4, new in `base`
- after pressing, the box turns rose and the missed rows become rose crosses, with no separate message line. The 1px rose edge stays, because it never appears without words.
- a confirm-password field, every wording on both forms (*Sign in* among them), and fifteen behaviour rules
- the hint slot became unnecessary: the checklist is the hint

**Upstream, not drawn:** US-01 and SRS UC-01 carry the rules. The server's half is SRS §9's second open backend item.

#### What question 4 returned for 2.3 and 2.4 — and what closed it

Asked on 2026-11-21 before drawing, in two rounds, after checking what the two list endpoints actually return: roles, counts, `access_status`, `visibility` and `background_color`, and no names, so no faces on tiles.

**Round 1 fired.** Nothing fixed the app bar's height, the tile's geometry or minimum width, the gap between tiles, or the swatch's shape and role. Six choices with no missing value came with them: how the switch marks the current list (state 18 is closed), a second Button kind, where *Delete workspace* lives, the empty states, *Owner* or *Admin*, and every wording. Nine choices went on a rendered page. The user took seven recommendations, chose the **roomy tile** and the **left rail**, and asked for **an illustration on the empty states**.

**Round 2** came from that note: the illustration had no colour, line, size or spacing. The user chose a **plum tint** (against the grey recommended), **whole-screen empty states only**, **centred** (against left-aligned), and **128**, and asked for the style to be written down.

**Settled by [ADR-033](adr/adr-033-workspace-screens.md)** (2026-11-21):
- both screens on `color.background`; `size.app-bar` 56, derived from ADR-017 §1 and ADR-019 §7
- the `Tile`: `spacing.tile-padding` 16, `spacing.tile-row-gap` 8, 84 / 88 tall (128 / 140 with an action), a left rail in the hue's `.strong`
- `spacing.tile-gap` 8 / 12 and `size.tile-min-width` 264, measured from the widest count row
- the `Segment` switch (plum's *where you are*), the Button's Outline `Kind`, and an owner-only workspace `⋯`
- `size.tile-skeleton-height` 84 / 88 for the cold loads, blocks on the page and never inside a tile
- the `Illustration` part: `color.illustration-ink`, `color.illustration-tint`, `border.width-illustration` 2, `size.illustration` 128, `spacing.illustration-gap` 16, and §9's ten rules for the next drawing
- *Owner · Admin · Member*, every wording, and the behaviour rules

Question 4 has now fired at twelve stages out of twelve.

#### What question 4 returned for the sidebar — and what closed it

Asked on 2026-11-21, before anything was drawn, while the sidebar itself was being grilled. A side menu had never been discussed, and 2.5 waited on it. Eleven structural questions went first, in text. Then eight visual choices went on a rendered page, each option in Light and Dark with the file's values: where the sidebar's top sits, its surface, the current row's mark, the board colour's mark, its width, 2.4's bar title, the button's icon and the dimming. **The user took all eight recommendations.**

**Settled by [ADR-034](adr/adr-034-side-menu.md)** (2026-11-21). It returned four values with no token: `size.sidebar-width` 240, `spacing.sidebar-padding` 8, `size.swatch` 12, and `color.scrim` (`#0C0A09` at 32% in Light, black at 60% in Dark). **All are in Penpot since 2026-11-21.** A colour token takes transparency, so `color.scrim` is stored as `rgba()`. Penpot refuses the space-separated form ADR-034 §13 wrote (ADR-034 *Operational note*).

Question 4 has now fired at thirteen stages out of thirteen.

#### What question 4 returned for the sidebar parts and the 2.4 redraw — and what closed it

Asked on 2026-11-21, before drawing, of `Sidebar row`, `Sidebar` and 2.4's six boards. It ran in two rounds, both decided on rendered pages.

**Round 1 found no missing value.** Every number binds a token the file had. It found five decisions: the list on a cold load, the list on a load failure, where the open switcher's names start, the space under *Your boards*, and where the user row's avatar sits. The user took three recommendations and chose a block for the switcher on cold load as well. They left notes on two choices: a cleaner retry with a red accent, and a pulsing dot to mark the selected workspace.

**Round 2 came from those notes.** Each ran into a rule ADR-018 had settled: rose's meanings and row 22's *zero channels*, and motion's *a teammate did this*. Each note's aim was drawn beside options that keep the rules, with the costs shown, and the pulse was not offered. The user chose a rose sentence over an uncoloured retry row, and a still plum dot. **The dot needed a size, so question 4 fired.**

**Settled by [ADR-035](adr/adr-035-sidebar-loading-failure-and-switcher-mark.md)** (2026-11-21):
- on a cold load, row-sized blocks for the switcher and three board rows
- on a load failure, *Your boards didn’t load.* in `color.destructive-text`, then a *Try again* row with `ArrowClockwise`. This narrows ADR-018's Rose row and row 22 for the sidebar only.
- workspace rows indented with the board names, and `size.dot` 8, new in `base`, in the selected workspace's slot
- `spacing.list-gap` under *Your boards*, and the user row's avatar in the slot's place
- one `Sidebar row` part for every row, and a `Sidebar` part with `Switcher` × `List`

The catalogue is now 200 tokens and the file has 17 parts. Question 4 has now fired at fourteen stages out of fourteen.

#### What question 4 returned for 2.5 — and what closed it

Asked on 2026-11-23 before drawing, of every value 2.5 needs against the live 206 tokens. Almost everything mapped: the sidebar, the bar, presence, the canvas, the columns, the header, the cards and their date colours, and the rail's size and colour. **One value had no token and no decision**: the space between a column's header and its first card. Both card-count sums assume 0, and at 0 the first card's edge sits on the header's rule.

Five decisions came with it: the rail's position (ADR-023 §2.4 had left it open), the bar's arrangement (ADR-031 had left it to 2.5), Add card's place and look, **late dates told apart by colour alone** (WCAG 1.4.1: in Light, amber is 1.52:1 and rose 1.21:1 against the grey dates), and empty flavour 5's wording. All six went on one rendered page, in Light and Dark with the file's values, with a deuteranopia view of the dates. The user took every recommendation but the rail's.

**Settled by [ADR-037](adr/adr-037-board-at-rest.md)** (2026-11-23):
- the first card sits `spacing.card-gap` below the header, bound by rule, so no token is added
- the rail is its own 4px strip under the bar. That was the user's call over a rail inside the bar, and it costs 4px of height.
- the bar: the name on the left; presence, *Add column* (Primary) and a `color.foreground` ⋯ on the right, `spacing.inline-gap` apart
- *Add card*, a new part: a quiet row after the last card that sticks to a long column's bottom, its stuck rule on a `Footer`. A phone then shows 5 cards and a peek in a long column, not ADR-019 §7's 6.
- a column is as tall as its content, at most to the bottom of the screen. That was the user's note, first misread as *full height*, and it narrows ADR-023 §2.3's wording. 2.5 was redrawn the same day.
- late dates in words: *Due today*, *Due tomorrow*, *Overdue · Nov 20*. *Due soon* means today or tomorrow.
- *No cards yet.* in an empty column
- the Card's meta row keeps a lone avatar on the right

The catalogue stays at 206 tokens, and the file has 18 parts. Question 4 has now fired at fifteen stages out of fifteen.

#### What question 4 returned for group d — and what closed it

Asked on 2026-11-23 before drawing, of every value 2.6 to 2.13 need against the live 206 tokens. Most of it mapped: the dimming, the lift, the drawer's text styles, *Close*, the Banner, the Skeleton block, the empty-board drawing's rules and both Button kinds. **Three values had no token**: the drawer's width, its field labels' width, and a rule on a shape's left edge. Fourteen decisions had never been made. All of them went on one rendered page, in Light and Dark with the file's values, and the user took every recommendation.

Their notes opened a second round. *Delete card* in red needed no decision. Drawings on 2.8, 2.11, 2.12 and 2.13 did, because ADR-033 had allowed drawings on empty screens only, and ADR-018 §7 spends no colour on a failure. Eight questions went on a second page, with that conflict shown once. The user took six recommendations, and chose the plum tint on all four drawings and a left-aligned 2.13.

**Settled by [ADR-038](adr/adr-038-board-states-group-d.md)** (2026-11-23):
- the drawer over everything, dimmed; `color.popover`, `shadow.rule-left` (new) and `shadow.overlay`; `size.drawer-width` 560 (new); labels on the left at `size.field-label` 96 (new); `spacing.screen-x` inside by rule
- Move is the *Column* field, and *Delete card* is a quiet row in `color.destructive-text` at the foot; the drawer shows *Due tomorrow · Nov 24*
- 2.7's banner under the rail, and 2.8 as a centred stack, narrowing ADR-021 §6
- the cold load draws the name and rail at once, with panels out to the right edge
- one centred stack for 2.8 and 2.10 to 2.12: the drawing, 16, the sentence within 400, 24, the button
- 2.12's two wordings, and the gap recorded: deleting a workspace sends no live message
- 2.13 waits for *Sign in* and then brings you back to the same page; the 401 handler raises a flag, and Login takes a same-site `next`
- five drawings, with ADR-033 §8–§9 widened, all tinted

The catalogue is 210 tokens and the file has 19 parts. Question 4 has now fired at sixteen stages out of sixteen.

#### What question 4 returned for group e — and what closed it

Asked on 2026-11-23 before drawing, of every value 2.14 and 2.15 need against the live 210 tokens. Almost everything mapped: the empty place, both shadows, the outline's colour, width and corners, the tint, the phone's bar, and the phone's 72px reservation, which is `spacing.screen-x` + `spacing.column-gap` + `size.target-min` at Coarse. **Three values had no token**: the lifted card's 2% growth, the phone column's 318, and the browser's 140. **Five decisions had never been made.** All eight went on one rendered page, and the user took seven recommendations and overrode one.

**Settled by [ADR-039](adr/adr-039-drag-and-phone-board.md)** (2026-11-23):
- **a short column takes a card only as far as it goes**: nothing changes height during a drag. That was the user's call over stretching every column to the bottom for the drag.
- the tint covers the cards' area only, and *Add card* stays untinted just under it, so a long column's footer keeps its rule
- the `Card` part gains `Armed` and `Lifted`, seven looks in all; their 2% and 1% growth is code only (§4.3)
- a column drag: its place at 60%, a lifted copy, a column-sized outline, nothing tinted
- *Add column* below 640 is first in the board ⋯ menu, and the empty board gains an *Add column* button there
- the phone board shows the browser as two labelled bands, types its 318, and has a drag board beside it
- the count, by eye: **5 whole cards and 44px of the sixth**

The catalogue stays at 210 tokens and the file at 19 parts. Question 4 has now fired at seventeen stages out of seventeen, and **Stage 4 is closed**.

**Stage 5 — the ~18 overlays** (§3).

Last, and not arbitrarily. §3's entire rule for what happens below 640px is written *relative to the card drawer already filling the screen on a phone*. Until 2.6 exists there is nothing to judge that against.

**Drawn in five groups**, recorded 2026-11-23 before anything was drawn. Each group settles its kind of container once, so the groups after it reuse it rather than re-decide it. Ask question 4 of each group before drawing it.

| Group | Overlays | What it settles once |
|---|---|---|
| **a. Menus and the confirmation** | Workspace `⋯`, board `⋯` (with *Add column* first below 640), column `⋯` with inline rename, card `⋯`, user menu, and the destructive confirmation (38) they open | The menu, as a popover from 640 up and a bottom action sheet below; the centred confirmation dialog |
| **b. The card drawer's pickers** | Date, assignee, and *Move card to…* | The inline picker; and the phone's sheet over a sheet ([ADR-038](adr/adr-038-board-states-group-d.md) *Out of scope*) |
| **c. Forms** | Create workspace, create board, board settings, invite people, and **create column** (added 2026-11-23: *Add column* opens it, and §3 had never listed it; settled 2026-11-24 as a new column on the board, not a form) | The form dialog from 640 up and its full-height sheet below — **done** ([ADR-042](adr/adr-042-the-forms.md)) |
| **d. Lists** | Workspace team panel, board members, add people, transfer ownership, presence overflow | The list dialog or drawer, and its rows — **done** ([ADR-043](adr/adr-043-the-lists.md)) |
| **e. The sidebar sliding over, and the toast** | The sidebar below 1024; the toast | Where a toast sits, the wordings for states 4, 27 and 32, and the archive toast's *Undo* ([ADR-040](adr/adr-040-menus-and-the-confirmation.md) §5) |

#### What question 4 returned for group a — and what closed it

Asked on 2026-11-23 before drawing, of every value the five menus and the confirmation need, against the live 210 tokens. The colours, corners, shadows, type and row sizes were all there: a menu row is the Sidebar row filled `color.popover`, and the dimming is `color.scrim`. Two measurements became rules. A menu needs an edge, since in Dark the popover is the column's colour (1.00:1) and in Light the card's. A dialog needs one too, at 1.09 to 1.17:1 against the dimmed screen in Dark. **Six values had no token:** the space inside a menu, its width, the dialog's width and inside, and the rose button's hover and busy fills. **Fourteen choices had never been made**, one of them a conflict: US-12c's *Undo* after archiving against ADR-030 §3's *no action button*.

All fourteen went on one rendered page, in Light and Dark with the file's values, at a 390 phone wherever a phone differs. The user took twelve recommendations. They chose a pale rose busy look over the shared plum one, and added *Leave board* to what is rose. The plugin cannot copy across pages, so the user pasted the four screens onto the new `Overlays` page.

**Settled by [ADR-040](adr/adr-040-menus-and-the-confirmation.md)** (2026-11-23):
- one menu surface: `color.popover`, an edge, `shadow.overlay`, `radius.surface`; `spacing.menu-padding` 4, worked out as 16 − 12; `size.menu-width` 224; 4 below its trigger, right edges lined up
- a Phosphor glyph before every entry, no rules between groups; entries you can't use are absent
- rose for the three deletes and *Leave board*; *Archive board* plain, with a Primary confirm
- all five actions ask first; *Undo* after archiving kept, superseding ADR-030 §3 for that toast (drawn in group e)
- *Delete workspace* asks you to type the name
- a 448 dialog (`size.dialog-width` = `size.form-width` + 2 × `spacing.dialog-padding` 24), 358 at 390, focus on *Cancel*
- the Button's rose kind, `color.destructive-hover`, and a pale rose busy (`color.destructive-subtle`, `-subtle-foreground`)
- below 640, a bottom sheet with a head and a Close
- inline rename fills the header; *Sign out*; every wording, and where each action leaves you
- parts on `Components`, and nine screens with an overlay open on a new `Overlays` page

The catalogue is 220 tokens and the file has 23 parts. Question 4 has now fired at eighteen stages out of eighteen.

#### What question 4 returned for group b — and what closed it

Asked on 2026-11-23 before drawing, of every value the three pickers need, against the live 220 tokens. **Every value mapped**, for the first time since ADR-032's first round. The rows are the `Menu row`'s geometry and the surface is group a's menu, without its shadow in the drawer. A calendar day is a control, `size.control-height` square, so the panel is 284 at Fine and 340 at Coarse. Today is `font.weight.semibold`, and the current choice is `color.primary`. Two measurements set rules: the value column is 262 at 390, too narrow for seven 44px days, and the picker's edge is 1.26:1 / 1.70:1 against the drawer. **Sixteen choices had never been made**: one reclassified *Move card to…* in ADR-021 §5's table, and one, the time of day a due date is stored at, touched the SRS. All sixteen went on one rendered page, and the user took every recommendation.

**Settled by [ADR-041](adr/adr-041-the-card-drawers-pickers.md)** (2026-11-23):
- **the sheet over a sheet never happens**: every picker opens in place under its field in the drawer, *Move card to…* too
- from the board, Move takes the card ⋯ menu's place: the popover from 640 up, the bottom sheet below. It leaves ADR-021 §5's *forms and lists*
- a column, then a place (*Top of column*, *After “card”*, *Bottom of column*); the place moves the card; one line per place
- under the value from 640 up, across the row below; the menu surface laid in the flow; plum text and a plum edge on the current choice; a caret on all three fields
- a six-week calendar, today semibold, the week from the browser's locale, past days pickable, no quick picks, *Remove due date*
- **a due date is a day at midnight UTC, read back by its date part** (SRS §3.6, edited first)
- the board's members, *Unassigned* first and you second, then A–Z, with a filter above eight members
- the drawer at 390, drawn for the first time

The catalogue stays at 220 tokens, and the file has 27 parts. Question 4 has now been asked at nineteen stages and fired at eighteen: group b is the first since ADR-032's first round where it found nothing missing.

#### What question 4 returned for group c — and what closed it

Asked on 2026-11-24 before drawing, of every value the five forms need, against the live 220 tokens. The frame was group a's dialog, which a 400 form fits exactly, and the phone's frame the drawer at 390. **One value had no token:** the space above and below the text of a field that grows, which ADR-022 §3 had left out on purpose. **Twenty choices had never been made**, two of them touching the API contract: what string a board's colour is stored as, and a workspace description's length. All twenty went on one rendered page. The user took sixteen recommendations and chose errors as you type, two choice rows, chips and the results under the field, and asked for a bigger colour square. Those picks needed a chip's height and the square's size, so **question 4 fired a second time**, on a second rendered round. Both resolved by rule, with no token. Four facts came from the backend code: names aren't trimmed, one bad address fails a whole invite, a colour is any 4 to 8 characters, and a workspace description already stops at 1,000 in the usecase.

**Settled by [ADR-042](adr/adr-042-the-forms.md)** (2026-11-24):
- the 448 dialog from 640 up, with no ×; a full-height sheet below, *Close* in its head and one full-width button; pressing outside does nothing, and Escape, *Cancel* and *Close* never ask
- a field of several lines grows from one line, with `spacing.control-padding-y` 8 / 12 above and below its text; that narrows ADR-022 §3
- a minimum length shows as you type, an empty field waits for the press; that narrows ADR-032 §8's rule 1
- Create a workspace warns that nothing can be renamed; its description stops at 1,000 (SRS UC-03)
- eight colour buttons, each holding its hue at 20 / 28, plum edge on the chosen one; the first unused hue first; **stored as the deep hex value** (SRS §4.3, edited first)
- *Workspace* and *Private* as two choice rows; *Save changes*, and a line before going private
- *Invite people* on 2.4's heading; addresses as chips, a target tall; results under the field with the checklist's marks
- *Add column* makes a new column on the board, its header a field, so it left the overlays
- every wording, and the behaviour rules

The catalogue is 222 tokens and the file has 32 parts. Question 4 has now been asked at twenty stages and fired at nineteen.

#### What question 4 returned for Stage 5's group d — and what closed it

Asked on 2026-11-24 before drawing, of every value the five lists need, against the live 222 tokens. The frame was group c's dialog and sheet, and a person's row was the user menu's *Identity* row: 20 + 16 is exactly `size.control-height` at Fine. **One value had no token:** how tall a list gets before it scrolls, since nothing capped a dialog's height. It resolved by rule. **Twenty choices had never been made**, two of them changes to the API contract and one a row of ADR-021 §5. All twenty went on one rendered page, and the user took nineteen recommendations and chose the role on the row's second line, with the email.

Five facts came from the backend code: the owner is always an admin, so the "last admin" refusals can't fire once removing the owner is refused; nothing refused removing the workspace owner; an admin can take a board's owner off by design; leaving a board told open lists nothing; and adding people is all-or-nothing and skips you. Two contradictions came from the design file: Product Team's counts, and Q3 Launch Plan's owner.

**Settled by [ADR-043](adr/adr-043-the-lists.md)** (2026-11-24):
- group c's 448 dialog and full-height sheet hold every list, with *Close*; **eight rows, then the list scrolls**: 8 × `size.control-height` + 7 × `spacing.list-gap`, no token
- a row: an Avatar, the name over *Role · email*, then a ⋯ or *Add* (the user's call over the role at the row's end); every row says its role; you first, then A–Z
- a ⋯ per row, opening the menu; ten actions ask first, three more rose
- *Leave workspace* in the workspace ⋯ for everyone but the owner, whose row has no ⋯; **backend item 4**, the server refuses removing the owner (SRS UC-06, edited first)
- adding people as the board list's second group, one person per press; you are left out until you join
- *Make owner* and a confirmation; the owner's sentence says how to leave; an admin can take the owner off
- **backend item 5**, leaving a board sends `MEMBER_REMOVED` (SRS UC-10, edited first)
- the presence list as a popover and a bottom sheet, which moves it to *Menus* in ADR-021 §5
- Product Team is 11 people and 6 boards everywhere, and you own Q3 Launch Plan

The catalogue stays at 222 tokens, and the file has 34 parts. Question 4 has now been asked at twenty-one stages and fired at twenty.

#### What question 4 returned for Stage 5's group e — and what closed it

Asked on 2026-11-24 before drawing, of every value the slide-over and the toast need, against the live 222 tokens. The panel was the `Sidebar` part as it stands, with `shadow.overlay` over `color.scrim`, and *Undo* fitted the `Button`'s Outline kind unchanged. **One value had no token:** the space between a toast and the screen's edge. It resolved by rule: `spacing.screen-x`, the value the toast's width is already derived from, drawn by position. **Fifteen choices had never been made**, one of them a gap between ADR-034 §4 and ADR-040 §9 (the user menu from the panel on a phone). All fifteen went on one rendered page; the user took fourteen recommendations and chose a stack of up to three. That pick needed the space between two stacked toasts, so **question 4 fired again** and resolved by the same rule: 8, drawn by position.

Seven facts came from the backend code: access revoked carries no actor; Undo is the archive request with `false`, safe to repeat; archiving evicts no one, so 2.12 still hears `BOARD_UNARCHIVED`, though a rejoin of an archived room is refused; a refused change returns 404, 409 or two kinds of 403; an admin edits without joining and a role change has no live message; admins stay in a room that goes private; and a timed-out request may still land.

**Settled by [ADR-044](adr/adr-044-the-sidebar-sliding-over-and-the-toast.md)** (2026-11-24):
- the slide-over drawn on the phone, closed and with the switcher open; focus on the current row; below 640 the user menu's bottom sheet opens over it, over a second dimming
- a toast 16 from the bottom and right, across the bottom on a phone, over the drawer's foot when it is open
- **up to three stacked, newest at the bottom, 8 apart** (the user's call over one at a time); no timer, so WCAG 2.2.1 never applies
- *Undo* as the Outline button; it takes you back to the board, and 2.12 brings teammates back on `BOARD_UNARCHIVED` (US-12c and SRS UC-12c edited first)
- a failure reads *Couldn’t move “X”.* and one of five reasons; after a 403 or 404 the board reloads, so 2.7, 2.8 or 2.12 shows
- every wording, and ten boards on `Overlays`

The catalogue stays at 222 tokens and the file at 34 parts. Question 4 has now been asked at twenty-two stages and fired at twenty-one. **Stage 5 is closed.**

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
  const M = s.layoutChild;
  if (M && !s.name.startsWith("Doc /")) {
    for (const [k, v] of [["marginTop", M.topMargin], ["marginRight", M.rightMargin],
                          ["marginBottom", M.bottomMargin], ["marginLeft", M.leftMargin]])
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
- **Margins were added on 2026-11-20**, with the presence cluster, the first part to bind one: its faces overlap by a negative left margin. Run across the whole file that day, the margin rows returned nothing ([ADR-031](adr/adr-031-presence-cluster.md) *Operational note*).
- **A bound stroke width can be any `strokeWidth*` key.** Binding `border.width` to `strokeWidth` reports all four sides too (ADR-023 *Operational note*). Likewise, an unset opacity reads as `null`, and `null < 1` is true in JavaScript, hence the guard.
- **No rule is a stroke any more.** Since 2026-11-22 every one-sided rule is a `shadow.rule-*` token on `shadow` ([ADR-036](adr/adr-036-one-sided-rules-as-inset-shadows.md)), so the check sees it as a bound shadow. A shape whose stroke reads `strokeWidthType: "multiple"` is a defect: Penpot saves per-side widths and never draws them.

**What it does not see.**
- **Width and height.** A size set by layout (fill, auto-height text) is indistinguishable here from a typed one.
- **Stroke style** (the placeholder's dashes).
- **Text content.**
- **Whether a shadow has a surface to cast from.** A shape with no fill casts its shadow from its children ([ADR-024](adr/adr-024-control-and-field-internals.md) *Operational note*). So `shadow.focus-ring` bound to an unfilled box passes this check and still draws the ring around the text. An inset shadow on an unfilled shape is worse: it draws nothing, so a `shadow.rule-*` on an unfilled shape passes and shows no rule ([ADR-036](adr/adr-036-one-sided-rules-as-inset-shadows.md) §3).

Look at those by eye.

**What it is expected to report.** Anything not on this list is a defect:

| Shape | Reported | Why it stays |
|---|---|---|
| Anything named `Hue / …`, **on the Foundations board only** | `fill` | Its identity section shows both pairings side by side, which a token-bound shape cannot: it only shows the active theme. Everywhere else the hues are tokens (`color.identity.<hue>.subtle` / `.strong`, [ADR-025](adr/adr-025-identity-hues-as-tokens.md)), so **a `Hue /` shape anywhere else is a defect** |
| The Card's `Menu` | `typography(13px)` | The `⋯` glyph size has no token |
| `Day`'s `Number` where `Today` is Yes: two on `Day — states`, one in every `Calendar` | `typography(14px)` | Its weight is bound to `font.weight.semibold`, which replaces the `type.body` binding, and the catalogue has no 14/20/600 style. Its family is bound; its 14/20 is typed ([ADR-041](adr/adr-041-the-card-drawers-pickers.md) *Operational note*) |
| The `Logo` part's `Name`, on the part and every copy | `typography(20px)`; `typography(17.5px)` in `Sidebar lockup` and every sidebar | The wordmark is Poppins SemiBold at 20 (17.5 in the sidebar), and the catalogue has no 600 style at either size. Its family, weight and colour are bound; its 20/28 is typed ([ADR-049](adr/adr-049-the-logo.md) §8) |
| Anything named `Fixed / …`, or inside one, **on `Brand` only** | `fill`; on the `Fixed / Surface /` boards also their padding and gaps; on `Fixed / App icon preview` its corners | They show the logo in colours that must not follow the theme: the Light and Dark pairing, one colour, and the app icon. A token-bound shape only shows the active theme. Everywhere else the logo is bound to `color.logo.*`, so **a `Fixed /` shape anywhere else is a defect**. Unbinding one takes `fills = []` before the colour: a typed colour equal to the token's value keeps the binding ([ADR-049](adr/adr-049-the-logo.md) *Operational note*) |

**Shapes that exist only to document.** A shape whose name starts `Doc /` has **geometry** — size, padding, gaps — that belongs to this file's documentation rather than to the product. Examples are a swatch's dimensions, the gap between a swatch and its label, and the gap between eight avatars in a row. That geometry is not a token (ADR-023 *Out of scope*). Borrowing a product token for it, such as the card gap or field gap, is the coupling ADR-023 rejected: the documentation would reflow on Fine → Coarse, and a change to the card gap would move it. **Everything else about a `Doc /` shape is bound like any other shape: fill, stroke colour and width, radius, type, opacity and shadow.** That way the Light/Dark flip still tests the documentation. The check skips padding and gaps on `Doc /` shapes and nothing else. It cannot see sizes on any shape.

Widened from sizes alone to all geometry on 2026-11-20, before the foundations board was drawn, once it was clear the board's own groups need inner gaps that no product token describes.

**By eye — only where looking genuinely beats checking.** Use `export_shape` on the phone board (2.15) and on the state boards as whole compositions. For everything else the two checks above are stricter than an eye.

**Rules show in exports.** Until 2026-11-22 this paragraph said no export shows a one-sided rule, and to confirm one from its data or on the canvas ([ADR-027](adr/adr-027-column-header.md) *Operational note*). That was wrong. The canvas did not draw them either, because Penpot draws only a stroke's general width. Rules are now inset shadows ([ADR-036](adr/adr-036-one-sided-rules-as-inset-shadows.md)), and exports draw them like everything else, so look at them in an export.

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

**Stage 3 — the remaining seven parts** (7) ✅ *(2026-11-20)*

- [x] Avatar *(2026-11-20; question 4 fired and [ADR-025](adr/adr-025-identity-hues-as-tokens.md) closed it)* · [x] Count chip *(2026-11-20; question 4 fired and [ADR-026](adr/adr-026-count-chip.md) closed it)* · [x] Column header *(2026-11-20; question 4 fired and [ADR-027](adr/adr-027-column-header.md) closed it)* · [x] Banner *(2026-11-20; question 4 fired and [ADR-028](adr/adr-028-banner.md) closed it)*
- [x] Skeleton block *(2026-11-20; question 4 fired and [ADR-029](adr/adr-029-skeleton-block.md) closed it)*
- [x] Toast *(2026-11-20; question 4 fired and [ADR-030](adr/adr-030-toast.md) closed it)*
- [x] Presence cluster *(2026-11-20; question 4 fired, and [ADR-031](adr/adr-031-presence-cluster.md) closed it)*
- [x] Binding check, margins included, re-run on all ten parts' boards on 2026-11-20: only the Card's `Menu → typography(13px)` rows (five on its own board, three in the Toast board's cards) · [x] Light → Dark · [x] Fine → Coarse · [x] **Stage 3 closed**: all ten parts drawn, and question 4 fired before every one

**Stage 4 — the remaining fourteen screens** (14) ✅ *(2026-11-23)*

- [x] 2.2 Register *(2026-11-21; question 4 fired in round 2, and [ADR-032](adr/adr-032-auth-form-errors-and-password-rules.md) closed it. Two parts added, `Form message` and `Password rule`; binding check clean on 2.1, 2.2 and both part boards; four Theme × Density combinations looked at)*
- [x] 2.3 Workspaces · [x] 2.4 Workspace detail *(2026-11-21; question 4 fired in two rounds, and [ADR-033](adr/adr-033-workspace-screens.md) closed it. Three parts added, `Tile`, `Segment` and `Illustration`, and `Kind` on the Button; ten boards; binding check clean on all ten and the four part boards; Dark + Fine and Light + Coarse looked at)*
- [x] The sidebar parts · [x] 2.4's six boards redrawn *(2026-11-21; question 4 fired in round 2, and [ADR-035](adr/adr-035-sidebar-loading-failure-and-switcher-mark.md) closed it. `Sidebar row` and `Sidebar` added, ADR-034's five entries and `size.dot` added (200 tokens), the Foundations catalogue regenerated; binding check clean on the two part boards and all six boards; Dark + Fine and Light + Coarse looked at)*
- [x] Every one-sided rule redrawn as an inset shadow *(2026-11-22, [ADR-036](adr/adr-036-one-sided-rules-as-inset-shadows.md). Reported by the user from the canvas: no per-side stroke had ever drawn. Three tokens, six entries (206); 54 rules redrawn; the Foundations catalogue regenerated; binding check clean on the whole file)*
- [x] **2.5 Board — at rest** *(2026-11-23; question 4 fired on the header gap, and [ADR-037](adr/adr-037-board-at-rest.md) closed it with five decisions. `Add card` added and the Card's meta row fixed, no token added (206); binding check: only the expected card `⋯` rows on 2.5 and the Card board, nothing on `Add card — states`; Dark + Fine and Light + Coarse looked at)*
- [x] 2.6 Card drawer · [x] 2.7 + 2.8 Read-only / join gate · [x] 2.9 Cold load · [x] 2.10 Empty board
- [x] 2.11 Load failed · [x] 2.12 Gone · [x] 2.13 Session expired *(2026-11-23; question 4 fired in two rounds, and [ADR-038](adr/adr-038-board-states-group-d.md) closed it. Four token entries (210), `Field` added, five Illustration subjects, the Foundations catalogue regenerated; binding check: only the expected card `⋯` rows on 2.6 and 2.7, nothing on the other seven or the two part boards; Dark + Fine and Light + Coarse looked at)*
- [x] 2.14 Drag · [x] 2.15 Phone board *(2026-11-23; question 4 fired on three values and five decisions, and [ADR-039](adr/adr-039-drag-and-phone-board.md) closed them. No token, no part; the `Card` gains `Armed` and `Lifted`; four boards, 2.14 as a card drag and a column drag, 2.15 with a drag board; binding check: only card `⋯` rows on all four and on `Card — states`; Dark + Fine and Light + Coarse looked at)*
- [x] **Stage 4 closed** *(2026-11-23)*: all fifteen screens drawn, question 4 asked before every group

**Stage 5 — overlays** (~18), in the five groups of §7.3

- [x] **Group a** *(2026-11-23; question 4 fired on six values and fourteen choices, and [ADR-040](adr/adr-040-menus-and-the-confirmation.md) closed them. Ten token entries (220), four parts (`Glyph`, `Menu row`, `Menu`, `Confirmation`) and the Button's rose kind; the `Overlays` page with nine boards; the Foundations catalogue regenerated; binding check: only card `⋯` rows on the Overlays boards and nothing on the five part boards; Dark + Fine and Light + Coarse looked at)*: [x] Destructive confirmation · [x] Board `⋯` · [x] Column `⋯` + inline rename · [x] Card `⋯` · [x] Workspace `⋯` · [x] User menu
- [x] **Group b** *(2026-11-23; question 4 found no missing value and sixteen choices, and [ADR-041](adr/adr-041-the-card-drawers-pickers.md) closed them. No token (220); four parts (`Picker row`, `Day`, `Calendar`, `Picker`) and four `Glyph` looks; eight boards on `Overlays`, the drawer at 390 among them; the SRS's due-date rule edited first; binding check: only card `⋯` rows and today's `Number`; Dark + Fine looked at, the file left in Light + Coarse)*: [x] Date + assignee pickers · [x] Move card to…
- [x] **Group c** *(2026-11-24; question 4 fired in two rounds, on one value and then on two that resolved by rule, and [ADR-042](adr/adr-042-the-forms.md) closed them with twenty-two choices. Two token entries (222), five parts (`Swatch`, `Choice row`, `Address chip`, `Invite result`, `Form`), the `Input`'s `Content` axis and three `Glyph` looks; ten boards on `Overlays`, and *Invite people* on 2.4's heading on eight boards; the SRS's UC-03 and §4.3 edited first; the Foundations catalogue regenerated; binding check: only card `⋯` rows on the Overlays boards, nothing on the part boards; Dark + Fine looked at, the file left in Light + Coarse)*: [x] Create workspace · [x] Create board · [x] Board settings · [x] Invite people · [x] Create column, as a new column on the board
- [x] **Group d** *(2026-11-24; question 4 fired on one value, which resolved by rule, and twenty choices, and [ADR-043](adr/adr-043-the-lists.md) closed them. No token (222); two parts (`Person row`, `List`), six `Glyph` looks, six `Menu` looks and six confirmation wordings; ten boards on `Overlays`; the SRS's UC-06 and UC-10 edited first; binding check: nothing on the part boards or the new 2.4 and phone boards, only card `⋯` rows on the 2.5 and 2.15 boards; Dark + Fine looked at, the file left in Light + Coarse)*: [x] Workspace team panel · [x] Board members · [x] Add people, as the members list's second group · [x] Transfer ownership, as *Make owner* and a confirmation · [x] Presence overflow
- [x] **Group e** *(2026-11-24; question 4 fired twice, on the toast's edge and on the stack's gap, both resolved by rule, and fifteen choices, and [ADR-044](adr/adr-044-the-sidebar-sliding-over-and-the-toast.md) closed them. No token (222), no part (34); the `Toast` gains a hidden *Undo*; ten boards on `Overlays`; US-12c and SRS UC-12c edited first; binding check: only card `⋯` rows on the Overlays boards and the Toast board; Light + Fine, Light + Coarse and Dark + Fine looked at)*: [x] The sidebar sliding over · [x] Toast content and placement, with the archive toast's *Undo*
- [x] **Stage 5 closed** *(2026-11-24)*: every overlay drawn, question 4 asked before every group

**Stage 6 — final check**

- [x] The screens and their overlays moved into three pages by area, so the check reviews the final layout *(2026-11-25, [ADR-045](adr/adr-045-design-file-pages-by-area.md): Account, Workspaces and Board; 31 boards moved, each copy checked against its original before the original was deleted)*
- [x] Binding check clean on every board *(2026-11-25: every row it returned is on §7.5's expected list, on all five pages, about 26,000 shapes)* · [x] Light/Dark flip reviewed, page by page *(2026-11-26, in Coarse and in Fine)* · [x] Fine/Coarse flip reviewed *(all four Theme × Density combinations, in the order Light + Coarse → Dark + Coarse → Dark + Fine → Light + Fine; the file is left in Light + Fine)*
- [x] **Stage 6 closed** *(2026-11-27; [ADR-046](adr/adr-046-stage-6-final-check.md). Four findings, all taken as recommended: the avatar's initials on a dimmed card and the contents of a dragged column's place are **recorded as an exception to 4.5:1**, the twelve unfilled documentation boards were filled, and the Count chip board's clipped chips were left as a known fault. Four Penpot repairs made. **Not looked at by eye:** 17 of Board's 30 desktop screens in Dark + Fine, and all parts but `Button` in Dark + Fine; the data checks covered everything)*
- [x] The Board page split in four, after Stage 6, so no page is much heavier than Workspaces *(2026-11-27, [ADR-048](adr/adr-048-board-page-split-in-four.md): Board, Board — what opens over it, Card, Board — phone; 34 boards and 3 labels moved, each copy checked against its original before the original was deleted, and every page counted again at the end)*
- [x] 2.15 exported and the "6 cards visible" number confirmed by eye *(2026-11-23: 5 whole cards and 44px of the sixth in a column longer than the screen, the count ADR-037 revised it to; [ADR-039](adr/adr-039-drag-and-phone-board.md) §9)*

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
| 4 | Rejected | **Not animated** — instant return, toast, dim lifts. The toast: *Couldn’t move “X”. Someone changed the board at the same time.*, on boards *2.5 + a move that didn’t stick* and *2.15 + a move that didn’t stick* ([ADR-044](adr/adr-044-the-sidebar-sliding-over-and-the-toast.md) §12) |
| 5 | Drag-armed | Card variant `Armed`: `--shadow-overlay`, where the ramp over the hold ends. The ramp and its `scale(1.01)` are code only (§4.3). Shown on the Card's own board, since a desktop drag has no hold ([ADR-039](adr/adr-039-drag-and-phone-board.md) §5) |
| 6, 7, 8, 9 | Drag source, overlay, placeholder, column tint | Boards *2.14 Drag — a card* (all four), *2.14 Drag — a column* (6, 7, 8 at column size, nothing tinted) and *2.15 Phone — dragging a card*. 6 is the Card's `DragSource`, 7 its `Lifted` ([ADR-039](adr/adr-039-drag-and-phone-board.md) §3–§4, §10) |
| 10 | Drag-rejected | No channel — instant return |
| 11 | Draggable at rest | Cursor at fine; **nothing at coarse** — the one accepted gap in the whole vocabulary |
| 12, 15 | Hover, pressed | Card and Button variants — border colour steps, never a fill |
| 13 | Hover-equivalent at coarse | The persistent `⋯` (§1.5) |
| 14 | Focus-visible | Every focusable component variant |
| 16, 17 | Not satisfied, busy | Button variants (§4.2). Not satisfied can still be pressed, and runs the checks; busy reads *Signing in…* / *Creating account…* ([ADR-032](adr/adr-032-auth-form-errors-and-password-rules.md) §7–§8) |
| 18 | Selected | **Does not exist in Phase 1** — which is what released 2px border weight for the placeholder. The Active / Archived switch's current option is plum's *where you are*, not this state ([ADR-033](adr/adr-033-workspace-screens.md) §5) |
| 19 | Cold load | Board 2.9, and the cold-load boards of 2.3 and 2.4 ([ADR-033](adr/adr-033-workspace-screens.md) §10) |
| 20 | Background refresh | **Nothing** — what's on screen is correct |
| 21 | Empty, **six** flavours | Boards 2.3, 2.4 (Active **and** Archived), 2.10, the in-column surface, and the board members list's second group, on `List — states` (§6.5). Flavours 1 to 4 carry an illustration ([ADR-033](adr/adr-033-workspace-screens.md) §8) |
| 22, 23 | Load failed, gone | Boards 2.11, 2.12; 22 also on the load-failed boards of 2.3 and 2.4 ([ADR-033](adr/adr-033-workspace-screens.md) §10) |
| 24, 25 | Read-only, join gate | Boards 2.7, 2.8 — **one component, two settings**, and 2.7's banner carries Join |
| 26 | Access revoked | **Navigates you off the board**, plus a toast naming the reason. Nothing reverts — this is not 27. Four reasons: three wordings, and the voluntary case shows nothing. After `removed_from_workspace` it goes to `/workspaces`, not the workspace ([ADR-030](adr/adr-030-toast.md) §5–6). Drawn on *2.3 + removed from a workspace* (ADR-044) |
| 27 | Action forbidden | Toast; the optimistic change reverts via the in-flight dim. After a 403 the board reloads, so 2.7's banner shows: board *2.7 + an edit you can’t make* ([ADR-044](adr/adr-044-the-sidebar-sliding-over-and-the-toast.md) §13) |
| 28 | Connecting | Part of cold load (2.9) |
| 29 | Reconnecting | Presence cluster variant `Reconnecting` (§4.2): grey faces on `color.avatar-neutral`, the label after the seat, or in its place below 640px ([ADR-031](adr/adr-031-presence-cluster.md) §5) |
| 30 | Live | **No treatment** — colour is an event, not a texture |
| 31 | Remote change arrived | The movement channel; not a static frame |
| 32 | Remote deletion of the open card | The drawer closes with *“X” was deleted.*: board *2.5 + the open card deleted* ([ADR-044](adr/adr-044-the-sidebar-sliding-over-and-the-toast.md)) |
| 33 | Presence change | Presence cluster: a face added or removed, and the seat's count. The movement itself is code only ([ADR-031](adr/adr-031-presence-cluster.md) *Out of scope*) |
| 34 | Private board | Filled Phosphor lock, on board tiles (§1.4) |
| 35 | Archived board | **Reachable after all** (§6.3) — but it still spends no channel. An archived board appears only in the Archived list, where every row is archived, so nothing on the tile needs to mark it. The filter is the signal |
| 36 | Field invalid | Input variant; board 2.2 (the taken email). A missed password rule is the `Password rule` part’s `Missed`, and the field’s box turns rose with it. A form-level failure has its own part, `Form message`, shown on its board as Login after a 401 ([ADR-032](adr/adr-032-auth-form-errors-and-password-rules.md)) |
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
- **The banner:** [ADR-028](adr/adr-028-banner.md) — `color.background` for Join's focus ring, the bottom rule, `spacing.screen-x`, `spacing.banner-padding-y`, `spacing.inline-gap`, one appearance with Join, no icon; extends ADR-017 §9
- **The skeleton block:** [ADR-029](adr/adr-029-skeleton-block.md) — `color.skeleton` because `color.muted` is the column in Light, `size.skeleton-height` 60 / 72, card-sized blocks in real column panels, never on the bare canvas, no `animate-pulse`; supersedes ADR-018 §7's fill only and extends ADR-017 §9
- **The toast:** [ADR-030](adr/adr-030-toast.md) — inverted, no edge, `spacing.toast-padding`, `size.toast-width`, a Phosphor `X` close shown at rest, one appearance, the three access-revoked wordings; supersedes ADR-018 §10's row 26 destination for `removed_from_workspace` only and extends ADR-017 §9
- **The presence cluster:** [ADR-031](adr/adr-031-presence-cluster.md) — `spacing.avatar-overlap`, `shadow.avatar-ring`, `color.avatar-neutral` for the grey faces and the seat, `radius.full`, `spacing.seat-padding-x`, the round `+N` seat, the caps of three and five, you first, the label's place, three appearances; extends ADR-017 §9, supersedes ADR-018 §10's fill, ADR-025 §4's grey circle and ADR-019 §8's chip
- **Where the auth forms report errors, and Register's password rules:** [ADR-032](adr/adr-032-auth-form-errors-and-password-rules.md) — the `Form message` above the button, a taken email on its field, the four rules as a live `Password rule` checklist (`spacing.list-gap`, met in `color.done`), confirm password, *Sign in*, every wording, fifteen behaviour rules; settles ADR-024's open question, extends ADR-017 §9, supersedes ADR-018's Emerald ledger row
- **Drag and the phone board:** [ADR-039](adr/adr-039-drag-and-phone-board.md) — a short column takes a card only as far as it goes, *Add card* untinted during a drag, the `Card`'s `Armed` and `Lifted`, growth as code only, *Add column* below 640 in the board ⋯ menu, the phone's browser bands and typed 318, the count checked by eye; settles ADR-037's *Out of scope*, amends ADR-038 §10 below 640
- **The menus and the confirmation:** [ADR-040](adr/adr-040-menus-and-the-confirmation.md) — Stage 5's group a: one menu surface as a popover and a bottom sheet, rose for the deletes and *Leave board*, five actions that ask first, *Undo* after archiving kept, typing a workspace's name to delete it, the 448 dialog, the rose button, inline rename, *Sign out*, the `Overlays` page
- **The card drawer's pickers:** [ADR-041](adr/adr-041-the-card-drawers-pickers.md) — Stage 5's group b: every picker in place in the drawer, *Move card to…* too, so no sheet over a sheet; from the board, Move in the card ⋯'s popover or sheet; a column, then a place; plum on the current choice; a six-week calendar; a due date as a day at midnight UTC; supersedes ADR-021 §5's placement of *Move card to…* only
- **The forms:** [ADR-042](adr/adr-042-the-forms.md) — Stage 5's group c: the 448 dialog and the full-height sheet, a field of several lines and `spacing.control-padding-y`, minimum lengths as you type, the colour buttons and the stored hex value, two choice rows, *Invite people* with chips and results under the field, *Add column* as a new column on the board
- **The lists:** [ADR-043](adr/adr-043-the-lists.md) — Stage 5's group d: the team and a board's members in group c's frame, eight rows before they scroll, the `Person row`, a ⋯ per row, adding one person at a time, *Make owner*, the presence popover, two backend items; supersedes ADR-021 §5's placement of presence overflow only
- **The sidebar sliding over, and the toast:** [ADR-044](adr/adr-044-the-sidebar-sliding-over-and-the-toast.md) — Stage 5's group e: the slide-over on the phone, focus on the current row, the user menu's sheet over it below 640; a toast 16 from the bottom and right, up to three stacked, no timer; *Undo* as the Outline button, back to the board, teammates back from 2.12; *Couldn’t move “X”.* and five reasons; the board reloads after a refusal; closes Stage 5
- **The design file's pages:** [ADR-045](adr/adr-045-design-file-pages-by-area.md) — three pages by area, Account, Workspaces and Board, one row per screen with what opens over it; supersedes ADR-040 §1's `Overlays` page only
- **The Board page, split in four:** [ADR-048](adr/adr-048-board-page-split-in-four.md) — Board, *Board — what opens over it*, Card and *Board — phone*, each row moved whole; supersedes ADR-045 §1, §5's page and §7's `Board` table only
- **What this document's own open questions settled into:** [ADR-021](adr/adr-021-design-inventory-corrections.md) — archive, column reorder, overlay presentation, the read-only exit
- **Routes and data flow:** [frontend-architecture.md](frontend-architecture.md)
- **Product intent:** [001-PRD.md](../product/001-PRD.md) · **Acceptance per feature:** [001-user-stories.md](../product/001-user-stories.md)
