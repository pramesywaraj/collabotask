# ADR-018: State Vocabulary — What Every Interaction State Looks Like

- **Status:** Accepted — **§6 given two things it left open by [ADR-024](adr-024-control-and-field-internals.md)** (2026-11-20): *pressed* (15) on a **filled** button reuses `--primary-hover`, since §6's pressed values are for an edge; and the focus ring's 2px offset is drawn as two stacked shadows (`shadow.focus-ring`), the way Tailwind's `ring-offset` renders it. Every value in §6 stands as written. **§7's skeleton fill superseded by [ADR-029](adr-029-skeleton-block.md)** (2026-11-20): `--muted` is the column's own colour in Light (`--board-column` is stored as `{color.muted}`), so a muted block inside a column cannot be seen. The skeleton is `--skeleton` instead. The rest of §7 stands: static, card-sized blocks, no shimmer. **§10's row 26 superseded for one of its three reasons by [ADR-030](adr-030-toast.md)** (2026-11-20): after `removed_from_workspace` the app goes to `/workspaces`, because the workspace can no longer be opened. The other two reasons still go to the workspace. The same ADR settles the toast's own design and the access-revoked wordings, which *Out of scope* deferred; placement stays deferred. **§10's reconnecting fill superseded by [ADR-031](adr-031-presence-cluster.md)** (2026-11-20): a flat `--muted` face is 1.09:1 on the white app bar, so the reconnecting faces sit on `--avatar-neutral` instead. The label, its wording and colour, and the rule that no hue is spent, stand. **The ledger's Emerald row superseded by [ADR-032](adr-032-auth-form-errors-and-password-rules.md)** (2026-11-21): emerald now has one Phase-1 render site, a met rule in Register's password checklist (`--done`). It means *done*, as the row says. Its glyph's shape, a tick against a rose cross, carries the meaning, because emerald and rose are 37 / 32 apart for a deuteranope. State 36's field-invalid also gains a form-level counterpart there, the `Form message` above the button. **The ledger's Plum row and §7's row 21 superseded by [ADR-033](adr-033-workspace-screens.md)** (2026-11-21), and nothing else in them: plum also tints the empty-state illustrations, a decorative use the user chose against the recommendation, and the whole-screen empty states (flavours 1 to 4) carry an illustration as well as their sentence. The same ADR marks the current option of the Active / Archived switch with plum's existing *where you are*, so state 18 stays closed. **§8's rule 3 narrowed by [ADR-034](adr-034-side-menu.md)** (2026-11-21), for one case: from 1024 up, the whole board glides 240px with the sidebar when you hide or show it. No board object moves relative to another, and the press is your own, so it cannot be read as *a teammate did this*. Every other rule and value in §8 stands. **The ledger's Rose row and §7's row 22 narrowed by [ADR-035](adr-035-sidebar-loading-failure-and-switcher-mark.md)** (2026-11-21), for the sidebar's board list only: when its request fails, the sentence *Your boards didn’t load.* is in `--destructive-text`, above an uncoloured *Try again* row. This was the user's call over an uncoloured sentence. The page failures on 2.3 and 2.4 still spend no channel. **§7's zero channels on a failure narrowed by [ADR-038](adr-038-board-states-group-d.md)** (2026-11-23), for one thing: the drawings on the board's *load failed* and *gone* screens keep the illustrations' plum tint. That was the user's call over an ink-and-paper drawing. Their sentences and buttons still spend no channel. **§5's busy look and §10's row 38 narrowed by [ADR-040](adr-040-menus-and-the-confirmation.md)** (2026-11-23). The Button's third kind, Destructive, waits in pale rose (`--destructive-subtle`, with `--destructive-subtle-foreground` text and spinner) rather than `--primary-subtle`: the user's call over one busy look for every button. Row 38's rose confirm holds for the three deletes and *Leave board*. *Archive board* also asks first, but can be undone, so its confirm is Primary. Rose keeps its meaning, destructive; §5's other two rows stand.
- **Date:** 2026-11-18
- **Scope:** Every interaction state in the **app zone** (React + Vite SPA — [ADR-011](adr-011-frontend-framework.md)), which visual channel expresses it, and its measured value. Spends the budget [ADR-017](adr-017-ui-foundations.md) §*Channel budget* reserved, under the four constraints [ADR-019](adr-019-responsive-touch-scope.md) §*What ADR-018 inherits* added. Also settles the three things ADR-017 deferred here by name: **motion specifics** (durations, easing, `prefers-reduced-motion`), the **icon family**, and the values for drag-over and the drop placeholder. **Supersedes nothing.** ADR-017's baseline and ADR-019's responsive scope both stand unchanged; this ADR re-points one unspent channel and strikes one unreachable meaning, both recorded in *Corrections* below. Does **not** close SC 4.1.3 (announcing changes that happen without user action) — PRD §5.4 keeps that open, and §10 only notes where the two touch.

## Context

No frontend code exists. ADR-017 fixed how the app rests — surfaces, type, radius, elevation, colour — and deliberately spent no channel on interaction state, listing what it left unspent so this ADR could spend it deliberately. ADR-019 then moved phones and tablets into Phase 1, which added four constraints and voided one channel outright.

**The failure this ADR is shaped to avoid is recorded in ADR-017, and it was not arithmetic.** The draft ADR-017 replaced *"had already assigned every channel ADR-018 needs… leaving nothing for pending, read-only or server-corrected."* It spent the budget before counting what needed paying for. So §1 below enumerates every state and assigns nothing; the first channel is not spent until §2.

Four facts shaped what follows.

1. **Every board mutation is optimistic, and the server's answer comes back twice.** [ADR-012](adr-012-frontend-state-management.md) applies the change locally in `onMutate`, then the REST response arrives, then the broadcast arrives. The REST response for a move carries the authoritative `CardResponse` including `position` ([card_handler.go](../../../collabotask-backend/internal/delivery/http/handler/card_handler.go)), while the broadcast is best-effort and swallow-on-error ([broadcaster.go](../../../collabotask-backend/internal/usecase/common/broadcaster.go)) and may never arrive at all. **The REST response is therefore the only reliable confirmation**, and it is what ends a pending state.

