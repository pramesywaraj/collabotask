# ADR-027: The Column Header

- **Status:** Accepted
- **Date:** 2026-11-20
- **Scope:** The inside of the column header: how far its contents sit from its edges, and the gap between its title, count and `⋯`. That is one new token and one reuse bound by a rule. Also the `⋯`'s glyph, tap area and visibility, and the Column header part's build. **Extends [ADR-017](adr-017-ui-foundations.md) §9 and supersedes nothing.** ADR-017 §2 fixed the header's outside: it shares the panel's fill, and its bottom rule is load-bearing. [ADR-019](adr-019-responsive-touch-scope.md) §7 fixed its height. Neither said what is inside it. The code gains no CSS variable.

## Context

**Found by question 4, before the header was drawn.** Part 5 of Stage 3 in [`design-file-inventory.md`](../design-file-inventory.md) §7.3 is the Column header. Mapping its values onto the live catalogue on 2026-11-20 gave this:

| Value | Needs | Live token | |
|---|---|---|---|
| Height | 36 / 44 (ADR-019 §7) | `size.column-header` | present |
| Fill | the panel's own (ADR-017 §2) | `color.board-column` | present |
| Bottom rule | 1px, bottom edge only | `color.border`, `border.width` | present |
| Title | 14/20 (ADR-017 §6) | `type.body` | present |
| Title colour | text on the panel | `color.foreground` | present |
| Count | the chip | `Count chip` instance | present |
| `⋯` colour | as on the Card | `color.muted-foreground` | present |
| `⋯` tap area | 24 / 44 | `size.target-min` | present |
| **Inset of the contents** | — | none | **missing** |
| **Gap between title, chip and `⋯`** | — | none | **missing** |

**`spacing.column-padding` is not the inset.** It is the column panel's own padding, and the header already sits inside it: ADR-019 §7's *"16 padding"* is the 8 above the header plus the 8 below the last card. Reusing it inside the header would put the title 16px from the panel's edge, which lines up with nothing.

Three more things had never been decided, and none of them is a token:
- the `⋯`'s glyph: the Card's is a typed 13px character
- whether the column's `⋯` is shown at rest or only on hover, which ADR-019 §6 settles for cards only
- whether the drag-over appearance in inventory §4.2 differs from rest at all, given that §2.14 leaves the header untinted

---

## Decision

### 1. One token, one entry

| Token | Set | Value |
|---|---|---|
| `spacing.inline-gap` | `base` | 8 |

The catalogue goes from 157 to **158**: `base` 29 → 30, the theme and density sets unchanged.

### 2. Where each value comes from

#### Derived

- **The inset is `spacing.card-padding-x`, bound by a rule: the column's title starts where its cards' titles start.** The header sits inside the panel's padding, so it is exactly a card's width, 264. With a 12px inset the title sits 20px from the panel's edge, where every card's title sits. On the right, the `⋯`'s tap area ends where the card's `⋯` slot ends. At Fine both are 24 wide, so the header's glyph sits directly above the cards' glyphs. At Coarse the header's area is 44 and the card's slot 32, so their right edges align and their centres sit 6px apart. This is a reuse [ADR-022](adr-022-form-control-geometry.md) §4's test allows: **one role, not two numbers that match.** If card padding ever moves, the header must move with it, or the column has two left edges. It is the argument [ADR-026](adr-026-count-chip.md) §2 made for binding the chip's height to its minimum width.
- **Height `size.column-header`, and no vertical padding.** The contents are centred, which is [ADR-022](adr-022-form-control-geometry.md) §3's reasoning for controls: a fixed height and a centred line need no second number.
- **Fill `color.board-column`, opaque.** ADR-017 §2: the header *shares* the panel's fill. It is filled rather than transparent because cards scroll under it. There is still no `--column-header` colour (ADR-017 §2).
- **Rule `color.border` and `border.width`, bound to the bottom edge only.** Built in the order [ADR-023](adr-023-spacing-between-objects.md)'s *Operational note* gives.
- **Title `type.body` in `color.foreground`.** The type-scale weights were decided on 2026-11-18, when the catalogue was built. That decision put column headers on `type.body` (400), together with body text, buttons and inputs, and gave card titles a step of their own (`type.card-title`, 500) so that promoting `type.body` would not bold every control. The consequence is accepted: **a column title is lighter than the card titles under it.** The weights are not yet in any ADR; this records the column header's.
- **`⋯` tap area `size.target-min`, 24 / 44.** This is the floor itself, so at Coarse the tap area fills the header's full 44.
- **`⋯` glyph `size.icon` 16 in `color.muted-foreground`**, the Card's `⋯` colour (§3).

#### Chosen

