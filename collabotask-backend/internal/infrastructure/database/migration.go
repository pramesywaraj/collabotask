package database

import (
	"collabotask/internal/config"
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func newMigrate(cfg *config.Config) (*migrate.Migrate, *sql.DB, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.PingContext(context.Background()); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to ping database: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	migrationPath, err := filepath.Abs("migrations")
	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to resolve migrations path: %w", err)
	}

	sourceDriver, err := iofs.New(os.DirFS(migrationPath), ".")
	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to create source driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return m, db, nil
}

func RunMigrations(cfg *config.Config) error {
	m, db, err := newMigrate(cfg)
	if err != nil {
		return err
	}
	defer db.Close()
	defer m.Close()

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func GetMigrationVersion(cfg *config.Config) (uint, bool, error) {
	m, db, err := newMigrate(cfg)
	if err != nil {
		return 0, false, err
	}
	defer db.Close()
	defer m.Close()

	version, dirty, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}

	return version, dirty, nil
}
