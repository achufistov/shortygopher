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

type Config struct {
	Address     string
	BaseURL     string
	FileStorage string
	DatabaseDSN string
	SecretKey   string
}

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
