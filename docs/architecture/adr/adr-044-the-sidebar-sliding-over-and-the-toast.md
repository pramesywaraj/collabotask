# ADR-044: The Sidebar Sliding Over, and the Toast — Stage 5, Group e

- **Status:** Accepted
- **Date:** 2026-11-24
- **Scope:** Question 4 for Stage 5's group e, the last of Stage 5: the sidebar sliding over the content below 1024 (ADR-034's *navigation* kind), and the toast's content and placement, with the archive toast's *Undo*. It settles:
  - what the file draws for the slide-over, where focus goes when it opens, and how the user menu opens from it on a phone
  - where a toast sits, the space between it and the screen's edge, and where it sits while the card drawer is open
  - how toasts stack, and how long one stays
  - what *Undo* looks like, where it leaves you, and what it does for teammates on 2.12
  - how a failure toast reads, and what the board does after a refused change
  - the wordings for states 4, 27 and 32, which ADR-030 left open
  - what the file draws for the toast

  **No token is added (222), and no part (34).** Question 4 fired twice and both times resolved by rule: the space between a toast and the screen's edge (§5), and, after the user's stacking pick, the space between two stacked toasts (§6). 21 of the 22 stages asked have now fired. **Fifteen choices had never been made.** The user took fourteen recommendations and chose a stack of up to three over one toast at a time (E07 B). The `Toast` part gains a hidden *Undo*, an Outline `Button` copy that its copies show, and its board gains a stack and an *Undoing…* look. The `Overlays` page gains ten boards (37 → 47).

  It **settles [ADR-030](adr-030-toast.md)'s *Out of scope***: where a toast sits, how two of them behave, how long one stays, and the wordings for 4, 27 and 32. It **draws [ADR-040](adr-040-menus-and-the-confirmation.md) §5's *Undo*** and **narrows ADR-040 §4** for one sentence: Undo now brings back teammates still on 2.12 (§11). It **clarifies [ADR-034](adr-034-side-menu.md) §4's** *it can never stack on another layer*: nothing opens the panel from inside another layer, but one layer can open over it, the user menu's bottom sheet below 640 (§3). It **draws ADR-034's slide-over**, which ADR-034 and [ADR-035](adr-035-sidebar-loading-failure-and-switcher-mark.md) left to Stage 5. **With it, Stage 5 is closed.**
- **Upstream edits:** two, made first. [US-12c](../../product/001-user-stories.md) gains two acceptance lines: Undo takes you back to the board, and teammates still on the "no longer available" screen are taken back too (§10, §11). [SRS UC-12c](../../spesifications/001-software-specifications.md) records what §11 relies on: archiving evicts no one, and a rejoin of an archived board's room is refused. **Neither changes the backend**, so the open backend items stay at five.

## Context

Inventory §7.3 draws Stage 5 in five groups, each settling its kind of container once. Group e is the last. Its two pieces share nothing but their timing: the sidebar's slide-over panel was left by [ADR-034](adr-034-side-menu.md) to *"Stage 5, with the user menu it holds"*, and the toast's placement, stacking and remaining wordings were left by [ADR-030](adr-030-toast.md) to whichever stage first drew a toast on a screen. [ADR-040](adr-040-menus-and-the-confirmation.md) §5 then kept US-12c's *Undo* after archiving and sent it here.

**Question 4 was asked of every value group e needs, against the live 222 tokens, before anything was drawn.** The panel is the `Sidebar` part as it stands, with `shadow.overlay` over `color.scrim`. The toast is ADR-030's part, and *Undo* fits the `Button`'s Outline kind with no change. **One value had no token:** the space between a toast and the screen's edge. **Fifteen choices had never been made.** One was a gap between two accepted ADRs (§3), and two came from the backend code (§11, §13). All fifteen went on one rendered page, drawn in Light and Dark at Fine on the desktop and at Coarse on a 390 phone wherever the phone differs. The user took fourteen recommendations and chose a stack of up to three (E07 B). That pick needed one more value, the space between two stacked toasts, so question 4 fired a second time, and it resolved by the same rule as the first.

