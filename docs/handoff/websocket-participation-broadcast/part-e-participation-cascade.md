# Part E — Participation cascade fan-out (step ④, final part)

**Written:** 2026-08-21, just-in-time before building Part E (per [index.md](./index.md) line 115).
**Re-read against current code:** `internal/usecase/workspace/{workspace_usecase,remove_member,leave_workspace}.go`, `internal/usecase/board/{utils,remove_member}.go`, `internal/usecase/common/{broadcaster,broadcast_events}.go`, `internal/realtime/broadcast/broadcaster.go`, `internal/realtime/hub.go`, `internal/domain/repository/{workspace_member,board}_repository.go`, `internal/repository/postgres/{workspace_member,workspace_member_queries,board_member,board_member_queries}.go`, `internal/injection/{providers,wire_gen}.go`.
**Design source:** [index.md §5 (continuous enforcement), §7 (participation cascade fan-out)](./index.md); [ADR-009](../../architecture/adr/adr-009-websocket-realtime-layer.md); SRS **§4.5** (UC-19b), **§5.2** (`ACCESS_REVOKED`, `MEMBER_REMOVED`, `CARD_UPDATED`); Part D's [evict-first ordering](./part-d-continuous-enforcement.md#ordering-within-remove_member-evict-first-invariant).

---

## Scope — what Part E owns vs. defers

Part E adds **workspace-scoped** eviction + broadcast fan-out to the two `WorkspaceUseCase` cascades that already run `RemoveWithParticipationCascade` but currently only write activities. It is the mirror of Part D, one layer up: Part D evicted from **one** board; Part E evicts a user from **every affected board in a workspace** in a single call, and closes the workspace-visible-watcher leak.

**Part E builds:**

| Trigger | Primitive | Signal to evicted | Usecase file |
|---|---|---|---|
| Removed from workspace (UC-06) | `EvictUserFromRooms(X, allWorkspaceBoards, "removed_from_workspace")` | `ACCESS_REVOKED{board_id, reason}` per board X was in | `workspace/remove_member.go` |
| Leaves workspace (UC-06c) | `EvictUserFromRooms(X, allWorkspaceBoards, "")` | silent (room sees `USER_LEFT` via hub presence) | `workspace/leave_workspace.go` |

Both then fan out, **per affected board**: `MEMBER_REMOVED{board_id, user_id}` + `CARD_UPDATED{assigned_to:null}` for every cleared card.

**Part E explicitly defers:**

| Deferred to | What | Why not Part E |
|---|---|---|
| **Phase 2** | Eviction on token expiry / logout | Stateless JWT — no server-side revocation hook (ADR-008 §7, index.md ledger A). Same deferral as Part D. |
| **⑤ integrity pass** | Composite-FK assignee invariant (ADR-010) that would auto-cascade unassignment and could supersede `RemoveWithParticipationCascade` | Needs the DB/repo harness; reopens shipped cascade code. Already the queued ⑤ work. |
| **Perf review** | Redis pub/sub multi-instance; hub lock sharding | index.md "Optimization notes" — not Phase-1. |

This is the **last part of step ④.** When it lands, ④ is complete and ⑤ opens.

---

## ⚠️ Decisions to approve before coding

Three genuine engineering-shape choices, each with a recommendation. Two touch the repo layer — which the `AffectedCard` struct comment optimistically said step ④ wouldn't need. Approve, redline, or ask; then I implement exactly what's approved.

### E1 — Where do "all workspace board IDs" come from? (the watcher-leak eviction set)

