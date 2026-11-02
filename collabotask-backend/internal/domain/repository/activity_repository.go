package repository

import (
	"collabotask/internal/domain/entity"
	"context"
)

type ActivityRepository interface {
	// Log inserts one activity row, joining the caller's ambient transaction when
	// the context carries one. A failure returns an error to the caller and, inside
	// WithinTransaction, rolls the enclosing mutation back: an activity write is
	// atomic with the mutation it describes, never swallowed (ADR-016).
	Log(ctx context.Context, a *entity.Activity) error
}
