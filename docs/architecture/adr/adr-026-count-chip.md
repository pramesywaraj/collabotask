# ADR-026: The Count Chip

- **Status:** Accepted
- **Date:** 2026-11-20
- **Scope:** The count chip's fill, height and side padding, as three tokens (four entries), and the Count chip part's build. **Extends [ADR-017](adr-017-ui-foundations.md) §9 and supersedes nothing.** §3.4 measured `--muted-foreground` against this chip, but §9 never gave the chip a variable of its own. The code gains one CSS variable, `--chip`.

## Context

**Found by question 4, before the chip was drawn.** Part 4 of Stage 3 in [`design-file-inventory.md`](../design-file-inventory.md) §7.3 is the Count chip. Mapping its values onto the live catalogue on 2026-11-20 gave this:

| Value | Needs | Live token | |
|---|---|---|---|
| Number colour | stone-600 / stone-300 | `color.muted-foreground` | present |
| Type | 12/16 (ADR-017 §6: "meta — due dates, counts …") | `type.meta` | present |
| Corners | 4px (ADR-017 §5: "badges, chips, count pills") | `radius.badge` | present |
| **Fill** | `#E7E5E4` / `#44403C` | `color.border`, by value | **borrowed** |
| **Height** | — | none | **missing** |
| **Side padding** | — | none | **missing** |
| `tabular-nums` | equal-width digits | no token type exists | not drawable |

**The fill was measured but never named.** ADR-017 §3.4 chose `--muted-foreground` *because of* this chip: stone-500 fails on `#E7E5E4` (3.82:1), and stone-400 fails on a stone-700 chip in Dark (4.07:1), so the muted text became stone-600 and stone-300. §9's `:root`, though, has no chip variable. The Light value appears in §3.4 next to `--board-canvas`, and the Dark value appears nowhere in §9. The only token that holds both values is `color.border`, and it holds them by coincidence. §9 comments `--border` as *"decorative edge"*, and §4 gives the border channel one meaning: an edge.

**No geometry ADR needed a chip.** [ADR-022](adr-022-form-control-geometry.md) sized controls, [ADR-023](adr-023-spacing-between-objects.md) spaced objects, and [ADR-024](adr-024-control-and-field-internals.md) filled in the inside of fields and controls. A chip is not a control, not a field and not a gap between objects.

---

## Decision

### 1. Three tokens, four entries

| Token | Set | Value |
|---|---|---|
| `color.chip` | `theme-light` | `#E7E5E4` (stone-200) |
| `color.chip` | `theme-dark` | `#44403C` (stone-700) |
| `size.chip-height` | `base` | 20 |
| `spacing.chip-padding-x` | `base` | 6 |

The catalogue goes from 153 to **157**: `base` 27 → 29, each theme set 51 → 52, density sets unchanged at 12.

### 2. Where each value comes from

#### Derived

- **`color.chip`** is ADR-017 §3.4's own pair of values, unchanged. Its contrast figures were measured against exactly these: the number on the chip is **6.08:1** in Light and **6.90:1** in Dark, recomputed for this ADR and matching §3.4.
- **`size.chip-height` 20** is `type.meta`'s 16px line plus 2px above and 2px below.
- **The minimum width is the same token**, bound to `layoutItemMinW`, so a one-digit count is a 20 × 20 square rather than a sliver. Height and minimum width change together by rule. That is the case [ADR-022](adr-022-form-control-geometry.md) §4 reserves for a shared token: one role, not two values that happen to match.
- **`spacing.chip-padding-x` 6** is `(20 − 8) ÷ 2`. Poppins 400 at 12px was measured in the file on 2026-11-20: `0` and `3`–`9` are 8px wide, `2` and `7` are 7px, and `1` is 4px. With 6px a side, the widest digit fills the square exactly, and every two-digit count is wider than it is tall. At shadcn's Badge padding of 8, a single digit would be 24 × 20 and the square would be gone.

#### Placed, not chosen

