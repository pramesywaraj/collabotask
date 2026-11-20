# ADR-020: The Design File Lives in Penpot

- **Status:** Accepted — **§3's type mapping extended by [ADR-023](adr-023-spacing-between-objects.md)** (2026-11-20): one `borderWidth` row, for the 1px edge and the 2px ring and placeholder. Everything else stands as written.
- **Date:** 2026-11-18
- **Scope:** Which tool holds the app zone's design file, and how [ADR-018](adr-018-state-vocabulary.md) §12's structure is expressed in it. **Supersedes [ADR-018](adr-018-state-vocabulary.md) §12 only.** Everything else in ADR-018 stands unchanged — the 39-state inventory, the channel ledger, the in-flight rule, drag, the three disabled states, affordance, loading/empty/gone, motion, icons, connection and permission states, and §11's token additions. This ADR changes the *medium*, not one value of the design. ADR-017's tokens and ADR-019's density scale are untouched. No code or token file is written here; this is a tooling and structure decision taken before the file is authored.

## Context

ADR-018 §12 specified the design file as **two variable collections with two modes each** — colour Light/Dark ([ADR-017](adr-017-ui-foundations.md) §7) and density Fine/Coarse ([ADR-019](adr-019-responsive-touch-scope.md) §7) — with every state a component variant and no typed-in numbers. Its reasoning was that two two-mode collections give all four combinations, where one four-mode collection would force every colour and every size to be written twice over.

That reasoning is sound and this ADR keeps it. What failed was the assumption that Figma could host it.

**Figma gates multiple modes per variable collection behind paid plans.** The account's two teams — *Team Panalaaar* and *Proyekan* — both report `tier: "starter"` from the Figma MCP's `whoami`. A starter-tier collection gets exactly one mode. §12's entire structure is two-mode collections, so on this account §12 is not buildable in Figma at all. Paying is out of scope: the project will not spend on design tooling.

Three tools were assessed against §12's requirements.

1. **`figma-bridge`** (a local plugin bridge, already connected) can create frames, text, shapes, SVG, auto-layout and shadows — but every write takes a literal hex or number. It has no variable creation, no token binding, no component or variant-set creation, and no instance placement. It can draw something that *looks* like the design while failing the only check §12 cares about.
2. **The official Figma MCP** (`use_figma`, arbitrary JS against the plugin API) can do all of §12 except modes, which the tier forbids. §12 minus its flip.
3. **Penpot** (`execute_code`, arbitrary JS against the Penpot plugin API) clears every requirement, free.

An intermediate position was held briefly and is recorded in *Options Considered*: that losing the mode flip was acceptable because ADR-019 §7 is already a complete, reviewed, version-controlled table. That argument was about **cost** — a flip is not worth a subscription. Penpot reduces the cost to zero, and a free check beats no check.

**The flip is verified, not assumed.** Two Penpot token sets each holding a token named `card.padding` — `8` in `density-fine`, `12` in `density-coarse` — with one flex board bound to all four paddings via `applyToken`. Toggling the active set moved the same board from `[8,8,8,8]` to `[12,12,12,12]`, with `shape.tokens` still reporting all four properties bound to `card.padding` and the board never touched after binding. One component carries both densities. §12's intent survives intact.

---

## Decision

### 1. Penpot holds the design file

The file is **`Collabotask App`**, authored through the Penpot MCP's `execute_code`. Figma is not used for the app zone's design file, and `figma-bridge` is retired from authoring — the official Figma MCP supersedes it on capability, and Penpot supersedes both on fit.

This decision is about the design file only. It implies nothing about the implementation, which remains React + Vite + Tailwind + shadcn/ui per ADR-011 and ADR-017.

### 2. Sets and themes replace collections and modes

Penpot's structure is not Figma's, and the mapping has to be stated or it will be re-derived wrongly. A Penpot **token set** is a named collection of uniquely-named tokens that is either active or inactive; a **theme** is a preset that activates a chosen combination of sets. Resolution follows the active set, so the same token name in two sets resolves differently depending on which is active. That is the mode mechanism.

Crucially, Penpot themes carry a **group**, which gives independent axes natively:

