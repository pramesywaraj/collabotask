# ADR-040: The Menus and the Confirmation — Stage 5, Group a

- **Status:** Accepted
- **Date:** 2026-11-23
- **Scope:** Question 4 for Stage 5's group a: the workspace ⋯, the board ⋯ (with *Add column* first below 640), the column ⋯ and its inline rename, the card ⋯, the user menu, and the destructive confirmation (state 38) they open. It settles:
  - the menu as a popover from 640 up and as a bottom action sheet below 640, once, for every later group
  - the centred confirmation dialog, once, for every later group
  - which actions ask first, which are rose, and what the rose button looks like in every state
  - typing a workspace's name to delete it
  - where inline rename sits in the column header
  - every wording, and where each action leaves you
  - where overlays live in the design file

  **Ten token entries are added (210 → 220):** `spacing.menu-padding`, `size.menu-width`, `spacing.dialog-padding` and `size.dialog-width` in `base`, and `color.destructive-hover`, `color.destructive-subtle` and `color.destructive-subtle-foreground` in both theme sets. **Four parts are added (19 → 23):** `Glyph`, `Menu row`, `Menu` and `Confirmation`. The `Button` gains `Kind` = Destructive, six looks, so it holds 17. A new page, **`Overlays`**, holds nine boards, each a screen with one overlay open.

  It **narrows [ADR-018](adr-018-state-vocabulary.md) twice**: §5's busy look is pale rose on the rose button (§8 below), and §10's row 38, *Rose confirm button*, holds for the deletes and *Leave board* but not for *Archive board*, whose confirm is plum (§3). It **supersedes [ADR-030](adr-030-toast.md) §3's *no action button* for one toast**, the one after you archive a board, which keeps US-12c's *Undo* (§5); group e draws it. It **widens [ADR-021](adr-021-design-inventory-corrections.md) §5's confirmations** from three to five. It **extends [ADR-024](adr-024-control-and-field-internals.md) §3** to the rose button, and **[ADR-017](adr-017-ui-foundations.md) §9** with the variables in §15.
- **Upstream edits:** one. [US-02b](../../product/001-user-stories.md)'s note says *Sign out* in place of *Log out* (§11). Nothing here touches an endpoint, the schema or a permission rule.

## Context

Stage 5 draws about eighteen overlays in five groups, each settling its kind of container once (inventory §7.3). Group a comes first because every later group opens a menu or a confirmation.

**Question 4 was asked of every value group a needs, against the live 210 tokens, before anything was drawn.** The colours, corners, shadows, type and row sizes were all there. The menu's rows are the Sidebar row's geometry filled with the popover's colour, as the Field is. The dimming is `color.scrim`, which ADR-034 §10.8 had already set aside for Stage 5's dialogs. Two measurements turned into rules rather than choices:
- **a menu needs an edge.** In Dark the popover is the column's own colour (1.00:1), and in Light it is the card's (1.00:1)
- **so does a dialog.** In Dark it is 1.09 to 1.17:1 against the dimmed screen

**Six values had no token**, so question 4 fired for the eighteenth stage out of eighteen: the space inside a menu, its width, the dialog's width and the space inside it, and the rose button's hover and busy fills. **Fourteen choices had never been made**, and one of them was a conflict: US-12c promises an *Undo* after archiving, which ADR-030 §3 had ruled out.

All fourteen went on one rendered page, drawn in Light and Dark with the file's values, at Fine on the desktop and at Coarse on a 390 phone wherever a phone shows it differently. **The user took twelve recommendations, chose a pale rose busy look over the shared plum one (A09), and added a note to A04: *Leave board* is rose too.**

One pick could not be carried out by the plugin alone: Penpot's plugin cannot copy or move a shape to another page. The user pasted the four screens onto `Overlays` by hand, and everything else was drawn there (§1).

## Decision

### 1. Parts on `Components`; each overlay drawn open over a screen on a new page, `Overlays`

- **The parts go on `Components`**, as every part has: `Glyph`, `Menu row`, `Menu` and `Confirmation`, each on its own holding board, and the Button's rose row on `Button — states`.
- **A new page, `Overlays`**, holds copies of the screens with one overlay open on each. Later groups add theirs to the same page.

