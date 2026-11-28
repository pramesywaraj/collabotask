# ADR-049: The Logo — Three Cards Make a C

- **Status:** Accepted — **§7 drawn, and §8's fourth look added, after acceptance** (later on 2026-11-28):
  - the `Logo` part's `Sidebar lockup` look (147 × 28)
  - the lockup in all five `Sidebar` looks' top rows, after the toggle. All 48 copies followed: 1 on Account, 14 on Workspaces, 33 on the Board pages
  - the full lockup on the left of 2.3's six app bars, where *CollaboTask* was
  - the full lockup first in the form on 2.1, 2.2 and 2.13. The form's gap, `spacing.section-gap`, puts it 24 above the heading (the user's call). Register's form grows to 640 and still fits its board
  - the Foundations catalogue's three rows, its title reading 228 and the theme table 70 each

  The binding check on every board touched reports only the expected `Name`, at 20 and at 17.5. **Left as it is:** the mark's cards start 5px inside its box, so on the forms the logo sits 5px in from the heading's left edge. **Not looked at by eye:** the sidebar copies on the four Board pages, which were not opened. Their data reads `Toggle | Logo`.
- **Date:** 2026-11-28
- **Scope:** CollaboTask's logo, and what follows from it. It settles:
  - **the mark**: three cards set out as a C, one of them tilted, on a 32 × 32 grid, with a straight version for small sizes
  - **its colours**, as three new tokens that follow the Theme switch, and the app icon's tile
  - **the name beside it**: *CollaboTask* in Poppins SemiBold, in deep plum
  - **clear space, the smallest sizes**, and **where the logo appears** in the app
  - **the files** in `brand/`, and what the code needs

  **Three tokens are added, once per theme set (222 → 228 entries). One part is added, `Logo` (34 → 35), and one page, `Brand`.** It **extends [ADR-017](adr-017-ui-foundations.md) §3.1 and §9**: plum gains a third place, the logo, and §9 gains three variables. It **supersedes [ADR-034](adr-034-side-menu.md) §10.1's *CollaboTask* in `type.title` and [ADR-033](adr-033-workspace-screens.md) §2's on 2.3 only**: both become the lockup. Everything else in those sections stands.
- **Upstream edits:** none. The PRD, the user stories and the SRS do not mention a logo.

## Context

The app had no logo. The sidebar's top row and 2.3's app bar wrote *CollaboTask* in `type.title`, the browser tab had no icon, and nothing could go on a phone's home screen.

The choice was made over three rounds on a rendered concept page, each mark shown at its real sizes, in Light and Dark, in one colour, in the sidebar and in a browser tab:

1. **Four concepts**: three cards make a C, a card with a face on it, a letter c and a dot, and a board mid-move. The user leaned to the first.
2. **Four versions of it**: as drawn, smaller corners (A1), square corners (A2), and the plum card aligned with the top card (A3), with the six combinations. The user chose **A1**, over my recommendation of A1 with A3.
3. **Colour**: five ways to colour the cards and five tile colours. The user chose **plum, dark and light** over my recommendation of near-black and plum, and the **plum tile with white cards**, which was my recommendation. The name in **deep plum** was the user's choice over near-black. Where the logo appears was my recommendation, taken as it was.

It was then drawn in Penpot. The user made a `Brand` page in this file rather than a separate file, which lets the logo use the file's tokens.

## Decision

### 1. The mark

Three cards on a 32 × 32 grid, read as the letter C. Two stand still. The third, at the bottom right, is tilted as if someone has just picked it up: the app's moment, work that moves while you watch.

| Card | Position and size | Corners |
|---|---|---|
| Long card | x 5, y 5, 8 × 22 | 1.33 |
| Top card | x 15, y 5, 12 × 8 | 1.33 |
| Moving card | x 15, y 19, 12 × 8, then moved 1.5 right and turned 8° anticlockwise about (21, 23) | 1.33 |

**The corners are 4/3, a sixth of a card's height**: the proportion of the app's buttons, 36 tall at Fine with 6px corners (ADR-017 §5, [ADR-022](adr-022-form-control-geometry.md)). The moving card sticks out past the top card's right edge on purpose. The aligned version was shown and not chosen.

### 2. The straight version, for small sizes

Below 24px the tilt goes soft, so a second drawing straightens the card and makes all three cards a little larger, with 2 grid squares between them, so each gap stays a whole pixel at 16:

| Card | Position and size | Corners |
|---|---|---|
| Long card | x 4, y 4, 9 × 24 | 1.5 |
| Top card | x 15, y 4, 13 × 9 | 1.5 |
| Moving card | x 15, y 19, 13 × 9 | 1.5 |

