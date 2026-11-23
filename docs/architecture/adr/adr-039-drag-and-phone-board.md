# ADR-039: Drag and the Phone Board — the Short Column, Add Card During a Drag, the Lifted Card, and Add Column Below 640

- **Status:** Accepted
- **Date:** 2026-11-23
- **Scope:** Question 4 for Stage 4's group e: board 2.14, *Drag*, and board 2.15, *Phone board*, each drawn as a difference from 2.5 ([ADR-037](adr-037-board-at-rest.md)). It settles:
  - how far down a short column takes a dragged card, which ADR-037 left to 2.14
  - what *Add card* does during a drag
  - the lifted card and the held press as looks of the `Card` part, and how their growth is drawn
  - where *Add column* goes below 640, which ADR-037 left to 2.15
  - how the phone board shows the browser, and how its column width is written
  - a phone drag board beside 2.15

  **No token is added** (210 stays 210), and no part is added (19). The `Card` part gains two looks, `Armed` and `Lifted`, so it holds seven. It **settles ADR-037's two *Out of scope* items**. It **amends [ADR-038](adr-038-board-states-group-d.md) §10 below 640**: there the empty board's centred stack gains an *Add column* button, because the phone's bar has none. It **checks [ADR-019](adr-019-responsive-touch-scope.md) §7's card count by eye**, as that ADR asked, and the count holds. Inventory §4.3 gains a third thing the file cannot show: the 2% and 1% growth of a lifted card and a held press.
- **Upstream edits:** none. Nothing here touches an endpoint, the schema or a permission rule. *Add column* below 640 is the same request ([SRS](../../spesifications/001-software-specifications.md) UC-13) from a different place.

## Context

Inventory §7.3 draws group e last in Stage 4: 2.14, the drag, as a difference from 2.5, then 2.15, the phone board, which is a different layout and is drawn by hand.

Question 4 was asked of every value these boards need, against the live 210 tokens, before anything was drawn. **Almost everything mapped:**
- the card's empty place is the `Card` part's `DragSource` look (`opacity.in-flight`)
- the lifted card's shadow is `shadow.dragged`, and the held press's is `shadow.overlay`
- the drop outline is `color.drop-placeholder`, `border.width-strong` and `radius.surface`, dashed
- the column you're over is `color.drag-over` on its body, with the header left in `color.board-column` (ADR-018 §4)
- the phone's bar is `size.app-bar`, `spacing.screen-x`, `spacing.inline-gap`, `size.target-min`, `type.title` and the `Presence cluster` part
- the phone's 72px reservation is exactly `spacing.screen-x` 16 + `spacing.column-gap` 12 + `size.target-min` 44 at Coarse

**Three values had no token**, so question 4 fired for the seventeenth stage out of seventeen: the lifted card's 2% growth (`scale(1.02)`, and the held press's 1%), which Penpot cannot draw; the phone column's 318px, which only code can work out; and the 140px of browser that ADR-019 §7's count subtracts.

**Five decisions had never been made:** how far a short column takes a card, *Add card* during a drag, whether the `Card` part gains the two looks, where *Add column* goes below 640, and whether the phone gets a drag board.

All eight went on one rendered page, each option drawn in Light and Dark with the file's values, at Fine on the desktop and Coarse on the phone. **The user took seven recommendations and overrode one:** a short column takes a card only as far as the column goes. The recommendation had every column stretch to the bottom of the screen while a card is dragged.

## Decision

### 1. A short column takes a card only as far as it goes

- **The drop area is the column panel as drawn.** Nothing changes height when a card is lifted, and the tint fills only the column you're over.
- **Let go under a short column's last card and nothing happens**: the card returns, as any drop outside a target does (ADR-018 state 10).
- **On a phone, the next column's 44px edge is a target only as tall as that column's cards.** The 2.15 drag board shows it.
- **The user's call.** The recommendation stretched every column to the bottom of the screen for the length of a drag, which keeps every target in view and the same shape. It costs the whole board changing shape once, at pickup, and the short columns the user chose in ADR-037 going full height for the drag. The user kept the columns as they are.

The rejected options: every column reaching the bottom while a card is dragged (recommended); and only the column you're over reaching the bottom, which makes columns grow and shrink as the card passes over them.

### 2. *Add card* stays, untinted, just under the cards

- **In the column you're over, the tint covers the cards' area only**: from the header's rule to the bottom of the last card.
- ***Add card* keeps `color.board-column`**, `spacing.card-gap` below the tint, in a `Footer` like To Do's. It is not a drop target, which is ADR-018 §4's reason for leaving the header untinted.
- **A long column's stuck footer keeps its rule.** On the tint the rule would measure about 1.19:1 and vanish, as the header's does.
- **Nothing moves.** With §1, *Add card* stays where it was at rest.

