# ADR-017: UI Foundations — Tokens, Density, Type and Elevation

- **Status:** Accepted — **§8 superseded by [ADR-019](adr-019-responsive-touch-scope.md)** (2026-11-17); the *Motion* reservation re-pointed and the *archive* icon struck by [ADR-018](adr-018-state-vocabulary.md) (2026-11-18); **§9 extended by [ADR-022](adr-022-form-control-geometry.md)** (2026-11-20) — form-control geometry, which §9 did not cover; **§9 extended again by [ADR-023](adr-023-spacing-between-objects.md)** (2026-11-20) — spacing between objects, the board rail's height and the two border widths, which §9 did not cover either; **§9 extended a third time by [ADR-024](adr-024-control-and-field-internals.md)** (2026-11-20) — the inside of a field and a control: the label gap, the icon's size and gap, the focus ring drawn as a shadow, the input's fill, and pressed on a filled button; **§9 extended a fourth time by [ADR-026](adr-026-count-chip.md)** (2026-11-20) — the count chip's fill, height and side padding: §3.4 measured the chip, but §9 never gave it a variable. Everything else stands as written.
- **Date:** 2026-11-13
- **Scope:** The visual baseline for the **app zone** (React + Vite SPA — [ADR-011](adr-011-frontend-framework.md)), consumed through shadcn/ui + Tailwind. Covers the token vocabulary and values, the density target, board structure, the type scale, the radius and elevation policy, colour modes, and — as first written — the responsive scope (see *Status*). **Supersedes** `docs/architecture/design-tokens.md` (an unadopted, uncommitted proposal dated 2026-11-12 — see *Corrections* below). This ADR decides the **baseline only**: it deliberately does **not** assign a visual channel to any interaction state. §*Channel budget* names the channels left unspent, which [ADR-018](adr-018-state-vocabulary.md) will spend. The public zone (Astro landing — [ADR-013](adr-013-public-zone-framework.md)) may diverge; a landing page can carry more colour than a dense board.

## Context

No frontend code exists. The architecture is settled (ADR-011/012/013, `../frontend-architecture.md`) but nothing had been decided about how the product *looks*, and the design work is happening in Figma before any scaffolding. Two things make that ordering load-bearing:

1. **The board is the product, and it is dense.** The backend has no pagination anywhere — `GetKanban` returns every card in every column ([column_response.go:21](../../../collabotask-backend/internal/delivery/http/response/column_response.go)), so a column is unbounded and the resting board is the screen a person stares at all day. Every token gets multiplied by card count.
2. **Components bake tokens in.** shadcn/ui is vendored into the repo, not installed — once twenty components reference `--primary`, changing its *role* (not just its value) is a sweep. This is a durable, hard-to-reverse decision, which is why it is an ADR rather than a token file.

Two constraints from the backend shaped the colour decisions directly:

- **A card has no status field.** `CardResponse` carries title, description, position, assignee, due date and timestamps — nothing else ([card_response.go:10-21](../../../collabotask-backend/internal/delivery/http/response/card_response.go)). Status is expressed by *which column a card is in*. So a vivid per-card colour would carry no information.
- **No board event carries the actor.** `CardMoved` sends card/from/to/position; `CardUpdated` sends card id and changed fields ([broadcaster.go:112-126](../../../collabotask-backend/internal/realtime/broadcast/broadcaster.go)). You cannot render "Bob moved this card." Presence therefore carries the entire weight of *other people are here*, which is why §*Identity palette* and the always-visible presence cluster are treated as foundations rather than decoration.

Every ratio quoted below is a computed WCAG 2.1 contrast ratio against the stated surface. AA requires **4.5:1** for body text and **3:1** for large text and for boundaries that identify a UI component (1.4.11).

---

## Decision

### 1. Density target

**A 2-line title clamp, with 8 cards visible at 1440×900 and a 9th peeking.**

| | |
|---|---|
| Card title | 14px / 20px line-height, clamped to 2 lines |
| Card padding | 8px vertical, 12px horizontal |
| Card avatar | 20px |
| Meta row | **Conditional** — rendered only when the card has an assignee or a due date |
| Card gap | 8px |
| Column width | 280px |

