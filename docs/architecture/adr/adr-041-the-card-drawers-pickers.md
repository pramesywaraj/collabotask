# ADR-041: The Card Drawer's Pickers — Stage 5, Group b

- **Status:** Accepted — **§7's mark spent once more by [ADR-042](adr-042-the-forms.md)** (2026-11-24): the chosen colour in Create a board takes the 1px plum edge, and the chosen visibility row takes plum text, a plum edge and a plum dot. §15's drawer at 390 is also the frame for every form below 640. **§11's filter applied to every list of people by [ADR-043](adr-043-the-lists.md)** (2026-11-24): the team list and the board's members list show it above eight people, counting everyone in the list, and a list scrolls after eight rows. **§13's refused-change toast settled by [ADR-044](adr-044-the-sidebar-sliding-over-and-the-toast.md)** (2026-11-24): *Couldn’t change “title”.* and one of five reasons, in the corner over the drawer's foot; after a 403 or 404 the board reloads. Everything else stands as written.
- **Date:** 2026-11-23
- **Scope:** Question 4 for Stage 5's group b: the three pickers the card drawer opens, *Due date*, *Assignee* and *Move card to…*, and the case [ADR-038](adr-038-board-states-group-d.md) left open, the phone's sheet over a sheet. It settles:
  - where *Move card to…* opens from the drawer and from the board, and how it asks for a place
  - the inline picker, once, for every later group: where it sits in the drawer, its surface, and how it marks the current choice
  - which drawer fields carry a caret
  - the calendar: its geometry, today's mark, the first day of the week, past days, and no quick picks
  - what a picked day is stored as
  - the assignee list: who is in it, its order, and when it has a filter
  - every wording, and the behaviour rules
  - the drawer at 390, drawn here for the first time

  **No token is added (220).** Every value maps by rule onto what group a and ADR-038 already made. Question 4 found no missing value for the first time since ADR-032's first round; it would have fired on one value, `spacing.row-padding-y`, had two-line places been picked (§4). **Four parts are added (23 → 27):** `Picker row`, `Day`, `Calendar` and `Picker`. The `Glyph` part gains four looks (19). Eight screens are added to the `Overlays` page (17).

  It **supersedes [ADR-021](adr-021-design-inventory-corrections.md) §5's placement of *Move card to…*** in the *Forms and lists* row, and nothing else in that table. In the drawer *Move card to…* is an inline picker like the other two. From the board it takes the menu's place, a popover from 640 up and the bottom sheet below. It **narrows [ADR-038](adr-038-board-states-group-d.md) §5**: *Column* no longer opens a picker of its own, and all three fields carry the caret (§8). It **settles ADR-038's *Out of scope*** for the pickers and the sheet over a sheet. It **extends [ADR-040](adr-040-menus-and-the-confirmation.md)**: the card ⋯ menu's *Move card to…* turns that menu into the picker, and group a's surface is reused in the flow. [ADR-018](adr-018-state-vocabulary.md)'s state 18 stays closed: the current choice is plum's *where you are*, and today is the weight channel's *emphasis within a block*.
- **Upstream edits:** one. [SRS](../../spesifications/001-software-specifications.md) §3.6 gains the rule that a card's `due_date` is a calendar day, sent as midnight UTC and read back by its date part; UC-14 and UC-16 point to it (§12). No endpoint, schema or permission rule changes: the server already stores what it is sent, and already clears on `due_date: null`.

## Context

Inventory §7.3 draws Stage 5 in five groups, each settling its kind of container once. Group b is the card drawer's pickers. It comes second because it reuses group a's menu surface, and because ADR-038 left it one structural question: on a phone the drawer is a full-height sheet, ADR-021 §5 filed *Move card to…* with the forms as a full-height sheet too, and the same section forbids a sheet over a sheet.

**Question 4 was asked of every value the pickers need, against the live 220 tokens, before anything was drawn. Every value mapped.** The pickers' rows are the `Menu row`'s geometry. Their surface is the menu's, without its shadow when it lies in the drawer. A calendar day is a control, so it is `size.control-height` square, and seven of them make the panel 284 wide at Fine and 340 at Coarse. Today's mark is `font.weight.semibold`, and the current choice is `color.primary` with `border.width`. The places need no slot glyph, and the way back is the `Glyph` part's `ArrowLeft`. Two measurements set the rules:
- **the value column is 262 wide at 390**, and seven 44px days with their gaps need 340, so on a phone a picker spans the whole row (§5)
- **the picker's edge is 1.26:1 against the drawer in Light and 1.70:1 in Dark**, enough to show where it ends and the next field begins (§6)