2. **The server can correct a position, but the correction is invisible.** `PositionStep = 1000` and `PositionRebalanceThreshold = 0.5` ([positioning.go](../../../collabotask-backend/internal/domain/positioning.go)) — halving a gap below the threshold takes **eleven consecutive drops into the same shrinking gap**. And when a rebalance fires it rewrites the partition with `ROW_NUMBER() OVER (ORDER BY position ASC, id ASC)` ([positioning.go](../../../collabotask-backend/internal/repository/postgres/positioning.go)), which **preserves visible order**. Nothing moves on screen.

3. **shadcn's vendored defaults collide with ADR-017 in two places**, and ADR-017 §10's *"Everything else is inherited"* means both arrive silently on day one. `disabled:opacity-50` collides with the opacity channel; `hover:bg-accent` resolves to `#F5F5F4`, which is `--board-column`'s own value, so a card hovering to accent dissolves into its column.

4. **The app has no actor information.** ADR-017's Context already records that no board event names who caused it. §4's assignment is built so it never needs one.

Every ratio quoted is a computed WCAG 2.1 contrast ratio against the stated surface. PRD §5.4 declares **WCAG 2.2 Level AA**, so 4.5:1 binds body text, 3:1 binds large text and component boundaries (1.4.11), and **24×24 CSS px binds pointer targets at fine pointer** (2.5.8) — a floor ADR-019 §7 left open at that density.

---

## Decision

### 1. The inventory — 39 states, assigned nothing

Counted before anything was spent. Grouped by cause, not by channel.

| | **A — my own mutation's lifecycle** | | **F — connection health** |
|---|---|---|---|
| 1 | Pending-create — no server identity yet | 28 | Connecting |
| 2 | Pending-change — update / move / delete | 29 | Reconnecting |
| 3 | Server-corrected | 30 | Live |
| 4 | Rejected | | **G — other people** |
| | **B — drag** | 31 | Remote change arrived |
| 5 | Drag-armed (coarse only, cancellable) | 32 | Remote deletion of the open card |
| 6 | Drag-source | 33 | Presence change |
| 7 | Drag-overlay | | **H — entity flags** |
| 8 | Drop placeholder | 34 | Private board |
| 9 | Drag-over column | 35 | Archived board |
| 10 | Drag-rejected | | **I — input & forms** |
| 11 | Draggable-at-rest | 36 | Field invalid |
| | **C — control affordance** | 37 | Inline editing |
| 12 | Hover (fine only) | | **J — destructive intent** |
| 13 | Hover-equivalent at coarse | 38 | Destructive confirmation pending |
| 14 | Focus-visible | | **— added during enumeration** |
| 15 | Pressed | 39 | Session expired mid-session |
| 16 | Disabled | | |
| 17 | Busy | | **D — data arrival** |
| 18 | Selected | 19 | Cold load |
| | **E — permission** | 20 | Background refresh |
| 24 | Read-only board | 21 | Empty (five flavours) |
| 25 | Join-required gate | 22 | Load failed |
| 26 | Access revoked mid-session | 23 | Gone |
| 27 | Action forbidden | | |

Three of these are closed by inspection rather than by assignment:

- **18 — Selected does not exist in Phase 1.** There are no bulk operations in the API, and ADR-019 §6 routes `M` into the move picker rather than into a keyboard grab. Recorded as an explicit no, which is what releases 2px border weight in §5.
- **35 — Archived board is unreachable.** See *Corrections*.
- **30 — Live is invisible.** ADR-017 §3: colour is an event, not a texture.

### 2. The channel ledger

ADR-017 left seven channels unspent. Four are confirmed as written, one is overturned, one is demoted, one is struck.

| ADR-017 reserved | Outcome |
|---|---|
| **Opacity** → "not fully real": pending *and* drag-source as one meaning | **Confirmed and strengthened** (§3) |
| **Dashed border** → the drop placeholder | **Confirmed** (§4) |
| **Column panel tint** → drag-over column | **Confirmed** (§4) |
| **Icon** → lock (private) *and* archive (archived) | **Lock confirmed, archive struck** (§9) |
| **Cursor** → "what you can and can't do here" | **Demoted** — ADR-019 #1 voids it at coarse pointer, so it can never be the sole carrier of a meaning. It becomes a fine-pointer echo of something carried elsewhere and leaves the budget |
| **Border weight (2px)** → "a candidate for selection" | **Absorbed** — selection does not exist (§1), so 2px goes to the drop placeholder, where it helps a dashed line read |
| **Motion** → "something just changed": server-corrected | **Overturned** (§8) |

Newly spent here: **skeleton** (§7) and **border colour** (§6) — ADR-017 reserved border *weight* for selection, not border colour, so the three-step hover/pressed ladder is a new spend on an existing channel. Blur, italic and strikethrough remain unspent and unassigned.

### 3. In flight — the one meaning dimming carries

**A change that has not reached the server renders its object at 60% opacity, and that object is inert.**

`opacity` on the whole card, not on its contents. The meta row drops `--muted-foreground` for `--foreground` while in flight.

| | Light | Dark |
|---|---|---|
| Title over the composited fill | **4.70:1** | 6.18:1 |
| Meta row, promoted | **4.70:1** | 6.18:1 |
| On the drag-over tint | 5.11:1 | 6.11:1 |

**50% is not available.** The whole card is one compositing layer, so its text blends toward the column behind it, not toward the card fill. At 50% the title itself is **3.44:1** — the conventional `opacity-50` for "saving" is a 1.4.3 failure that looks perfectly fine on screen.

**The meta row is the binding constraint, not the title.** Muted grey under a fade clears 4.5:1 only at **80%**, where a composited fill of `#FDFDFD` against a `#F5F5F4` panel is not a signal at all. Hence the promotion: while a card is de-emphasised as a whole, its *internal* hierarchy has nothing left to express, so flattening it costs nothing and buys compliance.

