# ADR-038: The Board's Other States — the Card Drawer, Read-only, the Join Gate, Cold Load, Empty, Load Failed, Gone and Session Expired

- **Status:** Accepted — **§10's *"No button of its own"* holds from 640 up only, by [ADR-039](adr-039-drag-and-phone-board.md)** (2026-11-23): below 640 the bar has no *Add column*, so there the empty board's centred stack ends in a Primary *Add column* button, placed as §12 places a button. Recorded for the code; 2.10 is not drawn at phone width. Everything else stands as written.
- **Date:** 2026-11-23
- **Scope:** Question 4 for Stage 4's group d: boards 2.6 to 2.13, each drawn as a difference from 2.5 ([ADR-037](adr-037-board-at-rest.md)). It settles:
  - the card drawer at 640 and up: how it sits over the board, its surface, width, inside, where Move and Delete go, and how it shows a due date
  - where 2.7's banner sits, and what 2.8's gate looks like
  - what a cold board draws at once, and how far its loading columns reach
  - one centred stack for every board with nothing to show (2.8, 2.10, 2.11, 2.12), with a drawing in each
  - what 2.12 says, and how 2.13 hands you over to Sign in and back
  - four new drawings, on screens that are not empty

  **Four token entries are added (206 → 210):** `size.drawer-width` and `size.field-label` in `base`, and `shadow.rule-left` in both theme sets. One part is added, `Field` (19 parts), and the `Illustration` part gains five subjects (8). It **narrows [ADR-021](adr-021-design-inventory-corrections.md) §6**: 2.7 and 2.8 keep its data rules and the same *Join board* button, but 2.8 is not drawn with the Banner part. It **widens [ADR-033](adr-033-workspace-screens.md) §8–§9**: drawings now also sit on 2.8, 2.11, 2.12 and 2.13, and rule 5 names what happened rather than what is missing. It **narrows [ADR-018](adr-018-state-vocabulary.md) §7 for one thing**: the drawings on the two failure screens keep their plum tint, which was the user's call over an uncoloured drawing. It **rewords ADR-037 §2**: the columns' tops touch whatever is above them, which on 2.7 is the banner.
- **Upstream edits:** none. Nothing here touches an endpoint, the schema or a permission rule. 2.13's way back after signing in is a frontend change, recorded in `frontend-architecture.md`. One backend gap is recorded and not fixed (§11).

## Context

Inventory §7.3 draws group d as differences from 2.5: the card drawer (2.6), the read-only board and the join gate (2.7, 2.8), and the five whole-board states (2.9 to 2.13).

Question 4 was asked of every value these boards need, against the live 206 tokens, before anything was drawn. **Most of it mapped.** The dimming is `color.scrim` and the lift `shadow.overlay`. The drawer's text is `type.title`, `type.description` and `type.meta`. *Close* is Phosphor `X` in a `size.target-min` box, as on the Toast. The banner is the `Banner` part, and cold-load blocks are the `Skeleton block` part. The empty board's drawing follows ADR-033 §9. *Try again* is the Outline `Button`, and *Sign in* the Primary one.

**Three values had no token**, so question 4 fired for the sixteenth stage out of sixteen: the drawer's width, the width of its field labels, and a rule on its left edge. The file had rules for the top, right and bottom of a shape ([ADR-036](adr-036-one-sided-rules-as-inset-shadows.md)), not the left.

**Fourteen decisions had never been made.** All fourteen went on one rendered page, each option drawn in Light and Dark with the file's values, and the user took every recommendation. Their notes on six picks opened a second round:
- *Delete card* in red. That needed no decision: rose is already ADR-018's colour for destructive actions.
- **A centred stack with a drawing, on 2.8, 2.11, 2.12 and 2.13.** ADR-033 §8 allowed drawings on empty screens only, and §9's rule 5 says a drawing shows something *empty*. None of these four is empty. Two are failures, where ADR-018 §7 spends no colour.
- **The cold load's skeleton across the whole board.**

Round 2 put eight questions on a second rendered page, and flagged the ADR-018 §7 conflict once. **The user took six recommendations and overrode two:** the plum tint goes on all four new drawings, failures included, and 2.13 is left-aligned like Login.

Three facts were found while checking the code:
- **Deleting a workspace sends no live message.** `DeleteWorkspace` deletes and returns. Someone with one of its boards open learns it only when their next request comes back *not found* (§11).
- **A workspace admin who hasn't joined a private board** can read its name, colour and member count. Its columns return *join required*, and the live room quietly refuses them, so 2.8 has no presence (§8).
- **Sign-in has no way back.** The 401 handler sends you to `/login` and drops where you were (§13).

