// Package db provides SQLite database connection management.
package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"

	_ "github.com/mattn/go-sqlite3"
	"github.com/revrost/code/counterspell/internal/db/sqlc"
)

//go:embed schema.sql
var schemaFS embed.FS

// DB wraps database/sql and sqlc queries.
type DB struct {
	DB      *sql.DB
	Queries *sqlc.Queries
}

// Connect creates a new SQLite database connection.
// If dbPath is empty, uses "./data/counterspell.db".
func Connect(ctx context.Context, dbPath string) (*DB, error) {
	if dbPath == "" {
		dbPath = "./data/counterspell.db"
	}

	// Open SQLite database
	sqlDB, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on&_journal_mode=WAL")
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
		DB:      sqlDB,
		Queries: sqlc.New(sqlDB),
	}, nil
}

// RunMigrations executes the schema.
func (db *DB) RunMigrations(ctx context.Context) error {
	// Read schema from embedded filesystem
	schemaBytes, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema: %w", err)
	}

	// Execute schema
	if _, err := db.DB.ExecContext(ctx, string(schemaBytes)); err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	slog.Info("Database schema initialized")

	return nil
}

// Close closes database connection.
func (db *DB) Close() {
	if err := db.DB.Close(); err != nil {
		slog.Error("Error closing database", "error", err)
	} else {
		slog.Info("Database connection closed")
	}
}
