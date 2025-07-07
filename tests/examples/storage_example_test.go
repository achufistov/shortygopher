package shortygopher_test

import (
	"fmt"
	"log"

	"github.com/achufistov/shortygopher.git/internal/app/storage"
)

// ExampleNewURLStorage demonstrates creating a new in-memory storage.
func ExampleNewURLStorage() {
	// Create new storage
	s := storage.NewURLStorage()

	// Check that storage is empty
	allURLs := s.GetAllURLs()
	fmt.Printf("URLs in new storage: %d\n", len(allURLs))

	// Output:
	// URLs in new storage: 0
}

// ExampleURLStorage_AddURL demonstrates adding a URL to storage.
func ExampleURLStorage_AddURL() {
	s := storage.NewURLStorage()

	// Add URL
	err := s.AddURL("abc123", "https://example.com", "user1")
	if err != nil {
		log.Fatal(err)
	}

	// Check that URL was added
	originalURL, exists, isDeleted := s.GetURL("abc123")
	fmt.Printf("URL exists: %v\n", exists)
	fmt.Printf("URL deleted: %v\n", isDeleted)
	fmt.Printf("Original URL: %s\n", originalURL)

	// Output:
	// URL exists: true
	// URL deleted: false
	// Original URL: https://example.com
}

// ExampleURLStorage_GetURL demonstrates getting a URL from storage.
func ExampleURLStorage_GetURL() {
	s := storage.NewURLStorage()

	// Add URL
	s.AddURL("test123", "https://test.com", "user1")

	// Get URL
	originalURL, exists, isDeleted := s.GetURL("test123")
	if exists && !isDeleted {
		fmt.Printf("Found URL: %s\n", originalURL)
	}

	// Try to get non-existent URL
	_, exists, _ = s.GetURL("nonexistent")
	fmt.Printf("Nonexistent URL found: %v\n", exists)

	// Output:
	// Found URL: https://test.com
	// Nonexistent URL found: false
}

// ExampleURLStorage_GetURLsByUser demonstrates getting all URLs for a user.
func ExampleURLStorage_GetURLsByUser() {
	s := storage.NewURLStorage()

	// Add several URLs for different users
	s.AddURL("url1", "https://example1.com", "user1")
	s.AddURL("url2", "https://example2.com", "user1")
	s.AddURL("url3", "https://example3.com", "user2")

	// Get URLs for user1
	urls, err := s.GetURLsByUser("user1")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("URLs for user1: %d\n", len(urls))
	fmt.Printf("URL1 value: %s\n", urls["url1"])

	// Output:
	// URLs for user1: 2
	// URL1 value: https://example1.com
}

// ExampleURLStorage_AddURLs demonstrates batch adding URLs.
func ExampleURLStorage_AddURLs() {
	s := storage.NewURLStorage()

	// Create batch URLs
	urls := map[string]string{
		"batch1": "https://batch1.com",
		"batch2": "https://batch2.com",
		"batch3": "https://batch3.com",
	}

	// Add all URLs at once
	err := s.AddURLs(urls, "batch-user")
	if err != nil {
		log.Fatal(err)
	}

	// Check that all URLs were added
	allURLs := s.GetAllURLs()
	fmt.Printf("Total URLs after batch: %d\n", len(allURLs))

	// Check specific URL
	originalURL, exists, _ := s.GetURL("batch2")
	fmt.Printf("Batch2 exists: %v, URL: %s\n", exists, originalURL)

	// Output:
	// Total URLs after batch: 3
	// Batch2 exists: true, URL: https://batch2.com
}

// ExampleURLStorage_DeleteURLs demonstrates deleting URLs.
func ExampleURLStorage_DeleteURLs() {
	s := storage.NewURLStorage()

	// Add URLs
	s.AddURL("delete1", "https://delete1.com", "user1")
	s.AddURL("delete2", "https://delete2.com", "user1")
	s.AddURL("keep1", "https://keep1.com", "user2")

	// Delete URLs for user1
	urlsToDelete := []string{"delete1", "delete2", "keep1"} // keep1 should not be deleted
	err := s.DeleteURLs(urlsToDelete, "user1")
	if err != nil {
		log.Fatal(err)
	}

	// Check URL status
	_, exists1, isDeleted1 := s.GetURL("delete1")
	_, exists2, isDeleted2 := s.GetURL("keep1")

	fmt.Printf("Delete1 - exists: %v, deleted: %v\n", exists1, isDeleted1)
	fmt.Printf("Keep1 - exists: %v, deleted: %v\n", exists2, isDeleted2)

	// Output:
	// Delete1 - exists: true, deleted: true
	// Keep1 - exists: true, deleted: false
}

// ExampleURLStorage_GetShortURLByOriginalURL demonstrates finding a short URL by original URL.
func ExampleURLStorage_GetShortURLByOriginalURL() {
	s := storage.NewURLStorage()

	// Add URL
	s.AddURL("find123", "https://findme.com", "user1")

	// Find short URL by original URL
	shortURL, found := s.GetShortURLByOriginalURL("https://findme.com")
	fmt.Printf("Found: %v\n", found)
	if found {
		fmt.Printf("Short URL: %s\n", shortURL)
	}

	// Search for non-existent URL
	_, found = s.GetShortURLByOriginalURL("https://notfound.com")
	fmt.Printf("Not found: %v\n", !found)

	// Output:
	// Found: true
	// Short URL: find123
	// Not found: true
}
