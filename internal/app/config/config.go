package config

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	Address string
	BaseURL string
}

func LoadConfig() (*Config, error) {
	// check the environment variables first
	address := os.Getenv("SERVER_ADDRESS")
	baseURL := os.Getenv("BASE_URL")

	// if the environment variables are not set, check the flags
	if address == "" {
		addressFlag := flag.String("a", "localhost:8080", "HTTP server address")
		baseURLFlag := flag.String("b", "http://localhost:8080", "Base URL for shortened links")
		flag.Parse()
		address = *addressFlag
		baseURL = *baseURLFlag
	}

	// check that the address and base URL are set
	if address == "" || baseURL == "" {
		return nil, fmt.Errorf("address and base URL must be provided")
	}

	return &Config{
		Address: address,
		BaseURL: baseURL,
	}, nil
}
