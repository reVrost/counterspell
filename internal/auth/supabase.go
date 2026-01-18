package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// SupabaseConfig holds Supabase configuration
type SupabaseConfig struct {
	URL string
	Key string
}

// Session represents the Supabase auth session returned by the token exchange
type Session struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	User         User   `json:"user"`
	// This is the GitHub PAT we need!
	ProviderToken string `json:"provider_token"`
}

type User struct {
	ID    string         `json:"id"`
	Email string         `json:"email"`
	Data  map[string]any `json:"user_metadata"`
}

// AuthService handles authentication with Supabase
type AuthService struct {
	supabaseURL string
	anonKey     string
}

// NewAuthService creates a new auth service
func NewAuthService(config *SupabaseConfig) (*AuthService, error) {
	if config.URL == "" || config.Key == "" {
		return nil, fmt.Errorf("supabase URL and key are required")
	}

	slog.Info("Auth service initialized", "url", config.URL)

	return &AuthService{supabaseURL: config.URL, anonKey: config.Key}, nil
}

// NewAuthServiceFromEnv creates auth service from environment variables
func NewAuthServiceFromEnv() (*AuthService, error) {
	url := os.Getenv("SUPABASE_URL")
	key := os.Getenv("SUPABASE_ANON_KEY")

	if url == "" || key == "" {
		return nil, fmt.Errorf("SUPABASE_URL and SUPABASE_ANON_KEY must be set")
	}

	return NewAuthService(&SupabaseConfig{
		URL: url,
		Key: key,
	})
}

// ExchangeCode swaps the auth code and verifier for a full session
func (s *AuthService) ExchangeCode(ctx context.Context, code, verifier string) (*Session, error) {
	// 1. Use Form Data instead of JSON
	data := url.Values{}
	data.Set("grant_type", "pkce")
	data.Set("code", code)
	data.Set("code_verifier", verifier)

	u := s.supabaseURL + "/auth/v1/token"

	// 2. Create the request with the form-encoded string
	req, err := http.NewRequestWithContext(ctx, "POST", u, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	// 3. Set standard headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("apikey", s.anonKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("exchange failed: %s", string(body))
	}

	var session Session
	json.Unmarshal(body, &session)
	return &session, nil
}

// GetOAuthURL returns an OAuth URL for a provider (github, google)
func (s *AuthService) GetOAuthURL(provider, redirectURL, codeChallenge, frontendRedirect string) string {
	params := url.Values{}
	params.Set("provider", provider)
	params.Set("redirect_to", redirectURL)
	// params.Set("state", frontendRedirect)
	params.Set("code_challenge", codeChallenge)
	params.Set("code_challenge_method", "s256")
	// params.Set("skip_http_redirect", "true")

	// Request GitHub scopes to get provider_token for repo access
	if provider == "github" {
		params.Set("scopes", "repo,read:user,read:org")
	}

	oauthURL := fmt.Sprintf("%s/auth/v1/authorize?%s", s.supabaseURL, params.Encode())

	slog.Info("Generated OAuth URL", "provider", provider, "url", oauthURL)

	return oauthURL
}

// GetSession retrieves the current session from the request
func (s *AuthService) GetSession(r *http.Request) (map[string]any, error) {
	// Get session from cookie
	cookie, err := r.Cookie("sb-access-token")

	if err != nil {
		return nil, fmt.Errorf("no session found")
	}

	// For now, we'll just check if token exists
	// In production, validate with Supabase API
	return map[string]any{
		"token": cookie.Value,
		"valid": len(cookie.Value) > 0,
	}, nil
}

// SetSessionCookie sets the session cookie
func (s *AuthService) SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "sb-access-token",
		Value:    token,
		Path:     "/",
		MaxAge:   3600 * 24 * 7, // 7 days
		Secure:   os.Getenv("ENV") == "production",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearSessionCookies clears all Supabase session cookies
func (s *AuthService) ClearSessionCookies(w http.ResponseWriter) {
	secure := os.Getenv("ENV") == "production"

	http.SetCookie(w, &http.Cookie{
		Name:     "sb-access-token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   secure,
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "sb-refresh-token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   secure,
		HttpOnly: true,
	})

	slog.Info("Supabase session cookies cleared")
}

// RefreshTokenResponse contains the response from Supabase token refresh
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// RefreshToken uses a refresh token to get a new access token from Supabase
func (s *AuthService) RefreshToken(refreshToken string) (*RefreshTokenResponse, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("refresh token is empty")
	}

	reqBody := map[string]string{"refresh_token": refreshToken}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/auth/v1/token?grant_type=refresh_token", s.supabaseURL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", s.anonKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("Supabase token refresh failed", "status", resp.StatusCode, "body", string(body))
		return nil, fmt.Errorf("refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp RefreshTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	slog.Info("Token refreshed successfully", "expires_in", tokenResp.ExpiresIn)
	return &tokenResp, nil
}
