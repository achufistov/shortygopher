package storage

import (
	"testing"
)

func TestNewURLStorage(t *testing.T) {
	storage := NewURLStorage()

	if storage == nil {
		t.Fatal("NewURLStorage() returned nil")
	}

	if storage.URLs == nil {
		t.Error("URLs map is not initialized")
	}

	if len(storage.URLs) != 0 {
		t.Errorf("Expected empty URLs map, got %d items", len(storage.URLs))
	}
}

func TestURLStorage_AddURL(t *testing.T) {
	storage := NewURLStorage()

	err := storage.AddURL("short1", "https://example.com", "user1")
	if err != nil {
		t.Errorf("AddURL() returned error: %v", err)
	}

	// Verify the URL was added
	originalURL, exists, isDeleted := storage.GetURL("short1")
	if !exists {
		t.Error("URL was not found after adding")
	}
	if originalURL != "https://example.com" {
		t.Errorf("Expected original URL 'https://example.com', got '%s'", originalURL)
	}
	if isDeleted {
		t.Error("URL should not be marked as deleted")
	}
}

func TestURLStorage_AddURLs(t *testing.T) {
	storage := NewURLStorage()

	urls := map[string]string{
		"short1": "https://example.com",
		"short2": "https://google.com",
		"short3": "https://github.com",
	}

	err := storage.AddURLs(urls, "user1")
	if err != nil {
		t.Errorf("AddURLs() returned error: %v", err)
	}

	// Verify all URLs were added
	for shortURL, expectedOriginal := range urls {
		originalURL, exists, isDeleted := storage.GetURL(shortURL)
		if !exists {
			t.Errorf("URL '%s' was not found after adding", shortURL)
		}
		if originalURL != expectedOriginal {
			t.Errorf("Expected original URL '%s', got '%s'", expectedOriginal, originalURL)
		}
		if isDeleted {
			t.Errorf("URL '%s' should not be marked as deleted", shortURL)
		}
	}

	// Check count
	if storage.Count() != 3 {
		t.Errorf("Expected 3 URLs, got %d", storage.Count())
	}
}

func TestURLStorage_GetURL(t *testing.T) {
	storage := NewURLStorage()

	// Test non-existent URL
	_, exists, _ := storage.GetURL("nonexistent")
	if exists {
		t.Error("Expected non-existent URL to not exist")
	}

	// Add URL and test retrieval
	storage.AddURL("short1", "https://example.com", "user1")
	originalURL, exists, isDeleted := storage.GetURL("short1")

	if !exists {
		t.Error("Expected URL to exist")
	}
	if originalURL != "https://example.com" {
		t.Errorf("Expected 'https://example.com', got '%s'", originalURL)
	}
	if isDeleted {
		t.Error("URL should not be deleted")
	}
}

func TestURLStorage_GetURLsByUser(t *testing.T) {
	storage := NewURLStorage()

	// Add URLs for different users
	storage.AddURL("short1", "https://example.com", "user1")
	storage.AddURL("short2", "https://google.com", "user1")
	storage.AddURL("short3", "https://github.com", "user2")

	// Get URLs for user1
	user1URLs, err := storage.GetURLsByUser("user1")
	if err != nil {
		t.Errorf("GetURLsByUser() returned error: %v", err)
	}

	if len(user1URLs) != 2 {
		t.Errorf("Expected 2 URLs for user1, got %d", len(user1URLs))
	}

	expectedURLs := map[string]string{
		"short1": "https://example.com",
		"short2": "https://google.com",
	}

	for shortURL, expectedOriginal := range expectedURLs {
		if originalURL, exists := user1URLs[shortURL]; !exists {
			t.Errorf("Expected URL '%s' not found for user1", shortURL)
		} else if originalURL != expectedOriginal {
			t.Errorf("Expected '%s', got '%s' for URL '%s'", expectedOriginal, originalURL, shortURL)
		}
	}

	// Get URLs for user2
	user2URLs, err := storage.GetURLsByUser("user2")
	if err != nil {
		t.Errorf("GetURLsByUser() returned error: %v", err)
	}

	if len(user2URLs) != 1 {
		t.Errorf("Expected 1 URL for user2, got %d", len(user2URLs))
	}

	// Get URLs for non-existent user
	emptyURLs, err := storage.GetURLsByUser("nonexistent")
	if err != nil {
		t.Errorf("GetURLsByUser() returned error: %v", err)
	}

	if len(emptyURLs) != 0 {
		t.Errorf("Expected 0 URLs for non-existent user, got %d", len(emptyURLs))
	}
}

