package services

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/revrost/code/counterspell/internal/config"
	"github.com/revrost/code/counterspell/internal/db"
)

// ErrTooManyTasks is returned when a user has too many concurrent tasks.
var ErrTooManyTasks = errors.New("too many concurrent tasks")

// ErrUserManagerShutdown is returned when the user manager is shutting down.
var ErrUserManagerShutdown = errors.New("user manager is shutting down")

// UserMetrics tracks per-user statistics.
type UserMetrics struct {
	TasksCreated  int64
	TasksComplete int64
	TasksFailed   int64
	TokensUsed    int64
	LastActive    time.Time
}

// UserManager manages resources for a single user.
// Each user gets their own goroutine for serialized DB writes.
type UserManager struct {
	userID      string
	cfg         *config.Config
	dbManager   *db.DBManager
	database    *db.DB
	activeTasks int32
	lastUsed    time.Time
	metrics     UserMetrics
	writeCh     chan func()
	done        chan struct{}
	mu          sync.RWMutex
}

// NewUserManager creates a new user manager.
func NewUserManager(userID string, cfg *config.Config, dbManager *db.DBManager) (*UserManager, error) {
	database, err := dbManager.GetDB(userID)
	if err != nil {
		return nil, err
	}

	um := &UserManager{
		userID:    userID,
		cfg:       cfg,
		dbManager: dbManager,
		database:  database,
		lastUsed:  time.Now(),
		writeCh:   make(chan func(), 100),
		done:      make(chan struct{}),
	}

	// Start the write processor goroutine
	go um.processWrites()

	slog.Info("Created user manager", "user_id", userID)
	return um, nil
}

// processWrites processes serialized database writes.
func (um *UserManager) processWrites() {
	for {
		select {
		case fn := <-um.writeCh:
			fn()
		case <-um.done:
			// Drain remaining writes
			for {
				select {
				case fn := <-um.writeCh:
					fn()
				default:
					return
				}
			}
		}
	}
}

// Write queues a database write operation.
func (um *UserManager) Write(fn func()) error {
	select {
	case um.writeCh <- fn:
		return nil
	case <-um.done:
		return ErrUserManagerShutdown
	}
}

// DB returns the user's database.
func (um *UserManager) DB() *db.DB {
	return um.database
}

// UserID returns the user ID.
func (um *UserManager) UserID() string {
	return um.userID
}

// CanStartTask checks if the user can start a new task.
func (um *UserManager) CanStartTask() bool {
	return atomic.LoadInt32(&um.activeTasks) < int32(um.cfg.MaxTasksPerUser)
}

// IncrementTasks increments the active task count.
// Returns an error if the limit would be exceeded.
func (um *UserManager) IncrementTasks() error {
	for {
		current := atomic.LoadInt32(&um.activeTasks)
		if current >= int32(um.cfg.MaxTasksPerUser) {
			return ErrTooManyTasks
		}
		if atomic.CompareAndSwapInt32(&um.activeTasks, current, current+1) {
			return nil
		}
	}
}

// DecrementTasks decrements the active task count.
func (um *UserManager) DecrementTasks() {
	atomic.AddInt32(&um.activeTasks, -1)
}

// ActiveTasks returns the number of active tasks.
func (um *UserManager) ActiveTasks() int {
	return int(atomic.LoadInt32(&um.activeTasks))
}

// Touch updates the last used time.
func (um *UserManager) Touch() {
	um.mu.Lock()
	um.lastUsed = time.Now()
	um.mu.Unlock()
}

// LastUsed returns when the user manager was last used.
func (um *UserManager) LastUsed() time.Time {
	um.mu.RLock()
	defer um.mu.RUnlock()
	return um.lastUsed
}

// RecordTaskCreated records a task creation.
func (um *UserManager) RecordTaskCreated() {
	atomic.AddInt64(&um.metrics.TasksCreated, 1)
	um.Touch()
}

// RecordTaskComplete records a task completion.
func (um *UserManager) RecordTaskComplete() {
	atomic.AddInt64(&um.metrics.TasksComplete, 1)
}

// RecordTaskFailed records a task failure.
func (um *UserManager) RecordTaskFailed() {
	atomic.AddInt64(&um.metrics.TasksFailed, 1)
}

