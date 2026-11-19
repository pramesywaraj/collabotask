# Design File Inventory — What To Draw in Penpot

> **What this doc is:** the authoring checklist for the Penpot file `Collabotask App` — every screen, overlay, whole-screen board and variant set the app zone needs, and what sits on each one.
> **What this doc is *not*:** a decision record. It **decides nothing.** Every row is derived from an accepted ADR, a user story or the route tree, and cites its source. Where no document settles a question, the row is marked **[unsettled]** and listed again in §6.
>
> **Status:** Living document. **Owner:** pramesywaraj. **Date:** 2026-11-18.
> **Audience:** whoever authors the design file — today that's the solo dev; written so a future teammate can pick it up without re-reading four ADRs.

---

## 0. Scope and sources

The **app zone** only — the React + Vite SPA ([ADR-011](adr/adr-011-frontend-framework.md)). The Astro landing page is explicitly outside this file: [ADR-020](adr/adr-020-design-tool-penpot.md) excludes it, and [ADR-017](adr/adr-017-ui-foundations.md) §3 leaves it free to carry more colour than a dense board. It needs its own, later effort.

Four documents fix everything here; this one only arranges it.

| Source | Fixes |
|---|---|
| [ADR-017](adr/adr-017-ui-foundations.md) | Tokens and values, fine density, contained columns, type scale, radius, elevation, colour modes |
| [ADR-019](adr/adr-019-responsive-touch-scope.md) | Two layouts (crossover 640px), touch drag, the coarse density scale |
| [ADR-018](adr/adr-018-state-vocabulary.md) | All 39 interaction states and the channel each one spends |
| [ADR-020](adr/adr-020-design-tool-penpot.md) | Penpot as the medium; token sets, theme groups, variants, which boards exist |

Screen boundaries come from the route tree in [frontend-architecture.md](frontend-architecture.md) §*Route tree*; screen *contents* come from [001-user-stories.md](../product/001-user-stories.md).

**The density rule, because it governs the whole file.** Fine and Coarse are a *theme group*, not a second set of frames ([ADR-020](adr/adr-020-design-tool-penpot.md) §2). Draw every board once at Fine and let the flip carry it to Coarse. Layout is the exception: no mode flip turns a fixed 280px column into a viewport-relative snap strip, so the phone board is drawn separately (§2.15).

---

## 1. The six screens

Six routes, one of which is an overlay over another. Every state named in a row is an [ADR-018](adr/adr-018-state-vocabulary.md) state number.

### 1.1 `/login`

| On it | Source |
|---|---|
| Heading at 28/36 — login and register own the only use of that step | ADR-017 §6 |
| Email and password fields with the `--input` stone-500 `#78716C` boundary, deliberately heavier than most shipped systems | ADR-017 §3.4 |
| Plum primary submit; inline plum link across to register | ADR-017 §3.1 |
| Variants: field invalid (36), busy (17 — `--primary-subtle` fill, spinner, label becomes the progressive verb), form not satisfied (16 — `--muted` fill), focus-visible (14 — 2px `--ring`, 2px offset) | ADR-018 §5, §6 |

No separate narrow board. A two-field form reflows at any width, and the density flip supplies the 44px targets.

### 1.2 `/register`

Same furniture plus a name field. One addition worth drawing: a **duplicate-email conflict renders as field-invalid, not as a toast** — it points at a field, so it belongs on the field.

### 1.3 `/workspaces`

| On it | Source |
|---|---|
| App bar, 56px — product mark, user menu | ADR-017 §1 derivation |
| Workspace tiles: name, **my role**, member count, active board count | US-05 |
| "Create workspace" primary action | US-03 |
| Variants: cold load (19 — static skeleton, no shimmer), empty (21, flavour 1), load failed (22) | ADR-018 §7 |

### 1.4 `/workspaces/:wsId`

One route carries both the boards and the team, so this is the densest screen after the board itself — and the one whose composition no document settles. **[unsettled — see §6.1]**