Verified against the locked values, not estimated:

| Measurement | Value | Derivation |
|---|---|---|
| Eight representative cards, stacked | **612px** | `8 + lines×20 + (24 if meta) + 8` per card, plus 8px gaps |
| Visible column body at 1440×900 | **704px** | 900 − 88 browser chrome − 56 app bar − 16 padding − 36 column header |
| Fully visible cards | **8** | target met, with 92px spare |
| Ninth card | ~1.1 rows | peeks below the fold, which is the desired cue that the column continues |

Card titles accept up to 500 characters (`max=500` — [card_request.go:10](../../../collabotask-backend/internal/delivery/http/request/card_request.go)), so **the clamp is mandatory, not cosmetic.** An unclamped title can be ~36 lines tall.

### 2. Board structure — contained columns

**Each column is a visible panel**: its own `--board-column` background, 8px radius, 8px padding, and a header carrying a bottom rule. Structure is *drawn*; the board reads as a row of containers.

Chosen by rendering three directions side by side in both themes rather than by argument. The two rejected directions and the measurements that killed them are in *Options Considered*.

Two consequences of this choice are accepted explicitly:

- **The panel→card step is faint.** `--board-column` → `--card` is **1.09:1**, and `--board-canvas` → `--board-column` is **1.15:1**. Three greys stack up and none of the steps is strong. This is the cost of drawn structure; the card's 1px border does the rest of the work.
- **The column header needs a rule.** Because columns are unbounded, cards scroll *under* the header. The header shares the panel's fill and is separated by `border-bottom: 1px solid var(--border)`. There is deliberately **no** `--column-header` colour token — the superseded draft defined one identical to `--column` (1.00:1), which is not a separable surface.

Direction A also leaves the column panel available as a tintable surface. That is a channel ADR-018 inherits (see §*Channel budget*) and the flatter directions would not have offered.

### 3. Colour

The resting board is **near-monochrome**: grey canvas, lighter grey column panels, white bordered shadowless cards, near-black text, muted initials circles, one thin coloured board rail, one plum button. Colour is an event, not a texture.

#### 3.1 Primary — plum

**`#6B4E9E`**, at 6.54:1 on white.

That single ratio is why it was chosen over the alternatives: it clears 4.5:1 as a filled button with white text, clears 4.5:1 as link text on white, **and** clears the 3:1 boundary minimum as a focus ring on white. One value, three jobs, so `--ring` is not a separate hue — it is `--primary`.

Hue separation from everything that carries meaning: **136°** from amber, **102°** from emerald, **85°** from rose. At badge size none of them blurs into the brand.

**Brand is confined to the primary action and to "where you are"** — the primary button, the focus ring, the active nav item, inline links. **Brand never carries state.** A plum card means nothing.

#### 3.2 Semantics — fill values are not text values

| Meaning | Allowed on | Light | Dark |
|---|---|---|---|
| Deadline approaching | due-soon date text **only** | `#B45309` 5.02:1 | `#FBBF24` 9.09:1 |
| Overdue / destructive | overdue date, destructive buttons and confirmations | `#BE123C` 6.29:1 | `#FB7185` 5.64:1 |
| Done | done / archived state | `#047857` 5.48:1 | `#34D399` 7.89:1 |

**The 500-level fill is never used as text.** Amber-500 `#F59E0B` as text is **2.15:1** and emerald-500 `#10B981` is **2.54:1** — both fail outright. The superseded draft shipped `--success: #10B981` unaudited while auditing `--success-text` beside it. Each semantic therefore gets one *text* value, which is the only form Phase 1 uses: dates and states are coloured text, not filled pills, because filled pills would reintroduce the colour volume §3 rules out.

#### 3.3 Identity palette — one set, two role assignments

Eight hues, assigned deterministically by user id (`hash(id) % 8`) so a person keeps the same colour across boards and sessions. **The same eight serve both avatars and board identity rails.**

| # | Hue | Deep | Pale |
|---|---|---|---|
| 1 | terracotta | `#744539` | `#F5EAE7` |
| 2 | ochre | `#746239` | `#F5F1E7` |
| 3 | olive | `#627439` | `#F1F5E7` |
| 4 | sage | `#39744F` | `#E7F5ED` |
| 5 | teal | `#397472` | `#E7F5F5` |
| 6 | steel | `#395974` | `#E7EFF5` |
| 7 | periwinkle | `#394174` | `#E7E9F5` |
| 8 | mauve | `#74395F` | `#F5E7F0` |

