# ADR-019: Responsive Scope and Touch Interaction

- **Status:** Accepted — **§8's `+N` overflow chip and the arithmetic built on it superseded by [ADR-031](adr-031-presence-cluster.md)** (2026-11-20): `+N` is a round seat in the stack, so the phone's cluster is 72–75px wide, not 96. §8's cap of three faces below 640px and its 8px overlap at Coarse stand; from 640px up the cap is five. **§2's 1024 and 1440 rows superseded by [ADR-034](adr-034-side-menu.md) while the sidebar is open** (2026-11-21): a 240px sidebar sits beside the content from 1024 up, so 1024 shows 2 columns + 176px and 1440 shows 4 + 16. Every row below 1024 stands, because there the sidebar slides over the content and takes no width. **§8's 44px "menu" slot is named by the same ADR**: it is the sidebar's show/hide button. Everything else stands as written.
- **Date:** 2026-11-17
- **Scope:** The responsive and input scope of the **app zone** (React + Vite SPA — [ADR-011](adr-011-frontend-framework.md)): how many board layouts exist and where they change over, how a drag begins and how far it reaches on a touchscreen, and the coarse-pointer density scale. **Supersedes [ADR-017](adr-017-ui-foundations.md) §8 only.** Everything else in ADR-017 stands — the token vocabulary and values, §1's fine-pointer density, §2's contained-column structure, the type scale, elevation, and the colour modes. This ADR *narrows* §8 rather than replacing the design behind it. The public zone (Astro landing — [ADR-013](adr-013-public-zone-framework.md)) is **not** covered: a landing page is responsive by default and needs no decision record. The numbering skips 018, which [ADR-017](adr-017-ui-foundations.md) already reserves by filename for the state vocabulary; that ADR was written the following day (2026-11-18) and picks up all four constraints in §*What ADR-018 inherits* below.

## Context

ADR-017 §8 set the app zone to **pointer input at ≥1024px, one density scale**, and committed a phone interface to Phase 2 as its own design. Its reasoning was sound on its own terms and two of its three arguments survive intact. The requirement changed: the product is to be usable on tablets and phones in Phase 1, with card dragging working on both.

The reversal is cheap to make **now and only now**. No frontend code exists. The cost of this decision is design time; the cost of the same decision taken after the board is built is rework across every component.

One argument in §8 does not survive, and it is worth naming precisely because the rest of this ADR depends on its failure:

> *"on a 375px screen only one 280px column is visible, so a drag must cross columns the user cannot see — which no amount of tuning fixes."*

That is true **only if the column stays 280px**. The premise is a fixed column width, not a property of phones. If the column is sized relative to the viewport, the next column is always partly on screen, and a drag to the adjacent column becomes a short move onto a visible target. The objection survives for *distant* moves — which §6 answers by a different route entirely.

Four facts from the existing system shaped what follows:

1. **Columns are unbounded.** `GetKanban` returns every card in every column with no pagination ([column_response.go](../../../collabotask-backend/internal/delivery/http/response/column_response.go)). A column can be two hundred cards long on a 390px screen, which rules out any layout that stacks columns vertically.
2. **The move API is position-based, not index-based.** `POST …/cards/{id}/move` takes `{to_column_id, to_position float}` ([card_request.go](../../../collabotask-backend/internal/delivery/http/request/card_request.go)). A non-drag move UI cannot offer "put it third" — it must name a **neighbour** and compute the midpoint client-side, exactly as a drop does.
3. **A read-only board is a normal state, not an edge case.** `CheckMutateAccess` denies every mutation to a non-joined plain member while `CheckViewAccess` lets them open the board — *"they can see the board, they just can't edit it"* ([board_access_check.go](../../../collabotask-backend/internal/usecase/common/board_access_check.go)). Any workspace member browsing a workspace-visible board they have not joined sees a complete, entirely uneditable board.
4. **No board event carries the actor.** Per ADR-017's Context, presence carries the whole weight of *other people are here*. It therefore cannot be dropped on narrow screens to reclaim space.

---

## Decision

### 1. Two axes — width drives layout, pointer drives density

**Screen width and input type are independent inputs, and they are not interchangeable.** An iPad in landscape is 1194px wide and wants the desktop *layout* with touch-sized *targets*. A desktop window dragged to 700px wants the narrow layout with mouse-sized targets. Keying everything off width alone gets both of those wrong.

| Axis | Drives | Mechanism |
|---|---|---|
| Viewport width | Which board layout, modal presentation, bar composition | Breakpoints (§2) |
| Pointer type | Card padding, avatar size, gap, minimum target | `pointer: coarse` (§7) |