### 3. Colours

The mark's colours are tokens, so the logo follows the Theme switch wherever it appears on a screen:

| Token | Light | Dark | Use |
|---|---|---|---|
| `color.logo.still` | `#3B2F57` | `#D6C7EE` | the long card and the top card |
| `color.logo.moving` | `#A78BD9` | `#A78BD9` | the moving card |
| `color.logo.name` | `{color.logo.still}` | `{color.logo.still}` | the name |

The values come from the app's plum ladder: `#3B2F57` is Dark's `primary-subtle`, `#D6C7EE` Dark's `primary-subtle-foreground`, and `#A78BD9` Dark's `primary`. In Light the moving card is the lighter one, and in Dark the darker one, so in both themes it stands apart from the still cards by shade.

**The pale moving card at 16px in Light** is 2.86:1 against white. It was checked and kept: it stays visible, a little softer than the still cards. A logo is not text or a control, so WCAG's contrast minimums do not bind it.

**Fixed colours**, which never follow the theme:
- **one colour**: `#1C1917` on a light ground, `#FFFFFF` on a dark one, for print and stamps
- **the app icon's tile**: `#6B4E9E`, cards `#FFFFFF`, moving card `#BBA3E3`

**Never in the logo: amber, rose or green.** In the app they mean due soon, overdue or delete, and done.

**This extends ADR-017 §3.1.** Plum stays confined to the primary action and to "where you are" inside the interface. The logo is a third place: an identity mark, which never carries state. That is why the logo uses its own tokens rather than `color.primary`.

### 4. The name and the lockup

- The name is written **CollaboTask**, with a capital T, everywhere. (The design file is still called *Collabotask App*, which predates this.)
- **Poppins SemiBold (600)**, letter spacing 0, in `color.logo.name`.
- **The lockup** is the mark, then the name: **mark 32, gap 8, name 20**. Scaled, the name is five-eighths of the mark's height and the gap a quarter. The gap is bound to `spacing.icon-gap`, the space between an icon and its label.
- I first proposed a 24 mark with a 16 name. It became 32 and 20 so that the name keeps the 20px the sidebar already uses (`type.title`'s size), and because 32 is the size at which every card edge lands on a whole pixel.
- In the 32 × 32 box, the name's baseline is at 23 and its letters run from 8.2 to 23.2.
- **In the sidebar it is drawn at 28** (§7): mark 28, gap 8, name 17.5.

### 5. Clear space and the smallest sizes

- **Clear space**: a quarter of the mark's height on every side, measured from its 32 × 32 box. That is the long card's width: 8 at 32, 32 at 128. The lockup takes the same.
- **24 and up**: the full mark. **16 to 23**: the straight version. **Never below 16.**
- **The lockup**: the mark at 24 or more.

### 6. The app icon

- A **plain square**, 512 × 512, filled `#6B4E9E`. Phones round the corners themselves, so the file has none.
- The mark is **320 wide** (10px to a grid square), **centred on its visible shapes rather than its box**. Centred on its box, it sat 9px right of centre, because the tilted card adds weight on the right. Its offset is 86.95, 93.9.
- Every shape lies within 164px of the tile's centre (measured to the corners of the visible shapes' bounding box), inside the 205px circle Android keeps for a "maskable" icon, so nothing is cut when a phone crops it to a circle or a squircle.

### 7. Where the logo appears

| Where | What |
|---|---|
| The sidebar's top row, on 2.4 and the board | the toggle, then the lockup at 28 (below) in place of *CollaboTask* in `type.title` (supersedes ADR-034 §10.1's text) |
| 2.3's app bar, on the left | the lockup in place of *CollaboTask* (supersedes ADR-033 §2's text) |
| Above the Sign in and Register forms (2.1, 2.2, and 2.13, a copy of 2.1) | the lockup, left-aligned with the form |
| The board screen's app bar | **nothing**: it holds the board's name and the plum *Add column*, and the sidebar already carries the logo |
| The browser tab | the straight mark alone (`favicon.svg`), never the tile |
| A phone's home screen, and the installed app | the app icon |

**This is decided here and not yet drawn on the screens.** Drawing it is the next step.

**In the sidebar the lockup is smaller: mark 28, gap 8, name 17.5, 147 wide, at both densities.** The user's call, as recommended. The full lockup would fit at Fine, with 9 to spare (16 + 24 + 8 + 167 + 16 = 231), but at Coarse the toggle is 44 (`size.target-min`) and the row would need 251, **11 too many**. At 147 the row is 231 at Coarse and 211 at Fine. One size at both densities follows [ADR-047](adr-047-illustration-size-192.md)'s rule that what is not a target does not change with the pointer. The name stays five-eighths of the mark, and the gap stays `spacing.icon-gap`.

