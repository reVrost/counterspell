package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// getRedirectURL returns the appropriate redirect URL for the current environment.
// In dev mode (Vite on :5173), uses Origin header or FRONTEND_URL env var.
// In prod, uses PUBLIC_URL or the request host.
func getRedirectURL(r *http.Request, path string) string {
	// Check for explicit frontend URL (useful for dev mode)
	if frontendURL := os.Getenv("FRONTEND_URL"); frontendURL != "" {
		return strings.TrimSuffix(frontendURL, "/") + path
	}

	// Check for PUBLIC_URL (production)
	if publicURL := os.Getenv("PUBLIC_URL"); publicURL != "" {
		return strings.TrimSuffix(publicURL, "/") + path
	}

	// Use Origin header (from frontend making requests)
	if origin := r.Header.Get("Origin"); origin != "" {
		return strings.TrimSuffix(origin, "/") + path
	}

	// Fall back to request host
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s%s", scheme, r.Host, path)
}

func (h *Handlers) HandleOAuth(w http.ResponseWriter, r *http.Request) {
	// Use Supabase OAuth flow if auth service is configured
	if h.authService != nil {
		scheme := "http"
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		host := r.Host
		if envHost := os.Getenv("PUBLIC_URL"); envHost != "" {
			host = envHost
			scheme = ""
		}

		var callbackURL string
		if scheme == "" {
			callbackURL = host + "/auth/callback"
		} else {
			callbackURL = fmt.Sprintf("%s://%s/auth/callback", scheme, host)
		}

		oauthURL, err := h.authService.GetOAuthURL("github", callbackURL)
		if err != nil {
			slog.Error("Failed to get Supabase OAuth URL", "error", err)
			http.Error(w, "OAuth error", http.StatusInternalServerError)
			return
		}

		slog.Info("Redirecting to Supabase OAuth", "url", oauthURL, "callback", callbackURL)
		http.Redirect(w, r, oauthURL, http.StatusTemporaryRedirect)
		return
	}

	// Direct GitHub OAuth
	h.HandleGitHubAuthorize(w, r)
}

