# ADR-030: The Toast

- **Status:** Accepted — **§3's *no action button* superseded for one toast by [ADR-040](adr-040-menus-and-the-confirmation.md)** (2026-11-23): after you archive the board you're on, the toast offers *Undo*, as US-12c and ADR-021 §1 promise. §3's *Phase 1 has no undo* was written without US-12c in view. The Toast part gains that look in Stage 5's group e, with where a toast sits and the wordings for states 4, 27 and 32. Every other toast is unchanged. **The toast's placement, stacking, duration and the wordings for states 4, 27 and 32 settled by [ADR-044](adr-044-the-sidebar-sliding-over-and-the-toast.md)** (2026-11-24), closing *Out of scope*'s first three bullets: a toast sits 16 from the bottom and the right of the screen (across the bottom on a phone), up to three stack with the newest at the bottom, 8 apart, and a toast stays until it is closed, pushed out by a fourth or the page changes, so no time limit applies. A failure reads *Couldn’t move “Final pricing sign-off”.* and one of five reasons. The archive toast's *Undo* is an Outline `Button` copy, hidden in the part's main and shown by its copies, so §3's *no axis* stands.
- **Date:** 2026-11-20
- **Scope:** The toast's surface and geometry: four new tokens in six entries, and five values reused by rule. Also its one appearance, its close button, the three access-revoked wordings, where losing a workspace lands you, and the Toast part's build. **Extends [ADR-017](adr-017-ui-foundations.md) §9**: the code gains `--toast`, `--toast-foreground` and `--toast-width`. **Supersedes [ADR-018](adr-018-state-vocabulary.md) §10's row 26 for one of its three reasons only**: after `removed_from_workspace` the app goes to `/workspaces`, not to the workspace. It settles the toast's own design and the access-revoked wordings, which ADR-018 and [ADR-021](adr-021-design-inventory-corrections.md) left out of scope. The other toasts' wordings and the toast's placement stay deferred. It also **corrects** [ADR-022](adr-022-form-control-geometry.md)'s *Out of scope* claim that a toast needs no size token.

## Context

**Found by question 4, before the toast was drawn.** Part 8 of Stage 3 in [`design-file-inventory.md`](../design-file-inventory.md) §7.3 is the Toast. Its row asks for *"the container and the longest message it must hold"*. Mapping its values onto the live catalogue on 2026-11-20 gave this:

| Value | Needs | Live token | |
|---|---|---|---|
| Fill | shadcn's toast surface | `color.popover` | present, **but cannot be seen** (below) |
| Edge | shadcn's | `color.border`, `border.width` | present, **and too faint to rescue the fill** |
| Corners | an overlay's | `radius.surface` | present |
| Elevation | an overlay's | `shadow.overlay` | present |
| **Padding** | — | none | **missing** |
| **Width** | — | none | **missing** |
| Message | body text | `type.body` | present |
| Gap before the close button | neighbours in one row | `spacing.inline-gap` | present |
| Close glyph and its tap area | an icon in a control | `size.icon`, `size.target-min` | present |

**A toast floats over everything, so its fill was measured against everything.** On the board it can land over the canvas, a column or a card. After losing access (state 26) you are on a workspace screen, over the page background. shadcn's fill is `color.popover`:

| Light | canvas | column | card | page |
|---|---|---|---|---|
| `color.popover` fill | 1.26 | 1.09 | **1.00** | **1.00** |
| `color.border` edge | 1.00 | 1.15 | 1.26 | 1.26 |

| Dark | canvas | column | card | page |
|---|---|---|---|---|
| `color.popover` fill | 1.13 | **1.00** | 1.15 | 1.13 |
| `color.border` edge | 1.92 | 1.70 | 1.48 | 1.92 |

In Light, a toast over a card or a white page is a 1.26:1 outline and a soft shadow. In Dark it disappears over a column. A board's cards already live at these ratios, but a card is always there and is looked for. A toast arrives unannounced, and it carries news such as *you were removed*.

**The padding and the width were missing because nothing before the toast was a floating box of fixed width.** ADR-022 said a toast needs no size token to be drawn. Drawing one needs a width that is either typed or a token, and §7.5's binding check cannot see which.

