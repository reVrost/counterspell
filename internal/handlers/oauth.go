package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/revrost/code/counterspell/internal/auth"
	"github.com/revrost/code/counterspell/internal/models"
	"github.com/revrost/code/counterspell/internal/services"
)

// HandleGitHubAuthorize initiates GitHub OAuth flow.
func (h *Handlers) HandleGitHubAuthorize(w http.ResponseWriter, r *http.Request) {
	connType := r.URL.Query().Get("type")
	fmt.Printf("GitHub authorize request - type: %s, clientID: %s\n", connType, h.clientID)

	redirectURI := h.redirectURI
	if redirectURI == "" {
		redirectURI = "http://localhost:8710/github/callback"
	}

	authURL := fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=repo,read:user,read:org&state=%s",
		h.clientID, redirectURI, connType)

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// HandleGitHubCallback handles GitHub OAuth return.
func (h *Handlers) HandleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	code := r.URL.Query().Get("code")
	connType := r.URL.Query().Get("state")

	if connType == "" {
		connType = "user"
	}

	if code == "" {
		http.Error(w, "No code provided", http.StatusBadRequest)
		return
	}

	svc, err := h.getServices(ctx)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	token, err := svc.GitHub.ExchangeCodeForToken(ctx, code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to exchange token: %v", err), http.StatusInternalServerError)
		return
	}

	login, avatarURL, err := svc.GitHub.GetUserInfo(ctx, token)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user info: %v", err), http.StatusInternalServerError)
		return
	}

	err = svc.GitHub.SaveConnection(ctx, connType, login, avatarURL, token, "repo,read:user,read:org")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to save connection: %v", err), http.StatusInternalServerError)
		return
	}

	// Fetch repos in background
	if conn, err := svc.GitHub.GetActiveConnection(ctx); err == nil {
		_ = svc.GitHub.FetchAndSaveRepositories(ctx, conn)
	}

	http.Redirect(w, r, "/app", http.StatusTemporaryRedirect)
}

func (h *Handlers) HandleDisconnect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slog.Info("[DISCONNECT] HandleDisconnect called", "method", r.Method, "url", r.URL.String())

	// Multi-tenant: clear Supabase session cookies
	if h.cfg != nil && h.cfg.MultiTenant && h.auth != nil {
		h.auth.ClearSessionCookies(w)
		slog.Info("[DISCONNECT] Cleared Supabase session")
	}

	svc, err := h.getServices(ctx)
	if err != nil {
		slog.Error("[DISCONNECT] Failed to get services", "error", err)
	} else {
		if err := svc.GitHub.DeleteConnection(ctx); err != nil {
			slog.Error("[DISCONNECT] Failed to delete connection", "error", err)
		} else {
			slog.Info("[DISCONNECT] Successfully deleted connection")
		}

		if err := svc.GitHub.DeleteAllProjects(ctx); err != nil {
			slog.Error("[DISCONNECT] Failed to delete projects", "error", err)
		} else {
			slog.Info("[DISCONNECT] Successfully deleted projects")
		}
	}

	slog.Info("[DISCONNECT] Redirecting to landing page")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Legacy/Stub Handlers
func (h *Handlers) HandleAuth(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/github/authorize", http.StatusTemporaryRedirect)
}

func (h *Handlers) HandleRegister(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/github/authorize", http.StatusTemporaryRedirect)
}

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
						// Fetch repos in background
						if conn, err := githubSvc.GetActiveConnection(r.Context()); err == nil {
							go func(c *models.GitHubConnection, svc *services.GitHubService) {
								if err := svc.FetchAndSaveRepositories(context.Background(), c); err != nil {
									slog.Error("Failed to fetch repositories", "error", err)
								} else {
									slog.Info("Repositories fetched successfully")
								}
							}(conn, githubSvc)
						}
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

func (h *Handlers) HandleAuthCheck(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID != "" && userID != "default" {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusUnauthorized)
}

func (h *Handlers) HandleLogout(w http.ResponseWriter, r *http.Request) {
	h.HandleDisconnect(w, r)
}