// RecordTokens records tokens used.
func (um *UserManager) RecordTokens(tokens int64) {
	atomic.AddInt64(&um.metrics.TokensUsed, tokens)
}

// Metrics returns the user's metrics.
func (um *UserManager) Metrics() UserMetrics {
	return UserMetrics{
		TasksCreated:  atomic.LoadInt64(&um.metrics.TasksCreated),
		TasksComplete: atomic.LoadInt64(&um.metrics.TasksComplete),
		TasksFailed:   atomic.LoadInt64(&um.metrics.TasksFailed),
		TokensUsed:    atomic.LoadInt64(&um.metrics.TokensUsed),
		LastActive:    um.LastUsed(),
	}
}

// Shutdown shuts down the user manager.
func (um *UserManager) Shutdown() {
	close(um.done)
	slog.Info("Shut down user manager", "user_id", um.userID)
}

// UserManagerRegistry manages all user managers.
type UserManagerRegistry struct {
	cfg       *config.Config
	dbManager *db.DBManager
	managers  map[string]*UserManager
	mu        sync.RWMutex
	done      chan struct{}
}

// NewUserManagerRegistry creates a new registry.
func NewUserManagerRegistry(cfg *config.Config, dbManager *db.DBManager) *UserManagerRegistry {
	r := &UserManagerRegistry{
		cfg:       cfg,
		dbManager: dbManager,
		managers:  make(map[string]*UserManager),
		done:      make(chan struct{}),
	}

	// Start cleanup goroutine
	go r.cleanupLoop()

	return r
}

// Get returns or creates a user manager for the given user ID.
func (r *UserManagerRegistry) Get(ctx context.Context, userID string) (*UserManager, error) {
	// Normalize to "default" in single-player mode
	if !r.cfg.MultiTenant {
		userID = "default"
	}

	// Check cache with read lock
	r.mu.RLock()
	if um, ok := r.managers[userID]; ok {
		r.mu.RUnlock()
		um.Touch()
		return um, nil
	}
	r.mu.RUnlock()

	// Create new manager with write lock
	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	if um, ok := r.managers[userID]; ok {
		um.Touch()
		return um, nil
	}

	um, err := NewUserManager(userID, r.cfg, r.dbManager)
	if err != nil {
		return nil, err
	}

	r.managers[userID] = um
	return um, nil
}

// cleanupLoop periodically cleans up inactive user managers.
func (r *UserManagerRegistry) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.cleanupInactive()
		case <-r.done:
			return
		}
	}
}

// cleanupInactive removes user managers that have been inactive too long.
func (r *UserManagerRegistry) cleanupInactive() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for userID, um := range r.managers {
		// Don't cleanup "default" user in single-player mode
		if !r.cfg.MultiTenant && userID == "default" {
			continue
		}

		// Don't cleanup if there are active tasks
		if um.ActiveTasks() > 0 {
			continue
		}

		if now.Sub(um.LastUsed()) > r.cfg.UserManagerTTL {
			um.Shutdown()
			delete(r.managers, userID)
			slog.Info("Cleaned up inactive user manager",
				"user_id", userID,
				"inactive_for", now.Sub(um.LastUsed()))
		}
	}
}

// Stats returns statistics about the registry.
func (r *UserManagerRegistry) Stats() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	userStats := make(map[string]interface{})
	for userID, um := range r.managers {
		userStats[userID] = map[string]interface{}{
			"active_tasks": um.ActiveTasks(),
			"last_used":    um.LastUsed(),
			"metrics":      um.Metrics(),
		}
	}

	return map[string]interface{}{
		"total_managers": len(r.managers),
		"users":          userStats,
	}
}

// Shutdown shuts down all user managers.
func (r *UserManagerRegistry) Shutdown() {
	close(r.done)

	r.mu.Lock()
	defer r.mu.Unlock()

	for userID, um := range r.managers {
		um.Shutdown()
		slog.Info("Shut down user manager during registry shutdown", "user_id", userID)
	}

	r.managers = make(map[string]*UserManager)
	slog.Info("User manager registry shut down")
}
