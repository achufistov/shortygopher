package storage

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/google/uuid"
)

type URLMapping struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// LoadURLMappings loads URLs from a file
func LoadURLMappings(filePath string) (map[string]string, error) {
	urlMap := make(map[string]string)

	// Проверяем, существует ли файл
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Если файл не существует, возвращаем пустую карту
		return urlMap, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		var mapping URLMapping
		if err := json.Unmarshal([]byte(line), &mapping); err != nil {
			return nil, err
		}
		urlMap[mapping.ShortURL] = mapping.OriginalURL
	}

	return urlMap, nil
}

// SaveURLMappings saves URLs to a file
func SaveURLMappings(filePath string, urlMap map[string]string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	for shortURL, originalURL := range urlMap {
		mapping := URLMapping{
			UUID:        generateUUID(),
			ShortURL:    shortURL,
			OriginalURL: originalURL,
		}
		line, err := json.Marshal(mapping)
		if err != nil {
			return err
		}
		file.Write(line)
		file.WriteString("\n")
	}

	return nil
}

// generateUUID generates a unique UUID
func generateUUID() string {
	return uuid.New().String()
}