**Frame-only translucency was measured and rejected.** Fading the fill and border while leaving text solid passes every ratio with room to spare, but its entire signal is a **4-unit-per-channel** shift in the fill and a **6-unit** shift on a 1px line. ADR-017 §3.4 rejected an argument on exactly this ground, calling a 3-unit difference *"below the just-noticeable difference for a flat field."* It is the draft's 1.02:1 tint in a different costume.

**Dim means inert.** A dimmed object cannot be dragged, opened, or operated through its menu. This is what makes the appearance honest, and it removes a real defect: two drags on the same card in quick succession issue two *absolute* position requests that can arrive out of order.

**Inertness requires a timeout, and there isn't one.** [frontend-architecture.md](../frontend-architecture.md) creates the axios client with `baseURL` and `withCredentials` and no `timeout`; axios defaults to zero, meaning none. A hung request would never fire `onError`, never restore the snapshot, and leave an inert card frozen until reload. **The axios instance sets `timeout: 10000`.** On timeout the mutation fails normally: snapshot restored, object live again, toast shown.

Opacity covers **1, 2, 6** — and covers them continuously. Pickup → drag-source → drop → pending → confirmed is one unbroken translucency, not two treatments handing off.

### 4. Drag

| State | Channel | Value |
|---|---|---|
| 5 drag-armed | **Shadow**, ramping | `--shadow-overlay`, built over the hold, with `scale(1.01)` |
| 6 drag-source | **Opacity** | §3 |
| 7 drag-overlay | **Shadow** | `--shadow-dragged`, `scale(1.02)` |
| 8 drop placeholder | **Dashed border, 2px** | `--drop-placeholder` |
| 9 drag-over column | **Column body tint** | `--drag-over` |
| 10 drag-rejected | — | Instant return to rest; no channel (§8) |
| 11 draggable-at-rest | **Cursor** at fine; nothing at coarse | `grab` / `grabbing` |

**The hold ramps rather than snapping.** ADR-019 §4 assigns shadow to the arming moment but does not say whether it appears during the hold or at its end. It builds across the hold, because a long-press is the one gesture on a phone that nothing else hints at — a fifth of a second of total silence reads as "my touch didn't register", and people lift off and retry. A drift beyond tolerance retracts it quickly and quietly as the board scrolls. The ramp is tied to the hold duration, which ADR-019 §10 tunes on device; it is deliberately **not** a token here, or tuning the gesture would silently retime the board.

**State 11 has no coarse-pointer treatment, and that is accepted rather than missed.** ADR-019 §4 makes the whole card the drag target with no grip handle, so on a touchscreen there is nothing to mark — a card's draggability is discovered by pressing it. This is the one place cursor is a sole carrier rather than the fine-pointer echo §2 demotes it to, and it is tolerable because the meaning is "everything here is draggable", which needs no per-object signal.

**Drag-over, and why the header stays out of it.**

| | Light | Dark |
|---|---|---|
| `--drag-over` | **`#D6D3D1`** (stone-300) | **`#3B3633`** |
| vs `--board-column` | **1.37:1** | **1.47:1** |
| vs `--board-canvas` | 1.19:1 — *darker than the canvas* | 1.66:1 |
| `--muted-foreground` on it | 5.12:1 | 8.00:1 |
| Card fill on it | 1.49:1 | 1.27:1 |

The draft's `#FEF3C7` measured **1.02:1** — a pale yellow almost exactly as bright as the grey it replaced. The floor this ADR states is **ADR-017's own canvas→column step, 1.15:1**: if that is enough to make columns read as separate panels permanently, it is the minimum for a transient signal. 1.37:1 clears it by 19%.

The real percept is stronger than the ratio: the tinted column becomes **the only surface on the board darker than the canvas behind it**, which the eye catches peripherally while following a card.

**The tint covers the column body only; the header keeps `--board-column`.** ADR-017 §2 makes the header's `border-bottom` load-bearing — columns are unbounded, so cards scroll *under* it. That rule is `--border #E7E5E4`, which on either tint sits at about **1.19:1 and vanishes**, exactly when the header most needs to hold. Leaving the header untinted also adds an internal step that makes the tint read more strongly, and is more precise: the header was never a drop target.

Dark inverts the card/panel brightness order during a drag — cards become darker than the panel they sit on. ADR-017 §4's requirement in dark is that the steps must not be **flattened**, and 1.27:1 exceeds the resting 1.15:1, so separation improves rather than degrades.

**The drop placeholder, and the dependency ADR-017 missed.**

| | Light | Dark |
|---|---|---|
| `--drop-placeholder` | **`#57534E`** (stone-600) | **`#A8A29E`** (stone-400) |
| on `--drag-over` | **5.12:1** | **4.73:1** |

The draft's `#D6D3D1` measured 1.37:1 on the bare panel and **1.00:1** on this tint — literally the same colour. But the finding that matters is this: **the placeholder only ever appears inside a column you are already over**, so the surface it must contrast against is the tint, never the plain panel. ADR-017 stated the 3:1 requirement without noticing it had already darkened that surface. Measured against it, **ADR-017's own `--input` (stone-500 `#78716C`) thins to 3.22:1**, down from 4.80:1 on white. It still clears 1.4.11, but by 0.22, and §3.4 chose it as the lightest compliant stone against a *white* field. A drop placeholder is not a boundary identifying a control — it is read peripherally, while attention follows the card being moved — so stone-600 at 5.12:1 is chosen for that job rather than for the floor.

2px rather than 1px: a dashed line at 1px loses most of its ink to the gaps. The weight is available because §1 closed selection.

Per ADR-019, all of this applies to the **44px peek** of the adjacent column, which is a legal drop target below 640px.

### 5. The three states that were all called "disabled"

