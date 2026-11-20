# ADR-031: The Presence Cluster

- **Status:** Accepted
- **Date:** 2026-11-20
- **Scope:** The presence cluster in the board's app bar. It covers five new tokens in nine entries, the round seat that holds `+N`, which faces the cluster shows and in what order, where *Reconnecting…* goes, the part's three appearances, and its build. It **extends [ADR-017](adr-017-ui-foundations.md) §9**: the code gains `--avatar-neutral`, `--avatar-overlap` and `--seat-padding-x`. It **supersedes three things and nothing else in them**:
  - [ADR-018](adr-018-state-vocabulary.md) §10's *"flat `--muted` fill"* for the reconnecting faces, which become `--avatar-neutral`.
  - [ADR-025](adr-025-identity-hues-as-tokens.md) §4's circle colour for the Muted and Unknown appearances, which also becomes `--avatar-neutral`.
  - [ADR-019](adr-019-responsive-touch-scope.md) §8's `+N` overflow chip and the arithmetic built on it. `+N` is a round seat in the stack. §8's cap of three faces below 640px stands, and from 640px up the cap is five.
- It also **corrects [ADR-026](adr-026-count-chip.md) §5's premise** that `+N` is its own 32px control. [ADR-018](adr-018-state-vocabulary.md) §6, written later, makes the whole cluster the one target.

## Context

Part 9 of the design file's Stage 3, and the last part. Inventory §7.3 places it last because it is a cluster of Avatars. Question 4 was asked before drawing, and it fired a tenth time in ten.

The faces themselves were in the catalogue: `size.avatar`, the initials, the Avatar's Muted look, `spacing.inline-gap` and `color.muted-foreground` for the label. Three things were not usable.

- **How far the faces overlap.** [ADR-019](adr-019-responsive-touch-scope.md) §8 gives 8px, worked out on a phone with 24px faces. Nothing said what a 20px desktop face does. Penpot also refuses a negative flex gap (*Value not valid: -8*), so an overlap has to be drawn another way (*Operational note*).
- **Anything between overlapping circles.** The worst-matched neighbours, ochre, sage and teal, measure 1.00–1.01 apart in Light and 1.03 in Dark. Two grey faces are identical, 1.00. Without something between them, overlapping circles merge, and the initials behind are cut by an edge nobody can see.
- **The grey circle could not be seen.** The Muted face sat on `color.muted`, which is **1.09:1** against the white app bar. The faintest live circle is 1.11. In Dark `color.muted` is 1.30, far fainter than the live circles' 2.05–3.83, so the cluster would fade when it reconnects. The Unknown face uses the same circle, and [ADR-014](adr-014-presence-frame-enrichment.md)'s best-effort profile lookup means a presence user can arrive with an empty name, so Unknown can appear in the cluster too.

Four choices no ADR had made came with them: what `+N` is, where *Reconnecting…* goes on a phone, how many faces show from 640px up, and whether you see yourself.

Every surface the cluster sits on is `color.background`. The app bar is *"the white app bar"* ([ADR-028](adr-028-banner.md) §*Options*), at every width. The overflow list the cluster opens is Stage 5 and draws its faces without overlap.

The options were compared on a page that rendered each one in Light and Dark with the catalogue's real values. The user took the recommendations, except for `+N`, where they chose the round seat. Question 4 was then asked again of the seat, and the user took its three recommendations.

## Decision

### 1. Five tokens, nine entries

| Token | Set | Value | Kind |
|---|---|---|---|
| `spacing.avatar-overlap` | `density-fine` · `density-coarse` | −6 · −8 | derived |
| `shadow.avatar-ring` | `theme-light` · `theme-dark` | `0 0 0 2 {color.background}` in each | chosen width, colour by rule |
| `color.avatar-neutral` | `theme-light` · `theme-dark` | `#E7E5E4` · `#57534E` | measured |
| `radius.full` | `base` | 9999 | by rule |
| `spacing.seat-padding-x` | `density-fine` · `density-coarse` | 2 · 4 | measured |

The catalogue goes from 169 to **178**: `base` 34, each theme set 57, each density set 15.

### 2. Where each value comes from

#### Measured

