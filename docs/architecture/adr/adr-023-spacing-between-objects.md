# ADR-023: Spacing Between Objects

- **Status:** Accepted
- **Date:** 2026-11-20
- **Scope:** The space *between* objects in the app zone — the screen's horizontal inset, the gap between board columns, the canvas's vertical padding, and the break between page sections — plus three values the token catalogue lacks for the same reason: the 4px board rail, the two border widths, and the size of avatar initials. **Extends [ADR-017](adr-017-ui-foundations.md) §9 and adds one row to [ADR-020](adr-020-design-tool-penpot.md) §3's type mapping; supersedes nothing.** Six of the eight values are derived from accepted ADRs — four stated outright, two implied only by their arithmetic — and are merely missing from the catalogue; two are choices, and are marked as such. No code is written here.

## Context

**Found the way [ADR-022](adr-022-form-control-geometry.md) was — before anything was drawn.** Stage 1 of [`design-file-inventory.md`](../design-file-inventory.md) §7.3 is the foundations board. Mapping every value it needs onto the live catalogue on 2026-11-20 found one family missing entirely.

**Every spacing token in the catalogue sits inside one object.** Card padding, column padding and control padding are inside an object; card gap and field gap separate items *within* one object (cards in a column, fields in a form). Nothing names the space between two objects side by side on a screen. ADR-017 §9's vocabulary is the inside of a board, and ADR-022's is the inside of a form, so neither ever needed it.

The foundations board is the first thing in the file that puts objects next to each other on a page, so it found the gap first. It is not the only thing that needs it:

| Value | Needed by |
|---|---|
| Inset from the screen edge | Every screen: canvas, app bar, page body |
| Gap between board columns | Board 2.5 and everything that varies from it; the elevation ladder on the foundations board |
| Canvas padding above and below the columns | The same |
| Break between page sections | The foundations board; boards 2.3 and 2.4 |

**Three more values surfaced in the same pass, of the same species.** The 4px board rail (ADR-017 §3.3), the 1px border (ADR-017 §2, §4) and the 2px ring and drop placeholder (ADR-018) are all fixed by accepted ADRs, yet none of them is a token. The border widths were invisible to the inventory's §7.5 binding check for a separate reason: as written, that check looks for property names Penpot shapes do not have. Eight of its nine watched names never match, so it cannot flag an unbound fill, padding or gap at all. A corrected predicate, run against the existing Card on 2026-11-20, reports exactly two literals: the 1px stroke width and the 13px `⋯` glyph. The first of those is closed here.

**Why this is an ADR and not a note** is ADR-022's reason, unchanged: every screen will reference these values, and changing a value's role after twenty components do is the sweep ADR-017's Context warns about.

---

## Decision

### 1. Six tokens, one zero, one reuse

| Token | Penpot type | Set | Fine | Coarse |
|---|---|---|---|---|
| `spacing.screen-x` | `spacing` | `base` | **16** | 16 |
| `spacing.column-gap` | `spacing` | both density sets | **8** | **12** |
| `spacing.section-gap` | `spacing` | `base` | **24** | 24 |
| `size.board-rail` | `sizing` | `base` | **4** | 4 |
| `border.width` | `borderWidth` | `base` | **1** | 1 |
| `border.width-strong` | `borderWidth` | `base` | **2** | 2 |

**Canvas vertical padding is 0**, and is not a token (§2.3).

**Avatar initials use the existing `type.meta`** (12/16, 400). No new type step (§2.6).

The catalogue goes from 109 tokens to **116**: `base` 19 → 24, each density set 11 → 12, the theme sets unchanged.

### 2. Where each value comes from

As in ADR-022, derived values and chosen ones are separated, so a later reader knows which are safe to re-open.

#### Derived

**2.1 — `spacing.screen-x` 16.** [ADR-019](adr-019-responsive-touch-scope.md) states it three times, independently:

