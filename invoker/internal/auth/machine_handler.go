package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/revrost/invoker/internal/cloudflare"
	"github.com/revrost/invoker/internal/config"
	"github.com/revrost/invoker/internal/db"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

// MachineHandler handles machine registration and management
type MachineHandler struct {
	db  db.Repository
	cfg *config.Config
	cf  *cloudflare.Client
}

// NewMachineHandler creates a new machine handler
func NewMachineHandler(database db.Repository, cfg *config.Config) *MachineHandler {
	var cf *cloudflare.Client
	if cfg.CloudflareAccountID != "" && cfg.CloudflareAPIToken != "" && cfg.CloudflareZoneName != "" {
		cf = cloudflare.NewClient(cfg.CloudflareAccountID, cfg.CloudflareAPIToken, cfg.CloudflareZoneName, cfg.CloudflareZoneID)
	}
	return &MachineHandler{
		db:  database,
		cfg: cfg,
		cf:  cf,
	}
}

// MachineRegisterRequest represents the machine registration request
type MachineRegisterRequest struct {
	MachineID string `json:"machine_id"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	Hostname  string `json:"hostname"`
	Version   string `json:"version"`
}

// MachineRegisterResponse represents the machine registration response
type MachineRegisterResponse struct {
	UserID         string `json:"user_id"`
	Subdomain      string `json:"subdomain"`
	TunnelToken    string `json:"tunnel_token"`
	TunnelProvider string `json:"tunnel_provider"`
}

// MachineInfo represents machine metadata for UI.
type MachineInfo struct {
	MachineID      string `json:"machine_id"`
	Subdomain      string `json:"subdomain"`
	TunnelProvider string `json:"tunnel_provider"`
	LastSeenAt     *int64 `json:"last_seen_at,omitempty"`
}

// MachineListResponse represents a list of machines for a user.
type MachineListResponse struct {
	Machines []MachineInfo `json:"machines"`
}

// ListMachines returns active machines for the authenticated user.
// GET /api/v1/machines
func (h *MachineHandler) ListMachines(w http.ResponseWriter, r *http.Request) {
	userID := extractUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	machines, err := h.db.GetMachineAuthByUserID(ctx, userID)
	if err != nil {
		slog.Error("Failed to list machines", "error", err)
		http.Error(w, "Failed to list machines", http.StatusInternalServerError)
		return
	}

	resp := MachineListResponse{
		Machines: make([]MachineInfo, 0, len(machines)),
	}

	for _, machine := range machines {
		var lastSeen *int64
		if machine.LastSeenAt != nil {
			ts := machine.LastSeenAt.UnixMilli()
			lastSeen = &ts
		}
		resp.Machines = append(resp.Machines, MachineInfo{
			MachineID:      machine.MachineID,
			Subdomain:      machine.Subdomain,
			TunnelProvider: machine.TunnelProvider,
			LastSeenAt:     lastSeen,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetMachine returns a single machine by id for the authenticated user.
// GET /api/v1/machines/{id}
func (h *MachineHandler) GetMachine(w http.ResponseWriter, r *http.Request) {
	userID := extractUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	machineID := chi.URLParam(r, "id")
	if machineID == "" {
		http.Error(w, "Missing machine id", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	machine, err := h.db.GetMachineAuthByMachineID(ctx, machineID)
	if err != nil {
		http.Error(w, "Machine not found", http.StatusNotFound)
		return
	}
	if machine.UserID != userID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	var lastSeen *int64
	if machine.LastSeenAt != nil {
		ts := machine.LastSeenAt.UnixMilli()
		lastSeen = &ts
	}

	resp := MachineInfo{
		MachineID:      machine.MachineID,
		Subdomain:      machine.Subdomain,
		TunnelProvider: machine.TunnelProvider,
		LastSeenAt:     lastSeen,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// RegisterMachine registers a machine and provisions a tunnel
// POST /api/v1/machines/register
func (h *MachineHandler) RegisterMachine(w http.ResponseWriter, r *http.Request) {
	var req MachineRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.MachineID == "" {
		http.Error(w, "Missing machine_id", http.StatusBadRequest)
		return
	}

	// Extract user ID from JWT
	userID := extractUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()

	// Check if machine already exists
	existingMachine, err := h.db.GetMachineAuthByMachineID(ctx, req.MachineID)
	if err == nil && existingMachine != nil {
		if existingMachine.TunnelProvider == "mock" && h.cf != nil {
			tunnelToken, tunnelProvider, err := h.provisionTunnel(ctx, existingMachine.Subdomain, req.MachineID)
			if err != nil {
				slog.Error("Failed to provision tunnel", "error", err)
				http.Error(w, "Failed to provision tunnel", http.StatusInternalServerError)
				return
			}
			updated, err := h.db.UpdateMachineAuthTunnel(ctx, existingMachine.MachineID, tunnelProvider, tunnelToken, existingMachine.Subdomain)
			if err != nil {
				slog.Error("Failed to update machine tunnel", "error", err)
				http.Error(w, "Failed to update machine", http.StatusInternalServerError)
				return
			}
			response := MachineRegisterResponse{
				UserID:         updated.UserID,
				Subdomain:      updated.Subdomain,
				TunnelToken:    updated.TunnelToken,
				TunnelProvider: updated.TunnelProvider,
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
		// Machine already registered, return existing data
		response := MachineRegisterResponse{
			UserID:         existingMachine.UserID,
			Subdomain:      existingMachine.Subdomain,
			TunnelToken:    existingMachine.TunnelToken,
			TunnelProvider: existingMachine.TunnelProvider,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Generate subdomain from user info
	user, err := h.db.GetUserByID(ctx, userID)
	if err != nil {
		slog.Error("Failed to get user", "user_id", userID, "error", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	subdomain := h.generateSubdomain(user.Username)

	// Ensure subdomain is unique
	subdomain, err = h.ensureUniqueSubdomain(ctx, subdomain)
	if err != nil {
		slog.Error("Failed to generate unique subdomain", "error", err)
		http.Error(w, "Failed to generate subdomain", http.StatusInternalServerError)
		return
	}

	tunnelToken, tunnelProvider, err := h.provisionTunnel(ctx, subdomain, req.MachineID)
	if err != nil {
		slog.Error("Failed to provision tunnel", "error", err)
		http.Error(w, "Failed to provision tunnel", http.StatusInternalServerError)
		return
	}

	// Create machine auth record
	machineAuth := &db.MachineAuth{
		ID:             generateUUID(),
		MachineID:      req.MachineID,
		UserID:         userID,
		Subdomain:      subdomain,
		TunnelProvider: tunnelProvider,
		TunnelToken:    tunnelToken,
		CreatedAt:      time.Now(),
		LastSeenAt:     nil,
		IsActive:       true,
	}

	if err := h.db.CreateMachineAuth(ctx, machineAuth); err != nil {
		slog.Error("Failed to create machine auth", "error", err)
		http.Error(w, "Failed to register machine", http.StatusInternalServerError)
		return
	}

	response := MachineRegisterResponse{
		UserID:         userID,
		Subdomain:      subdomain,
		TunnelToken:    tunnelToken,
		TunnelProvider: tunnelProvider,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	slog.Info("Machine registered", "machine_id", req.MachineID, "user_id", userID, "subdomain", subdomain)
}

// generateSubdomain generates a subdomain from username
func (h *MachineHandler) generateSubdomain(username string) string {
	// For now, just use username as subdomain
	// In production, you'd want to sanitize this more
	return username
}

// ensureUniqueSubdomain ensures subdomain is unique
func (h *MachineHandler) ensureUniqueSubdomain(ctx context.Context, subdomain string) (string, error) {
	originalSubdomain := subdomain
	counter := 1

	for {
		// Check if subdomain exists
		_, err := h.db.GetMachineAuthBySubdomain(ctx, subdomain)
		if err != nil {
			// Subdomain doesn't exist, we can use it
			return subdomain, nil
		}

		// Subdomain exists, try with counter
		subdomain = fmt.Sprintf("%s%d", originalSubdomain, counter)
		counter++

		// Safety limit
		if counter > 1000 {
			return "", fmt.Errorf("could not generate unique subdomain")
		}
	}
}

// generateTunnelToken generates a Cloudflare tunnel token (mock)
func (h *MachineHandler) generateTunnelToken(subdomain string) string {
	// In production, this would call Cloudflare API to create a tunnel
	// For now, return a mock token
	return fmt.Sprintf("mock_tunnel_token_%s_%d", subdomain, time.Now().Unix())
}

func (h *MachineHandler) provisionTunnel(ctx context.Context, subdomain, machineID string) (string, string, error) {
	if h.cf == nil {
		slog.Warn("Cloudflare not configured, using mock tunnel token")
		return h.generateTunnelToken(subdomain), "mock", nil
	}

	name := h.buildTunnelName(subdomain, machineID)
	tunnel, err := h.cf.CreateTunnel(ctx, name)
	if err != nil {
		return "", "", err
	}
	slog.Info("Cloudflare tunnel created", "tunnel_id", tunnel.ID, "name", name, "subdomain", subdomain)
	if err := h.cf.EnsureDNSRecord(ctx, subdomain, tunnel.ID); err != nil {
		return "", "", err
	}
	slog.Info("Cloudflare DNS record ensured", "subdomain", subdomain, "zone", h.cfg.CloudflareZoneName)
	return tunnel.Token, "cloudflare", nil
}

func (h *MachineHandler) buildTunnelName(subdomain, machineID string) string {
	id := strings.ReplaceAll(machineID, "-", "")
	if len(id) > 8 {
		id = id[:8]
	}
	name := fmt.Sprintf("cs-%s-%s", subdomain, id)
	name = strings.ToLower(name)
	if len(name) > 63 {
		name = name[:63]
	}
	return name
}

// RevokeMachine revokes a machine and its tunnel
// POST /api/v1/machines/{machine_id}/revoke
func (h *MachineHandler) RevokeMachine(w http.ResponseWriter, r *http.Request) {
	machineID := chi.URLParam(r, "machine_id")
	if machineID == "" {
		http.Error(w, "Missing machine_id", http.StatusBadRequest)
		return
	}

	// Extract user ID from JWT
	userID := extractUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()

	// Get machine to verify ownership
	machine, err := h.db.GetMachineAuthByMachineID(ctx, machineID)
	if err != nil {
		slog.Error("Machine not found", "machine_id", machineID, "error", err)
		http.Error(w, "Machine not found", http.StatusNotFound)
		return
	}

	// Verify user owns this machine
	if machine.UserID != userID {
		slog.Warn("Unauthorized revoke attempt", "machine_id", machineID, "user_id", userID, "owner_id", machine.UserID)
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Revoke machine
	if err := h.db.RevokeMachineAuth(ctx, machineID); err != nil {
		slog.Error("Failed to revoke machine", "machine_id", machineID, "error", err)
		http.Error(w, "Failed to revoke machine", http.StatusInternalServerError)
		return
	}

	// TODO: Revoke Cloudflare tunnel via API
	// TODO: Delete DNS record

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":     "revoked",
		"machine_id": machineID,
	})

	slog.Info("Machine revoked", "machine_id", machineID, "user_id", userID)
}

// extractUserIDFromContext extracts user ID from request context
func extractUserIDFromContext(ctx context.Context) string {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return ""
	}
	return userID
}

// MachineClaims represents JWT claims for machine authentication
type MachineClaims struct {
	jwt.RegisteredClaims
	UserID    string `json:"user_id"`
	MachineID string `json:"machine_id,omitempty"`
}

// RequireMachineJWT middleware validates machine JWT and adds user_id to context
func RequireMachineJWT(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
				return
			}

			// Remove "Bearer " prefix
			tokenString := authHeader
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				tokenString = authHeader[7:]
			}

			// Parse and validate JWT
			token, err := jwt.ParseWithClaims(tokenString, &MachineClaims{}, func(token *jwt.Token) (interface{}, error) {
				// Verify signing method
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err != nil {
				slog.Error("Failed to parse JWT", "error", err)
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(*MachineClaims)
			if !ok || !token.Valid {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			// Check expiration
			if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
				http.Error(w, "Token expired", http.StatusUnauthorized)
				return
			}

			// Add user_id to context
			ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetSystemInfo returns system information for machine registration
func GetSystemInfo() map[string]string {
	return map[string]string{
		"os":       runtime.GOOS,
		"arch":     runtime.GOARCH,
		"hostname": getHostname(),
	}
}

func getHostname() string {
	// In production, get actual hostname
	return "localhost"
}
