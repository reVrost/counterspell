// Package config provides application configuration from environment variables.
package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration.
type Config struct {
	// Multi-tenant mode
	MultiTenant bool

	// Supabase configuration (required when MultiTenant=true)
	SupabaseURL       string
	SupabaseAnonKey   string
	SupabaseJWTSecret string

	// Native backend allowlist (comma-separated commands)
	NativeAllowlist []string

	// Worker pool configuration
	WorkerPoolSize  int
	MaxTasksPerUser int
	UserManagerTTL  time.Duration

	// Sandbox configuration
	SandboxTimeout     time.Duration
	SandboxOutputLimit int64

	// Data directories
	DataDir string
}

// Load loads configuration from environment variables.
func Load() *Config {
	cfg := &Config{
		// Multi-tenant mode (default: false for backward compatibility)
		MultiTenant: getEnvBool("MULTI_TENANT", false),

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
		UserManagerTTL:  getEnvDuration("USER_MANAGER_TTL", 2*time.Hour),

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
	if c.MultiTenant {
		if c.SupabaseURL == "" {
			return &ConfigError{Field: "SUPABASE_URL", Message: "required when MULTI_TENANT=true"}
		}
		if c.SupabaseAnonKey == "" {
			return &ConfigError{Field: "SUPABASE_ANON_KEY", Message: "required when MULTI_TENANT=true"}
		}
		if c.SupabaseJWTSecret == "" {
			return &ConfigError{Field: "SUPABASE_JWT_SECRET", Message: "required when MULTI_TENANT=true"}
		}
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

func getEnvBool(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return defaultVal
	}
	return b
}

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
	for _, allowed := range c.NativeAllowlist {
		if cmd == allowed {
			return true
		}
	}
	return false
}

// DBPath returns the database path for a user.
// In single-player mode, returns path for "default" user.
func (c *Config) DBPath(userID string) string {
	if !c.MultiTenant {
		userID = "default"
	}
	return c.DataDir + "/db/" + userID + ".db"
}

// ReposPath returns the shared repos directory.
func (c *Config) ReposPath() string {
	return c.DataDir + "/repos"
}

// WorkspacesPath returns the workspaces directory for a user.
func (c *Config) WorkspacesPath(userID string) string {
	if !c.MultiTenant {
		userID = "default"
	}
	return c.DataDir + "/workspaces/" + userID
}
