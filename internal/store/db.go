package store

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
)

// Migration scripts
//
//go:embed migration/*.sql
var migrationFS embed.FS

// NewSQLite opens new SQLite database.
func NewSQLite(dbFilePath string, requiredVer uint) (*sqlx.DB, error) {
	// open DB
	db, err := sqlx.Open("sqlite", dbFilePath)
	if err != nil {
		return nil, fmt.Errorf("error opening DB: %w", err)
	}

	// SQLite does not support multiple connections!
	db.SetMaxOpenConns(1)

	if err = migrateDB(db, migrationFS, "migration", requiredVer); err != nil {
		return nil, fmt.Errorf("migration error: %w", err)
	}

	return db, nil
}

// migrateDB performs database migration.
func migrateDB(db *sqlx.DB, fs fs.FS, path string, requiredVer uint) error {
	data, err := iofs.New(fs, path)
	if err != nil {
		return fmt.Errorf("failed to load migration scripts: %w", err)
	}

	instance, err := sqlite.WithInstance(db.DB, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("failed to initialize database driver: %w", err)
	}
	m, err := migrate.NewWithInstance("iofs", data, "sqlite", instance)
	if err != nil {
		return fmt.Errorf("failed to initialize migration: %w", err)
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to perform migration: %w", err)
	}

	ver, isDirty, err := m.Version()
	if err != nil {
		return fmt.Errorf("failed to get database version: %w", err)
	}

	if isDirty {
		return errors.New("database is dirty")
	}

	if ver != requiredVer {
		return fmt.Errorf("unexpected database version: '%d', expected '%d'", ver, requiredVer)
	}

	return nil
}
