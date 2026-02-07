package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func (h *Handlers) HandleConnectorConnect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	connector := strings.TrimSpace(chi.URLParam(r, "connector"))
	if connector != "openai" {
		http.Error(w, "unsupported connector", http.StatusNotFound)
		return
	}

	var req struct {
		RedirectURI string `json:"redirect_uri"`
	}
	if err := render.DecodeJSON(r.Body, &req); err != nil && err != io.EOF {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	redirectURI := strings.TrimSpace(req.RedirectURI)
	if redirectURI == "" {
		redirectURI = connectorCallbackURL(r, connector)
	}

	authURL, err := h.openAIConnector.StartConnect(ctx, redirectURI)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to start openai connector flow: %v", err), http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, map[string]any{
		"auth_url": authURL,
	})
}

func (h *Handlers) HandleConnectorStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	connector := strings.TrimSpace(chi.URLParam(r, "connector"))
	if connector != "openai" {
		http.Error(w, "unsupported connector", http.StatusNotFound)
		return
	}
	status, err := h.openAIConnector.Status(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to load connector status: %v", err), http.StatusInternalServerError)
		return
	}
	render.JSON(w, r, status)
}

func (h *Handlers) HandleConnectorDisconnect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	connector := strings.TrimSpace(chi.URLParam(r, "connector"))
	if connector != "openai" {
		http.Error(w, "unsupported connector", http.StatusNotFound)
		return
	}
	if err := h.openAIConnector.Disconnect(ctx); err != nil {
		http.Error(w, fmt.Sprintf("failed to disconnect openai connector: %v", err), http.StatusInternalServerError)
		return
	}
	render.JSON(w, r, map[string]string{"status": "ok"})
}

func (h *Handlers) HandleConnectorCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	connector := strings.TrimSpace(chi.URLParam(r, "connector"))
	if connector != "openai" {
		http.Error(w, "unsupported connector", http.StatusNotFound)
		return
	}

	if oauthErr := strings.TrimSpace(r.URL.Query().Get("error")); oauthErr != "" {
		renderConnectorCallbackPage(w, http.StatusBadRequest, false, connector, "Authorization was denied.")
		return
	}

	code := strings.TrimSpace(r.URL.Query().Get("code"))
	state := strings.TrimSpace(r.URL.Query().Get("state"))
	if code == "" || state == "" {
		renderConnectorCallbackPage(w, http.StatusBadRequest, false, connector, "Missing code or state.")
		return
	}

	redirectURI := connectorCallbackURL(r, connector)
	if err := h.openAIConnector.CompleteConnect(ctx, code, state, redirectURI); err != nil {
		renderConnectorCallbackPage(w, http.StatusInternalServerError, false, connector, "Failed to complete connector connection.")
		return
	}

	renderConnectorCallbackPage(w, http.StatusOK, true, connector, "Connector is connected. You can close this window.")
}

func connectorCallbackURL(r *http.Request, connector string) string {
	scheme := "http"
	if strings.EqualFold(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")), "https") || r.TLS != nil {
		scheme = "https"
	}
	host := strings.TrimSpace(r.Header.Get("X-Forwarded-Host"))
	if host == "" {
		host = r.Host
	}
	return fmt.Sprintf("%s://%s/api/v1/connectors/%s/callback", scheme, host, connector)
}

func renderConnectorCallbackPage(w http.ResponseWriter, statusCode int, success bool, connector, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	statusLabel := "Connection Failed"
	color := "#ef4444"
	if success {
		statusLabel = "Connection Complete"
		color = "#10b981"
	}
	_, _ = w.Write([]byte(fmt.Sprintf(`<!doctype html>
<html>
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>%s</title>
    <style>
      body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: #0b0b10; color: #e5e7eb; display:flex; align-items:center; justify-content:center; min-height:100vh; margin:0; }
      .card { max-width: 460px; width: calc(100%% - 40px); border: 1px solid #1f2937; border-radius: 14px; background: #111827; padding: 24px; text-align:center; }
      .title { font-size: 20px; margin: 0 0 8px; color: %s; }
      .body { color: #9ca3af; line-height: 1.4; margin-bottom: 16px; }
      .link { color: #93c5fd; text-decoration: none; }
    </style>
  </head>
  <body>
    <div class="card">
      <h1 class="title">%s</h1>
      <p class="body">%s</p>
      <a class="link" href="/app/settings">Return to Settings</a>
    </div>
    <script>
      try {
        if (window.opener && !window.opener.closed) {
          window.opener.postMessage({ source: 'counterspell-connector', connector: '%s', connected: %t }, window.location.origin);
          window.close();
        }
      } catch (_) {}
    </script>
  </body>
</html>`, statusLabel, color, statusLabel, message, connector, success)))
}