| On it | Source |
|---|---|
| Board tiles: title, **identity swatch** from `board.background_color`, my role, joined / can-join, member count, card count | ADR-017 §3.3; US-08 |
| A **filled** Phosphor lock on private boards — the icon channel's only Phase-1 meaning, and the only place fill is permitted | ADR-018 §9 |
| A Join affordance on a board I may join but haven't | US-09 |
| Team list: name, email, role, promote / demote, remove, leave | US-06, US-06b, US-06c |
| Invite by email — a batch, with **per-email results**: unknown email reported not-found, existing member reported a conflict, and the valid invites still succeed | US-04 |
| Create board; delete workspace behind a destructive confirmation (38) | US-07, US-06d |
| Archived boards are **absent** from the list, not greyed | US-08 |
| Variants: cold load (19), empty boards (21, flavour 2), empty team, load failed (22), and the non-member case which is refused rather than rendered | ADR-018 §7, §10 |

The swatch is constrained to the eight identity hues and **must fall back on read** — `background_color` validates only `min=4,max=8` and is not a hex validator, so the string `"pink"` reaches the client (ADR-017 §3.3).

### 1.5 `/workspaces/:wsId/boards/:boardId` — the board

The product. Two layouts, crossing over at 640px (ADR-019 §2); from 640px up, ADR-017 §1–§2 apply unchanged.

| Region | Contents | Fine → Coarse |
|---|---|---|
| App bar, 56px | Board name 20/28; **presence cluster**; board `⋯`; add-column and add-card actions | — |
| Board accent | A 4px identity rail at the top, from the eight hues. **Not** a canvas colour — the canvas is always neutral `--board-canvas` | — |
| Canvas | `--board-canvas` `#E7E5E4`, horizontal scroll | — |
| Column panel | Own `--board-column` fill, 8px radius, 8px padding, 280px wide | 280px → `min(100vw − 72px, 420px)` below 640 |
| Column header | Shares the panel fill; `border-bottom: 1px solid var(--border)` is **load-bearing** because columns are unbounded and cards scroll under it. Title, count chip, `⋯` (rename, delete) | 36px → 44px |
| Card | 14/20 title clamped to **2 lines** (mandatory — titles accept 500 characters), 1px `--border`, **no resting shadow**, conditional meta row, `⋯` top-right | padding 8/12 → 12/12; avatar 20 → 24; meta row 24 → 28; gap 8 → 12; title right padding 24 → 32 |
| Meta row | Rendered **only** when the card has an assignee or a due date. Initials avatar, due date in `tabular-nums`, amber when due soon, rose when overdue | — |
| Column footer | Add card | — |

Per-card height is `8 + lines×20 + (24 if meta) + 8` at Fine and `12 + lines×20 + (28 if meta) + 12` at Coarse. The smallest card is 36px at Fine and exactly 44px at Coarse, so cards clear the target floor by construction rather than by special-casing.

The `⋯` is **hover-revealed at fine pointer and persistent at coarse** (ADR-019 §6), and it does not go in the meta row — that row is conditional, and putting `⋯` there would force it onto every card.

### 1.6 `.../cards/:cardId` — card detail

A URL-based overlay: the board renders underneath and browser back closes it. A drawer at 640px and above, a **full-height sheet** below (ADR-019 §3).

| On it | Source |
|---|---|
| Title at 20/28, editable inline (37 — the `--input` boundary doing the editing job) | ADR-017 §6; ADR-018 §10 |
| Description at 16/24 — the **only** use of that step in the app | ADR-017 §6 |
| Assignee picker, limited to board members | US-16; assignment-is-participation |
| Due-date picker; Phosphor calendar, Regular, 16px | ADR-018 §9 |
| **"Move to…"** — required here at every width | ADR-019 §6 |
| Delete, behind a destructive confirmation (38) | US-17 |
| Created / updated timestamps, `tabular-nums` | ADR-017 §6 |
| Variant: remote deletion of the open card (32) — the drawer closes with a toast | ADR-018 §10 |

---

## 2. Whole-screen boards

[ADR-020](adr/adr-020-design-tool-penpot.md) §5 fixes this list: a board only where **the whole screen differs**. Fifteen, numbered as they'll be named in the file.

