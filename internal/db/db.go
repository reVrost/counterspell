// Package db provides PostgreSQL database connection management.
package db

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/revrost/code/counterspell/internal/db/sqlc"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

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
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe

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

// RunMigrations executes all pending migrations.
func (db *DB) RunMigrations(ctx context.Context) error {
	// Create source from embedded filesystem
	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	// Get stdlib connection from pgxpool for migrate driver
	conn, err := db.Pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	// Convert pgx conn to stdlib *sql.DB
	sqlDB := stdlib.OpenDBFromPool(db.Pool)
	defer sqlDB.Close()

	// Create postgres driver
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Run migrations
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	version, dirty, _ := m.Version()
	if dirty {
		slog.Warn("Database migration state is dirty", "version", version)
	} else {
		slog.Info("Database migrations completed", "version", version)
	}

	return nil
}

// Close closes the connection pool.
func (db *DB) Close() {
	db.Pool.Close()
	slog.Info("Database connection pool closed")
}
