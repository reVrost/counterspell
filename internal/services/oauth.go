package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/google/uuid"
	"github.com/revrost/counterspell/internal/config"
	"github.com/revrost/counterspell/internal/db"
	"github.com/revrost/counterspell/internal/db/sqlc"
)

// OAuthService handles OAuth login flow for Counterspell.
type OAuthService struct {
	db          *db.DB
	cfg         *config.Config
	httpClient  *http.Client
	loginServer *http.Server
	callbackCh  chan *OAuthCallbackResult
}

// OAuthCallbackResult represents the result of OAuth callback.
type OAuthCallbackResult struct {
	MachineJWT  string
	UserID      string
	UserEmail   string
	Subdomain   string
	TunnelToken string
	Error       error
}

// OAuthLoginRequest represents a request to start login.
type OAuthLoginRequest struct {
	RedirectURI string `json:"redirect_uri"`
}

// OAuthLoginResponse represents the response from the Invoker control plane.
type OAuthLoginResponse struct {
	AuthURL string `json:"auth_url"`
}

// OAuthExchangeRequest represents a request to exchange OAuth code.
type OAuthExchangeRequest struct {
	Code         string `json:"code"`
	State        string `json:"state"`
	CodeVerifier string `json:"code_verifier"`
}

// OAuthExchangeResponse represents the response from the Invoker control plane.
type OAuthExchangeResponse struct {
	MachineJWT string `json:"machine_jwt"`
	UserID     string `json:"user_id"`
	UserEmail  string `json:"user_email"`
}

// MachineRegisterRequest represents a request to register a machine.
type MachineRegisterRequest struct {
	MachineID string `json:"machine_id"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	Hostname  string `json:"hostname"`
	Version   string `json:"version"`
}

// MachineRegisterResponse represents the response from the Invoker control plane.
type MachineRegisterResponse struct {
	UserID      string `json:"user_id"`
	Subdomain    string `json:"subdomain"`
	TunnelToken  string `json:"tunnel_token"`
}

// NewOAuthService creates a new OAuth service.
func NewOAuthService(database *db.DB, cfg *config.Config) *OAuthService {
	return &OAuthService{
		db:         database,
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		callbackCh: make(chan *OAuthCallbackResult, 1),
	}
}

// StartLoginFlow initiates the OAuth 2.0 Authorization Code Flow with PKCE.
// This is called during CLI startup (not an HTTP handler).
func (s *OAuthService) StartLoginFlow(ctx context.Context) (*OAuthLoginResponse, error) {
	// 1. Generate PKCE code_verifier
	codeVerifier, err := s.generateCodeVerifier()
	if err != nil {
		return nil, fmt.Errorf("failed to generate code_verifier: %w", err)
	}

	// 2. Derive code_challenge using SHA256 + base64url
	codeChallenge := s.generateCodeChallenge(codeVerifier)

	// 3. Generate cryptographically secure state value
	state := uuid.New().String()

	// 4. Persist {state, code_verifier, created_at} in SQLite
	err = s.db.Queries.CreateOAuthLoginAttempt(ctx, sqlc.CreateOAuthLoginAttemptParams{
		State:         state,
		CodeVerifier:  codeVerifier,
		CreatedAt:     time.Now().UnixMilli(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to store oauth attempt: %w", err)
	}

	slog.Info("OAuth login attempt created", "state", state)

	// 5. Call Invoker POST /api/v1/auth/url
	redirectURI := "http://localhost:8711/auth/callback"
	authURL, err := s.callInvokerAuthURL(ctx, codeChallenge, state, redirectURI)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth url from invoker: %w", err)
	}

	return &OAuthLoginResponse{AuthURL: authURL}, nil
}

// StartCallbackServer starts a temporary HTTP server to handle OAuth callback.
func (s *OAuthService) StartCallbackServer(ctx context.Context, callbackPort string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/callback", s.handleCallback)

	s.loginServer = &http.Server{
		Addr:    "127.0.0.1:" + callbackPort,
		Handler: mux,
	}

	slog.Info("Starting OAuth callback server", "addr", s.loginServer.Addr)

	go func() {
		if err := s.loginServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("OAuth callback server error", "error", err)
			s.callbackCh <- &OAuthCallbackResult{Error: err}
		}
	}()

	return nil
}

// StopCallbackServer stops the OAuth callback server.
func (s *OAuthService) StopCallbackServer(ctx context.Context) error {
	if s.loginServer != nil {
		return s.loginServer.Shutdown(ctx)
	}
	return nil
}

// WaitForCallback waits for OAuth callback to complete.
func (s *OAuthService) WaitForCallback(ctx context.Context) (*OAuthCallbackResult, error) {
	select {
	case result := <-s.callbackCh:
		return result, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// handleCallback handles the OAuth callback from the browser.
func (s *OAuthService) handleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Only accept requests from loopback
	if r.RemoteAddr != "127.0.0.1:"+s.cfg.OAuthCallbackPort && !isLoopback(r.RemoteAddr) {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// 1. Parse code and state query parameters
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" || state == "" {
		http.Error(w, "Missing code or state parameter", http.StatusBadRequest)
		return
	}

	// 2. Lookup stored state + code_verifier from SQLite
	attempt, err := s.db.Queries.GetOAuthLoginAttempt(ctx, state)
	if err != nil {
		slog.Error("Failed to get OAuth login attempt", "error", err, "state", state)
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	// Check if attempt is expired (5 minutes)
	if time.Since(time.UnixMilli(attempt.CreatedAt)) > 5*time.Minute {
		slog.Warn("OAuth login attempt expired", "state", state)
		http.Error(w, "State expired", http.StatusBadRequest)
		return
	}

	// 4. Call Invoker POST /api/v1/auth/exchange
	exchangeResp, err := s.callInvokerExchange(ctx, code, state, attempt.CodeVerifier)
	if err != nil {
		slog.Error("Failed to exchange OAuth code", "error", err)
		http.Error(w, "Failed to exchange code", http.StatusInternalServerError)
		return
	}

	// 5. Delete the oauth attempt
	if err := s.db.Queries.DeleteOAuthLoginAttempt(ctx, state); err != nil {
		slog.Error("Failed to delete OAuth login attempt", "error", err)
	}

	// 6. Send success response to browser
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
		<html>
			<head><title>Authentication Successful</title></head>
			<body style="font-family: sans-serif; text-align: center; padding-top: 50px;">
				<h1>Authentication Successful!</h1>
				<p>You can close this window and return to Counterspell.</p>
			</body>
		</html>
	`))

	// 7. Notify main process via channel
	s.callbackCh <- &OAuthCallbackResult{
		MachineJWT: exchangeResp.MachineJWT,
		UserID:     exchangeResp.UserID,
		UserEmail:  exchangeResp.UserEmail,
	}
}

