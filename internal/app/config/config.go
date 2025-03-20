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
	flagsDefined    = false
)

type Config struct {
	Address     string
	BaseURL     string
	FileStorage string
}

func LoadConfig() (*Config, error) {
	// Check if flags have already been defined
	if !flagsDefined {
		flag.Parse()
		flagsDefined = true
	}

	// Check the environment variables first
	address := os.Getenv("SERVER_ADDRESS")
	baseURL := os.Getenv("BASE_URL")
	fileStorage := os.Getenv("FILE_STORAGE_PATH")

	// If the environment variables are not set, use the flags
	if address == "" {
		address = *addressFlag
	}
	if baseURL == "" {
		baseURL = *baseURLFlag
	}
	if fileStorage == "" {
		fileStorage = *fileStoragePath
	}

	// Check that the address and base URL are set
	if address == "" || baseURL == "" || fileStorage == "" {
		return nil, fmt.Errorf("address, base URL, and file storage path must be provided")
	}

	return &Config{
		Address:     address,
		BaseURL:     baseURL,
		FileStorage: fileStorage,
	}, nil
}
