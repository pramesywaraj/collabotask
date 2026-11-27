# ADR-048: The Board Page, Split in Four

- **Status:** Accepted
- **Date:** 2026-11-27
- **Scope:** Which page each of the Board area's boards lives on in the Penpot file `Collabotask App`, so that no page is much heavier than the ones that work smoothly. It settles:
  - the four pages that replace `Board`, and their order
  - which row goes on which page
  - where the phone's user menu now lives

  **No token is added (222), no part (34), and no board is added or renamed.** The file still holds 73 screen boards; 34 of them changed page, with three row labels. It **supersedes [ADR-045](adr-045-design-file-pages-by-area.md) §1's list of pages, §5's page for the phone's user menu and §7's `Board` table only**. A row per screen, the row labels, the spacing and the rule for adding a board (§3, §4, §6, §8) stand, and Account and Workspaces are unchanged.
- **Upstream edits:** none. The PRD, the user stories and the SRS name no design-file page.

## Context

**The user reported that Penpot often freezes when they work in the file by hand**, sometimes until it can't be used. ADR-045 had already noted that a theme switch on a page of 45 screens blocks the tab for minutes, and ADR-046 that opening `Board` froze it for several minutes.

Every board was counted on 2026-11-27, read-only, before deciding:

| Page | Shapes | Drawn (not hidden) | Texts | Token bindings |
|---|---|---|---|---|
| Foundations | 926 | 921 | 572 | 2,312 |
| Components | 4,441 | 3,257 | 1,314 | 16,227 |
| Account | 345 | 242 | 93 | 1,458 |
| Workspaces | 4,026 | 2,664 | 961 | 16,888 |
| **Board** | **17,420** | 11,830 | 4,928 | 63,603 |

**`Board` held 64% of the file's 27,158 shapes, 4.3 times Workspaces.** By row:

| Row (ADR-045 §7) | Boards | Shapes | Bindings |
|---|---|---|---|
| 2.5, then 2.8 to 2.12 | 7 | 1,529 | 6,119 |
| 2.5 — what opens over it | 13 | 6,370 | 23,686 |
| 2.6 Card drawer | 9 | 4,387 | 15,560 |
| 2.7 Read-only board | 2 | 906 | 3,397 |
| 2.14 Drag | 2 | 993 | 3,649 |
| 2.15 Phone board | 12 | 3,235 | 11,192 |

Two things the count corrected:
- **A phone board is not light.** It averages about 270 shapes, half a desktop board (440 to 630).
- **Shadows are not the bulk.** There are about 18 per desktop screen, 625 on the page. The mass is shapes and text.

**The user said Workspaces, at about 4,000 shapes, feels smooth.** That set the target: no page much heavier than Workspaces.

## Decision

### 1. Eight pages

**Foundations, Components, Account, Workspaces, Board, Board — what opens over it, Card, Board — phone**, in that order.

### 2. Which row goes where

No row is cut. Each moves whole, with its label.

| Page | Rows | Boards | Shapes | Bindings |
|---|---|---|---|---|
| **Board** | 2.5 and its states (2.8 to 2.12), 2.7, 2.14 | 11 | 3,428 | 13,165 |
| **Board — what opens over it** | 2.5's menus, moving, forms, lists and toasts | 13 | 6,370 | 23,686 |
| **Card** | 2.6, its pickers, its confirmation and toast, and the drawer at 390 | 9 | 4,387 | 15,561 |
| **Board — phone** | 2.15, with what opens over it | 12 | 3,235 | 11,192 |

The heaviest page is now 1.6 times Workspaces, where it was 4.3 times. `Card` has one binding more than its rows had, from §6's repair.

### 3. Where things sit on each page

ADR-045 §3 holds on every page. **Each page starts with its row label at 0, 0**, and its boards 160 below. On *Board — what opens over it* the second line starts at y 1260. On `Board` the three remaining rows are at y 0, 1300 and 2600, 400 apart. Every board keeps its x. The layers panel reads in row order, top to bottom, on all four.

Because every page's origin holds a row label, ADR-045 §8's warning applies to all four: `instance()` and `createShapeFromSvg` drop their shape into it.

### 4. What the split costs

- **2.5 and its 13 overlays are on two pages**, which is the kind of separation ADR-045 was written to undo. It is milder: one screen, the next page in the list, and the page carries 2.5's name. Keeping that row on `Board` would have left the page at 9,800 shapes, 2.4 times Workspaces (*Options considered*).
- **`Card` reverses ADR-045 §1's reason that Board covers `card/`** (because the drawer opens over the board), for the page only. The drawer is still drawn over a board. The page matches the `card/` feature folder in frontend-architecture.md.

### 5. The phone's user menu is on Board — phone

