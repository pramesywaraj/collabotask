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

func (b *HubBroadcaster) EvictUser(boardID, userID uuid.UUID, reason string) {
	b.hub.EvictUser(boardID, userID, reason)
}

func (b *HubBroadcaster) EvictExcept(boardID uuid.UUID, allowed []uuid.UUID, reason string) {
	b.hub.EvictExcept(boardID, allowed, reason)
}

func (b *HubBroadcaster) EvictUserFromRooms(userID uuid.UUID, boardIDs []uuid.UUID, reason string) {
	b.hub.EvictUserFromRooms(userID, boardIDs, reason)
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
		userPayload := map[string]any{
			"id":         ev.User.ID,
			"email":      ev.User.Email,
			"name":       ev.User.Name,
			"avatar_url": ev.User.AvatarURL,
		}
		payload = map[string]any{
			"board_id": ev.BoardID,
			"user":     userPayload,
			"role":     ev.Role,
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
// Only keys listed in ChangedFields are included; the values come from
// CardToResponse so the shape matches the REST DTO exactly (Decision C3).
func projectCardFields(ev common.CardUpdated) map[string]any {
	r := response.CardToResponse(ev.Card, ev.Assignee)
	all := map[string]any{
		"title":       r.Title,
		"description": r.Description,
		"assigned_to": r.AssignedTo,
		"due_date":    r.DueDate,
	}
	out := make(map[string]any, len(ev.ChangedFields))
	for _, f := range ev.ChangedFields {
		if v, ok := all[f]; ok {
			out[f] = v
		}
	}
	return out
}

// projectBoardFields builds the changed-subset map from a BoardUpdated event.
func projectBoardFields(ev common.BoardUpdated) map[string]any {
	r := response.BoardToResponse(ev.Board)
	all := map[string]any{
		"title":            r.Title,
		"description":      r.Description,
		"background_color": r.BackgroundColor,
		"visibility":       r.Visibility,
	}
	out := make(map[string]any, len(ev.ChangedFields))
	for _, f := range ev.ChangedFields {
		if v, ok := all[f]; ok {
			out[f] = v
		}
	}
	return out
}