| Board | Screen | Open on it |
|---|---|---|
| *2.4 + workspace ⋯* | 2.4, the owner's view | the workspace ⋯ menu |
| *2.4 + user menu* | 2.4 | the user menu, opened from the sidebar |
| *2.5 + board ⋯* | 2.5 | the board ⋯ menu, for an admin who is on the board |
| *2.5 + column ⋯* | 2.5 | In Progress's ⋯ menu |
| *2.5 + renaming a column* | 2.5 | In Progress being renamed |
| *2.5 + card ⋯* | 2.5 | the ⋯ menu of *Send the press kit to the three trade outlets* |
| *2.6 + delete card* | 2.6 | the delete confirmation, over the drawer |
| *2.15 + board ⋯* | 2.15 | the board ⋯ as a bottom sheet, *Add column* first |
| *2.15 + delete column* | 2.15 | the delete-column confirmation, 358 wide |

The rejected options: parts only, each over a small `Doc /` strip, which never shows a menu in its real place; and the same nine boards on `Page 1`, which would mix screens with screens-plus-an-overlay.

### 2. The menu from 640 up

- **One surface.** `color.popover`, a 1px `color.border` edge drawn outside its box (`border.width`), `shadow.overlay`, `radius.surface` corners. The edge is the measured rule above, not a choice.
- **`spacing.menu-padding` 4 (new, `base`) inside, worked out rather than chosen:** `spacing.screen-x` 16 − `spacing.control-padding-x` 12. On a phone the sheet spans the screen, and its rows' words then start 16 from the edge, where the board's columns start.
- **`size.menu-width` 224 (new, `base`), the same for every menu (A02).** It is the sidebar row's width, so the user menu sits exactly over the row it opens from. The board ⋯ keeps one width whoever is looking, though its entries change.
- **Placement: 4 below the trigger, right edges lined up.** The user menu, opened from the sidebar's bottom row, opens upward with its left edge on the row's. The 4 is a placement in code (Radix's `sideOffset`) and is drawn by position. It is not a token, as 2.15's 318 is not.
- **Rows** are the `Menu row` part: the Sidebar row's geometry (`size.control-height`, `spacing.control-padding-x`, `spacing.icon-gap`, `radius.control`, `type.body`) filled `color.popover`, `spacing.list-gap` apart. No edge at rest; hover and pressed step the edge ladder; focus adds `shadow.focus-ring` (ADR-018 §6).
- **A Phosphor glyph before every entry (A03)**, Regular at 16 in `color.muted-foreground`. They are furniture, not the icon channel (ADR-018 §9): the lock stays its only meaning.

| Entry | Glyph |
|---|---|
| *Add column* | `Plus` |
| *Board settings…* | `Gear` |
| *Members…* | `Users` |
| *Archive board* | `Archive` |
| *Leave board* | `DoorOpen` |
| *Rename* | `PencilSimple` |
| *Move left* · *Move right* | `ArrowLeft` · `ArrowRight` |
| *Delete column* · *Delete card* · *Delete workspace* | `Trash` |
| *Open card* | `ArrowsOut` |
| *Move card to…* | `Kanban`, the drawer's *Column* glyph, since both open the same picker |
| *Sign out* | `SignOut` |
| the sheet's Close | `X`, in `color.foreground` as on the drawer |

- **No rules between groups of entries (A12).** The confirmations stand between a slip and its result, and rose marks the entries that lose something.
- **Entries you can't use are absent** (ADR-018 §5). The board ⋯ has five shapes: the owner sees Board settings, Members and Archive; an admin on the board sees all four; a plain member sees Members and Leave; an admin who hasn't joined sees Board settings, Members and Archive; a visitor on 2.7 sees Members only. Below 640, *Add column* comes first for everyone who may add one (ADR-039 §7). The leftmost column has no *Move left* and the rightmost no *Move right*. A read-only viewer sees no column ⋯ and no card ⋯, as 2.7 is drawn.
- **An open menu's trigger has no look of its own.** *Selected* (state 18) does not exist in Phase 1.

### 3. What is rose: the three deletes and *Leave board*

- ***Delete column*, *Delete card*, *Delete workspace* and *Leave board* are rose**: the row's words and glyph in `color.destructive-text` (the drawer's *Delete card*, ADR-038 §5), and the confirm button rose (§8). **That was the recommendation plus the user's note**, which added *Leave board*.
- ***Archive board* is an ordinary row, and its confirm button is Primary.** Archiving can be undone (§5).
- **So rose means *something is lost for good*.** Deleting loses the thing. Leaving clears your assignments on the board (UC-10), which joining again does not bring back, and on a private board someone has to add you back.

