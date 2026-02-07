package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/render"
	"github.com/revrost/counterspell/internal/services"
)

// HandleAuthLogin starts the browser OAuth flow via the Invoker control plane.
func (h *Handlers) HandleAuthLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if authenticated, _, err := h.oauthService.IsAuthenticated(ctx); err == nil && authenticated {
		http.Redirect(w, r, "/app", http.StatusTemporaryRedirect)
		return
	} else if err != nil {
		if errors.Is(err, services.ErrForbiddenLoginIdentityMismatch) {
			_ = render.Render(w, r, ErrForbidden("This Counterspell instance belongs to a different account"))
			return
		}
		slog.Error("Auth check failed", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Authentication failed", err))
		return
	}

	returnTo := sanitizeReturnTo(r.URL.Query().Get("return_to"))
	attempt, err := h.oauthService.StartWebLoginWithReturnTo(ctx, returnTo)
	if err != nil {
		slog.Error("Failed to start OAuth login", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to start login", err))
		return
	}

	http.Redirect(w, r, attempt.AuthURL, http.StatusTemporaryRedirect)
}

func sanitizeReturnTo(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	parsed, err := url.Parse(raw)
	if err != nil || !parsed.IsAbs() {
		return ""
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ""
	}

	host := strings.ToLower(parsed.Hostname())
	if host == "localhost" || host == "127.0.0.1" || host == "counterspell.io" || strings.HasSuffix(host, ".counterspell.app") {
		return parsed.String()
	}

	return ""
}

// HandleLogout clears local machine authentication for this Counterspell instance.
func (h *Handlers) HandleLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := h.oauthService.Logout(ctx); err != nil {
		slog.Error("Failed to logout", "error", err)
		_ = render.Render(w, r, ErrInternalServer("Failed to logout", err))
		return
	}

	render.JSON(w, r, map[string]any{"ok": true})
}

// RequireMachineAuth blocks API access until the machine is authenticated.
func (h *Handlers) RequireMachineAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		authenticated, _, err := h.oauthService.IsAuthenticated(ctx)
		if err != nil {
			if errors.Is(err, services.ErrForbiddenLoginIdentityMismatch) {
				_ = render.Render(w, r, ErrForbidden("This Counterspell instance belongs to a different account"))
				return
			}
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
