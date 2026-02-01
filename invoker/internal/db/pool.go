package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/revrost/invoker/internal/db/sqlc"
)

// DB represents database connection with sqlc queries
type DB struct {
	Pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// NewDB creates a new database connection with sqlc queries
func NewDB(databaseURL string) (*DB, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	queries := sqlc.New(pool)

	return &DB{
		Pool:    pool,
		queries: queries,
	}, nil
}

// Close closes the database connection pool
func (db *DB) Close() {
	db.Pool.Close()
}

// WithTx creates a new transaction and returns a new DB with transaction-bound queries
func (db *DB) WithTx(ctx context.Context) (*Tx, error) {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &Tx{
		Tx:      tx,
		queries: db.queries.WithTx(tx),
	}, nil
}

// Tx represents a database transaction
type Tx struct {
	Tx      pgx.Tx
	queries *sqlc.Queries
}

// Commit commits the transaction
func (t *Tx) Commit() error {
	return t.Tx.Commit(context.Background())
}

// Rollback rolls back the transaction
func (t *Tx) Rollback() error {
	return t.Tx.Rollback(context.Background())
}

// MustGetDB returns a database connection or panics
func MustGetDB() *DB {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		panic("DATABASE_URL is not set")
	}

	db, err := NewDB(databaseURL)
	if err != nil {
		panic(fmt.Sprintf("Failed to create database: %v", err))
	}

	return db
}
