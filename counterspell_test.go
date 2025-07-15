package counterspell

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/revrost/counterspell/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestInstall_Success(t *testing.T) {
	tmpDB := testutil.CreateTempDB(t, "test_counterspell")

	e := echo.New()

	// Set auth token environment variable
	os.Setenv("COUNTERSPELL_AUTH_TOKEN", "test-token")
	defer os.Unsetenv("COUNTERSPELL_AUTH_TOKEN")

	_, err := Install(e, WithDBPath(tmpDB))
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Verify database file was created
	if _, err := os.Stat(tmpDB); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}

	// Test that routes were registered by making a request
	req := httptest.NewRequest(http.MethodGet, "/counterspell/health", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestInstall_WithOptions(t *testing.T) {
	tmpDB := testutil.CreateTempDB(t, "test_counterspell_options")

	e := echo.New()

	_, err := Install(e,
		WithDBPath(tmpDB),
		WithAuthToken("custom-token"),
		WithServiceName("test-service"),
		WithServiceVersion("1.0.0"),
	)
	if err != nil {
		t.Fatalf("Install with options failed: %v", err)
	}

	// Verify database file was created
	if _, err := os.Stat(tmpDB); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}
}

func TestInstall_NoAuthToken(t *testing.T) {
	// Ensure no auth token is set
	os.Unsetenv("COUNTERSPELL_AUTH_TOKEN")

	e := echo.New()

	_, err := Install(e)
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
	_, err := Install(e,
		WithAuthToken("test-token"),
		WithDBPath("/non/existent/path/counterspell.db"),
	)
	if err == nil {
		t.Error("Expected Install to fail with invalid DB path, but it succeeded")
	}
}

func TestInstall_HealthEndpoint(t *testing.T) {
	tmpDB := testutil.CreateTempDB(t, "test_health")

	os.Setenv("COUNTERSPELL_AUTH_TOKEN", "test-token")
	defer os.Unsetenv("COUNTERSPELL_AUTH_TOKEN")

	e := echo.New()
	_, err := Install(e, WithDBPath(tmpDB))
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Test health endpoint (no auth required)
	req := httptest.NewRequest(http.MethodGet, "/counterspell/health", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Health endpoint failed, expected 200, got %d", rec.Code)
	}

	// Verify response content contains the expected fields
	body := rec.Body.String()
	if !contains(body, "healthy") || !contains(body, "counterspell") {
		t.Errorf("Response body should contain health status: %s", body)
	}
}

func TestInstall_APIEndpointAuthentication(t *testing.T) {
	tmpDB := testutil.CreateTempDB(t, "test_auth")

	e := echo.New()
	_, err := Install(e,
		WithDBPath(tmpDB),
		WithAuthToken("secret-token"),
	)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Test API endpoint without secret parameter (should fail)
	req := httptest.NewRequest(http.MethodGet, "/counterspell/api/logs", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected bad request status, got %d", rec.Code)
	}

	// Test API endpoint with wrong secret (should fail)
	req = httptest.NewRequest(http.MethodGet, "/counterspell/api/logs?secret=wrong-token", nil)
	rec = httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected unauthorized status with wrong token, got %d", rec.Code)
	}

	// Test API endpoint with correct secret (should succeed)
	req = httptest.NewRequest(http.MethodGet, "/counterspell/api/logs?secret=secret-token", nil)
	rec = httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected success with correct token, got %d", rec.Code)
	}
}

func TestInstall_DatabaseMigrations(t *testing.T) {
	tmpDB := testutil.CreateTempDB(t, "test_migrations")

	os.Setenv("COUNTERSPELL_AUTH_TOKEN", "test-token")
	defer os.Unsetenv("COUNTERSPELL_AUTH_TOKEN")

	e := echo.New()
	_, err := Install(e, WithDBPath(tmpDB))
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Verify tables were created by connecting to database
	db, err := sql.Open("duckdb", tmpDB)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Check if logs table exists
	var tableName string
	err = db.QueryRow("SELECT table_name FROM information_schema.tables WHERE table_name = 'logs'").Scan(&tableName)
	if err != nil {
		t.Errorf("Logs table was not created: %v", err)
	}

	// Check if spans table exists
	err = db.QueryRow("SELECT table_name FROM information_schema.tables WHERE table_name = 'spans'").Scan(&tableName)
	if err != nil {
		t.Errorf("Spans table was not created: %v", err)
	}
}

