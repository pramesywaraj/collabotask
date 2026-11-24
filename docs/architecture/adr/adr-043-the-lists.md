# ADR-043: The Lists — Stage 5, Group d

- **Status:** Accepted
- **Date:** 2026-11-24
- **Scope:** Question 4 for Stage 5's group d: the workspace team list, board members, adding people to a board, handing a board over, and the presence list the cluster opens. It settles:
  - the list's frame, once, for every list: group c's 448 dialog from 640 up and the full-height sheet below, and how tall a list gets before it scrolls
  - a person's row: what it shows, in what order, and where its actions live
  - which actions ask first, and where *Leave workspace* lives
  - the workspace owner's row, and a server guard that was missing
  - adding people as a second group in the board's list, one person at a time
  - handing a board over, telling an owner how to leave, and taking an owner off a board
  - a leave that open lists can see, a second server change
  - the presence list: its frame and what it shows
  - two places where the design file contradicted itself
  - every wording, and the behaviour rules

  **No token is added (222).** Question 4 fired on one value, how tall a list gets before it scrolls, and it resolved by rule (§2): 20 of the 21 stages asked have now fired. **Twenty choices had never been made.** The user took nineteen recommendations and chose the role on the second line, with the email (D03 B). **Two parts are added (32 → 34):** `Person row` and `List`. The `Glyph` part gains six looks (22 → 28) and the `Menu` part six (10 → 16). The `Confirmation` board gains six wordings, and the `Overlays` page gains the screens in §20.

  It **supersedes [ADR-021](adr-021-design-inventory-corrections.md) §5's placement of the presence overflow list** in *Forms and lists*, and nothing else in that table: it is menu-kind, a popover from 640 up and a bottom sheet below (§15). It **takes two rows out of inventory §3's overlays**: *Add people* is the second group of the board's list (§9), and *Transfer ownership* is a confirmation (§11). It **widens [ADR-040](adr-040-menus-and-the-confirmation.md)** three times: the workspace ⋯ appears for every member, with *Leave workspace* for everyone but the owner (§7); rose marks three more entries (§6); and ten actions now ask first, not five (§6). It **extends [ADR-042](adr-042-the-forms.md) §1–§3**: the form frame holds lists, with *Close* in place of the form's buttons. It **extends [ADR-041](adr-041-the-card-drawers-pickers.md) §11's filter** to every list of people. It **settles [ADR-031](adr-031-presence-cluster.md)'s *Out of scope* overflow list.** It **corrects the file's data** on 2.3, 2.4, the invite screens, the delete-workspace wording and two board ⋯ screens (§17).
- **Upstream edits:** two, both in the [SRS](../../spesifications/001-software-specifications.md), made first. **UC-06 gains the owner guard** (§8): removing the workspace owner is 403, as demoting the owner already is. **UC-10 gains `MEMBER_REMOVED`** (§14), and §4.5's eviction text and §5's event table follow. Both are marked *not built yet*, so **the open backend items go from three to five** (`CLAUDE.md` › `## Now`).

## Context

Inventory §7.3 draws Stage 5 in five groups, each settling its kind of container once. Group d is the lists. It reuses group c's frame, the 448 dialog from 640 up and the full-height sheet below (ADR-042 §1–§2), and the user menu's *Identity* row (ADR-040 §11): an Avatar, then a name over an email, `size.control-height` tall. At Fine those two lines are 20 + 16 = 36, exactly the control's height; at Coarse 44 leaves 4 above and below.

**Question 4 was asked of every value the lists need, against the live 222 tokens, before anything was drawn.** Every value mapped but one: nothing caps a dialog's height, and a list can be long. **Twenty choices had never been made.** Two of them changed the API contract, and one moved a row of ADR-021 §5's table. All twenty went on one rendered page, drawn in Light and Dark with the file's values, at Fine on the desktop and at Coarse on a 390 phone wherever the phone differs. The user took nineteen recommendations and chose the role on the second line, with the email (D03 B).

Five facts were found in the backend code:
- **The workspace owner is always an admin.** The owner is created as one, and the server refuses to demote the owner or let the owner leave (both 403). Once §8's guard exists, the two "last admin" refusals (409) can never fire for anyone else, so no list needs a message for them.
- **Nothing stopped an admin removing the workspace owner.** `RemoveMember` refuses only yourself; it has no owner check and no test for one. SRS §2.6 says a workspace must always have an owner, and only the owner can delete it. §8.
- **An admin can take a board's owner off the board, by design.** The board is then *owner-less but safe* (SRS §2.6), and a test covers it. §13.
- **Leaving a board told the room only `USER_LEFT`**, the message for a closed tab. Removing someone sends `MEMBER_REMOVED`, and so does leaving a workspace. §14.
- **Adding people to a board is all-or-nothing, and skips you without a word.** One person already on the board fails the whole request (409), and a request holding only you is a 400. §9, §10.