- **`color.avatar-neutral`.** The threshold is the faintest live identity circle against the bar, in the same theme: 1.11 in Light and 2.05 in Dark. Going grey then changes the colour and keeps the weight.
  - Light: `#F5F5F4` is 1.09 and fails. `#E7E5E4` is 1.26, the softest stone step that clears.
  - Dark: `#292524` is 1.30 and `#44403C` is 1.92, and both fail. `#57534E` is 2.59.
  - Initials and the Unknown glyph on it: 6.08 in Light and 5.12 in Dark. Both pass AA.
  - Written as values, not as references. `color.muted` itself cannot move, because in Light it is also the column's fill.
  - Side effect on the card: an Unknown face goes from 1.09 to 1.26 on a white card and from 1.00 to 1.99 on a Dark one.
- **`spacing.seat-padding-x`** is `(size.avatar − 16) ÷ 2`, where 16 is the width of `+5` at `type.meta`, measured in Penpot. A one-digit count then makes a circle exactly the size of a face: measured at 20×20 and 24×24. `+12` grows to 23 and 27, and `+128` to 31 and 35.

#### Derived

- **`spacing.avatar-overlap`.** −8 at Coarse is ADR-019 §8's number. At Fine, −6 keeps the same share of each covered face visible: `(avatar − overlap − ring) ÷ avatar` is 12 of 20 (60%) at Fine and 14 of 24 (58%) at Coarse. −8 at Fine would leave 10 of 20.
- The value is stored negative, because the file draws it as a margin and the code ships a negative margin.

#### Chosen

- **The ring is 2px wide.** It is Tailwind's `ring-2`, and the same 2px as the focus ring's offset ([ADR-024](adr-024-control-and-field-internals.md) §2).

#### By rule

- **The ring's colour is `{color.background}`**, the colour of the bar the cluster sits on. Its only job is to separate neighbours, so it needs no contrast of its own.
- **`radius.full` is 9999**, the code's `rounded-full`. Penpot clamps a radius to half the shape's height, so it draws a clean pill at any width (checked: a 24px seat draws with a 12px radius).
- **The seat's fill is `color.avatar-neutral`.** It does the grey faces' job, a neutral place in the avatar row. That is 1.26 / 2.59 on the bar, with `+N` on it at 6.08 / 5.12. While reconnecting, the whole stack is one grey.
- **The cluster's height is `size.target-min`** with its contents centred. That is ADR-018 §6's *"2px of vertical padding to reach 24px"* at Fine, and 10px at Coarse.
- **The gap before the label is `spacing.inline-gap`** ([ADR-027](adr-027-column-header.md) §2). The label is `type.meta` in `color.muted-foreground`: 7.63 / 13.26 on the bar.
- **The focused box is filled `color.background`, with corners on `radius.control`, a control's corners, and `shadow.focus-ring`.** The fill is there because an unfilled board casts its shadow from its children ([ADR-024](adr-024-control-and-field-internals.md) *Operational note*): the ring would circle each face, not the box.

### 3. The whole cluster is the button, and `+N` is a round seat

- [ADR-018](adr-018-state-vocabulary.md) §6 makes the cluster one target and says its circles do nothing on their own. So `+N` is a label inside the target, not a control. ADR-026 §5's worry, that a 32px `+N` falls under Coarse's 44px floor, does not arise: the target is the whole cluster, `size.target-min` tall.
- **`+N` is a round seat at a face's size**, the user's choice over a Count chip instance. The seat sits in the stack as the last face does: it overlaps by `spacing.avatar-overlap`, wears `shadow.avatar-ring`, is `size.avatar` tall with `size.avatar` as its minimum width, and takes `radius.full`. Its text is `type.meta` in `color.muted-foreground`.
- **What it costs on a phone.** ADR-019 §8 budgeted 32 for the chip plus an 8px gap. The seat adds only its visible part: 16px for `+5` and 19px for `+12`. On a 375px phone the cluster is 72–75px wide, against §8's 96, so the board name gets about 180px, against 159.
- The cluster opens the presence overflow list: a full-height sheet below 640px, a dialog or drawer above it ([ADR-021](adr-021-design-inventory-corrections.md) §5). The list is Stage 5.

### 4. Which faces, in what order

