# ADR-024: Control and Field Internals

- **Status:** Accepted
- **Date:** 2026-11-20
- **Scope:** The inside of one field and one control in the app zone: the gap between a field's label, box and message; the size of an icon inside a control and its gap from the label; how the focus ring's 2px offset is drawn; what fills an input's box; and what *pressed* looks like on a filled button. **Extends [ADR-017](adr-017-ui-foundations.md) §9; supersedes nothing.** It closes the focus-ring item [ADR-023](adr-023-spacing-between-objects.md) *Out of scope* deferred to this stage, and gives [ADR-018](adr-018-state-vocabulary.md) §6's *pressed* (15) a value for a filled button, where §6 gives values only for an edge. No code is written here.

## Context

**Found before drawing, like the two before it.** Stage 2 of [`design-file-inventory.md`](../design-file-inventory.md) §7.3 is the proof run: two parts (Input, Button) and one screen (board 2.1, Login) built from them. Its question 4, *"did you have to type a number or a colour because no shared value existed?"*, was asked of every value those three need on 2026-11-20, against the live catalogue of 116 tokens. It fired for the third stage running.

**Why these were missing.** [ADR-022](adr-022-form-control-geometry.md) gave a control its outside: how tall it is, how much room its text has, how far apart two fields sit, how wide a form is. [ADR-023](adr-023-spacing-between-objects.md) gave the space between objects. Neither needed the inside of one field (a label above a box, a message below it) or the inside of one control (an icon beside a label). Login is the first thing in the file to need both.

