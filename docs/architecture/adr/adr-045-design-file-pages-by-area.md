# ADR-045: The Design File's Pages, by Area of the App

- **Status:** Accepted — **§1's page list, §5's page and §7's `Board` table superseded by [ADR-048](adr-048-board-page-split-in-four.md)** (2026-11-27): `Board` held 64% of the file's shapes and froze when worked by hand, so it is now four pages: Board, *Board — what opens over it*, Card and *Board — phone*, each row moved whole with its label. §3, §4, §6 and §8 stand.
- **Date:** 2026-11-25
- **Scope:** Where the screens and the overlays live in the Penpot file `Collabotask App`, so that a developer finds a screen and everything that opens over it in one place. It settles:
  - the pages and their order
  - which board goes on which page
  - how a page is laid out, and how a row is labelled
  - where the phone's user menu lives

  **No token is added (222), no part (34), and no board is added or renamed.** The file still holds 73 screen boards; 31 of them changed page. It **supersedes [ADR-040](adr-040-menus-and-the-confirmation.md) §1's `Overlays` page only**: the parts stay on `Components`, and each overlay is still drawn open over a copy of its screen. It is a reorganisation made **before Stage 6** (inventory §7.5–§7.6), so that the final check reviews the final layout.
- **Upstream edits:** none. The PRD, the user stories and the SRS name no design-file page.

## Context

ADR-040 §1 put every screen with something open on a new page, `Overlays`, and left the fifteen screens and their states on `Page 1`. By the end of Stage 5 that was 26 boards on `Page 1` and 47 on `Overlays`, each page in the order its boards were drawn. *2.4 + workspace ⋯* sat at the top left of `Overlays`, beside *2.4 + user menu*, while 2.4 itself sat 4920 to the right on the other page. **The user's worry: a developer reading the file can't tell which flow to implement**, because a screen and what opens over it are drawn far apart.

Three ways were weighed (*Options considered*). The user chose to **split by area of the app**, for three reasons:
- a screen and its overlays drawn far apart is what confuses
- the areas roughly match the code's feature folders: frontend-architecture.md has `auth`, `workspace`, `board` and `card`
- two of the three pages stay light. That matters, because a theme switch on a page of 45 or more screens blocks the tab for minutes (ADR-044 *Operational note*)

## Decision

### 1. Five pages: Foundations, Components, Account, Workspaces, Board

- **Foundations** and **Components** are unchanged.
- **Account**: signing in, registering, a session that ended, and the user menu.
- **Workspaces**: 2.3 and 2.4, with their states and everything that opens over them.
- **Board**: 2.5 to 2.15, with their states and everything that opens over them, the card drawer included. The pages are named by area, not by folder, so Board covers both `board/` and `card/`: the card drawer opens over the board.
- *Account* was the user's pick over *Sign in*, which doesn't cover Register, and over *Auth*. In this app *auth* also reads as permissions: 2.7 read-only, 2.8 join gate and private boards, all of which live on Board or Workspaces.

  > *2026-11-27:* the Board area is now **four pages**: Board (2.5 and its states, 2.7, 2.14), *Board — what opens over it*, Card (2.6) and *Board — phone* (2.15). See [ADR-048](adr-048-board-page-split-in-four.md) §1–§2.

### 2. Two pages renamed, one created

`Page 1` became **Workspaces** and `Overlays` became **Board**, and **Account** was created. Renaming instead of emptying meant that 31 of the 73 boards moved, not all of them:

| From → to | Boards |
|---|---|
| `Page 1` → Account | 2.1 Login, 2.2 Register, 2.13 Session expired |
| `Overlays` → Account | *2.4 + user menu* |
| `Page 1` → Board | the 13 board screens: 2.5 to 2.11, both 2.12s, both 2.14s and both 2.15s |
| `Overlays` → Workspaces | the 14 workspace overlays |

The ten workspace screens stayed where they were, now on Workspaces, and the other 32 overlays stayed on what is now Board.