Two more facts shaped the lists. The candidate list (UC-11b) answers only a board's owner or a workspace admin; everyone else gets 403. And only board membership is live: `MEMBER_ADDED`, `MEMBER_REMOVED` and `OWNERSHIP_TRANSFERRED` reach everyone with the board open, while workspace roles change by plain requests with no room. Presence messages carry a name and a face, never an email (ADR-014).

The design file contradicted itself twice in ways these lists would have put on screen (§17).

## Decision

### 1. The frame from 640 up: group c's 448 dialog (D01)

- **Every list is the form dialog** (ADR-042 §1): `size.dialog-width` 448, `spacing.dialog-padding` 24, `color.popover`, a 1px `color.border` edge, `radius.surface`, `shadow.overlay`, over `color.scrim`. Inside, `spacing.section-gap` apart: the title in `type.title` and a sentence in `type.body` and `color.muted-foreground`, `spacing.field-inner-gap` apart; then the list, with the filter `spacing.field-gap` above it when there is one; then an Outline *Close* at the right.
- **Below 640, the full-height sheet** (ADR-042 §2): 390 × 704, `spacing.screen-x` inside, the title and *Close* in its head. A list sends nothing, so there is no button at its foot.
- **Closing works as a form's** (ADR-042 §3): pressing outside does nothing, and Escape or *Close* closes it.
- Rows fill the list's 400. Their avatars start `spacing.control-padding-x` in from the title, as a choice row's circle does in a form.
- The rejected option: a drawer from the right, 560 wide and full height, which would make two frames for overlays from 640 up, and the drawer has meant *a card* so far.

### 2. How tall a list gets: eight rows, then it scrolls (D02, question 4)

- **The list's own height stops at 8 × `size.control-height` + 7 × `spacing.list-gap`: 316 at Fine, 380 at Coarse.** Then it scrolls, and the title, the sentence, the filter and *Close* stay where they are. A group's label scrolls with its rows and counts toward the height.
- **It ties to [ADR-041](adr-041-the-card-drawers-pickers.md) §11's eight.** A list of more than eight people shows the filter, and scrolls. The filter now applies to every list of people, counting everyone in the list, both of §9's groups together.
- **No token.** The code writes the height as a sum of the two variables (§21), and the file draws a capped list by showing its first eight rows (§20).
- Below 640 the sheet is the cap: its head stays, and the rest scrolls.
- The rejected options: a dialog that grows until it is 24 from the screen's top and bottom, 852 tall on a 900 screen and changing height as you filter; and a token, `size.list-max-height` 480, for one height, which holds a different number of rows at each density.

### 3. What a person's row shows: the role, then the email (D03, D04)

- **A row is the `Person row` part**: an Avatar at `size.avatar`, `spacing.icon-gap`, then the name in `type.body` over a second line in `type.meta` and `color.muted-foreground`; at the end a ⋯ or an *Add* button. `size.control-height` tall, `spacing.control-padding-x` on the left, `radius.control`, `color.popover`, rows `spacing.list-gap` apart. With nothing at the end, `spacing.control-padding-x` on the right too.
- **The second line is the role, then the email**: *Admin · arif@example.com*. This was the user's call (D03 B) over the role at the row's end. The role is first, so a long email is what gets cut short with an ellipsis. The roles don't line up in a column, and the row's end holds only the ⋯.
- **Every row says its role** (D04): *Owner · Admin · Member* in the team, *Owner · Member* on a board ([ADR-033](adr-033-workspace-screens.md) §11). The board list doesn't mark workspace admins: its roles are the board's.
- **Your own row adds *(you)*** after your name, in `color.muted-foreground`, `spacing.inline-gap` after it, as the assignee list does (ADR-041 §11).
- **Someone not on the board** has no board role, so the second line is only the email (§9). **A presence row** has one line, the name (§16).
- The rejected options: the role at the row's end before the ⋯, which lines the roles up but cuts the email sooner on a phone; and the name alone, which can't tell two people with one name apart. Naming only *Owner* and *Admin* was rejected: nothing would say what a blank means.

### 4. The order: you first, then A–Z (D05)

- **You first, then everyone else A–Z by name**, in the team, in each of the board's two groups, as in the assignee list (ADR-041 §11). The server returns people in the order they joined; the browser sorts them.
- The presence list keeps the cluster's order instead (§16).
- The rejected options: the owner, then admins, then members, where making someone an admin moves their row while you look; and the order they joined, where nobody can be found by name without the filter.

### 5. A row's actions: a ⋯ at its end (D06)