The rejected option: all five rose, which puts rose on something that can be undone.

### 4. All five actions ask first

The three deletes, *Archive board* and *Leave board* each open the confirmation (A05). Archive takes the board away from everyone who has it open, and they are sent to 2.12; your Undo does not bring them back. Leave clears your assignments.

The rejected option: Archive acting at once, with its Undo only.

### 5. Archive keeps US-12c's *Undo*

- **The user story stands, since it is upstream.** After you archive the board you're on, you land on the workspace with a message that offers *Undo* (US-12c, ADR-021 §1).
- **ADR-030 §3's *no action button* is superseded for that one toast.** ADR-030 wrote *Phase 1 has no undo* without US-12c in view. Every other toast is unchanged.
- **Group e draws it**, with the toast's placement and its other wordings.
- The Archive dialog says *You can bring it back from the Archived list*, which stays true after the toast has gone.

The rejected option: dropping Undo and editing US-12c, which loses ADR-021's way of showing that the Archived list exists.

### 6. You type the name to delete a workspace, and only then

- ***Delete workspace* asks you to type the workspace's name (A07).** A field under the sentence, *Type “Product Team” to confirm*, is the Input part. The rose button stays in its *not satisfied* look (state 16, ADR-018 §5) until the name matches. This is the first use of that look outside Login and Register.
- **Matching:** spaces at either end are ignored, and capitals must match. The button turns rose the moment it matches.
- No other action asks for typing: deleting a column is ordinary tidying.

### 7. The confirmation dialog

- **`size.dialog-width` 448 (new, `base`), written as `{size.form-width} + {spacing.dialog-padding} * 2`**, with **`spacing.dialog-padding` 24 (new, `base`)** inside (A08). A 400 form fits a dialog exactly, which group c's forms will use. At 390 the dialog narrows to 358, the screen less `spacing.screen-x` each side, and is otherwise unchanged (ADR-021 §5).
- **The surface is the menu's**: `color.popover`, the edge outside its box, `shadow.overlay`, `radius.surface`, over `color.scrim`. Opened from inside the card drawer, it dims the drawer too, with a second scrim.
- **Inside:** the title in `type.title` and `color.foreground`, then the sentence in `type.body` and `color.muted-foreground`, `spacing.field-inner-gap` apart, as the drawer's *Description* and its text (ADR-038 §4). Then the field, if any, and the buttons, `spacing.section-gap` apart.
- **The buttons** sit at the right, `spacing.inline-gap` apart: an Outline *Cancel*, then the confirm. **Focus opens on *Cancel***, so a stray Enter cancels.
- **One part, five wordings** (inventory §3): `Confirmation` has no axis. Its copies change the words, show the field, and switch the confirm button's look.

The rejected options: 432 with 16 inside, which still needs its own token (a dialog is never the screen, ADR-022 §4); and shadcn's 512.

### 8. The rose button: the Button's third kind

| `State` | Fill | Label | |
|---|---|---|---|
| Rest | `color.destructive` | `color.destructive-foreground` | 4.70:1 Light, 6.50:1 Dark |
| Hover, Pressed | **`color.destructive-hover`** (new): `#BE123C` / `#FDA4AF` | `color.destructive-foreground` | 6.29:1, 9.25:1 |
| FocusVisible | `color.destructive` + `shadow.focus-ring` | | the ring is plum, as on every control |
| Busy | **`color.destructive-subtle`** (new): `#FFF1F2` / `#4C0519` | **`color.destructive-subtle-foreground`** (new): `#9F1239` / `#FECDD3`, with the spinner | 7.30:1, 11.08:1 |
| NotSatisfied | Primary's: `color.muted`, a `color.border` edge | `color.muted-foreground` | |

