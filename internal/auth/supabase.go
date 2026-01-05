package auth

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
)

// SupabaseConfig holds Supabase configuration
type SupabaseConfig struct {
	URL    string
	Key    string
}

// AuthService handles authentication with Supabase
type AuthService struct {
	supabaseURL string
}

// NewAuthService creates a new auth service
func NewAuthService(config *SupabaseConfig) (*AuthService, error) {
	if config.URL == "" || config.Key == "" {
		return nil, fmt.Errorf("supabase URL and key are required")
	}

	slog.Info("Auth service initialized", "url", config.URL)

	return &AuthService{supabaseURL: config.URL}, nil
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

// GetOAuthURL returns an OAuth URL for a provider (github, google)
func (s *AuthService) GetOAuthURL(provider, redirectURL string) (string, error) {
	params := url.Values{}
	params.Set("provider", provider)
	params.Set("redirect_to", redirectURL)
	params.Set("skip_http_redirect", "true")

	oauthURL := fmt.Sprintf("%s/auth/v1/authorize?%s", s.supabaseURL, params.Encode())

	slog.Info("Generated OAuth URL", "provider", provider, "url", oauthURL)

	return oauthURL, nil
}

// GetSession retrieves the current session from the request
func (s *AuthService) GetSession(r *http.Request) (map[string]interface{}, error) {
	// Get session from cookie
	cookie, err := r.Cookie("sb-access-token")

	if err != nil {
		return nil, fmt.Errorf("no session found")
	}

	// For now, we'll just check if token exists
	// In production, validate with Supabase API
	return map[string]interface{}{
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

// ClearSessionCookie clears the session cookie
func (s *AuthService) ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "sb-access-token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   os.Getenv("ENV") == "production",
		HttpOnly: true,
	})
}
