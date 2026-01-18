package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lithammer/shortuuid/v4"
	"github.com/revrost/code/counterspell/internal/db"
	"github.com/revrost/code/counterspell/internal/db/sqlc"
	"github.com/revrost/code/counterspell/internal/models"
)

// GitHubService handles GitHub OAuth API interactions.
type GitHubService struct {
	clientID     string
	clientSecret string
	redirectURI  string
	db           *db.DB
}

// NewGitHubService creates a new GitHub service.
func NewGitHubService(clientID, clientSecret, redirectURI string, db *db.DB) *GitHubService {
	fmt.Printf("GitHub Service initialized:\n")
	fmt.Printf("  Client ID: %s\n", maskString(clientID))
	fmt.Printf("  Client Secret: %s\n", maskString(clientSecret))
	fmt.Printf("  Redirect URI: %s\n", redirectURI)

	return &GitHubService{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		db:           db,
	}
}

// maskString masks sensitive values for logging
func maskString(s string) string {
	if s == "" {
		return "not set"
	}
	if len(s) <= 8 {
		return "***"
	}
	return s[:4] + "..." + s[len(s)-4:]
}

// ExchangeCodeForToken exchanges the OAuth code for an access token.
func (s *GitHubService) ExchangeCodeForToken(ctx context.Context, code string) (string, error) {
	type tokenRequest struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		Code         string `json:"code"`
		RedirectURI  string `json:"redirect_uri"`
	}

	type tokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}

	reqBody := tokenRequest{
		ClientID:     s.clientID,
		ClientSecret: s.clientSecret,
		Code:         code,
		RedirectURI:  s.redirectURI,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://github.com/login/oauth/access_token", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to exchange token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub token exchange failed: %s", string(body))
	}

	var tokenResp tokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("no access token in response")
	}

	return tokenResp.AccessToken, nil
}

// ValidateToken checks if a GitHub token is still valid by making a test API call.
func (s *GitHubService) ValidateToken(ctx context.Context, token string) bool {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return false
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// GetUserInfo fetches user info from GitHub.
func (s *GitHubService) GetUserInfo(ctx context.Context, token string) (login, avatarURL string, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("GitHub API failed: %s", string(body))
	}

	type userResponse struct {
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
	}

	var userResp userResponse
	if err := json.Unmarshal(body, &userResp); err != nil {
		return "", "", fmt.Errorf("failed to parse user response: %w", err)
	}

	return userResp.Login, userResp.AvatarURL, nil
}

