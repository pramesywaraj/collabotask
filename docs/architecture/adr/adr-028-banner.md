# ADR-028: The Banner

- **Status:** Accepted
- **Date:** 2026-11-20
- **Scope:** The read-only banner's surface and geometry: one new token, and five reuses each bound by a rule. Also its one appearance, its wording, and the Banner part's build. **Extends [ADR-017](adr-017-ui-foundations.md) §9 and supersedes nothing.** [ADR-018](adr-018-state-vocabulary.md) §10 and [ADR-019](adr-019-responsive-touch-scope.md) (*What ADR-018 inherits*, item 3) fixed that the banner is persistent and cannot be dismissed. [ADR-021](adr-021-design-inventory-corrections.md) §6 fixed that it carries Join. None of them said what it is drawn on or how big it is. It also **corrects** [`design-file-inventory.md`](../design-file-inventory.md) §4.2, which listed an appearance nobody can see. The code gains no CSS variable.

## Context

**Found by question 4, before the banner was drawn.** Part 6 of Stage 3 in [`design-file-inventory.md`](../design-file-inventory.md) §7.3 is the Banner. Mapping its values onto the live catalogue on 2026-11-20 gave this:

| Value | Needs | Live token | |
|---|---|---|---|
| Fill | the surface Join's focus ring is drawn for | `color.background` | present |
| Bottom rule | 1px, bottom edge only | `color.border`, `border.width` | present |
| Side padding | the screen's side inset | `spacing.screen-x` | present |
| Message | 14/20 (ADR-017 §6) | `type.body` | present |
| Message colour | text on the banner | `color.foreground` | present |
| Gap between the message and Join | offered by [ADR-027](adr-027-column-header.md) §2 | `spacing.inline-gap` | present |
| Join | the primary button | `Button` instance | present |
| **Space above and below the contents** | — | none | **missing** |

The banner holds a control 36 tall at Fine and 44 at Coarse (`size.control-height`), and nothing said how much room surrounds it.

Inventory §4.2 also listed two appearances, *informational* and *with an action*. That list predates ADR-021 §6, so it was checked against who can actually see a banner (§4).

Three more things had never been decided, and none of them is a token: whether the banner has an icon, where Join sits in the row, and what the banner says.

---

## Decision

### 1. One token, one entry

| Token | Set | Value |
|---|---|---|
| `spacing.banner-padding-y` | `base` | 8 |

The catalogue goes from 158 to **159**: `base` 30 → 31, the theme and density sets unchanged.

### 2. Where each value comes from

#### Derived

- **Fill `color.background`, for two reasons.** First, Join's focus ring ([ADR-024](adr-024-control-and-field-internals.md)) is two stacked shadows, and its 2px offset is drawn in `color.background`. On any other surface the offset shows as a band of the wrong colour around the button. Second, ADR-017 §10 keeps `--background` white *"so every shadcn component works unmodified on a white surface"*, and Join is the vendored shadcn Button. This is a requirement, not a coincidence of values.
- **Rule `color.border` and `border.width`, bound to the bottom edge only.** In Dark, `color.background` and `color.board-canvas` are both `#0C0A09`, so without the rule the banner has no lower edge. The rule is `#44403C` there, about 1.92:1, which is the border ladder's own contrast and what every column header draws. In Light the rule is the canvas's colour and merges into it, which is harmless: the fill changes from white to grey there. The edge is not the banner's signal. The message and Join are. Built in the order [ADR-023](adr-023-spacing-between-objects.md)'s *Operational note* gives.
- **Side padding `spacing.screen-x`.** ADR-023 §2 makes it the horizontal padding of the board canvas, the app bar and any page body. The banner spans the screen, so its message starts where the first column panel starts. One role, not two numbers that match ([ADR-022](adr-022-form-control-geometry.md) §4).
- **Message `type.body` in `color.foreground`.** ADR-017 §6's default step. Foreground rather than muted, because on touch the message is the only signal (ADR-019 item 3).
- **Join is a `Button` instance, primary, in `Rest`.** ADR-017 §3.1 confines brand to the primary action. On a read-only board every other mutating control is absent (ADR-018 §10), so Join is the only primary action on screen, and the banner gets the board's one plum shape without spending a hue. Its label is an override; nothing about the button is redrawn.
- **No height token.** The height follows the contents: Join's `size.control-height` plus the padding, so 52 at Fine and 60 at Coarse. When the message wraps to two lines on a phone it is 40 tall, which still fits beside Coarse's 44 button, so the banner stays 60.

#### Accepted

