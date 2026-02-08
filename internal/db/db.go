// Package db provides SQLite database connection management.
package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/revrost/counterspell/internal/db/sqlc"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// DB wraps database/sql and sqlc queries.
type DB struct {
	db      *sql.DB
	path    string
	Queries *sqlc.Queries
}

// Connect creates a new SQLite database connection.
// If dbPath is empty, uses "~/.counterspell/data/counterspell.db".
func Connect(ctx context.Context, dbPath string) (*DB, error) {
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("unable to get home directory: %w", err)
		}
		dbPath = filepath.Join(home, ".counterspell", "data", "counterspell.db")
	}

	// Ensure parent directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("unable to create database directory: %w", err)
	}

	// Open SQLite database
	sqlDB, err := sql.Open("sqlite", dbPath+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("unable to open database: %w", err)
	}

	// Verify connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL"); err != nil {
		slog.Warn("Failed to enable WAL mode", "error", err)
	}

	slog.Info("Connected to SQLite database", "path", dbPath)

	return &DB{
		db:      sqlDB,
		path:    dbPath,
		Queries: sqlc.New(sqlDB),
	}, nil
}

// RunMigrations runs database migrations using golang-migrate.
func (db *DB) RunMigrations(ctx context.Context) error {
	// Strip the "migrations/" prefix from embedded FS
	sqlFS, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("unable to create migrations sub-filesystem: %w", err)
	}

	// Create migration source from embedded FS
	sourceDriver, err := iofs.New(sqlFS, ".")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	// Create database driver from existing DB connection
	dbDriver, err := sqlite3.WithInstance(db.db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("failed to create database driver: %w", err)
	}

	// Create migrate instance with both drivers
	m, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite3", dbDriver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Run migrations
	// Note: We don't call m.Close() because it closes the underlying *sql.DB
	// connection when using WithInstance. See: https://github.com/golang-migrate/migrate/issues/97
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	if err == migrate.ErrNoChange {
		slog.Info("Database migrations up to date")
	} else {
		slog.Info("Database migrations applied successfully")
	}

	return nil
}

// Close closes database connection.
func (db *DB) Close() {
	if err := db.db.Close(); err != nil {
		slog.Error("Error closing database", "error", err)
	} else {
		slog.Info("Database connection closed")
	}
}
