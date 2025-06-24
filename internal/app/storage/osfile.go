package storage

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

type URLMapping struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
}

type BatchFileSaver struct {
	mu           sync.Mutex
	pendingURLs  map[string]string
	filePath     string
	saveInterval time.Duration
}

var (
	globalSaver     *BatchFileSaver
	globalSaverOnce sync.Once
)

func GetBatchSaver(filePath string) *BatchFileSaver {
	globalSaverOnce.Do(func() {
		globalSaver = &BatchFileSaver{
			pendingURLs:  make(map[string]string),
			filePath:     filePath,
			saveInterval: 5 * time.Second,
		}
		go globalSaver.periodicSave()
	})
	return globalSaver
}

func (b *BatchFileSaver) AddURL(shortURL, originalURL string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.pendingURLs[shortURL] = originalURL
}

func (b *BatchFileSaver) periodicSave() {
	ticker := time.NewTicker(b.saveInterval)
	defer ticker.Stop()

	for range ticker.C {
		b.forceSave()
	}
}

func (b *BatchFileSaver) forceSave() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.pendingURLs) == 0 {
		return nil
	}

	return b.saveToFile()
}

func (b *BatchFileSaver) saveToFile() error {
	tmpFile := b.filePath + ".tmp"
	file, err := os.Create(tmpFile)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for shortURL, originalURL := range b.pendingURLs {
		mapping := URLMapping{
			UUID:        generateUUID(),
			ShortURL:    shortURL,
			OriginalURL: originalURL,
			UserID:      "system",
		}
		line, err := json.Marshal(mapping)
		if err != nil {
			return err
		}
		writer.Write(line)
		writer.WriteString("\n")
	}

	b.pendingURLs = make(map[string]string)

	return os.Rename(tmpFile, b.filePath)
}

func LoadURLMappings(filePath string) (map[string]string, error) {
	urlMap := make(map[string]string)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return urlMap, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var mapping URLMapping
		if err := json.Unmarshal(scanner.Bytes(), &mapping); err != nil {
			continue
		}
		urlMap[mapping.ShortURL] = mapping.OriginalURL
	}

	return urlMap, scanner.Err()
}

func SaveURLMappings(filePath string, urlMap map[string]string) error {
	saver := GetBatchSaver(filePath)

	for shortURL, originalURL := range urlMap {
		saver.AddURL(shortURL, originalURL)
	}

	return saver.forceSave()
}

func SaveSingleURLMapping(filePath string, shortURL, originalURL string) error {
	saver := GetBatchSaver(filePath)
	saver.AddURL(shortURL, originalURL)
	return saver.forceSave()
}

func generateUUID() string {
	return uuid.New().String()
}
