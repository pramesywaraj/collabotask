# ADR-033: The Workspace Screens — Tiles, the Active / Archived Switch, an Outline Button, and Empty-State Illustrations

- **Status:** Accepted — **§2's app bar contents and §4's tile counts superseded by [ADR-034](adr-034-side-menu.md), on 2.4 only** (2026-11-21). On 2.4, *CollaboTask* and the user menu move into the sidebar, and the bar holds the show/hide button and the title *All boards*. With the sidebar open, the grid shows 4 tiles across at 1440 and 2 at 1024. 2.3 keeps §2 and §4 as written. **§8's empty flavour 5 given its wording by [ADR-037](adr-037-board-at-rest.md)** (2026-11-23): *No cards yet.*, one line with no illustration, as §8 says. **§8's list of screens with a drawing widened, and §9's rules 5 and 8 reworded, by [ADR-038](adr-038-board-states-group-d.md)** (2026-11-23): drawings also sit on the board's join gate, load-failed, gone and session-expired screens, and a drawing names what is missing *or what happened*. Five subjects are added (`NoColumns`, `PrivateBoard`, `LoadFailed`, `BoardGone`, `SessionEnded`), all in the plum tint, none with a sparkle. **§7's menu and confirmation drawn by [ADR-040](adr-040-menus-and-the-confirmation.md)** (2026-11-23): the workspace ⋯ opens a 224-wide menu with one rose entry, *Delete workspace*. Its confirmation asks you to type the workspace's name before the rose button can be pressed, then takes you to 2.3. **2.4's heading gains *Invite people* by [ADR-042](adr-042-the-forms.md)** (2026-11-24): an Outline button after the member count, for workspace admins, hidden on the cold load like the count. It opens the Invite people form. The heading's other buttons stand. **§7's workspace ⋯ widened by [ADR-043](adr-043-the-lists.md)** (2026-11-24): it appears for every member, holding *Leave workspace* for everyone but the owner, who keeps *Delete workspace*. *N members* opens the team list. The file's Product Team counts are corrected to *11 members* and *6 boards* on 2.3 and 2.4. **§2's product name superseded by [ADR-049](adr-049-the-logo.md)** (2026-11-28): on 2.3's bar, *CollaboTask* in `type.title` becomes the logo's lockup, still on the left and still linking to `/workspaces`. Everything else stands as written.
- **Date:** 2026-11-21
- **Scope:** Boards 2.3 Workspaces and 2.4 Workspace detail (Stage 4, group b). It settles:
  - the surface both screens sit on, which [ADR-029](adr-029-skeleton-block.md) §4 left open
  - the tile: its size, rows, padding, swatch shape and swatch colour
  - the tile grid's gap and minimum width
  - the app bar's height
  - how the Active / Archived switch shows which list you are on
  - a second Button kind, **Outline**
  - where *Delete workspace* lives
  - the empty states, which now carry **illustrations**, and the **rules for drawing the next illustration** (§9)
  - the role words, every wording on both screens, and the behaviour rules
  - three new parts (`Tile`, `Segment`, `Illustration`) and a `Kind` axis on `Button`
  - **fifteen token entries** (179 → 194)

  It **extends [ADR-017](adr-017-ui-foundations.md) §9**: the code gains the variables in §14. It **supersedes two things, and nothing else in them**:
  - [ADR-018](adr-018-state-vocabulary.md)'s channel ledger row *"Plum | the one action, and where you are — including the focus ring"*. Plum now also **tints the empty-state illustrations**, a decorative use. This was the user's call, against the recommendation (§8).
  - [ADR-018](adr-018-state-vocabulary.md) §7's row 21, *"Per-surface copy in `--muted-foreground`. Zero channels"*, **for the whole-screen empty states only** (flavours 1 to 4). They carry an illustration as well as the sentence. Flavours 5 and 6 are unchanged.
- **Upstream edits:** none. The PRD, US-03, US-05, US-05b, US-07, US-08, US-09, US-12c and SRS UC-03, UC-05, UC-05b, UC-07 and UC-08 were checked against every choice here. PRD §4.1 asks that each empty screen carry exactly one obvious next action. An illustration adds no action, so it conflicts with nothing.