**The roles swap by mode:**

| | Avatar fill | Avatar ink | Board rail |
|---|---|---|---|
| **Light** | pale | deep | deep |
| **Dark** | deep | pale | **pale** |

The dark rail must use the pale value. Every deep value fails the 3:1 boundary minimum against a dark card — **1.58:1 to 2.94:1** across the eight. The pale values clear it at 12.55:1 to 13.69:1. All eight deep/pale pairs clear 4.5:1 as initials (4.65:1 to 7.96:1), in both role assignments.

Two notes:

- **Pale fill with deep ink, not the reverse.** An earlier mid-saturation construction with white initials measured 3.86:1 to 4.45:1 and failed all eight. The inversion also serves eye comfort directly: a dense board shows ~30 pale dots rather than ~30 saturated ones.
- **Every avatar is initials-only in Phase 1.** Avatar uploads are a Phase-2 deferral (PRD §5.2), so `AssignedToResponse.AvatarURL` is always null and there is no image case to design.

`board.background_color` is **demoted from a canvas colour to an identity accent** — the board-list tile swatch and a 4px rail at the top of the board. The canvas itself is always the neutral `--board-canvas`. The field validates only `min=4,max=8` and is **not a hex validator** ([board_request.go:8](../../../collabotask-backend/internal/delivery/http/request/board_request.go)) — the string `"pink"` passes — so the frontend constrains the picker to these eight and **must defend on read**, falling back to the default when the stored value is not one of them.

#### 3.4 Neutrals — warm stone, and the muted-text correction

**Tailwind `stone`.** Measured against `zinc` and `neutral`, all three are functionally identical — every ratio lands within 0.25:1, and stone's canvas `RGB(231,229,228)` differs from zinc's `RGB(228,228,231)` by 3 units per channel, below the just-noticeable difference for a flat field. The original justification for warm neutrals was an amber primary, which is gone. Stone is retained on comfort grounds alone, and the choice is recorded as near-arbitrary rather than derived.

**`--muted-foreground` is stone-600 `#57534E`, not stone-500.** This correction is a direct consequence of §2: once a grey panel sits behind muted text, stone-500 `#78716C` fails.

| Surface | stone-500 | stone-600 |
|---|---|---|
| `--background` white | 4.80:1 | 7.63:1 |
| `--board-column` `#F5F5F4` | **4.40:1 fail** | 6.99:1 |
| `--board-canvas` / count chip `#E7E5E4` | **3.82:1 fail** | 6.08:1 |

Dark mode has the mirror problem: stone-400 `#A8A29E` passes on the canvas, column and card but drops to **4.07:1** on a stone-700 count chip. Dark `--muted-foreground` is therefore **stone-300 `#D6D3D1`** (6.90:1 worst case).

> **The count chip measured here has no variable in §9.** Since **[ADR-026](adr-026-count-chip.md)** (2026-11-20) it is `--chip`, with these two values unchanged: `#E7E5E4` and stone-700 `#44403C`. It has its own token rather than `--border`, which holds both values only by coincidence.

**`--input` is stone-500 `#78716C`** — a deliberate deviation. The superseded draft inherited shadcn's stone-300 `#D6D3D1` at 1.49:1 and flagged it as a known 1.4.11 gap without resolving it. Stone-500 is the *lightest* stone that reaches 3:1 on a white field (4.80:1), and it also clears it on both dark surfaces (3.16:1 on card, 3.65:1 on column) — so one value serves both modes. It looks more drawn than most shipped systems; login and register are the first screens a new user sees, and an invisible field boundary is a worse first impression than a slightly heavy one.

### 4. Elevation — three channels, one meaning each

| Channel | Means | Used on |
|---|---|---|
| **Surface brightness** | permanent hierarchy | canvas → column panel → card. Three steps, lighter as the layer comes forward — in **both** modes; dark is a compressed version of the same ladder, not an inverted one |
| **Shadow** | **temporarily lifted above the page — nothing else** | the dragged card's `DragOverlay` clone, menus, popovers, the card drawer, modals |
| **Border** | an edge | card edge, column header rule, input boundary |