### 3. One row per screen

- **A row holds a screen at rest, then its other states, then everything opened over it.** It ends with the phone versions of its desktop screen, at 390. 2.15 is a screen of its own, so it has a row of its own.
- **A long row wraps under itself at a natural break**: where its phone versions start (2.4's overlays), or between two kinds of overlay (2.5's: menus and forms above, lists and toasts below). No line is wider than eight desktop boards.
- In a line the boards are **200 apart**. The lines of one row are 200 apart too, so 1100 from top to top. **Rows are 400 apart**, and the row's label sits in that space. Every row starts at x 0.
- **The layers panel reads in the same order**: top to bottom, each row's label, then its boards from left to right.
- **2.7 has its own row**, with *2.7 + an edit you can’t make* beside it. The handoff put 2.7 among 2.5's states and that overlay "beside 2.7". Both could not hold without breaking *states, then what opens over it*. So 2.5's row keeps the states that take the whole board's place (2.8 to 2.12). 2.7 is the board with its banner, and it has an overlay of its own, so it reads as a screen row.

### 4. A row's label is a small `Doc /` board

- Each row starts with a board named `Doc / Row label / <heading>`, at the row's top left, 160 above its boards. It is a flex column filled `color.background`. It holds `Doc / Heading` in `type.heading` and `color.foreground`, and, where the row needs one, a one-line `Doc / Note` in `type.meta` and `color.muted-foreground`.
- **It is a board, not a bare text.** A text bound to `color.foreground` turns near-white in Dark, and it would sit on Penpot's grey canvas, which no theme changes. The Foundations board's section heads stay readable the same way: they sit on a filled board.
- Its padding (16 above and below, 24 at the sides) and its gap (4) are typed. That is documentation geometry, which `Doc /` shapes may type (inventory §7.5). Its fill and its type are bound, so the binding check returns nothing on any label.
- `type.heading`, 28, is the largest step the catalogue has. When you zoom out to see a whole row, the label is small, and Penpot's own board names do the job instead.

### 5. The phone's user menu stays on Board

*2.15 + sidebar, user menu* is the third of three boards that show the phone's sidebar step by step, after *2.15 + sidebar* and *2.15 + sidebar, switching workspace*. Moving it alone would break that sequence. Profile and *Sign out* live in the user menu, so the Account page's user-menu row carries a note pointing to it: *On a phone it opens over the sidebar: Board page, row 2.15, “2.15 + sidebar, user menu”.*

> *2026-11-27:* the row and its sequence moved together to **Board — phone**, and the note now names that page ([ADR-048](adr-048-board-page-split-in-four.md) §5).

### 6. Board names don't change

The ADRs refer to boards by name, so every board keeps its name. Only its page and its position changed.

### 7. What the file holds

In a line, a desktop board starts 1640 after the one before it (1440 + 200), and a phone board 590 after (390 + 200).

**Account**: 4 boards in 4 rows, each a single board at x 0.

| Row label (y) | Note | Boards (y) |
|---|---|---|
| 2.1 Login (0) | — | 2.1 Login (160) |
| 2.2 Register (1300) | — | 2.2 Register (1460) |
| 2.13 Session expired (2600) | *Over the page you were on; Sign in brings you back to it.* | 2.13 Session expired (2760) |
| User menu — profile and Sign out (3900) | §5's note | *2.4 + user menu* (4060) |

**Workspaces**: 24 boards in 3 rows.

| Row label (y) | Note | Boards (y) |
|---|---|---|
| 2.3 Workspaces (0) | *At rest and three states, then what opens over it.* | 160: 2.3 Workspaces, — cold load, — empty, — load failed, *2.3 + create workspace*, *2.3 + removed from a workspace* |
| 2.4 Workspace detail (1300) | *At rest and five states.* | 1460: 2.4 Workspace detail, — archived, — cold load, — empty, — archived, empty, — load failed |
| 2.4 Workspace detail — what opens over it (2600) | *The workspace ⋯, forms, lists and a toast; below, the same at 390.* | 2760: *2.4 + workspace ⋯*, *+ create board*, *+ create board, a short title*, *+ invite people*, *+ invite people, sent*, *+ members*, *+ removing a member*, *+ archived, with Undo*<br>3860: *Create a board at 390*, *Invite people at 390*, *Members at 390*, *Members at 390 + a person’s ⋯* |

