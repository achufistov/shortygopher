package models

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/url"
)

// URL is a model for a shortened URL
type URL struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
	UserID      string `json:"user_id"`
}

// URLRepository defines the interface for working with URL storage
type URLRepository interface {
	AddURL(shortURL, originalURL, userID string) error
	AddURLs(urls map[string]string, userID string) error
	GetURL(shortURL string) (string, error)
	GetURLsByUser(userID string) (map[string]string, error)
	GetAllURLs() map[string]string
	GetShortURLByOriginalURL(originalURL string) (string, bool)
	Ping() error
	Close() error
	DeleteURLs(shortURLs []string, userID string) error
	GetUserURLs(userID string) ([]URL, error)
}

// ValidateURL verifies the correctness of the URL
func ValidateURL(rawURL string) error {
	_, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return errors.New("invalid URL format")
	}
	return nil
}

// GenerateShortURL creates a short URL
func GenerateShortURL() string {
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)[:6]
}
