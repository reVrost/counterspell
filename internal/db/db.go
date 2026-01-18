// Package db provides PostgreSQL database connection management.
package db

import (
	"context"
	"embed"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/revrost/code/counterspell/internal/db/sqlc"
)

//go:embed schema.sql
var schemaFS embed.FS

// DB wraps pgxpool and sqlc queries.
type DB struct {
	Pool    *pgxpool.Pool
	Queries *sqlc.Queries
}

// Connect creates a new database connection pool.
func Connect(ctx context.Context, databaseURL string) (*DB, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database URL: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	slog.Info("Connected to PostgreSQL database")

	return &DB{
		Pool:    pool,
		Queries: sqlc.New(pool),
	}, nil
}

// RunMigrations executes the schema.sql file against the database.
func (db *DB) RunMigrations(ctx context.Context) error {
	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema: %w", err)
	}

	_, err = db.Pool.Exec(ctx, string(schema))
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	slog.Info("Database migrations completed")
	return nil
}

// Close closes the connection pool.
func (db *DB) Close() {
	db.Pool.Close()
	slog.Info("Database connection pool closed")
}
