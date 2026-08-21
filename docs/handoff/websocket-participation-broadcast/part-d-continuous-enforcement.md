# Part D — Continuous enforcement (step ④)

**Written:** 2026-08-20, just-in-time before building Part D (per [index.md](./index.md) line 110).
**Re-read against current code:** `internal/realtime/hub.go` (Parts A+B), `internal/realtime/broadcast/broadcaster.go` (Part C), `internal/usecase/common/broadcaster.go` + `broadcast_events.go` (Part C port), `internal/usecase/board/{remove_member,leave_board,update_board,utils}.go`, `.mockery.yaml`.
**Design source:** [index.md §5 (continuous enforcement), §7 (participation cascade fan-out)](./index.md); [ADR-009](../../architecture/adr/adr-009-websocket-realtime-layer.md); SRS **§4.5** (UC-19b), **§5.2** (`ACCESS_REVOKED`, `MEMBER_REMOVED`).

---

## Scope — what Part D owns vs. defers

Part D adds **board-scoped eviction** to three existing usecase mutations that were already complete but had `TODO(ws)` stubs. It does **not** touch workspace-scoped eviction (that is Part E).

**Part D builds:**

| Trigger | Primitive | Signal to evicted | Usecase file |
|---|---|---|---|
| Removed from board (UC-12d) | `EvictUser(board, X, "removed_from_board")` | `ACCESS_REVOKED{board_id, reason}` | `board/remove_member.go` |
| Leaves board voluntarily (UC-10) | `EvictUser(board, X, "")` | silent (room sees `USER_LEFT` via hub presence) | `board/leave_board.go` |
| Board flipped WORKSPACE→PRIVATE (UC-12b) | `EvictExcept(board, members∪admins, "board_made_private")` | `ACCESS_REVOKED` to each evicted user | `board/update_board.go` + new `board/utils.go` helper |

**Part D explicitly defers:**

| Deferred to | What | Why not Part D |
|---|---|---|
| **Part E** | Workspace remove/leave fan-out (`EvictUserFromRooms`), `MEMBER_REMOVED` + `CARD_UPDATED` per board | Cross-board cascade; `WorkspaceUseCase` gains the port in E. |
| **Part E** | `CARD_UPDATED` broadcast for cascade-cleared cards on workspace leave | Same — workspace-scoped; already returned in `AffectedBoardIDs`/`[]AffectedCard`. |
| **Phase 2** | Eviction on token expiry / logout | Stateless JWT — no server-side revocation hook. Requires session/refresh token model (ADR-008 §7, ledger A in index.md). |

---

## New artifacts

### `common/broadcaster.go` — new frame type constants + `EvictReason` type

```go
FrameMemberRemoved  FrameType = "MEMBER_REMOVED"
FrameAccessRevoked  FrameType = "ACCESS_REVOKED"

type EvictReason string
const (
    EvictReasonSilent       EvictReason = ""                   // voluntary — no ACCESS_REVOKED
    EvictReasonRemoved      EvictReason = "removed_from_board" // UC-12d
    EvictReasonBoardPrivate EvictReason = "board_made_private" // UC-12b →PRIVATE flip
)
```

`ACCESS_REVOKED` is sent **only to the evicted user** (targeted, not room-broadcast). `EvictReasonSilent` skips it; any other constant triggers it. `EvictReason` is threaded through the `Broadcaster` interface so call sites use named constants, not magic strings. Part E adds its workspace-removal reason constant here.

### `common/broadcast_events.go` — `MemberRemoved` event struct

```go
type MemberRemoved struct {
    BoardID uuid.UUID
    UserID  uuid.UUID
}
func (MemberRemoved) FrameType() FrameType { return FrameMemberRemoved }
```

`CARD_UPDATED` re-uses the existing `CardUpdated` struct with `ChangedFields: []string{"assigned_to"}` and a minimal `&entity.Card{ID: card.CardID}`.

### `realtime/hub.go` — `BroadcastToUser` method

```go
func (h *Hub) BroadcastToUser(boardID, userID uuid.UUID, msg []byte)
```

Enqueues `msg` on every connection belonging to `userID` in `boardID`'s room (multi-tab correct). Uses `RLock` — **never holds the lock during a socket write** (same invariant as `Broadcast`). No-op if board or user has no room entry.

Called by the adapter's `sendAccessRevoked` helper **before** `hub.EvictUser` / `hub.EvictExcept`. The "best-effort" contract applies: there is an accepted race between the `writePump` draining the `send` channel and the eviction teardown closing the socket. `ACCESS_REVOKED` lands on the channel first; whether the writePump flushes it before teardown is not guaranteed but is fine — the client will reconnect, see access denied, and handle it.

### `realtime/broadcast/broadcaster.go` — adapter changes

Three additions to the `HubBroadcaster` adapter:

1. **`MemberRemoved` marshal arm** in the existing `marshal()` switch: `{board_id, user_id}`.

