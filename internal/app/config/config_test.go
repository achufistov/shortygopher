package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_Success(t *testing.T) {
	// Create temporary secret file
	tempDir := t.TempDir()
	secretFile := filepath.Join(tempDir, "secret.key")
	secretContent := "test-secret-key"

	err := os.WriteFile(secretFile, []byte(secretContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test secret file: %v", err)
	}

	// Set environment variables
	os.Setenv("SERVER_ADDRESS", "localhost:9090")
	os.Setenv("BASE_URL", "http://localhost:9090")
	os.Setenv("FILE_STORAGE_PATH", "test_urls.json")
	os.Setenv("DATABASE_DSN", "postgres://user:pass@localhost/test")
	os.Setenv("JWT_SECRET_FILE", secretFile)

	defer func() {
		os.Unsetenv("SERVER_ADDRESS")
		os.Unsetenv("BASE_URL")
		os.Unsetenv("FILE_STORAGE_PATH")
		os.Unsetenv("DATABASE_DSN")
		os.Unsetenv("JWT_SECRET_FILE")
	}()

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	if config.Address != "localhost:9090" {
		t.Errorf("Expected Address to be 'localhost:9090', got '%s'", config.Address)
	}
	if config.BaseURL != "http://localhost:9090" {
		t.Errorf("Expected BaseURL to be 'http://localhost:9090', got '%s'", config.BaseURL)
	}
	if config.FileStorage != "test_urls.json" {
		t.Errorf("Expected FileStorage to be 'test_urls.json', got '%s'", config.FileStorage)
	}
	if config.DatabaseDSN != "postgres://user:pass@localhost/test" {
		t.Errorf("Expected DatabaseDSN to be 'postgres://user:pass@localhost/test', got '%s'", config.DatabaseDSN)
	}
	if config.SecretKey != secretContent {
		t.Errorf("Expected SecretKey to be '%s', got '%s'", secretContent, config.SecretKey)
	}
}

func TestLoadConfig_MissingSecretFile(t *testing.T) {
	// Set environment variables but point to non-existent secret file
	os.Setenv("SERVER_ADDRESS", "localhost:9090")
	os.Setenv("BASE_URL", "http://localhost:9090")
	os.Setenv("FILE_STORAGE_PATH", "test_urls.json")
	os.Setenv("JWT_SECRET_FILE", "/non/existent/file")

	defer func() {
		os.Unsetenv("SERVER_ADDRESS")
		os.Unsetenv("BASE_URL")
		os.Unsetenv("FILE_STORAGE_PATH")
		os.Unsetenv("JWT_SECRET_FILE")
	}()

	_, err := LoadConfig()
	if err == nil {
		t.Error("Expected LoadConfig() to fail with missing secret file, but it succeeded")
	}
}

func TestLoadConfig_SecretKeyTrimming(t *testing.T) {
	// Create temporary secret file with whitespace
	tempDir := t.TempDir()
	secretFile := filepath.Join(tempDir, "secret.key")
	secretContent := "  test-secret-key  \n"
	expectedSecret := "test-secret-key"

	err := os.WriteFile(secretFile, []byte(secretContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test secret file: %v", err)
	}

	// Set environment variables
	os.Setenv("SERVER_ADDRESS", "localhost:8080")
	os.Setenv("BASE_URL", "http://localhost:8080")
	os.Setenv("FILE_STORAGE_PATH", "urls.json")
	os.Setenv("JWT_SECRET_FILE", secretFile)

	defer func() {
		os.Unsetenv("SERVER_ADDRESS")
		os.Unsetenv("BASE_URL")
		os.Unsetenv("FILE_STORAGE_PATH")
		os.Unsetenv("JWT_SECRET_FILE")
	}()

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	if config.SecretKey != expectedSecret {
		t.Errorf("Expected SecretKey to be '%s', got '%s'", expectedSecret, config.SecretKey)
	}
}

func TestLoadConfig_EmptyDatabaseDSN(t *testing.T) {
	// Create temporary secret file
	tempDir := t.TempDir()
	secretFile := filepath.Join(tempDir, "secret.key")
	secretContent := "test-secret-key"

	err := os.WriteFile(secretFile, []byte(secretContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test secret file: %v", err)
	}

	// Set environment variables without DATABASE_DSN
	os.Setenv("SERVER_ADDRESS", "localhost:8080")
	os.Setenv("BASE_URL", "http://localhost:8080")
	os.Setenv("FILE_STORAGE_PATH", "urls.json")
	os.Setenv("JWT_SECRET_FILE", secretFile)

	defer func() {
		os.Unsetenv("SERVER_ADDRESS")
		os.Unsetenv("BASE_URL")
		os.Unsetenv("FILE_STORAGE_PATH")
		os.Unsetenv("JWT_SECRET_FILE")
	}()

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	if config.DatabaseDSN != "" {
		t.Errorf("Expected DatabaseDSN to be empty, got '%s'", config.DatabaseDSN)
	}
}
