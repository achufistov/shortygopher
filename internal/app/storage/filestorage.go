package storage

import (
	"encoding/json"
	"os"

	"github.com/achufistov/shortygopher.git/internal/app/models"
	"github.com/google/uuid"
)

// FileStorage реализует хранилище URL в файле
type FileStorage struct {
	filePath string
	storage  *URLStorage
}

// NewFileStorage создает новый экземпляр FileStorage
func NewFileStorage(filePath string) (*FileStorage, error) {
	storage := NewURLStorage()
	fs := &FileStorage{
		filePath: filePath,
		storage:  storage,
	}

	// Загружаем существующие URL из файла
	urlMap, err := fs.loadURLMappings()
	if err != nil {
		return nil, err
	}

	// Добавляем загруженные URL в хранилище
	for shortURL, originalURL := range urlMap {
		if err := storage.AddURL(shortURL, originalURL, "system"); err != nil {
			return nil, err
		}
	}

	return fs, nil
}

// loadURLMappings загружает URL из файла
func (fs *FileStorage) loadURLMappings() (map[string]string, error) {
	urlMap := make(map[string]string)
	if fs.filePath == "" {
		return urlMap, nil
	}

	if _, err := os.Stat(fs.filePath); os.IsNotExist(err) {
		return urlMap, nil
	}

	data, err := os.ReadFile(fs.filePath)
	if err != nil {
		return nil, err
	}

	var mappings []struct {
		UUID        string `json:"uuid"`
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
		UserID      string `json:"user_id"`
	}

	if err := json.Unmarshal(data, &mappings); err != nil {
		// Если не удалось распарсить как массив, пробуем как map
		var mapData map[string]string
		if err := json.Unmarshal(data, &mapData); err != nil {
			return nil, err
		}
		return mapData, nil
	}

	for _, mapping := range mappings {
		urlMap[mapping.ShortURL] = mapping.OriginalURL
	}
	return urlMap, nil
}

// saveURLMappings сохраняет URL в файл
func (fs *FileStorage) saveURLMappings() error {
	if fs.filePath == "" {
		return nil
	}

	file, err := os.OpenFile(fs.filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	urlMap := fs.storage.GetAllURLs()
	for shortURL, originalURL := range urlMap {
		mapping := struct {
			UUID        string `json:"uuid"`
			ShortURL    string `json:"short_url"`
			OriginalURL string `json:"original_url"`
			UserID      string `json:"user_id"`
		}{
			UUID:        uuid.New().String(),
			ShortURL:    shortURL,
			OriginalURL: originalURL,
			UserID:      "system",
		}
		line, err := json.Marshal(mapping)
		if err != nil {
			return err
		}
		if _, err := file.Write(line); err != nil {
			return err
		}
		if _, err := file.WriteString("\n"); err != nil {
			return err
		}
	}
	return nil
}

// AddURL добавляет URL в хранилище
func (fs *FileStorage) AddURL(shortURL, originalURL, userID string) error {
	if err := fs.storage.AddURL(shortURL, originalURL, userID); err != nil {
		return err
	}
	return fs.saveURLMappings()
}

// AddURLs добавляет несколько URL в хранилище
func (fs *FileStorage) AddURLs(urls map[string]string, userID string) error {
	if err := fs.storage.AddURLs(urls, userID); err != nil {
		return err
	}
	return fs.saveURLMappings()
}

// GetURL получает оригинальный URL по сокращенному
func (fs *FileStorage) GetURL(shortURL string) (string, error) {
	return fs.storage.GetURL(shortURL)
}

// GetURLsByUser получает все URL пользователя
func (fs *FileStorage) GetURLsByUser(userID string) (map[string]string, error) {
	return fs.storage.GetURLsByUser(userID)
}

// GetAllURLs получает все URL
func (fs *FileStorage) GetAllURLs() map[string]string {
	return fs.storage.GetAllURLs()
}

// GetShortURLByOriginalURL получает сокращенный URL по оригинальному
func (fs *FileStorage) GetShortURLByOriginalURL(originalURL string) (string, bool) {
	return fs.storage.GetShortURLByOriginalURL(originalURL)
}

// GetUserURLs получает все URL пользователя в виде слайса моделей
func (fs *FileStorage) GetUserURLs(userID string) ([]models.URL, error) {
	return fs.storage.GetUserURLs(userID)
}

// DeleteURLs удаляет URL пользователя
func (fs *FileStorage) DeleteURLs(shortURLs []string, userID string) error {
	if err := fs.storage.DeleteURLs(shortURLs, userID); err != nil {
		return err
	}
	return fs.saveURLMappings()
}

// Ping проверяет доступность хранилища
func (fs *FileStorage) Ping() error {
	return nil
}

// Close закрывает хранилище
func (fs *FileStorage) Close() error {
	return fs.saveURLMappings()
}