- **You are drawn, first.** The server adds you to the board's room before it sends `ACTIVE_USERS` (SRS §4.5), so the list already includes you. A person alone on a board sees one face. When the connection drops, *Reconnecting…* always has a face beside it.
- **At most three faces below 640px** (ADR-019 §8, unchanged), **and at most five from 640px up**, then the seat, which counts everyone not shown. Nobody hidden, no seat. At 640px on a Coarse tablet the board name still keeps about 410px.
- A face whose name came back empty shows Unknown, the Avatar's own appearance.

### 5. Reconnecting

- The faces switch to the Avatar's Muted look, now on `color.avatar-neutral`. *Reconnecting…*, ADR-018 §10's own wording, follows after `spacing.inline-gap`.
- **Below 640px the label takes the seat's place.** The label is 90px wide. After the seat, the phone's board name would keep about 80px. In the seat's place it keeps about 100px, for as long as the connection is down. The hidden count is as stale as the faces.
- From 640px up the label follows the seat.

### 6. One axis, three appearances

`State`: **Live**, **Reconnecting** and **FocusVisible**.

- Every appearance has the same box: filled `color.background`, `radius.control` on its four corners, `size.target-min` tall. On the bar the fill cannot be seen. FocusVisible adds `shadow.focus-ring`.
- The focused box has no side padding. Its ring sits 2px from the circles at the sides, and 4px above and below at Fine, 12px at Coarse.
- **No hover and no pressed look.** Both step an edge (ADR-018 §6), and the cluster has none. A focused grey cluster is not drawn as its own variant; it is FocusVisible's ring on Reconnecting's contents.

### 7. The part

- **The part's board**, on page `Components`, is `Presence cluster — states`.
  - `Label / Part` and the variant container `Presence cluster`, holding the three mains.
  - `Doc / Canvas, 375`, bound to `color.board-canvas`, with two `Doc / App bar, 375` strips. One holds a Live instance with Face 4 and Face 5 hidden and the seat reading `+5`. The other holds a Reconnecting instance with Face 4, Face 5 and the seat hidden.
  - `Doc / Canvas, 1024`, with two `Doc / App bar, 1024` strips: Live and Reconnecting, five faces and `+3`.
  - Each strip is filled `color.background`, with a bottom-only rule in `color.border` on `border.width` (`strokeWidthBottom`). Its width, height and padding are documentation geometry.
- **Build:**
  - The root is a flex row with its items centred. Its height binds `size.target-min`, its gap `spacing.inline-gap`, its fill `color.background`, and its four corners `radius.control`. FocusVisible's root also binds `shadow.focus-ring`. It does not clip.
  - `Faces` is a flex row with no fill and no gap, sized by its content. It does not clip, so the first face's ring shows.
  - `Face 1` to `Face 5` are Avatar instances: AR sage, MK periwinkle, JT ochre, SL mauve and DN teal, each switched to Muted in Reconnecting. Each binds `shadow.avatar-ring`. Faces 2 to 5 and the seat bind `spacing.avatar-overlap` to their left margin.
  - The `Seat` is a flex row with its contents centred. It binds `size.avatar` to height and minimum width, `spacing.seat-padding-x` to both side paddings, `color.avatar-neutral` to its fill, `radius.full` to its four corners, and `shadow.avatar-ring`. Its `Count` text binds `type.meta` and `color.muted-foreground`.
  - `Label` (Reconnecting only) binds `type.meta` and `color.muted-foreground`.
- **The Avatar part changes:** its Muted and Unknown circles are rebound from `color.muted` to `color.avatar-neutral`. Every instance follows, including the Card's.
- Each main was made a component where it stood, and the three were then grouped as variants. The instances were made in a later call, after the mains were verified ([ADR-028](adr-028-banner.md) *Operational note*).

### 8. The code

- `--avatar-neutral`: `#E7E5E4` in Light and `#57534E` in Dark. It is the Muted and Unknown faces' circle, and the seat's fill.
- `--avatar-overlap`: −6px at a fine pointer and −8px at a coarse one, in the density scales beside `--avatar`. It is a negative left margin on every face after the first and on the seat.
- `--seat-padding-x`: 2px fine and 4px coarse. The seat is `rounded-full`, with a height and minimum width of the avatar size, 12/16 text in `--muted-foreground`, and `tabular-nums`.
- The ring is `ring-2 ring-background` on every face and on the seat. It is a box-shadow, so it takes no layout room.
- The cluster is one `button` that opens the overflow list. It has a minimum height of the target size, items centred, an 8px gap, a `--background` fill and 6px corners. Its focus ring takes the same override ADR-024 gives Input and Button.
- At most 3 faces below 640px and 5 from 640px up. You come first. Below 640px, while reconnecting, the label replaces the seat.