**Sixteen choices had never been made.** One reclassifies *Move card to…* in ADR-021 §5's table, and one touches the API contract: the SRS says every timestamp is UTC but not what time of day a due date gets. All sixteen went on one rendered page, drawn in Light and Dark with the file's values, at Fine on the desktop and at Coarse on a 390 phone wherever the phone differs. **The user took every recommendation.** B06, the current choice's mark, was not saved on the page and was asked again; the user took the recommendation.

Two facts were found while checking the code and the file:
- **The server can clear a due date and an assignee.** `UpdateCardRequest` takes both as `OptionalPatch`, so a present `null` clears them. *Remove due date* and *Unassigned* need no backend change.
- **2.4's tile said *Q3 Launch Plan · 3 members*, while 2.5 assigns cards to six people**, and an assignee must be a member. The tile now reads *8 members*, the board drawn here: you, Arif Rahman, Budi Santoso, Dimas Nugroho, Joko Tirtana, Maya Kusuma, Rina Wijaya and Sari Lestari. Only Ayu Pratama and Joko Tirtana had been named in the file before.

## Decision

### 1. The sheet over a sheet: every picker opens in place in the drawer, *Move card to…* too (B01)

- **Pressing a field opens its picker right under it**, pushing the rows below down. *Column* opens *Move card to…*, *Assignee* the member list, *Due date* the calendar. Nothing stacks, at any width, so the phone never shows a sheet over a sheet.
- **This was ADR-038's own guess** (*likely an in-place list like the other pickers*), and it keeps ADR-021 §5's rule that a picker reached from the drawer is never a second full-height layer.
- **The cost:** a long column makes the drawer tall. To Do has 12 cards, so 12 places. The drawer scrolls as one piece (§13).

The rejected options: a centred dialog over the drawer at every width, which on a phone scrolls a 12-place list inside a 358 dialog over a dimmed drawer; and, below 640, the drawer's content turning into the picker with a back arrow, which needs a pattern (pages inside a sheet) that nothing else in the app uses.

### 2. From the board, *Move card to…* opens where the menu was (B02)

- **From 640 up, the card ⋯ popover turns into the picker in the same place**: the same surface, 224 wide, 4 below the ⋯, right edges lined up (ADR-040 §2). **`M` on a focused card** opens the same popover at the card. The board is not dimmed, so you can see the columns you are choosing between.
- **Below 640, the card ⋯ bottom sheet turns into the picker**: the same sheet, its head still the card's title and *Close*. Below 640 this is also the only way to move a card past the next column (ADR-019 §5).
- **So *Move card to…* leaves ADR-021 §5's *Forms and lists*.** In the drawer it is an inline picker; from the board it is a menu-kind overlay. Its content is one component wherever it opens.
- A long column makes a tall popover. Past the screen's edge, the popover scrolls inside; that is the one place a picker scrolls on its own.

The rejected options: a centred dialog from 640 up and a full-height sheet below, as ADR-021 §5's table had it, which dims the board for two taps and hides it entirely on a phone; and a centred dialog at every width.

### 3. Move asks for a column, then a place; the place moves the card (B03)

- **Two steps, two taps for every move.** First the columns, each with its name and a `Count chip` holding its card count, the card's own column marked (§7). Then the places in the column you pick, under a row with `ArrowLeft` and the column's name, which takes you back.
- **The places are ADR-019 §6's**: *Top of column*, *After “card”* for each card but the last, and *Bottom of column*. They keep an empty 16px slot, so they sit indented under the column's name, as ADR-035 §3's switcher rows do. Nothing moves until you pick a place.
- **The card's own column is in the list**, so picking it reorders. There the card itself is not offered, its current place is marked, and picking that place closes the picker with no request.
- ***After* the last card is *Bottom of column***, so it is not listed twice. **An empty column has one place**, so picking Review moves the card at once.
- **From the board** a *Move to* label in `type.meta` and `color.muted-foreground` sits over the columns, inset by `spacing.control-padding-x` and `spacing.list-gap` from the top. In the drawer the field already says it, so there is no label.