| # | Board | Why it is whole-screen |
|---|---|---|
| 2.1 | Login | §1.1 |
| 2.2 | Register | §1.2 |
| 2.3 | Workspaces | §1.3 |
| 2.4 | Workspace detail | §1.4 |
| 2.5 | Board — at rest | §1.5. The reference frame everything else varies from |
| 2.6 | Card drawer over the board | §1.6, at ≥640 |
| 2.7 | **Read-only board** (24) | Mutating controls are **absent**, plus a persistent, **non-dismissible** banner. On touch the banner is the only remaining signal — a read-only user presses and holds and nothing happens. Routine browsing, not an edge case |
| 2.8 | **Join-required gate** (25) | A private board an admin hasn't joined: they see it exists, not its contents. One Join action, and joining is recorded |
| 2.9 | **Cold load** (19) | Static skeleton blocks at the card and column geometry, `--muted` fill, **no shimmer** — shimmer is movement, and movement means a teammate acted |
| 2.10 | **Empty board** (21) | No columns. New boards seed To Do / In Progress / Done, so this is the deleted-them-all case |
| 2.11 | **Load failed** (22) | Message plus a retry control |
| 2.12 | **Gone** (23) | Deleted, archived, or access lost |
| 2.13 | **Session expired** (39) | A 7-day JWT dying mid-session, drawn *before* the hard navigation to `/login`. Without it, expiry looks exactly like a crash |
| 2.14 | **Drag** | Four states at once, which is why it earns a board: source at 60% opacity (6), overlay with `--shadow-dragged` and `scale(1.02)` (7), a 2px dashed `--drop-placeholder` (8), and the column **body** tinted `--drag-over` with the header left untinted (9) |
| 2.15 | **Phone board, 390×844** | Density is a variable; layout is a breakpoint. Also the only place ADR-019 §7's "6 cards visible" gets checked by eye — the one number in its table resting on an assumption |

**On 2.14, the two things easiest to draw wrong.** The tint covers the column body only: on either tint the header's `--border` rule sits at about 1.19:1 and vanishes, exactly when it is most needed. And the placeholder is `#57534E`, not `--input` — measured against the tint rather than against white, `--input` thins from 4.80:1 to 3.22:1.

**On 2.15**, the strip is `min(100vw − 72px, 420px)`, and the 72px reservation is not arbitrary: it leaves the next column peeking by exactly **44px**, one tap target, which is what makes it a legal drop target. Below 640px the board does **not** scroll during a drag — valid targets are the current column and that peek.

---

## 3. Overlays

Drawn as component variant sets rather than boards, and placed on one or two holding boards. The destructive confirmation is **one component with several copies of its copy**, not several components.

| Overlay | Notable content | Source |
|---|---|---|
| Create workspace | Name only | US-03 |
| Invite people | Email batch plus **per-row results** — the mixed batch is the normal case, not the error case | US-04 |
| Create board | Title, visibility, colour picker restricted to the eight hues | US-07 |
| Board settings | Title, description, colour, visibility | US-12b |
| Board members | Add, remove (**which unassigns their cards on this board**), transfer ownership, leave | US-11, US-12d, US-12e, US-10 |
| Add people — candidate list | "Who I can add" is **P2** (US-11b), so Phase 1 may select from the full workspace member list instead **[unsettled — see §6.5]** | US-11b |
| Transfer ownership | Target must **already** be a board member; a non-member is refused rather than auto-added | US-12e |
| **Move card to…** | A **column** plus a **neighbour** — top of column, after «card», bottom of column — because position is a float, not an index. Three entry points: card drawer, card `⋯`, and `M` on a focused card | ADR-019 §6 |
| Destructive confirmation (38) | Rose confirm. Its unsatisfied state is drawn in `--muted` tokens, **never** in transparency | ADR-018 §5 |
| Column `⋯` and inline rename (37) | Rename, delete | US-13b, US-13c |
| Card `⋯` | Open, Move, Delete | US-15, US-17 |
| User menu | Name, email, initials — and logout **[unsettled — see §6.2]** | US-02b |
| Presence overflow sheet | The `+N` chip's full list, below 640px | ADR-019 §8 |
| Toast | Carries several states' messages; its own design is deferred | ADR-018 *Out of scope* |
| Date picker, assignee picker | Phosphor Regular 16px throughout | ADR-018 §9 |

