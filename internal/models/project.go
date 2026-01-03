package models

// Project represents a connected GitHub project.
type Project struct {
	ID          string
	GitHubOwner string // org or username
	GitHubRepo  string
	CreatedAt   int64
}

// GitHubConnection represents a connected GitHub account.
type GitHubConnection struct {
	ID           string
	Type         string // "org" or "user"
	Login        string
	AvatarURL    string
	Token        string // encrypted
	Scope        string
	CreatedAt    int64
}
