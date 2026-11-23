# ADR-035: The Sidebar's Board List While It Loads and When It Fails, and the Selected Workspace's Dot

- **Status:** Accepted — **the *Operational note*'s two bullets on per-side strokes are corrected by [ADR-036](adr-036-one-sided-rules-as-inset-shadows.md)** (2026-11-22). Penpot draws no per-side width, so the sidebar's edge and its top row's rule drew nothing, and the user row drew as a full box. All three are now inset shadow tokens, and the *Closed, Rows* variant, which had no right edge even in its data, has one. Every decision here stands.
- **Date:** 2026-11-21
- **Scope:** Question 4 for [ADR-034](adr-034-side-menu.md)'s two parts (`Sidebar row` and `Sidebar`) and for the redraw of 2.4's six boards. It settles:
  - what the switcher and the board list show on a cold load (state 19) and on a load failure (state 22). This was the question ADR-034 left open.
  - how the sidebar's retry looks
  - where the open switcher's workspace names start, and what marks the selected workspace
  - the space under *Your boards*
  - where the user row's avatar sits
  - **one token, one entry**: `size.dot` 8 (199 → 200)

  It **extends [ADR-017](adr-017-ui-foundations.md) §9** (`--dot`) and ADR-034 §7, §8 and §10.9. It **supersedes two things in [ADR-018](adr-018-state-vocabulary.md), for the sidebar's board list only, and nothing else in them**:
  - the channel ledger's **Rose** row. Rose means overdue, destructive intent and invalid input. In the sidebar it also says *your boards didn't load*.
  - §7's **row 22**, *"zero channels"*. The sidebar's failure sentence spends rose. The page's own failure on 2.3 and 2.4 still spends nothing.
- **Upstream edits:** none. Everything here happens in the browser. No endpoint, schema or permission rule moves.

## Context

ADR-034 settled the sidebar and drew nothing. Its *Downstream edits* item 6 asked question 4 of the two parts and the 2.4 redraw before drawing, and inventory §1.7 named the one open question: what the board list shows on 2.4's cold load and load failure, since it shares the grid's request.

**First, the transparency check** ADR-034 §13 asked for. A Penpot colour token takes transparency, so `color.scrim` is a real token. The result, and the form Penpot refuses, are in ADR-034's *Operational note*. The five entries went in (194 → 199).

**Round 1 found no missing value.** Every number the two parts and the redrawn boards need binds a token the file already has. It found five decisions nobody had made. They went on a rendered page, each option drawn in Light and Dark at 1:1 with the file's values, with a picker. The user took three recommendations and chose the switcher block on cold load. They left notes on two choices:
- **The retry:** *a text button with a refresh icon and a red accent, to emphasise trying again.*
- **The open switcher:** *something in the empty slot that marks the selected workspace and catches the eye, like a pulsing dot.*

**Both notes ran into a rule ADR-018 had already settled.** Rose already has three meanings, and row 22 gives a load failure none. Movement means *a teammate did this*, which is why shadcn's Skeleton lost its pulse (ADR-029 §6). A second page drew each note's aim beside options that keep the rules, each with its cost, and the user's own idea was one of the options. The pulse itself was not offered. **The user chose against the recommendation both times**: a rose sentence over an uncoloured retry, and a still plum dot.

**The dot needed a size the file did not have, so question 4 fired**, in round 2. That makes fourteen stages out of fourteen.

## Decision

### 1. While the list loads (state 19): a block for every row that waits for data

- **The switcher row** is a Skeleton block at a row's geometry: `size.control-height` tall, `radius.control` corners, the list's full width, `color.skeleton`.
- **Under *Your boards*, three blocks** of the same size.
- **These draw at once:** the top row, *All boards* (current on 2.4), the *Your boards* label, and the user row. None of them waits for either request.
- **The blocks measure 1.26 in Light and 1.30 in Dark** against the sidebar, the same as 2.4's tile blocks against the page ([ADR-033](adr-033-workspace-screens.md) §1).
- **The same applies on the board's cold load (2.9)** whenever the list is not cached yet.

**The switcher's block was the user's call.** The recommendation left the switcher blank until the name arrived, as 2.4's heading does ([ADR-033](adr-033-workspace-screens.md) §10). The cost, shown on the page: on 2.4's cold load the sidebar shows a block for the workspace's name while the heading beside it shows nothing for the same name.