**A resting card has no shadow.** White fill, 1px `--border`, flat. The superseded draft gave resting cards both a lightness step *and* a shadow, which spends paint at unbounded density and burns the one channel the dragged card most needs. Two shadow levels exist, both reserved:

```css
--shadow-overlay: 0 4px 12px -2px rgb(28 25 23 / 0.10), 0 2px 4px -2px rgb(28 25 23 / 0.08);
--shadow-dragged: 0 12px 24px -6px rgb(28 25 23 / 0.22), 0 4px 8px -4px rgb(28 25 23 / 0.14);
```

Shadows are near-invisible on dark surfaces, so **the surface steps must not be flattened in dark mode** — they are the only thing separating a card from its column there.

### 5. Radius

Three steps, by object size class. `--radius: 0.5rem` is the shadcn base; the ladder is:

| Radius | Applies to |
|---|---|
| **4px** | badges, chips, count pills |
| **6px** | buttons, inputs, menu items |
| **8px** | cards, column panels, menus, popovers, the drawer, modals |

### 6. Type

**Poppins everywhere.** One family, three weights (400 / 500 / 600), **self-hosted** — ADR-011 commits to a strict `script-src 'self'` CSP, and self-hosting the font keeps the font-file origin under the same policy rather than adding a third-party host.

Five steps. The default is **14px**:

| Step | Use |
|---|---|
| 12 / 16 | meta — due dates, counts, timestamps, labels |
| **14 / 20** | **default** — card titles, body, column headers, buttons, inputs |
| 16 / 24 | card description in the open-card drawer |
| 20 / 28 | board name, open-card title |
| 28 / 36 | login and register headings |

`font-variant-numeric: tabular-nums` on **dates and counts**, so digits occupy equal width and a column of dates does not jitter as values change.

### 7. Modes — dark-ready, not dark-shipped

**Both values are authored; no toggle ships in Phase 1.**

- Every token in §9 carries a light and a dark value in CSS from the first commit.
- Figma uses a **2-mode collection** — a mode is a column in a variable collection, not a duplicate file or a duplicate frame.
- **No theme switch in the Phase 1 UI.** Nothing reads the dark values at runtime yet.
- ADR-018 maintains **light frames only** and flips the mode to spot-check. It does not draw every state twice.

This is what it costs, stated plainly: **30 of the 31 tokens in §9 carry a second value** (every one except `--radius`), and three of them are not simple darkenings — `--primary-foreground` inverts, the identity palette swaps roles, and `--muted-foreground` moves two steps rather than one. Shipping dark later is then a toggle plus an audit, not a re-derivation.

### 8. Responsive scope — pointer-only, 1024px and up

> **Superseded by [ADR-019](adr-019-responsive-touch-scope.md) (2026-11-17): phones and tablets are in Phase 1 and card dragging works on touch.** This section is kept as written because ADR-019 quotes it and takes its arguments apart one at a time — two of the three survive and are paid for in ADR-019 §7. Nothing below is in force.

**The app zone targets pointer input at ≥1024px. One density scale, not two.**

A Kanban board on a phone is a **different UI, not a narrower one**: horizontal column scrolling plus vertical card scrolling plus drag-to-reorder is a three-axis gesture problem that a reflowed layout does not solve. dnd-kit's `TouchSensor` needs long-press activation constraints and a different hit-target budget, which would force a second density scale and a second set of frames — and §1's whole budget assumes a 20px avatar and a 28px header that no touch guideline permits.

PRD §5.3 states "not a native mobile app in Phase 1 (responsive web is the target)" — a non-goal about native apps, which is not the same as a breakpoint commitment. PRD §5.2 now records it as a Phase 2 commitment (see *Downstream edits*).

**Below 1024px the app is neither blocked nor optimised.** It renders: columns are cramped, the board scrolls horizontally, and a press-and-hold drag competes with scroll. This is an accepted Phase-1 gap, not a supported configuration — no notice, no fallback layout, no touch sensors tuned for it. Login and register reflow fine, being simple forms.

