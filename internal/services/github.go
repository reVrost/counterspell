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
	"github.com/revrost/code/counterspell/internal/db"
	"github.com/revrost/code/counterspell/internal/db/sqlc"
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
		return "", fmt.Errorf("github oauth error: %d", resp.StatusCode)
	}

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Error != "" {
		return "", fmt.Errorf("github oauth error: %s", result.Error)
	}

	return result.AccessToken, nil
}

type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
}

func (s *GitHubService) GetUserInfo(ctx context.Context, token string) (*GitHubUser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api error: %d", resp.StatusCode)
	}

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

type GitHubRepo struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
	HTMLURL  string `json:"html_url"`
	CloneURL string `json:"clone_url"`
	Owner    struct {
		Login string `json:"login"`
	} `json:"owner"`
}

func (s *GitHubService) FetchUserRepos(ctx context.Context, token string) ([]GitHubRepo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user/repos?per_page=100&sort=updated", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api error: %d", resp.StatusCode)
	}

	var repos []GitHubRepo
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, err
	}

	return repos, nil
}

func (s *GitHubService) SyncConnection(ctx context.Context, user *GitHubUser, token string) (string, error) {
	existing, err := s.db.Queries.GetGithubConnection(ctx)
	now := time.Now().UnixMilli()

	if err == nil {
		// Update existing
		_, err = s.db.Queries.UpdateGithubConnection(ctx, sqlc.UpdateGithubConnectionParams{
			AccessToken: token,
			Username:    user.Login,
			AvatarUrl:   sql.NullString{String: user.AvatarURL, Valid: user.AvatarURL != ""},
			UpdatedAt:   now,
			ID:          existing.ID,
		})
		if err != nil {
			return "", err
		}
		return existing.ID, nil
	}

	// Create new
	id := uuid.New().String()
	_, err = s.db.Queries.CreateGithubConnection(ctx, sqlc.CreateGithubConnectionParams{
		ID:           id,
		GithubUserID: fmt.Sprintf("%d", user.ID),
		AccessToken:  token,
		Username:     user.Login,
		AvatarUrl:    sql.NullString{String: user.AvatarURL, Valid: user.AvatarURL != ""},
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

func (s *GitHubService) SyncRepos(ctx context.Context, connectionID string, repos []GitHubRepo) error {
	// Simple approach for single user: delete all and re-create
	if err := s.db.Queries.DeleteRepositoriesByConnection(ctx, connectionID); err != nil {
		return err
	}

	now := time.Now().UnixMilli()
	for _, r := range repos {
		_, err := s.db.Queries.CreateRepository(ctx, sqlc.CreateRepositoryParams{
			ID:           uuid.New().String(),
			ConnectionID: connectionID,
			Name:         r.Name,
			FullName:     r.FullName,
			Owner:        r.Owner.Login,
			IsPrivate:    r.Private,
			HtmlUrl:      r.HTMLURL,
			CloneUrl:     r.CloneURL,
			CreatedAt:    now,
			UpdatedAt:    now,
		})
		if err != nil {
			return err
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