**The access-revoked reasons were settled, but their words were not.** The SRS closes the set at `removed_from_board`, `board_made_private` and `removed_from_workspace`, and a voluntary leave sends no frame (ADR-021 *Out of scope*). No document had written what any of them says.

**One destination did not exist.** ADR-018 §10's row 26 sends you to *"the workspace"*, and [frontend-architecture.md](../frontend-architecture.md) to `/workspaces/:wsId`. After `removed_from_workspace` you can no longer open that workspace.

---

## Decision

### 1. Four tokens, six entries

| Token | Set | Value |
|---|---|---|
| `color.toast` | `theme-light` | `#44403C` |
| `color.toast` | `theme-dark` | `#E7E5E4` |
| `color.toast-foreground` | `theme-light` | `#FFFFFF` |
| `color.toast-foreground` | `theme-dark` | `#0C0A09` |
| `spacing.toast-padding` | `base` | 12 |
| `size.toast-width` | `base` | 358 |

The catalogue goes from 163 to **169**: `base` 31 → 33, each theme set 53 → 55, the density sets unchanged.

### 2. Where each value comes from

#### Measured

- **`color.toast` and `color.toast-foreground`: an inverted toast.** It is dark in Light and light in Dark. It is the only candidate that stands clear of every surface in both themes:

  | `color.toast` on | canvas | column | card | page |
  |---|---|---|---|---|
  | Light, `#44403C` | 8.18 | 9.42 | 10.27 | 10.27 |
  | Dark, `#E7E5E4` | 15.73 | 13.93 | 12.08 | 15.73 |

  The message and the close glyph in `color.toast-foreground` measure 10.27:1 in Light and 15.73:1 in Dark. Both clear 7:1, WCAG's AAA level for body text.
- **Grey, not the text colours.** The fill is stone-700 in Light and stone-200 in Dark, not the app's own text colours (*Options*). Each is the softest shade that still reads as a message and not as something switched off. One step further is stone-600 `#57534E` in Light and stone-300 `#D6D3D1` in Dark, and each of those is that theme's `color.muted-foreground`. Neither chosen value is used by any other token in its own theme.
- **No edge.** The fill is its own edge, at 8.18:1 at worst, so the toast has no stroke and no `border.width`.
- **`shadow.overlay` stays**, as for every overlay. In Dark it barely shows on the canvas. That is fine, because there the fill does the work.
- **Written as values, not as references**, following [ADR-026](adr-026-count-chip.md): a role gets its own values. Light's `#44403C` is Dark's `color.border` and `color.chip`, and Dark's `#E7E5E4` is Light's canvas, `color.border`, `color.chip` and `color.skeleton`. The two themes never show at once, so these are coincidences across themes, not shared roles.

#### Derived

