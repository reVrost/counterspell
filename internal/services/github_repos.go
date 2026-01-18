package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/revrost/code/counterspell/internal/db"
)

// GitHubRepo represents a GitHub repository.
type GitHubRepo struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	Owner         string `json:"owner"`
	Description   string `json:"description"`
	DefaultBranch string `json:"default_branch"`
	Private       bool   `json:"private"`
	Language      string `json:"language"`
	UpdatedAt     string `json:"updated_at"`
	IsFavorite    bool   `json:"is_favorite"`
}

// RepoListParams contains parameters for listing repos.
type RepoListParams struct {
	Page    int
	PerPage int
	Search  string
	Sort    string // "updated", "name", "created"
}

// RepoListResult contains the result of listing repos.
type RepoListResult struct {
	Repos      []GitHubRepo `json:"repos"`
	TotalCount int          `json:"total_count"`
	HasMore    bool         `json:"has_more"`
}

// FetchUserRepos fetches repositories for the authenticated user from GitHub API.
func FetchUserRepos(ctx context.Context, token string, params RepoListParams) (*RepoListResult, error) {
	if params.PerPage <= 0 {
		params.PerPage = 30
	}
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Sort == "" {
		params.Sort = "updated"
	}

	// Build GitHub API URL
	url := fmt.Sprintf(
		"https://api.github.com/user/repos?sort=%s&direction=desc&per_page=%d&page=%d",
		params.Sort, params.PerPage, params.Page,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repos: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", string(body))
	}

	type apiRepo struct {
		ID            int64  `json:"id"`
		Name          string `json:"name"`
		FullName      string `json:"full_name"`
		Description   string `json:"description"`
		DefaultBranch string `json:"default_branch"`
		Private       bool   `json:"private"`
		Language      string `json:"language"`
		UpdatedAt     string `json:"updated_at"`
		Owner         struct {
			Login string `json:"login"`
		} `json:"owner"`
	}

	var repos []apiRepo
	if err := json.Unmarshal(body, &repos); err != nil {
		return nil, fmt.Errorf("failed to parse repos: %w", err)
	}

	result := &RepoListResult{
		Repos: make([]GitHubRepo, 0, len(repos)),
	}

	for _, r := range repos {
		// Filter by search term if provided
		if params.Search != "" {
			searchLower := strings.ToLower(params.Search)
			if !strings.Contains(strings.ToLower(r.Name), searchLower) &&
				!strings.Contains(strings.ToLower(r.FullName), searchLower) {
				continue
			}
		}

		result.Repos = append(result.Repos, GitHubRepo{
			ID:            r.ID,
			Name:          r.Name,
			FullName:      r.FullName,
			Owner:         r.Owner.Login,
			Description:   r.Description,
			DefaultBranch: r.DefaultBranch,
			Private:       r.Private,
			Language:      r.Language,
			UpdatedAt:     r.UpdatedAt,
		})
	}

	// Check for more pages
	linkHeader := resp.Header.Get("Link")
	result.HasMore = strings.Contains(linkHeader, `rel="next"`)
	result.TotalCount = len(result.Repos)

	return result, nil
}

// CacheTTL is how long cached repos are valid before requiring refresh.
const CacheTTL = 1 * time.Hour

// RepoCache handles caching repo metadata in PostgreSQL.
type RepoCache struct {
	database *db.DB
}

// NewRepoCache creates a new repo cache.
func NewRepoCache(database *db.DB) *RepoCache {
	return &RepoCache{database: database}
}

// CacheRepo caches repo metadata.
func (c *RepoCache) CacheRepo(ctx context.Context, userID string, repo GitHubRepo) error {
	_, err := c.database.Pool.Exec(ctx, `
		INSERT INTO repo_cache (id, user_id, owner, name, default_branch, last_fetched_at, is_favorite)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT(user_id, owner, name) DO UPDATE SET
			default_branch = EXCLUDED.default_branch,
			last_fetched_at = EXCLUDED.last_fetched_at
	`, fmt.Sprintf("%d", repo.ID), userID, repo.Owner, repo.Name, repo.DefaultBranch, time.Now(), repo.IsFavorite)
	return err
}

