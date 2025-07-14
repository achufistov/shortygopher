package testutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/achufistov/shortygopher.git/internal/app/config"
)

// CreateTestConfig creates a test configuration with temporary files and environment variables.
// If secretContent is empty, uses a default test secret.
func CreateTestConfig(t *testing.T, secretContent string) *config.Config {
	t.Helper()

	// Use default secret if none provided
	if secretContent == "" {
		secretContent = "test-secret-key"
	}

	// Create temporary secret file
	tempDir := t.TempDir()
	secretFile := filepath.Join(tempDir, "secret.key")

	err := os.WriteFile(secretFile, []byte(secretContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test secret file: %v", err)
	}

	// Set environment variables
	os.Setenv("SERVER_ADDRESS", "localhost:8080")
	os.Setenv("BASE_URL", "http://localhost:8080")
	os.Setenv("FILE_STORAGE_PATH", "test_urls.json")
	os.Setenv("JWT_SECRET_FILE", secretFile)

	// Clean up environment variables after test
	t.Cleanup(func() {
		os.Unsetenv("SERVER_ADDRESS")
		os.Unsetenv("BASE_URL")
		os.Unsetenv("FILE_STORAGE_PATH")
		os.Unsetenv("JWT_SECRET_FILE")
	})

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	return cfg
}

// CreateTestConfigWithDefaults creates a test configuration with default secret content.
func CreateTestConfigWithDefaults(t *testing.T) *config.Config {
	return CreateTestConfig(t, "")
}
