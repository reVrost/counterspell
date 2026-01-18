// Package config provides application configuration from environment variables.
package config

import (
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration.
type Config struct {
	// Database
	DatabaseURL string

	// Supabase configuration (for auth)
	SupabaseURL       string
	SupabaseAnonKey   string
	SupabaseJWTSecret string

	// Native backend allowlist (comma-separated commands)
	NativeAllowlist []string

	// Worker pool configuration
	WorkerPoolSize  int
	MaxTasksPerUser int

	// Sandbox configuration
	SandboxTimeout     time.Duration
	SandboxOutputLimit int64

	// Data directories (for repos and workspaces)
	DataDir string
}

// Load loads configuration from environment variables.
func Load() *Config {
	cfg := &Config{
		// Database
		DatabaseURL: os.Getenv("DATABASE_URL"),

		// Supabase
		SupabaseURL:       os.Getenv("SUPABASE_URL"),
		SupabaseAnonKey:   os.Getenv("SUPABASE_ANON_KEY"),
		SupabaseJWTSecret: os.Getenv("SUPABASE_JWT_SECRET"),

		// Native allowlist
		NativeAllowlist: getEnvStringSlice("NATIVE_ALLOWLIST", []string{
			"git", "ls", "cat", "head", "tail", "grep", "find", "wc", "sort", "uniq",
		}),

		// Worker pool
		WorkerPoolSize:  getEnvInt("WORKER_POOL_SIZE", 20),
		MaxTasksPerUser: getEnvInt("MAX_TASKS_PER_USER", 5),

		// Sandbox
		SandboxTimeout:     getEnvDuration("SANDBOX_TIMEOUT", 10*time.Minute),
		SandboxOutputLimit: getEnvInt64("SANDBOX_OUTPUT_LIMIT", 1048576), // 1MB

		// Data directory
		DataDir: getEnvString("DATA_DIR", "./data"),
	}

	return cfg
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return &ConfigError{Field: "DATABASE_URL", Message: "required"}
	}
	if c.SupabaseJWTSecret == "" {
		return &ConfigError{Field: "SUPABASE_JWT_SECRET", Message: "required for auth"}
	}
	return nil
}

// ConfigError represents a configuration error.
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return e.Field + ": " + e.Message
}

// Helper functions

func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return i
}

func getEnvInt64(key string, defaultVal int64) int64 {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return defaultVal
	}
	return i
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		return defaultVal
	}
	return d
}

func getEnvString(key string, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func getEnvStringSlice(key string, defaultVal []string) []string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	parts := strings.Split(val, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// IsCommandAllowed checks if a command is in the native allowlist.
func (c *Config) IsCommandAllowed(cmd string) bool {
	return slices.Contains(c.NativeAllowlist, cmd)
}

// ReposPath returns the shared repos directory.
func (c *Config) ReposPath() string {
	return c.DataDir + "/repos"
}

// WorkspacesPath returns the workspaces directory for a user.
func (c *Config) WorkspacesPath(userID string) string {
	return c.DataDir + "/workspaces/" + userID
}
