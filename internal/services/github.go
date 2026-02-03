package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/revrost/counterspell/internal/db"
	"github.com/revrost/counterspell/internal/db/sqlc"
)

type GitHubService struct {
	db           *db.DB
	clientID     string
	clientSecret string
}

func NewGitHubService(database *db.DB, clientID, clientSecret string) *GitHubService {
	return &GitHubService{
		db:           database,
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

func (s *GitHubService) ExchangeCode(ctx context.Context, code string) (string, error) {
	data := url.Values{}
	data.Set("client_id", s.clientID)
	data.Set("client_secret", s.clientSecret)
	data.Set("code", code)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to exchange code: %s", resp.Status)
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.AccessToken, nil
}

type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
}

func (s *GitHubService) GetGitHubUser(ctx context.Context, accessToken string) (*GitHubUser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user: %s", resp.Status)
	}

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *GitHubService) CreateConnection(ctx context.Context, accessToken string) (string, error) {
	// Get user from GitHub
	user, err := s.GetGitHubUser(ctx, accessToken)
	if err != nil {
		return "", err
	}

	// Check if connection already exists (use user.login for now as simple lookup)
	// In practice, you'd use github_user_id
	conn, err := s.db.Queries.GetGithubConnection(ctx)
	if err == sql.ErrNoRows {
		// Create new connection
		id := uuid.New().String()
		now := time.Now().UnixMilli()
		if _, err := s.db.Queries.CreateGithubConnection(ctx, sqlc.CreateGithubConnectionParams{
			ID:           id,
			GithubUserID: fmt.Sprintf("%d", user.ID),
			AccessToken:  accessToken,
			Username:     user.Login,
			AvatarUrl:    sql.NullString{String: user.AvatarURL, Valid: user.AvatarURL != ""},
			CreatedAt:    now,
			UpdatedAt:    now,
		}); err != nil {
			return "", fmt.Errorf("failed to create connection: %w", err)
		}
		return id, nil
	} else if err != nil {
		return "", fmt.Errorf("failed to get connection: %w", err)
	}

	// Update existing connection
	if _, err := s.db.Queries.UpdateGithubConnection(ctx, sqlc.UpdateGithubConnectionParams{
		ID:          conn.ID,
		AccessToken: accessToken,
		Username:    user.Login,
		AvatarUrl:   sql.NullString{String: user.AvatarURL, Valid: user.AvatarURL != ""},
	}); err != nil {
		return "", fmt.Errorf("failed to update connection: %w", err)
	}
	return conn.ID, nil
}

type GitHubRepo struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Owner    struct {
		Login string `json:"login"`
	} `json:"owner"`
	Private  bool   `json:"private"`
	HTMLURL  string `json:"html_url"`
	CloneURL string `json:"clone_url"`
}

func (s *GitHubService) FetchRepos(ctx context.Context, accessToken string) ([]GitHubRepo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user/repos?visibility=all&affiliation=owner,collaborator", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get repos: %s", resp.Status)
	}

	var repos []GitHubRepo
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, err
	}

	return repos, nil
}

func (s *GitHubService) SyncRepos(ctx context.Context, connectionID string) error {
	// Get connection
	conn, err := s.db.Queries.GetGithubConnectionByID(ctx, connectionID)
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}

	// Get repos from GitHub
	repos, err := s.FetchRepos(ctx, conn.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to get repos: %w", err)
	}

	// Sync repos - upsert based on connection_id and full_name
	now := time.Now().UnixMilli()
	for _, r := range repos {
		// Use GitHub repo ID as our database ID for consistency
		repoID := fmt.Sprintf("%d", r.ID)
		if _, err := s.db.Queries.UpsertRepository(ctx, sqlc.UpsertRepositoryParams{
			ID:           repoID,
			ConnectionID: connectionID,
			Name:         r.Name,
			FullName:     r.FullName,
			Owner:        r.Owner.Login,
			IsPrivate:    r.Private,
			HtmlUrl:      r.HTMLURL,
			CloneUrl:     r.CloneURL,
			LocalPath:    sql.NullString{},
			CreatedAt:    now,
			UpdatedAt:    now,
		}); err != nil {
			return fmt.Errorf("failed to upsert repo %s: %w", r.FullName, err)
		}
	}

	return nil
}

func (s *GitHubService) GetRepos(ctx context.Context) ([]sqlc.Repository, error) {
	conn, err := s.db.Queries.GetGithubConnection(ctx)
	if err != nil {
		return nil, err
	}
	return s.db.Queries.ListRepositories(ctx, conn.ID)
}

func (s *GitHubService) GetConnection(ctx context.Context) (sqlc.GithubConnection, error) {
	return s.db.Queries.GetGithubConnection(ctx)
}

// CreatePullRequest creates a GitHub Pull Request.
func (s *GitHubService) CreatePullRequest(ctx context.Context, owner, repo, branch, title, body string) (string, error) {
	// Get connection
	conn, err := s.db.Queries.GetGithubConnection(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get connection: %w", err)
	}

	// Create PR request
	type PRRequest struct {
		Title string `json:"title"`
		Body  string `json:"body"`
		Head  string `json:"head"`
		Base  string `json:"base"`
	}

	prReq := PRRequest{
		Title: title,
		Body:  body,
		Head:  branch,
		Base:  "main",
	}

	reqBody, err := json.Marshal(prReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal PR request: %w", err)
	}

	// Create PR
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", owner, repo)
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(reqBody)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+conn.AccessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("failed to create PR: %s", resp.Status)
	}

	var result struct {
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode PR response: %w", err)
	}

	return result.HTMLURL, nil
}

// GetUserInfo returns GitHub user info for the connected account.
func (s *GitHubService) GetUserInfo(ctx context.Context) (*GitHubUser, error) {
	conn, err := s.db.Queries.GetGithubConnection(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	return s.GetGitHubUser(ctx, conn.AccessToken)
}

// SyncConnection syncs the current connection's user info and repos.
func (s *GitHubService) SyncConnection(ctx context.Context) error {
	conn, err := s.db.Queries.GetGithubConnection(ctx)
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}

	// Sync repos
	return s.SyncRepos(ctx, conn.ID)
}

// FetchUserRepos fetches repos from GitHub for the connected user.
func (s *GitHubService) FetchUserRepos(ctx context.Context) ([]GitHubRepo, error) {
	conn, err := s.db.Queries.GetGithubConnection(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	return s.FetchRepos(ctx, conn.AccessToken)
}
