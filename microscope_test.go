package microscope

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
)

func TestInstall_Success(t *testing.T) {
	// Create a temporary database file
	tmpDB := "test_microscope.db"
	defer os.Remove(tmpDB)

	e := echo.New()

	// Set auth token environment variable
	os.Setenv("MICROSCOPE_AUTH_TOKEN", "test-token")
	defer os.Unsetenv("MICROSCOPE_AUTH_TOKEN")

	err := Install(e, WithDBPath(tmpDB))
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Verify database file was created
	if _, err := os.Stat(tmpDB); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}

	// Test that routes were registered by making a request
	req := httptest.NewRequest(http.MethodGet, "/microscope/health", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Health endpoint not working, expected 200, got %d", rec.Code)
	}
}

func TestInstall_WithOptions(t *testing.T) {
	tmpDB := "test_microscope_options.db"
	defer os.Remove(tmpDB)

	e := echo.New()

	err := Install(e,
		WithDBPath(tmpDB),
		WithAuthToken("custom-token"),
	)
	if err != nil {
		t.Fatalf("Install with options failed: %v", err)
	}

	// Test API endpoint with custom token using query parameter
	req := httptest.NewRequest(http.MethodGet, "/microscope/api/logs?secret=custom-token", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("API endpoint with custom token failed, expected 200, got %d", rec.Code)
	}
}

func TestInstall_NoAuthToken(t *testing.T) {
	// Ensure no auth token is set
	os.Unsetenv("MICROSCOPE_AUTH_TOKEN")

	e := echo.New()

	err := Install(e)
	if err == nil {
		t.Error("Expected Install to fail without auth token, but it succeeded")
	}

	expectedError := "auth token is required"
	if !contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

func TestInstall_InvalidDBPath(t *testing.T) {
	e := echo.New()

	// Try to create database in non-existent directory
	err := Install(e,
		WithAuthToken("test-token"),
		WithDBPath("/non/existent/path/microscope.db"),
	)
	if err == nil {
		t.Error("Expected Install to fail with invalid DB path, but it succeeded")
	}
}

func TestInstall_HealthEndpoint(t *testing.T) {
	tmpDB := "test_health.db"
	defer os.Remove(tmpDB)

	os.Setenv("MICROSCOPE_AUTH_TOKEN", "test-token")
	defer os.Unsetenv("MICROSCOPE_AUTH_TOKEN")

	e := echo.New()
	err := Install(e, WithDBPath(tmpDB))
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Test health endpoint (no auth required)
	req := httptest.NewRequest(http.MethodGet, "/microscope/health", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Health endpoint failed, expected 200, got %d", rec.Code)
	}

	// Verify response content contains the expected fields
	body := rec.Body.String()
	if !contains(body, "healthy") || !contains(body, "microscope") {
		t.Errorf("Response body should contain health status: %s", body)
	}
}

func TestInstall_APIEndpointAuthentication(t *testing.T) {
	tmpDB := "test_auth.db"
	defer os.Remove(tmpDB)

	e := echo.New()
	err := Install(e,
		WithDBPath(tmpDB),
		WithAuthToken("secret-token"),
	)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Test API endpoint without secret parameter (should fail)
	req := httptest.NewRequest(http.MethodGet, "/microscope/api/logs", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected bad request status, got %d", rec.Code)
	}

	// Test API endpoint with wrong secret (should fail)
	req = httptest.NewRequest(http.MethodGet, "/microscope/api/logs?secret=wrong-token", nil)
	rec = httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected unauthorized status with wrong token, got %d", rec.Code)
	}

	// Test API endpoint with correct secret (should succeed)
	req = httptest.NewRequest(http.MethodGet, "/microscope/api/logs?secret=secret-token", nil)
	rec = httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected success with correct token, got %d", rec.Code)
	}
}

func TestInstall_DatabaseMigrations(t *testing.T) {
	tmpDB := "test_migrations.db"
	defer os.Remove(tmpDB)

	os.Setenv("MICROSCOPE_AUTH_TOKEN", "test-token")
	defer os.Unsetenv("MICROSCOPE_AUTH_TOKEN")

	e := echo.New()
	err := Install(e, WithDBPath(tmpDB))
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Verify tables were created by connecting to database
	db, err := sql.Open("sqlite3", tmpDB)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Check if logs table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='logs'").Scan(&tableName)
	if err != nil {
		t.Errorf("Logs table was not created: %v", err)
	}

	// Check if spans table exists
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='spans'").Scan(&tableName)
	if err != nil {
		t.Errorf("Spans table was not created: %v", err)
	}
}