- **A person's ⋯ opens the menu** ([ADR-040](adr-040-menus-and-the-confirmation.md) §2, §9): a 224 popover, 4 below the ⋯, right edges lined up; below 640 the bottom action sheet, its head the person's *Identity* (face, name, email) and *Close*. It is the column header's ⋯: `size.target-min` holding `DotsThree` at `size.icon` in `color.muted-foreground`, shown at rest at every pointer.
- **The menu closes before a confirmation opens**, so a phone shows the list's sheet and one more layer at most: a bottom action sheet, or a centred dialog. Never a second full-height sheet (ADR-021 §5).
- **Who sees a ⋯, and what is in it:**

| List | Seen by | Row | Entries |
|---|---|---|---|
| Team | a workspace admin | a member | *Make admin* · *Remove member* |
| | | another admin | *Make member* · *Remove member* |
| | | yourself, an admin who isn't the owner | *Make member* |
| | | the owner | no ⋯ (§8) |
| | a member | anyone | no ⋯ |
| Board | the board's owner or a workspace admin | another person on it | *Make owner* · *Remove from board* |
| | a workspace admin | the owner | *Remove from board* (§13) |
| | a workspace admin on the board | yourself | *Make owner* |
| | the owner | yourself | no ⋯ |
| | anyone else | anyone | no ⋯ |

- Nobody removes themselves from a list: you leave instead, from the workspace ⋯ (§7) or the board ⋯ (ADR-040). The server refuses removing yourself with 400.
- The rejected options: the role as a select and a × on every row, two controls per row and a × that removes someone from every board beside a name; and pressing a row to open its actions under it, which makes rows buttons and makes the list jump.

### 6. What asks first, and what is rose (D07)

- **Five more actions ask first**, so ten do in all ([ADR-040](adr-040-menus-and-the-confirmation.md) §4 had five):
  - *Remove member*, rose: it takes the person off every board in the workspace and unassigns their cards there
  - *Remove from board*, rose
  - *Leave workspace*, rose (§7)
  - *Make owner*, Primary: nothing is lost for good, but you can't take it back once you aren't the owner (§11)
  - *Make member* on your own row, Primary: you can't make yourself an admin again
- **Making someone an admin, or another admin a member, happens at once.** Any admin can undo it.
- **Rose marks three more menu entries** (ADR-040 §3): *Remove member*, *Remove from board* and *Leave workspace*, each with `UserMinus` or `DoorOpen` in rose.
- Every one is the `Confirmation` part with its own wording (§19). Focus opens on *Cancel* (ADR-040 §7).
- The rejected options: every change asks, a dialog for every promotion; and only the removes and leaving ask, so one stray press gives away a board or your own admin role.

### 7. *Leave workspace* is in the workspace ⋯ (D08)

- **The workspace ⋯ on 2.4 appears for every member.** The owner sees *Delete workspace*; everyone else sees *Leave workspace*, in rose with `DoorOpen`, then the confirmation. It mirrors *Leave board* in the board ⋯. This widens ADR-040's and ADR-033 §7's *the owner's only*.
- **After leaving: 2.3, with no message**, since you left by choice (ADR-030). The remembered workspace is cleared if it named this one.
- **The owner can't leave, and has no way to hand a workspace over.** No endpoint transfers a workspace, so the owner's only exit is deleting it. US-06c's *transfer or delete* names a transfer Phase 1 doesn't have (*Out of scope*).
- The rejected options: a rose button at the foot of the team list, where leaving hides inside a list about other people; and your own row's ⋯, a ⋯ on one row only.

### 8. The owner's row has no ⋯, and the server refuses removing the owner (D09)

- **The workspace owner's row has no ⋯ for anyone.** The owner can't be demoted, can't leave, and now can't be removed.
- **New backend item 4: `RemoveMember` refuses the owner with 403** (`ErrCannotRemoveOwner`), as demoting the owner already is. SRS UC-06 was edited first. Without it a direct request can remove the owner, after which nobody can delete the workspace.
- The rejected options: hiding it in the UI only, which leaves the request open; and allowing it, against SRS §2.6.

### 9. Adding people: one list, two groups, one person at a time (D10)

- **The board's list has two groups**: *On the board*, then *Not on the board yet*. Each group's label is the sidebar's *Your boards* look: `type.meta` in `color.muted-foreground`, `spacing.control-padding-x` in, `spacing.list-gap` above its rows. The groups are `spacing.section-gap` apart.
- **Each person not on the board has an *Add* button**, the Button's Outline look. It adds that one person at once (UC-11 with one id) and reads *Adding…*, in the Outline's Busy look, until the server answers. The row then moves up into the first group. No checkbox, no second screen, and a batch can never fail on one name.
- **Only the board's owner and workspace admins see the second group**, since only they can read the candidates (UC-11b). Everyone else sees *On the board* alone, with no label.
- **When everyone is already on the board**, the second group is one line, *Everyone in Product Team is already on this board.* That is inventory §6.5's flavour 6: one line, no drawing.
- **It takes *Add people* out of inventory §3's overlays.** It is this list's second group, not a panel of its own. The two batch behaviours inventory §3 warns about still differ, but adding one at a time means the all-or-nothing one can never bite.
- **On a board of eight or more, the second group starts below the eighth row**, so it opens out of sight. The filter finds anyone in either group. The file draws the list scrolled to its end, so the group shows (§20).
- The rejected options: an *Add people* row that swaps the list for a picker with checkboxes and one batch, which needs a checkbox look and fails whole on a 409; and its own entry in the board ⋯, a sixth board ⋯ shape with the same costs.