| Where | What it says |
|---|---|
| §2, the width table | "Usable (**−32px padding**)" at every width from 640 to 1440 |
| §3, the phone strip | "**16 left pad**" at 390, 430 and 375 |
| §8, the narrow app bar | "375px **−32 padding**" |

The same 16 applies at both layouts and on both pointers. §2's rows include iPads (coarse) and desktops (fine); §3's are phones (coarse). So the value goes in `base`. It is the horizontal padding of the board canvas, the app bar and any page body.

**2.2 — `spacing.column-gap` 8 / 12.** ADR-019 §3 states **12** outright for the phone strip: `16 left pad │ 12 gap │ 44px peek`, which is what makes the 72px reservation exact (16 + 12 + 44). §2's table never names a gap, but it reconciles only at **8**:

| Width | Usable | Arithmetic at 8px | Table says |
|---|---|---|---|
| 640 | 608 | 608 − 2×280 − 8 = **40** | 2 + a 40px peek |
| 744 | 712 | 712 − 2×280 − 8 = **144** | 2 + 144px |
| 834 | 802 | 802 − 2×280 − 8 = **234** | 2 + 234px |
| 1024 | 992 | 992 − 3×280 − 2×8 = **136** | 3 + 136px |

The 640 row alone forces it: 608 − 560 − *g* = 40 has only one solution. The 1440 row is the exception. It reads "5", which holds only with no gap at all; at 8px it is four columns and 256px of a fifth. It is rounded, and nothing depends on it.

**Which axis.** The two tables can be read as a *width* rule (8 at 640 and up, 12 below) or as a *pointer* rule (8 fine, 12 coarse). This ADR takes the **pointer** reading, because:

- ADR-019 §1 divides the work explicitly: width drives "which board layout, modal presentation, bar composition", and pointer drives "card padding, avatar size, **gap**, minimum target". Columns are drop targets, and space between targets is a size.
- Penpot has a pointer axis (the `Density` theme group, ADR-020 §2) and no width axis; the phone board is drawn by hand. A width rule would need two tokens for one role.
- It costs at most 4px per gap. An iPad (≥640, coarse) gets 12 rather than §2's implicit 8, so the iPad-mini-portrait peek goes from 144 to 140px. A desktop window narrower than 640 (fine) gets 8 rather than 12, so its peek grows from 44 to 48px. No column count changes, and neither peek falls below its floor.

**It equals `spacing.card-gap` at both densities, and stays a separate literal.** ADR-022 §4's test applies: the two match arithmetically, not semantically. The column gap is pinned by §3's peek arithmetic and the card gap by §7's card count. If the card gap ever moved and the column gap followed it, the 44px peek would break.

**2.3 — Canvas vertical padding 0.** Both accepted card-count derivations subtract exactly **16** between the app bar and the column header:

- ADR-017 §1: 900 − 88 − 56 − **16** − 36 = 704
- ADR-019 §7: 844 − 140 − 56 − **16** − 44 = 588

ADR-017 §2 gives the column panel 8px of padding on every side, and 8 above the header plus 8 below the body is that 16. So **both derivations already assume the canvas adds nothing above or below the columns.** The column panel runs from directly under the app bar to the bottom of the screen, separated by its own fill and padding.

This is recorded as a decision because the alternatives cost something ADR-019 asked for by name. With 16 above and below, the phone's seventh-card peek (ADR-019 §7's "the column continues" cue, 36px) falls to 4px, which is effectively gone. With 16 above only, it falls to 20px. At Fine, neither alternative breaks the eight-card target.

It is not a token. A zero has nothing to bind, and the corrected binding check flags only non-zero values.

**2.4 — `size.board-rail` 4.** ADR-017 §3.3: "a 4px rail at the top of the board". Already fixed; only missing. If the rail sits below the app bar rather than inside its 56px, it takes 4px that neither derivation in §2.3 subtracts: the phone peek goes from 36 to 32px and the fine spare from 92 to 88px. No target moves.