The rejected options: a column moving the card to its bottom at once, with a › in each row for an exact place, which puts two targets in every row and moves the card on a slip of the thumb; and a form with two drop-downs and a *Move card* button, which is a picker inside a picker.

### 4. A long card title in a place: one line (B15)

- **Every place is one line**, `size.control-height` tall. From the board the picker is 224 wide, which leaves 168px, about 24 characters. In the drawer it has 376 at 640 and up and 302 at 390.
- **In code a long title ends in an ellipsis**, and the full title is the row's spoken name. Penpot cannot draw an ellipsis, so the file's 224 popover types its three long places short: *After “Set up the waitli…”*, *After “Choose the lau…”*, *After “Agree the three…”*.

The rejected option: up to two lines, which gives rows two heights and needs `spacing.row-padding-y` 8 / 12, worked out as (`size.control-height` − 20) ÷ 2.

### 5. Where an in-place picker sits: under its value from 640 up, across the row below (B04)

- **From 640 up it drops from the field it belongs to**: its left edge on the value's, past a `size.field-label` indent, 432 wide.
- **Below 640 it spans the row**, 358 wide at 390. The value column there is 262, and the calendar's seven 44px days need 340.
- **The date picker hugs its grid** at both: 284 at Fine and 340 at Coarse. The lists fill their width.

The rejected options: across the row at every width, which starts the names 84px left of the value above them; and under the value at every width, which shrinks a phone's days to 32px, below ADR-019's 44.

### 6. Its surface: the menu's, laid in the flow (B05)

- **`color.popover`, a 1px `color.border` edge (`border.width`, inside its box), `radius.surface` corners, `spacing.menu-padding` inside, rows `spacing.list-gap` apart.** That is group a's menu without `shadow.overlay`, since in the drawer the picker is not lifted. So a picker looks the same in the drawer and floating over the board.
- **The edge measures 1.26:1 against the drawer in Light and 1.70:1 in Dark.** It is faint in Light. The rows' own hover edges sit inside it.

The rejected options: a sunk well in `color.muted`, 1.09:1 in Light and 1.15:1 in Dark against the drawer, with a fill nothing else uses; and no surface, which makes the list read as more fields.

### 7. The current choice: plum text and a plum edge, everywhere (B06)

- **The current assignee, the card's column, its place, and its due date all take the sidebar's *current* look**: the words in `color.primary` and a 1px `color.primary` edge inside the row or the day. In a row, *(you)* turns plum with the name.
- **Plum is 6.54:1 against the drawer in Light and 6.11:1 in Dark.** Against normal text it is only 2.67 / 2.74, so the edge carries the mark as a shape as well (ADR-034 §10.3).
- **That is plum's *where you are***, the Active / Archived switch's and the sidebar's. State 18, *selected*, stays closed.

The rejected options: edges on the rows and a filled plum day, which is the Primary button's shape, *the one action*, so the day reads as something to press; and a plum tick at the row's end with a filled day, which needs a new glyph look and marks choices two ways where the sidebar switcher already chose a dot over a tick (ADR-035 §3).

### 8. A caret on all three fields (B07)

- ***Column*, *Assignee* and *Due date* each show Phosphor `CaretUpDown`**, the `Field` part's own `Caret`, at every pointer. A field that opens something says so, and on a phone, where there is no hover, nothing else would.
- **This narrows ADR-038 §5**, where only *Column* carried it, as the one field that opened a picker of its own. With §1, all three open the same way.
- **The open field has no look of its own**, as an open menu's ⋯ has none (ADR-040 §2). Its picker shows itself.

The rejected options: no caret, which hides on a phone that the fields can be pressed; and *Column* only, which marks a difference §1 removed.

### 9. The calendar

- **A month of day buttons, one `Calendar` part.** Each day is `size.control-height` square (36 / 44), `radius.control`, `type.body`, `spacing.list-gap` apart both ways. It always draws **six week rows**, so it keeps its height as you page through months; days of other months are blank.
- **Its head is `size.control-height` tall**: the month's name in `type.card-title`, inset by `spacing.control-padding-x`, then *Previous month* and *Next month*, Phosphor `CaretLeft` and `CaretRight` at `size.icon` in `size.target-min` boxes. **Under it, the weekday initials** in `type.meta` and `color.muted-foreground`, at their own 16px line, each as wide as a day.
- **Under the month, *Remove due date***: a `Picker row` with Phosphor `X` in `color.muted-foreground`. It shows only when the card has a date, since entries you can't use are absent (ADR-018 §5).
- **It opens on the card's month**, or this month when there is no date. Picking a day saves it and closes the picker.