### 2. When the list's request fails (state 22): a rose sentence, then a refresh row

- ***Your boards didn’t load.***, in `type.body` and `color.destructive-text`, inset by `spacing.control-padding-x`. It measures 6.29 in Light and 7.34 in Dark against the sidebar.
- **Under it, `spacing.list-gap` below, a `Sidebar row`:** Phosphor `ArrowClockwise` in the slot in `color.muted-foreground`, and *Try again* in `color.foreground`. It has no edge at rest. Hover and pressed step the edge, and focus adds the ring, as on every row.
- **In the sidebar every control is a row**, so the retry needs no third Button kind. The pages keep their Outline *Try again* (ADR-033 §6).
- **It works the same on 2.4 and on a board.** On 2.4 one request feeds the grid and the list, so the screen has two *Try again*s and either one retries both. On a board the list has its own request, which can fail while the board loads fine. There the row is the only way to retry the list.

**The rose was the user's call.** The recommendation was the same row with no colour. The cost, shown on the page: rose otherwise means overdue, destructive intent and invalid input, and here nothing the user did was wrong. On 2.4, the page's own failure sentence beside it stays uncoloured, so the sidebar is louder than the page. The user's note asked for the red on the button. The option chosen puts it on the sentence instead, so the control does not look destructive.

### 3. The open switcher: indented rows, and a still plum dot on the selected workspace

- **Workspace rows and *All workspaces* keep the empty 16px slot**, so their names start 44 from the edge, with *All boards* and every board name. The switcher's own name stays at 20. The indent says the rows belong to the switcher.
- **The selected workspace's slot holds a dot**: `size.dot` 8, round, in `color.primary`. That is plum's existing *where you are*, so the ledger's Plum row is unchanged. The row also carries the current row's plum text and edge (ADR-034 §10.3). The other workspace slots stay empty, and *All workspaces* has no glyph.
- **The dot stays still.** The note asked for a pulse, and ADR-018 §8 gives movement one meaning, *a teammate did this*. A dot that pulses for as long as the list is open would say something is happening when nothing is.
- **The switcher's own row does not change** when the list opens. `CaretUpDown` shows either way, and `aria-expanded` says which.
- **While the list is open on 2.4, two rows are plum**: the selected workspace and *All boards*.

**The dot was the user's call.** The recommendation was a Phosphor `Check`, the usual mark of the chosen option in a list you pick from. The costs, shown on the page:
- the dot sits in the slot where a board shows its colour, so it can read as one more board colour. In Light, mauve is close to plum.
- a dot more often means *unread* or *online*.
- it needs a new token, where the check was already in the file

**`size.dot` 8**, in `base`, is half the slot (`size.icon` ÷ 2). The value is chosen.

### 4. The space under *Your boards*: `spacing.list-gap` 4

The label is treated as the list's first item, so everything under it sits 4 apart, and the row's own padding adds the rest. The label sits 12 above the first board name at Fine and 16 at Coarse, and 32 / 36 below *All boards*, so it reads with the boards. `spacing.field-inner-gap` was rejected because it would borrow a form field's token for a list ([ADR-022](adr-022-form-control-geometry.md) §4).

### 5. The user row: the avatar takes the slot's place

The Avatar (20 at Fine, 24 at Coarse) starts where a row's slot starts. Your name starts at 48 / 52, so it is 4 / 8 right of the board names above it. The row sits under its own rule, apart from the list. Lining the name up would have needed a new pair of density tokens for a 4–8px shift.

### 6. What follows by rule

- **The top row:** the show/hide button and *CollaboTask* sit `spacing.inline-gap` 8 apart, by ADR-027's rule for neighbours that are not an icon and its label. With the sidebar hidden, the same gap separates the button from *All boards* in 2.4's bar.
- ***Your boards*, the failure sentence and the seventh empty flavour** are inset by `spacing.control-padding-x`, so they start where a row's content starts, 20 from the edge.
- **A row-sized block takes a row's geometry.** That is ADR-029's rule that a block stands where the thing will.
- **On 2.4 the sidebar and the page are the same colour** in both themes (1.00). The rule (1.26 in Light, 1.92 in Dark) is the only edge, as where the app bar meets the page. ADR-034 §10.2 measured the sidebar against the board canvas only.
- **The focus ring** (2px, at a 2px offset) fits inside the 4px gap between rows and the 8px side padding.

