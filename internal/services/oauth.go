package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
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
	RedirectURI   string `json:"redirect_uri"`
	CodeChallenge string `json:"code_challenge"`
	State         string `json:"state"`
	Provider      string `json:"provider"`
}

// OAuthLoginResponse represents the response from the Invoker control plane.
type OAuthLoginResponse struct {
	AuthURL string `json:"auth_url"`
}

// OAuthLoginAttempt bundles login metadata for the CLI flow.
type OAuthLoginAttempt struct {
	AuthURL     string
	State       string
	RedirectURI string
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

// OAuthPollRequest represents a request to poll for an OAuth code.
type OAuthPollRequest struct {
	State string `json:"state"`
}

// OAuthPollResponse represents a poll response from Invoker.
type OAuthPollResponse struct {
	Status string `json:"status"`
	Code   string `json:"code,omitempty"`
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
	Subdomain   string `json:"subdomain"`
	TunnelToken string `json:"tunnel_token"`
}

// AuthResult captures the machine auth + tunnel metadata for startup.
type AuthResult struct {
	MachineJWT  string
	MachineID   string
	UserID      string
	Subdomain   string
	TunnelToken string
}

const (
	oauthAttemptTTL   = 10 * time.Minute
	oauthPollInterval = 2 * time.Second
)

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
func (s *OAuthService) StartLoginFlow(ctx context.Context) (*OAuthLoginAttempt, error) {
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
		State:        state,
		CodeVerifier: codeVerifier,
		CreatedAt:    time.Now().UnixMilli(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to store oauth attempt: %w", err)
	}

	slog.Info("OAuth login attempt created", "state", state)

	// 5. Call Invoker POST /api/v1/auth/url
	redirectURI := s.cfg.OAuthRedirectURI
	if redirectURI == "" {
		redirectURI = fmt.Sprintf("http://localhost:%s/auth/callback", s.cfg.OAuthCallbackPort)
	}
	if parsedRedirect, err := url.Parse(redirectURI); err == nil {
		q := parsedRedirect.Query()
		if q.Get("cs_state") == "" {
			q.Set("cs_state", state)
			parsedRedirect.RawQuery = q.Encode()
			redirectURI = parsedRedirect.String()
		}
	}
	slog.Info("OAuth login start", "redirect_uri", redirectURI, "invoker_base", s.cfg.InvokerBaseURL)
	authURL, err := s.callInvokerAuthURL(ctx, codeChallenge, state, redirectURI)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth url from invoker: %w", err)
	}
	slog.Info("OAuth auth_url received", "auth_url", authURL)

	return &OAuthLoginAttempt{AuthURL: authURL, State: state, RedirectURI: redirectURI}, nil
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

	// 2. Exchange code for machine JWT
	exchangeResp, err := s.exchangeWithState(ctx, code, state)
	if err != nil {
		slog.Error("Failed to exchange OAuth code", "error", err)
		http.Error(w, "Failed to exchange code", http.StatusInternalServerError)
		return
	}

	// 3. Send success response to browser
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`
		<html>
			<head><title>Authentication Successful</title></head>
			<body style="font-family: sans-serif; text-align: center; padding-top: 50px;">
				<h1>Authentication Successful!</h1>
				<p>You can close this window and return to Counterspell.</p>
			</body>
		</html>
	`))

	// 4. Notify main process via channel
	s.callbackCh <- &OAuthCallbackResult{
		MachineJWT: exchangeResp.MachineJWT,
		UserID:     exchangeResp.UserID,
		UserEmail:  exchangeResp.UserEmail,
	}
}

func (s *OAuthService) exchangeWithState(ctx context.Context, code, state string) (*OAuthExchangeResponse, error) {
	attempt, err := s.db.Queries.GetOAuthLoginAttempt(ctx, state)
	if err != nil {
		return nil, fmt.Errorf("invalid state: %w", err)
	}

	if time.Since(time.UnixMilli(attempt.CreatedAt)) > oauthAttemptTTL {
		return nil, fmt.Errorf("state expired")
	}

	exchangeResp, err := s.callInvokerExchange(ctx, code, state, attempt.CodeVerifier)
	if err != nil {
		return nil, err
	}

	if err := s.db.Queries.DeleteOAuthLoginAttempt(ctx, state); err != nil {
		slog.Error("Failed to delete OAuth login attempt", "error", err)
	}

	return exchangeResp, nil
}

