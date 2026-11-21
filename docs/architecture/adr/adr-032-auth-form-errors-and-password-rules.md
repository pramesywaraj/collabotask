# ADR-032: Where Login and Register Report Errors, and Register's Password Rules

- **Status:** Accepted
- **Date:** 2026-11-21
- **Scope:** Board 2.2 Register, and a second pass on 2.1 Login. It settles where every error on the two forms is reported, which [ADR-024](adr-024-control-and-field-internals.md) *Out of scope* left open for Login. It also records:
  - the two forms' wordings, including **Sign in** in place of *Log in*
  - Register's **four password rules**, a product decision taken against NIST's guidance, shown as a **live checklist**
  - a **confirm-password** field
  - fifteen behaviour rules
  - two new parts, `Form message` and `Password rule`
  - one new token, `spacing.list-gap`

  It **extends [ADR-017](adr-017-ui-foundations.md) §9**: the code gains `--list-gap`. It **supersedes two things and nothing else in them**:
  - [ADR-018](adr-018-state-vocabulary.md)'s channel ledger row *"Emerald | done / archived — no Phase-1 render site"*. Emerald now has one render site: a met password rule.
  - [ADR-024](adr-024-control-and-field-internals.md) §5's *"Register's password-minimum note uses the same step in `color.muted-foreground`"*. There is no single note now. The checklist (§4) replaces it.
- **Upstream edits** went first, as CLAUDE.md requires for a product change: US-01 and SRS UC-01 now carry the password rules and the confirm field. SRS §9 lists the server half as an open P1 item.

## Context

Question 4 was asked of 2.2 Register before drawing, in two rounds.

**Round 1** found no missing value on the recommended path. It was the first time in eleven stages. Every colour the two forms use clears its floor on `color.background`, the only surface either screen has. What was missing was not a number:

- **A home for a form-level error.** Neither inventory §1.1 nor the parts had one, and Login needs one. UC-02 returns the same `401 UNAUTHORIZED` for an unknown email and a wrong password, and US-02 forbids saying which. So that error can't sit under either field.
- **Any wording at all.** No document had error copy. The SRS's example 409 message, *"Email already registered"*, is not what the server sends. `domain.ErrEmailAlreadyExists` reads *"email already exists"*. Neither is meant for a person to read.
- **A way to point a 400 at a field.** `response/validation.go` returns `details` as a plain `[]string`, with no field key. Only the browser's own checks can point at a field.
- **One conflict, flagged and resolved.** The brief listed *email already taken* as a form-level error. Inventory §1.2 already puts it on the field, and §1.2 was kept (§2).

Two pairs measured close, and they shaped §4 and §6:

| Pair, Light | Ratio | Protanopia | Deuteranopia |
|---|---|---|---|
| Invalid edge `#E11D48` against the resting edge `#78716C` | **1.02** | 47 apart | 51 apart |
| Error text `#BE123C` against hint text `#57534E` | **1.21** | **21 apart** | 47 apart |

"Apart" is the sRGB distance between the two colours as a colour-blind reader sees them (Machado 2009). Under about 40, they read as the same colour. So in Light an invalid field differs from its neighbours by hue alone. A hint that turned into an error in the same place would change only its words for a protanope (red-blind reader).

The options were rendered side by side in Light and Dark with the file's values, and the user chose with a picker. The user took every recommendation and added three notes:
- *Sign in* in place of *Log in*
- *"Don't have an account yet? Create one"* on Login
- four password rules: at least 8 characters, an uppercase and a lowercase letter, a number, and a symbol

**Round 2** came out of the third note. The rules were checked against NIST SP 800-63B-4 (final, July 2025, §3.1.1.2). It says verifiers *"SHALL NOT impose other composition rules (e.g., requiring mixtures of different character types)"*. It requires two things instead: a 15-character minimum when the password is the only factor (8 only with multi-factor), and a comparison against *"a blocklist that contains known commonly used, expected, or compromised passwords"*. Three alternatives were offered, with NIST's 15 plus a blocklist recommended. **The user chose the four rules**, then a **live checklist** over a sentence, and **added a confirm-password field**.

The checklist needed one value the file did not have: the space between two rows. That is question 4's eleventh firing in eleven stages, in a second round. Four further choices came with it, decided on a second rendered page.

## Decision

### 1. A form-level message sits just above the button