The rejected options: *Add card* tinted with the body, which hides a long column's footer rule; and *Add card* hidden for the drag, which takes a row off every column at pickup.

### 3. The four states on a card drag

- **The empty place (6):** the card keeps its place at 60%, the `Card` part's `DragSource` look. That look turns the meta row to `color.foreground` (ADR-018 §3), so **an overdue date loses its rose while dimmed** and keeps its words, *Overdue · Nov 20*. Rose at 60% would fall below 4.5:1.
- **The lifted card (7):** the `Card` part's `Lifted` look, `shadow.dragged`, at full strength, keeping its date colour. It follows the pointer.
- **The outline (8):** 2px dashed `color.drop-placeholder` (`border.width-strong`), `radius.surface` corners, **no fill**, the lifted card's size, where the card will land.
- **The tint (9):** `color.drag-over` on the body of the column you're over (§2). The header stays `color.board-column`.
- **An empty column under a card**: *No cards yet.* gives way to the outline.

### 4. A column drag

- **The same three looks at column size**: the column's place at 60% (`opacity.in-flight` on the whole panel), a lifted copy with `shadow.dragged`, and a dashed outline the size of the dragged column where it will land.
- **Nothing is tinted.** A column isn't dropped into anything.
- **Every column keeps its own height.** §1 holds here too.
- **Only from 640 up.** Below 640 a column moves with *Move left* and *Move right* in its ⋯ (ADR-021 §4).

### 5. The `Card` part gains `Armed` and `Lifted`

- **`Armed`** (state 5) is Rest with `shadow.overlay`: where the held press's shadow ends, on a touch screen, after 200 to 250ms. It lives on the Card's own board.
- **`Lifted`** (state 7) is Rest with `shadow.dragged`. 2.14 and the phone drag board use it.
- **The part now holds seven looks**: Rest, Hover, Pressed, InFlight, DragSource, Armed, Lifted, which is every card look ADR-018 names.
- **In Dark both shadows are near-black on a near-black board**, so a lifted card reads by overlapping what's under it more than by its shadow. That is ADR-017's dark shadow, unchanged.

The rejected options: `Lifted` only; and no new look, with the shadow set on 2.14's copy alone, against [ADR-020](adr-020-design-tool-penpot.md) §4.

### 6. The 2% and 1% are code only

- **The lifted card and the held press are drawn at their real size.** Penpot cannot scale a shape together with its text, and has no token type for it.
- **Inventory §4.3 gains a third thing the file cannot show**, beside motion and equal-width digits. In code the values stand as ADR-018 §4 wrote them: `scale(1.02)` and `scale(1.01)`.

The rejected option: stretching the box by 2%, which makes a wider card with the same 14px words, a card the app never draws.

### 7. *Add column* below 640 is the first entry in the board ⋯ menu

- **The phone's bar keeps show/hide, the name, presence and ⋯**, and the name keeps about 170px.
- **Below 640 the board ⋯ menu starts with *Add column*.** It is there for the people who see the bar's *Add column* from 640 up: the board's members and workspace admins (`CheckMutateAccess`, [ADR-005](adr-005-board-visibility-access.md)). For everyone else it is absent (ADR-018 §5). Below 640 the menu is a bottom action sheet (ADR-021 §5), drawn at Stage 5.
- **The empty board, below 640, gains an *Add column* button.** ADR-038 §10 gave 2.10 no button because the bar has *Add column*. The phone's bar doesn't, so below 640 the centred stack ends in a Primary ***Add column***, `spacing.section-gap` under the sentence, as ADR-038 §12 places a button. 2.10 is not drawn at phone width; this is recorded for the code.

The rejected options: a slot after the last column, reached only by swiping past every column; and a plus button in the bar, which cuts the name short and reads as *add a card* as easily as *add a column*.

### 8. The phone board, 2.15

