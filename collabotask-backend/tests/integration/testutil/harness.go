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

// NewTestDBAtVersion starts a fresh postgres:16-alpine container and migrates the
// schema up to (and including) targetVersion, returning a pool, the migrate handle
// (so the caller can run the remaining migrations with m.Up() after seeding
// pre-migration state), and a terminate function. Used to exercise a migration's
// data-cleanup behavior — which the always-fully-migrated shared pool cannot reach.
// Call terminate when done.
func NewTestDBAtVersion(t *testing.T, targetVersion uint) (*pgxpool.Pool, *migrate.Migrate, func()) {
	t.Helper()
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
	require.NoError(t, err)

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	m, err := newMigrate(dsn)
	require.NoError(t, err)
	require.NoError(t, m.Migrate(targetVersion))

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)

	terminate := func() {
		pool.Close()
		_, _ = m.Close()
		_ = container.Terminate(context.Background())
	}

	return pool, m, terminate
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

// newMigrate builds a migrate handle pointed at the repo's migrations directory.
// The caller owns closing it.
func newMigrate(dsn string) (*migrate.Migrate, error) {
	_, callerFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, errors.New("could not determine harness file path")
	}

	// harness.go lives at tests/integration/testutil/ — three levels up is the backend root.
	backendRoot := filepath.Join(filepath.Dir(callerFile), "..", "..", "..")
	migrationsDir := filepath.Join(backendRoot, "migrations")

	return migrate.New(fmt.Sprintf("file://%s", migrationsDir), dsn)
}

func runMigrations(dsn string) error {
	m, err := newMigrate(dsn)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run up migrations: %w", err)
	}

	return nil
}
