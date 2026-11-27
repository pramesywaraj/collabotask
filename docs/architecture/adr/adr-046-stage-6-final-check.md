# ADR-046: Stage 6, the Final Check — What It Found, and One Recorded Exception

- **Status:** Accepted
- **Date:** 2026-11-27
- **Scope:** The result of the design file's final check (inventory §7.5–§7.6), run on 2026-11-25 to 2026-11-27 over all five pages in all four Theme × Density combinations. It settles:
  - **a contrast exception for inert, dimmed objects**: the avatar initials on a dimmed card, and everything inside the column a drag leaves behind
  - what the check changed in the file, and what it left alone on purpose
  - what the check did **not** look at

  **No token is added (222), no part (34), and no board is added or renamed.** It **extends [ADR-018](adr-018-state-vocabulary.md) §3** and does not supersede it: the title and meta row of an in-flight card keep their measured 4.70:1.
- **Upstream edits:** none. The PRD, the user stories and the SRS say nothing about the contrast of a dimmed object.

## Context

The check had two halves. §7.5's **binding check** looks for typed-in values that should be tokens. It came back clean on every page: every row it returned is on the inventory's expected list. The **flips** put the file into each combination and looked for three things on every page: text at the wrong size, text that does not fit its box, and text too faint to read. Exports were looked at by eye on the screens the ADRs say change most.

The flips found four things. Two are about contrast and need a decision, because they put the file's own numbers against WCAG 2.2 SC 1.4.3 (4.5:1 for body text). **The user took all four recommendations on 2026-11-27.**

### The measurements

Method: each text's colour composited at its own opacity, over the surface behind it, itself composited layer by layer. Ratios are rounded to two places.

| Where | What | Light | Dark |
|---|---|---|---|
| `Card — states`, the `InFlight` and `DragSource` looks | the avatar's initials | **2.38** | **3.92** |
| `2.14 Drag — a card`, the dimmed card | the avatar's initials | 2.94 | 4.48 |
| `2.14 Drag — a column`, the dimmed *In Progress* column | the count chip | 2.67 | 4.24 |
| | a due date: *Nov 28*, *Due tomorrow*, *Overdue · Nov 20* | 2.83, 2.44, 3.10 | 3.06 to 4.45 |
| | the initials *AP*, *JT*, *SL* | 3.07, 2.47, 2.95 | 4.08 and above |
| | the *Add card* label | 2.76 | passes |
| | the column title *In Progress* | 4.44 | passes |
| | the card titles inside it | 4.50, at the limit | passes |

[ADR-018](adr-018-state-vocabulary.md) §3 measured the **title** and the **meta row** of an in-flight card at 4.70:1, and promoted the meta row to `color.foreground` to get there. It never measured the avatar, which `Card` gained later ([ADR-025](adr-025-identity-hues-as-tokens.md) §5). [ADR-039](adr-039-drag-and-phone-board.md) §4 dims a dragged column's place as a whole panel, so nothing inside it is promoted. The same reasoning holds for the cards left inside it: they are in their Rest look, because only the panel is dimmed.

The phone's drag screen, `2.15 Phone — dragging a card`, was not flagged.

## Decision

### 1. Text inside a dimmed, inert object is exempt from 4.5:1

**What is exempt:**
- the avatar's initials on a card at `opacity.in-flight`
- every text inside a column at `opacity.in-flight` while its drag is under way: the title, the count, the dates, the initials, the *Add card* label, and the titles of the cards inside it

**Why it holds.** WCAG 2.2 SC 1.4.3 exempts text that is part of an **inactive user interface component**. ADR-018 §3 already makes a dimmed object inert: it cannot be dragged, opened, or reached through its menu. And nothing in either case is lost:
- **A dragged column** is shown in full strength at the same moment, as the lifted copy ([ADR-039](adr-039-drag-and-phone-board.md) §4). What stays behind is a slot.
- **An in-flight card** returns to full strength when the server answers, or after at most ten seconds, when the axios timeout fails the mutation and the card is restored (ADR-018 §3). Its title and meta row stay at 4.70:1, and the assignee's name is in the drawer.

**What is not exempt, and does not change:**
- the title and the meta row of an in-flight card stay at **4.70:1**. This ADR extends ADR-018 §3's promise only for the avatar and the dragged column.
- anything outside a dimmed object, and anything that can be operated.
- the user's choice here is **to record the exception, not to fix it**. The exception is narrow because the objects are inert and short-lived, not because 2.38:1 is acceptable for text a person has to read.

### 2. The twelve documentation boards without a fill are filled

Twelve boards on `Components` had no fill: `Card`, `Input`, `Button`, `Avatar`, `Count chip`, `Column header`, `Banner`, `Skeleton block`, `Toast`, `Presence cluster`, `Form message` and `Password rule`, each "— states". Their small `Label /` captions are `color.muted-foreground`, and with no fill they sit on whatever colour Penpot's own canvas is. Measured against white, in Dark they are **1.49:1**; eleven of the twelve are affected, and `Card — states` has no caption that was flagged.

**Done 2026-11-27:** each is filled with `color.background`, bound to the token, exactly as the 22 newer boards already were (`Segment — states` onward). Nothing in the product changes. All 34 boards on `Components` now have a bound fill.

### 3. The Count chip board's cut-off chips stay