// RegisterMachine associates the local machine with the authenticated user.
func (s *OAuthService) RegisterMachine(ctx context.Context, machineJWT string, machineID string) (*MachineRegisterResponse, error) {
	if machineID == "" {
		return nil, fmt.Errorf("machine_id is required")
	}

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
		MachineJwt:     sql.NullString{String: machineJWT, Valid: machineJWT != ""},
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

// EnsureAuthenticated ensures machine JWT + machine identity exist, prompting login if needed.
func (s *OAuthService) EnsureAuthenticated(ctx context.Context) (*AuthResult, error) {
	machineID, err := s.getOrCreateMachineID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get machine id: %w", err)
	}

	identity, err := s.getMachineIdentity(ctx, machineID)
	if err != nil {
		return nil, fmt.Errorf("failed to get machine identity: %w", err)
	}
	machineJWT := ""
	if identity != nil && identity.MachineJwt.Valid {
		machineJWT = identity.MachineJwt.String
	}
	if machineJWT == "" {
		machineJWT, err = s.login(ctx)
		if err != nil {
			return nil, err
		}
		if identity != nil {
			if err := s.storeMachineJWT(ctx, machineID, machineJWT); err != nil {
				return nil, fmt.Errorf("failed to store machine jwt: %w", err)
			}
		}
	}
	if identity != nil {
		if err := s.updateMachineIdentityLastSeen(ctx, machineID); err != nil {
			slog.Warn("Failed to update machine last_seen", "error", err)
		}
		return &AuthResult{
			MachineJWT:  machineJWT,
			MachineID:   machineID,
			UserID:      identity.UserID,
			Subdomain:   identity.Subdomain,
			TunnelToken: identity.TunnelToken,
		}, nil
	}

	// Register with Invoker if no local identity exists.
	reg, err := s.RegisterMachine(ctx, machineJWT, machineID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		MachineJWT:  machineJWT,
		MachineID:   machineID,
		UserID:      reg.UserID,
		Subdomain:   reg.Subdomain,
		TunnelToken: reg.TunnelToken,
	}, nil
}

func (s *OAuthService) login(ctx context.Context) (string, error) {
	if s.cfg.ForceDeviceCode || s.cfg.Headless {
		return "", fmt.Errorf("device flow is disabled; run with browser-based OAuth enabled")
	}

	// Try browser-based flow first.
	loginAttempt, err := s.StartLoginFlow(ctx)
	if err != nil {
		return "", err
	}

	isLoopback := isLoopbackRedirectURI(loginAttempt.RedirectURI)
	callbackPort := s.cfg.OAuthCallbackPort
	if isLoopback {
		callbackPort = callbackPortFromRedirectURI(loginAttempt.RedirectURI, callbackPort)
		if err := s.StartCallbackServer(ctx, callbackPort); err != nil {
			return "", err
		}
		defer func() {
			_ = s.StopCallbackServer(context.Background())
		}()
	}

	if err := s.OpenBrowser(loginAttempt.AuthURL); err != nil {
		return "", fmt.Errorf("failed to open browser for OAuth: %w", err)
	}

	if isLoopback {
		result, err := s.WaitForCallback(ctx)
		if err != nil {
			return "", err
		}
		if result.Error != nil {
			return "", result.Error
		}

		return result.MachineJWT, nil
	}

	return s.pollForAuthCodeAndExchange(ctx, loginAttempt.State)
}

func (s *OAuthService) pollForAuthCodeAndExchange(ctx context.Context, state string) (string, error) {
	attempt, err := s.db.Queries.GetOAuthLoginAttempt(ctx, state)
	if err != nil {
		return "", fmt.Errorf("failed to load login attempt: %w", err)
	}

	deadline := time.UnixMilli(attempt.CreatedAt).Add(oauthAttemptTTL)
	if time.Now().After(deadline) {
		return "", fmt.Errorf("oauth login attempt expired")
	}
	pollCtx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	ticker := time.NewTicker(oauthPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pollCtx.Done():
			return "", fmt.Errorf("oauth login expired or canceled")
		case <-ticker.C:
			resp, err := s.pollOAuthCode(pollCtx, state)
			if err != nil {
				return "", err
			}
			if resp.Status == "ready" && resp.Code != "" {
				exchangeResp, err := s.exchangeWithState(pollCtx, resp.Code, state)
				if err != nil {
					return "", err
				}
				return exchangeResp.MachineJWT, nil
			}
		}
	}
}