A form-level failure is one line of `type.body` in `color.destructive-text`. It fills the form's width and sits **inside the `Fields` stack, between the last field and the button**. `spacing.field-gap` sits above it and below it, and the button moves down 36px at Fine when it appears. It is the new part `Form message` (§9).

- It sits where your eyes are when you press. On a phone with the keyboard up it stays in view, where a line under the heading can scroll away.
- It is `type.body`, not a field error's `type.meta`, so its size also says it belongs to the whole form.
- It carries Login's 401, and on both forms *can't reach the server* and *something went wrong*.
- It is announced as an alert (`role="alert"`).
- Stage 5's forms (Create workspace, Invite people, Add people) will reuse it for their own server errors.

### 2. A taken email is reported on the email field

As inventory §1.2 already said: a 409 points at one field, so it shows there, as the Input's `Invalid` look with *"An account with this email already exists."* The server's own text is never shown. After a 409, focus moves to the email field.

### 3. The four password rules — a product decision against NIST

Register requires all four:
- at least 8 characters
- an uppercase and a lowercase letter
- a number
- a symbol

The exact character definitions are in SRS UC-01, so that the browser and the server count the same thing. Spaces are allowed but don't count as a symbol. The rules apply **at registration only**. Login checks only that the password isn't empty, so a password set before the rules keeps working.

**This was recommended against, and it is recorded so nobody mistakes it for a default.**
- Composition rules are met predictably: *Password1!* passes all four, and it is among the first passwords an attacker tries.
- NIST forbids them, and asks for length plus a blocklist instead.
- The user weighed this and chose the rules for the security awareness they give at sign-up.
- If the policy is revisited, the cheapest real improvement is a blocklist check on the server. It adds one message and leaves this ADR's layout intact.

**Consequences:**
- **US-01 and SRS UC-01 now state the rules.** A rule only the browser checks protects nothing: the API is public, and anyone can call it without the form.
- **The server change is open.** It is SRS §9's P1 item and a second open backend item, next to UC-08's `archived` parameter.
- **Found in the same pass:** `bcrypt.GenerateFromPassword` refuses passwords over 72 bytes, and today that surfaces as a 500. The backend item settles an upper limit and returns 400 instead.

### 4. The rules are shown as a live checklist

The checklist **is** the hint inventory §1.2 asks for (*"say it before they submit"*). It is always shown, from the first visit, under the password field.

- **One row per rule**: a 16px glyph (`size.icon`) and the rule in `type.meta`, `spacing.icon-gap` apart. The row aligns to the top, so when the long second rule wraps, its glyph stays on the first line.
- **Three looks, one per state:**

  | State | Glyph (Phosphor) | Glyph and text | Light | Dark |
  |---|---|---|---|---|
  | `Unmet` | `Circle` | `color.muted-foreground` | 7.63 | 13.26 |
  | `Met` | `Check` | `color.done` (emerald) | 5.48 | 10.28 |
  | `Missed` | `X` | `color.destructive-text` | 6.29 | 7.34 |

  Each colour is measured against `color.background`, the only surface the list sits on. All clear 4.5:1.
- **Rows are `spacing.list-gap` 4 apart** (new, in `base`, the same in both densities: a row is read, not pressed). Rows closer together than the list sits to the box read as one list belonging to the field. The list sits under the field's box at `spacing.field-inner-gap` 8, the same gap as an error line.
- **The glyph's shape carries the meaning, not its colour.** The three outlines differ: an empty ring, a bare tick and a bare cross. That matters because the colours don't separate:
  - Missed and Unmet are 1.21:1 in Light, and 21 apart for a protanope.
  - Met in emerald and Missed in rose are 1.15 / 1.40, and **37 / 32 apart for a deuteranope** (green-blind reader), the classic red–green pair.

  Ink (`color.foreground`) was recommended for Met because it stays furthest from both: 127 / 161 apart from rose. **The user chose emerald**, and WCAG 1.4.1 holds because the shape never depends on the colour. Emerald was defined in ADR-017 and colours nothing else in Phase 1. Its ledger meaning, *done*, is what a met rule is.
- **After pressing Create account** with a rule missed, the password box turns rose, and each missed row turns to a rose cross. Met rows stay ticked. The rows are the message, so the field shows no separate error line: the Input instance's own `Message` is hidden. §6's rule still holds, because the rose edge never appears without words beside it.
- **While typing**, rows tick and untick. A tick is never an error, so nothing turns rose until the button is pressed.

