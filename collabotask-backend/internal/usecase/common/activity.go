package common

import (
	"collabotask/internal/domain/entity"
	"collabotask/internal/domain/repository"
	"context"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// WriteActivity performs the best-effort, after-commit activity write (ADR-007 Option A):
// on error it logs and swallows — a logging failure must never fail the caller's mutation.
// actorID is set on a.UserID internally so callers never need a temporary copy to take &.
func WriteActivity(ctx context.Context, repo repository.ActivityRepository, actorID uuid.UUID, a *entity.Activity) {
	a.UserID = &actorID
	if err := repo.Log(ctx, a); err != nil {
		log.Error().Err(err).
			Str("board_id", a.BoardID.String()).
			Str("action", string(a.ActionType)).
			Str("entity", string(a.EntityType)).
			Msg("activity log write failed (swallowed)")
	}
}