shadcn ships `disabled:opacity-50`. Unoverridden, a disabled Confirm button and a card mid-save are the same picture, and the opacity channel carries two meanings on day one. **Nothing in this app is "disabled".**

| Situation | Treatment | Light | Dark |
|---|---|---|---|
| **You are not allowed** | The control is **absent**, plus a persistent banner | — | — |
| **In flight** (17) | `--primary-subtle` fill, `--primary-subtle-foreground` label, spinner, label changes to the progressive verb | 7.70:1 | 7.67:1 |
| **The form is not satisfied yet** (16) | `--muted` fill, `--muted-foreground` label, `--border` edge | 6.99:1 | 10.18:1 |

The first is ADR-017's read-only resolution — *"removing controls plus a banner — which costs zero channels"* — generalised. The second says *wait*, not *no*, so it keeps plum's shape and stays recognisably the live action. The third cannot be removed, since an unsatisfied Confirm is the entire point of a confirmation dialog, so it is drawn in tokens rather than in transparency: visibly inert, unmistakably still a button.

**The scaffold overrides shadcn's `disabled:` variant once, in the vendored button, before any other component is added.** ADR-017's Context is the reason: once twenty components reference a default, changing it is a sweep.

### 6. Affordance — hover, pressed, focus, and the coarse-pointer rule

**Hover cannot use a fill.** shadcn hovers to `--accent`, which ADR-017 §9 sets to `#F5F5F4` — `--board-column`'s own value. A card hovering to accent becomes the colour of the panel behind it and dissolves. The edge does the work instead, one palette step per state:

| | Light | vs `--card` | Dark | vs `--card` |
|---|---|---|---|---|
| Rest | `#E7E5E4` stone-200 | 1.26:1 | `#44403C` stone-700 | 1.48:1 |
| **Hover** (12) | **`#D6D3D1`** stone-300 | **1.49:1** | **`#57534E`** stone-600 | **1.99:1** |
| **Pressed** (15) | **`#A8A29E`** stone-400 | **2.52:1** | **`#78716C`** stone-500 | **3.16:1** |

Three consecutive stone steps in each mode. Pressed is also the only feedback a finger receives before the hold completes.

**Focus-visible (14)** is already spent: ADR-017 gives `--ring` the same value as `--primary` precisely so it clears 3:1 as a boundary. A 2px ring at `--ring` with a 2px offset, on every focusable thing, never removed.

**The coarse-pointer rule ADR-019 #2 asked for:**

> **A function must be reachable without hovering. Hover may make a control quicker to reach; it may never be the only way to reach it.**

This permits the hover-revealed `⋯` at fine pointer, because ADR-019 §6 already gives Move three entry points, of which two need no hover. It forbids a control whose only route is a hover.

**Minimum target size (SC 2.5.8).** Every pointer target has a hit area of at least **24×24px at fine pointer** and 44×44 at coarse (ADR-019 §7). The hit area may exceed the glyph: the `⋯` is a 13px mark inside a 24px box, and the title reserves 24px of right padding at fine pointer to hold it.

**Two things are declared explicitly not targets:** the 20px card avatar and the 20px presence avatar. They are identity markers, not controls. The presence *cluster* is one target with 2px of vertical padding to reach 24px; individual circles within it do nothing. This closes the gap PRD §5.4 recorded as *"to be checked if any of them become clickable"* without touching ADR-017's density.

The smallest possible card — one line, no meta — is 36px tall, so cards clear the floor by construction, as they do at coarse (ADR-019 §7).

Reserving 24px of title width at fine pointer leaves **216px** of the 240px content box. This nudges the representative card mix, which ADR-019 §7 already records as the one number in its table resting on an assumption rather than a measurement; it is re-measured against real content, not re-derived here.

### 7. Loading, empty, gone

| State | Treatment |
|---|---|
| 19 cold load | **Static skeleton** — blocks at the card and column geometry, `--muted` fill, **no shimmer** |
| 20 background refresh | **Nothing.** Data on screen is correct; the post-reconnect catch-up is covered by §8's presence marker |
| 21 empty | Per-surface copy in `--muted-foreground`. Zero channels |
| 22 load failed | Message plus a retry control. Zero channels |
| 23 gone | A whole-screen state. Zero channels |

**The skeleton does not shimmer, and that is forced rather than chosen.** §8 makes movement mean "a teammate did this". A shimmer is movement in the board's content area with no teammate behind it, so it would be the channel lying on first paint. A skeleton is also the right shape here rather than a spinner: the board's geometry is known in advance, and a spinner on a grey canvas says nothing about what is arriving.

> **The `--muted` fill in state 19's row cannot be seen where the blocks sit.** In Light, `--board-column` is `--muted` itself, so a muted block inside a column measures 1.00:1. **[ADR-029](adr-029-skeleton-block.md)** (2026-11-20) gives the skeleton its own colour, `--skeleton` (`#E7E5E4` / `#292524`, 1.15:1 on the column in both modes), and a height, `--skeleton-height` (60 / 72). It also reads *"the card and column geometry"* as card-sized blocks inside real column panels, since one colour cannot draw both sizes nested. Everything else above stands. One consequence for the code: shadcn's Skeleton ships `animate-pulse`, and a pulse is movement with no teammate behind it, so it is removed in the same pass as the `disabled:` override.
>
> **Row 21 has a second part on whole-screen empty states.** [ADR-033](adr-033-workspace-screens.md) (2026-11-21) puts an illustration above the sentence on flavours 1 to 4: the line in `--illustration-ink` and one or two shapes in `--illustration-tint`. Flavours 5 and 6 keep their one line. Row 22 is unchanged: a load failure is a sentence and an Outline *Try again*.
>
> **Row 22 spends rose in one place.** [ADR-035](adr-035-sidebar-loading-failure-and-switcher-mark.md) (2026-11-21): when the sidebar's board list fails to load, its sentence is in `--destructive-text`, and the retry is a sidebar row with a refresh glyph, uncoloured. The user chose this over an uncoloured sentence after the cost was shown. On 2.4 the page's own failure beside it stays uncoloured, so the sidebar is louder than the page. Every other load failure keeps zero channels. The same ADR keeps state 19 static in the sidebar: row-sized blocks, no pulse.

