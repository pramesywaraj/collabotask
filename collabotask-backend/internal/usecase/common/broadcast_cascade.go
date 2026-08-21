package common

import (
	"collabotask/internal/domain/entity"
	"collabotask/internal/domain/repository"
)

// BroadcastClearedCards emits CARD_UPDATED{assigned_to:null} for every card whose
// assignment was cleared by a participation cascade, routing each to its own board
// via AffectedCard.BoardID. Call it after eviction so the departing user is already
// out of the room. Shared by the board-scoped (UC-10/UC-12d) and workspace-scoped
// (UC-06/UC-06c) cascades — both populate BoardID, so routing is uniform.
func BroadcastClearedCards(b Broadcaster, cards []repository.AffectedCard) {
	for _, card := range cards {
		b.Broadcast(card.BoardID, CardUpdated{
			Card:          &entity.Card{ID: card.CardID},
			ChangedFields: []string{"assigned_to"},
		})
	}
}