2.3's bar and the Sign in and Register forms have room at both densities, and take the full lockup.

### 8. In the design file

- **Page `Brand`**, first in the file, in three rows. Each row starts with a `Doc / Row label`, as on the screen pages (ADR-045 §3):
  - **Logo**: the part, `Doc / Logo — construction` (the mark at 256 on its grid) and `Doc / Logo — sizes` (64, 32, 24, the straight 16, and the lockup)
  - **Logo — fixed colours**: `Doc / Logo — Light and Dark`, `Doc / Logo — one colour` and `Doc / Logo — app icon` (the 512 master and a rounded 120 preview)
  - **Logo — rules**: `Doc / Logo — space and size`
- **The `Logo` part**, with one axis: `Look` = Mark · Mark small · Lockup. The cards are paths set to **Scale**, so a resized copy draws the mark at any size. Each path's fill is bound to `color.logo.still` or `.moving`. The lockup holds a `Mark` copy and the `Name` text: `font.family`, `font.weight.semibold`, `color.logo.name`, 20/28 typed. The part lives on `Brand`, not `Components`, so the logo's whole record is on one page. **When the screens are drawn it gains a fourth look, `Sidebar lockup`** (mark 28, name 17.5), because a flex row holding text does not scale when it is resized.
- **`Fixed /` shapes** show colours that must not follow the theme: the Light and Dark pairing, one colour, and the app icon. They are detached copies, with their colours typed in. A `Fixed /` shape anywhere but `Brand` is a defect.
- **The binding check (inventory §7.5) gains two expected rows**:
  - the lockup's `Name`, on the part and its copies: `typography(20px)`, because its weight is bound and no type style is 20/600
  - anything named `Fixed / …`, or inside one, on `Brand` only: `fill`, the surfaces' typed padding and gaps, and the preview's corners

  Run on 2026-11-28, the check reported exactly these and nothing else: 220 shapes, 54 of them inside `Fixed /` boards.

### 9. The files

In `brand/`, at the root of the repository:

| File | What it is |
|---|---|
| `logo.svg` / `logo-dark.svg` | the lockup in Light / Dark colours, 167 × 32 |
| `logo-mono.svg` | the lockup in `currentColor`, so CSS sets its colour |
| `logo-mark.svg` / `logo-mark-dark.svg` | the mark alone |
| `favicon.svg` | the straight mark, with its Dark colours inside the file (a `prefers-color-scheme` rule) |
| `favicon.ico` | 16, 32 and 48, Light colours, for older browsers |
| `icon-512.png` / `icon-192.png` | the app icon, for the web app manifest |
| `apple-touch-icon.png` | the app icon at 180, for an iPhone's home screen |

How they were made:
- **The SVGs are written from the geometry in §1 and §2**, not exported from Penpot, so they are small and readable.
- **The name in the lockup files is outlined**, so it needs no font. The outlines come from `Poppins-SemiBold.ttf` (google/fonts, OFL), set with opentype.js at Penpot's layout. The name is 126.8 wide, the same as Penpot's, and its baseline is at 23.
- **The PNGs are exported from the Penpot master**: the 512 from `Fixed / App icon`, and the 192 and 180 scaled down from it. The `.ico` is assembled from three exported rasters.

### 10. The code

- **§9 gains three variables in both modes**: `--logo-still` (`#3B2F57` / `#D6C7EE`), `--logo-moving` (`#A78BD9`) and `--logo-name` (= `--logo-still`).
- **In the app, the logo is an inline SVG, not an `<img>`**, so it can read those variables and follow Dark. The mark is `aria-hidden`. **The name is real text**, Poppins 600 at 20 in `--logo-name` (17.5 beside a 28 mark in the sidebar), so a screen reader reads *CollaboTask* and the link has a name. The outlined `logo*.svg` files are for places without the font: the README, emails, a link preview.
- **In the `<head>`**: `favicon.svg` (`type="image/svg+xml"`), `favicon.ico` as the fallback, `apple-touch-icon.png`, and a manifest listing `icon-192.png` and `icon-512.png`, with `icon-512.png` listed a second time with `"purpose": "maskable"` (two entries rather than one `"any maskable"`, which some browsers warn about).
- **The files go in the `public/` folders** of the Vite app and the Astro landing, at the scaffold.

## Options considered

