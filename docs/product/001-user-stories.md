# User Stories & Acceptance Criteria — CollaboTask (Phase 1)

> **Purpose:** the user-facing behavior of each Phase-1 feature, written so it can be verified. Each story maps 1:1 to a use case in [001-software-specifications.md §4](./001-software-specifications.md). The *why* (product) lives in [001-PRD.md](./001-PRD.md); the *how* (API contracts, SQL, schema, internal mechanics) lives in the spec.
>
> **How to read:**
> - Each story is **what a user wants** ("As a… I want… so that…").
> - **Acceptance Criteria** (Given/When/Then) are the **observable behavior** that must hold — what a tester checks. They name endpoints, status codes, and live events because those *are* the contract; they do **not** describe internal database mechanics (those are in the spec).
> - `US-xx` ↔ `UC-xx` (same number). Priority: **P0** = correctness/security blocker · **P1** = core Phase-1 feature · **P2** = polish.
> - Build order and the audit→story map live in [spec §9.1 / §9.2](./001-software-specifications.md).

## Conventions for every story
These hold for **all** stories unless a story says otherwise — not repeated each time:
- Calling an authenticated endpoint without a valid login → `401 UNAUTHORIZED`.
- A malformed request body → `400 VALIDATION_ERROR`.
- "Done" for a story = its AC pass, the response matches the documented shape (spec §6), and the listed scenarios have tests.

---

## EPIC A — Authentication

### US-01 — Register (UC-01) · P1
**As** a new user, **I want** to sign up with my email, name, and password, **so that** I have an account and can start using CollaboTask.

**AC**
```
Scenario: Successful registration
  Given no account exists for "a@x.com"
  When I register with a valid email, name, and password (at least 8 characters)
  Then my account is created and I'm logged in (I receive my profile and a session token)
  And my password is never stored in readable form or returned to me

Scenario: Email already taken
  Given an account already exists for "a@x.com"
  When I register with that same email
  Then I get a 409 CONFLICT telling me the email is already registered

Scenario: Invalid input
  When I register with a malformed email OR a password shorter than 8 characters
  Then I get a 400 VALIDATION_ERROR and no account is created
```

### US-02 — Login (UC-02) · P1
**As** a returning user, **I want** to log in, **so that** I can get back into my workspaces.

**AC**
```
Scenario: Successful login
  Given I have an account with the right password
  When I log in with correct credentials
  Then I'm logged in and receive my profile and a session token

Scenario: Wrong credentials
  When I log in with an unknown email OR a wrong password
  Then I get a 401 UNAUTHORIZED
  And the message does NOT reveal which of the two was wrong
```
> Note: there is **no** "account disabled" path in Phase 1 (deactivation is deferred — see spec [§3.5](./001-software-specifications.md); product rationale in PRD §5.2).

### US-02b — View My Profile (UC-02b) · P1
**As** a logged-in user, **I want** to see my own profile, **so that** the app can show who I am.

**AC**
```
Scenario: Authenticated
  When I request my profile while logged in
  Then I get my id, email, name, and avatar
```

---

## EPIC B — Workspace Management

### US-03 — Create Workspace (UC-03) · P1
**As** a user, **I want** to create a workspace, **so that** I have a place to group my team's boards.

**AC**
```
Scenario: Create workspace
  When I create a workspace with a non-empty name
  Then the workspace is created and I am its owner and admin
  And I get it back with its member count and board count

Scenario: Empty name
  When I try to create a workspace with no name
  Then I get a 400 VALIDATION_ERROR
```

### US-04 — Invite People to a Workspace (UC-04) · P1
**As** a workspace admin, **I want** to invite people by email (several at once), **so that** my team can join.

**AC**
```
Scenario: Invite existing users
  Given I am an admin of the workspace
  When I invite ["a@x.com","b@y.com"]
  Then each person who has an account is added as a member
  And I get back the updated member list

Scenario: Mixed batch — problems are reported per email
  Given one email has no account and one is already a member
  When I invite the batch
  Then the unknown email is reported as not found (404)
  And the already-member is reported as a conflict (409)
  And the valid invites still succeed

Scenario: Not an admin
  Given I am only a member
  When I try to invite someone
  Then I get a 403
```

### US-05 — See My Workspaces (UC-05) · P1
**As** a user, **I want** to see all the workspaces I belong to, **so that** I can switch between teams.

**AC**
```
Scenario: List with context
  When I ask for my workspaces
  Then I get only the ones I belong to, most recent first
  And each shows my role, how many members it has, and how many active boards
```

