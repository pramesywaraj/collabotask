# ADR-022: Form-Control Geometry

- **Status:** Accepted — **§3's *no `spacing.control-padding-y`* narrowed by [ADR-042](adr-042-the-forms.md)** (2026-11-24): a field that grows with its text has no height to centre in, so the token exists, 8 at Fine and 12 at Coarse, as (`size.control-height` − 20) ÷ 2. A control of fixed height still centres its one line and doesn't use it. The visibility choice rows use it too. Everything else stands as written.
- **Date:** 2026-11-20
- **Scope:** The size of form controls in the app zone — control height, horizontal padding, the gap between stacked fields, and the width of a form container. **Extends [ADR-017](adr-017-ui-foundations.md) §9; supersedes nothing.** §9 was *silent* on control geometry, not wrong about it. Every value ADR-017 fixes is carried unchanged, and nothing in ADR-018, ADR-019, ADR-020 or ADR-021 moves. No code is written here.

## Context

**Found before a single part was drawn.** [`design-file-inventory.md`](../design-file-inventory.md) §7.3 opens its Stage 2 with a question no script can answer — *"did you have to type a number or a colour because no shared value existed?"* — and the answer came back *yes* when the live token catalogue was checked against board 2.1's requirements on 2026-11-20.

**Why the gap exists, and why it is nobody's mistake.** [ADR-020](adr-020-design-tool-penpot.md) *Out of scope* makes ADR-017 §9's token block "the authority for implementation values", and the 103 tokens in the Penpot file `Collabotask App` were derived from it in full. But §9's vocabulary is the **board** — cards, columns, canvas, avatars, meta rows. It is a complete account of the screen the product is built around.

Login and Register are **forms**, and no form geometry was ever needed to describe a board.

What ADR-017 *does* fix about a control:

| Fixed | Where |
|---|---|
| Text at 14/20 — "card titles, body, column headers, **buttons, inputs**" | §6 |
| Radius 6px — "buttons, inputs, menu items" | §5 |
| The `--input` boundary at stone-500 `#78716C`, chosen as the lightest stone reaching 3:1 on white | §3.4 |

What it leaves open: **how tall the control is, how much room the text has beside it, how far apart two stacked fields sit, and how wide the form is.**

`size.target-min` constrains this without answering it. At Coarse it is 44, so no control there may be shorter. At Fine it is 24 — a height no control carrying a 20px line can be.

**Why this is an ADR and not a note.** ADR-017's own Context already set this project's test, about this exact class of value:

> **Components bake tokens in.** shadcn/ui is vendored into the repo, not installed — once twenty components reference `--primary`, changing its *role* (not just its value) is a sweep. This is a durable, hard-to-reverse decision, which is why it is an ADR rather than a token file.

Control height will be referenced by every form control in the app. Changing it after twenty components do is the same sweep, so it goes where §9 went.

---

## Decision

### 1. Four tokens

| Token | Set | Fine | Coarse |
|---|---|---|---|
| `size.control-height` | both density sets | **36** | **44** |
| `spacing.control-padding-x` | `base` | **12** | 12 |
| `spacing.field-gap` | both density sets | **16** | **20** |
| `size.form-width` | `base` | **400** | 400 |

### 2. Where each value comes from

Two of the four derive from ladders already in the file. Two are choices. The difference is recorded because a later reader should know which ones are safe to re-open.

**Derived:**

- **`size.control-height` 36 / 44 is `size.column-header`'s existing ladder, unchanged.** A column header is already a bounded interactive row carrying 14/20 text with a `⋯` control in it — the same species of object as an input. Reusing its ladder adds no new number to the system.
- **`spacing.control-padding-x` 12 is `spacing.card-padding-x`**, which is 12 at both densities.