- **`size.toast-width`, 358.** That is `390 − 2 × spacing.screen-x`: the widest toast that fits the 390 phone the file draws at (the Banner's phone board) with the screen's gutters on both sides. On a narrower phone the code shrinks it to fit between the gutters (§8). So a message wraps the same way on a phone and on a desktop, and one drawing covers both.

#### Chosen

- **`spacing.toast-padding`, 12, on all four sides, in `base`.** Every other surface sets its text in by 12: the card (`spacing.card-padding-x`), a control (`spacing.control-padding-x`) and the column header, by rule. The toast does the same. It comes out 48 tall with one line at Fine, 64 with two, and 68 at Coarse. It is in `base` because the close button already grows at Coarse. This follows `spacing.banner-padding-y` ([ADR-028](adr-028-banner.md) §1).

#### By rule

- Corners `radius.surface`: ADR-017's 8px for *"cards, panels and overlays"*.
- Message `type.body`.
- `spacing.inline-gap` between the message and the close button: ADR-027's gap for neighbours in one row.
- The close glyph at `size.icon` in a `size.target-min` area, as the column header's `⋯`.

### 3. One appearance, no axis, no icon, no hue

- **One component, `Toast`, with no variant axis.** Every Phase-1 toast says one thing and offers nothing but closing it: a rejected move (4), access revoked (26), action forbidden (27), the open card deleted by someone else (32), and a request that timed out (ADR-018 §3). Phase 1 has no undo and no retry, so a toast has no action button.
- **No colour.** Rose is committed to overdue and destructive. ADR-018 §10's reasoning holds here too: a red message on a task board reads as *my work did not save*, and for 26 that would be false.
- **No icon.** The wording carries the reason, as on the Banner ([ADR-028](adr-028-banner.md) *Options*: *An icon*).
- **The inversion marks the container, not a state.** ADR-018 §10 prices every toast at zero channel cost, and this keeps it there. No other surface uses `color.toast`, so an inverted box reads as *a passing message* and as nothing else.

### 4. The close button

- **Phosphor `X`, Regular, at `size.icon`, in a `size.target-min` area, in `color.toast-foreground`, shown at rest at every pointer**, like the column's `⋯` ([ADR-027](adr-027-column-header.md) §4).
- **Why it exists:** a toast covers the bottom of the board, cards included. On touch nothing can hover to hold it, so there has to be a way to clear it.
- **Its focus ring works on the inverted fill because of the offset.** `shadow.focus-ring` paints a 2px band in `color.background` and then the ring ([ADR-024](adr-024-control-and-field-internals.md) §2). On the toast the band measures 10.27:1 in Light and 15.73:1 in Dark, and the ring measures 6.54 / 6.90 against the band. Against the toast directly, the ring alone would be 1.57 / 2.28, short of 3:1 in both themes. So the offset is what makes the ring visible here, and ADR-024's override of shadcn's offset-less ring is load-bearing on this surface.

### 5. The three access-revoked wordings

| Reason | Wording |
|---|---|
| `removed_from_board` | You were removed from “*board title*”. |
| `board_made_private` | “*board title*” is now private. Only its members can open it. |
| `removed_from_workspace` | You were removed from the workspace “*workspace name*”. |
| voluntary leave | no frame, so no toast |

- **None of them names who did it.** The frame carries `{ board_id, reason }` and no actor.
- **Each names what you lost**, because the toast is read on a different screen from the one you were on. The client already holds the board's title and the workspace's name.
- **`board_made_private` says who can still open the board.** Everyone it evicts is a plain workspace member who is not on the board. The access checker lets workspace admins and board members stay, so the people who receive this toast cannot join a private board.
- **The longest is `board_made_private`**, so the part is drawn holding it, with a title long enough to wrap to two lines.

### 6. Where losing the workspace lands you

**After `removed_from_workspace` the app goes to `/workspaces`.** For the other two reasons it goes to the workspace, as ADR-018 §10 says. This supersedes row 26's destination for that one reason and nothing else in §10. Without it, the navigation would land on a workspace you can no longer open, which is a load failure (22) and not the explanation 26 is meant to give.

### 7. The part

- **The part's board**, on page `Components`, is `Toast — states`.
  - `Doc / Canvas, 390` is bound to `color.board-canvas`, with 16px of documentation padding. It holds a `Doc / Surface / Column` bound as in [ADR-029](adr-029-skeleton-block.md) §5, with three Card instances. The main floats over them, positioned absolutely 16 from the canvas's left and bottom edges, so the Light check shows it over the third card, the column's edge and the canvas together.
  - `Doc / Page, 390` is bound to `color.background`. It holds one instance with the shortest wording, so the board shows both a one-line and a two-line toast.
  - Where the main sits is documentation geometry, not placement (*Out of scope*).
- **Build:**
  - A flex row with its items centred, filled `color.toast`, with all four corners on `radius.surface`, the shadow on `shadow.overlay` and all four paddings on `spacing.toast-padding`. Its gap is `spacing.inline-gap`, its width is `size.toast-width`, and its height follows its content.
  - `Message` uses `type.body` in `color.toast-foreground`. It fills the row and wraps.
  - `Close` is a `size.target-min` box centring an `Icon` board at `size.icon`. The board holds Phosphor's `X` path, filled `color.toast-foreground`, imported by ADR-024's recipe: 10 × 10 at Phosphor's own 3px inset.
  - That is 27 bindings in all. The main was made a component where it stood ([ADR-027](adr-027-column-header.md) *Operational note*). Its instance was made in a later call, after the main was verified ([ADR-028](adr-028-banner.md) *Operational note*).

### 8. The code

- `--toast`: `#44403C` in Light and `#E7E5E4` in Dark. `--toast-foreground`: `#FFFFFF` in Light and `#0C0A09` in Dark.
- `--toast-width: 358px`, used as `width: min(var(--toast-width), 100vw - 32px)`. The 32 is twice the screen inset.
- 12px padding, 8px between the message and the close button, 14/20 text, the overlay radius and the overlay shadow. No border.
- **If the scaffold vendors shadcn's Sonner wrapper** (current shadcn's toast), re-point its `--normal-bg` and `--normal-text` to `--toast` and `--toast-foreground`, and its `--normal-border` to `--toast`, so that no edge shows. Do not rely on Sonner's own `invert`: it inverts to Sonner's palette, not this one. Its shipped width, padding and text size also differ from the values above. **Confirm the variable names at scaffold time**, the same shape of warning as ADR-017 §9's Tailwind note.
- The close control sits at the end of the row, centred, as drawn. Its icon import is one of the `lucide-react` rewrites ADR-018 §9 lists.