### 8. Motion

**Movement of a board object means: *this changed because someone else acted*.**

ADR-017 reserved motion for server-correction. Context fact 2 shows that state needs eleven consecutive drops into one shrinking gap to fire and **preserves visible order when it does** — the channel had no render site. Meanwhile a teammate moving a card fires constantly and is the literal promise of PRD §4.2: *"A teammate drags a card to Done → it moves on your screen too, right away."*

**It needs no actor field, which is why it is buildable.** Movement is triggered by an object's position actually changing, not by a frame arriving. The broadcast for your own move is a no-op — the card is already there — so nothing moves and nothing animates. The only addition is that the drag handler renders its own drop without a transition, and that handler already knows it is dragging. Server-correction folds in as a degenerate case: if a card ever does land somewhere other than where you dropped it, it moved, so it animates.

| | Duration | Easing |
|---|---|---|
| Chrome state change — hover, pressed, border, fill | **120ms** | `linear` |
| **Board object moving** — the channel | **200ms** | `cubic-bezier(0.2, 0, 0, 1)` |
| Board object appearing / vanishing | **200ms** | arrive in; `cubic-bezier(0.4, 0, 1, 1)` out |
| Container — card drawer, phone sheet, toast | **260ms** | arrive / leave as above |
| Reduced-motion cross-fade | **110ms** | `linear` |

> **The sidebar is a container, and from 1024 up the board glides with it.** [ADR-034](adr-034-side-menu.md) (2026-11-21): the sidebar sliding over the content below 1024, and hiding or showing beside it from 1024 up, takes the container row's 260ms and eases, and cross-fades at 110ms under reduced motion. From 1024 up the board slides with it, so every card on screen moves 240px. **Rule 3 below is narrowed for this one case.** It guards *a teammate did this*, and the whole board shifting together right after your own press can't be read that way: no board object moves relative to another. The recommendation was a jump, the board taking its new width in one step, which kept rule 3 as written. The user chose the glide after its cost was shown (ADR-034, the note under §10.10). Under reduced motion the board moves in one step.

Things arriving decelerate; things leaving accelerate. One movement duration, so the app has one tempo.

Three rules follow from the meaning rather than from taste:

1. **A rejected move (4) is not animated.** ADR-012's own word is "snaps back", and it is correct: a refused move easing back into place would be indistinguishable from a teammate putting it there. Instant return, plus the toast, plus the card losing its dim — three signals at once, and the channel keeps telling the truth.
2. **Animate only when the object is visible before *and* after.** Columns are unbounded, so a card can travel thousands of pixels. That is not a movement anyone can follow; it is a vertical streak. Any transition crossing the visible bound cuts instead.
3. **Chrome transitions are not this channel.** A hover fade or a sheet sliding in is the mechanics of a control, not a claim about who acted. The boundary that keeps that honest: **chrome never displaces a board object.** This is also what rules out the shimmering skeleton in §7.

**`prefers-reduced-motion: reduce` replaces movement with a cross-fade; it does not remove the signal.**

A board object that would move instead fades out of its old position and in at its new one over 110ms. Nothing translates, so the vestibular trigger is gone, but the temporal cue survives — which matters here more than usual, because movement is the *only* thing explaining why the board changed. Removing it outright would take that explanation away from precisely the people least able to absorb a confusing board. Under the same query: the arming ramp keeps its shadow and drops its `scale`; the busy control drops its spinner and keeps its label change; containers cross-fade rather than slide.

This fallback and the open SC 4.1.3 gap want the same thing — a way of being told the board changed that does not depend on watching it. Whoever closes 4.1.3 should check both. Closing it is out of scope here.

### 9. Icons

**Phosphor (`@phosphor-icons/react` v2), Regular weight at 16px.**

Phosphor's weights are filled paths on a shared 256×256 grid, so line weight is baked into each weight and all weights are layout-identical. Read off the geometry:

| Weight | Units | At 16px | At 20px |
|---|---|---|---|
| Thin | 8 | 0.50px | 0.62px |
| Light | 12 | 0.75px | 0.94px |
| **Regular** | **16** | **1.00px** | 1.25px |
| Bold | 24 | 1.50px | 1.88px |

**Regular at 16px renders at exactly 1.00px**, matching ADR-017's 1px card border and column-header rule. Lucide, the alternative, would have rendered 1.33px.

**Regular is furniture. Fill is the channel.**

- **Regular** draws everything that labels an action — close, plus, chevron, `⋯`, calendar, trash. These are not state signals and spend nothing.
- **Fill** is permitted on exactly one icon: the **lock** that marks a private board (34), which is the icon channel's single Phase-1 meaning. A padlock reads better solid at 16px, and making fill the marker of the channel itself means a reviewer can tell furniture from state at a glance.
- Thin, Light, Bold and Duotone carry **no meaning**. Assigning one is a new channel and needs its own ADR.

Three implementation facts, because getting them wrong costs an afternoon:

- The package is `@phosphor-icons/react` **v2**. The older `phosphor-react` is a different, superseded library.
- v2 exports carry an **`Icon` suffix** — `LockIcon`, not `Lock` — and fill is a **prop**: `<LockIcon weight="fill" />`. There is no `LockFill` export.
- **This is not shadcn's default.** shadcn's components import `lucide-react`, so each component shipping an icon needs its import rewritten as it is vendored in — Dialog, Sheet, Select, DropdownMenu, Checkbox, Toast, Calendar, Command, Accordion. Roughly ten one-line edits, each done once, recurring only when a new component is added. The package carries around 9,000 icon components, so **confirm the import path at scaffold time**: pulling them through the main barrel can make Vite's dev server crawl. This is the same shape of warning as ADR-017 §9's Tailwind-version note.

### 10. Connection, permission, and the states that cost nothing