| Set | Holds |
|---|---|
| `base` | Everything neither theme- nor density-dependent: radius, the type scale, `--in-flight-opacity` |
| `theme-light` | ADR-017 §9's `:root` colour values |
| `theme-dark` | ADR-017 §9's `.dark` colour values, plus ADR-018 §11's four dark state colours |
| `density-fine` | ADR-019 §7's fine column, plus ADR-018 §6's two fine values (minimum target 24px, title right padding 24px) |
| `density-coarse` | ADR-019 §7's coarse column |

| Theme group | Names |
|---|---|
| `Theme` | `Light`, `Dark` |
| `Density` | `Fine`, `Coarse` |

Two groups of two give all four combinations without writing any value twice — which is **§12's own argument, carried over and vindicated by a different mechanism than the one it was written for.**

`base` is always active. Exactly one set from each axis is active at a time.

### 3. Token type mapping

ADR-017 §9 and ADR-018 §11 are CSS custom properties. Penpot's `TokenType` is a closed list, so the mapping is fixed here:

| Source | Penpot `TokenType` | Applied via `TokenProperty` |
|---|---|---|
| All colour values (`--background`, `--primary`, `--drag-over`, …) | `color` | `fill`, `strokeColor` |
| `--radius` | `borderRadius` | `borderRadius*` |
| `--shadow-overlay`, `--shadow-dragged` | `shadow` | `shadow` |
| `--in-flight-opacity` | `opacity` | `opacity` |
| `--target-min`, avatar size, meta-row height, column-header height | `sizing` | `width`, `height`, `layoutItemMinH` |
| Card padding, card gap, title right padding | `spacing` | `padding*`, `rowGap`, `columnGap` |
| ADR-017 §6's five size/line-height pairs | `typography` | `typography` |
| Poppins | `fontFamilies`, `fontWeights` | `fontFamilies`, `fontWeight` |

Two things deliberately do **not** become tokens:

- **Motion.** ADR-018 §11's four durations and two easing curves have no corresponding Penpot token type — the list has no duration or easing. Motion stays a code-only concern. This costs nothing: ADR-018 §8 re-pointed motion at "a teammate did this", which is a realtime multi-client behaviour no static design file could show regardless of tool.
- **The identity palette.** Unchanged from ADR-017 §9 — a lookup indexed by `hash(id) % 8`, living in a TypeScript constant with both role assignments, not in the token catalogue.

Where ADR-017 §9 notes a shared value with a distinct role — `--board-column` carrying the same value as `--muted` — Penpot's token **references** (`"{color.muted}"`) express that directly, which is strictly better than the CSS duplication it replaces.

### 4. States stay variants, not frames

§12's rule holds: **every state is a component variant.** Penpot's mechanism is `penpotUtils.createVariantContainer`, which takes shapes with named property values and builds a variant container with a structured swap UI.

The card carries one `State` axis — `Rest`, `Hover`, `Pressed`, `InFlight`, `DragSource` — resolving against whichever density and theme are active. Five variants × two densities × two themes is **one** component, not twenty frames.

### 5. Whole-screen boards

Unchanged from §12 in count and in reasoning. Boards only where the whole screen differs: read-only, join-required, cold-load skeleton, empty board, load failed, gone, session expired, plus one drag board (a drag is four states at once).

**And the one 390×844 phone board.** §12's reason for it was never density — it was that *density is a variable and layout is a breakpoint*, and no mode can turn a fixed 280px column into ADR-019 §3's `min(100vw − 72px, 420px)` snap strip. That holds exactly as written. It also remains the only place ADR-019 §7's "6 cards visible at 390×844" is checked by eye, which ADR-019 itself flags as the one number in its table resting on an assumption.

### 6. Verification is automated, not eyeballed

§12 enforced "no typed-in numbers" by flipping modes and looking. Penpot allows better. `shape.tokens` returns a map from property name to token name, and `penpotUtils.analyzeDescendants` walks a board applying an evaluator to every descendant.

So the completeness check becomes a **predicate**: walk each board and flag any descendant whose padding, gap, fill, radius or size is not token-bound. A missing binding is reported by name rather than noticed by eye. `export_shape` (PNG or SVG) covers what genuinely needs looking at — the phone board, and the state boards as compositions.

