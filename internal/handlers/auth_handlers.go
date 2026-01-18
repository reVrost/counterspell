package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v5"
	"github.com/revrost/code/counterspell/internal/models"
)

// getRedirectURL returns the appropriate redirect URL for the current environment.
// Always returns a relative path to work with any frontend (ngrok, localhost, production).

func (h *Handlers) HandleOAuthLogin(w http.ResponseWriter, r *http.Request) {
	if h.authService == nil {
		// Direct GitHub OAuth for self host single tenant
		h.HandleGitHubAuthorize(w, r)
	}

	// Use Supabase OAuth flow if auth service is configured
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	host := r.Host

	var callbackURL string
	if scheme == "" {
		callbackURL = host + "/api/v1/auth/callback"
	} else {
		callbackURL = fmt.Sprintf("%s://%s/api/v1/auth/callback", scheme, host)
	}

	// 1. Generate a random "Code Verifier" (save this in a secure cookie!)
	codeVerifier := "arandomstring123" // TODO: Generate a random string
	// 2. Hash it to create the "Code Challenge"
	hash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(hash[:])

	// 3. Store the verifier in a cookie so we can use it in the callback
	secure := os.Getenv("ENV") == "production"
	slog.Info("am i secure?", "secure", secure)
	http.SetCookie(w, &http.Cookie{
		Name:     "sb-code-verifier",
		Value:    codeVerifier,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure, // Set to false if testing locally without HTTPS
	})

	// // Get redirect URL from query params (frontend tells us where to go after auth)
	// // Pass it as state parameter so it survives the OAuth round trip
	// frontendRedirect := r.URL.Query().Get("redirect_url")
	// if frontendRedirect == "" {
	// 	frontendRedirect = "http://localhost:5173/dashboard"
	// }
	http.SetCookie(w, &http.Cookie{
		Name:  "return_to",
		Value: "http://localhost:5173/dashboard",
		Path:  "/",
	})

	supabaseAuthURL := h.authService.GetOAuthURL("github", callbackURL, codeChallenge, "")

	slog.Info("Redirecting to SupabaseAuthURL", "url", supabaseAuthURL, "callback", callbackURL)
	http.Redirect(w, r, supabaseAuthURL, http.StatusTemporaryRedirect)
}

func (h *Handlers) HandleAuthCallback(w http.ResponseWriter, r *http.Request) {
	slog.Info("[AUTH] HandleAuthCallback called", "url", r.URL.String())

	// 1. Check for standard OAuth error params from Supabase
	if errCode := r.URL.Query().Get("error"); errCode != "" {
		errDesc := r.URL.Query().Get("error_description")
		slog.Error("Supabase OAuth error", "code", errCode, "description", errDesc)
		http.Redirect(w, r, "/?error="+errCode, http.StatusTemporaryRedirect)
		return
	}

	// 2. Get the 'code' from the query string
	authCode := r.URL.Query().Get("code")
	if authCode == "" {
		slog.Error("No auth code present in callback")
		http.Redirect(w, r, "/?error=no_code", http.StatusTemporaryRedirect)
		return
	}
	// Debug: Log all cookies to see what the browser sent
	for _, c := range r.Cookies() {
		slog.Info("Cookie received", "name", c.Name)
	}

	// 3. Retrieve the PKCE verifier we saved in the cookie during HandleOAuth
	verifierCookie, err := r.Cookie("sb-code-verifier")
	if err != nil {
		slog.Error("PKCE verifier cookie missing", "error", err)
		http.Redirect(w, r, "/?error=session_expired", http.StatusTemporaryRedirect)
		return
	}

	// 4. EXCHANGE the code for a session (This is the critical PKCE step)
	// This call happens server-to-server. It returns the actual tokens.
	slog.Info("Attempting exchange", "code", authCode, "verifier", verifierCookie.Value)
	session, err := h.authService.ExchangeCode(r.Context(), authCode, verifierCookie.Value)
	if err != nil {
		slog.Error("Failed to exchange code for session", "error", err)
		http.Redirect(w, r, "/?error=exchange_failed", http.StatusTemporaryRedirect)
		return
	}

	// Cleanup the verifier cookie
	http.SetCookie(w, &http.Cookie{Name: "sb-code-verifier", MaxAge: -1, Path: "/"})

	// 5. Extract tokens from the session object returned by the exchange
	accessToken := session.AccessToken
	refreshToken := session.RefreshToken
	providerToken := session.ProviderToken // This is the GitHub PAT we need
	userID := session.User.ID              // Supabase Go SDK usually provides the user object here

	slog.Info("Supabase exchange successful",
		"user_id", userID,
		"has_provider_token", providerToken != "")

	// 6. Set your local session cookies
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

	// 7. If we have the GitHub token, sync everything
	if providerToken != "" {
		login, avatarURL, err := h.githubService.GetUserInfo(r.Context(), providerToken)
		if err != nil {
			slog.Error("Failed to get GitHub user info", "error", err)
		} else {
			err = h.githubService.SaveConnection(r.Context(), userID, "user", login, avatarURL, providerToken, "repo,read:user,read:org")
			if err != nil {
				slog.Error("Failed to save GitHub connection", "error", err)
			} else {
				slog.Info("GitHub connection saved", "login", login)

				// Sync repos in background
				go func(token, uid, loginName string) {
					bgCtx := context.Background()
					if err := h.repoCache.SyncReposFromGitHub(bgCtx, uid, token); err != nil {
						slog.Error("[OAuth] Cache sync failed", "error", err)
					}

					conn := &models.GitHubConnection{Type: "user", Login: loginName, Token: token}
					if err := h.githubService.FetchAndSaveRepositories(bgCtx, uid, conn); err != nil {
						slog.Error("[OAuth] DB save failed", "error", err)
					}
				}(providerToken, userID, login)
			}
		}
	}

	// 8. Final Redirect
	redirectURL := "http://localhost:5173/dashboard" // Default
	if c, err := r.Cookie("return_to"); err == nil {
		redirectURL = c.Value
		// Cleanup return_to cookie
		http.SetCookie(w, &http.Cookie{Name: "return_to", MaxAge: -1, Path: "/"})
	}

	slog.Info("Authentication complete", "redirecting_to", redirectURL)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
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

	_ = render.Render(w, r, Success("Logged out successfully"))
}

// HandleTokenRefresh refreshes an expired JWT using the refresh token
func (h *Handlers) HandleTokenRefresh(w http.ResponseWriter, r *http.Request) {
	if h.authService == nil {
		_ = render.Render(w, r, ErrInternalServer("Auth not configured", errors.New("Auth not configured")))
		return
	}
	slog.Info("[AUTH] HandleTokenRefresh called")

	// Get refresh token from cookie
	cookie, err := r.Cookie("sb-refresh-token")
	if err != nil || cookie.Value == "" {
		slog.Warn("Token refresh failed: no refresh token")
		_ = render.Render(w, r, ErrUnauthorized("No refresh token"))
		return
	}

	// Call Supabase to refresh
	tokens, err := h.authService.RefreshToken(cookie.Value)
	if err != nil {
		slog.Error("Token refresh failed", "error", err)
		_ = render.Render(w, r, ErrUnauthorized("Refresh failed"))
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
	render.NoContent(w, r)
}
