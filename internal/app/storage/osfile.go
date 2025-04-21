package storage

import (
	"encoding/json"
	"os"

	"github.com/google/uuid"
)

type URLMapping struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func LoadURLMappings(filePath string) (map[string]string, error) {
	urlMap := make(map[string]string)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return urlMap, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var mappings []URLMapping
	if err := json.Unmarshal(data, &mappings); err != nil {
		return nil, err
	}

	for _, mapping := range mappings {
		urlMap[mapping.ShortURL] = mapping.OriginalURL
	}

	return urlMap, nil
}

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

func generateUUID() string {
	return uuid.New().String()
}
