package storage

import (
	"sync"
)

// URLInfo contains information about a stored URL, including the deletion flag.
type URLInfo struct {
	OriginalURL string
	UserID      string
	IsDeleted   bool
}

// URLStorage represents an in-memory storage for URL mappings.
// Implements the Storage interface and supports concurrent access via sync.RWMutex.
//
// Example usage:
//
//	storage := NewURLStorage()
//	err := storage.AddURL("abc123", "https://example.com", "user1")
//	if err != nil {
//		log.Fatal(err)
//	}
type URLStorage struct {
	mu      sync.RWMutex
	URLs    map[string]URLInfo
	mapPool sync.Pool
}

// NewURLStorage creates a new URLStorage instance with an initialized URL map.
// Returns a ready-to-use storage object.
func NewURLStorage() *URLStorage {
	storage := &URLStorage{
		URLs: make(map[string]URLInfo, 1000),
	}

	storage.mapPool = sync.Pool{
		New: func() interface{} {
			return make(map[string]string, 100)
		},
	}

	return storage
}

// AddURL adds a new URL mapping to the storage.
// Thread-safe operation that stores the mapping with user association.
func (s *URLStorage) AddURL(shortURL, originalURL, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.URLs[shortURL] = URLInfo{OriginalURL: originalURL, UserID: userID}
	return nil
}

// AddURLs adds multiple URL mappings in a single operation.
// More efficient than multiple AddURL calls for batch operations.
func (s *URLStorage) AddURLs(urls map[string]string, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for shortURL, originalURL := range urls {
		s.URLs[shortURL] = URLInfo{OriginalURL: originalURL, UserID: userID}
	}
	return nil
}

// GetURL retrieves URL information by short URL.
// Returns original URL, existence flag, and deletion status.
func (s *URLStorage) GetURL(shortURL string) (string, bool, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	info, exists := s.URLs[shortURL]
	if !exists {
		return "", false, false
	}
	return info.OriginalURL, true, info.IsDeleted
}

// GetURLsByUser retrieves all URLs created by a specific user.
// Uses sync.Pool for efficient map allocation and reuse.
func (s *URLStorage) GetURLsByUser(userID string) (map[string]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	urls := s.mapPool.Get().(map[string]string)
	defer func() {
		for k := range urls {
			delete(urls, k)
		}
		s.mapPool.Put(urls)
	}()

	result := make(map[string]string, len(urls))

	for short, info := range s.URLs {
		if info.UserID == userID {
			result[short] = info.OriginalURL
		}
	}
	return result, nil
}

// GetAllURLs returns a copy of all stored URL mappings.
// Creates a new map to avoid exposing internal storage.
func (s *URLStorage) GetAllURLs() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	urlMap := make(map[string]string, len(s.URLs))
	for short, info := range s.URLs {
		urlMap[short] = info.OriginalURL
	}
	return urlMap
}

// IterateURLs calls the provided function for each URL mapping.
// Provides efficient iteration without copying all data.
func (s *URLStorage) IterateURLs(fn func(shortURL, originalURL string)) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for short, info := range s.URLs {
		fn(short, info.OriginalURL)
	}
}

// Count returns the total number of stored URL mappings.
// Thread-safe read operation.
func (s *URLStorage) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.URLs)
}

// GetShortURLByOriginalURL finds the short URL for a given original URL.
// Returns short URL and found flag by iterating through all mappings.
func (s *URLStorage) GetShortURLByOriginalURL(originalURL string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for short, info := range s.URLs {
		if info.OriginalURL == originalURL {
			return short, true
		}
	}
	return "", false
}

// DeleteURLs marks specified URLs as deleted for the given user.
// Only URLs owned by the user are marked for deletion.
func (s *URLStorage) DeleteURLs(shortURLs []string, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, shortURL := range shortURLs {
		if info, exists := s.URLs[shortURL]; exists && info.UserID == userID {
			info.IsDeleted = true
			s.URLs[shortURL] = info
		}
	}
	return nil
}

// Ping checks storage availability (always returns nil for in-memory storage).
func (s *URLStorage) Ping() error {
	return nil
}

// Close performs cleanup operations (no-op for in-memory storage).
func (s *URLStorage) Close() error {
	return nil
}

// GetStats returns statistics about the storage.
// Returns the number of URLs and unique users.
func (s *URLStorage) GetStats() (int, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	urlCount := len(s.URLs)
	userSet := make(map[string]bool)

	for _, info := range s.URLs {
		userSet[info.UserID] = true
	}

	userCount := len(userSet)
	return urlCount, userCount, nil
}
