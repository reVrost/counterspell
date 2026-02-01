package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken is returned when the JWT is malformed or invalid.
var ErrInvalidToken = errors.New("invalid token")

// ErrTokenExpired is returned when the JWT has expired.
var ErrTokenExpired = errors.New("token expired")

// Claims represents the JWT claims from Supabase.
type Claims struct {
	jwt.RegisteredClaims

	// Supabase-specific claims
	Email         string                 `json:"email"`
	Role          string                 `json:"role"`
	AppMetadata   map[string]any `json:"app_metadata"`
	UserMetadata  map[string]any `json:"user_metadata"`
	AMR           []AMREntry             `json:"amr"` // Authentication Method Reference
	SessionID     string                 `json:"session_id"`
	IsAnonymous   bool                   `json:"is_anonymous"`
	Authenticator string                 `json:"authenticator"`
}

// AMREntry represents an authentication method reference.
type AMREntry struct {
	Method    string `json:"method"`
	Timestamp int64  `json:"timestamp"`
}

// UserID returns the Supabase user ID.
func (c *Claims) UserID() string {
	return c.Subject
}

// GithubUsername returns the GitHub username from user metadata.
func (c *Claims) GithubUsername() string {
	if c.UserMetadata == nil {
		return ""
	}
	if username, ok := c.UserMetadata["user_name"].(string); ok {
		return username
	}
	if username, ok := c.UserMetadata["preferred_username"].(string); ok {
		return username
	}
	return ""
}

// AvatarURL returns the avatar URL from user metadata.
func (c *Claims) AvatarURL() string {
	if c.UserMetadata == nil {
		return ""
	}
	if url, ok := c.UserMetadata["avatar_url"].(string); ok {
		return url
	}
	return ""
}

// JWTValidator validates Supabase JWTs via remote validation.
type JWTValidator struct {
	supabaseURL string
	anonKey     string
}

// NewJWTValidator creates a new JWT validator.
// The secret parameter is kept for backwards compatibility.
func NewJWTValidator(secret string) *JWTValidator {
	return &JWTValidator{}
}

// NewJWTValidatorWithURL creates a validator that validates tokens via Supabase API.
func NewJWTValidatorWithURL(supabaseURL, anonKey string) *JWTValidator {
	return &JWTValidator{
		supabaseURL: supabaseURL,
		anonKey:     anonKey,
	}
}

// Validate validates a JWT and returns its claims.
func (v *JWTValidator) Validate(tokenString string) (*Claims, error) {
	// Parse without full validation to extract claims
	// We validate by calling Supabase's user endpoint
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, fmt.Errorf("%w: parse error: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("%w: invalid claims type", ErrInvalidToken)
	}

	// Check expiration locally first
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, ErrTokenExpired
	}

	// Validate token by calling Supabase's user endpoint
	if err := v.validateWithSupabase(tokenString); err != nil {
		return nil, err
	}

	return claims, nil
}

// validateWithSupabase validates the token by calling Supabase's /auth/v1/user endpoint.
func (v *JWTValidator) validateWithSupabase(token string) error {
	if v.supabaseURL == "" {
		return fmt.Errorf("%w: supabase URL not configured", ErrInvalidToken)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", v.supabaseURL+"/auth/v1/user", nil)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	req.Header.Set("apikey", v.anonKey)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: request failed: %v", ErrInvalidToken, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: supabase returned %d", ErrInvalidToken, resp.StatusCode)
	}

	return nil
}

// ExtractBearerToken extracts the token from a "Bearer <token>" string.
func ExtractBearerToken(authHeader string) string {
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(authHeader, "Bearer ")
}