---

## Options Considered

### The overlap at −8 at both densities

One entry in `base`, and ADR-019's only number. Rejected by §2: on a desktop each covered face would show 10 of 20px, half of it, against 14 of 24 on a phone.

### No ring

Nothing new in the catalogue. Rejected by the Context's measurements: neighbouring circles at 1.00–1.03 merge, and two grey faces are identical.

### Keep `color.muted` for the grey circle

No new token. Rejected: 1.09:1 on the white bar, under the faintest live circle. In a stack, the circle in front cuts off the initials behind at an edge nobody can see.

### `#44403C` in Dark

Softer than `#57534E`, with 6.90 for the initials. Rejected narrowly: 1.92 against the bar, just under the faintest live circle at 2.05.

### Change Light only, and keep `#292524` in Dark

The smallest change. Rejected: at 1.30 the Dark cluster fades when it reconnects, against live circles of 2.05–3.83.

### `+N` as a Count chip instance

Recommended at first: no new shape, and every value already bound. The user chose the round seat, which matches the faces, sits in the stack and costs less width on a phone. The chip is 20px tall with 4px corners beside 24px circles at Coarse.

### Seat padding 6 at both densities

The Count chip's spacing, as one entry in `base`. Rejected: `+5` becomes 28 wide at both sizes, so the seat is always a pill and never the round face chosen.

### Seat fill `color.chip`

The same in Light. In Dark it is one step darker than the grey faces, 1.92 on the bar. Rejected: while reconnecting, the seat would read as a different object from the faces beside it.

### A radius token at half the avatar, 10 / 12

Exact rather than clamped. Rejected: two entries, and it only fits shapes that are `size.avatar` tall. `radius.full` fits any pill, and it is the value the code ships.

### The label after the seat at every width

One rule. Rejected by §5: on a phone the board name would keep about 80px while reconnecting.

### Three faces at every width

One rule, and the smallest cluster. Rejected: a desktop would show as few faces as a phone, and presence carries the whole weight of *other people are here* ([ADR-017](adr-017-ui-foundations.md) *Context*).

### Leave yourself out

The common pattern in document editors. Rejected by §4: alone on a board, the cluster is empty, and *Reconnecting…* has nothing to sit beside.

### Two axes, Connection × Focus

Four variants, so a focused grey cluster is drawn too. Rejected: one more variant, and a second axis on a part that needs only one. FocusVisible's ring is the same on either contents.

---

## Out of scope (deferred, recorded)

- **The presence overflow list** the cluster opens. Stage 5, inventory §3. Its rows draw an Avatar beside a name, with no overlap and no ring.
- **The app bar itself**: the board name, the gap between the name and the cluster, the board `⋯` and the add-column action. Boards 2.5 and 2.15 draw it.
- **The order of the faces after yours.** Code. Join order is the natural reading of `USER_JOINED`, but nothing here decides it.
- **Names.** The cluster shows initials only. Names appear in the overflow list, and no hover tooltip is drawn or decided.
- **Motion when someone arrives or leaves** (state 33). Code only, like all motion (inventory §4.3).
- **Announcing presence changes to assistive technology.** SC 4.1.3, kept open by PRD §5.4 (ADR-018 *Out of scope*).
- **The Foundations board's radius ladder.** It still shows ADR-017's three steps. `radius.full` is in the catalogue table only.
- **A `+1` seat in the file.** Penpot ignores a minimum width on a board sized by its content ([ADR-026](adr-026-count-chip.md) *Operational note*), so a `+1` seat draws 17 wide at Fine. The app draws it 20.

## Downstream edits this ADR requires