## Context

Question 4 was asked of both screens before drawing, in two rounds.

**What the API gives a tile was checked against the code first**, because a tile can show only what the list endpoints return:
- The workspace list returns `role`, `member_count`, `board_count` (active boards only) and `owner_id` ([workspace_response.go](../../../collabotask-backend/internal/delivery/http/response/workspace_response.go)).
- The board list returns `user_role` (empty when not joined), `access_status` (`JOINED` or `CAN_JOIN`), `member_count`, `card_count`, `visibility` and `background_color` ([board_response.go](../../../collabotask-backend/internal/delivery/http/response/board_response.go)). The counts come back for every board the list returns, including a private board an admin has not joined ([board_queries.go](../../../collabotask-backend/internal/repository/postgres/board_queries.go), `getUserBoardsInWorkspace`).
- **Neither returns names or faces**, so no tile shows avatars. The workspace detail returns the member list, so *N members* in 2.4's header is that list's length.

**Round 1 fired.** No value existed for:
- the app bar's height
- the tile's geometry and minimum width
- the gap between tiles ([ADR-023](adr-023-spacing-between-objects.md) *Out of scope* left it open)
- the swatch's shape and its colour role ([ADR-025](adr-025-identity-hues-as-tokens.md) *Out of scope* left the role to this stage)

There were also decisions with no value missing:
- **How the switch marks the current list.** ADR-018 closed state 18, *Selected*, because nothing can be bulk-selected.
- **A second Button kind.** An admin's grid can hold a dozen Join buttons, so Join cannot be the plum Primary that means *the one action*.
- **Where Delete workspace lives.**
- **What the empty states look like.**
- **Whether the owner reads *Owner* or *Admin*.**
- **Every wording.** None existed.

Nine choices went onto a rendered page, each option drawn in Light and Dark with the file's values, with a picker. **The user took seven recommendations.** They chose the **roomy tile** over the card-sized one, and the **left rail** over the top rail. They also asked for **an illustration on the empty states, "like Notion"**.

**Round 2 came out of that note.** The illustration had no line colour, grey, line weight, size or spacing. Four choices went on a second rendered page:

| Choice | Recommended | Chosen |
|---|---|---|
| Style | One ink line and one grey, spending no channel | **The same, with a plum tint**, against the recommendation |
| Which empty states | Whole-screen only | Whole-screen only |
| Where it sits | Above the sentence, aligned left | **Centred in the grid's place**, against the recommendation |
| Size | 128 | 128 |

The user's note on round 2 asked for the style to be written down, so that a later illustration can follow it. That is §9.

**The roomy tile needed four more values.** They follow from the pick, so they were listed on the round-2 page without a question: the tile's padding, the gap between its rows, a second skeleton height, and a re-measured minimum width (§3, §4).

## Decision

### 1. Both screens sit on `color.background`

The page is the same white (and in Dark, near-black) the forms use. This also fixes what ADR-029 §4 left open: a skeleton block may sit on the page.

| Pair | Light | Dark |
|---|---|---|
| Tile fill on the page | 1.00: set off by its edge | 1.30 |
| Tile edge on the page | 1.26 | 1.92 |
| Skeleton block on the page | 1.26 | 1.30 |

[ADR-030](adr-030-toast.md) already measured the toast over this surface. On `color.board-canvas` instead, a Light skeleton block and the tile's edge both measure 1.00.

### 2. The app bar

- **`size.app-bar` 56**, in `base`. It is derived, not chosen: ADR-017 §1 (900 − 88 − **56** − 16 − 36) and ADR-019 §7 (844 − 140 − **56** − 16 − 44) both subtract it, at both densities.
- Fill `color.background`, a bottom rule in `color.border` at `border.width`, and `spacing.screen-x` at the sides.
- The product name, *CollaboTask*, sits on the left in `type.title` and links to `/workspaces`.

  > *2026-11-28:* it is now **the logo's lockup**: the mark at 32, a gap of 8 and the name at 20 in Poppins SemiBold, deep plum. It still sits on the left and links to `/workspaces`. See [ADR-049](adr-049-the-logo.md) §7.
