package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/revrost/counterspell/internal/db"
	"github.com/revrost/counterspell/internal/db/sqlc"
)

const (
	openAIConnectorID       = "openai"
	openAIAuthClientID      = "app_EMoamEEZ73f0CkXaXp7hrann"
	openAIAuthURL           = "https://auth.openai.com/oauth/authorize"
	openAIAuthTokenURL      = "https://auth.openai.com/oauth/token"
	openAIAuthScope         = "openid profile email offline_access"
	openAIAuthAttemptTTL    = 10 * time.Minute
	openAIAuthOriginator    = "counterspell"
	openAIJWTAuthClaimKey   = "https://api.openai.com/auth"
	openAIAccountIDClaimKey = "chatgpt_account_id"
)

// OpenAIConnectorStatus exposes machine-local connector status to the UI.
type OpenAIConnectorStatus struct {
	Connected      bool   `json:"connected"`
	AccountID      string `json:"account_id,omitempty"`
	ExpiresAt      int64  `json:"expires_at,omitempty"`
	ConnectedAt    int64  `json:"connected_at,omitempty"`
	TokenExpired   bool   `json:"token_expired"`
	NeedsReconnect bool   `json:"needs_reconnect"`
}

type openAIAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type OpenAIConnectorService struct {
	db         *db.DB
	httpClient *http.Client
	clientID   string
	authorize  string
	tokenURL   string
}

func NewOpenAIConnectorService(database *db.DB) *OpenAIConnectorService {
	return &OpenAIConnectorService{
		db:         database,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		clientID:   openAIAuthClientID,
		authorize:  openAIAuthURL,
		tokenURL:   openAIAuthTokenURL,
	}
}