### US-05b — See a Workspace's Details (UC-05b) · P1
**As** a workspace member, **I want** to open a workspace and see its team, **so that** I know who's in it.

**AC**
```
Scenario: Member views details
  Given I belong to the workspace
  When I open it
  Then I see the workspace and its member list

Scenario: Non-member blocked
  Given I do NOT belong to the workspace
  When I try to open it
  Then I'm refused (403, or 404 to avoid revealing it exists)
```

### US-06 — Remove a Workspace Member (UC-06) · P1
**As** a workspace admin, **I want** to remove someone from the workspace, **so that** I can manage the team.

**AC**
```
Scenario: Removing a member also removes them from the team's boards
  Given I am an admin and X is a member
  When I remove X
  Then X is no longer in the workspace
  And X is no longer on any board in this workspace
  And any cards X was assigned to (across this workspace's boards) are unassigned
  And people viewing those boards see the assignee disappear
  And if X owned a board, that board is fine — admins still manage it (it's not an error)

Scenario: Can't remove yourself this way
  When I (an admin) target my own account
  Then I'm refused (I should leave the workspace instead)

Scenario: Not an admin
  Given I am a member
  When I try to remove anyone
  Then I get a 403
```

### US-06b — Promote / Demote a Member (UC-06b) · P1
**As** a workspace admin, **I want** to make another member an admin (or step one back down), **so that** I can share admin duties as the team grows.

**AC**
```
Scenario: Promote to admin
  Given I am an admin
  When I change a member's role to admin
  Then they become an admin

Scenario: Guard — keep at least one admin
  Given there is only one admin
  When I try to demote that admin
  Then I'm refused (a workspace must always have at least one admin)

Scenario: Guard — the workspace owner stays an admin
  When I try to demote the workspace owner
  Then I'm refused
```

### US-06c — Leave a Workspace (UC-06c) · P1
**As** a workspace member, **I want** to leave a workspace, **so that** I can step away from a team I'm no longer part of.

**AC**
```
Scenario: Member leaves
  Given I am a member (not the owner, not the only admin)
  When I leave the workspace
  Then I'm removed from it and from its boards
  And any cards I was assigned to (across this workspace's boards) are unassigned

Scenario: Guards
  Given I am the workspace owner → I can't leave (I must transfer or delete it first)
  Given I am the only admin → I can't leave
```

### US-06d — Delete a Workspace (UC-06d) · P1
**As** the workspace owner, **I want** to delete a workspace, **so that** I can remove it entirely.

**AC**
```
Scenario: Owner deletes (with confirmation)
  Given I am the workspace owner
  When I delete it after confirming
  Then the workspace and everything in it (boards, columns, cards, memberships) is gone

Scenario: Non-owner blocked
  Given I am an admin but not the owner
  When I try to delete it
  Then I get a 403
```

---

## EPIC C — Board Management

### US-07 — Create a Board (UC-07) · P1
**As** a workspace member, **I want** to create a board and choose whether the whole team can see it, **so that** I can start a project.

**AC**
```
Scenario: Create a board ready to use
  Given I belong to the workspace
  When I create a board with a title (it's workspace-visible by default)
  Then the board is created, I'm its owner, and it already has "To Do", "In Progress", and "Done" columns

Scenario: Private board
  When I create the board as private
  Then only its members and workspace admins can see and open it

Scenario: Guards
  Given I don't belong to the workspace → 403
  Given I give no title → 400
```

### US-08 — Browse Boards in a Workspace (UC-08) · P1
**As** a workspace member, **I want** to see the boards I'm allowed to see, **so that** I can find my work.

**AC**
```
Scenario: I see what I'm allowed to
  When I browse a workspace's boards
  Then as an admin I see every board
  And as a member I see workspace-visible boards plus private boards I'm on
  And each board shows my role, whether I've joined or can join, and its member and card counts
  And archived boards are not shown
```

### US-09 — Join a Board (UC-09) · P0
**As** a workspace admin (or a member looking at a workspace-visible board), **I want** to join a board, **so that** I can work on it — and when an admin opens a private board, that's recorded.

**AC**
```
Scenario: Member joins a workspace-visible board
  Given the board is workspace-visible and I'm a workspace member
  When I join it
  Then I become a board member
  And everyone viewing the board sees that I joined

Scenario: Admin opens a private board (recorded)
  Given the board is private and I'm a workspace admin who isn't on it yet
  When I join
  Then I become a board member, the join is recorded in the activity history, and viewers see I joined

Scenario: Not allowed to join
  Given the board is private and I'm a plain member who isn't on it
  When I try to join
  Then I get a 403
```

