package storage

import (
	"sync"
)

type URLInfo struct {
	OriginalURL string
	UserID      string
	DeletedFlag bool
}

type URLStorage struct {
	mu      sync.RWMutex
	URLs    map[string]URLInfo
	mapPool sync.Pool
}

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

func (s *URLStorage) GetURL(shortURL string) (string, bool, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	info, exists := s.URLs[shortURL]
	if !exists {
		return "", false, false
	}
	return info.OriginalURL, true, info.DeletedFlag
}

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

func (s *URLStorage) GetAllURLs() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	urlMap := make(map[string]string, len(s.URLs))
	for short, info := range s.URLs {
		urlMap[short] = info.OriginalURL
	}
	return urlMap
}

func (s *URLStorage) IterateURLs(fn func(shortURL, originalURL string)) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for short, info := range s.URLs {
		fn(short, info.OriginalURL)
	}
}

func (s *URLStorage) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.URLs)
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

func (s *URLStorage) Ping() error {
	return nil
}

func (s *URLStorage) Close() error {
	return nil
}