**A phone interface is a committed Phase 2 effort**, designed as its own thing rather than a reflow of this one. The likely shape is a "Move card to…" picker in place of drag: on a 375px screen only one 280px column is visible, so a drag must cross columns the user cannot see — which no amount of tuning fixes.

### 9. Token reference

> **§9 fixes what a control looks like, not how big it is.** Control height, horizontal padding, the gap between stacked fields and the form container width are in **[ADR-022](adr-022-form-control-geometry.md)** (2026-11-20), which *extends* this section. Nothing below is changed or withdrawn by it.
>
> **§9 also says nothing about the space *between* objects**: the screen's side inset, the gap between board columns, the canvas's vertical padding, and section breaks. It also gives no token for the 4px rail (§3.3) or for the 1px / 2px border widths. Those are in **[ADR-023](adr-023-spacing-between-objects.md)** (2026-11-20), which extends this section the same way.
>
> **Nor does §9 say anything about the inside of one field or one control**: the gap between a label, its box and its message, an icon's size and its gap from a label, how the focus ring's offset is drawn, what fills an input, or what *pressed* looks like on a filled button. Those are in **[ADR-024](adr-024-control-and-field-internals.md)** (2026-11-20), which extends this section the same way.
>
> **Nor does §9 give the count chip a variable**, though §3.4 chose `--muted-foreground` by measuring against it. Its fill (`--chip`, the two values §3.4 names), its height and its side padding are in **[ADR-026](adr-026-count-chip.md)** (2026-11-20), which extends this section the same way.

shadcn/ui vocabulary, Tailwind v4 / current shadcn format — full colour values, consumed as `var(--primary)`.

> **If the project scaffolds on Tailwind v3**, shadcn stores bare HSL triplets and reads them as `hsl(var(--primary))`. These hex values will silently break that setup. Confirm the Tailwind major version at scaffold time and convert if needed.

```css
:root {
  --radius: 0.5rem;

  /* surfaces — shadcn vocabulary, unchanged */
  --background: #FFFFFF;
  --foreground: #1C1917;                 /* 17.49:1 */
  --card: #FFFFFF;
  --card-foreground: #1C1917;
  --popover: #FFFFFF;
  --popover-foreground: #1C1917;

  /* board surfaces — the two additions */
  --board-canvas: #E7E5E4;
  --board-column: #F5F5F4;               /* same value as --muted, different role */

  /* primary — plum */
  --primary: #6B4E9E;                    /* 6.54:1 on white: fill, ring and link text */
  --primary-foreground: #FFFFFF;         /* 6.54:1 */
  --primary-hover: #5B4287;              /* 8.16:1 with white */
  --primary-subtle: #F4F0FA;
  --primary-subtle-foreground: #573F80;  /* 7.70:1 on the subtle fill */
  --ring: #6B4E9E;                       /* = --primary; no separate ring value */

  /* neutral utility — shadcn semantics, kept grey */
  --secondary: #F5F5F4;
  --secondary-foreground: #1C1917;
  --muted: #F5F5F4;
  --muted-foreground: #57534E;           /* stone-600 — see §3.4 */
  --accent: #F5F5F4;
  --accent-foreground: #1C1917;
  --border: #E7E5E4;                     /* decorative edge */
  --input: #78716C;                      /* 4.80:1 — a boundary, not decoration */

  /* semantics — text values; the 500-level fills are never used as text */
  --destructive: #E11D48;
  --destructive-foreground: #FFFFFF;     /* 4.70:1 */
  --destructive-text: #BE123C;           /* 6.29:1 */
  --due-soon: #B45309;                   /* 5.02:1 — amber means deadlines, only */
  --overdue: #BE123C;                    /* 6.29:1 */
  --done: #047857;                       /* 5.48:1 */

  /* elevation — shadow means "temporarily lifted", nothing else */
  --shadow-overlay: 0 4px 12px -2px rgb(28 25 23 / 0.10), 0 2px 4px -2px rgb(28 25 23 / 0.08);
  --shadow-dragged: 0 12px 24px -6px rgb(28 25 23 / 0.22), 0 4px 8px -4px rgb(28 25 23 / 0.14);
}

.dark {
  --background: #0C0A09;
  --foreground: #FAFAF9;
  --card: #292524;
  --card-foreground: #FAFAF9;            /* 14.52:1 */
  --popover: #1C1917;
  --popover-foreground: #FAFAF9;

  --board-canvas: #0C0A09;
  --board-column: #1C1917;

  --primary: #A78BD9;                    /* 6.11:1 on column, 5.30:1 on card */
  --primary-foreground: #1C1917;         /* 6.11:1 — INVERTS from white */
  --primary-hover: #BBA3E3;              /* 7.88:1 with the dark ink */
  --primary-subtle: #3B2F57;
  --primary-subtle-foreground: #D6C7EE;  /* 7.67:1 on the subtle fill */
  --ring: #A78BD9;                       /* 5.30:1 on card */

  --secondary: #292524;
  --secondary-foreground: #FAFAF9;
  --muted: #292524;
  --muted-foreground: #D6D3D1;           /* stone-300 — see §3.4 */
  --accent: #292524;
  --accent-foreground: #FAFAF9;
  --border: #44403C;
  --input: #78716C;                      /* same value as light mode */

  --destructive: #FB7185;
  --destructive-foreground: #1C1917;     /* 6.50:1 */
  --destructive-text: #FB7185;           /* 5.64:1 on card */
  --due-soon: #FBBF24;                   /* 9.09:1 */
  --overdue: #FB7185;                    /* 5.64:1 */
  --done: #34D399;                       /* 7.89:1 */

  --shadow-overlay: 0 4px 12px -2px rgb(0 0 0 / 0.50);
  --shadow-dragged: 0 12px 24px -6px rgb(0 0 0 / 0.60);
}
```

