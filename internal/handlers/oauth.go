package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/revrost/code/counterspell/internal/auth"
	"github.com/revrost/code/counterspell/internal/models"
)

// HandleGitHubAuthorize initiates GitHub OAuth flow.
func (h *Handlers) HandleGitHubAuthorize(w http.ResponseWriter, r *http.Request) {
	connType := r.URL.Query().Get("type")
	fmt.Printf("GitHub authorize request - type: %s, clientID: %s\n", connType, h.clientID)

	redirectURI := h.redirectURI
	if redirectURI == "" {
		redirectURI = "http://localhost:8710/api/v1/github/callback"
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
	userID := auth.UserIDFromContext(ctx)

	if connType == "" {
		connType = "user"
	}

	if code == "" {
		http.Error(w, "No code provided", http.StatusBadRequest)
		return
	}

	token, err := h.githubService.ExchangeCodeForToken(ctx, code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to exchange token: %v", err), http.StatusInternalServerError)
		return
	}

	login, avatarURL, err := h.githubService.GetUserInfo(ctx, token)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user info: %v", err), http.StatusInternalServerError)
		return
	}

	err = h.githubService.SaveConnection(ctx, userID, connType, login, avatarURL, token, "repo,read:user,read:org")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to save connection: %v", err), http.StatusInternalServerError)
		return
	}

	// Sync repos in background (not blocking the redirect)
	go func() {
		bgCtx := context.Background()

		// Sync repos to cache
		if err := h.repoCache.SyncReposFromGitHub(bgCtx, userID, token); err != nil {
			slog.Error("[OAuth] Failed to sync repos to cache", "error", err)
		} else {
			slog.Info("[OAuth] Repos synced to cache successfully")
		}

		// Also save repos to projects table
		conn := &models.GitHubConnection{
			Type:  connType,
			Login: login,
			Token: token,
		}
		if err := h.githubService.FetchAndSaveRepositories(bgCtx, userID, conn); err != nil {
			slog.Error("[OAuth] Failed to save projects to DB", "error", err)
		} else {
			slog.Info("[OAuth] Projects saved to DB successfully")
		}
	}()

	redirectURL := getRedirectURL(r, "/dashboard")
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

func (h *Handlers) HandleDisconnect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := auth.UserIDFromContext(ctx)
	slog.Info("[DISCONNECT] HandleDisconnect called", "method", r.Method, "url", r.URL.String())

	// Clear auth session cookies
	if h.authService != nil {
		h.authService.ClearSessionCookies(w)
		slog.Info("[DISCONNECT] Cleared auth session")
	}

	if err := h.githubService.DeleteConnection(ctx, userID); err != nil {
		slog.Error("[DISCONNECT] Failed to delete connection", "error", err)
	} else {
		slog.Info("[DISCONNECT] Successfully deleted connection")
	}

	if err := h.githubService.DeleteAllProjects(ctx, userID); err != nil {
		slog.Error("[DISCONNECT] Failed to delete projects", "error", err)
	} else {
		slog.Info("[DISCONNECT] Successfully deleted projects")
	}

	slog.Info("[DISCONNECT] Redirecting to landing page")
	redirectURL := getRedirectURL(r, "/")
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}