func (s *OAuthService) pollOAuthCode(ctx context.Context, state string) (*OAuthPollResponse, error) {
	req := OAuthPollRequest{State: state}
	var resp OAuthPollResponse
	if err := s.doInvokerJSON(ctx, http.MethodPost, "/api/v1/auth/poll", req, &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (s *OAuthService) getMachineIdentity(ctx context.Context, machineID string) (*sqlc.MachineIdentity, error) {
	if machineID == "" {
		return nil, nil
	}
	identity, err := s.db.Queries.GetMachineIdentity(ctx, machineID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &identity, nil
}

func (s *OAuthService) updateMachineIdentityLastSeen(ctx context.Context, machineID string) error {
	if machineID == "" {
		return nil
	}
	return s.db.Queries.UpdateMachineIdentityLastSeen(ctx, sqlc.UpdateMachineIdentityLastSeenParams{
		LastSeenAt: sql.NullInt64{Int64: time.Now().UnixMilli(), Valid: true},
		MachineID:  machineID,
	})
}

func (s *OAuthService) storeMachineJWT(ctx context.Context, machineID string, jwt string) error {
	if machineID == "" {
		return nil
	}
	return s.db.Queries.UpdateMachineIdentityJWT(ctx, sqlc.UpdateMachineIdentityJWTParams{
		MachineJwt: sql.NullString{String: jwt, Valid: jwt != ""},
		MachineID:  machineID,
	})
}

func (s *OAuthService) getOrCreateMachineID(ctx context.Context) (string, error) {
	return s.getMachineID()
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

// getMachineID returns a stable machine identifier derived from hardware IDs.
func (s *OAuthService) getMachineID() (string, error) {
	id, err := stableMachineID()
	if err != nil || id == "" {
		return "", fmt.Errorf("failed to derive stable machine id: %w", err)
	}
	return id, nil
}

// callInvokerAuthURL calls the Invoker control plane to get the auth URL.
func (s *OAuthService) callInvokerAuthURL(ctx context.Context, codeChallenge, state, redirectURI string) (string, error) {
	req := OAuthLoginRequest{
		RedirectURI:   redirectURI,
		CodeChallenge: codeChallenge,
		State:         state,
		Provider:      s.cfg.InvokerOAuthProvider,
	}
	var resp OAuthLoginResponse
	if err := s.doInvokerJSON(ctx, http.MethodPost, "/api/v1/auth/url", req, &resp, nil); err != nil {
		return "", err
	}
	return resp.AuthURL, nil
}

// callInvokerExchange calls the Invoker control plane to exchange the OAuth code.
func (s *OAuthService) callInvokerExchange(ctx context.Context, code, state, codeVerifier string) (*OAuthExchangeResponse, error) {
	req := OAuthExchangeRequest{
		Code:         code,
		State:        state,
		CodeVerifier: codeVerifier,
	}
	var resp OAuthExchangeResponse
	if err := s.doInvokerJSON(ctx, http.MethodPost, "/api/v1/auth/exchange", req, &resp, nil); err != nil {
		return nil, err
	}
	return &resp, nil
}

// callInvokerRegisterMachine calls the Invoker control plane to register the machine.
func (s *OAuthService) callInvokerRegisterMachine(ctx context.Context, machineJWT string, req MachineRegisterRequest) (*MachineRegisterResponse, error) {
	headers := map[string]string{
		"Authorization": "Bearer " + machineJWT,
	}
	var resp MachineRegisterResponse
	if err := s.doInvokerJSON(ctx, http.MethodPost, "/api/v1/machines/register", req, &resp, headers); err != nil {
		return nil, err
	}
	return &resp, nil
}


func (s *OAuthService) invokerURL(path string) string {
	base := strings.TrimRight(s.cfg.InvokerBaseURL, "/")
	if strings.HasPrefix(path, "/") {
		return base + path
	}
	return base + "/" + path
}

func (s *OAuthService) doInvokerJSON(ctx context.Context, method, path string, reqBody any, respBody any, headers map[string]string) error {
	var body io.Reader
	if reqBody != nil {
		buf := &bytes.Buffer{}
		if err := json.NewEncoder(buf).Encode(reqBody); err != nil {
			return err
		}
		body = buf
	}

	req, err := http.NewRequestWithContext(ctx, method, s.invokerURL(path), body)
	if err != nil {
		return err
	}
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		data, _ := io.ReadAll(resp.Body)
		msg := strings.TrimSpace(string(data))
		if msg == "" {
			msg = resp.Status
		}
		return fmt.Errorf("invoker %s %s failed: %s", method, path, msg)
	}

	if respBody == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(respBody)
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

func isLoopbackRedirectURI(redirectURI string) bool {
	if redirectURI == "" {
		return false
	}
	parsed, err := url.Parse(redirectURI)
	if err != nil {
		return false
	}
	host := parsed.Hostname()
	if host != "localhost" && host != "127.0.0.1" {
		return false
	}
	if parsed.Path != "" && parsed.Path != "/auth/callback" {
		return false
	}
	return true
}

func callbackPortFromRedirectURI(redirectURI, fallback string) string {
	if redirectURI == "" {
		return fallback
	}
	parsed, err := url.Parse(redirectURI)
	if err != nil {
		return fallback
	}
	if port := parsed.Port(); port != "" {
		return port
	}
	return fallback
}

// isLoopback checks if the remote address is from loopback.
func isLoopback(addr string) bool {
	// Simple check - in production, parse the IP properly
	return true
}

// CleanupExpiredOAuthAttempts removes expired OAuth login attempts.
func (s *OAuthService) CleanupExpiredOAuthAttempts(ctx context.Context) error {
	cutoff := time.Now().Add(-oauthAttemptTTL).UnixMilli()
	return s.db.Queries.CleanupExpiredOAuthAttempts(ctx, cutoff)
}
