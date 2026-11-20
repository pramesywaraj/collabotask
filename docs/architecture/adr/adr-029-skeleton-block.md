# ADR-029: The Skeleton Block

- **Status:** Accepted
- **Date:** 2026-11-20
- **Scope:** The cold-load skeleton block's fill and height: two new tokens in four entries, three values reused by rule, what a block may be drawn on, its one appearance, and the Skeleton block part's build. **Supersedes [ADR-018](adr-018-state-vocabulary.md) §7's "`--muted` fill" only.** The rest of §7 stands: a static skeleton, blocks at the card geometry, **no shimmer**. **Extends [ADR-017](adr-017-ui-foundations.md) §9**: the code gains `--skeleton` and `--skeleton-height`. It also **corrects** [`design-file-inventory.md`](../design-file-inventory.md) §2's row 2.9, which repeats ADR-018's fill.

## Context

**Found by question 4, before the block was drawn.** Part 7 of Stage 3 in [`design-file-inventory.md`](../design-file-inventory.md) §7.3 is the Skeleton block. Mapping its values onto the live catalogue on 2026-11-20 gave this:

| Value | Needs | Live token | |
|---|---|---|---|
| Fill | ADR-018 §7: `--muted` | `color.muted` | present, **but cannot be seen** (below) |
| Corners | a card's | `radius.surface` | present |
| Width | a card's: the column's inside width | none needed: the block fills the column, as a card does | — |
| **Height** | a card's | none | **missing** |
| Gap between blocks | a card's | `spacing.card-gap` | present |
| What it sits on | a column panel | `color.board-column`, `size.column-width`, `spacing.column-padding` | present |

**The fill was there but could not be used.** In Light, `color.board-column` is not a separate value that happens to match `color.muted`. It is stored as the reference `{color.muted}`. So a muted block inside a column is exactly the column's colour, 1.00:1, for as long as that reference stands. Dark works only because there `color.muted` holds `color.card`'s `#292524`, 1.15:1 on the column.

ADR-018 §7 asks for blocks "at the card and column geometry". A column-sized muted block is visible on the canvas in both themes, but then card-sized muted blocks inside it are invisible in both. **No single fill can draw both sizes nested.**

**The height was missing because a card's height comes from what it holds.** ADR-017 §1 gives it as `padding + lines × 20 + (24 if meta) + padding`, and the Card part is built the same way: a title box and a `size.meta-row` row, inside `spacing.card-padding-y`. A skeleton holds nothing, so nothing sets its height.

---

## Decision

### 1. Two tokens, four entries

| Token | Set | Value |
|---|---|---|
| `color.skeleton` | `theme-light` | `#E7E5E4` |
| `color.skeleton` | `theme-dark` | `#292524` |
| `size.skeleton-height` | `density-fine` | 60 |
| `size.skeleton-height` | `density-coarse` | 72 |

The catalogue goes from 159 to **163**: each theme set 52 → 53, each density set 12 → 13, `base` unchanged.

### 2. Where each value comes from

#### Measured

- **`color.skeleton`, `#E7E5E4` / `#292524`.** Measured against the column panel, the only surface a block sits on in board 2.9:

  | Fill (Light / Dark) | Light, on the column | Dark, on the column |
  |---|---|---|
  | `color.muted` (ADR-018 §7) | **1.00** | 1.15 |
  | **`color.skeleton`** | **1.15** | **1.15** |
  | `color.chip`, borrowed | 1.15 | 1.70 |
  | `color.board-canvas`, borrowed | 1.15 | 1.13 |
  | `color.card`, borrowed, no outline | 1.09 | 1.15 |

  It is the only candidate that is the same in both themes. Its 1.15:1 is Light's canvas → column step, the faintest step ADR-017 §2 accepts as drawn structure. Its Dark value is the card's own fill, so in Dark the block becomes the card it stood for, and only the card's border and text arrive.
- **Written as values, not as references** (`{color.board-canvas}` in Light, `{color.card}` in Dark). This follows [ADR-026](adr-026-count-chip.md): a role gets its own values so that the surfaces it currently matches can move without moving it.

#### Derived

- **`size.skeleton-height`, 60 / 72.** ADR-017 §1's formula with one title line and a meta row: `8 + 20 + 24 + 8` at Fine, `12 + 20 + 28 + 12` at Coarse. A card is 36 to 80 tall at Fine. ADR-017 §1's eight representative cards average 69.5: `(612 − 7 × 8) ÷ 8`. 60 is the real card shape nearest that average (80 is 10.5 away). The values are written as numbers in the density sets because the card's padding and meta row grow at Coarse. The formula's 20 is the title's line height and does not grow, since type stays the same at both densities.
- **Corners `radius.surface`, gap `spacing.card-gap`, width filling the column.** Each is bound by one rule: a block stands where a card will.

