# ADR-042: The Forms — Stage 5, Group c

- **Status:** Accepted — **§1–§3's frame given the lists by [ADR-043](adr-043-the-lists.md)** (2026-11-24): the same 448 dialog and full-height sheet hold the team list and the board's members, with *Close* in place of the form's buttons; a list stops at eight rows and scrolls. **The invite screens' people corrected** the same day: they add Citra Dewi and Hendra Gunawan, not Dimas and Maya, who were already on Q3 Launch Plan, and *2.4 + invite people* reads *9 members*. Rina's mistyped address stays. *Out of scope*'s team panel is settled there. ***Out of scope*'s *Where a toast sits* settled by [ADR-044](adr-044-the-sidebar-sliding-over-and-the-toast.md)** (2026-11-24): 16 from the bottom and right; forms still report their own failures in the `Form message`, never in a toast. Everything else stands as written.
- **Date:** 2026-11-24
- **Scope:** Question 4 for Stage 5's group c: *Create a workspace*, *Create a board*, *Board settings*, *Invite people* and *Add column*. It settles:
  - the form's frame, once, for every later group: group a's 448 dialog from 640 up, a full-height sheet below, and how each one closes
  - a field of several lines, and the space above and below its text
  - how a form shows a minimum length
  - Create a workspace's warning, and a workspace description's length
  - the colour choice: its look, its size, which colour comes first, and what string is stored
  - who can see a board: two choice rows
  - how Board settings saves, and what it says before a board goes private
  - Invite people: where it opens, the addresses as chips, the results under the field, and how each outcome is marked
  - Add column: a new column at the end, its header a field, instead of a form
  - every wording, and the behaviour rules

  **One token is added, in two entries (220 → 222):** `spacing.control-padding-y`, 8 at Fine and 12 at Coarse. Question 4 fired in two rounds (19 of the 20 stages have now fired). Round 1 found that value. Round 2 came from the user's picks, and needed two values that resolved by rule with no token: a chip's height and a colour square's size. **Five parts are added (27 → 32):** `Swatch`, `Choice row`, `Address chip`, `Invite result` and `Form`. The `Input` part gains a `Content` axis and five looks (4 → 9), and the `Glyph` part three looks (19 → 22). Ten screens are added to the `Overlays` page (17 → 27), and 2.4's heading gains *Invite people* on eight boards.

  It **narrows [ADR-022](adr-022-form-control-geometry.md) §3**: `spacing.control-padding-y` now exists, for a control that grows with its content. A control of fixed height still centres its one line and doesn't use it. It **narrows [ADR-032](adr-032-auth-form-errors-and-password-rules.md) §8's rule 1 for minimum lengths only**: a name or title shorter than its minimum shows its error while you type. An empty field still waits for the press, and Login and Register don't change (§6). It **settles [ADR-021](adr-021-design-inventory-corrections.md) §5's *centred dialog or drawer*** for forms: the dialog. It **takes *Create column* out of the overlays** (§14). Inventory §3 listed it with the forms, and it becomes a column on the board. It **extends [ADR-040](adr-040-menus-and-the-confirmation.md) §7**: the confirmation's dialog holds the forms, and pressing outside it does nothing. It **extends [ADR-041](adr-041-the-card-drawers-pickers.md) §7**: the plum mark for a current choice now also marks the chosen colour and the chosen row. It **widens ADR-032's Emerald row in [ADR-018](adr-018-state-vocabulary.md)'s ledger** by one render site, an invite's *Added*. It **adds a button to [ADR-033](adr-033-workspace-screens.md)'s 2.4 heading**, *Invite people*, an Outline button for workspace admins. State 18 stays closed.
- **Upstream edits:** two, both in the [SRS](../../spesifications/001-software-specifications.md). UC-03 now gives a workspace name's 2 to 255 characters and a description's 1,000. That limit was already built: the request binding has none, but `CreateWorkspaceInput` validates `max=1000` in the usecase, so a longer one is 400. §4.3 now names the eight strings the app stores as `background_color` (§9). No endpoint, schema or permission rule changes, and no backend item is added.

## Context

Inventory §7.3 draws Stage 5 in five groups, each settling its kind of container once. Group c is the forms. It reuses group a's 448 dialog, which a 400 form fits exactly (`size.dialog-width` = `size.form-width` + 2 × `spacing.dialog-padding`), and it settles the full-height sheet that ADR-021 §5 puts every form in below 640.

