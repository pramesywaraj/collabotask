# ADR-025: Identity Hues as Design-File Tokens

- **Status:** Accepted
- **Date:** 2026-11-20
- **Scope:** The Penpot design file only. The 16 identity colours of [ADR-017](adr-017-ui-foundations.md) §3.3 become 32 token entries that follow the `Theme` switch; `size.card-avatar` is renamed `size.avatar`; and the Avatar part's appearances are fixed, including the no-name fallback inventory §1.5 left open. **Supersedes [ADR-020](adr-020-design-tool-penpot.md) §3's second exclusion, the identity palette, and nothing else.** The code is untouched: ADR-017 §9's rule that the palette lives in a TypeScript constant stands as written.

## Context

**Found by question 4, before the Avatar was drawn.** Stage 3 of [`design-file-inventory.md`](../design-file-inventory.md) §7.3 starts with the Avatar. Mapping its values onto the live catalogue on 2026-11-20 found no missing number: size, initials, circle and the muted colours were all covered. It found a structural problem instead.

**ADR-017 §3.3 swaps the roles of its colours by mode:**

| | Avatar fill | Avatar initials | Board rail |
|---|---|---|---|
| Light | pale | deep | deep |
| Dark | deep | pale | **pale** |

The rail must flip because every deep value fails the 3:1 boundary minimum against a dark card (1.58–2.94:1); the pale values clear it at 12.55–13.69:1. Recomputed for this ADR, all three ranges match ADR-017's.

**A typed-in colour cannot follow that swap in Penpot.** ADR-020 §3 kept the 16 values out of the catalogue, so every avatar and rail in the file would be a literal, and the `Theme` switch moves only what is a token. The consequences, had the rule stood:

- **The Avatar needs a second axis**, Light or Dark pairing: 8 hues × 2, plus two neutral looks, is 18 variants.
- **The Light → Dark check goes blind on people and boards.** On board 2.5 and most boards built from it, every Dark check would show avatars in the Light pairing and a deep rail failing 3:1: a known failure, every time, in exactly the place ADR-017 says Dark is subtle.
- **A Dark mockup means swapping by hand**: about 30 avatar instances plus the rail on board 2.5, by ADR-017 §3.3's own estimate of a dense board.
- **The binding check's expected list grows with every board**, since each avatar and rail reports `fill` and has to be named `Hue / …` to pass.

**Where the rule came from.** ADR-017 §9 says the palette *"is **not** a CSS variable set. It is a lookup indexed by `hash(id) % 8`, so it lives in a TypeScript constant with both role assignments, not in `:root`."* That is a reason about **code**. ADR-020 §3 carried it into Penpot as *"Unchanged from ADR-017 §9"* without weighing the `Theme` switch, which the code has no equivalent of.

The two options were compared side by side on 2026-11-20 with a live mock board, the same board under each, one Light/Dark switch, and the rail's contrast read out under each. The user chose tokens.

---

## Decision

### 1. Thirty-two tokens, two per hue in each theme set

| Token | Set | Light | Dark |
|---|---|---|---|
| `color.identity.<hue>.subtle` | `theme-light`, `theme-dark` | the **pale** value | the **deep** value |
| `color.identity.<hue>.strong` | `theme-light`, `theme-dark` | the **deep** value | the **pale** value |

For each of the eight hues in ADR-017 §3.3's order: terracotta, ochre, olive, sage, teal, steel, periwinkle, mauve. The values are ADR-017's, unchanged.

The catalogue goes from 121 to **153**: each theme set 35 → 51. With §3's rename, the density sets stay at 12.

### 2. Two roles, named for the surface they sit on

**`subtle` is the avatar's fill; `strong` is its initials and the board rail.** The names say how each value sits against the card behind it, which is the one thing that decides it in both modes:

