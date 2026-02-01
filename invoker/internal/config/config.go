package config

import (
	"os"
)

// Config holds application configuration
type Config struct {
	// Supabase
	SupabaseURL         string
	SupabaseAnonKey     string
	SupabaseServiceRole string

	// Server
	Port        string
	AppVersion  string
	Environment string

	// Fly.io
	FlyAPIToken    string
	FlyOrg         string
	FlyRegion      string
	FlyAppName     string
	FlyDockerImage string

	// Database
	DatabaseURL string

	// Security
	JWTSecret string

	// Device auth
	DeviceVerificationURL string

	// Cloudflare tunnel provisioning
	CloudflareAccountID string
	CloudflareAPIToken  string
	CloudflareZoneName  string
	CloudflareZoneID    string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		SupabaseURL:           os.Getenv("SUPABASE_URL"),
		SupabaseAnonKey:       os.Getenv("SUPABASE_ANON_KEY"),
		SupabaseServiceRole:   os.Getenv("SUPABASE_SERVICE_ROLE_KEY"),
		Port:                  getEnv("PORT", "8079"),
		AppVersion:            getEnv("APP_VERSION", "0.1.0"),
		Environment:           getEnv("ENVIRONMENT", "development"),
		FlyAPIToken:           os.Getenv("FLY_API_TOKEN"),
		FlyOrg:                os.Getenv("FLY_ORG"),
		FlyRegion:             getEnv("FLY_REGION", "iad"),
		FlyAppName:            getEnv("FLY_APP_NAME", "counterspell-data-plane"),
		FlyDockerImage:        getEnv("FLY_DOCKER_IMAGE", "registry.fly.io/counterspell-data-plane:latest"),
		DatabaseURL:           os.Getenv("DATABASE_URL"),
		JWTSecret:             os.Getenv("SUPABASE_JWT_SECRET"),
		DeviceVerificationURL: getEnv("DEVICE_VERIFICATION_URL", "https://counterspell.io/device"),
		CloudflareAccountID:   os.Getenv("CLOUDFLARE_ACCOUNT_ID"),
		CloudflareAPIToken:    os.Getenv("CLOUDFLARE_API_TOKEN"),
		CloudflareZoneName:    getEnv("CLOUDFLARE_ZONE_NAME", "counterspell.app"),
		CloudflareZoneID:      os.Getenv("CLOUDFLARE_ZONE_ID"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
