# ADR-036: One-Sided Rules Are Drawn as Inset Shadows

- **Status:** Accepted — *Out of scope*'s **rule on a shape that needs another shadow met for the first time by [ADR-037](adr-037-board-at-rest.md)** (2026-11-23): the `Add card` row binds `shadow.focus-ring`, so the top rule it shows while stuck to a long column's bottom sits on a `Footer` wrapper, as suggested here. **A fourth rule, `shadow.rule-left`, added by [ADR-038](adr-038-board-states-group-d.md)** (2026-11-23), for the card drawer's left edge: x offset `{border.width}`, once per theme set. Everything here stands as written.
- **Date:** 2026-11-22
- **Scope:** How the design file draws a line along one edge of a shape: the sidebar's right edge, the bottom line of the app bar, the column header, the banner and the sidebar's top row, and the line above the sidebar's user row. That is **three tokens, six entries** (200 → 206). **The design file only: the code is unchanged.** It **replaces the drawing method** in three earlier operational notes, and changes no decision in any of them:
  - [ADR-023](adr-023-spacing-between-objects.md)'s first two paragraphs (per-side stroke widths, and the order to build them in)
  - [ADR-027](adr-027-column-header.md)'s, whose diagnosis was wrong
  - [ADR-035](adr-035-sidebar-loading-failure-and-switcher-mark.md)'s two bullets on per-side strokes

  Every rule keeps the colour, width and edge its ADR gave it: `color.border`, `border.width`, one side.
- **Upstream edits:** none.

## Context

**Reported by the user on 2026-11-22, looking at the canvas:** there was no line between the sidebar, the app bar and the page on 2.4. In the design panel, the `Sidebar` part's *Closed, Rows* variant showed a stroke width of 0.

**The cause.** Penpot keeps two kinds of stroke width: one general width, and a width for each side (`strokeWidthTop`, `…Right`, `…Bottom`, `…Left`, with `strokeWidthType: "multiple"`). The plugin API documents the per-side widths and the file saves them. **Nothing draws them.** The design panel shows the general width, and the canvas and the exporter draw only that. Every one-sided rule in the file had been built as a general width of 0 plus 1 on one side, the method [ADR-023](adr-023-spacing-between-objects.md)'s *Operational note* recorded at Stage 1. So every one was invisible. That was **50 shapes**, in five parts and on every screen board drawn so far.

**How it went unnoticed.** [ADR-027](adr-027-column-header.md) saw the column header's rule missing from its exports. It blamed the exporter, confirmed the rule from the shape's data and generated CSS, and said to check the canvas. Nobody looked at the canvas until 2026-11-22. [ADR-028](adr-028-banner.md) and ADR-035 followed the same advice. Two consequences:
- [ADR-017](adr-017-ui-foundations.md) §2 calls the column header's rule load-bearing, and it has never been visible in the file.
- [ADR-028](adr-028-banner.md) §2 relies on the banner's rule as its only lower edge in Dark, and it never drew.

The decisions were right. Only the drawings were wrong.

**Two more defects, found while measuring:**
- **The user row (ADR-035 §5) drew as a full box.** Its rule had a general width of 1 and widths 1 / 0 / 0 / 0. Since only the general width is drawn, it outlined the whole row. ADR-035's note recorded this as an export quirk.
- **The `Sidebar` part's *Closed, Rows* variant had no right edge even in its data.** Its stroke read 0 on all four sides. It is the variant on 2.4's main board, *Archived* and *Archived, empty*, so three of 2.4's six boards lacked the rule twice over.

## Decision

### 1. Three shadow tokens, one per edge

A rule is a **1px inset shadow with no blur and no spread, in `color.border`**, offset inward from the edge it marks.

| Token | Offset x | Offset y | Colour | Light | Dark |
|---|---|---|---|---|---|
| `shadow.rule-top` | 0 | `{border.width}` | `{color.border}` | 0, 1, `#E7E5E4` | 0, 1, `#44403C` |
| `shadow.rule-right` | `{border.width} * -1` | 0 | `{color.border}` | −1, 0, `#E7E5E4` | −1, 0, `#44403C` |
| `shadow.rule-bottom` | 0 | `{border.width} * -1` | `{color.border}` | 0, −1, `#E7E5E4` | 0, −1, `#44403C` |

- **Each token is in both `theme-light` and `theme-dark`**, as `shadow.focus-ring` and `shadow.avatar-ring` are. The values are identical, since both references follow the active theme.
- **The thickness stays bound to `border.width`.** Penpot evaluates the arithmetic in a shadow token's offset, and `{border.width} * -1` resolves to −1. So a change to `border.width` moves every rule, as it did when rules were strokes.
- **No left-edge token.** No part has a left-only rule. Add `shadow.rule-left` if one ever does.

### 2. Every one-sided rule uses one, and carries no stroke

The stroke is removed and the token is bound to `shadow`. That is **54 shapes**: the 50 above, plus the *Closed, Rows* variant and its three instances, which now draw the right edge they never had.

