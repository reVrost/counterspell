package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
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
	AppMetadata   map[string]interface{} `json:"app_metadata"`
	UserMetadata  map[string]interface{} `json:"user_metadata"`
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

// JWTValidator validates Supabase JWTs.
type JWTValidator struct {
	supabaseURL string
	anonKey     string
	jwksURL     string
	keys        map[string]*ecdsa.PublicKey
	keysMu      sync.RWMutex
	lastFetch   time.Time
}

// NewJWTValidator creates a new JWT validator.
// The secret parameter is kept for backwards compatibility but not used for ES256.
func NewJWTValidator(secret string) *JWTValidator {
	return &JWTValidator{
		keys: make(map[string]*ecdsa.PublicKey),
	}
}

// NewJWTValidatorWithURL creates a validator that fetches JWKS from Supabase.
func NewJWTValidatorWithURL(supabaseURL, anonKey string) *JWTValidator {
	return &JWTValidator{
		supabaseURL: supabaseURL,
		anonKey:     anonKey,
		jwksURL:     supabaseURL + "/auth/v1/jwks",
		keys:        make(map[string]*ecdsa.PublicKey),
	}
}

// Validate validates a JWT and returns its claims.
func (v *JWTValidator) Validate(tokenString string) (*Claims, error) {
	// Parse without full validation to extract claims
	// We'll validate by calling Supabase's user endpoint
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

// getKey retrieves the public key for the given key ID.
func (v *JWTValidator) getKey(kid string) (*ecdsa.PublicKey, error) {
	v.keysMu.RLock()
	key, ok := v.keys[kid]
	v.keysMu.RUnlock()

	if ok {
		return key, nil
	}

	// Fetch JWKS
	if err := v.fetchJWKS(); err != nil {
		return nil, err
	}

	v.keysMu.RLock()
	key, ok = v.keys[kid]
	v.keysMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("key not found: %s", kid)
	}

	return key, nil
}

// JWKS represents a JSON Web Key Set.
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key.
type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

// fetchJWKS fetches the JWKS from Supabase.
func (v *JWTValidator) fetchJWKS() error {
	if v.jwksURL == "" {
		return errors.New("JWKS URL not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", v.jwksURL, nil)
	if err != nil {
		return err
	}

	// Add API key for authentication
	if v.anonKey != "" {
		req.Header.Set("apikey", v.anonKey)
		req.Header.Set("Authorization", "Bearer "+v.anonKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS fetch failed: %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return err
	}

	v.keysMu.Lock()
	defer v.keysMu.Unlock()

	for _, jwk := range jwks.Keys {
		if jwk.Kty != "EC" || jwk.Alg != "ES256" {
			continue
		}

		key, err := jwkToECDSA(jwk)
		if err != nil {
			continue
		}
		v.keys[jwk.Kid] = key
	}

	v.lastFetch = time.Now()
	return nil
}

// jwkToECDSA converts a JWK to an ECDSA public key.
func jwkToECDSA(jwk JWK) (*ecdsa.PublicKey, error) {
	// Decode X and Y coordinates (base64url encoded)
	xBytes, err := jwt.NewParser().DecodeSegment(jwk.X)
	if err != nil {
		return nil, fmt.Errorf("failed to decode X: %w", err)
	}

	yBytes, err := jwt.NewParser().DecodeSegment(jwk.Y)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Y: %w", err)
	}

	return &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}, nil
}

// ExtractBearerToken extracts the token from a "Bearer <token>" string.
func ExtractBearerToken(authHeader string) string {
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(authHeader, "Bearer ")
}