The identity palette (§3.3) is **not** a CSS variable set. It is a lookup indexed by `hash(id) % 8`, so it lives in a TypeScript constant with both role assignments, not in `:root`.

`shadcn init` writes the standard mappings. The additions that need exposing as Tailwind utilities are `--board-canvas`, `--board-column`, `--primary-hover`, `--primary-subtle`, `--primary-subtle-foreground`, `--destructive-text`, `--due-soon`, `--overdue`, `--done`, `--shadow-overlay` and `--shadow-dragged`.

### 10. Deviations from shadcn defaults

Everything else is inherited. Five changes:

| Token | shadcn default | Here | Why |
|---|---|---|---|
| `--primary` | near-black | plum `#6B4E9E` | the product needs a brand; §3.1 |
| `--destructive` | red-500 | rose `#E11D48` | red sits close enough to amber in hue that the two blur at badge size |
| `--radius` | `0.5rem` | kept, but with an explicit 4/6/8 ladder | §5 |
| `--input` | stone-300 | stone-500 | shadcn's default fails 1.4.11; §3.4 |
| — | — | `+ --board-canvas`, `+ --board-column` | `--background` stays white so every shadcn component works unmodified on a white surface; the board's two greys are additions rather than a redefinition |

The last row is the important one. Redefining `--background` to the canvas grey would have made every inherited component inherit a grey page, which is not what any of them are drawn for.

---

## Channel budget — what ADR-018 inherits

A visual channel gets **exactly one meaning**. These are spent:

| Channel | Meaning |
|---|---|
| Plum | "the one action here" and "where you are" — never state |
| Amber | deadline approaching — **deadlines only** |
| Rose | overdue, and destructive intent |
| Emerald | done / archived |
| The eight identity hues | who, and which board |
| Surface brightness (3 steps) | permanent hierarchy |
| Shadow | temporarily lifted above the page |
| Border | an edge |
| Radius (3 steps) | object size class |
| Type scale (5 steps) | reading order |
| Weight (400/500/600) | emphasis within a block |

**Left deliberately unspent, for ADR-018:**

