package storage

import (
	"errors"
	"sync"

	"github.com/achufistov/shortygopher.git/internal/app/models"
)

// Storage defines the interface for working with the storage
type Storage interface {
	AddURL(shortURL, originalURL, userID string) error
	AddURLs(urls map[string]string, userID string) error
	GetURL(shortURL string) (string, error)
	GetURLsByUser(userID string) (map[string]string, error)
	GetAllURLs() map[string]string
	GetShortURLByOriginalURL(originalURL string) (string, bool)
	Ping() error
	Close() error
	DeleteURLs(shortURLs []string, userID string) error
	GetUserURLs(userID string) ([]models.URL, error)
}

// URLInfo contains information about a URL
type URLInfo struct {
	OriginalURL string
	UserID      string
	DeletedFlag bool
}

// URLStorage implements in-memory URL storage
type URLStorage struct {
	mu   sync.RWMutex
	URLs map[string]URLInfo
}

// NewURLStorage creates a new URLStorage instance
func NewURLStorage() *URLStorage {
	return &URLStorage{
		URLs: make(map[string]URLInfo),
	}
}

// AddURL adds a URL to the storage
func (s *URLStorage) AddURL(shortURL, originalURL, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.URLs[shortURL] = URLInfo{OriginalURL: originalURL, UserID: userID}
	return nil
}

// AddURLs adds multiple URLs to the storage
func (s *URLStorage) AddURLs(urls map[string]string, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for shortURL, originalURL := range urls {
		s.URLs[shortURL] = URLInfo{OriginalURL: originalURL, UserID: userID}
	}
	return nil
}

// GetURL gets the original URL by the shortened URL
func (s *URLStorage) GetURL(shortURL string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	info, exists := s.URLs[shortURL]
	if !exists {
		return "", errors.New("URL not found")
	}
	if info.DeletedFlag {
		return "", errors.New("URL is deleted")
	}
	return info.OriginalURL, nil
}

// GetURLsByUser gets all user URLs in the form of a map
func (s *URLStorage) GetURLsByUser(userID string) (map[string]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	urls := make(map[string]string)
	for short, info := range s.URLs {
		if info.UserID == userID && !info.DeletedFlag {
			urls[short] = info.OriginalURL
		}
	}
	return urls, nil
}

// GetAllURLs gets all URLs
func (s *URLStorage) GetAllURLs() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	urlMap := make(map[string]string)
	for short, info := range s.URLs {
		if !info.DeletedFlag {
			urlMap[short] = info.OriginalURL
		}
	}
	return urlMap
}

// GetShortURLByOriginalURL gets the shortened URL by the original URL
func (s *URLStorage) GetShortURLByOriginalURL(originalURL string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for short, info := range s.URLs {
		if info.OriginalURL == originalURL && !info.DeletedFlag {
			return short, true
		}
	}
	return "", false
}

// GetUserURLs gets all user URLs in the form of a slice of models
func (s *URLStorage) GetUserURLs(userID string) ([]models.URL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var urls []models.URL
	for short, info := range s.URLs {
		if info.UserID == userID && !info.DeletedFlag {
			urls = append(urls, models.URL{
				OriginalURL: info.OriginalURL,
				ShortURL:    short,
				UserID:      info.UserID,
			})
		}
	}
	return urls, nil
}

// DeleteURLs deletes URLs
func (s *URLStorage) DeleteURLs(shortURLs []string, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, shortURL := range shortURLs {
		if info, exists := s.URLs[shortURL]; exists && info.UserID == userID {
			info.DeletedFlag = true
			s.URLs[shortURL] = info
		}
	}
	return nil
}

// Ping checks the storage availability
func (s *URLStorage) Ping() error {
	return nil
}

// Close closes the storage
func (s *URLStorage) Close() error {
	return nil
}