### 10. Today, the week, past days, and no quick picks (B08–B11)

- **Today is a semibold number** (`font.weight.semibold`), with no colour and no shape. That is the weight channel's *emphasis within a block*, and `aria-current="date"` says it to a screen reader. When today is the card's date it is both semibold and plum-edged (B08).
- **The week starts where the browser's locale says**, read from `Intl.Locale`'s week information where the browser has it, and on Monday where it does not. Firefox has no week information yet, so it gets Monday. The file draws Monday first (B09).
- **A past day can be picked**, with no look of its own. A card dated last Friday then reads *Overdue · Nov 20*, like any late card (B10).
- **There are no quick picks.** Today is one tap away in the grid, and *Due today* and *Due tomorrow* stay the card's words, not buttons (B11).

The rejected options: a `color.muted` fill for today, a fill where every other state is an edge, 1.09:1 in Light; a dot under it, which already means the selected workspace (ADR-035 §3) and needs a size with no token; Monday or Sunday always; past days that can't be picked, a look ADR-018 §5 does not have; and *Today* and *Tomorrow* buttons above the month.

### 11. The assignee list (B13, B14)

- **The board's members only** (US-16; assignment is participation). An admin who hasn't joined is not in it.
- ***Unassigned* first**, with the Avatar's `Unknown` look; picking it clears the card (US-16's *I can clear the assignee*). **Then you, marked *(you)*** in `color.muted-foreground`, `spacing.inline-gap` after your name. **Then everyone else, A–Z by name.** Each row is an Avatar at `size.avatar` in the slot's place, then the name.
- **A filter field appears when the board has more than eight members.** Eight rows at Coarse are about 380px, under half a phone screen. It sits first in the picker: `size.control-height`, filled `color.background` with a `color.input` edge, as the Input's box is (ADR-024), Phosphor `MagnifyingGlass` and *Find a member* in `color.muted-foreground`. From 640 up focus starts in it. On a phone focus starts on the list, so the keyboard waits until you tap the field.

The rejected options: no filter, relying on a listbox's typeahead, which a phone cannot type into; a filter field always, which costs a row on boards of three; and *Unassigned* then everyone A–Z, where your place moves from board to board.

### 12. A picked day is stored as the day at midnight UTC (B12)

- **The app sends `2026-11-24T00:00:00Z` and reads back only `2026-11-24`**, never through the reader's local clock. Joko in Jakarta (UTC+7) and Sari in London (UTC+0) both see *Nov 24*.
- ***Due today* compares that day with each viewer's own today**, written the same way, so the words can differ by time zone on the day itself, which is true, while the date never does.
- **The SRS says so first** (§3.6; UC-14 and UC-16). The server stores and returns the value unchanged, so nothing on the backend changes.

The rejected option: the viewer's local midnight converted to UTC. Joko's pick would be stored as `2026-11-23T17:00:00Z`, Sari would see *Nov 23*, and on the 23rd her card would say *Due today* while Joko's said *Due tomorrow*.

### 13. The behaviour rules

- **One picker is open at a time.** Pressing its field again, picking, or Escape closes it. Escape closes the picker, not the drawer, and focus returns to the field.
- **Picking saves at once.** There is no Save button. The field shows the new value and the card on the board behind goes 60% until the server answers (ADR-018 §3). A refused change puts the old value back with a toast (state 27, group e).
- **After a move from the drawer** the drawer stays open and *Column* shows the new column. **After a move from the board** the card moves at once, focus follows it to its new place, and a polite live region says where it went.
- **A teammate's change while a picker is open:** the picker stays open and its mark follows the new value. A member who leaves the board leaves the list. A card deleted under you closes the drawer (state 32, group e).
- **A visitor who hasn't joined** (2.7) sees the drawer's fields as plain text: no caret, no edge, no picker, and no *Delete card*.
- **On a phone an opened picker is scrolled into view, and the drawer scrolls as one piece.** Nothing inside it scrolls on its own. The floating picker from 640 up is the one exception (§2).
- **Keyboard:** the lists are listboxes, where the arrow keys move and Enter picks. The calendar is the WAI-ARIA date grid: arrows by day and week, Page Up and Page Down by month, Enter picks.