---

## Options Considered

### Pay for Figma Professional

Rejected. It buys one check — mode-flip completeness over eight numbers that ADR-019 §7 already records in a reviewed, version-controlled table. The project will not spend on design tooling, and Penpot supplies the same check for nothing.

### Figma at starter tier, single mode, official MCP

Rejected once Penpot was verified, though it was the leading option while the flip was believed unavailable anywhere free. It builds §12's components, variants, auto-layout and variables correctly and only loses modes. Viable, strictly worse than Penpot, and the fallback if Penpot is ever abandoned.

### Accept the loss of the flip, on the grounds that ADR-019 §7 is already complete

Held briefly, then withdrawn. The argument was never that the check is worthless — only that it was not worth a subscription. Penpot moves the price to zero and the argument dissolves. Recorded because the reasoning was explicit and reversing it silently would hide why §12's scope was restored rather than cut.

### `figma-bridge` drawing the design

Rejected on capability, not preference. No token binding, no components, no variant sets, no instance placement; every fill a literal hex and every padding a typed-in number. That is §12's rejected "draw both densities" option arrived at by a different route — and §12's objection to it ("one hardcoded 8px padding and the mode flip silently stops being a check") applies with more force, since there would be no flip to stop.

### Penpot's themes as a single four-value axis

Rejected for exactly §12's reason. One axis of `Light-Fine / Light-Coarse / Dark-Fine / Dark-Coarse` writes every colour twice and every size twice. Two groups of two is the same decision §12 made, in Penpot's vocabulary.

---

## Out of scope (deferred, recorded)

- **The implementation.** This ADR decides where the design file lives, not how the app is built. React + Vite + Tailwind + shadcn/ui stands per ADR-011 and ADR-017, and the design file is a reference, not a source of generated code.
- **Design-to-code generation.** Penpot exposes `generateStyle` and `generateMarkup`, and the official Figma MCP has a mature design-to-code path. Neither is adopted here. ADR-017 §9's token block is already the authority for implementation values; generating a second set from the design file would create two sources of truth for values the ADRs already fix.
- **The Astro landing page.** Unchanged — responsive by default, visually permitted to diverge, and no part of the app zone's design file.
- **The scaffold-time checklist.** ADR-018 §*Downstream edits* item 6 is unaffected by this ADR and still applies to the Vite app: override shadcn's `disabled:opacity-50`, set `timeout: 10000` on axios, confirm Phosphor's import path and default status, and confirm the Tailwind major version.
- **Phase-2 deferrals** — comments, labels, attachments, public boards, refresh tokens, account deletion — untouched.

## Downstream edits this ADR requires

1. **`CLAUDE.md` › `## Now`** — record that the design file is Penpot (`Collabotask App`), that ADR-020 supersedes ADR-018 §12, and that scaffolding and the design file are independent roots with no blocking edge between them.
2. **`docs/build-log.md`** — an entry when the design file is authored, not now.
3. **ADR-018** — no edit. §12 is superseded by reference, per the project's rule that an existing ADR is never edited.
4. **ADR-017, ADR-019** — no change. Every value either ADR fixes is carried unchanged; only the container changes.
5. **SRS** — no change. This is a design-tooling decision, not a technical contract. Same conclusion ADR-018 and ADR-019 both reached about their own content.
6. **`CONTEXT.md`** — no change. "Token set", "theme" and "variant" are tool vocabulary, not domain terms, and nothing in Workspace / Board / Column / Card / Position / the role model moves. Worth confirming rather than adding: **the design file is not a domain artefact.** It records no state the backend knows about, and no board, column or card in it corresponds to a row anywhere.

## Operational note

A `no heartbeat` / `plugin tab appears to be suspended` error from the Penpot MCP **does not mean the code did not execute.** An attempt that returned that error had in fact created both token sets and applied both toggles. The stale-heartbeat counter also climbs monotonically across browser-side fixes, which makes a recoverable condition look permanent and server-side. Inspect state before re-running, and never assume a failed destructive call was a no-op.
