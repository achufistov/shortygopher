package storage

import (
	"encoding/json"
	"os"

	"github.com/achufistov/shortygopher.git/internal/app/models"
	"github.com/google/uuid"
)

// FileStorage implements URL storage in a file
type FileStorage struct {
	filePath string
	storage  *URLStorage
}

// NewFileStorage creates a new FileStorage instance
func NewFileStorage(filePath string) (*FileStorage, error) {
	storage := NewURLStorage()
	fs := &FileStorage{
		filePath: filePath,
		storage:  storage,
	}

	// Load existing URLs from a file
	urlMap, err := fs.loadURLMappings()
	if err != nil {
		return nil, err
	}

	// Add loaded URLs to the storage
	for shortURL, originalURL := range urlMap {
		if err := storage.AddURL(shortURL, originalURL, "system"); err != nil {
			return nil, err
		}
	}

	return fs, nil
}

// loadURLMappings loads URLs from a file
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
		// If it's not possible to parse as an array, try parsing as a map
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

// saveURLMappings saves URLs to a file
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

// AddURL adds a URL to the storage
func (fs *FileStorage) AddURL(shortURL, originalURL, userID string) error {
	if err := fs.storage.AddURL(shortURL, originalURL, userID); err != nil {
		return err
	}
	return fs.saveURLMappings()
}

// AddURLs adds multiple URLs to the storage
func (fs *FileStorage) AddURLs(urls map[string]string, userID string) error {
	if err := fs.storage.AddURLs(urls, userID); err != nil {
		return err
	}
	return fs.saveURLMappings()
}

// GetURL gets the original URL by the shortened URL
func (fs *FileStorage) GetURL(shortURL string) (string, error) {
	return fs.storage.GetURL(shortURL)
}

// GetURLsByUser gets all user URLs
func (fs *FileStorage) GetURLsByUser(userID string) (map[string]string, error) {
	return fs.storage.GetURLsByUser(userID)
}

// GetAllURLs gets all URLs
func (fs *FileStorage) GetAllURLs() map[string]string {
	return fs.storage.GetAllURLs()
}

// GetShortURLByOriginalURL gets the shortened URL by the original URL
func (fs *FileStorage) GetShortURLByOriginalURL(originalURL string) (string, bool) {
	return fs.storage.GetShortURLByOriginalURL(originalURL)
}

// GetUserURLs gets all user URLs in the form of a slice of models
func (fs *FileStorage) GetUserURLs(userID string) ([]models.URL, error) {
	return fs.storage.GetUserURLs(userID)
}

// DeleteURLs deletes URLs
func (fs *FileStorage) DeleteURLs(shortURLs []string, userID string) error {
	if err := fs.storage.DeleteURLs(shortURLs, userID); err != nil {
		return err
	}
	return fs.saveURLMappings()
}

// Ping checks the storage availability
func (fs *FileStorage) Ping() error {
	return nil
}

// Closes the storage
func (fs *FileStorage) Close() error {
	return fs.saveURLMappings()
}