- **`spacing.inline-gap` 8.** It is the gap between neighbours in one row that are *not* an icon and its own label: here, between the title and the chip, and the minimum between the title-and-chip group and the `⋯`. 8 keeps to the 4px grid and to the other inner gaps in the catalogue. It is a literal, not a reference to `spacing.icon-gap`, by ADR-022 §4: the same number, a different job.
- **In `base`, not the density sets.** Type does not grow at Coarse (ADR-019 §7), and the row's only tap target, the `⋯`, already grows to 44 by itself.
- **The name is general on purpose.** The banner (a message and Join) and the board tile (two chips side by side) look likely to need the same gap. Each still asks its own question 4: the token is offered to them, not imposed on them.

> **The banner took it.** Its message and Join are neighbours in one row, and neither is an icon with its own label, so `spacing.inline-gap` is the gap between them. See **[ADR-028](adr-028-banner.md)** §2 (2026-11-20). The board tile has yet to ask.

### 3. The `⋯` is Phosphor `DotsThree`, not a typed character

[ADR-018](adr-018-state-vocabulary.md) §9 makes Phosphor the icon set, so the code renders `DotsThree` rather than a `⋯` character. The file draws Phosphor's Regular path at `size.icon`, in a box bound to `size.icon` ([ADR-024](adr-024-control-and-field-internals.md)'s *Operational note*), filled `color.muted-foreground`. Every value is bound, so §7.5's list of expected rows does not grow.

At 16px Regular the three dots are 1.5px across. That is what the code ships.

**The Card's `⋯` stays a typed 13px character** and stays an expected §7.5 row. Bringing it into line is a separate change, not made here.

### 4. The column's `⋯` is always shown, at every pointer

ADR-019 §6 hides the card's `⋯` until hover at a fine pointer and shows it at a coarse one. **The column's is shown at rest at every pointer:**
- A board holds a handful of columns and may hold hundreds of cards. Cards hide theirs to keep a dense board quiet; one `⋯` per column costs almost nothing.
- Below 640px it is the only way to reorder columns ([ADR-021](adr-021-design-inventory-corrections.md) §4), and a touchscreen has no hover.

The drawing is the same either way, because the Card already draws its `⋯` in every state. This decides what the code does.

### 5. The part: one appearance, no axis

- **One component, `Column header`, with no variant axis.** Inventory §4.2 listed *rest* and *drag-over*. But §2.14 leaves the header untinted, so drag-over changes nothing on the header itself, and a second variant would be a pixel copy of the first that could drift from it. The Card's identical `InFlight` and `DragSource` pair is not a precedent: those are states *of the card*, and drag-over is a state of the column body.
- **The part's board shows it twice instead**, on page `Components`, board `Column header — states`. Each copy sits on a `Doc / Surface / Column` panel bound to the real column tokens (`size.column-width`, `color.board-column`, `radius.surface`, `spacing.column-padding`). Under the first is an untinted body, and under the second a body tinted `color.drag-over`, so the rule can be seen against the tint. The body strip is documentation: it does not decide where board 2.14's tint begins or ends.
- **Build:** a row with centred items. `size.column-header` is bound to height and minimum and maximum height, `spacing.card-padding-x` to the left and right padding, and `spacing.inline-gap` to the gap. The fill and the bottom rule are as in §2. Inside are `Lead` (fills the width: `Title` and a `Count chip` instance, `spacing.inline-gap` apart) and `Menu` (`size.target-min` on all six size properties, centring an `Icon` box bound to `size.icon` that holds the `DotsThree` path).
- **Sample text:** `In Progress` and `12`. Two digits, because a board that sizes itself to its content ignores the chip's minimum width on the canvas (ADR-026's *Operational note*).
- **The menu itself is not this part.** Rename, Move left, Move right and Delete ([ADR-021](adr-021-design-inventory-corrections.md) §3) are Stage 5's *Column `⋯` and inline rename* overlay. This part draws only the button that opens it.
- **A long title truncates to one line in code**, and the chip and the `⋯` never shrink. Penpot cannot draw that truncation, so the board uses a short title.

### 6. The code

No new variable. The header renders 36px tall at a fine pointer and 44px at a coarse one, with 12px side padding and 8px between the title, the count and the `⋯`. It has `--board-column` behind it and `border-bottom: 1px solid var(--border)` under it. The title is one line and truncates. The chip and the `⋯` do not shrink, and the `⋯` is pushed to the right edge. The `⋯` is an icon button: Phosphor `DotsThree` at 16 in `--muted-foreground`, in a 24 × 24 area at a fine pointer and 44 × 44 at a coarse one. It is visible at rest, and its accessible name includes the column's title.

---

## Options Considered

### No inset

Needs no token, and the header's edges become the cards' edges. Rejected: the column's title would sit 12px left of every card title beneath it, so one column would have two left edges.

### Bind `spacing.control-padding-x`

