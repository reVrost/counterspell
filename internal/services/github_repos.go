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
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

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
		return nil, fmt.Errorf("GitHub API failed (%d): %s", resp.StatusCode, string(body))
	}

	type repoResponse struct {
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

	var repos []repoResponse
	if err := json.Unmarshal(body, &repos); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to our type
	result := &RepoListResult{
		Repos: make([]GitHubRepo, 0, len(repos)),
	}

	for _, r := range repos {
		// Filter by search if provided
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

// RepoCache handles caching repo metadata in SQLite.
type RepoCache struct {
	database *db.DB
}

// NewRepoCache creates a new repo cache.
func NewRepoCache(database *db.DB) *RepoCache {
	return &RepoCache{database: database}
}

// CacheRepo caches repo metadata.
func (c *RepoCache) CacheRepo(ctx context.Context, repo GitHubRepo) error {
	_, err := c.database.Exec(`
		INSERT INTO repo_cache (id, owner, name, default_branch, last_fetched_at, is_favorite)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			default_branch = excluded.default_branch,
			last_fetched_at = excluded.last_fetched_at
	`, fmt.Sprintf("%d", repo.ID), repo.Owner, repo.Name, repo.DefaultBranch, time.Now().Unix(), repo.IsFavorite)
	return err
}

// GetCachedRepos returns cached repos.
func (c *RepoCache) GetCachedRepos(ctx context.Context) ([]GitHubRepo, error) {
	rows, err := c.database.Query(`
		SELECT id, owner, name, default_branch, is_favorite
		FROM repo_cache
		ORDER BY is_favorite DESC, name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

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

// IsCacheStale checks if the cache needs refreshing.
func (c *RepoCache) IsCacheStale(ctx context.Context) bool {
	var lastFetched int64
	err := c.database.QueryRow(`
		SELECT COALESCE(MAX(last_fetched_at), 0) FROM repo_cache
	`).Scan(&lastFetched)
	if err != nil || lastFetched == 0 {
		return true // No cache or error, consider stale
	}
	return time.Since(time.Unix(lastFetched, 0)) > CacheTTL
}

// CacheCount returns the number of cached repos.
func (c *RepoCache) CacheCount(ctx context.Context) int {
	var count int
	_ = c.database.QueryRow(`SELECT COUNT(*) FROM repo_cache`).Scan(&count)
	return count
}

// SetFavorite marks a repo as favorite.
func (c *RepoCache) SetFavorite(ctx context.Context, owner, name string, favorite bool) error {
	_, err := c.database.Exec(`
		UPDATE repo_cache SET is_favorite = ? WHERE owner = ? AND name = ?
	`, favorite, owner, name)
	return err
}

// GetFavorites returns favorite repos.
func (c *RepoCache) GetFavorites(ctx context.Context) ([]GitHubRepo, error) {
	rows, err := c.database.Query(`
		SELECT id, owner, name, default_branch, is_favorite
		FROM repo_cache
		WHERE is_favorite = 1
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

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

// SyncReposFromGitHub fetches repos from GitHub and caches them.
func (c *RepoCache) SyncReposFromGitHub(ctx context.Context, token string) error {
	slog.Info("[GITHUB] Syncing repos from GitHub")

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
			if err := c.CacheRepo(ctx, repo); err != nil {
				slog.Warn("Failed to cache repo", "repo", repo.FullName, "error", err)
			}
		}

		if !result.HasMore {
			break
		}
		page++
	}

	slog.Info("[GITHUB] Repo sync complete")
	return nil
}