**Question 4 was asked of every value the forms need, against the live 220 tokens, before anything was drawn.** Every value mapped but one: the space above and below the text of a field that holds several lines. ADR-022 §3 had left it out on purpose, because a one-line control centres its text in its height, and a field that grows has no height to centre in. **Twenty choices had never been made.** Two touched the API contract: what string a board's colour is stored as, and how long a workspace description can be. All twenty went on one rendered page, drawn in Light and Dark with the file's values, at Fine on the desktop and at Coarse on a 390 phone wherever the phone differs. **The user took sixteen recommendations** and chose four others: errors as you type (C06 C), two choice rows for who can see a board (C12 B), chips for addresses (C16 B), and the results under the field (C17 B). On C09 they asked for a bigger colour square.

Those picks needed two values the file didn't have, so **question 4 fired a second time**, on a second rendered round. The user took the recommendation for the chip (D01 A, no token) and chose the larger colour square at Coarse (D02 B, no token).

Four facts were found in the backend code:
- **The server doesn't trim a workspace name, a board title or a column title.** Two spaces pass the name's 2-character minimum. So the browser trims every value before it checks it and before it sends it (§15).
- **One malformed address fails the whole invite** with a 400, before anyone is added (`dive,email` on `InviteMemberRequest.Emails`). So the browser checks every address before sending (§12).
- **A board colour is any 4 to 8 characters**, and nothing said which string means which hue. Two hue names don't fit in 8 (*periwinkle*, *terracotta*). §9 settles it.
- **A workspace description already stops at 1,000 characters on the server.** The decision page said it had no limit, because only the request binding was checked. The usecase validates `max=1000`, so C08 needed only the SRS to say so.

## Decision

### 1. The frame from 640 up: the confirmation's dialog (C01)

- **Every form is group a's 448 dialog** (ADR-040 §7): `color.popover`, a 1px `color.border` edge, `radius.surface`, `shadow.overlay`, over `color.scrim`. Inside are `spacing.dialog-padding` 24 and, from top to bottom: the title in `type.title`, an optional sentence in `type.body` and `color.muted-foreground` `spacing.field-inner-gap` below it, the fields `spacing.field-gap` apart, and *Cancel* (Outline) then the Primary button at the right, `spacing.inline-gap` apart. These blocks are `spacing.section-gap` apart.
- **No ×.** *Cancel* and Escape close it, and the app keeps one dialog shape.
- The rejected options: the same dialog with a × by the title, as shadcn's `Dialog` ships, which puts two ways to close 400px apart; and a drawer from the right like the card drawer's, which is a full-height panel for one to four fields.

### 2. Below 640: a full-height sheet with one full-width button (C02)

- **The drawer at 390's frame** (ADR-041 §15): 390 × 704 between the browser's bars, `spacing.screen-x` inside, no edge, no shadow. Its head holds the title and *Close*, a Phosphor `X` in a `size.target-min` square. The sentence sits `spacing.field-inner-gap` below the head.
- **The fields, then the Primary button across the form**, `spacing.section-gap` below the last field, as on Login. *Cancel* is dropped, because *Close* does the same.
- When the keyboard is up the button can sit under it. The keyboard's Go key sends the form, and the sheet scrolls.
- The rejected options: buttons pinned in a foot, which phones handle differently when the keyboard opens (two behaviours to test, for forms that fit anyway); and the dialog 358 wide, which a four-field form with its buttons doesn't fit with the keyboard up.

### 3. Closing: pressing outside does nothing (C03)

- **Pressing the dimmed screen does nothing.** A stray click can't lose a half-written description, as the confirmation (shadcn's `AlertDialog`) already behaves.
- **Escape, *Cancel* and *Close* close it at once and never ask.** What was typed is dropped.
- The rejected options: pressing outside closes it too, shadcn's default; and asking *Discard what you typed?*, a dialog over a dialog for a draft that isn't lost for good.

### 4. A field of several lines grows from one line (C04)

- **At rest it is `size.control-height` tall, like every other field**, and each new line adds 20. The form grows with it, and the dialog scrolls if it outgrows the screen.
- Workspace and board descriptions use it. Nothing at rest shows that it takes several lines; the label says it.
- The rejected options: three lines at rest, which needs a `size.textarea-min-height` and costs 40px on forms whose descriptions mostly stay empty; and one line that never grows, which scrolls a 1,000-character description sideways.

