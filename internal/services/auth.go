package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"

	"github.com/revrost/code/counterspell/internal/db"
	"github.com/revrost/code/counterspell/internal/db/sqlc"
)

// AuthService handles authentication and machine registration.
type AuthService struct {
	db               *db.DB
	controlPlane     *ControlPlaneClient
	machineID        string
	machineName      string
	defaultMachineID string

	// PKCE & State for auth flow
	codeVerifier string
	codeChallenge string
	state        string
}

// NewAuthService creates a new auth service.
func NewAuthService(database *db.DB, controlPlane *ControlPlaneClient, defaultMachineID string) *AuthService {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	machineName := fmt.Sprintf("%s-%s", runtime.GOOS, hostname)

	as := &AuthService{
		db:               database,
		controlPlane:     controlPlane,
		machineID:        defaultMachineID,
		machineName:      machineName,
		defaultMachineID: defaultMachineID,
	}

	// Generate PKCE code_verifier and code_challenge
	as.codeVerifier = generateRandomString(128)
	as.codeChallenge = generateCodeChallenge(as.codeVerifier)

	// Generate random state for CSRF protection
	as.state = generateRandomString(32)

	return as
}

// generateRandomString generates a cryptographically random string.
func generateRandomString(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic(fmt.Sprintf("failed to generate random string: %v", err))
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}

