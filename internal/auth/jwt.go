package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// ErrInvalidToken is returned when the JWT is malformed or invalid.
var ErrInvalidToken = errors.New("invalid token")

// ErrTokenExpired is returned when the JWT has expired.
var ErrTokenExpired = errors.New("token expired")

// Claims represents the JWT claims from Supabase.
type Claims struct {
	// Standard JWT claims
	Sub string `json:"sub"` // User ID (Supabase UUID)
	Aud string `json:"aud"` // Audience
	Exp int64  `json:"exp"` // Expiration time
	Iat int64  `json:"iat"` // Issued at
	Iss string `json:"iss"` // Issuer

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
	return c.Sub
}

// IsExpired checks if the token has expired.
func (c *Claims) IsExpired() bool {
	return time.Now().Unix() > c.Exp
}

// GithubProviderToken returns the GitHub access token if available from provider tokens.
func (c *Claims) GithubProviderToken() string {
	if c.AppMetadata == nil {
		return ""
	}
	providers, ok := c.AppMetadata["providers"].([]interface{})
	if !ok {
		return ""
	}
	for _, p := range providers {
		if provider, ok := p.(map[string]interface{}); ok {
			if provider["name"] == "github" {
				if token, ok := provider["access_token"].(string); ok {
					return token
				}
			}
		}
	}
	return ""
}

// JWTValidator validates Supabase JWTs locally.
type JWTValidator struct {
	secret []byte
}

// NewJWTValidator creates a new JWT validator with the given secret.
func NewJWTValidator(secret string) *JWTValidator {
	return &JWTValidator{
		secret: []byte(secret),
	}
}

// Validate validates a JWT and returns its claims.
func (v *JWTValidator) Validate(tokenString string) (*Claims, error) {
	// Split the token into parts
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	headerB64, payloadB64, signatureB64 := parts[0], parts[1], parts[2]

	// Verify signature
	if !v.verifySignature(headerB64, payloadB64, signatureB64) {
		return nil, ErrInvalidToken
	}

	// Decode payload
	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Parse claims
	var claims Claims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	// Check expiration
	if claims.IsExpired() {
		return nil, ErrTokenExpired
	}

	return &claims, nil
}

// verifySignature verifies the HMAC-SHA256 signature.
func (v *JWTValidator) verifySignature(header, payload, signature string) bool {
	// Create the signing input
	signingInput := header + "." + payload

	// Create HMAC-SHA256
	h := hmac.New(sha256.New, v.secret)
	h.Write([]byte(signingInput))
	expectedSig := h.Sum(nil)

	// Decode the provided signature
	providedSig, err := base64.RawURLEncoding.DecodeString(signature)
	if err != nil {
		return false
	}

	// Compare signatures
	return hmac.Equal(expectedSig, providedSig)
}

// ExtractBearerToken extracts the token from a "Bearer <token>" string.
func ExtractBearerToken(authHeader string) string {
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(authHeader, "Bearer ")
}
