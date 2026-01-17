package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/revrost/code/counterspell/internal/db"
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

	// Sync repos to cache in background (not blocking the redirect)
	go func() {
		userDB := db.DBFromContext(ctx)
		if userDB == nil {
			slog.Error("[OAuth] No database in context for repo sync")
			return
		}
		cache := services.NewRepoCache(userDB)
		if err := cache.SyncReposFromGitHub(context.Background(), token); err != nil {
			slog.Error("[OAuth] Failed to sync repos to cache", "error", err)
		} else {
			slog.Info("[OAuth] Repos synced to cache successfully")
		}
	}()

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