## Decision

### 1. The drawer sits over everything, and dims the board

- **Full height, on the right, over the sidebar and the app bar.** The whole screen behind it takes `color.scrim`.
- **The board can't be used while it is open.** Escape, the *Close* button, a press on the dimmed board, or the browser's back button closes it. Focus stays inside it.
- **It behaves as the phone's full-height sheet does**, so one piece of code serves both (ADR-019 §3).
- **The open card is not marked on the board.** A *selected* look (ADR-018 state 18) does not exist in Phase 1.

The rejected option: a panel beside a live board, under the bar, with no dimming. It hides the right-hand columns without saying so, focus can wander between two places, and it behaves differently from the phone's sheet.

### 2. Its surface: the popover colour, a left rule and the overlay shadow

- **`color.popover`**, with a 1px rule down its left edge, **`shadow.rule-left` (new)**, and **`shadow.overlay`**.
- **`shadow.rule-left`** is an inset 1px in `{color.border}` with x offset `{border.width}`, once per theme set, like its three partners (ADR-036). Penpot draws a positive offset on the left edge.
- **A shape binds one shadow token**, so the rule sits on an inner `Panel` and the overlay shadow on the outer `Drawer`, both filled `color.popover`. That is the route ADR-037 §4 used for *Add card*'s footer.
- **Measured in Dark**, the rule is 1.70:1 against the drawer and 1.86 to 1.99 against the dimmed board. In Light the drawer itself is 2.15 to 2.64 against the dimmed board.

The rejected options: `color.background` and a shadow, which is shadcn's Sheet as shipped and measures 1.00 to 1.04 against the dimmed board in Dark; and `color.popover` with no rule, 1.10 against the dimmed cards in Dark.

### 3. It is 560 wide: `size.drawer-width` (new, `base`)

- About **60 characters** of description a line, in the middle of the comfortable 45 to 75. The description is the only `type.description` (16/24) text in the app, so the drawer is sized for it.
- It leaves 640px of dimmed board at 1440, 224 at 1024 with the sidebar open, and 80 at 640. **Below 640 the drawer is the full-height sheet**, so this width only applies from 640 up.

The rejected widths: 480 (about 51 characters) and 640 (about 69, and at a 640 screen it covers everything).

### 4. The inside: labels on the left, the screen's inset by rule

- **The padding is `spacing.screen-x` (16), bound by rule.** Below 640 the drawer *is* the screen, so the sheet and the drawer share one inside.
- **Blocks are `spacing.section-gap` (24) apart**: the head, the fields and the description.
- **The head** is the title in `type.title`, wrapping, then *Close* (Phosphor `X` at `size.icon` in a `size.target-min` box), `spacing.inline-gap` apart.
- **The fields**, `spacing.list-gap` (4) apart. Each is a label in `type.body` and `color.muted-foreground`, **`size.field-label` (96, new, `base`) wide**, then its value, a `Field` instance filling the rest.
- **The description**: *Description* in the label style, then the text in `type.description`, `spacing.field-inner-gap` (8) apart.
- **The foot**: *Created Nov 18 · Updated Nov 22* in `type.meta` and `color.muted-foreground`, then *Delete card* (§5). It has a top rule (`shadow.rule-top`), `spacing.screen-x` at the sides, and **`spacing.banner-padding-y` (8) above and below**, bound by rule. It is a strip holding one row of controls, as the banner is. The value is what the decision page showed. The token was chosen while drawing.
- **At rest the title is plain text.** Pressing it turns it into the Input part's InlineEditing look (state 37). The description works the same way.

The rejected options: 24 padding, which needs `spacing.drawer-padding` and wastes 8px a side on a phone; and labels above their values, as on the forms, which takes about 224px for three fields instead of 116.

### 5. Move is the *Column* field; *Delete card* sits at the foot, in rose

- **A *Column* field shows the card's column**, *In Progress*. Pressing it opens *Move card to…*, which [ADR-019](adr-019-responsive-touch-scope.md) §6 requires from the drawer at every width. The drawer also says which column the card is in. The word *Move* shows in the picker's title and the field's spoken name.
- **The *Column* field carries a caret** (Phosphor `CaretUpDown`), because it is the one field that opens a picker of its own. *Assignee* and *Due date* expand in place below their row (ADR-021 §5).
- ***Delete card* is a quiet row at the foot**: Phosphor `Trash` and the words in **`color.destructive-text`**, 6.29:1 on the drawer in Light and 6.50:1 in Dark. **That was the user's note**, so the action is seen for what it is. Rose is ADR-018's colour for destructive intent, so nothing is narrowed. Pressing it still opens the confirmation (state 38) before anything is deleted.

