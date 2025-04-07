// config.go
package config

import (
	"flag"
	"fmt"
	"os"
)

var (
	addressFlag     = flag.String("a", "localhost:8080", "HTTP server address")
	baseURLFlag     = flag.String("b", "http://localhost:8080", "Base URL for shortened links")
	fileStoragePath = flag.String("f", "urls.json", "File for storing urls")
	databaseDSNFlag = flag.String("d", "", "Database connection string")
	flagsDefined    = false
)

type Config struct {
	Address     string
	BaseURL     string
	FileStorage string
	DatabaseDSN string
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

	if address == "" || baseURL == "" || fileStorage == "" || databaseDSN == "" {
		return nil, fmt.Errorf("address, base URL, file storage path, and database DSN must be provided")
	}

	return &Config{
		Address:     address,
		BaseURL:     baseURL,
		FileStorage: fileStorage,
		DatabaseDSN: databaseDSN,
	}, nil
}