Also 12. [ADR-022](adr-022-form-control-geometry.md) §2 calls a column header *"the same species of object as an input"*, and borrowed `size.column-header`'s height ladder for controls on that ground. Rejected for the inset: what the header's contents must line up with is the cards beneath it, not a control elsewhere on screen. If card padding moved and control padding did not, this binding would split the column's left edge, which is exactly what the inset exists to prevent. (Not listed when question 4 was reported, and added here so the record is complete.)

### `spacing.column-padding` as the inset

Rejected in *Context*: it is already spent outside the header, and spending it again puts the title at 16, which aligns with nothing.

### Borrow `spacing.icon-gap` for the gap

Same value. Rejected by ADR-022 §4: its job is an icon next to its own label inside a control, and a count chip is not an icon.

### A column-only gap token (`spacing.column-header-gap`)

Narrower, so it could never be misused. Rejected: the same gap is likely in the banner and the board tile, and three tokens holding one role is the drift ADR-022 §4 guards against in the other direction.

### Repeat the Card's typed 13px `⋯`

Consistent with the Card today. Rejected by §3: it adds a second expected `typography(13px)` row, and it draws a character the code will never render.

### Hover-revealed `⋯` at a fine pointer, as on cards

Rejected by §4.

### Two identical variants, `Rest` and `DragOver`

Follows §4.2's list to the letter. Rejected by §5.

---

## Out of scope (deferred, recorded)

- **The `⋯` button's own hover, pressed and focus-visible looks.** ADR-018's states govern them. No part yet draws a ghost icon button, and §4.2 does not ask for one here.
- **The column `⋯` menu and inline rename (37).** Stage 5.
- **The Card's `⋯` glyph.** §3.
- **The Foundations board's two loose column headers.** Its elevation ladder has `Column header` boards holding only a title. They predate this part, as the Card's placeholder avatar predated the Avatar. Replacing them with instances is a follow-up.
- **Where a column drag is picked up.** At 640px and up the header is the obvious handle, but §4.2 asks for no appearance for it. The column placeholder belongs to board 2.14.

## Downstream edits this ADR requires

1. **[ADR-017](adr-017-ui-foundations.md): annotation, no rewrite.** Its Status line records §9's fifth extension, and a note under §9's opening points here.
2. **[ADR-021](adr-021-design-inventory-corrections.md): annotation.** A note under §3 says the menu's trigger is always shown (§4).
3. **Penpot `Collabotask App`.** **Done 2026-11-20.**
   - `spacing.inline-gap` added. The catalogue reads 158.
   - The part was drawn: component `Column header`, its main on the `Column header — states` board with one instance beside it.
   - The Foundations board's token catalogue was regenerated: one row added, and the heading, section name and `base` table title now read 158 and 30.
   - Binding check: the part's board reports nothing, the Count chip instance and the icon included. Foundations still reports only its 48 `Hue /` shapes and the five Card `⋯` glyphs.
   - Switched to Dark, to Coarse and to Light + Coarse, and looked at. Dark resolves the header to `#1C1917`, the rule and chip to `#44403C`, the title to `#FAFAF9` and the dots to `#D6D3D1`. Coarse makes the header and the `⋯` area 44 and leaves the title, chip and icon at Fine's sizes. Restored to `theme-light` + `density-fine`. **The rule was confirmed from the shape's data and the generated CSS, not from an export** (*Operational note*).
4. **[`design-file-inventory.md`](../design-file-inventory.md):**
   - §0 and *Related documents* gain this ADR.
   - §1.5's column header row says the `⋯` is shown at rest.
   - §4.2's Column header row says there is no axis, and why.
   - §7.2 records the part and the new count.
   - §7.3 records what the header's question 4 returned.
   - §7.5's *By eye* paragraph notes that an export cannot show the rule.
   - §7.6 ticks Column header.
5. **`CLAUDE.md` › `## Now`:** a pointer.
6. **`docs/build-log.md`:** no entry now, following ADR-020 to ADR-026.
7. **SRS** and **`CONTEXT.md`:** no change. Presentation only.

## Operational note

**Penpot's exports do not draw a stroke whose sides differ in width.** Both PNG and SVG export the column header's rule as a full outline at `stroke-width: 0`, so it is invisible in every export. The Foundations board's headers, built the same way at Stage 1, export identically, so this is the exporter and not the part. The data is right:
- the stroke stores bottom 1 and the other three sides 0
- `border.width` is bound to `strokeWidthBottom`
- `penpot.generateStyle` emits `border-block-end-width: 1px` in `#e7e5e4`, so the rule reaches developer handoff

Confirm the rule from those, or on the canvas. Do not expect to find it in an export.

**`createComponent` on a single board makes that board the main, in place.** It keeps its parent, its position and its bindings, so a part can be built inside its documentation panel and turned into a component where it stands.

**Setting `characters` on a cloned text keeps its typography and fill bindings.** Adding a row to a Foundations catalogue table is therefore: clone the last row, append it, rename it, and set its three cells' text. No rebinding is needed.