- **Concepts B, C and D.** A card with a face read as a notification badge. The letter and dot said nothing about the app. The board mid-move turned to noise at 16 and looked like Trello's mark. Not chosen.
- **A as drawn, A2 square corners, A3 aligned, and the combinations.** A's rounder corners are friendlier. A2 matched nothing in the app, which is rounded everywhere. A3 was my recommendation, with A1. **The user chose A1.**
- **Near-black and plum.** My recommendation: plum would have stayed a small accent. **The user chose plum, dark and light**, which makes the whole mark plum at the cost of more plum beside the sidebar's plum current row.
- **All plum, warm grey and plum, team colours.** All plum loses the moving card's difference at 16. Grey reads as switched off. Team colours borrow the board and people colours.
- **The white, near-black and pale plum tiles, and the plum tile with a white moving card.** The plum tile with white cards was chosen, as recommended: outside the app nothing competes with plum, and a coloured tile is easier to find.
- **The name in near-black.** My recommendation, as ordinary text. **The user chose deep plum**, to match the mark.
- **Typed-in colours instead of tokens.** Would have left the logo in its Light colours on every Dark screen. Not chosen.
- **In the sidebar at Coarse: the full lockup at Fine and a smaller one at Coarse, or the mark alone at Coarse.** The first gives each density the biggest lockup that fits, but changes with the pointer and needs a density token. The second drops the name exactly where the slide-over covers the screen. **The user chose one smaller lockup at both densities**, as recommended (§7).
- **A separate Penpot file.** My first suggestion, to keep the finished file untouched. The user made the page in this file instead, which is what makes the tokens possible.

## Out of scope (deferred, recorded)

- **Drawing the logo on the screens** (§7): the part's `Sidebar lockup` look, the sidebar's top row in `Sidebar` and its copies, 2.3's app bar, 2.1, 2.2 and 2.13. That is the next step.
- **The Foundations catalogue.** Its tables are static text, so the three new tokens need three rows in the theme table, which goes to 70 each, and the title reads 228.
- **A link-preview image** (`og-image.png`, 1200 × 630), with the Astro landing.
- **The app's `Logo` component**, at the scaffold (§10).
- **Letter spacing.** Left at 0. Tightening it by eye, around *oT* especially, was not tried.
- **Not looked at:** the live Theme switch to Dark with the part on screen. The Dark values were checked on the typed-in pairing, which uses the same values, and in the token sets. The file was not switched, because a switch makes part copies on every page show their main's text size until each page is opened again ([ADR-043](adr-043-the-lists.md) *Operational note*).

## Downstream edits this ADR requires

- **Done in the file**: the three tokens, the `Logo` part, the `Brand` page and its boards, and the binding check's run.
- **Done in the repository**: `brand/`, staged and not yet committed.
- **Done in the docs**:
  - **Inventory** §7.2: the token sets row (`theme-light` 70, `theme-dark` 70, **228**), a `Logo` row and a `Brand` page row. §7.5: the two expected rows in §8.
  - **ADR-017 §3.1 and §9, ADR-033 §2, ADR-034 §10.1 and ADR-045 §1** carry a note pointing here. They are annotated, not edited.
  - **frontend-architecture.md**: the `Logo` component, the head tags and the manifest (§10).
  - `CLAUDE.md` `## Now` and `docs/build-log.md`.
- **Not yet**: the screens and the Foundations catalogue (*Out of scope*).

## Operational note

Found while drawing and exporting:

- **`penpot.flatten()` does not outline text.** It made a path of the text's bounding box, and the copy it flattened was inserted into the lockup's flex row, pushing the name 135px right. Outline text outside Penpot.
- **Penpot's SVG export keeps text as `<text>`**, with an `@font-face` pointing at Penpot's own font server. It is not usable as a logo file.
- **Setting a fill to the colour its token already holds keeps the binding.** The Light pairing's cards and name stayed bound to `color.logo.*` after their colours were typed in, so a switch to Dark would have recoloured a "fixed" copy. To unbind, set `fills = []` first, then the colour.
- **Straight after `createVariantContainer`, `library.local.components` did not list two of the three looks.** `container.variants.variantComponents()` did. Look components up through the container.
- **A flex column whose children were built off-canvas and appended later slipped to `horizontalSizing: "fix"`**, at the width it had before they arrived. Set it back to `auto`, and check `horizontalSizing` after building a column that way.
- **Wrapped text is measured as one line** until its `growType` goes to `fixed` and then back to `auto-height`, in two calls. A flex row's auto height also stuck at a stale 75 until it went `fix` and then `auto`, in two calls.
- **`export_shape` with `page` timed out on `Brand`.** Export boards one at a time.