**2.5 — `border.width` 1 and `border.width-strong` 2.**

- **1px** is the card edge and the column header rule (ADR-017 §2, §4). §4 puts the input boundary on the same border channel; §3.4 fixes its colour, never a width, so it takes this one. It appears on nearly every bordered object in the app.
- **2px** is the focus-visible ring and the dashed drop placeholder, both from [ADR-018](adr-018-state-vocabulary.md), as quoted in inventory §1.1 and §2.14. ADR-017's channel budget left "border weight (2px)" unspent, and ADR-018 spent it once state 18 (*Selected*) was found not to exist in Phase 1.

This adds a row to [ADR-020](adr-020-design-tool-penpot.md) §3's mapping, which that ADR calls fixed:

| Source | Penpot `TokenType` | Applied via `TokenProperty` |
|---|---|---|
| Border widths (1px edge, 2px ring and placeholder) | `borderWidth` | `strokeWidth`; `strokeWidthBottom` for the column header's bottom rule |

**Why not leave them as known literals.** Two reasons. A width that carries an assigned meaning (2px = "where it will land") cannot be checked unless it is a token. And with the check corrected, a literal 1px is flagged on every bordered shape in the file. A check that is always noisy stops being read.

#### Chosen

**2.6 — Initials at `type.meta`, 12/16, weight 400.** No ADR fixes it. It fits both avatar sizes (a 16px line in 20 and 24px circles), and ADR-017 §3.3's initials contrast range, 4.65:1 to 7.96:1, clears 4.5:1 for normal-weight text at its lowest pair. So no heavier weight is needed for legibility, and no seventh type step is warranted. `type.card-title` became a sixth step to keep controls from turning bold; initials have no such conflict.

**2.7 — `spacing.section-gap` 24.** The rule: **a section break must be larger than every gap inside a group, at either density**, or a reader cannot tell a new section from the next item. The largest within-group gap in the catalogue is `spacing.field-gap` at Coarse, 20. On the 4px grid every spacing value sits on, the smallest value above 20 is 24. A section break is not spacing between touch targets, so it does not flip with density and goes in `base`.

It is used between page-level blocks, and above the first and below the last block of a page that scrolls. The board canvas is not such a page: its vertical padding is §2.3's zero.

---

## Options Considered

### Read the column gap as a width rule (8 at 640 and up, 12 below)

The literal reading of ADR-019's two tables. Rejected for the reasons in §2.2: §1 puts gaps on the pointer axis, Penpot has no width axis, and the cost of the pointer reading is ≤4px per gap with no visible consequence.

### Pad the canvas by 16 on every side

Uniform with `spacing.screen-x`, and it would look more conventional. Rejected on §2.3's arithmetic: it removes the seventh-card cue ADR-019 §7 describes, and contradicts the derivation both card-count tables already rest on.

### A numeric spacing scale (`space.4`, `space.8` … `space.32`)

What most design systems ship. Rejected because it names numbers rather than roles. Every spacing token in this catalogue is named for what it separates, which is what lets ADR-022 §4 reason about whether two equal values are the same value. A numeric scale invites "pick the nearest step", which is precisely what the binding check exists to prevent, and it would hide which values are derived.

### Borrow existing tokens for the foundations board's layout

Lay the board out with `spacing.field-gap` or `spacing.card-gap`. Rejected: both flip with density, so the documentation's own layout would move on Fine → Coarse, and changing the card gap would reflow the foundations board. Borrowing a token for a role it does not have is the coupling ADR-022 §4 refused.

### Add the values to ADR-017 §9 directly

Rejected for ADR-022's reasons, unchanged: later ADRs cite §9's contents by reference, and the dates would lie.

---

## Out of scope (deferred, recorded)