### 10. You, when you aren't on the board: left out (D11)

- **An admin who hasn't joined isn't in either group.** You appear once you're on the board; the banner's *Join* is how you get there (2.7).
- When everyone else is on the board, the line reads *Everyone else in Product Team is already on this board.*
- The rejected option: listing you with *Join the board to be on it.*, a row that only repeats the banner.

### 11. Handing a board over: *Make owner*, then a confirmation (D12)

- ***Make owner* is in a person's ⋯ on the board's list**, for the board's owner and for workspace admins. Its confirmation's button is Primary, *Make owner*, busy *Saving…*. The person must already be on the board (the server refuses anyone else with 400), and the list holds only people on it.
- **The old owner stays on as a member.** If that was you and you aren't an admin, your ⋯s disappear, since you no longer manage the board.
- **It takes *Transfer ownership* out of inventory §3's overlays.** It is this confirmation, not a list of its own.
- The rejected options: *Transfer ownership…* in the board ⋯, a second list of the same people; and doing it at once with *Undo* in a toast, an undo that is another transfer and fails once you no longer manage the board.

### 12. An owner learns how to leave from the list (D13)

- **To the owner, the board list's sentence adds** *You own this board. To leave it, make someone else the owner first.* Nothing else changes: *Leave board* stays absent from the owner's board ⋯ (ADR-040 §2).
- The rejected options: *Leave board* shown to the owner, opening an explanation, which is an entry that can't do what it says; and nothing at all.

### 13. An admin can take the owner off a board (D14)

- **For a workspace admin, the owner's row has *Remove from board*.** Its confirmation adds *Q3 Launch Plan will have no owner; workspace admins still manage it.*
- **A board with no owner**: no row says *Owner*, the sentence adds *Q3 Launch Plan has no owner.*, and *Make owner* is on every row for an admin. Appointing an owner is that same action (SRS §2.6, UC-12e).
- The rejected option: no way to take the owner off, so an owner who has stopped answering stays.

### 14. A leave that open lists can see: backend item 5 (D15)

- **New backend item 5: `LeaveBoard` sends `MEMBER_REMOVED`** after the silent eviction, as leaving a workspace already does. Open board lists and the assignee picker then drop the person, and nobody can be assigned who isn't on the board. SRS UC-10 was edited first, with §4.5 and §5.
- The rejected options: reloading the members on every `USER_LEFT`, a request from everyone watching each time anyone closes a tab; and leaving lists stale until they reload.

### 15. The presence list is menu-kind (D16)

- **From 640 up, a popover under the cluster**: group a's menu surface, `size.menu-width` 224, `spacing.menu-padding` 4, 4 below the cluster, right edges lined up. Pressing outside closes it, like a menu.
- **Below 640, the bottom sheet**, its head *8 here now* and *Close*.
- **It supersedes [ADR-021](adr-021-design-inventory-corrections.md) §5's placement** of presence overflow in *Forms and lists*. It has nothing to do: no actions, nothing to send.
- **Its head is a label, *8 here now***, the sidebar's *Your boards* look. Rows follow at most eight deep, then it scrolls (§2).
- The rejected option: the list dialog and the full-height sheet, a dimmed screen to read eight names.

### 16. What it shows: you first, then the cluster's order (D17)