**Board**: 45 boards in 6 rows.

| Row label (y) | Note | Boards (y) |
|---|---|---|
| 2.5 Board (0) | *At rest, then the states that take the whole board’s place.* | 160: 2.5 Board — at rest, 2.8 Join gate, 2.9 Cold load, 2.10 Empty board, 2.11 Load failed, 2.12 Gone — archived, 2.12 Gone — not found |
| 2.5 Board — what opens over it (1300) | *Menus, moving and forms; below, lists and toasts.* | 1460: *2.5 + board ⋯*, *+ column ⋯*, *+ renaming a column*, *+ card ⋯*, *+ moving from the card ⋯*, *+ board settings*, *+ adding a column*<br>2560: *2.5 + members*, *+ members, a person’s ⋯*, *+ make owner*, *+ who’s here*, *+ a move that didn’t stick*, *+ the open card deleted* |
| 2.6 Card drawer (3700) | *At rest, its pickers, a confirmation and a toast, then the drawer at 390.* | 3860: 2.6 Card drawer over the board, *2.6 + assignee*, *+ due date*, *+ moving, a column*, *+ delete card*, *+ a change that didn’t save*, then from x 9840 *2.6 at 390*, *2.6 at 390 + due date*, *2.6 at 390 + moving, a place* |
| 2.7 Read-only board (5000) | *At rest, then the toast for an edit you can’t make.* | 5160: 2.7 Read-only board, *2.7 + an edit you can’t make* |
| 2.14 Drag (6300) | *A card, then a column.* | 6460: 2.14 Drag — a card, 2.14 Drag — a column |
| 2.15 Phone board (7600) | *At rest and dragging a card, then what opens over it.* | 7760: 2.15 Phone board, 2.15 Phone — dragging a card, *2.15 + board ⋯*, *+ delete column*, *+ adding a column*, *+ moving a card*, *Board members at 390*, *2.15 + who’s here*, *+ a move that didn’t stick*, *+ sidebar*, *+ sidebar, switching workspace*, *+ sidebar, user menu* |

> *2026-11-27:* the `Board` table above is superseded. Its rows are now on four pages, each starting at 0, 0 with its label: see [ADR-048](adr-048-board-page-split-in-four.md) §2–§3, and inventory §7.2 for today's positions.

### 8. Adding a board later

- A new state or overlay goes in its screen's row, in §3's order, and a phone version at the row's end. Move the boards after it along; the label stays where it is.
- A new screen gets a new row on its area's page, with a label made as in §4.
- **The board at each page's canvas origin is now a row label.** `instance()` and `createShapeFromSvg` drop their shape into whatever board sits at the origin (ADR-040 and ADR-044, *Operational note*). On these three pages that is `Doc / Row label / 2.1 Login`, `… / 2.3 Workspaces` or `… / 2.5 Board`. Move the new shape out in the same call.

## Options considered

- **Everything on one page**, `Page 1` and `Overlays` merged. A screen and its overlays could sit side by side, but at 73 boards on one page, and a theme switch that already takes minutes on 47 would take longer.
- **Leave the pages as they are and add a map**: a `Doc /` index board pointing from each screen to its overlays. Nothing moves, so nothing can break, but the screen and what opens over it stay far apart, which was the problem.
- **Split by area** (chosen).
- **Naming the first page** *Sign in* or *Auth* (§1).
- **A bare `Doc /` text as a row's label**, which vanishes in Dark (§4).

## Out of scope (deferred, recorded)

