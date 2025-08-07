package shortygopher_test

import (
	"fmt"
	"log"
	"os"

	"github.com/achufistov/shortygopher.git/internal/app/config"
)

// ExampleLoadConfig demonstrates loading configuration with default settings.
func ExampleLoadConfig() {
	// Clear any existing environment variables that might interfere
	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("GRPC_ADDRESS")
	
	// Create a temporary file with JWT secret for the example
	tmpfile, err := os.CreateTemp("", "example_secret")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte("example-jwt-secret")); err != nil {
		log.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}

	// Set JWT secret path via environment variable
	os.Setenv("JWT_SECRET_FILE", tmpfile.Name())
	defer os.Unsetenv("JWT_SECRET_FILE")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Output main settings
	fmt.Printf("Server address: %s\n", cfg.Address)
	fmt.Printf("Base URL: %s\n", cfg.BaseURL)
	fmt.Printf("File storage: %s\n", cfg.FileStorage)
	fmt.Printf("Has secret key: %v\n", len(cfg.SecretKey) > 0)

	// Output:
	// Server address: localhost:8080
	// Base URL: http://localhost:8080
	// File storage: urls.json
	// Has secret key: true
}

// ExampleLoadConfig_withEnvironmentVariables demonstrates loading configuration
// with settings from environment variables.
func ExampleLoadConfig_withEnvironmentVariables() {
	// Create a temporary file with JWT secret
	tmpfile, err := os.CreateTemp("", "example_secret")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte("env-jwt-secret")); err != nil {
		log.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}

	// Set environment variables
	os.Setenv("SERVER_ADDRESS", "localhost:9000")
	os.Setenv("BASE_URL", "https://myshortener.com")
	os.Setenv("FILE_STORAGE_PATH", "custom_urls.json")
	os.Setenv("JWT_SECRET_FILE", tmpfile.Name())

	defer func() {
		os.Unsetenv("SERVER_ADDRESS")
		os.Unsetenv("BASE_URL")
		os.Unsetenv("FILE_STORAGE_PATH")
		os.Unsetenv("JWT_SECRET_FILE")
	}()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Output settings from environment variables
	fmt.Printf("Custom address: %s\n", cfg.Address)
	fmt.Printf("Custom base URL: %s\n", cfg.BaseURL)
	fmt.Printf("Custom file storage: %s\n", cfg.FileStorage)

	// Output:
	// Custom address: localhost:9000
	// Custom base URL: https://myshortener.com
	// Custom file storage: custom_urls.json
}