### 7. What the file draws

**Two parts, on `Components`:**
- **`Sidebar row`**, on `Sidebar row — states`: `State` = Rest · Hover · Pressed · FocusVisible · Current. It draws every row in the sidebar.
  - It is built from `size.control-height`, `spacing.control-padding-x`, `radius.control`, `spacing.icon-gap`, `type.body`, and a `color.background` fill, so the focus ring has a surface (ADR-024).
  - Its 16px slot (`size.icon`) holds the board's square (`size.swatch`, `radius.badge`, the hue's `.strong`), and hides the `SquaresFour` glyph, the `ArrowClockwise` glyph and the dot until an instance shows one.
  - A hidden Avatar, a `Name` group (the title, then a hidden filled lock, `spacing.icon-gap` apart) that fills the row, and a hidden `CaretUpDown` follow.
  - Current turns the edge, the title and every glyph `color.primary`. The square keeps its hue.
- **`Sidebar`**, on `Sidebar — states`: two axes, `Switcher` = Closed · Open and `List` = Rows · Loading · Failed · Empty. That gives five variants, because Open is drawn with Rows only.
  - It is 240 × 900 (`size.sidebar-width`) on `color.background`, with a right-only rule.
  - The top row is `size.app-bar` tall, with `spacing.screen-x` at the sides and a bottom rule. Then comes the list, `spacing.sidebar-padding` in from the sides and below the top row, its two groups `spacing.section-gap` apart and their rows `spacing.list-gap` apart. The user row sits at the bottom under a top rule, with `spacing.sidebar-padding` around it.
  - The same board shows 2.4's app bar with the sidebar hidden, which is drawn nowhere else.

**2.4's six boards, redrawn on `Page 1`:**
- Each is a row: a `Sidebar` instance, then `Main`, which holds the app bar and the page.
- The bar reads *All boards*.
- The grid has 4 columns. At 1440 the tiles are 286 wide at Fine and 283 at Coarse.
- The sidebar variant per board:

| Board | Sidebar |
|---|---|
| *2.4 Workspace detail* | Closed · Rows |
| *— archived* | Closed · Rows: the Archived filter is its own request, so the list still shows the joined active boards |
| *— cold load* | Closed · Loading. The grid's five blocks became four, one full row |
| *— empty* | Closed · Empty: no boards at all means none joined, so flavour 7 shows |
| *— archived, empty* | Closed · Rows |
| *— load failed* | Closed · Failed |

The sample list is the four boards 2.4's grid shows as joined, in the grid's order, newest first: *Q3 Launch Plan*, *Hiring* (private), *Brand guidelines* and *Support rota*.

### 8. The wordings

| Where | Wording |
|---|---|
| Sidebar, the list failed | *Your boards didn’t load.* |
| Sidebar, retry row | *Try again* |

Every other wording is ADR-034 §12's.

### 9. The code

- **§9 gains `--dot: 8px`.**
- **The retry is a `SidebarMenuButton`** with Phosphor `ArrowClockwise`, not a `Button`, restyled to ADR-034 §10.3 like every row.
- **The failure sentence** uses `--destructive-text`.
- **The cold-load blocks** are shadcn's Skeleton (no `animate-pulse`, ADR-029 §6) at `h-[var(--control-height)]` with the control's radius.
- **The dot** is `aria-hidden`, and the selected workspace's link carries `aria-current`. It is `"page"` on 2.4, where that link leads to the page you are on, as *All boards* does, and `"true"` on a board.

## Options considered

### While the list loads
- **Blocks for the boards, and a blank switcher**, recommended. Each part waits the way the page beside it waits for the same data.
- **Nothing until it loads.** The quietest. The empty space under *Your boards* can read as *no boards*.

### When the list fails
- **A sentence and an Outline button**, round 1's drawing. Heavier than a row in a list of rows, which the user's note pointed out.
- **A sentence only.** On a board whose list failed, nothing is left to press.
- **Hiding the list.** A failed list looks like one you can't see.
- **The refresh row with no colour**, recommended in round 2.
- **The refresh row in rose**, the note as written. A rose control reads as destructive.