### US-10 — Leave a Board (UC-10) · P1
**As** a board member, **I want** to leave a board, **so that** I'm no longer involved in it.

**AC**
```
Scenario: Member leaves
  Given I'm a member of the board
  When I leave
  Then I'm no longer on it and viewers see I left
  And any cards I was assigned to on this board are unassigned, and viewers see the assignee disappear

Scenario: The owner must hand off first
  Given I'm the board's owner
  When I try to leave
  Then I'm refused (I must transfer ownership first)

Scenario: An admin who leaves keeps their authority
  Given I'm a workspace admin who had joined this board
  When I leave
  Then I'm no longer "on" the board, but I can still manage it as a workspace admin
  And my card assignments on this board are cleared too — I keep governance (managing it), not participation (holding the work)
```

### US-11 — Add People to a Board (UC-11) · P1
**As** a board owner or workspace admin, **I want** to add workspace members to a board, **so that** they're involved.

**AC**
```
Scenario: Add members
  Given I can manage the board
  When I add one or more workspace members
  Then each is added and viewers see them appear

Scenario: Validation
  Given someone I add isn't a workspace member → 400
  Given someone I add is already on the board → 409
  Given I can't manage the board → 403
```

### US-11b — See Who I Can Add (UC-11b) · P2
**As** a board owner or workspace admin, **I want** to see which workspace members aren't on the board yet, **so that** adding people is easy.

**AC**
```
Scenario: List candidates
  Given I can manage the board
  When I ask who I can add
  Then I get the workspace members who aren't on the board yet
```

### US-12 — Open a Board (UC-12) · P0
**As** a user with access, **I want** to open a board and its Kanban, **so that** I can see and work the columns and cards live.

**AC**
```
Scenario: Open a board I'm allowed to see
  When I open a board and its Kanban
  Then I can open it if I'm a workspace admin, OR it's workspace-visible, OR I created it, OR I'm a member of it
  And I see its columns and cards in order

Scenario: A private board an admin hasn't joined
  Given the board is private and I'm an admin who hasn't joined
  When I try to open its contents
  Then I'm asked to join first (I can see it exists, but not its contents until I join)

Scenario: The board comes alive
  When the board opens
  Then I'm connected live and can see who else is currently here
```

### US-12b — Edit Board Settings (UC-12b) · P0
**As** a board owner or workspace admin, **I want** to edit a board's title, description, color, and visibility, **so that** I can manage it.

**AC**
```
Scenario: Owner or admin can edit
  Given I'm the board owner OR a workspace admin
  When I change the board's settings
  Then the board updates and everyone viewing it sees the change

Scenario: A plain board member cannot edit settings
  Given I'm only a board member (and not a workspace admin)
  When I try to change settings
  Then I get a 403
```

### US-12c — Archive / Unarchive a Board (UC-12c) · P1
**As** a board owner or workspace admin, **I want** to archive a board, **so that** I can hide finished work without deleting it.

**AC**
```
Scenario: Archive and restore
  Given I can manage the board
  When I archive it
  Then it drops out of the normal board list and viewers see it was archived
  And when I unarchive it, it comes back
```

### US-12d — Remove Someone from a Board (UC-12d) · P1
**As** a board owner or workspace admin, **I want** to remove a member from a board, **so that** I can manage who's involved.

**AC**
```
Scenario: Remove a member
  Given I can manage the board
  When I remove a member
  Then they're no longer on it and viewers see them leave
  And any cards they were assigned to on this board are unassigned, and viewers see the assignee disappear
```

### US-12e — Transfer Board Ownership (UC-12e) · P1
**As** the current board owner (or a workspace admin), **I want** to hand ownership to another board member, **so that** responsibility moves cleanly to one clear person.

**AC**
```
Scenario: Hand off ownership
  Given I'm the current owner or a workspace admin, and the target is already a board member
  When I transfer ownership to them
  Then they become the sole owner and the previous owner becomes a regular member
  And the handoff is recorded and viewers see it

Scenario: Target isn't on the board
  Given the target isn't a board member yet
  When I try to transfer to them
  Then I'm refused (they must join first)

Scenario: A former owner who is also an admin keeps authority
  Given the outgoing owner is also a workspace admin
  When the transfer completes
  Then they can still manage the board — as an admin, not as the owner
```

---

## EPIC D — Columns & Cards
> Everything here needs board access: the board's owner, a board member, or a workspace admin. A workspace member who isn't on the board is refused (403) for all of these. Every change below is seen live by everyone viewing the board.

### US-13 — Add a Column (UC-13) · P1
**As** a team member working on a board, **I want** to add a column, **so that** I can represent a new stage of work.