| Rule | Token | Where |
|---|---|---|
| Sidebar, right edge | `shadow.rule-right` | the five `Sidebar` variants, and the instance on each of 2.4's six boards |
| Sidebar top row, bottom | `shadow.rule-bottom` | inside each variant and instance |
| App bar, bottom | `shadow.rule-bottom` | 2.3's four boards, 2.4's six, `Sidebar — states`' *App bar — sidebar hidden*, and the four `Doc / App bar` strips on `Components` |
| Column header, bottom | `shadow.rule-bottom` | the part's main, the drag-over instance, and the two loose headers on `Foundations` |
| Banner, bottom | `shadow.rule-bottom` | the part's main and the phone instance |
| User row, top | `shadow.rule-top` | inside each `Sidebar` variant and instance |

### 3. A shape with a rule must be filled

**An inset shadow draws nothing on a shape with no fill.** A test board showed this. Thirteen shapes had no fill, and each now has the colour it sits on, so nothing else changes on screen:
- the 11 user rows take `color.background`
- the 2 loose column headers on `Foundations` take `color.board-column`

The part's own Column header main was already filled `color.board-column`.

**A board's children never cover its inset shadow.** The test board showed this too. It matters for the sidebar, whose top row is full width and filled white.

### 4. The code is unchanged

These three tokens exist only in the design file, and the code gains no CSS variable. A rule is still built as a border in the code: 1px in `--border` on its one side (`border-r`, `border-b`, `border-t`). Penpot's inspect tab will report these shapes as `box-shadow: inset …`. **Build them as borders anyway.** An inset shadow is painted beneath a box's children, and a border is not.

## Options considered

- **Per-side stroke widths, the old method.** Rejected: Penpot saves them and draws nothing.
- **A 1px rectangle pinned to the edge.** It draws, and it was tested. It was rejected for three reasons:
  - It adds 54 layers, each positioned absolutely inside a layout.
  - Its thickness is a width or a height, which takes a sizing token, not `border.width`, so it would need a token of its own.
  - The inspect tab would show an extra element.
- **A full stroke clipped by the parent screen.** Rejected without testing: every part's main sits unclipped on `Components` and would show a box.
- **An inset shadow (chosen, the user's go-ahead on 2026-11-22).** It draws on the canvas and in exports, follows the Theme switch, stays bound to `border.width`, adds no layers, and stretches with its shape.

## Out of scope (deferred, recorded)

- **A rule on a shape that needs another shadow.** A shape binds one shadow token. No rule-carrying shape has another shadow today. If one ever does, such as a focusable header, give it a combined token or move the rule to a wrapper. Decide that when it arises.
- **Whether the rule is too faint in Light.** It measures 1.26:1 on white, which ADR-034 §10.2 accepted as decorative. The user has now seen it drawn for the first time. Strengthening it would be a design decision and a new ADR.

## Downstream edits this ADR requires

1. **ADR-023, ADR-027, ADR-028 and ADR-035: annotation, no rewrite.** Each Status line points here.
2. **[`design-file-inventory.md`](../design-file-inventory.md):** §7.2's token count; §7.5's bullet on bound stroke widths, its *What it does not see* list, and its closing paragraph on exports, which now inverts; and a line in §7.6's checklist.
3. **The Foundations board's catalogue**, which is static text: three rows added to the theme table, cloned from `shadow.avatar-ring` with each swatch bound again to its own token, and the headings updated to *206 tokens* and *63 each*.
3. **CLAUDE.md › `## Now`:** the token count and a pointer here.
4. **ADR-017 §9, ADR-020 §3, the SRS and `CONTEXT.md`:** no change. No CSS variable, no new token type, and nothing in the domain moves.

## Operational note

**Redrawn 2026-11-22.** No per-side stroke is left anywhere in the file. The whole-file binding check reports only what §7.5 expects:
- the 48 `Hue /` pairings on `Foundations`
- the Card's `⋯` glyphs (13 across the file)
- each page's `Root Frame`, the page background, which the check is not meant to see because it is run per board

Exports were looked at:
- 2.4's main board and *Archived* in Light + Coarse
- the cold-load board and the Column part in Dark + Fine

The file was left in Light + Coarse.

What drawing found:
- **The archived board's `Sidebar` instance had an 8px radius on all four corners.** ADR-035's check did not report it, so it appeared afterwards, and how is not known. It was reset to 0, like its main.
- **Writes fail on a page that is not open** (*Cannot modify a page that is not currently active*). `penpot.openPage` takes effect only after the call ends, so open the page in one call and write in the next.
- **`shape.strokes = []` also removes the stroke's token bindings.** But `applyToken` lands after a call's other writes. So binding a stroke and clearing it in the same call leaves the stroke in place, bound.
- **A density flip can lay out a stale board again.** 2.4's main board exported at Fine sizes while Coarse was active, until a flip to Dark + Fine and back laid it out at Coarse.
- **An export draws text at the size stored in the file, and that size can be stale.** On 2026-11-22 the banner's phone board exported with its message cut after one line and *Join board* cut off: the label was stored 47 wide. That was reported to the user as a defect. It was not one. On 2026-11-23, after the user opened `Components` in the editor, Penpot had measured the text again: the message 250 × 40 on two lines, the label 76. The export came out right, as ADR-028 §3 and §4 say. **Before calling text clipped, open its page in the editor, or ask the user to look.** Strokes were different: the canvas did not draw per-side widths either.
