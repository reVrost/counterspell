package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/go-chi/render"
)

// HandleGitHubLogin redirects to GitHub OAuth.
func (h *Handlers) HandleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	redirectURL := r.URL.Query().Get("redirect_url")
	if redirectURL == "" {
		redirectURL = "/dashboard"
	}

	// Build the callback URL - this is where GitHub will redirect after auth
	// In dev, this needs to go through the proxy, so we use the request host
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	callbackURL := fmt.Sprintf("%s://%s/api/v1/github/callback", scheme, r.Host)

	scope := "repo,user"
	githubURL := fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&scope=%s&redirect_uri=%s&state=%s",
		h.cfg.GitHubClientID,
		scope,
		url.QueryEscape(callbackURL),
		url.QueryEscape(redirectURL),
	)

	http.Redirect(w, r, githubURL, http.StatusTemporaryRedirect)
}

// HandleGitHubCallback handles the OAuth callback.
func (h *Handlers) HandleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state") // This is our redirect_url

	if code == "" {
		http.Error(w, "Code required", http.StatusBadRequest)
		return
	}

	// 1. Exchange code for token
	token, err := h.githubService.ExchangeCode(ctx, code)
	if err != nil {
		slog.Error("Failed to exchange github code", "error", err)
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	// 2. Get user info
	user, err := h.githubService.GetUserInfo(ctx, token)
	if err != nil {
		slog.Error("Failed to get github user info", "error", err)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	// 3. Sync connection
	connectionID, err := h.githubService.SyncConnection(ctx, user, token)
	if err != nil {
		slog.Error("Failed to sync github connection", "error", err)
		http.Error(w, "Failed to save connection", http.StatusInternalServerError)
		return
	}

	// 4. Fetch and sync repos
	repos, err := h.githubService.FetchUserRepos(ctx, token)
	if err != nil {
		slog.Error("Failed to fetch github repos", "error", err)
		// Don't fail the whole login if repo sync fails
	} else {
		if err := h.githubService.SyncRepos(ctx, connectionID, repos); err != nil {
			slog.Error("Failed to sync github repos", "error", err)
		}
	}

	// Redirect back to the app
	if state != "" {
		http.Redirect(w, r, state, http.StatusTemporaryRedirect)
	} else {
		http.Redirect(w, r, "/dashboard", http.StatusTemporaryRedirect)
	}
}

// HandleGitHubRepos returns the list of synced repositories.
func (h *Handlers) HandleGitHubRepos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	repos, err := h.githubService.GetRepos(ctx)
	if err != nil {
		slog.Error("Failed to get github repos from db", "error", err)
		render.JSON(w, r, []any{})
		return
	}

	render.JSON(w, r, repos)
}