*2.15 + sidebar, user menu* moved with its row and its sequence (ADR-045 §5's reason). The note on Account's user-menu row now reads: *On a phone it opens over the sidebar: Board — phone page, row 2.15, “2.15 + sidebar, user menu”.*

### 6. Two repairs, found while checking

- **45 size bindings were lost on `Card` after the paste, and put back.** *2.6 Card drawer over the board* lost 40, *2.6 + delete card* 3 and *2.6 + assignee* 2, all `width` bindings: `size.sidebar-width`, `size.column-width`, `size.drawer-width`, `size.title-reserve`, `size.avatar`, `size.target-min`, `size.icon`. Every size was still right, so nothing changed on screen. Each was re-applied to the same property, and they still held on a second count.
- **An older fault on *2.6 + delete card*:** the sidebar row *Q3 Launch Plan* had no height binding and was **44 tall in Fine, where it should be 36**, which is the Coarse value it kept. It is now bound to `size.control-height` and 36 (the user's call). It predates this ADR: the §7.5 binding check does not look at sizes.

## Options considered

- **Leave `Board` whole.** Nothing moves, and the page stays 4.3 times Workspaces.
- **Only the 15 phone boards on a page of their own** (ADR-045 *Out of scope*). Least work, but phone boards are not light enough for that to be enough, and taking 2.6's three phone boards out of the drawer's row breaks ADR-045 §3.
- **Three pages, the overlays left on Board**: Board 9,798 shapes (24 boards), Card and Board — phone as here. Every row keeps its screen, but Board stays 2.4 times Workspaces.
- **Four pages.** **Chosen by the user**, once they confirmed that Workspaces feels smooth.
- **Separate files**: a library file for the parts and tokens, and screen files. Not tried. Whether Penpot shares tokens between files was not checked, and the tokens are what the file is for.
- **Draw the one-sided rules as 1px rectangles instead of inset shadows** ([ADR-036](adr-036-one-sided-rules-as-inset-shadows.md)'s alternative). Not done. The count showed shadows are a small share of the page.

## Out of scope (deferred, recorded)

- **A theme or density switch still rewrites the whole file**, about 100,000 bindings, whichever page is open. The split only makes the redraw afterwards smaller. Switch from `Foundations` (ADR-046 *Operational note*).
- **Faster ways to work that don't change the file**, discussed with the user and not decided here: close the Claude plugin when working by hand, use Chrome with graphics acceleration on, work zoomed in with focus mode (**F**), and Penpot 2.16's *WebGL rendering (Beta)* in the account's settings. The beta draws text with a new engine, and every note in these ADRs about the plugin measuring text was learned on the old renderer.
- **§7.5's binding check does not look at width and height.** That is how §6's 44-tall row went unseen through Stage 6. Not extended here.

## Downstream edits this ADR requires

1. **[ADR-045](adr-045-design-file-pages-by-area.md): annotation.** §1's page list, §5's page and §7's `Board` table are superseded. It is annotated, not edited.
2. **[`design-file-inventory.md`](../design-file-inventory.md):** the sources table, the sidebar's *Where and when*, §7.2's `Account` row and the `Board` row (now four rows), §7.6, and *Related documents*.
3. **`CLAUDE.md` › `## Now`** and **`docs/build-log.md`**.
4. **PRD, user stories, SRS, `CONTEXT.md`, frontend-architecture.md:** no change.

## Operational note

**Moved 2026-11-27**, in three rounds of ADR-045's routine, lightest first: 2.15 (13 with its label), 2.6 (10), then 2.5's overlays (14). In each round the plugin selected the row on `Board` by id, and the user pressed ⌘C, opened the new page and pressed ⌘V. The plugin compared every copy with its original (size within 1px, shapes, part copies, part links, bindings) and placed it. The user agreed before each delete of the originals.

**Checks.**
- **Before moving:** no board on `Board` held a part's main, every part copy was linked, and every board name was unique.
- **After:** 27,158 shapes in the file and 34 parts, as before. Every page's bindings matched their rows, apart from `Card` (§6). Nothing overlaps, and the layers read in row order.
- **Looked at:** one export per new page: *2.15 + sidebar, user menu*, *2.6 + due date*, *2.5 + members, a person’s ⋯* and *2.6 + delete card*.

**What moving found:**
- **Counting works from any page and is fast.** Shapes, bindings, part links and shadows read correctly on a page that isn't open, about 2 seconds for every 12 to 20 boards. Only text sizes need the page open (ADR-046).
- **A paste can lose size bindings after it has checked clean.** `Card` matched its originals straight after the paste and was 45 bindings short a few minutes later, after the copies were placed and one was exported. The other three pages did not lose any. The cause wasn't found. **Count every page again at the end**, not only straight after each paste.
- **After the third paste the plugin still reported `Board` as open.** The copies were checked from `Board`, the originals deleted there, and the copies placed once the user had opened the new page. Check `currentPage` before every write.
- **A row label measures 1px narrower on a new page** (365 → 364): its texts are measured again when the page is first drawn. Harmless.
- **Setting a `Doc /` text's `characters` kept both its bindings**, and its auto-width label grew to fit.
- **One export failed with a 500 timeout** and worked on retry.