- **You first, then the cluster's order**, so its first five rows are the five faces in the bar and the rest are who the seat counted. Each row is the `Person row` with nothing at its end and one line: the face and the name. Presence messages carry no email.
- **Nobody is marked as only viewing.** Who can edit is the members list's job.
- **It is live**: people come and go while it is open. While the connection is lost, its faces turn grey (the Avatar's Muted look) and the head reads *Reconnecting…*, as the cluster does (ADR-031 §5).
- The rejected options: marking *Viewing only* under people not on the board, which needs the list to match two sources; and A–Z, where the rows stop matching the faces.

### 17. Two contradictions in the design file, corrected (D18, D19)

- **You own Q3 Launch Plan** (D18). Its tile on 2.4 said so. *2.5 + board ⋯* and *2.15 + board ⋯* had been drawn for an admin on the board, with *Leave board*, which an owner never sees. Both now show the owner's entries: *Board settings…*, *Members…* and *Archive board* (with *Add column* first on the phone). The `Menu` part's Board look still shows all four, and the *Leave board* confirmation is unchanged.
- **Product Team has 11 people and 6 boards** (D19). Everyone on a board is in its workspace, so 2.4's *5 members* was wrong beside Q3 Launch Plan's 8, and 2.3's *4 boards* beside 2.4's six. The people: you (Owner); Arif Rahman and Maya Kusuma (Admin); Budi Santoso, Citra Dewi, Dimas Nugroho, Hendra Gunawan, Joko Tirtana, Lina Marpaung, Rina Wijaya and Sari Lestari. Citra, Hendra and Lina aren't on Q3 Launch Plan.
  - *11 members* on 2.4 and its state boards, on `Page 1` and `Overlays`. *2.4 + invite people* reads *9 members*, before Citra and Hendra are added.
  - 2.3's Product Team tile: *11 members · 6 boards*.
  - The delete-workspace wording: *for all 11 members*.
  - **The invite screens add Citra and Hendra**, in place of Dimas and Maya, who were already on Q3 Launch Plan: `citra@example.com` and `hendra@example.com`, wherever the addresses appear. Rina's mistyped `rina.w@example.com` stays; her own address is `rina.wijaya@example.com`.
- ADR-040 and ADR-042 are annotated, not edited.

### 18. The behaviour rules

- **Where each list opens**: the team from *N members* on 2.4, for every member; the board's list from *Members…* in the board ⋯, for everyone who sees that menu; the presence list from the cluster.
- **Focus**: from 640 up it opens in the filter when there is one, else on *Close*; on a phone, on *Close*, so the keyboard waits. It returns to what opened the list.
- **The filter** searches names and emails. Nothing found: *No member matches “xy”.*
- **A role change** shows *Saving…* in place of the role on the second line until the server answers. A failure puts the role back and says so on the list's last line, as the `Form message` (ADR-032 §1).
- **Removing, leaving and handing over wait in their dialogs**, the button Busy, as *Leave board* does (ADR-040 §13). After a removal the row leaves the list, and *N members* on 2.4 counts one fewer. After *Leave workspace*: 2.3.
- **An *Add* that fails** returns the button and says so on the list's last line. A 409, someone added meanwhile, counts as done: the row moves up. A 400, the person has left the workspace, takes the row away with *Citra Dewi is no longer in Product Team.*
- **Live, and not.** The board's list follows `MEMBER_ADDED`, `MEMBER_REMOVED` (with §14, a leave too) and `OWNERSHIP_TRANSFERRED` while open. The team list loads each time it opens; if your role changed meanwhile, your next action fails with 403 and its wording (§19). The presence list follows presence.
- **Someone removed from a board or a workspace who has it open** is taken out with the access-revoked toast (ADR-030), unchanged.

### 19. The wordings (D20)

| Where | Wording |
|---|---|
| The team | Title *Members* · *11 people in Product Team.* · filter *Find a member*; nothing found *No member matches “xy”.* · second lines *Owner · ayu@example.com*, *Admin · …*, *Member · …* · *(you)* after your name · *Close* |
| A person's ⋯ | Spoken *Actions for Dimas Nugroho* · *Make admin* (`ShieldCheck`) · *Make member* (`User`) · *Remove member* (`UserMinus`, rose) · on a board: *Make owner* (`Crown`) · *Remove from board* (`UserMinus`, rose) |
| Workspace ⋯ | *Leave workspace* (`DoorOpen`, rose), for everyone but the owner |
| The board's list | Title *Members* · *8 people on Q3 Launch Plan.* · to the owner, after it, *You own this board. To leave it, make someone else the owner first.* · with no owner, after it, *Q3 Launch Plan has no owner.* · groups *On the board* · *Not on the board yet* · *Add*, busy *Adding…*, spoken *Add Citra Dewi to Q3 Launch Plan* · heard after *Citra Dewi added to Q3 Launch Plan.* · empty *Everyone in Product Team is already on this board.* (*Everyone else…* when you aren't on it) |
| Remove member (rose) | *Remove Dimas Nugroho from “Product Team”?* · *They’re taken off its boards and unassigned from their cards there. To come back, they’ll need to be invited again.* · *Remove* · *Removing…* |
| Remove from board (rose) | *Remove Dimas Nugroho from “Q3 Launch Plan”?* · *They’re unassigned from their 2 cards on it.* With none: *They can join it again from the workspace.* On a private board: *It’s private, so someone will need to add them back.* in place of joining again · for the owner, after it: *Q3 Launch Plan will have no owner; workspace admins still manage it.* · *Remove* · *Removing…* |
| Make owner (Primary) | *Make Dimas Nugroho the owner of “Q3 Launch Plan”?* · if you own it: *You’ll stay on the board as a member.* · if someone else does: *Ayu Pratama stays on the board as a member.* · if nobody does: *They can change its settings, add and remove people, and archive it.* · *Make owner* · *Saving…* |
| Demote yourself (Primary) | *Stop being an admin of “Product Team”?* · *You won’t be able to invite people, change roles or remove anyone. Another admin can make you one again.* · *Make me a member* · *Saving…* |
| Leave workspace (rose) | *Leave “Product Team”?* · *You’ll be taken off its boards and unassigned from your cards on them. Someone will need to invite you back.* · *Leave workspace* · *Leaving…* |
| Heard after | *Dimas Nugroho is now an admin.* · *Dimas Nugroho is now a member.* · *Dimas Nugroho removed from Product Team.* · *Dimas Nugroho is now the owner of Q3 Launch Plan.* |
| Who's here | The cluster, spoken *8 people here. Show who.* · head *8 here now* (*Just you here* when alone) · *Ayu Pratama (you)* · while reconnecting, the head reads *Reconnecting…* and the faces go grey |
| Failures | A role change: *Couldn’t change Dimas Nugroho’s role. Try again.* · an *Add*: *Couldn’t add Citra Dewi. Try again.* · someone who left: *Citra Dewi is no longer in Product Team.* · ADR-032's *Can’t reach CollaboTask. Check your connection and try again.* · a 403 after your role changed: *Only workspace admins can change roles.* · *Only the board’s owner or a workspace admin can manage its members.* |

