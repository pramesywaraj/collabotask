# ADR-037: The Board at Rest — the Header Gap, the Rail, the App Bar, Add Card, Late Dates and the Empty Column

- **Status:** Accepted — **§2's *"the columns' tops touch it"* reworded by [ADR-038](adr-038-board-states-group-d.md)** (2026-11-23): the columns' tops touch whatever is above them, which on 2.7 is the read-only banner. Everything else stands as written.
- **Date:** 2026-11-23
- **Scope:** Question 4 for board 2.5, *Board — at rest*, the reference frame for the rest of Stage 4. It settles:
  - the space between a column's header and its first card
  - where the board's 4px colour rail sits
  - how the board's app bar is arranged, which [ADR-031](adr-031-presence-cluster.md) left to 2.5
  - where *Add card* sits and what it looks like
  - **how tall a column is: as tall as its content, up to the bottom of the screen**
  - how a due-soon or overdue date is told apart without colour, and what *due soon* means
  - an empty column's wording (empty flavour 5)

  **No token is added** (206 stays 206). One part is added, `Add card`, and the `Card` part's meta row is corrected. It **extends** [ADR-017](adr-017-ui-foundations.md) §3.2 (late dates carry words as well as colour), and **settles** what [ADR-023](adr-023-spacing-between-objects.md) §2.4 and ADR-031's *Out of scope* left open. It **narrows ADR-023 §2.3's wording**: a column runs from under the rail as far down as its content, and at most to the bottom of the screen. §2.3's decision, no padding above or below the columns, stands. It **revises one measured number in [ADR-019](adr-019-responsive-touch-scope.md) §7**: a phone shows 6 whole cards in a column, but 5 and a peek once the column is longer than the screen (§4).
- **Upstream edits:** none. The PRD and user stories say nothing about how a due date is shown, and nothing here touches an endpoint, the schema or a permission rule.

## Context

2.5 is the first board drawn with columns, cards, the board's app bar and its rail together. Inventory §7.3 calls it *"the reference frame everything else varies from"*: 2.6 to 2.14 are drawn as differences from it.

Question 4 was asked of every value 2.5 needs, against the live catalogue of 206 tokens, before anything was drawn. **Almost everything mapped**:
- the sidebar is the `Sidebar` part, with the board's row marked current instead of *All boards* (ADR-034 §8)
- the app bar's height, fill, rule and side inset are ADR-033 §2's
- presence is the `Presence cluster` part, and the column header the `Column header` part
- the canvas, the column panels, the cards, their gap and the date colours are ADR-017 and ADR-023's
- the rail's size and colour are `size.board-rail` and the hue's `.strong`

**One value had no token and no decision, so question 4 fired**, for the fifteenth stage out of fifteen: the space between a column's header and its first card. Both card-count sums assume 0 (ADR-017 §1, ADR-019 §7). At 0 the first card's top edge sits on the header's rule, and the two lines read as one thick line.