- The **user menu** sits on the right: an Avatar in a `size.target-min` box. The menu it opens is Stage 5.

> **On 2.4 the bar changes.** [ADR-034](adr-034-side-menu.md) (2026-11-21) puts a sidebar beside 2.4 and the board. On 2.4, *CollaboTask* moves to the sidebar's top row and the user menu to its bottom. The bar holds the show/hide button (while the sidebar is hidden, or below 1024) and the title *All boards*, in `type.title`. The heading below it is unchanged. On 2.3, which has no sidebar, this section stands as written.

### 3. The tile

One part serves both screens. A workspace tile is the same part with its rail hidden, because a workspace has no colour.

| Piece | Value |
|---|---|
| Padding | `spacing.tile-padding` 16 on all four sides, at both densities |
| Rows | title row, count row, and an action row only when there is an action |
| Gap between rows | `spacing.tile-row-gap` 8. No existing token has this role (ADR-022 §4) |
| Title row | the title in `type.card-title`, `color.card-foreground`; the private lock (Phosphor `Lock`, fill weight, `size.icon`, `color.muted-foreground`) `spacing.icon-gap` after it |
| Count row | `size.meta-row` tall: the role in `type.meta`, `color.muted-foreground`, then two Count chips, all `spacing.inline-gap` apart |
| Action row | one Outline button (§6): *Join* or *Unarchive* |
| Surface | `color.card`, a 1px `color.border` edge, `radius.surface` |
| Height | 84 at Fine and 88 at Coarse without an action; 128 and 140 with one |
| Swatch | a **left rail**, `size.board-rail` (4) wide and the tile's full height, in the hue's `.strong`, clipped by the corner radius |
| States | `State` = Rest · Hover · Pressed · FocusVisible. Hover and pressed step the edge (`color.border-hover`, `color.border-pressed`), as on the Card. Focus adds `shadow.focus-ring` |

**The rail reuses `size.board-rail`** because its role is the same, the board's colour strip, turned on its side. **It is `.strong`** because every `.subtle` hue fails 3:1 against the tile in both themes:

| | Light | Dark |
|---|---|---|
| `.subtle` against `color.card` | 1.11 – 1.21 | 1.58 – 2.94 |
| `.strong` against `color.card` | 5.15 – 9.62 | 12.55 – 13.69 |

The counts read *12 cards · 3 members* on a board tile and *5 members · 4 boards* on a workspace tile, with *1 card*, *1 member* and *1 board* in the singular ([ADR-026](adr-026-count-chip.md) §4). The count row takes `spacing.inline-gap`, which [ADR-027](adr-027-column-header.md) §2 offered.

### 4. The tile grid

- **`spacing.tile-gap` 8 at Fine and 12 at Coarse**, in both directions. The values match `spacing.column-gap`, but that gap is fixed by the phone strip's 44px peek (ADR-023 §2.2), and a tile grid has no peek. ADR-022 §4 keeps them apart.
- **`size.tile-min-width` 264**, in `base`, measured in Penpot. The action has its own row, so the widest row is the counts. With two-digit counts it is *Member · 88 members · 88 boards*: 51 + 8 + 89 + 8 + 74 = 230, plus 2 × 16 of padding = 262, rounded up to 264.
- In code the grid is `repeat(auto-fill, minmax(var(--tile-min-width), 1fr))`. That gives 5 tiles across at 1440, 3 at 1024 and 1 on a 390 phone.
- **A row stretches to its tallest tile**, the grid's default. A row with one Join tile makes all its tiles 128 tall. The user saw that on the round-1 page.

> **On 2.4, with the sidebar open, fewer tiles fit.** [ADR-034](adr-034-side-menu.md) (2026-11-21) takes 240px for the sidebar from 1024 up, so 2.4's grid shows 4 tiles across at 1440, 3 at 1280 and 2 at 1024. With the sidebar hidden, and on 2.3, the counts above stand.

### 5. The Active / Archived switch

A part, `Segment`, with `State` = Current · Other · Hover · Pressed · FocusVisible. Two instances sit side by side, `spacing.inline-gap` apart, between the header and the grid.

