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
	// If no Supabase URL, validator is not set indicating local dev mode
	// No auth is required in this case
	if cfg.SupabaseURL != "" {
		// Use JWKS-based validation for ES256 tokens
		validator = NewJWTValidatorWithURL(cfg.SupabaseURL, cfg.SupabaseAnonKey)
	}
	return &Middleware{
		cfg:       cfg,
		validator: validator,
	}
}

// RequireAuth is middleware that requires a valid JWT.
// If no Supabase is configured (validator is nil), uses "default" user ID.
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// If no validator configured, use default user (local dev mode)
		if m.validator == nil {
			ctx = context.WithValue(ctx, UserIDKey, "default")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Validate JWT
		userID, claims, err := m.validateRequest(r)
		if err != nil {
			slog.Warn("Auth failed: unauthorized request",
				"error", err,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
				"auth_header_present", r.Header.Get("Authorization") != "",
				"cookie_present", func() bool { _, err := r.Cookie("sb-access-token"); return err == nil }())
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
// If no Supabase is configured (validator is nil), uses "default" user ID.
// Without a valid token and with Supabase configured, userID is empty.
func (m *Middleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// If no validator configured, use default user (local dev mode)
		if m.validator == nil {
			ctx = context.WithValue(ctx, UserIDKey, "default")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Try to validate, but don't fail if missing
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
			slog.Debug("Found token in cookie", "token_len", len(token))
		} else {
			slog.Debug("No cookie found", "error", err)
		}
	}

	// No token found
	if token == "" {
		slog.Debug("No token found in request")
		return "", nil, ErrInvalidToken
	}

	// Validate the token
	if m.validator == nil {
		slog.Error("JWT validator is nil")
		return "", nil, ErrInvalidToken
	}

	claims, err := m.validator.Validate(token)
	if err != nil {
		slog.Error("JWT validation failed", "error", err)
		return "", nil, err
	}

	slog.Info("JWT validated successfully", "user_id", claims.UserID(), "email", claims.Email)
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
