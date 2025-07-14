package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateUUID(t *testing.T) {
	uuid1 := generateUUID()
	uuid2 := generateUUID()

	// Check that UUIDs are generated
	if uuid1 == "" {
		t.Error("Expected non-empty UUID")
	}
	if uuid2 == "" {
		t.Error("Expected non-empty UUID")
	}

	// Check that UUIDs are different
	if uuid1 == uuid2 {
		t.Error("Expected different UUIDs")
	}

	// Check UUID format (should be 36 characters with dashes)
	if len(uuid1) != 36 {
		t.Errorf("Expected UUID length 36, got %d", len(uuid1))
	}
	if len(uuid2) != 36 {
		t.Errorf("Expected UUID length 36, got %d", len(uuid2))
	}
}

func TestURLMapping(t *testing.T) {
	mapping := URLMapping{
		UUID:        "test-uuid",
		ShortURL:    "short1",
		OriginalURL: "https://example.com",
		UserID:      "user1",
	}

	if mapping.UUID != "test-uuid" {
		t.Errorf("Expected UUID 'test-uuid', got '%s'", mapping.UUID)
	}
	if mapping.ShortURL != "short1" {
		t.Errorf("Expected ShortURL 'short1', got '%s'", mapping.ShortURL)
	}
	if mapping.OriginalURL != "https://example.com" {
		t.Errorf("Expected OriginalURL 'https://example.com', got '%s'", mapping.OriginalURL)
	}
	if mapping.UserID != "user1" {
		t.Errorf("Expected UserID 'user1', got '%s'", mapping.UserID)
	}
}

func TestLoadURLMappings_NonExistentFile(t *testing.T) {
	tempDir := t.TempDir()
	nonExistentFile := filepath.Join(tempDir, "nonexistent.json")

	urlMap, err := LoadURLMappings(nonExistentFile)
	if err != nil {
		t.Errorf("LoadURLMappings() should not return error for non-existent file, got: %v", err)
	}

	if len(urlMap) != 0 {
		t.Errorf("Expected empty map for non-existent file, got %d items", len(urlMap))
	}
}

func TestLoadURLMappings_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	emptyFile := filepath.Join(tempDir, "empty.json")

	// Create empty file
	err := os.WriteFile(emptyFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	urlMap, err := LoadURLMappings(emptyFile)
	if err != nil {
		t.Errorf("LoadURLMappings() should not return error for empty file, got: %v", err)
	}

	if len(urlMap) != 0 {
		t.Errorf("Expected empty map for empty file, got %d items", len(urlMap))
	}
}

func TestLoadURLMappings_ValidFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.json")

	// Create test file with JSON lines
	testData := `{"uuid":"1","short_url":"short1","original_url":"https://example.com","user_id":"user1"}
{"uuid":"2","short_url":"short2","original_url":"https://google.com","user_id":"user2"}`

	err := os.WriteFile(testFile, []byte(testData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	urlMap, err := LoadURLMappings(testFile)
	if err != nil {
		t.Errorf("LoadURLMappings() returned error: %v", err)
	}

	if len(urlMap) != 2 {
		t.Errorf("Expected 2 URLs, got %d", len(urlMap))
	}

	expectedURLs := map[string]string{
		"short1": "https://example.com",
		"short2": "https://google.com",
	}

	for shortURL, expectedOriginal := range expectedURLs {
		if originalURL, exists := urlMap[shortURL]; !exists {
			t.Errorf("Expected URL '%s' not found", shortURL)
		} else if originalURL != expectedOriginal {
			t.Errorf("Expected '%s', got '%s' for URL '%s'", expectedOriginal, originalURL, shortURL)
		}
	}
}

func TestLoadURLMappings_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.json")

	// Create test file with mix of valid and invalid JSON
	testData := `{"uuid":"1","short_url":"short1","original_url":"https://example.com","user_id":"user1"}
invalid json line
{"uuid":"2","short_url":"short2","original_url":"https://google.com","user_id":"user2"}`

	err := os.WriteFile(testFile, []byte(testData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	urlMap, err := LoadURLMappings(testFile)
	if err != nil {
		t.Errorf("LoadURLMappings() returned error: %v", err)
	}

	// Should load only valid JSON lines
	if len(urlMap) != 2 {
		t.Errorf("Expected 2 URLs (invalid JSON should be skipped), got %d", len(urlMap))
	}
}

func TestSaveSingleURLMapping(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_single.json")

	err := SaveSingleURLMapping(testFile, "short1", "https://example.com")
	if err != nil {
		t.Errorf("SaveSingleURLMapping() returned error: %v", err)
	}

	// Verify file was created and contains the URL
	urlMap, err := LoadURLMappings(testFile)
	if err != nil {
		t.Errorf("LoadURLMappings() returned error: %v", err)
	}

	if len(urlMap) < 1 {
		t.Errorf("Expected at least 1 URL, got %d", len(urlMap))
	}

	// Check if our URL is present (there might be other URLs from global saver)
	if originalURL, exists := urlMap["short1"]; !exists {
		t.Error("Expected 'short1' URL not found")
	} else if originalURL != "https://example.com" {
		t.Errorf("Expected 'https://example.com', got '%s'", originalURL)
	}
}
