package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// ControlPlaneClient talks to the control plane API (counterspell.io).
// The control plane handles auth with Supabase internally - the binary doesn't know about Supabase.
type ControlPlaneClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewControlPlaneClient creates a new control plane client.
func NewControlPlaneClient(baseURL string) *ControlPlaneClient {
	return &ControlPlaneClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// AuthURLRequest is the request to get an auth URL.
type AuthURLRequest struct {
	MachineName    string `json:"machine_name"`
	RedirectURL    string `json:"redirect_url"`
	CodeChallenge  string `json:"code_challenge"` // PKCE - hash of code_verifier
	State          string `json:"state"`          // CSRF protection
}

// AuthURLResponse is the response containing the auth URL.
type AuthURLResponse struct {
	AuthURL string `json:"auth_url"`
	State   string `json:"state"`
}

// GetAuthURL returns the auth URL that the user should visit to log in.
func (c *ControlPlaneClient) GetAuthURL(ctx context.Context, machineName, redirectURL string) (*AuthURLResponse, error) {
	// Deprecated - use GetAuthURLWithPKCE for security
	return c.GetAuthURLWithPKCE(ctx, machineName, redirectURL, "", "")
}

// GetAuthURLWithPKCE returns the auth URL with PKCE for security.
func (c *ControlPlaneClient) GetAuthURLWithPKCE(ctx context.Context, machineName, redirectURL, codeChallenge, state string) (*AuthURLResponse, error) {
	reqBody := AuthURLRequest{
		MachineName:   machineName,
		RedirectURL:   redirectURL,
		CodeChallenge: codeChallenge,
		State:         state,
	}

	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/auth/url", &readerCloser{Reader: reqData})
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call control plane: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("control plane returned status %d", resp.StatusCode)
	}

	var response AuthURLResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// ExchangeCodeRequest exchanges an auth code for a JWT token.
type ExchangeCodeRequest struct {
	Code         string `json:"code"`
	CodeVerifier string `json:"code_verifier"` // PKCE - original random string
	State        string `json:"state"`
}

// ExchangeCodeResponse contains the JWT token and user info.
type ExchangeCodeResponse struct {
	JWT       string `json:"jwt"`
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	ExpiresAt int64  `json:"expires_at"` // Unix ms
}

// ExchangeCode exchanges the auth callback code for a JWT token.
func (c *ControlPlaneClient) ExchangeCode(ctx context.Context, code, codeVerifier, state string) (*ExchangeCodeResponse, error) {
	reqBody := ExchangeCodeRequest{
		Code:         code,
		CodeVerifier: codeVerifier,
		State:        state,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/auth/exchange", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Body = &readerCloser{Reader: body}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call control plane: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("control plane returned status %d", resp.StatusCode)
	}

	var response ExchangeCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// RegisterMachineRequest registers a machine with the control plane.
type RegisterMachineRequest struct {
	MachineID     string                 `json:"machine_id"`
	MachineName   string                 `json:"machine_name"`
	Mode          string                 `json:"mode"` // "local" or "cloud"
	Capabilities  map[string]interface{} `json:"capabilities"`
}

// RegisterMachineResponse contains the tunnel/subdomain info.
type RegisterMachineResponse struct {
	Subdomain string `json:"subdomain"`
	TunnelURL string `json:"tunnel_url"`
}

// RegisterMachine registers this machine and gets tunnel configuration.
func (c *ControlPlaneClient) RegisterMachine(ctx context.Context, jwt, machineID, machineName string, capabilities map[string]interface{}) (*RegisterMachineResponse, error) {
	reqBody := RegisterMachineRequest{
		MachineID:    machineID,
		MachineName:  machineName,
		Mode:         "local",
		Capabilities: capabilities,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/machines/register", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Body = &readerCloser{Reader: body}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call control plane: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Failed to register machine", "status", resp.StatusCode)
		return nil, fmt.Errorf("control plane returned status %d", resp.StatusCode)
	}

	var response RegisterMachineResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// ValidateToken validates a JWT token with the control plane.
func (c *ControlPlaneClient) ValidateToken(ctx context.Context, jwt string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/v1/auth/validate", nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+jwt)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to call control plane: %w", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// readerCloser wraps a []byte to implement io.ReadCloser.
type readerCloser struct {
	Reader []byte
}

func (r *readerCloser) Read(p []byte) (int, error) {
	return copy(p, r.Reader), nil
}

func (r *readerCloser) Close() error {
	return nil
}
