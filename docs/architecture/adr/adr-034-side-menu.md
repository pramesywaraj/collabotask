# ADR-034: The Side Menu — a Workspace Sidebar Beside 2.4 and the Board

- **Status:** Accepted — **§10.10 completed after acceptance** (2026-11-21): from 1024 up, the board glides with the sidebar when you hide or show it, which narrows [ADR-018](adr-018-state-vocabulary.md) §8's rule 3 for that one case (the note under §10.10). **§9 gains one rule after acceptance** (2026-11-21): the shortcut does nothing while a card or a column is being dragged (the note under §9). **Its one open question, and question 4 for its parts, settled by [ADR-035](adr-035-sidebar-loading-failure-and-switcher-mark.md)** (2026-11-21):
  - on a cold load the switcher and the board rows are row-sized blocks
  - on a load failure, *Your boards didn’t load.* in rose, then a *Try again* refresh row
  - the open switcher's workspace rows keep the empty slot, and a still plum dot (`size.dot` 8) marks the selected one
  - *Your boards* sits `spacing.list-gap` above its rows
  - the user row's avatar takes the slot's place

  The parts and 2.4's six boards are drawn. **§5's user menu drawn by [ADR-040](adr-040-menus-and-the-confirmation.md)** (2026-11-23): your face, name and email, then *Sign out*, 224 wide, opening upward from the user row with its left edge on the row's. Below 640 it is a bottom sheet whose head is you. §10.8's `color.scrim` dims behind every dialog and sheet. **§4's slide-over drawn, and its *never stacks on another layer* clarified, by [ADR-044](adr-044-the-sidebar-sliding-over-and-the-toast.md)** (2026-11-24): it is drawn on the phone, closed and with the switcher open, and focus opens on the current row. Nothing opens the panel from inside another layer, which stays true, but one layer opens over it: below 640 the user menu is the bottom sheet over a second dimming. *Out of scope*'s *Drawing the slide-over panel* is done. **§10.1's *CollaboTask* superseded by [ADR-049](adr-049-the-logo.md)** (2026-11-28): the top row holds the button and the logo's lockup, the mark at 28 and the name at 17.5, at both densities. Everything else stands as written.
- **Date:** 2026-11-21
- **Scope:** A side menu (the **sidebar**) in the app zone. It settles:
  - what the sidebar holds, and which screens show it
  - how it hides, the width at which it stops sitting beside the content, and what it becomes below that width
  - what moves out of the app bar, and what the app bar names
  - whether 2.3 Workspaces keeps its own screen, and where the app opens
  - the workspace switcher, the board list, and the way back to 2.4
  - keyboard and screen-reader behaviour, and a shortcut
  - its look: layout, surface, the current row's mark, each board's colour mark, width, the button's icon, and how the content dims
  - **four tokens, five entries** (194 → 199)

  It **extends [ADR-017](adr-017-ui-foundations.md) §9** (the code gains the variables in §13) and **[ADR-021](adr-021-design-inventory-corrections.md) §5** (a fifth presentation kind, *navigation*). It also fills a gap in **[ADR-019](adr-019-responsive-touch-scope.md) §8**: that section's 44px "menu" slot was never defined, and it is the sidebar's button. It **supersedes three things, and nothing else in them**:
  - [ADR-033](adr-033-workspace-screens.md) §2's app bar contents, **on 2.4 only**. *CollaboTask* and the user menu move into the sidebar. The bar holds the show/hide button and the title *All boards*. 2.3 keeps §2 as written.
  - [ADR-033](adr-033-workspace-screens.md) §4's tile counts, **on 2.4 only, while the sidebar is open**: 4 tiles across at 1440 and 2 at 1024, not 5 and 3.
  - [ADR-019](adr-019-responsive-touch-scope.md) §2's **1024 and 1440 rows, while the sidebar is open** (§11). Every row below 1024 stands, because below 1024 the sidebar slides over the content and takes no width from it.
- **Upstream edits:** none. The PRD (§3.2, §4.1, §5.1), US-05, US-05b, US-08, US-09, US-10, US-12 and US-12c, and SRS UC-05, UC-05b, UC-08, UC-09 and §5 were checked against every choice here. Everything in this ADR happens in the browser: no endpoint, schema or permission rule moves. A new account still lands on the empty 2.3 (PRD §4.1).

## Context

A side menu had never been discussed. Nothing in the PRD, the user stories, the SRS, the ADRs or the design inventory mentioned one. The design assumed, without deciding it, **a top app bar only, with drill-down navigation**: workspace list → workspace → board, with *CollaboTask* linking back to `/workspaces`. The route tree had no shared layout beyond the auth check. 2.3 and 2.4 were drawn (ADR-033), and 2.5 Board — at rest waited on this decision.