`EvictUserFromRooms` must evict X from **every board in the workspace**, not just the boards X was a *member* of. The leak (index §5 note): X watches a **WORKSPACE-visible** board without joining it — permitted, since join is only required for PRIVATE boards — then is removed from the workspace. That connection is **not** in `result.AffectedBoardIDs` (only member-boards land there), so it survives unless we evict over the full workspace board set. Passing the full set is correct *and* harmless (evicting a room X isn't in is a no-op).

`BoardRepository` has no method that returns all board IDs for a workspace (`GetUserBoardsInWorkspace` is user-scoped and returns list DTOs — wrong shape). Options:

- **(a) — RECOMMENDED.** Add `GetBoardIDsByWorkspace(ctx, workspaceID) ([]uuid.UUID, error)` to `BoardRepository` + Postgres impl (`SELECT id FROM boards WHERE workspace_id = $1`), and add `boardRepo` as a new dependency of `WorkspaceUseCase`. The board-list read then runs **outside** the cascade TX, in the best-effort post-commit path — consistent with "broadcast is best-effort, never touches the mutation." The method is genuinely reusable (Redis-swap, future workspace-board features) and honestly named.
  - *Cost:* one new dep on `WorkspaceUseCase` (`boardRepo`) — a normal, honest relationship (workspace ops relate to their boards).
- **(b).** Extend `WorkspaceCascadeResult` with `AllBoardIDs []uuid.UUID`, populated **inside** the cascade TX with an extra `SELECT`. Avoids the new usecase dep, but puts a read that exists *only to serve realtime* inside the mutation/commit path, and grows a result struct with a field only the broadcaster consumes.

> Recommend **(a)** — keeps realtime-only reads out of the commit path and yields a reusable query, at the cost of one honest dependency. Reject (b) because it couples the mutation TX to broadcast plumbing.

### E2 — How does `CARD_UPDATED` reach the *right* board room? (`AffectedCard` has no board ID)

`AffectedCard` is `{CardID, ColumnID}` — **no `BoardID`**. Part D's board cascade didn't need one: every cleared card was in the single known `input.BoardID`, so `broadcastClearedCards(input.BoardID, cards)` routes correctly. But a **workspace** cascade clears cards across **many boards**, and `Broadcast` is board-scoped (`Broadcast(boardID, …)`). Without a board ID per card we cannot route `CARD_UPDATED` to the correct room. The workspace unassign SQL already `JOIN boards b`, so the board id is one column away.

- **(a) — RECOMMENDED.** Add `BoardID uuid.UUID` to the shared `AffectedCard` struct; change `unassignCardsForUserQuery` `RETURNING c.id, c.column_id` → `RETURNING c.id, c.column_id, b.id` and scan it. For symmetry, set `card.BoardID = boardID` in the **board_member** cascade scan too (its board is the known param) so the field is never half-populated. Update the struct's now-obsolete "step ④ is a pure add (no repo change)" comment.
- **(b).** Introduce a separate `WorkspaceAffectedCard{CardID, ColumnID, BoardID}` type used only by the workspace cascade. Avoids touching the board path, at the cost of a near-duplicate type + a second scan shape.
- **(c).** Skip per-card `CARD_UPDATED` on workspace cascade; rely on `MEMBER_REMOVED` + client refetch. **Reject** — index §7 and the Part D deferral table both explicitly call for `CARD_UPDATED` per cleared card here.

> Recommend **(a)** — one field on the existing struct, both cascade paths fully populate it, no type duplication. This is the repo change the earlier "pure wire-up" note didn't foresee; it's small and additive (new `RETURNING` column, new struct field).

### E3 — New `EvictReason` constant

index §5: workspace **remove** (UC-06) is involuntary → `ACCESS_REVOKED`; workspace **leave** (UC-06c) is voluntary → silent. Add one constant beside Part D's, per `common/broadcaster.go`'s own note ("New workspace-removal reasons (Part E) add constants here"):

```go
EvictReasonRemovedFromWorkspace EvictReason = "removed_from_workspace" // UC-06 remove_member (involuntary)
```

Leave uses the existing `EvictReasonSilent`. This adds one value to the `ACCESS_REVOKED.reason` vocabulary (SRS §5.2 contract touch — additive).

### Design nuance to hold (not a fork, but easy to get wrong)

**Two different board sets are in play in one method.** Conflating them is the most likely bug:

| Set | Source | Used for |
|---|---|---|
| **All workspace boards** (superset) | `boardRepo.GetBoardIDsByWorkspace` (E1) | `EvictUserFromRooms` target — catches the watcher leak |
| **Affected boards / cards** (subset) | `result.AffectedBoardIDs` / `result.AffectedCards` | `MEMBER_REMOVED` per board + `CARD_UPDATED` per card — only boards X actually held membership/assignments on |

You evict over the superset; you broadcast membership/card changes over the subset.

---

## New / changed artifacts

### `domain/repository/board_repository.go` — new method (E1)

```go
// GetBoardIDsByWorkspace returns the IDs of every board in the workspace
// (archived included — evicting a room no one is in is a harmless no-op).
// Used by the participation-cascade fan-out to evict a removed user from all
// workspace boards, including WORKSPACE-visible boards they watched without joining.
GetBoardIDsByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]uuid.UUID, error)
```
Postgres impl + query `SELECT id FROM boards WHERE workspace_id = $1`.

### `domain/repository/workspace_member_repository.go` — `AffectedCard.BoardID` (E2)

```go
type AffectedCard struct {
	CardID   uuid.UUID
	ColumnID uuid.UUID
	BoardID  uuid.UUID // added Part E — routes CARD_UPDATED to the right room
}
```
Drop the obsolete "pure add (no repo change)" clause from the doc comment.

### `common/broadcaster.go` — new `EvictReason` constant (E3)

```go
EvictReasonRemovedFromWorkspace EvictReason = "removed_from_workspace"
```

### `realtime/broadcast/broadcaster.go` — `EvictUserFromRooms` gains ACCESS_REVOKED fan-out

Currently a pure hub passthrough (flagged "Noted, no action" in Part D's review — left for E). Make it mirror `EvictUser`: for a non-silent reason, send a targeted `ACCESS_REVOKED` per board **before** hub teardown.

```go
func (b *HubBroadcaster) EvictUserFromRooms(userID uuid.UUID, boardIDs []uuid.UUID, reason common.EvictReason) {
	if reason != common.EvictReasonSilent {
		for _, boardID := range boardIDs {
			b.sendAccessRevoked(boardID, userID, reason) // no-op if user not in that room
		}
	}
	b.hub.EvictUserFromRooms(userID, boardIDs, string(reason))
}
```
Reuses the existing `sendAccessRevoked` / `buildAccessRevoked` helpers unchanged — each evicted board gets an `ACCESS_REVOKED{board_id, reason}` carrying its own board id.

### `workspace/workspace_usecase.go` — struct gains `broadcaster` + `boardRepo`

Add `broadcaster common.Broadcaster` and `boardRepo repository.BoardRepository` fields + constructor params (mirrors `BoardUseCase`).

### `workspace/util.go` (or workspace_io.go) — new fan-out helper

`broadcastClearedCards` lives on `BoardUseCase` (`board/utils.go`) and is single-board; Part E's is cross-board, so it can't be reused as-is. Add a workspace-local helper that routes each card by its own `BoardID`:

```go
func (wu *WorkspaceUseCase) broadcastClearedCards(cards []repository.AffectedCard) {
	for _, card := range cards {
		wu.broadcaster.Broadcast(card.BoardID, common.CardUpdated{
			Card:          &entity.Card{ID: card.CardID},
			ChangedFields: []string{"assigned_to"},
		})
	}
}
```
(Same `CardUpdated` shape as Part D — minimal card + `assigned_to` changed-field.)

---

## Wiring — files to touch

| File | Change |
|---|---|
| `domain/repository/board_repository.go` | + `GetBoardIDsByWorkspace` (E1) |
| `domain/repository/workspace_member_repository.go` | + `AffectedCard.BoardID`; fix comment (E2) |
| `repository/postgres/board.go` + `board_queries.go` | impl + `SELECT id FROM boards WHERE workspace_id=$1` (E1) |
| `repository/postgres/workspace_member_queries.go` | `unassignCardsForUserQuery` RETURNING `+ b.id` (E2) |
| `repository/postgres/workspace_member.go` | scan the extra `board_id` into `AffectedCard.BoardID` (E2) |
| `repository/postgres/board_member.go` | set `card.BoardID = boardID` in cascade scan (E2 symmetry) |
| `usecase/common/broadcaster.go` | + `EvictReasonRemovedFromWorkspace` (E3) |
| `realtime/broadcast/broadcaster.go` | `EvictUserFromRooms` ACCESS_REVOKED fan-out |
| `usecase/workspace/workspace_usecase.go` | struct + ctor: `broadcaster`, `boardRepo` |
| `usecase/workspace/remove_member.go` | evict-first fan-out (involuntary) |
| `usecase/workspace/leave_workspace.go` | evict-first fan-out (silent) |
| `usecase/workspace/util.go` | `broadcastClearedCards` helper |
| `injection/providers.go` | `ProvideWorkspaceUseCase`: + `broadcaster`, `boardRepo` params |
| `injection/wire_gen.go` | regenerate via `cd internal/injection && go generate` |
| `.mockery.yaml` mocks | regenerate via `mockery` (BoardRepository gains a method; Broadcaster mock already exists) |

---

## Ordering within each usecase (evict-first invariant)

Same rule as Part D (index §7, §5 UC-19b): **evict → `MEMBER_REMOVED` → `CARD_UPDATED`.** Evicting first keeps the room broadcasts clean of the departing user. Activity writes stay exactly where they are; the realtime block is appended after them, in the best-effort post-commit tail.

**`remove_member.go` (UC-06, involuntary):**
```go
result, err := wu.workspaceMemberRepo.RemoveWithParticipationCascade(ctx, wsID, targetID)
// … existing per-board WriteActivity loop stays …

boardIDs, err := wu.boardRepo.GetBoardIDsByWorkspace(ctx, wsID)
if err != nil {
	boardIDs = result.AffectedBoardIDs // best-effort fallback: still evicts members, may miss a watcher
}
wu.broadcaster.EvictUserFromRooms(targetID, boardIDs, common.EvictReasonRemovedFromWorkspace)
for _, boardID := range result.AffectedBoardIDs {
	wu.broadcaster.Broadcast(boardID, common.MemberRemoved{BoardID: boardID, UserID: targetID})
}
wu.broadcastClearedCards(result.AffectedCards)
```

**`leave_workspace.go` (UC-06c, voluntary):** identical, but `EvictReasonSilent` and `requesterID` as the subject — no `ACCESS_REVOKED`; the room learns via the hub's `USER_LEFT` presence edge.

The `GetBoardIDsByWorkspace` error path is **best-effort**: log-swallow and fall back to `result.AffectedBoardIDs` as the eviction set (still evicts every member-board; only the rare workspace-visible-watcher edge is missed). Never fail the mutation — it already committed.

---

## Testing

### Adapter (`realtime/broadcast/broadcaster_test.go`) — 2 new
| Test | Verifies |
|---|---|
| `TestEvictUserFromRooms_NonSilent_SendsAccessRevokedPerBoard` | one `ACCESS_REVOKED{board_id, reason}` enqueued to the user per board, before hub teardown |
| `TestEvictUserFromRooms_Silent_NoAccessRevoked` | `EvictReasonSilent` → no ACCESS_REVOKED; hub still evicts |

### Workspace broadcast-contract (`usecase/workspace/workspace_broadcast_test.go`, new) — mock `Broadcaster` + `BoardRepository`
| Test | Key assertions |
|---|---|
| `TestRemoveMemberBroadcast` | evict-first ordering; `EvictUserFromRooms(target, allWorkspaceBoards, RemovedFromWorkspace)`; `MEMBER_REMOVED` per `AffectedBoardID`; `CARD_UPDATED` per affected card **routed to `card.BoardID`** |
| `TestRemoveMemberBroadcast_CascadeError` | cascade error → return before any evict/broadcast |
| `TestRemoveMemberBroadcast_BoardListError` | `GetBoardIDsByWorkspace` error → eviction falls back to `AffectedBoardIDs`; `MEMBER_REMOVED`/`CARD_UPDATED` still emitted |
| `TestLeaveWorkspaceBroadcast` | `EvictReasonSilent`; `MEMBER_REMOVED` + `CARD_UPDATED` per affected board/card |
| `TestLeaveWorkspaceBroadcast_CascadeError` | cascade error → no evict/broadcast |

### Existing test fixes
- `remove_member_test.go`, `leave_workspace_test.go`, `workspace_activity_test.go`: the `WorkspaceUseCase` constructor now takes a mock `Broadcaster` + mock `BoardRepository` — update the setup helpers. Add `.Maybe()` for the new broadcaster/`GetBoardIDsByWorkspace` calls in the parallel table loops so unrelated subtests don't fail on unexpected calls (same technique as Part D's board-test fixes).
- Run `go test -race ./...` (Part D confirmed `-count=3` on the hot package); expect the count to climb from 584.

### Regeneration
```bash
mockery                                   # BoardRepository.GetBoardIDsByWorkspace mock
cd internal/injection && go generate      # WorkspaceUseCase new deps
```

---

## Checkpoint (Done when)

Per index.md: *"Removing/leaving a workspace member fans out + evicts across all affected boards."* Concretely:
- A user removed from a workspace is evicted from **every** workspace board they had open — including a WORKSPACE-visible board they watched without joining — each with an `ACCESS_REVOKED{board_id}`.
- Each affected board room sees `MEMBER_REMOVED` + `CARD_UPDATED{assigned_to:null}` for that user's cleared cards.
- Leaving a workspace does the same but silently (rooms see `USER_LEFT`).
- `go build` / `go vet` / `go test -race ./...` clean; then a two-axis `/code-review` against this doc + index §5/§7 + SRS §4.5/§5.2 (the Part D pattern).

When this passes review and merges, **step ④ is complete** and ⑤ (post-Phase-1 integrity pass) opens.

---

## Post-Part-E — the gofmt sweep (from index ledger C)

index.md ledger C parks a `gofmt -w ./...` cleanup for **after Part E lands** (7 pre-existing dirty files, none introduced by realtime work). Do it as a **single isolated commit** once E merges, so formatting churn never mixes into a feature diff. Optionally add a CI `gofmt` check afterward.

---

## Build status — 2026-08-21

**Built to plan.** Decisions E1(a) / E2(a) / E3 implemented exactly as approved.

- **Repo (E1/E2):** `BoardRepository.GetBoardIDsByWorkspace` + Postgres impl/query; `AffectedCard.BoardID` field; workspace unassign SQL `RETURNING … b.id` + scan; board-member cascade sets `BoardID` from its known param (symmetry); obsolete "pure add" struct comment corrected.
- **Port/adapter:** `EvictReasonRemovedFromWorkspace` constant; `HubBroadcaster.EvictUserFromRooms` now sends a targeted `ACCESS_REVOKED` per board on a non-silent reason (was a bare passthrough).
- **Usecase:** `WorkspaceUseCase` gained `boardRepo` + `broadcaster`; single `fanOutParticipationCascade` helper (evict-first, all-boards evict + per-affected-board `MEMBER_REMOVED` + per-card `CARD_UPDATED` routed by `card.BoardID`, best-effort board-list fallback to `AffectedBoardIDs`); wired into `remove_member.go` (involuntary) + `leave_workspace.go` (silent).
- **Wiring:** `ProvideWorkspaceUseCase` + `mockery` + `go generate` regenerated.

**Tests:** adapter eviction (2, real hub + stub socket, `-race` stable ×5); workspace broadcast-contract (7 subtests: happy, board-list-error fallback, cascade-error, silent leave); existing 13 workspace constructor sites updated (`.Maybe()` broadcaster/boardRepo in the two fan-out files + activity tests). **`go build` / `go vet` clean; `go test -race ./...` = 596 pass; no new gofmt debt.**

**Next:** two-axis `/code-review` (this doc + index §5/§7 + SRS §4.5/§5.2). On pass → ④ complete; then update index map + `## Now` + memory, and run the ledger-C `gofmt` sweep. Manual two-client smoke remains pending on the same local DB-env friction noted for Parts C/D (not a code issue).
