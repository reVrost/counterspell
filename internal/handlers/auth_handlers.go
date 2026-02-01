package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/render"
)

// HandleAuthLogin starts the browser OAuth flow via the Invoker control plane.
func (h *Handlers) HandleAuthLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if authenticated, _, err := h.oauthService.IsAuthenticated(ctx); err == nil && authenticated {
		http.Redirect(w, r, "/dashboard", http.StatusTemporaryRedirect)
		return
	}

	attempt, err := h.oauthService.StartWebLogin(ctx)
	if err != nil {
		slog.Error("Failed to start OAuth login", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to start login", err))
		return
	}

	http.Redirect(w, r, attempt.AuthURL, http.StatusTemporaryRedirect)
}

// RequireMachineAuth blocks API access until the machine is authenticated.
func (h *Handlers) RequireMachineAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		authenticated, _, err := h.oauthService.IsAuthenticated(ctx)
		if err != nil {
			slog.Error("Auth check failed", "error", err)
			_ = render.Render(w, r, ErrInternalServer("Authentication failed", err))
			return
		}
		if !authenticated {
			_ = render.Render(w, r, ErrUnauthorized("Authentication required"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