func TestURLStorage_GetShortURLByOriginalURL(t *testing.T) {
	storage := NewURLStorage()

	// Test non-existent original URL
	_, found := storage.GetShortURLByOriginalURL("https://nonexistent.com")
	if found {
		t.Error("Expected non-existent original URL to not be found")
	}

	// Add URL and test retrieval
	storage.AddURL("short1", "https://example.com", "user1")
	shortURL, found := storage.GetShortURLByOriginalURL("https://example.com")

	if !found {
		t.Error("Expected original URL to be found")
	}
	if shortURL != "short1" {
		t.Errorf("Expected 'short1', got '%s'", shortURL)
	}
}

func TestURLStorage_DeleteURLs(t *testing.T) {
	storage := NewURLStorage()

	// Add URLs for different users
	storage.AddURL("short1", "https://example.com", "user1")
	storage.AddURL("short2", "https://google.com", "user1")
	storage.AddURL("short3", "https://github.com", "user2")

	// Delete URLs for user1
	err := storage.DeleteURLs([]string{"short1", "short2", "short3"}, "user1")
	if err != nil {
		t.Errorf("DeleteURLs() returned error: %v", err)
	}

	// Check that user1's URLs are marked as deleted
	_, exists1, isDeleted1 := storage.GetURL("short1")
	if !exists1 {
		t.Error("URL short1 should still exist")
	}
	if !isDeleted1 {
		t.Error("URL short1 should be marked as deleted")
	}

	_, exists2, isDeleted2 := storage.GetURL("short2")
	if !exists2 {
		t.Error("URL short2 should still exist")
	}
	if !isDeleted2 {
		t.Error("URL short2 should be marked as deleted")
	}

	// Check that user2's URL is not deleted (different user)
	_, exists3, isDeleted3 := storage.GetURL("short3")
	if !exists3 {
		t.Error("URL short3 should still exist")
	}
	if isDeleted3 {
		t.Error("URL short3 should not be marked as deleted")
	}
}

func TestURLStorage_GetAllURLs(t *testing.T) {
	storage := NewURLStorage()

	// Test empty storage
	allURLs := storage.GetAllURLs()
	if len(allURLs) != 0 {
		t.Errorf("Expected empty map, got %d items", len(allURLs))
	}

	// Add some URLs
	storage.AddURL("short1", "https://example.com", "user1")
	storage.AddURL("short2", "https://google.com", "user2")

	allURLs = storage.GetAllURLs()
	if len(allURLs) != 2 {
		t.Errorf("Expected 2 URLs, got %d", len(allURLs))
	}

	expectedURLs := map[string]string{
		"short1": "https://example.com",
		"short2": "https://google.com",
	}

	for shortURL, expectedOriginal := range expectedURLs {
		if originalURL, exists := allURLs[shortURL]; !exists {
			t.Errorf("Expected URL '%s' not found", shortURL)
		} else if originalURL != expectedOriginal {
			t.Errorf("Expected '%s', got '%s' for URL '%s'", expectedOriginal, originalURL, shortURL)
		}
	}
}

func TestURLStorage_Count(t *testing.T) {
	storage := NewURLStorage()

	// Test empty storage
	if count := storage.Count(); count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}

	// Add URLs
	storage.AddURL("short1", "https://example.com", "user1")
	if count := storage.Count(); count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}

	storage.AddURL("short2", "https://google.com", "user1")
	if count := storage.Count(); count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

func TestURLStorage_IterateURLs(t *testing.T) {
	storage := NewURLStorage()

	// Add test URLs
	expectedURLs := map[string]string{
		"short1": "https://example.com",
		"short2": "https://google.com",
		"short3": "https://github.com",
	}

	for short, original := range expectedURLs {
		storage.AddURL(short, original, "user1")
	}

	// Iterate and collect URLs
	collectedURLs := make(map[string]string)
	storage.IterateURLs(func(shortURL, originalURL string) {
		collectedURLs[shortURL] = originalURL
	})

	// Verify collected URLs match expected
	if len(collectedURLs) != len(expectedURLs) {
		t.Errorf("Expected %d URLs, collected %d", len(expectedURLs), len(collectedURLs))
	}

	for shortURL, expectedOriginal := range expectedURLs {
		if originalURL, exists := collectedURLs[shortURL]; !exists {
			t.Errorf("Expected URL '%s' not found in iteration", shortURL)
		} else if originalURL != expectedOriginal {
			t.Errorf("Expected '%s', got '%s' for URL '%s'", expectedOriginal, originalURL, shortURL)
		}
	}
}

func TestURLStorage_Ping(t *testing.T) {
	storage := NewURLStorage()

	err := storage.Ping()
	if err != nil {
		t.Errorf("Ping() should not return error for in-memory storage, got: %v", err)
	}
}

func TestURLStorage_Close(t *testing.T) {
	storage := NewURLStorage()

	err := storage.Close()
	if err != nil {
		t.Errorf("Close() should not return error for in-memory storage, got: %v", err)
	}
}