// SaveConnection saves a GitHub connection to the database.
func (s *GitHubService) SaveConnection(ctx context.Context, userID, connType, login, avatarURL, token, scope string) error {
	err := s.db.Queries.CreateGitHubConnection(ctx, sqlc.CreateGitHubConnectionParams{
		ID:        shortuuid.New(),
		UserID:    userID,
		Type:      connType,
		Login:     login,
		AvatarUrl: pgtype.Text{String: avatarURL, Valid: avatarURL != ""},
		Token:     token,
		Scope:     pgtype.Text{String: scope, Valid: scope != ""},
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to save connection: %w", err)
	}

	return nil
}

// GetActiveConnection retrieves the active GitHub connection.
func (s *GitHubService) GetActiveConnection(ctx context.Context, userID string) (*models.GitHubConnection, error) {
	fmt.Println("[GITHUB] GetActiveConnection: querying database")
	conn, err := s.db.Queries.GetActiveGitHubConnection(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		fmt.Println("[GITHUB] GetActiveConnection: no connection found")
		return nil, nil
	}
	if err != nil {
		fmt.Printf("[GITHUB] GetActiveConnection: error=%v\n", err)
		return nil, err
	}

	result := &models.GitHubConnection{
		ID:        conn.ID,
		Type:      conn.Type,
		Login:     conn.Login,
		AvatarURL: conn.AvatarUrl.String,
		Token:     conn.Token,
		Scope:     conn.Scope.String,
		CreatedAt: conn.CreatedAt.Time.Unix(),
	}

	fmt.Printf("[GITHUB] GetActiveConnection: found connection login=%s type=%s\n", result.Login, result.Type)
	return result, nil
}

// SaveProject saves a project to the database.
func (s *GitHubService) SaveProject(ctx context.Context, userID, owner, repo string) error {
	_, err := s.db.Queries.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:          shortuuid.New(),
		UserID:      userID,
		GithubOwner: owner,
		GithubRepo:  repo,
		CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	return nil
}

// GetProject retrieves a single project by ID.
func (s *GitHubService) GetProject(ctx context.Context, userID, projectID string) (*models.Project, error) {
	p, err := s.db.Queries.GetProject(ctx, sqlc.GetProjectParams{
		ID:     projectID,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query project: %w", err)
	}
	return &models.Project{
		ID:          p.ID,
		GitHubOwner: p.GithubOwner,
		GitHubRepo:  p.GithubRepo,
		CreatedAt:   p.CreatedAt.Time.Unix(),
	}, nil
}

// GetProjects retrieves all projects.
func (s *GitHubService) GetProjects(ctx context.Context, userID string) ([]models.Project, error) {
	projects, err := s.db.Queries.GetProjects(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}

	result := make([]models.Project, len(projects))
	for i, p := range projects {
		result[i] = models.Project{
			ID:          p.ID,
			GitHubOwner: p.GithubOwner,
			GitHubRepo:  p.GithubRepo,
			CreatedAt:   p.CreatedAt.Time.Unix(),
		}
	}

	return result, nil
}

// GetRecentProjects retrieves the first 5 recent projects.
func (s *GitHubService) GetRecentProjects(ctx context.Context, userID string) ([]models.Project, error) {
	projects, err := s.db.Queries.GetRecentProjects(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent projects: %w", err)
	}

	result := make([]models.Project, len(projects))
	for i, p := range projects {
		result[i] = models.Project{
			ID:          p.ID,
			GitHubOwner: p.GithubOwner,
			GitHubRepo:  p.GithubRepo,
			CreatedAt:   p.CreatedAt.Time.Unix(),
		}
	}

	return result, nil
}

// GetProjectByRepo retrieves a project by repository name.
func (s *GitHubService) GetProjectByRepo(ctx context.Context, userID, repo string) (*models.Project, error) {
	p, err := s.db.Queries.GetProjectByRepo(ctx, sqlc.GetProjectByRepoParams{
		UserID:     userID,
		GithubRepo: repo,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project by repo: %w", err)
	}

	result := &models.Project{
		ID:          p.ID,
		GitHubOwner: p.GithubOwner,
		GitHubRepo:  p.GithubRepo,
		CreatedAt:   p.CreatedAt.Time.Unix(),
	}

	return result, nil
}

// DeleteConnection deletes the active GitHub connection.
func (s *GitHubService) DeleteConnection(ctx context.Context, userID string) error {
	fmt.Println("[GITHUB] DeleteConnection: starting")
	result, err := s.db.Queries.DeleteAllGitHubConnections(ctx, userID)
	if err != nil {
		fmt.Printf("[GITHUB] DeleteConnection: error=%v\n", err)
		return fmt.Errorf("failed to delete connection: %w", err)
	}
	rows := result.RowsAffected()
	fmt.Printf("[GITHUB] DeleteConnection: success, rows_affected=%d\n", rows)
	return nil
}

// DeleteAllProjects deletes all projects.
func (s *GitHubService) DeleteAllProjects(ctx context.Context, userID string) error {
	fmt.Println("[GITHUB] DeleteAllProjects: starting")
	result, err := s.db.Queries.DeleteAllProjects(ctx, userID)
	if err != nil {
		fmt.Printf("[GITHUB] DeleteAllProjects: error=%v\n", err)
		return fmt.Errorf("failed to delete projects: %w", err)
	}
	rows := result.RowsAffected()
	fmt.Printf("[GITHUB] DeleteAllProjects: success, rows_affected=%d\n", rows)
	return nil
}

// FetchAndSaveRepositories fetches repositories from GitHub and saves them to database.
func (s *GitHubService) FetchAndSaveRepositories(ctx context.Context, userID string, conn *models.GitHubConnection) error {
	type repoResponse struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		Owner    struct {
			Login string `json:"login"`
		} `json:"owner"`
	}

	var url string
	if conn.Type == "org" {
		url = "https://api.github.com/orgs/" + conn.Login + "/repos"
	} else {
		url = "https://api.github.com/user/repos"
	}

	fmt.Printf("Fetching from URL: %s\n", url)

	repoCount := 0
	for {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+conn.Token)
		req.Header.Set("Accept", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to fetch repos: %w", err)
		}

		linkHeader := resp.Header.Get("Link")
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		fmt.Printf("Response status: %d\n", resp.StatusCode)

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("GitHub API failed: %s", string(body))
		}

		var repos []repoResponse
		if err := json.Unmarshal(body, &repos); err != nil {
			return fmt.Errorf("failed to parse repos: %w", err)
		}

		fmt.Printf("Found %d repos in this page\n", len(repos))

		// Save each repository
		for _, repo := range repos {
			owner := repo.Owner.Login
			repoName := repo.Name
			fmt.Printf("Checking repo: %s/%s\n", owner, repoName)

			// Check if already exists
			exists, err := s.db.Queries.ProjectExists(ctx, sqlc.ProjectExistsParams{
				UserID:      userID,
				GithubOwner: owner,
				GithubRepo:  repoName,
			})
			if err != nil {
				return fmt.Errorf("failed to check existing project: %w", err)
			}

			if !exists {
				if err := s.SaveProject(ctx, userID, owner, repoName); err != nil {
					// Log error but continue with other repos
					fmt.Printf("Failed to save project %s/%s: %v\n", owner, repoName, err)
				} else {
					fmt.Printf("Saved project %s/%s\n", owner, repoName)
					repoCount++
				}
			} else {
				fmt.Printf("Project %s/%s already exists\n", owner, repoName)
			}
		}

		// Check if there's a next page
		url = ""
		if linkHeader != "" {
			links := strings.Split(linkHeader, ",")
			for _, link := range links {
				if strings.Contains(link, `rel="next"`) {
					parts := strings.Split(link, ";")
					url = strings.TrimSpace(parts[0])
					url = strings.TrimPrefix(url, "<")
					url = strings.TrimSuffix(url, ">")
					break
				}
			}
		}

		if url == "" {
			break
		}
		fmt.Printf("Fetching next page: %s\n", url)
	}

	fmt.Printf("Total repositories saved: %d\n", repoCount)
	return nil
}

