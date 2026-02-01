package db

import (
	"context"

	"embed"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	pgxv5 "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"log/slog"
)

//go:embed migrations/*.sql
var EmbedMigrations embed.FS

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	// iofs source from embedded FS
	source, err := iofs.New(EmbedMigrations, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source %v", err)
	}

	// pgx driver uses a single connection internally
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire pgx connection %v", err)
	}
	defer conn.Release()
	db := stdlib.OpenDBFromPool(pool)

	driver, err := pgxv5.WithInstance(db, &pgxv5.Config{})
	if err != nil {
		return fmt.Errorf("failed to create pgx driver %v", err)
	}

	m, err := migrate.NewWithInstance(
		"iofs",
		source,
		"pgx",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to migrate up %v", err)
	}

	slog.Info("database migrations applied")
	return nil
}
