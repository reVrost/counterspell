package auth

import (
	"net/http"
	"os"
	"time"
)

// SessionManager handles session cookies.
type SessionManager struct {
	secure bool // Use secure cookies in production
}

// NewSessionManager creates a new session manager.
func NewSessionManager() *SessionManager {
	return &SessionManager{
		secure: os.Getenv("ENV") == "production",
	}
}

// SetAccessToken sets the access token cookie.
func (s *SessionManager) SetAccessToken(w http.ResponseWriter, token string, maxAge time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     "sb-access-token",
		Value:    token,
		Path:     "/",
		MaxAge:   int(maxAge.Seconds()),
		Secure:   s.secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// SetRefreshToken sets the refresh token cookie.
func (s *SessionManager) SetRefreshToken(w http.ResponseWriter, token string, maxAge time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     "sb-refresh-token",
		Value:    token,
		Path:     "/",
		MaxAge:   int(maxAge.Seconds()),
		Secure:   s.secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// GetAccessToken returns the access token from cookies.
func (s *SessionManager) GetAccessToken(r *http.Request) string {
	cookie, err := r.Cookie("sb-access-token")
	if err != nil {
		return ""
	}
	return cookie.Value
}

// GetRefreshToken returns the refresh token from cookies.
func (s *SessionManager) GetRefreshToken(r *http.Request) string {
	cookie, err := r.Cookie("sb-refresh-token")
	if err != nil {
		return ""
	}
	return cookie.Value
}

// ClearSession clears all session cookies.
func (s *SessionManager) ClearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "sb-access-token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   s.secure,
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "sb-refresh-token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   s.secure,
		HttpOnly: true,
	})
}

// SetSession sets both access and refresh tokens.
func (s *SessionManager) SetSession(w http.ResponseWriter, accessToken, refreshToken string) {
	// Access token: expires in 1 hour (matching Supabase default)
	s.SetAccessToken(w, accessToken, 1*time.Hour)
	// Refresh token: expires in 7 days
	s.SetRefreshToken(w, refreshToken, 7*24*time.Hour)
}