- **`subtle`** is the value that is close to the card: pale on a white card, deep on a dark one. It is the avatar's fill, chosen for comfort ("a dense board shows ~30 pale dots rather than ~30 saturated ones", ADR-017 §3.3).
- **`strong`** is the value that stands out from the card: deep on white, pale on dark. The initials need it against the fill (4.65–7.96:1) and the rail needs it against the card (3:1). That is why ADR-017's table flips them **together**. They are one job, contrast against the surface, so they share one token, which passes [ADR-022](adr-022-form-control-geometry.md) §4's test from the other side: same role, same value.

The names follow the catalogue's existing `color.primary-subtle`.

### 3. `size.card-avatar` is renamed `size.avatar`

Same values, 20 / 24, in both density sets. The card avatar and the presence avatar are one size: [ADR-018](adr-018-state-vocabulary.md) §6 names both at 20px as identity markers rather than targets, and [ADR-019](adr-019-responsive-touch-scope.md) §8 lays three presence avatars out at 24px on touch, the card's own coarse value (§7). The catalogue names tokens for their job, and binding a presence avatar to a "card" token is the borrowing [ADR-023](adr-023-spacing-between-objects.md) turned down.

ADR-020 §3's mapping row says only "avatar size" and never named the token, so it needs no change.

### 4. The Avatar part: one axis, ten appearances

`Hue` = the eight hues, plus:

| Appearance | Circle | Content | For |
|---|---|---|---|
| **Muted** | `color.muted` | initials in `color.muted-foreground` | The presence cluster's *reconnecting* state, where "hues drop to flat `--muted`" (inventory §4.2). It is an Avatar appearance, so it lives in the Avatar; the cluster only arranges it |
| **Unknown** | `color.muted` | Phosphor `User` (Regular) at `size.icon`, in `color.muted-foreground` | A person whose name came back empty (inventory §1.5) |

**Unknown is a glyph, not an empty circle.** This settles inventory §1.5's *"a neutral `--muted` circle, or the Phosphor user glyph"*. A muted circle against its card is **1.09:1** in Light, almost invisible, and **1.00:1** in Dark, where `color.muted` *is* the card colour (`#292524`). An empty circle is the blank on screen §1.5 warns about. The glyph is 6.99:1 and 10.18:1. It stays 16px when the circle grows to 24 at Coarse, as the initials stay 12.

Every appearance is the same build: a board sized by `size.avatar`, an ellipse behind it sized by the same token (so no corner radius to bind), and the content centred by layout.

### 5. The Card uses the Avatar

The Card's placeholder avatar, a grey circle with a border, is replaced by an Avatar instance in all five Card variants. Card heights are unchanged: 80 at Fine and 92 at Coarse, ADR-017 §1's and ADR-019 §7's figures exactly.

### 6. One place keeps the hues typed in

**The Foundations board's identity section** shows both pairings side by side, which is its job: a reader judges the Light pairing and the Dark pairing at once. A token-bound shape shows only the active theme, so those 16 values stay typed in and its 48 `Hue / …` shapes stay on the binding check's expected list. **A `Hue /` shape anywhere else is now a defect.**

### 7. The code does not change

ADR-017 §9 stands: the app's palette is a TypeScript constant with both role assignments, indexed by `hash(id) % 8`. [ADR-020](adr-020-design-tool-penpot.md) is explicit that the design file generates no code and exists as a reference, so tokens here impose nothing there. The 16 values now live in two places, Penpot tokens and the TypeScript constant, which is true of every colour in the system (Penpot tokens and CSS variables).

---

## Options Considered

### Keep the colours typed in, with a pairing axis

ADR-020 §3 as written. Costs nothing in decisions and adds no tokens. Rejected by the user on 2026-11-20 after the side-by-side comparison, for the four consequences in *Context*: 18 variants, a blind Dark check, Dark mockups swapped by hand, and an expected-list that grows with every board.

### Name the tokens by index (`color.identity.1.subtle`)

Would mirror the code's `hash(id) % 8` literally. Rejected: ADR-017 §3.3 and the Foundations board both name the hues, and a name is what a designer recognises in Penpot's token panel. The index is ADR-017's table order, so the mapping is one line wherever it is needed.

### A third role for the rail (`color.identity.<hue>.rail`)