### 5. Its padding: `spacing.control-padding-y` 8 / 12 (C05)

- **(36 − 20) ÷ 2 = 8 at Fine and (44 − 20) ÷ 2 = 12 at Coarse**, written into both density sets as literals (ADR-022 §4). The first line then sits exactly where a one-line field's does.
- **It narrows ADR-022 §3**: a control of fixed height still centres its text and doesn't use it.
- **The choice rows use it too** (§10), so they need no token of their own.
- The rejected option: borrowing `spacing.card-padding-y` by rule. It is the same 8 / 12 today, but a change to the card would move every field. ADR-022 gave controls their own `control-padding-x` for that reason.

### 6. A minimum length shows as you type (C06)

- **While a name or title holds fewer characters than it needs, the field shows its error**: the rose edge and *Use at least 3 characters.* (or *2* for a workspace name), from the first character until the minimum is met.
- **An empty field waits for the press.** Pressing the Primary button with it empty shows *Enter a title.* (or *Enter a name.*) and moves focus there.
- **The Primary button looks not satisfied until the form can be sent**, as on Login (ADR-032 §8, ADR-018 §5). It can still be pressed.
- **This narrows ADR-032 §8's rule 1 for minimum lengths only.** Login and Register have no minimum-length field apart from the password, whose checklist already ticks as you type, so they don't change. An address's format waits until it becomes a chip (§12).
- This was the user's call. The recommendation was Login's rule, which mentions the 3 only after a press. A hint under the field was also offered.

### 7. Create a workspace says that nothing can be renamed (C07, C08)

- **One sentence in the confirmation's sentence slot**: *You can’t rename a workspace or change its description later.* No endpoint renames a workspace or edits its description (PRD §5.1).
- **The description stops at 1,000 characters, in the browser and on the server.** The server already checks it (§ *Context*).
- The rejected options: a hint under each field, two lines saying the same thing; no warning; and no limit, which would push 2.4's boards down the screen.

### 8. The colour choice: eight buttons, each holding its colour (C09, D02)

- **Each hue is a button the size of a control**: a `size.control-height` square, `radius.control`, filled `color.popover`, `spacing.list-gap` from its neighbours. Inside it a square in the hue's `.strong`, `radius.badge`, fills the button inside `spacing.inline-gap` 8 on every side. That is **20 at Fine and 28 at Coarse**: the user's call (D02 B) over 20 at both. It costs a gap token used as padding, so a change to the space between neighbours would resize the squares.
- **Hover and press step the edge** (`color.border-hover`, `color.border-pressed`), and **the chosen one takes ADR-041 §7's plum edge**. The colour needs no other mark.
- **Eight in a row from 640 up** (316 at Fine, 380 at Coarse, inside 400), and **two rows of four below**, since eight 44s need 380 and the phone's form is 358.
- **Spoken as a radio group named *Colour***, each by its hue's name, which also shows as a tooltip on hover.
- The rejected options: filled squares with a check, a mark unlike any other current choice; and filled circles with a ring, which is the focus ring's shape (ADR-018 §6).

### 9. Which colour comes first, and what is stored (C10, C11)

- **The first hue that no active board in the workspace uses**, in ADR-017 §3.3's order, so a new board tends to differ from its neighbours. When all eight are used, terracotta. Product Team's six boards give olive.
- **`background_color` is the hue's deep value**, a 7-character hex string: terracotta `#744539`, ochre `#746239`, olive `#627439`, sage `#39744F`, teal `#397472`, steel `#395974`, periwinkle `#394174`, mauve `#74395F`. It names the hue; it is not drawn as is, because the deep and pale values swap with the theme.
- **It is read back by exact match, ignoring case.** Anything else, including the legacy `#0079BF`, maps from the board id, as SRS §4.3 already said. Create a board always sends one of the eight, so the legacy default applies only to boards made by a direct API call. No backend change.
- If ADR-017 ever changes a deep value, boards keep the old string, and the app will need a table of old values to read them.
- The rejected options: always terracotta first; none chosen; and short keys (*terra*, *peri*), two made-up abbreviations in a field named `background_color`.

### 10. Who can see it: two choice rows (C12)