- **The gap between tiles on boards 2.3 and 2.4.** Not fixed anywhere. `spacing.column-gap` is the natural candidate (tiles, like columns, are side-by-side targets), but that is a choice for when those boards are drawn, not a derivation available now.
- **The focus ring's 2px offset** (ADR-018, inventory §1.1). Fixed by ADR-018 but not a stroke width; Penpot strokes have an alignment but no offset property, so how it binds is a Stage 2 question. Expected to surface there as Stage 2's question 4.
- **Count chip and overlay interior geometry.** Stage 3 and Stage 5 will ask their own question 4. Neither is needed to draw the foundations board.
- **Sizes used only by the foundations board** (swatch size and similar). Not product values, so not tokens. How the binding check treats them is a rule for inventory §7.5, which is that document's own working plan and freely revisable.
- **Motion and the identity palette.** Unchanged, and still not tokens (ADR-020 §3). The 16 identity hex values on the foundations board remain deliberate literals.
- **The Astro landing page** and **Phase-2 deferrals.** Untouched.

## Downstream edits this ADR requires

1. **[ADR-017](adr-017-ui-foundations.md): annotation, no rewrite.** Its Status line gains `§9 extended again by ADR-023 (date) — spacing between objects, rail height, border widths.`, and §9's head note gains a pointer. No decision text changes.
2. **[ADR-020](adr-020-design-tool-penpot.md): annotation, no rewrite.** Its Status line gains `§3's type mapping extended by ADR-023 — borderWidth.`
3. **[ADR-019](adr-019-responsive-touch-scope.md): no edit.** Its §2 table's rounded 1440 row is recorded in §2.2 above rather than corrected there.
4. **Penpot `Collabotask App`: add the six tokens.** Five go into `base` and `spacing.column-gap` into each density set. Then verify two things not yet tried in this file: that a `borderWidth` token binds to `strokeWidth` and to `strokeWidthBottom`, and that the flip moves the column gap from 8 to 12 with its set made active, not read (ADR-022's operational note). **Done 2026-11-20**, on a throwaway probe board that was then removed. All seven entries were added and the catalogue reads 116. `border.width` resolved to 1 on a full stroke and on a bottom-only rule (0 / 0 / 1 / 0). `border.width-strong` resolved to 2. `size.board-rail` gave a 4px height. With `density-coarse` made active, the column gap moved 8 → 12, the `base` values did not move, and every binding survived. The active sets were restored to `theme-light` + `density-fine`.
5. **[`design-file-inventory.md`](../design-file-inventory.md):** §0's source table gains this ADR. §7.2's counts move to 116. §7.3 records Stage 1's question 4 and that it closed here. §7.5's quality check is replaced by the corrected predicate, together with the rule for documentation-only sizes.
6. **`CLAUDE.md` › `## Now`:** a pointer, as for ADR-022.
7. **`docs/build-log.md`:** no entry now, following ADR-020, ADR-021 and ADR-022.
8. **SRS** and **`CONTEXT.md`:** no change. Presentation only; nothing in Workspace / Board / Column / Card / Position / the role model moves.

## Operational note

**Per-side stroke widths live on the stroke, not on the shape.** The Penpot plugin overview documents `shape.strokeWidthTop` / `Right` / `Bottom` / `Left`. In this file those read as `undefined` on a board, and writing `strokeWidthTop = 0` to the shape silently did nothing. The values are on `shape.strokes[0]`: `strokeWidthTop`, …, plus `strokeWidthType: "multiple"`.

**Rewriting `strokes` drops every stroke binding**, the way rewriting `fills` drops a fill binding. So a column-header rule is built in this order: first write the stroke with top, right and left at 0 and bottom at 1, then `applyToken(border.width, ["strokeWidthBottom"])`. Bind first and write the shape second, and the binding is gone without any error.

**Binding `border.width` to `strokeWidth` also binds all four sides.** `shape.tokens` reports `strokeWidth` and the four `strokeWidth*` properties together. A check for "is this stroke width bound" should therefore accept any `strokeWidth*` key, or a bottom-only rule will be reported as unbound.
