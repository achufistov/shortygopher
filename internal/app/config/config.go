package config

import (
	"flag"
	"fmt"
	"os"
)

var (
	addressFlag  = flag.String("a", "localhost:8080", "HTTP server address")
	baseURLFlag  = flag.String("b", "http://localhost:8080", "Base URL for shortened links")
	flagsDefined = false
)

type Config struct {
	Address string
	BaseURL string
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

	// If the environment variables are not set, use the flags
	if address == "" {
		address = *addressFlag
	}
	if baseURL == "" {
		baseURL = *baseURLFlag
	}

	// Check that the address and base URL are set
	if address == "" || baseURL == "" {
		return nil, fmt.Errorf("address and base URL must be provided")
	}

	return &Config{
		Address: address,
		BaseURL: baseURL,
	}, nil
}