Seven facts were found in the backend code:
- **Access revoked carries `{ board_id, reason }` and nothing else**, so no toast can name who did something. Its three reasons and ADR-030 §5's wordings stand.
- **Undo is `POST …/board/:id/archive` with `is_archived: false`**, the request archiving uses. It needs the board's owner or a workspace admin, it skips the access checker so it works on an archived board, and a repeat is a no-op 200.
- **Archiving evicts no one.** It sends `BOARD_ARCHIVED` and the room stays connected, so a teammate still on 2.12 hears `BOARD_UNARCHIVED`. A rejoin of an archived board's room is refused, silently.
- **What a refused board change returns:** 404 when the card, its column or the board is gone (an archived board too) or the card moved column meanwhile; 409 for a move across boards; 403 `FORBIDDEN` when you may no longer edit; 403 `BOARD_JOIN_REQUIRED` when the board went private and you are an admin who hasn't joined. The 404s look alike, so the browser cannot tell a gone card from a gone board.
- **A workspace admin edits a workspace board without joining it, and workspace roles have no live message.** An admin who is made a member learns of it at their next edit, as a 403. This is state 27's real case.
- **When a board goes private, workspace admins stay in its room.** A non-joined admin's next edit is 403 `BOARD_JOIN_REQUIRED`.
- **A request that times out may still land.** axios stops waiting at 10 s (ADR-018 §3); if the server made the change, its live message then puts the card where you dropped it.

## Decision

### 1. The slide-over: the `Sidebar` part over a dimmed screen, drawn on the phone (E01)

- **It is the sidebar as it stands** (ADR-034 §4): the `Sidebar` part, `size.sidebar-width` 240, `color.background`, `shadow.rule-right`, full height between the browser's bars (704 at 390 × 844), sliding in from the left. It takes `shadow.overlay` (ADR-034 §10.8) on a `Sidebar panel` wrapper filled `color.background`, since the part already binds its rule as a shadow. `color.scrim` covers the app bar and the board; 150 of dimmed board shows beside it at 390. Against the dimmed canvas the panel measures 2.64:1 in Light and 1.04:1 in Dark, so in Dark its rule and the shadow are its edge.
- **Its button lies exactly over the bar's**, 16 from the edge, and reads *Hide sidebar*.
- **The file draws it on the phone only** (E01 A): *2.15 + sidebar* and *2.15 + sidebar, switching workspace*, at Coarse. The switcher opens inside the panel (ADR-034 §7), with ADR-035 §3's dot on Product Team, so two rows are plum: the selected workspace and Q3 Launch Plan.
- The rejected option: a tablet in portrait as well, 834 × 1194, a screen width the file draws nowhere else, for a panel that looks the same.

### 2. Focus opens on the current row (E02)

- **When the panel opens, focus goes to the current row**: *Q3 Launch Plan* on a board, *All boards* on 2.4. A screen reader starts at where you are; Tab goes down the list and Shift + Tab up to the switcher and the button.
- It is a modal dialog named *Sidebar* (ADR-034 §9): Tab stays inside it, and focus returns to the button when it closes.
- The rejected option: the button, the first thing in it, which announces *Hide sidebar* rather than where you are.

### 3. Below 640, the user menu is the bottom sheet over a second dimming (E03)