Would keep the rail separable from the initials. Rejected by §2: the rail equals `strong` in both modes for the same reason the initials do, so a separate token would name a difference that does not exist.

### Rename the token in place

Penpot's token `name` is writable, so this looked free. Rejected after a probe: renaming does not carry shape bindings (*Operational note*).

### An empty muted circle for the unknown person

Zero new parts. Rejected by §4's contrast figures.

---

## Out of scope (deferred, recorded)

- **The board tile's identity swatch** (inventory §1.4). Which role it takes, `subtle` or `strong`, is decided when boards 2.3 and 2.4 are drawn.
- **The rail itself.** Drawn with board 2.5; it binds `strong`.
- **The presence cluster's own geometry**: the 8px overlap ADR-019 §8's arithmetic uses, any separating ring between overlapping avatars, and the `+N` chip. Part 9 asks its own question 4.
- **The create-board colour picker** (inventory §3). Stage 5.
- **Avatar uploads.** Still Phase 2 (PRD §5.2), so there is still no photo appearance.

## Downstream edits this ADR requires

1. **[ADR-020](adr-020-design-tool-penpot.md): annotation, no rewrite.** Its Status line records that §3's identity exclusion is superseded here, for the design file only.
2. **[ADR-017](adr-017-ui-foundations.md): no edit.** §3.3's values and §9's code rule both stand, and nothing ADR-017 decides moves.
3. **Penpot `Collabotask App`.** **Done 2026-11-20.**
   - The 32 entries were added and `size.avatar` was added to both density sets.
   - All 42 shapes bound to `size.card-avatar` were rebound or replaced, then the old token was removed. The catalogue reads 153.
   - The Avatar part was drawn (`Avatar — states` on `Components`) and the Card's placeholder replaced.
   - The Foundations board's token catalogue was regenerated to 153 rows. That also adds the five ADR-024 tokens, which it had been missing since Stage 2. Its identity notes were rewritten.
   - Binding check: the Avatar part reports nothing; the Card reports only its five `⋯` glyphs; Foundations reports only its 48 `Hue /` shapes and those five glyphs.
   - Switched to Dark and to Coarse and looked at, then restored to `theme-light` + `density-fine`.
4. **[`design-file-inventory.md`](../design-file-inventory.md):**
   - §0 and *Related documents* gain this ADR.
   - §1.5's fallback "or" is settled.
   - §4.2's Avatar row and §4.3's identity paragraph are corrected.
   - §7.2 records the Avatar and the new count.
   - §7.3 records what Avatar's question 4 returned.
   - §7.5's `Hue /` row narrows to the Foundations board.
   - §7.6 ticks Avatar.
5. **`CLAUDE.md` › `## Now`:** a pointer.
6. **`docs/build-log.md`:** no entry now, following ADR-020 to ADR-024.
7. **SRS** and **`CONTEXT.md`:** no change. Presentation only.

## Operational note

**Renaming a token does not move its bindings.** Setting `token.name` succeeds, but every shape bound to it keeps the old name, which then resolves to nothing, silently. A rename is therefore: add the new token, rebind every shape (search all pages for the old name in `shape.tokens`), confirm none remain, then remove the old one. Instances follow their main and need no rebinding.

**`children` returns fresh wrapper objects on every read**, so `parent.children.indexOf(child)` is −1 even for a real child, and `insertChild(-1, …)` put the new shape **first**. In a flex row, `appendChild` places a shape last visually.

**Removing a layer inside a component's main deletes it.** The Card's placeholder ellipse was gone, not hidden; the hide-instead-of-remove behaviour applies to copies.

**Penpot's server sits behind Cloudflare, whose filter can refuse an `execute_code` payload** with a "Sorry, you have been blocked" page. It happened once, on a read whose code contained `|| "-"`, which looks like SQL string concatenation. The plugin and the file were unaffected, and the same read without that pattern ran. Keep quoted strings away from `||` in code. Also keep angle-bracketed words such as `<hue>` out of text written into the file, since saving it passes the same filter.