---

## Options Considered

### shadcn's default: `color.popover`, `color.border` and `shadow.overlay`

Needs no token. Rejected by *Context*: 1.00:1 over a card or the page in Light, and 1.00:1 over a column in Dark, with an edge of 1.26 at best in Light.

### `color.popover` with `color.input` as the edge

It is the one existing edge that clears 3:1 on every surface: 3.82 to 4.80 in Light and 3.16 to 4.12 in Dark. Rejected: `color.input` is the boundary that means *you can type here* (ADR-017 §3.4), and a heavy grey outline round a white box reads as a field.

### `color.popover` with a stronger edge of its own

At `color.border-pressed`'s values it is 2.01 to 2.52 in Light. Rejected: it needs a token anyway, it is weaker than the inversion, and the fill still measures 1.00 over a card.

### `color.toast` at the text colours, `#1C1917` / `#FAFAF9`

This is the maximum inversion, 13.93:1 or better on every surface, and it was the first one drawn. Rejected after looking at it. `#1C1917` is Light's `color.foreground`, so the toast read as a block of ink, heavier than a passing message should be. `#FAFAF9` is Dark's, and it glared on a near-black board. The grey keeps more than twice the contrast a non-text edge needs (3:1) everywhere. One step darker in Light, stone-800 `#292524`, looks almost no different from black.

### `color.toast` as references

Rejected by §2, following ADR-026.

### Padding 16, Sonner's default

Rejected: it sets the text further in than any other surface does, and makes a one-line toast 56 tall.

### Borrow the Banner's padding (`spacing.screen-x` across, `spacing.banner-padding-y` down)

The two have the same anatomy, a message and a control in a row. Rejected: it borrows two roles, and a one-line toast would be 40 tall, cramped inside an 8px corner.

### Padding in the density sets

Rejected: the close button already grows at Coarse, as Join does in the Banner.

### Width `size.form-width`, 400

Rejected: the two would match only by coincidence, and 400 does not fit a 390 phone with its gutters.

### No width token: fill the phone, and a typed maximum on a desktop

Rejected: a typed maximum is exactly the number question 4 exists to catch.

### No close button: swipe or wait

Rejected by §4: on touch there is no hover to hold a toast, and it covers cards.

### An icon per reason, or rose for 4 and 27

Rejected by §3.

### Navigate to the workspace after `removed_from_workspace`, as ADR-018 wrote

Rejected by §6.

---

## Out of scope (deferred, recorded)

- **Where a toast sits on screen, and how two toasts stack or replace each other.** This is decided when the first screen draws a toast, at Stage 4 or 5. The part's board places the main only to show it over the board.
- **The other toasts' wordings**: a rejected move (4), action forbidden (27), the open card deleted (32), a timed-out request. They belong to inventory §7.6's Stage 5 item *Toast content*. Whatever they say, they fit this container: it wraps.
- **How long a toast stays**, and whether it waits while hovered or focused or can be swiped away. Code, settled at scaffold. WCAG 2.2.1 is the constraint to check against.
- **Announcing a toast to assistive technology.** A toast is a status message in SC 4.1.3's sense, which PRD §5.4 keeps open (ADR-018 *Out of scope*).
- **The close button's hover, pressed and focus-visible looks.** Those belong to an icon button on an inverted surface, which is not yet a part. Only its rest look is drawn.
- **Truncating a very long board title** inside a wording. Code only.
- **Motion**: ADR-018 §8's 260ms container timing. A design file cannot draw it (inventory §4.3).
- **One toast per workspace removal.** `removed_from_workspace` is sent once for each board room you were in (handoff part E). The client shows one toast however many frames arrive. Code.