**Presentation below 640px.** ADR-019 §3 commits only the *card modal* to becoming a full-height sheet. Whether the other fourteen follow is not settled anywhere. **[unsettled — see §6.4]**

---

## 4. Not screens, but drawn before them

### 4.1 Foundations board

| Holds |
|---|
| The token catalogue, as sets: `base`, `theme-light`, `theme-dark`, `density-fine`, `density-coarse` (ADR-020 §2) |
| The eight identity hues in **both role assignments** — pale fill with deep ink at Light, deep fill with pale ink at Dark, and the rail flipping to pale at Dark because every deep value fails 3:1 against a dark card |
| The elevation ladder: canvas → column → card, in both modes, with dark as a *compressed* version of the same ladder, never an inverted one |
| The radius ladder: 4px chips, 6px buttons and inputs, 8px cards, panels and overlays |
| The five type steps, and the **14px-against-16px specimen** ADR-017 names as the first thing on the page — it exists so the default can be revisited by looking rather than re-argued |

### 4.2 Variant containers

Per [ADR-020](adr/adr-020-design-tool-penpot.md) §4, every state is a variant. The card's `State` axis is `Rest · Hover · Pressed · InFlight · DragSource` — five variants across two densities and two themes is **one** component, not twenty frames.

| Component | Axis |
|---|---|
| Card | the five above |
| Button | rest, hover, pressed, focus-visible, busy, not-satisfied |
| Input | rest, focus, invalid, inline-editing |
| Avatar | eight hues × two role assignments |
| Column header | rest, drag-over (header **untinted**) |
| Count chip, banner, toast, skeleton block | — |
| Presence cluster | live, **reconnecting** — hues drop to flat `--muted` with a "Reconnecting…" label beside them, spending no hue |

Hover, pressed and the border ladder are three consecutive stone steps on the **edge**, never a fill: shadcn hovers to `--accent`, which is `--board-column`'s own value, so a card hovering to accent dissolves into the panel behind it.

### 4.3 What cannot be drawn

**Motion.** Penpot has no duration or easing token type, and the channel's meaning — "a teammate did this" — is a multi-client behaviour no static file can show. It stays code-only: 120ms chrome, 200ms board objects, 260ms containers, 110ms reduced-motion cross-fade.

**The identity palette as tokens.** It is a lookup indexed by `hash(id) % 8` living in a TypeScript constant, in both role assignments — not in the catalogue.

---

## 5. What the file totals

| | Count |
|---|---|
| Whole-screen boards | **15** (§2) |
| Overlay variant sets | ~**15** (§3) |
| Foundations board | **1** |
| Component variant containers | ~**10** (§4.2) |

Verification is a **predicate, not an eyeball** (ADR-020 §6): walk each board with `analyzeDescendants` and flag any descendant whose padding, gap, fill, radius or size is not token-bound, reading `shape.tokens`. Export only what genuinely needs looking at — the phone board, and the state boards as compositions.

---

## 6. Open questions — settle before drawing

Five. Three change what gets drawn; two are one-liners.

### 6.1 Workspace detail composition — **changes the board count**

One route carries both the board grid and the team list, and nothing picks a composition: tabs, a persistent side panel, or team-behind-a-dialog. Worth settling first because it bakes into component structure, and ADR-017's own Context is the argument — changing a thing's *role* after components reference it is a sweep.

### 6.2 There is no profile surface

US-02b (P1) wants a profile; the route tree has none. Either it's the user-menu popover or a route is missing. Reversible either way, so decide it in a line.

### 6.3 Archive has no exit — **a scope conflict, not a drawing question**

US-12c (P1) promises archive **and unarchive**. [ADR-018](adr/adr-018-state-vocabulary.md) §9 established that archived boards are filtered out of every list and 404 on every access path, and struck the archive icon on exactly that basis. So there is nowhere to click Unarchive, and nothing to click it on.

Three ways out, and this one flows **upstream** to the PRD and user stories rather than into the design file: draw Archive as one-way for Phase 1 and amend US-12c; add an archived-boards view and reopen the icon decision; or drop archive from Phase 1. Drawing the button before this is settled draws a trapdoor.

