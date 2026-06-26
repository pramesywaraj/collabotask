# CollaboTask — Domain Glossary

The canonical vocabulary for this project. When naming things in code, issues, tests, or proposals — use these terms exactly. Do not invent synonyms.

---

## Core Entities

**Workspace** — The top-level container. Groups boards for an organization or team. Use "Workspace" — never "Workplace" (earlier draft term, now retired).

**Board** — A project container within a workspace. Has columns, cards, members, and a visibility setting. Every board belongs to exactly one workspace.

**Column** — An ordered lane within a board representing a stage of work (e.g. To Do / In Progress / Done). Position is stored as fractional NUMERIC; one column is never aware of another's position.

**Card** — The atomic unit of work. Lives in a column, moves across columns to show progress. Canonical term is **Card** — "Ticket" is an accepted alias in the spec but Card is preferred in code and conversation.

---

## Roles

There are two independent role layers. **Effective permission is always computed from both — never from board role alone.**

**Workspace ADMIN** — Administrative superuser. Has full authority over every board in the workspace without needing to be a board member. Authority source: `workspace_members.role = 'ADMIN'`.

**Workspace MEMBER** — A participant in the workspace. Can create boards and self-join WORKSPACE-visible boards. Limited administrative privileges.

**BOARD_OWNER** — The single accountable owner of a board. Created by default as its creator; can be transferred cleanly to one other person. There is always exactly one owner (or none, if the owner left without transferring — which is safe, not an emergency).

**BOARD_MEMBER** — A user involved in a board (works on cards/columns). Board membership = *involvement*, not *authority*. A BOARD_MEMBER row never grants administrative power.

---

## Critical Distinctions

**Membership ≠ Authority.** A workspace admin who joins a PRIVATE board receives a `BOARD_MEMBER` row — but that row does not downgrade or change their authority. Their power comes from the workspace layer, always.

**Governance ≠ Participation.** Two distinct capabilities. **Governance** = controlling a board (settings, archive, delete, manage members) — admins hold this via their workspace role, *without joining*. **Participation** = doing the work (being assigned a card, presence, "my cards") — this requires board membership. An admin can govern a board they never joined, but to *participate* in its work they must join. When an admin leaves a board they keep governance but lose participation.

**Visibility ≠ Membership.** A user can see and open a WORKSPACE-visible board without being a board member. "Seeing a board" and "being a member of a board" are separate states.

**One owner, never co-owners.** Ownership transfer moves the title from one person to another. The concept of "co-owner" does not exist in this product.

---

## Board Visibility

**WORKSPACE** (default) — Every workspace member can see the board metadata, open it, and self-join.

**PRIVATE** — Only board members and workspace admins can see the board. Admins always see metadata, but must Join to open content (break-glass, logged).

**PUBLIC** — Deferred to Phase 2. Not a valid value in Phase 1.

---

## Key Rules

**Break-glass** — When a workspace admin opens the content of a PRIVATE board, they must Join first. This is a one-click action that is recorded in the activity log. It exists so "an admin looked inside a private board" is always auditable.

**Layered permission** — `can_administer_board(user, board)` = user is BOARD_OWNER of that board **OR** user is ADMIN of that board's workspace. Checks must always evaluate both conditions.

**Orphan-safe boards** — If a board owner leaves without transferring, the board has no owner but is never locked. Workspace admins retain full authority via the layered model. A new owner is appointed manually.

**Assignment = participation** — A card's assignee must be a **board member** (not merely a workspace member). When a user loses board participation — removed from / leaves a board, or removed from / leaves the workspace — their assignments on the affected boards are cleared (`assigned_to` → null) and a `CARD_UPDATED` is broadcast. Admins are not exempt: an admin who leaves a board is unassigned, keeping only governance. See [[assignment-participation-model]].

**Real-time is not a feature** — It is a layer applied to every state-changing board action. Every mutation that happens on a board broadcasts an event to connected clients in Phase 1.

---

## Terms to Avoid

| Avoid | Use instead | Why |
| --- | --- | --- |
| Workplace | Workspace | Retired term from earlier drafts |
| Ticket | Card | Card is canonical |
| Co-owner | — | Concept does not exist; use transfer |
| Admin join without logging | Break-glass (logged) | The log is the point |