#### Chosen

- **No outline.** The card's 1px edge would make the block read as an empty card, and a hollow shape is what the drop placeholder already is (2px dashed, inventory §2.14). A skeleton block is a fill and nothing else.
- **One height.** Mixing heights would invent a spread of card sizes the client does not yet know, and inventory §4.2 gives the part no axis.

### 3. The column is drawn as itself

ADR-018 §7's *"blocks at the card and column geometry"* is read as **card-sized blocks inside a real column panel**. A column-sized block in `color.skeleton` would hide the card blocks inside it, which is the same failure as `color.muted` with a different colour (*Context*). The panel holds no data, so drawing it claims nothing. How many columns a cold load shows is board 2.9's question.

### 4. Where a block may sit

Measured against every surface in the catalogue:

| Surface | Light | Dark |
|---|---|---|
| `color.background` | 1.26 | 1.30 |
| `color.board-column` | 1.15 | 1.15 |
| `color.popover` | 1.26 | 1.15 |
| `color.board-canvas` | **1.00** | 1.30 |
| `color.card` | 1.26 | **1.00** |

**A skeleton block sits on a column panel or on the page background. It never sits directly on the canvas, and never inside a card-coloured surface.** The inventory's other two cold loads, `/workspaces` (§1.3) and workspace detail (§1.4), are drawn at Stage 4 with blocks at tile geometry, and what their page sits on is not yet decided. If it turns out to be the canvas, this section has to be revisited before those boards are drawn. The rule is not a cost of this ADR: `color.muted` fails on the same two surfaces (the column in Light, the card in Dark) and is weaker on the page background (1.09 in Light).

### 5. The part: one appearance, no axis

- **One component, `Skeleton block`, with no variant axis**, as inventory §4.2 says. It is static. A design file cannot draw motion, so the file cannot get "no shimmer" wrong. The code can (§6).
- **The part's board**, on page `Components`, is `Skeleton block — states`. It holds a `Doc / Canvas` bound to `color.board-canvas`, with 16px of documentation padding (ADR-023 *Out of scope*). Inside that is a `Doc / Surface / Column` bound to `color.board-column`, `size.column-width`, `spacing.column-padding`, `radius.surface` and, for its row gap, `spacing.card-gap`. The column holds the main and two instances. The canvas is there so that the Light check shows the full ladder: canvas, then column, then block.
- **Build:** a board with no layout, filled `color.skeleton`, with all four corners bound to `radius.surface`, its height bound to `size.skeleton-height`, and its width set to fill the column. The main was made a component where it stood, inside the column ([ADR-027](adr-027-column-header.md) *Operational note*). Its instances were made in a later call, after the main was verified ([ADR-028](adr-028-banner.md) *Operational note*).

### 6. The code

- `--skeleton`: `#E7E5E4` in Light and `#292524` in Dark.
- `--skeleton-height`: 60px, or 72px under the coarse-pointer density switch ([ADR-019](adr-019-responsive-touch-scope.md) §1), alongside the other density values.
- shadcn's Skeleton ships `animate-pulse`. **Remove it**, because ADR-018 §7 makes the skeleton static: a pulse is movement in the board's content area with no teammate behind it. Do this in the same pass as ADR-018's `disabled:` override.
- Its shipped fill is `bg-muted` or `bg-accent`, depending on the version, and both hold the old value. Replace it with the new variable.
- Its corners are the Card's, whatever class the Card uses. Blocks fill the column's width.

---

## Options Considered

### Keep `color.muted`, as ADR-018 §7 wrote it

Needs no token. Rejected: 1.00:1 in Light, because the column *is* `color.muted` there.

### Keep `color.muted` and drop the column panels

The blocks then sit on the canvas: 1.15 in Light and 1.30 in Dark, so both visible, and ADR-018 stays untouched. Rejected: the skeleton then shows no columns. In Light every block is a column-coloured slab on the canvas, and the board's structure appears only once the data arrives, which is the opposite of what a skeleton is for (ADR-018 §7: *"the board's geometry is known in advance"*).

### Borrow `color.chip`

