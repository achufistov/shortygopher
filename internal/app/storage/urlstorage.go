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

// AddURL добавляет URL в хранилище
func (s *URLStorage) AddURL(shortURL, originalURL string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.URLs[shortURL] = originalURL
}

// GetURL возвращает оригинальный URL по короткому
func (s *URLStorage) GetURL(shortURL string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	originalURL, exists := s.URLs[shortURL]
	return originalURL, exists
}

// GetAllURLs возвращает копию всех URL
func (s *URLStorage) GetAllURLs() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copyMap := make(map[string]string)
	for k, v := range s.URLs {
		copyMap[k] = v
	}
	return copyMap
}
