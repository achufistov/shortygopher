package storage

import (
	"sync"
)

type URLInfo struct {
	OriginalURL string
	UserID      string
}

type URLStorage struct {
	mu   sync.RWMutex
	URLs map[string]URLInfo
}

func NewURLStorage() *URLStorage {
	return &URLStorage{
		URLs: make(map[string]URLInfo),
	}
}

func (s *URLStorage) AddURL(shortURL, originalURL, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.URLs[shortURL] = URLInfo{OriginalURL: originalURL, UserID: userID}
	return nil
}

func (s *URLStorage) AddURLs(urls map[string]string, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for shortURL, originalURL := range urls {
		s.URLs[shortURL] = URLInfo{OriginalURL: originalURL, UserID: userID}
	}
	return nil
}

func (s *URLStorage) GetURL(shortURL string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	info, exists := s.URLs[shortURL]
	if !exists {
		return "", false
	}
	return info.OriginalURL, true
}

func (s *URLStorage) GetURLsByUser(userID string) (map[string]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	urls := make(map[string]string)
	for short, info := range s.URLs {
		if info.UserID == userID {
			urls[short] = info.OriginalURL
		}
	}
	return urls, nil
}

func (s *URLStorage) GetAllURLs() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	urlMap := make(map[string]string)
	for short, info := range s.URLs {
		urlMap[short] = info.OriginalURL
	}
	return urlMap
}

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

func (s *URLStorage) Ping() error {
	return nil
}

func (s *URLStorage) Close() error {
	return nil
}
