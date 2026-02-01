package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/supabase-community/gotrue-go/types"
	"github.com/supabase-community/supabase-go"
)

// SupabaseAuth handles Supabase authentication using official client
type SupabaseAuth struct {
	client      *supabase.Client
	jwtSecret   string
	supabaseURL string
	anonKey     string
	httpClient  *http.Client
}

// NewSupabaseAuth creates a new Supabase auth client
func NewSupabaseAuth(supabaseURL, anonKey, jwtSecret string) (*SupabaseAuth, error) {
	client, err := supabase.NewClient(supabaseURL, anonKey, &supabase.ClientOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create Supabase client: %w", err)
	}

	return &SupabaseAuth{
		client:      client,
		jwtSecret:   jwtSecret,
		supabaseURL: supabaseURL,
		anonKey:     anonKey,
		httpClient:  &http.Client{Timeout: 15 * time.Second},
	}, nil
}

// Claims represents JWT claims from Supabase tokens
type Claims struct {
	jwt.RegisteredClaims
	Email        string                 `json:"email"`
	Phone        string                 `json:"phone"`
	AppMetadata  map[string]interface{} `json:"app_metadata"`
	UserMetadata map[string]interface{} `json:"user_metadata"`
	Role         string                 `json:"role"`
	AAL          string                 `json:"aal"`
	AMR          []struct {
		Method    string `json:"method"`
		Timestamp string `json:"timestamp"`
	} `json:"amr"`
	SessionID string `json:"session_id"`
}

// ValidateToken validates a Supabase JWT token using JWT secret
func (sa *SupabaseAuth) ValidateToken(tokenString string) (*Claims, error) {
	if sa.jwtSecret == "" {
		return nil, errors.New("SUPABASE_JWT_SECRET not configured")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing algorithm (Supabase uses HS256 with JWT_SECRET)
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unsupported signing method: %v", token.Header["alg"])
		}
		return []byte(sa.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("token expired")
	}

	return claims, nil
}

// Signup registers a new user with Supabase Auth and returns JWT token
func (sa *SupabaseAuth) Signup(email, password string) (*SupabaseAuthResponse, error) {
	// Use Supabase Auth client for signup
	authResp, err := sa.client.Auth.Signup(types.SignupRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return nil, fmt.Errorf("signup failed: %w", err)
	}

	// Convert to our response format
	response := &SupabaseAuthResponse{
		AccessToken: authResp.AccessToken,
		TokenType:   authResp.TokenType,
		ExpiresIn:   authResp.ExpiresIn,
	}

	// Signup response has User directly if autoconfirm is off, or Session with User if autoconfirm is on
	var user types.User
	if authResp.User.ID != (types.User{}).ID {
		user = authResp.User
	} else if authResp.Session.User.ID != (types.User{}).ID {
		user = authResp.Session.User
	}

	if user.ID != (types.User{}).ID {
		response.User = &SupabaseUser{
			ID:               user.ID.String(),
			Email:            user.Email,
			CreatedAt:        user.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:        user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			UserMetadata:     user.UserMetadata,
			AppMetadata:      user.AppMetadata,
			PhoneConfirmedAt: formatTimeToString(user.PhoneConfirmedAt),
		}
		if user.EmailConfirmedAt != nil {
			confirmedAt := user.EmailConfirmedAt.Format("2006-01-02T15:04:05Z")
			response.User.EmailConfirmedAt = &confirmedAt
		}
		if user.LastSignInAt != nil {
			lastSignIn := user.LastSignInAt.Format("2006-01-02T15:04:05Z")
			response.User.LastSignInAt = &lastSignIn
		}
	}

	return response, nil
}

// Login authenticates a user with Supabase Auth and returns JWT token
func (sa *SupabaseAuth) Login(email, password string) (*SupabaseAuthResponse, error) {
	// Use Supabase Auth client for login
	authResp, err := sa.client.Auth.SignInWithEmailPassword(email, password)
	if err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	return sa.convertResponse(&authResp.Session), nil
}

func (sa *SupabaseAuth) convertResponse(session *types.Session) *SupabaseAuthResponse {
	response := &SupabaseAuthResponse{
		AccessToken:  session.AccessToken,
		RefreshToken: session.RefreshToken,
		TokenType:    session.TokenType,
		ExpiresIn:    session.ExpiresIn,
	}

	if session.User.ID != (types.User{}).ID {
		user := session.User
		response.User = &SupabaseUser{
			ID:               user.ID.String(),
			Email:            user.Email,
			CreatedAt:        user.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:        user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			UserMetadata:     user.UserMetadata,
			AppMetadata:      user.AppMetadata,
			PhoneConfirmedAt: formatTimeToString(user.PhoneConfirmedAt),
		}
		if user.EmailConfirmedAt != nil {
			confirmedAt := user.EmailConfirmedAt.Format("2006-01-02T15:04:05Z")
			response.User.EmailConfirmedAt = &confirmedAt
		}
		if user.LastSignInAt != nil {
			lastSignIn := user.LastSignInAt.Format("2006-01-02T15:04:05Z")
			response.User.LastSignInAt = &lastSignIn
		}
	}

	return response
}

