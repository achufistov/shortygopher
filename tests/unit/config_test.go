package unit_test

import (
	"os"
	"testing"

	"github.com/achufistov/shortygopher.git/internal/app/config"
)

func TestLoadConfig_DefaultValues(t *testing.T) {

	tmpfile, err := os.CreateTemp("", "secret")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte("test-secret")); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	os.Setenv("JWT_SECRET_FILE", tmpfile.Name())
	defer os.Unsetenv("JWT_SECRET_FILE")

	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")
	os.Unsetenv("FILE_STORAGE_PATH")
	os.Unsetenv("DATABASE_DSN")

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Address != "localhost:8080" {
		t.Fatalf("Expected localhost:8080, got %s", cfg.Address)
	}

	if cfg.BaseURL != "http://localhost:8080" {
		t.Fatalf("Expected http://localhost:8080, got %s", cfg.BaseURL)
	}

	if cfg.FileStorage != "urls.json" {
		t.Fatalf("Expected urls.json, got %s", cfg.FileStorage)
	}

	if cfg.DatabaseDSN != "" {
		t.Fatalf("Expected empty string, got %s", cfg.DatabaseDSN)
	}

	if cfg.SecretKey != "test-secret" {
		t.Fatalf("Expected test-secret, got %s", cfg.SecretKey)
	}
}

func TestLoadConfig_EnvironmentVariables(t *testing.T) {

	tmpfile, err := os.CreateTemp("", "secret")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte("env-secret")); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	os.Setenv("SERVER_ADDRESS", "localhost:9090")
	os.Setenv("BASE_URL", "http://localhost:9090")
	os.Setenv("FILE_STORAGE_PATH", "custom.json")
	os.Setenv("DATABASE_DSN", "postgres://user:pass@localhost/db")
	os.Setenv("JWT_SECRET_FILE", tmpfile.Name())

	defer func() {
		os.Unsetenv("SERVER_ADDRESS")
		os.Unsetenv("BASE_URL")
		os.Unsetenv("FILE_STORAGE_PATH")
		os.Unsetenv("DATABASE_DSN")
		os.Unsetenv("JWT_SECRET_FILE")
	}()

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Address != "localhost:9090" {
		t.Fatalf("Expected localhost:9090, got %s", cfg.Address)
	}

	if cfg.BaseURL != "http://localhost:9090" {
		t.Fatalf("Expected http://localhost:9090, got %s", cfg.BaseURL)
	}

	if cfg.FileStorage != "custom.json" {
		t.Fatalf("Expected custom.json, got %s", cfg.FileStorage)
	}

	if cfg.DatabaseDSN != "postgres://user:pass@localhost/db" {
		t.Fatalf("Expected postgres://user:pass@localhost/db, got %s", cfg.DatabaseDSN)
	}

	if cfg.SecretKey != "env-secret" {
		t.Fatalf("Expected env-secret, got %s", cfg.SecretKey)
	}
}

func TestLoadConfig_MissingSecretFile(t *testing.T) {

	os.Setenv("JWT_SECRET_FILE", "/nonexistent/file")
	defer os.Unsetenv("JWT_SECRET_FILE")

	_, err := config.LoadConfig()
	if err == nil {
		t.Fatal("Expected error for missing secret file")
	}
}
