package services

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/revrost/code/counterspell/internal/db"
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
func (s *GitHubService) SaveConnection(ctx context.Context, connType, login, avatarURL, token, scope string) error {
	conn := models.GitHubConnection{
		ID:        uuid.New().String(),
		Type:      connType,
		Login:     login,
		AvatarURL: avatarURL,
		Token:     token,
		Scope:     scope,
		CreatedAt: time.Now().Unix(),
	}

	query := `INSERT INTO github_connections (id, type, login, avatar_url, token, scope, created_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.ExecContext(ctx, query, conn.ID, conn.Type, conn.Login, conn.AvatarURL, conn.Token, conn.Scope, conn.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to save connection: %w", err)
	}

	return nil
}

// GetActiveConnection retrieves the active GitHub connection.
func (s *GitHubService) GetActiveConnection(ctx context.Context) (*models.GitHubConnection, error) {
	var conn models.GitHubConnection

	query := `SELECT id, type, login, avatar_url, token, scope, created_at
	          FROM github_connections
	          ORDER BY created_at DESC
	          LIMIT 1`

	err := s.db.QueryRowContext(ctx, query).Scan(&conn.ID, &conn.Type, &conn.Login, &conn.AvatarURL, &conn.Token, &conn.Scope, &conn.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &conn, nil
}

// SaveProject saves a project to the database.
func (s *GitHubService) SaveProject(ctx context.Context, owner, repo string) error {
	project := models.Project{
		ID:          uuid.New().String(),
		GitHubOwner: owner,
		GitHubRepo:  repo,
		CreatedAt:   time.Now().Unix(),
	}

	query := `INSERT INTO projects (id, github_owner, github_repo, created_at)
	          VALUES (?, ?, ?, ?)`

	_, err := s.db.ExecContext(ctx, query, project.ID, project.GitHubOwner, project.GitHubRepo, project.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to save project: %w", err)
	}

	return nil
}

// GetProjects retrieves all projects.
func (s *GitHubService) GetProjects(ctx context.Context) ([]models.Project, error) {
	query := `SELECT id, github_owner, github_repo, created_at FROM projects ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Printf("Failed to close rows: %v\n", err)
		}
	}()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		if err := rows.Scan(&p.ID, &p.GitHubOwner, &p.GitHubRepo, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, p)
	}

	return projects, nil
}

// GetRecentProjects retrieves the first 5 recent projects.
func (s *GitHubService) GetRecentProjects(ctx context.Context) ([]models.Project, error) {
	query := `SELECT id, github_owner, github_repo, created_at FROM projects ORDER BY created_at DESC LIMIT 5`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent projects: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Printf("Failed to close rows: %v\n", err)
		}
	}()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		if err := rows.Scan(&p.ID, &p.GitHubOwner, &p.GitHubRepo, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, p)
	}

	return projects, nil
}

// GetProjectByRepo retrieves a project by repository name.
func (s *GitHubService) GetProjectByRepo(ctx context.Context, repo string) (*models.Project, error) {
	query := `SELECT id, github_owner, github_repo, created_at FROM projects WHERE github_repo = ? LIMIT 1`

	var p models.Project
	err := s.db.QueryRowContext(ctx, query, repo).Scan(&p.ID, &p.GitHubOwner, &p.GitHubRepo, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project by repo: %w", err)
	}

	return &p, nil
}

// DeleteConnection deletes the active GitHub connection.
func (s *GitHubService) DeleteConnection(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM github_connections`)
	if err != nil {
		return fmt.Errorf("failed to delete connection: %w", err)
	}
	return nil
}

// DeleteAllProjects deletes all projects.
func (s *GitHubService) DeleteAllProjects(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM projects`)
	if err != nil {
		return fmt.Errorf("failed to delete projects: %w", err)
	}
	return nil
}

// FetchAndSaveRepositories fetches repositories from GitHub and saves them to database.
func (s *GitHubService) FetchAndSaveRepositories(ctx context.Context, conn *models.GitHubConnection) error {
	type repoResponse struct {
		Name         string `json:"name"`
		FullName     string `json:"full_name"`
		Owner        struct {
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
			var exists bool
			checkQuery := `SELECT EXISTS(SELECT 1 FROM projects WHERE github_owner = ? AND github_repo = ?)`
			err := s.db.QueryRowContext(ctx, checkQuery, owner, repoName).Scan(&exists)
			if err != nil && err != sql.ErrNoRows {
				return fmt.Errorf("failed to check existing project: %w", err)
			}

			if !exists {
				if err := s.SaveProject(ctx, owner, repoName); err != nil {
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
					url = strings.Trim(strings.Trim(parts[0], "<>"), " ")
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
