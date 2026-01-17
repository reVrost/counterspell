package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
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

	// Default to 'user' if no type specified (e.g., from landing page login)
	if connType == "" {
		connType = "user"
	}

	if code == "" {
		http.Error(w, "No code provided", http.StatusBadRequest)
		return
	}

	token, err := h.github.ExchangeCodeForToken(ctx, code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to exchange token: %v", err), http.StatusInternalServerError)
		return
	}

	login, avatarURL, err := h.github.GetUserInfo(ctx, token)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user info: %v", err), http.StatusInternalServerError)
		return
	}

	err = h.github.SaveConnection(ctx, connType, login, avatarURL, token, "repo,read:user,read:org")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to save connection: %v", err), http.StatusInternalServerError)
		return
	}

	// Fetch repos in background or sync
	if conn, err := h.github.GetActiveConnection(ctx); err == nil {
		_ = h.github.FetchAndSaveRepositories(ctx, conn)
	}

	http.Redirect(w, r, "/app", http.StatusTemporaryRedirect)
}

func (h *Handlers) HandleDisconnect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slog.Info("[DISCONNECT] HandleDisconnect called", "method", r.Method, "url", r.URL.String())
	
	if err := h.github.DeleteConnection(ctx); err != nil {
		slog.Error("[DISCONNECT] Failed to delete connection", "error", err)
	} else {
		slog.Info("[DISCONNECT] Successfully deleted connection")
	}
	
	if err := h.github.DeleteAllProjects(ctx); err != nil {
		slog.Error("[DISCONNECT] Failed to delete projects", "error", err)
	} else {
		slog.Info("[DISCONNECT] Successfully deleted projects")
	}
	
	slog.Info("[DISCONNECT] Redirecting to landing page")
	http.Redirect(w, r, "/", http.StatusSeeOther) // 303 converts POST to GET - goes to public landing
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
		// Build callback URL for our app
		scheme := "http"
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		host := r.Host
		if envHost := os.Getenv("PUBLIC_URL"); envHost != "" {
			// Use PUBLIC_URL if set (for production)
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
		// Supabase sends tokens via URL fragment (#access_token=...)
		// But if using PKCE flow, it sends code in query params
		
		// Check for error from Supabase
		if errCode := r.URL.Query().Get("error"); errCode != "" {
			errDesc := r.URL.Query().Get("error_description")
			slog.Error("Supabase OAuth error", "code", errCode, "description", errDesc)
			http.Redirect(w, r, "/?error="+errCode+"&error_description="+errDesc, http.StatusTemporaryRedirect)
			return
		}
		
		// Check for access_token in query (Supabase implicit flow)
		accessToken := r.URL.Query().Get("access_token")
		refreshToken := r.URL.Query().Get("refresh_token")
		
		if accessToken != "" {
			// Set session cookies
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
			slog.Info("Supabase OAuth successful, redirecting to app")
			http.Redirect(w, r, "/app", http.StatusTemporaryRedirect)
			return
		}
		
		// No token - Supabase uses URL fragment, need client-side handling
		// Serve a page that extracts the fragment and sends it to us
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>Completing login...</title></head>
<body>
<script>
// Supabase sends tokens in URL fragment
const hash = window.location.hash.substring(1);
if (hash) {
	const params = new URLSearchParams(hash);
	const accessToken = params.get('access_token');
	const refreshToken = params.get('refresh_token');
	if (accessToken) {
		// Redirect with tokens in query params so server can set cookies
		window.location.href = '/auth/callback?access_token=' + encodeURIComponent(accessToken) + 
			(refreshToken ? '&refresh_token=' + encodeURIComponent(refreshToken) : '');
	} else {
		window.location.href = '/?error=no_token';
	}
} else {
	// No fragment, check if we have an error
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

func (h *Handlers) HandleAuthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"authenticated": true}`))
}

func (h *Handlers) HandleLogout(w http.ResponseWriter, r *http.Request) {
	h.HandleDisconnect(w, r)
}