- **Both geometry tokens live in `base`, not the density sets.** The chip is a type step plus 2px, and [ADR-019](adr-019-responsive-touch-scope.md) §7 leaves type unchanged at Coarse. Nobody taps a count, so Coarse's 44px tap floor does not apply either. The column header grows from 36 to 44 around the chip, and the chip stays centred. Flipped to Coarse on 2026-11-20, every chip on the part's board measured the same as at Fine.

**The chip's edge is faint, and that is ADR-017's look rather than a defect.** Against its surfaces the fill is **1.15:1** on a Light column, 1.26:1 on a Light card, 1.70:1 on a Dark column and 1.48:1 on a Dark card. The chip is not a control, and its edge identifies nothing. The number carries the information, at the ratios above.

### 3. The fill is its own role, not `color.border`

Same value today, different job. ADR-017 §9 already sets the precedent in its own `:root`: `--board-column: #F5F5F4; /* same value as --muted, different role */`.

**The values are literals in each theme set, not `{color.border}`.** A reference would say "the chip is whatever the border is", and that is false. The border is the first step of an edge ladder: inventory §4.2 puts border, hover and pressed on three consecutive stone steps. The chip is a surface, pinned by §3.4's contrast measurement. If the edge ladder ever moves, the chip must stay where it is, or `--muted-foreground` is resting on a value nobody measured. That is ADR-022 §4's reasoning exactly: equal arithmetically, not semantically.

### 4. The part: one appearance, any text

- **One component, `Count chip`, with no variant axis**, as inventory §4.2's "—" says. It sits on page `Components`, board `Count chip — states`.
- **It holds text, not only a number.** The column header shows `12`. A board tile carries two counts (inventory §1.4: member count and card count), and two bare numbers side by side cannot be told apart, so a tile shows `12 cards` and `3 members`. The text is an override on an instance, not an appearance, so the part needs no axis for it. The tile's final wording is settled when boards 2.3 and 2.4 are drawn.
- **No icon slot.** An icon at `size.icon` 16 inside a 20px chip would need a gap value of its own, and words already say which count is which.
- **Build:** a board with a centred flex row. `size.chip-height` is bound to height and minimum width, `spacing.chip-padding-x` to left and right padding, `color.chip` to the fill and `radius.badge` to all four corners. The text is bound to `type.meta` and `color.muted-foreground`. The width follows the content.

### 5. Two things question 4 raised that are not gaps

- **`tabular-nums` cannot be drawn.** ADR-017 §6 requires it on counts. Penpot has no token type for it and no setting for a font's alternate figures, so nothing is typed in instead and nothing is missing from the catalogue. It is code-only, like motion (inventory §4.3). The file shows Poppins's **default** digits, which are proportional, as measured in §2: `11` draws 8px wide and `88` 16px.
- **ADR-019 §8's `+N` overflow chip is not this part.** It opens the presence list as a sheet, so it is a control, sized at 32 to sit in the avatar row. A count chip is a label. §7.3 already places the overflow sheet with the presence cluster (part 9), and that part's question 4 should note that 32 is below Coarse's 44px tap floor.

  > **Not this part, for a different reason.** [ADR-018](adr-018-state-vocabulary.md) §6 makes the whole presence cluster the one target, so `+N` is a label inside it and the 44px floor is met by the cluster, `size.target-min` tall. **[ADR-031](adr-031-presence-cluster.md)** (2026-11-20) then draws `+N` as a round seat at a face's size rather than as a Count chip instance, by the user's choice: it sits in the stack, overlaps the last face and wears the same ring.

### 6. The code

ADR-017 §9 gains one variable in each mode: `--chip: #E7E5E4` in `:root` and `--chip: #44403C` in `.dark`, exposed as a Tailwind utility alongside `--board-canvas`. The chip renders at 20px height and minimum width, 6px side padding, the 4px radius of §5, 12/16 text in `--muted-foreground`, with `tabular-nums`.

**Scaffold check:** confirm that the self-hosted Poppins build carries the `tnum` feature. If it does not, `tabular-nums` does nothing, a count jitters as it changes, and a `1` chip is held square only by its minimum width.

---

## Options Considered

### Bind the fill to `color.border`

No new tokens, and the values match in both modes today. Rejected by §3: it ties a surface to the edge ladder, so a change to card edges would silently move the surface §3.4's contrast proof was measured on.