**AC**
```
Scenario: Add a column
  When I add a column with a title
  Then it appears at the end of the board for everyone viewing it
```

### US-13b — Rename a Column (UC-13b) · P1
**As** a team member working on a board, **I want** to rename a column.
```
Scenario: Rename
  When I rename a column
  Then the new name shows for everyone viewing the board
```

### US-13c — Delete a Column (UC-13c) · P1
**As** a team member working on a board, **I want** to delete a column and its cards, with a confirmation.
```
Scenario: Delete
  When I delete a column after confirming
  Then it and its cards are removed for everyone viewing the board
```

### US-13d — Reorder Columns (UC-13d) · P1
**As** a team member working on a board, **I want** to drag a column to a new position.
```
Scenario: Reorder
  When I drop a column in a new spot
  Then it settles there for everyone viewing the board
```

### US-14 — Add a Card (UC-14) · P1
**As** a team member working on a board, **I want** to add a card, **so that** I can capture a task.

**AC**
```
Scenario: Add a card
  When I add a card with a title (optionally a description, assignee, and due date)
  Then it appears in the column for everyone viewing the board

Scenario: Assignee must be a board member
  Given the person I assign isn't a member of this board
  When I try to add the card
  Then I'm refused (they must be added to the board first — assignment requires board membership)

Scenario: Missing title
  When I try to add a card with no title
  Then I get a 400 VALIDATION_ERROR
```

### US-15 — Move a Card / Drag & Drop (UC-15) · P0
**As** a team member working on a board, **I want** to drag a card within or between columns, **so that** I can show progress live.

**AC**
```
Scenario: Drag a card
  When I drag a card to a new spot or column
  Then it moves immediately for me
  And everyone viewing the board sees it land in the same spot

Scenario: A move that isn't allowed snaps back
  Given the move is rejected (for example, I've lost access)
  When I try it
  Then my card snaps back to where it was and I see an error
```

### US-16 — Edit a Card (UC-16) · P1
**As** a team member working on a board, **I want** to edit a card.
```
Scenario: Edit or unassign
  When I change any of a card's title, description, assignee, or due date
  Then the change shows for everyone viewing the board
  And I can clear the assignee to leave the card unassigned
  And if I set an assignee, they must be a member of this board (as in US-14)
```

### US-17 — Delete a Card (UC-17) · P1
**As** a team member working on a board, **I want** to delete a card, with a confirmation.
```
Scenario: Delete
  When I delete a card after confirming
  Then it's removed for everyone viewing the board
```

---

## EPIC E — Real-time Collaboration

### US-18 — Live Connection (UC-18) · P0
**As** a user on a board, **I want** a live connection to it, **so that** I see changes as they happen.

**AC**
```
Scenario: Connect and join a board
  Given I'm logged in
  When I connect and join a board
  Then my access to that board is checked, I'm added to its live session, everyone sees me join, and I'm shown who's already here

Scenario: Bad or expired login
  Given my login is invalid or expired
  When I try to connect
  Then the connection is refused

Scenario: I drop off
  When my connection drops
  Then I'm removed from the live session and others see me leave

Scenario: I come back
  When my connection recovers
  Then it reconnects on its own and refreshes the board so I'm back in sync
```

### US-19 — See Who's Here / Presence (UC-19) · P1
**As** a user on a board, **I want** to see who else is active on it, **so that** collaboration feels live.
```
Scenario: Who's here
  Given other people are on the same board
  When I join
  Then I'm shown who's currently active on it
```

### US-20 — Instant Feedback / Optimistic UI (UC-20) · P1
**As** a user, **I want** my actions to take effect instantly, **so that** the app feels fast.
```
Scenario: Act now, confirm in the background
  When I make a change
  Then it shows immediately for me
  And once the server confirms, everyone else sees it too
  And if the server rejects it, my view snaps back and I see an error
```

---

## Coverage check (story ↔ use case)

| Stories | Use cases | Notes |
| --- | --- | --- |
| US-01 … US-02b | UC-01 … UC-02b | Authentication |
| US-03 … US-06d | UC-03 … UC-06d | Workspace management |
| US-07 … US-12e | UC-07 … UC-12e | Board management |
| US-13 … US-17 | UC-13 … UC-17 | Columns & cards |
| US-18 … US-20 | UC-18 … UC-20 | Real-time collaboration |

> Every Phase-1 use case (UC-01 … UC-20) has a matching story. For build order and how each story relates to the implementation audit, see [spec §9.1 / §9.2](./001-software-specifications.md).