### 14. The wordings

| Where | Wording |
|---|---|
| The fields, as a screen reader names them | *Column, In Progress. Move card* · *Assignee, Joko Tirtana. Change assignee* · *Due date, due tomorrow, November 24. Change due date* |
| Assignee | *Unassigned* · *Ayu Pratama (you)* · the members by name. Filter: placeholder *Find a member*; nothing found: *No member matches “xy”.* |
| Due date | *November 2026* · *Previous month* · *Next month* (spoken) · *Mo Tu We Th Fr Sa Su* · a day, spoken *Tuesday, November 24, 2026* · *Remove due date*. After removing, the field reads *No due date* (ADR-038 §6) |
| Move card to… | From the board, *Move to* over the columns · a column row: its name and its count · the row back: the column's name, spoken *Back to columns* · the places: *Top of column* · *After “Choose the launch date”* · *Bottom of column* · the sheet's head: the card's title and *Close* |
| Heard after a move | *Moved to Done, at the bottom.* · *Moved to Done, after “Choose the launch date”.* · *Moved to the top of In Progress.* |

### 15. The drawer at 390

- **Drawn here for the first time**: a full-height sheet between the browser's bars, 390 × 704, with ADR-038 §4's inside unchanged. The labels are `size.field-label` 96 and the values 262, and blocks are `spacing.section-gap` apart.
- **It has no edge and no shadow**, since nothing sits beside it: `shadow.overlay` and `shadow.rule-left` are off. The foot stays at the bottom, and the body scrolls under it.

### 16. What the file draws

**Tokens:** none added (220).

**Parts, on `Components` (23 → 27):**
- **`Glyph`** gains four looks (15 → 19): `X` Muted, `CaretLeft`, `CaretRight` and `MagnifyingGlass`, each Phosphor's Regular outline scaled from 256 to 16, filled `color.muted-foreground`. Its holding board was widened to 953.
- **`Picker row`**, on `Picker row — states` (y 9680). `State` = Rest · Hover · Pressed · FocusVisible · Current. A flex row 224 wide at rest:
  - `size.control-height`, `spacing.control-padding-x`, `spacing.icon-gap`, `radius.control`, filled `color.popover`; an inner 1px edge on Hover (`color.border-hover`), Pressed (`color.border-pressed`) and Current (`color.primary`); `shadow.focus-ring` on FocusVisible
  - inside: `Slot` (`size.icon`, holding a `Glyph`), a hidden `Avatar`, `Label` (filling the row, `spacing.inline-gap`: `Name` in `type.body`, and a hidden `You`, *(you)* in `color.muted-foreground`), and a hidden `Count chip`. Current turns `Name` and `You` plum
