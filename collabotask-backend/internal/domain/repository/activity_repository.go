package repository

import (
	"collabotask/internal/domain/entity"
	"context"
)

type ActivityRepository interface {
	// Log inserts one activity row. A failure returns an error to the caller;
	// the swallow policy lives in common.WriteActivity, not here.
	Log(ctx context.Context, a *entity.Activity) error
}