### 5. A confirm-password field

It goes directly under the password field and its checklist, labelled *Confirm password*. It is checked when the button is pressed:
- if it is empty: *"Confirm your password."*
- if it differs from the password: *"Passwords don't match."*

The mismatch goes away as soon as the two match. The browser compares the two, and the request body is unchanged: the API still takes one `password`.

### 6. The 1px rose edge stays

An invalid edge is 1.02:1 against a resting one in Light, so it differs by hue alone. That is enough because **no look in the file shows the edge without words**: a field's message, or for the password, its missed rows. The edge never carries the meaning alone, and WCAG 1.4.1 is met. This is recorded so that nobody later drops the message and relies on the edge. `border.width-strong` was considered and not taken.

### 7. The wordings

They follow ADR-030's voice: say what happened and what to do, with no apology.

| Where | Wording |
|---|---|
| Login heading, button | Sign in |
| Login button, busy | Signing in… |
| Login, under the form | Don't have an account yet? **Create one** |
| Register heading | Create an account |
| Register button / busy | Create account / Creating account… |
| Register, under the form | Already have an account? **Sign in** |
| Name label / left empty | Name / Enter your name. |
| Email left empty | Enter your email address. |
| Email not an email | Enter a valid email address. |
| Email taken (409) | An account with this email already exists. |
| Password left empty (Login) | Enter your password. |
| Checklist row 1 | At least 8 characters |
| Checklist row 2 | At least one uppercase letter (A-Z) and one lowercase letter (a-z) |
| Checklist row 3 | At least one number (0-9) |
| Checklist row 4 | At least one symbol (e.g., ! @ # $ % ^ & *) |
| After each row, screen readers only | met / not met yet |
| Confirm label / empty / differs | Confirm password / Confirm your password. / Passwords don't match. |
| Wrong email or password (401) | Wrong email or password. |
| No answer, or a timeout | Can't reach CollaboTask. Check your connection and try again. |
| Anything else (5xx, a stray 400 or 403) | Something went wrong on our side. Try again. |

*Sign in* is in sentence case, like every other label. The user asked for *Sign In*, then chose sentence case when asked. The route stays `/login`, and the board keeps the name 2.1 Login. The rows' *(A-Z)*, *(a-z)* and *(0-9)* are the user's wording. SRS UC-01 accepts any uppercase or lowercase letter and any digit, so the server accepts more than the rows say, never less.

### 8. Behaviour

The drawing depends on these rules, so they are recorded here, not left to the code.

1. **Errors appear when the button is pressed**, not while you type and not when you leave a field. (The checklist's ticks are not errors; see 9.)
2. The not-satisfied button can still be pressed, and pressing it runs the same checks. ADR-018 §5 already says nothing is disabled.
3. On a failed press, **focus moves to the first field with an error**, top to bottom: name, email, password, confirm. After a 409 it moves to the email field.
4. A field's error goes away as soon as its value passes. The taken-email error goes away when the email changes.
5. The form message is announced as an alert. It goes away on the next press, or when any field changes.
6. Every value is kept after a failure, the passwords included.
7. The name field stops at 255 characters, so its too-long error can never appear.
8. Autocomplete hints: `name`, `email` and `new-password` (both password fields) on Register; `email` and `current-password` on Login.
9. The checklist is always shown, and its rows tick as you type.
10. After a press, a missed row turns back into a tick as soon as it is met, and the rose box returns to stone once all four are met.
11. For screen readers, the list is the password field's description (`aria-describedby`), and each row ends with a hidden *met* or *not met yet*. Ticks are not announced while typing, which would interrupt every keystroke. After a failed press, focus lands on the password field, which reads the list.
12. Confirm password is checked on press: empty, or different. The mismatch goes away as soon as the two match.
13. The confirmation is never sent to the server.
14. The four rules apply at registration only. Sign-in checks presence.
15. The server checks the same four rules (SRS UC-01). Until it does, a direct API call can still set an 8-character password with no other rule.

### 9. What the file draws

- **2.1 Login** stays at first arrival, now with *Sign in* and the new line under the form.
- **2.2 Register** is drawn at the moment the server says the email is taken:
  - *Ada Lovelace*, and *ada@example.com* on the Input's `Invalid` look with the 409 wording
  - the password filled, with all four rows `Met`
  - the confirmation filled
  - the button at `Rest`

  In the app, focus would also move to the email field and add the ring. The file draws one look at a time, as every part board does, and the code combines them.
- **`Form message`**: one component, no axis. Its board, `Form message — states`, shows Login straight after a 401, the way the Toast's board shows a toast over cards.
- **`Password rule`**: one axis, `State` = `Unmet` · `Met` · `Missed`. Its board shows the three looks, and two demos of the field: while typing `abcde1!`, and after pressing.
- **No whole-screen board for a failed moment.** ADR-020 §5 allows a whole screen only where the whole screen differs, and a failure differs by a line.

### 10. The code

- §9 gains `--list-gap: 4px`.
- The form message is `<p role="alert">`.
- The checklist is a `<ul>` the password input points at with `aria-describedby`. Each `<li>` carries its glyph `aria-hidden` and its status as visually hidden text.
- A field's error is shadcn's `FormMessage`.
- The four checks live in one shared function. Its definitions match SRS UC-01 exactly, so the browser and the server agree.
- The Button's sample label in the file is *Sign in*. Login uses it, and other screens override it.

## Options considered

### Where the form message sits
- **A tinted box under the heading** (`color.destructive-subtle`, `#FFF1F2` / `#4C0519`). It is easy to spot, but it needs a new colour in each theme, so question 4 would have fired. It also scrolls away under a phone's keyboard.
- **A line just below the button.** It can fall behind the keyboard, and it reads as a footnote to the button rather than the reason the press failed.

### The taken email as a form message
Rejected. It points at one field, and inventory §1.2 already said so.

### How the password requirement is said (before the rules changed)
- **One line that turns rose in place.** Rejected: for a protanope in Light it changes only its words (21 apart).
- **In the label.** Rejected: the label is what screen readers announce on every focus.

### The password policy
- **15 characters plus a common-password check**: NIST's rule for a single-factor password. Recommended. It needs the same spec and server work as the four rules, plus one message.
- **8 characters plus a common-password check.** It blocks the most-guessed passwords but stays below NIST's 15.
- **8 characters only**, as the spec stood. Nothing reopens, but it is the weakest.
- **The four rules.** Chosen (§3).

### How the four rules are shown
- **A sentence before submitting, and an error naming what is missing.** Recommended: no new part, no new value, and it fits rule 1 as first written. The user chose the checklist for the progress it shows before submitting.

### The checklist's details
- **All three glyphs inside a circle** (`Circle`, `CheckCircle`, `XCircle`). Tidier as a column, but only a mark about 6px wide changes, so the shape separates the states less well.
- **Met in ink.** Recommended for separation (§4). Emerald was chosen.
- **Rows 8 apart, reusing `spacing.field-inner-gap`.** No new token, but the rows read as four loose lines. **Rows touching (0).** The 16px glyphs end up about 3px apart.
- **Only the rows change after pressing, and the box stays stone.** Quieter, but it would be the only field whose box doesn't say *wrong*. **Rows, box and a message line.** It says the same thing twice and adds 24px.

### A 2px invalid edge (`border.width-strong`)
It is not needed while §6 holds. It would also spend a channel ADR-018 gave to the drop placeholder.

### *Sign In* in title case
It would be the only title-case label in the app.

## Out of scope (deferred, recorded)

- **The server half of §3**: the four checks, and 400 rather than 500 above bcrypt's 72 bytes. SRS §9 P1; a backend change, so it is planned before it is built.
- **A common-password blocklist.** Recommended and not adopted. It would add one checklist-independent message on the password field.
- **A show-password toggle.** Not drawn and not decided. It is the usual alternative to a confirm field; both together are allowed.
- **Field-level errors for the remaining fields** (name empty, email empty, confirm empty or different). The wordings are settled in §7, and the Input's `Invalid` look shows the pattern. No board draws each one.
- **Register at 390.** ADR-020 §5 keeps exactly one phone board, 2.15. The form reflows, and row 2 wraps at any width.
- **Password reset.** A Phase-2 deferral ([ADR-021](adr-021-design-inventory-corrections.md) *Out of scope*).

## Downstream edits this ADR requires

1. **[ADR-017](adr-017-ui-foundations.md): annotation, no rewrite.** Its Status line records §9's tenth extension, and a note under §9's opening points here.
2. **[ADR-018](adr-018-state-vocabulary.md): annotation.** Its Status line records that emerald now has a render site, and a note under the channel ledger points here.
3. **[ADR-024](adr-024-control-and-field-internals.md): annotation.** Its Status line records the settled question. Notes go under §5's error-message row and under *Out of scope*'s wrong-password bullet.
4. **[US-01](../../product/001-user-stories.md) and [SRS](../../spesifications/001-software-specifications.md) UC-01 and §9.** **Done 2026-11-21**, before the drawing. UC-08's "the one backend item still open" now reads one of two.
5. **Penpot `Collabotask App`.** **Done 2026-11-21.**
   - `spacing.list-gap` 4 was added to `base`. The catalogue reads 179.
   - The Button's six sample labels read *Sign in* and *Signing in…*. Login's heading reads *Sign in*, and its line under the form reads *Don't have an account yet? Create one*.
   - New part `Form message`, on the board `Form message — states`.
   - New part `Password rule`, `State` = Unmet · Met · Missed, on the board `Password rule — states`. Its glyphs are Phosphor `Circle`, `Check` and `X`, imported at 256 and scaled to 16.
   - Board `2.2 Register` on `Page 1`, made from a copy of 2.1 so that every binding came with it. Then:
     - Email was switched to `Invalid` and the button to `Rest`.
     - A Name and a Confirm instance were added.
     - The password Input and a `Password rules` list were wrapped in a `Password field` board.
   - The Foundations catalogue gained a `spacing.list-gap` row. The heading, the section name and the `base` title read 179 and 35.
   - **Binding check:**
     - 2.1, 2.2, both new part boards, the Button board and the catalogue section return nothing.
     - The Card board returns only its five expected `Menu → typography(13px)`.
     - Two text fills had dropped their binding and were rebound before the check came back clean (*Operational note*).
   - **Flips:** 2.2 was looked at in Light + Fine, Dark + Fine, Dark + Coarse and Light + Coarse, and both part boards in Dark.
     - Coarse makes the controls 44 and the field gap 20. The row gap stays 4.
     - Dark draws Met in `#34D399` and Missed in `#FB7185`.
     - Restored to `theme-light` + `density-fine`.
6. **[`design-file-inventory.md`](../design-file-inventory.md):**
   - §0's table and *Related documents* gain this ADR.
   - §1.1 and §1.2 are rewritten to match.
   - §4.2 gains the two parts, and §5 counts 12.
   - §7.2 records the parts, the count and 2.2.
   - §7.3 gains a *What question 4 returned for 2.2 Register* section.
   - §7.6 ticks 2.2.
   - The Appendix's rows 16, 17 and 36 point here.
7. **`CLAUDE.md` › `## Now`:** a pointer, including the second open backend item.
8. **`docs/build-log.md`:** no entry, following ADR-020 to ADR-031. Stage 4 is not closed.
9. **`CONTEXT.md`** and **[frontend-architecture.md](../frontend-architecture.md):** no change. No domain term moved. The architecture names the routes, not their copy.

## Operational note

**`createShapeFromSvg` drops its group into whatever board lies under the viewport.** With nothing selected, all three glyph groups landed inside the Card's main component. They were moved out at once, and the Card was checked back to its two rows. Move the path out of the group immediately, delete the group, and read the target part's children afterwards. The group is also the SVG's `viewBox` size (256), whatever `width` and `height` say. Scale the path by 1/16 and place it by its offset ÷ 16. The result matches the Toast's existing `X` exactly: 10.0 at 3.0.

**Writing an instance text's characters can drop its fill binding.** On both `Invalid` instances, the `Value` lost `fill → color.foreground` after its characters were set and the same token was applied in the same call. The main binds the same token, so the override may have been taken as a no-op. A second `applyToken` in a separate call held. The binding check found both, which is what it is for. Run it before calling a board done.

**A copied screen keeps its hidden overrides.** 2.1's password `Value` is hidden, so that the box draws empty at first arrival. The copy kept that, and 2.2's password box was blank until it was unhidden. After copying a screen, list every hidden child.

**`switchVariant(0, value)` keeps the instance's id and its overrides.** Email kept its label, and the button kept its fill width.

**A hidden child is left out of the flex layout, in an instance too.** Hiding the `Invalid` instance's `Message` gives a 64-tall field, not 88.

**The Penpot tab sleeps whenever it is not in front**, including while the user reads a decision page. Calls fail with *no heartbeat*. Ask the user to keep the tab in front for the whole drawing run, and inspect the state before re-running anything, because a call can run even when no heartbeat comes back.