- **Hover is one step darker in Light and lighter in Dark**, as `color.primary-hover` is. **Pressed reuses it**, ADR-024 §3's rule for a filled button.
- **Busy is pale rose, the user's call (A09)** over the plum busy look every other button shares. The button keeps its colour while it waits. That narrows ADR-018 §5, whose busy is `--primary-subtle`, for this kind only.

### 9. Below 640: the bottom action sheet

- **The same surface, full width, docked above the browser's bottom bar** (ADR-021 §5), over `color.scrim`. Its top corners are `radius.surface`, and its edge is `shadow.rule-top` on an inner `Panel`, since the outer box binds `shadow.overlay`. In Dark it is 1.13:1 against the dimmed board without that edge.
- **Rows are 44 tall at Coarse**, their words 16 from the screen's edge (§2).
- **It has a head (A11):** the name of what it acts on (the board, column, card or workspace), one line in `type.body` and `color.muted-foreground`, and a Close button (Phosphor `X` in a `size.target-min` box). The user menu's head is you: your face, name and email. Tapping the dimmed screen also closes it.

### 10. Inline rename: the field takes the whole header

- ***Rename* turns In Progress's header into a field the header's own size (A10).** The field is the Input part's *InlineEditing* look: the `color.input` edge, no label (state 37). Its box is `size.control-height`, which is `size.column-header` at both densities (36 / 44), and its words start 12 in, where the title sat. So the name doesn't move. The count and the ⋯ step aside until you finish.
- **The name stays in `type.body`.** *InlineEditing* draws the drawer's 20/28 title, so on the board the field's text is changed to 14/20.
- **The Column header part is not given a *Renaming* look.** It has over a hundred copies, and ADR-027 §5's one appearance stands. The renaming header is drawn as the Input in the header's place, on *2.5 + renaming a column*, detached so its 14/20 stays (*Operational note*).
- **Behaviour:** rename starts from the menu only. From 640 up the header is the column's drag handle, so pressing the title starts nothing new. Enter or leaving the field saves; Escape puts the old name back; an empty name puts the old name back, with no error. The field stops at 255 characters, the server's limit (UC-13b).

The rejected option: a field in the title's place only, which moves the name 12px right.

### 11. The user menu

- **You, then one action:** an Avatar at `size.avatar`, your name in `type.card-title` and your email in `type.meta` and `color.muted-foreground`, in a row `size.control-height` tall. Then ***Sign out*** (A13), which pairs with ADR-032's *Sign in*. US-02b's note is edited to match. A long email ends in an ellipsis in code; the file uses a short one, since Penpot cannot draw that.
- **From the sidebar's bottom row** (2.4 and the board) it opens upward, its left edge on the row's. **From 2.3's app bar** it opens downward, right edges lined up with the avatar's box. Only the first is drawn; both use the same part.

### 12. The wordings

| Menu | Its ⋯, as a screen reader names it | Entries, in order |
|---|---|---|
| Workspace ⋯ | *Workspace actions* | *Delete workspace* |
| Board ⋯ | *Board actions* | *Add column* (below 640 only) · *Board settings…* · *Members…* · *Archive board* · *Leave board* |
| Column ⋯ | *Actions for “In Progress”* | *Rename* · *Move left* · *Move right* · *Delete column* |
| Card ⋯ | *Actions for “Final pricing sign-off”* | *Open card* · *Move card to…* · *Delete card* |
| User menu | *Account* | your face, name and email · *Sign out* |
| Sheet head | — | the name of what it acts on · a Close button named *Close* |

| Confirmation | Title | Sentence | Button · while it waits |
|---|---|---|---|
| Delete column | *Delete “Review”?* | *This deletes the column and its 4 cards for everyone on the board. You can’t undo this.* With no cards: *This deletes the column for everyone on the board. You can’t undo this.* | *Delete column* · closes at once |
| Delete card | *Delete “Final pricing sign-off”?* | *This deletes the card for everyone on the board. You can’t undo this.* | *Delete card* · closes at once |
| Delete workspace | *Delete “Product Team”?* | *This deletes the workspace and its 6 boards, with every column and card on them, for all 5 members. You can’t undo this.* Field: *Type “Product Team” to confirm* | *Delete workspace* · *Deleting…* |
| Archive board | *Archive “Q3 Launch Plan”?* | *It leaves the board list for everyone, and anyone who has it open is taken out of it. You can bring it back from the Archived list.* | *Archive board* · *Archiving…* |
| Leave board | *Leave “Q3 Launch Plan”?* | *You’ll be unassigned from your 2 cards on it.* With none: *You can join it again from the workspace.* On a private board, either one adds *It’s private, so someone will need to add you back.* in place of joining again | *Leave board* · *Leaving…* |
| Every one | | | *Cancel*, an Outline button, before the confirm |