1.15 in Light and 1.70 in Dark. Rejected for the reason ADR-026 gave the chip its own role: the two would match only by coincidence, and moving one would move the other. It is also uneven between the themes.

### Borrow `color.board-canvas`

1.15 in Light and 1.13 in Dark, with blocks reading as holes in the panel. Rejected: it borrows a role. In Dark it is also darker than the column while the card is lighter, so every block would flip from dark to light as its card arrived.

### Borrow `color.card`, with no outline

1.09 in Light and 1.15 in Dark. Rejected: in Light it is the faintest step on the board with nothing to help it. ADR-017 §2 relies on the card's 1px border to *"do the rest of the work"*, and a skeleton block has none.

### The card's outline on a `color.muted` block

Visible in Light, but only by its edge. Rejected by §2: a hollow card reads as an empty card or as the drop placeholder.

### `color.skeleton` as references

Rejected by §2, following ADR-026.

### Height 80 / 92, as the drawn Card

Draws a skeleton block the same size as the Card part. Rejected: it is the tallest card the formula allows, so a column of blocks overstates the column.

### Height 36 / 44, the shortest card

Rejected: it understates the column, and it equals `size.control-height` and `size.column-header` only by coincidence, so binding either would join two jobs that merely share a number ([ADR-022](adr-022-form-control-geometry.md) §4).

### Height in `base`

Rejected: the card's padding and meta row grow at Coarse, so the block would understate every Coarse card.

### Blocks of mixed heights

Looks less mechanical. Rejected by §2: it invents data, and it needs either an axis or several tokens.

---

## Out of scope (deferred, recorded)

- **Board 2.9's composition**: how many columns and blocks, what a column's header shows while loading (empty, or a bar where the title goes), and the space between a header and the first block. Stage 4, group d.
- **The cold loads of `/workspaces` and workspace detail** (inventory §1.3, §1.4): blocks at tile geometry, under §4's rule. Stage 4, group b.
- **How a loading region is announced to assistive technology.** Code, settled at scaffold.

## Downstream edits this ADR requires

1. **[ADR-018](adr-018-state-vocabulary.md): annotation, no rewrite.** Its Status line records that §7's fill is superseded, and a note under §7 points here. The note also carries this ADR's §6 `animate-pulse` removal, because ADR-018 §7 is where that rule comes from.
2. **[ADR-017](adr-017-ui-foundations.md): annotation.** Its Status line records §9's seventh extension, and a note under §9's opening points here.
3. **Penpot `Collabotask App`.** **Done 2026-11-20.**
   - The four entries were added. The catalogue reads 163.
   - The part was drawn: component `Skeleton block`, with its main and two instances in one column on the `Skeleton block — states` board.
   - The Foundations board's token catalogue was regenerated: a `color.skeleton` row (its swatch bound) and a `size.skeleton-height` row were added. The heading, the section name and the two table titles now read 163, 53 each and 13 each.
   - Binding check: the part's board reports nothing, both instances included. Foundations still reports only its 48 `Hue /` shapes and the five Card `⋯` glyphs.
   - It was switched to Dark and to Light + Coarse, and looked at each way. Dark resolves the blocks to `#292524` on a `#1C1917` column on the `#0C0A09` canvas. Coarse makes the blocks 72 tall with 12 between them and leaves the column's padding at 8. Restored to `theme-light` + `density-fine`.
4. **[`design-file-inventory.md`](../design-file-inventory.md):**
   - §0 and *Related documents* gain this ADR.
   - §2's row 2.9 says `color.skeleton`, card-sized blocks in real column panels.
   - §1.3's and §1.4's cold-load entries point at §4's rule.
   - §4.2 gives the Skeleton block its own row.
   - §7.2 records the part and the new count.
   - §7.3's table row 7 and a new *What question 4 returned* section.
   - §7.6 ticks Skeleton block.
5. **`CLAUDE.md` › `## Now`:** a pointer. The scaffold checklist gains the `animate-pulse` removal.
6. **`docs/build-log.md`:** no entry now, following ADR-020 to ADR-028.
7. **SRS** and **`CONTEXT.md`:** no change. Presentation only.

## Operational note

**A token's `resolvedValue` in an inactive set reports the active set's value.** Right after `color.skeleton` was added, the `theme-dark` entry read `value=#292524` but `resolvedValue=#E7E5E4`, and the Coarse height read `72` but resolved to `60`. The stored values were right, and both resolved correctly once their sets were switched on. So check a new entry's `value` in its own set, and check `resolvedValue` only after a flip.