- **The current option:** text and a 1px edge in `color.primary`. That is 6.54 against the page in Light and 6.90 in Dark.
- **The other option:** `color.muted-foreground` text and no edge. Its hover and pressed step the edge ladder, and focus adds `shadow.focus-ring`.
- It is built from `size.control-height`, `spacing.control-padding-x`, `radius.control`, `border.width` and `color.background`. **No new token, and no new channel.**
- **This is not state 18.** ADR-018's ledger gives plum *"where you are"*, and the list you are on is where you are. It is not plum-tinted, because `color.primary-subtle` is Busy's fill.

### 6. The Button gains a second kind, Outline

The `Button` part has two axes now, `State` and `Kind` = Primary · Outline. Its six existing variants are `Primary`.

- **Outline:** fill `color.background`, a 1px `color.border` edge, `color.foreground` text. Hover and pressed step the edge, as ADR-024 §3 already said an outline button would. Focus adds `shadow.focus-ring`. **Busy** looks the same as Primary's Busy (ADR-018 §5).
- It is used for *N members*, *Join*, *Unarchive* and *Try again*.
- Primary stays *the one action* on each screen: *Create workspace* on 2.3 and *Create board* on 2.4.
- In Dark, an Outline button inside a tile is darker than the tile (1.30), because its fill is the page. The focus ring's offset is drawn in that colour, so this is what keeps the ring clean. The edge is decorative (1.26 in Light), and the label identifies the button, as on shadcn's own outline.

### 7. Delete workspace sits behind a `⋯` in the heading row

The column header's `DotsThree`, at `size.icon` in a `size.target-min` box, after *Create board*. **Only the owner sees it.** For everyone else it is absent, not disabled (US-06d, ADR-018 §5). It opens a menu with *Delete workspace*, which leads to confirmation 38. The menu and the confirmation are Stage 5. **Inventory §3 gains a *Workspace `⋯`* menu.**

### 8. Empty states, and the illustration

- **Which ones:** the whole-screen empty states, flavours 1 to 4 (inventory §6.5). Flavours 1, 2 and 3 are drawn here. Flavour 4, a board with no columns, gets one when 2.10 is drawn, measured against the board canvas then. Flavours 5 (an empty column) and 6 (the add-people panel) keep one line: a 128 drawing does not fit a 280 column or a panel list.
- **Where:** centred in the grid's place. A block with `spacing.section-gap` above and below holds the illustration, `spacing.illustration-gap` (16), and the sentence, centred, in `type.body` and `color.muted-foreground`.
- **The one button is still the heading's** *Create workspace* or *Create board*. The empty state adds no button of its own.
- **Size:** `size.illustration` 128, at both densities. A drawing is not a touch target.

  > *2026-11-27:* **192**, not 128. [ADR-047](adr-047-illustration-size-192.md) supersedes this size and §9 rule 1's box. The line stays 2px, the colours are unchanged, and the shapes in the part now scale with their box.
- **Colours:**

| Token | Light | Dark | On the page | Ink on it |
|---|---|---|---|---|
| `color.illustration-ink` | `#57534E` | `#A8A29E` | 7.63 / 7.83 | — |
| `color.illustration-tint` | `#F4F0FA` | `#3B2F57` | 1.12 / 1.63 | 6.79 / 4.81 |
| `color.background` (paper) | `#FFFFFF` | `#0C0A09` | — | 7.63 / 7.83 |

The line is `border.width-illustration` 2. It is not `border.width-strong`, which means *where it will land*.