- **Two radio rows, *Workspace* and *Private*, each with its sentence under its name.** A row is a `Choice row`: `spacing.control-padding-y` above and below, `spacing.control-padding-x` at the sides, `radius.control`, `color.popover`, rows `spacing.list-gap` apart. In it are a 16 circle (`size.icon`, `radius.full`, a 1px `color.input` edge), `spacing.icon-gap` before the name in `type.body`, and the sentence in `type.meta` and `color.muted-foreground` under the name.
- **The chosen row takes ADR-041 §7's look**: its name and edge in plum, the circle's edge in plum, and a plum dot, `size.dot` 8, in it. Hover and press step the row's edge.
- Workspace is chosen when Create a board opens (UC-07).
- This was the user's call. The recommendation was two Segments with one sentence for the chosen option. A checkbox, *Private board*, was also offered.

### 11. Board settings saves with a button, and says who loses access (C13, C14)

- **The same form as Create a board, filled in, with *Save changes*.** Nothing changes until it is pressed; then everyone on the board sees the change. One form component serves both.
- **Choosing *Private* adds a line under the two rows, in the text colour**: *Members of Product Team who aren’t on it lose access, and are taken out if they have it open.* It is there before you save, and saving doesn't ask again.
- The rejected options: each field saving on its own, like the card drawer, which would take people out of a board on the tap; *Save* asking first, a form that turns into a confirmation; and no warning.

### 12. Invite people: from 2.4's heading, the addresses as chips (C15, C16, D01)

- **An Outline button, *Invite people*, in 2.4's heading after the member count**, for workspace admins only. It opens the form as a dialog or a sheet on its own. A new workspace's next two steps, *Invite people* and *Create board*, then sit side by side. On a phone the heading's buttons wrap onto their own row. Inviting from inside the team panel was rejected: that panel is a full-height sheet on a phone, so it would put a sheet over a sheet.
- **Each address becomes a chip.** A comma, a semicolon, a space, Enter or a paste turns what was typed into chips, lower-cased and trimmed, and a repeat is dropped as it is made. Backspace in the empty field removes the last chip.
- **A chip is `size.target-min` tall, 24 / 44, with its × in a square of the same size** (D01 A), so on a phone the × is a full target. It is filled `color.chip`, `radius.badge`, with the address in `type.body` `spacing.inline-gap` from its left edge. The field holds the chips `spacing.list-gap` from its edge and from each other, and grows a row at a time. On a phone four addresses make the field about 200 tall.
- **A chip that isn't an email address turns rose when it is made** (a 1px `color.destructive` edge), and the field's message names it: *“rina.w@” isn’t an email address.* While one is rose, *Invite* looks not satisfied.
- This was the user's call over one field of several lines. The rejected chip sizes: fitted to keep the field at 36 / 44, which leaves a 34 × on touch, and the same with no × on touch.

### 13. The results sit under the field, marked like the checklist (C17, C18)

- **After sending, a row for every address sits under the field**, `spacing.field-gap` below it, the rows `spacing.field-inner-gap` apart. A row has the glyph, `spacing.icon-gap`, the address in `type.body`, and the outcome under it in `type.meta`.
- **The marks are the password checklist's**: *Added*, an emerald check (`color.done`); *No CollaboTask account uses this address.*, a rose × and rose words (`color.destructive-text`); *Already in Product Team.*, a grey check, since they're in and nothing changed. That gives emerald a second render site (ADR-032 gave it the first).
- **The field keeps the addresses that found no account, as chips**, so a typo can be fixed and sent again. The added ones and the ones already in leave the field.
- **After a send, *Cancel* reads *Close*.** People were added, so nothing is cancelled. A new send replaces the results.
- **The summary is spoken, not shown**: a polite live region says *2 of 4 added to Product Team.* (*All 4 added to Product Team.*, *None of the 4 were added.*). Each row already shows its own outcome.
- This was the user's call over the results taking the form's place with *Invite more* and *Done*, so those two wordings are dropped. A toast was also offered.

### 14. Add column: a new column at the end, its header a field (C19)

- **The bar's *Add column* (and the board ⋯'s below 640, ADR-039) scrolls the board to its end, where a new column appears with its header a field**, focused, reading *Column name*. It is the rename field's look (ADR-040 §10): the `Input` part's InlineEditing look in the header's place.
- **Enter, or leaving the field with a name, adds the column.** Escape, or leaving it empty, removes it. Until then it exists only in your browser.
- **So *Create column* is not an overlay.** It leaves the form dialog, and group c draws four forms. It is still the same use case (UC-13).
- The rejected options: the form dialog with one field, and a full-height sheet for it on a phone; and a popover under *Add column*, after which the new column appears where you can't see it.