### 13. The behaviour rules

- **Cards and columns are deleted at once** after you confirm: the dialog closes, and the card or column goes 60% while the server answers, as every board change does (ADR-018 §3, ADR-012).
- **Delete workspace, Archive and Leave wait for the server.** The button goes Busy, then:
  - after *Delete workspace*: 2.3, with no message, since its tile is gone. The remembered workspace is cleared if it named this one
  - after *Archive board*: the workspace, with the Undo message (§5)
  - after *Leave board*: the workspace, with no message. ADR-030 shows nothing when you leave by choice
- **Names in a title** are quoted and wrap to two lines at most, then end in an ellipsis. A card title can be 500 characters.
- **Counts in a sentence come from what the screen already holds**: the column's cards, the workspace's boards and members, your assigned cards on the board. No request is made to word a dialog.
- **A menu closes** when you pick an entry, press Escape, or press outside it. The arrow keys move between rows and Enter picks one. Escape closes a menu, a sheet or a dialog, and focus returns to what opened it.

### 14. What the file draws

**Tokens (210 → 220):**
- in `base`: `spacing.menu-padding` 4, `size.menu-width` 224, `spacing.dialog-padding` 24, and `size.dialog-width` stored as `{size.form-width} + {spacing.dialog-padding} * 2`, which resolves to 448
- in `theme-light` and `theme-dark`: `color.destructive-hover`, `color.destructive-subtle` and `color.destructive-subtle-foreground`

**Parts, on `Components` (19 → 23):**
- **`Glyph`**, on `Glyph — states`. `Name` × `Tone`, fifteen looks: twelve Muted (`Plus`, `Gear`, `Users`, `Archive`, `DoorOpen`, `PencilSimple`, `ArrowLeft`, `ArrowRight`, `Trash`, `ArrowsOut`, `Kanban`, `SignOut`), `Trash` and `DoorOpen` Destructive, and `X` Foreground. Each is a 16 board (`size.icon`) holding Phosphor's Regular outline, scaled from its 256 grid, filled `color.muted-foreground`, `color.destructive-text` or `color.foreground`. It is the first glyph part. The older parts keep their own glyph paths.
- **`Menu row`**, on `Menu row — states`. `State` (Rest · Hover · Pressed · FocusVisible) × `Tone` (Default · Destructive), eight looks. A 16 `Slot` holds a `Glyph`; a copy switches the glyph by its `Name`.
- **`Menu`**, on `Menu — states`. `Of` (Workspace · Board · Column · Card · User) × `Presentation` (Popover · Sheet), ten looks. The rows are `Menu row` copies inside a `Panel`; the sheet's `Panel` carries `shadow.rule-top` and a `Head`.
- **`Confirmation`**, on `Confirmation — states`. One part, no axis, worded for *Delete column*, with five copies beside it: *Delete card*, *Delete workspace* empty (*not satisfied*) and typed (*Deleting…*), *Archive board* (Primary) and *Leave board*.
- **`Button`** gains `Kind` = Destructive: six looks on a third row of `Button — states`, labelled *Delete* and *Deleting…*. It now holds 17.

**The `Overlays` page**: the nine boards of §1. The four screens were pasted by the user; five are clones of those.

**The Foundations catalogue** gains seven rows (four in `base`, three in the theme table, swatches bound) and reads *220 tokens*, *base · 52*, *67 each*, as of 2026-11-23.

### 15. The code

