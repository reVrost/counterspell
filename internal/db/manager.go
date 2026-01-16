// Package db provides database management for multi-tenant SQLite databases.
package db

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/revrost/code/counterspell/internal/config"
)

// DBManager manages per-user SQLite database connections.
// It caches open connections and handles lazy migration.
type DBManager struct {
	cfg       *config.Config
	databases map[string]*DB
	mu        sync.RWMutex
	lastUsed  map[string]time.Time
}

// NewDBManager creates a new database manager.
func NewDBManager(cfg *config.Config) *DBManager {
	return &DBManager{
		cfg:       cfg,
		databases: make(map[string]*DB),
		lastUsed:  make(map[string]time.Time),
	}
}

// GetDB returns the database for a user, creating it if necessary.
// In single-player mode (MULTI_TENANT=false), always returns the "default" user's DB.
func (m *DBManager) GetDB(userID string) (*DB, error) {
	// Normalize to "default" in single-player mode
	if !m.cfg.MultiTenant {
		userID = "default"
	}

	// Check cache first with read lock
	m.mu.RLock()
	if db, ok := m.databases[userID]; ok {
		m.mu.RUnlock()
		// Update last used time (needs write lock, but do it async to not block)
		go m.updateLastUsed(userID)
		return db, nil
	}
	m.mu.RUnlock()

	// Need to create/open the database
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if db, ok := m.databases[userID]; ok {
		m.lastUsed[userID] = time.Now()
		return db, nil
	}

	// Open the database
	db, err := m.openUserDB(userID)
	if err != nil {
		return nil, err
	}

	m.databases[userID] = db
	m.lastUsed[userID] = time.Now()

	slog.Info("Opened user database", "user_id", userID)
	return db, nil
}

// openUserDB opens or creates a database for a user.
func (m *DBManager) openUserDB(userID string) (*DB, error) {
	dbPath := m.cfg.DBPath(userID)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	// Open the database (this runs migrations automatically)
	db, err := Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database for user %s: %w", userID, err)
	}

	return db, nil
}

// updateLastUsed updates the last used time for a user's database.
func (m *DBManager) updateLastUsed(userID string) {
	m.mu.Lock()
	m.lastUsed[userID] = time.Now()
	m.mu.Unlock()
}

// CloseDB closes the database for a specific user.
func (m *DBManager) CloseDB(userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	db, ok := m.databases[userID]
	if !ok {
		return nil
	}

	if err := db.Close(); err != nil {
		return fmt.Errorf("failed to close database for user %s: %w", userID, err)
	}

	delete(m.databases, userID)
	delete(m.lastUsed, userID)

	slog.Info("Closed user database", "user_id", userID)
	return nil
}

// CloseInactive closes databases that haven't been used for the given duration.
func (m *DBManager) CloseInactive(maxAge time.Duration) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	closed := 0

	for userID, lastUsed := range m.lastUsed {
		if now.Sub(lastUsed) > maxAge {
			db, ok := m.databases[userID]
			if ok {
				if err := db.Close(); err != nil {
					slog.Error("Failed to close inactive database",
						"user_id", userID,
						"error", err)
					continue
				}
				delete(m.databases, userID)
				delete(m.lastUsed, userID)
				closed++
				slog.Info("Closed inactive database", "user_id", userID, "inactive_for", now.Sub(lastUsed))
			}
		}
	}

	return closed
}

// CloseAll closes all open databases.
func (m *DBManager) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for userID, db := range m.databases {
		if err := db.Close(); err != nil {
			errs = append(errs, fmt.Errorf("user %s: %w", userID, err))
		}
	}

	m.databases = make(map[string]*DB)
	m.lastUsed = make(map[string]time.Time)

	if len(errs) > 0 {
		return fmt.Errorf("failed to close some databases: %v", errs)
	}

	slog.Info("Closed all databases")
	return nil
}

// Stats returns statistics about open databases.
func (m *DBManager) Stats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"open_databases": len(m.databases),
		"user_ids":       m.getUserIDs(),
	}
}

func (m *DBManager) getUserIDs() []string {
	ids := make([]string, 0, len(m.databases))
	for id := range m.databases {
		ids = append(ids, id)
	}
	return ids
}

// DBFromContext extracts the database from context.
// This is a convenience type for context-based DB access.
type contextDBKey struct{}

// ContextWithDB returns a context with the database attached.
func ContextWithDB(ctx context.Context, db *DB) context.Context {
	return context.WithValue(ctx, contextDBKey{}, db)
}

// DBFromContext extracts the database from context.
// Returns nil if not present.
func DBFromContext(ctx context.Context) *DB {
	db, _ := ctx.Value(contextDBKey{}).(*DB)
	return db
}