- **Found while mapping.** ADR-034 §4 says the panel *never stacks on another layer*, and ADR-040 §9 makes the user menu a bottom sheet below 640. On a phone the panel is the only way to the user menu, so the two meet.
- **ADR-034 §4 meant that nothing opens the panel from inside another layer**, and that stays true. One layer can open over it: the user menu's bottom sheet, over a second `color.scrim` that dims the panel too, as a drawer is dimmed under its delete dialog (ADR-040 §7). That is within [ADR-043](adr-043-the-lists.md) §5's rule: a full-height layer and one more, a bottom action sheet or a dialog. Every menu below 640 stays a sheet.
- **From 640 to 1023** the user menu opens upward from the row as the popover (ADR-040 §11), inside the panel, since 224 is the row's width.
- Closing the sheet leaves you in the panel.
- The rejected options: the desktop popover inside the panel below 640, which narrows ADR-021 §5's menus row for one trigger; and closing the panel before the sheet opens, which loses the panel you were in.

### 4. A toast sits at the bottom right; on a phone, across the bottom (E04, E06)

- **One rule at every width: 16 from the bottom and the right edge of the screen.** On a 390 phone the 358 toast then spans the screen less 16 at each side, 16 above the browser's bar.
- **It is the top layer**: over the card drawer, a dialog's dimming and the slide-over panel.
- **While the card drawer is open it stays in the same corner, over the drawer's foot** (E06 A), covering *Delete card* and the timestamps until it is closed.
- The rejected options: bottom centre of the content, over the middle columns where the work is; top centre under the app bar, over the column headers and far from the thumb; and beside the drawer while it is open, a second position plus a third rule for windows too narrow for one.

### 5. The space to the edge is `spacing.screen-x`, by rule (E05, question 4)

- **16, `spacing.screen-x`, at the side and the bottom, at every width.** The toast's width is already `390 − 2 × spacing.screen-x` (ADR-030 §2), so on a 390 phone it is the only value that fits. **No token.**
- **It is a placement, drawn by position**, as the menu's 4 below its ⋯ is (ADR-040 §2). The code writes it as `var(--screen-x)`.
- The rejected option: `spacing.toast-offset` 24 from 640 up, Sonner's desktop offset, a token for one placement that changes at a breakpoint, which no token set can express (a tablet is Coarse too).

### 6. Toasts stack, up to three, newest at the bottom (E07)