- **§9 gains** `--menu-padding: 4px`, `--menu-width: 224px`, `--dialog-padding: 24px`, `--dialog-width: calc(var(--form-width) + 2 * var(--dialog-padding))`, and `--destructive-hover`, `--destructive-subtle` and `--destructive-subtle-foreground` in both modes.
- **Menus from 640 up** are shadcn's `DropdownMenu`, restyled: `w-[var(--menu-width)]`, `p-[var(--menu-padding)]`, `gap-[var(--list-gap)]`, `bg-popover`, `border`, `shadow` from `shadow.overlay`, `rounded-lg`. Items are restyled to the edge ladder, not shadcn's `focus:bg-accent` (ADR-018 §6). Use `align="end"` and `sideOffset={4}`; the user menu in the sidebar uses `side="top"` and `align="start"`.
- **Below 640 the same menu renders as a bottom `Sheet`**, `side="bottom"`, with the head row and the overlay in `--scrim`. Pick the component by the 640 breakpoint (ADR-019), not by pointer type.
- **Confirmations are shadcn's `AlertDialog`**: `w-[min(var(--dialog-width),calc(100vw-2*var(--screen-x)))]`, `p-[var(--dialog-padding)]`, the overlay in `--scrim`, and initial focus on *Cancel*. One component takes the wording, the confirm kind and the optional name field.
- **The Button gains a `destructive` variant restyled to §8**, including its busy look. shadcn's `destructive` uses `bg-destructive/90` on hover, so replace that with `--destructive-hover`.
- **Inline rename** is a plain input in the header's place, at the header's height, `text-sm`, selecting its text on open.
- **Phosphor imports**: `PlusIcon`, `GearIcon`, `UsersIcon`, `ArchiveIcon`, `DoorOpenIcon`, `PencilSimpleIcon`, `ArrowLeftIcon`, `ArrowRightIcon`, `TrashIcon`, `ArrowsOutIcon`, `KanbanIcon`, `SignOutIcon`, `XIcon`.
- **Navigation after an action** (§13), and clearing the remembered workspace after *Delete workspace*, are recorded in `frontend-architecture.md`.

## Options considered

Every option is named, with its cost, under its decision above, and all were drawn side by side on the decision page. In short:
- **Where overlays live:** parts only over `Doc /` strips; the screen copies on `Page 1`.
- **Menu width:** fitted to the longest entry, at least 160.
- **Glyphs:** words only.
- **Rules between groups:** a rule before the entries that ask first.
- **What is rose:** all five actions.
- **Which actions ask:** Archive acting at once, with Undo only.
- **Undo:** dropping it and editing US-12c.
- **Typing the name:** never; also for a column that holds cards.
- **Dialog:** 432 with 16 inside; 512 with 24.
- **The rose button's busy look:** the plum busy every button shares, the recommendation, which the user turned down.
- **The sheet:** rows only, with no head.
- **Rename:** the field in the title's place only.
- **The action's word:** *Log out*, as US-02b wrote it.

## Out of scope (deferred, recorded)

- **The create-column form.** *Add column* opens it, and inventory §3 never listed it. It joins group c, with the other forms.
- **The pickers and *Move card to…***, including the phone's sheet over a sheet (ADR-038). Group b.
- **The toast's *Undo*, where a toast sits, and the wordings for states 4, 27 and 32.** Group e.
- **The user menu from 2.3's app bar.** The same part, opening downward; not drawn.
- **Hover, pressed and focus on the overlay boards.** The parts carry those looks.
- **Movement.** Menus and sheets open and close in code only (inventory §4.3).
- **The overlays at a tablet's Coarse density.** The desktop boards are drawn at Fine (*Operational note*).

## Downstream edits this ADR requires

