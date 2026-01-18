package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/revrost/code/counterspell/internal/config"
)

func TestMiddleware_NoSupabase(t *testing.T) {
	t.Parallel()

	// When no Supabase is configured, middleware allows all requests with "default" userID
	cfg := &config.Config{}
	m := NewMiddleware(cfg)

	handler := m.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := UserIDFromContext(r.Context())
		if userID != "default" {
			t.Errorf("userID = %q, want 'default'", userID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/app", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestMiddleware_WithSupabaseNoToken(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		SupabaseURL:       "https://example.supabase.co",
		SupabaseJWTSecret: "test-secret-key-at-least-32-chars-long",
	}
	m := NewMiddleware(cfg)

	handler := m.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called without token")
	}))

	req := httptest.NewRequest("GET", "/app", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestUserIDFromContext(t *testing.T) {
	t.Parallel()

	t.Run("empty context", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		if got := UserIDFromContext(ctx); got != "" {
			t.Errorf("UserIDFromContext(empty) = %q, want empty", got)
		}
	})

	t.Run("with user ID", func(t *testing.T) {
		t.Parallel()
		ctx := context.WithValue(context.Background(), UserIDKey, "user-123")
		if got := UserIDFromContext(ctx); got != "user-123" {
			t.Errorf("UserIDFromContext = %q, want 'user-123'", got)
		}
	})
}

func TestOptionalAuth_NoSupabase(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{}
	m := NewMiddleware(cfg)

	handler := m.OptionalAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := UserIDFromContext(r.Context())
		if userID != "default" {
			t.Errorf("userID = %q, want 'default'", userID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestOptionalAuth_WithSupabaseNoToken(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		SupabaseURL:       "https://example.supabase.co",
		SupabaseJWTSecret: "test-secret-key-at-least-32-chars-long",
	}
	m := NewMiddleware(cfg)

	handler := m.OptionalAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := UserIDFromContext(r.Context())
		if userID != "" {
			t.Errorf("userID = %q, want empty (no token)", userID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}