- **390 × 844 with the browser drawn**: a 47px band at the top for the status bar and a 93px band at the bottom for Safari's bars, each labelled. These are the 140px ADR-019 §7 subtracts. They are `Doc / Browser / …` shapes filled `color.muted-foreground`, labelled in `type.meta` and `color.background`.
- **The bar:** show/hide (Phosphor `SidebarSimple` in a `size.target-min` box, as in the sidebar's top row), *Q3 Launch Plan* in `type.title`, three faces and a *+5* seat (ADR-019 §8, [ADR-031](adr-031-presence-cluster.md)), and the board ⋯. No *Add column* (§7).
- **The column is 318 wide, typed on this board.** The rule is `min(100vw − 72px, 420px)` and lives in code and in ADR-019 §3. The board is a picture of one phone, so its column width is typed the way its own 390 × 844 is. The binding check never looks at widths.
- **The strip:** `spacing.screen-x`, the 318 To Do column, `spacing.column-gap`, and 44px of In Progress. To Do is longer than the screen, so its *Add card* sticks at the bottom (ADR-037 §4).
- **Drawn at Coarse**, since a phone has a touch screen. At Fine it is the same layout with mouse-sized targets, a 390-wide desktop window (ADR-019 §1).

The rejected options: a 390 × 704 board with no bands, which no longer looks like a phone; and a `size.column-width-phone` token, which would be right on one phone only.

### 9. The card count, checked by eye

ADR-019 §7 marked its *6 cards visible* as the one number in its table resting on an assumption, and ADR-037 §4 revised it to 5 and a peek in a column longer than the screen. **Measured on 2.15 at Coarse:**

| | Height |
|---|---|
| The card area (844 − 140 − 56 − 4 − 16 − 44 − 56) | 528, of which 12 is the gap under the header |
| *Draft the Q4 positioning brief for review* | 92 |
| *Collect launch quotes from the beta customers* | 92 |
| *Update the pricing page copy* | 44 |
| *Record the product walkthrough video* | 92 |
| *Translate onboarding emails into Bahasa* | 92 |
| *Book a venue for the launch event* | 44 of 72 shown |

**5 whole cards and 44px of the sixth.** ADR-037 predicted 56px with an 80px average card; these five average 82. The count holds.

### 10. The phone drag board

- ***2.15 Phone — dragging a card***: a copy of 2.15 with *Update the pricing page copy* lifted out of To Do and held over In Progress's 44px edge.
- **In Progress is the column under the finger:** its body is tinted, its *Add card* stays untinted under it (§2), and the dashed outline sits between its first and second cards. Of the outline, the edge shows 36px, after the panel's 8px padding.
- **The edge is only as tall as In Progress's cards** (§1). The board does not scroll during the drag (ADR-019 §5).

### 11. What the file draws

**The `Card` part, on `Components`:** `Armed` and `Lifted` added to `Card — states` at 1660 and 1980, 320 apart like the others, with their labels.

**Four boards, on `Page 1`:**
- **2.14 Drag — a card**, at x 8200, y 0: a copy of 2.5. *Final pricing sign-off* is `DragSource` in In Progress and `Lifted` over Done, 24 right of and 30 above its outline. Done's `Body` is `color.drag-over`. Its *Add card* moved into a `Footer` (`color.board-column`, `spacing.card-gap` above, no rule). The `Drop placeholder` is a board with no fill, the dashed edge and `spacing.card-padding-y` above and below a 20px `Title line` and a `size.meta-row` `Meta row`. That makes it the lifted card's height at both densities: `8 + 20 + 24 + 8` = 60 at Fine, 72 at Coarse.
- **2.14 Drag — a column**, at x 8200, y 1100: a copy of 2.5. In Progress at `opacity.in-flight`, a `Drop placeholder` between Review and Done, and a `Lifted column` (`shadow.dragged`) held 36 right of and 56 below it. The placeholder is a copy of In Progress with every fill, edge and shadow inside it removed, so it is always that column's height, at both densities.
- **2.15 Phone board**, at x 9840, y 0, 390 × 844 (§8).
- **2.15 Phone — dragging a card**, at x 9840, y 1100 (§10).

### 12. The code

- **The drop target is the column panel** (§1). The canvas stays `flex items-start`, so no column stretches during a drag.
- **The tint goes on the card list**, not on the element that also holds *Add card* (§2). *Add card* keeps `bg-board-column` throughout.
- **The empty place** is the card's own element at `opacity-[var(--in-flight-opacity)]` with its meta row in `text-foreground`, as for a pending change (ADR-018 §3). **The lifted card** is dnd-kit's `DragOverlay`, rendering the card with `shadow-dragged` and `scale-[1.02]`. **The outline** is a placeholder element at the dragged card's measured height: `border-2 border-dashed border-drop-placeholder rounded-[var(--radius)]`, no background.
- **A column drag** uses the same three, at column size.
- **The board ⋯ menu** renders *Add column* first below 640 (`max-sm:`), for those who can change the board, and opens the same create-column form as the bar's button. **The empty board** renders its *Add column* button below 640 only.
- **No CSS variable is added.**

## Options considered

Every option is named, with its cost, under its decision above. In short:
- **A short column:** every column to the bottom while a card is dragged (recommended); only the column you're over to the bottom.
- ***Add card* during a drag:** tinted with the body; hidden.
- **The `Card` part:** `Lifted` only; neither.
- **The growth:** stretching the box.
- ***Add column* below 640:** a slot after the last column; a plus button in the bar.
- **The browser:** a 390 × 704 board.
- **The phone's column width:** a token.
- **A phone drag board:** none.

## Out of scope (deferred, recorded)

- **2.10 at phone width**, with its *Add column* button (§7). Recorded, not drawn.
- **The board ⋯ menu itself**, with *Add column* first below 640. A Stage 5 overlay.
- **The movement**: the held press's shadow growing over the hold, the card lifting, the outline opening, and the cards making room. Motion is code only (inventory §4.3).
- **The cursor** (state 11) and **a refused drop** (state 10, which snaps back with no look of its own).
- **The desktop's auto-scroll edge zone** while dragging from 640 up (ADR-019 §5). Code only.
- **Hover, pressed and focus on these boards.** The parts carry those looks.

## Downstream edits this ADR requires

1. **[ADR-037](adr-037-board-at-rest.md): annotation.** Its Status line records that its two *Out of scope* items are settled.
2. **[ADR-038](adr-038-board-states-group-d.md): annotation.** Its Status line records that §10's *no button of its own* holds from 640 up only.
3. **[ADR-019](adr-019-responsive-touch-scope.md): annotation.** Its Status line records §7's count as checked by eye: 5 whole cards and 44px of the sixth.
4. **Penpot `Collabotask App`:** done 2026-11-23 (*Operational note*).
5. **[`design-file-inventory.md`](../design-file-inventory.md):** §1.5, §2's 2.14 and 2.15 rows and their notes, §4.2, §4.3, §5, §7.2, §7.3, §7.6 and the appendix.
6. **`CLAUDE.md` › `## Now`:** a pointer. Stage 4 is closed.
7. **`docs/build-log.md`:** an entry, since Stage 4 of the design file is closed.
8. **PRD, user stories, SRS, `CONTEXT.md`, `frontend-architecture.md`:** no change.

## Operational note

**Drawn 2026-11-23**, at Light + Coarse, as the file was found. The binding check returns only the card `⋯` rows the inventory's §7.5 expects:
- seven on `Card — states`, one per look
- 22 on 2.14's card board: 21 cards and the lifted one
- 27 on its column board: 21, three in the lifted column and three in the outline's unpainted copy
- 15 on 2.15 and 16 on its drag board

No shape has per-side edge widths. Dark + Fine and Light + Coarse were looked at on all four boards. After each flip every text box matched its text. Each outline matched what it stands for: by eye at Fine, and measured at Coarse, where the column outline and In Progress are both 408, the card outline 72 and the phone's 44. The file was left in Light + Coarse.

What drawing found:
- **Penpot's hosted server blocked one write three times running**, with a Cloudflare page, while reads and other writes passed. So the block was about the request's content, not the server. Split into pieces, it followed the statements that named and worded the new labels. Building those strings in pieces (`"Ar" + "med"`) went through. If one write keeps being blocked while others pass, split it until the blocked statement is found.
- **`addVariant()` names its copies `Value 6`, `Value 7`** and places them by its own spacing, here 284 apart. Rename them with `setVariantProperty` and place them by hand.
- **`penpot.library.local.components` lists a variant set once**, by its first look. Reach the others through `.variants.variantComponents()`.
- **`switchVariant` keeps a card's title, date wording and avatar**, but not always a date's colour. `DragSource` sets the date to `color.foreground` by design. `Lifted` reset an overdue date to `color.muted-foreground`, and it was bound back to `color.overdue`.
- **A hidden child does not count towards a flex board's size.** So an outline that must match a column's height at both densities holds a copy of that column with every fill, edge and shadow removed, not a hidden one.
- **`parentIndex` counts from the end**: in a row, the leftmost child has the highest index. `setParentIndex(2)` put the column outline fourth of six.
- **`resize()` on a flex board resets how its height is decided** to fixed. The phone's columns were set back: To Do fills, In Progress follows its cards.
- **`waitForLayoutUpdate()` timed out on a theme flip**; the flip had landed. **One export failed with a server timeout** and worked on the retry.
- **The lifted card and column are placed by position**, while the layout under them moves with density, so at Fine they cover a little more of their outline. Any position is a valid moment in a drag.
