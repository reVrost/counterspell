package db

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaFS embed.FS

// DB wraps the SQL database connection.
type DB struct {
	*sql.DB
}

// Open opens a SQLite database at the given path and runs migrations.
func Open(path string) (*DB, error) {
	sqlDB, err := sql.Open("sqlite", path+"?_foreign_keys=1&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations
	if err := migrate(sqlDB); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	slog.Info("Database opened", "path", path)

	return &DB{sqlDB}, nil
}

// migrate runs the schema migrations.
func migrate(db *sql.DB) error {
	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema: %w", err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	slog.Info("Migrations completed")
	return nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.DB.Close()
}