**Three gaps turned up while reading the docs**, and a sidebar closes all three:
- **The board screen had no way back.** Inventory §1.5's app bar holds the board name, presence, `⋯` and add-column. There was no link up to the workspace.
- **The board screen had no user menu.** Only 2.3 and 2.4 drew one.
- **ADR-019 §8's phone bar reserves 44px for a "menu"** (`375 − 32 − 44 menu − 44 actions`) and never says what it is.

**What the API can fill a sidebar with** was checked against the SRS and the code:
- `GET /workspace` lists your workspaces (UC-05). 2.3 already loads it.
- `GET /workspace/:id/board` lists one workspace's boards (UC-08), newest first (`ORDER BY b.created_at DESC`). Each board carries `access_status` (`JOINED` or `CAN_JOIN`), `visibility` and `background_color`. 2.4 already loads it.
- **No endpoint lists your boards across workspaces.** "Every board in every workspace" would need one request per workspace, or a new endpoint.
- **Realtime messages reach only the people inside a board** (SRS §5.2). A board a teammate creates, or a board you are added to, appears in any list only when that list reloads. 2.4's grid already works this way.
- **The app has no other top-level pages**: no *My cards*, settings or notifications. A sidebar can hold only workspaces and boards.

**The width arithmetic was re-checked first.** It assumed a 240px sidebar, kept the 16px padding on each side of the content, and used the right column gap for each pointer (8 Fine, 12 Coarse, [ADR-023](adr-023-spacing-between-objects.md) §2.2). One correction came out of it: the first figures used 8 on tablets, which are Coarse. §11 has the table.