**Five decisions had never been made:**
- the rail's position, which ADR-023 §2.4 costed and left open
- the app bar's arrangement, which ADR-031 handed to 2.5
- *Add card*'s place and look. The inventory puts it at the foot of each column (§1.5), and says nothing more.
- **late dates are told apart by colour alone.** ADR-017 §3.2 colours a due-soon date amber and an overdue one rose, and changes nothing else. WCAG 1.4.1 (PRD §5.4's AA target) asks that colour not be the only sign. In Light, amber measures 1.52:1 and rose 1.21:1 against the grey dates. In Dark they are 1.12:1 and 1.81:1. *Due soon* itself had no definition anywhere.
- empty flavour 5. ADR-033 §8 says an empty column keeps one line and no drawing, but no wording existed and no board drew it.

All six went on a rendered page. Each option was drawn in Light and Dark with the file's values, at Coarse where density changed the answer, and late dates also through a deuteranopia simulation. The user picked with a database-backed picker. **They took every recommendation but one**: the rail sits in its own strip under the bar, not along the bar's top edge.

**Their note on *Add card* settled one thing more.** It read *"The column length will be stretched as the length of the display right?"* It was first answered as a question, *yes, every column is full height* (ADR-023 §2.3), and 2.5 was drawn that way. The user then said what they meant: **a column's height follows its content, and the screen's height is its maximum.** 2.5 was redrawn the same day (§4).

## Decision

### 1. The first card sits a card gap below the header

**`spacing.card-gap`**, 8 at Fine and 12 at Coarse, bound as the column body's top padding. The first card sits as far below the header as the next card sits below it.

It is bound by rule, like ADR-027's header inset, not by a matching number. `spacing.column-padding` was rejected because it stays 8 at Coarse, which would put the first card closer to the header than to its neighbour.

### 2. The rail is its own 4px strip, under the app bar

- **`size.board-rail` (4) tall, `color.identity.<hue>.strong`**, between the app bar and the canvas. The columns' tops touch it.
- **It spans the board area only**, to the right of the sidebar, because the sidebar runs the full height (ADR-034 §10.1). With the sidebar hidden, it spans the whole width.
- **It is drawn under the bar's rule**, so the bar keeps its own line and the sidebar's top-row rule still meets it at 56.
- **It costs 4px of board height**, as ADR-023 §2.4 worked out. §4 gives the totals.
- In Dark, `.strong` is the pale tint, so the rail is pale (ADR-017 §3.3).

**The user's call.** The recommendation put the rail along the top of the bar, inside its 56px, which costs no height. A third option replaced the bar's grey rule with the board's colour.

### 3. The app bar: the name on the left, then presence, *Add column* and ⋯ on the right

- **The board's name** in `type.title` takes all the room that is left, and is cut with an ellipsis when it runs out.
- **Then, on the right:** the `Presence cluster` (Live: five faces and a seat, you first), ***Add column*** as a Primary `Button`, and the **board ⋯**, in that order.
- **All four sit `spacing.inline-gap` (8) apart**, by ADR-027's rule for neighbours that are not an icon and its label.
- **⋯ is last**, where a menu usually sits. It is Phosphor `DotsThree` at `size.icon` in a `size.target-min` box (ADR-027's rule), **in `color.foreground`**, like the sidebar's show/hide button on the same line (ADR-034 §10.7). The column ⋯ stays `color.muted-foreground`.
- ***Add column* is Primary** because ADR-017 §3 gives the resting board *"one plum button"*. It is that button.
- **Below 640** the bar has no room for *Add column* (ADR-019 §8). Where it goes there is 2.15's question.

The rejected options: presence straight after the name, which leaves the faces mid-bar beside a short name; and *Add column* as a slot after the last column, which leaves the resting board with no plum button.

### 4. *Add card*: a quiet row after the last card, which sticks to the bottom when the column is long

**Where it sits.**
- **Right after the last card**, `spacing.card-gap` below it.
- **A column is as tall as its content, up to the bottom of the screen.** That was the user's call, from their note. The panel starts under the rail and ends `spacing.column-padding` below the *Add card* row, so a short column ends just under its last card and the canvas shows beneath it. A column with more than the screen holds stops at the bottom of the screen.
- **When the column reaches the bottom of the screen, the row sticks there** and the cards scroll under it. While it sticks, a 1px rule along its top (`shadow.rule-top` in the file, a top border in code) separates it from the cards, as the header's rule does at the top.
- **The rule is on a wrapper, not on the row.** The row's focus look is `shadow.focus-ring`, and a shape binds one shadow token, so the rule goes on a `Footer` around it. That is the route [ADR-036](adr-036-one-sided-rules-as-inset-shadows.md)'s *Out of scope* named for this case. The footer is filled `color.board-column`, with `spacing.card-gap` above the row.

**What it looks like.** The `Add card` part, `State` = Rest · Hover · Pressed · FocusVisible:
- Phosphor `Plus` at `size.icon`, then *Add card* in `type.body`, both in `color.muted-foreground`
- `size.control-height` tall, `spacing.control-padding-x` at the sides, `spacing.icon-gap` between glyph and words, `radius.control`
- filled `color.board-column`, the surface it sits on, so the focus ring has something to cast from ([ADR-024](adr-024-control-and-field-internals.md) *Operational note*)
- no edge at rest. Hover and pressed step the edge (`color.border-hover`, `color.border-pressed`, `border.width`, inside the shape), and FocusVisible adds `shadow.focus-ring`, as on the sidebar's rows
- the glyph starts where the card titles start (8 + 12 = 20 from the panel's edge). The words start 24 further right. The decision page said the words lined up, and that was wrong.

An Outline button was rejected: it puts a box with an edge in every column.

**What sticking costs.** Together with §1 and §2, the card counts move:

| | Before | A column that fits | A column longer than the screen |
|---|---|---|---|
| Desktop, 1440 × 900, Fine (ADR-017 §1) | 8 cards, 92 spare | 8 cards, 80 spare | **8 cards, and 28px of the ninth** |
| Phone, 390 × 844, Coarse (ADR-019 §7) | 6 cards, 36px of the seventh | 6 cards, 20px of the seventh | **5 cards, and 56px of the sixth** |

The desktop sum: 900 − 88 chrome − 56 bar − 4 rail − 16 panel padding − 36 header = 700, then − 8 for §1, and − 44 (a 36 row and its 8 gap) when the row sticks. The phone sum: 844 − 140 − 56 − 4 − 16 − 44 = 584, then − 12, and − 56 (44 + 12) when the row sticks.

**ADR-019 §7's *"6 cards visible"* becomes 5 and a peek in a long column**, and stays 6 in one that fits. The user saw this cost on the decision page and chose it. The row is always in reach, and on a phone it sits under the thumb. ADR-019 §7 already marked that number as the one resting on an assumption, and 2.15 checks it by eye.

The rejected options: the row scrolling with the cards, which keeps every count but puts *Add card* out of reach in a long column; and the row always at the bottom of a full-height column, which separates it from the cards in a short one. With columns that follow their content, that option and the chosen one draw the same.

**The card counts are unchanged by the column's height.** They describe a column that reaches the bottom of the screen, and that column is as tall as before.

### 5. Late dates say so in words: *Due today*, *Due tomorrow*, *Overdue · Nov 20*

| The due date is | The meta row shows | Colour |
|---|---|---|
| before today | *Overdue · Nov 20* | `color.overdue` |
| today | *Due today* | `color.due-soon` |
| tomorrow | *Due tomorrow* | `color.due-soon` |
| later | *Nov 28* | `color.muted-foreground` |

- **"Due soon" means today or tomorrow**, counted in the viewer's own calendar days. A narrow window keeps amber rare, as ADR-017 §3 wants colour to be an event and not a texture.
- **The words carry the state, and the colour repeats it.** A board read in greyscale, or by someone who can't tell amber and rose from grey, says the same thing. This answers WCAG 1.4.1.
- **The longest wording, *Overdue · Nov 20*, is about 100px** of the meta row's 240 at Fine. The row now gives the date the width the avatar leaves (§7), so it never runs under the avatar.
- **Nothing else changes.** Amber still means deadlines only and rose overdue, destructive intent and invalid input ([ADR-018](adr-018-state-vocabulary.md)'s ledger). No icon is spent: ADR-018 §9 keeps the filled lock as the only icon that marks a state.
- **The app cannot know which column means finished**, so a card in *Done* with a past date still reads *Overdue*. The sample board puts no dated card in *Done*. Whether finished work should stop being overdue is a product question for later (*Out of scope*).

The rejected options: colour only, as ADR-017 §3.2 wrote; and a Phosphor `Clock` or `WarningCircle` before the date, which would have needed ADR-018 §9 narrowed. The windows rejected were 3 days and 7 days.

### 6. An empty column says *No cards yet.*

- ***No cards yet.*** in `type.body` and `color.muted-foreground`, at the top of the column body, `spacing.card-gap` below the header.
- **Inset by `spacing.card-padding-x`**, so it starts where a card title starts.
- ***Add card* follows**, `spacing.card-gap` below.
- **No illustration**, as ADR-033 §8 says for flavours 5 and 6.

The rejected options: the *Add card* row alone, which drops the line ADR-033 §8 expects; and drawing no empty column on 2.5.

### 7. The Card part's meta row: a lone avatar stays on the right

The `Card` part spread its two meta items to the row's ends. A card with an assignee and no date would therefore have drawn its avatar on the **left**.

All five variants now:
- **pack the meta row to its end**, so a lone avatar stays on the right
- **let the date fill the width the avatar leaves**, so a date alone still starts on the left
- **put `spacing.inline-gap` (8) between date and avatar**, by ADR-027's rule

Heights are unchanged (80 at Fine, 92 at Coarse), and every Card instance in the file follows. In code this is `ml-auto` on the avatar.

### 8. What the file draws

**One new part, on `Components`:** `Add card`, on `Add card — states`, with the four states in §4. It sits on a board filled `color.board-column`, its home surface. The file now has **18 parts**.

**Board 2.5, on `Page 1`** at x 6560, 1440 × 900, built like 2.4:
- **a `Sidebar` instance** (Closed · Rows). *Q3 Launch Plan* is the current row and *All boards* is at rest.
- **`Main`:** the app bar (§3), the `Rail` (§2, periwinkle), and the `Canvas` (`color.board-canvas`, `spacing.screen-x` on the left, `spacing.column-gap`, nothing above or below).
- **five columns**, each a panel (`color.board-column`, `radius.surface`, `spacing.column-padding`, `size.column-width`) holding a `Column header` instance, a `Body` (`spacing.card-gap` above and between) and the `Add card` row. The `Canvas` aligns them to the top. **The four short ones follow their content** (Fine heights: In Progress 340, Review 124, Done 360, Later 208). **To Do fills to the bottom**, 840, because it holds more than the screen:

| Column | Holds |
|---|---|
| To Do · 12 | Longer than the screen, so its *Add card* sticks in a `Footer` under the cards. One- and two-line titles, with and without a meta row, one with an avatar and no date |
| In Progress · 3 | *Write the launch announcement* (*Nov 28*), *Send the press kit to the three trade outlets* (*Due tomorrow*), *Final pricing sign-off* (*Overdue · Nov 20*) |
| Review · 0 | *No cards yet.* (§6) |
| Done · 4 | No dates (§5) |
| Later · 2 | The fifth column. 32px of it shows at Fine and 16 at Coarse |

- **The presence cluster starts with you**: AP, Ayu Pratama, in periwinkle, as in the sidebar's user row. Then AR sage, MK steel, JT ochre and SL mauve, and *+3*. MK moves from periwinkle to steel so no hue repeats.
- **Today, on this board, is 2026-11-23**, so *Due tomorrow* is Nov 24.

### 9. The code

- **The canvas** is `flex items-start`, so columns don't stretch. **Each column** is `flex flex-col max-h-full`. Its body takes `min-h-0 overflow-y-auto`, so a short column ends after its content and a long one stops at the bottom and scrolls.
- **The column body:** `pt-[var(--card-gap)]` and `gap-[var(--card-gap)]`.
- ***Add card*** is a ghost button restyled to the edge ladder, like `SidebarMenuButton` (ADR-034 §14). It sits after the last card inside the column's scroll area with `position: sticky; bottom: 0` and the column's background. Its top border shows only while it sticks: watch a 1px sentinel after the last card with an `IntersectionObserver`, or use a `scroll-state(stuck: bottom)` container query where the browser supports it.
- **The rail** is a `h-1` strip under the app bar, coloured from ADR-017 §9's identity constant (the `.strong` role), with the same fallback on read as the tile.
- **The app bar:** `flex items-center gap-2`, the title `min-w-0 flex-1 truncate`, the ⋯ `text-foreground`.
- **Late dates:** one function returns the wording and the colour from the due date and today's local date. The card's accessible name uses the same words.
- **No CSS variable is added.**

## Options considered

### The header gap
- **0**, what both card-count sums assumed. The rule looks twice as thick.
- **`spacing.column-padding`, 8 at both densities.** At Coarse the first card sits nearer the header than its neighbour.

### The rail
- **Along the top of the bar, inside its 56px**, recommended. No height cost.
- **In place of the bar's bottom rule**, 4px in the board's colour. No height cost. Where it meets the sidebar, a 1px grey line runs into a 4px coloured one.

### The app bar
- **Presence straight after the name.** With a short name the faces float mid-bar.
- ***Add column* as a slot after the last column.** It works at every width, but the board loses its one plum button and you scroll to find it.

### *Add card*
- **Scrolling with the cards.** Keeps every card count. In a long column it is out of reach.
- **Always at the bottom of the column.** In a short column it sits far from the cards.
- **An Outline button.** A box with an edge in every column.

### Late dates
- **Colour only.** Fails WCAG 1.4.1.
- **A glyph before the date.** Spends the icon channel on a second state, against ADR-018 §9.
- **A 3- or 7-day window.** More of a board turns amber.

### The empty column
- **Only the *Add card* row.** ADR-033 §8 expects a line.
- **Not drawn on 2.5.**

## Out of scope (deferred, recorded)

- **Where *Add column* goes below 640.** The phone's bar has no room for it. 2.15 decides, likely the board ⋯ menu.
- **A short column's drop area**, for 2.14. A column that follows its content is a shorter drop target, and on a phone the next column's 44px peek (ADR-019 §3) is only as tall as that column's cards. 2.14 decides whether the drop area and its `color.drag-over` tint stop at the panel or reach the bottom of the screen. Reaching the bottom keeps short columns easy to hit.
- **Whether a card in a finished column stops being overdue.** Phase 1 has no notion of a finished column. Deciding it would need one, which is a product change.
- **The date format itself**: other years, and the locale. *Nov 20* is the format the `Card` part has used since Stage 0.
- **Hover, pressed and focus on 2.5.** 2.5 is at rest. The parts carry those looks.

## Downstream edits this ADR requires

1. **[ADR-017](adr-017-ui-foundations.md): annotation.** Its Status line records that §3.2's late dates carry words, and a note goes under §3.2.
2. **[ADR-019](adr-019-responsive-touch-scope.md): annotation.** Its Status line records §7's revised count in a long column.
3. **[ADR-023](adr-023-spacing-between-objects.md): annotation.** Its Status line records that §2.4's rail sits under the bar, and that §2.3's columns run only as far down as their content.
4. **[ADR-031](adr-031-presence-cluster.md): annotation.** Its Status line records that the app bar it left to 2.5 is settled.
5. **[ADR-033](adr-033-workspace-screens.md): annotation.** Its Status line records flavour 5's wording.
6. **[ADR-036](adr-036-one-sided-rules-as-inset-shadows.md): annotation.** Its Status line records the first rule on a wrapper, the column footer's.
7. **Penpot `Collabotask App`:** done 2026-11-23 (*Operational note*).
8. **[`design-file-inventory.md`](../design-file-inventory.md):** §1.5, §2's 2.5 row, §4.2, §5, §6.5, §7.2, §7.3 and §7.6.
9. **`CLAUDE.md` › `## Now`:** a pointer.
10. **`docs/build-log.md`:** no entry. Stage 4 is not closed.
11. **PRD, user stories, SRS, `CONTEXT.md`:** no change.

## Operational note

**Drawn 2026-11-23**, at Light + Fine. The binding check on 2.5 returns only its 21 Card `⋯` glyphs, one per card, which §7.5 expects. It returns nothing on `Add card — states`, and only the five expected `⋯` rows on `Card — states`. Dark + Fine was looked at: the rail and swatches turn pale, the dates take their Dark values, and the sidebar's rule is the only edge between sidebar and canvas. Light + Coarse was looked at and measured first (headers, buttons, the ⋯ box and *Add card* 44, gap 12, a two-line card 92). The file was left in Light + Coarse, as it was found.

**Redrawn the same day** with columns that follow their content (*Context*). Dark + Fine and Light + Coarse were looked at again, and the binding check still returns only the 21 card `⋯` rows.

What drawing found:
- **`createShapeFromSvg` keeps a glyph at its 256 scale** when the path is lifted out of the group it imports. Phosphor's `Plus` came out 192 wide in a 16 box. Resize it to 12 × 12 at offset 2, which is the glyph at 16px. `DotsThree` is 10 × 1.5 at offset 3, 7.25, read from the `Column header` part.
- **`switchVariant` on a nested Avatar drops the face's margin token.** The presence faces kept their −6 overlap as a typed value after their hue was switched, and a re-bound token was lost again at the next density flip. The fix that held: **a fresh `Presence cluster` instance, recoloured by overriding each face's `Circle` and initials fill tokens and letters**, with no variant switch. Its overlap stays the part's own binding.
- **A density flip does not recompute a margin token's value.** After Fine → Coarse → back, the faces were 24 wide (Coarse) with Fine's −6 overlap, still bound. The `Presence cluster` part's own main does the same, so this predates ADR-037. The binding is right, and the check passes. The drawn value updates when the token is applied again.
- **A density flip can put copied texts back at the part's stored size.** After Fine → Coarse, six one-line card titles took the Card main's two-line height, *Add column* took the 47px width of the Button's *Sign in*, and some count digits drew blank. Switching each to fixed size and back to auto in a later call made Penpot measure it again. Compare a text's box with its `textBounds` after every flip, and re-measure before exporting.
- **`switchVariant` on a nested `Sidebar row` keeps its overrides.** *All boards* kept its glyph, and *Q3 Launch Plan* its title and square.
- **The presence cluster's covered initials are the part's design**, not a stale export. Each face covers 6px of the one behind it (ADR-031), and the part's own main exports the same way.
- **A one-line title can wrap at Coarse.** The title's right reserve widens from 24 to 32, so *Update the pricing page copy* takes two lines there. That is ADR-019 §7's geometry, not a defect.
- **At 1440 the fifth column shows 32px at Fine.** ADR-034 §11 reads 16, because it subtracted a 16px inset on the right that the canvas does not have while it scrolls.
- **One call timed out with the tab in the background.** Its writes had landed. Read the file's state before re-running anything (ADR-035 recorded the same).
