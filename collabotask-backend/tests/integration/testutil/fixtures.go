package testutil

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"collabotask/internal/domain/entity"
	pgRepos "collabotask/internal/repository/postgres"
)

// Fixtures holds the shared pool and exposes helpers for seeding test data.
// Create one per test package via NewFixtures(testPool).
type Fixtures struct {
	pool *pgxpool.Pool
}

func NewFixtures(pool *pgxpool.Pool) *Fixtures {
	return &Fixtures{pool: pool}
}

// CreateTestUser inserts a user with a unique email and returns the created entity.
func (f *Fixtures) CreateTestUser(t *testing.T) *entity.User {
	t.Helper()
	repo := pgRepos.NewUserRepository(f.pool)
	user := &entity.User{
		Email:        fmt.Sprintf("user-%s@test.com", uuid.New().String()[:8]),
		Name:         "Test User",
		PasswordHash: "$2a$10$testhashfortestingonly",
		SystemRole:   entity.SystemRoleUser,
	}
	require.NoError(t, repo.Create(context.Background(), user))
	return user
}

// CreateTestWorkspace inserts a workspace owned by ownerID and returns the created entity.
func (f *Fixtures) CreateTestWorkspace(t *testing.T, ownerID uuid.UUID) *entity.Workspace {
	t.Helper()
	repo := pgRepos.NewWorkspaceRepository(f.pool)
	ws := &entity.Workspace{
		Name:    fmt.Sprintf("Workspace-%s", uuid.New().String()[:8]),
		OwnerID: ownerID,
	}
	require.NoError(t, repo.Create(context.Background(), ws))
	return ws
}