The rejected options: an Outline *Move to…* and a bin button in a top row; and both behind a ⋯.

### 6. The drawer shows a due date in words, then the date

| The due date is | The drawer shows | The card shows | Colour |
|---|---|---|---|
| before today | *Overdue · Nov 20* | *Overdue · Nov 20* | `color.overdue` |
| today | *Due today · Nov 23* | *Due today* | `color.due-soon` |
| tomorrow | *Due tomorrow · Nov 24* | *Due tomorrow* | `color.due-soon` |
| later | *Nov 28* | *Nov 28* | `color.foreground` |
| none | *No due date* | (no meta row) | `color.muted-foreground` |

The words come first, in the order *Overdue · Nov 20* already uses. ADR-037 §5's one function gains a longer form for the drawer. *Unassigned* is the empty Assignee, also in `color.muted-foreground`.

The rejected options: the card's words only, which hide which day *today* is; and the plain date, which drops the lateness the card shows.

### 7. 2.7, the read-only board: the banner under the rail

- **The bar, the rail, then the `Banner` part, then the columns.** The bar and rail stay together, as on 2.5. The banner keeps its bottom rule, which in Dark is its only edge against the canvas (ADR-028).
- **The columns' tops touch the banner.** ADR-037 §2's *"the columns' tops touch it"* now reads *touch whatever is above them*.
- **Everything that changes the board is absent** (ADR-018 §5): *Add column*, every column ⋯, every *Add card*, and every card ⋯. The card ⋯'s only remaining entry, *Open*, is what pressing the card already does.
- **The board ⋯ stays**, because *Members…* is open to anyone who can read the board. Presence stays.
- **The board is not in the sidebar's list**, which holds only the boards you've joined (ADR-034), so no row is current.

The rejected option: the banner between the bar and the rail. Its rule then sits on the rail and does nothing.

### 8. 2.8, the join gate: a centred stack

- **The bar** shows the board's name and the filled lock in `color.muted-foreground`, as its tile and sidebar row do. **No presence, no ⋯, no *Add column*.** The rail is the board's colour, which the server returns.
- **The canvas holds one stack, centred both ways** (§12): the `PrivateBoard` drawing, *This board is private. Join it to see its cards. Its members will see that you joined.*, and a Primary ***Join board***.
- **The user's note** asked for the stack and for the lock to become a drawing. It was confirmed before drawing.
- **This narrows ADR-021 §6**, which made 2.7 and 2.8 *"one component with two settings"*. Read literally, 2.8 was the banner over a bare canvas, which reads as a board that failed to draw. 2.7 and 2.8 keep §6's data rules, its roster rule and the same button.

### 9. 2.9, the cold load: the frame at once, panels to the edge

- **The sidebar, the bar's name and the rail draw at once.** The board's name and colour come from the board list the sidebar already loaded. In code the board query starts from that list entry. A link opened straight to a board, before any list has loaded, shows an empty bar for a moment.
- **Presence, *Add column* and the board ⋯ wait**, because who you are on this board isn't known yet.
- **Column panels reach the right edge, as many as fit**, the last cut off as on 2.5. That was the user's note. At 1440 with the sidebar open that is four and the edge of a fifth.
- **Each panel holds `Skeleton block` instances**, 3, 2, 3, 1 and 2, never on the bare canvas (ADR-029 §4). Nothing shimmers.
- **The headers draw empty**, as bands with their rule, because a block has one height and the titles aren't known yet. No *Add card*.

The rejected options: everything waiting, including the name you just pressed; and three panels stretched to share the width, which jump to 280 when the board arrives.

### 10. 2.10, the empty board: a whiteboard

- **The stack** (§12): the `NoColumns` drawing, a blank whiteboard with its marker on the tray, and *This board has no columns yet.* in `color.muted-foreground`, like every empty state.
- **No button of its own.** The one action is the bar's *Add column* (ADR-033 §8). The bar is 2.5's.

The rejected drawings: a bare corkboard, which reads as *no cards*; and a clipboard, which reads as a form.

### 11. 2.11 and 2.12: load failed, and gone