### `color.chip` as a reference to `{color.border}`

Names the role, so a shape binds `color.chip`. Rejected: the link stays in place under a new name. That is the relationship ADR-022 §4 declines to assert between two values that are only arithmetically equal.

### Bind the fill to `color.board-canvas`

Matches §3.4's Light row, where the chip shares the canvas value. Rejected: the Dark canvas is `#0C0A09`, darker than the column the chip sits on, so the chip would read as a hole in the header.

### Vertical padding instead of a fixed height

`2` above and below, with the height following the text. Rejected: the square needs a known height for its minimum width, and the code ships a fixed height.

### Put the geometry in the density sets

Would grow the chip at Coarse. Rejected by §2: its type does not grow, and it is not a target.

### Padding 8, as shadcn's Badge

Rejected by §2's arithmetic: one digit would be 24 × 20.

### An icon slot for the tile's counts

Rejected by §4.

---

## Out of scope (deferred, recorded)

- **The board tile's composition and wording** (inventory §1.4). Settled with boards 2.3 and 2.4. This ADR only shows that the chip can hold the words.
  > *2026-11-21:* settled in [ADR-033](adr-033-workspace-screens.md) §3 and §12. A board tile reads *12 cards · 3 members* and a workspace tile *5 members · 4 boards*, with singulars, after the role, `spacing.inline-gap` apart.
- **The `+N` overflow chip.** Part 9, the presence cluster.

  > **Settled as a seat.** **[ADR-031](adr-031-presence-cluster.md)** §3 (2026-11-20): `size.avatar` tall, `radius.full`, `spacing.seat-padding-x` 2 / 4, filled `color.avatar-neutral`.
- **A cap on large counts** (`999+`). Columns are unbounded (ADR-017 §2), so the chip simply grows. Truncating a count is a product decision nobody has asked for.

## Downstream edits this ADR requires

1. **[ADR-017](adr-017-ui-foundations.md): annotation, no rewrite.** Its Status line records §9's extension. A note under §9's opening and another under §3.4 point here.
2. **Penpot `Collabotask App`.** **Done 2026-11-20.**
   - The four entries were added. The catalogue reads 157.
   - The part was drawn: `Count chip — states` on `Components`, with `3`, `12` and `128` on a column-coloured surface, and `12 cards` and `3 members` on a tile-coloured one.
   - The Foundations board's token catalogue was regenerated: three rows added, and its heading and table titles now read 157, 29 and 52.
   - Binding check: the part's board reports nothing. Foundations still reports only its 48 `Hue /` shapes and the five Card `⋯` glyphs.
   - Switched to Dark and to Coarse and looked at. Dark resolves the chip to `#44403C` with `#D6D3D1` numbers. Coarse leaves every chip at Fine's size. Restored to `theme-light` + `density-fine`.
3. **[`design-file-inventory.md`](../design-file-inventory.md):**
   - §0 and *Related documents* gain this ADR.
   - §4.2's Count chip row and §4.3's *What cannot be drawn* gain a line each.
   - §7.2 records the part and the new count.
   - §7.3 records what the chip's question 4 returned.
   - §7.6 ticks Count chip.
4. **`CLAUDE.md` › `## Now`:** a pointer, and the `tnum` check on the scaffolding list.
5. **`docs/build-log.md`:** no entry now, following ADR-020 to ADR-025.
6. **SRS** and **`CONTEXT.md`:** no change. Presentation only.

## Operational note

**Penpot does not draw a minimum width on a board whose width follows its content.** The chip's `layoutItemMinW` is bound, and the shape stores it (`layoutChild.minWidth` reads 20). Penpot's generated CSS emits `min-width: 20px`, so the rule reaches developer handoff. The canvas ignores it, though: a chip holding `1` draws 16 wide. Two workarounds were tried on 2026-11-20 and neither took effect: writing the minimum again, and putting a minimum on the text instead. So in this file a chip holding `1`, `2` or `7` draws narrower than the app will render it. The part's board uses `3`, whose digit fills the square exactly.

**A minimum width cannot be cleared by writing `null`.** `layoutChild.minWidth = null` is refused with *Value not valid*. Replace the instance instead.