func (h *Handlers) HandleAuthCallback(w http.ResponseWriter, r *http.Request) {
	// Handle Supabase callback if auth service is configured
	if h.authService != nil {
		if errCode := r.URL.Query().Get("error"); errCode != "" {
			errDesc := r.URL.Query().Get("error_description")
			slog.Error("Supabase OAuth error", "code", errCode, "description", errDesc)
			http.Redirect(w, r, "/?error="+errCode+"&error_description="+errDesc, http.StatusTemporaryRedirect)
			return
		}

		accessToken := r.URL.Query().Get("access_token")
		refreshToken := r.URL.Query().Get("refresh_token")
		providerToken := r.URL.Query().Get("provider_token")

		slog.Info("Supabase callback received",
			"has_access_token", accessToken != "",
			"has_refresh_token", refreshToken != "",
			"has_provider_token", providerToken != "")

		if accessToken != "" {
			h.authService.SetSessionCookie(w, accessToken)
			if refreshToken != "" {
				http.SetCookie(w, &http.Cookie{
					Name:     "sb-refresh-token",
					Value:    refreshToken,
					Path:     "/",
					MaxAge:   3600 * 24 * 7,
					HttpOnly: true,
					SameSite: http.SameSiteLaxMode,
				})
			}

			userID := h.extractUserIDFromToken(accessToken)
			if userID == "" {
				slog.Error("Failed to extract user ID from token")
				http.Redirect(w, r, "/?error=invalid_token", http.StatusTemporaryRedirect)
				return
			}

			// If we have a GitHub provider token, save connection and fetch repos
			if providerToken != "" {
				login, avatarURL, err := h.githubService.GetUserInfo(r.Context(), providerToken)
				if err != nil {
					slog.Error("Failed to get GitHub user info", "error", err)
				} else {
					err = h.githubService.SaveConnection(r.Context(), userID, "user", login, avatarURL, providerToken, "repo,read:user,read:org")
					if err != nil {
						slog.Error("Failed to save GitHub connection", "error", err)
					} else {
						slog.Info("GitHub connection saved for user", "login", login, "user_id", userID)
						// Sync repos to cache in background
						go func(token, uid string) {
							if err := h.repoCache.SyncReposFromGitHub(context.Background(), uid, token); err != nil {
								slog.Error("[OAuth] Failed to sync repos to cache", "error", err)
							} else {
								slog.Info("[OAuth] Repos synced to cache successfully")
							}
						}(providerToken, userID)
					}
				}
			} else {
				slog.Warn("No provider_token in Supabase callback - repos won't be fetched")
			}

			redirectURL := getRedirectURL(r, "/dashboard")
			slog.Info("Supabase OAuth successful, redirecting", "url", redirectURL)
			http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
			return
		}

		// No token - serve page that extracts fragment
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>Completing login...</title></head>
<body>
<script>
const hash = window.location.hash.substring(1);
if (hash) {
	const params = new URLSearchParams(hash);
	const accessToken = params.get('access_token');
	const refreshToken = params.get('refresh_token');
	const providerToken = params.get('provider_token');
	if (accessToken) {
		let url = '/auth/callback?access_token=' + encodeURIComponent(accessToken);
		if (refreshToken) url += '&refresh_token=' + encodeURIComponent(refreshToken);
		if (providerToken) url += '&provider_token=' + encodeURIComponent(providerToken);
		window.location.href = url;
	} else {
		window.location.href = '/?error=no_token';
	}
} else {
	const params = new URLSearchParams(window.location.search);
	if (params.get('error')) {
		window.location.href = '/?error=' + params.get('error') + '&error_description=' + (params.get('error_description') || '');
	} else {
		window.location.href = '/?error=invalid_callback';
	}
}
</script>
<p>Completing login...</p>
</body>
</html>`))
		return
	}

	// Direct GitHub callback
	h.HandleGitHubCallback(w, r)
}

// extractUserIDFromToken extracts the user ID (sub claim) from a Supabase JWT
func (h *Handlers) extractUserIDFromToken(tokenString string) string {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		slog.Error("Failed to parse JWT", "error", err)
		return ""
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return ""
	}

	if sub, ok := claims["sub"].(string); ok {
		return sub
	}
	return ""
}

func (h *Handlers) HandleLogout(w http.ResponseWriter, r *http.Request) {
	slog.Info("[AUTH] HandleLogout called")

	// Clear auth session cookies
	if h.authService != nil {
		h.authService.ClearSessionCookies(w)
		slog.Info("[AUTH] Cleared auth session")
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Logged out successfully",
	})
}

// HandleTokenRefresh refreshes an expired JWT using the refresh token
func (h *Handlers) HandleTokenRefresh(w http.ResponseWriter, r *http.Request) {
	if h.authService == nil {
		http.Error(w, "Auth not configured", http.StatusServiceUnavailable)
		return
	}
	slog.Info("[AUTH] HandleTokenRefresh called")

	// Get refresh token from cookie
	cookie, err := r.Cookie("sb-refresh-token")
	if err != nil || cookie.Value == "" {
		slog.Warn("Token refresh failed: no refresh token")
		http.Error(w, "No refresh token", http.StatusUnauthorized)
		return
	}

	// Call Supabase to refresh
	tokens, err := h.authService.RefreshToken(cookie.Value)
	if err != nil {
		slog.Error("Token refresh failed", "error", err)
		http.Error(w, "Refresh failed", http.StatusUnauthorized)
		return
	}

	// Set new cookies
	secure := os.Getenv("ENV") == "production"
	http.SetCookie(w, &http.Cookie{
		Name:     "sb-access-token",
		Value:    tokens.AccessToken,
		Path:     "/",
		MaxAge:   tokens.ExpiresIn,
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "sb-refresh-token",
		Value:    tokens.RefreshToken,
		Path:     "/",
		MaxAge:   3600 * 24 * 7, // 7 days
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	slog.Info("Token refreshed via endpoint", "expires_in", tokens.ExpiresIn)
	w.WriteHeader(http.StatusOK)
}