**Connection health (28, 29, 30) is marked in grey, on the presence cluster.**

When the socket drops, your work still saves — mutations go over REST; the socket only receives. What stops is **other people's changes**, and the presence cluster freezes showing whoever happened to be there. ADR-017 treats that cluster as a foundation precisely because no event names an actor, so when it goes stale it becomes the most confidently wrong thing on screen.

So the marker goes on the row that is lying: the avatars drop their identity hues for a flat `--muted` fill, and a `--muted-foreground` "Reconnecting…" label appears beside them. **No hue is spent.** Every attention colour is committed — plum to the one action, amber to deadlines only, rose to overdue and destructive — and the honest message is "you may be behind", not "something is broken". A red bar on a task board reads as *my work did not save*, which would be false. Placement carries it instead of colour.

Reconnection retries forever with capped backoff, so there is no terminal failure state to draw. **Connecting (28)** is part of cold load. **Live (30)** has no treatment.

> **The grey has to be seen.** `--muted` is 1.09:1 on the white app bar, under the faintest identity circle (1.11), and in an overlapping stack the face in front cuts off the initials behind at an edge nobody can see. **[ADR-031](adr-031-presence-cluster.md)** (2026-11-20) puts the reconnecting faces on `--avatar-neutral` (`#E7E5E4` / `#57534E`), matched to the faintest live circle in each theme, so the cluster changes colour without changing weight. It also places the label: after the `+N` seat from 640px up, and in the seat's place below 640px.

| State | Treatment | Channel cost |
|---|---|---|
| 24 read-only board | Mutating controls absent + a **persistent, non-dismissible** banner | zero |
| 25 join-required gate | A whole-screen gate with the Join action | zero |
| 26 access revoked | Navigate to the workspace + a toast naming the reason | zero |
| 27 action forbidden | Toast; the optimistic change reverts via §3 | zero |
| 32 remote deletion of the open card | The drawer closes with a toast | zero |
| 36 field invalid | `--destructive-text` message + `--destructive` field edge | already spent |
| 37 inline editing | The `--input` boundary ADR-017 §3.4 made deliberately heavy | already spent |
| 38 destructive confirmation | Rose confirm button | already spent |
| 39 session expired | A whole-screen state before the redirect to `/login` | zero |

> **Row 26's destination does not exist for one of its reasons.** After `removed_from_workspace` you can no longer open the workspace, so navigating there would end in a load failure (22). **[ADR-030](adr-030-toast.md)** §6 (2026-11-20) sends that reason to `/workspaces` instead. The other two reasons still go to the workspace, and the toast still names the reason. ADR-030 §5 writes the three wordings. Every row's channel cost stands: the toast is inverted, which marks the container and not a state.

**24 is persistent by inheritance, not by preference.** ADR-019 #3: on touch the banner is the only remaining signal, since a read-only user presses and holds and simply nothing happens. Per ADR-019's Context fact 3 this is routine browsing, not an edge case.

**39 was missed in the first pass and is worth naming.** [frontend-architecture.md](../frontend-architecture.md)'s axios interceptor hard-navigates to `/login` on any non-profile 401. Without a treatment, a 7-day JWT expiring mid-session looks exactly like a crash.

### 11. Token additions

```css
:root {
  /* state vocabulary — ADR-018 */
  --drag-over: #D6D3D1;            /* 1.37:1 on --board-column; body only, not the header */
  --drop-placeholder: #57534E;     /* 5.12:1 on --drag-over — not --input, which thins to 3.22:1 there */
  --border-hover: #D6D3D1;         /* 1.49:1 on --card */
  --border-pressed: #A8A29E;       /* 2.52:1 on --card */

  --in-flight-opacity: 0.6;        /* 4.70:1 title and promoted meta; 50% fails at 3.44:1 */
  --target-min: 24px;              /* SC 2.5.8 at fine pointer; 44px at coarse (ADR-019 §7) */

  --motion-chrome: 120ms;
  --motion-object: 200ms;
  --motion-container: 260ms;
  --motion-reduced: 110ms;
  --ease-arrive: cubic-bezier(0.2, 0, 0, 1);
  --ease-leave:  cubic-bezier(0.4, 0, 1, 1);
}

.dark {
  --drag-over: #3B3633;            /* 1.47:1 on --board-column */
  --drop-placeholder: #A8A29E;     /* 4.73:1 on --drag-over */
  --border-hover: #57534E;         /* 1.99:1 on --card */
  --border-pressed: #78716C;       /* 3.16:1 on --card */
}
```

Busy and not-satisfied add **no** tokens — they are built from `--primary-subtle` / `--primary-subtle-foreground` and `--muted` / `--muted-foreground`, which ADR-017 §9 already defines. The additions needing exposure as Tailwind utilities are `--drag-over`, `--drop-placeholder`, `--border-hover` and `--border-pressed`.

The axios client gains `timeout: 10000` (§3). The vendored shadcn button's `disabled:` variant is overridden at scaffold (§5).

### 12. Figma — two densities without doubling the frames

**The state vocabulary is density-independent.** Every value above is a colour, an opacity, a stroke weight or a duration. Two rows of ADR-019 §7's table carried only a coarse value, and §6 fills in both fine values — **minimum target 24px** and **title right padding 24px**. Neither is a new variable; the slots already existed. **ADR-018 adds no new row to the density scale.**

- **A second variable collection with two modes, Fine / Coarse**, separate from ADR-017 §7's colour collection. Theme and density are independent, so two two-mode collections give all four combinations; one four-mode collection would force every colour to be written twice over and every size twice over.
- It holds ADR-019 §7's seven numbers: card padding, avatar size, meta-row height, card gap, column-header height, title right-padding, minimum target.
- **Every state is a component variant, not a frame.** One card component carrying rest / hover / pressed / in-flight / drag-source resolves to whichever mode its parent frame is set to.
- **Auto-layout throughout, and no typed-in numbers.** One hardcoded 8px padding and the mode flip silently stops being a check.
- Whole-screen frames only where the whole screen differs: read-only, join-required, cold-load skeleton, empty board, load failed, gone, session expired, plus one drag frame (a drag is four states at once).

