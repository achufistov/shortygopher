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

// LoadURLMappings loads URLs from a file
func LoadURLMappings(filePath string) (map[string]string, error) {
	urlMap := make(map[string]string)

	// Checking if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// If the file does not exist, we return an empty card
		return urlMap, nil
	}

	// Reading the entire file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Decoding an array of JSON objects
	var mappings []URLMapping
	if err := json.Unmarshal(data, &mappings); err != nil {
		return nil, err
	}

	// Filling out the card
	for _, mapping := range mappings {
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