2. **`sendAccessRevoked(boardID, *userID, reason)` helper** — marshals `{type:"ACCESS_REVOKED", payload:{board_id, reason}}` and calls `h.hub.BroadcastToUser`. Swallows marshal errors (logs + returns).

3. **Updated `EvictUser`/`EvictExcept`** — when `reason != ""`, call `sendAccessRevoked` per targeted user **before** delegating to the hub. `EvictExcept` computes the evicted set (all `hub.ActiveUsers(boardID)` minus the allowed set) to send targeted `ACCESS_REVOKED` frames.

`buildAccessRevoked` is a package-private pure function tested directly.

---

## Ordering within remove_member (evict-first invariant)

Per index §7 and §5 (UC-19b), the order within `remove_member.go` is:

```
EvictUser(boardID, targetID, "removed_from_board")   ← closes socket + queues ACCESS_REVOKED before this
Broadcast(boardID, MemberRemoved{…})                  ← room sees MEMBER_REMOVED without evicted user
for _, card := range affectedCards {
    Broadcast(boardID, CardUpdated{Card: &entity.Card{ID: card.CardID}, ChangedFields: ["assigned_to"], Assignee: nil})
}
```

Evicting first ensures the evicted user's connection does not receive the subsequent `MEMBER_REMOVED` or `CARD_UPDATED` frames it triggered.

The `affectedCards` slice comes from the `RemoveWithParticipationCascade` return value (already returned but previously discarded with `_, err =` — Part D captures it).

---

## WORKSPACE→PRIVATE eviction

`update_board.go` detects the flip by comparing `board.Visibility == entity.BoardVisibilityPrivate` and `oldVisibility != ""` (non-zero means there was a prior value to flip from). It calls `bu.evictNonAllowed(ctx, board)` after the `BOARD_UPDATED` broadcast.

`evictNonAllowed` (in `board/utils.go`) fetches the board's member list and the workspace's member list, builds `allowed = board members ∪ workspace admins`, and calls `EvictExcept(board.ID, allowed, "board_made_private")`. Best-effort: any repo error is swallowed and eviction is skipped rather than failing the mutation.

**Why workspace admins are allowed:** break-glass (index §5, SRS §2.3 — workspace admin authority outranks board authority). An admin watching a board they're not a member of must not be evicted on a →PRIVATE flip.

**No eviction on PRIVATE→WORKSPACE:** access only widens; existing connections may stay. The table in `update_board.go` only triggers on `→PRIVATE`.

---

## Testing

### New hub tests — "Slice 6 — BroadcastToUser" (3 cases)
`internal/realtime/hub_test.go`

| Test | Verifies |
|---|---|
| `TestBroadcastToUser_DeliversOnlyToTargetUser` | Bystander in same room does not receive the message |
| `TestBroadcastToUser_MultiTab_AllConnsReceive` | Multi-tab user (2 conns) receives on both |
| `TestBroadcastToUser_UnknownBoard_IsNoop` | No panic when board has no room |

### New adapter tests (2 cases)
`internal/realtime/broadcast/broadcaster_test.go`

| Test | Verifies |
|---|---|
| `TestMarshal_MemberRemoved` | `{type:"MEMBER_REMOVED", payload:{board_id, user_id}}` wire shape |
| `TestBuildAccessRevoked` | `{type:"ACCESS_REVOKED", payload:{board_id, reason}}` envelope |

### New usecase broadcast-contract tests (4 functions)
`internal/usecase/board/board_broadcast_test.go`

| Test | Subtests | Key assertions |
|---|---|---|
| `TestRemoveMemberBroadcast` | happy path, no affected cards | evict-first ordering; `EvictUser` with non-silent reason; `Broadcast(MemberRemoved)`; `Broadcast(CardUpdated)` per cleared card |
| `TestRemoveMemberCascadeError` | cascade error | cascade error → no `EvictUser`, no broadcast (return before realtime calls) |
| `TestLeaveBoardBroadcast` | silent leave, cleared-cards leave, cascade error | `EvictUser` with `EvictReasonSilent`; `CARD_UPDATED` per cleared card; cascade error → no broadcast |
| `TestUpdateBoardPrivateEviction` | WORKSPACE→PRIVATE flip, PRIVATE→WORKSPACE flip, title-only update | `EvictExcept` called iff flipping to PRIVATE; `GetMembersByBoard` + `GetMembersByWorkspace` fetched for allowed list |

### Existing test fixes

- **`.Maybe()` expectations** added to the parallel test loops in `leave_board_test.go`, `remove_member_test.go`, `update_board_test.go`, `board_activity_test.go` so the new broadcaster/repo calls don't fail unrelated subtests.
- **Race fix** (`update_board_test.go`): the "boardRepo.Update fails" subtest was returning the shared `existingBoard` pointer from `GetByID`. With `input.Title` present, `UpdateBoard` writes `board.Title` at line 75 while parallel subtests do `b := *existingBoard`. Fixed with `b := *existingBoard; Return(&b, nil)` in that case's `setupMocks`.