// GetCachedRepos returns cached repos for a user.
func (c *RepoCache) GetCachedRepos(ctx context.Context, userID string) ([]GitHubRepo, error) {
	rows, err := c.database.Pool.Query(ctx, `
		SELECT id, owner, name, default_branch, is_favorite
		FROM repo_cache
		WHERE user_id = $1
		ORDER BY is_favorite DESC, name ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []GitHubRepo
	for rows.Next() {
		var repo GitHubRepo
		var id string
		if err := rows.Scan(&id, &repo.Owner, &repo.Name, &repo.DefaultBranch, &repo.IsFavorite); err != nil {
			continue
		}
		repo.FullName = fmt.Sprintf("%s/%s", repo.Owner, repo.Name)
		repos = append(repos, repo)
	}

	return repos, nil
}

// IsCacheStale checks if the cache needs refreshing for a user.
func (c *RepoCache) IsCacheStale(ctx context.Context, userID string) bool {
	var lastFetched *time.Time
	err := c.database.Pool.QueryRow(ctx, `
		SELECT MAX(last_fetched_at) FROM repo_cache WHERE user_id = $1
	`, userID).Scan(&lastFetched)
	if err != nil || lastFetched == nil {
		return true // No cache or error, consider stale
	}
	return time.Since(*lastFetched) > CacheTTL
}

// CacheCount returns the number of cached repos for a user.
func (c *RepoCache) CacheCount(ctx context.Context, userID string) int {
	var count int
	_ = c.database.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM repo_cache WHERE user_id = $1`, userID).Scan(&count)
	return count
}

// SetFavorite marks a repo as favorite for a user.
func (c *RepoCache) SetFavorite(ctx context.Context, userID, owner, name string, favorite bool) error {
	_, err := c.database.Pool.Exec(ctx, `
		UPDATE repo_cache SET is_favorite = $1 WHERE user_id = $2 AND owner = $3 AND name = $4
	`, favorite, userID, owner, name)
	return err
}

// GetFavorites returns favorite repos for a user.
func (c *RepoCache) GetFavorites(ctx context.Context, userID string) ([]GitHubRepo, error) {
	rows, err := c.database.Pool.Query(ctx, `
		SELECT id, owner, name, default_branch, is_favorite
		FROM repo_cache
		WHERE user_id = $1 AND is_favorite = true
		ORDER BY name ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []GitHubRepo
	for rows.Next() {
		var repo GitHubRepo
		var id string
		if err := rows.Scan(&id, &repo.Owner, &repo.Name, &repo.DefaultBranch, &repo.IsFavorite); err != nil {
			continue
		}
		repo.FullName = fmt.Sprintf("%s/%s", repo.Owner, repo.Name)
		repos = append(repos, repo)
	}

	return repos, nil
}

// SyncReposFromGitHub fetches repos from GitHub and caches them for a user.
func (c *RepoCache) SyncReposFromGitHub(ctx context.Context, userID, token string) error {
	slog.Info("[GITHUB] Syncing repos from GitHub", "user_id", userID)

	// Fetch all pages
	page := 1
	for {
		result, err := FetchUserRepos(ctx, token, RepoListParams{
			Page:    page,
			PerPage: 100,
			Sort:    "updated",
		})
		if err != nil {
			return err
		}

		for _, repo := range result.Repos {
			if err := c.CacheRepo(ctx, userID, repo); err != nil {
				slog.Warn("Failed to cache repo", "repo", repo.FullName, "error", err)
			}
		}

		if !result.HasMore {
			break
		}
		page++
	}

	slog.Info("[GITHUB] Repo sync complete", "user_id", userID)
	return nil
}