- **Up to three show at once, each its own box, newest at the bottom**, nearest the corner. The stack grows upward. Each × closes its own. **This was the user's call** over one at a time, the newest replacing the one showing. Its cost was on the page: three boxes cover about 230 of the board's corner, and on a phone a quarter of the screen.
- **A fourth removes the oldest**, the top one. The same wording twice shows once, so `removed_from_workspace`, which comes once per board room you were in, is one toast (ADR-030 *Out of scope*).
- **8 apart, by the same rule as §5** (question 4's second firing): a placement in code, drawn by position, as the decision page drew it. No token.
- **Going to another page closes the stack**, except the toast that the navigation itself brings (§7).
- The rejected options: one at a time (recommended), where a new toast replaces an *Undo*; and a queue, where an old message shows after its moment has passed.

### 7. A toast stays until you close it, another pushes it out, or you leave the page (E08)

- **No timer**, so WCAG 2.2.1's time limits never apply, and *Undo* waits for as long as you stay on 2.4. PRD §5.4 targets WCAG 2.2 AA.
- **A toast that comes with a navigation shows on the page it lands on**: access revoked (26) after you are taken out, the archive toast after you land on 2.4, *was unarchived* after a teammate is brought back (§11).
- The cost, accepted: a message stays over the board's corner until it is closed; on a phone, over a long column's sticky *Add card*.
- The rejected options: 6 seconds and 10 for *Undo*, waiting while pointed at or focused, a time limit that someone reading slowly or using a screen reader can miss; and timed messages with a waiting *Undo*, two behaviours, where an access-revoked reason can vanish before it is read.

### 8. *Undo* is the Outline button (E09)

- **A copy of the `Button`'s Outline kind, unchanged**, between the message and the close, `spacing.inline-gap` apart. Its fill is the page's, so it is white on the Light toast and near-black on the Dark one: 10.27:1 and 15.73:1 against the toast, its label 17.49:1 and 18.92:1. Its hover, pressed, focus and Busy looks already exist; Busy (*Undoing…*) is 9.14:1 / 9.67:1 against the toast. **No new value.**
- The toast grows from 48 to 60 at Fine with *Undo* (68 at Coarse, as before).
- The rejected options: underlined text, a new part with every look still to draw and a reading as a link; and plum text, 1.57:1 in Light and 2.28:1 in Dark against the toast.

### 9. The archive toast

- After you archive the board you're on, you land on 2.4 (ADR-040 §13) with ***“Q3 Launch Plan” is archived.*** and *Undo*. The board has left the grid and the sidebar.
- *Undo* reads *Undoing…*, the Outline's Busy look, until the server answers.

### 10. After *Undo*, you are back on the board (E10)

- **Undo puts you back where you were**: the board opens again, and the toast closes. A screen reader hears *“Q3 Launch Plan” is back.* This is how Gmail's *Undo* returns you to the message.
- **If Undo fails**, the toast is replaced by *Couldn’t bring back “Q3 Launch Plan”. Try again from the Archived list.* A 200 with no change, because someone else already unarchived it, counts as done.
- [US-12c](../../product/001-user-stories.md) gained the line *And Undo takes me back to the board*, first.
- The rejected option: staying on 2.4 with the board back in the grid, one more press to where you were.

### 11. Teammates on 2.12 are brought back (E11)

- **2.12 stays in the board's room while it shows *This board was archived.*, and on `BOARD_UNARCHIVED` it opens the board again** with the toast *“Q3 Launch Plan” was unarchived.* Browser only; the backend already sends the message to the room (§*Context*).
- **It is best effort.** Someone whose connection dropped while the board was archived can't rejoin its room, so they stay on 2.12. Someone who left 2.12 sees nothing.
- **This narrows [ADR-040](adr-040-menus-and-the-confirmation.md) §4's sentence** *your Undo does not bring them back*, which was written as a fact. Archive still asks first: a teammate's drag or typing is still interrupted.
- US-12c gained *And if I undo while they're still on that screen, they're taken back to the board too*, and SRS UC-12c records that archiving evicts no one. Both first.
- The rejected option: 2.12 as ADR-040 §4 had it, where a teammate reads that a board is gone while it exists again.

### 12. A failure toast names what didn't happen, then why (E12)

- ***Couldn’t* + a verb + the thing's name, then one of five reasons** (§16): *Couldn’t move “Final pricing sign-off”. Someone changed the board at the same time.* The verbs are *add*, *rename*, *change*, *move* and *delete*. This follows ADR-030 §5's rule that each toast names what you lost, because the toast sits in a corner, away from the card, and several changes can be in flight.
- The change itself goes back at once, not animated (ADR-018 §8 rule 1), and loses its dim.
- **Forms and lists don't toast.** Their failures stay in their own `Form message` (ADR-032 §1, ADR-042, ADR-043).
- The rejected options: one first line for every failure, *Your change didn’t save.*, where nothing says which change bounced; and the reason only, which doesn't say anything went back.

### 13. After a refused change, the board reloads (E13)

- **After a 403 or a 404, the board, its columns, its members and the workspace's detail load again**, so the screen shows what is true: 2.7's banner with *Join* after a 403, 2.8 after `BOARD_JOIN_REQUIRED`, 2.12 if the board is gone. Live messages already correct what teammates changed; nothing tells the page that a role changed.
- After a 409, a timeout or a 5xx nothing reloads: the snapshot is enough.
- The rejected option: the toast only, so after a role change every next edit fails the same way.

### 14. What the file draws

- **No token: 222. No part: 34.**
- **`Toast`**: its main gains a hidden `Undo`, an Outline `Button` copy labelled *Undo*, between `Message` and `Close`. A copy shows it. ADR-030 §3's *one appearance, no axis* stands: the part has no variant axis, as the `Confirmation` holds a hidden field.
- **`Toast — states`** is widened to 1710 with two `Doc / Page, 390` boards: *a stack of three* (the archive toast with *Undo*, then a refused move, then a rename that couldn't reach the server, 8 apart, newest at the bottom), and *Undoing…* (the Outline's Busy look). Its three Card copies' `Meta row` gaps, typed 8 since ADR-037 §7, are bound to `spacing.inline-gap`, as ADR-038 did for the Foundations cards.
- **Ten boards on `Overlays`, 37 → 47**, all clones made on the page except 2.7 (*Operational note*):
  - y 13200, Fine: *2.4 + archived, with Undo*, *2.5 + a move that didn’t stick* (4), *2.7 + an edit you can’t make* (27, after the reload) and *2.6 + a change that didn’t save* (over the drawer)
  - y 14300, Fine: *2.5 + the open card deleted* (32; the card is gone from In Progress, whose chip reads 2) and *2.3 + removed from a workspace* (26; Design Guild's tile is gone)
  - y 15400, Coarse: *2.15 + a move that didn’t stick*, *2.15 + sidebar*, *2.15 + sidebar, switching workspace* and *2.15 + sidebar, user menu*
- Every toast is an absolute child of its screen, placed by position 16 from the bottom and right: the desktop ones for Fine, the phone one for Coarse inside `Main`, between the browser's bands.

### 15. The behaviour rules

- **Where toasts come from**: a refused board change (4, 27), the open card deleted by someone else (32), access revoked (26), archiving (the *Undo* toast), and a teammate's board coming back (§11).
- **The open card deleted (32)**: the drawer closes and the toast names the card. If its column was deleted, the client knows from `COLUMN_DELETED`, and the toast says so (§16).
- **A toast never takes focus.** Its *Undo* and × are reachable from the keyboard (*The code*). Announcing it without your action is SC 4.1.3, which PRD §5.4 keeps open.
- **The slide-over** closes when you pick a row (a board, *All boards*, a workspace, *All workspaces*), press the dimmed strip, press Escape or press its button. When its list is taller than the screen, it scrolls between the top row and your row, which stay. **Crossing 1024 while it is open closes it**, and from 1024 up the remembered choice applies. Under reduced motion it cross-fades (ADR-034 §10.10). Ctrl + \\ does nothing during a drag (ADR-034 §9).

### 16. The wordings (E15)

| Where | Wording |
|---|---|
| The archive toast | *“Q3 Launch Plan” is archived.* · *Undo*, spoken *Undo archiving “Q3 Launch Plan”*, busy *Undoing…* · the close, spoken *Close* |
| After Undo | back on the board; heard *“Q3 Launch Plan” is back.* · if it fails: *Couldn’t bring back “Q3 Launch Plan”. Try again from the Archived list.* |
| A teammate on 2.12 | back on the board, with *“Q3 Launch Plan” was unarchived.* |
| A change that didn't save | *Couldn’t add “Write the FAQ”.* · *Couldn’t rename “Review”.* · *Couldn’t change “Final pricing sign-off”.* (a title, description, assignee or due date) · *Couldn’t move “Final pricing sign-off”.* · *Couldn’t delete “Review”.*, then one of the five reasons below |
| Why: the board changed (4) | 404 or 409: *Someone changed the board at the same time.* |
| Why: not allowed any more (27) | 403: *You can’t edit this board any more.* · 403 `BOARD_JOIN_REQUIRED`: *It’s private now. Join it to keep editing.* |
| Why: no answer | no answer, or a timeout: *Can’t reach CollaboTask. Check your connection and try again.* ([ADR-032](adr-032-auth-form-errors-and-password-rules.md)) |
| Why: anything else | *Something went wrong on our side. Try again.* (ADR-032) |
| The open card deleted (32) | *“Final pricing sign-off” was deleted.* · with its column: *“Final pricing sign-off” was deleted with its column.* |
| Access revoked (26) | [ADR-030](adr-030-toast.md) §5, unchanged |
| The panel | [ADR-034](adr-034-side-menu.md) §12, unchanged: *Show sidebar* · *Hide sidebar* · the dialog is named *Sidebar* |

Names are quoted and wrap; a very long one is cut short in code (ADR-030 *Out of scope*).

### 17. The code

- **No CSS variable is added.**
- **The toaster is shadcn's Sonner**, re-pointed as ADR-030 §8 says: `position="bottom-right"`, both offsets `var(--screen-x)`, `gap={8}`, `visibleToasts={3}`, `expand` (always shown apart, never piled), `duration={Infinity}`, and ADR-030's own close button. Sonner's mobile layout already spans the screen below its breakpoint; check it gives 16 at each side. **Confirm the prop names at scaffold time**, as ADR-030 §8 warns. A repeated message is deduplicated by an `id` per wording.
- ***Undo* is the `Button`'s `outline` variant** (ADR-033 §14) rendered as the toast's action. Pressing it posts `{ is_archived: false }`, shows the busy look, then invalidates `['boards', wsId]` and navigates to the board.
- **One error helper for board mutations**: restore the snapshot, call `failureToast(verb, name, error)`, and after a 403 or 404 invalidate `['kanban', boardId]`, `['board', boardId]`, `['board-members', boardId]` and the workspace's detail. The reason comes from the response: no response or `ECONNABORTED`, 404/409, 403 with `BOARD_JOIN_REQUIRED`, other 403, anything else.
- **2.12 keeps the board's room**: the board route stays joined while it renders the gone state, and `BOARD_UNARCHIVED` for that board refetches it and shows the toast.
- **The slide-over** is ADR-034 §14's `Sidebar` sheet below 1024: `onOpenAutoFocus` moves focus to the row with `aria-current`, and the overlay is `bg-scrim`. Below 640 the user menu is a bottom `Sheet` opened inside it; Radix stacks the second dialog's focus trap and overlay over the first.

## Options considered

Every option is named, with its cost, under its decision above, and all were drawn side by side on the decision page. In short:
- **The slide-over:** a tablet board as well. **Focus:** the button. **The user menu below 640:** the popover inside the panel; closing the panel first.
- **Where:** bottom centre; top centre. **The edge:** a 24 token. **Over the drawer:** beside it.
- **Two at once:** one at a time (recommended, not chosen); a queue. **How long:** timed; timed with a waiting *Undo*.
- ***Undo*:** underlined text; plum text. **After:** staying on 2.4. **Teammates:** 2.12 stays.
- **Failures:** one first line for all; the reason only. **After a refusal:** the toast only.
- **The file:** two toast boards.

## Out of scope (deferred, recorded)

- **Hover, pressed and focus on the overlay boards**, and on the toast's close. The parts carry *Undo*'s looks; the close's are still those of an icon button on an inverted surface, not yet a part (ADR-030).
- **The slide-over at a tablet's width**, and on 2.4 at 390: the same column over a different screen.
- **Movement**: the slide (ADR-034 §10.10) and the toast's 260ms arrival (ADR-018 §8). Code only (inventory §4.3).
- **SC 4.1.3**, announcing a toast without your action. PRD §5.4 keeps it open.
- **The teammate's return when their connection dropped** while the board was archived. Bringing them back would need the server to let a socket wait on an archived room. Not in Phase 1.
- **Truncating a long name inside a wording.** Code.

## Downstream edits this ADR requires

1. **[US-12c](../../product/001-user-stories.md):** two acceptance lines. Done first.
2. **[SRS](../../spesifications/001-software-specifications.md) UC-12c:** archiving evicts no one; an archived room refuses a rejoin. Done first.
3. **[ADR-030](adr-030-toast.md): annotation.** *Out of scope*'s placement, stacking, duration and wordings are settled; §3's part gains a hidden *Undo*.
4. **[ADR-040](adr-040-menus-and-the-confirmation.md): annotation.** §4's sentence on teammates is narrowed; §5's *Undo* is drawn.
5. **[ADR-034](adr-034-side-menu.md): annotation.** §4's *never stacks* is clarified; the slide-over is drawn.
6. **[ADR-018](adr-018-state-vocabulary.md): annotation.** *Out of scope*'s toast placement is settled; §10's rows 27 and 32 have wordings.
7. **[ADR-021](adr-021-design-inventory-corrections.md), [ADR-035](adr-035-sidebar-loading-failure-and-switcher-mark.md), [ADR-038](adr-038-board-states-group-d.md), [ADR-041](adr-041-the-card-drawers-pickers.md), [ADR-042](adr-042-the-forms.md): annotations.** What each left to group e is settled.
8. **[`design-file-inventory.md`](../design-file-inventory.md):** the sources table, §3, §4.2, §5, §7.2, §7.3, §7.6, the appendix and *Related documents*.
9. **[`frontend-architecture.md`](../frontend-architecture.md):** the toaster, the error helper, *Undo*, 2.12's room, and the slide-over's focus.
10. **`CLAUDE.md` › `## Now`:** a pointer. Stage 5 is closed; Stage 6, the final check, is next.
11. **`docs/build-log.md`:** an entry. Stage 5 is closed.
12. **PRD, `CONTEXT.md`:** no change.

## Operational note

**Drawn 2026-11-24.** The phone boards were placed at Light + Coarse and the desktop ones at Light + Fine. The file was left in Light + Coarse, as it was found. The last switch, back to Coarse, was made with `Page 1` open after a reconnect, so open `Overlays` in the editor before exporting from it (ADR-043 *Operational note*).

**Checks.**
- **Binding check, parts:** on `Toast — states`, only its three Card copies' `Menu → typography(13px)`, the rows §7.5 expects, after their `Meta row` gaps were bound.
- **Binding check, Overlays:** nothing on *2.4 + archived, with Undo* or *2.3 + removed from a workspace*; only card `⋯` rows elsewhere: 21 on *2.5 + a move that didn’t stick*, *2.6 + a change that didn’t save* and *2.7 + an edit you can’t make*, 20 on *2.5 + the open card deleted*, and 15 on each 2.15 board.
- **Layout:** the check that compares every auto-sized board with its children, gaps, padding and margins found nothing on the ten boards.
- **The looks:** Light + Coarse on the four phone boards and `Toast — states`; Light + Fine on the six desktop boards; Dark + Fine on *2.4 + archived, with Undo* (a `#E7E5E4` toast with a near-black *Undo*) and *2.15 + sidebar* (the panel `#0C0A09`, its rule its only edge against the dimmed board, as ADR-034 §10.8 measured). After each switch, *Add column* still measured 86, so no text width went stale.

**What drawing found:**
- **A theme switch on `Overlays` blocks the tab for minutes**, now that the page holds 47 screens. The switch call timed out and the plugin lost its heartbeat, twice, but the switch landed each time, and the plugin's storage survived a reconnect. Make each switch with the Penpot window beside VS Code, return from the call at once, and read the result in a later call.
- **Moving tiles between grid cells by `layoutCell` can leave one drawn in its old place** while it reports the new one. Leadership offsite read row 2, column 1 and drew at column 4. Re-appending it with `grid.appendChild(shape, row, col)` placed it.
- **`instance()` lands inside the board at the canvas origin**: on `Components` that was the `Card — states` holding board this time, and on `Overlays` it is *2.4 + workspace ⋯*. Each copy was moved in the same call, and both boards were checked for strays.
- **2.7 came across as a temporary part**, as in ADR-042: a clone on `Page 1` made a part, a copy placed on `Overlays` and detached, then the part removed with its clone. The library is back to 34.
- **`switchVariant` gives a nested copy the new variant's words**: after the switch to Busy, the *Undo* copy read the Busy main's *Joining…*, and *Undoing…* was set in a separate call.
