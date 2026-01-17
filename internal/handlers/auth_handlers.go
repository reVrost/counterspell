package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/revrost/code/counterspell/internal/db"
	"github.com/revrost/code/counterspell/internal/services"
)

func (h *Handlers) HandleOAuth(w http.ResponseWriter, r *http.Request) {
	// Multi-tenant: use Supabase OAuth flow
	if h.cfg != nil && h.cfg.MultiTenant && h.auth != nil {
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

		oauthURL, err := h.auth.GetOAuthURL("github", callbackURL)
		if err != nil {
			slog.Error("Failed to get Supabase OAuth URL", "error", err)
			http.Error(w, "OAuth error", http.StatusInternalServerError)
			return
		}

		slog.Info("Redirecting to Supabase OAuth", "url", oauthURL, "callback", callbackURL)
		http.Redirect(w, r, oauthURL, http.StatusTemporaryRedirect)
		return
	}

	// Single-player: direct GitHub OAuth
	h.HandleGitHubAuthorize(w, r)
}

func (h *Handlers) HandleAuthCallback(w http.ResponseWriter, r *http.Request) {
	// Multi-tenant: handle Supabase callback
	if h.cfg != nil && h.cfg.MultiTenant && h.auth != nil {
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
			h.auth.SetSessionCookie(w, accessToken)
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

			// Get user manager and services
			um, err := h.registry.Get(r.Context(), userID)
			if err != nil {
				slog.Error("Failed to get user manager", "error", err, "user_id", userID)
				http.Redirect(w, r, "/?error=db_error", http.StatusTemporaryRedirect)
				return
			}

			// Create services with user's DB
			userDB := um.DB()
			githubSvc := services.NewGitHubService(h.clientID, h.clientSecret, h.redirectURI, userDB)

			// If we have a GitHub provider token, save connection and fetch repos
			if providerToken != "" {
				login, avatarURL, err := githubSvc.GetUserInfo(r.Context(), providerToken)
				if err != nil {
					slog.Error("Failed to get GitHub user info", "error", err)
				} else {
					err = githubSvc.SaveConnection(r.Context(), "user", login, avatarURL, providerToken, "repo,read:user,read:org")
					if err != nil {
						slog.Error("Failed to save GitHub connection", "error", err)
					} else {
						slog.Info("GitHub connection saved for multi-tenant user", "login", login, "user_id", userID)
						// Sync repos to cache in background
						go func(token string, userDB *db.DB) {
							cache := services.NewRepoCache(userDB)
							if err := cache.SyncReposFromGitHub(context.Background(), token); err != nil {
								slog.Error("[OAuth] Failed to sync repos to cache", "error", err)
							} else {
								slog.Info("[OAuth] Repos synced to cache successfully")
							}
						}(providerToken, userDB)
					}
				}
			} else {
				slog.Warn("No provider_token in Supabase callback - repos won't be fetched")
			}

			slog.Info("Supabase OAuth successful, redirecting to app")
			http.Redirect(w, r, "/app", http.StatusTemporaryRedirect)
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

	// Single-player: direct GitHub callback
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
	h.HandleDisconnect(w, r)
}

// HandleTokenRefresh refreshes an expired JWT using the refresh token
func (h *Handlers) HandleTokenRefresh(w http.ResponseWriter, r *http.Request) {
	if h.auth == nil {
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
	tokens, err := h.auth.RefreshToken(cookie.Value)
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