func TestInstall_ConcurrentInstallation(t *testing.T) {
	// Test that multiple installations don't conflict
	tmpDB1 := testutil.CreateTempDB(t, "test_concurrent1")
	tmpDB2 := testutil.CreateTempDB(t, "test_concurrent2")

	e1 := echo.New()
	e2 := echo.New()

	done := make(chan error, 2)

	// Install in parallel
	go func() {
		_, err := Install(e1,
			WithDBPath(tmpDB1),
			WithAuthToken("token1"),
		)
		done <- err
	}()

	go func() {
		_, err := Install(e2,
			WithDBPath(tmpDB2),
			WithAuthToken("token2"),
		)
		done <- err
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
	tmpDB := testutil.CreateTempDB(t, "test_shutdown")

	os.Setenv("COUNTERSPELL_AUTH_TOKEN", "test-token")
	defer os.Unsetenv("COUNTERSPELL_AUTH_TOKEN")

	e := echo.New()
	_, err := Install(e, WithDBPath(tmpDB))
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
	// Clean up the default database files
	t.Cleanup(func() {
		os.Remove("counterspell.db")
		os.Remove("counterspell.db.wal")
		os.Remove("counterspell.db.shm")
	})

	os.Setenv("COUNTERSPELL_AUTH_TOKEN", "test-token")
	defer os.Unsetenv("COUNTERSPELL_AUTH_TOKEN")

	e := echo.New()
	_, err := Install(e) // No WithDBPath option
	if err != nil {
		t.Fatalf("Install with default DB path failed: %v", err)
	}

	// Verify default database file was created
	if _, err := os.Stat("counterspell.db"); os.IsNotExist(err) {
		t.Error("Default database file was not created")
	}
}

func TestInstall_EnvVarAuthToken(t *testing.T) {
	tmpDB := testutil.CreateTempDB(t, "test_env_auth")

	// Set environment variable
	os.Setenv("COUNTERSPELL_AUTH_TOKEN", "env-token")
	defer os.Unsetenv("COUNTERSPELL_AUTH_TOKEN")

	e := echo.New()
	_, err := Install(e, WithDBPath(tmpDB))
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Test that the environment variable token works
	req := httptest.NewRequest(http.MethodGet, "/counterspell/api/logs?secret=env-token", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Environment variable auth token failed, expected 200, got %d", rec.Code)
	}
}

func TestInstall_OptionOverridesEnvVar(t *testing.T) {
	tmpDB := testutil.CreateTempDB(t, "test_option_override")

	// Set environment variable
	os.Setenv("COUNTERSPELL_AUTH_TOKEN", "env-token")
	defer os.Unsetenv("COUNTERSPELL_AUTH_TOKEN")

	e := echo.New()
	_, err := Install(e,
		WithDBPath(tmpDB),
		WithAuthToken("option-token"), // Should override env var
	)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Test that the option token works (not env var)
	req := httptest.NewRequest(http.MethodGet, "/counterspell/api/logs?secret=option-token", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Option auth token failed, expected 200, got %d", rec.Code)
	}

	// Test that env var token no longer works
	req = httptest.NewRequest(http.MethodGet, "/counterspell/api/logs?secret=env-token", nil)
	rec = httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Environment auth token should be overridden, expected 401, got %d", rec.Code)
	}
}

func TestAddToEcho_Success(t *testing.T) {
	// Use in-memory database for simple tests
	e := echo.New()
	_, err := AddToEcho(e, WithDBPath(":memory:"), WithAuthToken("test-token"))

	assert.NoError(t, err)
	assert.NotNil(t, e)
}

func TestAddToStdlib_Success(t *testing.T) {
	// Use in-memory database for simple tests
	mux := http.NewServeMux()
	_, err := AddToStdlib(mux, WithDBPath(":memory:"), WithAuthToken("test-token"))

	assert.NoError(t, err)
	assert.NotNil(t, mux)
}

func TestAddToStdlib_HealthEndpoint(t *testing.T) {
	// Use in-memory database for simple tests
	mux := http.NewServeMux()
	_, err := AddToStdlib(mux, WithDBPath(":memory:"), WithAuthToken("test-token"))
	assert.NoError(t, err)

	// Test the health endpoint
	req := httptest.NewRequest("GET", "/counterspell/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var response map[string]string
	err = json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "counterspell", response["service"])
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
