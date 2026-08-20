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

### `common/broadcaster.go` — two new frame type constants

```go
FrameMemberRemoved  FrameType = "MEMBER_REMOVED"
FrameAccessRevoked  FrameType = "ACCESS_REVOKED"
```

`ACCESS_REVOKED` is sent **only to the evicted user** (targeted, not room-broadcast). A non-empty `reason` in `EvictUser`/`EvictExcept` triggers it; voluntary leaves pass `""`.

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

### New usecase broadcast-contract tests (3 functions)
`internal/usecase/board/board_broadcast_test.go`

| Test | Subtests | Key assertions |
|---|---|---|
| `TestRemoveMemberBroadcast` | happy path, cascade-error-skips-broadcast | evict-first ordering; `EvictUser` with non-empty reason; `Broadcast(MemberRemoved)`; `Broadcast(CardUpdated)` per cleared card |
| `TestLeaveBoardBroadcast` | success | `EvictUser` called with empty reason (no `ACCESS_REVOKED`) |
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
