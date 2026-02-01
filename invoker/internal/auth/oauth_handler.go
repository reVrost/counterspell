package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/revrost/invoker/internal/config"
	"github.com/revrost/invoker/internal/db"
	"github.com/revrost/invoker/internal/db/sqlc"
	"github.com/revrost/invoker/pkg/models"
)

// OAuthHandler handles OAuth-related requests for Counterspell CLI authentication
type OAuthHandler struct {
	supabase *SupabaseAuth
	db       db.Repository
	cfg      *config.Config
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(supabase *SupabaseAuth, database db.Repository, cfg *config.Config) *OAuthHandler {
	return &OAuthHandler{
		supabase: supabase,
		db:       database,
		cfg:      cfg,
	}
}

// CreateAuthURLRequest represents the request to create an auth URL
type CreateAuthURLRequest struct {
	RedirectURI   string `json:"redirect_uri"`
	CodeChallenge string `json:"code_challenge"`
	State         string `json:"state"`
	Provider      string `json:"provider"`
}

// CreateAuthURLResponse represents the response containing the auth URL
type CreateAuthURLResponse struct {
	AuthURL string `json:"auth_url"`
}

// CreateCLIAuthURL creates an OAuth authorization URL for a Counterspell CLI login
// POST /api/v1/auth/url
func (h *OAuthHandler) CreateCLIAuthURL(w http.ResponseWriter, r *http.Request) {
	var req CreateAuthURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	slog.Info("OAuth auth_url request",
		"redirect_uri", req.RedirectURI,
		"provider", req.Provider,
		"state_present", req.State != "",
	)

	// Validate redirect_uri against allowlist
	if err := h.validateRedirectURI(req.RedirectURI); err != nil {
		slog.Error("Invalid redirect_uri", "redirect_uri", req.RedirectURI, "error", err)
		http.Error(w, "Invalid redirect_uri", http.StatusBadRequest)
		return
	}

	// Generate state if not provided
	if req.State == "" {
		state, err := generateState()
		if err != nil {
			slog.Error("Failed to generate state", "error", err)
			http.Error(w, "Failed to generate state", http.StatusInternalServerError)
			return
		}
		req.State = state
	}

	// Generate code challenge if not provided
	if req.CodeChallenge == "" {
		// In PKCE, the client generates the code_challenge from code_verifier
		// For now, we'll use what's provided or skip validation
	}

	if req.Provider == "" {
		req.Provider = "google"
	}

	// Persist pending login record
	ctx := r.Context()
	login := &db.PendingOAuthLogin{
		ID:            generateUUID(),
		State:         req.State,
		CodeChallenge: req.CodeChallenge,
		RedirectURI:   req.RedirectURI,
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(10 * time.Minute), // 10 minute expiry
	}

	if err := h.db.CreatePendingOAuthLogin(ctx, login); err != nil {
		slog.Error("Failed to create pending OAuth login", "error", err)
		http.Error(w, "Failed to create pending login", http.StatusInternalServerError)
		return
	}

	// Generate Supabase OAuth URL with PKCE
	authURL, err := h.generateSupabaseAuthURL(req.RedirectURI, req.CodeChallenge, req.State, req.Provider)
	if err != nil {
		slog.Error("Failed to generate Supabase auth URL", "error", err)
		http.Error(w, "Failed to generate auth URL", http.StatusInternalServerError)
		return
	}

	slog.Info("OAuth auth_url generated", "auth_url", authURL)

	response := CreateAuthURLResponse{
		AuthURL: authURL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// validateRedirectURI validates that the redirect URI is in the allowlist
func (h *OAuthHandler) validateRedirectURI(redirectURI string) error {
	// Allowlist of valid redirect URIs
	allowedURIs := []string{
		"http://localhost:8711/auth/callback",
		"https://counterspell.io/api/v1/auth/callback",
		"https://counterspell.io/auth/callback",
	}

	parsedURI, err := url.Parse(redirectURI)
	if err != nil {
		return fmt.Errorf("invalid redirect URI format")
	}

	if isLocalhost(parsedURI.Hostname()) {
		if parsedURI.Path == "/api/v1/auth/callback" || parsedURI.Path == "/auth/callback" {
			return nil
		}
	}

	for _, allowed := range allowedURIs {
		parsedAllowed, err := url.Parse(allowed)
		if err != nil {
			continue
		}
		if parsedURI.Host == parsedAllowed.Host && parsedURI.Path == parsedAllowed.Path {
			return nil
		}
	}

	return fmt.Errorf("redirect URI not in allowlist")
}

func isLocalhost(host string) bool {
	return host == "localhost" || host == "127.0.0.1"
}

// generateSupabaseAuthURL generates the Supabase OAuth URL with PKCE
func (h *OAuthHandler) generateSupabaseAuthURL(redirectURI, codeChallenge, state, provider string) (string, error) {
	if h.supabase == nil || h.supabase.supabaseURL == "" {
		return "", fmt.Errorf("Supabase not configured")
	}

	baseURL := h.supabase.supabaseURL
	if baseURL[len(baseURL)-1] != '/' {
		baseURL += "/"
	}

	authEndpoint := baseURL + "auth/v1/authorize"

	redirectTo := redirectURI
	if state != "" {
		if parsedRedirect, err := url.Parse(redirectURI); err == nil {
			q := parsedRedirect.Query()
			if q.Get("cs_state") == "" {
				q.Set("cs_state", state)
				parsedRedirect.RawQuery = q.Encode()
				redirectTo = parsedRedirect.String()
			}
		}
	}

	params := url.Values{}
	params.Set("provider", provider)
	params.Set("redirect_to", redirectTo)
	if codeChallenge != "" {
		params.Set("code_challenge", codeChallenge)
		params.Set("code_challenge_method", "S256")
	}
	authURL := authEndpoint + "?" + params.Encode()
	if parsed, err := url.Parse(authEndpoint); err == nil {
		slog.Info("Supabase authorize URL built",
			"supabase_host", parsed.Host,
			"redirect_to", redirectTo,
			"provider", provider,
		)
	}
	return authURL, nil
}

// generateState generates a cryptographically secure random state
func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// generateUUID generates a random UUID
func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// generateCodeChallenge generates a PKCE code challenge from a verifier
func generateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// generateCodeVerifier generates a PKCE code verifier
func generateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// OAuthCallbackRequest represents the OAuth callback request
type OAuthCallbackRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

// OAuthCallback handles OAuth redirect from Supabase after user login
// GET /auth/callback
func (h *OAuthHandler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	csState := r.URL.Query().Get("cs_state")
	if csState != "" {
		state = csState
	}

	if code == "" || state == "" {
		http.Error(w, "Missing code or state parameter", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Validate state against pending login
	if _, err := h.db.GetPendingOAuthLoginByState(ctx, state); err != nil {
		slog.Error("Invalid or expired state", "state", state, "error", err)
		http.Error(w, "Invalid or expired state", http.StatusBadRequest)
		return
	}

	// Store auth code for CLI polling
	if _, err := h.db.Queries().UpdatePendingOAuthLoginAuthCode(ctx, sqlc.UpdatePendingOAuthLoginAuthCodeParams{
		State:    state,
		AuthCode: code,
	}); err != nil {
		slog.Error("Failed to store auth code", "error", err)
		http.Error(w, "Failed to store auth code", http.StatusInternalServerError)
		return
	}

	// Note: User sync happens in the exchange endpoint (POST /api/v1/auth/exchange)
	// This is a browser redirect callback, so we just validate and respond
	// The CLI will poll /api/v1/auth/poll to retrieve the code

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
}

// syncUserFromSupabase ensures user exists in our database
func (h *OAuthHandler) syncUserFromSupabase(ctx context.Context, supabaseResp *SupabaseAuthResponse) (*models.User, error) {
	if supabaseResp == nil || supabaseResp.User == nil {
		return nil, fmt.Errorf("missing user in supabase response")
	}
	if supabaseResp.User.Email == "" {
		return nil, fmt.Errorf("missing email in supabase response")
	}
	if supabaseResp.User.ID == "" {
		return nil, fmt.Errorf("missing user id in supabase response")
	}
	// Check if user exists
	existingUser, err := h.db.GetUserByEmail(ctx, supabaseResp.User.Email)
	if err == nil && existingUser != nil {
		return existingUser, nil
	}

	// Create new user
	// This is simplified - in production you'd extract more fields from metadata
	meta := supabaseResp.User.UserMetadata
	firstName := safeMetaString(meta, "first_name")
	lastName := safeMetaString(meta, "last_name")
	if firstName == "" && lastName == "" {
		fullName := safeMetaString(meta, "full_name")
		if fullName == "" {
			fullName = safeMetaString(meta, "name")
		}
		if fullName != "" {
			parts := strings.Fields(fullName)
			if len(parts) > 0 {
				firstName = parts[0]
			}
			if len(parts) > 1 {
				lastName = strings.Join(parts[1:], " ")
			}
		}
	}
	if firstName == "" {
		firstName = "User"
	}

	username := generateUsernameFromEmail(supabaseResp.User.Email)
	user := &models.User{
		ID:        supabaseResp.User.ID,
		Email:     supabaseResp.User.Email,
		FirstName: firstName,
		LastName:  lastName,
		Username:  username,
		Tier:      "free",
	}

	// Ensure unique username
	user.Username, err = ensureUniqueUsername(ctx, h.db, username)
	if err != nil {
		return nil, err
	}

	if err := h.db.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func safeMetaString(meta map[string]interface{}, key string) string {
	if meta == nil {
		return ""
	}
	val, ok := meta[key]
	if !ok || val == nil {
		return ""
	}
	if s, ok := val.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", val)
}

// generateUsernameFromEmail generates a username from email
func generateUsernameFromEmail(email string) string {
	localPart := email
	if atIndex := strings.LastIndex(email, "@"); atIndex > 0 {
		localPart = email[:atIndex]
	}

	// Remove special characters, keep only alphanumeric
	var b strings.Builder
	for _, c := range localPart {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			b.WriteRune(c)
		}
	}

	username := strings.ToLower(b.String())
	if len(username) < 3 {
		username = "user"
	}
	if len(username) > 20 {
		username = username[:20]
	}

	return username
}

// ExchangeAuthRequest represents the auth exchange request
type ExchangeAuthRequest struct {
	Code         string `json:"code"`
	CodeVerifier string `json:"code_verifier"`
	State        string `json:"state"`
}

// ExchangeAuthResponse represents the response with machine JWT
type ExchangeAuthResponse struct {
	MachineJWT string `json:"machine_jwt"`
	UserID     string `json:"user_id"`
	UserEmail  string `json:"user_email"`
}

// PollAuthRequest represents a poll request for CLI auth
type PollAuthRequest struct {
	State string `json:"state"`
}

// PollAuthResponse represents a poll response for CLI auth
type PollAuthResponse struct {
	Status string `json:"status"`
	Code   string `json:"code,omitempty"`
}

// PollOAuthCode polls for an OAuth code tied to a pending login
// POST /api/v1/auth/poll
func (h *OAuthHandler) PollOAuthCode(w http.ResponseWriter, r *http.Request) {
	var req PollAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.State == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	pendingLogin, err := h.db.GetPendingOAuthLoginByState(ctx, req.State)
	if err != nil {
		http.Error(w, "Pending login not found", http.StatusNotFound)
		return
	}

	if pendingLogin.AuthCode == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(PollAuthResponse{Status: "pending"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PollAuthResponse{Status: "ready", Code: pendingLogin.AuthCode})
}

// ExchangeOAuthCode exchanges OAuth code for a machine-scoped JWT
// POST /api/v1/auth/exchange
func (h *OAuthHandler) ExchangeOAuthCode(w http.ResponseWriter, r *http.Request) {
	var req ExchangeAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Code == "" || req.State == "" {
		http.Error(w, "Missing code or state", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Validate state against pending login
	pendingLogin, err := h.db.GetPendingOAuthLoginByState(ctx, req.State)
	if err != nil {
		slog.Error("Invalid or expired state", "state", req.State, "error", err)
		http.Error(w, "Invalid or expired state", http.StatusBadRequest)
		return
	}

	// Verify code_challenge if provided (PKCE)
	if req.CodeVerifier != "" && pendingLogin.CodeChallenge != "" {
		expectedChallenge := generateCodeChallenge(req.CodeVerifier)
		if expectedChallenge != pendingLogin.CodeChallenge {
			slog.Error("Invalid code_challenge", "state", req.State)
			http.Error(w, "Invalid code challenge", http.StatusBadRequest)
			return
		}
	}

	// Exchange authorization code with Supabase
	supabaseResp, err := h.supabase.ExchangeCodeForToken(req.Code, req.CodeVerifier)
	if err != nil {
		slog.Error("Failed to exchange code with Supabase", "error", err)
		http.Error(w, "Failed to exchange code", http.StatusInternalServerError)
		return
	}

	// Create or get user
	user, err := h.syncUserFromSupabase(ctx, supabaseResp)
	if err != nil {
		slog.Error("Failed to sync user", "error", err)
		http.Error(w, "Failed to sync user", http.StatusInternalServerError)
		return
	}

	// Mint machine-scoped JWT
	machineJWT, err := h.generateMachineJWT(user.ID)
	if err != nil {
		slog.Error("Failed to generate machine JWT", "error", err)
		http.Error(w, "Failed to generate machine JWT", http.StatusInternalServerError)
		return
	}

	// Delete pending login
	if err := h.db.DeletePendingOAuthLogin(ctx, req.State); err != nil {
		slog.Error("Failed to delete pending login", "error", err)
		// Don't fail the request
	}

	response := ExchangeAuthResponse{
		MachineJWT: machineJWT,
		UserID:     user.ID,
		UserEmail:  user.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// generateMachineJWT generates a machine-scoped JWT
func (h *OAuthHandler) generateMachineJWT(userID string) (string, error) {
	return generateMachineJWT(h.cfg.JWTSecret, userID, "")
}