### 15. The behaviour rules

- **Creating puts you inside what you made** (inventory §1.3, §1.4). After *Create workspace* the app opens the new workspace, which has no boards yet. After *Create board* it opens the new board, with To Do, In Progress and Done.
- **Where each form opens**: *Create workspace* on 2.3's heading; *Create board* on 2.4's heading, for every member; *Invite people* on 2.4's heading, for workspace admins; *Board settings…* in the board ⋯, for the board's owner or a workspace admin; *Add column* in the bar, and in the board ⋯ below 640. Anyone who can't use one doesn't see it (ADR-018 §5).
- **Focus opens on the first field**, and returns to the button that opened the form when it closes.
- **Enter in a one-line field sends the form.** In a field of several lines, Enter starts a new line and Ctrl + Enter (⌘ + Enter) sends. In the address field, Enter makes a chip of what was typed, and sends when nothing is typed.
- **Fields stop at the server's limits**: 255 characters for names and titles, 1,000 for descriptions. So a too-long error never appears (ADR-032 rule 7).
- **The browser trims every value before it checks it and before it sends it.**
- **The Primary button is the form's one action.** It reads *Creating…*, *Saving…* or *Inviting…* while the server answers, in its Busy look. A server failure shows as the `Form message`, the last line of the fields, `spacing.field-gap` below the last field and `spacing.section-gap` above the buttons, and everything typed is kept (ADR-032 §1, §8).
- **Board settings sends only what changed.** With nothing changed, *Save changes* closes with no request. Emptying the description clears it. If a teammate saves first, the last save wins, since the server keeps no versions.
- **Inviting adds people at once and sends no email** (US-04, UC-04). Only someone who already has an account can be added. The answer has a row for every address; HTTP is 200 even when some fail.

### 16. The wordings

| Where | Wording |
|---|---|
| Create a workspace | Title *Create a workspace* · *You can’t rename a workspace or change its description later.* · *Name* · *Description (optional)* · *Create workspace*, busy *Creating…* · errors *Enter a name.* and *Use at least 2 characters.* |
| Create a board | Title *Create a board* · *Title* · *Description (optional)* · *Colour* · *Who can see it*: *Workspace* · *Private* · *Everyone in Product Team can find it and open it.* · *Only people on the board can open it. Workspace admins can see it and join.* · *Create board*, busy *Creating…* · errors *Enter a title.* and *Use at least 3 characters.* |
| Board settings | Title *Board settings* · the same fields, filled in · choosing Private adds *Members of Product Team who aren’t on it lose access, and are taken out if they have it open.* · *Save changes*, busy *Saving…* |
| The colours, spoken | A radio group named *Colour*: *Terracotta*, *Ochre*, *Olive*, *Sage*, *Teal*, *Steel*, *Periwinkle*, *Mauve*. The same names show as a tooltip on hover. |
| Invite people | Title *Invite people* · *They’re added to Product Team straight away. Nobody gets an email, so each person needs a CollaboTask account first.* · *Email addresses*, placeholder *name@example.com, another@example.com* · a chip's ×, spoken *Remove dimas@example.com* · *Invite*, busy *Inviting…* · errors *Enter at least one email address.*, *“rina.w@” isn’t an email address.*, and for several *These aren’t email addresses: “rina.w@”, “budi”.* |
| The results | *Added* · *No CollaboTask account uses this address.* · *Already in Product Team.* · spoken *2 of 4 added to Product Team.* (*All 4 added to Product Team.* · *None of the 4 were added.*) · *Close* after a send |
| Add column | Placeholder *Column name*, spoken *New column name* · heard after: *Column “QA” added.* |
| Every form | *Cancel* · the phone's *Close* · ADR-032's *Can’t reach CollaboTask. Check your connection and try again.* and *Something went wrong on our side. Try again.* · a 403 after your role changed: *Only workspace admins can invite people.* · *Only the board’s owner or a workspace admin can change its settings.* |

*Description (optional)* draws *(optional)* in `color.muted-foreground`.

### 17. What the file draws

**Tokens (220 → 222):** `spacing.control-padding-y`, `8` in `density-fine` and `12` in `density-coarse`.

