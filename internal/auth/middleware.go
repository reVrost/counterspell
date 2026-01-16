package auth

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/revrost/code/counterspell/internal/config"
)

// contextKey is used for storing values in context.
type contextKey string

const (
	// UserIDKey is the context key for the user ID.
	UserIDKey contextKey = "user_id"
	// ClaimsKey is the context key for the full JWT claims.
	ClaimsKey contextKey = "claims"
)

// Middleware provides authentication middleware.
type Middleware struct {
	cfg       *config.Config
	validator *JWTValidator
}

// NewMiddleware creates a new auth middleware.
func NewMiddleware(cfg *config.Config) *Middleware {
	var validator *JWTValidator
	if cfg.MultiTenant && cfg.SupabaseJWTSecret != "" {
		validator = NewJWTValidator(cfg.SupabaseJWTSecret)
	}
	return &Middleware{
		cfg:       cfg,
		validator: validator,
	}
}

// RequireAuth is middleware that requires a valid JWT.
// In single-player mode (MULTI_TENANT=false), it sets userID to "default".
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Single-player mode: skip auth, use default user
		if !m.cfg.MultiTenant {
			ctx = context.WithValue(ctx, UserIDKey, "default")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Multi-tenant mode: validate JWT
		userID, claims, err := m.validateRequest(r)
		if err != nil {
			slog.Debug("Auth failed", "error", err, "path", r.URL.Path)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Add user context
		ctx = context.WithValue(ctx, UserIDKey, userID)
		if claims != nil {
			ctx = context.WithValue(ctx, ClaimsKey, claims)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth is middleware that validates JWT if present, but doesn't require it.
// In single-player mode, it sets userID to "default".
// In multi-tenant mode without a valid token, userID is empty.
func (m *Middleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Single-player mode: always use default user
		if !m.cfg.MultiTenant {
			ctx = context.WithValue(ctx, UserIDKey, "default")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Multi-tenant mode: try to validate, but don't fail if missing
		userID, claims, err := m.validateRequest(r)
		if err == nil && userID != "" {
			ctx = context.WithValue(ctx, UserIDKey, userID)
			if claims != nil {
				ctx = context.WithValue(ctx, ClaimsKey, claims)
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateRequest extracts and validates the JWT from the request.
func (m *Middleware) validateRequest(r *http.Request) (string, *Claims, error) {
	// Try Authorization header first
	authHeader := r.Header.Get("Authorization")
	token := ExtractBearerToken(authHeader)

	// Fall back to cookie
	if token == "" {
		cookie, err := r.Cookie("sb-access-token")
		if err == nil {
			token = cookie.Value
		}
	}

	// No token found
	if token == "" {
		return "", nil, ErrInvalidToken
	}

	// Validate the token
	if m.validator == nil {
		return "", nil, ErrInvalidToken
	}

	claims, err := m.validator.Validate(token)
	if err != nil {
		return "", nil, err
	}

	return claims.UserID(), claims, nil
}

// UserIDFromContext extracts the user ID from the context.
// Returns empty string if not present.
func UserIDFromContext(ctx context.Context) string {
	userID, _ := ctx.Value(UserIDKey).(string)
	return userID
}

// ClaimsFromContext extracts the JWT claims from the context.
// Returns nil if not present.
func ClaimsFromContext(ctx context.Context) *Claims {
	claims, _ := ctx.Value(ClaimsKey).(*Claims)
	return claims
}

// MustUserID extracts the user ID from context, panics if not present.
// Use only after RequireAuth middleware.
func MustUserID(ctx context.Context) string {
	userID := UserIDFromContext(ctx)
	if userID == "" {
		panic("MustUserID called without RequireAuth middleware")
	}
	return userID
}
