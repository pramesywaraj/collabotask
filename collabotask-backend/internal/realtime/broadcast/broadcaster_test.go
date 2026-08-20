package broadcast

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"collabotask/internal/delivery/http/response"
	"collabotask/internal/domain/entity"
	"collabotask/internal/usecase/common"
)

// unmarshalEnvelope parses the raw JSON produced by marshal into a generic envelope.
func unmarshalEnvelope(t *testing.T, raw []byte) (frameType string, payload map[string]any) {
	t.Helper()
	var env struct {
		Type    string         `json:"type"`
		Payload map[string]any `json:"payload"`
	}
	require.NoError(t, json.Unmarshal(raw, &env))
	return env.Type, env.Payload
}

// --- Card events ---

func TestMarshal_CardCreated(t *testing.T) {
	cardID := uuid.New()
	colID := uuid.New()
	userID := uuid.New()
	assigneeID := uuid.New()

	card := &entity.Card{
		ID:       cardID,
		ColumnID: colID,
		Title:    "Fix bug",
		Position: 1000,
		AssignedTo: &assigneeID,
		CreatedBy: userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	assignee := &entity.User{ID: assigneeID, Name: "Alice", AvatarURL: nil}

	ev := common.CardCreated{Card: card, Assignee: assignee}
	raw, err := marshal(ev)
	require.NoError(t, err)

	ft, payload := unmarshalEnvelope(t, raw)
	assert.Equal(t, "CARD_CREATED", ft)
	require.Contains(t, payload, "card")

	// payload.card must match CardToResponse byte-for-byte (shape guard)
	wantCard := response.CardToResponse(card, assignee)
	wantJSON, _ := json.Marshal(wantCard)
	gotCardJSON, _ := json.Marshal(payload["card"])
	assert.JSONEq(t, string(wantJSON), string(gotCardJSON))
}

func TestMarshal_CardMoved(t *testing.T) {
	cardID := uuid.New()
	fromColID := uuid.New()
	toColID := uuid.New()

	ev := common.CardMoved{
		CardID:       cardID,
		FromColumnID: fromColID,
		ToColumnID:   toColID,
		Position:     2000,
	}
	raw, err := marshal(ev)
	require.NoError(t, err)

	ft, payload := unmarshalEnvelope(t, raw)
	assert.Equal(t, "CARD_MOVED", ft)
	assert.Equal(t, cardID.String(), payload["card_id"])
	assert.Equal(t, fromColID.String(), payload["from_column_id"])
	assert.Equal(t, toColID.String(), payload["to_column_id"])
	assert.Equal(t, float64(2000), payload["position"])
}

func TestMarshal_CardUpdated_OnlyChangedFields(t *testing.T) {
	cardID := uuid.New()
	colID := uuid.New()
	userID := uuid.New()

	card := &entity.Card{
		ID:       cardID,
		ColumnID: colID,
		Title:    "New Title",
		CreatedBy: userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	ev := common.CardUpdated{
		Card:          card,
		Assignee:      nil,
		ChangedFields: []string{"title"},
	}
	raw, err := marshal(ev)
	require.NoError(t, err)

	ft, payload := unmarshalEnvelope(t, raw)
	assert.Equal(t, "CARD_UPDATED", ft)
	assert.Equal(t, cardID.String(), payload["card_id"])

	fields, ok := payload["fields"].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, fields, "title")
	assert.NotContains(t, fields, "description")
	assert.NotContains(t, fields, "due_date")
	assert.Equal(t, "New Title", fields["title"])
}

func TestMarshal_CardDeleted(t *testing.T) {
	cardID := uuid.New()
	colID := uuid.New()

	ev := common.CardDeleted{CardID: cardID, ColumnID: colID}
	raw, err := marshal(ev)
	require.NoError(t, err)

	ft, payload := unmarshalEnvelope(t, raw)
	assert.Equal(t, "CARD_DELETED", ft)
	assert.Equal(t, cardID.String(), payload["card_id"])
	assert.Equal(t, colID.String(), payload["column_id"])
}

// --- Column events ---

func TestMarshal_ColumnCreated(t *testing.T) {
	colID := uuid.New()
	boardID := uuid.New()

	col := &entity.Column{
		ID:        colID,
		BoardID:   boardID,
		Title:     "In Progress",
		Position:  1000,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	ev := common.ColumnCreated{Column: col}
	raw, err := marshal(ev)
	require.NoError(t, err)

	ft, payload := unmarshalEnvelope(t, raw)
	assert.Equal(t, "COLUMN_CREATED", ft)
	require.Contains(t, payload, "column")

	wantCol := response.ColumnToResponse(col)
	wantJSON, _ := json.Marshal(wantCol)
	gotColJSON, _ := json.Marshal(payload["column"])
	assert.JSONEq(t, string(wantJSON), string(gotColJSON))
}

func TestMarshal_ColumnUpdated(t *testing.T) {
	colID := uuid.New()

	ev := common.ColumnUpdated{ColumnID: colID, Title: "Done"}
	raw, err := marshal(ev)
	require.NoError(t, err)

	ft, payload := unmarshalEnvelope(t, raw)
	assert.Equal(t, "COLUMN_UPDATED", ft)
	assert.Equal(t, colID.String(), payload["column_id"])
	assert.Equal(t, "Done", payload["title"])
}

func TestMarshal_ColumnDeleted(t *testing.T) {
	colID := uuid.New()

	ev := common.ColumnDeleted{ColumnID: colID}
	raw, err := marshal(ev)
	require.NoError(t, err)

	ft, payload := unmarshalEnvelope(t, raw)
	assert.Equal(t, "COLUMN_DELETED", ft)
	assert.Equal(t, colID.String(), payload["column_id"])
}

func TestMarshal_ColumnMoved(t *testing.T) {
	colID := uuid.New()

	ev := common.ColumnMoved{ColumnID: colID, Position: 3000}
	raw, err := marshal(ev)
	require.NoError(t, err)

	ft, payload := unmarshalEnvelope(t, raw)
	assert.Equal(t, "COLUMN_MOVED", ft)
	assert.Equal(t, colID.String(), payload["column_id"])
	assert.Equal(t, float64(3000), payload["position"])
}

// --- Board events ---

func TestMarshal_MemberAdded(t *testing.T) {
	boardID := uuid.New()
	userID := uuid.New()

	user := &entity.User{ID: userID, Email: "alice@example.com", Name: "Alice", AvatarURL: nil}

	ev := common.MemberAdded{
		BoardID: boardID,
		User:    user,
		Role:    entity.BoardRoleMember,
	}
	raw, err := marshal(ev)
	require.NoError(t, err)

	ft, payload := unmarshalEnvelope(t, raw)
	assert.Equal(t, "MEMBER_ADDED", ft)
	assert.Equal(t, boardID.String(), payload["board_id"])
	assert.Equal(t, string(entity.BoardRoleMember), payload["role"])

	userMap, ok := payload["user"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, userID.String(), userMap["id"])
	assert.Equal(t, "alice@example.com", userMap["email"])
	assert.Equal(t, "Alice", userMap["name"])
}

func TestMarshal_OwnershipTransferred_NilFromUser(t *testing.T) {
	boardID := uuid.New()
	toUserID := uuid.New()

	ev := common.OwnershipTransferred{
		BoardID:    boardID,
		FromUserID: nil,
		ToUserID:   toUserID,
	}
	raw, err := marshal(ev)
	require.NoError(t, err)

	ft, payload := unmarshalEnvelope(t, raw)
	assert.Equal(t, "OWNERSHIP_TRANSFERRED", ft)
	assert.Equal(t, boardID.String(), payload["board_id"])
	assert.Nil(t, payload["from_user_id"])
	assert.Equal(t, toUserID.String(), payload["to_user_id"])
}

func TestMarshal_OwnershipTransferred_WithFromUser(t *testing.T) {
	boardID := uuid.New()
	fromUserID := uuid.New()
	toUserID := uuid.New()

	ev := common.OwnershipTransferred{
		BoardID:    boardID,
		FromUserID: &fromUserID,
		ToUserID:   toUserID,
	}
	raw, err := marshal(ev)
	require.NoError(t, err)

	ft, payload := unmarshalEnvelope(t, raw)
	assert.Equal(t, "OWNERSHIP_TRANSFERRED", ft)
	assert.Equal(t, fromUserID.String(), payload["from_user_id"])
	assert.Equal(t, toUserID.String(), payload["to_user_id"])
}

func TestMarshal_BoardUpdated_OnlyChangedFields(t *testing.T) {
	boardID := uuid.New()
	wsID := uuid.New()
	userID := uuid.New()

	board := &entity.Board{
		ID:          boardID,
		WorkspaceID: wsID,
		Title:       "New Board Title",
		CreatedBy:   userID,
		Visibility:  entity.BoardVisibilityWorkspace,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	ev := common.BoardUpdated{
		Board:         board,
		ChangedFields: []string{"title"},
	}
	raw, err := marshal(ev)
	require.NoError(t, err)

	ft, payload := unmarshalEnvelope(t, raw)
	assert.Equal(t, "BOARD_UPDATED", ft)
	assert.Equal(t, boardID.String(), payload["board_id"])

	fields, ok := payload["fields"].(map[string]any)
	require.True(t, ok)
	assert.Contains(t, fields, "title")
	assert.NotContains(t, fields, "visibility")
	assert.Equal(t, "New Board Title", fields["title"])
}

func TestMarshal_BoardArchivedSet_Archived(t *testing.T) {
	boardID := uuid.New()

	ev := common.BoardArchivedSet{BoardID: boardID, Archived: true}
	raw, err := marshal(ev)
	require.NoError(t, err)

	ft, payload := unmarshalEnvelope(t, raw)
	assert.Equal(t, "BOARD_ARCHIVED", ft)
	assert.Equal(t, boardID.String(), payload["board_id"])
}

func TestMarshal_BoardArchivedSet_Unarchived(t *testing.T) {
	boardID := uuid.New()

	ev := common.BoardArchivedSet{BoardID: boardID, Archived: false}
	raw, err := marshal(ev)
	require.NoError(t, err)

	ft, payload := unmarshalEnvelope(t, raw)
	assert.Equal(t, "BOARD_UNARCHIVED", ft)
	assert.Equal(t, boardID.String(), payload["board_id"])
}

// --- Error cases ---

func TestMarshal_UnknownEvent_ReturnsError(t *testing.T) {
	type unknownEvent struct{}
	// unknownEvent doesn't implement Event — test using a hand-rolled stub
	// that satisfies the interface but is not handled.
	type stubEvent struct{}
	_ = common.Event(nil) // just to import the package

	// We can't easily test the default branch without an unregistered type
	// because marshal is unexported. Use a type assertion trick: cast a struct
	// that implements Event (we'll define a local one) and call marshal.
	// Simplest: verify the error message format via a compile-time safe wrapper.
	_, err := marshal(dummyUnknownEvent{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown event")
}

// dummyUnknownEvent is an Event implementation not handled by marshal.
type dummyUnknownEvent struct{}

func (dummyUnknownEvent) FrameType() common.FrameType { return "UNKNOWN" }
