package unit_test

import (
	"testing"

	"github.com/achufistov/shortygopher.git/internal/app/storage"
)

func TestNewURLStorage(t *testing.T) {
	storageInstance := storage.NewURLStorage()
	if storageInstance == nil {
		t.Fatal("NewURLStorage returned nil")
	}
}

func TestURLStorage_AddURL(t *testing.T) {
	storageInstance := storage.NewURLStorage()

	err := storageInstance.AddURL("short1", "http://example.com", "user1")
	if err != nil {
		t.Fatalf("AddURL failed: %v", err)
	}

	originalURL, exists, isDeleted := storageInstance.GetURL("short1")
	if !exists {
		t.Fatal("URL not found after adding")
	}
	if originalURL != "http://example.com" {
		t.Fatalf("Expected http://example.com, got %s", originalURL)
	}
	if isDeleted {
		t.Fatal("URL should not be deleted")
	}
}

func TestURLStorage_GetURL(t *testing.T) {
	storageInstance := storage.NewURLStorage()

	_, exists, _ := storageInstance.GetURL("nonexistent")
	if exists {
		t.Fatal("Non-existent URL reported as existing")
	}

	storageInstance.AddURL("short1", "http://example.com", "user1")
	originalURL, exists, isDeleted := storageInstance.GetURL("short1")

	if !exists {
		t.Fatal("URL not found")
	}
	if originalURL != "http://example.com" {
		t.Fatalf("Expected http://example.com, got %s", originalURL)
	}
	if isDeleted {
		t.Fatal("URL should not be deleted")
	}
}

func TestURLStorage_GetURLsByUser(t *testing.T) {
	storageInstance := storage.NewURLStorage()

	storageInstance.AddURL("short1", "http://example1.com", "user1")
	storageInstance.AddURL("short2", "http://example2.com", "user1")
	storageInstance.AddURL("short3", "http://example3.com", "user2")

	urls, err := storageInstance.GetURLsByUser("user1")
	if err != nil {
		t.Fatalf("GetURLsByUser failed: %v", err)
	}

	if len(urls) != 2 {
		t.Fatalf("Expected 2 URLs for user1, got %d", len(urls))
	}

	if urls["short1"] != "http://example1.com" {
		t.Fatalf("Expected http://example1.com, got %s", urls["short1"])
	}

	if urls["short2"] != "http://example2.com" {
		t.Fatalf("Expected http://example2.com, got %s", urls["short2"])
	}

	urls, err = storageInstance.GetURLsByUser("nonexistent")
	if err != nil {
		t.Fatalf("GetURLsByUser failed for nonexistent user: %v", err)
	}

	if len(urls) != 0 {
		t.Fatalf("Expected 0 URLs for nonexistent user, got %d", len(urls))
	}
}

func TestURLStorage_GetAllURLs(t *testing.T) {
	storageInstance := storage.NewURLStorage()

	urls := storageInstance.GetAllURLs()
	if len(urls) != 0 {
		t.Fatalf("Expected 0 URLs, got %d", len(urls))
	}

	storageInstance.AddURL("short1", "http://example1.com", "user1")
	storageInstance.AddURL("short2", "http://example2.com", "user1")

	urls = storageInstance.GetAllURLs()
	if len(urls) != 2 {
		t.Fatalf("Expected 2 URLs, got %d", len(urls))
	}

	if urls["short1"] != "http://example1.com" {
		t.Fatalf("Expected http://example1.com, got %s", urls["short1"])
	}
}

func TestURLStorage_GetShortURLByOriginalURL(t *testing.T) {
	storageInstance := storage.NewURLStorage()

	_, exists := storageInstance.GetShortURLByOriginalURL("http://nonexistent.com")
	if exists {
		t.Fatal("Non-existent URL reported as existing")
	}

	storageInstance.AddURL("short1", "http://example.com", "user1")
	shortURL, exists := storageInstance.GetShortURLByOriginalURL("http://example.com")

	if !exists {
		t.Fatal("URL not found by original URL")
	}

	if shortURL != "short1" {
		t.Fatalf("Expected short1, got %s", shortURL)
	}
}

func TestURLStorage_DeleteURLs(t *testing.T) {
	storageInstance := storage.NewURLStorage()

	storageInstance.AddURL("short1", "http://example1.com", "user1")
	storageInstance.AddURL("short2", "http://example2.com", "user1")
	storageInstance.AddURL("short3", "http://example3.com", "user2")

	err := storageInstance.DeleteURLs([]string{"short1", "short2", "short3"}, "user1")
	if err != nil {
		t.Fatalf("DeleteURLs failed: %v", err)
	}

	_, exists, isDeleted := storageInstance.GetURL("short1")
	if !exists {
		t.Fatal("URL should still exist after deletion")
	}
	if !isDeleted {
		t.Fatal("URL should be marked as deleted")
	}

	_, exists, isDeleted = storageInstance.GetURL("short3")
	if !exists {
		t.Fatal("URL should exist")
	}
	if isDeleted {
		t.Fatal("URL should not be deleted (different user)")
	}
}

func TestURLStorage_AddURLs(t *testing.T) {
	storageInstance := storage.NewURLStorage()

	urls := map[string]string{
		"short1": "http://example1.com",
		"short2": "http://example2.com",
		"short3": "http://example3.com",
	}

	err := storageInstance.AddURLs(urls, "user1")
	if err != nil {
		t.Fatalf("AddURLs failed: %v", err)
	}

	for shortURL, expectedOriginal := range urls {
		originalURL, exists, isDeleted := storageInstance.GetURL(shortURL)
		if !exists {
			t.Fatalf("URL %s not found after batch add", shortURL)
		}
		if originalURL != expectedOriginal {
			t.Fatalf("Expected %s, got %s", expectedOriginal, originalURL)
		}
		if isDeleted {
			t.Fatalf("URL %s should not be deleted", shortURL)
		}
	}
}

func TestURLStorage_Count(t *testing.T) {
	storageInstance := storage.NewURLStorage()

	if count := storageInstance.Count(); count != 0 {
		t.Fatalf("Expected 0, got %d", count)
	}

	storageInstance.AddURL("short1", "http://example1.com", "user1")
	storageInstance.AddURL("short2", "http://example2.com", "user1")

	if count := storageInstance.Count(); count != 2 {
		t.Fatalf("Expected 2, got %d", count)
	}
}

func TestURLStorage_IterateURLs(t *testing.T) {
	storageInstance := storage.NewURLStorage()

	storageInstance.AddURL("short1", "http://example1.com", "user1")
	storageInstance.AddURL("short2", "http://example2.com", "user1")

	count := 0
	storageInstance.IterateURLs(func(shortURL, originalURL string) {
		count++
		if shortURL != "short1" && shortURL != "short2" {
			t.Fatalf("Unexpected shortURL: %s", shortURL)
		}
	})

	if count != 2 {
		t.Fatalf("Expected 2 iterations, got %d", count)
	}
}

func TestURLStorage_Ping(t *testing.T) {
	storageInstance := storage.NewURLStorage()

	err := storageInstance.Ping()
	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
}

func TestURLStorage_Close(t *testing.T) {
	storageInstance := storage.NewURLStorage()

	err := storageInstance.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}
