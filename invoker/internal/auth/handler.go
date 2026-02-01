package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/revrost/invoker/internal/db"
	"github.com/revrost/invoker/pkg/models"
)

// Handler handles authentication requests
type Handler struct {
	supabase   *SupabaseAuth
	db         db.Repository
	flyService interface{} // Will be *fly.Service, but keep as interface{} to avoid circular dependency
}

// NewHandler creates a new auth handler
func NewHandler(supabase *SupabaseAuth, database db.Repository, flyService interface{}) *Handler {
	return &Handler{
		supabase:   supabase,
		db:         database,
		flyService: flyService,
	}
}

// SyncProfile creates/updates a user profile in the database after Supabase auth
func (h *Handler) SyncProfile(w http.ResponseWriter, r *http.Request) {
	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	// Remove "Bearer " prefix
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
		return
	}

	// Validate token with Supabase
	claims, err := h.supabase.ValidateToken(token)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
		return
	}

	// Create a mock response with user info
	supabaseResp := &SupabaseAuthResponse{
		AccessToken: token,
		User: &SupabaseUser{
			ID:    ExtractUserID(claims),
			Email: ExtractEmail(claims),
			UserMetadata: map[string]interface{}{
				"full_name": claims.UserMetadata["full_name"],
			},
		},
	}

	// Sync user to database
	user, err := h.syncProfile(r.Context(), supabaseResp)
	if err != nil {
		slog.Error("Failed to sync user profile", "error", err)
		// Don't fail the request - frontend already has the token
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"profile": user,
	}); err != nil {
		slog.Error("Failed to encode sync response", "error", err)
	}
}

// GetProfile returns the user profile for the authenticated user
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by middleware)
	userID := ExtractUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	user, err := h.db.GetUserByID(r.Context(), userID)
	if err != nil {
		slog.Error("Failed to fetch user profile", "user_id", userID, "error", err)
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		slog.Error("Failed to encode profile response", "error", err)
	}
}

// syncProfile ensures Supabase user has a profile in our database
func (h *Handler) syncProfile(ctx context.Context, supabaseResp *SupabaseAuthResponse) (*models.User, error) {
	email := supabaseResp.User.Email

	existingUser, err := h.db.GetUserByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return existingUser, nil
	}

	// Extract metadata or defaults
	firstName := "User"
	lastName := ""
	if meta, ok := supabaseResp.User.UserMetadata["full_name"].(string); ok {
		parts := strings.Split(meta, " ")
		firstName = parts[0]
		if len(parts) > 1 {
			lastName = strings.Join(parts[1:], " ")
		}
	} else {
		if fn, ok := supabaseResp.User.UserMetadata["first_name"].(string); ok {
			firstName = fn
		}
		if ln, ok := supabaseResp.User.UserMetadata["last_name"].(string); ok {
			lastName = ln
		}
	}

	username := generateUsername(firstName, email)
	username, err = ensureUniqueUsername(ctx, h.db, username)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:        supabaseResp.User.ID,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Username:  username,
		Tier:      "free",
	}

	if err := h.db.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	// Provision Fly.io machine for new user
	if h.flyService != nil {
		if flySvc, ok := h.flyService.(interface {
			ProvisionUserMachine(context.Context, *models.User) (*models.MachineRegistry, error)
		}); ok {
			machine, err := flySvc.ProvisionUserMachine(ctx, user)
			if err != nil {
				// Don't fail user creation, just log the error
				slog.Error("Failed to provision Fly.io machine for user", "user_id", user.ID, "error", err)
			} else {
				slog.Info("Fly.io machine provisioned for user", "user_id", user.ID, "machine_id", machine.FlyMachineID)
			}
		}
	}

	return user, nil
}

// generateUsername generates a username from first name and email
func generateUsername(firstName, email string) string {
	// Clean first name: lowercase, remove special chars
	cleanName := strings.ToLower(firstName)
	reg := regexp.MustCompile("[^a-z0-9]")
	cleanName = reg.ReplaceAllString(cleanName, "")

	// If name is too short, use part of email
	if len(cleanName) < 3 {
		emailParts := strings.Split(email, "@")
		username := strings.ToLower(emailParts[0])
		reg := regexp.MustCompile("[^a-z0-9]")
		username = reg.ReplaceAllString(username, "")
		if len(username) > 3 {
			username = username[:3]
		}
		return username
	}

	// Truncate to max 20 chars
	if len(cleanName) > 20 {
		cleanName = cleanName[:20]
	}

	return cleanName
}

// ensureUniqueUsername ensures username is unique
func ensureUniqueUsername(ctx context.Context, db db.Repository, username string) (string, error) {
	originalUsername := username
	counter := 1

	for {
		exists, err := db.UsernameExists(ctx, username)
		if err != nil {
			return "", err
		}

		if !exists {
			return username, nil
		}

		// Append counter and try again
		newUsername := fmt.Sprintf("%s%d", originalUsername, counter)
		if len(newUsername) > 20 {
			// Truncate and append
			maxLen := 20 - len(fmt.Sprintf("%d", counter))
			newUsername = fmt.Sprintf("%s%d", originalUsername[:maxLen], counter)
		}

		username = newUsername
		counter++

		// Safety limit
		if counter > 1000 {
			return "", fmt.Errorf("could not generate unique username")
		}
	}
}
