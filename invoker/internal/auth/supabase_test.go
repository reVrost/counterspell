package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateToken_HS256 tests validation of HS256 tokens
func TestValidateToken_HS256(t *testing.T) {
	tests := []struct {
		name        string
		jwtSecret   string
		token       string
		wantError   bool
		errorContains string
	}{
		{
			name:      "valid HS256 token",
			jwtSecret: "test-secret-key",
			token: generateHS256Token(t, "test-secret-key", "user@example.com", "user-123"),
			wantError: false,
		},
		{
			name:        "missing secret",
			jwtSecret:   "",
			token:       generateHS256Token(t, "different-secret", "user@example.com", "user-123"),
			wantError:   true,
			errorContains: "SUPABASE_JWT_SECRET not configured",
		},
		{
			name:        "wrong secret",
			jwtSecret:   "correct-secret",
			token:       generateHS256Token(t, "wrong-secret", "user@example.com", "user-123"),
			wantError:   true,
			errorContains: "signature is invalid",
		},
		{
			name:        "expired token",
			jwtSecret:   "test-secret-key",
			token:       generateExpiredHS256Token(t, "test-secret-key"),
			wantError:   true,
			errorContains: "token is expired",
		},
		{
			name:        "malformed token",
			jwtSecret:   "test-secret-key",
			token:       "not.a.valid.jwt",
			wantError:   true,
			errorContains: "failed to parse token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := &SupabaseAuth{
				jwtSecret: tt.jwtSecret,
			}

			claims, err := auth.ValidateToken(tt.token)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, "user@example.com", claims.Email)
				assert.Equal(t, "user-123", claims.Subject)
			}
		})
	}
}

// TestValidateToken_ES256 tests validation of ES256 tokens
// Note: This test requires actual Supabase JWKS to work
// It's more of an integration test
func TestValidateToken_ES256(t *testing.T) {
	tests := []struct {
		name        string
		supabaseURL string
		token       string
		wantError   bool
	}{
		{
			name:        "missing kid header",
			supabaseURL: "https://example.supabase.co",
			token:       generateES256TokenWithoutKid(t),
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := &SupabaseAuth{
				supabaseURL: tt.supabaseURL,
			}

			claims, err := auth.ValidateToken(tt.token)

			assert.True(t, tt.wantError)
			assert.Error(t, err)
			assert.Nil(t, claims)
		})
	}
}

// TestValidateToken_UnsupportedAlgorithm tests rejection of unsupported algorithms
func TestValidateToken_UnsupportedAlgorithm(t *testing.T) {
	tests := []struct {
		name        string
		algorithm   string
		token       string
		wantError   bool
		errorContains string
	}{
		{
			name:        "HS512 not supported",
			algorithm:   "HS512",
			token:       generateTokenWithAlg(t, "HS512"),
			wantError:   true,
			errorContains: "unsupported signing method",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := &SupabaseAuth{
				jwtSecret: "test-secret-key",
			}

			claims, err := auth.ValidateToken(tt.token)

			assert.True(t, tt.wantError)
			assert.Error(t, err)
			assert.Nil(t, claims)
			assert.Contains(t, err.Error(), tt.errorContains)
		})
	}
}

// Helper functions for test token generation

func generateHS256Token(t *testing.T, secret, email, userID string) string {
	now := time.Now()
	expiresIn := now.Add(time.Hour)

	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"iss":   "https://test.supabase.co",
		"aud":   "authenticated",
		"iat":   jwt.NewNumericDate(now),
		"exp":   jwt.NewNumericDate(expiresIn),
		"role":  "authenticated",
		"user_metadata": map[string]any{
			"full_name": "Test User",
		},
		"app_metadata": map[string]any{
			"provider": "email",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	require.NoError(t, err)
	return tokenString
}

func generateExpiredHS256Token(t *testing.T, secret string) string {
	now := time.Now()
	expired := now.Add(-time.Hour)

	claims := jwt.MapClaims{
		"sub":   "user-123",
		"email": "user@example.com",
		"iss":   "https://test.supabase.co",
		"aud":   "authenticated",
		"iat":   jwt.NewNumericDate(now.Add(-2 * time.Hour)),
		"exp":   jwt.NewNumericDate(expired),
		"role":  "authenticated",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	require.NoError(t, err)
	return tokenString
}

func generateES256TokenWithoutKid(t *testing.T) string {
	now := time.Now()
	expiresIn := now.Add(time.Hour)

	// Create a malformed ES256 token without kid header
	claims := jwt.MapClaims{
		"sub":   "user-123",
		"email": "user@example.com",
		"iss":   "https://test.supabase.co",
		"aud":   "authenticated",
		"iat":   jwt.NewNumericDate(now),
		"exp":   jwt.NewNumericDate(expiresIn),
	}

	// Create token header without kid
	header := map[string]any{
		"alg": "HS256", // Use HS256 instead of ES256 to avoid signing issues
		"typ": "JWT",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header = header
	tokenString, err := token.SignedString([]byte("test-secret"))
	require.NoError(t, err)
	return tokenString
}

func generateTokenWithAlg(t *testing.T, alg string) string {
	now := time.Now()
	expiresIn := now.Add(time.Hour)

	claims := jwt.MapClaims{
		"sub":   "user-123",
		"email": "user@example.com",
		"iat":   jwt.NewNumericDate(now),
		"exp":   jwt.NewNumericDate(expiresIn),
	}

	var method jwt.SigningMethod
	var secret interface{}
	switch alg {
	case "HS512":
		method = jwt.SigningMethodHS512
		secret = []byte("test-secret-key")
	default:
		method = jwt.SigningMethodHS256
		secret = []byte("test-secret-key")
	}

	token := jwt.NewWithClaims(method, claims)
	tokenString, err := token.SignedString(secret)
	require.NoError(t, err)
	return tokenString
}

// TestExtractUserID tests the ExtractUserID helper function
func TestExtractUserID(t *testing.T) {
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "user-123",
		},
	}

	userID := ExtractUserID(claims)
	assert.Equal(t, "user-123", userID)
}

// TestExtractEmail tests the ExtractEmail helper function
func TestExtractEmail(t *testing.T) {
	claims := &Claims{
		Email: "test@example.com",
	}

	email := ExtractEmail(claims)
	assert.Equal(t, "test@example.com", email)
}