1. **[US-02b](../../product/001-user-stories.md):** its note says *Sign out*.
2. **[ADR-018](adr-018-state-vocabulary.md): annotation.** Its Status line records §5's pale rose busy for the rose button, and row 38's plum confirm for *Archive board*.
3. **[ADR-021](adr-021-design-inventory-corrections.md): annotation.** §5's confirmations are five.
4. **[ADR-030](adr-030-toast.md): annotation.** §3's *no action button* is superseded for the archive toast.
5. **[ADR-024](adr-024-control-and-field-internals.md): annotation.** §3's pressed rule covers the rose button.
6. **[ADR-017](adr-017-ui-foundations.md): annotation.** §9's seven new variables.
7. **[ADR-027](adr-027-column-header.md), [ADR-033](adr-033-workspace-screens.md), [ADR-034](adr-034-side-menu.md), [ADR-038](adr-038-board-states-group-d.md), [ADR-039](adr-039-drag-and-phone-board.md): annotations.** The menus, the rename, the user menu, the delete confirmation and the board ⋯ sheet they left to Stage 5 are drawn.
8. **Penpot `Collabotask App`:** done 2026-11-23 (*Operational note*).
9. **[`design-file-inventory.md`](../design-file-inventory.md):** the sources table, §1.5, §1.7, §3, §4.2, §5, §7.2, §7.3, §7.6 and *Related documents*.
10. **[`frontend-architecture.md`](../frontend-architecture.md):** where these actions leave you, and the remembered workspace after a delete.
11. **`CLAUDE.md` › `## Now`:** a pointer.
12. **`docs/build-log.md`:** no entry. Stage 5 is not closed.
13. **PRD, SRS, `CONTEXT.md`:** no change.

## Operational note

**Drawn 2026-11-23.** The desktop overlays were placed at Light + Fine and the phone ones at Light + Coarse. The file was left in Light + Coarse, as it was found.

**Checks.**
- **Binding check, Overlays page:** only the card `⋯` rows §7.5 expects (21 on each 2.5 and 2.6 board, 15 on each 2.15 board), and nothing on the two 2.4 boards.
- **Binding check, parts:** nothing on `Glyph — states`, `Menu row — states`, `Menu — states`, `Confirmation — states` or `Button — states`. On `Card — states`, only its seven `⋯` rows.
- **Binding check, Foundations:** its expected 53 rows.
- **The looks:** Dark + Fine and Light + Coarse were looked at, on the part boards and the Overlays boards. After each flip every text box matched its text, apart from the drawer's 96-wide labels, which are fixed by design.

**What drawing found:**
- **`penpot.createShapeFromSvg()` creates its shape at the canvas origin, inside whatever board is there.** On `Components` that is the Card part's main, and the glyph spread to all 205 Card copies across four pages. Removing it from the main removed every copy, and the Card's seven looks were checked back at their heights. Glyphs were then drawn with `createPath()`, which goes to the page root, from Phosphor's outlines scaled from 256 to 16 by hand. A part's `instance()` is also created at the origin, inside the board there, and was moved at once.
- **The plugin cannot copy or move a shape to another page.** `appendChild` across pages throws *Cannot modify a page that is not currently active*. The user pasted the four screens onto `Overlays`.
- **`penpotUtils.findShapeById` searches every page, and timed out on this file.** Use `penpot.currentPage.getShapeById`.
- **The tab slept whenever VS Code covered it**, even with the tab selected. Windows side by side kept it awake. After one reload the plugin's storage was empty, and the last writes before the sleep had not been saved: fifteen glyph boards and a move. Read the file before re-running anything.
- **Penpot's open page can change under the plugin** when the Penpot window is clicked. Check `penpot.currentPage` before every write.
- **A padding bound right after a flex board is made can be reported but not laid out.** The sheets' heads read 12 and drew 0. Setting the value to 0 and back made Penpot lay them out.
- **A minimum height bound on a board that sizes itself to its content is ignored.** The user menu's identity row read 44 and drew 36. Bind the height instead.
- **Two tokens applied to one shape in one call can leave only one bound.** Apply the second in its own call and read it back.
- **A typography token changed inside a copy went back to the main's on a theme switch.** The rename field, an Input copy, returned to `type.title`. It was detached so its `type.body` stays. Colour changes inside copies survived.
- **A copy's width bound in its main snapped back after the first resize.** Set its sizing to fixed and resize again; the phone dialog then kept 358.
- **A shape appended to a flex board is laid out with its siblings.** Set `layoutChild.absolute` before placing an overlay.
- **Moving a part's holding board, or a main, is slow while Penpot re-syncs the copies.** The calls time out, but the change lands.
- **The desktop overlays are placed by position for Fine.** At Coarse their triggers grow from 24 to 44 and the cards move. The bar's and the column's menus then overlap their ⋯ by about 10px, and the card ⋯ menu covers part of its card's title. That is the same as 2.14's lifted card (ADR-039).
- **An export can lag a theme switch.** The Menu board exported once in Light while Dark was on, then correctly.
