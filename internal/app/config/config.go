// Package config provides functions for loading and managing application configuration.
package config

import (
	"encoding/json"
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
	configFile      = flag.String("c", "", "Path to JSON configuration file (can also use -config)")
	enableHTTPS     = flag.Bool("s", false, "Enable HTTPS server")
	certFile        = flag.String("cert", "cert.pem", "Path to TLS certificate file")
	keyFile         = flag.String("key", "key.pem", "Path to TLS private key file")
	trustedSubnet   = flag.String("t", "", "Trusted subnet in CIDR notation")
)

// Config contains all configuration parameters for the URL shortening service.
// Configuration can be set via environment variables, command line flags, or JSON config file.
//
// Setting priority (from highest to lowest):
//  1. Environment variables
//  2. Command line flags
//  3. JSON config file
//  4. Default values
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
	Address string `json:"server_address"`

	// BaseURL defines the base URL for generating shortened links
	BaseURL string `json:"base_url"`

	// FileStorage defines the path to the file for persistent URL storage (can be empty)
	FileStorage string `json:"file_storage_path"`

	// DatabaseDSN contains the database connection string (can be empty)
	DatabaseDSN string `json:"database_dsn"`

	// SecretKey contains the secret key for JWT token signing
	SecretKey string `json:"-"`

	// EnableHTTPS indicates whether to enable HTTPS server
	EnableHTTPS bool `json:"enable_https"`

	// CertFile is the path to the TLS certificate file
	CertFile string `json:"cert_file"`

	// KeyFile is the path to the TLS private key file
	KeyFile string `json:"key_file"`

	// TrustedSubnet defines the trusted subnet in CIDR notation for internal endpoints
	TrustedSubnet string `json:"trusted_subnet"`
}

// LoadConfig loads configuration from environment variables, command line flags, and JSON config file.
// Returns a pointer to Config struct or an error if configuration loading fails.
//
// Supported environment variables:
//   - SERVER_ADDRESS: server address
//   - BASE_URL: base URL
//   - FILE_STORAGE_PATH: storage file path
//   - DATABASE_DSN: database connection string
//   - JWT_SECRET_FILE: path to JWT secret file
//   - ENABLE_HTTPS: enable HTTPS server (true/false)
//   - TLS_CERT_FILE: path to TLS certificate file
//   - TLS_KEY_FILE: path to TLS private key file
//   - CONFIG: path to JSON configuration file
//
// Supported flags:
//   - -a: server address
//   - -b: base URL
//   - -f: storage file path
//   - -d: database connection string
//   - -jwt-secret-file: path to JWT secret file
//   - -s: enable HTTPS server
//   - -cert: path to TLS certificate file
//   - -key: path to TLS private key file
//   - -c, -config: path to JSON configuration file
func LoadConfig() (*Config, error) {
	// Initialize config with default values
	config := &Config{
		Address:       *addressFlag,
		BaseURL:       *baseURLFlag,
		FileStorage:   *fileStoragePath,
		DatabaseDSN:   *databaseDSNFlag,
		CertFile:      *certFile,
		KeyFile:       *keyFile,
		EnableHTTPS:   *enableHTTPS,
		TrustedSubnet: *trustedSubnet,
	}

	// Load from JSON config file if specified
	configPath := os.Getenv("CONFIG")
	if configPath == "" {
		if *configFile != "" {
			configPath = *configFile
		}
	}

	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := json.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Override with command line flags
	if *enableHTTPS {
		config.EnableHTTPS = true
		config.CertFile = *certFile
		config.KeyFile = *keyFile
	}

	// Override with environment variables
	if envAddr := os.Getenv("SERVER_ADDRESS"); envAddr != "" {
		config.Address = envAddr
	}
	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		config.BaseURL = envBaseURL
	}
	if envFileStorage := os.Getenv("FILE_STORAGE_PATH"); envFileStorage != "" {
		config.FileStorage = envFileStorage
	}
	if envDSN := os.Getenv("DATABASE_DSN"); envDSN != "" {
		config.DatabaseDSN = envDSN
	}
	if os.Getenv("ENABLE_HTTPS") == "true" {
		config.EnableHTTPS = true
	}
	if envCertFile := os.Getenv("TLS_CERT_FILE"); envCertFile != "" {
		config.CertFile = envCertFile
	}
	if envKeyFile := os.Getenv("TLS_KEY_FILE"); envKeyFile != "" {
		config.KeyFile = envKeyFile
	}
	if envTrustedSubnet := os.Getenv("TRUSTED_SUBNET"); envTrustedSubnet != "" {
		config.TrustedSubnet = envTrustedSubnet
	}

	// Load JWT secret
	secretFile := os.Getenv("JWT_SECRET_FILE")
	if secretFile == "" {
		secretFile = *jwtSecretFile
	}

	secretKeyBytes, err := os.ReadFile(secretFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read JWT secret file: %w", err)
	}
	config.SecretKey = strings.TrimSpace(string(secretKeyBytes))

	// Validate required fields
	if config.Address == "" || config.BaseURL == "" || config.FileStorage == "" {
		return nil, fmt.Errorf("address, base URL, file storage path must be provided")
	}

	return config, nil
}
