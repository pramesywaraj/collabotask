package broadcast

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"collabotask/internal/delivery/http/response"
	"collabotask/internal/realtime"
	"collabotask/internal/usecase/common"
)

// HubBroadcaster implements common.Broadcaster over a *realtime.Hub.
type HubBroadcaster struct {
	hub *realtime.Hub
}

func NewHubBroadcaster(hub *realtime.Hub) *HubBroadcaster {
	return &HubBroadcaster{hub: hub}
}

func (b *HubBroadcaster) Broadcast(boardID uuid.UUID, e common.Event) {
	msg, err := marshal(e)
	if err != nil {
		log.Error().Err(err).
			Str("board_id", boardID.String()).
			Str("type", string(e.FrameType())).
			Msg("realtime marshal failed (swallowed)")
		return
	}
	b.hub.Broadcast(boardID, msg)
}

// EvictUser evicts all connections of userID from boardID. A non-silent reason
// triggers an ACCESS_REVOKED frame to the target before teardown (involuntary).
// EvictReasonSilent is a voluntary eviction (e.g. leave_board).
func (b *HubBroadcaster) EvictUser(boardID, userID uuid.UUID, reason common.EvictReason) {
	if reason != common.EvictReasonSilent {
		b.sendAccessRevoked(boardID, userID, reason)
	}
	b.hub.EvictUser(boardID, userID, string(reason))
}

// EvictExcept evicts every user in boardID not in allowed. A non-silent reason
// triggers ACCESS_REVOKED to each evicted user before teardown.
func (b *HubBroadcaster) EvictExcept(boardID uuid.UUID, allowed []uuid.UUID, reason common.EvictReason) {
	if reason != common.EvictReasonSilent {
		allowSet := make(map[uuid.UUID]struct{}, len(allowed))
		for _, id := range allowed {
			allowSet[id] = struct{}{}
		}
		for _, userID := range b.hub.ActiveUsers(boardID) {
			if _, ok := allowSet[userID]; !ok {
				b.sendAccessRevoked(boardID, userID, reason)
			}
		}
	}
	b.hub.EvictExcept(boardID, allowed, string(reason))
}

func (b *HubBroadcaster) EvictUserFromRooms(userID uuid.UUID, boardIDs []uuid.UUID, reason common.EvictReason) {
	b.hub.EvictUserFromRooms(userID, boardIDs, string(reason))
}

// sendAccessRevoked marshals and enqueues an ACCESS_REVOKED frame to a specific
// user in a room. Swallows marshal errors (logs + returns).
func (b *HubBroadcaster) sendAccessRevoked(boardID uuid.UUID, userID uuid.UUID, reason common.EvictReason) {
	msg, err := buildAccessRevoked(boardID, reason)
	if err != nil {
		log.Error().Err(err).
			Str("board_id", boardID.String()).
			Msg("realtime: failed to build ACCESS_REVOKED (swallowed)")
		return
	}
	b.hub.BroadcastToUser(boardID, userID, msg)
}

func buildAccessRevoked(boardID uuid.UUID, reason common.EvictReason) ([]byte, error) {
	return json.Marshal(envelope{
		Type: common.FrameAccessRevoked,
		Payload: map[string]any{
			"board_id": boardID,
			"reason":   string(reason),
		},
	})
}

type envelope struct {
	Type    common.FrameType `json:"type"`
	Payload any              `json:"payload"`
}

// marshal type-switches the concrete event to build the payload (reusing the REST
// response mappers), then wraps the {type,payload} envelope. Unknown type → error
// (swallowed by Broadcast). All transport/DTO knowledge is confined to this file.
func marshal(e common.Event) ([]byte, error) {
	var payload any
	switch ev := e.(type) {
	case common.CardCreated:
		payload = map[string]any{
			"card": response.CardToResponse(ev.Card, ev.Assignee),
		}
	case common.CardMoved:
		payload = map[string]any{
			"card_id":        ev.CardID,
			"from_column_id": ev.FromColumnID,
			"to_column_id":   ev.ToColumnID,
			"position":       ev.Position,
		}
	case common.CardUpdated:
		payload = map[string]any{
			"card_id": ev.Card.ID,
			"fields":  projectCardFields(ev),
		}
	case common.CardDeleted:
		payload = map[string]any{
			"card_id":   ev.CardID,
			"column_id": ev.ColumnID,
		}
	case common.ColumnCreated:
		payload = map[string]any{
			"column": response.ColumnToResponse(ev.Column),
		}
	case common.ColumnUpdated:
		payload = map[string]any{
			"column_id": ev.ColumnID,
			"title":     ev.Title,
		}
	case common.ColumnDeleted:
		payload = map[string]any{
			"column_id": ev.ColumnID,
		}
	case common.ColumnMoved:
		payload = map[string]any{
			"column_id": ev.ColumnID,
			"position":  ev.Position,
		}
	case common.MemberAdded:
		payload = map[string]any{
			"board_id": ev.BoardID,
			"user":     response.BoardMemberUserToResponse(ev.User, ev.JoinedAt),
			"role":     ev.Role,
		}
	case common.MemberRemoved:
		payload = map[string]any{
			"board_id": ev.BoardID,
			"user_id":  ev.UserID,
		}
	case common.OwnershipTransferred:
		payload = map[string]any{
			"board_id":     ev.BoardID,
			"from_user_id": ev.FromUserID,
			"to_user_id":   ev.ToUserID,
		}
	case common.BoardUpdated:
		payload = map[string]any{
			"board_id": ev.Board.ID,
			"fields":   projectBoardFields(ev),
		}
	case common.BoardArchivedSet:
		payload = map[string]any{
			"board_id": ev.BoardID,
		}
	default:
		return nil, fmt.Errorf("realtime: unknown event %T", e)
	}

	return json.Marshal(envelope{Type: e.FrameType(), Payload: payload})
}

// projectCardFields builds the changed-subset map from a CardUpdated event.
// Values come from CardToResponse so the shape matches the REST DTO (Decision C3).
func projectCardFields(ev common.CardUpdated) map[string]any {
	r := response.CardToResponse(ev.Card, ev.Assignee)
	return projectFields(map[string]any{
		"title":       r.Title,
		"description": r.Description,
		"assigned_to": r.AssignedTo,
		"due_date":    r.DueDate,
	}, ev.ChangedFields)
}

// projectBoardFields builds the changed-subset map from a BoardUpdated event.
func projectBoardFields(ev common.BoardUpdated) map[string]any {
	r := response.BoardToResponse(ev.Board)
	return projectFields(map[string]any{
		"title":            r.Title,
		"description":      r.Description,
		"background_color": r.BackgroundColor,
		"visibility":       r.Visibility,
	}, ev.ChangedFields)
}

// projectFields returns the subset of all whose keys are listed in changed —
// honouring §5.2's `fields:{…}` "changed subset only" rule (Decision C3).
func projectFields(all map[string]any, changed []string) map[string]any {
	out := make(map[string]any, len(changed))
	for _, f := range changed {
		if v, ok := all[f]; ok {
			out[f] = v
		}
	}
	return out
}