// CreatePullRequest creates a GitHub PR for the given branch.
// Returns the PR URL on success.
func (s *GitHubService) CreatePullRequest(ctx context.Context, userID, owner, repo, branch, title, body string) (string, error) {
	conn, err := s.GetActiveConnection(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("no GitHub connection: %w", err)
	}
	if conn == nil {
		return "", fmt.Errorf("no GitHub connection found")
	}

	// Create PR via GitHub API
	prData := map[string]string{
		"title": title,
		"body":  body,
		"head":  branch,
		"base":  "main",
	}

	jsonData, err := json.Marshal(prData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal PR data: %w", err)
	}

	prURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", owner, repo)
	req, err := http.NewRequestWithContext(ctx, "POST", prURL, bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+conn.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to create PR: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		// Check if PR already exists
		if resp.StatusCode == 422 && strings.Contains(string(respBody), "A pull request already exists") {
			return "", fmt.Errorf("PR already exists for this branch")
		}
		return "", fmt.Errorf("GitHub API failed (%d): %s", resp.StatusCode, string(respBody))
	}

	var prResponse struct {
		HTMLURL string `json:"html_url"`
		Number  int    `json:"number"`
	}
	if err := json.Unmarshal(respBody, &prResponse); err != nil {
		return "", fmt.Errorf("failed to parse PR response: %w", err)
	}

	return prResponse.HTMLURL, nil
}
