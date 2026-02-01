package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name     string
		setupEnv func()
		cleanup  func()
		check    func(t *testing.T, cfg *Config)
	}{
		{
			name: "load config with all env vars set",
			setupEnv: func() {
				if err := os.Setenv("SUPABASE_URL", "https://supabase.example.com"); err != nil {
					t.Fatalf("Failed to set SUPABASE_URL: %v", err)
				}
				if err := os.Setenv("SUPABASE_ANON_KEY", "anon-key"); err != nil {
					t.Fatalf("Failed to set SUPABASE_ANON_KEY: %v", err)
				}
				if err := os.Setenv("SUPABASE_SERVICE_ROLE_KEY", "service-role-key"); err != nil {
					t.Fatalf("Failed to set SUPABASE_SERVICE_ROLE_KEY: %v", err)
				}
				if err := os.Setenv("PORT", "9000"); err != nil {
					t.Fatalf("Failed to set PORT: %v", err)
				}
				if err := os.Setenv("APP_VERSION", "1.0.0"); err != nil {
					t.Fatalf("Failed to set APP_VERSION: %v", err)
				}
				if err := os.Setenv("ENVIRONMENT", "production"); err != nil {
					t.Fatalf("Failed to set ENVIRONMENT: %v", err)
				}
				if err := os.Setenv("FLY_API_TOKEN", "fly-token"); err != nil {
					t.Fatalf("Failed to set FLY_API_TOKEN: %v", err)
				}
				if err := os.Setenv("FLY_ORG", "test-org"); err != nil {
					t.Fatalf("Failed to set FLY_ORG: %v", err)
				}
				if err := os.Setenv("FLY_REGION", "ewr"); err != nil {
					t.Fatalf("Failed to set FLY_REGION: %v", err)
				}
				if err := os.Setenv("DATABASE_URL", "postgres://localhost/db"); err != nil {
					t.Fatalf("Failed to set DATABASE_URL: %v", err)
				}
				if err := os.Setenv("SUPABASE_JWT_SECRET", "jwt-secret"); err != nil {
					t.Fatalf("Failed to set SUPABASE_JWT_SECRET: %v", err)
				}
			},
			cleanup: func() {
				_ = os.Unsetenv("SUPABASE_URL")
				_ = os.Unsetenv("SUPABASE_ANON_KEY")
				_ = os.Unsetenv("SUPABASE_SERVICE_ROLE_KEY")
				_ = os.Unsetenv("PORT")
				_ = os.Unsetenv("APP_VERSION")
				_ = os.Unsetenv("ENVIRONMENT")
				_ = os.Unsetenv("FLY_API_TOKEN")
				_ = os.Unsetenv("FLY_ORG")
				_ = os.Unsetenv("FLY_REGION")
				_ = os.Unsetenv("DATABASE_URL")
				_ = os.Unsetenv("SUPABASE_JWT_SECRET")
			},
			check: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "https://supabase.example.com", cfg.SupabaseURL)
				assert.Equal(t, "anon-key", cfg.SupabaseAnonKey)
				assert.Equal(t, "service-role-key", cfg.SupabaseServiceRole)
				assert.Equal(t, "9000", cfg.Port)
				assert.Equal(t, "1.0.0", cfg.AppVersion)
				assert.Equal(t, "production", cfg.Environment)
				assert.Equal(t, "fly-token", cfg.FlyAPIToken)
				assert.Equal(t, "test-org", cfg.FlyOrg)
				assert.Equal(t, "ewr", cfg.FlyRegion)
				assert.Equal(t, "postgres://localhost/db", cfg.DatabaseURL)
				assert.Equal(t, "jwt-secret", cfg.JWTSecret)
			},
		},
		{
			name:     "load config with defaults",
			setupEnv: func() {},
			cleanup:  func() {},
			check: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "", cfg.SupabaseURL)
				assert.Equal(t, "", cfg.SupabaseAnonKey)
				assert.Equal(t, "", cfg.SupabaseServiceRole)
				assert.Equal(t, "8079", cfg.Port)
				assert.Equal(t, "0.1.0", cfg.AppVersion)
				assert.Equal(t, "development", cfg.Environment)
				assert.Equal(t, "", cfg.FlyAPIToken)
				assert.Equal(t, "", cfg.FlyOrg)
				assert.Equal(t, "iad", cfg.FlyRegion)
				assert.Equal(t, "", cfg.DatabaseURL)
				assert.Equal(t, "", cfg.JWTSecret)
			},
		},
		{
			name: "load config with partial env vars",
			setupEnv: func() {
				if err := os.Setenv("PORT", "3000"); err != nil {
					t.Fatalf("Failed to set PORT: %v", err)
				}
				if err := os.Setenv("ENVIRONMENT", "staging"); err != nil {
					t.Fatalf("Failed to set ENVIRONMENT: %v", err)
				}
				if err := os.Setenv("DATABASE_URL", "postgres://localhost/testdb"); err != nil {
					t.Fatalf("Failed to set DATABASE_URL: %v", err)
				}
			},
			cleanup: func() {
				_ = os.Unsetenv("PORT")
				_ = os.Unsetenv("ENVIRONMENT")
				_ = os.Unsetenv("DATABASE_URL")
			},
			check: func(t *testing.T, cfg *Config) {
				assert.Equal(t, "", cfg.SupabaseURL)
				assert.Equal(t, "3000", cfg.Port)
				assert.Equal(t, "0.1.0", cfg.AppVersion)
				assert.Equal(t, "staging", cfg.Environment)
				assert.Equal(t, "", cfg.FlyAPIToken)
				assert.Equal(t, "", cfg.FlyOrg)
				assert.Equal(t, "iad", cfg.FlyRegion)
				assert.Equal(t, "postgres://localhost/testdb", cfg.DatabaseURL)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			defer tt.cleanup()

			cfg := LoadConfig()
			tt.check(t, cfg)
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		defaultValue  string
		setupEnv      func()
		cleanup       func()
		expectedValue string
	}{
		{
			name:         "env var exists",
			key:          "TEST_VAR",
			defaultValue: "default",
			setupEnv: func() {
				if err := os.Setenv("TEST_VAR", "actual"); err != nil {
					t.Fatalf("Failed to set TEST_VAR: %v", err)
				}
			},
			cleanup:       func() { _ = os.Unsetenv("TEST_VAR") },
			expectedValue: "actual",
		},
		{
			name:          "env var does not exist",
			key:           "NON_EXISTENT_VAR",
			defaultValue:  "default",
			setupEnv:      func() {},
			cleanup:       func() {},
			expectedValue: "default",
		},
		{
			name:         "env var is empty",
			key:          "EMPTY_VAR",
			defaultValue: "default",
			setupEnv: func() {
				if err := os.Setenv("EMPTY_VAR", ""); err != nil {
					t.Fatalf("Failed to set EMPTY_VAR: %v", err)
				}
			},
			cleanup:       func() { _ = os.Unsetenv("EMPTY_VAR") },
			expectedValue: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			defer tt.cleanup()

			result := getEnv(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expectedValue, result)
		})
	}
}