| Channel | Reserved for |
|---|---|
| **Opacity** | "not fully real" — covers **pending** *and* **drag-source** as a single meaning |
| **Motion** | "something just changed" — **server-corrected**. *Re-pointed by [ADR-018](adr-018-state-vocabulary.md) §8 to "a teammate did this": a rebalance preserves visible order, so this meaning had no render site* |
| **Dashed border** | "where it will land" — the drop placeholder |
| **Cursor** | "what you can and can't do here" |
| **Icon** | lock (private board), ~~archive (archived)~~. *Archive struck by [ADR-018](adr-018-state-vocabulary.md) §9 — archived boards are unreachable in Phase 1* |
| **Column panel tint** | drag-over column. Available because §2 chose contained panels |
| Border weight (2px) | unassigned — a candidate for selection |
| Blur, skeleton/shimmer, italic, strikethrough | unassigned |

The four states named in the brief are all covered without a new hue: **pending** and **drag-source** share opacity, **server-corrected** takes motion, and **read-only** is best expressed by **removing controls plus a banner — which costs zero channels.**

Two requirements ADR-018 inherits as constraints rather than suggestions:

1. **The drag-over tint must have a real step.** The superseded draft used `#FEF3C7` against the `#F5F5F4` panel — **1.02:1**, an invisible "which column am I over" signal on the P0 drag path.
2. **The drop placeholder must reach 3:1.** The draft's `#D6D3D1` against the panel is **1.37:1**, which fails 1.4.11. The draft flagged `--input` at 1.49:1 as a known gap and missed this one.

---

## Options Considered

### Amber `#F59E0B` primary — the original brainstorm

Rejected on four structural grounds, not comfort:

1. **White text on amber is 2.15:1.** Every amber button must carry near-black text permanently — there is no amber button with white text at any size.
2. **Primary could not equal the ring.** Amber-500 is 2.15:1 on white, below the 3:1 boundary minimum, so rings needed amber-600 and ring-abutting-text needed amber-700. Three values for one role.
3. **It spends the best attention-grabber on the most mundane job.** Amber is the strongest "look here" colour available, and a primary button is the one element on screen that already announces itself by position and fill.
4. **It collapses the deadline ladder.** Due-soon loses its natural colour, so neutral → amber → rose becomes neutral → rose, and "due tomorrow" and "three weeks overdue" land in the same bucket.

The draft's own first open question anticipated this: *"Amber is a high-arousal color. The original brief was 'focus and hard work' — amber delivers the second half well and works against the first."* Its arithmetic was otherwise sound; two values were understated in its own favour (`primary-foreground on primary` is 8.14:1, not 7.3; amber on the dark background is 9.20:1, not 8.3).

### The Monday.com palette

Rejected on mechanism. **Seven of its eight colours fail white text**, but the decisive problem is informational: in Monday, status *is* a coloured field, so the colour volume is load-bearing. CollaboTask has no status field — status is which column a card sits in — so vivid pills would carry no information and would be pure texture at 8 cards per column.

### Board structure B — tracked

No column panel: a header with a hairline rule, cards directly on the canvas. Genuinely attractive, and it buys a **larger** card step than direction A — 1.26:1 canvas-to-card against A's 1.09:1 panel-to-card. Not chosen, but it was close, and the render is what settled it rather than the measurement.

Cost of not choosing it: `--board-column` would have been unnecessary, and ADR-018 would have had one fewer tintable surface for drag-over.

### Board structure C — bare

Rejected **by measurement, before taste.** A card with no surface step has nothing but its border to identify it, and a border doing that job unaided must reach 3:1. A hairline manages 1.43:1; the lightest compliant value is stone-500 `#78716C` (4.59:1 on a near-white canvas, 3.65:1 on the dark one). So the *least drawn* direction ends up with the **heaviest visible edges of the three** — it defeats its own purpose. In dark it is worse: card and canvas are both `#1C1917`, a literal 1.00:1 step, leaving the outline as the only cue.

The "make it like Linear" instinct points here, but **Linear's calm comes from being a list** — rows separated by dividers, with no card objects at all. That pattern does not survive the move to columns.

### Inter, and a two-family pairing

Inter is the safer choice at 14px — a taller lowercase means it holds up better at small sizes, where Poppins' lowercase is roughly three-quarters of Inter's at the same nominal size. Poppins was chosen anyway, for character. The practical consequence is recorded rather than hidden: **Poppins reads smaller than its nominal size**, which is exactly why the 14px-against-16px comparison is the first specimen on the Figma Foundations page.