**Fine frames, plus exactly one 390×844 phone board.** Density is a variable; layout is a breakpoint. No mode can turn a fixed 280px column into ADR-019 §3's `min(100vw − 72px, 420px)` snap strip, so that needs a frame regardless — and it is the only place ADR-019 §7's "6 cards visible at 390×844" is checked by eye, which ADR-019 itself flags as the one number in its table resting on an assumption.

---

## Channel budget — the ledger after this ADR

| Channel | Meaning | Spent by |
|---|---|---|
| Plum | the one action, and where you are — including the focus ring | ADR-017 |
| Amber | deadline approaching — deadlines only | ADR-017 |
| Rose | overdue, destructive intent, invalid input | ADR-017 |
| Emerald | done / archived — **no Phase-1 render site** | ADR-017 |
| Eight identity hues | who, and which board | ADR-017 |
| Surface brightness (3 steps) | permanent hierarchy | ADR-017 |
| Shadow | temporarily lifted — including the armed long-press | ADR-017 / ADR-019 |
| Border | an edge — and, by colour step, hover and pressed | ADR-017 / **ADR-018** |
| Radius (3 steps) | object size class | ADR-017 |
| Type scale (5 steps) | reading order | ADR-017 |
| Weight (400/500/600) | emphasis within a block | ADR-017 |
| **Opacity** | not fully real — pending and drag-source, and inert | **ADR-018** |
| **Motion** | a teammate did this | **ADR-018** |
| **Dashed border (2px)** | where it will land | **ADR-018** |
| **Column body tint** | the column you are over | **ADR-018** |
| **Icon fill** | this icon is a state, not an action | **ADR-018** |
| **Skeleton** | the shape of what is loading | **ADR-018** |

> **Emerald has a render site now.** [ADR-032](adr-032-auth-form-errors-and-password-rules.md) (2026-11-21) colours a met password rule with `--done`, the meaning this row already gives it. The rule's glyph (a tick, against a rose cross for a missed rule) carries the meaning on its own: emerald and rose are the red–green pair.

> **Plum gains a decorative use, and spends its *where you are* once more.** [ADR-033](adr-033-workspace-screens.md) (2026-11-21) tints the empty-state illustrations with `--illustration-tint` (`#F4F0FA` / `#3B2F57`), the same values as `--primary-subtle` held as literals. The recommendation was a grey; the user chose plum after the cost was shown once. The same ADR marks the current option of the Active / Archived switch with plum text and a 1px plum edge. That is this row's *where you are*, not state 18, which stays closed.

> **Plum spends its *where you are* once more.** [ADR-034](adr-034-side-menu.md) (2026-11-21) marks the sidebar's current row (the open board, or *All boards* on 2.4) with plum text and a 1px plum edge, the Active / Archived switch's look. The edge carries the mark as a shape as well, because plum against normal text is only 2.67 / 2.74. State 18 stays closed. The same ADR's `--scrim`, which dims the content under the slide-over sidebar, is not the Opacity row: it covers the whole screen at once, never one object, so it cannot be read as one thing being "not fully real".

> **Rose gains a meaning in one place, and plum marks one more *where you are*.** [ADR-035](adr-035-sidebar-loading-failure-and-switcher-mark.md) (2026-11-21):
> - **Rose:** in the sidebar, rose also says *your boards didn't load*. Elsewhere it still means overdue, destructive intent and invalid input. It is on the sentence, not the retry, so the control does not read as destructive.
> - **Plum:** a still 8px dot marks the selected workspace in the open switcher. That is this row's *where you are*, so no channel moves.
> - **Motion:** the note asking for that dot asked for a pulse. The pulse was not offered, because Motion means *a teammate did this*.

**Still unspent:** blur, italic, strikethrough. **Left the budget:** cursor, which is now a fine-pointer echo and never a carrier.

Every one of the 39 states is covered, with one stated gap: **11 has no coarse-pointer treatment** (§4). No channel carries two meanings, with one declared exception: **Border** marks an edge permanently and, by colour step, hover and pressed (§6) — the two cannot be confused, because the step is a change *to* the edge rather than a second thing drawn beside it.

---

## Corrections to ADR-017

ADR-017 is **not** superseded. Two of its entries are corrected, both for the same reason it corrected the draft it replaced — a value or a meaning with nothing to attach to.

| Entry | Finding | Resolution |
|---|---|---|
| **Motion** → server-corrected | Needs 11 consecutive drops into one gap, and preserves visible order when it fires. **Zero render sites.** | Re-pointed to "a teammate did this" (§8) |
| **Icon** → lock *and* archive | Archived boards are filtered out of the list (`AND b.is_archived = FALSE`, [board_queries.go](../../../collabotask-backend/internal/repository/postgres/board_queries.go)) and every access path 404s them ([board_access_check.go](../../../collabotask-backend/internal/usecase/common/board_access_check.go)). **Unreachable in Phase 1**, and so is Unarchive | Struck; the icon channel carries one meaning (§9) |

Three further observations, recorded rather than acted on:

- **`--done` / emerald has no Phase-1 render site either.** Cards have no status field, archived is unreachable, and a column titled "Done" is free text. The hue is defined and audited at 5.48:1 but colours nothing. Left as ADR-017 wrote it.
  > *2026-11-21:* no longer true. [ADR-032](adr-032-auth-form-errors-and-password-rules.md) gives it one render site, a met rule in Register's password checklist.
- **ADR-017's `--input` is too thin for the drop placeholder** at 3.22:1 on the tint, down from 4.80:1 on white. It clears 1.4.11 by 0.22, so this is a margin finding rather than a failure. Not an error in §3.4 — the surface underneath simply changed. §4 uses stone-600 at 5.12:1, because a placeholder read peripherally needs more than the floor.
- **ADR-012 under-uses the REST response body.** Its comment says the broadcast *"confirms the authoritative position"*, but the move endpoint already returns it, and the broadcast is droppable. Nothing contradicts ADR-012; §3 states the reading it implies.