**Touch sensors are registered unconditionally** — never behind a pointer check. Hybrid devices (an iPad with a keyboard case, a Surface) report their primary pointer inconsistently across OS versions, and the failure mode of guessing wrong is a drag that silently does nothing. Pointer type adjusts *sizes*; it never gates *capability*.

A consequence for scaffolding: Tailwind v4 ships `pointer-coarse` as a first-class variant; on v3 it is an arbitrary variant. This is the **second** reason to confirm the Tailwind major version at scaffold time — ADR-017 §9 already warns that its hex token values break a v3 shadcn setup.

### 2. Two board layouts, crossover at 640px

ADR-017's board is fixed 280px columns scrolling horizontally. That design **degrades gracefully far below 1024px**; you simply see fewer columns:

| Width | Usable (−32px padding) | Columns visible |
|---|---|---|
| 640px | 608 | 2 + a 40px peek |
| 744px *(iPad mini, portrait)* | 712 | 2 + 144px |
| 834px *(iPad 11″, portrait)* | 802 | 2 + 234px |
| 1024px | 992 | 3 + 136px |
| 1440px | 1408 | 5 |

> **With the sidebar open, the 1024 and 1440 rows change.** [ADR-034](adr-034-side-menu.md) (2026-11-21) puts a 240px sidebar beside the content from 1024 up, on 2.4 and the board. While it is open, 1024 shows 2 columns + 176px and 1440 shows 4 + 16. Those peeks are measured after the gap before the next column; on that measure this table's own 1024 row would read 128. The rows from 640 to 834 are unchanged: below 1024 the sidebar slides over the content and takes no width. ADR-034 §11 has the full table, including 1194 and 1280.

Nothing breaks at any of those widths. **A tablet needs touch targets, not a new layout** — and targets come from the pointer axis. So the board has two layouts, not three, and ADR-017 §1–§2 apply unchanged from 640px up.

Below 640px it does break down: 1.4 columns at 430px, 1.2 at 375px — a column plus a useless sliver.

**The crossover is 640px (Tailwind `sm`), not 768px.** This is a deliberate, load-bearing choice of number: **an iPad mini in portrait is 744px.** At a 768px crossover that device falls into the phone layout and renders a single ~670px-wide column — one absurd column on a tablet. 640 catches it on the correct side.

### 3. The phone board — a horizontal snap strip

Below 640px, columns remain side by side and become viewport-relative with scroll-snap:

```
column width:  min(100vw − 72px, 420px)

at 390px →  318px column │ 16 left pad │ 12 gap │ 44px peek
at 430px →  358px column │ 16 left pad │ 12 gap │ 44px peek
at 375px →  303px column │ 16 left pad │ 12 gap │ 44px peek
```

The 72px reservation is not arbitrary: it leaves the next column peeking by exactly **44px — one standard tap target wide**. The peek is therefore a legal drop target by the same rule that governs every other control, which is what makes §5's adjacent drag work. The 420px cap stops the column becoming absurd as the viewport approaches the crossover.

Vertical column scrolling and horizontal board scrolling both remain; the card modal becomes a **full-height sheet** rather than a centred dialog below 640px.

Rejected: one column with a tab switcher, and a vertical accordion. Both in *Options Considered*.

### 4. Drag pickup on touch — long-press, no handle

**A drag arms on a press-and-hold of 200–250ms with a drift tolerance of ~5px.** Drift beyond the tolerance during the hold is a scroll, not a drag. The whole card is the drag target; there is no grip handle.

The constraint that forces this: dnd-kit documents `touch-action: none` on draggable elements, but **cards cover roughly 90% of a column's surface**, so applying it to the card kills both scroll axes. Long-press is the only option that leaves scrolling intact — `touch-action` stays permissive and `preventDefault` fires only once the drag has armed.

- **The pickup cue costs no new channel.** ADR-017 spends **shadow** on *"temporarily lifted above the page"*. That is precisely the moment a long-press arms.
- **The hold duration is a tuning range, not a token.** 250ms is dnd-kit's suggested default, iOS long-press is nearer 500ms, Trello's feels faster than both. The final value comes from §10, not from this document.
- **iOS Safari requires suppression.** A long-press on text raises Safari's selection callout and magnifier, which will appear on top of a lifted card unless `-webkit-touch-callout: none` and `user-select: none` are set on the card.

### 5. Drag reach — adjacent only below 640px

**Below 640px the board does not scroll while a drag is in flight.** Valid drop targets are the current column and the peeking neighbour. Distant moves go through §6.