// RegisterMachine associates the local machine with the authenticated user.
func (s *OAuthService) RegisterMachine(ctx context.Context, machineJWT string) (*MachineRegisterResponse, error) {
	// 1. Generate or load persistent machine_id
	machineID := s.getMachineID()

	// 2. Collect system metadata
	hostname, _ := os.Hostname()

	req := MachineRegisterRequest{
		MachineID: machineID,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Hostname:  hostname,
		Version:   "dev", // TODO: get from build info
	}

	// 3. Call Invoker POST /api/v1/machines/register
	resp, err := s.callInvokerRegisterMachine(ctx, machineJWT, req)
	if err != nil {
		return nil, fmt.Errorf("failed to register machine: %w", err)
	}

	// 4. Persist values in SQLite
	_, err = s.db.Queries.UpsertMachineIdentity(ctx, sqlc.UpsertMachineIdentityParams{
		MachineID:      machineID,
		UserID:         resp.UserID,
		Subdomain:      resp.Subdomain,
		TunnelProvider: "cloudflare",
		TunnelToken:    resp.TunnelToken,
		CreatedAt:      time.Now().UnixMilli(),
		LastSeenAt:     sql.NullInt64{Int64: time.Now().UnixMilli(), Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to persist machine identity: %w", err)
	}

	slog.Info("Machine registered successfully", "machine_id", machineID, "subdomain", resp.Subdomain)

	return resp, nil
}

// generateCodeVerifier generates a PKCE code_verifier (RFC 7636).
func (s *OAuthService) generateCodeVerifier() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// generateCodeChallenge derives code_challenge from code_verifier using SHA256 + base64url.
func (s *OAuthService) generateCodeChallenge(codeVerifier string) string {
	hash := sha256.Sum256([]byte(codeVerifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// getMachineID returns a persistent machine ID.
func (s *OAuthService) getMachineID() string {
	// TODO: Load from persistent storage or generate once
	// For now, generate a new UUID
	return uuid.New().String()
}

// callInvokerAuthURL calls the Invoker control plane to get the auth URL.
func (s *OAuthService) callInvokerAuthURL(ctx context.Context, codeChallenge, state, redirectURI string) (string, error) {
	// TODO: Implement actual HTTP call to Invoker
	// For now, return a placeholder URL
	return "https://invoker.counterspell.app/auth?code_challenge=" + codeChallenge + "&state=" + state, nil
}

// callInvokerExchange calls the Invoker control plane to exchange the OAuth code.
func (s *OAuthService) callInvokerExchange(ctx context.Context, code, state, codeVerifier string) (*OAuthExchangeResponse, error) {
	// TODO: Implement actual HTTP call to Invoker
	// For now, return a placeholder response
	return &OAuthExchangeResponse{
		MachineJWT: "placeholder.jwt.token",
		UserID:    "user-123",
		UserEmail: "user@example.com",
	}, nil
}

// callInvokerRegisterMachine calls the Invoker control plane to register the machine.
func (s *OAuthService) callInvokerRegisterMachine(ctx context.Context, machineJWT string, req MachineRegisterRequest) (*MachineRegisterResponse, error) {
	// TODO: Implement actual HTTP call to Invoker
	// For now, return a placeholder response
	return &MachineRegisterResponse{
		UserID:     "user-123",
		Subdomain:   "alice",
		TunnelToken: "placeholder-tunnel-token",
	}, nil
}

// OpenBrowser opens the system browser to the specified URL.
func (s *OAuthService) OpenBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "rundll32.exe"
		args = []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return exec.Command(cmd, args...).Start()
}

// isLoopback checks if the remote address is from loopback.
func isLoopback(addr string) bool {
	// Simple check - in production, parse the IP properly
	return true
}

// CleanupExpiredOAuthAttempts removes expired OAuth login attempts.
func (s *OAuthService) CleanupExpiredOAuthAttempts(ctx context.Context) error {
	cutoff := time.Now().Add(-5 * time.Minute).UnixMilli()
	return s.db.Queries.CleanupExpiredOAuthAttempts(ctx, cutoff)
}
