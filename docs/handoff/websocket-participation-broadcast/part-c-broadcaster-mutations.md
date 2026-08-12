# Part C — Broadcaster port + mutation broadcasts (step ④)

**Written:** 2026-08-11, just-in-time before building Part C (per [index.md](./index.md) line 110).
**Re-read against current code:** `internal/realtime/*` (Hub public API from Parts A+B), `internal/usecase/{card,column,board}/*`, `internal/usecase/common/{activity,board_access_check}.go`, `internal/delivery/http/response/{card,column,board}_response.go` (the DTO mappers), `internal/injection/{providers,wire}.go`, `.mockery.yaml`.
**Design source:** [index.md §1 (layering), §6 (broadcast wiring)](./index.md); [ADR-009](../../architecture/adr/adr-009-websocket-realtime-layer.md); SRS **§5.2** (the event table) + **§5.3** (client handling).

---

## Scope — what Part C owns vs. defers (read first)

Part C makes **card / column / board mutations appear live** in a joined room. It adds the `Broadcaster` **port** (the usecase-layer seam), the **typed event structs**, the **envelope adapter** over the Part-A hub, and one `Broadcast(...)` call **next to each existing `common.WriteActivity(...)`** — but **only for the additive "something changed, tell the room" events**. Anything coupled to *eviction* or the *workspace cascade* is out.

**Part C builds (12 broadcast call sites → 13 frame types):**

| Usecase file | `WriteActivity` line | Broadcasts (§5.2) |
|---|---|---|
| `card/create_card.go` | 74 | `CARD_CREATED` |
| `card/move_card.go` | 79 (guarded) | `CARD_MOVED` |
| `card/update_card.go` | 118 (guarded) | `CARD_UPDATED` |
| `card/delete_card.go` | 50 | `CARD_DELETED` |
| `column/create_column.go` | 36 | `COLUMN_CREATED` |
| `column/update_column.go` | 43 (guarded) | `COLUMN_UPDATED` |
| `column/delete_column.go` | 39 | `COLUMN_DELETED` |
| `column/update_column_position.go` | 45 (guarded) | `COLUMN_MOVED` |
| `board/invite_member.go` | 91 (per-member loop) | `MEMBER_ADDED` |
| `board/transfer_ownership.go` | 53 (has a `TODO(ws)`) | `OWNERSHIP_TRANSFERRED` |
| `board/update_board.go` | 110 (guarded) | `BOARD_UPDATED` |
| `board/set_archived.go` | 58 (guarded) | `BOARD_ARCHIVED` / `BOARD_UNARCHIVED` |