**2.11, load failed.** The bar keeps the name and the rail, and drops presence, *Add column* and the ⋯. The stack: the `LoadFailed` drawing, a plug pulled from its socket, then *This board didn't load.* in `color.foreground`, then an Outline ***Try again***.

**2.12, gone, has two wordings**, both with the `BoardGone` drawing, a closed and taped box:

| When | The stack says | Button | The bar |
|---|---|---|---|
| The live archive message arrives | *This board was archived.* | Outline ***Go to All boards*** | Name and rail |
| Any *not found* answer | *This board can't be opened. It may have been archived, or its workspace deleted.* | Outline ***Go to your workspaces*** | Empty, and the rail in `color.skeleton` |

- **The board leaves the sidebar's list**, so no row is current.
- **The gap is recorded, not fixed.** `DeleteWorkspace` sends no live message, so after a workspace is deleted an open board finds out at its next request. The *not found* wording covers that, and an old link to an archived board. A live message on workspace deletion would reopen the backend a third time.

The rejected options: one wording for both, which drops the reason when the app knows it; and adding that live message now.

### 12. The centred stack

Every board with nothing to show says why in the same place, 2.8 and 2.10 to 2.12:
- **centred both ways on the canvas**
- the drawing at `size.illustration`, then **`spacing.illustration-gap` (16)**, then the sentence, centred, **wrapping within `size.form-width` (400)**
- then **`spacing.section-gap` (24)**, then the button. That is the gap Login puts above its button. It was the user's pick over 16.
- the sentence is `type.body`. It is `color.muted-foreground` on the empty board, like every empty state, and `color.foreground` on 2.8, 2.11 and 2.12, like 2.4's failure line.

**2.4's own load failure is unchanged** (ADR-033 §10): one row, top-left, where its grid would start.

### 13. 2.13, session expired: a page that waits, then back to where you were

- **A whole page, like Login**: a `size.form-width` block on `color.background`, its parts `spacing.section-gap` apart, **left-aligned like Login's**. That was the user's call over a centred block.
- **The `SessionEnded` drawing**, an hourglass that has run out, sits `spacing.illustration-gap` (16) above the heading, at the block's left edge.
- ***Your session has ended.*** in `type.title`, not the 28/36 heading, which belongs to Login and Register alone (ADR-017 §6). Then *Sign in again to pick up where you left off.* in `type.body` and `color.muted-foreground`. Then a full-width Primary ***Sign in***.
- **Nothing moves until you press *Sign in*.** After signing in you come back to the page you were on, not to your last workspace.
- **In code:** the axios 401 handler sets a *session ended* flag that the app layout renders as 2.13, instead of assigning `window.location.href`. *Sign in* goes to `/login?next=<path>`, and Login, after success, navigates to `next` **only if it is a same-site path** (it starts with `/` and not `//`), else to the last workspace (ADR-034 §6).

The rejected options: the same page with a five-second jump, which someone who looked away misses; and a dialog over the frozen board, which contradicts ADR-018 row 39 and looks as if it could still save.

### 14. The drawings on screens that are not empty