What the catalogue already covered is recorded so a reader can see the gap was narrow. Every colour each state needs was present: rest, hover, busy and not-satisfied fills and labels; `color.input` for the box edge; `color.destructive` and `color.destructive-text` for invalid (ADR-018's ledger, state 36). So were the control's height, padding and radius, all five type steps, the ring's colour, and the page inset, section break and form width Login sits in.

What was missing:

| Needed by | Value | Why nothing covered it |
|---|---|---|
| Every field | The gap between label and box, and between box and message | `spacing.field-gap` is *between* fields |
| Every focusable thing | The focus ring's 2px offset (ADR-018 §6) | Penpot strokes have an alignment but no offset; ADR-023 *Out of scope* expected this stage to meet it |
| Button, busy (17) | The spinner's size | No icon size anywhere in the catalogue |
| Button, busy (17) | The gap between spinner and label | Nothing inside a control was ever two things side by side |
| Button, pressed (15) | A fill for pressed | ADR-018 §6's pressed ladder is three stone steps on an **edge**, written for cards. A filled plum button has `color.primary-hover` and nothing past it |

**Why this is an ADR and not a note** is ADR-022's reason, unchanged: every form and every control will reference these values, and changing one after twenty components do is the sweep ADR-017's Context warns about.

---

## Decision

### 1. Four tokens, five entries

| Token | Penpot type | Set | Fine | Coarse |
|---|---|---|---|---|
| `spacing.field-inner-gap` | `spacing` | `base` | **8** | 8 |
| `spacing.icon-gap` | `spacing` | `base` | **8** | 8 |
| `size.icon` | `sizing` | `base` | **16** | 16 |
| `shadow.focus-ring` | `shadow` | `theme-light` and `theme-dark` | two layers, below | same |

`shadow.focus-ring`, written identically into both theme sets:

| Layer | offset X / Y | blur | spread | colour |
|---|---|---|---|---|
| 1 (on top) | 0 / 0 | 0 | **2** | `{color.background}` |
| 2 | 0 / 0 | 0 | **4** | `{color.ring}` |

The catalogue goes from 116 tokens to **121**: `base` 24 → 27, each theme set 34 → 35, the density sets unchanged.

**And three decisions that need no token:** pressed on a filled button reuses `color.primary-hover` (§3); an input's box is filled `color.background` (§4); and four mappings onto existing tokens (§5).

### 2. Where each value comes from

Derived values and chosen ones are separated, as in ADR-022 and ADR-023, so a later reader knows which are safe to re-open.

#### Derived

**2.1 — `size.icon` 16.** The inventory quotes Phosphor at 16px for every icon it names inside a control: the busy spinner, and the due-date picker's calendar (§1.6, citing ADR-018 §9: *"Phosphor calendar, Regular, 16px"*). It is also the size shadcn's Button gives any icon inside it (`[&_svg]:size-4`). It goes in `base` because it does not flip: at Coarse the control grows from 36 to 44 around an icon that stays 16, just as the 14/20 label does.

It does **not** cover the Card's `⋯`. That is a 13px text glyph (ADR-018 §6), not an icon, and stays the known literal §7.5 already expects.

**2.2 — `shadow.focus-ring`: the values.** ADR-018 §6 fixes them: *"a 2px ring at `--ring` with a 2px offset, on every focusable thing, never removed."* A 2px offset band, then a 2px ring outside it, is spreads of 2 and 4. The offset band is painted in `color.background`, which is what Tailwind's `ring-offset` does too: it is a solid band in the page colour, not a transparent gap.

#### Chosen

**2.3 — `shadow.focus-ring`: the mechanism.** How the ring is drawn is a choice, made because Penpot cannot draw it as ADR-023 §2.5 implied (a 2px stroke). Its strokes have no offset property. Two stacked shadows can draw it, and they have three advantages over every alternative (*Options Considered*):

- **It is how the code renders it.** Tailwind's `ring-2 ring-offset-2` compiles to two box-shadows, the inner in the offset colour and the outer in the ring colour. The file draws what the scaffold will ship.
- **It never moves anything.** A shadow sits outside the shape's bounds, so focus arriving on a field shifts nothing in the form around it.
- **It is checked.** One token, bound to `shadow`, which the §7.5 binding check already watches. Both colours are *references*, so the ring follows Light/Dark without a second hex value: verified as `#FFFFFF` / `#6B4E9E` at Light and `#0C0A09` / `#A78BD9` at Dark.

**The cost, recorded:** ADR-023 §2.5 names `border.width-strong` (2) as the ring's width. The ring's 2px now lives inside this token as the difference between its two spreads, and `border.width-strong` keeps the drop placeholder alone. If the ring's width ever changes, it changes here. The spreads are literals inside the token, the way `shadow.overlay`'s offsets and blur already are; the token, not its parts, is what gets bound and checked.

**2.4 — `spacing.field-inner-gap` 8.** One value for both gaps inside a field: label to box, and box to message. The constraint: it must be smaller than `spacing.field-gap` at **both** densities (16 / 20), or a label stops reading as belonging to its own box and starts reading as belonging to the box above it. On the 4px grid every spacing value sits on, 4, 8 and 12 qualify; 8 is shadcn's own (`FormItem` is `grid gap-2`). It goes in `base`: it is space inside one target, since a label focuses its box, not space between two targets.

**2.5 — `spacing.icon-gap` 8.** shadcn's Button is `gap-2`. A choice.

**It equals `spacing.field-inner-gap` and stays a separate token.** [ADR-022](adr-022-form-control-geometry.md) §4's test applies: the two match arithmetically, not semantically. One is vertical, inside a field, and constrained by the field gap; the other is horizontal, inside a control, and constrained by nothing yet. If one moved, the other should not follow.

### 3. Pressed on a filled button reuses `color.primary-hover`

**Decided by the user, 2026-11-20.** The plum button's *pressed* (15) uses the same fill as its *hover* (12). No new colour.

- **On touch, where pressed matters most.** ADR-018 §6 says pressed is the only feedback a finger receives. There is no hover on touch, so the darker plum appears only on press and is unambiguous.
- **On a mouse,** pressing looks the same as hovering. A submit then turns busy at once (17), which is its own feedback.
- **Two identical variants are not new in this file.** The Card's InFlight and DragSource variants are already pixel-identical by design. The Button keeps Pressed as its own variant so the state stays named and the code has a hook if it ever diverges.

This applies to **filled** buttons. An outline or ghost button, when one is drawn, takes ADR-018 §6's edge ladder (`color.border-hover`, `color.border-pressed`) as written.

### 4. An input's box is filled `color.background`

**The box is not transparent.** This was found by drawing, not by asking, and it is load-bearing.

- **In Penpot:** a board with no fill casts its shadow from its *children*. A transparent input box with `shadow.focus-ring` drew the ring around the typed letters, not around the box. Any shape carrying the ring must have a fill.
- **In code:** the 2px-ring-at-2px-offset idiom ADR-018 chose is the one shadcn's **Tailwind v3** Input shipped, and that Input was `bg-background`. **Current** shadcn, the format ADR-017 §9 is written in, ships `bg-transparent` with a 3px ring and no offset. So the scaffold overrides the vendored Input's focus classes regardless (see *Downstream edits*), and this fill goes in the same override.

On Login, where the page is `color.background`, the fill is invisible. Inside a surface of another colour (the card drawer in Dark, `color.popover` `#1C1917`), the box reads as a slightly darker well (`#0C0A09`). That is the v3 look the ring comes from.

### 5. Mappings onto existing tokens

No new value. Recorded because each is a choice a later screen will copy.

| What | Token | Why |
|---|---|---|
| A field's label | `type.body`, `color.foreground` | Regular weight. Controls stay at 400; `type.card-title` exists precisely so they do not turn medium |
| An error message | `type.meta`, `color.destructive-text` | No ADR assigns helper text a step. 12/16 at `#BE123C` on white is above 4.5:1. Register's password-minimum note uses the same step in `color.muted-foreground` |
| The invalid box edge | `color.destructive` at `border.width` | ADR-018's ledger, state 36 |
| Inline title editing (37) | `type.title` in a `size.control-height` box | A 28px line fits: 36 − 28 leaves 4 above and below at Fine, 44 − 28 leaves 8 at Coarse, so ADR-022 §3's "height and a centred line, no vertical padding token" holds |

**No space is reserved for an error message.** It appears below the box, and the form grows by 24px at Fine (8 gap + a 16px line). Reserving the slot would leave 24px of permanent space under every field for an error most people never see. It would also need a typed height in Penpot, since a hidden child takes no room in a flex layout. The cost: the button moves down when an error appears.

---

## Options Considered

### Draw the ring as a separate rectangle around each part

An absolutely positioned outline 4px outside the box, carrying `border.width-strong` and `color.ring`. Rejected: it needs a typed −4 position, a +8 size and a 10px corner (6 + 2 + 2), none of which is a token, and it would be one more layer to keep aligned in every variant.

### Draw the ring as an outer-aligned stroke on the box itself

Binds `border.width-strong` directly, with no new token. Rejected: it draws the ring touching the box, which drops ADR-018 §6's offset. A stroke can be inner, centre or outer, never apart from the shape.

### Express the spreads as references (`{border.width-strong}`)

Would keep ADR-023 §2.5's mapping literally true. Not taken. Penpot's shadow token accepts a reference in `spread`, but documents it as a reference to a *number* token, and `border.width-strong` is a `borderWidth` token. Whether that resolves is unverified. The literal is consistent with how every shadow token in the file already holds its geometry.

### A third plum step, `color.primary-pressed`

One palette step past hover per theme, mirroring §6's "one palette step per state" on edges. It would distinguish press from hover on a mouse too. Rejected by the user in favour of §3: it costs two new hex values from ADR-017 §3.1's plum scale, each needing a contrast measurement, for a difference only a mouse ever shows, at the moment a submit turns busy anyway.

### Keep the input transparent, as current shadcn does

Rejected by §4: the ring would draw around the text in the design file, and in code the offset idiom already departs from current shadcn.

### One shared 8px token for the field's inner gap and the icon gap

Fewer names. Rejected by ADR-022 §4's test, as in §2.5.

### Reserve space for the error message

Rejected in §5.

---

## Out of scope (deferred, recorded)

- **Button kinds other than primary.** Login has one button and it is primary, so the Button has one axis, `State`. Secondary, ghost and destructive become a second axis when a screen needs one; §3 already says which pressed rule each takes.
- **Whether the inline title editor also shows the ring.** State 37 is drawn as *"the `--input` boundary doing the editing job"*, without the ring. Browsers match `:focus-visible` on any focus of a text input, so in code it would show both. Board 2.6 settles which.
- **Where a wrong password is reported on Login.** A failed sign-in is not about one field, and neither inventory §1.1 nor the two parts give it a home. Board 2.1 is drawn at first arrival and does not need it yet; a later pass on 2.1 or Stage 4's Register does.
- **The spinner's motion.** Motion stays code-only ([ADR-020](adr-020-design-tool-penpot.md) §3; inventory §4.3). The file draws Phosphor's `CircleNotch` at rest.
- **The Astro landing page** and **Phase-2 deferrals.** Untouched.

## Downstream edits this ADR requires

1. **[ADR-017](adr-017-ui-foundations.md): annotation, no rewrite.** Its Status line gains `§9 extended a third time by ADR-024`, and §9's head note gains a pointer. No decision text changes.
2. **[ADR-018](adr-018-state-vocabulary.md): annotation, no rewrite.** Its Status line records that §6's pressed has a filled-button value and its ring a construction, both here.
3. **[ADR-020](adr-020-design-tool-penpot.md): annotation.** Its Status line's note on ADR-023's `borderWidth` row says that row covers "the 2px ring"; it gains a pointer saying the ring is now a shadow.
4. **[ADR-023](adr-023-spacing-between-objects.md): annotation.** Its Status line records that its *Out of scope* focus-ring item is closed here, and that `border.width-strong` now carries the placeholder alone.
5. **Penpot `Collabotask App`: add the five entries.** **Done 2026-11-20.** The ring was proven first on a throwaway probe board, then removed: both layers rendered in the right order, the binding survived the switch to Dark, and both colours followed it. The catalogue reads 121. The Input and Button parts and board 2.1 were then drawn on these values; the binding check on 2.1 returns nothing, and all four Theme × Density combinations were looked at. Active sets restored to `theme-light` + `density-fine`.
6. **[`design-file-inventory.md`](../design-file-inventory.md):** §0's source table and *Related documents* gain this ADR. §7.2 records the two parts and board 2.1. §7.3 records what Stage 2's question 4 returned and that it closed here. §7.5 gains the shadow-needs-a-fill caveat, and §7.6 ticks Stage 2.
7. **Scaffold (no edit now; recorded here, as ADR-022 did).** These values reach Tailwind's config alongside [ADR-018](adr-018-state-vocabulary.md) §*Downstream edits* item 6. One consequence is new: **the vendored Input and Button need their focus classes overridden** from current shadcn's `focus-visible:ring-[3px] ring-ring/50` to `ring-2 ring-ring ring-offset-2 ring-offset-background`, and the Input from `bg-transparent` to `bg-background`. Do this in the same pass as ADR-018's `disabled:` override, before any other component is added.
8. **`CLAUDE.md` › `## Now`:** a pointer.
9. **`docs/build-log.md`:** no entry now, following ADR-020 to ADR-023.
10. **SRS** and **`CONTEXT.md`:** no change. Presentation only.

## Operational note

Four Penpot behaviours met while drawing Stage 2. None is in ADR-020, ADR-022 or ADR-023.

**A board with no fill casts its shadow from its children.** §4's reason. The §7.5 binding check cannot see this: an unfilled box with `shadow.focus-ring` bound passes the check and still draws the ring around its text. Look at it.

**`createShapeFromSvg` ignores `width` and `height` and imports at the viewBox's scale**, which for Phosphor is 256. It also wraps the path in a group with a transparent `base-background` rectangle. To get a 16px icon, move the path into a board, set the path's constraints to `scale` on both axes, resize the board to 16, then bind `size.icon` to the board, not the path. The glyph then keeps Phosphor's own inset (13 × 12.5 inside 16).

**`resize()` resets a flex board's sizing to `fix`.** Set `horizontalSizing` / `verticalSizing` to `auto` *after* the resize, not before, or the board stops growing with its content and its height is silently wrong.

**In this file all four themes read inactive.** The sets are switched on directly, which Penpot treats as a custom theme. So a flip means toggling the sets, not the themes; restore by making `base`, `theme-light` and `density-fine` the active sets.