## Downstream edits this ADR requires

1. **[ADR-017](adr-017-ui-foundations.md): annotation, no rewrite.** Its Status line records §9's eighth extension, and a note under §9's opening points here.
2. **[ADR-018](adr-018-state-vocabulary.md): annotation.**
   - Its Status line records that row 26's destination is superseded for `removed_from_workspace`.
   - A note under §10's table points here.
   - A note under *Out of scope*'s toast bullet says the design and the access-revoked wordings are settled, and placement is not.
3. **[ADR-021](adr-021-design-inventory-corrections.md): annotation.** A note under *Out of scope*'s *Toast content* bullet points at §5.
4. **[ADR-022](adr-022-form-control-geometry.md): annotation.** A note under *Out of scope*'s toast bullet says that drawing the toast needed a width and a padding.
5. **[frontend-architecture.md](../frontend-architecture.md):** the `ACCESS_REVOKED` handler goes to `/workspaces` for `removed_from_workspace`, and shows the toast.
6. **Penpot `Collabotask App`.** **Done 2026-11-20.**
   - The six entries were added. The catalogue reads 169.
   - The part was drawn: component `Toast`, with its main and one instance on the `Toast — states` board.
   - The Foundations board's token catalogue was regenerated: two `base` rows, and two theme rows with their swatches rebound. The heading, the section name and the two table titles now read 169, 33 and 55 each.
   - Binding check: the part's board reports only its three Card instances' `Menu → typography(13px)`, the row §7.5 expects. Foundations still reports only its 48 `Hue /` shapes and the five Card `⋯` glyphs.
   - It was switched to Dark and to Light + Coarse, and looked at each way.
     - Dark resolves the toast to `#E7E5E4` and its text and glyph to `#0C0A09`, over a `#1C1917` column, `#292524` cards and the `#0C0A09` canvas and page.
     - Coarse makes the close area 44 and both toasts 68 tall. The one-line wording wraps to two at the narrower 282 message width. Nothing overflows.
     - Restored to `theme-light` + `density-fine`.
   - **Greyed the same day**, after the part was looked at (*Options*). `color.toast` went from `#1C1917` / `#FAFAF9` to `#44403C` / `#E7E5E4`, and the two theme-table cells were rewritten. The swatch follows the token. A value change does not re-resolve bound shapes until a set is toggled, so the shapes showed the old fill until the flip. The board was then flipped and looked at again, the binding check re-run with the same result, and the file restored to `theme-light` + `density-fine`.
7. **[`design-file-inventory.md`](../design-file-inventory.md):**
   - §0 and *Related documents* gain this ADR.
   - §3's Toast row says the design and the access-revoked wordings are settled.
   - §4.2 gives the Toast its own row.
   - §7.2 records the part and the new count.
   - §7.3's table row 8 and a new *What question 4 returned* section.
   - §7.6 ticks Toast.
8. **`CLAUDE.md` › `## Now`:** a pointer. The scaffold checklist gains re-pointing Sonner's variables.
9. **`docs/build-log.md`:** no entry now, following ADR-020 to ADR-029.
10. **SRS** and **`CONTEXT.md`:** no change. The SRS fixes the frame, not where the client goes.

## Operational note

**A child made absolute inside a flex board stays where it was created.** Penpot created the toast at the page origin. Setting `layoutChild.absolute = true` took it out of the canvas's flow but left it there, at `(0, 0)` on the page, not inside the canvas. Place it with `penpotUtils.setParentXY` *after* making it absolute, and read its position back in a separate call.

**Changing a token's `value` does not repaint the shapes bound to it.** After `color.toast` was rewritten, the token read the new value but every bound shape still had the old `fillColor`. Toggling the active sets, which the Dark check does anyway, re-resolved them. So after editing a value, flip, and read the fills back in a separate call.

**`createComponent` returns the component, not the shape.** Its `id` is the library component's, so `export_shape` and `findShapeById` refuse it. Use the main shape's own id, or `component.mainInstance()`. The shape that call returns has no `isMainComponent()`.