**Parts, on `Components` (27 → 32):**
- **`Glyph`** gains three looks (19 → 22): `Check` Done (`color.done`), `Check` Muted and `X` Destructive (`color.destructive-text`). The check is the password checklist's outline. Its holding board was widened to 1060.
- **`Input`** gains a second axis, `Content` (Text · Lines · Chips), and five looks (4 → 9). The four it had are `Text`.
  - Lines: Rest and Focus. The `Box` sizes itself to its text, `spacing.control-padding-y` above and below, at least `size.control-height`. Its label is a `Label row`: *Description*, then *(optional)* in `color.muted-foreground`, `spacing.list-gap` apart, because a token can't colour part of a text.
  - Chips: Rest, Focus and Invalid. The `Box` wraps four `Address chip` copies, `spacing.list-gap` from its edge and apart. Invalid keeps the box's `color.input` edge, turns the third chip Invalid, and shows the message.
  - The holding board, `Input — states`, is 4360 wide, with a label over each new look.
- **`Swatch`**, on `Swatch — states` (y 12000). `State` = Rest · Hover · Pressed · FocusVisible · Current. A `size.control-height` square, `radius.control`, `color.popover`, padded `spacing.inline-gap` on every side, holding a `Colour` square that fills it, `radius.badge`, bound to `color.identity.terracotta.strong`. Copies re-bind it to their hue.
- **`Choice row`**, on `Choice row — states` (y 12264). `State` = Rest · Hover · Pressed · FocusVisible · Current, 400 wide (`size.form-width`). A `Head` (the `Radio` circle and `Name`, `spacing.icon-gap` apart, centred on the name's line) over a `Sub` (an `Indent` the circle's width, then `Sentence`). The 16 circle beside a 20 line is centred by the `Head`, not by a typed 2px margin. Current turns `Name`, the edge and the circle plum and shows the `Dot`.
- **`Address chip`**, on `Address chip — states` (y 12544). `State` = Rest · Invalid. `size.target-min` tall, `color.chip`, `radius.badge`, `spacing.inline-gap` on the left; an `Address` text and a `Remove` square, `size.target-min`, holding a `Glyph` X Muted. Invalid adds a 1px `color.destructive` edge.
- **`Invite result`**, on `Invite result — states` (y 12808). `Outcome` = Added · NoAccount · AlreadyIn, 400 wide, built like the `Choice row`'s `Head` and `Sub`: the glyph and the `Address`, then the `Outcome` under it.
- **`Form`**, on `Form — states` (y 13064). `Form` × `Frame`, seven looks: Workspace, Board, Settings, Invite and Invite sent as `Dialog`; Board and Invite as `Sheet`. A Dialog is §1's frame; a Sheet is §2's, 390 wide, sized to its content (the screens stretch it to 704). Each holds `Text`, `Fields` and the buttons, built from copies of `Input`, `Swatch`, `Choice row`, `Invite result`, `Button` and a hidden `Form message`. The colour field is a `Swatches` row of eight, or two `Row`s of four on a Sheet. Invite/Sheet shows the Invalid chips and *Invite* not satisfied.

**The `Overlays` page gains ten boards (17 → 27):**

| Board | Where | Open on it |
|---|---|---|
| *2.3 + create workspace* | x 0, y 6600 | Create a workspace, *Marketing* typed |
| *2.4 + create board* | x 1640, y 6600 | Create a board, olive chosen, focus in *Title* |
| *2.4 + create board, a short title* | x 3280, y 6600 | *Q4*: the error as you type, *Create board* not satisfied |
| *2.5 + board settings* | x 4920, y 6600 | Q3 Launch Plan's settings, Private chosen, the line under it |
| *2.4 + invite people* | x 0, y 7700 | four addresses as chips |
| *2.4 + invite people, sent* | x 1640, y 7700 | Rina's chip kept, the four results, *Close* |
| *2.5 + adding a column* | x 3280, y 7700 | the board scrolled to its end, the new column's field |
| *Create a board at 390* | x 0, y 8800 | the full-height sheet, two rows of colours |
| *Invite people at 390* | x 590, y 8800 | a chip that isn't an address, and its message |
| *2.15 + adding a column* | x 1180, y 8800 | the phone board at its end, the new column's field |

The desktop screens are clones of *2.4 + workspace ⋯* and *2.5 + board ⋯* without their menus, and a detached copy of 2.3 (*Operational note*). The phone sheets are clones of *2.6 at 390* with the board and drawer removed. The board scrolled to its end is drawn by setting the canvas's columns to the end (`justifyContent: end`) with `spacing.screen-x` on its right.

