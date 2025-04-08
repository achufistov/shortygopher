package storage

import (
	"sync"
)

type URLStorage struct {
	mu   sync.RWMutex
	URLs map[string]string
}

func NewURLStorage() *URLStorage {
	return &URLStorage{
		URLs: make(map[string]string),
	}
}

func (s *URLStorage) AddURL(shortURL, originalURL string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.URLs[shortURL] = originalURL
}

func (s *URLStorage) AddURLs(urls map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for shortURL, originalURL := range urls {
		s.URLs[shortURL] = originalURL
	}
	return nil
}

func (s *URLStorage) GetURL(shortURL string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	originalURL, exists := s.URLs[shortURL]
	return originalURL, exists
}

func (s *URLStorage) GetAllURLs() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copyMap := make(map[string]string)
	for k, v := range s.URLs {
		copyMap[k] = v
	}
	return copyMap
}

func (s *URLStorage) Ping() error {
	return nil // In-memory storage doesn't need to ping anything
}

func (s *URLStorage) Close() error {
	return nil // In-memory storage doesn't need to close anything
}
