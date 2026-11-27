# ADR-047: Illustrations Are 192 Wide, Not 128

- **Status:** Accepted
- **Date:** 2026-11-27
- **Scope:** The size of the empty-state and failure drawings, and what has to follow from it. It settles:
  - `size.illustration` goes from 128 to **192**
  - the drawing's shapes **scale with the box**, so the token alone drives the size
  - the line stays 2px
  - the part's own board in the file is laid out for the new size

  **No token is added (222), no part (34), and no board is added or renamed.** One token's value changes. It **supersedes [ADR-033](adr-033-workspace-screens.md) §8's size and §9 rule 1's box only**: the colours, the line, the rules for what to draw, and the one-line empty states are unchanged.
- **Upstream edits:** none. The PRD, the user stories and the SRS give no drawing a size.

## Context

[ADR-033](adr-033-workspace-screens.md) §8 fixed `size.illustration` at 128 *at both densities* and drew three subjects at that size. [ADR-038](adr-038-board-states-group-d.md) added five more. **After Stage 6** ([ADR-046](adr-046-stage-6-final-check.md)) the user tried to change the token from 128 to 256 in Penpot. **Only the wrapper grew.** The drawing did not.

That was a defect in the part, not in the token. Each look is a square board holding four to eight loose vector shapes, and every shape was pinned to the board's top-left corner (Penpot's default constraint). Resizing the board moved nothing inside it.

The user asked to see it larger, on one screen first. Two test copies of *2.4 Workspace detail — empty* were made, the drawing's shapes set to **Scale**, and the board resized to 192 and to 256. The user looked at both and chose 192.

## Decision

### 1. `size.illustration` is 192

The token's value in `base` is 192. Both density sets inherit it, as before: **a drawing is not a touch target**, so it does not change with the pointer.

The three things that came with it, all measured on the file:
- **Every drawing grows by exactly 1.5.** The *No boards* frame went from 96 × 72 to 144 × 108; every other shape followed.
- **The block grows by 64px.** The drawing is 64 taller and the 16px gap to the sentence is unchanged, so each empty stack grew from about 212 to 276. All nine screens that show a drawing still fit inside their screen, with the stack's bottom edge between 416 and 594 of 900.
- **The line does not scale.** All 41 strokes are still 2px, bound to `border.width-illustration`. At 192 the drawing reads a little lighter than at 128, which the user judged acceptable.

### 2. The part's shapes scale with its box

In every look of `Illustration`, each shape's horizontal and vertical constraint is **Scale**. A change to `size.illustration` now resizes the board and the drawing inside it together. A copy on a screen follows its main.

### 3. The part's own board is re-laid out

`Illustration — states` was drawn for 128: its looks sat 200 apart. At 192 that left 8px between drawings and put them past the board's bottom edge.

- the eight looks are **260 apart**, and the board is 2184 × 314 (it was 1700 × 250), with its inner frame 2072 × 252
- the three captions moved with their looks
- the **19 boards below it on `Components` moved down 64**, so the usual 100px between boards holds

### 4. The rules for drawing the next illustration

[ADR-033](adr-033-workspace-screens.md) §9 stands, with these changes:

- **Rule 1, the box:** 192 × 192. Draw on the 128 grid as before, or at 192; either way the shapes of a new `Subject` get **Scale** on both axes the moment they are drawn, or the token will not move them.
- **Rule 6, accents:** a sparkle is **12 to 18** long, not 8 to 12, at the new size. The clear edge of the box is 12.
- **Rule 2, the line, and the colours (§8):** unchanged. The line is 2px.

### 5. The code

- §9's `--illustration-size` becomes **`192px`**.
- **The line must stay 2px in the browser too.** An inline SVG drawn on a 128 grid and shown at 192 scales its strokes to 3px. Either **export the drawing at 192** from the file, whose strokes are already 2px, or keep the 128 grid and put `vector-effect: non-scaling-stroke` on every stroke. Either gives the lighter line the user chose. §9 rule 10's `stroke-width="2"` holds in both.
- The empty-state block is the drawing, `--illustration-gap` (16) and the sentence, centred. The block is 64px taller. On a phone it is 192 wide in a 390 screen, with `--screen-x` (16) either side, so it fits. **No phone screen in the file shows a drawing**, so that was not drawn.

## Options considered

- **Keep 128.** The size the screens were laid out around. The drawings read as small marks on a 1440 screen.
- **192.** **Chosen by the user**, after looking at the real screen at 1.5 ×.
- **256.** Tried on a test copy. The drawing becomes the main thing on the screen and the 2px line reads like a sketch. It would also be two thirds of a phone's width. Not chosen.
- **Scale the line with the drawing.** Would keep the weight but change `border.width-illustration` and every stroke bound to it. Not done: the user did not ask for it, and 192 looked right.

## Out of scope (deferred, recorded)

- **2.13 *Session expired*** is left-aligned with the form, as [ADR-038](adr-038-board-states-group-d.md) set. The hourglass is drawn in the middle of its box, so it now sits about 48px in from the text's left edge, where it was about 31. It reads as slightly indented. Not changed; it is the user's left-aligned choice.
- **Flavours 5 and 6**, an empty column and the add-people panel, keep one line of text. A 192 drawing fits them even less than a 128 one ([ADR-033](adr-033-workspace-screens.md) §8).
- **A pre-existing overlap on `Components`**, found while moving the boards: `Confirmation — states` (made taller on 2026-11-24, when six wordings were added) runs 168px into `Picker row — states`, at the same x. It is older than this ADR and untouched.

## Downstream edits this ADR requires

- **Done in the file:** the token, the constraints on all eight looks, the part's board, the 19 boards moved, and the Foundations catalogue's `size.illustration` value (192, bindings intact).
- **Inventory** §7.2: the `Illustration` row, and the Foundations row. **ADR-033 §8 and §9** carry a note pointing here. They are annotated, not edited.
- `CLAUDE.md` `## Now` and `docs/build-log.md`.
- **Frontend:** `--illustration-size: 192px`, and the stroke rule in §5, at the scaffold. [ADR-033](adr-033-workspace-screens.md) §14 lists the variable at 128 and is not edited.

## Operational note

Found while making the change:

- **Setting a token's value from the plugin does not resize the shapes bound to it.** `token.value = "192"` changed the value, and `resolvedValue` read 192, while every bound board stayed 128. Changing it in Penpot's panel does re-apply it. From the plugin, re-apply the token to each bound shape.
- **Applying a token that is already applied removes it**, like a toggle. The first `applyToken` on a bound board left it unbound and at 128. The second bound it and resized it. Read `tokens` after every apply, and apply once more if it came back empty.
- **A token change across the file freezes the tab for about a minute.** It lands anyway. Read the result in a later call.
- **Resizing a board only moves its children if they are set to Scale.** Penpot's default for a child is Left / Top. Check `constraintsHorizontal` and `constraintsVertical` on any part meant to be resized by a token.
- **To try a size on one screen, clone the screen and detach the copy.** The constraint lives in the part, so a copy cannot be set up on its own while it is linked. Delete the clone afterwards.
- **A shape object read twice is not the same JavaScript object.** `indexOf` on `children` returns -1. Compare by `id`.
