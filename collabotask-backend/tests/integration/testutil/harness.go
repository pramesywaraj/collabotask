package testutil

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// NewTestDB starts a postgres:16-alpine container, runs all migrations, and returns a
// pool ready for use. The returned terminate function stops the container and closes
// the pool — call it in TestMain after m.Run().
func NewTestDB() (*pgxpool.Pool, func()) {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("collabotask_test"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
	)
	if err != nil {
		panic(fmt.Sprintf("testutil: failed to start postgres container: %v", err))
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = container.Terminate(context.Background())
		panic(fmt.Sprintf("testutil: failed to get connection string: %v", err))
	}

	if err := runMigrations(dsn); err != nil {
		_ = container.Terminate(context.Background())
		panic(fmt.Sprintf("testutil: failed to run migrations: %v", err))
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		_ = container.Terminate(context.Background())
		panic(fmt.Sprintf("testutil: failed to create pool: %v", err))
	}

	terminate := func() {
		pool.Close()
		_ = container.Terminate(context.Background())
	}

	return pool, terminate
}

// TruncateAll removes all rows from every table, resetting sequences.
// Call it in t.Cleanup at the top of each test.
func TruncateAll(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `
		TRUNCATE users, workspaces, workspace_members,
		         boards, board_members, columns, cards, activities
		RESTART IDENTITY CASCADE
	`)
	require.NoError(t, err)
}

func runMigrations(dsn string) error {
	_, callerFile, _, ok := runtime.Caller(0)
	if !ok {
		return errors.New("could not determine harness file path")
	}

	// harness.go lives at tests/integration/testutil/ — three levels up is the backend root.
	backendRoot := filepath.Join(filepath.Dir(callerFile), "..", "..", "..")
	migrationsDir := filepath.Join(backendRoot, "migrations")

	m, err := migrate.New(fmt.Sprintf("file://%s", migrationsDir), dsn)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run up migrations: %w", err)
	}

	return nil
}
