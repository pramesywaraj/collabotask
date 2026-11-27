# CollaboTask

Collaborative task management app (Trello-style). This is a monorepo.

The backend (Go) is implemented and lives in `collabotask-backend/`.
**All technical detail — architecture, conventions, commands, the "add an endpoint" recipe — is in** `collabotask-backend/README.md` **(conventions + recipe) and** `collabotask-backend/TESTING.md` **(test approach). Read those for any code work.**
The frontend architecture is settled and documented (`docs/architecture/frontend-architecture.md`, ADR-011/012/013). App zone: React + Vite SPA — login, register and the Kanban app. Public zone: Astro (SSG) — landing only. Frontend implementation is a separate effort.

## Now

**Phase 1 backend is complete. Frontend implementation — not started.**

Build order (spec §9.1): ① P0 fixes ✅ → ② positioning ✅ → ③ REST features ✅ → ③.5 cookie auth ✅ → ④ WebSocket ✅ → ⑤ integrity pass ✅ — **all closed.**

Step ⑤ closed 2026-10-26 on branch `test/deferred-sql-tests`, **merged to `main` 2026-11-02 (PR #15).** The deferred repository-layer SQL tests are written (31 integration tests), the suite has been run end to end, and the Transactor carry-over items are all discharged. `-race` is clean after fixing a pre-existing shared-fixture race in `set_archived_test.go` (unit tree re-verified 8/8 on 2026-10-27; `tests/integration` verified 5/5 on 2026-10-26).

**One open backend item:** five files fail `gofmt` — own branch, no rush, nothing enforces formatting. See `docs/handoff/handoff-integrity-pass-findings.md` › #3.

The default-column spacing defect (boards seeded at 0/1/2 instead of `PositionStep` multiples) was **fixed 2026-11-02** — new boards now seed at 1000/2000/3000, so the frontend's drag-and-drop starts with the gaps ADR-004 intends.

Frontend: architecture is settled (ADR-011/012/013, `docs/architecture/frontend-architecture.md`), the **visual baseline is settled — [ADR-017](docs/architecture/adr/adr-017-ui-foundations.md)** (supersedes the uncommitted `docs/architecture/design-tokens.md`), the **responsive + touch scope is settled — [ADR-019](docs/architecture/adr/adr-019-responsive-touch-scope.md)** (supersedes ADR-017 §8: phones and tablets are in scope for Phase 1, card dragging works on touch), and the **state vocabulary is settled — [ADR-018](docs/architecture/adr/adr-018-state-vocabulary.md)** (2026-11-18, written after ADR-019): 39 interaction states enumerated before any channel was spent, every value measured, motion re-pointed from server-correction to "a teammate did this", Phosphor icons. The **design tool is settled — [ADR-020](docs/architecture/adr/adr-020-design-tool-penpot.md)** (2026-11-18, supersedes ADR-018 §12 only): the design file is **Penpot**, named `Collabotask App`. The Figma account is starter tier on both teams, which gates the two-mode variable collections §12 required; Penpot token sets + theme groups deliver the Fine/Coarse flip for free — verified empirically, so §12's scope is restored rather than cut. Still **no code written**. The design file is **started, not empty** — verified against the live Penpot file on 2026-11-19: the five token sets (206 tokens as of 2026-11-22), both theme groups (`Theme`: Light/Dark, `Density`: Fine/Coarse) and the Card's five-state variant container all exist in `Collabotask App`. The **foundations board is drawn** (2026-11-20, page `Foundations`; binding check clean, both flips looked at). **Stage 2, the proof run, is done** (2026-11-20): the Input and Button variant containers on `Components`, and board **2.1 Login** on `Page 1` built from them; the binding check returns nothing, and all four Theme × Density combinations were looked at. **Stage 3 is closed** (2026-11-20): the Avatar, the Count chip, the Column header, the Banner, the Skeleton block, the Toast and the Presence cluster are done, and the Card uses the Avatar. All ten parts exist. **Stage 4 has begun** (2026-11-21): board 2.2 Register is drawn, with the two parts it needed, and **group b is done**: 2.3 Workspaces and 2.4 Workspace detail, with eight state boards and three more parts. **The sidebar's two parts are drawn, and 2.4 and its five state boards are redrawn with it** (ADR-035). **Group c is done**: 2.5 Board — at rest, the reference frame (ADR-037, 2026-11-23). **Group d is done**: 2.6 to 2.13, nine boards (ADR-038, 2026-11-23). **Group e is done, so Stage 4 is closed**: 2.14 Drag and 2.15 Phone (ADR-039, 2026-11-23). **Stage 5 has begun**: group a, the menus and the confirmation, and group b, the card drawer's pickers, are done (ADR-040, ADR-041, 2026-11-23), and so are group c, the forms (ADR-042, 2026-11-24), group d, the lists (ADR-043, 2026-11-24), and group e, the sidebar sliding over and the toast (ADR-044, 2026-11-24). **Stage 5 is closed.** **The boards were then regrouped by area, before Stage 6** (ADR-045, 2026-11-25): the pages are Foundations, Components, **Account**, **Workspaces** and **Board**, and `Page 1` and `Overlays` are gone. **Stage 6 closed 2026-11-27** ([ADR-046](docs/architecture/adr/adr-046-stage-6-final-check.md)): the binding check is clean on every board, and all four Theme × Density combinations were reviewed. It **records an exception to 4.5:1 for text inside a dimmed, inert object** (the avatar's initials on an in-flight card, and the contents of a dragged column's place; the user's call), fills the twelve oldest documentation boards, leaves the Count chip board's clipped chips as a known fault, and says what it did not look at by eye. No token (222), no part (34). **Later the same day, illustrations went from 128 to 192** ([ADR-047](docs/architecture/adr/adr-047-illustration-size-192.md), the user's call after looking at 1.5× and 2× on a test copy): the part's shapes now scale with its box, the line stays 2px, and the code needs `--illustration-size: 192px` and a 2px stroke that does not scale.

**The design inventory's five open questions are settled — [ADR-021](docs/architecture/adr/adr-021-design-inventory-corrections.md)** (2026-11-19). Grilled against the running API rather than the other ADRs, which turned three of them into facts and found four gaps §6 never asked about: archive is **two-way and always was** (`SetArchived` is the one board path that skips `BoardAccessChecker`, so unarchive works today — only a *list* was missing, so ADR-018's "Unarchive is unreachable" is corrected, and the SRS had said so all along); **column reorder gets Move left / Move right** in the column `⋯` menu, closing a WCAG 2.5.7 gap ADR-019 §6 built for cards and skipped for columns; **overlays below 640px follow four rules, not fifteen decisions**; and the **read-only banner carries Join** — its only possible occupant is always allowed to join. Also settled: workspace screen = boards full-screen + team behind a member-count button, profile = the user menu, six empty flavours not five, Create drops you inside what you made, and password reset is now a recorded Phase-2 deferral rather than an unnoticed hole.

**Form-control geometry is settled — [ADR-022](docs/architecture/adr/adr-022-form-control-geometry.md)** (2026-11-20, **extends** ADR-017 §9 — supersedes nothing). Found by running the design file's own Stage 2 question 4 against the live catalogue *before* anything was drawn: ADR-017 fixes what a control looks like (14/20 text, 6px radius, the `--input` boundary) but never how big one is, because §9's vocabulary is the board and Login is a form. Four values — `size.control-height` 36/44, `spacing.control-padding-x` 12, `spacing.field-gap` 16/20, `size.form-width` 400 — now in Penpot and flip-verified. **ADR-017 is annotated, not edited**, which is its own established pattern.

**Spacing between objects is settled — [ADR-023](docs/architecture/adr/adr-023-spacing-between-objects.md)** (2026-11-20, **extends** ADR-017 §9 and adds a `borderWidth` row to ADR-020 §3 — supersedes nothing). Stage 1's question 4 fired before drawing: every spacing token sat inside one object, and nothing named the space *between* two. Six of its eight values are derived from ADR-017/018/019 rather than chosen — `spacing.screen-x` 16, `spacing.column-gap` 8/12 (the only gap ADR-019 §2's width table reconciles at), **zero vertical canvas padding** (what both card-count derivations already assume), `size.board-rail` 4, `border.width` 1 / `-strong` 2. The two choices: `spacing.section-gap` 24 and initials at `type.meta`. In Penpot and flip-verified. The inventory's §7.5 binding check was also **rewritten**: as first written it could not see fills, padding or gaps at all.

**The inside of a field and a control is settled — [ADR-024](docs/architecture/adr/adr-024-control-and-field-internals.md)** (2026-11-20, **extends** ADR-017 §9 — supersedes nothing). Stage 2's question 4 fired a third time before drawing: five values Login needed had no token. `spacing.field-inner-gap` 8 and `spacing.icon-gap` 8 (chosen, kept apart by ADR-022 §4), `size.icon` 16 (Phosphor's size inside a control), and **`shadow.focus-ring`**: ADR-018's 2px ring at 2px offset drawn as two stacked shadows, because Penpot strokes cannot be offset; it is also how Tailwind's `ring-offset` renders. **Pressed on a filled button reuses `color.primary-hover`** (user's call). Found while drawing: an input's box is filled `color.background`, since Penpot casts an unfilled board's shadow from its children. **Scaffold consequence:** current shadcn's Input and Button focus classes (`ring-[3px]`, no offset, `bg-transparent`) must be overridden in the same pass as ADR-018's `disabled:` override. ADR-017, -018, -020 and -023 annotated. Open, not blocking: where a wrong password is reported on Login.

**The identity hues are tokens in the design file — [ADR-025](docs/architecture/adr/adr-025-identity-hues-as-tokens.md)** (2026-11-20, **supersedes ADR-020 §3's identity exclusion only**). Avatar's question 4 found no missing number but a rule: ADR-017 §3.3 swaps the hues' roles by mode, and typed-in colours cannot follow Penpot's Theme switch. That would have meant 18 variants, Dark mockups swapped by hand, and a Dark check blind to every avatar and rail. Compared side by side, the user chose tokens:
- `color.identity.<hue>.subtle` (avatar fill) and `.strong` (initials and rail), 32 entries
- `size.card-avatar` renamed `size.avatar`
- the Avatar's ten appearances, including a Phosphor `User` glyph for an empty name and a Muted look for presence's reconnecting state

**The code is unchanged**: ADR-017 §9's TypeScript constant stands. The Foundations board's side-by-side pairings stay typed in, and they are the only `Hue /` shapes the binding check still expects.

**The count chip is settled — [ADR-026](docs/architecture/adr/adr-026-count-chip.md)** (2026-11-20, **extends** ADR-017 §9 — supersedes nothing). Question 4 fired a fifth time in five: ADR-017 §3.4 measured the muted text against the chip, but §9 never gave the chip a variable. Its values matched `color.border` only by coincidence, and it had no geometry at all. Now:
- `color.chip` (`#E7E5E4` / `#44403C`), literals rather than `{color.border}`
- `size.chip-height` 20, also the minimum width
- `spacing.chip-padding-x` 6, measured as `(20 − 8) ÷ 2` against Poppins's widest digit

The chip holds text, so tiles can say `12 cards`. `tabular-nums` cannot be drawn, and the `+N` chip belongs to the presence cluster. **Code consequence:** §9 gains `--chip` in both modes.

**The column header is settled — [ADR-027](docs/architecture/adr/adr-027-column-header.md)** (2026-11-20, **extends** ADR-017 §9 — supersedes nothing). Question 4 fired a sixth time in six: ADR-017 §2 and ADR-019 §7 fixed the header's outside, and nothing said what is inside it. The user took all the recommendations:
- the inset binds `spacing.card-padding-x` 12 by rule, so the column title starts where the card titles start
- `spacing.inline-gap` 8, new in `base`, between title, count and `⋯`
- the `⋯` is Phosphor `DotsThree` at `size.icon` in a `size.target-min` area, **shown at rest at every pointer**, unlike the card's
- one appearance, no axis: drag-over leaves the header untinted

**No export shows the header's rule**: Penpot's exporter drops strokes whose sides differ, so check it on the canvas. ADR-017 and ADR-021 §3 annotated.

**The banner is settled — [ADR-028](docs/architecture/adr/adr-028-banner.md)** (2026-11-20, **extends** ADR-017 §9 — supersedes nothing). Question 4 fired a seventh time in seven, on one value: the space above and below the Join button. The user took all the recommendations:
- `spacing.banner-padding-y` 8, new in `base`; padding, not a height, because the message wraps on a phone
- fill `color.background`, because Join's focus ring draws its offset in it, plus a bottom rule, because Dark's background and canvas are the same colour
- `spacing.screen-x` at the sides, `spacing.inline-gap` before Join
- **one appearance, no informational variant**: every visitor to a read-only board can join, so inventory §4.2's second appearance had nobody to show it to
- no icon; the wording is *You’re viewing this board. Join it to add, edit and move cards.* with a *Join board* button

**Make a part's instances only after verifying its main, in a separate call**: an instance made in the same call missed one late binding, and `resetOverrides()` did not restore it. ADR-017, ADR-021 §6 and ADR-027 §2 annotated.

**The skeleton block is settled — [ADR-029](docs/architecture/adr/adr-029-skeleton-block.md)** (2026-11-20, **supersedes ADR-018 §7's `--muted` fill only**, extends ADR-017 §9). Question 4 fired an eighth time in eight, and for the first time on a value that existed but could not be seen: in Light `color.board-column` is stored as `{color.muted}`, so a muted block inside a column measures 1.00:1. The user took the recommendations:
- `color.skeleton` `#E7E5E4` / `#292524`, written as values: 1.15:1 on the column in both themes, and in Dark it is the card's own fill
- `size.skeleton-height` 60 / 72 in the density sets: ADR-017 §1's formula for one title line and a meta row
- card-sized blocks inside **real column panels**, because one colour cannot draw both sizes nested; no outline, one height, no axis
- **a block never sits directly on the canvas** (1.00:1 in Light) or inside a card-coloured surface (1.00:1 in Dark). The `/workspaces` and workspace-detail cold loads must respect this at Stage 4.

**Code consequence:** §9 gains `--skeleton` and `--skeleton-height`, and shadcn's Skeleton loses `animate-pulse`. ADR-017 and ADR-018 §7 annotated.

**The toast is settled — [ADR-030](docs/architecture/adr/adr-030-toast.md)** (2026-11-20, **supersedes ADR-018 §10's row-26 destination for `removed_from_workspace` only**, extends ADR-017 §9). Question 4 fired a ninth time in nine. shadcn's `color.popover` is 1.00:1 over a card in Light and over a column in Dark, and the toast had no padding or width. The user took the recommendations:
- an **inverted** toast: `color.toast` `#44403C` / `#E7E5E4` (grey, not the text colours) with `color.toast-foreground` `#FFFFFF` / `#0C0A09`, and no edge. It is 8.18:1 or better over every surface.
- `spacing.toast-padding` 12 and `size.toast-width` 358 (`390 − 2 × screen-x`), both in `base`. The catalogue is now 169 tokens.
- one appearance, no axis, no icon, no hue, and a Phosphor `X` close shown at rest
- three access-revoked wordings, which name what you lost and not who did it
- **`/workspaces` after `removed_from_workspace`**, since that workspace can no longer be opened (frontend-architecture.md updated)

**Code consequence:** §9 gains `--toast`, `--toast-foreground` and `--toast-width`. If the scaffold uses shadcn's Sonner, its `--normal-*` variables are re-pointed. Still deferred: where a toast sits, and the wordings for 4, 27 and 32. ADR-017, -018, -021 and -022 annotated.

**The presence cluster is settled — [ADR-031](docs/architecture/adr/adr-031-presence-cluster.md)** (2026-11-20, **supersedes ADR-018 §10's reconnecting fill, ADR-025 §4's grey circle and ADR-019 §8's `+N` chip only**, extends ADR-017 §9). Question 4 fired a tenth time in ten, and Stage 3 closes with it. The overlap had no Fine value, nothing separated overlapping circles, and `color.muted` is 1.09:1 on the white app bar. The options were compared on a rendered decision page. The user took the recommendations except one: `+N` is a round seat, not a Count chip.
- `spacing.avatar-overlap` −6 / −8, drawn as a negative left margin, because Penpot refuses negative gaps
- `shadow.avatar-ring` (2px in `{color.background}`) and `color.avatar-neutral` (`#E7E5E4` / `#57534E`), which is now also the Avatar's Muted and Unknown circle
- the seat: `radius.full` 9999, `spacing.seat-padding-x` 2 / 4 and the neutral fill. The catalogue is now 178 tokens.
- the whole cluster is the button, `size.target-min` tall. You come first, with three faces below 640px and five above. On a phone, *Reconnecting…* takes the seat's place.
- `State` = Live · Reconnecting · FocusVisible

**Code consequence:** §9 gains `--avatar-neutral`, `--avatar-overlap` and `--seat-padding-x`, and the faces and the seat wear `ring-2 ring-background`. §7.5's binding check now sees margins. **One slip:** a name lookup overwrote the Banner board's `Label / Phone`, whose wording was not recorded. It was rebuilt as *Phone, 390 — two lines*. ADR-017, -018, -019, -025 and -026 annotated.

**Where Login and Register report errors is settled, and Register has four password rules — [ADR-032](docs/architecture/adr/adr-032-auth-form-errors-and-password-rules.md)** (2026-11-21, settles ADR-024's open question, **supersedes ADR-018's Emerald ledger row only**, extends ADR-017 §9). Question 4 found no missing value in round 1, the first time in eleven stages. It fired in round 2. Both rounds were decided on rendered pages.
- a form-level failure is a `Form message`, one line just above the button, carrying Login's 401 (*Wrong email or password.*). A taken email stays on its field.
- **the four password rules were the user's call, against NIST SP 800-63B-4**, which forbids composition rules. That is recorded in the ADR, with what NIST asks for instead. They are shown as a live `Password rule` checklist (Unmet · Met · Missed, met in `color.done` emerald, the glyph's shape carrying the state), rows `spacing.list-gap` 4 apart. The catalogue is now 179 tokens.
- a confirm-password field, client-only
- *Sign in* in place of *Log in*, every wording on both forms, and fifteen behaviour rules
- **US-01 and SRS UC-01 changed first.** A rule only the browser checks protects nothing.

**The workspace screens are settled and drawn — [ADR-033](docs/architecture/adr/adr-033-workspace-screens.md)** (2026-11-21, **supersedes ADR-018's Plum ledger row and §7's row 21, for whole-screen empties only**; extends ADR-017 §9). Question 4 fired a twelfth time in twelve, in two rounds, both decided on rendered pages. The user took seven of nine recommendations in round 1 and chose the roomy tile and a left rail. A note asking for illustrations on the empty states started round 2.
- both screens on `color.background`; `size.app-bar` 56; the `Tile` part (padding 16, rows 8 apart, a left rail in the hue's `.strong`), `spacing.tile-gap` 8 / 12, `size.tile-min-width` 264
- the Active / Archived `Segment` (plum text and edge: plum's *where you are*, state 18 stays closed); an **Outline** `Kind` on the Button for *N members*, *Join*, *Unarchive* and *Try again*; *Delete workspace* behind an owner-only `⋯`
- **empty-state illustrations** on whole-screen empties (flavours 1–4), centred, 128, a 2px ink line and a **plum tint, the user's call against a grey**. ADR-033 §9 holds **the rules for drawing the next one**; 2.10's empty board is the next to need one.
- *Owner · Admin · Member*, every wording, and the behaviour rules. The catalogue is now 194 tokens and the file has 15 parts.

**Code consequence:** §9 gains the tile, grid, app-bar and illustration variables, and shadcn's `outline` button is restyled to the edge ladder. ADR-017, -018, -021, -023 to -027 and -029 annotated.

**The side menu is settled — [ADR-034](docs/architecture/adr/adr-034-side-menu.md)** (2026-11-21, **supersedes ADR-033 §2's app bar and §4's tile counts on 2.4 only, and ADR-019 §2's 1024 and 1440 rows while the sidebar is open**; extends ADR-017 §9 and ADR-021 §5). A side menu had never been discussed. It was grilled in eleven questions, then eight visual choices went on a rendered page, and the user took every recommendation.
- a sidebar on **2.4 and the board, not 2.3**: a workspace switcher that opens inside it, *All boards* (→ 2.4), then only the boards you've joined, filtered in the browser. *CollaboTask* sits at its top and the user menu at its bottom
- **beside the content from 1024 up**: full height, 240 wide, page white with a rule, open by default and remembered on the device, toggled with Ctrl / ⌘ + \\. **Below 1024 it slides over** the content (a fifth presentation kind, *navigation*), with `color.scrim` dimming the content
- plum text and edge on the current row; a 12px colour square per board; the app bar names the current row (*All boards* on 2.4)
- 2.3 stays, and the app opens in your last workspace (`localStorage`, frontend only)
- four tokens, five entries (194 → 199), in Penpot since 2026-11-21. A colour token takes transparency, so `color.scrim` is stored as `rgba()`; Penpot refuses the space-separated form

**Code consequence:** shadcn's `Sidebar` differs in five places (width, shortcut, cookie, breakpoint, icon; ADR-034 §14), and the route tree gains a shared layout route and a `/` redirect (frontend-architecture.md). ADR-017, -018, -019, -021 and -033 annotated. **One exception to know about:** from 1024 up the board **glides** 240px with the sidebar. That was the user's call over a jump, and it narrows ADR-018 §8's rule 3 for this one case. Build it as one width change plus a `transform` slide, not a width animation (ADR-034, the note under §10.10).

**The sidebar's loading, failure and switcher mark are settled, and the sidebar is drawn — [ADR-035](docs/architecture/adr/adr-035-sidebar-loading-failure-and-switcher-mark.md)** (2026-11-21). It answers ADR-034's open question and question 4 for its parts, and **narrows ADR-018's Rose row and row 22 for the sidebar's list only**. Round 1 found no missing value, only five decisions. Round 2 came from two notes, and the dot's size fired question 4 (14 out of 14). Both rounds were decided on rendered pages.
- cold load: row-sized blocks for the switcher and the board rows. The user's call over a blank switcher.
- load failure: *Your boards didn’t load.* in rose, then an uncoloured *Try again* row with `ArrowClockwise`. The user's call over no rose. The pages' own failures stay uncoloured.
- the open switcher: workspace rows indented with the board names, and **a still plum dot** (`size.dot` 8, new) on the selected one. The user's call over a check. A pulse was not offered, because motion means *a teammate did this*.
- `spacing.list-gap` under *Your boards*, and the user row's avatar in the slot's place
- parts `Sidebar row` (one for every row) and `Sidebar` (`Switcher` × `List`, five variants). The catalogue is 200 tokens and the file has 17 parts.

**Code consequence:** §9 gains `--dot`, and the sidebar's retry is a `SidebarMenuButton`, not a `Button`. ADR-017, -018 and -034 annotated.

**One-sided rules are drawn as inset shadows — [ADR-036](docs/architecture/adr/adr-036-one-sided-rules-as-inset-shadows.md)** (2026-11-22). The user found the lines missing on the canvas. Penpot saves a stroke's per-side widths but draws only its general width, so **no one-sided rule had ever drawn**: 50 shapes since Stage 1, including ADR-017 §2's load-bearing column-header rule and the banner's only lower edge in Dark. The user row drew as a full box. ADR-027's "it's only the exporter" was wrong.
- `shadow.rule-top` / `-right` / `-bottom`, once per theme set (200 → 206): an inset 1px in `{color.border}`, offset `{border.width}` (Penpot evaluates `* -1`). The user's go-ahead over a pinned 1px rectangle.
- 54 rules redrawn, with no strokes left. A rule's shape must be filled, because an inset shadow draws nothing on an unfilled shape.
- **No code change**: build rules as borders (`border-r` / `-b` / `-t`), whatever Penpot's inspect tab says.

**2.5 Board — at rest is settled and drawn — [ADR-037](docs/architecture/adr/adr-037-board-at-rest.md)** (2026-11-23). Question 4 fired on one value, the space between a column's header and its first card (15 out of 15). Five decisions came with it. All six were decided on one rendered page, and the user took every recommendation but the rail's.
- the first card sits `spacing.card-gap` below the header, bound by rule, so **no token is added (206)**
- the rail is its own 4px strip under the app bar. That was the user's call over a rail inside the bar, and it costs 4px of height.
- the bar: the name on the left; presence, *Add column* (Primary, the board's one plum button) and a `color.foreground` ⋯ on the right
- **`Add card`, a new part (18)**: a quiet row after the last card that sticks to a long column's bottom, its stuck rule on a `Footer` wrapper. **A phone then shows 5 cards and a peek in a long column**, not ADR-019 §7's 6.
- **a column is as tall as its content, at most to the bottom of the screen** (the user's call, narrowing ADR-023 §2.3's wording). 2.14 must decide whether a short column's drop area still reaches the bottom.
- **late dates in words**, because colour alone failed WCAG 1.4.1: *Due today*, *Due tomorrow*, *Overdue · Nov 20*. *Due soon* means today or tomorrow.
- *No cards yet.* in an empty column, and the Card's meta row now keeps a lone avatar on the right

**Code consequence:** no CSS variable is added. The column body takes `pt`/`gap` of `--card-gap`, *Add card* is `position: sticky` with a top border only while stuck, and one function turns a due date into its wording and colour. Still open: where *Add column* goes below 640 (2.15). ADR-017, -019, -023, -031, -033 and -036 annotated.

**Group d is settled and drawn — [ADR-038](docs/architecture/adr/adr-038-board-states-group-d.md)** (2026-11-23). Question 4 fired on three values (16 out of 16): the drawer's width, its labels' width, and a left-edge rule. Two rendered rounds, 22 decisions; the user took every recommendation but two.
- **the card drawer**: over everything, dimmed; `color.popover` with `shadow.rule-left` (new) and `shadow.overlay`; `size.drawer-width` 560 and `size.field-label` 96 (new; 210 tokens); a new `Field` part (19 parts); *Column* opens *Move card to…*; *Delete card* in rose at the foot (the user's note); dates as *Due tomorrow · Nov 24*
- **2.7**'s banner sits under the rail. **2.8** is a centred stack, which narrows ADR-021 §6
- **one centred stack** on 2.8 and 2.10 to 2.12: drawing, 16, sentence within 400, 24, button. **2.9** draws the name and rail at once, with panels out to the edge
- **drawings on screens that aren't empty** (padlock, plug, closed box, hourglass), widening ADR-033 §8–§9. **All tinted plum, the user's call over ADR-018 §7's uncoloured failures**
- **2.13 waits for *Sign in* and then returns you to the same page.** Left-aligned like Login (the user's call)

**Code consequence:** `--drawer-width` and `--field-label`. The axios 401 handler sets `uiStore.sessionEnded` instead of navigating, and Login honours a same-site `?next=` (frontend-architecture.md). **Now open backend item 3:** `DeleteWorkspace` sends no live message, so an open board learns of it at its next request (ADR-038 §11). Still for Stage 5: *Move card to…* from the drawer on a phone is a sheet over a sheet, which ADR-021 §5 forbids. ADR-017, -018, -021, -028, -033, -036 and -037 annotated.

**Group e is settled and drawn, and Stage 4 is closed — [ADR-039](docs/architecture/adr/adr-039-drag-and-phone-board.md)** (2026-11-23). Question 4 fired on three values (17 out of 17): the lifted card's 2% growth, the phone column's 318 and the browser's 140. Five decisions came with them. One rendered page; the user took seven recommendations and overrode one.
- **a short column takes a card only as far as it goes**: nothing changes height during a drag. That was the user's call over stretching every column to the bottom for the drag, and it settles ADR-037's open question
- the tint covers the cards' area only; *Add card* stays untinted just under it, so a long column's footer keeps its rule
- the `Card` part gains `Armed` and `Lifted` (seven looks); the 2% and 1% growth is code only (inventory §4.3). A dimmed overdue date keeps its words, not its rose
- **2.14 is two boards**, a card drag and a column drag; a column drag tints nothing
- **below 640, *Add column* is first in the board ⋯ menu**, and the empty board's stack gains an *Add column* button there (amends ADR-038 §10)
- **2.15** shows the browser as two labelled bands, types its 318, and has a drag board beside it. **Counted by eye: 5 whole cards and 44px of the sixth**, so ADR-019 §7's revised count holds

**Code consequence:** none in CSS. The drop target is the column panel and the tint goes on the card list, not on what holds *Add card*. The board ⋯ menu renders *Add column* first below 640. No token (210) and no part (19) added. ADR-019, -037 and -038 annotated.

**Stage 5's group a is settled and drawn: the menus and the confirmation — [ADR-040](docs/architecture/adr/adr-040-menus-and-the-confirmation.md)** (2026-11-23). Question 4 fired on six values (18 out of 18) and fourteen choices, one of them a conflict. One rendered page; the user took twelve recommendations, chose a pale rose busy look, and added *Leave board* to what is rose.
- **one menu surface**: `color.popover` with an edge (in Dark it is the column's colour), `size.menu-width` 224, `spacing.menu-padding` 4 (16 − 12), 4 below its ⋯; a glyph per entry; below 640 a bottom sheet with a head and a Close
- **rose for the three deletes and *Leave board***; *Archive board* plain, with a plum confirm. All five ask first
- **US-12c's *Undo* after archiving is kept**, superseding ADR-030 §3's *no action button* for that toast. Group e draws it
- **a 448 dialog** (`size.dialog-width` = `size.form-width` + 2 × `spacing.dialog-padding` 24), 358 at 390, focus on *Cancel*. *Delete workspace* asks you to type the name
- **the Button's rose kind**: `color.destructive-hover`, and Busy in `color.destructive-subtle` / `-subtle-foreground` (narrows ADR-018 §5)
- inline rename fills the header; *Sign out* (US-02b edited); every wording, and where each action leaves you (frontend-architecture.md)
- 220 tokens; 23 parts (`Glyph`, `Menu row`, `Menu`, `Confirmation`); **a new page, `Overlays`**, with nine screens each showing one overlay open. The plugin can't copy across pages, so the four screens were pasted by hand. `Overlays` is now Board, split by area (ADR-045)

**Code consequence:** §9 gains `--menu-padding`, `--menu-width`, `--dialog-padding`, `--dialog-width` and the three rose variables; shadcn's `DropdownMenu`, `Sheet` (bottom, below 640), `AlertDialog` and the `destructive` button are restyled (ADR-040 §15). **Found while drawing:** `createShapeFromSvg` drops its shape into whatever board is at the canvas origin, and on `Components` that is the Card's main. Draw glyphs with `createPath`. Keep the Penpot window beside VS Code, not behind it, or it sleeps. ADR-017, -018, -021, -024, -027, -030, -033, -034, -038 and -039 annotated.

**Stage 5's group b is settled and drawn: the card drawer's pickers — [ADR-041](docs/architecture/adr/adr-041-the-card-drawers-pickers.md)** (2026-11-23). Question 4 found **no missing value**, the first stage since ADR-032's first round (18 of 19 have fired), and sixteen choices. One rendered page; the user took every recommendation.
- **no sheet over a sheet**: in the drawer every picker opens in place under its field, *Move card to…* too. That supersedes ADR-021 §5's placement of *Move card to…* only, and narrows ADR-038 §5: all three fields carry the caret
- from the board, Move takes the card ⋯ menu's place, the popover from 640 up and the bottom sheet below. A column, then a place (*Top of column*, *After “card”*, *Bottom of column*); the place moves the card
- under the value (432) from 640 up, across the row below; group a's menu surface laid in the flow; plum text and a plum edge on the current choice, so state 18 stays closed
- a six-week calendar: today semibold, the week from the locale, past days pickable, no quick picks, *Remove due date*
- **a due date is a calendar day at midnight UTC, read back by its date part**. SRS §3.6, UC-14 and UC-16 were edited first; no backend change
- the board's members only: *Unassigned*, you, then A–Z, with a filter above eight members
- 220 tokens; 27 parts (`Picker row`, `Day`, `Calendar`, `Picker`), and the `Glyph` part at 19 looks; eight new `Overlays` boards, among them the first drawer at 390. 2.4's Q3 tile corrected to *8 members*

**Code consequence:** no CSS variable. shadcn's `Calendar` is restyled (no fill on the picked day, no `bg-accent` on today, `fixedWeeks`, no outside days). The in-place pickers are disclosures, not popovers, and Move is one component in three places; that and the due-date rule are in frontend-architecture.md. **Found while drawing:** binding a text's weight drops its typography binding, so today's number is a new expected row in §7.5's check, and `setParentIndex` counts from the other end of `children`. ADR-018, -019, -021, -037, -038 and -040 annotated.

**Stage 5's group c is settled and drawn: the forms — [ADR-042](docs/architecture/adr/adr-042-the-forms.md)** (2026-11-24). Question 4 fired in two rounds (19 of 20 have fired). Round 1 found one value and twenty choices. The user took sixteen recommendations and chose errors as you type, two choice rows, chips and the results under the field. Round 2 settled a chip's height and the colour square's size, both by rule with no token.
- **one form frame**: group a's 448 dialog from 640 up, with no ×, and pressing outside it does nothing. Below 640, a full-height sheet with *Close* in its head and one full-width button. Escape, *Cancel* and *Close* never ask
- **`spacing.control-padding-y` 8 / 12** for a field of several lines, which grows from one line. This narrows ADR-022 §3; 222 tokens
- **a minimum length shows as you type** (the user's call); an empty field waits for the press. This narrows ADR-032 §8's rule 1 for minimum lengths only, and Login and Register don't change
- **colour** as eight buttons holding their hue (20 at Fine, 28 at Coarse, the user's call), with a plum edge on the chosen one. The first hue no active board uses comes first. **`background_color` stores the deep hex value**, read back by case-insensitive match (SRS §4.3, edited first)
- *Workspace* and *Private* as two choice rows. Board settings has *Save changes*, and a line saying who loses access before a board goes private
- **Invite people** is an Outline button on 2.4's heading, for admins, now drawn on eight boards. Addresses are chips, a target tall. After sending, the results sit under the field with the checklist's marks, and the field keeps the addresses with no account
- **Add column is not a form**: it adds a new column at the end of the board, its header the rename field
- 32 parts (`Swatch`, `Choice row`, `Address chip`, `Invite result`, `Form`), the `Input`'s `Content` axis, three `Glyph` looks, and ten new `Overlays` boards

**Code consequence:** §9 gains `--control-padding-y`. One `FormFrame` component (shadcn `Dialog`/`Sheet`), `Textarea` restyled, the colour and visibility choices as Radix `RadioGroup`s, `hueFromBoard` and `firstFreeHue`, trim before validating, and a client-only draft column (frontend-architecture.md). **Corrected while checking:** the decision page said a workspace description had no server limit. It does: the usecase validates `max=1000`. So no backend item was added. **Found while drawing:** `createPath()` gives a path a black 1px stroke, so remove it. A screen can cross pages through a temporary part: make the part on the source page, place a copy on the target page, detach it, then delete the part. A canvas set to `justifyContent: end` draws the board scrolled to its end. ADR-017, -018, -021, -022, -032, -033, -039, -040 and -041 annotated.

**Stage 5's group d is settled and drawn: the lists — [ADR-043](docs/architecture/adr/adr-043-the-lists.md)** (2026-11-24). Question 4 fired on one value, how tall a list gets before it scrolls, and it resolved by rule (20 of 21 have fired). Twenty choices, on one rendered page; the user took nineteen recommendations and chose the role on the row's second line, with the email.
- **group c's frame holds every list**: the 448 dialog from 640 up, the full-height sheet below, with *Close*. **Eight rows, then it scrolls** (8 × `size.control-height` + 7 × `spacing.list-gap`); no token, so 222
- **a person's row**: an Avatar, the name and *(you)* over *Role · email*, then a ⋯ or *Add*. Every row says its role; you first, then A–Z
- a ⋯ per row opens the menu; **ten actions ask first**; *Remove member*, *Remove from board* and *Leave workspace* are rose. The entry is *Remove member* because *Remove from workspace* didn't fit a 224 menu row (§19)
- *Leave workspace* is in the workspace ⋯ for everyone but the owner, and the owner's row has no ⋯
- **adding people is the board list's second group**, *Not on the board yet*, one person per press, without you until you join. **Handing a board over** is *Make owner* and a confirmation; the owner's line says how to leave; an admin can take the owner off
- **the presence list is a popover from 640 up and a bottom sheet below**, in the cluster's order, with names. That supersedes ADR-021 §5's placement of it only
- two contradictions in the file fixed: Product Team is 11 people and 6 boards everywhere, and you own Q3 Launch Plan
- 34 parts (`Person row`, `List`), six `Glyph` looks, six `Menu` looks, six confirmation wordings, ten new `Overlays` boards

**Code consequence:** no CSS variable. One `PersonRow`; `FormFrame` holds the lists; the presence list is shadcn's `Popover`; `MEMBER_ADDED`, `MEMBER_REMOVED` and `OWNERSHIP_TRANSFERRED` patch the board's roster (frontend-architecture.md). **Two backend items added**, 4 and 5 below, the SRS edited first. **Found while drawing:** a theme switch copies a part's text size from its main onto every copy that has its own words, and a copy gets its own size back only when the editor draws its page. Make the switch while the tab is awake and the page open, and export only after that. A flex row left laid out for the old size is fixed by hiding one child and showing it again, in two calls. ADR-018, -021, -031, -033, -040, -041 and -042 annotated.

**Stage 5's group e is settled and drawn, and Stage 5 is closed: the sidebar sliding over and the toast — [ADR-044](docs/architecture/adr/adr-044-the-sidebar-sliding-over-and-the-toast.md)** (2026-11-24). Question 4 fired twice, on the toast's edge and on the stack's gap, both resolved by rule (21 of 22 have fired), and fifteen choices. One rendered page; the user took fourteen recommendations and chose a stack of up to three.
- **the slide-over** is the `Sidebar` part over `color.scrim`, drawn on the phone, closed and with the switcher open; focus opens on the current row. Below 640 the user menu's bottom sheet opens over it, over a second dimming, which clarifies ADR-034 §4's *never stacks*
- **a toast sits 16 from the bottom and right** (`spacing.screen-x`, by rule), across the bottom on a phone, over the drawer's foot when it is open. **Up to three stack**, newest at the bottom, 8 apart (the user's call). **No timer**: it stays until closed, pushed out or the page changes, so WCAG 2.2.1 never applies
- **Undo is the Outline button**, hidden in the `Toast` part's main. It takes you back to the board, and **2.12 brings teammates back on `BOARD_UNARCHIVED`**, since archiving evicts no one (narrows ADR-040 §4; best effort). US-12c and SRS UC-12c were edited first; no backend change
- **a failure reads *Couldn’t move “X”.* and one of five reasons**; after a 403 or 404 the board reloads, so a demoted admin sees 2.7's banner. State 32 reads *“X” was deleted.*
- 222 tokens and 34 parts, none added; ten new `Overlays` boards (47)

**Code consequence:** no CSS variable. Sonner at the bottom right with offsets of `--screen-x`, `gap={8}`, `visibleToasts={3}`, `expand` and `duration={Infinity}`; one error helper for board mutations that reloads after a 403 or 404; 2.12 keeps the board's room joined (frontend-architecture.md). **Found while drawing:** a theme switch on `Overlays` (47 screens) blocks the tab for minutes and drops the plugin's heartbeat, but lands; return from the call at once and read the result later. Moving grid tiles by `layoutCell` can leave one drawn in its old cell; re-append it with `grid.appendChild`. ADR-018, -021, -030, -034, -035, -038, -040, -041 and -042 annotated.

**The design file's screens and overlays are regrouped into three pages by area — [ADR-045](docs/architecture/adr/adr-045-design-file-pages-by-area.md)** (2026-11-25, **supersedes ADR-040 §1's `Overlays` page only**). The user's worry was that developers wouldn't know which flow to implement, with a screen on `Page 1` and what opens over it on `Overlays`. The user chose to split by area over one merged page and over adding a map.
- **Pages: Foundations, Components, Account, Workspaces, Board.** `Page 1` was renamed Workspaces and `Overlays` Board, so 31 of 73 boards moved. Account holds 2.1, 2.2, 2.13 and *2.4 + user menu* (4 boards), Workspaces holds 2.3 and 2.4 with everything over them (24), and Board holds 2.5 to 2.15 with everything over them (45)
- **one row per screen**: at rest, its states, what opens over it, its phone versions last, under a `Doc / Row label` board (`color.background`, `type.heading`), since a bare text vanishes in Dark. 2.7 has its own row
- the phone's user menu stays on Board, in the 2.15 sidebar sequence, and a note on Account points to it. **No board renamed**
- moved by four ⌘C/⌘V rounds after the plugin set `penpot.selection`. Every copy matched its original in shapes, part links and token bindings before the original was deleted

**Code consequence:** none. **Found while moving:** the board at each page's origin is now a row label, so `instance()` lands inside it. The plugin's `storage` can be lost on a reconnect. ADR-040 annotated; the other ADRs' `Page 1` and `Overlays` positions are history, and inventory §7.2 has today's.

**Five backend items are open.** All five reopen a backend otherwise complete — small, but say so rather than discover it. Do them together in one backend pass, on their own branch, before the frontend reaches the screens that need them.
1. `GET /workspace/:id/board` needs an `archived` query parameter: handler → usecase → a parameterised predicate in `getUserBoardsInWorkspace`, which currently hardcodes `AND b.is_archived = FALSE`. SRS UC-08 documents it, and ADR-021 cannot deliver it in docs.
2. **Password rules on the server** (ADR-032, SRS UC-01 and §9 P1). `RegisterRequest.Password` is `min=8` only. Add the four checks, matching the browser's definitions, and return 400 rather than 500 above bcrypt's 72 bytes.
3. **A live message when a workspace is deleted** (ADR-038 §11). `DeleteWorkspace` deletes and returns, so anyone with one of its boards open learns of it only at their next request. Collect the board IDs with `GetBoardIDsByWorkspace` before the delete, then message each board's room after commit, as the remove-member fan-out in `internal/usecase/workspace/util.go` already does. **SRS first**: decide between a new message type and a new `ACCESS_REVOKED` reason. Revoked access navigates you out with a toast, but this case is closer to *gone* (2.12). Then add 2.12's third wording, *This board's workspace was deleted.*, to the design file.
4. **Refuse removing the workspace owner** (ADR-043 §8, SRS UC-06). `internal/usecase/workspace/remove_member.go` has no owner guard, so an admin can remove the owner and leave the workspace with nobody who can delete it. Add `ErrCannotRemoveOwner` beside `ErrCannotDemoteOwner` in `internal/domain/sentinel_errors.go`, return it when the target is the owner, and map it to 403 `FORBIDDEN` in `domain_mapper.go`, as demoting the owner already is.
5. **Leaving a board sends `MEMBER_REMOVED`** (ADR-043 §14, SRS UC-10, §4.5, §5). `LeaveBoard` sends only `USER_LEFT`, which is presence, so an open members list keeps showing someone who left. After commit, send `MEMBER_REMOVED` to the room as the remove-member path does, then the `CARD_UPDATED`s for their unassigned cards.

Next: **two independent roots, no blocking edge between them; the second is now finished.** ① **Scaffolding** — the Vite app zone before the Astro landing; ADR-018 §*Downstream edits* item 6 is the checklist (override shadcn's `disabled:opacity-50`, remove `animate-pulse` from shadcn's Skeleton — ADR-029 §6, re-point shadcn Sonner's `--normal-*` variables to `--toast` / `--toast-foreground` — ADR-030 §8, set `timeout: 10000` on axios, confirm Phosphor's import path and default status, confirm the Tailwind major version, confirm the self-hosted Poppins carries `tnum` — ADR-026 §6). ② The **Penpot design file** per ADR-020 §§2–6 — [`docs/architecture/design-file-inventory.md`](docs/architecture/design-file-inventory.md) is the authoring checklist, and **[§7](docs/architecture/design-file-inventory.md#7-the-order-to-draw-in-and-what-is-already-drawn) is where a drawing session starts**: the six-stage order (proof run first — two parts, one screen, one full check, before the other seven parts), what is already drawn, and a paste-in query that reports the file's real state instead of trusting a list. **Stages 1 to 5 are done** (2026-11-24): all fifteen screens are drawn (the last two with ADR-039) and every overlay (ADR-040 to ADR-044). **Stage 6, the final check, is closed** (2026-11-27, [ADR-046](docs/architecture/adr/adr-046-stage-6-final-check.md)): the design file is finished and left in Light + Fine. Keep the Penpot window beside VS Code, where part of it stays visible, because a covered window sleeps. A screen and what opens over it share a row on their area's page (ADR-045 §3, §8). To copy a board to another page, have the plugin select it, then ask the user to press ⌘C, open the page and press ⌘V. **If the file is touched again:** read a page only after opening it, open `Board` by hand, and read ADR-046's *Operational note* first. **The design root is complete, so ① Scaffolding is what remains** (with the five backend items, done in one pass before the frontend reaches the screens that need them). The `archived` parameter is only needed before the Archived list is *built*, the password rules before Register is *built*, and the workspace-deletion message before the board page is *built*, not before any of them is drawn.

**Don't touch:** Phase-2 deferrals — comments, labels, attachments, public boards, refresh tokens, account deletion.

Per-step history, decisions and dates: `docs/build-log.md`.

## Core Concepts

- Hierarchy: **Workspace → Board → Column → Card**
- **Workspace Admin authority outranks Board authority.**
- Board visibility: `workspace-visible` or `private`.
- Private boards are visible to workspace admins but require **joining before opening contents**.
- Realtime: board changes (card create/update/delete/move, column reorder, member join/leave) are broadcast to connected clients. Clients are optimistic; **the server is the source of truth.**



## Tech Stack

- Backend: Go, Gin, PostgreSQL (pgx), Google Wire, JWT, WebSocket
- Frontend (app zone): React + Vite SPA, TypeScript, React Router v7, TanStack Query, Zustand, axios, shadcn/ui, Tailwind, dnd-kit
- Frontend (public zone): Astro (SSG) — landing only; React islands where interactivity is needed



## Docs

- Build log — `docs/build-log.md` (what's been built, in build order, with links to each ADR/handoff)
- PRD — `docs/product/001-PRD.md`
- User stories — `docs/product/001-user-stories.md`
- Software spec — `docs/spesifications/001-software-specifications.md`
- Backend technical — `collabotask-backend/README.md` (conventions + "add an endpoint" recipe), `collabotask-backend/TESTING.md` (test approach)
- Architecture — `docs/architecture/backend-architecture.md` (backend), `docs/architecture/frontend-architecture.md` (frontend), `docs/architecture/adr/`
- Design file inventory — `docs/architecture/design-file-inventory.md` (what to draw in Penpot: the 15 whole-screen boards, ~16 overlays, a 39-state coverage check — derived from ADR-017/018/019/020/021, decides nothing. Its §6 is closed; the answers live in ADR-021)



### Updating docs (flow downstream: PRD → user-stories → SRS → code)

- `## Now` **above** — every time I switch focus (the only frequently-edited thing). Keep it a pointer, not a log.
- `docs/build-log.md` — append an entry when a build-order step completes; move the detail out of `## Now`.
- `CONTEXT.md` — only when a domain *term* is added/renamed/retired. Not for features.
- **PRD / user-stories** — scope or product-decision changes (edit first, upstream).
- **SRS** — technical-contract changes (API, schema, permission rules, closing §9 audit items).
- **ADRs** — never edit an existing one; write a new ADR that supersedes it.



## Working With Me

- **Don't scan the repo.** For a feature, read only the relevant user story + spec section.
- **Non-trivial change → I plan first and wait for your approval before coding.** One-liners I just do.
- **If a request conflicts with the spec or architecture, I flag it before building** — not after.
- Specific decisions (e.g. "admins bypass the member check") go to memory, not this file.
- **Task tracking is doc-driven + just-in-time:** spec §9.1 is the roadmap, `## Now` is the pointer. I break a step into a TodoWrite checklist only when you start it — no GitHub issues, no separate task files.



## Agent skills



### Issue tracker

Issues live in GitHub Issues; use the `gh` CLI. External PRs are not a triage surface. See `docs/agents/issue-tracker.md`.

### Triage labels

Default label vocabulary — no custom mappings. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context layout: `CONTEXT.md` at repo root + `docs/architecture/adr/`. See `docs/agents/domain.md`.