// generateCodeChallenge creates PKCE code_challenge from code_verifier.
func generateCodeChallenge(codeVerifier string) string {
	hash := sha256.Sum256([]byte(codeVerifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// generateMachineID generates a unique machine ID.
func (a *AuthService) generateMachineID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate machine ID: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// GetCodeVerifier returns the PKCE code_verifier.
func (a *AuthService) GetCodeVerifier() string {
	return a.codeVerifier
}

// GetCodeChallenge returns the PKCE code_challenge.
func (a *AuthService) GetCodeChallenge() string {
	return a.codeChallenge
}

// GetState returns the random state for CSRF protection.
func (a *AuthService) GetState() string {
	return a.state
}

// EnsureMachine ensures this machine exists in the database.
func (a *AuthService) EnsureMachine(ctx context.Context) (string, error) {
	machineID := a.defaultMachineID
	if machineID == "" {
		var err error
		machineID, err = a.generateMachineID()
		if err != nil {
			return "", err
		}
		a.machineID = machineID
	}

	now := time.Now().UnixMilli()
	capabilities := map[string]interface{}{
		"os":   runtime.GOOS,
		"arch": runtime.GOARCH,
		"cpus": runtime.NumCPU(),
	}
	capabilitiesJSON, _ := json.Marshal(capabilities)

	// Try to get existing machine
	existing, err := a.db.Queries.GetMachine(ctx, machineID)
	if err != nil {
		// Create new machine
		var capabilitiesPtr sql.NullString
		if len(capabilitiesJSON) > 0 {
			capabilitiesPtr = sql.NullString{String: string(capabilitiesJSON), Valid: true}
		}
		_, err = a.db.Queries.CreateMachine(ctx, sqlc.CreateMachineParams{
			ID:          machineID,
			Name:        a.machineName,
			Mode:        "local",
			Capabilities: capabilitiesPtr,
			CreatedAt:   now,
			UpdatedAt:   now,
			LastSeenAt:  now,
		})
		if err != nil {
			return "", fmt.Errorf("failed to create machine: %w", err)
		}
		slog.Info("Created new machine", "machine_id", machineID, "name", a.machineName)
	} else {
		// Update last seen
		err = a.db.Queries.UpdateMachineLastSeen(ctx, sqlc.UpdateMachineLastSeenParams{
			LastSeenAt: now,
			UpdatedAt:  now,
			ID:         existing.ID,
		})
		if err != nil {
			slog.Warn("Failed to update machine last seen", "error", err)
		}
	}

	return machineID, nil
}

// GetStoredAuth retrieves stored auth for this machine.
func (a *AuthService) GetStoredAuth(ctx context.Context, machineID string) (*sqlc.Auth, error) {
	auth, err := a.db.Queries.GetAuthByMachineID(ctx, machineID)
	if err != nil {
		return nil, err
	}
	return &auth, nil
}

// StoreAuth stores auth token and user info.
func (a *AuthService) StoreAuth(ctx context.Context, machineID, jwt, userID, email string, expiresAt int64) error {
	now := time.Now().UnixMilli()

	// Check if auth exists for this machine
	existing, err := a.db.Queries.GetAuthByMachineID(ctx, machineID)
	if err == nil {
		// Update existing
		err = a.db.Queries.UpdateAuth(ctx, sqlc.UpdateAuthParams{
			JwtToken:   jwt,
			ExpiresAt:  expiresAt,
			UpdatedAt:  now,
			ID:         existing.ID,
		})
		return err
	}

	// Create new auth
	var emailPtr sql.NullString
	if email != "" {
		emailPtr = sql.NullString{String: email, Valid: true}
	}
	_, err = a.db.Queries.CreateAuth(ctx, sqlc.CreateAuthParams{
		MachineID:  machineID,
		JwtToken:   jwt,
		UserID:     userID,
		Email:      emailPtr,
		ExpiresAt:  expiresAt,
		CreatedAt:  now,
		UpdatedAt:  now,
	})
	return err
}

// ValidateToken validates the stored JWT token.
func (a *AuthService) ValidateToken(ctx context.Context, jwt string) (bool, error) {
	if a.controlPlane == nil {
		return false, fmt.Errorf("control plane client not configured")
	}
	return a.controlPlane.ValidateToken(ctx, jwt)
}

// IsAuthenticated checks if this machine has valid auth.
func (a *AuthService) IsAuthenticated(ctx context.Context, machineID string) (bool, error) {
	auth, err := a.GetStoredAuth(ctx, machineID)
	if err != nil {
		return false, nil // No auth stored
	}

	// Check if token is expired
	now := time.Now().UnixMilli()
	if auth.ExpiresAt < now {
		return false, nil // Token expired
	}

	// Validate with control plane
	valid, err := a.ValidateToken(ctx, auth.JwtToken)
	if err != nil {
		slog.Error("Failed to validate token", "error", err)
		return false, nil
	}

	return valid, nil
}

// RegisterMachine registers this machine with the control plane.
func (a *AuthService) RegisterMachine(ctx context.Context, jwt, machineID string) (string, error) {
	if a.controlPlane == nil {
		return "", fmt.Errorf("control plane client not configured")
	}

	capabilities := map[string]interface{}{
		"os":   runtime.GOOS,
		"arch": runtime.GOARCH,
		"cpus": runtime.NumCPU(),
	}

	resp, err := a.controlPlane.RegisterMachine(ctx, jwt, machineID, a.machineName, capabilities)
	if err != nil {
		return "", fmt.Errorf("failed to register machine: %w", err)
	}

	return resp.Subdomain, nil
}

// StartAuthFlow starts the auth flow - returns auth URL and state.
func (a *AuthService) StartAuthFlow(ctx context.Context) (string, string, error) {
	if a.controlPlane == nil {
		return "", "", fmt.Errorf("control plane client not configured")
	}

	// For CLI flow, redirect URL can be a local callback or just the app URL
	redirectURL := "counterspell://auth/callback"

	resp, err := a.controlPlane.GetAuthURL(ctx, a.machineName, redirectURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to get auth URL: %w", err)
	}

	return resp.AuthURL, resp.State, nil
}

// CompleteAuthFlow exchanges the auth code for a JWT.
func (a *AuthService) CompleteAuthFlow(ctx context.Context, code, state string) (*ExchangeCodeResponse, error) {
	if a.controlPlane == nil {
		return nil, fmt.Errorf("control plane client not configured")
	}

	// Validate state matches to prevent CSRF
	if state != a.state {
		return nil, fmt.Errorf("invalid state - potential CSRF attack")
	}

	return a.controlPlane.ExchangeCode(ctx, code, a.codeVerifier, state)
}