// SupabaseAuthResponse represents response from Supabase Auth API
type SupabaseAuthResponse struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresIn    int           `json:"expires_in"`
	TokenType    string        `json:"token_type"`
	User         *SupabaseUser `json:"user"`
}

// SupabaseUser represents user data from Supabase Auth API
type SupabaseUser struct {
	ID               string                 `json:"id"`
	Email            string                 `json:"email"`
	CreatedAt        string                 `json:"created_at"`
	UpdatedAt        string                 `json:"updated_at"`
	UserMetadata     map[string]interface{} `json:"user_metadata"`
	AppMetadata      map[string]interface{} `json:"app_metadata"`
	EmailConfirmedAt *string                `json:"email_confirmed_at"`
	PhoneConfirmedAt *string                `json:"phone_confirmed_at"`
	LastSignInAt     *string                `json:"last_sign_in_at"`
}

// ExtractUserID extracts user ID from JWT claims
func ExtractUserID(claims *Claims) string {
	return claims.Subject
}

// ExtractEmail extracts email from JWT claims
func ExtractEmail(claims *Claims) string {
	return claims.Email
}

// Helper function to format time.Time pointer to string pointer
func formatTimeToString(t *time.Time) *string {
	if t == nil {
		return nil
	}
	formatted := t.Format("2006-01-02T15:04:05Z")
	return &formatted
}

// ExchangeCodeForToken exchanges an OAuth authorization code for a session
func (sa *SupabaseAuth) ExchangeCodeForToken(code, codeVerifier string) (*SupabaseAuthResponse, error) {
	if sa.supabaseURL == "" {
		return nil, fmt.Errorf("supabase URL not configured")
	}
	if sa.anonKey == "" {
		return nil, fmt.Errorf("supabase anon key not configured")
	}

	endpoint := strings.TrimRight(sa.supabaseURL, "/") + "/auth/v1/token?grant_type=pkce"
	payload := map[string]string{
		"auth_code":     code,
		"code_verifier": codeVerifier,
	}
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return nil, fmt.Errorf("encode exchange payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, buf)
	if err != nil {
		return nil, fmt.Errorf("build exchange request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", sa.anonKey)
	req.Header.Set("Authorization", "Bearer "+sa.anonKey)

	resp, err := sa.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("exchange request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		msg := strings.TrimSpace(string(body))
		if msg == "" {
			msg = resp.Status
		}
		return nil, fmt.Errorf("supabase exchange failed: %s", msg)
	}

	var supaResp struct {
		AccessToken  string                 `json:"access_token"`
		RefreshToken string                 `json:"refresh_token"`
		ExpiresIn    int                    `json:"expires_in"`
		TokenType    string                 `json:"token_type"`
		User         map[string]interface{} `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&supaResp); err != nil {
		return nil, fmt.Errorf("decode exchange response: %w", err)
	}

	response := &SupabaseAuthResponse{
		AccessToken:  supaResp.AccessToken,
		RefreshToken: supaResp.RefreshToken,
		ExpiresIn:    supaResp.ExpiresIn,
		TokenType:    supaResp.TokenType,
	}

	if supaResp.User != nil {
		user := &SupabaseUser{
			ID:           toString(supaResp.User["id"]),
			Email:        toString(supaResp.User["email"]),
			CreatedAt:    toString(supaResp.User["created_at"]),
			UpdatedAt:    toString(supaResp.User["updated_at"]),
			UserMetadata: toMap(supaResp.User["user_metadata"]),
			AppMetadata:  toMap(supaResp.User["app_metadata"]),
		}
		if v, ok := supaResp.User["email_confirmed_at"].(string); ok && v != "" {
			user.EmailConfirmedAt = &v
		}
		if v, ok := supaResp.User["phone_confirmed_at"].(string); ok && v != "" {
			user.PhoneConfirmedAt = &v
		}
		if v, ok := supaResp.User["last_sign_in_at"].(string); ok && v != "" {
			user.LastSignInAt = &v
		}
		response.User = user
	}

	return response, nil
}

func toString(val interface{}) string {
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

func toMap(val interface{}) map[string]interface{} {
	if m, ok := val.(map[string]interface{}); ok {
		return m
	}
	return map[string]interface{}{}
}