Counts in a sentence come from what the screen already holds (ADR-040 §13): the board's cards for *2 cards*, the list's length for *11 people*.

**The menu entry is *Remove member*, not *Remove from workspace* as the decision page had it.** Found while drawing: a 224 menu row leaves 168 for its words (224 − 2 × 4 menu padding − 2 × 12 inset − 16 glyph − 8 gap), and *Remove from workspace* is 175 at `type.body`, so it wrapped onto a second line. In a team list "member" already means "of this workspace", and the confirmation still names the workspace (*Remove Dimas Nugroho from “Product Team”?*). Every other wording is unchanged.

### 20. What the file draws

- **No token: 222.**
- **Two parts, 32 → 34.**
  - **`Person row`**, board `Person row — states` (y 13964) on `Components`. `Trail` = Menu · Add · Adding · None. A row `size.control-height` tall and `size.form-width` wide, `spacing.control-padding-x` on its left, `spacing.icon-gap`, `radius.control`, `color.popover`. Inside: an Avatar; `Who`, a `Label` (the name in `type.body`, a hidden *(you)* in `color.muted-foreground`, `spacing.inline-gap` apart) over `Sub` (*Member · dimas@example.com*, `type.meta`, `color.muted-foreground`); then `Menu` (`size.target-min`, a `Glyph` `DotsThree` Muted) or an Outline *Add* / Busy *Adding…*. None adds `spacing.control-padding-x` on the right.
  - **`List`**, board `List — states` (y 14228). `Of` × `Frame`, six looks:
    - *Team/Dialog*: the `Form` part's 448 frame, holding `Text`, a `Body` (the filter, then the rows) and `Actions` with *Close*. Eleven rows, you first and then A–Z, the last three hidden, so it stands eight rows tall: 380 at Coarse, 316 at Fine.
    - *Members/Dialog*: drawn scrolled to its end (§9). *On the board*'s label and first four rows are hidden, and *Not on the board yet* shows Citra Dewi, Hendra Gunawan and Lina Marpaung with *Add*.
    - *Team/Sheet* and *Members/Sheet*: 390, `spacing.screen-x` inside, a `Head` with *Close*, no `Actions`, every row shown.
    - *Here/Popover*: `size.menu-width`, *8 here now* in `type.meta` (padded `spacing.control-padding-x` / `spacing.list-gap`), then eight `Person row` copies with `Trail` None and `Sub` hidden, in the cluster's order. *Here/Sheet*: a `Panel` with `shadow.rule-top`, a `Head` with *8 here now* and *Close*.
