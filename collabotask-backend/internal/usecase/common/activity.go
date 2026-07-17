package common

import (
	"collabotask/internal/domain/entity"
	"collabotask/internal/domain/repository"
	"context"

	"github.com/rs/zerolog/log"
)

// WriteActivity performs the best-effort, after-commit activity write (ADR-007 Option A):
// on error it logs and swallows — a logging failure must never fail the caller's mutation.
func WriteActivity(ctx context.Context, repo repository.ActivityRepository, a *entity.Activity) {
	if err := repo.Log(ctx, a); err != nil {
		log.Error().Err(err).
			Str("board_id", a.BoardID.String()).
			Str("action", string(a.ActionType)).
			Str("entity", string(a.EntityType)).
			Msg("activity log write failed (swallowed)")
	}
}
