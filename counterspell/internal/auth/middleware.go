package auth

import (
	"context"
	"net/http"
)

// contextKey is used for storing values in context.
type contextKey string

const (
	// UserIDKey is context key for user ID.
	UserIDKey contextKey = "user_id"
)

// Middleware provides authentication middleware.
type Middleware struct {
	// No fields needed for local-first mode
}

// NewMiddleware creates a new auth middleware (no-op for local-first).
func NewMiddleware() *Middleware {
	return &Middleware{}
}

// RequireAuth is middleware that doesn't require auth for local-first mode.
// Always sets userID to "default".
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), UserIDKey, "default")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth is middleware that doesn't require auth for local-first mode.
// Always sets userID to "default".
func (m *Middleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), UserIDKey, "default")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// UserIDFromContext extracts user ID from context.
func UserIDFromContext(ctx context.Context) string {
	userID, _ := ctx.Value(UserIDKey).(string)
	return userID
}

// MustUserID extracts user ID from context, panics if not present.
func MustUserID(ctx context.Context) string {
	userID := UserIDFromContext(ctx)
	if userID == "" {
		panic("MustUserID called without auth middleware")
	}
	return userID
}