The ladder is self-checking: 36 − 20 = 16, so a centred 14/20 line leaves 8 above and below, which is `spacing.card-padding-y` at Fine. 44 − 20 = 24, leaving 12, which is `spacing.card-padding-y` at Coarse. The height ladder and the padding ladder agree without being tuned to.

Both heights clear their floor the way ADR-017 says cards do — by construction rather than by special-casing. 36 clears Fine's 24; 44 **is** Coarse's 44.

**Chosen:**

- **`spacing.field-gap` 16 / 20.** It must exceed `spacing.card-gap` (8 / 12): a form field is a taller, heavier object than a card and needs more air between instances. 16 and 20 keep to the 4px grid every existing value in the catalogue sits on.
- **`size.form-width` 400.** The one value with no constraint behind it. It partners the phone strip's `min(100vw − 72px, 420px)` ceiling from [ADR-019](adr-019-responsive-touch-scope.md) §3 without exceeding it.

### 3. Vertical padding is not a token

**There is deliberately no `spacing.control-padding-y`.** A control with a fixed height and a centred single line does not need one — the vertical space is a consequence of the height, not an independent value. In Penpot it is a flex layout with `alignItems: "center"`; in code it is the `h-9` / `items-center` pattern shadcn already uses for Button and Input.

> **Narrowed by [ADR-042](adr-042-the-forms.md)** (2026-11-24). This holds for a control of fixed height. A field of several lines grows with its text, so it has no height to centre in: its text sits `spacing.control-padding-y` from the top and bottom, 8 at Fine and 12 at Coarse, so its first line lands where a one-line field's does.

Adding the token would create a second way to express the same measurement, and the two would eventually disagree. Same reasoning as ADR-017 §2's refusal of a `--column-header` colour: a value that is not separable should not be named.

### 4. Density values are written into both density sets, as literals

The two density-dependent tokens are **written twice, once into `density-fine` and once into `density-coarse`, each with its own literal number.** They are not expressed as references to the card ladder.

**The reference mechanism does work** — verified on 2026-11-20 against the live file. `size.title-reserve` is `{spacing.title-right-padding}` in both density sets, and toggling the active set moved it from 24 to 32 and back, so a reference resolves per-density exactly as [ADR-020](adr-020-design-tool-penpot.md) §2's flip requires.

It is still the wrong tool here. `spacing.control-padding-x` equals `spacing.card-padding-x` **arithmetically, not semantically**: the control value comes out of the height ladder, and the card value out of ADR-017 §9. If card padding ever moves to 10, control padding must *not* follow, or the height ladder breaks. [ADR-020](adr-020-design-tool-penpot.md) §3 reserves references for "a shared value with a distinct role" — `--board-column` carrying `--muted`'s value. This is a coincidence of two ladders, which is a different thing, and encoding it as a reference would assert a relationship that does not exist.

---

## Options Considered

### Add the values to ADR-017 §9 directly

Rejected, though it is the most natural-looking home and the one that raised this question.

Two concrete problems. **Later ADRs cite §9's contents by reference:** ADR-020 §2 maps "ADR-017 §9's `:root` colour values" into `theme-light`, ADR-018 §11 adds four dark state colours to that set, and the 103 Penpot tokens were derived from §9 as it stood. If §9 grows silently, no reader can tell which version those citations saw — which is not hypothetical, since answering that question on 2026-11-20 required querying Penpot rather than reading a document. **And the dates would lie:** ADR-017 is dated 2026-11-13, and a decision taken a week later does not belong inside it.

ADR-017 is instead **annotated**, which is this project's established practice — its Status line already records §8's supersession by ADR-019 and two ADR-018 changes. See *Downstream edits*.

### Record them in `design-file-inventory.md`

Rejected on the inventory's own terms. Its header states it **decides nothing** and that every row derives from an accepted ADR, a user story or the route tree. That promise is the only reason the document can be trusted as a checklist, and a decision living there would void it.

### A token file in the repo

Rejected by ADR-017's own argument, quoted in *Context*. There is also no scaffold yet to hold one.