---

## Options Considered

### Movement meaning "something changed", from any source

One code path, one animation, nothing to keep straight. Rejected because on a board that changes constantly, "a thing happened" is almost always true and therefore carries almost nothing — the same failure ADR-017 diagnosed in the draft's drag-over tint, present in the arithmetic and useless on screen. It also animates your own optimistic change *before* the server has agreed to it, so a rejected move eases smoothly in and then smoothly back out: two pieces of movement for an action that did not happen.

### Keeping movement on server-correction, or releasing it

These produce **pixel-identical boards** — nothing animates either way. The only difference is whether the ADR admits it. Both were rejected because a card silently appearing in a different column is the main source of "wait, did I do that?", and PRD §4.2 promises the opposite.

### Frame-only translucency for the in-flight state

Measured and genuinely tempting: every ratio passes with enormous headroom because the text never fades. Rejected on ADR-017's own just-noticeable-difference standard — a 4-unit fill shift and a 6-unit shift on a 1px line. It is correct arithmetic describing something nobody would see.

### Letting dimming mean both "disabled" and "in flight"

The zero-work option: keep shadcn's default and rely on a button and a card being obviously different objects. Rejected because they are not always different objects — a card's own menu button, dimmed, is genuinely ambiguous between "forbidden" and "still saving" — and because one channel carrying two meanings is the precise mechanism ADR-017 says sank the draft.

### Rose or amber for the dropped connection

Rose would be unmissable, but on a task board a red bar reads as *my work did not save*, which is false — mutations still succeed over REST. It also widens a colour ADR-017 deliberately kept narrow. Amber fits the "attention, time-sensitive" meaning, but ADR-017 bolds **deadlines only** twice, and unlike motion, amber has real things it colours today: a card due tomorrow would share a colour with a network blip.

### A stronger drag-over tint, with darker muted text to compensate

Would buy the loudest possible signal. Rejected because it reopens `--muted-foreground`, which ADR-017 §3.4 derived carefully across four surfaces in both modes, to gain 8% on one transient state.

### Lucide

shadcn's default, so it needs no import rewrites at all — a real advantage, and the one argument that nearly carried it. Rejected on rendering: at 16px its 24-grid, 2-unit stroke lands at 1.33px against ADR-017's 1px borders, where Phosphor Regular lands at exactly 1.00px. Lucide also has no designed filled weight — fills are officially unsupported and work only incidentally on some icons — so the Regular/Fill distinction of §9 would not have been available.

### A separate cue for the arming window

Clearest possible separation between "arming" and "armed", but it spends a channel on a 220ms window, and ADR-019 §4 explicitly says that cue should cost nothing.

### Instant change under `prefers-reduced-motion`

The conventional reading, and the simplest build. Rejected because movement is the only carrier for "a teammate did this": removing it hands the unexplained-board problem specifically to people using an accessibility setting. The cross-fade honours the setting — nothing translates — while keeping the meaning.

### Drawing both densities in Figma

Rejected by §12's finding. With the state values identical at both densities, most of a second frame set would differ only in padding, which is what a mode flip is for.

---

## Out of scope (deferred, recorded)

- **SC 4.1.3 — announcing changes that happen without user action.** PRD §5.4 keeps it open. §8 notes that the reduced-motion fallback and this gap want the same non-visual signal, and that both should be checked by whoever closes it. Not solved here.
- **Toast content and placement.** Several states deliver their message by toast; the component's own design is scaffolding work, not a channel decision.

  > **The design and part of the content are now settled; placement is not.** **[ADR-030](adr-030-toast.md)** (2026-11-20) draws the toast as one inverted appearance with a close button, and writes the three access-revoked wordings. Where it sits on screen, and the wordings for 4, 27 and 32, are still deferred.
- **Empty-state copy.** Five surfaces need it; writing it is not a vocabulary decision.
- **The Astro landing page.** Unbound by this ADR, as it is by ADR-017 §3.
- **Column width at fine pointer.** 280px remains ungrilled, as ADR-017 and ADR-019 both record.
- **The representative card mix.** §6 narrows the fine-pointer title box by 24px, which nudges it. ADR-019 §7 already records re-measuring against real content.
- **`--done` / emerald.** Defined, audited and unused. Left alone rather than re-pointed; a Phase-2 feature may claim it.
- **Phase-2 deferrals** — comments, labels, attachments, public boards, refresh tokens, account deletion — untouched.

## Downstream edits this ADR requires

1. **PRD §5.4** — the SC 2.5.8 gap **closes**. §6 sets a 24×24 floor at fine pointer and declares both 20px avatars non-targets, which is exactly the check the gap row asked for. SC 4.1.3 stays open.
2. **`CLAUDE.md` › `## Now`** — ADR-018 is written; the next step is scaffolding, Vite app zone first.
3. **`docs/build-log.md`** — an entry for the state vocabulary.
4. **SRS** — no change. This is frontend presentation end to end; the SRS is a backend contract, which is the same conclusion ADR-019 reached.
5. **`CONTEXT.md`** — no change. "In flight", "drag-over", "placeholder" and "skeleton" are interface vocabulary, not domain terms, and nothing in Workspace / Board / Column / Card / Position / the role model moves. Worth confirming rather than adding: **a dimmed card is not a domain state.** There is no `pending` column anywhere; it is a client-side window between `onMutate` and the response, invisible to the backend and to every other client.
6. **Scaffold-time checklist**, carried from the ADR body so it is not rediscovered: override shadcn's `disabled:opacity-50`; set `timeout: 10000` on the axios client; confirm Phosphor's import path against Vite dev-server startup; confirm Phosphor is still not shadcn's default before rewriting imports; and confirm the Tailwind major version, which ADR-017 §9 and ADR-019 §1 both already ask for.