- **ADR-033 §8's list widens** to 2.8, 2.11, 2.12 and 2.13, beside the four empty flavours.
- **Rule 5 becomes:** one everyday object that names what is missing **or what happened**. The new subjects are a padlock (`PrivateBoard`), a plug pulled from its socket (`LoadFailed`), a closed and taped box (`BoardGone`, the partner of *Nothing archived*'s open, empty box) and an hourglass that has run out (`SessionEnded`). Rule 8's names follow: a PascalCase name for what is missing or what happened.
- **None of the four has a sparkle.** In this app a sparkle means *new*.
- **All four take the plum tint, the user's call.** The recommendation kept the tint only where the screen's button is plum (2.8, 2.13) and drew 2.11 and 2.12 in ink and paper, because ADR-018 §7 spends no colour on a failure. The conflict was shown once, side by side. **ADR-018 §7 is narrowed for the drawings' tint only**: the failure's message and button stay uncoloured.
- **Measured where they sit:** on the board canvas the tint is 1.12 in Light and 1.63 in Dark, the ink 6.08 and 7.83. On 2.13's page they are ADR-033 §8's numbers.

### 15. The wordings

| Board | Where | Wording |
|---|---|---|
| 2.6 | Field labels | *Column · Assignee · Due date* |
| 2.6 | No assignee · no date | *Unassigned · No due date* |
| 2.6 | Description, its label and when empty | *Description · Add a description…* |
| 2.6 | Timestamps | *Created Nov 18 · Updated Nov 22* |
| 2.6 | Delete, and the close button's spoken name | *Delete card · Close* |
| 2.7 | Banner (ADR-028, unchanged) | *You're viewing this board. Join it to add, edit and move cards.* · *Join board* |
| 2.8 | Gate | *This board is private. Join it to see its cards. Its members will see that you joined.* · *Join board* |
| 2.10 | Empty board | *This board has no columns yet.* |
| 2.11 | Load failed | *This board didn't load.* · *Try again* |
| 2.12 | After the archive message | *This board was archived.* · *Go to All boards* |
| 2.12 | After *not found* | *This board can't be opened. It may have been archived, or its workspace deleted.* · *Go to your workspaces* |
| 2.13 | Session expired | *Your session has ended.* · *Sign in again to pick up where you left off.* · *Sign in* |

### 16. What the file draws

**Tokens (206 → 210):** `size.drawer-width` 560 and `size.field-label` 96 in `base`; `shadow.rule-left` in `theme-light` and `theme-dark`.

**One new part, on `Components`:** `Field`, on `Field — states`, `State` = Rest · Hover · Pressed · FocusVisible. It is the Sidebar row's geometry on the drawer's surface:
- `size.control-height` tall, `spacing.control-padding-x` at the sides, `spacing.icon-gap` between the slot and the value, `radius.control`, filled `color.popover`
- a 16 `Slot` holding Phosphor `Kanban` (*Column*) or `CalendarBlank` (*Due date*), or an `Avatar` in its place (*Assignee*); the value in `type.body`; and a `Caret`, shown on *Column* only
- no edge at rest. Hover and pressed step the edge (`color.border-hover`, `color.border-pressed`, `border.width`, inside), and FocusVisible adds `shadow.focus-ring`

The file now has **19 parts**. **The `Illustration` part gains five subjects**, `NoColumns`, `PrivateBoard`, `LoadFailed`, `BoardGone` and `SessionEnded`, drawn as paths with every fill and line bound.

**Nine boards, on `Page 1`**, under 2.5 at x 6560, 1440 × 900 each, 1100 apart: 2.6 Card drawer, 2.7 Read-only board, 2.8 Join gate, 2.9 Cold load, 2.10 Empty board, 2.11 Load failed, 2.12 Gone — archived, 2.12 Gone — not found, 2.13 Session expired. The drawer is drawn with *Send the press kit to the three trade outlets* open: In Progress, JT, due tomorrow. 2.8 shows *Salary bands 2027* in terracotta.

### 17. The code

- **CSS variables:** `--drawer-width: 560px` and `--field-label: 96px`. No colour is added.
- **The drawer** is shadcn's `Sheet` with `side="right"`, `w-[var(--drawer-width)] max-w-full`, `bg-popover`, `border-l`, `shadow-*` from `shadow.overlay`, and its overlay in `--scrim`. Below 640 the same component is the full-height sheet.
- **A `Field`** is a ghost button restyled to the edge ladder, like `SidebarMenuButton` (ADR-034 §14). *Column* opens *Move card to…*. *Assignee* and *Due date* expand their picker in place.
- ***Delete card*** is a ghost button in `text-destructive-text` that opens the confirmation.
- **The date function** (ADR-037 §5) returns a short form for the card and a long form for the drawer.
- **The cold load** renders as many skeleton panels as the canvas width holds, plus one.
- **2.13:** the 401 interceptor and the `next` parameter in §13, recorded in `frontend-architecture.md`.

## Options considered

Every option is named, with its cost, under its decision above. In short:
- **Presentation:** a panel beside a live board.
- **Surface:** `color.background` and a shadow; `color.popover` with no rule.
- **Width:** 480; 640.
- **Inside:** 24 padding; labels above.
- **Move and Delete:** buttons along the top; both behind a ⋯.
- **Drawer date:** the card's words only; the plain date.
- **2.7's banner:** between the bar and the rail.
- **2.8:** the banner over a bare canvas, as ADR-021 §6 was written; a board with a padlock on it.
- **Cold load:** everything waits; three wider panels.
- **2.10:** a corkboard; a clipboard.
- **Message position:** top-left, as on 2.4.
- **2.11's drawing:** a paper plane come down.
- **2.12:** one wording; a live message on workspace deletion; a filing cabinet.
- **2.13:** a five-second jump; a dialog over the board; a key; a centred block.
- **The stack's button gap:** 16.
- **The tint:** only where the button is plum.

## Out of scope (deferred, recorded)

- **The pickers, *Move card to…* and the delete confirmation.** Stage 5's overlays.
- **State 32, the open card deleted by a teammate.** It needs a toast wording and a toast position, and both are still open in Stage 5.
- **The sheet-on-sheet below 640.** *Move card to…* is a full-height sheet on a phone, and it opens from the card drawer, which already is one. ADR-021 §5 forbids that. Stage 5 decides, likely an in-place list like the other pickers.
- **A live message on workspace deletion.** A backend change, recorded here and not scheduled (§11).
- **Hover, pressed and focus on these boards.** They are drawn at rest. The parts carry those looks.

## Downstream edits this ADR requires

1. **[ADR-018](adr-018-state-vocabulary.md): annotation.** Its Status line records that §7's zero-channel failures keep the drawings' plum tint.
2. **[ADR-021](adr-021-design-inventory-corrections.md): annotation.** Its Status line records that §6's *one component* is narrowed: 2.8 is a centred stack.
3. **[ADR-028](adr-028-banner.md): annotation.** Its Status line records that on 2.7 the banner sits under the rail.
4. **[ADR-033](adr-033-workspace-screens.md): annotation.** Its Status line records that §8's list widens and §9's rules 5 and 8 name what happened.
5. **[ADR-036](adr-036-one-sided-rules-as-inset-shadows.md): annotation.** Its Status line records `shadow.rule-left`.
6. **[ADR-037](adr-037-board-at-rest.md): annotation.** Its Status line records §2's reworded sentence.
7. **[ADR-017](adr-017-ui-foundations.md): annotation.** Its Status line records §9's two new variables.
8. **Penpot `Collabotask App`:** done 2026-11-23 (*Operational note*).
9. **[`design-file-inventory.md`](../design-file-inventory.md):** §1.6, §2's rows 2.6 to 2.13, §4.2, §5, §7.2, §7.3 and §7.6.
10. **[`frontend-architecture.md`](../frontend-architecture.md):** the 401 handler and the `next` parameter (§13).
11. **`CLAUDE.md` › `## Now`:** a pointer.
12. **`docs/build-log.md`:** no entry. Stage 4 is not closed.
13. **PRD, user stories, SRS, `CONTEXT.md`:** no change.

## Operational note

**Drawn 2026-11-23**, at Light + Coarse, as the file was found. The binding check returns only the 21 card `⋯` rows on 2.6 and on 2.7, which §7.5 expects; on 2.7 they are hidden but still there. It returns nothing on the other seven boards, on `Field — states` or on `Illustration — states`. Dark + Fine was looked at on 2.6, 2.8 and 2.9: the drawer's left rule is its edge against the dimmed board, the padlock's tint reads on the dark canvas, and the blocks show inside their panels. Light + Coarse was looked at on all nine. After each flip, every text box matched its text to within 1px, apart from texts set to fill their row. The file was left in Light + Coarse.

**The Foundations catalogue was regenerated:** three rows (`size.drawer-width`, `size.field-label`, `shadow.rule-left`), *base · 48*, *64 each*, and *210 tokens*. The check found one older defect there and it was fixed: five `Card` instances' meta rows kept a typed gap of 8 after ADR-037 §7, and are now bound to `spacing.inline-gap`.

What drawing found:
- **`penpot.createPath()` gives a new path a default stroke.** The check caught it on the Field's two glyphs. Set `strokes = []` on a glyph path.
- **`penpot.openPage()` is ignored while the tab is in the background**, and so is `addVariant()` on a part whose page is not the open one. Neither throws. Read `penpot.currentPage` before writing.
- **`addVariant()` places the copy by the container's own layout**, 20 apart, so the Field's states sit 452 apart, not 480. One call made one variant too many, which was removed.
- **`transformInVariant()` on a part made from an unlinked instance** put its container inside the nearest board, `Card — states`. It was moved to its own `Field — states`.
- **Copying a nested shape out of an instance does nothing**, with no error: the sidebar row's `Lock` did not arrive in 2.8's bar. Build the glyph from its path instead. Copying a whole top-level instance, such as a `Button`, works.
- **Login's submit button is drawn `NotSatisfied`**, the grey look of an empty form. 2.13's copy was switched to `Rest`.
- **In Light, the *not found* rail in `color.skeleton` is the canvas's own colour**, so it disappears. That is the intent, since no colour is known. In Dark it shows.
- **Penpot's hosted server refused every request for a few minutes**, with a Cloudflare block page, then accepted them again. Nothing was lost. Read the file's state before re-running anything.
- **The tab slept once** and the run resumed after it was brought to the front.