`Count chip — states` draws *128*, *12 cards* and *3 members* clipped, because the three copies stay at 23px wide where they need 31, 61 and 81. **Five repairs were tried and none held**: resizing a copy, toggling its sizing, toggling its text's grow type, re-measuring the words, and a resize followed by auto sizing. Each took effect and was put back to the main's 23px within about fifteen seconds.

**The chips on the real screens are correct**: the tiles on 2.3 and 2.4 and the column headers on the board all show their full text. [ADR-026](adr-026-count-chip.md)'s *Operational note* already records that Penpot does not draw the chip's minimum width, and this is the same family of fault. It is left as a known fault of one documentation board, **and a developer must not read the cut-off as the design.**

### 4. The check's four Penpot repairs

Found by looking, fixed on 2026-11-26, none of them a design decision:

| Where | What was wrong | Fix |
|---|---|---|
| the `List` part's `Here / Sheet` head, and its copies | the head's left padding of 12 was ignored, so the title sat flush with the edge | the head's padding mode went from *same on both sides* to *one value per side* |
| the `Form` part's `Fields`, and every copy | `Fields` clipped its content, so a focused field's ring lost its outer edge | `clipContent` off on `Fields`, in all seven looks |
| `Invite people at 390` | one email chip's text drew blank | the words were re-set, so Penpot measured them again |
| the two `Footer` boards under *Add card* on the drag screens | the same padding-mode trap, harmless today because the sides differ | switched to *one value per side*, bound and 12 / 0 as before |

## What the check did not look at

- **`Board`, in Dark + Fine:** 13 of its 30 desktop screens were looked at by eye. The other 17 are overlays over 2.5 built from the same parts, and the data checks covered all 30.
- **The parts, in Dark + Fine:** only `Button` and the Foundations board were looked at. The data checks covered every part.
- **Light + Fine** was looked at on a few Workspaces screens only. It is the combination the file is left in.
- **The fit check could not measure a page that was not drawn.** On `Board` in Dark + Fine it read every text's size as 0 × 0 and passed everything. It was re-run with the page drawn, in Light + Fine, and returned only the cut-offs that are designed: the fifth column that peeks past the screen's edge, the column pushed off by a drag, and the board scrolled to its end.

## Options considered

For §1, the contrast of a dimmed object:

- **Raise the dim to about 80%.** ADR-018 §3 measured this: at 80% a composited card against its panel *is not a signal at all*. It would pass, and the state would stop reading as one.
- **Promote the contents to `color.foreground`**, as the meta row already is. It would help the dates, and ADR-018 §3's reasoning holds that internal hierarchy has nothing left to say while an object is de-emphasised as a whole. It was **not measured** for the count, the initials or the *Add card* label, and it would flatten a dragged column until it no longer looked like the column the lifted copy shows.
- **Frame-only translucency.** Rejected in ADR-018 §3 for its 4-unit-per-channel signal.
- **Record the exception.** **Chosen.**

For §3: **fixing the chips by drawing each one at its real width on a plain board** was not tried. A copy that is not following its main would stop testing the part.

## Out of scope (deferred, recorded)

- Whether an audit tool in the build flags the dimmed text. Automated checkers (axe, for one) usually skip an element marked `aria-disabled`, and the dimmed object is inert, so it can be marked. **Decide at build, and check the tool.**
- The Foundations board's contrast notes. It states ratios for the Rest looks only, and is not changed.

## Downstream edits this ADR requires

- **Inventory** §7.6: Stage 6 ticked, with these findings. §7.2: the Foundations row brought up to 222 tokens, and a note on the Count chip board.
- **ADR-018 §3** and **ADR-039 §4** carry a short note pointing here. They are annotated, not edited.
- `CLAUDE.md` `## Now`, `docs/build-log.md` and the memory note for the design tool.
- **Frontend:** nothing at scaffold time. See *Out of scope* for the one build-time question.

## Operational note

Found while running the check, in the order a later session will meet it:

- **A page that is not open reads as 0 × 0.** Reading a page from another page works for fills, bindings, sizes and tokens, but every text's `textBounds` is zero, so a fit check passes everything. Exports of such a page draw stale text, like the button that read *Add cc* for *Add column*. **Open the page first**, and read it afterwards.
- **A flex board whose padding mode is *simple* uses its right padding for the left** when only one side is set, and ignores a typed or bound left value. The trap is cleared by setting `paddingType` to `"multiple"`. Run `paddingType === "simple"` with unequal sides across the file to find it. Only the two `Footer` boards were left.
- **A board that clips its content cuts a child's shadow**, and a focus ring is a shadow. `clipContent` on a part that holds a focusable field is a defect.
- **Opening `Board` from the plugin froze the tab** in Dark + Coarse (once) and in Dark + Fine (three times), for several minutes, and one wait ended in a lost plugin connection. In Light + Fine it was opened by hand, with the window beside VS Code, and it settled. Open it by hand and leave it alone; the tab answers the plugin's heartbeat again when it has finished drawing.
- **`/mcp` in Claude Code reconnects Claude to the server, not the plugin.** The plugin has to be started again from Penpot's plugin menu. The plugin's `storage` is empty after every reconnect.
- **A flip lands even when the tab stops answering.** The Theme switch back to Light, made while `Foundations` was open, left the tab responsive. The Density switch to Fine, made from `Account`, froze it for about a minute and landed all the same. After any switch, read `themes` and `sets` and confirm exactly one of each group is on.
