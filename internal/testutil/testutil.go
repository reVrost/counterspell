package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/revrost/counterspell/internal/db"
)

// CreateTempDB creates a temporary database file path and sets up cleanup.
// This helps minimize WAL file creation during tests.
func CreateTempDB(t *testing.T, name string) string {
	tempDir := t.TempDir() // Automatically cleaned up after test
	dbPath := filepath.Join(tempDir, name+".db")

	// Additional cleanup for any WAL files that might be created
	t.Cleanup(func() {
		os.Remove(dbPath + ".wal")
		os.Remove(dbPath + ".shm")
	})

	return dbPath
}

// CreateInMemoryDB creates an in-memory database for tests that don't need persistence.
// This is the preferred option for most tests as it doesn't create any files.
func CreateInMemoryDB(t *testing.T) db.DBTX {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory test database: %v", err)
	}

	t.Cleanup(func() {
		database.Close()
	})

	return database
}

// InMemoryPath returns the in-memory database path string.
// Use this for functions that need a database path but you want in-memory storage.
func InMemoryPath() string {
	return ":memory:"
}