This is chosen as much for what it removes as for what it provides. Auto-scrolling a scroll-snapped container during a drag means disabling snap on drag start and restoring it on drop, on cancel, and on a drag interrupted by an incoming WebSocket update — while aiming a thumb at a ~40px edge zone. **Freezing the board deletes that entire class of problem**, and since only two columns are ever on screen, there is nothing off-screen for the user to aim at: it reads as "nothing else to drop on" rather than as a limit.

**At 640px and above, drag keeps full reach with ordinary dnd-kit auto-scroll** — there is no snap strip there to conflict with. The edge zone widens at coarse pointer so a thumb can hit it.

### 6. The move picker — required at every breakpoint

A **"Move to…"** path exists at every width, not only on phones.

This is a conformance requirement, not a convenience. **WCAG 2.2 SC 2.5.7 (Dragging Movements, Level AA)** requires that any function using a dragging movement also be achievable with a single pointer without dragging, unless the dragging is essential. Reordering a Kanban card is not essential dragging in that sense. ADR-017 already treats WCAG AA as binding throughout its colour work; this is the same standard applied to gesture.

**Removing the picker at any breakpoint is a conformance regression, not a simplification.** Anyone proposing it must supersede this ADR.

Three entry points, one code path:

| Entry | Availability |
|---|---|
| Card modal → Move | All widths |
| `⋯` on the card | Hover-revealed at fine pointer; **persistent** at coarse |
| `M` on a focused card | All widths — also the keyboard move path |

Because the move API takes a float position, the picker offers a **column** plus a **neighbour** — *top of column*, *after "«card»"*, *bottom of column* — and computes the midpoint exactly as a drop does.

The persistent `⋯` goes in the card's **top-right corner**, with the title block reserving 32px of right padding at coarse pointer. It explicitly does **not** go in the meta row: ADR-017 §1 makes that row *conditional* on an assignee or due date, and putting `⋯` there would force it to render on every card.

### 7. The coarse-pointer density scale

| | Fine (ADR-017 §1) | **Coarse** |
|---|---|---|
| Card title | 14px / 20px, 2-line clamp | *unchanged* |
| Card padding | 8px / 12px | **12px / 12px** |
| Card avatar | 20px | **24px** |
| Meta row (conditional) | 24px | **28px** |
| Card gap | 8px | **12px** |
| Column header | 36px | **44px** |
| Title right padding | — | **32px** (for `⋯`) |
| Minimum tap target | — | **44px** |

Per-card height becomes `12 + lines×20 + (28 if meta) + 12`.

**The smallest possible card is exactly 44px tall** — one line, no meta row — so every card is a legal tap target by construction rather than by special-casing.

The type scale is deliberately untouched. ADR-017 spends the 5-step scale on reading order; changing sizes for touch would spend a channel that is already committed.

Anchored the way §1 is, against a real device:

| Measurement | Value | Derivation |
|---|---|---|
| Reference viewport | **390 × 844** | iPhone 14/15 |
| Visible column body | **588px** | 844 − 140 iOS Safari chrome − 56 app bar − 16 padding − 44 column header |
| Representative card | **~80px** | §1's mix, re-measured against the coarse padding and meta row |
| Fully visible cards | **6** | 6×80 + 5×12 gaps = 540px |
| Seventh card | ~36px | peeks, which is §1's intended "the column continues" cue (588 − 540 − one 12px gap) |

The card *sizes* are arithmetic and are decided here. The **representative card mix** is carried over from ADR-017 §1 and should be re-measured once real content exists — that is the one number in this table resting on an assumption rather than a measurement.

### 8. Presence on a narrow bar

ADR-017 treats the always-visible presence cluster as a foundation, because no board event carries an actor. It therefore survives narrow screens rather than collapsing.

**Below 640px: at most 3 avatars plus a `+N` overflow chip**, the chip opening the full list as a sheet.

```
375px  −32 padding              = 343
       −44 menu   −44 actions   = 255 for title + presence

3 avatars @24px, 8px overlap    =  56
overflow chip                   =  32
gap                             =   8
                                  ───
presence                        =  96
title                           = 159px  ≈ 18–20 characters
```

The cap is set by that arithmetic. Five avatars leaves the title ~15 characters, which truncates real board names.

> **The chip became a seat, and the sum shrank.** **[ADR-031](adr-031-presence-cluster.md)** (2026-11-20) draws `+N` as a round seat at a face's size that overlaps the last face like another face, with no gap before it. It adds only its visible 16–19px, so three faces and the seat are 72–75px and the title gets about 180px. While reconnecting, the 90px *Reconnecting…* label takes the seat's place below 640px. The whole cluster, not the seat, is the one target (ADR-018 §6), `size.target-min` tall. From 640px up it shows at most five faces, with you first.