**Part C explicitly defers (boundary notes so later parts don't get re-litigated):**

| Deferred to | What | Why not Part C |
|---|---|---|
| **Part B (done)** | `USER_JOINED` / `USER_LEFT` / `ACTIVE_USERS` presence frames | Presence is **connection**-scoped, emitted by the hub on the WS join/leave edge — not a REST mutation broadcast. |
| **Part D** | `board/remove_member.go` → `MEMBER_REMOVED` **+** `EvictUser`; `board/leave_board.go` → `USER_LEFT`/evict; `→PRIVATE` flip eviction (`EvictExcept`); the **`ACCESS_REVOKED`** frame | These pair a broadcast with an **eviction**; keeping each removal whole in one PR is cleaner than splitting broadcast (C) from evict (D). See **Decision C2**. |
| **Part E** | `workspace/*` cascade fan-out (`MEMBER_REMOVED` + `CARD_UPDATED` per affected board, `EvictUserFromRooms`) | Cross-board cascade; `WorkspaceUseCase` gains the port in E, not C. |
| **Part D** | The `→PRIVATE` **eviction** on `UpdateBoard` | Part C broadcasts `BOARD_UPDATED` on every change **including** a visibility flip; D layers the conditional `EvictExcept` on top (purely additive). See **Decision C6**. |

**Not a broadcast site (Decision C1):** `board/self_join_board.go` (UC-09). §5.2 maps UC-09 → `USER_JOINED`, and `USER_JOINED` is a **presence** frame owned by the hub (Part B) — it fires when the self-joined user opens the board and its socket hits the 0→1 edge. A REST self-join changes the *roster*, not the *in-memory room*; emitting a presence frame for a not-yet-connected user would violate the "`ACTIVE_USERS` is built from the in-memory room, never the DB roster" invariant (index §4). So self-join gets **no** usecase-layer broadcast in Part C. (The existing comment at `self_join_board.go:69` — "activity log / broadcast stays silent" on the idempotent no-op path — already anticipates this.)

---

## Layering & artifacts (the one real design decision)

The dependency rule from Part A holds (**[part-a §Package & files](./part-a-hub-core.md), index §1**): *the usecase owns the port; `realtime` implements it.* Three artifacts, three homes, one import direction:

```
usecase/{card,column,board}  ──emits──►  common.Event  (typed structs, hold *entity.*)
                             ──calls───►  common.Broadcaster (port: Broadcast + evict family)
                                                 ▲
                                                 │ implements
realtime/broadcast (adapter) ────────────────────┘
   marshals common.Event ──► {type,payload} bytes ──► realtime.Hub.Broadcast(boardID, []byte)
   reuses delivery/http/response.*ToResponse for payload shape (client dedupe, UC-20)
```

**Why events carry `*entity.*`, not wire DTOs.** The usecase already holds `*entity.Card` (+ resolved `*entity.User` assignee), `*entity.Column`, `*entity.Board` at each call site. It emits those. The **adapter** — the only component that knows the transport — maps entity → wire shape by **reusing the exact `response.CardToResponse` / `ColumnToResponse` / `BoardToResponse` mappers the REST handlers use**. That guarantees the broadcast payload is byte-for-byte the REST DTO shape, which is what lets optimistic clients dedupe a broadcast against their own REST response (§5.2 broadcast rule, §5.3, UC-20). Duplicating the shape inside `realtime` would invite drift — the one thing to avoid here.

**Import direction (no cycles).** `usecase/common` gains **no** new imports (events reference only `entity` + `uuid`, which it already uses). The **adapter** imports `usecase/common` (Event types + port), `realtime` (the Hub), and `delivery/http/response` (DTO mappers). `common` never imports `realtime` → dependency rule intact. Verified acyclic: `common` ⇸ `realtime`; `response` ⇸ `realtime`.

**Adapter package (Decision C5).** Put the adapter in a **new package `internal/realtime/broadcast`**, *not* in `realtime` itself — that keeps the Part-A/B hub package pure (no `delivery/http/response` import leaking into the concurrency core). The hub stays "raw `[]byte` fan-out"; `broadcast` is the thin marshaling shell.

---

## Files

**New:**
```
internal/usecase/common/broadcaster.go        Broadcaster port + Event interface + FrameType consts
internal/usecase/common/broadcast_events.go    the 12 typed event structs (hold *entity.*)
internal/realtime/broadcast/broadcaster.go     HubBroadcaster adapter: Marshal(Event)→bytes, delegates to *realtime.Hub
internal/realtime/broadcast/broadcaster_test.go  marshaling unit tests (envelope + payload shape per event)
```

**Touched:**
```
internal/usecase/card/card_usecase.go          + broadcaster field + ctor param
internal/usecase/card/{create,move,update,delete}_card.go   + Broadcast next to WriteActivity
internal/usecase/column/column_usecase.go       + broadcaster field + ctor param
internal/usecase/column/{create,update,delete}_column.go, update_column_position.go   + Broadcast
internal/usecase/board/board_usecase.go          + broadcaster field + ctor param
internal/usecase/board/{invite_member,transfer_ownership,update_board,set_archived}.go  + Broadcast (remove the transfer_ownership TODO)
internal/injection/providers.go                  + ProvideBroadcaster; add param to Card/Column/Board usecase providers
internal/injection/wire.go                       add ProvideBroadcaster to UseCaseSet
internal/injection/wire_gen.go                   regenerate (go generate ./... or wire)
.mockery.yaml                                    add usecase/common: Broadcaster
internal/mocks/common_mocks.go                   regenerate (mockery)
```

---

## Signatures

### Port + Event interface (`usecase/common/broadcaster.go`)

```go
package common

import "github.com/google/uuid"

// FrameType is the outgoing "type" tag of a server→client mutation frame (§5.2).
// (Presence frame types — USER_JOINED/USER_LEFT/ACTIVE_USERS — live in the realtime
// package; they are emitted by the hub, not through this port.)
type FrameType string

const (
	FrameCardCreated  FrameType = "CARD_CREATED"
	FrameCardMoved    FrameType = "CARD_MOVED"
	FrameCardUpdated  FrameType = "CARD_UPDATED"
	FrameCardDeleted  FrameType = "CARD_DELETED"
	FrameColumnCreated FrameType = "COLUMN_CREATED"
	FrameColumnUpdated FrameType = "COLUMN_UPDATED"
	FrameColumnDeleted FrameType = "COLUMN_DELETED"
	FrameColumnMoved   FrameType = "COLUMN_MOVED"
	FrameMemberAdded          FrameType = "MEMBER_ADDED"
	FrameOwnershipTransferred FrameType = "OWNERSHIP_TRANSFERRED"
	FrameBoardUpdated         FrameType = "BOARD_UPDATED"
	FrameBoardArchived        FrameType = "BOARD_ARCHIVED"
	FrameBoardUnarchived      FrameType = "BOARD_UNARCHIVED"
)

// Event is a typed realtime mutation event. Concrete structs live in
// broadcast_events.go and carry *entity.* — the transport (envelope + payload
// shape) is the adapter's job, so the usecase never learns the wire format.
type Event interface {
	FrameType() FrameType
}

// Broadcaster is the usecase-layer port over the realtime hub. Best-effort,
// after-commit, swallow-on-error — identical contract to WriteActivity: a failed
// broadcast NEVER fails or rolls back the mutation. The concrete adapter lives in
// internal/realtime/broadcast.
//
// Part C uses only Broadcast. EvictUser/EvictExcept/EvictUserFromRooms are the
// settled surface (index §1) and the hub already implements them (Part A); they
// are declared here now so Parts D/E add call sites without re-touching the port
// or the adapter. See Decision C4.
type Broadcaster interface {
	Broadcast(boardID uuid.UUID, event Event)
	EvictUser(boardID, userID uuid.UUID, reason string)
	EvictExcept(boardID uuid.UUID, allowed []uuid.UUID, reason string)
	EvictUserFromRooms(userID uuid.UUID, boardIDs []uuid.UUID, reason string)
}
```

### Typed events (`usecase/common/broadcast_events.go`) — one per §5.2 payload

```go
// Cards
type CardCreated struct { Card *entity.Card; Assignee *entity.User }      // → { card{…} }
type CardMoved struct { CardID, FromColumnID, ToColumnID uuid.UUID; Position float64 } // → { card_id, from_column_id, to_column_id, position }
type CardUpdated struct { Card *entity.Card; Assignee *entity.User; ChangedFields []string } // → { card_id, fields:{…} }  (adapter projects the changed subset — Decision C3)
type CardDeleted struct { CardID, ColumnID uuid.UUID }                    // → { card_id, column_id }

// Columns
type ColumnCreated struct { Column *entity.Column }                       // → { column{ id,board_id,title,position } }
type ColumnUpdated struct { ColumnID uuid.UUID; Title string }           // → { column_id, title }
type ColumnDeleted struct { ColumnID uuid.UUID }                          // → { column_id }
type ColumnMoved struct { ColumnID uuid.UUID; Position float64 }          // → { column_id, position }

// Board
type MemberAdded struct { BoardID uuid.UUID; User *entity.User; Role entity.BoardRole } // → { board_id, user{…}, role }
type OwnershipTransferred struct { BoardID uuid.UUID; FromUserID *uuid.UUID; ToUserID uuid.UUID } // → { board_id, from_user_id, to_user_id }  (from may be null: orphan-safe)
type BoardUpdated struct { Board *entity.Board; ChangedFields []string }  // → { board_id, fields:{…} }  (adapter projects subset — C3)
type BoardArchivedSet struct { BoardID uuid.UUID; Archived bool }         // FrameType() → BOARD_ARCHIVED | BOARD_UNARCHIVED; payload → { board_id }

// each implements FrameType(); e.g.
func (BoardArchivedSet e) FrameType() common.FrameType {
	if e.Archived { return FrameBoardArchived }
	return FrameBoardUnarchived
}
```

### Adapter (`internal/realtime/broadcast/broadcaster.go`)

```go
package broadcast

// HubBroadcaster implements common.Broadcaster over a *realtime.Hub.
type HubBroadcaster struct { hub *realtime.Hub }

func NewHubBroadcaster(hub *realtime.Hub) *HubBroadcaster { return &HubBroadcaster{hub: hub} }

func (b *HubBroadcaster) Broadcast(boardID uuid.UUID, e common.Event) {
	msg, err := marshal(e)          // {type: e.FrameType(), payload: <shaped>}
	if err != nil {                 // best-effort: log + swallow, mirror WriteActivity
		log.Error().Err(err).Str("board_id", boardID.String()).
			Str("type", string(e.FrameType())).Msg("realtime marshal failed (swallowed)")
		return
	}
	b.hub.Broadcast(boardID, msg)   // hub is itself non-blocking / slow-consumer-drop (Part A)
}

func (b *HubBroadcaster) EvictUser(boardID, userID uuid.UUID, reason string) { b.hub.EvictUser(boardID, userID, reason) }
func (b *HubBroadcaster) EvictExcept(boardID uuid.UUID, allowed []uuid.UUID, reason string) { b.hub.EvictExcept(boardID, allowed, reason) }
func (b *HubBroadcaster) EvictUserFromRooms(userID uuid.UUID, boardIDs []uuid.UUID, reason string) { b.hub.EvictUserFromRooms(userID, boardIDs, reason) }

// marshal type-switches the concrete event to build the payload (reusing the REST
// response mappers), then wraps the {type,payload} envelope. Unknown type → error
// (swallowed by Broadcast). Keeps ALL transport/DTO knowledge in this one file.
func marshal(e common.Event) ([]byte, error) {
	type envelope struct { Type common.FrameType `json:"type"`; Payload any `json:"payload"` }
	var payload any
	switch ev := e.(type) {
	case common.CardCreated:
		payload = map[string]any{"card": response.CardToResponse(ev.Card, ev.Assignee)}
	case common.CardMoved:
		payload = map[string]any{"card_id": ev.CardID, "from_column_id": ev.FromColumnID, "to_column_id": ev.ToColumnID, "position": ev.Position}
	case common.CardUpdated:
		payload = map[string]any{"card_id": ev.Card.ID, "fields": projectCardFields(ev)} // C3
	// … one arm per event; column/board mirror the REST mappers …
	default:
		return nil, fmt.Errorf("realtime: unknown event %T", e)
	}
	return json.Marshal(envelope{Type: e.FrameType(), Payload: payload})
}
```

### Call-site pattern (example: `create_card.go`, right after the existing `WriteActivity`)

```go
common.WriteActivity(ctx, cru.activityRepo, input.RequesterID, &entity.Activity{ /* … unchanged … */ })

cru.broadcaster.Broadcast(column.BoardID, common.CardCreated{Card: card, Assignee: assignee})
```

- **Placement = after commit, next to `WriteActivity`, inside the same state-change guard.** For the *guarded* sites (`move_card`, `update_card`, `update_column`, `update_column_position`, `update_board`, `set_archived`) the `Broadcast` goes **inside** the existing `if changed { … }` / `if newPos != oldPosition { … }` block so a no-op mutation broadcasts nothing — mirroring log-on-state-change (ADR-007). `set_archived` already early-returns on the idempotent no-op before the write, so its `Broadcast` sits unconditionally next to `WriteActivity`.
- **`invite_member`**: emit one `MemberAdded` per added member, inside the existing `for _, m := range membersToAdd` loop next to `WriteActivity`, using `users[m.UserID]` for the `*entity.User`.
- **`board_id` for the room arg**: use the same board id the `WriteActivity` uses (`column.BoardID` for card/column events; `input.BoardID` / `board.ID` for board events).

---

## Wiring

**`providers.go`** — new provider (shares the one Hub singleton with the WS handler, so broadcasts reach the same rooms connections registered on):

```go
func ProvideBroadcaster(hub *realtime.Hub) common.Broadcaster {
	return broadcast.NewHubBroadcaster(hub)
}
```

Add `broadcaster common.Broadcaster` as a param to `ProvideCardUseCase`, `ProvideColumnUseCase`, `ProvideBoardUseCase` and thread it into each `New*UseCase(...)`. Update the three `New*UseCase` constructors + structs to hold it (`card_usecase.go`, `column_usecase.go`, `board_usecase.go`).

**`wire.go`** — add `ProvideBroadcaster` to `UseCaseSet`. `ProvideHub` stays in `HandlerSet`; Wire resolves the shared `*realtime.Hub` across sets (single global graph). Regenerate `wire_gen.go`.

**`.mockery.yaml`** — under `collabotask/internal/usecase/common` add `Broadcaster: {}` (alongside `BoardAccessChecker`), then regenerate → `MockBroadcaster` lands in `internal/mocks/common_mocks.go`.

---

## Decisions locked in this doc (confirm at build if any smell)

- **C1 — self-join emits nothing** (see Scope). Presence covers it; no roster ping needed since the actor *is* the new member.
- **C2 — `MEMBER_REMOVED` + board `leave` → Part D**, kept whole with their evictions rather than split broadcast-here / evict-there.
- **C3 — `*_UPDATED` `fields` = the changed subset**, honoring §5.2's `fields:{…}` literally. The adapter projects only `ChangedFields` out of the full `response.*ToResponse`, so shapes still match REST (e.g. `assigned_to` renders as `{id,name,avatar_url}`). Alternative (send full DTO, let client replace) rejected: diverges from §5.2 and sends more than changed.
- **C4 — declare the full `Broadcaster` interface now** (Broadcast + 3 evict methods), even though Part C only calls `Broadcast`. The hub already implements all four (Part A), so the adapter is complete in one pass and D/E add *call sites only* — no port/adapter churn, no re-mock. (If a reviewer objects to "unused" methods in C's diff, the fallback is Broadcast-only now + grow in D; not recommended — three interface edits instead of one.)
- **C5 — adapter in `internal/realtime/broadcast`**, keeping the hub package free of the `delivery/http/response` import.
- **C6 — `UpdateBoard` broadcasts `BOARD_UPDATED` on every change including a `→PRIVATE` flip; the eviction is Part D.** Intermediate state (C merged, D not) = a to-be-evicted member briefly sees a board-update frame; harmless and additive.
- **C7 — port + events live in `usecase/common`** (`broadcaster.go` + `broadcast_events.go`), chosen over a dedicated `usecase/realtime` package. Rationale: the `Broadcaster` port is a genuine sibling of `WriteActivity`/`BoardAccessChecker` already in `common`, and this avoids a second package named `realtime`. **Accepted trade-off:** `common` sits at the bottom of the `response → usecase/board → usecase/common` chain, so it can **never import above the domain layer** (importing `delivery/http/response` would cycle) — fine here because all DTO-shaping stays in the adapter (C5), never in the events. **Extraction trigger (revisit, don't pre-build):** when the event vocabulary grows enough to erode `common`'s cohesion — most likely **Phase 2** (comments/labels/attachments each add broadcast events). Extraction is then a mechanical package-rename (events → `usecase/realtime`, keep the port in `common`), no dependents beyond the usecases + adapter. Deliberately deferred (YAGNI); flagged so a future bloat isn't mistaken for an accident.

---

## Test checklist

**Usecase-layer emission (mirror the activity-contract tests; add a `MockBroadcaster` to each `new*Deps(t)` helper):**
- [ ] `CreateCard` → asserts one `Broadcast(boardID, CardCreated{…})` with the created card + resolved assignee.
- [ ] `MoveCard` → `CardMoved` **emitted on actual move**; **no** broadcast when column & position unchanged (guard parity).
- [ ] `UpdateCard` → `CardUpdated` with the right `ChangedFields`; **no** broadcast when nothing changed; assigned-to change carries the assignee.
- [ ] `DeleteCard` → `CardDeleted{card_id, column_id}`.
- [ ] `CreateColumn` / `UpdateColumn` / `DeleteColumn` → respective frames; `UpdateColumn` no-op emits nothing.
- [ ] `UpdateColumnPosition` → `ColumnMoved` only when `newPos != oldPosition`.
- [ ] `InviteMember` → one `MemberAdded` **per** added member; skipped/duplicate members emit none.
- [ ] `TransferOwnership` → `OwnershipTransferred` with `from` possibly nil; idempotent (target already owner) emits nothing.
- [ ] `UpdateBoard` → `BoardUpdated` with changed subset; no-op emits nothing.
- [ ] `SetArchived` → `BoardArchived` vs `BoardUnarchived` by target state; idempotent no-op emits nothing.
- [ ] **Best-effort:** a `Broadcaster` whose `Broadcast` is a no-op/panic-free stub never affects the usecase's return (broadcast failures are swallowed at the adapter, but assert the usecase doesn't depend on a return value — the port method returns nothing).

**Adapter marshaling (`internal/realtime/broadcast/broadcaster_test.go`, pure, no hub needed — inject a fake or assert `marshal` directly):**
- [ ] Each event marshals to `{"type": "<FRAME>", "payload": {…}}` with the exact §5.2 keys.
- [ ] `CardCreated.payload.card` matches `response.CardToResponse` byte-for-byte (shape-consistency guard).
- [ ] `CardUpdated`/`BoardUpdated` `fields` contains **only** the changed keys.
- [ ] `BoardArchivedSet{Archived:true/false}` → `BOARD_ARCHIVED` / `BOARD_UNARCHIVED`.
- [ ] `OwnershipTransferred` with nil `FromUserID` → `from_user_id: null`.
- [ ] Unknown event type → `marshal` returns an error (and `Broadcast` swallows it — no panic, no hub call).

No DB harness and no `-race` needed at this layer (usecase tests use the mock; the adapter test is pure marshaling). The hub's own concurrency is already covered by Part A.

---

## Done when (checkpoint)

A card/column/board mutation issued over REST **appears live** in a joined room (manual: two clients, one joins via WS, the other mutates via REST → first sees the frame), and every usecase test asserts emission (and non-emission on no-op) via the mock `Broadcaster`. `go build ./...` + `go vet ./...` clean; mocks + `wire_gen.go` regenerated. Then a two-axis `/code-review` (Standards + Spec-vs-§5.2), same cadence as Parts A/B.