- **Board's 15 phone boards on a page of their own.** It would make Board's theme switch quicker. The user knows about the option, and it is not done now.
- **Links between boards**, Penpot's prototype flows. A row reads in order; nothing in it is clickable.
- **Stage 6 itself**: the binding check on every board, and the two flips page by page.

## Downstream edits this ADR requires

1. **[ADR-040](adr-040-menus-and-the-confirmation.md): annotation.** §1's `Overlays` page is superseded. Parts on `Components` and overlays drawn open over a screen stand.
2. **ADR-032 to ADR-039 and ADR-041 to ADR-044: no annotation.** Their `Page 1` and `Overlays` positions record where each board was drawn at the time, and they stay true as history. To find a board now, look in inventory §7.2, not in the ADR that drew it.
3. **[`design-file-inventory.md`](../design-file-inventory.md):** the sources table; §7.2, where the screens row and the `Overlays` row become the three pages; §7.6; and *Related documents*.
4. **`CLAUDE.md` › `## Now`:** a pointer, and the line on where overlays are drawn.
5. **`docs/build-log.md`:** a line in the Stage 5 entry.
6. **PRD, user stories, SRS, `CONTEXT.md`, frontend-architecture.md:** no change.

## Operational note

**Moved 2026-11-25**, in four rounds of copy and paste, because the plugin cannot copy or move a shape to another page (ADR-040 *Operational note*). In each round the plugin selected the boards by id. The user pressed ⌘C, opened the target page and pressed ⌘V. The plugin checked the copies, and the user agreed before each delete of the originals.
- R1, `Page 1` → Account: 2.1, 2.2, 2.13
- R2, `Page 1` → Board: the 13 board screens
- R3, Board → Workspaces: the 14 workspace overlays
- R4, Board → Account: *2.4 + user menu*

**Checks.**
- **Before moving:** no board on `Page 1` or `Overlays` held a part's main, so deleting an original could not delete a part.
- **Each copy against its original:** the same size, the same number of shapes, the same number of part copies, every copy still linked and to the same part, and the same number of token bindings. All 31 matched, from 2.13 (11 shapes, 2 part copies, 47 bindings) to *2.14 Drag — a column* (531, 97, 1,939). No part was duplicated: the library still holds 34.
- **The counts after:** Foundations 1, Components 34, Account 4 boards and 4 labels, Workspaces 24 and 3, Board 45 and 6.
- **The layout:** on Board, the boards in every line are exactly 200 apart, and none overlap.
- **The binding check** on the 13 labels: nothing.
- **The looks:** 2.5 Board — at rest and *2.15 + sidebar, user menu* were exported after the move and compared with *2.5 + board ⋯*, which never moved; they draw the same. No theme was switched, so no copy's text size can have gone stale (ADR-043 *Operational note*). The file is in Light + Coarse, as ADR-044 left it.

**What moving found:**
- **The plugin can set the selection**, on the page that is open: `penpot.selection = shapes` selected the boards for the user's ⌘C in every round, so no drag-select was needed.
- **A pasted copy keeps every part link and every token binding.** The copies landed 451 to the right of and 510 below their originals; the offset did not matter, since every board was then placed.
- **The component checks are `isComponentMainInstance()`, `isComponentCopyInstance()` and `isComponentHead()`.** The first check for mains was written as `s.isMainComponent && s.isMainComponent()`. No such method exists, so the guard made it match nothing, and it looked like a clean result. Call the method without a guard, so that a wrong name throws.
- **The plugin's `storage` can be lost on a reconnect**, although it survived one in ADR-044. Keep the ids that matter in the conversation too.
- **One call over Board's 45 boards can time out and still land.** Placing them and making the six labels timed out, then the tab lost its heartbeat, but everything had been placed. Check what landed before running such a call again. The layers were then reordered in three calls of 13 to 19 boards each.
- **The plugin cannot reorder pages.** The user dragged them.
- **`export_shape` with `page` fails** on these pages, with a screenshot timeout. Export the boards one at a time.