### 6.4 One sheet rule, or fourteen decisions

ADR-019 §3 commits only the card modal to becoming a full-height sheet below 640px. Decide the rule once for create-board, settings, members, confirmations and the rest; discovering it per-overlay while drawing is the expensive order.

### 6.5 The five empty flavours are not enumerated

ADR-018 counts "empty (five flavours)" and defers the *copy* as not a vocabulary decision — but the five surfaces themselves are never listed. My read, recorded as a reading and not a decision: **no workspaces, no boards in a workspace, no columns on a board, no cards in a column, nobody left to add to a board.** Worth pinning before the copy is written.

---

## Appendix — the 39 states, and where each is drawn

A completeness check against [ADR-018](adr/adr-018-state-vocabulary.md) §1. Three states are closed by inspection rather than drawn.

| # | State | Drawn on |
|---|---|---|
| 1, 2 | Pending create / change | Card variant `InFlight` — 60% opacity, meta row promoted from `--muted-foreground` to `--foreground` |
| 3 | Server-corrected | Folds into the movement channel; no separate frame |
| 4 | Rejected | **Not animated** — instant return, toast, dim lifts |
| 5 | Drag-armed | Board 2.14; `--shadow-overlay` ramping over the hold, with `scale(1.01)` |
| 6, 7, 8, 9 | Drag source, overlay, placeholder, column tint | Board 2.14 |
| 10 | Drag-rejected | No channel — instant return |
| 11 | Draggable at rest | Cursor at fine; **nothing at coarse** — the one accepted gap in the whole vocabulary |
| 12, 15 | Hover, pressed | Card and Button variants — border colour steps, never a fill |
| 13 | Hover-equivalent at coarse | The persistent `⋯` (§1.5) |
| 14 | Focus-visible | Every focusable component variant |
| 16, 17 | Not satisfied, busy | Button variants (§4.2) |
| 18 | Selected | **Does not exist in Phase 1** — which is what released 2px border weight for the placeholder |
| 19 | Cold load | Board 2.9 |
| 20 | Background refresh | **Nothing** — what's on screen is correct |
| 21 | Empty, five flavours | Boards 2.3, 2.4, 2.10 plus two in-column surfaces (§6.5) |
| 22, 23 | Load failed, gone | Boards 2.11, 2.12 |
| 24, 25 | Read-only, join gate | Boards 2.7, 2.8 |
| 26, 27 | Access revoked, action forbidden | Toast; the optimistic change reverts |
| 28 | Connecting | Part of cold load (2.9) |
| 29 | Reconnecting | Presence cluster variant (§4.2) |
| 30 | Live | **No treatment** — colour is an event, not a texture |
| 31 | Remote change arrived | The movement channel; not a static frame |
| 32 | Remote deletion of the open card | Card drawer variant (§1.6) |
| 33 | Presence change | Presence cluster |
| 34 | Private board | Filled Phosphor lock, on board tiles (§1.4) |
| 35 | Archived board | **Unreachable in Phase 1** — see §6.3 |
| 36 | Field invalid | Input variant; boards 2.1, 2.2 |
| 37 | Inline editing | Input variant; column rename, card title |
| 38 | Destructive confirmation | Overlay (§3) |
| 39 | Session expired | Board 2.13 |

---

## Related documents

- **Visual baseline:** [ADR-017](adr/adr-017-ui-foundations.md) — tokens, density, structure, type, elevation, modes
- **Responsive and touch:** [ADR-019](adr/adr-019-responsive-touch-scope.md) — layouts, drag, the coarse scale
- **State vocabulary:** [ADR-018](adr/adr-018-state-vocabulary.md) — all 39 states and their measured values
- **The medium:** [ADR-020](adr/adr-020-design-tool-penpot.md) — Penpot sets, themes, variants, verification
- **Routes and data flow:** [frontend-architecture.md](frontend-architecture.md)
- **Product intent:** [001-PRD.md](../product/001-PRD.md) · **Acceptance per feature:** [001-user-stories.md](../product/001-user-stories.md)