- **`spacing.inline-gap` 8 between the message and Join.** [ADR-027](adr-027-column-header.md) §2 offered it to the banner by name, for neighbours in one row that are not an icon and its own label. A message and a button fit that exactly. The message fills the row and Join sits at the end, so 8 is the minimum gap, as between the column header's `Lead` and `Menu`.

#### Chosen

- **`spacing.banner-padding-y` 8.** It keeps the plum button 8px off the strip's top and bottom edges, and it keeps to the 4px grid and to the catalogue's other inner steps.
- **In `base`, not the density sets.** Join already grows from 36 to 44 at Coarse. Growing the padding too would cost another 8px of board at Coarse, which means on phones, where height is scarcest.
- **Named for the banner.** No other part in §7.3 holds a control in a full-width strip. The app bar comes closest, but both card-count derivations treat it as a fixed 56px height (ADR-017 §1, ADR-019 §7), not as padding around its contents, so a general name would not describe it.

### 3. Padding, not a height

A fixed `size.banner-height` with centred contents is how the column header and the controls avoid a second number (ADR-022 §3). It does not work here. The message wraps on a narrow screen, so a fixed height clips it. A minimum height would not clip, but Penpot does not draw a minimum size on a board that sizes itself to its content ([ADR-026](adr-026-count-chip.md)'s *Operational note*). Padding is the only form that the file and the code both draw the same way.

### 4. The part: one appearance, no axis

- **Nobody would see an informational banner.** [ADR-021](adr-021-design-inventory-corrections.md) Context §4 reads it off the access checker: state 24 has exactly one population, a plain workspace member on a workspace-visible board who has not joined, and `SelfJoinBoard` lets every one of them join. None of the other states that say "you can't" is a banner. 25 is a whole-screen gate. 26 and 27 are toasts. 28 to 30 are shown on the presence cluster. 23 and 39 are whole screens. The inventory and ADR-017 to ADR-021 were searched for "banner", and every mention is state 24. A banner without Join is the one ADR-021 rejected (*Join on the board tile only, leaving the read-only banner informational*), so drawing it would draw a rejected design.
- **One component, `Banner`, with no variant axis**, like the Count chip and the Column header. Join's own looks (hover, pressed, focus-visible, and busy as *Joining…* per ADR-018 §5) belong to the Button. A failed join is a toast (27), and the banner stays.
- **The part's board**, on page `Components`, is `Banner — states`. The main sits in `Doc / Screen, 1024`, where the message fits on one line. An instance sits in `Doc / Screen, 390`, where it wraps to two. Under each is a `Doc / Canvas` strip bound to `color.board-canvas`, so the Dark flip shows that the rule is the banner's only edge there. The screens' widths and the strip's height are documentation geometry (ADR-023 *Out of scope*).
- **Build:** a row with centred items. The fill and the rule are as in §2. `spacing.screen-x` is bound to the left and right padding, `spacing.banner-padding-y` to the top and bottom, and `spacing.inline-gap` to the gap. The height follows the content. Inside are `Message` (fills the width, grows downward) and `Join` (a `Button` instance, `State=Rest`).
- **Wording:** *"You’re viewing this board. Join it to add, edit and move cards."* and a **Join board** button. The first sentence says why pressing does nothing, which is the banner's job on touch (ADR-019 item 3). The second reads as an invitation, not a refusal (ADR-021 §6).
- **No icon.** The message and the plum Join already carry the signal, and no ADR has chosen a glyph. An icon would be the only element in the banner that says nothing the text does not.

### 5. What it costs the board

The banner takes 52px at Fine and 60px at Coarse from the column body of board 2.7. On ADR-019 §7's phone that leaves 528px of column body instead of 588. The card counts in ADR-017 §1 and ADR-019 §7 subtract no banner because they describe board 2.5 at rest, so they stand. The *"6 cards visible"* check belongs to board 2.15, also at rest, and is unaffected.

### 6. The code

No new variable. The banner is full width, on `--background`, with `border-bottom: 1px solid var(--border)`, 16px side padding and 8px top and bottom padding (`px-4 py-2`). Its contents are a row, centred vertically, 8px apart. The message takes the remaining width and wraps. Join is the vendored Button in its primary kind; it does not shrink, and it sits at the row's end. The banner cannot be dismissed (ADR-018 §10).

---

## Options Considered

### A plum-tinted banner (`color.primary-subtle`)

The usual look for a notice. Rejected twice. ADR-017 §3.1 says brand never carries state, and read-only is a state. And Join's busy fill is `color.primary-subtle` (ADR-018 §5), so the button would vanish into its own banner while the join was in flight.

### A grey banner (`color.muted` or `color.board-column`)

Sets the banner apart from the white app bar. Rejected by §2: Join's focus ring would show its `color.background` offset as a white band on grey. In Dark, `color.muted` holds `color.card`'s value, so the banner would read as a card-coloured slab.

### A banner colour of its own (`color.banner`)

The precedent is [ADR-026](adr-026-count-chip.md) §3, where the chip got its own role. Rejected: any value other than `color.background` breaks the focus ring's offset, so the token could only ever hold a copy of it. A role with no value of its own adds nothing.

### No vertical padding

Needs no token, as with the column header. Rejected: the plum button would touch both edges of the strip. The column header gets away with it only because its `⋯` has no fill.

### A fixed height (`size.banner-height`)

Rejected by §3.

### Borrow `spacing.card-padding-y` or `spacing.column-padding`

Both are 8 at Fine. Rejected by ADR-022 §4: the same number, a different job. Card padding is also 12 at Coarse, which would make the banner 68.

### Padding in the density sets (8 / 12)

Rejected by §2: Join already grows at Coarse.

### Keep the informational appearance

Follows §4.2 to the letter. Rejected by §4.

### An icon (Phosphor `Eye` or `Info`)

Every value it needs exists (`size.icon`, `spacing.icon-gap`, `color.muted-foreground`), so this was a choice, not a gap. Rejected by §4.

### Join right after the message

Keeps the action close to the sentence on a wide screen. Rejected: on a narrow screen the message must fill the row in order to wrap, and one layout serves both widths only with Join at the end, the same structure as the column header.

---

## Out of scope (deferred, recorded)

- **Where the banner sits on board 2.7**: under the app bar, and above or below the 4px board rail. ADR-023 §2.4 already flags that the rail's placement costs height. Board 2.7 decides both.
- **Its accessible role and where it sits in the focus order.** Code, settled at scaffold.
- **Board 2.8's join gate.** It is not a banner; it is a whole screen with one Join (ADR-021 §6). Stage 4, group d.
- **What replaces the banner after a successful join.** The controls return and the banner goes, which is board 2.5. That is the user's own action, so it needs no movement (ADR-018 §8).

## Downstream edits this ADR requires

1. **[ADR-017](adr-017-ui-foundations.md): annotation, no rewrite.** Its Status line records §9's sixth extension, and a note under §9's opening points here.
2. **[ADR-021](adr-021-design-inventory-corrections.md): annotation.** A note under §6 says the banner part has one appearance, and that inventory §4.2's informational one is withdrawn (§4).
3. **[ADR-027](adr-027-column-header.md): annotation.** A note under §2 *Chosen* says the banner took the offered gap. The board tile has yet to ask.
4. **Penpot `Collabotask App`.** **Done 2026-11-20.**
   - `spacing.banner-padding-y` added. The catalogue reads 159.
   - The part was drawn: component `Banner`, its main on the `Banner — states` board with one instance beside it.
   - The Foundations board's token catalogue was regenerated: one row added, and the heading, the section name and the `base` table title now read 159 and 31.
   - Binding check: the part's board reports nothing, both Join instances included. Foundations still reports only its 48 `Hue /` shapes and the five Card `⋯` glyphs.
   - Switched to Dark and to Light + Coarse, and looked at. Dark resolves the banner to `#0C0A09` (the canvas's own value), the rule to `#44403C`, the message to `#FAFAF9` and Join to `#A78BD9`. Coarse makes both banners 60 tall and Join 44, and leaves the padding at 8 and the message at 14/20. Restored to `theme-light` + `density-fine`. **The rule was confirmed from the shape's data and the generated CSS** (`border-block-end-width: 1px`), not from an export ([ADR-027](adr-027-column-header.md)'s *Operational note*).
5. **[`design-file-inventory.md`](../design-file-inventory.md):**
   - §0 and *Related documents* gain this ADR.
   - §4.2's Banner row says there is one appearance, and why.
   - §7.2 records the part and the new count.
   - §7.3 records what the banner's question 4 returned.
   - §7.6 ticks Banner.
6. **`CLAUDE.md` › `## Now`:** a pointer.
7. **`docs/build-log.md`:** no entry now, following ADR-020 to ADR-027.
8. **SRS** and **`CONTEXT.md`:** no change. Presentation only.

## Operational note

**An instance made in the same call as its main's token bindings can miss one.** Penpot applies tokens asynchronously. The `Banner` main and its first instance were made in one `execute_code` call. The main ended up with all eight bindings, and the instance with seven: it lacked `columnGap`. On screen the gap was 8 either way, so only the binding check caught it. `resetOverrides()` did not bring the binding back, and it reset the instance's width to the main's 1024 although its sizing still read `fill`. A fresh instance made in a later call carried all eight. So: **bind the main, verify it in a separate call, and only then make its instances.**
