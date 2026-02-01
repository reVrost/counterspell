package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// OAuthHandlers handles OAuth-related HTTP requests.
type OAuthHandlers struct {
	handlers *Handlers
}

// RegisterOAuthRoutes registers OAuth-related routes.
func (h *Handlers) RegisterOAuthRoutes(r chi.Router) {
	oauth := &OAuthHandlers{handlers: h}

	r.Get("/auth/callback", oauth.OAuthCallback)
}

// OAuthCallback handles the OAuth redirect from the browser.
// This is NOT used when the OAuth service starts its own temporary server on localhost:8711.
// It's provided for completeness if the main Counterspell server needs to handle callbacks.
func (h *OAuthHandlers) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	slog.Info("OAuth callback received", "path", r.URL.Path)
	slog.Warn("OAuth callback should be handled by temporary server on localhost:8711")

	// Return error page
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte(`
		<html>
			<head><title>OAuth Error</title></head>
			<body style="font-family: sans-serif; text-align: center; padding-top: 50px;">
				<h1>OAuth Callback Error</h1>
				<p>OAuth callback should be handled by the temporary server on localhost:8711.</p>
				<p>Please restart the Counterspell login flow.</p>
			</body>
		</html>
	`))
}
