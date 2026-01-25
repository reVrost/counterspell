package cli

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/revrost/code/counterspell/internal/services"
)

// AuthCallbackServer handles OAuth callback from browser.
type AuthCallbackServer struct {
	server     *http.Server
	resultChan chan *AuthResult
	authURL    string
	state      string
}

type AuthResult struct {
	Code  string `json:"code"`
	State string `json:"state"`
	Error string `json:"error,omitempty"`
}

// NewAuthCallbackServer creates a new callback server.
func NewAuthCallbackServer(port int) *AuthCallbackServer {
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:   10 * time.Second,
		WriteTimeout:  10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	acs := &AuthCallbackServer{
		server:     server,
		resultChan: make(chan *AuthResult, 1),
	}

	mux.HandleFunc("/callback", acs.handleCallback)
	mux.HandleFunc("/ping", acs.handlePing)

	return acs
}

// Start starts the callback server.
func (a *AuthCallbackServer) Start() error {
	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Callback server error", "error", err)
		}
	}()
	return nil
}

// Stop stops the callback server.
func (a *AuthCallbackServer) Stop(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}

// WaitForResult waits for the auth callback result with timeout.
func (a *AuthCallbackServer) WaitForResult(timeout time.Duration) (*AuthResult, error) {
	select {
	case result := <-a.resultChan:
		return result, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for callback")
	}
}

// GetCallbackURL returns the URL that the browser should redirect to.
func (a *AuthCallbackServer) GetCallbackURL() string {
	return fmt.Sprintf("http://localhost%s/callback", a.server.Addr)
}

func (a *AuthCallbackServer) handlePing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (a *AuthCallbackServer) handleCallback(w http.ResponseWriter, r *http.Request) {
	// Close the resultChan when done
	defer close(a.resultChan)

	var result AuthResult

	// Check for error in query params
	if err := r.URL.Query().Get("error"); err != "" {
		result.Error = err
		a.sendResult(w, &result)
		return
	}

	// Get authorization code from query params
	code := r.URL.Query().Get("code")
	if code == "" {
		result.Error = "no authorization code provided"
		a.sendResult(w, &result)
		return
	}

	// Get state
	state := r.URL.Query().Get("state")

	// Store code and state for back-end exchange
	result.Code = code
	result.State = state

	a.sendResult(w, &result)
}

func (a *AuthCallbackServer) sendResult(w http.ResponseWriter, result *AuthResult) {
	// Send result to channel (non-blocking)
	select {
	case a.resultChan <- result:
	default:
	}

	// Send HTML response to browser
	html := `<!DOCTYPE html>
<html>
<head>
	<title>Authentication Successful</title>
	<style>
		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
			display: flex;
			align-items: center;
			justify-content: center;
			height: 100vh;
			margin: 0;
			background: #f5f5f5;
		}
		.container {
			text-align: center;
			padding: 40px;
			background: white;
			border-radius: 8px;
			box-shadow: 0 2px 10px rgba(0,0,0,0.1);
			max-width: 400px;
		}
		.success {
			color: #10b981;
			font-size: 48px;
			margin-bottom: 16px;
		}
		.error {
			color: #ef4444;
			font-size: 48px;
			margin-bottom: 16px;
		}
		h1 {
			margin: 0 0 8px 0;
			font-size: 24px;
		}
		p {
			color: #666;
			margin: 0 0 24px 0;
		}
		.close-hint {
			font-size: 14px;
			color: #999;
		}
	</style>
</head>
<body>`

	if result.Error != "" {
		html += `
	<div class="container">
		<div class="error">✕</div>
		<h1>Authentication Failed</h1>
		<p>` + result.Error + `</p>
		<p class="close-hint">You can close this window and try again.</p>
	</div>
</body>
</html>`
	} else {
		html += `
	<div class="container">
		<div class="success">✓</div>
		<h1>Authentication Successful</h1>
		<p>Exchanging authorization code for token...</p>
	</div>
</body>
</html>`
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, html)

	slog.Info("Auth callback received", "has_code", result.Code != "", "error", result.Error)
}

// AuthWithCallback performs auth using local callback server.
type AuthWithCallback struct {
	authService  *services.AuthService
	callbackPort int
	controlPlane *services.ControlPlaneClient
}

// NewAuthWithCallback creates auth with callback.
func NewAuthWithCallback(authService *services.AuthService, controlPlane *services.ControlPlaneClient, callbackPort int) *AuthWithCallback {
	return &AuthWithCallback{
		authService:  authService,
		callbackPort:  callbackPort,
		controlPlane:  controlPlane,
	}
}

// Authenticate performs auth flow with browser callback.
func (a *AuthWithCallback) Authenticate(ctx context.Context, machineID string) (string, *services.ExchangeCodeResponse, error) {
	// Start callback server
	callbackServer := NewAuthCallbackServer(a.callbackPort)
	if err := callbackServer.Start(); err != nil {
		return "", nil, fmt.Errorf("failed to start callback server: %w", err)
	}
	defer callbackServer.Stop(ctx)

	callbackURL := callbackServer.GetCallbackURL()

	fmt.Println("\n" + "============================================================")
	fmt.Println("Welcome to Counterspell!")
	fmt.Println("============================================================")
	fmt.Println("\nYou need to authenticate to use Counterspell.")
	fmt.Println("This allows us to:")
	fmt.Println("  - Provide your personal subdomain (e.g., username.counterspell.app)")
	fmt.Println("  - Create a secure tunnel to your machine")
	fmt.Println("  - Manage your cloud deployments")
	fmt.Println()

	// Get auth URL with our callback URL (with PKCE)
	authResp, err := a.controlPlane.GetAuthURLWithPKCE(
		ctx,
		machineID,
		callbackURL,
		a.authService.GetCodeChallenge(),
		a.authService.GetState(),
	)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get auth URL: %w", err)
	}

	fmt.Printf("\n1. Opening your browser...\n")
	fmt.Printf("\n   If it doesn't open automatically, visit:\n\n")
	fmt.Printf("   %s\n\n", authResp.AuthURL)
	fmt.Printf("2. Log in or create an account\n")
	fmt.Printf("3. After logging in, we'll automatically exchange code for a JWT\n\n")

	// Open browser (optional - let user do it)
	fmt.Printf("Waiting for authorization... (press Ctrl+C to cancel)\n")

	// Wait for callback with timeout
	result, err := callbackServer.WaitForResult(5 * time.Minute)
	if err != nil {
		return "", nil, fmt.Errorf("timeout waiting for authentication: %w", err)
	}

	if result.Error != "" {
		return "", nil, fmt.Errorf("authentication failed: %s", result.Error)
	}

	fmt.Println("\n✓ Authorization code received!")
	fmt.Println("Exchanging code for JWT...")

	// Perform back-end code exchange (JWT never goes through URL!)
	exchangeResp, err := a.authService.CompleteAuthFlow(ctx, result.Code, result.State)
	if err != nil {
		return "", nil, fmt.Errorf("failed to exchange code for JWT: %w", err)
	}

	fmt.Println("\n✓ Successfully authenticated!")

	return exchangeResp.UserID, exchangeResp, nil
}