**The grilling ran in two parts.** Eleven structural questions were asked in text, one at a time, each with a recommendation. The user took the recommendation on all eleven. Two answers came after a clarifying question: whether the button stays when the sidebar is hidden, and whether *remembered on this device* and the joined-only filter need the backend (they don't). Eight visual choices then went on a rendered page, each option drawn in Light and Dark at 1:1 with the file's values, with a picker. **The user took all eight recommendations** and left no notes.

That page was also question 4 for the sidebar: it named the four values the sidebar needs and has no token for (§13). Question 4 for 2.5 is still to be asked before 2.5 is drawn.

## Decision

### 1. What it holds

**One workspace at a time:**
1. *CollaboTask*, at the top.
2. **The workspace switcher**, showing the current workspace's name (§7).
3. **An *All boards* row**, which goes to 2.4 (§8).
4. **The boards you've joined in this workspace**, under a *Your boards* label (§8).
5. **The user menu**, at the bottom (§5).

Nothing else. *Create board* and *Create workspace* stay on 2.4 and 2.3, and the sidebar has no "+".

This uses only the two lists 2.3 and 2.4 already load, so the backend does not change. It serves both people in PRD §3.2: the Doer, who "lives inside a board or two", and the Organizer, who may run more than one team. With teams of 2–15, most people have one workspace, so the switcher usually holds one name.

### 2. Which screens show it

**2.4 Workspace detail and the board, with all of the board's states (2.6–2.15). Not 2.3, and not Login or Register.**

The sidebar's content is *this workspace's boards*, and only 2.4 and the board have a current workspace. On 2.3 the sidebar would show a switcher with nothing chosen, next to a page listing the same workspaces. Showing it on both 2.4 and the board keeps the layout still as you move between them.

The board's states each meet it in their own way:
- **2.7, read-only** (a board you haven't joined): the board isn't in *Your boards*, so no row is marked current. It appears once you join. The read-only banner spans the content, to the right of the sidebar.
- **2.8, the join gate**: the same, no row marked.
- **2.12, gone**: the board leaves the list.
- **2.6, the card drawer**: opens over the board as before. The sidebar stays where it is.

### 3. Hiding it

- **One button hides and shows it.** It sits at the left of the app bar's line, 16px from the screen's edge, whether the sidebar is open or hidden (§10.1), so it is always in the same spot.
- **It starts open** from 1024 up. Your choice is **remembered on this device**, for both screens.
- **Ctrl + \\ (⌘ + \\ on a Mac)** does the same, at every width (§9).
- **Hidden means fully hidden.** None of its items can receive focus.

A narrow strip of icons was rejected: a board has only a name and one of 8 colours, so a strip of colour marks cannot tell boards apart.

### 4. Beside the content from 1024 up; below it, a panel that slides over

| Width | The sidebar |
|---|---|
| **1024 and up** | Sits beside the content, open by default, and pushes the content right |
| **Below 1024** (tablets in portrait, narrow windows, phones) | Hidden by default. The button slides it in from the left, **over** the content. The content dims (`color.scrim`, §10.8). It closes when you pick a board or workspace, tap outside it, press Escape, or press the button. **It always starts closed**: the remembered choice applies only from 1024 up |

**Why 1024.** It is a number the docs already use: ADR-017's original floor, and Tailwind's `lg`. Every width that has the sidebar beside the content keeps at least two full columns. Below it, a sidebar beside the content would leave one column (744 → 1 + 180, 834 → 1 + 270), which is the case worth avoiding.

**The panel is the same column** as the sidebar beside the content: the same width (240), the same rows, the same top row. On a 375 phone, 135px of dimmed board shows beside it.

**ADR-021 §5 gains a fifth kind, with one rule:**

| Kind | Members | Below 1024 | 1024 and up |
|---|---|---|---|
| **Navigation** | The sidebar | **Slides over the content from the left**, with the content dimmed | Beside the content |

It can never stack on another layer, which was ADR-021 §5's worry. It opens only from the app bar's button or the shortcut, never from inside a sheet, and while it is open nothing behind it can be reached. Its switcher opens **inside** it (§7), which is the rule ADR-021 §5 already gives the card drawer's pickers.

### 5. What moves out of the app bar, and what the bar names

**On 2.4 and the board, *CollaboTask* moves to the sidebar's top row and the user menu to its bottom.** The app bar then holds the show/hide button, a title and the page's own controls.

**The bar names the sidebar row you're on:**

| Screen | The app bar |
|---|---|
| **The board**, sidebar open (1024 and up) | Board name (`type.title`), presence, board `⋯`, add column. The button is in the sidebar's top row, at the same spot |
| **The board**, sidebar hidden or below 1024 | The button, then the same |
| **2.4**, sidebar open | *All boards* (`type.title`). 2.4's heading below it is unchanged: the workspace name in `type.heading`, the role, the description, *N members*, *Create board* and the owner's `⋯` |
| **2.4**, sidebar hidden or below 1024 | The button, then *All boards* |
| **2.3** | Unchanged ([ADR-033](adr-033-workspace-screens.md) §2): *CollaboTask* on the left, linking to `/workspaces`, and the user menu on the right. No button, because 2.3 has no sidebar |

On 2.4 two names stack: *All boards* in the bar, and the workspace's name in the heading. The user accepted that cost (§10.6).

**The phone's board bar keeps its title width.** ADR-019 §8's 44px "menu" slot is the button. The bar is the button, the title, presence and `⋯`, and the title keeps the ~180px ADR-031 left it. Putting the user menu in the bar as well (44 + a gap) would have cut the title to about 128px, roughly 13 characters, and ADR-019 §8 had already rejected 15 as too short for real board names. The user menu is rarely used: the profile is view-only and logging out is occasional.

While the sidebar is hidden, or below 1024, the user menu is two presses away.

### 6. 2.3 stays, and the app opens in your last workspace

**2.3 keeps its own screen** for the three jobs only it does:
- it is the empty screen a new account sees first (PRD §4.1)
- it holds *Create workspace*
- it shows each workspace's role and counts

**The app opens in the last workspace you were in**, remembered in the browser:
- **When:** only at the app's entry point (`/`) and right after sign-in. `/workspaces` always shows 2.3, so *All workspaces* (§7) always reaches it.
- **What's kept:** the workspace id, in `localStorage`, under your user id, so someone else signing in on the same browser is not sent to your workspace. It is written whenever a workspace-scoped screen opens.
- **When it points nowhere:** if the workspace can't be opened (it was deleted, or you were removed), the value is cleared and you land on 2.3. After `removed_from_workspace` ([ADR-030](adr-030-toast.md) §6) the value is cleared too, if it names that workspace.
- **Its limits:** per device and per browser. A new phone, another browser, a private window or cleared site data lands on 2.3 the first time.
- **Frontend only.** Following you across devices would need a server-side column and endpoint, a scope change, and is deferred.

### 7. The workspace switcher

- **A row** at the top of the sidebar's list: the workspace's name in `type.card-title`, with Phosphor `CaretUpDown` in `color.muted-foreground` at the right.
- **Pressing it opens your workspaces inside the sidebar**, pushing the boards down. It is a show-more list (a disclosure), not a menu. On a phone the sidebar is already a layer over the content, so a separate menu or bottom sheet would stack a second layer on it.
- **The list:** your workspaces as rows, newest first as `GET /workspace` returns them, with the current one marked like a current board (§10.3). Workspaces have no colour, so their rows have no swatch. The list **ends with *All workspaces***, which goes to 2.3.
- **Picking a workspace goes to its 2.4**, not to the last board you had open there. That would need a second remembered value, and it could point to a board that has since been archived.
- *Create workspace* stays on 2.3 only.

### 8. The board list

- **An *All boards* row** comes first and goes to 2.4. It is marked current while you're on 2.4, so on every screen with the sidebar something in it is marked current. It names what 2.4 has that the sidebar doesn't: every board, not only yours, plus the team, the Archived list, *Create board* and *Delete workspace*.
- **A *Your boards* label**, in `type.meta` and `color.muted-foreground`, with `spacing.section-gap` above it.
- **Only the boards you've joined**, newest first. The frontend keeps the boards whose `access_status` is `JOINED`. This is tidiness, not permission: the boards it hides are ones 2.4 shows you anyway. An admin can see every private board in the workspace, and under the full list their sidebar would be a long run of locked boards they had never opened.
- **Each row:** the board's colour as a 12px square (§10.4), its title, and the filled Phosphor `Lock` after the title if the board is private, as on the tile.
- **One request serves two places.** The sidebar reads the same cached `GET /workspace/:id/board` (active boards) that 2.4's Active grid reads. 2.4's Archived filter is a separate request and doesn't touch the sidebar.
- **No joined boards** is a seventh empty flavour, alongside ADR-033's six: *You haven’t joined a board yet.*, in `type.body` and `color.muted-foreground`, with no illustration, like flavours 5 and 6. *All boards* above it is the way forward.

**When the list refreshes.** The cached list is refetched or patched:
- after you join, leave, create, archive or unarchive a board ([ADR-033](adr-033-workspace-screens.md) §11 already needs the refresh after joining, so the tile can show your new role)
- when the open board is renamed or recoloured (`BOARD_UPDATED`), archived (`BOARD_ARCHIVED`), or your access to it is revoked (`ACCESS_REVOKED`)
- otherwise when the list reloads, as on 2.4. A board you are added to while elsewhere appears then, not at once (*Out of scope*).

### 9. Keyboard and screen readers

These follow from WCAG 2.2 AA, PRD §5.4's target:

- **The sidebar is a navigation landmark** named *Sidebar*. The current row (a board, or *All boards* on 2.4) carries `aria-current="page"`.
- **The button** is labelled *Hide sidebar* or *Show sidebar*, carries `aria-expanded` and `aria-controls`, and shows its shortcut in its tooltip (*Hide sidebar · Ctrl + \\*, or *⌘ + \\* on a Mac) and in `aria-keyshortcuts`.
- **From 1024 up**, a ***Skip to content*** link is the first thing on 2.4 and the board that can receive focus, visible only while focused. Without it, keyboard users would tab through every board before reaching the board itself (SC 2.4.1). The sidebar comes before the content in Tab order.
- **Below 1024** the panel is a modal dialog: focus moves into it, Tab stays inside it, Escape closes it, focus returns to the button, and the dimmed page can't be reached.
- **The switcher** is a button with `aria-expanded`, controlling a list of links.
- **The shortcut is Ctrl + \\ (⌘ + \\ on a Mac).** It needs a modifier key, so it can't fire while you type a card title, and mainstream browsers don't use it. A single key such as `[` would break SC 2.1.4: a single-key shortcut must be possible to turn off, and the app has no settings screen to turn it off in. `M` on a card ([ADR-019](adr-019-responsive-touch-scope.md) §6) is allowed only because it works only while the card has focus.

> **Added after acceptance (2026-11-21): the shortcut does nothing during a drag.** While a card or a column is being dragged, Ctrl + \\ is ignored. Otherwise, from 1024 up, the board would glide 240px under the dragged card and carry every drop target with it. Below 1024, the slide-over would cover the board you are dropping onto. The button can't be pressed mid-drag anyway, because the pointer or finger is holding the card, so this closes the one way in. When the drag ends or is cancelled, the shortcut works again. Phase 1 has no keyboard drag (ADR-019 §6 sends `M` to the move picker), so pointer and touch drags are the only case. In code, the shortcut's handler returns early while `uiStore` holds a drag ([frontend-architecture.md](../frontend-architecture.md)).

### 10. The look

Each of these was chosen on the rendered page. Only that choice changed between options, and every option was drawn in Light and Dark with the file's values.

**10.1 Full height.** The sidebar runs from the top of the screen to the bottom. Its top row is `size.app-bar` (56) tall, with `spacing.screen-x` (16) at the sides, and holds the button and *CollaboTask* (`type.title`). Its bottom rule continues the app bar's. The app bar starts at the sidebar's edge. When the sidebar is hidden, the app bar runs the full width and the button sits at its left, at the same spot on screen. Below 1024 the slide-over panel is this same column. The cost: when the sidebar hides, the page title slides 240px left.

> *2026-11-28:* the top row holds the button and **the logo's lockup** in place of *CollaboTask* in `type.title`: the mark at 28, a gap of 8 and the name at 17.5 in Poppins SemiBold, deep plum, at both densities, so it fits beside the 44 button at Coarse. See [ADR-049](adr-049-the-logo.md) §7.

**10.2 Page white, with a rule.** The fill is `color.background`, with a `border.width` rule in `color.border` on its right. This is how the app bar is drawn.

| | Light | Dark |
|---|---|---|
| Sidebar against the board canvas | 1.26 | **1.00**: the rule is the only edge |
| The rule against the sidebar | 1.26, decorative, like the bar's | 1.92 |

In Dark the sidebar and the canvas are the same colour, as the banner and the canvas are ([ADR-028](adr-028-banner.md)), and the rule separates them. `color.muted` was rejected: in Light it is the column panels' own colour, so the sidebar read as one more column.

**10.3 The current row: plum text and a 1px plum edge.** This is the Active / Archived switch's look ([ADR-033](adr-033-workspace-screens.md) §5), so ADR-018's *where you are* looks the same everywhere. The edge adds a shape as well as a colour. That matters because plum against normal text is only 2.67 in Light (2.74 in Dark), under the 3:1 at which colour alone could be relied on.

| | Light | Dark |
|---|---|---|
| Plum text and edge on the sidebar | 6.54 | 6.90 |

The other rows have `color.foreground` text and no edge at rest. **Hover and pressed step the edge** (`color.border-hover`, `color.border-pressed`), and **focus adds `shadow.focus-ring`**, as on the switch. A row's icon takes the row's text colour when the row is current, and `color.muted-foreground` otherwise. **This is not state 18**, which stays closed: it is plum's existing *where you are*.

**10.4 Each board's colour: a 12px square** (`size.swatch`, new) in the row's 16px icon slot (`size.icon`), with `radius.badge` corners, in the hue's `.strong`. `background_color` falls back on read, as on the tile ([ADR-017](adr-017-ui-foundations.md) §3.3).

| | Light | Dark |
|---|---|---|
| `.strong` against the sidebar | 5.15 – 9.62 | 16.34 – 17.83 |

This is the board colour's third shape, after the tile's left rail and the board's top rail. The tile's rail was rejected for rows: a coloured bar at a row's left is the usual sign of *the current item*, so every row would have looked current. *All boards* holds Phosphor `SquaresFour` in the same slot, so the names line up.

**10.5 Width: 240** (`size.sidebar-width`, new), at both densities. There is room for about 176px of name, around 23 characters, or 152 with a lock. Longer names end in an ellipsis. 224 and 256 (shadcn's default) were drawn: every 16px of sidebar comes off the board, and at 256 nothing of the fifth column shows at 1440.

**10.6 The app bar names the current row** (§5): the board's name on a board, *All boards* on 2.4. The alternatives were the workspace's name in the bar, which left 2.4's heading without a name, and an empty bar on 2.4, which left a 56px strip with nothing in it.

**10.7 The button's icon: Phosphor `SidebarSimple`**, Regular, at `size.icon` in a `size.target-min` box (24 Fine, 44 Coarse), in `color.foreground`. It shows what the button does. Phosphor `List`, three lines, was rejected: on a phone it usually opens a menu of actions, not a panel of places.

**10.8 The content dims below 1024: `color.scrim`**, new.

| | Light | Dark |
|---|---|---|
| `color.scrim` | `#0C0A09` at 32% | `#000000` at 60% |
| The panel against the dimmed canvas | 2.64 | 1.04: the panel's rule and `shadow.overlay` carry its edge |

The panel takes `shadow.overlay`, because it is temporarily lifted above the page ([ADR-017](adr-017-ui-foundations.md) §4). **The dimming is not the opacity channel.** ADR-017 reserves see-through for a single object that "isn't fully real". The scrim covers the whole screen at once, so it can't be read that way. 50% / 70% was drawn and dimmed the board further than needed. No dimming made the board look usable when a tap would only close the panel. **Stage 5's dialogs reuse `color.scrim`.**

**10.9 Rows.** `size.control-height` (36 Fine, 44 Coarse), `spacing.control-padding-x` 12, `radius.control`, `type.body`, `spacing.icon-gap` 8 after the icon slot. Rows sit `spacing.list-gap` (4) apart ([ADR-032](adr-032-auth-form-errors-and-password-rules.md)). The list sits **`spacing.sidebar-padding` (8, new)** in from the sidebar's sides and below its top row. The user menu row at the bottom holds an Avatar at `size.avatar` and your name in `type.body`, under a `color.border` rule, with the same 8 around it.

**`spacing.sidebar-padding` 8 is derived, at Fine.** The button's glyph sits at 16 + (24 − 16) ÷ 2 = 20 from the edge, and a row's content starts at 8 + 12 = 20, so the names line up under the glyph. At Coarse the button's box is 44, so the glyph sits at 30, 10px right of the names. That is accepted: changing `spacing.screen-x` for one button would move the whole app bar. It is not `spacing.column-padding`, which matches it only arithmetically ([ADR-022](adr-022-form-control-geometry.md) §4).

**10.10 Motion.** Sliding in and out is a container moving, so it takes [ADR-018](adr-018-state-vocabulary.md) §8's 260ms, with its arrive and leave eases. From 1024 up, hiding and showing move the same way. With `prefers-reduced-motion: reduce`, the sidebar cross-fades instead (110ms), as ADR-018 §8 says containers do.

> **Settled after acceptance (2026-11-21): the board glides.** The paragraph above left open what the board does from 1024 up while the sidebar slides. The two options ran side by side, in Light and Dark and in slow motion, on a rendered page:
> - **A jump**, recommended: the board takes its new width in one step, and only the sidebar slides. It kept ADR-018 §8's rule 3 (panels never move board objects) true as written.
> - **A glide**, chosen by the user: the board slides with the sidebar over the same 260ms and eases, so every card on screen moves 240px.
>
> The user asked whether the glide would slow the app, and the answer recorded is **no**. Its cost exists only during the 260ms after your press. The columns keep their 280px width, so no title re-wraps and the browser redraws only what is on screen. Nothing keeps running afterwards. The worst case is a choppy slide on a weak computer with a busy board, and building it as described below removes most of that.
>
> - **Rule 3 is narrowed for this one case.** The rule guards *a teammate did this*. The whole board moving together right after your own press can't be read that way, and no board object moves relative to another. Every other board movement is still the channel.
> - **Under reduced motion** the sidebar cross-fades and the board moves in one step, so A and B look the same.
> - **Build it as one width change and a slide, not a width animation.** Give the board its final width once, then move it into place with a `transform`, which the graphics chip handles. Current shadcn's `Sidebar` animates a spacer's `width` (`transition-[width]`), which re-measures the board on every frame. Replace that at scaffold time, along with the five differences §14 lists.
> - **Below 1024 nothing changes.** The sidebar slides over the content, and the board doesn't move.

### 11. What the board and the tiles fit now

**The board**, with the sidebar open. A peek is the visible part of the next column, measured after the gap before it. (ADR-019 §2's rows from 640 to 1024 measured their peeks with that gap included, so its 1024 row reads 136 where this reads 128. ADR-023 §2.2 noted the same difference at 1440.)

| Width | Pointer | Sidebar hidden | Sidebar open |
|---|---|---|---|
| 1024 | Fine | 3 + 128 | **2 + 176** |
| 1194, iPad landscape | Coarse | 4 | **3 + 46** |
| 1280 | Fine | 4 + 96 | **3 + 144** |
| 1440 | Fine | 4 + 256 | **4 + 16** |
| 834, 744, 640 and phones | — | ADR-019 §2 and §3, unchanged | Unchanged: the sidebar slides over |

**The tile grid on 2.4** (`repeat(auto-fill, minmax(264px, 1fr))`), with the sidebar open:

| Width | Sidebar hidden | Sidebar open |
|---|---|---|
| 1024 | 3 | **2** |
| 1194 (Coarse) | 4 | **3** |
| 1280 | 4 | **3** |
| 1440 | 5 | **4** |

2.3 has no sidebar, so its grid is unchanged. ADR-017 §1's height arithmetic is untouched: the sidebar takes width, not height, and it adds no row above the board.

### 12. The wordings

| Where | Wording |
|---|---|
| Sidebar, top row | *CollaboTask* |
| Switcher | the workspace's name |
| Switcher list, last row | *All workspaces* |
| Row to 2.4 | *All boards* |
| Label above the boards | *Your boards* |
| No joined boards | *You haven’t joined a board yet.* |
| Button label | *Hide sidebar* · *Show sidebar* |
| Button tooltip | *Hide sidebar · Ctrl + \\* (*⌘ + \\* on a Mac), and the same for *Show sidebar* |
| Skip link | *Skip to content* |
| Landmark name | *Sidebar* |
| 2.4's app bar title | *All boards* |

### 13. Tokens

| Token | Set | Value | Derived or chosen |
|---|---|---|---|
| `size.sidebar-width` | `base` | **240** | Chosen (§10.5) |
| `spacing.sidebar-padding` | `base` | **8** | Derived at Fine (§10.9) |
| `size.swatch` | `base` | **12** | Chosen (§10.4) |
| `color.scrim` | `theme-light` / `theme-dark` | **`rgb(12 10 9 / 0.32)`** / **`rgb(0 0 0 / 0.60)`** | Chosen (§10.8) |

Four tokens, five entries: **194 → 199**.

**Check before adding `color.scrim`:** that a Penpot colour token accepts transparency (`rgba()` or an 8-digit hex). If it doesn't, the scrim is drawn as a solid fill at layer opacity, bound by hand, and the binding check is told to expect it. Record the result in this ADR's *Operational note*.

### 14. The code

- **§9 gains** `--sidebar-width: 240px`, `--sidebar-padding: 8px`, `--swatch: 12px`, and `--scrim` in both modes.
- **The sidebar is shadcn's `Sidebar`, `collapsible="offcanvas"`**, restyled. Current shadcn differs from this ADR in five places, to confirm at scaffold time:
  - its width is `16rem`: set `SIDEBAR_WIDTH` to `var(--sidebar-width)`, and the mobile width to the same
  - its shortcut is ⌘/Ctrl + B: set `SIDEBAR_KEYBOARD_SHORTCUT` to `\`
  - it remembers its state in a `sidebar_state` cookie: keep it in `localStorage` instead, since nothing on a server reads it (ADR-011)
  - it switches to its mobile `Sheet` at 768 (`useIsMobile`): the sidebar switches at **1024**, with its own breakpoint, leaving ADR-019's 640 alone
  - its trigger uses Lucide's `PanelLeft`: use Phosphor `SidebarSimple` ([ADR-018](adr-018-state-vocabulary.md) §9)
- **Its colours:** point `--sidebar` at `--background`, `--sidebar-foreground` at `--foreground`, `--sidebar-border` at `--border`, `--sidebar-primary` at `--primary` and `--sidebar-ring` at `--ring`. `SidebarMenuButton` fills hovered and active rows with `--sidebar-accent`. Restyle it to §10.3 instead, an edge rather than a fill, as ADR-033 §14 restyled the Outline button. Don't use `SidebarRail`.
- **The sheet's overlay** (`bg-black/50`) becomes `bg-scrim`.
- **Routes** ([frontend-architecture.md](../frontend-architecture.md)): 2.4 and the board share a **layout route that renders the sidebar** and writes the remembered workspace. 2.3 stays outside it. The entry route `/` redirects to the remembered workspace, or to `/workspaces`, and sign-in lands there.
- **State:** whether the sidebar is open (from 1024 up) lives in a small persisted Zustand store, read inside `try`/`catch`. The last workspace is a `localStorage` value keyed by user id. The board list is the TanStack Query cache entry 2.4 already uses, filtered in a selector.

## Options considered

### What it holds
- **The current workspace's boards only.** It costs the same 240px and holds less.
- **Workspaces only.** The sidebar would not help someone move between boards, the commoner job.
- **Every board in every workspace, grouped**, the shape Trello and Notion use. It needs a new endpoint (or one request per workspace), and it gets long.

### Which screens show it
- **All three.** On 2.3 it repeats the page, with nothing chosen.
- **The board only.** The sidebar would appear and disappear as you go in and out of a board.
- **2.4 only.** The board keeps its width, but the sidebar then doesn't do its main job, switching board from inside a board.

### Hiding it
- **Never hidden.** It would cost the board a column at 1024 with no way back.
- **Hidden by default on the board.** Two defaults to learn, and hidden on the one screen where switching board matters.
- **A narrow strip of icons.** Boards have no icons, and 8 colours repeat.

### The width at which it stops sitting beside the content
- **640**, where the board's layout changes. Tablets in portrait would open the board with one column.
- **1280.** iPads in landscape and 1024-wide windows would lose it.

### Below 1024
- **A full-height sheet below 640**, as a "form or list". Two looks for one thing.
- **A bottom action sheet below 640**, as a "menu". Bottom sheets are for short menus, and a board list has no length limit.
- **No sidebar below 640**, with a back arrow instead. Phones would get around differently from every other device.

### *CollaboTask* and the user menu
- **Both stay in the app bar.** On the phone's board bar the title falls to about 128px.
- ***CollaboTask* moves, the user menu stays.** The same cost on the phone.

### 2.3
- **Kept as it is**, with the app opening on it. Most people would tap the same single tile each time they open the app.
- **Dropped**, with `/workspaces` sending you to your most recent workspace. It loses the roles and counts, and the onboarding screen becomes a special case.

### The switcher
- **A separate menu** (a popover, or a bottom sheet below 640). On a phone it stacks a second layer on the panel.
- **Reopening the last board** you had open in that workspace. A second remembered value, which can point to an archived board.

### The board list
- **Every active board, as on 2.4.** For an admin, a long run of locked boards they've never opened.
- **Two groups**, *Your boards* then *Other boards*. The same long list, split in two.

### The way back to 2.4
- **The list's heading as a link.** Easy to miss.
- **No separate way**, using the switcher. Hidden.
- **A breadcrumb in the app bar.** It costs the phone's board bar the width that kept the user menu out of it.

### The shortcut
- **None.** Keyboard users would tab to the top of the page to get a column back.
- **A single key, `[`.** It breaks SC 2.1.4.

### The look
- **Under a full-width bar.** The title stays still, but below 1024 the panel has a different shape from the sidebar beside the content, and *CollaboTask* sits under the bar.
- **A `color.muted` fill.** In Light it is the column panels' colour.
- **A plum tint for the current row.** It is Busy's fill, which ADR-033 kept off the switch, and at 1.12 in Light the mark rests on colour alone.
- **Grey and bold for the current row.** *Where you are* would look one way on the switch and another way here.
- **A left rail, or no colour, for each board.** The rail reads as *current*. With no colour, boards are told apart by name alone.
- **224 or 256 wide.**
- **The workspace's name, or nothing, in 2.4's bar.**
- **Phosphor `List`.**
- **Stronger dimming (50% / 70%), or none.**

## Out of scope (deferred, recorded)

- **A list of your boards across workspaces.** It needs an endpoint. Not in Phase 1.
- **The last workspace following you across devices.** It needs a server-side column and endpoint.
- **Live updates for boards you're not in.** No realtime message reaches you outside a board, so a board you are added to appears when the list reloads. A workspace-level realtime channel would be a backend change.
- **Starred or pinned boards, reordering the list, and *Create board* or *Create workspace* in the sidebar.**
- **Drawing the slide-over panel.** It is an overlay, so it is drawn at Stage 5, with the user menu it holds. 2.15 Phone board shows the button in its bar.
- **A tablet-width board.** ADR-020 §5 keeps none, and none is needed: below 1024 the board is unchanged.
- **SC 4.1.3**, announcing changes that happen without user action. PRD §5.4 keeps it open. The sidebar adds no realtime change of its own to announce.

## Downstream edits this ADR requires

Done before 2.5 is drawn.

1. **[ADR-017](adr-017-ui-foundations.md): annotation.** Its Status line records §9's twelfth extension. A note under §9's opening points here. A note under §3.3 records the board colour's third shape, the sidebar's square.
2. **[ADR-018](adr-018-state-vocabulary.md): annotation**, under the channel ledger: plum spends its *where you are* once more, on the sidebar's current row, and state 18 stays closed. A note under §8's container row adds the sidebar.
3. **[ADR-019](adr-019-responsive-touch-scope.md): annotation.** Its Status line records the supersession of §2's 1024 and 1440 rows while the sidebar is open. A note under §2's table points to §11 here. A note under §8 names the 44px slot as the sidebar's button.
4. **[ADR-021](adr-021-design-inventory-corrections.md): annotation**, under §5's table: the fifth kind, *navigation*.
5. **[ADR-033](adr-033-workspace-screens.md): annotation.** Its Status line records the two supersessions, on 2.4 only. Notes go under §2 and §4.
6. **Penpot `Collabotask App`:**
   - the five token entries, after the transparency check in §13
   - the sidebar's parts: a row with its `State` (Rest · Hover · Pressed · FocusVisible · Current) and its icon slot, and the sidebar itself
   - **2.4 and its five state boards redrawn** with the sidebar and the new app bar
   - the Foundations catalogue's new rows
   - the binding check and the flips, as at every stage

   Question 4 is asked of each before drawing.
7. **[`design-file-inventory.md`](../design-file-inventory.md):**
   - a section on the sidebar
   - §1.4 and §1.5's app bars
   - §2's rows for 2.4, 2.5, 2.7, 2.8 and 2.15
   - §3's presentation table gains *navigation*
   - §4.2 gains the new parts
   - §6.5 gains the seventh empty flavour
   - §7 records the stage
8. **[frontend-architecture.md](../frontend-architecture.md):** the route tree (the shared layout route and the `/` redirect) and *State ownership* (the sidebar's open state and the last workspace).
9. **`CLAUDE.md` › `## Now`:** a pointer.
10. **`docs/build-log.md`:** no entry, following ADR-020 to ADR-033. Stage 4 is not closed.
11. **PRD, user stories, SRS, `CONTEXT.md`:** no change. No scope, contract or domain term moved. *Sidebar* is a screen word, not a domain term.

## Operational note

**The transparency check (2026-11-21): a Penpot colour token takes transparency.** Three throwaway tokens went into `theme-light`, each bound to a rectangle's fill:
- `rgba(12, 10, 9, 0.32)`: accepted. The fill reads `#0c0a09` with its own opacity at 0.32. The layer's opacity stays unset, so §7.5's binding check sees a bound fill and reports nothing.
- `#0C0A0952`, an 8-digit hex: accepted, but `52` is 82 ÷ 255, so the opacity comes out at 0.3216.
- `rgb(12 10 9 / 0.32)`, the form §13 writes: **refused** as an invalid token value.

The token's resolved value leaves the alpha out (it reports `#0c0a09`), so check the fill, not the token. **`color.scrim` is stored as `rgba(12, 10, 9, 0.32)` in `theme-light` and `rgba(0, 0, 0, 0.6)` in `theme-dark`.** A shape bound to it read `#000000` at 0.6 after the switch to Dark. The code keeps §13's form, since the refusal comes from Penpot's parser, not from CSS. The probes were removed, and the five entries are in the file (194 → 199).