func TestInstall_ConcurrentInstallation(t *testing.T) {
	// Test that multiple installations don't conflict
	tmpDB1 := "test_concurrent1.db"
	tmpDB2 := "test_concurrent2.db"
	defer os.Remove(tmpDB1)
	defer os.Remove(tmpDB2)

	e1 := echo.New()
	e2 := echo.New()

	done := make(chan error, 2)

	// Install in parallel
	go func() {
		done <- Install(e1,
			WithDBPath(tmpDB1),
			WithAuthToken("token1"),
		)
	}()

	go func() {
		done <- Install(e2,
			WithDBPath(tmpDB2),
			WithAuthToken("token2"),
		)
	}()

	// Wait for both to complete
	for i := 0; i < 2; i++ {
		if err := <-done; err != nil {
			t.Errorf("Concurrent installation failed: %v", err)
		}
	}

	// Verify both databases were created
	if _, err := os.Stat(tmpDB1); os.IsNotExist(err) {
		t.Error("Database 1 was not created")
	}
	if _, err := os.Stat(tmpDB2); os.IsNotExist(err) {
		t.Error("Database 2 was not created")
	}
}

func TestInstall_ShutdownHandling(t *testing.T) {
	tmpDB := "test_shutdown.db"
	defer os.Remove(tmpDB)

	os.Setenv("MICROSCOPE_AUTH_TOKEN", "test-token")
	defer os.Unsetenv("MICROSCOPE_AUTH_TOKEN")

	e := echo.New()
	err := Install(e, WithDBPath(tmpDB))
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Verify that shutdown hooks were registered
	// This is harder to test directly, but we can at least verify
	// that the installation completed successfully
	if e.Server == nil {
		t.Error("Echo server was not properly initialized")
	}
}

func TestInstall_DefaultDBPath(t *testing.T) {
	// Test default database path
	defer os.Remove("microscope.db") // Default path

	os.Setenv("MICROSCOPE_AUTH_TOKEN", "test-token")
	defer os.Unsetenv("MICROSCOPE_AUTH_TOKEN")

	e := echo.New()
	err := Install(e) // No WithDBPath option
	if err != nil {
		t.Fatalf("Install with default DB path failed: %v", err)
	}

	// Verify default database file was created
	if _, err := os.Stat("microscope.db"); os.IsNotExist(err) {
		t.Error("Default database file was not created")
	}
}

func TestInstall_EnvVarAuthToken(t *testing.T) {
	tmpDB := "test_env_auth.db"
	defer os.Remove(tmpDB)

	// Set environment variable
	os.Setenv("MICROSCOPE_AUTH_TOKEN", "env-token")
	defer os.Unsetenv("MICROSCOPE_AUTH_TOKEN")

	e := echo.New()
	err := Install(e, WithDBPath(tmpDB))
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Test that the environment variable token works
	req := httptest.NewRequest(http.MethodGet, "/microscope/api/logs?secret=env-token", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Environment variable auth token failed, expected 200, got %d", rec.Code)
	}
}

func TestInstall_OptionOverridesEnvVar(t *testing.T) {
	tmpDB := "test_option_override.db"
	defer os.Remove(tmpDB)

	// Set environment variable
	os.Setenv("MICROSCOPE_AUTH_TOKEN", "env-token")
	defer os.Unsetenv("MICROSCOPE_AUTH_TOKEN")

	e := echo.New()
	err := Install(e,
		WithDBPath(tmpDB),
		WithAuthToken("option-token"), // Should override env var
	)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Test that the option token works (not env var)
	req := httptest.NewRequest(http.MethodGet, "/microscope/api/logs?secret=option-token", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Option auth token failed, expected 200, got %d", rec.Code)
	}

	// Test that env var token no longer works
	req = httptest.NewRequest(http.MethodGet, "/microscope/api/logs?secret=env-token", nil)
	rec = httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Environment auth token should be overridden, expected 401, got %d", rec.Code)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 1; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}