**Also changed:** *Invite people* (an Outline button, after the member count) on 2.4 and its five state boards on `Page 1`, and on *2.4 + workspace ⋯* and *2.4 + user menu*. On the cold load it is hidden, like the member count, because your role is not known yet.

### 18. The code

- **One CSS variable is added**: `--control-padding-y`, `8px` at Fine and `12px` at Coarse, beside `--control-padding-x`.
- **The form dialog is shadcn's `Dialog`, restyled**: `w-[var(--dialog-width)]`, `p-[var(--dialog-padding)]`, the confirmation's surface, `showCloseButton={false}`, and `onInteractOutside={e => e.preventDefault()}`. Below 640 it is shadcn's `Sheet`, full height, with a head holding the title and a *Close* button. One `FormFrame` component switches between them at 640.
- **A field of several lines is shadcn's `Textarea`**, restyled: `field-sizing-content`, `min-h-[var(--control-height)]`, `py-[var(--control-padding-y)]`, `rows={1}`, `resize-none`.
- **The colour choice and the visibility choice are Radix `RadioGroup`s.** A colour item is `size-[var(--control-height)] p-[var(--inline-gap)]` with a `size-full rounded-sm` square inside; the checked one gets a 1px inset `ring-primary`. A visibility item is a row whose sentence is its `aria-describedby`.
- **The address field is a small component of its own**: a list of chips, each with a remove button, and a text input after them. A paste is split on commas, semicolons, spaces and new lines.
- **`hueFromBoard(board)`** reads `background_color` by case-insensitive match against the eight deep values, and falls back to the id. **`firstFreeHue(boards)`** picks the default (§9).
- **Validation runs on the trimmed value.** A minimum length is checked on change; a missing value is checked on submit. The Primary button's not-satisfied look comes from the same check.
- **Add column** keeps a client-only draft column after the last one; Enter or blur with a name calls UC-13, and the draft is replaced by the server's column. The announcement uses a polite live region.
- **Phosphor imports:** `XIcon` (the chip and *Close*), `CheckIcon` (the results).

## Options considered

Every option is named, with its cost, under its decision above, and all were drawn side by side on the decision page. In short:
- **The frame:** a × by the title; a drawer from the right.
- **The phone:** buttons pinned in a foot; the dialog 358 wide.
- **Closing:** pressing outside closes it; asking before Escape or *Cancel*.
- **Several lines:** three lines at rest; one line that never grows.
- **Its padding:** `spacing.card-padding-y` by rule.
- **A minimum:** Login's rule; a hint under the field.
- **The warning:** a hint under each field; none.
- **The description:** 1,000 in the browser only; no limit.
- **The colours:** filled squares with a check; circles with a ring. The square: 20 at both densities; 24 at both, with a new token.
- **The first colour:** always terracotta; none chosen. **Stored as:** a short key.
- **Visibility:** two Segments and one sentence; a checkbox.
- **Settings:** each field saving on its own. **Going private:** asking on save; no warning.
- **Invite:** inside the team panel; both. **Addresses:** one field of several lines; one field per address. **Chips:** fitted to the field, with a 34 × or no × on touch.
- **Results:** taking the form's place; a toast. **Marks:** words only; only the failure in colour.
- **Add column:** the form dialog; a popover.

## Out of scope (deferred, recorded)

- **The team panel**, opened from the member count. It is a list, group d's to draw; *Invite people* doesn't open from it.
- **Hover, pressed and focus on the overlay boards.** The parts carry those looks.
- **Board settings and Create a workspace at 390.** They use the same sheet as Create a board.
- **Where a toast sits.** Group e.
- **Movement.** Forms and sheets open and close in code only (inventory §4.3).
- **The overlays at a tablet's Coarse density.** The desktop boards are drawn at Fine (ADR-040 *Operational note*).

## Downstream edits this ADR requires