### The open switcher
- **Flush with the switcher's name.** Two name edges run down the sidebar, 24px apart.
- **A check on the selected workspace**, recommended in round 2.
- **A letter mark on every workspace.** It fills every slot, but needs a new part, and two workspaces with the same first letter look alike.
- **A pulsing dot**, from the note. Not offered: it breaks ADR-018 §8.

### The space under the label, and the user row
- **`spacing.field-inner-gap` 8.** It borrows a form token.
- **The name lined up with the boards.** It needs a new density token pair for 4–8px.

## Out of scope (deferred, recorded)

- **The slide-over panel's own drawing**, below 1024. It is Stage 5, as ADR-034 records.
- **Announcing the failure to a screen reader** without the user's action. SC 4.1.3 stays open (PRD §5.4).
- **2.3's and 2.4's page failures.** They keep ADR-033 §10's uncoloured sentence and Outline *Try again*.

## Downstream edits this ADR requires

1. **[ADR-017](adr-017-ui-foundations.md): annotation.** Its Status line records §9's thirteenth extension, and a note under §9 adds `--dot`.
2. **[ADR-018](adr-018-state-vocabulary.md): annotation.** Its Status line records the narrowing, and notes go under §7's table and the channel ledger.
3. **[ADR-034](adr-034-side-menu.md): annotation.** Its Status line records that its open question and question 4 for its parts are settled here.
4. **Penpot `Collabotask App`:** done 2026-11-21. `size.dot` and ADR-034's five entries are in (200 tokens), the two parts are drawn, 2.4's six boards are redrawn, the Foundations catalogue gained six rows, and the binding check and both flips ran (*Operational note*).
5. **[`design-file-inventory.md`](../design-file-inventory.md):** §1.7, §2's 2.4 row, §4.2, §5, §6.5, §7.2, §7.3 and §7.6.
6. **`CLAUDE.md` › `## Now`:** a pointer.
7. **`docs/build-log.md`:** no entry. Stage 4 is not closed.
8. **PRD, user stories, SRS, `CONTEXT.md`:** no change.

## Operational note

**Drawn 2026-11-21.** The binding check returns nothing on `Sidebar row — states`, `Sidebar — states` or any of 2.4's six boards. On the Foundations board it returns only its expected 53 rows: 48 `Hue /` pairings and the Card's five `⋯` glyphs. Dark + Fine was looked at and its numbers read back (rows 36, the button 24, tiles 286, both surfaces `#0C0A09`). Light + Coarse was looked at and measured the same way, and the file was left in it, as it was found. In Dark, the four board squares are the pale `.strong` tints and hard to tell apart. That is the measured palette ADR-034 §10.4 chose, and the tiles' rails show the same.

What drawing found:
- **`penpot.createPath()` adds a 1px black stroke by default.** Every new glyph had one, so the plum `SquaresFour` exported black. The binding check caught it as an unbound `strokeColor` and `strokeWidth`. Clear `strokes` on the mains, since instances inherit.
- **On a board with a layout, `flex.horizontalSizing` and `layoutChild.horizontalSizing` are one setting.** Writing `fix` to the first after `fill` to the second left the `Name` group 100 wide. Set the child sizing alone.
- **Per-side stroke widths go on the stroke object**, not on the shape. Assigning `shape.strokeWidthTop` throws *object is not extensible*, and the writes before it in the call still land. ADR-023 recorded this. It was not re-read before the first call.
- **A top-only rule whose general width reads 1 exports as a full box.** The user row's rule has widths 1 / 0 / 0 / 0 and draws correctly in the generated CSS. Check per-side rules on the canvas, as ADR-027 says.
- **A grid's `children` are not in visual order.** Reading them as if they were reversed the sidebar's list before it was caught. Read `layoutCell.row` and `.column`.
- **Removing a grid column:** move its children into other cells first, then call `removeColumn`.
- **`penpotUtils.findShapeById` called through a bare alias loses `this`** and throws before doing anything. Wrap it in a function.
- **Cloudflare's firewall blocked one read-only call** that used template strings and `JSON.stringify`. The same read in plain concatenation passed.
- **Calls that timed out while the tab was in the background still ran.** Check the file's state before re-running anything.