func (s *OpenAIConnectorService) StartConnect(ctx context.Context, redirectURI string) (string, error) {
	verifier, err := generatePKCEVerifier()
	if err != nil {
		return "", fmt.Errorf("generate code verifier: %w", err)
	}

	challenge := generatePKCEChallenge(verifier)
	state := uuid.NewString()
	now := time.Now().UnixMilli()

	if err := s.db.Queries.CreateConnectorOAuthAttempt(ctx, sqlc.CreateConnectorOAuthAttemptParams{
		Connector:    openAIConnectorID,
		State:        state,
		CodeVerifier: verifier,
		CreatedAt:    now,
	}); err != nil {
		return "", fmt.Errorf("persist openai connector attempt: %w", err)
	}

	if err := s.db.Queries.CleanupExpiredConnectorOAuthAttempts(ctx, sqlc.CleanupExpiredConnectorOAuthAttemptsParams{
		Connector: openAIConnectorID,
		CreatedAt: now - openAIAuthAttemptTTL.Milliseconds(),
	}); err != nil {
		return "", fmt.Errorf("cleanup expired openai connector attempts: %w", err)
	}

	authURL, err := url.Parse(s.authorize)
	if err != nil {
		return "", fmt.Errorf("parse openai authorize url: %w", err)
	}
	q := authURL.Query()
	q.Set("response_type", "code")
	q.Set("client_id", s.clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("scope", openAIAuthScope)
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	q.Set("state", state)
	q.Set("codex_cli_simplified_flow", "true")
	q.Set("originator", openAIAuthOriginator)
	authURL.RawQuery = q.Encode()

	return authURL.String(), nil
}

func (s *OpenAIConnectorService) CompleteConnect(ctx context.Context, code, state, redirectURI string) error {
	attempt, err := s.db.Queries.GetConnectorOAuthAttempt(ctx, sqlc.GetConnectorOAuthAttemptParams{
		Connector: openAIConnectorID,
		State:     state,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("invalid connector state")
		}
		return fmt.Errorf("load connector attempt: %w", err)
	}

	if time.Since(time.UnixMilli(attempt.CreatedAt)) > openAIAuthAttemptTTL {
		return fmt.Errorf("connector state expired")
	}

	tokens, err := s.exchangeAuthorizationCode(ctx, code, attempt.CodeVerifier, redirectURI)
	if err != nil {
		return err
	}

	accountID, err := accountIDFromAccessToken(tokens.AccessToken)
	if err != nil {
		return err
	}

	now := time.Now().UnixMilli()
	expiresAt := now + (tokens.ExpiresIn * 1000)
	if err := s.db.Queries.UpsertConnectorAuth(ctx, sqlc.UpsertConnectorAuthParams{
		Connector:    openAIConnectorID,
		AccessToken:  sql.NullString{String: tokens.AccessToken, Valid: tokens.AccessToken != ""},
		RefreshToken: sql.NullString{String: tokens.RefreshToken, Valid: tokens.RefreshToken != ""},
		AccountID:    sql.NullString{String: accountID, Valid: accountID != ""},
		MetadataJson: sql.NullString{},
		ExpiresAt:    sql.NullInt64{Int64: expiresAt, Valid: expiresAt > 0},
		ConnectedAt:  now,
		UpdatedAt:    now,
	}); err != nil {
		return fmt.Errorf("persist connector auth: %w", err)
	}

	if err := s.db.Queries.DeleteConnectorOAuthAttempt(ctx, sqlc.DeleteConnectorOAuthAttemptParams{
		Connector: openAIConnectorID,
		State:     state,
	}); err != nil {
		return fmt.Errorf("delete connector attempt: %w", err)
	}

	return nil
}

func (s *OpenAIConnectorService) Disconnect(ctx context.Context) error {
	if err := s.db.Queries.DeleteConnectorAuth(ctx, openAIConnectorID); err != nil {
		return fmt.Errorf("delete connector auth: %w", err)
	}
	return nil
}

func (s *OpenAIConnectorService) Status(ctx context.Context) (*OpenAIConnectorStatus, error) {
	row, err := s.db.Queries.GetConnectorAuth(ctx, openAIConnectorID)
	if err != nil {
		if err == sql.ErrNoRows {
			return &OpenAIConnectorStatus{Connected: false, NeedsReconnect: true}, nil
		}
		return nil, fmt.Errorf("load connector status: %w", err)
	}

	now := time.Now().UnixMilli()
	expiresAt := row.ExpiresAt.Int64
	tokenExpired := row.ExpiresAt.Valid && expiresAt > 0 && now >= expiresAt
	refreshToken := strings.TrimSpace(row.RefreshToken.String)
	accountID := strings.TrimSpace(row.AccountID.String)

	return &OpenAIConnectorStatus{
		Connected:      refreshToken != "" && accountID != "",
		AccountID:      accountID,
		ExpiresAt:      expiresAt,
		ConnectedAt:    row.ConnectedAt,
		TokenExpired:   tokenExpired,
		NeedsReconnect: refreshToken == "" || tokenExpired,
	}, nil
}

func (s *OpenAIConnectorService) exchangeAuthorizationCode(ctx context.Context, code, verifier, redirectURI string) (*openAIAuthTokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", s.clientID)
	form.Set("code", code)
	form.Set("code_verifier", verifier)
	form.Set("redirect_uri", redirectURI)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build openai token exchange request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("exchange openai authorization code: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai token exchange failed: %s", strings.TrimSpace(string(body)))
	}

	var tokenResp openAIAuthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decode openai token exchange response: %w", err)
	}
	if tokenResp.AccessToken == "" || tokenResp.RefreshToken == "" || tokenResp.ExpiresIn <= 0 {
		return nil, fmt.Errorf("openai token exchange returned incomplete credentials")
	}

	return &tokenResp, nil
}

func accountIDFromAccessToken(accessToken string) (string, error) {
	claims := jwt.MapClaims{}
	token, _, err := jwt.NewParser(jwt.WithoutClaimsValidation()).ParseUnverified(accessToken, claims)
	if err != nil {
		return "", fmt.Errorf("parse access token claims: %w", err)
	}
	if token == nil {
		return "", fmt.Errorf("invalid access token")
	}

	authClaimRaw, ok := claims[openAIJWTAuthClaimKey]
	if !ok {
		return "", fmt.Errorf("access token missing %s claim", openAIJWTAuthClaimKey)
	}
	authClaim, ok := authClaimRaw.(map[string]any)
	if !ok {
		return "", fmt.Errorf("access token auth claim has invalid format")
	}
	accountID, _ := authClaim[openAIAccountIDClaimKey].(string)
	accountID = strings.TrimSpace(accountID)
	if accountID == "" {
		return "", fmt.Errorf("access token missing chatgpt account id")
	}
	return accountID, nil
}

func generatePKCEVerifier() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate random bytes for verifier: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func generatePKCEChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
