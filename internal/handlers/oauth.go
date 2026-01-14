package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
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
		h.github.FetchAndSaveRepositories(ctx, conn)
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
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
	
	slog.Info("[DISCONNECT] Redirecting to /")
	http.Redirect(w, r, "/", http.StatusSeeOther) // 303 converts POST to GET
}

// Legacy/Stub Handlers
func (h *Handlers) HandleAuth(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/github/authorize", http.StatusTemporaryRedirect)
}

func (h *Handlers) HandleRegister(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/github/authorize", http.StatusTemporaryRedirect)
}

func (h *Handlers) HandleOAuth(w http.ResponseWriter, r *http.Request) {
	// Provider usage or redirect
	h.HandleGitHubAuthorize(w, r)
}

func (h *Handlers) HandleAuthCallback(w http.ResponseWriter, r *http.Request) {
	h.HandleGitHubCallback(w, r)
}

func (h *Handlers) HandleAuthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"authenticated": true}`))
}

func (h *Handlers) HandleLogout(w http.ResponseWriter, r *http.Request) {
	h.HandleDisconnect(w, r)
}
