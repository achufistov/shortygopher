package storage

import (
	"errors"
	"sync"

	"github.com/achufistov/shortygopher.git/internal/app/models"
)

// Storage определяет интерфейс для работы с хранилищем
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

// URLInfo содержит информацию о URL
type URLInfo struct {
	OriginalURL string
	UserID      string
	DeletedFlag bool
}

// URLStorage реализует in-memory хранилище URL
type URLStorage struct {
	mu   sync.RWMutex
	URLs map[string]URLInfo
}

// NewURLStorage создает новый экземпляр URLStorage
func NewURLStorage() *URLStorage {
	return &URLStorage{
		URLs: make(map[string]URLInfo),
	}
}

// AddURL добавляет URL в хранилище
func (s *URLStorage) AddURL(shortURL, originalURL, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.URLs[shortURL] = URLInfo{OriginalURL: originalURL, UserID: userID}
	return nil
}

// AddURLs добавляет несколько URL в хранилище
func (s *URLStorage) AddURLs(urls map[string]string, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for shortURL, originalURL := range urls {
		s.URLs[shortURL] = URLInfo{OriginalURL: originalURL, UserID: userID}
	}
	return nil
}

// GetURL получает оригинальный URL по сокращенному
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

// GetURLsByUser получает все URL пользователя в виде map
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

// GetAllURLs получает все URL
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

// GetShortURLByOriginalURL получает сокращенный URL по оригинальному
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

// GetUserURLs получает все URL пользователя в виде слайса моделей
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

// DeleteURLs удаляет URL пользователя
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

// Ping проверяет доступность хранилища
func (s *URLStorage) Ping() error {
	return nil
}

// Close закрывает хранилище
func (s *URLStorage) Close() error {
	return nil
}
