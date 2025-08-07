package config

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	// Parse test flags first
	flag.Parse()

	// Run tests
	os.Exit(m.Run())
}

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

func TestLoadConfig_HTTPSWithFlags(t *testing.T) {
	// Create temporary secret file
	tempDir := t.TempDir()
	secretFile := filepath.Join(tempDir, "secret.key")
	secretContent := "test-secret-key"

	err := os.WriteFile(secretFile, []byte(secretContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test secret file: %v", err)
	}

	// Set required environment variables
	os.Setenv("SERVER_ADDRESS", "localhost:9090")
	os.Setenv("BASE_URL", "http://localhost:9090")
	os.Setenv("FILE_STORAGE_PATH", "test_urls.json")
	os.Setenv("JWT_SECRET_FILE", secretFile)
	os.Setenv("ENABLE_HTTPS", "true")
	os.Setenv("TLS_CERT_FILE", "flag_cert.pem")
	os.Setenv("TLS_KEY_FILE", "flag_key.pem")

	defer func() {
		os.Unsetenv("SERVER_ADDRESS")
		os.Unsetenv("BASE_URL")
		os.Unsetenv("FILE_STORAGE_PATH")
		os.Unsetenv("JWT_SECRET_FILE")
		os.Unsetenv("ENABLE_HTTPS")
		os.Unsetenv("TLS_CERT_FILE")
		os.Unsetenv("TLS_KEY_FILE")
	}()

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	if !config.EnableHTTPS {
		t.Error("Expected EnableHTTPS to be true when -s flag is set")
	}
	if config.CertFile != "flag_cert.pem" {
		t.Errorf("Expected CertFile to be 'flag_cert.pem', got '%s'", config.CertFile)
	}
	if config.KeyFile != "flag_key.pem" {
		t.Errorf("Expected KeyFile to be 'flag_key.pem', got '%s'", config.KeyFile)
	}
}

func TestLoadConfig_JSONConfig(t *testing.T) {
	// Create temporary secret file
	tempDir := t.TempDir()
	secretFile := filepath.Join(tempDir, "secret.key")
	secretContent := "test-secret-key"

	err := os.WriteFile(secretFile, []byte(secretContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test secret file: %v", err)
	}

	// Create temporary JSON config file
	configFile := filepath.Join(tempDir, "config.json")
	configContent := `{
		"server_address": "localhost:9999",
		"base_url": "http://localhost:9999",
		"file_storage_path": "json_urls.json",
		"database_dsn": "postgres://json:pass@localhost/test",
		"enable_https": true,
		"cert_file": "json_cert.pem",
		"key_file": "json_key.pem"
	}`

	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Set minimal required environment variables
	os.Setenv("JWT_SECRET_FILE", secretFile)
	os.Setenv("CONFIG", configFile)

	defer func() {
		os.Unsetenv("JWT_SECRET_FILE")
		os.Unsetenv("CONFIG")
	}()

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	// Check that values from JSON file were loaded
	if config.Address != "localhost:9999" {
		t.Errorf("Expected Address to be 'localhost:9999', got '%s'", config.Address)
	}
	if config.BaseURL != "http://localhost:9999" {
		t.Errorf("Expected BaseURL to be 'http://localhost:9999', got '%s'", config.BaseURL)
	}
	if config.FileStorage != "json_urls.json" {
		t.Errorf("Expected FileStorage to be 'json_urls.json', got '%s'", config.FileStorage)
	}
	if config.DatabaseDSN != "postgres://json:pass@localhost/test" {
		t.Errorf("Expected DatabaseDSN to be 'postgres://json:pass@localhost/test', got '%s'", config.DatabaseDSN)
	}
	if !config.EnableHTTPS {
		t.Error("Expected EnableHTTPS to be true")
	}
	if config.CertFile != "json_cert.pem" {
		t.Errorf("Expected CertFile to be 'json_cert.pem', got '%s'", config.CertFile)
	}
	if config.KeyFile != "json_key.pem" {
		t.Errorf("Expected KeyFile to be 'json_key.pem', got '%s'", config.KeyFile)
	}
}

func TestLoadConfig_DatabaseDSNFlag(t *testing.T) {
	// Create temporary secret file
	tempDir := t.TempDir()
	secretFile := filepath.Join(tempDir, "secret.key")
	secretContent := "test-secret-key"

	err := os.WriteFile(secretFile, []byte(secretContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test secret file: %v", err)
	}

	// Set required environment variables
	os.Setenv("SERVER_ADDRESS", "localhost:9090")
	os.Setenv("BASE_URL", "http://localhost:9090")
	os.Setenv("FILE_STORAGE_PATH", "test_urls.json")
	os.Setenv("JWT_SECRET_FILE", secretFile)
	os.Setenv("DATABASE_DSN", "postgres://flag:pass@localhost/test")

	defer func() {
		os.Unsetenv("SERVER_ADDRESS")
		os.Unsetenv("BASE_URL")
		os.Unsetenv("FILE_STORAGE_PATH")
		os.Unsetenv("JWT_SECRET_FILE")
		os.Unsetenv("DATABASE_DSN")
	}()

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	if config.DatabaseDSN != "postgres://flag:pass@localhost/test" {
		t.Errorf("Expected DatabaseDSN to be 'postgres://flag:pass@localhost/test', got '%s'", config.DatabaseDSN)
	}
}

func TestLoadConfig_TrustedSubnet(t *testing.T) {
	// Create temporary secret file
	tempDir := t.TempDir()
	secretFile := filepath.Join(tempDir, "secret.key")
	secretContent := "test-secret-key"

	err := os.WriteFile(secretFile, []byte(secretContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test secret file: %v", err)
	}

	// Set required environment variables
	os.Setenv("SERVER_ADDRESS", "localhost:9090")
	os.Setenv("BASE_URL", "http://localhost:9090")
	os.Setenv("FILE_STORAGE_PATH", "test_urls.json")
	os.Setenv("JWT_SECRET_FILE", secretFile)
	os.Setenv("TRUSTED_SUBNET", "192.168.1.0/24")

	defer func() {
		os.Unsetenv("SERVER_ADDRESS")
		os.Unsetenv("BASE_URL")
		os.Unsetenv("FILE_STORAGE_PATH")
		os.Unsetenv("JWT_SECRET_FILE")
		os.Unsetenv("TRUSTED_SUBNET")
	}()

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	if config.TrustedSubnet != "192.168.1.0/24" {
		t.Errorf("Expected TrustedSubnet to be '192.168.1.0/24', got '%s'", config.TrustedSubnet)
	}
}

func TestLoadConfig_TrustedSubnetJSON(t *testing.T) {
	// Create temporary secret file
	tempDir := t.TempDir()
	secretFile := filepath.Join(tempDir, "secret.key")
	secretContent := "test-secret-key"

	err := os.WriteFile(secretFile, []byte(secretContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test secret file: %v", err)
	}

	// Create temporary JSON config file with trusted subnet
	configFile := filepath.Join(tempDir, "config.json")
	configContent := `{
		"server_address": "localhost:9999",
		"base_url": "http://localhost:9999",
		"file_storage_path": "json_urls.json",
		"trusted_subnet": "10.0.0.0/8"
	}`

	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Set minimal required environment variables
	os.Setenv("JWT_SECRET_FILE", secretFile)
	os.Setenv("CONFIG", configFile)

	defer func() {
		os.Unsetenv("JWT_SECRET_FILE")
		os.Unsetenv("CONFIG")
	}()

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	if config.TrustedSubnet != "10.0.0.0/8" {
		t.Errorf("Expected TrustedSubnet to be '10.0.0.0/8', got '%s'", config.TrustedSubnet)
	}
}