- **`Glyph`, 22 → 28 looks**: `Crown`, `ShieldCheck`, `User`, `UserMinus` Muted and Destructive, and `DotsThree` Muted. The board widened to 1276.
- **`Menu`, 10 → 16 looks**: *Teammate*, *Board member* and *Workspace member*, each Popover and Sheet. The sheets' heads show Dimas Nugroho's face and name. The board widened to 4192.
- **`Confirmation`, six more wordings** in a second row 378 below the first: *Remove member*, *Remove from board* and *Remove owner* (rose), *Make owner* and *Stop being an admin* (Primary), and *Leave workspace* (rose). The board is now 648 tall.
- **Ten boards on `Overlays`, 27 → 37**:
  - y 9900: *2.4 + members* (the team, Dimas Nugroho's ⋯ open), *2.4 + removing a member*, *2.5 + members* and *2.5 + make owner*
  - y 11000: *2.5 + members, a person's ⋯* and *2.5 + who's here*
  - y 12100: *Members at 390*, *Members at 390 + a person's ⋯*, *Board members at 390* and *2.15 + who's here*
- **The hues** on every new face: AP periwinkle, AR sage, BS olive, CD periwinkle, DN teal, HG ochre, JT ochre, LM mauve, MK steel, RW terracotta, SL mauve.
- **Corrected across the file (§16–§17):**
  - **You own Q3 Launch Plan.** *Leave board* is hidden in the board ⋯ on *2.5 + board ⋯* (the menu is now 148 tall) and *2.15 + board ⋯* (244, moved to y 460).
  - **Product Team is 11 people and 6 boards.** That is the 2.4 heading on `Page 1` and `Overlays`, 2.3 and *2.3 + create workspace*, and the delete-workspace confirmations' *for all 11 members*. The two invite screens read *9 members*, since Citra and Hendra are the two being invited.
  - **The invite addresses** are citra@ and hendra@ in place of dimas@ and maya.k@, on three `Overlays` boards and in the `Input`, `Address chip`, `Invite result` and `Form` parts. Dimas and Maya are already members.

### 21. The code

- **No CSS variable is added.** A list's scroll region is `max-h-[calc(8*var(--control-height)+7*var(--list-gap))] overflow-y-auto`.
- **The frame is ADR-042's `FormFrame`**, with *Close* in place of the form's buttons: shadcn's `Dialog` from 640 up, its `Sheet` below.
- **The presence list is shadcn's `Popover`** from 640 up (`align="end"`, `sideOffset={4}`, `w-[var(--menu-width)]`) and a bottom `Sheet` below 640, the menu's two presentations.
- **A person's ⋯ is shadcn's `DropdownMenu`**, the bottom `Sheet` below 640, as every menu (ADR-040 §15).
- **One `PersonRow` component**: an Avatar, the name and *(you)* over a second line, and a slot at the end for a ⋯ or an *Add* button. Its second line is `` `${role} · ${email}` ``, truncated.
- **One sort**: you first, then `localeCompare` on the name.
- **Data.** The team is UC-05b's member list, fetched each time the list opens (no WebSocket). The board's list is UC-12's roster, plus UC-11b's candidates for the owner and admins; the second group is the candidates with `is_board_member` false, less you. `MEMBER_ADDED`, `MEMBER_REMOVED` and `OWNERSHIP_TRANSFERRED` patch the board's roster in the query cache.
- **Adding** posts `{ user_ids: [id] }`, one person per press; a 409 is treated as done.
- **Phosphor imports**: `CrownIcon`, `ShieldCheckIcon`, `UserIcon`, `UserMinusIcon`, `DotsThreeIcon`.

## Options considered

Every option is named, with its cost, under its decision above, and all were drawn side by side on the decision page. In short:
- **The frame:** a drawer from the right. **Its height:** growing to the screen's edges; a token.
- **The row:** the role at the end (recommended, not chosen); the name alone. **Roles:** only Owner and Admin named. **Order:** by role; by joining.
- **Actions:** a role select and a ×; actions opening under the row.
- **Asking:** every change; only the removes and leaving.
- **Leaving a workspace:** at the list's foot; on your own row.
- **The owner's row:** hidden in the UI only; allowed.
- **Adding:** a picker with checkboxes in the list; its own board ⋯ entry. **You, off the board:** listed with a pointer to Join.
- **Handing over:** a board ⋯ entry and a second list; at once with *Undo*. **Telling the owner:** *Leave board* that explains; nothing. **Removing the owner:** not allowed.
- **A leave:** reloading on every `USER_LEFT`; nothing.
- **Presence:** the list frame. **Its content:** *Viewing only* marks; A–Z.
- **The file:** Maya Kusuma as Q3 Launch Plan's owner; leaving the counts as drawn.

## Out of scope (deferred, recorded)

- **Handing a workspace over.** No endpoint transfers a workspace, so its owner's only exit is deleting it. US-06c's guard says *transfer or delete*; the transfer is not a Phase 1 use case. Raise it with the PRD before anyone builds a *Transfer workspace* entry.
- **Hover, pressed and focus on the overlay boards.** The parts carry those looks.
- **2.4 as a member sees it, with *Leave workspace* open.** The `Menu` part's *Workspace member* look and the confirmation are drawn; a whole 2.4 for Budi would need every tile's role redrawn.
- **A board with no owner, and the presence list while reconnecting.** Their wordings are in §19; neither is drawn.
- **Movement.** Lists open and close in code only (inventory §4.3).
- **The overlays at a tablet's Coarse density.** The desktop boards are drawn at Fine (ADR-040 *Operational note*).

## Downstream edits this ADR requires

1. **[SRS](../../spesifications/001-software-specifications.md):** UC-06's owner guard, UC-10's `MEMBER_REMOVED`, §4.5's eviction text and §5's event table. Done first.
2. **[ADR-021](adr-021-design-inventory-corrections.md): annotation.** §5's table places presence overflow in *Forms and lists*; it is menu-kind.
3. **[ADR-040](adr-040-menus-and-the-confirmation.md): annotation.** §2's workspace ⋯ shapes, §3's rose entries and §4's actions that ask first; *2.5 + board ⋯* is now drawn for the owner; the delete-workspace wording reads *11 members*.
4. **[ADR-042](adr-042-the-forms.md): annotation.** §1's frame holds lists; the invite screens add Citra and Hendra.
5. **[ADR-041](adr-041-the-card-drawers-pickers.md): annotation.** §11's filter applies to every list of people.
6. **[ADR-031](adr-031-presence-cluster.md): annotation.** *Out of scope*'s overflow list is settled here.
7. **[ADR-033](adr-033-workspace-screens.md): annotation.** §7's workspace ⋯ appears for every member; 2.3 and 2.4's Product Team counts read 11 and 6.
8. **[ADR-018](adr-018-state-vocabulary.md): annotation.** Rose spent on three more entries; state 18 stays closed.
9. **[`design-file-inventory.md`](../design-file-inventory.md):** the sources table, §3, §5, §7.2, §7.3, §7.6 and *Related documents*.
10. **[`frontend-architecture.md`](../frontend-architecture.md):** the list frame, the presence popover, the sort, and the roster patches.
11. **`CLAUDE.md` › `## Now`:** a pointer, and the open backend items from three to five.
12. **`docs/build-log.md`:** no entry. Stage 5 is not closed.
13. **PRD, user stories, `CONTEXT.md`:** no change. US-06c's *transfer* is recorded under *Out of scope*.

## Operational note

**Drawn 2026-11-24.** The desktop overlays were placed and looked at in Light + Fine, and the phone ones in Light + Coarse. The file was left in Light + Coarse.

**Checks.**
- **Binding check, parts:** nothing on `Person row — states`, `List — states`, `Glyph — states`, `Menu — states` or `Confirmation — states`.
- **Binding check, Overlays:** nothing on *2.4 + members*, *2.4 + removing a member* or the four phone boards, apart from the card `⋯` rows §7.5 expects: 21 on each 2.5 board and 15 on each 2.15 board.
- **Bindings by name:** each `List` look was also checked for its expected tokens by name. §7.5's check can't see a binding that was lost while its value was 0 (below).
- **Foundations:** not re-run. No token changed, so its catalogue tables still stand.
- **Layout:** a check that compares every auto-sized board with its children, gaps, padding and margins found nothing on `Page 1` or `Overlays` after the repair below.
- **The looks:** Light + Fine on *2.4 + members*, *2.4 + removing a member*, *2.5 + members*, *2.5 + make owner* and *2.5 + who's here*; Light + Coarse on *Members at 390*, its person's ⋯, *Board members at 390* and *2.5 + make owner*; Dark + Fine on *2.4 + members*. In the presence cluster each face hides the right edge of the one before it, as ADR-031 §2 measured it, so the covered initials read *A*, *M*, *J*: that is the settled overlap, not a fault.

**What drawing found:**
- ***Remove from workspace* didn't fit.** It is 175 wide at `type.body`, and a menu row leaves 168, so it wrapped. The entry became *Remove member* (§19).
- **A theme switch stamps each part's text size from its main onto every copy that has its own words, and the copies get their own size back only when the editor draws their page.** The switch to Dark was made just as the tab went to sleep. Afterwards every *Add column* was stored 47 wide, the width of the Button main's *Sign in*, and the boxes around such texts were laid out to it. Exports draw the real text, so *2.5 + make owner* came out with its title running into its sentence, its buttons clipped and *Add column* cut off. Opening `Page 1` and then `Overlays` in the editor, with the tab awake, put every width back: 86 for *Add column* again, and the layout check came back clean. So: make each switch with the tab in front, open every page with a screen to export, and export only after that. Opening `Overlays` (37 screens) stalls the tab for over a minute.
- **A flex row laid out for an old size is re-laid by hiding one child and showing it again, in two calls.** That fixed seven *(you)* labels and four *Close* buttons, all on `Overlays`. One of the *(you)* labels is group b's *2.6 + assignee*, where it had sat on top of the name. Setting a text to fixed width and back did nothing here, because the texts' own widths were already right.
- **Binding a padding in the same call that creates its shape can be lost.** The section labels' padding came out 0 in an export, and §7.5 couldn't see it, because nothing was bound. Padding was set, then bound again in a separate call.
- **`resize()` turns an auto height fixed.** The sheets needed `flex.verticalSizing = 'auto'` set again.
- **A cloned menu row lands first, and `setParentIndex(0)` moves it last**, since the index counts from the other end of `children`. The clone's label also went back to the main's and was set again.
- **`switchVariant` and edits to the result in one call fail**; they were split into two calls. **`addVariant` times out but still lands**; two were lost while the connection was down, so the count was checked before redoing them.
