package postgres

import (
	"collabotask/internal/domain/entity"
	"collabotask/internal/domain/repository"
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type activityRepository struct {
	db *pgxpool.Pool
}

func NewActivityRepository(db *pgxpool.Pool) repository.ActivityRepository {
	return &activityRepository{db: db}
}

func (ar *activityRepository) Log(ctx context.Context, a *entity.Activity) error {
	metadata, err := json.Marshal(a.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal activity metadata: %w", err)
	}

	err = ar.db.QueryRow(ctx, insertActivityQuery,
		a.BoardID,
		a.UserID,
		string(a.ActionType),
		string(a.EntityType),
		a.EntityID,
		metadata,
	).Scan(&a.ID, &a.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert activity: %w", err)
	}

	return nil
}