### Reuse `size.target-min` as the control height

Rejected on meaning. It is a **floor**, introduced by [ADR-018](adr-018-state-vocabulary.md) §6 to satisfy WCAG 2.5.8, not a size. At Coarse it coincides with the answer; at Fine it is 24, which cannot hold a 20px line with any padding at all. A floor that happens to equal a height in one of two densities is not the same value.

### Express vertical padding as `{spacing.card-padding-y}`

Rejected after verifying that it would work. See *Decision* §4 — the mechanism is sound, the relationship is not.

---

## Out of scope (deferred, recorded)

- **Every control geometry already fixed elsewhere.** The column header's 36 / 44, the card's padding and gap ladders, the 56px app bar and the 280px column are untouched; this ADR adds the form case and changes nothing it finds.
- **Toast, menu and overlay geometry.** The toast's own design is already deferred by ADR-018, and overlay presentation is settled as four rules by [ADR-021](adr-021-design-inventory-corrections.md) §5. Neither needs a size token to be drawn.

  > **The toast did need two.** Drawing it on 2026-11-20 needed a width and a padding, and typing either is what question 4 exists to catch. **[ADR-030](adr-030-toast.md)** adds `size.toast-width` (358) and `spacing.toast-padding` (12). Overlays and menus are unaffected by it.
- **Motion and the identity palette.** Unchanged, and still not tokens — ADR-020 §3.
- **The Astro landing page.** Outside the app zone's design file entirely.
- **Phase-2 deferrals** — comments, labels, attachments, public boards, refresh tokens, account deletion.

## Downstream edits this ADR requires

1. **[ADR-017](adr-017-ui-foundations.md) — two annotations, no rewrite.** Its Status line gains `§9 extended by ADR-022 (2026-11-20) — form-control geometry, which §9 did not cover.`, and §9's head gains one line pointing here. This matches how §8's supersession and ADR-018's two changes are already recorded. **No decision text in ADR-017 is altered.**
2. **Penpot `Collabotask App`** — add the four tokens: two into `base`, two into each density set. Then re-run the flip to confirm 36 → 44 and 16 → 20.
3. **[`design-file-inventory.md`](../design-file-inventory.md)** — §7.2's "form-control geometry is missing" row and §7.3's Stage 2 gap block both close, and point here instead. Stage 2 returns to being a drawing task.
4. **[`docs/build-log.md`](../../build-log.md)** — **no entry now.** [ADR-020](adr-020-design-tool-penpot.md) already defers the design-file entry until the file is authored, and [ADR-021](adr-021-design-inventory-corrections.md) did the same; neither is in the log. This ADR follows that, and will be named in that single entry rather than starting a parallel log of decisions.
5. **[ADR-018](adr-018-state-vocabulary.md) — no edit.** Its §*Downstream edits* item 6 holds the scaffold-time checklist, and these four values will need to reach Tailwind's config alongside it. That consequence is recorded **here** rather than added there, because this project does not edit an accepted ADR — the same course [ADR-020](adr-020-design-tool-penpot.md) took with the same checklist.
6. **SRS** — no change. Frontend presentation end to end, which is the conclusion ADR-018, ADR-019 and ADR-020 each reached about their own content.
7. **`CONTEXT.md`** — no change. "Control", "field" and "form" are interface vocabulary, not domain terms, and nothing in Workspace / Board / Column / Card / Position / the role model moves.

## Operational note

**`resolvedValue` is global to the active set configuration, not per-set.** Reading a token object fetched from `density-fine` returned `32` while `density-coarse` was the active set. Only `value` is per-set and trustworthy. So a density value cannot be verified by reading the set that holds it — the set has to be made active first, and the previous active state restored afterwards.

This compounds the warning ADR-020 already records: a `no heartbeat` error does not mean the code did not run. Toggling active sets is a write. Inspect state before re-running.