A display/body pairing was considered and dropped. With a 5-step scale topping out at 28px there is no display size for a second family to occupy, and one variable font file is one request.

### A 16px default

Rejected in favour of 14px. 16px is more comfortable in isolation; it costs words per card and cards per screen, and §1's 8-card target is the binding constraint. The side-by-side specimen exists so this can be revisited by looking rather than re-argued.

### Dark mode shipped in Phase 1

Rejected. The cost is not the CSS — it is that ADR-018 would have to draw every interaction state twice, doubling the frame count of the session that matters most. §7's dark-ready posture keeps the values honest while deferring that cost.

### Two density scales for touch

Rejected with §8. A second scale is the visible tip of a different interaction model, and committing to it would mean designing a phone Kanban, which is Phase 2 work.

**[ADR-019](adr-019-responsive-touch-scope.md) adopted them.** This reasoning was not refuted: ADR-019 agrees a second scale is real cost and pays it in its §7, and agrees a phone board deserves design rather than reflow. What changed is the requirement — and the timing, since with no frontend code written the layout half is nearly free.

---

## Corrections to the superseded draft

`docs/architecture/design-tokens.md` is **superseded in full** by this ADR. It was never committed and never adopted; it is self-labelled "Proposal, not adopted". Its elevation ladder (`#E7E5E4` canvas → `#F5F5F4` column → `#FFFFFF` card) is carried forward unchanged. Four measured defects are not:

| Token | Draft value | Measured | Resolution |
|---|---|---|---|
| `--drag-over` | `#FEF3C7` on `#F5F5F4` | **1.02:1** | deferred to ADR-018 with a stated minimum |
| `--drop-placeholder` | `#D6D3D1` on `#F5F5F4` | **1.37:1**, fails 1.4.11 | deferred to ADR-018 with a stated minimum |
| `--success` | `#10B981` on white | **2.54:1**, shipped unaudited | replaced by `--done: #047857` at 5.48:1 |
| `--column-header` | identical to `--column` | **1.00:1** | token removed; the header carries a rule (§2) |

The draft's deeper problem was scope rather than arithmetic: it had **already assigned every channel ADR-018 needs** — amber for acting, teal for persistent state, opacity for drag-source, dashed border for the placeholder, and ring and shadow each doing two jobs — leaving nothing for pending, read-only or server-corrected. §*Channel budget* exists to prevent that recurring.

Its avatar palette is also replaced. All eight of its hues sat in the cool half of the wheel, which put every chromatic primary candidate 0–12° from one of them; the collision could not be solved by choosing a different brand hue. §3.3 resolves it with **saturation as the separator** rather than hue.

---

### Out of scope (deferred, recorded)

- **The state vocabulary** — which channel each interaction state claims. ADR-018, with the budget above.
- **Column width.** 280px is used throughout §1 and in the comparison render so that the structure comparison was not confounded, but the value itself was never grilled.
- **Motion specifics** — durations, easing, `prefers-reduced-motion` behaviour. Motion is reserved as a channel in §*Channel budget*; what it does is ADR-018's.
- **The Astro landing page's visual language.** A landing page can carry more colour than a dense board; it is not bound by §3.
- ~~**Touch and phone layouts** (§8)~~ — **moved into Phase 1 by [ADR-019](adr-019-responsive-touch-scope.md)** (2026-11-17). **Avatar images** (§3.3) and **labels, comments and attachments** remain Phase 2.
- **Icon set selection.** The icon channel is reserved; which family supplies it is unresolved.

### Downstream edits this ADR requires

1. **PRD §5.2** — ✅ done (2026-11-13), then **undone by [ADR-019](adr-019-responsive-touch-scope.md)** (2026-11-17): the deferral row was removed and the §5.3 non-goal rewritten, because phones and tablets are now Phase 1. Recorded here so this checkmark is not read as describing the current PRD.
2. **`docs/architecture/design-tokens.md`** — ✅ deleted (2026-11-13), superseded by this ADR.
3. **SRS** — ✅ done (2026-11-13). §4.3 now defines `background_color` as an identity accent limited to the eight hues of §3.3, with the client-side fallback rule; the schema default `'#0079BF'` is annotated as a legacy value clients map on read. No backend change.
