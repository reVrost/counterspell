package auth

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/revrost/invoker/internal/db"
)

// EdgeMiddleware provides middleware for edge request validation
type EdgeMiddleware struct {
	supabase *SupabaseAuth
	db       db.Repository
}

// NewEdgeMiddleware creates a new edge middleware
func NewEdgeMiddleware(supabase *SupabaseAuth, database db.Repository) *EdgeMiddleware {
	return &EdgeMiddleware{
		supabase: supabase,
		db:       database,
	}
}

// RequireUserSession validates that an incoming request to *.counterspell.app
// is from an authenticated user who owns the subdomain
func (m *EdgeMiddleware) RequireUserSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract subdomain from Host header
		subdomain := extractSubdomain(r.Host)
		if subdomain == "" {
			http.Error(w, "Invalid subdomain", http.StatusBadRequest)
			return
		}

		// Get session token from cookie or Authorization header
		sessionToken := extractSessionToken(r)
		if sessionToken == "" {
			http.Error(w, "Unauthorized - no session", http.StatusUnauthorized)
			return
		}

		// Validate Supabase session
		claims, err := m.supabase.ValidateToken(sessionToken)
		if err != nil {
			slog.Error("Invalid session token", "error", err)
			http.Error(w, "Unauthorized - invalid session", http.StatusUnauthorized)
			return
		}

		userID := ExtractUserID(claims)
		if userID == "" {
			http.Error(w, "Unauthorized - no user ID", http.StatusUnauthorized)
			return
		}

		ctx := r.Context()

		// Verify user owns the subdomain
		machineAuth, err := m.db.GetMachineAuthBySubdomain(ctx, subdomain)
		if err != nil {
			slog.Error("Machine not found for subdomain", "subdomain", subdomain, "error", err)
			http.Error(w, "Machine not found", http.StatusNotFound)
			return
		}

		if machineAuth.UserID != userID {
			slog.Warn("User does not own subdomain", "user_id", userID, "subdomain", subdomain, "owner_id", machineAuth.UserID)
			http.Error(w, "Forbidden - not your machine", http.StatusForbidden)
			return
		}

		if !machineAuth.IsActive {
			slog.Warn("Machine is inactive", "subdomain", subdomain)
			http.Error(w, "Machine is inactive", http.StatusForbidden)
			return
		}

		// Inject identity headers for downstream (Data Plane)
		r.Header.Set("X-Counterspell-User-ID", userID)
		r.Header.Set("X-Counterspell-Session-ID", claims.SessionID)

		// Add to context
		ctx = context.WithValue(ctx, "user_id", userID)
		ctx = context.WithValue(ctx, "subdomain", subdomain)
		ctx = context.WithValue(ctx, "session_id", claims.SessionID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractSubdomain extracts subdomain from Host header
// Example: alice.counterspell.app -> alice
func extractSubdomain(host string) string {
	// Remove port if present
	if colonIndex := strings.Index(host, ":"); colonIndex > 0 {
		host = host[:colonIndex]
	}

	// Split by dots
	parts := strings.Split(host, ".")

	// Expecting format: <subdomain>.counterspell.app
	// or <subdomain>.counterspell.localhost for local dev
	if len(parts) >= 3 {
		// Return first part as subdomain
		return parts[0]
	}

	// For localhost development: subdomain.localhost
	if len(parts) == 2 && parts[1] == "localhost" {
		return parts[0]
	}

	return ""
}

// extractSessionToken extracts session token from request
// Checks Authorization header first, then session cookie
func extractSessionToken(r *http.Request) string {
	// Check Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		// Remove "Bearer " prefix if present
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			return authHeader[7:]
		}
		return authHeader
	}

	// Check session cookie
	cookie, err := r.Cookie("sb-access-token")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	return ""
}

// RequireEdgeIdentity validates that a request was authorized by Invoker (edge)
// This is used by the Data Plane (Counterspell binary) to validate requests
func RequireEdgeIdentity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read headers injected by edge
		userID := r.Header.Get("X-Counterspell-User-ID")
		sessionID := r.Header.Get("X-Counterspell-Session-ID")

		if userID == "" || sessionID == "" {
			slog.Warn("Missing edge identity headers")
			http.Error(w, "Unauthorized - missing edge identity", http.StatusUnauthorized)
			return
		}

		// Attach values to request context
		ctx := context.WithValue(r.Context(), "user_id", userID)
		ctx = context.WithValue(ctx, "session_id", sessionID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