**The tint is plum, and that is the user's call.** The recommendation was a grey. The page showed the cost once: the tint is `color.primary-subtle`'s value (Busy's fill), on a screen with nothing busy, and plum also means *the one action*, which is the Create button beside it. `color.illustration-tint` holds the same values as **literals**, not `{color.primary-subtle}`. A change to Busy's fill then does not repaint the drawings, and the drawings are one place to change.

The drawings are decorative, so WCAG 1.4.11 does not require 3:1. The ink passes anyway.

### 9. The illustration style: rules for drawing the next one

Every illustration in the app follows these rules. A new one is a new `Subject` on the `Illustration` part.

1. **The box.** 128 × 128, on whole or half pixels. The subject sits roughly centred and fills most of the box, with at least 8px clear at the edges. Only accents go nearer.
2. **One line.** 2px (`border.width-illustration`), round caps and round joins, in `color.illustration-ink`. No second weight and no dashes, because dashes mean *where it will land*.
3. **Two fills, and only two.**
   - `color.illustration-tint` goes on one or two shapes: the part that shows the emptiness, such as the pot, the board's columns, or the inside of the box.
   - `color.background` ("paper") goes on shapes that must hide the lines behind them, such as leaves, flaps or a box front.
   - Nothing else: no identity hue, no amber, rose or emerald, no gradient, no shadow, no transparency.
4. **Outlines.** Every shape has the ink line, except a tinted shape that sits inside an outlined one, like the board's columns. That shape may go without.
5. **The subject.** One everyday object, shown empty, that names what is missing:
   - *No workspaces* is a sprout in a pot: nothing has grown yet.
   - *No boards* is an empty board with three bare columns.
   - *Nothing archived* is an open, empty box.

   Never a person or a face. Never words or numbers inside the drawing. Never anything that looks like a control, such as a button or a plus sign that seems clickable.
6. **Accents.** At most two small sparkles (a plus of two lines, 8 to 12 long) or one wisp. They mean *new* or *air*. They never move: in this app, movement means *a teammate did this* (ADR-018 §8).
7. **The view.** Flat front, or a simple three-quarter view, drawn with clean geometry rather than a hand-drawn wobble. Corners of 2 to 6 belong to the drawing, not to the radius tokens.
8. **In the file.** Each drawing is one variant of `Illustration`, `Subject` = a PascalCase name for what is missing (`NoWorkspaces`, `NoBoards`, `NothingArchived`). Every stroke and fill is bound to a token. The geometry is the drawing's own, like a Phosphor glyph, so it is drawn as paths (§13).
9. **Measure it where it sits.** Measure the ink and the tint against the surface the drawing sits on, in both themes. On the board canvas (2.10), the Light tint is 1.12 and the ink 6.08.
10. **In code.** An inline SVG with `aria-hidden="true"`, because the sentence carries the meaning. Strokes use `var(--illustration-ink)`, fills use `var(--illustration-tint)` or `var(--background)`, and `stroke-width="2"`, `stroke-linecap="round"`, `stroke-linejoin="round"`.

### 10. Cold loads and load failures

- **Cold load (19).** Skeleton block instances stand in the tile grid, one per tile, directly on the page. That is allowed by ADR-029 §4, at 1.26 / 1.30. **None sits inside a tile**, where Dark measures 1.00. A block is a tile with no action: `size.tile-skeleton-height`, 84 at Fine and 88 at Coarse, in the density sets. That is the second height option 2B listed. The part keeps its one height, and these instances bind the second.
- **On 2.4 the header waits for its data.** The name, role, description, *N members* and `⋯` all depend on the response, so they show nothing until it arrives: a block has one height, so a heading-shaped block was not drawn (ADR-029 §5). *Create board* and the switch draw at once, because every member can create a board.
- **Load failed (22).** One sentence in `type.body`, `color.foreground`, and an Outline *Try again*, in one row, `spacing.inline-gap` apart, in the grid's place. Zero channels, as ADR-018 §7 says.

### 11. Roles and behaviour

- **Workspace roles read Owner · Admin · Member.** *Owner* is worked out from `owner_id`, which both the list and the detail return. The API's own role is `ADMIN` or `MEMBER`, and the owner is always an admin. *Owner* tells the owner why they alone have a `⋯`.
- **Board roles read Owner · Member**, from `user_role`.
- A **joined** tile shows your role.
- A **`CAN_JOIN`** tile shows no role, and shows *Join* in its action row.
- An **archived** tile cannot be opened, so it has no hover and no pressed. It keeps its role and counts. It shows *Unarchive* only to the board owner or a workspace admin, and is absent for everyone else. It never shows *Join*.
- **Create board** is shown to every workspace member (UC-07).
- **Pressing a tile** opens the workspace or the board. A `CAN_JOIN` board opens read-only with the banner if it is workspace-visible, or at the join gate if it is private (boards 2.7 and 2.8, [US-09](../../product/001-user-stories.md)).
- **Pressing *Join*** joins you, and the tile then shows your role (US-09). Whether it also opens the board is not decided here.
- **Pressing *Unarchive*** moves the board back to Active (US-12c).
- **The switch is a filter, not a route** ([ADR-021](adr-021-design-inventory-corrections.md) §1). Only the current option looks current.

### 12. The wordings

| Where | Wording |
|---|---|
| App bar | *CollaboTask* |
| 2.3 heading and button | *Workspaces* · *Create workspace* |
| Workspace tile counts | *1 member* · *3 members* · *1 board* · *4 boards* |
| Board tile counts | *1 card* · *12 cards* · *1 member* · *3 members* |
| 2.4 members button | *5 members* (*1 member*) |
| 2.4 button | *Create board* |
| Switch | *Active* · *Archived* |
| Tile actions | *Join* · *Unarchive* |
| Workspace `⋯` | *Delete workspace* |
| Empty 1 (2.3) | *You’re not in a workspace yet. Create one, or ask a teammate to invite you.* |
| Empty 2 (2.4, Active) | *No boards to show here yet.* |
| Empty 3 (2.4, Archived) | *No archived boards to show.* |
| Load failed (2.3) | *Your workspaces didn’t load.* |
| Load failed (2.4) | *This workspace’s boards didn’t load.* |
| Retry | *Try again* |

### 13. What the file draws

**New parts, on `Components`:**
- **`Tile`**: `State` = Rest · Hover · Pressed · FocusVisible, on `Tile — states`. The main is a joined board tile with its lock and action row hidden.
- **`Segment`**: `State` = Current · Other · Hover · Pressed · FocusVisible, on `Segment — states`.
- **`Illustration`**: `Subject` = NoWorkspaces · NoBoards · NothingArchived, on `Illustration — states`.
- **`Button`** gained `Kind`: five Outline variants (Rest, Hover, Pressed, FocusVisible, Busy) on a second row of `Button — states`.

**Ten boards on `Page 1`:**

| 2.3, at x 3280 | 2.4, at x 4920 |
|---|---|
| *2.3 Workspaces*: four workspace tiles, including *1 member · 1 board* | *2.4 Workspace detail*: six board tiles. Two can be joined, one of them private; two are private. The viewer is the owner |
| *— cold load*: five blocks | *— archived*: three archived tiles, each with *Unarchive* |
| *— empty*: the sprout | *— cold load*: blocks; the header waits |
| *— load failed* | *— empty*: the empty board |
| | *— archived, empty*: the empty box |
| | *— load failed* |

The **overlays are not drawn**, only what opens them: the user menu, *N members*, the two Create buttons and the `⋯`. They are Stage 5.

### 14. The code

- §9 gains `--app-bar: 56px`, `--tile-min-width: 264px`, `--tile-padding: 16px`, `--tile-row-gap: 8px`, `--tile-gap` (8 / 12), `--tile-skeleton-height` (84 / 88), `--illustration-size: 128px`, `--illustration-gap: 16px`, and `--illustration-ink` and `--illustration-tint` in both modes.
- The Outline button is shadcn's `outline` variant, restyled to §6: `bg-background`, the edge ladder on hover and active, and ADR-024's focus ring.
- The switch is a single-choice toggle group (shadcn `ToggleGroup type="single"`) restyled to §5, not tabs: it filters one grid rather than switching panels.
- The tile is a link (`<a>`) for a joined board or a workspace. An archived tile is not a link. *Join* and *Unarchive* are buttons inside it, so the link wraps only the title.
- The illustrations are inline SVG components (§9 rule 10).

## Options considered

### What the pages sit on
**`color.board-canvas`**, to match the board screen. Rejected: a Light skeleton block and the tile's edge both measure 1.00 on it, against ADR-029 §4.

### The tile's size
**Card geometry**, recommended: padding 12 × 8, one title line and a meta row, 60 / 72 tall, which already equals `size.skeleton-height`. The action sat on the right, centred. It needed one new token (`size.tile-min-width` 320). The user chose the roomy tile for its calmer reading.

### The gap between tiles
**Reuse `spacing.column-gap`.** One fewer token, but the tile grid would move whenever the phone strip's peek moved the column gap.

### The swatch
- **A top rail**, recommended: the same rail, in the same place, as on the board itself.
- **A leading 16px square**: it reads as an icon, beside the lock, which is the icon channel's only Phase-1 meaning (ADR-018 §9).

### The switch
- **shadcn Tabs** (a muted track with a white current segment): 1.09 in Light, so the current option cannot be seen without shadcn's shadow, and shadow means *lifted*.
- **A 2px plum underline**: it spends the 2px weight ADR-018 gave to *where it will land*.

### The second button kind
- **Primary everywhere**: plum on every Join, so plum stops meaning *the one action*.
- **Plum link text**: no edge to step for hover or pressed, and a 24px target at Fine.

### Delete workspace
- **A rose outline button in the heading row**: always in view for the one person who can press it, right next to *Create board*.
- **A section at the bottom of the page**: below a grid that grows with every board.

### The empty states
**A centred block with its own button**: it hides the heading's button on the first screen a new account sees, and needs a new gap.

### The illustration
- **One ink line and one grey**, recommended: it spends no channel. The grey would have been `#E7E5E4` / `#292524`.
- **Grey shapes with no line**: at 1.26 / 1.30 the shapes hardly read.
- **On all six empty states**: a drawing in a column is taller than three cards.
- **Only on the first screen**: flavours 2 and 3 look plainer beside it.
- **Above the sentence, aligned left**, recommended: it keeps the sentence beside the Create button it points at. The user chose centred, like Notion.
- **160**: nearly half a phone's width.

### Workspace roles
**Admin · Member**, the enum as it is: the owner would read *Admin*.

## Out of scope (deferred, recorded)

- **The overlays** these screens open: the user menu, the workspace team panel, Create workspace, Create board, the workspace `⋯` menu, and confirmation 38. Stage 5.
- **Flavour 4's illustration**, on 2.10.
- **These screens at 390.** ADR-020 §5 keeps one phone board, 2.15. The grid drops to one column, and the heading row wraps.
- **The viewer as a plain member.** Not drawn. §11 states what changes: no `⋯`, and *Unarchive* only on boards they own.
- **The workspace description on its tile.** The list returns it, but inventory §1.3 does not put it on the tile. It shows in 2.4's header.
- **The `archived` query parameter** (SRS UC-08). The one open backend item for this screen. It decides whether the Archived list can be filled with real data, not whether it can be drawn.
- **Tile order.** The board list is newest first (`ORDER BY b.created_at DESC`). No sort control is drawn.

## Downstream edits this ADR requires

1. **[ADR-017](adr-017-ui-foundations.md): annotation.** Its Status line records §9's eleventh extension, and a note under §9's opening points here.
2. **[ADR-018](adr-018-state-vocabulary.md): annotation.** Its Status line records both supersessions. Notes go under the channel ledger and under §7's table.
3. **[ADR-021](adr-021-design-inventory-corrections.md): annotation**, under §1's interface paragraph: the switch's look, the archived tile's actions, and the empty states.
4. **[ADR-023](adr-023-spacing-between-objects.md): annotation**, under *Out of scope*'s tile-gap bullet.
5. **[ADR-024](adr-024-control-and-field-internals.md): annotation**, under *Out of scope*'s button-kinds bullet.
6. **[ADR-025](adr-025-identity-hues-as-tokens.md): annotation**, under *Out of scope*'s swatch bullet.
7. **[ADR-026](adr-026-count-chip.md): annotation**, under *Out of scope*'s tile bullet.
8. **[ADR-027](adr-027-column-header.md): annotation**, under §2's note on who took `spacing.inline-gap`.
9. **[ADR-029](adr-029-skeleton-block.md): annotation**, under §4's rule and *Out of scope*'s cold-load bullet.
10. **Penpot `Collabotask App`.** **Done 2026-11-21.**
    - Fifteen token entries: seven in `base`, two in each theme set, two in each density set. The catalogue reads 194.
    - The parts and boards in §13.
    - The Foundations catalogue gained eleven rows. The two theme rows' swatches were rebound to their own tokens. The heading, the section name, all three table titles and the *as of* date were updated.
    - **Binding check:**
      - All ten boards and the four part boards return nothing.
      - The Foundations board returns only its expected rows: 48 `Hue /` swatches and the Card's five `Menu → typography(13px)`.
      - Three bindings had dropped and were rebound before the check came back clean (*Operational note*).
    - **Flips:**
      - Dark + Fine: 2.4, 2.3 empty and 2.4 cold load.
      - Light + Coarse: 2.3 and 2.4. Tiles are 272 wide with a 12 gap, 88 or 140 tall. The user menu is 44, and the skeleton blocks are 88.
      - Restored to `theme-light` + `density-fine`.
11. **[`design-file-inventory.md`](../design-file-inventory.md):**
    - §1.3 and §1.4 match this ADR.
    - §3 gains the workspace `⋯` menu.
    - §4.2 gains the three parts and Button's `Kind`.
    - §6.5 records which flavours carry an illustration.
    - §7.2, §7.3 and §7.6 record the stage.
12. **`CLAUDE.md` › `## Now`:** a pointer.
13. **`docs/build-log.md`:** no entry, following ADR-020 to ADR-032. Stage 4 is not closed.
14. **PRD, user stories, SRS, `CONTEXT.md`, [frontend-architecture.md](../frontend-architecture.md):** no change. No scope, contract or domain term moved.

## Operational note

**Setting an instance text's characters does not re-measure it.** A Count chip set to *12 cards* kept the 11px width of *12*, and the text overflowed the chip. Re-applying the typography, changing the font size, or flipping `growType` twice in one call did nothing. What works is two calls:
1. Set the characters, `growType = 'fixed'` and a rough `resize()`.
2. In a **separate** call, set `growType = 'auto-width'`.

The text and every auto-sized parent then measure correctly. Plain texts on a screen were treated the same way, to be safe.

**`resize()` turns a layout child back to a fixed size.** The skeleton instances were set to `fill` and then resized to 84 tall, and they stayed 264 wide in the grid. Set `fill` after the last `resize()`. The same thing happened to the Segment mains' `auto` width.

**Adding a variant axis and variants:**
- `variants.addProperty()` names the axis *Property 2*, with the value *Value 1* on every variant. Rename it with `renameProperty(pos, name)`, then `setVariantProperty(pos, value)` on each.
- `component.addVariant()` copies that variant to the container's right edge, with a fresh *Value N*. The container grows to fit, so resize it after moving the copies.

**`penpot.createPath()` with `d` draws a glyph without `createShapeFromSvg`**, so nothing lands in the board under the viewport. The path's coordinates are page coordinates:
- A Phosphor glyph comes in at 256 scale and is resized to its 16 box.
- An illustration designed on a 128 grid is shifted by the board's origin before it is set. Every command was absolute, and arcs (`A`) shift only their end point.

**A text's alignment is `align`, not `textAlign`.** Setting `textAlign` throws *object is not extensible*. Every write before it in the same call still lands.

**`switchVariant` swaps in the target variant's text unless that text was overridden.** Switching the *Active* segment to `Other` made it read *Archived*. Set the labels again afterwards.

**A token fix must skip the variant container.** It shares its component's name. A search for `Tile` boards with an unbound stroke width matched the container, and its own 2px frame was bound to `border.width`. Penpot has no call that removes a token. Writing the stroke back by hand (`strokes = [...]`) removed the binding, and the frame was restored to 2px.

**Re-applying a token the main already holds drops the binding**, as [ADR-032](adr-032-auth-form-errors-and-password-rules.md) found for text fills. The periwinkle rail on two tiles (the main's own hue) came back unbound. A second `applyToken` in a separate call held.

**An export of a freshly made `— states` board can fail** with *locator… hidden svg 0.01*. Exporting the main inside it, or any screen board, worked. The board was not broken.

**"Cannot modify a page that is not currently active"** means the Penpot tab changed page underneath the run, here to `Components`. Reopen the page in its own call and carry on.
