// Package config provides functions for loading and managing application configuration.
package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

var (
	addressFlag     = flag.String("a", "localhost:8080", "HTTP server address")
	baseURLFlag     = flag.String("b", "http://localhost:8080", "Base URL for shortened links")
	fileStoragePath = flag.String("f", "urls.json", "File for storing urls")
	databaseDSNFlag = flag.String("d", "", "Database connection string")
	jwtSecretFile   = flag.String("jwt-secret-file", "secret.key", "Path to JWT secret file")
	flagsDefined    = false
)

// Config contains all configuration parameters for the URL shortening service.
// Configuration can be set via environment variables or command line flags.
//
// Setting priority (from highest to lowest):
//  1. Environment variables
//  2. Command line flags
//  3. Default values
//
// Example usage:
//
//	cfg, err := config.LoadConfig()
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Printf("Server started on %s\n", cfg.Address)
type Config struct {
	// Address defines the address and port for the HTTP server (e.g., "localhost:8080")
	Address string

	// BaseURL defines the base URL for generating shortened links
	BaseURL string

	// FileStorage defines the path to the file for persistent URL storage (can be empty)
	FileStorage string

	// DatabaseDSN contains the database connection string (can be empty)
	DatabaseDSN string

	// SecretKey contains the secret key for JWT token signing
	SecretKey string
}

// LoadConfig loads configuration from environment variables and command line flags.
// Returns a pointer to Config struct or an error if configuration loading fails.
//
// Supported environment variables:
//   - SERVER_ADDRESS: server address
//   - BASE_URL: base URL
//   - FILE_STORAGE_PATH: storage file path
//   - DATABASE_DSN: database connection string
//   - JWT_SECRET_FILE: path to JWT secret file
//
// Supported flags:
//   - -a: server address
//   - -b: base URL
//   - -f: storage file path
//   - -d: database connection string
//   - -jwt-secret-file: path to JWT secret file
func LoadConfig() (*Config, error) {
	if !flagsDefined {
		flag.Parse()
		flagsDefined = true
	}

	address := os.Getenv("SERVER_ADDRESS")
	baseURL := os.Getenv("BASE_URL")
	fileStorage := os.Getenv("FILE_STORAGE_PATH")
	databaseDSN := os.Getenv("DATABASE_DSN")

	if address == "" {
		address = *addressFlag
	}
	if baseURL == "" {
		baseURL = *baseURLFlag
	}
	if fileStorage == "" {
		fileStorage = *fileStoragePath
	}
	if databaseDSN == "" {
		databaseDSN = *databaseDSNFlag
	}

	secretFile := os.Getenv("JWT_SECRET_FILE")
	if secretFile == "" {
		secretFile = *jwtSecretFile
	}

	secretKeyBytes, err := os.ReadFile(secretFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read JWT secret file: %v", err)
	}
	secretKey := strings.TrimSpace(string(secretKeyBytes))

	if address == "" || baseURL == "" || fileStorage == "" {
		return nil, fmt.Errorf("address, base URL, file storage path must be provided")
	}

	return &Config{
		Address:     address,
		BaseURL:     baseURL,
		FileStorage: fileStorage,
		DatabaseDSN: databaseDSN,
		SecretKey:   secretKey,
	}, nil
}