> **The 44px "menu" is the sidebar's button.** [ADR-034](adr-034-side-menu.md) (2026-11-21) adds a sidebar on 2.4 and the board, and its show/hide button takes this slot. On a phone the board's bar is the button, the title, presence and `⋯`, so the title keeps the ~180px above. The user menu is not in this bar. It sits at the bottom of the sidebar, because a 44px avatar here would cut the title to about 128px.

### 9. Build order — structural now, touch input as one pass

The two halves of this change have different retrofit costs, and the work is cut along that line.

**Layout flexibility is structural** — it is baked into how each component is written, Tailwind's narrow-first defaults run with it, and retrofitting means reopening every file. It is done **as each screen is written**.

**Touch input is additive** — a sensor added to an existing `DndContext`, a dialog calling a mutation that already exists. It lands as **one focused increment**, which is also the increment where §10's device tuning happens, rather than being smeared across twenty component PRs.

Three things must exist from the first commit regardless, because they are the expensive ones to add later:

1. **`moveCard(cardId, toColumnId, toPosition)` standing on its own**, not inside the drag handler. It has three callers: drag, picker, keyboard.
2. **Both density scales defined in the token file**, even while only one is consumed. ADR-017's own Context warns that changing a token's *role* after components reference it is a sweep.
3. **The `⋯` slot reserved in the card layout**, even when it renders nothing.

### 10. Verification

**The gesture is verified by hand on real devices; everything else is automated through the picker.**

Because §6 puts the picker at every breakpoint and it calls the same `moveCard`, the **logic** — that a move lands where intended, that the picker only offers valid targets, that each width renders its layout — is ordinary automated UI testing with no touchscreen involved. Only the **feel** needs hands: does the card lift, does scrolling survive, is the hold the right length.

Minimum device set, each covering something the others do not:

| Device | Covers |
|---|---|
| iPhone (Safari) | The selection-callout conflict of §4; the strictest gesture quirks |
| Android phone (Chrome) | Different scroll/drag arbitration |
| iPad | Tablet layout *with* fingers — a combination neither phone reaches |

Browser touch emulation is explicitly **not** sufficient: it is wrong about press timing, scroll-versus-drag arbitration and momentum, which is the exact list of things §4 and §5 decide.

---

## What ADR-018 inherits

Four constraints this ADR adds to the state-vocabulary brief:

1. **The cursor channel is void at coarse pointer.** ADR-017's channel budget reserves *Cursor* for *"what you can and can't do here"*. A touchscreen has no cursor, so that meaning needs a home — or is unavailable to touch users.
2. **Hover is void at coarse pointer too.** Hover is not a channel in ADR-017's sense — it is a *trigger* — but revealing a secondary control on hover is the default habit, and at coarse pointer there is nothing to trigger it. **Any affordance revealed by hover needs a coarse-pointer equivalent that does not depend on hovering.** §6's persistent `⋯` is the one instance this ADR settles; ADR-018 states the general rule.
3. **The read-only banner must be persistent, not dismissible.** ADR-017 resolves read-only as *"removing controls plus a banner — which costs zero channels"*, which still holds. But on touch the banner is the **only** remaining signal: a read-only user presses and holds, nothing happens, and without the banner they cannot distinguish "no permission" from "broken". Per Context fact 3, this is a routine browsing state.
4. **Drag states must cover an armed gesture, not just a mouse-down.** There is a 200–250ms window between touch and pickup with no desktop equivalent, and it is the only state in the set that can be **cancelled without completing** — a drift beyond tolerance turns it back into a scroll.

ADR-017's existing drag-state channels are unaffected: opacity still covers drag-source, the dashed border still marks the drop placeholder, and the column panel tint still marks drag-over — including the 44px peek, which is a drop target under §5.

---

## Options Considered

### Keeping ADR-017 §8 and deferring to Phase 2

The status-quo option, and its two surviving arguments are real: a second density scale is genuine cost (§7 is the bill), and a phone Kanban deserves design rather than reflow. It fails on timing. The decision is being made when no frontend code exists, which is the only moment the layout half is nearly free; deferring converts §9's "structural now" into a retrofit of every component. The third argument — that cross-column drag cannot work on a phone — rests on a fixed 280px column and does not survive §3.

### Width-only adaptation

One axis, one scale per breakpoint, nothing to detect. Rejected on real devices: an iPad in landscape (1194px) would take the desktop layout *and* mouse-sized targets — 20px avatars and 8px gaps aimed at a thumb — while a narrow desktop window would get touch sizing for a mouse.