1. **[SRS](../../spesifications/001-software-specifications.md):** UC-03's name and description lengths, and §4.3's stored colour strings. Done first.
2. **[ADR-022](adr-022-form-control-geometry.md): annotation.** §3's *no `spacing.control-padding-y`* is narrowed to controls of fixed height.
3. **[ADR-032](adr-032-auth-form-errors-and-password-rules.md): annotation.** §8's rule 1 is narrowed for minimum lengths.
4. **[ADR-021](adr-021-design-inventory-corrections.md): annotation.** §5's forms are the dialog, not a drawer, from 640 up.
5. **[ADR-040](adr-040-menus-and-the-confirmation.md): annotation.** §7's dialog holds the forms; pressing outside does nothing.
6. **[ADR-041](adr-041-the-card-drawers-pickers.md): annotation.** §7's mark also marks a chosen colour and a chosen row.
7. **[ADR-018](adr-018-state-vocabulary.md): annotation.** Emerald's second render site; plum's *where you are* spent on the colour and visibility choices; state 18 stays closed.
8. **[ADR-033](adr-033-workspace-screens.md): annotation.** 2.4's heading gains *Invite people*.
9. **[ADR-017](adr-017-ui-foundations.md): annotation.** §3.3's `background_color` is stored as the deep value.
10. **[ADR-039](adr-039-drag-and-phone-board.md): annotation.** The board ⋯'s *Add column* below 640 makes the new column on the board.
11. **Penpot `Collabotask App`:** done 2026-11-24 (*Operational note*).
12. **[`design-file-inventory.md`](../design-file-inventory.md):** the sources table, §1.5, §3, §5, §7.2, §7.3, §7.6 and *Related documents*.
13. **[`frontend-architecture.md`](../frontend-architecture.md):** the stored colour, the form frame, and the trimming and address rules.
14. **`CLAUDE.md` › `## Now`:** a pointer.
15. **`docs/build-log.md`:** no entry. Stage 5 is not closed.
16. **PRD, user stories, `CONTEXT.md`:** no change.

## Operational note

**Drawn 2026-11-24.** The parts were built at Light + Coarse, as the file was found. The desktop overlays were placed at Light + Fine and the phone ones at Light + Coarse. The file was left in Light + Coarse.

**Checks.**
- **Binding check, parts:** nothing on `Swatch — states`, `Choice row — states`, `Address chip — states`, `Invite result — states`, `Form — states`, `Glyph — states` or `Input — states`. On `Card — states`, only its seven `⋯` rows.
- **Binding check, Overlays:** only the card `⋯` rows §7.5 expects (21 on *2.5 + board settings* and *2.5 + adding a column*, 15 on *2.15 + adding a column*), and nothing on the other seven new boards, *2.4 + workspace ⋯* or *2.4 + user menu*.
- **Binding check, Foundations:** its expected 53 rows, after the density table gained `spacing.control-padding-y` and the catalogue's heading became *222 tokens*.
- **The looks:** Light + Coarse on the `Form` looks *Settings/Dialog*, *Invite sent/Dialog* and *Board/Sheet*, and on *Invite people at 390* and *2.15 + adding a column*; Light + Fine on *2.4 + create board, a short title* and *2.5 + adding a column*; Dark + Fine on *2.5 + board settings* and *2.4 + invite people, sent*. In Dark the colour squares are the pale values, as ADR-017 §3.3's role swap says, and the dialog's edge, the plum marks and the emerald checks all show.

**What drawing found:**
- **`penpot.createPath()` gives the path a black 1px stroke.** The two new checks drew black in exports while their data said emerald and grey. Their strokes were removed. The older glyphs have none.
- **A screen can cross pages without the user pasting it.** 2.3 was cloned on `Page 1`, the clone made a temporary part, a copy of it placed on `Overlays` (it landed inside the board at the origin, *2.4 + workspace ⋯*, and was moved to the root at once), the copy detached, and the part deleted from `Page 1`, which removed the clone. The part count went 32 → 33 → 32.
- **A hidden child still takes room in a flex box.** An empty description with its `Value` hidden stayed 64 tall. A one-space text keeps it at `size.control-height`.
- **Setting `layoutChild.maxHeight` dropped the box's `layoutItemMinH` binding**, and a `height` binding survived `verticalSizing: auto` until the box was resized to another value. Both were read back and fixed.
- **A look's `Box` lost its fill width after its edits**, so the phone sheet's description field stayed 400 inside 358. Found by reading every field's width on the Sheet looks.
- **The board scrolled to its end** is drawn with the canvas's `justifyContent` set to `end` and `spacing.screen-x` on its right, so no offset is typed.
- **The tab slept on every Theme and Density switch.** Each switch landed, and the tab woke when the user focused it. The documents were written while it slept.
- **Switching a nested copy's variant kept its texts only partly**: the Input's label held, its value and message went back to the main's. They were set again.