1. **[ADR-017](adr-017-ui-foundations.md): annotation, no rewrite.** Its Status line records §9's ninth extension, and a note under §9's opening points here.
2. **[ADR-018](adr-018-state-vocabulary.md): annotation.** Its Status line records that §10's reconnecting fill is superseded, and a note under §10's connection-health paragraph points here.
3. **[ADR-019](adr-019-responsive-touch-scope.md): annotation.** Its Status line records that §8's overflow chip and arithmetic are superseded, and a note under §8 gives the new widths and the cap from 640px up.
4. **[ADR-025](adr-025-identity-hues-as-tokens.md): annotation.** Its Status line records the Muted and Unknown circle colour. Notes go under §4's table and under *Out of scope*'s cluster-geometry bullet.
5. **[ADR-026](adr-026-count-chip.md): annotation.** Notes go under §5's `+N` bullet and under *Out of scope*'s `+N` bullet.
6. **Penpot `Collabotask App`.** **Done 2026-11-20.**
   - The nine entries were added. The catalogue reads 178.
   - The Avatar's Muted and Unknown circles were rebound to `color.avatar-neutral`. No shape anywhere still binds a `Circle` to `color.muted`.
   - The part was drawn: component `Presence cluster`, with `State` = Live, Reconnecting and FocusVisible, and four instances on the `Presence cluster — states` board.
   - The Foundations board's catalogue was regenerated: one `base` row, two theme rows with their swatches rebound, and two density rows. The heading, the section name and the three table titles now read 178, 34, 57 each and 15 each.
   - Binding check: the part's board returns nothing. A file-wide check of non-zero layout margins also returns nothing. Foundations still reports only its 48 `Hue /` shapes and the five Card `⋯` glyphs. Re-run with margins on every board on `Components`, it reports only the Card's `Menu → typography(13px)`: five on its own board and three in the Toast board's cards.
   - It was switched to Dark and to Light + Coarse, and looked at each way.
     - Dark resolves the bar and the rings to `#0C0A09`, the grey faces and the seat to `#57534E` with `#D6D3D1` text, periwinkle to `#394174` with `#E7E9F5` initials, and the focus ring to `#A78BD9`.
     - Coarse makes the faces 24, steps them by 16, makes the `+3` seat exactly 24×24, and makes the cluster 44 tall. The phone's reconnecting cluster is 154 wide.
     - Restored to `theme-light` + `density-fine`.
   - **One slip, repaired.** A name lookup for `Label / Phone` searched every page and found the Banner board's label first, and overwrote its text. Its original wording had not been recorded. It was rewritten as *Phone, 390 — two lines*, after its sibling *Wide, 1024 — one line*, with its bindings intact.
7. **[`design-file-inventory.md`](../design-file-inventory.md):**
   - *Related documents* gains this ADR.
   - §4.2's Presence cluster row gives the three appearances, and its Avatar row names the new circle.
   - §7.2 records the part and the new count.
   - §7.3's table row 9 points here, and a new *What question 4 returned* section is added.
   - §7.5's binding check gains layout margins.
   - §7.6 ticks the Presence cluster and closes Stage 3.
   - The Appendix's rows 29 and 33 point here.
8. **`CLAUDE.md` › `## Now`:** a pointer. Stage 3 is closed, and Stage 4 is next.
9. **`docs/build-log.md`:** no entry, following ADR-020 to ADR-030.
10. **SRS**, **`CONTEXT.md`** and **[frontend-architecture.md](../frontend-architecture.md):** no change. The SRS fixes the frames, and the architecture stores presence as a list without saying how it is drawn.

## Operational note

**Penpot refuses a negative flex gap.** Setting `columnGap = -8` fails with *Value not valid: -8*. A negative **margin** works, and a spacing token can hold a negative value and bind to `marginLeft`.

**Bind a margin only after writing it.** Binding the token alone stores the margin and emits `margin-inline-start` in the generated CSS, but the canvas does not move the shape. Write the margin directly first (`layoutChild.leftMargin = -6`), then bind, and the canvas follows. Writing any other side afterwards also works. Either way the binding survives, and a density flip re-resolves it: −6 became −8 on the canvas.

**§7.5's binding check could not see margins.** Every earlier part had none. The check now looks at the four `layoutChild` margins too.

**A hidden child is left out of a flex layout**, so an instance shows fewer faces by hiding the rest: `hidden = true` on Face 4 and Face 5, and on the seat.

**`penpot.library.local.components` lists a variant group once**, under its first variant. A lookup of the other variants by id there comes back empty. Reach them through the container's `variants.variantComponents()`.

**Shapes are found by name across every page, and boards reuse names.** `penpotUtils.findShape` with a name searches the whole file, and `Label / Phone` exists on two boards. Find a shape by its id, or search within the part's own board.
