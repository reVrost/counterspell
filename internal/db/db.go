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

// DB wraps SQL database connection.
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

// migrate runs schema migrations.
func migrate(db *sql.DB) error {
	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema: %w", err)
	}

	// Execute schema - this will fail for existing tables, which is fine
	_, err = db.Exec(string(schema))
	if err != nil {
		// Log error but continue - likely means table already exists
		slog.Info("Schema execution (may have expected errors)", "error", err.Error())
	}

	// Run additional migrations for existing databases
	if err := runMigrations(db); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	slog.Info("Migrations completed")
	return nil
}

// runMigrations executes specific migrations for existing databases.
func runMigrations(db *sql.DB) error {
	// Migration 1: Add project_id column to tasks table if it doesn't exist
	// Check if column exists
	var hasColumn bool
	err := db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM pragma_table_info('tasks')
		WHERE name = 'project_id'
	`).Scan(&hasColumn)

	if err != nil {
		return fmt.Errorf("failed to check project_id column: %w", err)
	}

	// Add column if it doesn't exist
	if !hasColumn {
		slog.Info("Adding project_id column to tasks table")
		_, err = db.Exec(`
			ALTER TABLE tasks
			ADD COLUMN project_id TEXT NOT NULL DEFAULT ''
		`)
		if err != nil {
			return fmt.Errorf("failed to add project_id column: %w", err)
		}

		// Create a default project if none exists
		var projectCount int
		_ = db.QueryRow("SELECT COUNT(*) FROM projects").Scan(&projectCount)

		if projectCount == 0 {
			slog.Info("Creating default project")
			_, err = db.Exec(`
				INSERT INTO projects (id, github_owner, github_repo, created_at)
				VALUES ('default', 'local', 'local-project', datetime('now'))
			`)
			if err != nil {
				slog.Warn("Failed to create default project", "error", err)
			}
		}

		// Update existing tasks to use the first available project or default
		_, err = db.Exec(`
			UPDATE tasks
			SET project_id = COALESCE(
				(SELECT id FROM projects LIMIT 1),
				'default'
			)
			WHERE project_id = ''
		`)
		if err != nil {
			slog.Warn("Failed to update tasks with project_id", "error", err)
		}
	}

	// Migration 2: Add message_history column to tasks table
	var hasMessageHistory bool
	err = db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM pragma_table_info('tasks')
		WHERE name = 'message_history'
	`).Scan(&hasMessageHistory)

	if err != nil {
		return fmt.Errorf("failed to check message_history column: %w", err)
	}

	if !hasMessageHistory {
		slog.Info("Adding message_history column to tasks table")
		_, err = db.Exec(`ALTER TABLE tasks ADD COLUMN message_history TEXT`)
		if err != nil {
			return fmt.Errorf("failed to add message_history column: %w", err)
		}
	}

	// Migration 3: Add agent_backend column to user_settings table
	var hasAgentBackend bool
	err = db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM pragma_table_info('user_settings')
		WHERE name = 'agent_backend'
	`).Scan(&hasAgentBackend)

	if err != nil {
		return fmt.Errorf("failed to check agent_backend column: %w", err)
	}

	if !hasAgentBackend {
		slog.Info("Adding agent_backend column to user_settings table")
		_, err = db.Exec(`ALTER TABLE user_settings ADD COLUMN agent_backend TEXT DEFAULT 'native'`)
		if err != nil {
			return fmt.Errorf("failed to add agent_backend column: %w", err)
		}
	}

	return nil
}

// Close closes database connection.
func (db *DB) Close() error {
	return db.DB.Close()
}
