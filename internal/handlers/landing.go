package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/revrost/code/counterspell/internal/ui"
)

// HandleLanding renders the landing page.
func (h *Handlers) HandleLanding(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if err := ui.LandingPage().Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleAuth renders the auth page (login/register).
func (h *Handlers) HandleAuth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if err := ui.AuthPage("Sign In").Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleRegister renders the register page (same as auth page with different title).
func (h *Handlers) HandleRegister(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if err := ui.AuthPage("Get Started").Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleOAuth initiates OAuth flow for GitHub or Google.
func (h *Handlers) HandleOAuth(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")

	if h.auth == nil {
		http.Error(w, "Auth service not configured", http.StatusNotImplemented)
		return
	}

	// Get redirect URL (the current host)
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	redirectURL := fmt.Sprintf("%s://%s/auth/callback", scheme, r.Host)

	// Generate OAuth URL
	oauthURL, err := h.auth.GetOAuthURL(provider, redirectURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to OAuth provider
	http.Redirect(w, r, oauthURL, http.StatusTemporaryRedirect)
}

// HandleAuthCallback handles OAuth callback from Supabase.
func (h *Handlers) HandleAuthCallback(w http.ResponseWriter, r *http.Request) {
	// Extract tokens from URL parameters or session
	// For now, we'll just set a dummy session cookie
	// In production, validate with Supabase

	// Get access token from query params
	accessToken := r.URL.Query().Get("access_token")

	if accessToken != "" {
		h.auth.SetSessionCookie(w, accessToken)
		// Redirect to board
		http.Redirect(w, r, "/board", http.StatusTemporaryRedirect)
		return
	}

	// If no token, redirect to auth page
	http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)
}

// HandleLogout clears the session.
func (h *Handlers) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if h.auth != nil {
		h.auth.ClearSessionCookie(w)
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// HandleAuthCheck checks if user is authenticated.
func (h *Handlers) HandleAuthCheck(w http.ResponseWriter, r *http.Request) {
	if h.auth == nil {
		w.Write([]byte(`{"authenticated": false}`))
		return
	}

	session, err := h.auth.GetSession(r)
	if err != nil {
		w.Write([]byte(`{"authenticated": false}`))
		return
	}

	w.Write([]byte(`{"authenticated": true}`))
	fmt.Printf("Session: %+v\n", session)
}
