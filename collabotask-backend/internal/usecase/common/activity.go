package common

import (
	"collabotask/internal/domain/entity"
	"collabotask/internal/domain/repository"
	"context"

	"github.com/google/uuid"
)

// LogActivity sets a.UserID to actorID and calls repo.Log. It returns the error
// so the caller can propagate it and roll back the surrounding transaction (ADR-016).
func LogActivity(ctx context.Context, repo repository.ActivityRepository, actorID uuid.UUID, a *entity.Activity) error {
	a.UserID = &actorID
	return repo.Log(ctx, a)
}
