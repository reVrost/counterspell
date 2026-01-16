package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/revrost/code/counterspell/internal/config"
	"github.com/revrost/code/counterspell/internal/db"
)

// ErrTokenNotFound is returned when no GitHub token is found.
var ErrTokenNotFound = errors.New("GitHub token not found")

// TokenStore handles dual storage of GitHub tokens (SQLite + Supabase vault).
type TokenStore struct {
	cfg         *config.Config
	supabaseURL string
	supabaseKey string
}

// NewTokenStore creates a new token store.
func NewTokenStore(cfg *config.Config) *TokenStore {
	return &TokenStore{
		cfg:         cfg,
		supabaseURL: cfg.SupabaseURL,
		supabaseKey: cfg.SupabaseAnonKey,
	}
}

// GetToken retrieves the GitHub token for a user.
// First tries SQLite, then falls back to Supabase vault if multi-tenant mode.
// If token is fetched from vault, it's written back to SQLite.
func (s *TokenStore) GetToken(ctx context.Context, userID string, database *db.DB) (string, error) {
	// Try SQLite first
	token, err := s.getTokenFromSQLite(ctx, database)
	if err == nil && token != "" {
		slog.Debug("Got token from SQLite", "user_id", userID)
		return token, nil
	}

	// In single-player mode, SQLite is the only source
	if !s.cfg.MultiTenant {
		if err != nil && err != sql.ErrNoRows {
			return "", fmt.Errorf("failed to get token from SQLite: %w", err)
		}
		return "", ErrTokenNotFound
	}

	// Try Supabase vault
	token, err = s.getTokenFromVault(ctx, userID)
	if err != nil {
		slog.Debug("Token not in vault either", "user_id", userID, "error", err)
		return "", ErrTokenNotFound
	}

	// Write back to SQLite for caching
	if token != "" {
		if err := s.saveTokenToSQLite(ctx, database, token); err != nil {
			slog.Warn("Failed to cache token in SQLite", "error", err)
		} else {
			slog.Debug("Cached vault token in SQLite", "user_id", userID)
		}
	}

	return token, nil
}

// SaveToken saves a GitHub token to both SQLite and Supabase vault (if multi-tenant).
func (s *TokenStore) SaveToken(ctx context.Context, userID string, database *db.DB, token string) error {
	// Always save to SQLite
	if err := s.saveTokenToSQLite(ctx, database, token); err != nil {
		return fmt.Errorf("failed to save token to SQLite: %w", err)
	}

	// In multi-tenant mode, also save to vault
	if s.cfg.MultiTenant {
		if err := s.saveTokenToVault(ctx, userID, token); err != nil {
			slog.Warn("Failed to save token to vault", "user_id", userID, "error", err)
			// Don't return error - SQLite save succeeded
		}
	}

	return nil
}

// DeleteToken removes the GitHub token from both stores.
func (s *TokenStore) DeleteToken(ctx context.Context, userID string, database *db.DB) error {
	// Delete from SQLite
	if err := s.deleteTokenFromSQLite(ctx, database); err != nil {
		slog.Warn("Failed to delete token from SQLite", "error", err)
	}

	// In multi-tenant mode, also delete from vault
	if s.cfg.MultiTenant {
		if err := s.deleteTokenFromVault(ctx, userID); err != nil {
			slog.Warn("Failed to delete token from vault", "user_id", userID, "error", err)
		}
	}

	return nil
}

// getTokenFromSQLite retrieves the token from the user's SQLite database.
func (s *TokenStore) getTokenFromSQLite(ctx context.Context, database *db.DB) (string, error) {
	conn, err := database.Queries.GetActiveGitHubConnection(ctx)
	if err != nil {
		return "", err
	}
	return conn.Token, nil
}

// saveTokenToSQLite saves the token to the user's SQLite database.
func (s *TokenStore) saveTokenToSQLite(ctx context.Context, database *db.DB, token string) error {
	// This updates the existing connection's token
	// The actual save is done through GitHubService.SaveConnection
	// This method is for updating existing tokens
	_, err := database.Exec(`
		UPDATE github_connections
		SET token = ?
		WHERE id = (SELECT id FROM github_connections LIMIT 1)
	`, token)
	return err
}

// deleteTokenFromSQLite removes the token from SQLite.
func (s *TokenStore) deleteTokenFromSQLite(ctx context.Context, database *db.DB) error {
	_, err := database.Queries.DeleteAllGitHubConnections(ctx)
	return err
}

// getTokenFromVault retrieves the token from Supabase vault.
func (s *TokenStore) getTokenFromVault(ctx context.Context, userID string) (string, error) {
	if s.supabaseURL == "" {
		return "", errors.New("supabase not configured")
	}

	// Call Supabase vault API
	// The vault stores secrets per user with a service role key
	url := fmt.Sprintf("%s/rest/v1/vault?user_id=eq.%s&select=github_token", s.supabaseURL, userID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("apikey", s.supabaseKey)
	req.Header.Set("Authorization", "Bearer "+s.supabaseKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("vault API failed (%d): %s", resp.StatusCode, string(body))
	}

	var results []struct {
		GithubToken string `json:"github_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return "", err
	}

	if len(results) == 0 || results[0].GithubToken == "" {
		return "", ErrTokenNotFound
	}

	return results[0].GithubToken, nil
}

// saveTokenToVault saves the token to Supabase vault.
func (s *TokenStore) saveTokenToVault(ctx context.Context, userID, token string) error {
	if s.supabaseURL == "" {
		return errors.New("supabase not configured")
	}

	// Upsert into vault table
	url := fmt.Sprintf("%s/rest/v1/vault", s.supabaseURL)

	body := map[string]string{
		"user_id":      userID,
		"github_token": token,
	}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("apikey", s.supabaseKey)
	req.Header.Set("Authorization", "Bearer "+s.supabaseKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "resolution=merge-duplicates")
	req.Body = io.NopCloser(io.NopCloser(io.NopCloser(io.NopCloser(
		http.NoBody,
	))))

	// Use PUT with upsert
	req.Method = "POST"
	req.Body = io.NopCloser(http.NoBody)

	// Re-create with proper body
	req, _ = http.NewRequestWithContext(ctx, "POST", url, io.NopCloser(
		&readCloserWrapper{data: jsonBody},
	))
	req.Header.Set("apikey", s.supabaseKey)
	req.Header.Set("Authorization", "Bearer "+s.supabaseKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "resolution=merge-duplicates")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("vault save failed (%d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// deleteTokenFromVault removes the token from Supabase vault.
func (s *TokenStore) deleteTokenFromVault(ctx context.Context, userID string) error {
	if s.supabaseURL == "" {
		return errors.New("supabase not configured")
	}

	url := fmt.Sprintf("%s/rest/v1/vault?user_id=eq.%s", s.supabaseURL, userID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("apikey", s.supabaseKey)
	req.Header.Set("Authorization", "Bearer "+s.supabaseKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// readCloserWrapper wraps a byte slice as an io.ReadCloser.
type readCloserWrapper struct {
	data []byte
	pos  int
}

func (r *readCloserWrapper) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func (r *readCloserWrapper) Close() error {
	return nil
}