**Final count:** 584 tests pass, `go test -race ./...` clean (confirmed with `-count=3` on the board package).

---

## What Part E needs from Part D

Part E (`WorkspaceUseCase` + cascade fan-out) depends on:
- The `EvictUser` / `EvictExcept` / `EvictUserFromRooms` port surface (settled in Part C, unchanged here)
- The `MemberRemoved` event struct (added here) — reused by E's per-board fan-out
- The `CardUpdated` pattern for cleared assignees (established here) — same pattern in E
- The evict-first ordering convention (established here) — E follows the same rule across boards

Part E does **not** need any new hub primitives beyond `EvictUserFromRooms` (already in the port, hub-implemented in Part A).

Part E's workspace-removal reason constant (`EvictReasonWorkspaceRemoved` or similar) slots into `common/broadcaster.go` alongside the Part D constants. The `broadcastClearedCards` helper in `board/utils.go` is shared — Part E calls it per affected board.


---

## Review findings (2026-08-21 `/code-review`, two-axis)

Two-axis review against this doc + index §5/§7 + ADR-009 + SRS §4.5/§5.2. **Verdict: spec-faithful, builds/vets clean, 244 affected tests pass — 0 hard standards violations, 0 material spec gaps.** The items below are refactors/cleanups, not blockers. **Do the two pre-Part-E actions before starting Part E** so E inherits the clean shape instead of copying the rough one.

### Pre-Part-E actions (do these first)

1. **Extract the CARD_UPDATED cascade fan-out into a shared helper** *(Duplicated Code — highest priority).*
   `leave_board.go` and `remove_member.go` contain the identical loop:
   ```go
   for _, card := range affectedCards {
       bu.broadcaster.Broadcast(input.BoardID, common.CardUpdated{
           Card:          &entity.Card{ID: card.CardID},
           ChangedFields: []string{"assigned_to"},
       })
   }
   ```
   Part E's workspace-cascade fan-out is the **third** copy of this exact shape (per §7 / the deferral table). Extract e.g. `bu.broadcastClearedCards(boardID uuid.UUID, cards []AffectedCard)` on the board usecase now and call it from both existing sites; Part E then calls the same helper per affected board instead of hand-rolling copy #3. This turns a looming three-site divergence into one definition.

2. **Type the eviction `reason`** *(Primitive Obsession).*
   `reason string` carries a load-bearing invariant — `""` ⇒ silent (no `ACCESS_REVOKED`), non-empty ⇒ involuntary (frame sent) — enforced only by scattered `if reason != ""` checks and the magic literals `"removed_from_board"` / `"board_made_private"` at call sites. The sibling `FrameType` constants are already centralized in `broadcaster.go`; `reason` should be too. Introduce a small type so the contract lives in the type system, not prose:
   ```go
   type EvictReason string
   const (
       EvictReasonSilent       EvictReason = ""                    // voluntary leave — no ACCESS_REVOKED
       EvictReasonRemoved      EvictReason = "removed_from_board"
       EvictReasonBoardPrivate EvictReason = "board_made_private"
   )
   ```
   Thread it through `EvictUser`/`EvictExcept`/`sendAccessRevoked` and the three call sites. Part E's workspace-removal reason then slots in as one more constant. *(Judgement call — no repo rule mandates it; recommended because eviction triggers keep multiplying.)*

### Optional cleanups (any time)

3. **`sendAccessRevoked` pointer param is Speculative Generality.** Signature is `sendAccessRevoked(boardID, userID *uuid.UUID, reason)`, but every caller passes `&userID` (never nil) and the body unconditionally dereferences `*userID`. The doc-comment's "(if non-nil)" branch does not exist. Change to a value param `userID uuid.UUID` and drop the `&`.

4. **Rename/complete the `TestRemoveMemberBroadcast` cascade subtest** *(cosmetic — spec axis).* This doc (Testing table) names the second subtest **"cascade-error-skips-broadcast"**, but the implemented subtest is **"no affected cards — evict + MEMBER_REMOVED, no CARD_UPDATED"** (it asserts the empty-cascade path, not a cascade-*error* path). The code is correct either way — both usecases `return err` before any broadcast when the cascade errors — but that early-return path currently has no test. Either rename this doc's line to match the implemented subtest, or add the genuine cascade-error subtest (preferred: the error path deserves coverage). Not a behaviour gap.

### Noted, no action

- **`EvictUserFromRooms` is a pure hub passthrough** on `HubBroadcaster`, asymmetric with `EvictUser`/`EvictExcept` (which now do adapter-side `ACCESS_REVOKED` work). This is **intentional** — it was declared in Part C for Part E to wire up; Part E adds the adapter-side fan-out logic. Flagged only so E doesn't mistake the passthrough for the finished shape.