### Pointer-only adaptation

Rejected: it adapts targets but never the layout, so a 390px phone still renders 280px columns side by side. It does not deliver the requirement.

### A third, purpose-built tablet layout (640–1023px)

Rejected by the table in §2. The 280px board already handles every width in that band without complaint; a third layout would cost a design pass, two more Figma frames (one per density scale) and a third width class in testing, to improve widths that are not broken.

### One column with a tab switcher at phone width

The best *reading* layout at 390px — full-width cards show more title before the 2-line clamp. Rejected because it makes cross-column drag genuinely impossible: the only cross-column target would be the tab strip, which is a long-established bad pattern. It would have reduced drag to within-column reordering, which does not meet the requirement.

### A vertical accordion at phone width

Rejected on Context fact 1. Columns are unbounded, so a 200-card section makes the accordion unusable, and a vertical-only drag conflicts with page scroll more directly than any other option here.

### A drag handle instead of long-press

Instant pickup with no delay, and `touch-action: none` safely scoped to the grip. Rejected because the card's affordance budget is spent on the `⋯` of §6: a grip would be a second persistent control on a card whose density ADR-017 §1 derived tightly, and it is a smaller target than the whole card. Tracked — if §10 finds the long-press delay unacceptable on device, this is the fallback, and it reopens §7's title padding.

### Full-reach drag with edge auto-scroll below 640px

Matches the desktop model exactly and imposes no reach limit. Rejected on the cost in §5: snap toggling across drop, cancel and WebSocket interruption, a thumb wider than the edge zone, and the hardest behaviour in the plan to verify — to duplicate, worse, what §6 already does well.

### Paged reach — one column per edge-hold

A genuine middle path: deterministic, and it cooperates with scroll-snap instead of fighting it. Rejected because it is a gesture no user has encountered before, with cooldown timing to tune, for a job §6 already does with a familiar control.

### Recording SC 2.5.7 in the SRS

Considered and rejected on inspection of the document. The SRS is a backend contract end to end — use cases, WebSocket frames, DTOs, schema, audit — with no section describing client behaviour and no existing accessibility content; its only non-functional line sits in §10 *Open Notes* marked "indicative". A lone WCAG line there would be read by nobody about to delete a React component. §6 carries it instead, where the project's never-edit-an-ADR rule means dropping the picker requires writing down that AA conformance was given up.

---

## Out of scope (deferred, recorded)

- **The Astro landing page.** Responsive by default; a landing page needs no decision record to work on a phone. ADR-017's note that the public zone may diverge visually is unaffected.
- **The work a declared AA target implies.** The *target itself* is no longer open: **PRD §5.4 declares WCAG 2.2 Level AA** (2026-11-17), closing the gap this ADR first recorded — that the project practised AA without any document stating it, leaving every accessibility argument to re-derive its own authority. The two gaps PRD §5.4 records stay out of scope here: a **24×24px minimum target at *fine* pointer** (SC 2.5.8) and **announcing realtime board changes** (SC 4.1.3). Note that §7's table sets a minimum target size only at coarse pointer — the fine-pointer floor is a constraint ADR-018 now inherits.
- **The state vocabulary.** Still ADR-018, now with the three additional constraints above.
- **Column width at fine pointer.** 280px remains ungrilled, as ADR-017 recorded.
- **Native applications.** Unchanged: out of scope, and not what this ADR provides.
- **Phase-2 deferrals** — comments, labels, attachments, public boards, refresh tokens, account deletion — untouched.

## Downstream edits this ADR requires

1. **PRD §5.2** — remove the *Phone & touch interface* deferral row; it is no longer deferred.
2. **PRD §5.3** — rewrite the non-goal bullet. "Responsive web is the target" becomes true as written for the first time, and the clause deferring touch to Phase 2 is withdrawn.
3. **`CLAUDE.md` › `## Now`** — record that ADR-019 precedes ADR-018 in the frontend sequence.
4. **SRS** — no change. See *Options Considered*.
5. **`CONTEXT.md`** — no change. "Snap strip", "long-press" and "picker" are interface vocabulary, not domain terms, and nothing in Workspace / Board / Column / Card / Position / the role model moves. `Position` already states it is *"a coordinate on an infinite line, not an index"*, which is the constraint §6's picker honours. Worth confirming rather than adding: **moving a card is one domain action with three interface triggers** — `ActivityActionMoved` is correctly indifferent to whether a finger, a mouse or the `M` key caused it.
6. **`docs/build-log.md`** — an entry when the work completes, not now.