- **`Day`**, on `Day — states` (y 9944). `State` × `Today`, seven looks: Rest, Hover, Pressed, FocusVisible and Current, with Rest and Current also as Today. A `size.control-height` square, `radius.control`, filled `color.popover`, centring a `Number` in `type.body`; the edge ladder as on the row; Current plum. On Today the number's weight is bound to `font.weight.semibold` and its family to `font.family` (*Operational note*).
- **`Calendar`**, on `Calendar — states` (y 10208). One part, no axis: November 2026, Monday first, the 23rd today and the 24th current. A flex column, `spacing.list-gap` apart: `Head` (`Month`, `Previous`, `Next`), `Weekdays` (seven), `Weeks` (six `Week` rows of seven `Day` copies or `Blank` boards, each `size.control-height`).
- **`Picker`**, on `Picker — states` (y 10720), filled `color.board-canvas`. `Of` × `Presentation`, nine looks:
  - Inline: `Assignee`, `Assignee filtered` (a twelve-member board, under the `Filter` field), `Due date`, `Column` and `Place` (Done's)
  - Popover: `Column` and `Place`
  - Sheet: `Column` and `Place`

  Inline is the menu surface in the flow (§6), 432 wide, apart from Due date, which hugs its grid. Popover and Sheet are detached copies of the `Menu` part's Card looks, holding `Picker row` copies in their `Panel`, with a *Move to* `Label` over the columns.

**The `Overlays` page gains eight boards:**

| Board | Where | Open on it |
|---|---|---|
| *2.6 + assignee* | x 0, y 4400 | the assignee list under Joko's field |
| *2.6 + due date* | x 1640, y 4400 | the calendar under the date |
| *2.6 + moving, a column* | x 3280, y 4400 | the columns under *Column* |
| *2.5 + moving from the card ⋯* | x 4920, y 4400 | the column picker where the card ⋯ menu was |
| *2.6 at 390* | x 0, y 5500 | the drawer as a full-height sheet, at rest |
| *2.6 at 390 + due date* | x 590, y 5500 | the calendar across the row |
| *2.6 at 390 + moving, a place* | x 1180, y 5500 | In Progress's places, the card's own marked |
| *2.15 + moving a card* | x 1770, y 5500 | the column picker in the bottom sheet |

The desktop drawers are clones of *2.6 + delete card* without its dialog; the phone drawers are a clone of its drawer, resized. Each picker sits in the drawer's `Fields` as a `Row / … picker`. On the desktop that row starts with an `Indent` board bound to `size.field-label`.

**Also corrected:** the carets on *Assignee* and *Due date* on 2.6 (`Page 1`) and *2.6 + delete card*; and Q3 Launch Plan's tile, *8 members*, on 2.4, *2.4 + workspace ⋯* and *2.4 + user menu*.

### 17. The code

- **No CSS variable is added.**
- **The in-place pickers are not popovers.** Each `Field` is a button with `aria-expanded` and `aria-controls`, and its picker renders in the flow below it. The lists are listboxes; the filtered assignee list is shadcn's `Command`, restyled to the menu's rows.
- **The calendar is shadcn's `Calendar` (react-day-picker), restyled**: days `size-[var(--control-height)]` with `rounded-md` and the edge ladder; the selected day `text-primary` with a 1px inset `ring-primary`, with no fill; today `font-semibold`, with no `bg-accent`; `fixedWeeks`, `showOutsideDays={false}`; `weekStartsOn` from the locale (§10); nothing disabled.
- **Move card to… is one component in three places**: in the drawer, under *Column*; from the card ⋯, the `DropdownMenu` closes and a `Popover` on the same trigger opens (from 640 up), or the bottom `Sheet`'s content is replaced (below 640); `M` on a focused card opens the `Popover` at the card. It computes the target position exactly as a drop does (UC-15).
- **The due date** is sent as `` `${day}T00:00:00Z` `` and read as `due_date.slice(0, 10)`, never through `new Date()` for display. ADR-037 §5's date function compares that day with the viewer's local today, written as `YYYY-MM-DD`. Recorded in `frontend-architecture.md`.
- **Phosphor imports:** `CaretUpDownIcon`, `CaretLeftIcon`, `CaretRightIcon`, `MagnifyingGlassIcon`, `ArrowLeftIcon`, `XIcon`, `UserIcon` (the `Unknown` avatar).

## Options considered

Every option is named, with its cost, under its decision above, and all were drawn side by side on the decision page. In short:
- **The sheet over a sheet:** a centred dialog over the drawer; the drawer turning into the picker below 640.
- **Move from the board:** a dialog from 640 up and a full-height sheet below; a dialog at every width.
- **How Move asks:** a column moving the card to its bottom, with a › for an exact place; a form with a *Move card* button.
- **Long titles:** two lines.
- **Where a picker sits:** across the row everywhere; under the value everywhere.
- **Its surface:** a sunk well; none.
- **The current choice:** a filled plum day; a plum tick and a filled day.
- **Carets:** none; *Column* only.
- **Today:** a muted fill; a dot.
- **The week:** Monday always; Sunday always.
- **Past days:** not pickable.
- **Quick picks:** *Today* and *Tomorrow*.
- **The stored day:** local midnight in UTC.
- **The filter:** none; always.
- **The order:** everyone A–Z.

## Out of scope (deferred, recorded)

- **Setting an assignee or a date while creating a card.** *Add card* takes a title (ADR-037 §4); US-14's optional fields are set in the drawer afterwards. The create-column form is group c.
- **Hover, pressed and focus on the overlay boards.** The parts carry those looks.
- **The filtered list on a screen.** Q3 Launch Plan has eight members, so its screens show no filter; the twelve-member look is on the part's board.
- **The popover from `M` on a focused card.** The same part at the card; not drawn.
- **Movement.** Pickers open and close in code only (inventory §4.3).
- **The overlays at a tablet's Coarse density.** The desktop boards are drawn at Fine (ADR-040 *Operational note*).

## Downstream edits this ADR requires

1. **[SRS](../../spesifications/001-software-specifications.md):** §3.6's calendar-day rule, and UC-14 and UC-16 pointing to it. Done first.
2. **[ADR-021](adr-021-design-inventory-corrections.md): annotation.** §5's table places *Move card to…* in *Forms and lists*; it is an inline picker in the drawer and a menu-kind overlay from the board.
3. **[ADR-038](adr-038-board-states-group-d.md): annotation.** §5's lone caret and *Column*'s own picker are narrowed; *Out of scope*'s pickers and sheet over a sheet are settled.
4. **[ADR-040](adr-040-menus-and-the-confirmation.md): annotation.** The card ⋯'s *Move card to…* turns the menu into the picker.
5. **[ADR-019](adr-019-responsive-touch-scope.md): annotation.** §6's picker is drawn, and it is required at every width as before.
6. **[ADR-018](adr-018-state-vocabulary.md): annotation.** Plum's *where you are* marks a picker's current choice, and the weight channel marks today; state 18 stays closed.
7. **[ADR-037](adr-037-board-at-rest.md): annotation.** §5's date function reads the stored day by its date part (§12).
8. **Penpot `Collabotask App`:** done 2026-11-23 (*Operational note*).
9. **[`design-file-inventory.md`](../design-file-inventory.md):** the sources table, §1.6, §3, §4.2, §5, §7.2, §7.3, §7.5's expected rows, §7.6 and *Related documents*.
10. **[`frontend-architecture.md`](../frontend-architecture.md):** the due-date rule, and *Move card to…*'s three entry points.
11. **`CLAUDE.md` › `## Now`:** a pointer.
12. **`docs/build-log.md`:** no entry. Stage 5 is not closed.
13. **PRD, user stories, `CONTEXT.md`:** no change.

## Operational note

**Drawn 2026-11-23**, at Light + Coarse as the file was found. Dark + Fine was looked at on *2.6 + moving, a column*, *2.6 + due date* and *2.5 + moving from the card ⋯*. The picker's edge shows against the drawer, the plum mark reads, and at Fine the calendar is 284 wide and the popover sits 4 below the card's ⋯. The file was put back to Light + Coarse. After the flips every text in the new pickers fitted its box.

**Checks.**
- **Binding check, parts:** nothing on `Glyph — states` or `Picker row — states`. On `Day — states`, `Calendar — states` and `Picker — states`, only today's `Number` (2, 1 and 1), which §7.5 now expects.
- **Binding check, Overlays:** only the card `⋯` rows (21 on each desktop board, 15 on each phone board) and today's `Number` on the two date screens. Nothing on *2.4 + workspace ⋯* or *2.4 + user menu* after the tile's text changed.
- **The Card part** was counted before the first `instance()` and after every batch, and never changed (ADR-040 found copies can land inside it).

**What drawing found:**
- **Binding a text's weight replaces its typography binding.** Today's number lost `type.body` when `font.weight.semibold` was bound, and the catalogue has no 14/20/600 style. It keeps the weight and `font.family` bound, with 14/20 typed, and the binding check reports it as `typography(14px)`. That is now an expected row, like the Card's `⋯`. Adding a typography token for one number was not worth breaking the catalogue's count.
- **`penpot.library.local.components` lists one look per part.** Take a part's looks from that component's `variants.variantComponents()`.
- **`addVariant()` gives the new look a placeholder value** (*Value 4*) on the other axis. Set it with `setVariantProperty` straight after.
- **`createVariantContainer` dropped a comma from a value**: *Assignee, filtered* became *Assignee filtered*.
- **`setParentIndex(k)` counts from the other end of `children`**: to put a shape at position *p* of *n*, pass *n* − 1 − *p*.
- **Tokens applied to a new shape can read back empty in the same call** and be bound when read in the next. Read bindings in a separate call before re-applying anything.
- **`resize()` refuses `'auto'`**, after the board is already made, which left an empty board to remove.
- **The Penpot tab slept during a theme switch**, and the call timed out. The switch had landed, and the tab woke when the user focused it.
- **An export timed out once on Penpot's server** (500) and worked on the retry.
